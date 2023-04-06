// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package limitorders

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ava-labs/subnet-evm/rpc"
	"github.com/ethereum/go-ethereum/common"
)

type OrderBookAPI struct {
	db LimitOrderDatabase
}

func NewOrderBookAPI(database LimitOrderDatabase) *OrderBookAPI {
	return &OrderBookAPI{
		db: database,
	}
}

type OrderBookResponse struct {
	Orders []OrderMin
}

type OpenOrdersResponse struct {
	Orders []OrderForOpenOrders
}

type OrderMin struct {
	Market
	Price   string
	Size    string
	Signer  string
	OrderId string
}

type OrderForOpenOrders struct {
	Market
	Price      string
	Size       string
	FilledSize string
	Timestamp  uint64
	Salt       string
	OrderId    string
}

func (api *OrderBookAPI) GetDetailedOrderBookData(ctx context.Context) InMemoryDatabase {
	return api.db.GetOrderBookData()
}

func (api *OrderBookAPI) GetOrderBook(ctx context.Context, marketStr string) (*OrderBookResponse, error) {
	// market is a string cuz it's an optional param
	allOrders := api.db.GetOrderBookData().OrderMap
	orders := []OrderMin{}

	if len(marketStr) > 0 {
		market, err := strconv.Atoi(marketStr)
		if err != nil {
			return nil, fmt.Errorf("invalid market")
		}
		marketOrders := map[common.Hash]*LimitOrder{}
		for hash, order := range allOrders {
			if order.Market == Market(market) {
				if order.PositionType == "short" /* || order.Price.Cmp(big.NewInt(20e6)) <= 0 */ {
					marketOrders[hash] = order
				}
			}
		}
		allOrders = marketOrders
	}

	for hash, order := range allOrders {
		orders = append(orders, OrderMin{
			Market:  order.Market,
			Price:   order.Price.String(),
			Size:    order.GetUnFilledBaseAssetQuantity().String(),
			Signer:  order.UserAddress,
			OrderId: hash.String(),
		})
	}

	return &OrderBookResponse{Orders: orders}, nil
}

func (api *OrderBookAPI) GetOpenOrders(ctx context.Context, trader string) OpenOrdersResponse {
	traderOrders := []OrderForOpenOrders{}
	orderMap := api.db.GetOrderBookData().OrderMap
	for hash, order := range orderMap {
		if strings.EqualFold(order.UserAddress, trader) {
			traderOrders = append(traderOrders, OrderForOpenOrders{
				Market:     order.Market,
				Price:      order.Price.String(),
				Size:       order.BaseAssetQuantity.String(),
				FilledSize: order.FilledBaseAssetQuantity.String(),
				Salt:       getOrderFromRawOrder(order.RawOrder).Salt.String(),
				OrderId:    hash.String(),
			})
		}
	}

	return OpenOrdersResponse{Orders: traderOrders}
}

func (api *OrderBookAPI) GetAggregatedOrderBookState(ctx context.Context, market int) *AggregatedOrderBookState {
	return aggregatedOrderBookState(api.db, Market(market))
}

func (api *OrderBookAPI) AggregatedOrderBookState(ctx context.Context, market int) (*rpc.Subscription, error) {
	notifier, _ := rpc.NotifierFromContext(ctx)
	rpcSub := notifier.CreateSubscription()

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				notifier.Notify(rpcSub.ID, aggregatedOrderBookState(api.db, Market(market)))
			case <-notifier.Closed():
				ticker.Stop()
				return
			}
		}
	}()

	return rpcSub, nil
}

func aggregatedOrderBookState(db LimitOrderDatabase, market Market) *AggregatedOrderBookState {
	longOrders := db.GetLongOrders(market)
	shortOrders := db.GetShortOrders(market)
	return &AggregatedOrderBookState{
		Market: market,
		Longs:  aggregateOrdersByPrice(longOrders),
		Shorts: aggregateOrdersByPrice(shortOrders),
	}
}

func aggregateOrdersByPrice(orders []LimitOrder) map[string]string {
	aggregatedOrders := map[string]string{}
	for _, order := range orders {
		aggregatedBaseAssetQuantity, ok := aggregatedOrders[order.Price.String()]
		if ok {
			quantity, _ := big.NewInt(0).SetString(aggregatedBaseAssetQuantity, 10)
			quantity.Add(quantity, order.BaseAssetQuantity)
		} else {
			aggregatedOrders[order.Price.String()] = order.BaseAssetQuantity.String()
		}
	}
	return aggregatedOrders
}

type AggregatedOrderBookState struct {
	Market Market            `json:"market"`
	Longs  map[string]string `json:"longs"`
	Shorts map[string]string `json:"shorts"`
}
