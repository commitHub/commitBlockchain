package keeper

import (
	"github.com/commitHub/commitBlockchain/modules/assetFactory"
	"github.com/commitHub/commitBlockchain/modules/bank"
	channelexported "github.com/commitHub/commitBlockchain/modules/ibc/04-channel/exported"
	channeltypes "github.com/commitHub/commitBlockchain/modules/ibc/04-channel/types"
	"github.com/commitHub/commitBlockchain/modules/ibc/20-transfer/types"
	commitment "github.com/commitHub/commitBlockchain/modules/ibc/23-commitment"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SendTransfer handles transfer sending logic
func (k Keeper) SendTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	amount sdk.Coins,
	sender,
	receiver sdk.AccAddress,
	isSourceChain bool,
) error {
	// get the port and channel of the counterparty
	channel, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return channeltypes.ErrChannelNotFound(k.codespace, sourcePort, sourceChannel)
	}

	destinationPort := channel.Counterparty.PortID
	destinationChannel := channel.Counterparty.ChannelID

	// get the next sequence
	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return channeltypes.ErrSequenceNotFound(k.codespace, "send")
	}

	coins := make(sdk.Coins, len(amount))
	// prefix := types.GetDenomPrefix(destinationPort, destinationChannel)
	switch {
	case isSourceChain:
		// build the receiving denomination prefix
		for i, coin := range amount {
			coins[i] = sdk.NewCoin(coin.Denom, coin.Amount)
		}
	default:
		coins = amount
	}

	return k.createOutgoingPacket(ctx, sequence, sourcePort, sourceChannel, destinationPort, destinationChannel, coins, sender, receiver, isSourceChain)
}

// ReceivePacket handles receiving packet
func (k Keeper) ReceivePacket(ctx sdk.Context, packet channelexported.PacketI, proof commitment.ProofI, height uint64) error {
	_, err := k.channelKeeper.RecvPacket(ctx, packet, proof, height, nil, k.storeKey)
	if err != nil {
		return err
	}

	var data types.PacketData
	err = data.UnmarshalJSON(packet.GetData())
	if err != nil {
		return sdk.NewError(types.DefaultCodespace, types.CodeInvalidPacketData, "invalid packet data")
	}

	return k.ReceiveTransfer(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetDestPort(), packet.GetDestChannel(), data)
}

// ReceiveTransfer handles transfer receiving logic
func (k Keeper) ReceiveTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	data types.PacketData,
) error {
	if data.Source {
		// prefix := types.GetDenomPrefix(destinationPort, destinationChannel)
		// for _, coin := range data.Amount {
		// 	if !strings.HasPrefix(coin.Denom, prefix) {
		// 		return sdk.ErrInvalidCoins(fmt.Sprintf("%s doesn't contain the prefix '%s'", coin.Denom, prefix))
		// 	}
		// }

		// mint new tokens if the source of the transfer is the same chain
		err := k.supplyKeeper.MintCoins(ctx, types.GetModuleAccountName(), data.Amount)
		if err != nil {
			return err
		}

		// send to receiver
		return k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.GetModuleAccountName(), data.Receiver, data.Amount)
	}

	// unescrow tokens

	// check the denom prefix
	// prefix := types.GetDenomPrefix(sourcePort, sourceChannel)
	// coins := make(sdk.Coins, len(data.Amount))
	// for i, coin := range data.Amount {
	// 	if !strings.HasPrefix(coin.Denom, prefix) {
	// 		return sdk.ErrInvalidCoins(fmt.Sprintf("%s doesn't contain the prefix '%s'", coin.Denom, prefix))
	// 	}
	// 	coins[i] = sdk.NewCoin(coin.Denom[len(prefix):], coin.Amount)
	// }

	escrowAddress := types.GetEscrowAddress(destinationPort, destinationChannel)
	// return k.bankKeeper.SendCoins(ctx, escrowAddress, data.Receiver, coins)
	return k.bankKeeper.SendCoins(ctx, escrowAddress, data.Receiver, data.Amount)

}

func (k Keeper) createOutgoingPacket(
	ctx sdk.Context,
	seq uint64,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	amount sdk.Coins,
	sender sdk.AccAddress,
	receiver sdk.AccAddress,
	isSourceChain bool,
) error {
	if isSourceChain {
		// escrow tokens if the destination chain is the same as the sender's
		escrowAddress := types.GetEscrowAddress(sourcePort, sourceChannel)

		// prefix := types.GetDenomPrefix(destinationPort, destinationChannel)
		// coins := make(sdk.Coins, len(amount))
		// for i, coin := range amount {
		// 	if !strings.HasPrefix(coin.Denom, prefix) {
		// 		return sdk.ErrInvalidCoins(fmt.Sprintf("%s doesn't contain the prefix '%s'", coin.Denom, prefix))
		// 	}
		// 	coins[i] = sdk.NewCoin(coin.Denom[len(prefix):], coin.Amount)
		// }

		// err := k.bankKeeper.SendCoins(ctx, sender, escrowAddress, coins)
		err := k.bankKeeper.SendCoins(ctx, sender, escrowAddress, amount)
		if err != nil {
			return err
		}

	} else {
		// burn vouchers from the sender's balance if the source is from another chain
		// prefix := types.GetDenomPrefix(sourcePort, sourceChannel)
		// for _, coin := range amount {
		// 	if !strings.HasPrefix(coin.Denom, prefix) {
		// 		return sdk.ErrInvalidCoins(fmt.Sprintf("%s doesn't contain the prefix '%s'", coin.Denom, prefix))
		// 	}
		// }

		// transfer the coins to the module account and burn them
		err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, sender, types.GetModuleAccountName(), amount)
		if err != nil {
			return err
		}

		// burn from supply
		err = k.supplyKeeper.BurnCoins(ctx, types.GetModuleAccountName(), amount)
		if err != nil {
			return err
		}
	}

	packetData := types.PacketData{
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Source:   isSourceChain,
	}

	// TODO: This should be binary-marshaled and hashed (for the commitment in the store).
	packetDataBz, err := packetData.MarshalJSON()
	if err != nil {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidPacketData, "invalid packet data")
	}

	packet := channeltypes.NewPacket(
		seq,
		uint64(ctx.BlockHeight())+DefaultPacketTimeout,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		packetDataBz,
	)

	// TODO: Remove this, capability keys are never generated when sending packets. Not sure why this is here.
	key := sdk.NewKVStoreKey(types.BoundPortID)

	return k.channelKeeper.SendPacket(ctx, packet, key)
}

func (k Keeper) IssueAssetTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	issueAsset bank.IssueAsset,
) error {
	// get the port and channel of the counterparty
	channel, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return channeltypes.ErrChannelNotFound(k.codespace, sourcePort, sourceChannel)
	}

	destinationPort := channel.Counterparty.PortID
	destinationChannel := channel.Counterparty.ChannelID

	// get the next sequence
	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return channeltypes.ErrSequenceNotFound(k.codespace, "issueAsset")
	}

	if !issueAsset.AssetPeg.GetModerated() {
		return sdk.ErrInternal("Cannot do unmoderated issue asset.")
	}

	return k.createIssueAssetOutgoingPacket(ctx, sequence, sourcePort, sourceChannel, destinationPort, destinationChannel, issueAsset)
}

func (k Keeper) createIssueAssetOutgoingPacket(
	ctx sdk.Context,
	seq uint64,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	issueAsset bank.IssueAsset,
) error {

	issuedAssetPeg, err := k.bankKeeper.IssueAssetsToWallets(ctx, issueAsset)
	if err != nil {
		return err
	}
	issueAsset.AssetPeg.SetPegHash(issuedAssetPeg.GetPegHash())
	issueAssetPacketData := types.IssueAssetPacketData{
		IssueAsset: types.NewIssueAsset(issueAsset),
	}

	// TODO: This should be binary-marshaled and hashed (for the commitment in the store).
	issueAssetPacketDataBz, err2 := issueAssetPacketData.MarshalJSON()
	if err2 != nil {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidPacketData, "invalid issue asset packet data")
	}

	packet := channeltypes.NewPacket(
		seq,
		uint64(ctx.BlockHeight())+DefaultPacketTimeout,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		issueAssetPacketDataBz,
	)

	// TODO: Remove this, capability keys are never generated when sending packets. Not sure why this is here.
	key := sdk.NewKVStoreKey(types.BoundPortID)

	return k.channelKeeper.SendPacket(ctx, packet, key)
}

func (k Keeper) ReceiveIssueAssetPacket(ctx sdk.Context, packet channelexported.PacketI, proof commitment.ProofI, height uint64) error {
	_, err := k.channelKeeper.RecvPacket(ctx, packet, proof, height, nil, k.storeKey)
	if err != nil {
		return err
	}

	var data types.IssueAssetPacketData
	err = data.UnmarshalJSON(packet.GetData())
	if err != nil {
		return sdk.NewError(types.DefaultCodespace, types.CodeInvalidPacketData, "invalid packet data!!")
	}

	return k.ReceiveIssueAssetTransfer(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetDestPort(), packet.GetDestChannel(), data)
}

func (k Keeper) ReceiveIssueAssetTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	data types.IssueAssetPacketData,
) error {
	issueAsset := assetFactory.NewIssueAsset(data.IssueAsset.IssuerAddress, data.IssueAsset.ToAddress, &data.IssueAsset.AssetPeg)
	return k.assetKeeper.InstantiateAndAssignAsset(ctx, issueAsset)
}
