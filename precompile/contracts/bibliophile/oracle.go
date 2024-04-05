package bibliophile

import (
	"math/big"

	hu "github.com/ava-labs/subnet-evm/hubbleutils"
	"github.com/ava-labs/subnet-evm/precompile/contract"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

var (
	RED_STONE_VALUES_MAPPING_STORAGE_LOCATION  = common.HexToHash("0x4dd0c77efa6f6d590c97573d8c70b714546e7311202ff7c11c484cc841d91bfc") // keccak256("RedStone.oracleValuesMapping");
	RED_STONE_LATEST_ROUND_ID_STORAGE_LOCATION = common.HexToHash("0xc68d7f1ee07d8668991a8951e720010c9d44c2f11c06b5cac61fbc4083263938") // keccak256("RedStone.latestRoundId");

	AGGREGATOR_MAP_SLOT         int64 = 1
	RED_STONE_ADAPTER_SLOT      int64 = 2
	CUSTOM_ORACLE_ROUND_ID_SLOT int64 = 0
	CUSTOM_ORACLE_ENTRIES_SLOT  int64 = 1
)

const (
	// this slot is from TestOracle.sol
	TEST_ORACLE_PRICES_MAPPING_SLOT int64 = 3
)

func getUnderlyingPrice(stateDB contract.StateDB, market common.Address) *big.Int {
	return getUnderlyingPrice_(stateDB, getUnderlyingAssetAddress(stateDB, market))
}

func getUnderlyingPrice_(stateDB contract.StateDB, underlying common.Address) *big.Int {
	oracle := getOracleAddress(stateDB) // this comes from margin account

	// 1. Check for redstone feed id
	feedId := getRedStoneFeedId(stateDB, oracle, underlying)
	if feedId.Big().Sign() != 0 {
		// redstone oracle is configured for this market
		redStoneAdapter := getRedStoneAdapterAddress(stateDB, oracle)
		redstonePrice := getRedStonePrice(stateDB, redStoneAdapter, feedId)
		return redstonePrice
	}

	// 2. Check for custom oracle
	aggregator := getAggregatorAddress(stateDB, oracle, underlying)
	if aggregator.Big().Sign() != 0 {
		// custom oracle is configured for this market
		price := getCustomOraclePrice(stateDB, aggregator)
		log.Info("custom-oracle-price", "underlying", underlying, "price", price)
		return price
	}

	// 3. neither red stone nor custom oracle is enabled for this market, we use the default TestOracle
	slot := crypto.Keccak256(append(common.LeftPadBytes(underlying.Bytes(), 32), common.BigToHash(big.NewInt(TEST_ORACLE_PRICES_MAPPING_SLOT)).Bytes()...))
	return fromTwosComplement(stateDB.GetState(oracle, common.BytesToHash(slot)).Bytes())
}

func getMidPrice(stateDB contract.StateDB, market common.Address) *big.Int {
	asksHead := getAsksHead(stateDB, market)
	bidsHead := getBidsHead(stateDB, market)
	if asksHead.Sign() == 0 || bidsHead.Sign() == 0 {
		return getUnderlyingPrice(stateDB, market)
	}
	return hu.Div(hu.Add(asksHead, bidsHead), big.NewInt(2))
}

func getRedStoneAdapterAddress(stateDB contract.StateDB, oracle common.Address) common.Address {
	return common.BytesToAddress(stateDB.GetState(oracle, common.BigToHash(big.NewInt(RED_STONE_ADAPTER_SLOT))).Bytes())
}

func getRedStonePrice(stateDB contract.StateDB, adapterAddress common.Address, redStoneFeedId common.Hash) *big.Int {
	latestRoundId := getlatestRoundId(stateDB, adapterAddress)
	slot := common.BytesToHash(crypto.Keccak256(append(append(redStoneFeedId.Bytes(), common.LeftPadBytes(latestRoundId.Bytes(), 32)...), RED_STONE_VALUES_MAPPING_STORAGE_LOCATION.Bytes()...)))
	return new(big.Int).Div(fromTwosComplement(stateDB.GetState(adapterAddress, slot).Bytes()), big.NewInt(100)) // we use 6 decimals precision everywhere
}

func getlatestRoundId(stateDB contract.StateDB, adapterAddress common.Address) *big.Int {
	return fromTwosComplement(stateDB.GetState(adapterAddress, RED_STONE_LATEST_ROUND_ID_STORAGE_LOCATION).Bytes())
}

func aggregatorMapSlot(underlying common.Address) *big.Int {
	return new(big.Int).SetBytes(crypto.Keccak256(append(common.LeftPadBytes(underlying.Bytes(), 32), common.BigToHash(big.NewInt(AGGREGATOR_MAP_SLOT)).Bytes()...)))
}

func getRedStoneFeedId(stateDB contract.StateDB, oracle, underlying common.Address) common.Hash {
	aggregatorMapSlot := aggregatorMapSlot(underlying)
	return stateDB.GetState(oracle, common.BigToHash(aggregatorMapSlot))
}

func getAggregatorAddress(stateDB contract.StateDB, oracle, underlying common.Address) common.Address {
	aggregatorMapSlot := aggregatorMapSlot(underlying)
	aggregatorSlot := hu.Add(aggregatorMapSlot, big.NewInt(1))
	return common.BytesToAddress(stateDB.GetState(oracle, common.BigToHash(aggregatorSlot)).Bytes())
}

func getCustomOraclePrice(stateDB contract.StateDB, aggregator common.Address) *big.Int {
	roundId := stateDB.GetState(aggregator, common.BigToHash(big.NewInt(CUSTOM_ORACLE_ROUND_ID_SLOT))).Bytes()
	entriesSlot := new(big.Int).SetBytes(crypto.Keccak256(append(common.LeftPadBytes(roundId, 32), common.BigToHash(big.NewInt(CUSTOM_ORACLE_ENTRIES_SLOT)).Bytes()...)))
	priceSlot := hu.Add(entriesSlot, big.NewInt(1))
	return hu.Div(fromTwosComplement(stateDB.GetState(aggregator, common.BigToHash(priceSlot)).Bytes()), big.NewInt(100)) // we use 6 decimals precision everywhere
}
