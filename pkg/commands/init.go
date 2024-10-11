package commands

import (
	"crypto/rand"
	"encoding/json"
	"net"
	"os"

	"github.com/spf13/cobra"
)

func randomBuffer(size int) []byte {
	buf := make([]byte, size)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	return buf
}

func RandomSubnet(network net.IPNet, prefix int) net.IPNet {
	ones, _ := network.Mask.Size()

	randomBuffer := randomBuffer(len(network.IP))

	newIp := make([]byte, len(network.IP))
	copy(newIp, network.IP)

	// Set the  random bits from the random buffer
	// on bits from size to prefix
	for bitIndex := ones; bitIndex < prefix; bitIndex++ {
		byteIndex := bitIndex / 8

		newIp[byteIndex] |= randomBuffer[byteIndex] & (1 << uint(7-bitIndex%8))
	}

	return net.IPNet{
		IP:   newIp,
		Mask: net.CIDRMask(prefix, 32),
	}
}

func NewInitCommand() *cobra.Command {
	var meshIpNet string
	var subnetPrefix int
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration for ikto",
		Run: func(cmd *cobra.Command, args []string) {
			config := DefaultConfig()
			_, ipNet, err := net.ParseCIDR(meshIpNet)
			if err != nil {
				panic(err)
			}

			config.MeshIPNet = ipNet.String()

			config.HostPrefixLength = subnetPrefix

			config.PrivateAddress = RandomSubnet(*ipNet, subnetPrefix).IP.String()
			bytes, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				panic(err)
			}
			os.Stdout.Write(append(bytes, '\n'))
		},
	}

	cmd.Flags().StringVarP(&meshIpNet, "mesh-ip-net", "m", "fd10::/16", "Mesh IP Net")
	cmd.Flags().IntVarP(&subnetPrefix, "subnet-prefix-length", "p", 48, "Subnet Prefix")

	return cmd
}
