package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	etcd "go.etcd.io/etcd/client/v3"
)

type SessionStoreConfig struct {
	Etcd *etcd.Client
}

func NewSessionStoreClient(config SessionStoreConfig) *SessionStore {
	return &SessionStore{
		etcd: config.Etcd,
	}
}

type SessionStore struct {
	etcd *etcd.Client
}

func (s *SessionStore) GetHandoffSession(ctx context.Context, key uint32) (*HandoffSession, error) {
	res, err := s.etcd.Get(ctx, fmt.Sprintf("handoff_sessions:%d", key))
	if err != nil {
		return nil, err
	}
	if len(res.Kvs) == 1 {
		return nil, errors.New("invalid key")
	}
	var presence HandoffSession
	if err := json.Unmarshal(res.Kvs[0].Value, &presence); err != nil {
		return nil, err
	}
	return &presence, nil
}

func (s *SessionStore) SetHandoffSession(
	ctx context.Context,
	key uint32,
	presence *HandoffSession,
) error {
	d, err := json.Marshal(presence)
	if err != nil {
		return err
	}
	_, err = s.etcd.Put(ctx, fmt.Sprintf("handoff_sessions:%d", key), string(d))
	return err
}

func (s *SessionStore) DeleteHandoffSession(ctx context.Context, key uint32) error {
	_, err := s.etcd.Delete(ctx, fmt.Sprintf("handoff_sessions:%d", key))
	return err
}

type HandoffSession struct {
	Identifier   string
	SessionToken string
}
