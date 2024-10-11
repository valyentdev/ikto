package types

import (
	"encoding/json"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type PublicKey [32]byte

var _ json.Marshaler = (*PublicKey)(nil)
var _ json.Unmarshaler = (*PublicKey)(nil)

func (k *PublicKey) String() string {
	return k.WG().String()
}

func (k PublicKey) WG() wgtypes.Key {
	return wgtypes.Key(k)
}

func (n PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

func (n *PublicKey) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	key, err := ParseKey(s)
	if err != nil {
		return err
	}

	*n = key
	return nil
}

func ParseKey(s string) (PublicKey, error) {
	key, err := wgtypes.ParseKey(s)
	if err != nil {
		return PublicKey{}, err
	}

	return PublicKey(key), nil
}
