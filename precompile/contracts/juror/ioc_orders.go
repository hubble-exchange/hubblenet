package juror

import (
	"errors"
	"math/big"

	"github.com/ava-labs/subnet-evm/plugin/evm/orderbook"
	b "github.com/ava-labs/subnet-evm/precompile/contracts/bibliophile"
)

func ValidatePlaceIOCorder(bibliophile b.BibliophileClient, inputStruct *ValidatePlaceIOCOrderInput) (response ValidatePlaceIOCOrderOutput) {
	order := inputStruct.Order
	trader := order.Trader

	var err error
	response.OrderHash, err = GetIOCOrderHash(&orderbook.IOCOrder{
		OrderType: order.OrderType,
		ExpireAt:  order.ExpireAt,
		BaseOrder: orderbook.BaseOrder{
			AmmIndex:          order.AmmIndex,
			Trader:            order.Trader,
			BaseAssetQuantity: order.BaseAssetQuantity,
			Price:             order.Price,
			Salt:              order.Salt,
			ReduceOnly:        order.ReduceOnly,
		},
	})
	if err != nil {
		response.Err = err.Error()
		return
	}

	if trader != inputStruct.Sender && !bibliophile.IsTradingAuthority(trader, inputStruct.Sender) {
		response.Err = ErrNoTradingAuthority.Error()
		return
	}
	blockTimestamp := bibliophile.GetAccessibleState().GetBlockContext().Timestamp()
	expireWithin := blockTimestamp + bibliophile.IOC_GetExpirationCap().Uint64()
	if order.BaseAssetQuantity.Sign() == 0 {
		response.Err = ErrInvalidFillAmount.Error()
		return
	}
	if OrderType(order.OrderType) != IOC {
		response.Err = errors.New("not_ioc_order").Error()
		return
	}
	if order.ExpireAt.Uint64() < blockTimestamp {
		response.Err = errors.New("ioc expired").Error()
		return
	}
	if order.ExpireAt.Uint64() > expireWithin {
		response.Err = errors.New("ioc expiration too far").Error()
		return
	}
	minSize := bibliophile.GetMinSizeRequirement(order.AmmIndex.Int64())
	if new(big.Int).Mod(order.BaseAssetQuantity, minSize).Sign() != 0 {
		response.Err = ErrNotMultiple.Error()
		return
	}

	if OrderStatus(bibliophile.IOC_GetOrderStatus(response.OrderHash)) != Invalid {
		response.Err = ErrInvalidOrder.Error()
		return
	}
	// this check is as such not required, because even if this order is not reducing the position, it will be rejected by the matching engine and expire away
	// this check is sort of also redundant because either ways user can circumvent this by placing several reduceOnly order
	// if order.ReduceOnly {}
	return response
}

func validateExecuteIOCOrder(bibliophile b.BibliophileClient, order *orderbook.IOCOrder, side Side, fillAmount *big.Int) (metadata *Metadata, err error) {
	if OrderType(order.OrderType) != IOC {
		return nil, errors.New("not ioc order")
	}
	if order.ExpireAt.Uint64() < bibliophile.GetAccessibleState().GetBlockContext().Timestamp() {
		return nil, errors.New("ioc expired")
	}
	orderHash, err := GetIOCOrderHash(order)
	if err != nil {
		return nil, err
	}
	if err := validateLimitOrderLike(bibliophile, &order.BaseOrder, bibliophile.IOC_GetOrderFilledAmount(orderHash), OrderStatus(bibliophile.IOC_GetOrderStatus(orderHash)), side, fillAmount); err != nil {
		return nil, err
	}
	return &Metadata{
		AmmIndex:          order.AmmIndex,
		Trader:            order.Trader,
		BaseAssetQuantity: order.BaseAssetQuantity,
		BlockPlaced:       bibliophile.IOC_GetBlockPlaced(orderHash),
		Price:             order.Price,
		OrderHash:         orderHash,
		OrderType:         IOC,
	}, nil
}
