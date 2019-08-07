// nolint
// autogenerated code using github.com/rigelrozanski/multitool
// aliases generated for the following subdirectories:
// ALIASGEN: github.com/commitHub/commitBlockchain/modules/crisis/types
package crisis

import (
	"github.com/commitHub/commitBlockchain/modules/crisis/internal/keeper"
	"github.com/commitHub/commitBlockchain/modules/crisis/internal/types"
)

const (
	DefaultCodespace  = types.DefaultCodespace
	CodeInvalidInput  = types.CodeInvalidInput
	ModuleName        = types.ModuleName
	DefaultParamspace = types.DefaultParamspace
)

var (
	// functions aliases
	RegisterCodec         = types.RegisterCodec
	ErrNilSender          = types.ErrNilSender
	ErrUnknownInvariant   = types.ErrUnknownInvariant
	NewGenesisState       = types.NewGenesisState
	DefaultGenesisState   = types.DefaultGenesisState
	NewMsgVerifyInvariant = types.NewMsgVerifyInvariant
	ParamKeyTable         = types.ParamKeyTable
	NewInvarRoute         = types.NewInvarRoute
	NewKeeper             = keeper.NewKeeper
	
	// variable aliases
	ModuleCdc                = types.ModuleCdc
	ParamStoreKeyConstantFee = types.ParamStoreKeyConstantFee
)

type (
	GenesisState = types.GenesisState
	MsgVerifyInvariant = types.MsgVerifyInvariant
	InvarRoute = types.InvarRoute
	Keeper = keeper.Keeper
)
