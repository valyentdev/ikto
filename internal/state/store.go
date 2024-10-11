package state

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/valyentdev/ikto/pkg/types"
)

type Store struct {
	kv jetstream.KeyValue
}

func NewStore(kv jetstream.KeyValue) *Store {
	return &Store{
		kv: kv,
	}
}

func (s *Store) CreatePeer(ctx context.Context, peer types.Peer) (uint64, error) {
	bytes, err := json.Marshal(peer)
	if err != nil {
		return 0, err
	}

	revision, err := s.kv.Create(ctx, getKey(peer.AllowedIP), bytes)
	if err != nil {
		return 0, err
	}

	return revision, nil
}

func getKey(ip string) string {
	return "peers." + base64.URLEncoding.EncodeToString([]byte(ip))
}

func (s *Store) GetPeer(ctx context.Context, ip string) (types.Peer, uint64, error) {
	entry, err := s.kv.Get(ctx, getKey(ip))
	if err != nil {
		return types.Peer{}, 0, err
	}

	var peer types.Peer
	if err := json.Unmarshal(entry.Value(), &peer); err != nil {
		return types.Peer{}, 0, err
	}

	return peer, entry.Revision(), nil
}

func (s *Store) UpdatePeer(ctx context.Context, peer types.Peer, revision uint64) (uint64, error) {
	bytes, err := json.Marshal(peer)
	if err != nil {
		return 0, err
	}

	r, err := s.kv.Update(ctx, getKey(peer.AllowedIP), bytes, revision)
	if err != nil {
		return 0, err
	}

	return r, err
}

func (s *Store) DeletePeer(ctx context.Context, ip string, revision uint64) error {
	return s.kv.Delete(ctx, getKey(ip), jetstream.LastRevision(revision))
}
