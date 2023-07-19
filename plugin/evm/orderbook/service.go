// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package orderbook

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/eth"
	"github.com/ava-labs/subnet-evm/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
)

type OrderBookAPI struct {
	db            LimitOrderDatabase
	backend       *eth.EthAPIBackend
	configService IConfigService
}

func NewOrderBookAPI(database LimitOrderDatabase, backend *eth.EthAPIBackend, configService IConfigService) *OrderBookAPI {
	return &OrderBookAPI{
		db:            database,
		backend:       backend,
		configService: configService,
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
	ReduceOnly bool
	OrderType  OrderType
}

type GetDebugDataResponse struct {
	MarginFraction   map[common.Address]*big.Int
	AvailableMargin  map[common.Address]*big.Int
	PendingFunding   map[common.Address]*big.Int
	Margin           map[common.Address]*big.Int
	UtilisedMargin   map[common.Address]*big.Int
	ReservedMargin   map[common.Address]*big.Int
	NotionalPosition map[common.Address]*big.Int
	UnrealizePnL     map[common.Address]*big.Int
	LastPrice        map[Market]*big.Int
	OraclePrice      map[Market]*big.Int
}

func (api *OrderBookAPI) GetDebugData(ctx context.Context, trader string) GetDebugDataResponse {
	traderHash := common.HexToAddress(trader)
	response := GetDebugDataResponse{
		MarginFraction:   map[common.Address]*big.Int{},
		AvailableMargin:  map[common.Address]*big.Int{},
		PendingFunding:   map[common.Address]*big.Int{},
		Margin:           map[common.Address]*big.Int{},
		NotionalPosition: map[common.Address]*big.Int{},
		UnrealizePnL:     map[common.Address]*big.Int{},
		UtilisedMargin:   map[common.Address]*big.Int{},
		ReservedMargin:   map[common.Address]*big.Int{},
		LastPrice:        map[Market]*big.Int{},
		OraclePrice:      map[Market]*big.Int{},
	}

	traderMap := api.db.GetAllTraders()
	if trader != "" {
		traderMap = map[common.Address]Trader{
			traderHash: traderMap[traderHash],
		}
	}

	minAllowableMargin := api.configService.getMinAllowableMargin()
	prices := api.configService.GetUnderlyingPrices()
	lastPrices := api.db.GetLastPrices()
	oraclePrices := map[Market]*big.Int{}
	count := api.configService.GetActiveMarketsCount()
	markets := make([]Market, count)
	for i := int64(0); i < count; i++ {
		markets[i] = Market(i)
		oraclePrices[Market(i)] = prices[Market(i)]
	}

	for addr, trader := range traderMap {
		pendingFunding := getTotalFunding(&trader, markets)
		margin := new(big.Int).Sub(getNormalisedMargin(&trader), pendingFunding)
		notionalPosition, unrealizePnL := getTotalNotionalPositionAndUnrealizedPnl(&trader, margin, Min_Allowable_Margin, oraclePrices, lastPrices, markets)
		marginFraction := calcMarginFraction(&trader, pendingFunding, oraclePrices, lastPrices, markets)
		availableMargin := getAvailableMargin(&trader, pendingFunding, oraclePrices, lastPrices, api.configService.getMinAllowableMargin(), markets)
		utilisedMargin := divideByBasePrecision(new(big.Int).Mul(notionalPosition, minAllowableMargin))

		response.MarginFraction[addr] = marginFraction
		response.AvailableMargin[addr] = availableMargin
		response.PendingFunding[addr] = pendingFunding
		response.Margin[addr] = getNormalisedMargin(&trader)
		response.UtilisedMargin[addr] = utilisedMargin
		response.NotionalPosition[addr] = notionalPosition
		response.UnrealizePnL[addr] = unrealizePnL
		response.ReservedMargin[addr] = trader.Margin.Reserved
	}

	response.LastPrice = lastPrices
	response.OraclePrice = oraclePrices
	return response
}

func (api *OrderBookAPI) GetDetailedOrderBookData(ctx context.Context) InMemoryDatabase {
	return api.db.GetOrderBookData()
}

func (api *OrderBookAPI) GetOrderBook(ctx context.Context, marketStr string) (*OrderBookResponse, error) {
	market, err := parseMarket(marketStr)
	if err != nil {
		return nil, err
	}
	allOrders := api.db.GetAllOrders()
	orders := []OrderMin{}
	for _, order := range allOrders {
		if market == nil || order.Market == Market(*market) {
			orders = append(orders, order.ToOrderMin())
		}
	}
	return &OrderBookResponse{Orders: orders}, nil
}

func parseMarket(marketStr string) (*int, error) {
	var market *int
	if len(marketStr) > 0 {
		_market, err := strconv.Atoi(marketStr)
		if err != nil {
			return nil, fmt.Errorf("invalid market")
		}
		market = &_market
	}
	return market, nil
}

func (api *OrderBookAPI) GetOpenOrders(ctx context.Context, trader string, marketStr string) (*OpenOrdersResponse, error) {
	market, err := parseMarket(marketStr)
	if err != nil {
		return nil, err
	}
	traderOrders := []OrderForOpenOrders{}
	traderHash := common.HexToAddress(trader)
	orders := api.db.GetOpenOrdersForTrader(traderHash, LimitOrderType)
	for _, order := range orders {
		if strings.EqualFold(order.UserAddress, trader) && (market == nil || order.Market == Market(*market)) {
			traderOrders = append(traderOrders, OrderForOpenOrders{
				Market:     order.Market,
				Price:      order.Price.String(),
				Size:       order.BaseAssetQuantity.String(),
				FilledSize: order.FilledBaseAssetQuantity.String(),
				Salt:       order.Salt.String(),
				OrderId:    order.Id.String(),
				ReduceOnly: order.ReduceOnly,
				OrderType:  order.OrderType,
			})
		}
	}
	return &OpenOrdersResponse{Orders: traderOrders}, nil
}

// NewOrderBookState send a notification each time a new (header) block is appended to the chain.
func (api *OrderBookAPI) NewOrderBookState(ctx context.Context) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}

	rpcSub := notifier.CreateSubscription()

	go func() {
		var (
			headers    = make(chan core.ChainHeadEvent)
			headersSub event.Subscription
		)

		headersSub = api.backend.SubscribeChainHeadEvent(headers)
		defer headersSub.Unsubscribe()

		for {
			select {
			case <-headers:
				orderBookData := api.GetDetailedOrderBookData(ctx)
				notifier.Notify(rpcSub.ID, &orderBookData)
			case <-rpcSub.Err():
				headersSub.Unsubscribe()
				return
			case <-notifier.Closed():
				headersSub.Unsubscribe()
				return
			}
		}
	}()

	return rpcSub, nil
}

func (api *OrderBookAPI) GetDepthForMarket(ctx context.Context, market int) *MarketDepth {
	return getDepthForMarket(api.db, Market(market))
}

func (api *OrderBookAPI) StreamDepthUpdateForMarket(ctx context.Context, market int) (*rpc.Subscription, error) {
	notifier, _ := rpc.NotifierFromContext(ctx)
	rpcSub := notifier.CreateSubscription()

	ticker := time.NewTicker(1 * time.Second)

	var oldMarketDepth = &MarketDepth{}

	go func() {
		for {
			select {
			case <-ticker.C:
				newMarketDepth := getDepthForMarket(api.db, Market(market))
				depthUpdate := getUpdateInDepth(newMarketDepth, oldMarketDepth)
				notifier.Notify(rpcSub.ID, depthUpdate)
				oldMarketDepth = newMarketDepth
			case <-notifier.Closed():
				ticker.Stop()
				return
			}
		}
	}()

	return rpcSub, nil
}

func getUpdateInDepth(newMarketDepth *MarketDepth, oldMarketDepth *MarketDepth) *MarketDepth {
	var diff = &MarketDepth{
		Market: newMarketDepth.Market,
		Longs:  map[string]string{},
		Shorts: map[string]string{},
	}
	for price, depth := range newMarketDepth.Longs {
		oldDepth := oldMarketDepth.Longs[price]
		if oldDepth != depth {
			diff.Longs[price] = depth
		}
	}
	for price := range oldMarketDepth.Longs {
		if newMarketDepth.Longs[price] == "" {
			diff.Longs[price] = big.NewInt(0).String()
		}
	}
	for price, depth := range newMarketDepth.Shorts {
		oldDepth := oldMarketDepth.Shorts[price]
		if oldDepth != depth {
			diff.Shorts[price] = depth
		}
	}
	for price := range oldMarketDepth.Shorts {
		if newMarketDepth.Shorts[price] == "" {
			diff.Shorts[price] = big.NewInt(0).String()
		}
	}
	return diff
}

func getDepthForMarket(db LimitOrderDatabase, market Market) *MarketDepth {
	// currentBlock number only needs to be passed in for the retry logic for failed orders.
	// There are some orders in the book that could have been marked failed,
	// but because of our retry logic they might be retried every 100 blocks
	// So, one could argue that is this not a super accurate representation of the order book
	// BUT for the argument sake, we could also say that these retry orders can be treated as "fresh" orders
	longOrders := db.GetLongOrders(market, nil /* lowerbound */, nil /* currentBlock */)
	shortOrders := db.GetShortOrders(market, nil /* upperbound */, nil /* currentBlock */)
	return &MarketDepth{
		Market: market,
		Longs:  aggregateOrdersByPrice(longOrders),
		Shorts: aggregateOrdersByPrice(shortOrders),
	}
}

func aggregateOrdersByPrice(orders []Order) map[string]string {
	aggregatedOrders := map[string]string{}
	for _, order := range orders {
		aggregatedBaseAssetQuantity, ok := aggregatedOrders[order.Price.String()]
		if ok {
			quantity, _ := big.NewInt(0).SetString(aggregatedBaseAssetQuantity, 10)
			aggregatedOrders[order.Price.String()] = quantity.Add(quantity, order.GetUnFilledBaseAssetQuantity()).String()
		} else {
			aggregatedOrders[order.Price.String()] = order.GetUnFilledBaseAssetQuantity().String()
		}
	}
	return aggregatedOrders
}

type MarketDepth struct {
	Market Market            `json:"market"`
	Longs  map[string]string `json:"longs"`
	Shorts map[string]string `json:"shorts"`
}
