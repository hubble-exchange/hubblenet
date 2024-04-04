package orderbook

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"testing"

	hu "github.com/ava-labs/subnet-evm/hubbleutils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestDecodeLimitOrder(t *testing.T) {
	t.Run("long order", func(t *testing.T) {
		testDecodeTypeAndEncodedOrder(
			t,
			strings.TrimPrefix("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c80000000000000000000000000000000000000000000000004563918244f4000000000000000000000000000000000000000000000000000000000000000f42400000000000000000000000000000000000000000000000000000018a82b01e9d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", "0x"),
			strings.TrimPrefix("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c80000000000000000000000000000000000000000000000004563918244f4000000000000000000000000000000000000000000000000000000000000000f42400000000000000000000000000000000000000000000000000000018a82b01e9d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", "0x"),
			Limit,
			LimitOrder{
				BaseOrder: hu.BaseOrder{
					AmmIndex:          big.NewInt(0),
					Trader:            common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8"),
					BaseAssetQuantity: big.NewInt(5000000000000000000),
					Price:             big.NewInt(1000000),
					Salt:              big.NewInt(1694409694877),
					ReduceOnly:        false,
				},
				PostOnly: false,
			},
		)
	})

	t.Run("long order reduce only", func(t *testing.T) {
		testDecodeTypeAndEncodedOrder(
			t,
			strings.TrimPrefix("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c80000000000000000000000000000000000000000000000004563918244f4000000000000000000000000000000000000000000000000000000000000000f42400000000000000000000000000000000000000000000000000000018a82b4121c00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000", "0x"),
			strings.TrimPrefix("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c80000000000000000000000000000000000000000000000004563918244f4000000000000000000000000000000000000000000000000000000000000000f42400000000000000000000000000000000000000000000000000000018a82b4121c00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000", "0x"),
			Limit,
			LimitOrder{
				BaseOrder: hu.BaseOrder{
					AmmIndex:          big.NewInt(0),
					Trader:            common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8"),
					BaseAssetQuantity: big.NewInt(5000000000000000000),
					Price:             big.NewInt(1000000),
					Salt:              big.NewInt(1694409953820),
					ReduceOnly:        true,
				},
				PostOnly: false,
			},
		)
	})

	t.Run("short order", func(t *testing.T) {
		order := LimitOrder{
			BaseOrder: hu.BaseOrder{
				AmmIndex:          big.NewInt(0),
				Trader:            common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8"),
				BaseAssetQuantity: big.NewInt(-5000000000000000000),
				Price:             big.NewInt(1000000),
				Salt:              big.NewInt(1694410024592),
				ReduceOnly:        false,
			},
			PostOnly: false,
		}
		orderHash, err := order.Hash()
		assert.Nil(t, err)
		assert.Equal(t, "0x0d87f0d9a37bc19fc3557db4085088cbecc5d6f3ff63c05f6db33684b8145108", orderHash.Hex())
		testDecodeTypeAndEncodedOrder(
			t,
			strings.TrimPrefix("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c8ffffffffffffffffffffffffffffffffffffffffffffffffba9c6e7dbb0c000000000000000000000000000000000000000000000000000000000000000f42400000000000000000000000000000000000000000000000000000018a82b5269000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", "0x"),
			strings.TrimPrefix("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c8ffffffffffffffffffffffffffffffffffffffffffffffffba9c6e7dbb0c000000000000000000000000000000000000000000000000000000000000000f42400000000000000000000000000000000000000000000000000000018a82b5269000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", "0x"),
			Limit,
			order,
		)
	})

	t.Run("short order reduce only", func(t *testing.T) {
		testDecodeTypeAndEncodedOrder(
			t,
			strings.TrimPrefix("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c8ffffffffffffffffffffffffffffffffffffffffffffffffba9c6e7dbb0c000000000000000000000000000000000000000000000000000000000000000f42400000000000000000000000000000000000000000000000000000018a82b7597700000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000", "0x"),
			strings.TrimPrefix("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c8ffffffffffffffffffffffffffffffffffffffffffffffffba9c6e7dbb0c000000000000000000000000000000000000000000000000000000000000000f42400000000000000000000000000000000000000000000000000000018a82b7597700000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000", "0x"),
			Limit,
			LimitOrder{
				BaseOrder: hu.BaseOrder{
					AmmIndex:          big.NewInt(0),
					Trader:            common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8"),
					BaseAssetQuantity: big.NewInt(-5000000000000000000),
					Price:             big.NewInt(1000000),
					Salt:              big.NewInt(1694410168695),
					ReduceOnly:        true,
				},
				PostOnly: false,
			},
		)
	})
	t.Run("short order reduce only with post order", func(t *testing.T) {
		testDecodeTypeAndEncodedOrder(
			t,
			strings.TrimPrefix("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c8ffffffffffffffffffffffffffffffffffffffffffffffffba9c6e7dbb0c000000000000000000000000000000000000000000000000000000000000000f42400000000000000000000000000000000000000000000000000000018a82b8382e00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001", "0x"),
			strings.TrimPrefix("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c8ffffffffffffffffffffffffffffffffffffffffffffffffba9c6e7dbb0c000000000000000000000000000000000000000000000000000000000000000f42400000000000000000000000000000000000000000000000000000018a82b8382e00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001", "0x"),
			Limit,
			LimitOrder{
				BaseOrder: hu.BaseOrder{
					AmmIndex:          big.NewInt(0),
					Trader:            common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8"),
					BaseAssetQuantity: big.NewInt(-5000000000000000000),
					Price:             big.NewInt(1000000),
					Salt:              big.NewInt(1694410225710),
					ReduceOnly:        true,
				},
				PostOnly: true,
			},
		)
	})
}

func testDecodeTypeAndEncodedOrder(t *testing.T, typedEncodedOrder string, encodedOrder string, orderType OrderType, expectedOutput interface{}) {
	testData, err := hex.DecodeString(typedEncodedOrder)
	assert.Nil(t, err)

	decodeStep, err := hu.DecodeTypeAndEncodedOrder(testData)
	assert.Nil(t, err)

	assert.Equal(t, orderType, decodeStep.OrderType)
	assert.Equal(t, encodedOrder, hex.EncodeToString(decodeStep.EncodedOrder))
	testDecodeLimitOrder(t, encodedOrder, expectedOutput)
}

func testDecodeLimitOrder(t *testing.T, encodedOrder string, expectedOutput interface{}) {
	testData, err := hex.DecodeString(encodedOrder)
	assert.Nil(t, err)

	result, err := hu.DecodeLimitOrder(testData)
	fmt.Println(result)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assertLimitOrderEquality(t, expectedOutput.(LimitOrder).BaseOrder, result.BaseOrder)
	assert.Equal(t, expectedOutput.(LimitOrder).PostOnly, result.PostOnly)
}

func TestDecodeIOCOrder(t *testing.T) {
	t.Run("long order", func(t *testing.T) {
		order := &IOCOrder{
			OrderType: 1,
			ExpireAt:  big.NewInt(1688994854),
			BaseOrder: hu.BaseOrder{
				AmmIndex:          big.NewInt(0),
				Trader:            common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8"),
				BaseAssetQuantity: big.NewInt(5000000000000000000),
				Price:             big.NewInt(1000000000),
				Salt:              big.NewInt(1688994806105),
				ReduceOnly:        false,
			},
		}
		h, err := order.Hash()
		assert.Nil(t, err)
		assert.Equal(t, "0xc989b9a5bf196036dbbae61f56179f31172cc04aa91238bc1b7c828bebf0fe5e", h.Hex())

		typeEncodedOrder := strings.TrimPrefix("0x00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000064ac0426000000000000000000000000000000000000000000000000000000000000000000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c80000000000000000000000000000000000000000000000004563918244f40000000000000000000000000000000000000000000000000000000000003b9aca00000000000000000000000000000000000000000000000000000001893fef79590000000000000000000000000000000000000000000000000000000000000000", "0x")
		encodedOrder := strings.TrimPrefix("0x00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000064ac0426000000000000000000000000000000000000000000000000000000000000000000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c80000000000000000000000000000000000000000000000004563918244f40000000000000000000000000000000000000000000000000000000000003b9aca00000000000000000000000000000000000000000000000000000001893fef79590000000000000000000000000000000000000000000000000000000000000000", "0x")
		b, err := order.EncodeToABI()
		assert.Nil(t, err)
		assert.Equal(t, typeEncodedOrder, hex.EncodeToString(b))
		testDecodeTypeAndEncodedIOCOrder(t, typeEncodedOrder, encodedOrder, IOC, order)
	})

	t.Run("short order", func(t *testing.T) {
		order := &IOCOrder{
			OrderType: 1,
			ExpireAt:  big.NewInt(1688994854),
			BaseOrder: hu.BaseOrder{
				AmmIndex:          big.NewInt(0),
				Trader:            common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8"),
				BaseAssetQuantity: big.NewInt(-5000000000000000000),
				Price:             big.NewInt(1000000000),
				Salt:              big.NewInt(1688994806105),
				ReduceOnly:        false,
			},
		}
		h, err := order.Hash()
		assert.Nil(t, err)
		assert.Equal(t, "0x4f92bf62284e2080d3d3cf7c15dcddf1c1a496902c1742de78737d3d9a870661", h.Hex())

		typeEncodedOrder := strings.TrimPrefix("0x00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000064ac0426000000000000000000000000000000000000000000000000000000000000000000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c8ffffffffffffffffffffffffffffffffffffffffffffffffba9c6e7dbb0c0000000000000000000000000000000000000000000000000000000000003b9aca00000000000000000000000000000000000000000000000000000001893fef79590000000000000000000000000000000000000000000000000000000000000000", "0x")
		encodedOrder := strings.TrimPrefix("0x00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000064ac0426000000000000000000000000000000000000000000000000000000000000000000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c8ffffffffffffffffffffffffffffffffffffffffffffffffba9c6e7dbb0c0000000000000000000000000000000000000000000000000000000000003b9aca00000000000000000000000000000000000000000000000000000001893fef79590000000000000000000000000000000000000000000000000000000000000000", "0x")
		b, err := order.EncodeToABI()
		assert.Nil(t, err)
		assert.Equal(t, typeEncodedOrder, hex.EncodeToString(b))
		testDecodeTypeAndEncodedIOCOrder(t, typeEncodedOrder, encodedOrder, IOC, order)
	})
}

func testDecodeTypeAndEncodedIOCOrder(t *testing.T, typedEncodedOrder string, encodedOrder string, orderType OrderType, expectedOutput *IOCOrder) {
	testData, err := hex.DecodeString(typedEncodedOrder)
	assert.Nil(t, err)

	decodeStep, err := hu.DecodeTypeAndEncodedOrder(testData)
	assert.Nil(t, err)

	assert.Equal(t, orderType, decodeStep.OrderType)
	assert.Equal(t, encodedOrder, hex.EncodeToString(decodeStep.EncodedOrder))
	testDecodeIOCOrder(t, decodeStep.EncodedOrder, expectedOutput)
}

func testDecodeIOCOrder(t *testing.T, encodedOrder []byte, expectedOutput *IOCOrder) {
	result, err := hu.DecodeIOCOrder(encodedOrder)
	assert.NoError(t, err)
	fmt.Println(result)
	assert.NotNil(t, result)
	assertIOCOrderEquality(t, expectedOutput, result)
}

func assertIOCOrderEquality(t *testing.T, expected, actual *IOCOrder) {
	assert.Equal(t, expected.OrderType, actual.OrderType)
	assert.Equal(t, expected.ExpireAt.Int64(), actual.ExpireAt.Int64())
	assertLimitOrderEquality(t, expected.BaseOrder, actual.BaseOrder)
}

func assertLimitOrderEquality(t *testing.T, expected, actual hu.BaseOrder) {
	assert.Equal(t, expected.AmmIndex.Int64(), actual.AmmIndex.Int64())
	assert.Equal(t, expected.Trader, actual.Trader)
	assert.Equal(t, expected.BaseAssetQuantity, actual.BaseAssetQuantity)
	assert.Equal(t, expected.Price, actual.Price)
	assert.Equal(t, expected.Salt, actual.Salt)
	assert.Equal(t, expected.ReduceOnly, actual.ReduceOnly)
}
