package utils

import (
	"io"
	"log"
)

type writeLogger struct {
	prefix string
	w      io.Writer
}

func (l *writeLogger) Write(p []byte) (n int, err error) {
	n, err = l.w.Write(p)
	if err != nil {
		log.Printf("%s %v: %v", l.prefix, p[0:n], err)
	} else {
		log.Printf("%s %v", l.prefix, p[0:n])
	}
	return
}

// NewWriteLogger returns a writer that behaves like w except
// that it logs (using log.Printf) each write to standard error,
// printing the prefix and the hexadecimal data written.
func NewWriteLogger(prefix string, w io.Writer) io.Writer {
	return &writeLogger{prefix, w}
}

type readLogger struct {
	prefix string
	r      io.Reader
}

func (l *readLogger) Read(p []byte) (n int, err error) {
	n, err = l.r.Read(p)
	if err != nil {
		log.Printf("%s %v: %v", l.prefix, p[0:n], err)
	} else {
		log.Printf("%s %v", l.prefix, p[0:n])
	}
	return
}

// NewReadLogger returns a reader that behaves like r except
// that it logs (using log.Printf) each read to standard error,
// printing the prefix and the hexadecimal data read.
func NewReadLogger(prefix string, r io.Reader) io.Reader {
	return &readLogger{prefix, r}
}
