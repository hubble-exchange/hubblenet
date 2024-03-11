package hubbleutils

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _hState = &HubbleState{
	Assets: []Collateral{
		{
			Price:    big.NewInt(1.01 * 1e6), // 1.01
			Weight:   big.NewInt(1e6),        // 1
			Decimals: 6,
		},
		{
			Price:    big.NewInt(54.36 * 1e6), // 54.36
			Weight:   big.NewInt(0.7 * 1e6),   // 0.7
			Decimals: 6,
		},
	},
	MidPrices: map[Market]*big.Int{
		0: big.NewInt(1544.21 * 1e6), // 1544.21
		1: big.NewInt(19.5 * 1e6),    // 19.5
	},
	OraclePrices: map[Market]*big.Int{
		0: big.NewInt(1503.21 * 1e6),
		1: big.NewInt(17.5 * 1e6),
	},
	ActiveMarkets: []Market{
		0, 1,
	},
	MinAllowableMargin: big.NewInt(100000), // 0.1
	MaintenanceMargin:  big.NewInt(200000), // 0.2
	UpgradeVersion:     V1,
}

var userState = &UserState{
	Positions: map[Market]*Position{
		0: {
			Size:         big.NewInt(0.582 * 1e18), // 0.582
			OpenNotional: big.NewInt(875 * 1e6),    // 875, openPrice = 1503.43
		},
		1: {
			Size:         Scale(big.NewInt(-101), 18), // -101
			OpenNotional: big.NewInt(1767.5 * 1e6),    // 1767.5, openPrice = 17.5
		},
	},
	Margins: []*big.Int{
		big.NewInt(30.5 * 1e6), // 30.5
		big.NewInt(14 * 1e6),   // 14
	},
	PendingFunding: big.NewInt(-50 * 1e6), // +50
	ReservedMargin: big.NewInt(60 * 1e6),  // 60
	AccountPreferences: map[Market]*AccountPreferences{
		0: {
			MarginType:     Cross_Margin,
			MarginFraction: big.NewInt(0.2 * 1e6), // 0.2
		},
		1: {
			MarginType:     Isolated_Margin,
			MarginFraction: big.NewInt(0.1 * 1e6), // 0.1
		},
	},
}

func TestWeightedAndSpotCollateral(t *testing.T) {
	assets := _hState.Assets
	margins := userState.Margins
	expectedWeighted := Unscale(Mul(Mul(margins[0], assets[0].Price), assets[0].Weight), assets[0].Decimals+6)
	expectedWeighted.Add(expectedWeighted, Unscale(Mul(Mul(margins[1], assets[1].Price), assets[1].Weight), assets[1].Decimals+6))

	expectedSpot := Unscale(Mul(margins[0], assets[0].Price), assets[0].Decimals)
	expectedSpot.Add(expectedSpot, Unscale(Mul(margins[1], assets[1].Price), assets[1].Decimals))

	resultWeighted, resultSpot := WeightedAndSpotCollateral(assets, margins)
	fmt.Println(resultWeighted, resultSpot)
	assert.Equal(t, expectedWeighted, resultWeighted)
	assert.Equal(t, expectedSpot, resultSpot)

	normalisedMargin := GetNormalizedMargin(assets, margins)
	assert.Equal(t, expectedWeighted, normalisedMargin)

}

func TestGetNotionalPosition(t *testing.T) {
	price := Scale(big.NewInt(1200), 6)
	size := Scale(big.NewInt(5), 18)
	expected := Scale(big.NewInt(6000), 6)

	result := GetNotionalPosition(price, size)

	assert.Equal(t, expected, result)
}

func TestGetPositionMetadata(t *testing.T) {
	price := big.NewInt(20250000)        // 20.25
	openNotional := big.NewInt(75369000) // 75.369 (size * 18.5)
	size := Scale(big.NewInt(40740), 14) // 4.074
	margin := big.NewInt(20000000)       // 20

	notionalPosition, unrealisedPnl, marginFraction := GetPositionMetadata(price, openNotional, size, margin)

	expectedNotionalPosition := big.NewInt(82498500) // 82.4985
	expectedUnrealisedPnl := big.NewInt(7129500)     // 7.1295
	expectedMarginFraction := big.NewInt(328848)     // 0.328848

	assert.Equal(t, expectedNotionalPosition, notionalPosition)
	assert.Equal(t, expectedUnrealisedPnl, unrealisedPnl)
	assert.Equal(t, expectedMarginFraction, marginFraction)

	// ------ when size is negative ------
	size = Scale(big.NewInt(-40740), 14) // -4.074
	openNotional = big.NewInt(75369000)  // 75.369 (size * 18.5)
	notionalPosition, unrealisedPnl, marginFraction = GetPositionMetadata(price, openNotional, size, margin)
	fmt.Println("notionalPosition", notionalPosition, "unrealisedPnl", unrealisedPnl, "marginFraction", marginFraction)

	expectedNotionalPosition = big.NewInt(82498500) // 82.4985
	expectedUnrealisedPnl = big.NewInt(-7129500)    // -7.1295
	expectedMarginFraction = big.NewInt(156008)     // 0.156008

	assert.Equal(t, expectedNotionalPosition, notionalPosition)
	assert.Equal(t, expectedUnrealisedPnl, unrealisedPnl)
	assert.Equal(t, expectedMarginFraction, marginFraction)
}

func TestGetOptimalPnlV2(t *testing.T) {
	margin := big.NewInt(20 * 1e6) // 20
	market := 0
	position := userState.Positions[market]
	marginMode := Maintenance_Margin

	notionalPosition, uPnL := getOptimalPnl(_hState, position, margin, market, marginMode)

	// mid price pnl is more than oracle price pnl
	expectedNotionalPosition := Unscale(Mul(position.Size, _hState.MidPrices[market]), 18)
	expectedUPnL := Sub(expectedNotionalPosition, position.OpenNotional)
	fmt.Println("Maintenace_Margin_Mode", "notionalPosition", notionalPosition, "uPnL", uPnL)

	assert.Equal(t, expectedNotionalPosition, notionalPosition)
	assert.Equal(t, expectedUPnL, uPnL)

	// ------ when marginMode is Min_Allowable_Margin ------

	marginMode = Min_Allowable_Margin
	notionalPosition, uPnL = getOptimalPnl(_hState, position, margin, market, marginMode)

	expectedNotionalPosition = Unscale(Mul(position.Size, _hState.OraclePrices[market]), 18)
	expectedUPnL = Sub(expectedNotionalPosition, position.OpenNotional)
	fmt.Println("Min_Allowable_Margin_Mode", "notionalPosition", notionalPosition, "uPnL", uPnL)

	assert.Equal(t, expectedNotionalPosition, notionalPosition)
	assert.Equal(t, expectedUPnL, uPnL)
}

func TestGetOptimalPnlV1(t *testing.T) {
	margin := big.NewInt(20 * 1e6) // 20
	market := 0
	position := userState.Positions[market]
	marginMode := Maintenance_Margin

	notionalPosition, uPnL := getOptimalPnl(_hState, position, margin, market, marginMode)

	// mid price pnl is more than oracle price pnl
	expectedNotionalPosition := Unscale(Mul(position.Size, _hState.MidPrices[market]), 18)
	expectedUPnL := Sub(expectedNotionalPosition, position.OpenNotional)
	fmt.Println("Maintenace_Margin_Mode", "notionalPosition", notionalPosition, "uPnL", uPnL)

	assert.Equal(t, expectedNotionalPosition, notionalPosition)
	assert.Equal(t, expectedUPnL, uPnL)

	// ------ when marginMode is Min_Allowable_Margin ------

	marginMode = Min_Allowable_Margin
	notionalPosition, uPnL = getOptimalPnl(_hState, position, margin, market, marginMode)

	expectedNotionalPosition = Unscale(Mul(position.Size, _hState.OraclePrices[market]), 18)
	expectedUPnL = Sub(expectedNotionalPosition, position.OpenNotional)
	fmt.Println("Min_Allowable_Margin_Mode", "notionalPosition", notionalPosition, "uPnL", uPnL)

	assert.Equal(t, expectedNotionalPosition, notionalPosition)
	assert.Equal(t, expectedUPnL, uPnL)
}

func TestGetTotalNotionalPositionAndUnrealizedPnlV2(t *testing.T) {
	margin := GetNormalizedMargin(_hState.Assets, userState.Margins)
	marginMode := Maintenance_Margin
	notionalPosition, uPnL := GetTotalNotionalPositionAndUnrealizedPnl(_hState, userState, margin, marginMode)

	// mid price pnl is more than oracle price pnl for long position
	expectedNotionalPosition := Unscale(Mul(userState.Positions[0].Size, _hState.MidPrices[0]), 18)
	expectedUPnL := Sub(expectedNotionalPosition, userState.Positions[0].OpenNotional)
	// oracle price pnl is more than mid price pnl for short position
	expectedNotional2 := Abs(Unscale(Mul(userState.Positions[1].Size, _hState.OraclePrices[1]), 18))
	expectedNotionalPosition.Add(expectedNotionalPosition, expectedNotional2)
	expectedUPnL.Add(expectedUPnL, Sub(userState.Positions[1].OpenNotional, expectedNotional2))

	assert.Equal(t, expectedNotionalPosition, notionalPosition)
	assert.Equal(t, expectedUPnL, uPnL)

	// ------ when marginMode is Min_Allowable_Margin ------

	marginMode = Min_Allowable_Margin
	notionalPosition, uPnL = GetTotalNotionalPositionAndUnrealizedPnl(_hState, userState, margin, marginMode)
	fmt.Println("Min_Allowable_Margin_Mode ", "notionalPosition = ", notionalPosition, "uPnL = ", uPnL)

	expectedNotionalPosition = Unscale(Mul(userState.Positions[0].Size, _hState.OraclePrices[0]), 18)
	expectedUPnL = Sub(expectedNotionalPosition, userState.Positions[0].OpenNotional)
	expectedNotional2 = Abs(Unscale(Mul(userState.Positions[1].Size, _hState.MidPrices[1]), 18))
	expectedNotionalPosition.Add(expectedNotionalPosition, expectedNotional2)
	expectedUPnL.Add(expectedUPnL, Sub(userState.Positions[1].OpenNotional, expectedNotional2))

	assert.Equal(t, expectedNotionalPosition, notionalPosition)
	assert.Equal(t, expectedUPnL, uPnL)
}

func TestGetTotalNotionalPositionAndUnrealizedPnl(t *testing.T) {
	margin := GetNormalizedMargin(_hState.Assets, userState.Margins)
	marginMode := Maintenance_Margin
	_hState.UpgradeVersion = V2
	notionalPosition, uPnL := GetTotalNotionalPositionAndUnrealizedPnl(_hState, userState, margin, marginMode)

	// mid price pnl is more than oracle price pnl for long position
	expectedNotionalPosition := Unscale(Mul(userState.Positions[0].Size, _hState.OraclePrices[0]), 18)
	expectedUPnL := Sub(expectedNotionalPosition, userState.Positions[0].OpenNotional)
	// oracle price pnl is more than mid price pnl for short position
	expectedNotional2 := Abs(Unscale(Mul(userState.Positions[1].Size, _hState.OraclePrices[1]), 18))
	expectedNotionalPosition.Add(expectedNotionalPosition, expectedNotional2)
	expectedUPnL.Add(expectedUPnL, Sub(userState.Positions[1].OpenNotional, expectedNotional2))

	assert.Equal(t, expectedNotionalPosition, notionalPosition)
	assert.Equal(t, expectedUPnL, uPnL)

	// ------ when marginMode is Min_Allowable_Margin ------

	marginMode = Min_Allowable_Margin
	notionalPosition, uPnL = GetTotalNotionalPositionAndUnrealizedPnl(_hState, userState, margin, marginMode)
	fmt.Println("Min_Allowable_Margin_Mode ", "notionalPosition = ", notionalPosition, "uPnL = ", uPnL)

	expectedNotionalPosition = Unscale(Mul(userState.Positions[0].Size, _hState.OraclePrices[0]), 18)
	expectedUPnL = Sub(expectedNotionalPosition, userState.Positions[0].OpenNotional)
	expectedNotional2 = Abs(Unscale(Mul(userState.Positions[1].Size, _hState.OraclePrices[1]), 18))
	expectedNotionalPosition.Add(expectedNotionalPosition, expectedNotional2)
	expectedUPnL.Add(expectedUPnL, Sub(userState.Positions[1].OpenNotional, expectedNotional2))

	assert.Equal(t, expectedNotionalPosition, notionalPosition)
	assert.Equal(t, expectedUPnL, uPnL)
}

func TestGetNotionalPositionAndRequiredMargin(t *testing.T) {
	t.Run("one market in cross and other in isolated mode", func(t *testing.T) {
		expectedMargin := GetNormalizedMargin(_hState.Assets, userState.Margins)
		fmt.Println(expectedMargin)
		notionalPosition, margin, requiredMargin := GetNotionalPositionAndRequiredMargin(_hState, userState)
		expectedNotionalPosition := Unscale(Mul(userState.Positions[0].Size, _hState.OraclePrices[0]), 18)
		expectedUPnL := Sub(expectedNotionalPosition, userState.Positions[0].OpenNotional)
		expectedMargin = Sub(Add(expectedMargin, expectedUPnL), userState.PendingFunding)
		expectedRequiredMargin := Unscale(Mul(expectedNotionalPosition, userState.AccountPreferences[0].MarginFraction), 6)
		assert.Equal(t, expectedNotionalPosition, notionalPosition)
		assert.Equal(t, expectedMargin, margin)
		assert.Equal(t, expectedRequiredMargin, requiredMargin)
	})

	t.Run("both markets in cross mode", func(t *testing.T) {
		userState.AccountPreferences[1].MarginType = Cross_Margin
		notionalPosition, margin, requiredMargin := GetNotionalPositionAndRequiredMargin(_hState, userState)
		expectedNotionalPosition := big.NewInt(0)
		expectedRequiredMargin := big.NewInt(0)
		expectedMargin := GetNormalizedMargin(_hState.Assets, userState.Margins)
		for _, market := range _hState.ActiveMarkets {
			notional := Abs(Unscale(Mul(userState.Positions[market].Size, _hState.OraclePrices[market]), 18))
			expectedNotionalPosition = Add(expectedNotionalPosition, notional)
			multiplier := big.NewInt(1)
			if userState.Positions[market].Size.Sign() == -1 {
				multiplier = big.NewInt(-1)
			}
			expectedUPnL := Mul(Sub(notional, userState.Positions[market].OpenNotional), multiplier)
			expectedMargin = Add(expectedMargin, expectedUPnL)
			expectedRequiredMargin = Add(expectedRequiredMargin, Unscale(Mul(notional, userState.AccountPreferences[market].MarginFraction), 6))
		}
		expectedMargin = Sub(expectedMargin, userState.PendingFunding)
		assert.Equal(t, expectedNotionalPosition, notionalPosition)
		assert.Equal(t, expectedMargin, margin)
		assert.Equal(t, expectedRequiredMargin, requiredMargin)
	})
}
