// Code generated
// This file is a generated precompile contract config with stubbed abstract functions.
// The file is generated by a template. Please inspect every code and comment in this file before use.

package traderviewer

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ava-labs/subnet-evm/accounts/abi"
	"github.com/ava-labs/subnet-evm/precompile/contract"
	"github.com/ava-labs/subnet-evm/vmerrs"

	_ "embed"

	"github.com/ethereum/go-ethereum/common"
)

const (
	// Gas costs for each function. These are set to 1 by default.
	// You should set a gas cost for each function in your contract.
	// Generally, you should not set gas costs very low as this may cause your network to be vulnerable to DoS attacks.
	// There are some predefined gas costs in contract/utils.go that you can use.
	GetCrossMarginAccountDataGasCost              uint64 = 1 /* SET A GAS COST HERE */
	GetNotionalPositionAndMarginGasCost           uint64 = 1 /* SET A GAS COST HERE */
	GetTotalFundingForCrossMarginPositionsGasCost uint64 = 1 /* SET A GAS COST HERE */
	GetTraderDataForMarketGasCost                 uint64 = 1 /* SET A GAS COST HERE */
)

// CUSTOM CODE STARTS HERE
// Reference imports to suppress errors from unused imports. This code and any unnecessary imports can be removed.
var (
	_ = abi.JSON
	_ = errors.New
	_ = big.NewInt
	_ = vmerrs.ErrOutOfGas
	_ = common.Big0
)

// Singleton StatefulPrecompiledContract and signatures.
var (

	// TraderViewerRawABI contains the raw ABI of TraderViewer contract.
	//go:embed contract.abi
	TraderViewerRawABI string

	TraderViewerABI = contract.ParseABI(TraderViewerRawABI)

	TraderViewerPrecompile = createTraderViewerPrecompile()
)

type GetCrossMarginAccountDataInput struct {
	Trader common.Address
	Mode   uint8
}

type GetCrossMarginAccountDataOutput struct {
	NotionalPosition *big.Int
	RequiredMargin   *big.Int
	UnrealizedPnl    *big.Int
	PendingFunding   *big.Int
}

type GetNotionalPositionAndMarginInput struct {
	Trader                 common.Address
	IncludeFundingPayments bool
	Mode                   uint8
}

type GetNotionalPositionAndMarginOutput struct {
	NotionalPosition *big.Int
	Margin           *big.Int
	RequiredMargin   *big.Int
}

type GetTraderDataForMarketInput struct {
	Trader   common.Address
	AmmIndex *big.Int
	Mode     uint8
}

type GetTraderDataForMarketOutput struct {
	IsIsolated       bool
	NotionalPosition *big.Int
	UnrealizedPnl    *big.Int
	RequiredMargin   *big.Int
	PendingFunding   *big.Int
}

// UnpackGetCrossMarginAccountDataInput attempts to unpack [input] as GetCrossMarginAccountDataInput
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackGetCrossMarginAccountDataInput(input []byte) (GetCrossMarginAccountDataInput, error) {
	inputStruct := GetCrossMarginAccountDataInput{}
	// The strict mode in decoding is disabled after Durango. You can re-enable by changing the last argument to true.
	err := TraderViewerABI.UnpackInputIntoInterface(&inputStruct, "getCrossMarginAccountData", input, false)

	return inputStruct, err
}

// PackGetCrossMarginAccountData packs [inputStruct] of type GetCrossMarginAccountDataInput into the appropriate arguments for getCrossMarginAccountData.
func PackGetCrossMarginAccountData(inputStruct GetCrossMarginAccountDataInput) ([]byte, error) {
	return TraderViewerABI.Pack("getCrossMarginAccountData", inputStruct.Trader, inputStruct.Mode)
}

// PackGetCrossMarginAccountDataOutput attempts to pack given [outputStruct] of type GetCrossMarginAccountDataOutput
// to conform the ABI outputs.
func PackGetCrossMarginAccountDataOutput(outputStruct GetCrossMarginAccountDataOutput) ([]byte, error) {
	return TraderViewerABI.PackOutput("getCrossMarginAccountData",
		outputStruct.NotionalPosition,
		outputStruct.RequiredMargin,
		outputStruct.UnrealizedPnl,
		outputStruct.PendingFunding,
	)
}

// UnpackGetCrossMarginAccountDataOutput attempts to unpack [output] as GetCrossMarginAccountDataOutput
// assumes that [output] does not include selector (omits first 4 func signature bytes)
func UnpackGetCrossMarginAccountDataOutput(output []byte) (GetCrossMarginAccountDataOutput, error) {
	outputStruct := GetCrossMarginAccountDataOutput{}
	err := TraderViewerABI.UnpackIntoInterface(&outputStruct, "getCrossMarginAccountData", output)

	return outputStruct, err
}

func getCrossMarginAccountData(accessibleState contract.AccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, GetCrossMarginAccountDataGasCost); err != nil {
		return nil, 0, err
	}
	// attempts to unpack [input] into the arguments to the GetCrossMarginAccountDataInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [inputStruct] variable in your code
	inputStruct, err := UnpackGetCrossMarginAccountDataInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	// CUSTOM CODE STARTS HERE
	_ = inputStruct                            // CUSTOM CODE OPERATES ON INPUT
	var output GetCrossMarginAccountDataOutput // CUSTOM CODE FOR AN OUTPUT
	packedOutput, err := PackGetCrossMarginAccountDataOutput(output)
	if err != nil {
		return nil, remainingGas, err
	}

	// Return the packed output and the remaining gas
	return packedOutput, remainingGas, nil
}

// UnpackGetNotionalPositionAndMarginInput attempts to unpack [input] as GetNotionalPositionAndMarginInput
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackGetNotionalPositionAndMarginInput(input []byte) (GetNotionalPositionAndMarginInput, error) {
	inputStruct := GetNotionalPositionAndMarginInput{}
	// The strict mode in decoding is disabled after Durango. You can re-enable by changing the last argument to true.
	err := TraderViewerABI.UnpackInputIntoInterface(&inputStruct, "getNotionalPositionAndMargin", input, false)

	return inputStruct, err
}

// PackGetNotionalPositionAndMargin packs [inputStruct] of type GetNotionalPositionAndMarginInput into the appropriate arguments for getNotionalPositionAndMargin.
func PackGetNotionalPositionAndMargin(inputStruct GetNotionalPositionAndMarginInput) ([]byte, error) {
	return TraderViewerABI.Pack("getNotionalPositionAndMargin", inputStruct.Trader, inputStruct.IncludeFundingPayments, inputStruct.Mode)
}

// PackGetNotionalPositionAndMarginOutput attempts to pack given [outputStruct] of type GetNotionalPositionAndMarginOutput
// to conform the ABI outputs.
func PackGetNotionalPositionAndMarginOutput(outputStruct GetNotionalPositionAndMarginOutput) ([]byte, error) {
	return TraderViewerABI.PackOutput("getNotionalPositionAndMargin",
		outputStruct.NotionalPosition,
		outputStruct.Margin,
		outputStruct.RequiredMargin,
	)
}

// UnpackGetNotionalPositionAndMarginOutput attempts to unpack [output] as GetNotionalPositionAndMarginOutput
// assumes that [output] does not include selector (omits first 4 func signature bytes)
func UnpackGetNotionalPositionAndMarginOutput(output []byte) (GetNotionalPositionAndMarginOutput, error) {
	outputStruct := GetNotionalPositionAndMarginOutput{}
	err := TraderViewerABI.UnpackIntoInterface(&outputStruct, "getNotionalPositionAndMargin", output)

	return outputStruct, err
}

func getNotionalPositionAndMargin(accessibleState contract.AccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, GetNotionalPositionAndMarginGasCost); err != nil {
		return nil, 0, err
	}
	// attempts to unpack [input] into the arguments to the GetNotionalPositionAndMarginInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [inputStruct] variable in your code
	inputStruct, err := UnpackGetNotionalPositionAndMarginInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	// CUSTOM CODE STARTS HERE
	_ = inputStruct                               // CUSTOM CODE OPERATES ON INPUT
	var output GetNotionalPositionAndMarginOutput // CUSTOM CODE FOR AN OUTPUT
	packedOutput, err := PackGetNotionalPositionAndMarginOutput(output)
	if err != nil {
		return nil, remainingGas, err
	}

	// Return the packed output and the remaining gas
	return packedOutput, remainingGas, nil
}

// UnpackGetTotalFundingForCrossMarginPositionsInput attempts to unpack [input] into the common.Address type argument
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackGetTotalFundingForCrossMarginPositionsInput(input []byte) (common.Address, error) {
	// The strict mode in decoding is disabled after Durango. You can re-enable by changing the last argument to true.
	res, err := TraderViewerABI.UnpackInput("getTotalFundingForCrossMarginPositions", input, false)
	if err != nil {
		return common.Address{}, err
	}
	unpacked := *abi.ConvertType(res[0], new(common.Address)).(*common.Address)
	return unpacked, nil
}

// PackGetTotalFundingForCrossMarginPositions packs [trader] of type common.Address into the appropriate arguments for getTotalFundingForCrossMarginPositions.
// the packed bytes include selector (first 4 func signature bytes).
// This function is mostly used for tests.
func PackGetTotalFundingForCrossMarginPositions(trader common.Address) ([]byte, error) {
	return TraderViewerABI.Pack("getTotalFundingForCrossMarginPositions", trader)
}

// PackGetTotalFundingForCrossMarginPositionsOutput attempts to pack given totalFunding of type *big.Int
// to conform the ABI outputs.
func PackGetTotalFundingForCrossMarginPositionsOutput(totalFunding *big.Int) ([]byte, error) {
	return TraderViewerABI.PackOutput("getTotalFundingForCrossMarginPositions", totalFunding)
}

// UnpackGetTotalFundingForCrossMarginPositionsOutput attempts to unpack given [output] into the *big.Int type output
// assumes that [output] does not include selector (omits first 4 func signature bytes)
func UnpackGetTotalFundingForCrossMarginPositionsOutput(output []byte) (*big.Int, error) {
	res, err := TraderViewerABI.Unpack("getTotalFundingForCrossMarginPositions", output)
	if err != nil {
		return new(big.Int), err
	}
	unpacked := *abi.ConvertType(res[0], new(*big.Int)).(**big.Int)
	return unpacked, nil
}

func getTotalFundingForCrossMarginPositions(accessibleState contract.AccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, GetTotalFundingForCrossMarginPositionsGasCost); err != nil {
		return nil, 0, err
	}
	// attempts to unpack [input] into the arguments to the GetTotalFundingForCrossMarginPositionsInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [inputStruct] variable in your code
	inputStruct, err := UnpackGetTotalFundingForCrossMarginPositionsInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	// CUSTOM CODE STARTS HERE
	_ = inputStruct // CUSTOM CODE OPERATES ON INPUT

	var output *big.Int // CUSTOM CODE FOR AN OUTPUT
	packedOutput, err := PackGetTotalFundingForCrossMarginPositionsOutput(output)
	if err != nil {
		return nil, remainingGas, err
	}

	// Return the packed output and the remaining gas
	return packedOutput, remainingGas, nil
}

// UnpackGetTraderDataForMarketInput attempts to unpack [input] as GetTraderDataForMarketInput
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackGetTraderDataForMarketInput(input []byte) (GetTraderDataForMarketInput, error) {
	inputStruct := GetTraderDataForMarketInput{}
	// The strict mode in decoding is disabled after Durango. You can re-enable by changing the last argument to true.
	err := TraderViewerABI.UnpackInputIntoInterface(&inputStruct, "getTraderDataForMarket", input, false)

	return inputStruct, err
}

// PackGetTraderDataForMarket packs [inputStruct] of type GetTraderDataForMarketInput into the appropriate arguments for getTraderDataForMarket.
func PackGetTraderDataForMarket(inputStruct GetTraderDataForMarketInput) ([]byte, error) {
	return TraderViewerABI.Pack("getTraderDataForMarket", inputStruct.Trader, inputStruct.AmmIndex, inputStruct.Mode)
}

// PackGetTraderDataForMarketOutput attempts to pack given [outputStruct] of type GetTraderDataForMarketOutput
// to conform the ABI outputs.
func PackGetTraderDataForMarketOutput(outputStruct GetTraderDataForMarketOutput) ([]byte, error) {
	return TraderViewerABI.PackOutput("getTraderDataForMarket",
		outputStruct.IsIsolated,
		outputStruct.NotionalPosition,
		outputStruct.UnrealizedPnl,
		outputStruct.RequiredMargin,
		outputStruct.PendingFunding,
	)
}

// UnpackGetTraderDataForMarketOutput attempts to unpack [output] as GetTraderDataForMarketOutput
// assumes that [output] does not include selector (omits first 4 func signature bytes)
func UnpackGetTraderDataForMarketOutput(output []byte) (GetTraderDataForMarketOutput, error) {
	outputStruct := GetTraderDataForMarketOutput{}
	err := TraderViewerABI.UnpackIntoInterface(&outputStruct, "getTraderDataForMarket", output)

	return outputStruct, err
}

func getTraderDataForMarket(accessibleState contract.AccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, GetTraderDataForMarketGasCost); err != nil {
		return nil, 0, err
	}
	// attempts to unpack [input] into the arguments to the GetTraderDataForMarketInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [inputStruct] variable in your code
	inputStruct, err := UnpackGetTraderDataForMarketInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	// CUSTOM CODE STARTS HERE
	_ = inputStruct                         // CUSTOM CODE OPERATES ON INPUT
	var output GetTraderDataForMarketOutput // CUSTOM CODE FOR AN OUTPUT
	packedOutput, err := PackGetTraderDataForMarketOutput(output)
	if err != nil {
		return nil, remainingGas, err
	}

	// Return the packed output and the remaining gas
	return packedOutput, remainingGas, nil
}

// createTraderViewerPrecompile returns a StatefulPrecompiledContract with getters and setters for the precompile.

func createTraderViewerPrecompile() contract.StatefulPrecompiledContract {
	var functions []*contract.StatefulPrecompileFunction

	abiFunctionMap := map[string]contract.RunStatefulPrecompileFunc{
		"getCrossMarginAccountData":              getCrossMarginAccountData,
		"getNotionalPositionAndMargin":           getNotionalPositionAndMargin,
		"getTotalFundingForCrossMarginPositions": getTotalFundingForCrossMarginPositions,
		"getTraderDataForMarket":                 getTraderDataForMarket,
	}

	for name, function := range abiFunctionMap {
		method, ok := TraderViewerABI.Methods[name]
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
