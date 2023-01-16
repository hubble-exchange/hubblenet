package limitorders

import (
	"math"
	"sort"

	"github.com/ethereum/go-ethereum/common"
)

const maintenanceMargin float64 = 0.1
const spreadRatioThreshold = 20

type Liquidable struct {
	Address        common.Address
	Size           float64
	MarginFraction float64
}

func (db *InMemoryDatabase) GetLiquidableTraders(market Market, markPrice float64, oraclePrice float64) []Liquidable {

	toLiquidate := []Liquidable{}

	overSpreadLimit := isOverSpreadLimit(markPrice, oraclePrice)

	for addr, trader := range db.GetAllTraders() {
		position := trader.Positions[market]
		notionalPosition := position.Size * markPrice
		var unrealisedPnL float64
		if position.Size > 0 {
			unrealisedPnL = notionalPosition - position.OpenNotional
		} else {
			unrealisedPnL = position.OpenNotional - notionalPosition
		}

		margin := trader.GetNormalisedMargin() - position.UnrealisedFunding
		marginFraction := (margin + unrealisedPnL) / notionalPosition
		if overSpreadLimit {
			var oracleBasedUnrealizedPnl float64

			oracleBasedNotional := oraclePrice * math.Abs(position.Size)
			if position.Size > 0 {
				oracleBasedUnrealizedPnl = oracleBasedNotional - position.OpenNotional
			} else if position.Size < 0 {
				oracleBasedUnrealizedPnl = position.OpenNotional - oracleBasedNotional
			}

			oracleBasedmarginFraction := (margin + oracleBasedUnrealizedPnl) / oracleBasedNotional
			marginFraction = math.Max(marginFraction, oracleBasedmarginFraction)
		}

		if marginFraction < maintenanceMargin {
			toLiquidate = append(toLiquidate, Liquidable{
				Address:        addr,
				Size:           position.Size,
				MarginFraction: marginFraction,
			})
		}
	}

	// lower margin fraction positions should be liquidated first
	sort.Slice(toLiquidate, func(i, j int) bool {
		return toLiquidate[i].MarginFraction < toLiquidate[j].MarginFraction
	})
	return toLiquidate
}

func isOverSpreadLimit(markPrice float64, oraclePrice float64) bool {
	diff := math.Abs(markPrice - oraclePrice)
	spreadRatioAbs := diff * 100 / oraclePrice
	return spreadRatioAbs >= spreadRatioThreshold
}
