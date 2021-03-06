package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	cTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/commitHub/commitBlockchain/types"

	rest2 "github.com/commitHub/commitBlockchain/client/rest"
	fiatFactoryTypes "github.com/commitHub/commitBlockchain/modules/fiatFactory/internal/types"
)

type redeemFiatReq struct {
	BaseReq        rest.BaseReq `json:"base_req"`
	OwnerAddress   string       `json:"ownerAddress" `
	PegHash        string       `json:"pegHash" `
	RedeemedAmount int64        `json:"redeemedAmount" `
	Password       string       `json:"password"`
	Mode           string       `json:"mode"`
}

func RedeemFiatHandlerFunction(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req redeemFiatReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, name, err := context.GetFromFields(req.BaseReq.From, false)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx = cliCtx.WithFromAddress(fromAddr)
		cliCtx = cliCtx.WithFromName(name)

		ownerAddr, err := cTypes.AccAddressFromBech32(req.OwnerAddress)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		pegHashHex, err := types.GetFiatPegHashHex(req.PegHash)
		fiatPeg := types.BaseFiatPeg{
			PegHash:        pegHashHex,
			RedeemedAmount: req.RedeemedAmount,
		}

		msg := fiatFactoryTypes.BuildRedeemFiatMsg(fromAddr, ownerAddr, req.RedeemedAmount, types.FiatPegWallet{fiatPeg})
		rest2.SignAndBroadcast(req.BaseReq, cliCtx, req.Mode, req.Password, []cTypes.Msg{msg})
	}

}
