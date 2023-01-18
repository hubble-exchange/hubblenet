package limitorders

import (
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
)

var maintenanceMargin = big.NewInt(1e5)

var spreadRatioThreshold = big.NewInt(20 * 1e4)

var BASE_PRECISION = big.NewInt(1e18)

type Liquidable struct {
	Address        common.Address
	Size           *big.Int
	MarginFraction *big.Int
	FilledSize     *big.Int
}

func (liq Liquidable) GetUnfilledSize() *big.Int {
	return big.NewInt(0).Sub(liq.Size, liq.FilledSize)
}

func (db *InMemoryDatabase) GetLiquidableTraders(market Market, oraclePrice *big.Int) (longPositions []Liquidable, shortPositions []Liquidable) {
	longPositions = []Liquidable{}
	shortPositions = []Liquidable{}
	markPrice := db.lastPrice[market]

	overSpreadLimit := isOverSpreadLimit(markPrice, oraclePrice)

	for addr, trader := range db.GetAllTraders() {
		position := trader.Positions[market]
		notionalPosition := big.NewInt(0).Div(big.NewInt(0).Mul(position.Size, markPrice), BASE_PRECISION) // position.size * markPrice / BASE_PRECISION
		var unrealisedPnL *big.Int
		if position.Size.Sign() == 1 {
			unrealisedPnL.Sub(notionalPosition, position.OpenNotional)
		} else {
			unrealisedPnL.Sub(position.OpenNotional, notionalPosition)
		}

		margin := big.NewInt(0).Sub(trader.GetNormalisedMargin(), position.UnrealisedFunding)
		totalMargin := big.NewInt(0).Mul( big.NewInt(0).Add(margin, unrealisedPnL), big.NewInt(1e6))
		marginFraction := big.NewInt(0).Div(totalMargin, notionalPosition) // margin + unrealisedPnL / notionalPosition
		if overSpreadLimit {
			var oracleBasedUnrealizedPnl *big.Int

			oracleBasedNotional := big.NewInt(0).Div(big.NewInt(0).Mul(oraclePrice, position.Size.Abs(position.Size)), BASE_PRECISION)
			if position.Size.Sign() == 1 {
				oracleBasedUnrealizedPnl = big.NewInt(0).Sub(oracleBasedNotional, position.OpenNotional)
			} else if position.Size.Sign() == -1 {
				oracleBasedUnrealizedPnl = big.NewInt(0).Sub(position.OpenNotional, oracleBasedNotional)
			}

			oracleBasedmarginFraction := big.NewInt(0).Div(big.NewInt(0).Add(margin, oracleBasedUnrealizedPnl), oracleBasedNotional)
			if oracleBasedmarginFraction.Cmp(marginFraction) == 1 {
				marginFraction = oracleBasedmarginFraction
			}
		}

		if marginFraction.Cmp(maintenanceMargin) == -1 {
			liquidable := Liquidable{
				Address:        addr,
				Size:           position.Size,
				MarginFraction: marginFraction,
				FilledSize:     big.NewInt(0),
			}
			if position.Size.Sign() == -1 {
				shortPositions = append(shortPositions, liquidable)
			} else {
				longPositions = append(longPositions, liquidable)
			}
		}
	}

	// lower margin fraction positions should be liquidated first
	sort.Slice(longPositions, func(i, j int) bool {
		return longPositions[i].MarginFraction.Cmp(longPositions[j].MarginFraction) == -1
	})
	sort.Slice(shortPositions, func(i, j int) bool {
		return shortPositions[i].MarginFraction.Cmp(shortPositions[j].MarginFraction) == -1
	})
	return longPositions, shortPositions
}

func isOverSpreadLimit(markPrice *big.Int, oraclePrice *big.Int) bool {
	// diff := abs(markPrice - oraclePrice)
	diff := big.NewInt(0).Abs(big.NewInt(0).Sub(markPrice, oraclePrice))
	// spreadRatioAbs := diff * 100 / oraclePrice
	spreadRatioAbs := big.NewInt(0).Div(big.NewInt(0).Mul(diff, big.NewInt(1e6)), oraclePrice)
	if spreadRatioAbs.Cmp(spreadRatioThreshold) >= 0 {
		return true
	} else {
		return false
	}
}
