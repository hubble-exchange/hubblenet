package traderviewer

import (
	"math/big"

	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	hu "github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
	b "github.com/ava-labs/subnet-evm/precompile/contracts/bibliophile"
	juror "github.com/ava-labs/subnet-evm/precompile/contracts/juror"
	gomock "github.com/golang/mock/gomock"
)

func TestValidateCancelLimitOrderV2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBibliophile := b.NewMockBibliophileClient(ctrl)
	ammIndex := big.NewInt(0)
	longBaseAssetQuantity := big.NewInt(5000000000000000000)
	shortBaseAssetQuantity := big.NewInt(-5000000000000000000)
	price := big.NewInt(100000000)
	salt := big.NewInt(121)
	reduceOnly := false
	postOnly := false
	trader := common.HexToAddress("0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC")
	ammAddress := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	assertLowMargin := false
	assertOverPositionCap := false

	t.Run("when sender is not the trader and is not trading authority, it returns error", func(t *testing.T) {
		sender := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C9")
		t.Run("it returns error for a long order", func(t *testing.T) {
			order := getOrder(ammIndex, trader, longBaseAssetQuantity, price, salt, reduceOnly, postOnly)
			input := getValidateCancelLimitOrderV2Input(order, sender, assertLowMargin, assertOverPositionCap)
			mockBibliophile.EXPECT().IsTradingAuthority(order.Trader, sender).Return(false).Times(1)
			output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
			assert.Equal(t, ErrUnauthorizedCancellation.Error(), output.Err)
		})
		t.Run("it returns error for a short order", func(t *testing.T) {
			order := getOrder(ammIndex, trader, shortBaseAssetQuantity, price, salt, reduceOnly, postOnly)
			input := getValidateCancelLimitOrderV2Input(order, sender, assertLowMargin, assertOverPositionCap)
			mockBibliophile.EXPECT().IsTradingAuthority(order.Trader, sender).Return(false).Times(1)
			output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
			assert.Equal(t, ErrUnauthorizedCancellation.Error(), output.Err)
		})
	})
	t.Run("when either sender is trader or a trading authority", func(t *testing.T) {
		t.Run("When order status is not placed", func(t *testing.T) {
			t.Run("when order status was never placed", func(t *testing.T) {
				t.Run("it returns error for a longOrder", func(t *testing.T) {
					longOrder := getOrder(ammIndex, trader, longBaseAssetQuantity, price, salt, reduceOnly, postOnly)
					orderHash := getOrderHash(longOrder)
					mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Invalid)).Times(1)
					input := getValidateCancelLimitOrderV2Input(longOrder, trader, assertLowMargin, assertOverPositionCap)
					output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
					assert.Equal(t, "Invalid", output.Err)
					assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
					assert.Equal(t, common.Address{}, output.Res.Amm)
					assert.Equal(t, big.NewInt(0), output.Res.UnfilledAmount)
				})
				t.Run("it returns error for a shortOrder", func(t *testing.T) {
					shortOrder := getOrder(ammIndex, trader, shortBaseAssetQuantity, price, salt, reduceOnly, postOnly)
					orderHash := getOrderHash(shortOrder)
					mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Invalid)).Times(1)
					input := getValidateCancelLimitOrderV2Input(shortOrder, trader, assertLowMargin, assertOverPositionCap)
					output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
					assert.Equal(t, "Invalid", output.Err)
					assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
					assert.Equal(t, common.Address{}, output.Res.Amm)
					assert.Equal(t, big.NewInt(0), output.Res.UnfilledAmount)
				})
			})
			t.Run("when order status is cancelled", func(t *testing.T) {
				t.Run("it returns error for a longOrder", func(t *testing.T) {
					longOrder := getOrder(ammIndex, trader, longBaseAssetQuantity, price, salt, reduceOnly, postOnly)
					orderHash := getOrderHash(longOrder)
					mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Cancelled)).Times(1)
					input := getValidateCancelLimitOrderV2Input(longOrder, trader, assertLowMargin, assertOverPositionCap)
					output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
					assert.Equal(t, "Cancelled", output.Err)
					assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
					assert.Equal(t, common.Address{}, output.Res.Amm)
					assert.Equal(t, big.NewInt(0), output.Res.UnfilledAmount)
				})
				t.Run("it returns error for a shortOrder", func(t *testing.T) {
					shortOrder := getOrder(ammIndex, trader, shortBaseAssetQuantity, price, salt, reduceOnly, postOnly)
					orderHash := getOrderHash(shortOrder)
					mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Cancelled)).Times(1)
					input := getValidateCancelLimitOrderV2Input(shortOrder, trader, assertLowMargin, assertOverPositionCap)
					output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
					assert.Equal(t, "Cancelled", output.Err)
					assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
					assert.Equal(t, common.Address{}, output.Res.Amm)
					assert.Equal(t, big.NewInt(0), output.Res.UnfilledAmount)
				})
			})
			t.Run("when order status is filled", func(t *testing.T) {
				t.Run("it returns error for a longOrder", func(t *testing.T) {
					longOrder := getOrder(ammIndex, trader, longBaseAssetQuantity, price, salt, reduceOnly, postOnly)
					orderHash := getOrderHash(longOrder)
					mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Filled)).Times(1)
					input := getValidateCancelLimitOrderV2Input(longOrder, trader, assertLowMargin, assertOverPositionCap)
					output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
					assert.Equal(t, "Filled", output.Err)
					assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
					assert.Equal(t, common.Address{}, output.Res.Amm)
					assert.Equal(t, big.NewInt(0), output.Res.UnfilledAmount)
				})
				t.Run("it returns error for a shortOrder", func(t *testing.T) {
					shortOrder := getOrder(ammIndex, trader, shortBaseAssetQuantity, price, salt, reduceOnly, postOnly)
					orderHash := getOrderHash(shortOrder)
					mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Filled)).Times(1)
					input := getValidateCancelLimitOrderV2Input(shortOrder, trader, assertLowMargin, assertOverPositionCap)
					output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
					assert.Equal(t, "Filled", output.Err)
					assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
					assert.Equal(t, common.Address{}, output.Res.Amm)
					assert.Equal(t, big.NewInt(0), output.Res.UnfilledAmount)
				})
			})
		})
		t.Run("When order status is placed", func(t *testing.T) {
			t.Run("when assertLowMargin is true", func(t *testing.T) {
				assertLowMargin := true
				t.Run("when availableMargin >= zero", func(t *testing.T) {
					t.Run("when availableMargin == 0 ", func(t *testing.T) {
						t.Run("it returns error for a longOrder", func(t *testing.T) {
							longOrder := getOrder(ammIndex, trader, longBaseAssetQuantity, price, salt, reduceOnly, postOnly)
							orderHash := getOrderHash(longOrder)

							mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Placed)).Times(1)
							mockBibliophile.EXPECT().GetTimeStamp().Return(hu.V1ActivationTime).Times(1)
							mockBibliophile.EXPECT().GetAvailableMargin(longOrder.Trader, hu.V1).Return(big.NewInt(0)).Times(1)
							mockBibliophile.EXPECT().IsValidator(longOrder.Trader).Return(true).Times(1)
							mockBibliophile.EXPECT().GetOrderFilledAmount(orderHash).Return(big.NewInt(0)).Times(1)
							input := getValidateCancelLimitOrderV2Input(longOrder, trader, assertLowMargin, assertOverPositionCap)
							output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
							assert.Equal(t, "Not Low Margin", output.Err)
							assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
							assert.Equal(t, common.Address{}, output.Res.Amm)
							assert.Equal(t, longOrder.BaseAssetQuantity, output.Res.UnfilledAmount)
						})
						t.Run("it returns error for a shortOrder", func(t *testing.T) {
							shortOrder := getOrder(ammIndex, trader, shortBaseAssetQuantity, price, salt, reduceOnly, postOnly)
							orderHash := getOrderHash(shortOrder)

							mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Placed)).Times(1)
							mockBibliophile.EXPECT().GetTimeStamp().Return(hu.V1ActivationTime).Times(1)
							mockBibliophile.EXPECT().GetAvailableMargin(shortOrder.Trader, hu.V1).Return(big.NewInt(0)).Times(1)
							mockBibliophile.EXPECT().IsValidator(shortOrder.Trader).Return(true).Times(1)
							mockBibliophile.EXPECT().GetOrderFilledAmount(orderHash).Return(big.NewInt(0)).Times(1)
							input := getValidateCancelLimitOrderV2Input(shortOrder, trader, assertLowMargin, assertOverPositionCap)
							output := ValidateCancelLimitOrderV2(mockBibliophile, &input)

							assert.Equal(t, "Not Low Margin", output.Err)
							assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
							assert.Equal(t, common.Address{}, output.Res.Amm)
							assert.Equal(t, shortOrder.BaseAssetQuantity, output.Res.UnfilledAmount)
						})
					})
					t.Run("when availableMargin > 0 ", func(t *testing.T) {
						newMargin := hu.Mul(price, longBaseAssetQuantity)
						t.Run("it returns error for a longOrder", func(t *testing.T) {
							longOrder := getOrder(ammIndex, trader, longBaseAssetQuantity, price, salt, reduceOnly, postOnly)
							orderHash := getOrderHash(longOrder)

							mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Placed)).Times(1)
							mockBibliophile.EXPECT().GetTimeStamp().Return(hu.V1ActivationTime).Times(1)
							mockBibliophile.EXPECT().GetAvailableMargin(longOrder.Trader, hu.V1).Return(newMargin).Times(1)
							mockBibliophile.EXPECT().IsValidator(longOrder.Trader).Return(true).Times(1)
							mockBibliophile.EXPECT().GetOrderFilledAmount(orderHash).Return(big.NewInt(0)).Times(1)
							input := getValidateCancelLimitOrderV2Input(longOrder, trader, assertLowMargin, assertOverPositionCap)
							output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
							assert.Equal(t, "Not Low Margin", output.Err)
							assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
							assert.Equal(t, common.Address{}, output.Res.Amm)
							assert.Equal(t, longOrder.BaseAssetQuantity, output.Res.UnfilledAmount)
						})
						t.Run("it returns error for a shortOrder", func(t *testing.T) {
							shortOrder := getOrder(ammIndex, trader, shortBaseAssetQuantity, price, salt, reduceOnly, postOnly)
							orderHash := getOrderHash(shortOrder)

							mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Placed)).Times(1)
							mockBibliophile.EXPECT().GetTimeStamp().Return(hu.V1ActivationTime).Times(1)
							mockBibliophile.EXPECT().GetAvailableMargin(shortOrder.Trader, hu.V1).Return(newMargin).Times(1)
							mockBibliophile.EXPECT().IsValidator(shortOrder.Trader).Return(true).Times(1)
							mockBibliophile.EXPECT().GetOrderFilledAmount(orderHash).Return(big.NewInt(0)).Times(1)
							input := getValidateCancelLimitOrderV2Input(shortOrder, trader, assertLowMargin, assertOverPositionCap)
							output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
							assert.Equal(t, "Not Low Margin", output.Err)
							assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
							assert.Equal(t, common.Address{}, output.Res.Amm)
							assert.Equal(t, shortOrder.BaseAssetQuantity, output.Res.UnfilledAmount)
						})
					})
				})
				t.Run("when availableMargin < zero", func(t *testing.T) {
					t.Run("for an unfilled Order", func(t *testing.T) {
						t.Run("for a longOrder it returns err = nil, with ammAddress and unfilled amount of cancelled Order", func(t *testing.T) {
							longOrder := getOrder(ammIndex, trader, longBaseAssetQuantity, price, salt, reduceOnly, postOnly)
							orderHash := getOrderHash(longOrder)

							mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Placed)).Times(1)
							mockBibliophile.EXPECT().GetTimeStamp().Return(hu.V1ActivationTime).Times(1)
							mockBibliophile.EXPECT().GetAvailableMargin(longOrder.Trader, hu.V1).Return(big.NewInt(-1)).Times(1)
							mockBibliophile.EXPECT().GetOrderFilledAmount(orderHash).Return(big.NewInt(0)).Times(1)
							mockBibliophile.EXPECT().GetMarketAddressFromMarketID(longOrder.AmmIndex.Int64()).Return(ammAddress).Times(1)
							mockBibliophile.EXPECT().IsValidator(longOrder.Trader).Return(true).Times(1)

							input := getValidateCancelLimitOrderV2Input(longOrder, trader, assertLowMargin, assertOverPositionCap)
							output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
							assert.Equal(t, "", output.Err)
							assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
							assert.Equal(t, ammAddress, output.Res.Amm)
							assert.Equal(t, longOrder.BaseAssetQuantity, output.Res.UnfilledAmount)
						})
						t.Run("for a shortOrder it returns err = nil, with ammAddress and unfilled amount of cancelled Order", func(t *testing.T) {
							shortOrder := getOrder(ammIndex, trader, shortBaseAssetQuantity, price, salt, reduceOnly, postOnly)
							orderHash := getOrderHash(shortOrder)

							mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Placed)).Times(1)
							mockBibliophile.EXPECT().GetTimeStamp().Return(hu.V1ActivationTime).Times(1)
							mockBibliophile.EXPECT().GetAvailableMargin(shortOrder.Trader, hu.V1).Return(big.NewInt(-1)).Times(1)
							mockBibliophile.EXPECT().GetOrderFilledAmount(orderHash).Return(big.NewInt(0)).Times(1)
							mockBibliophile.EXPECT().GetMarketAddressFromMarketID(shortOrder.AmmIndex.Int64()).Return(ammAddress).Times(1)
							mockBibliophile.EXPECT().IsValidator(shortOrder.Trader).Return(true).Times(1)

							input := getValidateCancelLimitOrderV2Input(shortOrder, trader, assertLowMargin, assertOverPositionCap)
							output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
							assert.Equal(t, "", output.Err)
							assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
							assert.Equal(t, ammAddress, output.Res.Amm)
							assert.Equal(t, shortOrder.BaseAssetQuantity, output.Res.UnfilledAmount)
						})
					})
					t.Run("for a partially filled Order", func(t *testing.T) {
						t.Run("for a longOrder it returns err = nil, with ammAddress and unfilled amount of cancelled Order", func(t *testing.T) {
							longOrder := getOrder(ammIndex, trader, longBaseAssetQuantity, price, salt, reduceOnly, postOnly)
							orderHash := getOrderHash(longOrder)
							filledAmount := hu.Div(longOrder.BaseAssetQuantity, big.NewInt(2))

							mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Placed)).Times(1)
							mockBibliophile.EXPECT().GetTimeStamp().Return(hu.V1ActivationTime).Times(1)
							mockBibliophile.EXPECT().GetAvailableMargin(longOrder.Trader, hu.V1).Return(big.NewInt(-1)).Times(1)
							mockBibliophile.EXPECT().GetOrderFilledAmount(orderHash).Return(filledAmount).Times(1)
							mockBibliophile.EXPECT().GetMarketAddressFromMarketID(longOrder.AmmIndex.Int64()).Return(ammAddress).Times(1)
							mockBibliophile.EXPECT().IsValidator(longOrder.Trader).Return(true).Times(1)

							input := getValidateCancelLimitOrderV2Input(longOrder, trader, assertLowMargin, assertOverPositionCap)
							output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
							assert.Equal(t, "", output.Err)
							assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
							assert.Equal(t, ammAddress, output.Res.Amm)
							expectedUnfilleAmount := hu.Sub(longOrder.BaseAssetQuantity, filledAmount)
							assert.Equal(t, expectedUnfilleAmount, output.Res.UnfilledAmount)
						})
						t.Run("for a shortOrder it returns err = nil, with ammAddress and unfilled amount of cancelled Order", func(t *testing.T) {
							shortOrder := getOrder(ammIndex, trader, shortBaseAssetQuantity, price, salt, reduceOnly, postOnly)
							orderHash := getOrderHash(shortOrder)
							filledAmount := hu.Div(shortOrder.BaseAssetQuantity, big.NewInt(2))

							mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Placed)).Times(1)
							mockBibliophile.EXPECT().GetTimeStamp().Return(hu.V1ActivationTime).Times(1)
							mockBibliophile.EXPECT().GetAvailableMargin(shortOrder.Trader, hu.V1).Return(big.NewInt(-1)).Times(1)
							mockBibliophile.EXPECT().GetOrderFilledAmount(orderHash).Return(filledAmount).Times(1)
							mockBibliophile.EXPECT().GetMarketAddressFromMarketID(shortOrder.AmmIndex.Int64()).Return(ammAddress).Times(1)
							mockBibliophile.EXPECT().IsValidator(shortOrder.Trader).Return(true).Times(1)

							input := getValidateCancelLimitOrderV2Input(shortOrder, trader, assertLowMargin, assertOverPositionCap)
							output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
							assert.Equal(t, "", output.Err)
							assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
							assert.Equal(t, ammAddress, output.Res.Amm)
							expectedUnfilleAmount := hu.Sub(shortOrder.BaseAssetQuantity, filledAmount)
							assert.Equal(t, expectedUnfilleAmount, output.Res.UnfilledAmount)
						})
					})
				})
			})
			t.Run("when assertLowMargin is false", func(t *testing.T) {
				assertLowMargin := false
				t.Run("for an unfilled Order", func(t *testing.T) {
					t.Run("for a longOrder it returns err = nil, with ammAddress and unfilled amount of cancelled Order", func(t *testing.T) {
						longOrder := getOrder(ammIndex, trader, longBaseAssetQuantity, price, salt, reduceOnly, postOnly)
						orderHash := getOrderHash(longOrder)

						mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Placed)).Times(1)
						mockBibliophile.EXPECT().GetOrderFilledAmount(orderHash).Return(big.NewInt(0)).Times(1)
						mockBibliophile.EXPECT().GetMarketAddressFromMarketID(longOrder.AmmIndex.Int64()).Return(ammAddress).Times(1)

						input := getValidateCancelLimitOrderV2Input(longOrder, trader, assertLowMargin, assertOverPositionCap)
						output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
						assert.Equal(t, "", output.Err)
						assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
						assert.Equal(t, ammAddress, output.Res.Amm)
						assert.Equal(t, longOrder.BaseAssetQuantity, output.Res.UnfilledAmount)
					})
					t.Run("for a shortOrder it returns err = nil, with ammAddress and unfilled amount of cancelled Order", func(t *testing.T) {
						shortOrder := getOrder(ammIndex, trader, shortBaseAssetQuantity, price, salt, reduceOnly, postOnly)
						orderHash := getOrderHash(shortOrder)

						mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Placed)).Times(1)
						mockBibliophile.EXPECT().GetOrderFilledAmount(orderHash).Return(big.NewInt(0)).Times(1)
						mockBibliophile.EXPECT().GetMarketAddressFromMarketID(shortOrder.AmmIndex.Int64()).Return(ammAddress).Times(1)

						input := getValidateCancelLimitOrderV2Input(shortOrder, trader, assertLowMargin, assertOverPositionCap)
						output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
						assert.Equal(t, "", output.Err)
						assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
						assert.Equal(t, ammAddress, output.Res.Amm)
						assert.Equal(t, shortOrder.BaseAssetQuantity, output.Res.UnfilledAmount)
					})
				})
				t.Run("for a partially filled Order", func(t *testing.T) {
					t.Run("for a longOrder it returns err = nil, with ammAddress and unfilled amount of cancelled Order", func(t *testing.T) {
						longOrder := getOrder(ammIndex, trader, longBaseAssetQuantity, price, salt, reduceOnly, postOnly)
						orderHash := getOrderHash(longOrder)
						filledAmount := hu.Div(longOrder.BaseAssetQuantity, big.NewInt(2))

						mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Placed)).Times(1)
						mockBibliophile.EXPECT().GetOrderFilledAmount(orderHash).Return(filledAmount).Times(1)
						mockBibliophile.EXPECT().GetMarketAddressFromMarketID(longOrder.AmmIndex.Int64()).Return(ammAddress).Times(1)

						input := getValidateCancelLimitOrderV2Input(longOrder, trader, assertLowMargin, assertOverPositionCap)
						output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
						assert.Equal(t, "", output.Err)
						assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
						assert.Equal(t, ammAddress, output.Res.Amm)
						expectedUnfilleAmount := hu.Sub(longOrder.BaseAssetQuantity, filledAmount)
						assert.Equal(t, expectedUnfilleAmount, output.Res.UnfilledAmount)
					})
					t.Run("for a shortOrder it returns err = nil, with ammAddress and unfilled amount of cancelled Order", func(t *testing.T) {
						shortOrder := getOrder(ammIndex, trader, shortBaseAssetQuantity, price, salt, reduceOnly, postOnly)
						orderHash := getOrderHash(shortOrder)
						filledAmount := hu.Div(shortOrder.BaseAssetQuantity, big.NewInt(2))

						mockBibliophile.EXPECT().GetOrderStatus(orderHash).Return(int64(juror.Placed)).Times(1)
						mockBibliophile.EXPECT().GetOrderFilledAmount(orderHash).Return(filledAmount).Times(1)
						mockBibliophile.EXPECT().GetMarketAddressFromMarketID(shortOrder.AmmIndex.Int64()).Return(ammAddress).Times(1)

						input := getValidateCancelLimitOrderV2Input(shortOrder, trader, assertLowMargin, assertOverPositionCap)
						output := ValidateCancelLimitOrderV2(mockBibliophile, &input)
						assert.Equal(t, "", output.Err)
						assert.Equal(t, orderHash, common.BytesToHash(output.OrderHash[:]))
						assert.Equal(t, ammAddress, output.Res.Amm)
						expectedUnfilleAmount := hu.Sub(shortOrder.BaseAssetQuantity, filledAmount)
						assert.Equal(t, expectedUnfilleAmount, output.Res.UnfilledAmount)
					})
				})
			})
		})
	})
}

func getValidateCancelLimitOrderV2Input(order ILimitOrderBookOrder, sender common.Address, assertLowMargin bool, assertOverPositionCap bool) ValidateCancelLimitOrderV2Input {
	return ValidateCancelLimitOrderV2Input{
		Order:                 order,
		Sender:                sender,
		AssertLowMargin:       assertLowMargin,
		AssertOverPositionCap: assertOverPositionCap,
	}
}

func getOrder(ammIndex *big.Int, trader common.Address, baseAssetQuantity *big.Int, price *big.Int, salt *big.Int, reduceOnly bool, postOnly bool) ILimitOrderBookOrder {
	return ILimitOrderBookOrder{
		AmmIndex:          ammIndex,
		BaseAssetQuantity: baseAssetQuantity,
		Trader:            trader,
		Price:             price,
		Salt:              salt,
		ReduceOnly:        reduceOnly,
		PostOnly:          postOnly,
	}
}

func getOrderHash(order ILimitOrderBookOrder) common.Hash {
	orderHash, err := GetLimitOrderHashFromContractStruct(&order)
	if err != nil {
		panic("error in getting order hash")
	}
	return orderHash
}
