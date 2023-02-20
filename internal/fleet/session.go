package fleet

import (
	"context"
	"errors"
	"fmt"

	fleetv1alpha1 "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/sirupsen/logrus"
	etcd "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
)

type SessionStoreConfig struct {
	Etcd etcd.KV
}

func NewSessionStoreClient(config SessionStoreConfig) *SessionStore {
	return &SessionStore{
		etcd: config.Etcd,
	}
}

type SessionStore struct {
	etcd etcd.KV
}

func (s *SessionStore) GetHandoffSession(
	ctx context.Context,
	key uint32,
) (*fleetv1alpha1.Handoff, error) {
	res, err := s.etcd.Get(ctx, fmt.Sprintf("handoff_sessions:%d", key))
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got %d handoff sessions", len(res.Kvs))
	if len(res.Kvs) == 1 {
		return nil, errors.New("invalid key")
	}
	var presence fleetv1alpha1.Handoff
	if err := proto.Unmarshal(res.Kvs[0].Value, &presence); err != nil {
		return nil, err
	}
	return &presence, nil
}

func (s *SessionStore) SetHandoffSession(
	ctx context.Context,
	key uint32,
	handoff *fleetv1alpha1.Handoff,
) error {
	d, err := proto.Marshal(handoff)
	if err != nil {
		return err
	}
	res, err := s.etcd.Put(ctx, fmt.Sprintf("handoff_sessions:%d", key), string(d))
	if err != nil {
		return err
	}
	logrus.Debugf("fleet: set handoff session: %v", res)
	return nil
}

func (s *SessionStore) DeleteHandoffSession(ctx context.Context, key uint32) error {
	res, err := s.etcd.Delete(ctx, fmt.Sprintf("handoff_sessions:%d", key))
	if err != nil {
		return err
	}
	logrus.Debugf("fleet: deleted handoff session: %v", res)
	return nil
}

type HandoffSession struct {
	Identifier   string
	SessionToken string
}
