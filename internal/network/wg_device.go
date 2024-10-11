package network

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/valyentdev/ikto/pkg/types"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type WGDevice struct {
	name       string
	port       int
	wg         *wgctrl.Client
	privateKey wgtypes.Key
}

func New(name string, port int, privateKey wgtypes.Key) (*WGDevice, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create wgctrl client: %w", err)
	}

	return &WGDevice{
		wg:         client,
		name:       name,
		port:       port,
		privateKey: privateKey,
	}, nil
}

func (d *WGDevice) Ensure() error {
	link, err := netlink.LinkByName(d.name)
	if err != nil {
		switch err.(type) {
		case netlink.LinkNotFoundError:
			link = &netlink.Wireguard{
				LinkAttrs: netlink.LinkAttrs{
					Name: d.name,
				},
			}

			if err := netlink.LinkAdd(link); err != nil {
				return err
			}
		default:
			return fmt.Errorf("failed to check link %s: %w", d.name, err)
		}
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to set link up: %w", err)
	}

	return nil
}

func (d WGDevice) SetAddr(ipnet net.IPNet) error {
	link, err := netlink.LinkByName(d.name)
	if err != nil {
		return fmt.Errorf("failed to get link: %w", err)
	}

	newAddr := &netlink.Addr{
		IPNet: &ipnet,
	}

	if err := netlink.AddrReplace(link, newAddr); err != nil {
		return fmt.Errorf("failed to add addr: %w", err)
	}
	list, err := netlink.AddrList(link, netlink.FAMILY_V6)
	if err != nil {
		return fmt.Errorf("failed to list addrs: %w", err)
	}

	for _, addr := range list {
		if !addr.Equal(*newAddr) {
			if err := netlink.AddrDel(link, &addr); err != nil {
				slog.Error("failed to delete addr", "error", err)
			}
		}
	}

	return nil
}

func (m *WGDevice) InitConfig() error {
	return m.wg.ConfigureDevice(m.name, wgtypes.Config{
		PrivateKey: &m.privateKey,
		ListenPort: &m.port,
	})
}

func (m *WGDevice) Remove() error {
	link, err := netlink.LinkByName("wg-ravel")
	if err != nil {
		switch err.(type) {
		case netlink.LinkNotFoundError:
			return nil
		default:
			return fmt.Errorf("failed to get link: %w", err)
		}
	}

	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	return nil

}

func (m *WGDevice) RemovePeer(publicKey wgtypes.Key) error {
	return m.wg.ConfigureDevice(m.name, wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: publicKey,
				Remove:    true,
			},
		},
	})
}

func (m *WGDevice) AddPeer(peer types.Peer) error {
	return m.configurePeers([]types.Peer{peer}, false)
}

func (m *WGDevice) ReplacePeers(peers []types.Peer) error {
	return m.configurePeers(peers, true)
}

func (m *WGDevice) configurePeers(peers []types.Peer, replacePeers bool) error {
	peerConfigs := make([]wgtypes.PeerConfig, 0, len(peers))
	for _, member := range peers {
		peerConfig, err := member.WGPeerConfig()
		if err != nil {
			slog.Error("failed to get peer config", "error", err)
			return err
		}

		peerConfigs = append(peerConfigs, peerConfig)
	}

	return m.wg.ConfigureDevice(m.name, wgtypes.Config{
		Peers:        peerConfigs,
		ReplacePeers: replacePeers,
	})
}
