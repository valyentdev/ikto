package types

import (
	"fmt"
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Peer struct {
	Name             string    `json:"name"`
	PublicKey        PublicKey `json:"public_key"`
	AdvertiseAddress string    `json:"advertise_address"`
	PrivateCIDR      string    `json:"private_cidr"`
	WGPort           int       `json:"wg_port"`
}

func (p *Peer) WGPeerConfig() (wgtypes.PeerConfig, error) {
	_, ipnet, err := net.ParseCIDR(p.PrivateCIDR)
	if err != nil {
		return wgtypes.PeerConfig{}, err
	}

	advertiseAddress := net.ParseIP(p.AdvertiseAddress)
	if advertiseAddress == nil {
		return wgtypes.PeerConfig{}, fmt.Errorf("invalid public address: %s", p.AdvertiseAddress)
	}

	return wgtypes.PeerConfig{
		PublicKey: p.PublicKey.WG(),
		Endpoint: &net.UDPAddr{
			IP:   advertiseAddress,
			Port: p.WGPort,
		},
		AllowedIPs: []net.IPNet{*ipnet},
	}, nil
}
