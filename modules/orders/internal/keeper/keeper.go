package keeper

import (
	cTypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/commitHub/commitBlockchain/codec"
	"github.com/commitHub/commitBlockchain/modules/auth"
	orderTypes "github.com/commitHub/commitBlockchain/modules/orders/internal/types"
	"github.com/commitHub/commitBlockchain/types"
)

type Keeper struct {
	storeKey      cTypes.StoreKey
	cdc           *codec.Codec
	AccountKeeper auth.AccountKeeper
}

func NewKeeper(storeKey cTypes.StoreKey, cdc *codec.Codec, accountKeeper auth.AccountKeeper) Keeper {

	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		AccountKeeper: accountKeeper,
	}
}

func (k Keeper) SetOrder(ctx cTypes.Context, order types.Order) {
	negotiationID := order.GetNegotiationID()
	store := ctx.KVStore(k.storeKey)

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(order)
	if err != nil {
		panic(err)
	}
	storeKey := orderTypes.GetOrderKey(negotiationID)
	store.Set(storeKey, bz)

}

func (k Keeper) GetOrder(ctx cTypes.Context, negotiationID types.NegotiationID) types.Order {
	store := ctx.KVStore(k.storeKey)
	storeKey := orderTypes.GetOrderKey(negotiationID)
	bz := store.Get(storeKey)
	if bz == nil {
		return nil
	}
	var order types.Order
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &order)
	return order
}

func (k Keeper) IterateOrders(ctx cTypes.Context, process func(types.Order) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := cTypes.KVStorePrefixIterator(store, orderTypes.OrdersKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var order types.Order
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &order)
		if process(order) {
			break
		}
	}
}

func (k Keeper) NewOrder(buyerAddress cTypes.AccAddress, sellerAddress cTypes.AccAddress, pegHash types.PegHash) types.Order {
	order := types.BaseOrder{}
	negotiationID := types.NegotiationID(append(append(buyerAddress.Bytes(), sellerAddress.Bytes()...), pegHash.Bytes()...))
	order.SetNegotiationID(negotiationID)
	return &order
}

func (keeper Keeper) SendAssetsToOrder(ctx cTypes.Context, fromAddress cTypes.AccAddress, toAddress cTypes.AccAddress, assetPeg types.AssetPeg) cTypes.Error {
	// negotiationID := negotiation.GetOrderKey(toAddress, fromAddress, assetPeg.GetPegHash())
	negotiationID := types.NegotiationID(append(append(toAddress.Bytes(), fromAddress.Bytes()...), assetPeg.GetPegHash().Bytes()...))
	order := keeper.GetOrder(ctx, negotiationID)
	if order == nil {
		order = keeper.NewOrder(toAddress, fromAddress, assetPeg.GetPegHash())
	}
	order.SetAssetPegWallet(types.AddAssetPegToWallet(assetPeg, order.GetAssetPegWallet()))
	keeper.SetOrder(ctx, order)
	return nil
}

// SendFiatsToOrder fiat pegs to order
func (keeper Keeper) SendFiatsToOrder(ctx cTypes.Context, fromAddress cTypes.AccAddress, toAddress cTypes.AccAddress, pegHash types.PegHash, fiatPegWallet types.FiatPegWallet) cTypes.Error {
	negotiationID := types.NegotiationID(append(append(fromAddress.Bytes(), toAddress.Bytes()...), pegHash.Bytes()...))
	order := keeper.GetOrder(ctx, negotiationID)
	if order == nil {
		order = keeper.NewOrder(fromAddress, toAddress, pegHash)
	}
	order.SetFiatPegWallet(types.AddFiatPegToWallet(order.GetFiatPegWallet(), fiatPegWallet))
	keeper.SetOrder(ctx, order)
	return nil
}

// GetOrderDetails : get the order details
func (keeper Keeper) GetOrderDetails(ctx cTypes.Context, buyerAddress cTypes.AccAddress, sellerAddress cTypes.AccAddress, pegHash types.PegHash) (cTypes.Error, types.AssetPegWallet, types.FiatPegWallet, string, string) {
	negotiationID := types.NegotiationID(append(append(buyerAddress.Bytes(), sellerAddress.Bytes()...), pegHash.Bytes()...))
	order := keeper.GetOrder(ctx, negotiationID)
	if order == nil {
		return cTypes.ErrInvalidAddress("Order not found!"), nil, nil, "", ""
	}
	return nil, order.GetAssetPegWallet(), order.GetFiatPegWallet(), order.GetFiatProofHash(), order.GetAWBProofHash()
}

// SetOrderFiatProofHash : Set FiatProofHash to Order
func (keeper Keeper) SetOrderFiatProofHash(ctx cTypes.Context, buyerAddress cTypes.AccAddress, sellerAddress cTypes.AccAddress, pegHash types.PegHash, fiatProofHash string) {
	negotiationID := types.NegotiationID(append(append(buyerAddress.Bytes(), sellerAddress.Bytes()...), pegHash.Bytes()...))
	order := keeper.GetOrder(ctx, negotiationID)
	order.SetFiatProofHash(fiatProofHash)
	keeper.SetOrder(ctx, order)
}

// SetOrderAWBProofHash : Set AWBProofHash to Order
func (keeper Keeper) SetOrderAWBProofHash(ctx cTypes.Context, buyerAddress cTypes.AccAddress, sellerAddress cTypes.AccAddress, pegHash types.PegHash, awbProofHash string) {
	negotiationID := types.NegotiationID(append(append(buyerAddress.Bytes(), sellerAddress.Bytes()...), pegHash.Bytes()...))
	order := keeper.GetOrder(ctx, negotiationID)
	order.SetAWBProofHash(awbProofHash)
	keeper.SetOrder(ctx, order)
}

// SendAssetFromOrder asset peg to buyer
func (keeper Keeper) SendAssetFromOrder(ctx cTypes.Context, fromAddress cTypes.AccAddress, toAddress cTypes.AccAddress, assetPeg types.AssetPeg) types.AssetPegWallet {
	negotiationID := types.NegotiationID(append(append(fromAddress.Bytes(), toAddress.Bytes()...), assetPeg.GetPegHash().Bytes()...))
	order := keeper.GetOrder(ctx, negotiationID)
	_, updatedAssetPegWallet := types.SubtractAssetPegFromWallet(assetPeg.GetPegHash(), order.GetAssetPegWallet())
	order.SetAssetPegWallet(updatedAssetPegWallet)
	keeper.SetOrder(ctx, order)
	return updatedAssetPegWallet
}

// SendFiatsFromOrder fiat pegs to seller
func (keeper Keeper) SendFiatsFromOrder(ctx cTypes.Context, fromAddress cTypes.AccAddress, toAddress cTypes.AccAddress, pegHash types.PegHash, fiatPegWallet types.FiatPegWallet) types.FiatPegWallet {
	negotiationID := types.NegotiationID(append(append(fromAddress.Bytes(), toAddress.Bytes()...), pegHash.Bytes()...))
	order := keeper.GetOrder(ctx, negotiationID)
	updatedFiatPegWallet := types.SubtractFiatPegWalletFromWallet(fiatPegWallet, order.GetFiatPegWallet())
	order.SetFiatPegWallet(updatedFiatPegWallet)
	keeper.SetOrder(ctx, order)
	return updatedFiatPegWallet
}