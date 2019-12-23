package negotiation

import (
	"github.com/commitHub/commitBlockchain/modules/negotiation/internal/keeper"
	"github.com/commitHub/commitBlockchain/modules/negotiation/internal/types"
	types2 "github.com/commitHub/commitBlockchain/types"
)

const (
	StoreKey     = types.StoreKey
	ModuleName   = types.ModuleName
	RouterKey    = types.RouterKey
	QuerierRoute = types.QuerierRoute

	CodeInvalidSignature = types.CodeInvalidSignature
	DefaultCodeSpace     = types.DefaultCodeSpace
)

var (
	RegisterCodec = types.RegisterCodec
	ModuleCdc     = types.ModuleCdc

	DefaultGenesisState = types.DefaultGenesisState
	ValidateGenesis     = types.ValidateGenesis

	NewQuerier = keeper.NewQuerier
	NewKeeper  = keeper.NewKeeper

	ErrUnauthorized = types.ErrUnauthorized
)

type (
	GenesisState = types.GenesisState
	Keeper       = keeper.Keeper

	BaseNegotiation = types2.BaseNegotiation

	MsgChangeBuyerBids   = types.MsgChangeBuyerBids
	MsgChangeSellerBids  = types.MsgChangeSellerBids
	MsgConfirmBuyerBids  = types.MsgConfirmBuyerBids
	MsgConfirmSellerBids = types.MsgConfirmSellerBids

	ChangeBid  = types.ChangeBid
	ConfirmBid = types.ConfirmBid

	Signature = types2.Signature
)