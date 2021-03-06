package types

import (
	"github.com/commitHub/commitBlockchain/types"
)

const (
	ModuleName   = "fiat"
	StoreKey     = ModuleName
	RouterKey    = StoreKey
	QuerierRoute = RouterKey
)

var (
	PegHashKey = []byte{0x01}
)

func FiatPegHashStoreKey(fiatPegHash types.PegHash) []byte {
	return append(PegHashKey, fiatPegHash.Bytes()...)
}
