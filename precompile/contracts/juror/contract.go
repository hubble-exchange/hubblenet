// Code generated
// This file is a generated precompile contract config with stubbed abstract functions.
// The file is generated by a template. Please inspect every code and comment in this file before use.

package juror

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ava-labs/subnet-evm/accounts/abi"
	"github.com/ava-labs/subnet-evm/precompile/contract"
	"github.com/ava-labs/subnet-evm/precompile/contracts/bibliophile"

	_ "embed"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

const (
	// Gas costs for each function. These are set to 1 by default.
	// You should set a gas cost for each function in your contract.
	// Generally, you should not set gas costs very low as this may cause your network to be vulnerable to DoS attacks.
	// There are some predefined gas costs in contract/utils.go that you can use.
	GetBaseQuoteGasCost                                  uint64 = 69 /* SET A GAS COST HERE */
	GetPrevTickGasCost                                   uint64 = 69 /* SET A GAS COST HERE */
	GetQuoteGasCost                                      uint64 = 69 /* SET A GAS COST HERE */
	SampleImpactAskGasCost                               uint64 = 69 /* SET A GAS COST HERE */
	SampleImpactBidGasCost                               uint64 = 69 /* SET A GAS COST HERE */
	ValidateCancelLimitOrderGasCost                      uint64 = 69 /* SET A GAS COST HERE */
	ValidateLiquidationOrderAndDetermineFillPriceGasCost uint64 = 69 /* SET A GAS COST HERE */
	ValidateOrdersAndDetermineFillPriceGasCost           uint64 = 69 /* SET A GAS COST HERE */
	ValidatePlaceIOCOrdersGasCost                        uint64 = 69 /* SET A GAS COST HERE */
	ValidatePlaceLimitOrderGasCost                       uint64 = 69 /* SET A GAS COST HERE */
)

// CUSTOM CODE STARTS HERE
// Reference imports to suppress errors from unused imports. This code and any unnecessary imports can be removed.
var (
	_ = abi.JSON
	_ = errors.New
	_ = big.NewInt
)

// Singleton StatefulPrecompiledContract and signatures.
var (

	// JurorRawABI contains the raw ABI of Juror contract.
	//go:embed contract.abi
	JurorRawABI string

	JurorABI = contract.ParseABI(JurorRawABI)

	JurorPrecompile = createJurorPrecompile()
)

// IClearingHouseInstruction is an auto generated low-level Go binding around an user-defined struct.
type IClearingHouseInstruction struct {
	AmmIndex  *big.Int
	Trader    common.Address
	OrderHash [32]byte
	Mode      uint8
}

// IImmediateOrCancelOrdersOrder is an auto generated low-level Go binding around an user-defined struct.
type IImmediateOrCancelOrdersOrder struct {
	OrderType         uint8
	ExpireAt          *big.Int
	AmmIndex          *big.Int
	Trader            common.Address
	BaseAssetQuantity *big.Int
	Price             *big.Int
	Salt              *big.Int
	ReduceOnly        bool
}

// ILimitOrderBookOrderV2 is an auto generated low-level Go binding around an user-defined struct.
type ILimitOrderBookOrderV2 struct {
	AmmIndex          *big.Int       `json: "ammIndex"`
	Trader            common.Address `json: "trader"`
	BaseAssetQuantity *big.Int       `json: "baseAssetQuantity"`
	Price             *big.Int       `json: "price"`
	Salt              *big.Int       `json: "salt"`
	ReduceOnly        bool           `json: "reduceOnly"`
	PostOnly          bool           `json: "postOnly"`
}

// IOrderHandlerCancelOrderRes is an auto generated low-level Go binding around an user-defined struct.
type IOrderHandlerCancelOrderRes struct {
	UnfilledAmount *big.Int
	Amm            common.Address
}

// IOrderHandlerPlaceOrderRes is an auto generated low-level Go binding around an user-defined struct.
type IOrderHandlerPlaceOrderRes struct {
	ReserveAmount *big.Int
	Amm           common.Address
}

type GetBaseQuoteInput struct {
	Amm           common.Address
	QuoteQuantity *big.Int
}

type GetPrevTickInput struct {
	Amm   common.Address
	IsBid bool
	Tick  *big.Int
}

type GetQuoteInput struct {
	Amm               common.Address
	BaseAssetQuantity *big.Int
}

type ValidateCancelLimitOrderInput struct {
	Order           ILimitOrderBookOrderV2
	Trader          common.Address
	AssertLowMargin bool
}

type ValidateCancelLimitOrderOutput struct {
	Err       string
	OrderHash [32]byte
	Res       IOrderHandlerCancelOrderRes
}

type ValidateLiquidationOrderAndDetermineFillPriceInput struct {
	Data              []byte
	LiquidationAmount *big.Int
}

type ValidateLiquidationOrderAndDetermineFillPriceOutput struct {
	Instruction  IClearingHouseInstruction
	OrderType    uint8
	EncodedOrder []byte
	FillPrice    *big.Int
	FillAmount   *big.Int
}

type ValidateOrdersAndDetermineFillPriceInput struct {
	Data       [2][]byte
	FillAmount *big.Int
}

type ValidateOrdersAndDetermineFillPriceOutput struct {
	Instructions  [2]IClearingHouseInstruction
	OrderTypes    [2]uint8
	EncodedOrders [2][]byte
	FillPrice     *big.Int
}

type ValidatePlaceIOCOrdersInput struct {
	Orders []IImmediateOrCancelOrdersOrder
	Sender common.Address
}

type ValidatePlaceLimitOrderInput struct {
	Order  ILimitOrderBookOrderV2
	Trader common.Address
}

type ValidatePlaceLimitOrderOutput struct {
	Errs      string
	Orderhash [32]byte
	Res       IOrderHandlerPlaceOrderRes
}

// UnpackGetBaseQuoteInput attempts to unpack [input] as GetBaseQuoteInput
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackGetBaseQuoteInput(input []byte) (GetBaseQuoteInput, error) {
	inputStruct := GetBaseQuoteInput{}
	err := JurorABI.UnpackInputIntoInterface(&inputStruct, "getBaseQuote", input)

	return inputStruct, err
}

// PackGetBaseQuote packs [inputStruct] of type GetBaseQuoteInput into the appropriate arguments for getBaseQuote.
func PackGetBaseQuote(inputStruct GetBaseQuoteInput) ([]byte, error) {
	return JurorABI.Pack("getBaseQuote", inputStruct.Amm, inputStruct.QuoteQuantity)
}

// PackGetBaseQuoteOutput attempts to pack given base of type *big.Int
// to conform the ABI outputs.
func PackGetBaseQuoteOutput(base *big.Int) ([]byte, error) {
	return JurorABI.PackOutput("getBaseQuote", base)
}

func getBaseQuote(accessibleState contract.AccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, GetBaseQuoteGasCost); err != nil {
		return nil, 0, err
	}
	// attempts to unpack [input] into the arguments to the GetBaseQuoteInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [inputStruct] variable in your code
	inputStruct, err := UnpackGetBaseQuoteInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	// CUSTOM CODE STARTS HERE
	bibliophile := bibliophile.NewBibliophileClient(accessibleState)
	baseQuote := GetBaseQuote(bibliophile, inputStruct.Amm, inputStruct.QuoteQuantity)

	packedOutput, err := PackGetBaseQuoteOutput(baseQuote)
	if err != nil {
		return nil, remainingGas, err
	}

	// Return the packed output and the remaining gas
	return packedOutput, remainingGas, nil
}

// UnpackGetPrevTickInput attempts to unpack [input] as GetPrevTickInput
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackGetPrevTickInput(input []byte) (GetPrevTickInput, error) {
	inputStruct := GetPrevTickInput{}
	err := JurorABI.UnpackInputIntoInterface(&inputStruct, "getPrevTick", input)

	return inputStruct, err
}

// PackGetPrevTick packs [inputStruct] of type GetPrevTickInput into the appropriate arguments for getPrevTick.
func PackGetPrevTick(inputStruct GetPrevTickInput) ([]byte, error) {
	return JurorABI.Pack("getPrevTick", inputStruct.Amm, inputStruct.IsBid, inputStruct.Tick)
}

// PackGetPrevTickOutput attempts to pack given prevTick of type *big.Int
// to conform the ABI outputs.
func PackGetPrevTickOutput(prevTick *big.Int) ([]byte, error) {
	return JurorABI.PackOutput("getPrevTick", prevTick)
}

func getPrevTick(accessibleState contract.AccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, GetPrevTickGasCost); err != nil {
		return nil, 0, err
	}
	// attempts to unpack [input] into the arguments to the GetPrevTickInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [inputStruct] variable in your code
	inputStruct, err := UnpackGetPrevTickInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	// CUSTOM CODE STARTS HERE
	bibliophile := bibliophile.NewBibliophileClient(accessibleState)
	prevTick, err := GetPrevTick(bibliophile, inputStruct)
	if err != nil {
		return nil, remainingGas, err
	}

	packedOutput, err := PackGetPrevTickOutput(prevTick)
	if err != nil {
		return nil, remainingGas, err
	}

	// Return the packed output and the remaining gas
	return packedOutput, remainingGas, nil
}

// UnpackGetQuoteInput attempts to unpack [input] as GetQuoteInput
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackGetQuoteInput(input []byte) (GetQuoteInput, error) {
	inputStruct := GetQuoteInput{}
	err := JurorABI.UnpackInputIntoInterface(&inputStruct, "getQuote", input)

	return inputStruct, err
}

// PackGetQuote packs [inputStruct] of type GetQuoteInput into the appropriate arguments for getQuote.
func PackGetQuote(inputStruct GetQuoteInput) ([]byte, error) {
	return JurorABI.Pack("getQuote", inputStruct.Amm, inputStruct.BaseAssetQuantity)
}

// PackGetQuoteOutput attempts to pack given quote of type *big.Int
// to conform the ABI outputs.
func PackGetQuoteOutput(quote *big.Int) ([]byte, error) {
	return JurorABI.PackOutput("getQuote", quote)
}

func getQuote(accessibleState contract.AccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, GetQuoteGasCost); err != nil {
		return nil, 0, err
	}
	// attempts to unpack [input] into the arguments to the GetQuoteInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [inputStruct] variable in your code
	inputStruct, err := UnpackGetQuoteInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	// CUSTOM CODE STARTS HERE
	bibliophile := bibliophile.NewBibliophileClient(accessibleState)
	quote := GetQuote(bibliophile, inputStruct.Amm, inputStruct.BaseAssetQuantity)

	packedOutput, err := PackGetQuoteOutput(quote)
	if err != nil {
		return nil, remainingGas, err
	}

	// Return the packed output and the remaining gas
	return packedOutput, remainingGas, nil
}

// UnpackSampleImpactAskInput attempts to unpack [input] into the common.Address type argument
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackSampleImpactAskInput(input []byte) (common.Address, error) {
	res, err := JurorABI.UnpackInput("sampleImpactAsk", input)
	if err != nil {
		return *new(common.Address), err
	}
	unpacked := *abi.ConvertType(res[0], new(common.Address)).(*common.Address)
	return unpacked, nil
}

// PackSampleImpactAsk packs [amm] of type common.Address into the appropriate arguments for sampleImpactAsk.
// the packed bytes include selector (first 4 func signature bytes).
// This function is mostly used for tests.
func PackSampleImpactAsk(amm common.Address) ([]byte, error) {
	return JurorABI.Pack("sampleImpactAsk", amm)
}

// PackSampleImpactAskOutput attempts to pack given impactAsk of type *big.Int
// to conform the ABI outputs.
func PackSampleImpactAskOutput(impactAsk *big.Int) ([]byte, error) {
	return JurorABI.PackOutput("sampleImpactAsk", impactAsk)
}

func sampleImpactAsk(accessibleState contract.AccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, SampleImpactAskGasCost); err != nil {
		return nil, 0, err
	}
	// attempts to unpack [input] into the arguments to the SampleImpactAskInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [inputStruct] variable in your code
	inputStruct, err := UnpackSampleImpactAskInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	// CUSTOM CODE STARTS HERE
	bibliophile := bibliophile.NewBibliophileClient(accessibleState)
	output := SampleImpactAsk(bibliophile, inputStruct)

	packedOutput, err := PackSampleImpactAskOutput(output)
	if err != nil {
		return nil, remainingGas, err
	}

	// Return the packed output and the remaining gas
	return packedOutput, remainingGas, nil
}

// UnpackSampleImpactBidInput attempts to unpack [input] into the common.Address type argument
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackSampleImpactBidInput(input []byte) (common.Address, error) {
	res, err := JurorABI.UnpackInput("sampleImpactBid", input)
	if err != nil {
		return *new(common.Address), err
	}
	unpacked := *abi.ConvertType(res[0], new(common.Address)).(*common.Address)
	return unpacked, nil
}

// PackSampleImpactBid packs [amm] of type common.Address into the appropriate arguments for sampleImpactBid.
// the packed bytes include selector (first 4 func signature bytes).
// This function is mostly used for tests.
func PackSampleImpactBid(amm common.Address) ([]byte, error) {
	return JurorABI.Pack("sampleImpactBid", amm)
}

// PackSampleImpactBidOutput attempts to pack given impactBid of type *big.Int
// to conform the ABI outputs.
func PackSampleImpactBidOutput(impactBid *big.Int) ([]byte, error) {
	return JurorABI.PackOutput("sampleImpactBid", impactBid)
}

func sampleImpactBid(accessibleState contract.AccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, SampleImpactBidGasCost); err != nil {
		return nil, 0, err
	}
	// attempts to unpack [input] into the arguments to the SampleImpactBidInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [ammAddress] variable in your code
	ammAddress, err := UnpackSampleImpactBidInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	// CUSTOM CODE STARTS HERE
	bibliophile := bibliophile.NewBibliophileClient(accessibleState)
	output := SampleImpactBid(bibliophile, ammAddress)

	packedOutput, err := PackSampleImpactBidOutput(output)
	if err != nil {
		return nil, remainingGas, err
	}

	// Return the packed output and the remaining gas
	return packedOutput, remainingGas, nil
}

// UnpackValidateCancelLimitOrderInput attempts to unpack [input] as ValidateCancelLimitOrderInput
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackValidateCancelLimitOrderInput(input []byte) (ValidateCancelLimitOrderInput, error) {
	inputStruct := ValidateCancelLimitOrderInput{}
	err := JurorABI.UnpackInputIntoInterface(&inputStruct, "validateCancelLimitOrder", input)

	return inputStruct, err
}

// PackValidateCancelLimitOrder packs [inputStruct] of type ValidateCancelLimitOrderInput into the appropriate arguments for validateCancelLimitOrder.
func PackValidateCancelLimitOrder(inputStruct ValidateCancelLimitOrderInput) ([]byte, error) {
	return JurorABI.Pack("validateCancelLimitOrder", inputStruct.Order, inputStruct.Trader, inputStruct.AssertLowMargin)
}

// PackValidateCancelLimitOrderOutput attempts to pack given [outputStruct] of type ValidateCancelLimitOrderOutput
// to conform the ABI outputs.
func PackValidateCancelLimitOrderOutput(outputStruct ValidateCancelLimitOrderOutput) ([]byte, error) {
	return JurorABI.PackOutput("validateCancelLimitOrder",
		outputStruct.Err,
		outputStruct.OrderHash,
		outputStruct.Res,
	)
}

func validateCancelLimitOrder(accessibleState contract.AccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, ValidateCancelLimitOrderGasCost); err != nil {
		return nil, 0, err
	}
	// attempts to unpack [input] into the arguments to the ValidateCancelLimitOrderInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [inputStruct] variable in your code
	inputStruct, err := UnpackValidateCancelLimitOrderInput(input)
	if err != nil {
		return nil, remainingGas, err
	}
	// CUSTOM CODE STARTS HERE
	bibliophile := bibliophile.NewBibliophileClient(accessibleState)
	output := ValidateCancelLimitOrderV2(bibliophile, &inputStruct)
	packedOutput, err := PackValidateCancelLimitOrderOutput(*output)
	if err != nil {
		return nil, remainingGas, err
	}
	// Return the packed output and the remaining gas
	return packedOutput, remainingGas, nil
}

// UnpackValidateLiquidationOrderAndDetermineFillPriceInput attempts to unpack [input] as ValidateLiquidationOrderAndDetermineFillPriceInput
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackValidateLiquidationOrderAndDetermineFillPriceInput(input []byte) (ValidateLiquidationOrderAndDetermineFillPriceInput, error) {
	inputStruct := ValidateLiquidationOrderAndDetermineFillPriceInput{}
	err := JurorABI.UnpackInputIntoInterface(&inputStruct, "validateLiquidationOrderAndDetermineFillPrice", input)

	return inputStruct, err
}

// PackValidateLiquidationOrderAndDetermineFillPrice packs [inputStruct] of type ValidateLiquidationOrderAndDetermineFillPriceInput into the appropriate arguments for validateLiquidationOrderAndDetermineFillPrice.
func PackValidateLiquidationOrderAndDetermineFillPrice(inputStruct ValidateLiquidationOrderAndDetermineFillPriceInput) ([]byte, error) {
	return JurorABI.Pack("validateLiquidationOrderAndDetermineFillPrice", inputStruct.Data, inputStruct.LiquidationAmount)
}

// PackValidateLiquidationOrderAndDetermineFillPriceOutput attempts to pack given [outputStruct] of type ValidateLiquidationOrderAndDetermineFillPriceOutput
// to conform the ABI outputs.
func PackValidateLiquidationOrderAndDetermineFillPriceOutput(outputStruct ValidateLiquidationOrderAndDetermineFillPriceOutput) ([]byte, error) {
	return JurorABI.PackOutput("validateLiquidationOrderAndDetermineFillPrice",
		outputStruct.Instruction,
		outputStruct.OrderType,
		outputStruct.EncodedOrder,
		outputStruct.FillPrice,
		outputStruct.FillAmount,
	)
}

func validateLiquidationOrderAndDetermineFillPrice(accessibleState contract.AccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, ValidateLiquidationOrderAndDetermineFillPriceGasCost); err != nil {
		return nil, 0, err
	}
	// attempts to unpack [input] into the arguments to the ValidateLiquidationOrderAndDetermineFillPriceInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [inputStruct] variable in your code
	inputStruct, err := UnpackValidateLiquidationOrderAndDetermineFillPriceInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	// CUSTOM CODE STARTS HERE
	bibliophile := bibliophile.NewBibliophileClient(accessibleState)
	output, err := ValidateLiquidationOrderAndDetermineFillPrice(bibliophile, &inputStruct)
	if err != nil {
		return nil, remainingGas, err
	}
	packedOutput, err := PackValidateLiquidationOrderAndDetermineFillPriceOutput(*output)
	if err != nil {
		return nil, remainingGas, err
	}

	// Return the packed output and the remaining gas
	return packedOutput, remainingGas, nil
}

// UnpackValidateOrdersAndDetermineFillPriceInput attempts to unpack [input] as ValidateOrdersAndDetermineFillPriceInput
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackValidateOrdersAndDetermineFillPriceInput(input []byte) (ValidateOrdersAndDetermineFillPriceInput, error) {
	inputStruct := ValidateOrdersAndDetermineFillPriceInput{}
	err := JurorABI.UnpackInputIntoInterface(&inputStruct, "validateOrdersAndDetermineFillPrice", input)

	return inputStruct, err
}

// PackValidateOrdersAndDetermineFillPrice packs [inputStruct] of type ValidateOrdersAndDetermineFillPriceInput into the appropriate arguments for validateOrdersAndDetermineFillPrice.
func PackValidateOrdersAndDetermineFillPrice(inputStruct ValidateOrdersAndDetermineFillPriceInput) ([]byte, error) {
	return JurorABI.Pack("validateOrdersAndDetermineFillPrice", inputStruct.Data, inputStruct.FillAmount)
}

// PackValidateOrdersAndDetermineFillPriceOutput attempts to pack given [outputStruct] of type ValidateOrdersAndDetermineFillPriceOutput
// to conform the ABI outputs.
func PackValidateOrdersAndDetermineFillPriceOutput(outputStruct ValidateOrdersAndDetermineFillPriceOutput) ([]byte, error) {
	return JurorABI.PackOutput("validateOrdersAndDetermineFillPrice",
		outputStruct.Instructions,
		outputStruct.OrderTypes,
		outputStruct.EncodedOrders,
		outputStruct.FillPrice,
	)
}

func validateOrdersAndDetermineFillPrice(accessibleState contract.AccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, ValidateOrdersAndDetermineFillPriceGasCost); err != nil {
		return nil, 0, err
	}
	// attempts to unpack [input] into the arguments to the ValidateOrdersAndDetermineFillPriceInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [inputStruct] variable in your code
	inputStruct, err := UnpackValidateOrdersAndDetermineFillPriceInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	// CUSTOM CODE STARTS HERE
	bibliophile := bibliophile.NewBibliophileClient(accessibleState)
	output, err := ValidateOrdersAndDetermineFillPrice(bibliophile, &inputStruct)
	if err != nil {
		log.Error("validateOrdersAndDetermineFillPrice", "order0", formatOrder(inputStruct.Data[0]), "order1", formatOrder(inputStruct.Data[1]), "fillAmount", inputStruct.FillAmount, "err", err, "block", accessibleState.GetBlockContext().Number())
		return nil, remainingGas, err
	}
	packedOutput, err := PackValidateOrdersAndDetermineFillPriceOutput(*output)
	if err != nil {
		return nil, remainingGas, err
	}

	// Return the packed output and the remaining gas
	return packedOutput, remainingGas, nil
}

// UnpackValidatePlaceIOCOrdersInput attempts to unpack [input] as ValidatePlaceIOCOrdersInput
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackValidatePlaceIOCOrdersInput(input []byte) (ValidatePlaceIOCOrdersInput, error) {
	inputStruct := ValidatePlaceIOCOrdersInput{}
	err := JurorABI.UnpackInputIntoInterface(&inputStruct, "validatePlaceIOCOrders", input)

	return inputStruct, err
}

// PackValidatePlaceIOCOrders packs [inputStruct] of type ValidatePlaceIOCOrdersInput into the appropriate arguments for validatePlaceIOCOrders.
func PackValidatePlaceIOCOrders(inputStruct ValidatePlaceIOCOrdersInput) ([]byte, error) {
	return JurorABI.Pack("validatePlaceIOCOrders", inputStruct.Orders, inputStruct.Sender)
}

// PackValidatePlaceIOCOrdersOutput attempts to pack given orderHashes of type [][32]byte
// to conform the ABI outputs.
func PackValidatePlaceIOCOrdersOutput(orderHashes [][32]byte) ([]byte, error) {
	return JurorABI.PackOutput("validatePlaceIOCOrders", orderHashes)
}

func validatePlaceIOCOrders(accessibleState contract.AccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, ValidatePlaceIOCOrdersGasCost); err != nil {
		return nil, 0, err
	}
	// attempts to unpack [input] into the arguments to the ValidatePlaceIOCOrdersInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [inputStruct] variable in your code
	inputStruct, err := UnpackValidatePlaceIOCOrdersInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	// CUSTOM CODE STARTS HERE
	bibliophile := bibliophile.NewBibliophileClient(accessibleState)
	output, err := ValidatePlaceIOCOrders(bibliophile, &inputStruct)
	if err != nil {
		log.Error("validatePlaceIOCOrders", "error", err, "inputStruct", inputStruct, "block", accessibleState.GetBlockContext().Number())
		return nil, remainingGas, err
	}
	packedOutput, err := PackValidatePlaceIOCOrdersOutput(output)
	if err != nil {
		return nil, remainingGas, err
	}

	// Return the packed output and the remaining gas
	return packedOutput, remainingGas, nil
}

// UnpackValidatePlaceLimitOrderInput attempts to unpack [input] as ValidatePlaceLimitOrderInput
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackValidatePlaceLimitOrderInput(input []byte) (ValidatePlaceLimitOrderInput, error) {
	inputStruct := ValidatePlaceLimitOrderInput{}
	err := JurorABI.UnpackInputIntoInterface(&inputStruct, "validatePlaceLimitOrder", input)

	return inputStruct, err
}

// PackValidatePlaceLimitOrder packs [inputStruct] of type ValidatePlaceLimitOrderInput into the appropriate arguments for validatePlaceLimitOrder.
func PackValidatePlaceLimitOrder(inputStruct ValidatePlaceLimitOrderInput) ([]byte, error) {
	return JurorABI.Pack("validatePlaceLimitOrder", inputStruct.Order, inputStruct.Trader)
}

// PackValidatePlaceLimitOrderOutput attempts to pack given [outputStruct] of type ValidatePlaceLimitOrderOutput
// to conform the ABI outputs.
func PackValidatePlaceLimitOrderOutput(outputStruct ValidatePlaceLimitOrderOutput) ([]byte, error) {
	log.Info("validatePlaceLimitOrder", "outputStruct", outputStruct)
	return JurorABI.PackOutput("validatePlaceLimitOrder",
		outputStruct.Errs,
		outputStruct.Orderhash,
		outputStruct.Res,
	)
}

func validatePlaceLimitOrder(accessibleState contract.AccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, ValidatePlaceLimitOrderGasCost); err != nil {
		return nil, 0, err
	}
	// attempts to unpack [input] into the arguments to the ValidatePlaceLimitOrderInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [inputStruct] variable in your code
	inputStruct, err := UnpackValidatePlaceLimitOrderInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	// CUSTOM CODE STARTS HERE
	bibliophile := bibliophile.NewBibliophileClient(accessibleState)
	output := ValidatePlaceLimitOrderV2(bibliophile, inputStruct.Order, inputStruct.Trader)
	packedOutput, err := PackValidatePlaceLimitOrderOutput(*output)
	if err != nil {
		return nil, remainingGas, err
	}

	// Return the packed output and the remaining gas
	return packedOutput, remainingGas, nil
}

// createJurorPrecompile returns a StatefulPrecompiledContract with getters and setters for the precompile.

func createJurorPrecompile() contract.StatefulPrecompiledContract {
	var functions []*contract.StatefulPrecompileFunction

	abiFunctionMap := map[string]contract.RunStatefulPrecompileFunc{
		// "getBaseQuote":             getBaseQuote,
		"getPrevTick": getPrevTick,
		// "getQuote":                 getQuote,
		"sampleImpactAsk":                               sampleImpactAsk,
		"sampleImpactBid":                               sampleImpactBid,
		"validateCancelLimitOrder":                      validateCancelLimitOrder,
		"validateLiquidationOrderAndDetermineFillPrice": validateLiquidationOrderAndDetermineFillPrice,
		"validateOrdersAndDetermineFillPrice":           validateOrdersAndDetermineFillPrice,
		"validatePlaceIOCOrders":                        validatePlaceIOCOrders,
		"validatePlaceLimitOrder":                       validatePlaceLimitOrder,
	}

	for name, function := range abiFunctionMap {
		method, ok := JurorABI.Methods[name]
		if !ok {
			panic(fmt.Errorf("given method (%s) does not exist in the ABI", name))
		}
		functions = append(functions, contract.NewStatefulPrecompileFunction(method.ID, function))
	}
	// Construct the contract with no fallback function.
	statefulContract, err := contract.NewStatefulPrecompileContract(nil, functions)
	if err != nil {
		panic(err)
	}
	return statefulContract
}
