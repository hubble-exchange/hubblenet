package hubbleutils

import (
	"encoding/json"
	"math/big"
)

type Collateral struct {
	Price    *big.Int // scaled by 1e6
	Weight   *big.Int // scaled by 1e6
	Decimals uint8
}

type Market = int

type Position struct {
	OpenNotional         *big.Int `json:"open_notional"`
	Size                 *big.Int `json:"size"`
	UnrealisedFunding    *big.Int `json:"unrealised_funding"`
	LastPremiumFraction  *big.Int `json:"last_premium_fraction"`
	LiquidationThreshold *big.Int `json:"liquidation_threshold"`
}

func (p *Position) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		OpenNotional         string `json:"open_notional"`
		Size                 string `json:"size"`
		UnrealisedFunding    string `json:"unrealised_funding"`
		LastPremiumFraction  string `json:"last_premium_fraction"`
		LiquidationThreshold string `json:"liquidation_threshold"`
	}{
		OpenNotional:         p.OpenNotional.String(),
		Size:                 p.Size.String(),
		UnrealisedFunding:    p.UnrealisedFunding.String(),
		LastPremiumFraction:  p.LastPremiumFraction.String(),
		LiquidationThreshold: p.LiquidationThreshold.String(),
	})
}

type Trader struct {
	Positions map[Market]*Position `json:"positions"` // position for every market
	Margin    Margin               `json:"margin"`    // available margin/balance for every market
}

type Margin struct {
	Reserved  *big.Int                `json:"reserved"`
	Deposited map[Collateral]*big.Int `json:"deposited"`
}

func GetNormalizedMargin(assets []Collateral, margins []*big.Int) *big.Int {
	weighted, _ := WeightedAndSpotCollateral(assets, margins)
	return weighted
}

func WeightedAndSpotCollateral(assets []Collateral, margins []*big.Int) (weighted, spot *big.Int) {
	weighted = big.NewInt(0)
	spot = big.NewInt(0)
	for i, asset := range assets {
		if margins[i].Sign() == 0 {
			continue
		}
		numerator := Mul(margins[i], asset.Price) // margin[i] is scaled by asset.Decimal
		spot.Add(spot, Unscale(numerator, asset.Decimals))
		weighted.Add(weighted, Unscale(Mul(numerator, asset.Weight), asset.Decimals+6))
	}
	return weighted, spot
}

func GetAvailableMargin(positions map[Market]*Position, margins []*big.Int, pendingFunding *big.Int, reservedMargin *big.Int, assets []Collateral, oraclePrices map[Market]*big.Int, lastPrices map[Market]*big.Int, minAllowableMargin *big.Int, markets []Market) *big.Int {
	notionalPosition, margin := GetNotionalPositionAndMargin(positions, margins, Min_Allowable_Margin, pendingFunding, assets, oraclePrices, lastPrices, markets)
	return GetAvailableMargin_(notionalPosition, margin, reservedMargin, minAllowableMargin)
}

func GetAvailableMargin_(notionalPosition, margin, reservedMargin, minAllowableMargin *big.Int) *big.Int {
	utilisedMargin := Div1e6(Mul(notionalPosition, minAllowableMargin))
	return Sub(margin, Add(utilisedMargin, reservedMargin))
}

type MarginMode = uint8

const (
	Maintenance_Margin MarginMode = iota
	Min_Allowable_Margin
)

func GetNotionalPositionAndMargin(positions map[Market]*Position, margins []*big.Int, marginMode MarginMode, pendingFunding *big.Int, assets []Collateral, oraclePrices map[Market]*big.Int, lastPrices map[Market]*big.Int, markets []Market) (*big.Int, *big.Int) {
	margin := Sub(GetNormalizedMargin(assets, margins), pendingFunding)
	notionalPosition, unrealizedPnl := GetTotalNotionalPositionAndUnrealizedPnl(positions, margin, marginMode, oraclePrices, lastPrices, markets)
	return notionalPosition, Add(margin, unrealizedPnl)
}

func GetTotalNotionalPositionAndUnrealizedPnl(positions map[Market]*Position, margin *big.Int, marginMode MarginMode, oraclePrices map[Market]*big.Int, lastPrices map[Market]*big.Int, markets []Market) (*big.Int, *big.Int) {
	notionalPosition := big.NewInt(0)
	unrealizedPnl := big.NewInt(0)
	for _, market := range markets {
		_notionalPosition, _unrealizedPnl := GetOptimalPnl(market, oraclePrices[market], lastPrices[market], positions, margin, marginMode)
		notionalPosition.Add(notionalPosition, _notionalPosition)
		unrealizedPnl.Add(unrealizedPnl, _unrealizedPnl)
	}
	return notionalPosition, unrealizedPnl
}

func GetOptimalPnl(market Market, oraclePrice *big.Int, lastPrice *big.Int, positions map[Market]*Position, margin *big.Int, marginMode MarginMode) (notionalPosition *big.Int, uPnL *big.Int) {
	position := positions[market]
	if position == nil || position.Size.Sign() == 0 {
		return big.NewInt(0), big.NewInt(0)
	}

	// based on last price
	notionalPosition, unrealizedPnl, lastPriceBasedMF := GetPositionMetadata(
		lastPrice,
		position.OpenNotional,
		position.Size,
		margin,
	)
	// log.Info("in getOptimalPnl", "notionalPosition", notionalPosition, "unrealizedPnl", unrealizedPnl, "lastPriceBasedMF", lastPriceBasedMF)

	// based on oracle price
	oracleBasedNotional, oracleBasedUnrealizedPnl, oracleBasedMF := GetPositionMetadata(
		oraclePrice,
		position.OpenNotional,
		position.Size,
		margin,
	)
	// log.Info("in getOptimalPnl", "oracleBasedNotional", oracleBasedNotional, "oracleBasedUnrealizedPnl", oracleBasedUnrealizedPnl, "oracleBasedMF", oracleBasedMF)

	if (marginMode == Maintenance_Margin && oracleBasedMF.Cmp(lastPriceBasedMF) == 1) || // for liquidations
		(marginMode == Min_Allowable_Margin && oracleBasedMF.Cmp(lastPriceBasedMF) == -1) { // for increasing leverage
		return oracleBasedNotional, oracleBasedUnrealizedPnl
	}
	return notionalPosition, unrealizedPnl
}

func GetPositionMetadata(price *big.Int, openNotional *big.Int, size *big.Int, margin *big.Int) (notionalPosition *big.Int, unrealisedPnl *big.Int, marginFraction *big.Int) {
	// log.Info("in GetPositionMetadata", "price", price, "openNotional", openNotional, "size", size, "margin", margin)
	notionalPosition = getNotionalPosition(price, size)
	uPnL := new(big.Int)
	if notionalPosition.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0), big.NewInt(0), big.NewInt(0)
	}
	if size.Cmp(big.NewInt(0)) > 0 {
		uPnL = new(big.Int).Sub(notionalPosition, openNotional)
	} else {
		uPnL = new(big.Int).Sub(openNotional, notionalPosition)
	}
	mf := new(big.Int).Div(Mul1e6(new(big.Int).Add(margin, uPnL)), notionalPosition)
	return notionalPosition, uPnL, mf
}

func getNotionalPosition(price *big.Int, size *big.Int) *big.Int {
	return big.NewInt(0).Abs(Div1e18(Mul(size, price)))
}
