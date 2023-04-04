package limitorders

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAggregatedOrderBook(t *testing.T) {
	t.Run("it aggregates long and short orders by price and returns aggregated data in json format with blockNumber", func(t *testing.T) {
		db := NewInMemoryDatabase()
		service := NewOrderBookAPI(db)

		longOrder1 := getLongOrder()
		db.Add(getIdFromLimitOrder(longOrder1), &longOrder1)

		longOrder2 := getLongOrder()
		longOrder2.Salt.Add(longOrder2.Salt, big.NewInt(100))
		longOrder2.Price.Mul(longOrder2.Price, big.NewInt(2))
		db.Add(getIdFromLimitOrder(longOrder2), &longOrder2)

		shortOrder1 := getShortOrder()
		shortOrder1.Salt.Add(shortOrder1.Salt, big.NewInt(200))
		db.Add(getIdFromLimitOrder(shortOrder1), &shortOrder1)

		shortOrder2 := getShortOrder()
		shortOrder2.Salt.Add(shortOrder1.Salt, big.NewInt(300))
		shortOrder2.Price.Mul(shortOrder2.Price, big.NewInt(2))
		db.Add(getIdFromLimitOrder(shortOrder2), &shortOrder2)

		ctx := context.TODO()
		response := service.GetAggregatedOrderBookState(ctx, int(AvaxPerp))
		expectedAggregatedOrderBookState := AggregatedOrderBookState{
			Market:      AvaxPerp,
			BlockNumber: big.NewInt(1),
			Longs: map[int64]*big.Int{
				longOrder1.Price.Int64(): longOrder1.BaseAssetQuantity,
				longOrder2.Price.Int64(): longOrder2.BaseAssetQuantity,
			},
			Shorts: map[int64]*big.Int{
				shortOrder1.Price.Int64(): shortOrder1.BaseAssetQuantity,
				shortOrder2.Price.Int64(): shortOrder2.BaseAssetQuantity,
			},
		}
		assert.Equal(t, expectedAggregatedOrderBookState, *response)
	})
}
