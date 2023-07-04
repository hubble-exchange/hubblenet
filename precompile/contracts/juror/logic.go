package juror

import (
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type LimitOrder struct {
	AmmIndex          *big.Int
	Trader            common.Address
	BaseAssetQuantity *big.Int
	Price             *big.Int
	Salt              *big.Int
	ReduceOnly        bool
}

type DecodeStep struct {
	OrderType    uint8
	EncodedOrder []byte
}

type Metadata struct {
	AmmIndex          *big.Int
	Trader            common.Address
	BaseAssetQuantity *big.Int
	Price             *big.Int
	BlockPlaced       *big.Int
	OrderHash         common.Hash
}

type Side uint8

const (
	Long Side = iota
	Short
	Liquidation
)

type OrderStatus uint8

const (
	Invalid OrderStatus = iota
	Placed
	Filled
	Cancelled
)

var (
	ErrInvalidFillAmount        = errors.New("invalid fillAmount")
	ErrNotLongOrder             = errors.New("OB_order_0_is_not_long")
	ErrNotShortOrder            = errors.New("OB_order_1_is_not_short")
	ErrNotSameAMM               = errors.New("OB_orders_for_different_amms")
	ErrNoMatch                  = errors.New("OB_orders_do_not_match")
	ErrInvalidOrder             = errors.New("OB_invalid_order")
	ErrNotMultiple              = errors.New("OB.not_multiple")
	ErrTooLow                   = errors.New("OB_long_order_price_too_low")
	ErrTooHigh                  = errors.New("OB_short_order_price_too_high")
	ErrOverFill                 = errors.New("OB_filled_amount_higher_than_order_base")
	ErrReduceOnlyAmountExceeded = errors.New("OB_reduce_only_amount_exceeded")
)

// Business Logic
func ValidateOrdersAndDetermineFillPrice(bibliophile Bibliophile, inputStruct *ValidateOrdersAndDetermineFillPriceInput) (*ValidateOrdersAndDetermineFillPriceOutput, error) {
	if inputStruct.FillAmount.Sign() <= 0 {
		return nil, ErrInvalidFillAmount
	}

	decodeStep0, err := decodeTypeAndEncodedOrder(inputStruct.Data[0])
	if err != nil {
		return nil, err
	}
	m0, err := validateOrder(bibliophile, decodeStep0.OrderType, decodeStep0.EncodedOrder, Long, inputStruct.FillAmount)
	if err != nil {
		return nil, err
	}

	decodeStep1, err := decodeTypeAndEncodedOrder(inputStruct.Data[1])
	if err != nil {
		return nil, err
	}
	m1, err := validateOrder(bibliophile, decodeStep1.OrderType, decodeStep1.EncodedOrder, Short, new(big.Int).Neg(inputStruct.FillAmount))
	if err != nil {
		return nil, err
	}

	if m0.AmmIndex.Cmp(m1.AmmIndex) != 0 {
		return nil, ErrNotSameAMM
	}

	if m0.Price.Cmp(m1.Price) == -1 {
		return nil, ErrNoMatch
	}

	_fillPriceAndModes, err := bibliophile.DetermineFillPrice(m0.AmmIndex.Int64(), inputStruct.FillAmount, m0.Price, m1.Price, m0.BlockPlaced, m1.BlockPlaced)
	if err != nil {
		return nil, err
	}

	output := &ValidateOrdersAndDetermineFillPriceOutput{
		Instructions: [2]IClearingHouseInstruction{
			IClearingHouseInstruction{
				AmmIndex:  m0.AmmIndex,
				Trader:    m0.Trader,
				OrderHash: m0.OrderHash,
				Mode:      _fillPriceAndModes.Mode0,
			},
			IClearingHouseInstruction{
				AmmIndex:  m1.AmmIndex,
				Trader:    m1.Trader,
				OrderHash: m1.OrderHash,
				Mode:      _fillPriceAndModes.Mode1,
			},
		},
		OrderTypes: [2]uint8{decodeStep0.OrderType, decodeStep1.OrderType},
		EncodedOrders: [2][]byte{
			decodeStep0.EncodedOrder,
			decodeStep1.EncodedOrder,
		},
		FillPrice: _fillPriceAndModes.FillPrice,
	}
	return output, nil
}

func decodeTypeAndEncodedOrder(data []byte) (*DecodeStep, error) {
	abiJSON := `[{"name": "orderType", "type": "uint8"}, {"name": "encodedOrder", "type": "bytes"}]` // check
	myABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}

	// Create an instance of the Order struct
	decodeStep := &DecodeStep{}

	// Decode the encodedOrder
	err = myABI.UnpackIntoInterface(decodeStep, "DecodeStep", data)
	if err != nil {
		return nil, err
	}
	return decodeStep, nil
}

func validateOrder(bibliophile Bibliophile, orderType uint8, encodedOrder []byte, side Side, fillAmount *big.Int) (metadata *Metadata, err error) {
	if orderType == 0 {
		order, err := decodeLimitOrder(encodedOrder)
		if err != nil {
			return nil, err
		}
		return validateExecuteLimitOrder(bibliophile, order, side, fillAmount)
	}
	return nil, errors.New("invalid order type")
}

func decodeLimitOrder(encodedOrder []byte) (*LimitOrder, error) {
	return &LimitOrder{
		AmmIndex:          new(big.Int).SetBytes(encodedOrder[0:32]),
		Trader:            common.BytesToAddress(encodedOrder[44:64]),
		BaseAssetQuantity: fromTwosComplement(encodedOrder[64:96]),
		Price:             new(big.Int).SetBytes(encodedOrder[96:128]),
		Salt:              new(big.Int).SetBytes(encodedOrder[128:160]),
		ReduceOnly:        new(big.Int).SetBytes(encodedOrder[160:192]).Sign() != 0,
	}, nil
}

func validateExecuteLimitOrder(bibliophile Bibliophile, order *LimitOrder, side Side, fillAmount *big.Int) (metadata *Metadata, err error) {
	orderHash, err := GetLimitOrderHash(order)
	if err != nil {
		return nil, err
	}
	if err := validateLimitOrderLike(bibliophile, order, bibliophile.GetOrderFilledAmount(orderHash), OrderStatus(bibliophile.GetOrderStatus(orderHash)), side, fillAmount); err != nil {
		return nil, err
	}
	return &Metadata{
		AmmIndex:          order.AmmIndex,
		Trader:            order.Trader,
		BaseAssetQuantity: order.BaseAssetQuantity,
		BlockPlaced:       bibliophile.GetBlockPlaced(orderHash),
		Price:             order.Price,
		OrderHash:         orderHash,
	}, nil
}

func validateLimitOrderLike(bibliophile Bibliophile, order *LimitOrder, filledAmount *big.Int, status OrderStatus, side Side, fillAmount *big.Int) error {
	if status != Placed {
		return ErrInvalidOrder
	}

	// in case of liquidations, side of the order is determined by the sign of the base asset quantity, so basically base asset quantity check is redundant
	if side == Liquidation {
		if order.BaseAssetQuantity.Sign() > 0 {
			side = Long
		} else if order.BaseAssetQuantity.Sign() < 0 {
			side = Short
		}
	}

	market := bibliophile.GetMarketAddressFromMarketID(order.AmmIndex.Int64())
	if side == Long {
		if order.BaseAssetQuantity.Sign() <= 0 {
			return ErrNotLongOrder
		}
		if fillAmount.Sign() <= 0 {
			return ErrInvalidFillAmount
		}
		if new(big.Int).Add(filledAmount, fillAmount).Cmp(order.BaseAssetQuantity) > 0 {
			return ErrOverFill
		}
		if order.ReduceOnly {
			posSize := bibliophile.GetSize(market, &order.Trader)
			// posSize should be closed to continue to be Short
			// this also returns err if posSize >= 0, which should not happen because we are executing a long reduceOnly order on this account
			if new(big.Int).Add(posSize, fillAmount).Sign() > 0 {
				return ErrReduceOnlyAmountExceeded
			}
		}
	} else if side == Short {
		if order.BaseAssetQuantity.Sign() >= 0 {
			return ErrNotShortOrder
		}
		if fillAmount.Sign() >= 0 {
			return ErrInvalidFillAmount
		}
		if new(big.Int).Add(filledAmount, fillAmount).Cmp(order.BaseAssetQuantity) < 0 { // all quantities are -ve
			return ErrOverFill
		}
		if order.ReduceOnly {
			posSize := bibliophile.GetSize(market, &order.Trader)
			// posSize should continue to be Long
			// this also returns is posSize <= 0, which should not happen because we are executing a short reduceOnly order on this account
			if new(big.Int).Add(posSize, fillAmount).Sign() < 0 {
				return ErrReduceOnlyAmountExceeded
			}
		}
	} else {
		return errors.New("OB_invalid_side")
	}
	return nil
}
