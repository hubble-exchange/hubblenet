package hubblebibliophile

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/contract"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	VAR_LAST_PRICE_SLOT int64 = 0
	VAR_POSITIONS_SLOT  int64 = 1 // tbd
)

// Reader

func getLastPrice(stateDB contract.StateDB, amm *common.Address) *big.Int {
	return stateDB.GetState(*amm, common.BigToHash(big.NewInt(VAR_LAST_PRICE_SLOT))).Big()
}

func getSize(stateDB contract.StateDB, amm *common.Address, trader *common.Address) *big.Int {
	return stateDB.GetState(*amm, common.BigToHash(positionsStorageSlot(trader))).Big()
}

func getOpenNotional(stateDB contract.StateDB, amm *common.Address, trader *common.Address) *big.Int {
	return stateDB.GetState(*amm, common.BigToHash(new(big.Int).Add(positionsStorageSlot(trader), big.NewInt(1)))).Big()
}

func positionsStorageSlot(trader *common.Address) *big.Int {
	return new(big.Int).SetBytes(crypto.Keccak256(append(common.LeftPadBytes(trader.Bytes(), 32), common.LeftPadBytes(big.NewInt(VAR_POSITIONS_SLOT).Bytes(), 32)...)))
}

// utilities

type MarginMode uint8

const (
	Maintenance_Margin MarginMode = iota
	Min_Allowable_Margin
)

func GetMarginMode(marginMode uint8) MarginMode {
	if marginMode == 0 {
		return Maintenance_Margin
	}
	return Min_Allowable_Margin
}

func GetTotalNotionalPositionAndUnrealizedPnl(stateDB contract.StateDB, trader *common.Address, margin *big.Int, marginMode MarginMode) (*big.Int, *big.Int) {
	notionalPosition := big.NewInt(0)
	unrealizedPnl := big.NewInt(0)
	markets := []*common.Address{}
	for _, market := range markets {
		lastPrice := getLastPrice(stateDB, market)
		// oraclePrice := getUnderlyingPrice(stateDB, market)
		oraclePrice := multiply1e6(big.NewInt(1800))
		_notionalPosition, _unrealizedPnl := getOptimalPnl(stateDB, market, oraclePrice, lastPrice, trader, margin, marginMode)
		notionalPosition.Add(notionalPosition, _notionalPosition)
		unrealizedPnl.Add(unrealizedPnl, _unrealizedPnl)
	}
	return notionalPosition, unrealizedPnl
}

func getOptimalPnl(stateDB contract.StateDB, market *common.Address, oraclePrice *big.Int, lastPrice *big.Int, trader *common.Address, margin *big.Int, marginMode MarginMode) (notionalPosition *big.Int, uPnL *big.Int) {
	size := getSize(stateDB, market, trader)
	if size.Sign() == 0 {
		return big.NewInt(0), big.NewInt(0)
	}

	openNotional := getOpenNotional(stateDB, market, trader)
	// based on last price
	notionalPosition, unrealizedPnl, lastPriceBasedMF := getPositionMetadata(
		lastPrice,
		openNotional,
		size,
		margin,
	)

	// based on oracle price
	oracleBasedNotional, oracleBasedUnrealizedPnl, oracleBasedMF := getPositionMetadata(
		oraclePrice,
		openNotional,
		size,
		margin,
	)

	if (marginMode == Maintenance_Margin && oracleBasedMF.Cmp(lastPriceBasedMF) == 1) || // for liquidations
		(marginMode == Min_Allowable_Margin && oracleBasedMF.Cmp(lastPriceBasedMF) == -1) { // for increasing leverage
		return oracleBasedNotional, oracleBasedUnrealizedPnl
	}
	return notionalPosition, unrealizedPnl
}

func getPositionMetadata(price *big.Int, openNotional *big.Int, size *big.Int, margin *big.Int) (notionalPos *big.Int, uPnl *big.Int, marginFraction *big.Int) {
	notionalPos = divide1e18(new(big.Int).Mul(price, new(big.Int).Abs(size)))
	if notionalPos.Sign() == 0 {
		return big.NewInt(0), big.NewInt(0), big.NewInt(0)
	}
	if size.Sign() == 1 {
		uPnl = new(big.Int).Sub(notionalPos, openNotional)
	} else {
		uPnl = new(big.Int).Sub(openNotional, notionalPos)
	}
	marginFraction = new(big.Int).Div(multiply1e6(new(big.Int).Add(margin, uPnl)), notionalPos)
	return notionalPos, uPnl, marginFraction
}

func divide1e18(number *big.Int) *big.Int {
	return big.NewInt(0).Div(number, big.NewInt(1e18))
}

func divide1e6(number *big.Int) *big.Int {
	return big.NewInt(0).Div(number, big.NewInt(1e6))
}

func multiply1e6(number *big.Int) *big.Int {
	return new(big.Int).Div(number, big.NewInt(1e6))
}
