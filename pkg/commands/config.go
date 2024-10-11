package commands

import (
	"fmt"
	"net"

	"github.com/valyentdev/ikto/pkg/ikto"
)

type Config struct {
	Name string `json:"name"`

	AdvertiseAddress string `json:"advertise_address"`
	PrivateAddress   string `json:"private_address"`
	HostPrefixLength int    `json:"subnet_prefix"`
	MeshIPNet        string `json:"mesh_cidr"`

	WGDevName      string `json:"wg_dev_name"`
	WGPort         int    `json:"wg_port"`
	PrivateKeyPath string `json:"private_key_path"`

	NatsCreds string `json:"nats_creds"`
	NatsURL   string `json:"nats_url"`
	NatsKV    string `json:"nats_kv"`
}

func (c *Config) Validate() (ikto.Config, error) {
	_, ipnet, err := net.ParseCIDR(c.MeshIPNet)
	if err != nil {
		return ikto.Config{}, err
	}

	advertiseAddress := net.ParseIP(c.AdvertiseAddress)
	if advertiseAddress == nil {
		return ikto.Config{}, fmt.Errorf("advertise address is required")
	}

	privateAddress := net.ParseIP(c.PrivateAddress)
	if privateAddress == nil {
		return ikto.Config{}, fmt.Errorf("private address is invalid")
	}
	if len(ipnet.IP) == 4 {
		privateAddress = privateAddress.To4()
	} else {
		privateAddress = privateAddress.To16()
	}

	fmt.Println(privateAddress)
	if !ipnet.Contains(privateAddress) {
		return ikto.Config{}, fmt.Errorf("private address is not in mesh network")
	}

	return ikto.Config{
		Name: c.Name,

		NatsCreds: c.NatsCreds,
		NatsURL:   c.NatsURL,
		NatsKV:    c.NatsKV,

		AdvertiseAddress: advertiseAddress,
		PrivateAddress:   privateAddress,
		MeshIPNet:        *ipnet,
		HostPrefixLength: c.HostPrefixLength,

		WGDevName:      c.WGDevName,
		WGPort:         c.WGPort,
		PrivateKeyPath: c.PrivateKeyPath,
	}, nil
}

func DefaultConfig() *Config {
	return &Config{
		Name:             "",
		NatsCreds:        "",
		NatsURL:          "nats://",
		NatsKV:           "ikto-mesh",
		PrivateKeyPath:   "",
		AdvertiseAddress: "",
		MeshIPNet:        "",
		WGDevName:        "wg-ikto",
		WGPort:           51820,
	}
}
