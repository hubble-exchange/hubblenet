package jurorv2

import (
	"errors"
	"math/big"
	"strings"

	ob "github.com/ava-labs/subnet-evm/plugin/evm/orderbook"
	hu "github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
	b "github.com/ava-labs/subnet-evm/precompile/contracts/bibliophile"
	"github.com/ava-labs/subnet-evm/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type Metadata struct {
	AmmIndex          *big.Int
	Trader            common.Address
	BaseAssetQuantity *big.Int
	Price             *big.Int
	BlockPlaced       *big.Int
	OrderHash         common.Hash
	OrderType         ob.OrderType
	PostOnly          bool
}

type Side uint8

const (
	Long Side = iota
	Short
	Liquidation
)

type OrderStatus uint8

// has to be exact same as IOrderHandler
const (
	Invalid OrderStatus = iota
	Placed
	Filled
	Cancelled
)

var (
	ErrTwoOrders         = errors.New("need 2 orders")
	ErrInvalidFillAmount = errors.New("invalid fillAmount")
	ErrNotLongOrder      = errors.New("not long")
	ErrNotShortOrder     = errors.New("not short")
	ErrNotSameAMM        = errors.New("OB_orders_for_different_amms")
	ErrNoMatch           = errors.New("OB_orders_do_not_match")
	ErrBothPostOnly      = errors.New("both orders are post only")
	ErrNotMultiple       = errors.New("not multiple")

	ErrInvalidOrder                       = errors.New("invalid order")
	ErrNotIOCOrder                        = errors.New("not_ioc_order")
	ErrInvalidPrice                       = errors.New("invalid price")
	ErrPricePrecision                     = errors.New("invalid price precision")
	ErrInvalidMarket                      = errors.New("invalid market")
	ErrCancelledOrder                     = errors.New("cancelled order")
	ErrFilledOrder                        = errors.New("filled order")
	ErrOrderAlreadyExists                 = errors.New("order already exists")
	ErrTooLow                             = errors.New("long price below lower bound")
	ErrTooHigh                            = errors.New("short price above upper bound")
	ErrOverFill                           = errors.New("overfill")
	ErrReduceOnlyAmountExceeded           = errors.New("not reducing pos")
	ErrBaseAssetQuantityZero              = errors.New("baseAssetQuantity is zero")
	ErrReduceOnlyBaseAssetQuantityInvalid = errors.New("reduce only order must reduce position")
	ErrNetReduceOnlyAmountExceeded        = errors.New("net reduce only amount exceeded")
	ErrStaleReduceOnlyOrders              = errors.New("cancel stale reduce only orders")
	ErrInsufficientMargin                 = errors.New("insufficient margin")
	ErrCrossingMarket                     = errors.New("crossing market")
	ErrIOCOrderExpired                    = errors.New("IOC order expired")
	ErrOpenOrders                         = errors.New("open orders")
	ErrOpenReduceOnlyOrders               = errors.New("open reduce only orders")
	ErrNoTradingAuthority                 = errors.New("no trading authority")
	ErrNoReferrer                         = errors.New("no referrer")
)

type BadElement uint8

// DO NOT change this ordering because it is critical for the orderbook to determine the problematic order
const (
	Order0 BadElement = iota
	Order1
	Generic
	NoError
)

// Business Logic
func ValidateOrdersAndDetermineFillPrice(bibliophile b.BibliophileClient, inputStruct *ValidateOrdersAndDetermineFillPriceInput) ValidateOrdersAndDetermineFillPriceOutput {
	if len(inputStruct.Data) != 2 {
		return getValidateOrdersAndDetermineFillPriceErrorOutput(ErrTwoOrders, Generic, common.Hash{})
	}

	if inputStruct.FillAmount.Sign() <= 0 {
		return getValidateOrdersAndDetermineFillPriceErrorOutput(ErrInvalidFillAmount, Generic, common.Hash{})
	}

	decodeStep0, err := hu.DecodeTypeAndEncodedOrder(inputStruct.Data[0])
	if err != nil {
		return getValidateOrdersAndDetermineFillPriceErrorOutput(err, Order0, common.Hash{})
	}
	m0, err := validateOrder(bibliophile, decodeStep0.OrderType, decodeStep0.EncodedOrder, Long, inputStruct.FillAmount)
	if err != nil {
		return getValidateOrdersAndDetermineFillPriceErrorOutput(err, Order0, m0.OrderHash)
	}

	decodeStep1, err := hu.DecodeTypeAndEncodedOrder(inputStruct.Data[1])
	if err != nil {
		return getValidateOrdersAndDetermineFillPriceErrorOutput(err, Order1, common.Hash{})
	}
	m1, err := validateOrder(bibliophile, decodeStep1.OrderType, decodeStep1.EncodedOrder, Short, new(big.Int).Neg(inputStruct.FillAmount))
	if err != nil {
		return getValidateOrdersAndDetermineFillPriceErrorOutput(err, Order1, m1.OrderHash)
	}

	if m0.AmmIndex.Cmp(m1.AmmIndex) != 0 {
		return getValidateOrdersAndDetermineFillPriceErrorOutput(ErrNotSameAMM, Generic, common.Hash{})
	}

	if m0.Price.Cmp(m1.Price) < 0 {
		return getValidateOrdersAndDetermineFillPriceErrorOutput(ErrNoMatch, Generic, common.Hash{})
	}

	// check 11
	if m0.PostOnly && m1.PostOnly {
		return getValidateOrdersAndDetermineFillPriceErrorOutput(ErrBothPostOnly, Generic, common.Hash{})
	}

	minSize := bibliophile.GetMinSizeRequirement(m0.AmmIndex.Int64())
	if new(big.Int).Mod(inputStruct.FillAmount, minSize).Cmp(big.NewInt(0)) != 0 {
		return getValidateOrdersAndDetermineFillPriceErrorOutput(ErrNotMultiple, Generic, common.Hash{})
	}

	fillPriceAndModes, err, element := determineFillPrice(bibliophile, m0, m1)
	if err != nil {
		orderHash := common.Hash{}
		if element == Order0 {
			orderHash = m0.OrderHash
		} else if element == Order1 {
			orderHash = m1.OrderHash
		}
		return getValidateOrdersAndDetermineFillPriceErrorOutput(err, element, orderHash)
	}

	return ValidateOrdersAndDetermineFillPriceOutput{
		Err:     "",
		Element: uint8(NoError),
		Res: IOrderHandlerMatchingValidationRes{
			Instructions: [2]IClearingHouseInstruction{
				IClearingHouseInstruction{
					AmmIndex:  m0.AmmIndex,
					Trader:    m0.Trader,
					OrderHash: m0.OrderHash,
					Mode:      uint8(fillPriceAndModes.Mode0),
				},
				IClearingHouseInstruction{
					AmmIndex:  m1.AmmIndex,
					Trader:    m1.Trader,
					OrderHash: m1.OrderHash,
					Mode:      uint8(fillPriceAndModes.Mode1),
				},
			},
			OrderTypes: [2]uint8{uint8(decodeStep0.OrderType), uint8(decodeStep1.OrderType)},
			EncodedOrders: [2][]byte{
				decodeStep0.EncodedOrder,
				decodeStep1.EncodedOrder,
			},
			FillPrice: fillPriceAndModes.FillPrice,
		},
	}
}

type executionMode uint8

// DO NOT change this ordering because it is critical for the clearing house to determine the correct fill mode
const (
	Taker executionMode = iota
	Maker
)

type FillPriceAndModes struct {
	FillPrice *big.Int
	Mode0     executionMode
	Mode1     executionMode
}

func determineFillPrice(bibliophile b.BibliophileClient, m0, m1 *Metadata) (*FillPriceAndModes, error, BadElement) {
	output := FillPriceAndModes{}
	upperBound, lowerBound := bibliophile.GetUpperAndLowerBoundForMarket(m0.AmmIndex.Int64())
	if m0.Price.Cmp(lowerBound) == -1 {
		return nil, ErrTooLow, Order0
	}
	if m1.Price.Cmp(upperBound) == 1 {
		return nil, ErrTooHigh, Order1
	}

	blockDiff := m0.BlockPlaced.Cmp(m1.BlockPlaced)
	if blockDiff == -1 {
		// order0 came first, can't be IOC order
		if m0.OrderType == ob.IOC {
			return nil, ErrIOCOrderExpired, Order0
		}
		// order1 came second, can't be post only order
		if m1.OrderType == ob.Limit && m1.PostOnly {
			return nil, ErrCrossingMarket, Order1
		}
		output.Mode0 = Maker
		output.Mode1 = Taker
	} else if blockDiff == 1 {
		// order1 came first, can't be IOC order
		if m1.OrderType == ob.IOC {
			return nil, ErrIOCOrderExpired, Order1
		}
		// order0 came second, can't be post only order
		if m0.OrderType == ob.Limit && m0.PostOnly {
			return nil, ErrCrossingMarket, Order0
		}
		output.Mode0 = Taker
		output.Mode1 = Maker
	} else {
		// both orders were placed in same block
		if m1.OrderType == ob.IOC {
			// order1 is IOC, order0 is Limit or post only
			output.Mode0 = Maker
			output.Mode1 = Taker
		} else {
			// scenarios:
			// 1. order0 is IOC, order1 is Limit or post only
			// 2. both order0 and order1 are Limit or post only (in that scenario we default to long being the taker order, which can sometimes result in a better execution price for them)
			output.Mode0 = Taker
			output.Mode1 = Maker
		}
	}

	if output.Mode0 == Maker {
		output.FillPrice = utils.BigIntMin(m0.Price, upperBound)
	} else {
		output.FillPrice = utils.BigIntMax(m1.Price, lowerBound)
	}
	return &output, nil, NoError
}

func ValidateLiquidationOrderAndDetermineFillPrice(bibliophile b.BibliophileClient, inputStruct *ValidateLiquidationOrderAndDetermineFillPriceInput) ValidateLiquidationOrderAndDetermineFillPriceOutput {
	fillAmount := new(big.Int).Set(inputStruct.LiquidationAmount)
	if fillAmount.Sign() <= 0 {
		return getValidateLiquidationOrderAndDetermineFillPriceErrorOutput(ErrInvalidFillAmount, Generic, common.Hash{})
	}

	decodeStep0, err := hu.DecodeTypeAndEncodedOrder(inputStruct.Data)
	if err != nil {
		return getValidateLiquidationOrderAndDetermineFillPriceErrorOutput(err, Order0, common.Hash{})
	}
	m0, err := validateOrder(bibliophile, decodeStep0.OrderType, decodeStep0.EncodedOrder, Liquidation, fillAmount)
	if err != nil {
		return getValidateLiquidationOrderAndDetermineFillPriceErrorOutput(err, Order0, m0.OrderHash)
	}

	if m0.BaseAssetQuantity.Sign() < 0 {
		fillAmount = new(big.Int).Neg(fillAmount)
	}

	minSize := bibliophile.GetMinSizeRequirement(m0.AmmIndex.Int64())
	if new(big.Int).Mod(fillAmount, minSize).Cmp(big.NewInt(0)) != 0 {
		return getValidateLiquidationOrderAndDetermineFillPriceErrorOutput(ErrNotMultiple, Generic, common.Hash{})
	}

	fillPrice, err := determineLiquidationFillPrice(bibliophile, m0)
	if err != nil {
		return getValidateLiquidationOrderAndDetermineFillPriceErrorOutput(err, Order0, m0.OrderHash)
	}

	return ValidateLiquidationOrderAndDetermineFillPriceOutput{
		Err:     "",
		Element: uint8(NoError),
		Res: IOrderHandlerLiquidationMatchingValidationRes{
			Instruction: IClearingHouseInstruction{
				AmmIndex:  m0.AmmIndex,
				Trader:    m0.Trader,
				OrderHash: m0.OrderHash,
				Mode:      uint8(Maker),
			},
			OrderType:    uint8(decodeStep0.OrderType),
			EncodedOrder: decodeStep0.EncodedOrder,
			FillPrice:    fillPrice,
			FillAmount:   fillAmount,
		},
	}
}

func determineLiquidationFillPrice(bibliophile b.BibliophileClient, m0 *Metadata) (*big.Int, error) {
	liqUpperBound, liqLowerBound := bibliophile.GetAcceptableBoundsForLiquidation(m0.AmmIndex.Int64())
	upperBound, lowerBound := bibliophile.GetUpperAndLowerBoundForMarket(m0.AmmIndex.Int64())
	if m0.BaseAssetQuantity.Sign() > 0 {
		// we are liquidating a long position
		// do not allow liquidation if order.Price < liqLowerBound, because that gives scope for malicious activity to a validator
		if m0.Price.Cmp(liqLowerBound) == -1 {
			return nil, ErrTooLow
		}
		return utils.BigIntMin(m0.Price, upperBound /* oracle spread upper bound */), nil
	}

	// we are liquidating a short position
	if m0.Price.Cmp(liqUpperBound) == 1 {
		return nil, ErrTooHigh
	}
	return utils.BigIntMax(m0.Price, lowerBound /* oracle spread lower bound */), nil
}

func validateOrder(bibliophile b.BibliophileClient, orderType ob.OrderType, encodedOrder []byte, side Side, fillAmount *big.Int) (metadata *Metadata, err error) {
	if orderType == ob.Limit {
		order, err := hu.DecodeLimitOrder(encodedOrder)
		if err != nil {
			return &Metadata{OrderHash: common.Hash{}}, err
		}
		return validateExecuteLimitOrder(bibliophile, order, side, fillAmount)
	}
	if orderType == ob.IOC {
		order, err := hu.DecodeIOCOrder(encodedOrder)
		if err != nil {
			return &Metadata{OrderHash: common.Hash{}}, err
		}
		return validateExecuteIOCOrder(bibliophile, order, side, fillAmount)
	}
	if orderType == ob.Signed {
		order, err := hu.DecodeSignedOrder(encodedOrder)
		if err != nil {
			return &Metadata{OrderHash: common.Hash{}}, err
		}
		return validateExecuteSignedOrder(bibliophile, order, side, fillAmount)
	}
	return &Metadata{OrderHash: common.Hash{}}, errors.New("invalid order type")
}

func validateExecuteLimitOrder(bibliophile b.BibliophileClient, order *ob.LimitOrder, side Side, fillAmount *big.Int) (metadata *Metadata, err error) {
	orderHash, err := order.Hash()
	if err != nil {
		return &Metadata{OrderHash: common.Hash{}}, err
	}
	if err := validateLimitOrderLike(bibliophile, &order.BaseOrder, bibliophile.GetOrderFilledAmount(orderHash), OrderStatus(bibliophile.GetOrderStatus(orderHash)), side, fillAmount); err != nil {
		return &Metadata{OrderHash: orderHash}, err
	}
	return &Metadata{
		AmmIndex:          order.AmmIndex,
		Trader:            order.Trader,
		BaseAssetQuantity: order.BaseAssetQuantity,
		BlockPlaced:       bibliophile.GetBlockPlaced(orderHash),
		Price:             order.Price,
		OrderHash:         orderHash,
		OrderType:         ob.Limit,
		PostOnly:          order.PostOnly,
	}, nil
}

func validateExecuteIOCOrder(bibliophile b.BibliophileClient, order *ob.IOCOrder, side Side, fillAmount *big.Int) (metadata *Metadata, err error) {
	orderHash, err := order.Hash()
	if err != nil {
		return &Metadata{OrderHash: common.Hash{}}, err
	}
	if ob.OrderType(order.OrderType) != ob.IOC {
		return &Metadata{OrderHash: orderHash}, errors.New("not ioc order")
	}
	if order.ExpireAt.Uint64() < bibliophile.GetTimeStamp() {
		return &Metadata{OrderHash: orderHash}, errors.New("ioc expired")
	}
	if err := validateLimitOrderLike(bibliophile, &order.BaseOrder, bibliophile.IOC_GetOrderFilledAmount(orderHash), OrderStatus(bibliophile.IOC_GetOrderStatus(orderHash)), side, fillAmount); err != nil {
		return &Metadata{OrderHash: orderHash}, err
	}
	return &Metadata{
		AmmIndex:          order.AmmIndex,
		Trader:            order.Trader,
		BaseAssetQuantity: order.BaseAssetQuantity,
		BlockPlaced:       bibliophile.IOC_GetBlockPlaced(orderHash),
		Price:             order.Price,
		OrderHash:         orderHash,
		OrderType:         ob.IOC,
		PostOnly:          false,
	}, nil
}

func validateExecuteSignedOrder(bibliophile b.BibliophileClient, order *hu.SignedOrder, side Side, fillAmount *big.Int) (metadata *Metadata, err error) {
	// these fields are only set in plugin/evm/limit_order.go.NewLimitOrderProcesser
	// however, the above is not invoked after the bootstrap, so we need to set these fields here to serve the precompile validation during the bootstrap
	if hu.VerifyingContract == "" || hu.ChainId == 0 {
		chainId := bibliophile.GetAccessibleState().GetSnowContext().ChainID
		if strings.EqualFold(chainId.String(), "2jfjkB7NkK4v8zoaoWmh5eaABNW6ynjQvemPFZpgPQ7ugrmUXv") { // mainnet
			hu.SetChainIdAndVerifyingSignedOrdersContract(1992, "0x211682829664a5e289885DE21897B094eF289d18")
		}
	}

	orderHash, err := order.Hash()
	if err != nil {
		return &Metadata{OrderHash: common.Hash{}}, err
	}
	trader, signer, err := hu.ValidateSignedOrder(
		order,
		hu.SignedOrderValidationFields{
			OrderHash:          orderHash,
			Now:                bibliophile.GetTimeStamp(),
			ActiveMarketsCount: bibliophile.GetActiveMarketsCount(),
			MinSize:            bibliophile.GetMinSizeRequirement(order.AmmIndex.Int64()),
			PriceMultiplier:    bibliophile.GetPriceMultiplier(bibliophile.GetMarketAddressFromMarketID(order.AmmIndex.Int64())),
			Status:             bibliophile.GetSignedOrderStatus(orderHash),
		},
	)
	if err != nil {
		return &Metadata{OrderHash: orderHash}, err
	}

	log.Info("validateExecuteSignedOrder", "trader", trader, "signer", signer, "orderHash", orderHash)
	if trader != signer && !bibliophile.IsTradingAuthority(trader, signer) {
		return &Metadata{OrderHash: orderHash}, hu.ErrNoTradingAuthority
	}

	// M1, M2
	if err := validateLimitOrderLike(bibliophile, &order.BaseOrder, bibliophile.GetSignedOrderFilledAmount(orderHash), Placed, side, fillAmount); err != nil {
		return &Metadata{OrderHash: orderHash}, err
	}

	// M3
	if !bibliophile.HasReferrer(order.Trader) {
		return &Metadata{OrderHash: orderHash}, ErrNoReferrer
	}

	return &Metadata{
		AmmIndex:          order.AmmIndex,
		Trader:            order.Trader,
		BaseAssetQuantity: order.BaseAssetQuantity,
		BlockPlaced:       big.NewInt(0), // will always be treated as a maker order
		Price:             order.Price,
		OrderHash:         orderHash,
		OrderType:         ob.Signed,
		PostOnly:          true,
	}, nil
}

func validateLimitOrderLike(bibliophile b.BibliophileClient, order *hu.BaseOrder, filledAmount *big.Int, status OrderStatus, side Side, fillAmount *big.Int) error {
	if status != Placed {
		return ErrInvalidOrder
	}

	// in case of liquidations, side of the order is determined by the sign of the base asset quantity, so basically base asset quantity check is redundant
	if side == Liquidation {
		if order.BaseAssetQuantity.Sign() > 0 {
			side = Long
		} else if order.BaseAssetQuantity.Sign() < 0 {
			side = Short
			fillAmount = new(big.Int).Neg(fillAmount)
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
		return errors.New("invalid side")
	}
	return nil
}

// Common
func reducesPosition(positionSize *big.Int, baseAssetQuantity *big.Int) bool {
	if positionSize.Sign() == 1 && baseAssetQuantity.Sign() == -1 && big.NewInt(0).Add(positionSize, baseAssetQuantity).Sign() != -1 {
		return true
	}
	if positionSize.Sign() == -1 && baseAssetQuantity.Sign() == 1 && big.NewInt(0).Add(positionSize, baseAssetQuantity).Sign() != 1 {
		return true
	}
	return false
}

func getRequiredMargin(bibliophile b.BibliophileClient, order ILimitOrderBookOrder) *big.Int {
	price := order.Price
	upperBound, _ := bibliophile.GetUpperAndLowerBoundForMarket(order.AmmIndex.Int64())
	if order.BaseAssetQuantity.Sign() == -1 && order.Price.Cmp(upperBound) == -1 {
		price = upperBound
	}
	quoteAsset := big.NewInt(0).Abs(big.NewInt(0).Div(new(big.Int).Mul(order.BaseAssetQuantity, price), big.NewInt(1e18)))
	requiredMargin := big.NewInt(0).Div(big.NewInt(0).Mul(bibliophile.GetMinAllowableMargin(), quoteAsset), big.NewInt(1e6))
	takerFee := big.NewInt(0).Div(big.NewInt(0).Mul(quoteAsset, bibliophile.GetTakerFee()), big.NewInt(1e6))
	requiredMargin.Add(requiredMargin, takerFee)
	return requiredMargin
}

func formatOrder(orderBytes []byte) interface{} {
	decodeStep0, err := hu.DecodeTypeAndEncodedOrder(orderBytes)
	if err != nil {
		return orderBytes
	}

	if decodeStep0.OrderType == ob.Limit {
		order, err := hu.DecodeLimitOrder(decodeStep0.EncodedOrder)
		if err != nil {
			return decodeStep0
		}
		orderJson := order.Map()
		orderHash, err := order.Hash()
		if err != nil {
			return orderJson
		}
		orderJson["hash"] = orderHash.String()
		return orderJson
	}
	if decodeStep0.OrderType == ob.IOC {
		order, err := hu.DecodeIOCOrder(decodeStep0.EncodedOrder)
		if err != nil {
			return decodeStep0
		}
		orderJson := order.Map()
		orderHash, err := order.Hash()
		if err != nil {
			return orderJson
		}
		orderJson["hash"] = orderHash.String()
		return orderJson
	}
	return nil
}

func getValidateOrdersAndDetermineFillPriceErrorOutput(err error, element BadElement, orderHash common.Hash) ValidateOrdersAndDetermineFillPriceOutput {
	// need to provide an empty res because PackValidateOrdersAndDetermineFillPriceOutput fails if FillPrice is nil, and if res.Instructions[0].AmmIndex is nil
	emptyRes := IOrderHandlerMatchingValidationRes{
		Instructions: [2]IClearingHouseInstruction{
			IClearingHouseInstruction{AmmIndex: big.NewInt(0)},
			IClearingHouseInstruction{AmmIndex: big.NewInt(0)},
		},
		OrderTypes:    [2]uint8{},
		EncodedOrders: [2][]byte{},
		FillPrice:     big.NewInt(0),
	}

	var errorString string
	if err != nil {
		// should always be true
		errorString = err.Error()
	}
	if (element == Order0 || element == Order1) && orderHash != (common.Hash{}) {
		emptyRes.Instructions[element].OrderHash = orderHash
	}
	return ValidateOrdersAndDetermineFillPriceOutput{Err: errorString, Element: uint8(element), Res: emptyRes}
}

func getValidateLiquidationOrderAndDetermineFillPriceErrorOutput(err error, element BadElement, orderHash common.Hash) ValidateLiquidationOrderAndDetermineFillPriceOutput {
	emptyRes := IOrderHandlerLiquidationMatchingValidationRes{
		Instruction:  IClearingHouseInstruction{AmmIndex: big.NewInt(0)},
		OrderType:    0,
		EncodedOrder: []byte{},
		FillPrice:    big.NewInt(0),
		FillAmount:   big.NewInt(0),
	}

	var errorString string
	if err != nil {
		// should always be true
		errorString = err.Error()
	}
	if element == Order0 && orderHash != (common.Hash{}) {
		emptyRes.Instruction.OrderHash = orderHash
	}
	return ValidateLiquidationOrderAndDetermineFillPriceOutput{Err: errorString, Element: uint8(element), Res: emptyRes}
}
