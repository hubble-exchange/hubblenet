package bibliophile

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/contract"

	hu "github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	MARGIN_ACCOUNT_GENESIS_ADDRESS       = "0x0300000000000000000000000000000000000001"
	ORACLE_SLOT                    int64 = 4
	SUPPORTED_COLLATERAL_SLOT      int64 = 8
	VAR_MARGIN_MAPPING_SLOT        int64 = 10
	VAR_RESERVED_MARGIN_SLOT       int64 = 11
)

func GetNormalizedMargin(stateDB contract.StateDB, trader common.Address) *big.Int {
	assets := getCollaterals(stateDB)
	numAssets := len(assets)
	margin := make([]*big.Int, numAssets)
	for i := 0; i < numAssets; i++ {
		margin[i] = getMargin(stateDB, big.NewInt(int64(i)), trader)
	}
	return hu.GetNormalizedMargin(assets, margin)
}

func getMargin(stateDB contract.StateDB, idx *big.Int, trader common.Address) *big.Int {
	marginStorageSlot := crypto.Keccak256(append(common.LeftPadBytes(idx.Bytes(), 32), common.LeftPadBytes(big.NewInt(VAR_MARGIN_MAPPING_SLOT).Bytes(), 32)...))
	marginStorageSlot = crypto.Keccak256(append(common.LeftPadBytes(trader.Bytes(), 32), marginStorageSlot...))
	return fromTwosComplement(stateDB.GetState(common.HexToAddress(MARGIN_ACCOUNT_GENESIS_ADDRESS), common.BytesToHash(marginStorageSlot)).Bytes())
}

func getReservedMargin(stateDB contract.StateDB, trader common.Address) *big.Int {
	baseMappingHash := crypto.Keccak256(append(common.LeftPadBytes(trader.Bytes(), 32), common.LeftPadBytes(big.NewInt(VAR_RESERVED_MARGIN_SLOT).Bytes(), 32)...))
	return stateDB.GetState(common.HexToAddress(MARGIN_ACCOUNT_GENESIS_ADDRESS), common.BytesToHash(baseMappingHash)).Big()
}

func GetAvailableMargin(stateDB contract.StateDB, trader common.Address) *big.Int {
	includeFundingPayment := true
	mode := uint8(1) // Min_Allowable_Margin
	output := getNotionalPositionAndMargin(stateDB, &GetNotionalPositionAndMarginInput{Trader: trader, IncludeFundingPayments: includeFundingPayment, Mode: mode})
	notionalPostion := output.NotionalPosition
	margin := output.Margin
	utitlizedMargin := hu.Div1e6(big.NewInt(0).Mul(notionalPostion, GetMinAllowableMargin(stateDB)))
	reservedMargin := getReservedMargin(stateDB, trader)
	// log.Info("GetAvailableMargin", "trader", trader, "notionalPostion", notionalPostion, "margin", margin, "utitlizedMargin", utitlizedMargin, "reservedMargin", reservedMargin)
	return big.NewInt(0).Sub(big.NewInt(0).Sub(margin, utitlizedMargin), reservedMargin)
}

func getOracleAddress(stateDB contract.StateDB) common.Address {
	return common.BytesToAddress(stateDB.GetState(common.HexToAddress(MARGIN_ACCOUNT_GENESIS_ADDRESS), common.BigToHash(big.NewInt(ORACLE_SLOT))).Bytes())
}

func GetCollaterals(stateDB contract.StateDB) []hu.Collateral {
	numAssets := getCollateralCount(stateDB)
	assets := make([]hu.Collateral, numAssets)
	for i := uint8(0); i < numAssets; i++ {
		assets[i] = getCollateralAt(stateDB, i)
	}
	return assets
}

func getCollateralCount(stateDB contract.StateDB) uint8 {
	rawVal := stateDB.GetState(common.HexToAddress(MARGIN_ACCOUNT_GENESIS_ADDRESS), common.BytesToHash(common.LeftPadBytes(big.NewInt(SUPPORTED_COLLATERAL_SLOT).Bytes(), 32)))
	return uint8(new(big.Int).SetBytes(rawVal.Bytes()).Uint64())
}

func getCollateralAt(stateDB contract.StateDB, idx uint8) hu.Collateral {
	baseSlot := hu.Add(collateralStorageSlot(), big.NewInt(int64(idx)))
	tokenAddress := common.BytesToAddress(stateDB.GetState(common.HexToAddress(MARGIN_ACCOUNT_GENESIS_ADDRESS), common.BigToHash(baseSlot)).Bytes())
	return hu.Collateral{
		Weight:   stateDB.GetState(common.HexToAddress(MARGIN_ACCOUNT_GENESIS_ADDRESS), common.BigToHash(hu.Add(baseSlot, big.NewInt(1)))).Big(),
		Decimals: uint8(stateDB.GetState(common.HexToAddress(MARGIN_ACCOUNT_GENESIS_ADDRESS), common.BigToHash(hu.Add(baseSlot, big.NewInt(2)))).Big().Uint64()),
		Price:    getUnderlyingPrice_(stateDB, tokenAddress),
	}
}

func collateralStorageSlot() *big.Int {
	return new(big.Int).SetBytes(crypto.Keccak256(common.LeftPadBytes(big.NewInt(SUPPORTED_COLLATERAL_SLOT).Bytes(), 32)))
}
