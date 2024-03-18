package traderviewer

import (
	"errors"
	"math/big"

	ob "github.com/ava-labs/subnet-evm/plugin/evm/orderbook"
	hu "github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
	b "github.com/ava-labs/subnet-evm/precompile/contracts/bibliophile"
	juror "github.com/ava-labs/subnet-evm/precompile/contracts/juror"
	"github.com/ethereum/go-ethereum/common"
)

var	(
	ErrUnauthorizedCancellation = errors.New("unauthorized cancellation")
	ErrNotOverPositionCap = errors.New("not over position cap")
)

func ValidateCancelLimitOrderV2(bibliophile b.BibliophileClient, inputStruct *ValidateCancelLimitOrderV2Input) (response ValidateCancelLimitOrderV2Output) {
	order := inputStruct.Order
	sender := inputStruct.Sender
	assertLowMargin := inputStruct.AssertLowMargin
	assertOverPositionCap := inputStruct.AssertOverPositionCap

	response.Res.UnfilledAmount = big.NewInt(0)

	trader := order.Trader
	if (!assertLowMargin && trader != sender && !bibliophile.IsTradingAuthority(trader, sender)) ||
		(assertLowMargin && !bibliophile.IsValidator(sender)) ||
		(assertOverPositionCap && !bibliophile.IsValidator(sender)) {
		response.Err = ErrUnauthorizedCancellation.Error()
		return
	}
	orderHash, err := GetLimitOrderHashFromContractStruct(&order)
	response.OrderHash = orderHash
	if err != nil {
		response.Err = err.Error()
		return
	}
	switch status := juror.OrderStatus(bibliophile.GetOrderStatus(orderHash)); status {
	case juror.Invalid:
		response.Err = "Invalid"
		return
	case juror.Filled:
		response.Err = "Filled"
		return
	case juror.Cancelled:
		response.Err = "Cancelled"
		return
	default:
	}
	response.Res.UnfilledAmount = hu.Sub(order.BaseAssetQuantity, bibliophile.GetOrderFilledAmount(orderHash))
	response.Res.Amm = bibliophile.GetMarketAddressFromMarketID(order.AmmIndex.Int64())
	if assertLowMargin && bibliophile.GetAvailableMargin(trader, hu.UpgradeVersionV0orV1(bibliophile.GetTimeStamp())).Sign() != -1 {
		response.Err = "Not Low Margin"
		return
	} else if assertOverPositionCap {
		posSize := bibliophile.GetSize(response.Res.Amm, &trader)
		if hu.Abs(hu.Add(posSize, response.Res.UnfilledAmount)).Cmp(bibliophile.GetPositionCap(order.AmmIndex.Int64(), trader)) <= 0 {
			response.Err = ErrNotOverPositionCap.Error()
			return
		}
	}

	return response
}

func ILimitOrderBookOrderToLimitOrder(o *ILimitOrderBookOrder) *ob.LimitOrder {
	return &ob.LimitOrder{
		BaseOrder: hu.BaseOrder{
			AmmIndex:          o.AmmIndex,
			Trader:            o.Trader,
			BaseAssetQuantity: o.BaseAssetQuantity,
			Price:             o.Price,
			Salt:              o.Salt,
			ReduceOnly:        o.ReduceOnly,
		},
		PostOnly: o.PostOnly,
	}
}

func GetLimitOrderHashFromContractStruct(o *ILimitOrderBookOrder) (common.Hash, error) {
	return ILimitOrderBookOrderToLimitOrder(o).Hash()
}

func GetRequiredMargin(bibliophile b.BibliophileClient, inputStruct *GetRequiredMarginInput) *big.Int {
	return bibliophile.GetRequiredMargin(inputStruct.BaseAssetQuantity, inputStruct.Price, inputStruct.AmmIndex.Int64(), &inputStruct.Trader)
}
