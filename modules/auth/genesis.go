package auth

import (
	cTypes "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis - Init store state from genesis data
//
// CONTRACT: old coins from the FeeCollectionKeeper need to be transferred through
// a genesis port script to the new fee collector account
func InitGenesis(ctx cTypes.Context, ak AccountKeeper, data GenesisState) {
	ak.SetParams(ctx, data.Params)
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx cTypes.Context, ak AccountKeeper) GenesisState {
	params := ak.GetParams(ctx)
	return NewGenesisState(params)
}