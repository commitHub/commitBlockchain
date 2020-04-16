package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/persistenceOne/comdexCrust/modules/ibc/02-client/exported"
	"github.com/persistenceOne/comdexCrust/modules/ibc/02-client/types/tendermint"
)

var SubModuleCdc *codec.Codec

// RegisterCodec registers the IBC client interfaces and types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.Blockchain)(nil), nil)
	cdc.RegisterInterface((*exported.ConsensusState)(nil), nil)
	cdc.RegisterInterface((*exported.Evidence)(nil), nil)
	cdc.RegisterInterface((*exported.Header)(nil), nil)
	cdc.RegisterInterface((*exported.Misbehaviour)(nil), nil)

	cdc.RegisterConcrete(MsgCreateClient{}, "ibc/client/MsgCreateClient", nil)
	cdc.RegisterConcrete(MsgUpdateClient{}, "ibc/client/MsgUpdateClient", nil)
	cdc.RegisterConcrete(MsgSubmitMisbehaviour{}, "ibc/client/MsgSubmitMisbehaviour", nil)

	cdc.RegisterConcrete(tendermint.ConsensusState{}, "ibc/client/tendermint/ConsensusState", nil)
	cdc.RegisterConcrete(tendermint.Header{}, "ibc/client/tendermint/Header", nil)

	SetSubModuleCodec(cdc)
}

func SetSubModuleCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}
