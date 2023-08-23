package bibliophile

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/contract"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	MARGIN_ACCOUNT_GENESIS_ADDRESS        = "0x0300000000000000000000000000000000000001"
	VAR_MARGIN_MAPPING_STORAGE_SLOT int64 = 10
	VAR_RESERVED_MARGIN_SLOT        int64 = 11
	MA_MIN_ALLOWABLE_MARGIN_SLOT    int64 = 12
)

func GetNormalizedMargin(stateDB contract.StateDB, trader common.Address) *big.Int {
	// this is only written for single hUSD collateral
	// TODO: generalize for multiple collaterals
	return getMargin(stateDB, big.NewInt(0), trader)
}

// GetMinAllowableMarginMA returns the minimum allowable margin for a trader set in MarginAccount
func getMinAllowableMarginMA(stateDB contract.StateDB) *big.Int {
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(CLEARING_HOUSE_GENESIS_ADDRESS), common.BytesToHash(common.LeftPadBytes(big.NewInt(MA_MIN_ALLOWABLE_MARGIN_SLOT).Bytes(), 32))).Bytes())
}

func getMargin(stateDB contract.StateDB, collateralIdx *big.Int, trader common.Address) *big.Int {
	marginStorageSlot := crypto.Keccak256(append(common.LeftPadBytes(collateralIdx.Bytes(), 32), common.LeftPadBytes(big.NewInt(VAR_MARGIN_MAPPING_STORAGE_SLOT).Bytes(), 32)...))
	marginStorageSlot = crypto.Keccak256(append(common.LeftPadBytes(trader.Bytes(), 32), marginStorageSlot...))
	return fromTwosComplement(stateDB.GetState(common.HexToAddress(MARGIN_ACCOUNT_GENESIS_ADDRESS), common.BytesToHash(marginStorageSlot)).Bytes())
}

func getReservedMargin(stateDB contract.StateDB, trader common.Address) *big.Int {
	baseMappingHash := crypto.Keccak256(append(common.LeftPadBytes(trader.Bytes(), 32), common.LeftPadBytes(big.NewInt(VAR_RESERVED_MARGIN_SLOT).Bytes(), 32)...))
	return stateDB.GetState(common.HexToAddress(MARGIN_ACCOUNT_GENESIS_ADDRESS), common.BytesToHash(baseMappingHash)).Big()
}

func GetAvailableMargin(stateDB contract.StateDB, trader common.Address) *big.Int {
	//As this function is implemented after V2ActivationDate we can hardcode. Change this if we do another migration
	timestamp := big.NewInt(0).Add(V2ActivationDate, big.NewInt(1))
	includeFundingPayment := true
	mode := uint8(1) //Min_Allowable_Margin
	output := GetNotionalPositionAndMargin(stateDB, &GetNotionalPositionAndMarginInput{Trader: trader, IncludeFundingPayments: includeFundingPayment, Mode: mode}, timestamp)
	notionalPostion := output.NotionalPosition
	margin := output.Margin
	utitlizedMargin := divide1e6(big.NewInt(0).Mul(notionalPostion, getMinAllowableMarginMA(stateDB)))
	reservedMargin := getReservedMargin(stateDB, trader)
	return big.NewInt(0).Sub(big.NewInt(0).Sub(margin, utitlizedMargin), reservedMargin)
}
