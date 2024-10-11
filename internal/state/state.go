package state

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/valyentdev/ikto/pkg/types"
)

type SyncedState struct {
	stop   chan struct{}
	config Config
	peers  map[string]types.Peer
	mutex  sync.RWMutex
}

type Config struct {
	KV         jetstream.KeyValue
	IgnorePeer types.PublicKey

	OnPeerPut    func(peer types.Peer)
	OnPeerDelete func(peer types.Peer)
	OnInitPeers  func(map[string]types.Peer)
}

func New(config Config) *SyncedState {
	if config.OnPeerPut == nil {
		config.OnPeerPut = func(peer types.Peer) {}
	}

	if config.OnPeerDelete == nil {
		config.OnPeerDelete = func(peer types.Peer) {}
	}

	if config.OnInitPeers == nil {
		config.OnInitPeers = func(peers map[string]types.Peer) {}
	}

	return &SyncedState{
		stop:   make(chan struct{}),
		config: config,
	}
}

func (w *SyncedState) Start(ctx context.Context) error {
	kv := w.config.KV
	sub := "peers.*"
	watcher, err := kv.Watch(ctx, sub)
	if err != nil {
		return fmt.Errorf("failed to watch: %w", err)
	}

	slog.Info("Initializing watcher", "subject", sub)
	updates := watcher.Updates()
	w.init(updates)

	go func() {
		w.sync(updates)
	}()

	return nil
}

func (w *SyncedState) init(entries <-chan jetstream.KeyValueEntry) {
	for entry := range entries {
		if entry == nil {
			break
		}

		if entry.Operation() != jetstream.KeyValuePut {
			continue
		}

		peer, err := readPeer(entry.Value())
		if err != nil {
			slog.Error("failed to read peer", "error", err)
			continue
		}

		if peer.PublicKey == w.config.IgnorePeer {
			continue
		}

		w.peers[peer.PrivateCIDR] = peer
	}
}

func (w *SyncedState) sync(entries <-chan jetstream.KeyValueEntry) {
	for entry := range entries {
		if entry == nil {
			continue
		}

		key := entry.Key()

		switch entry.Operation() {
		case jetstream.KeyValuePut:
			peer, err := readPeer(entry.Value())
			if err != nil {
				slog.Error("failed to read peer", "error", err)
				continue
			}
			w.onPeerPut(key, peer)
		case jetstream.KeyValueDelete:
			w.onPeerDelete(key)
		case jetstream.KeyValuePurge:
			w.onPeerDelete(key)
		}
	}
}

func (w *SyncedState) onPeerPut(key string, peer types.Peer) {
	if peer.PublicKey == w.config.IgnorePeer {
		return
	}
	w.mutex.Lock()
	w.peers[key] = peer
	w.mutex.Unlock()
	w.config.OnPeerPut(peer)

}

func (w *SyncedState) onPeerDelete(key string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	peer, ok := w.peers[key]
	if !ok {
		return
	}

	delete(w.peers, key)
	w.config.OnPeerDelete(peer)
}

func (w *SyncedState) Stop() {
	close(w.stop)
}

func (s *SyncedState) ListPeers() []types.Peer {
	s.mutex.RLock()
	peers := make([]types.Peer, 0, len(s.peers))
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	s.mutex.RUnlock()
	return peers
}

func readPeer(data []byte) (types.Peer, error) {
	var p types.Peer
	err := json.Unmarshal(data, &p)
	if err != nil {
		return types.Peer{}, err
	}

	return p, nil

}
