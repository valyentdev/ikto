package network

import (
	"crypto/rand"
	"net"
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
