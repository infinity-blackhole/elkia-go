package nostale

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ServerContextKey is a context key. It can be used in NosTale
	// handlers with Context.Value to access the server that
	// started the handler. The associated value will be of
	// type *Server.
	ServerContextKey = &contextKey{"nostale-server"}
)

// A Handler is a simple interface for handling NosTale messages.
type Handler interface {
	ServeNosTale(net.Conn)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as NosTale handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(net.Conn)

// ServeNosTale calls f(w, r).
func (f HandlerFunc) ServeNosTale(c net.Conn) {
	f(c)
}

// A Server defines parameters for running an NosTale server.
// The zero value for Server is a valid configuration.
type ServerConfig struct {
	// Addr optionally specifies the TCP address for the server to listen on,
	// in the form "host:port".
	// The service names are defined in RFC 6335 and assigned by IANA.
	// See net.Dial for details of the address format.
	Addr string

	Handler Handler // handler to invoke.

	// BaseContext optionally specifies a function that returns
	// the base context for incoming messages on this server.
	// The provided Listener is the specific Listener that's
	// about to start accepting messages.
	// If BaseContext is nil, the default is context.Background().
	// If non-nil, it must return a non-nil context.
	BaseContext func(net.Listener) context.Context

	// ConnContext optionally specifies a function that modifies
	// the context used for a new connection c. The provided ctx
	// is derived from the base context and has a ServerContextKey
	// value.
	ConnContext func(ctx context.Context, c net.Conn) context.Context
}

func NewServer(cfg ServerConfig) *Server {
	return &Server{
		addr:        cfg.Addr,
		handler:     cfg.Handler,
		baseContext: cfg.BaseContext,
		connContext: cfg.ConnContext,
		listeners:   make(map[*net.Listener]struct{}),
		activeConn:  make(map[net.Conn]struct{}),
	}
}

type Server struct {
	// addr optionally specifies the TCP address for the server to listen on,
	// in the form "host:port".
	// The service names are defined in RFC 6335 and assigned by IANA.
	// See net.Dial for details of the address format.
	addr string

	handler Handler // handler to invoke.

	// baseContext optionally specifies a function that returns
	// the base context for incoming messages on this server.
	// The provided Listener is the specific Listener that's
	// about to start accepting messages.
	// If baseContext is nil, the default is context.Background().
	// If non-nil, it must return a non-nil context.
	baseContext func(net.Listener) context.Context

	// connContext optionally specifies a function that modifies
	// the context used for a new connection c. The provided ctx
	// is derived from the base context and has a ServerContextKey
	// value.
	connContext func(ctx context.Context, c net.Conn) context.Context

	inShutdown atomic.Bool // true when server is in shutdown

	mu         sync.Mutex
	listeners  map[*net.Listener]struct{}
	activeConn map[net.Conn]struct{}
	doneChan   chan struct{}

	listenerGroup sync.WaitGroup
}

func (s *Server) getDoneChan() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getDoneChanLocked()
}

func (s *Server) getDoneChanLocked() chan struct{} {
	if s.doneChan == nil {
		s.doneChan = make(chan struct{})
	}
	return s.doneChan
}

func (s *Server) closeDoneChanLocked() {
	ch := s.getDoneChanLocked()
	select {
	case <-ch:
		// Already closed. Don't close again.
	default:
		// Safe to close here. We're the only closer, guarded
		// by s.mu.
		close(ch)
	}
}

// Close immediately closes all active net.Listeners and any
// connections. For a graceful shutdown, use Shutdown.
//
// Close returns any error returned from closing the Server's
// underlying Listener(s).
func (srv *Server) Close() error {
	srv.inShutdown.Store(true)
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.closeDoneChanLocked()
	err := srv.closeListenersLocked()

	// Unlock srv.mu while waiting for listenerGroup.
	// The group Add and Done calls are made with srv.mu held,
	// to avoid adding a new listener in the window between
	// us setting inShutdown above and waiting here.
	srv.mu.Unlock()
	srv.listenerGroup.Wait()
	srv.mu.Lock()

	for c := range srv.activeConn {
		c.Close()
		delete(srv.activeConn, c)
	}
	return err
}

// shutdownPollIntervalMax is the max polling interval when checking
// quiescence during Server.Shutdown. Polling starts with a small
// interval and backs off to the max.
// Ideally we could find a solution that doesn't involve polling,
// but which also doesn't have a high runtime cost (and doesn't
// involve any contentious mutexes), but that is left as an
// exercise for the reader.
const shutdownPollIntervalMax = 500 * time.Millisecond

// Shutdown gracefully shuts down the server without interrupting any
// active connections. Shutdown works by first closing all open
// listeners, then closing all idle connections, and then waiting
// indefinitely for connections to return to idle and then shut down.
// If the provided context expires before the shutdown is complete,
// Shutdown returns the context's error, otherwise it returns any
// error returned from closing the Server's underlying Listener(s).
//
// When Shutdown is called, Serve, ListenAndServe, and
// ListenAndServeTLS immediately return ErrServerClosed. Make sure the
// program doesn't exit and waits instead for Shutdown to return.
//
// Once Shutdown has been called on a server, it may not be reused;
// future calls to methods such as Serve will return ErrServerClosed.
func (srv *Server) Shutdown(ctx context.Context) error {
	srv.inShutdown.Store(true)

	srv.mu.Lock()
	lnerr := srv.closeListenersLocked()
	srv.closeDoneChanLocked()
	srv.mu.Unlock()
	srv.listenerGroup.Wait()

	pollIntervalBase := time.Millisecond
	nextPollInterval := func() time.Duration {
		// Add 10% jitter.
		interval := pollIntervalBase + time.Duration(rand.Intn(int(pollIntervalBase/10)))
		// Double and clamp for next time.
		pollIntervalBase *= 2
		if pollIntervalBase > shutdownPollIntervalMax {
			pollIntervalBase = shutdownPollIntervalMax
		}
		return interval
	}

	timer := time.NewTimer(nextPollInterval())
	defer timer.Stop()
	for {
		if srv.closeIdleConns() {
			return lnerr
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(nextPollInterval())
		}
	}
}

// closeIdleConns closes all idle connections and reports whether the
// server is quiescent.
func (s *Server) closeIdleConns() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	quiescent := true
	for c := range s.activeConn {
		// TODO: Add state abstraction and verify the current conn state
		c.Close()
		delete(s.activeConn, c)
	}
	return quiescent
}

func (s *Server) closeListenersLocked() error {
	var err error
	for ln := range s.listeners {
		if cerr := (*ln).Close(); cerr != nil && err == nil {
			err = cerr
		}
	}
	return err
}

// ListenAndServe listens on the TCP network address addr and then calls
// Serve with handler to handle messages on incoming connections.
// Accepted connections are configured to enable TCP keep-alives.
//
// The handler is typically nil, in which case the DefaultServeMux is used.
//
// ListenAndServe always returns a non-nil error.
func ListenAndServe(addr string, handler Handler) error {
	server := &Server{addr: addr, handler: handler}
	return server.ListenAndServe()
}

// ListenAndServe listens on the TCP network address srv.Addr and then
// calls Serve to handle messages on incoming connections.
// Accepted connections are configured to enable TCP keep-alives.
//
// ListenAndServe always returns a non-nil error. After Shutdown or Close,
// the returned error is ErrServerClosed.
func (srv *Server) ListenAndServe() error {
	if srv.inShutdown.Load() {
		return ErrServerClosed
	}
	ln, err := net.Listen("tcp", srv.addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

// ErrServerClosed is returned by the Server's Serve and ListenAndServe
// methods after a call to Shutdown or Close.
var ErrServerClosed = errors.New("nostale: Server closed")

// Serve accepts incoming connections on the Listener l, creating a
// new service goroutine for each. The service goroutines read messages and
// then call srv.Handler to reply to them.
//
// Serve always returns a non-nil error and closes l.
// After Shutdown or Close, the returned error is ErrServerClosed.
func (srv *Server) Serve(l net.Listener) error {
	origListener := l
	l = &onceCloseListener{Listener: l}
	defer l.Close()

	if !srv.trackListener(&l) {
		return ErrServerClosed
	}
	defer srv.untrackListener(&l)

	baseCtx := context.Background()
	if srv.baseContext != nil {
		baseCtx = srv.baseContext(origListener)
		if baseCtx == nil {
			panic("BaseContext returned a nil context")
		}
	}

	ctx := context.WithValue(baseCtx, ServerContextKey, srv)
	for {
		rw, err := l.Accept()
		if err != nil {
			select {
			case <-srv.getDoneChan():
				return ErrServerClosed
			default:
			}
			return err
		}
		connCtx := ctx
		if cc := srv.connContext; cc != nil {
			connCtx = cc(connCtx, rw)
			if connCtx == nil {
				panic("ConnContext returned nil")
			}
		}
		go srv.handler.ServeNosTale(rw)
	}
}

// onceCloseListener wraps a net.Listener, protecting it from
// multiple Close calls.
type onceCloseListener struct {
	net.Listener
	once     sync.Once
	closeErr error
}

func (oc *onceCloseListener) Close() error {
	oc.once.Do(oc.close)
	return oc.closeErr
}

func (oc *onceCloseListener) close() { oc.closeErr = oc.Listener.Close() }

// trackListener adds or removes a net.Listener to the set of tracked
// listeners.
//
// We store a pointer to interface in the map set, in case the
// net.Listener is not comparable. This is safe because we only call
// trackListener via Serve and can track+defer untrack the same
// pointer to local variable there. We never need to compare a
// Listener from another caller.
//
// It reports whether the server is still up (not Shutdown or Closed).
func (s *Server) trackListener(ln *net.Listener) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inShutdown.Load() {
		return false
	}
	s.listeners[ln] = struct{}{}
	s.listenerGroup.Add(1)
	return true
}

func (s *Server) untrackListener(ln *net.Listener) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.listeners, ln)
	s.listenerGroup.Done()
	return true
}

func (s *Server) trackConn(c net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeConn == nil {
		s.activeConn = make(map[net.Conn]struct{})
	}
	s.activeConn[c] = struct{}{}
}

func (s *Server) untrackConn(c net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeConn == nil {
		s.activeConn = make(map[net.Conn]struct{})
	}
	delete(s.activeConn, c)
}
