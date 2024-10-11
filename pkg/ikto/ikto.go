package ikto

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/valyentdev/ikto/internal/network"
	"github.com/valyentdev/ikto/internal/state"
	"github.com/valyentdev/ikto/pkg/types"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Config struct {
	Name string

	AdvertiseAddress net.IP
	PrivateAddress   net.IP
	MeshIPNet        net.IPNet
	HostPrefixLength int

	WGDevName      string
	WGPort         int
	PrivateKeyPath string

	NatsCreds string
	NatsURL   string
	NatsKV    string
}

func (c *Config) getPrivateCIDR() string {
	return fmt.Sprintf("%s/%d", c.PrivateAddress.String(), c.HostPrefixLength)
}

type Ikto struct {
	config Config
	nc     *nats.Conn
	js     jetstream.JetStream
	kv     jetstream.KeyValue

	store *state.Store
	self  types.Peer
	state *state.SyncedState
	wg    *network.WGDevice
}

var ErrAddressAlreadyInUse = fmt.Errorf("address already in use")

func NewIkto(c *Config) (*Ikto, error) {
	privateKeyFile, err := os.ReadFile(c.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	privateKey, err := wgtypes.ParseKey(strings.TrimSpace(string(privateKeyFile)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey := privateKey.PublicKey()

	self := types.Peer{
		Name:             c.Name,
		PublicKey:        types.PublicKey(publicKey),
		AdvertiseAddress: c.AdvertiseAddress.String(),
		AllowedIP:        c.getPrivateCIDR(),
		WGPort:           c.WGPort,
	}

	slog.Info("Starting with self config", "name", self.Name, "public_key", self.PublicKey.String(), "advertise_address", self.AdvertiseAddress, "allowed_ip", self.AllowedIP, "wg_port", self.WGPort, "wg_dev_name", c.WGDevName)

	wg, err := network.New(fmt.Sprintf(c.WGDevName), c.WGPort, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create wg service: %w", err)
	}

	err = wg.Ensure()
	if err != nil {
		return nil, fmt.Errorf("failed to ensure wireguard device: %w", err)
	}

	err = wg.InitConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to init wireguard config: %w", err)
	}

	meshOnes, _ := c.MeshIPNet.Mask.Size()

	err = wg.SetAddr(net.IPNet{
		IP:   c.PrivateAddress,
		Mask: net.CIDRMask(meshOnes, len(c.PrivateAddress)*8),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set address: %w", err)
	}

	nc, err := nats.Connect(c.NatsURL, nats.UserCredentials(c.NatsCreds, c.NatsCreds))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream client: %w", err)
	}

	kv, err := js.KeyValue(context.Background(), c.NatsKV)
	if err != nil {
		return nil, fmt.Errorf("failed to get key value store: %w", err)
	}

	store := state.NewStore(kv)

	state := state.New(state.Config{
		KV:         kv,
		IgnorePeer: self.PublicKey,

		OnPeerPut: func(peer types.Peer) {
			err := wg.AddPeer(peer)
			if err != nil {
				fmt.Printf("failed to add peer: %v", err)
			}
		},
		OnPeerDelete: func(peer types.Peer) {
			err := wg.RemovePeer(peer.PublicKey.WG())
			if err != nil {
				fmt.Printf("failed to remove peer: %v", err)
			}
		},
		OnInitPeers: func(m map[string]types.Peer) {
			peers := make([]types.Peer, 0, len(m))
			for _, peer := range m {
				peers = append(peers, peer)
			}

			err := wg.ReplacePeers(peers)
			if err != nil {
				fmt.Printf("failed to replace peers: %v", err)
			}
		},
	})
	return &Ikto{
		config: *c,
		nc:     nc,
		js:     js,
		kv:     kv,

		store: store,
		self:  self,
		state: state,
		wg:    wg,
	}, nil
}

func (i *Ikto) init() error {
	previous, revision, err := i.store.GetPeer(context.Background(), i.self.AllowedIP)
	if err != nil && err != jetstream.ErrKeyNotFound {
		return fmt.Errorf("failed to get self: % w", err)
	}
	if err == jetstream.ErrKeyNotFound {
		_, err := i.store.CreatePeer(context.Background(), i.self)
		if err != nil {
			return fmt.Errorf("failed to create self: %w", err)
		}
		return nil
	}

	if previous.PublicKey != i.self.PublicKey {
		return ErrAddressAlreadyInUse
	}

	_, err = i.store.UpdatePeer(context.Background(), i.self, revision)
	if err != nil {
		return fmt.Errorf("failed to update self: %w", err)
	}

	return nil
}

func (i *Ikto) Start() error {
	if err := i.init(); err != nil {
		return fmt.Errorf("failed to init: %w", err)
	}

	err := i.state.Start(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start state: %w", err)
	}

	return nil
}

func (i *Ikto) Stop() {
	slog.Info("stopping")
	i.state.Stop()
	i.nc.Close()
}

func (i *Ikto) Leave() {
	slog.Error("leaving network")
}

func (i *Ikto) Self() types.Peer {
	return i.self
}

func (i *Ikto) Peers() []types.Peer {
	return i.state.ListPeers()
}
