package limitorders

import (
	"math"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Market int

const (
	AvaxPerp Market = iota
	EthPerp
)

func GetActiveMarkets() []Market {
	return []Market{AvaxPerp}
}

type Collateral int

const (
	USDC Collateral = iota
	Avax
	Eth
)

var collateralWeightMap map[Collateral]float64 = map[Collateral]float64{USDC: 1, Avax: 0.8, Eth: 0.8}

type Status string

const (
	Open        = "open"
	Unfulfilled = "unfulfilled"
	Fulfilled   = "fulfilled"
	Cancelled   = "cancelled"
)

type LimitOrder struct {
	Market
	id                      uint64
	PositionType            string
	UserAddress             string
	BaseAssetQuantity       int
	FilledBaseAssetQuantity int
	Price                   float64
	Status                  Status
	Salt                    string
	Signature               []byte
	RawOrder                interface{}
	RawSignature            interface{}
	BlockNumber             uint64
	Locked                  bool
}

type Position struct {
	OpenNotional        float64
	Size                float64
	UnrealisedFunding   float64
	LastPremiumFraction float64
}

type Trader struct {
	Positions   map[Market]Position    // position for every market
	Margins     map[Collateral]float64 // available margin/balance for every market
	BlockNumber uint32
}

type LimitOrderDatabase interface {
	GetAllOrders() []LimitOrder
	Add(order *LimitOrder)
	UpdateFilledBaseAssetQuantity(quantity uint, signature []byte)
	GetLongOrders(market Market) []LimitOrder
	GetShortOrders(market Market) []LimitOrder
	UpdatePosition(trader common.Address, market Market, size float64, openNotional float64)
	UpdateMargin(trader common.Address, collateral Collateral, addAmount float64)
	UpdateUnrealisedFunding(market Market, cumulativePremiumFraction float64)
	ResetUnrealisedFunding(market Market, trader common.Address, cumulativePremiumFraction float64)
	UpdateNextFundingTime()
	GetNextFundingTime() uint64
	GetLiquidableTraders(market Market, markPrice float64, oraclePrice float64) []Liquidable
	UpdateLastPrice(market Market, lastPrice float64)
	GetLastPrice(market Market) float64
}

type InMemoryDatabase struct {
	orderMap        map[string]*LimitOrder     // signature => order
	traderMap       map[common.Address]*Trader // address => trader info
	nextFundingTime uint64
	lastPrice       map[Market]float64
}

func NewInMemoryDatabase() *InMemoryDatabase {
	orderMap := map[string]*LimitOrder{}
	lastPrice := map[Market]float64{}
	traderMap := map[common.Address]*Trader{}
	nextFundingTime := uint64(getNextHour().Unix())

	return &InMemoryDatabase{
		orderMap:        orderMap,
		traderMap:       traderMap,
		nextFundingTime: nextFundingTime,
		lastPrice:       lastPrice,
	}
}

func (db *InMemoryDatabase) GetAllOrders() []LimitOrder {
	allOrders := []LimitOrder{}
	for _, order := range db.orderMap {
		allOrders = append(allOrders, *order)
	}
	return allOrders
}

func (db *InMemoryDatabase) Add(order *LimitOrder) {
	db.orderMap[string(order.Signature)] = order
}

func (db *InMemoryDatabase) UpdateFilledBaseAssetQuantity(quantity uint, signature []byte) {
	limitOrder := db.orderMap[string(signature)]
	if uint(math.Abs(float64(limitOrder.BaseAssetQuantity))) == quantity {
		deleteOrder(db, signature)
		return
	} else {
		if limitOrder.PositionType == "long" {
			limitOrder.FilledBaseAssetQuantity = int(quantity)
		}
		if limitOrder.PositionType == "short" {
			limitOrder.FilledBaseAssetQuantity = -int(quantity)
		}
	}
}

func (db *InMemoryDatabase) GetNextFundingTime() uint64 {
	return db.nextFundingTime
}

func (db *InMemoryDatabase) UpdateNextFundingTime() {
	db.nextFundingTime = uint64(getNextHour().Unix())
}

func (db *InMemoryDatabase) GetLongOrders(market Market) []LimitOrder {
	var longOrders []LimitOrder
	for _, order := range db.orderMap {
		if order.PositionType == "long" && order.Market == market {
			longOrders = append(longOrders, *order)
		}
	}
	sortLongOrders(longOrders)
	return longOrders
}

func (db *InMemoryDatabase) GetShortOrders(market Market) []LimitOrder {
	var shortOrders []LimitOrder
	for _, order := range db.orderMap {
		if order.PositionType == "short" && order.Market == market {
			shortOrders = append(shortOrders, *order)
		}
	}
	sortShortOrders(shortOrders)
	return shortOrders
}

func (db *InMemoryDatabase) UpdateMargin(trader common.Address, collateral Collateral, addAmount float64) {
	if _, ok := db.traderMap[trader]; !ok {
		db.traderMap[trader] = &Trader{
			Positions: map[Market]Position{},
			Margins:   map[Collateral]float64{},
		}
	}

	if _, ok := db.traderMap[trader].Margins[collateral]; !ok {
		db.traderMap[trader].Margins[collateral] = 0
	}

	db.traderMap[trader].Margins[collateral] += addAmount
}

func (db *InMemoryDatabase) UpdatePosition(trader common.Address, market Market, size float64, openNotional float64) {
	if _, ok := db.traderMap[trader]; !ok {
		db.traderMap[trader] = &Trader{
			Positions: map[Market]Position{},
			Margins:   map[Collateral]float64{},
		}
	}

	if _, ok := db.traderMap[trader].Positions[market]; !ok {
		db.traderMap[trader].Positions[market] = Position{}
	}

	position := db.traderMap[trader].Positions[market]

	position.Size = size
	position.OpenNotional = openNotional
	db.traderMap[trader].Positions[market] = position
}

func (db *InMemoryDatabase) UpdateUnrealisedFunding(market Market, cumulativePremiumFraction float64) {
	for addr, trader := range db.traderMap {
		position := trader.Positions[market]
		position.UnrealisedFunding = (cumulativePremiumFraction - position.LastPremiumFraction) * position.Size
		db.traderMap[addr].Positions[market] = position
	}
}

func (db *InMemoryDatabase) ResetUnrealisedFunding(market Market, trader common.Address, cumulativePremiumFraction float64) {
	if db.traderMap[trader] != nil {
		if _, ok := db.traderMap[trader].Positions[market]; ok {
			position := db.traderMap[trader].Positions[market]
			position.UnrealisedFunding = 0
			position.LastPremiumFraction = cumulativePremiumFraction
			db.traderMap[trader].Positions[market] = position
				}
	}
}

func (db *InMemoryDatabase) GetAllTraders() map[common.Address]*Trader {
	return db.traderMap
}

func (db *InMemoryDatabase) UpdateLastPrice(market Market, lastPrice float64) {
	db.lastPrice[market] = lastPrice
}

func (db *InMemoryDatabase) GetLastPrice(market Market) float64 {
	return db.lastPrice[market]
}

func (trader *Trader) GetNormalisedMargin() float64 {
	return trader.Margins[USDC]

	// this will change after multi collateral
	// var normalisedMargin float64
	// for coll, margin := range trader.Margins {
	// 	normalisedMargin += margin * priceMap[coll] * collateralWeightMap[coll]
	// }

	// return normalisedMargin
}

func sortLongOrders(orders []LimitOrder) []LimitOrder {
	sort.SliceStable(orders, func(i, j int) bool {
		if orders[i].Price > orders[j].Price {
			return true
		}
		if orders[i].Price == orders[j].Price {
			if orders[i].BlockNumber < orders[j].BlockNumber {
				return true
			}
		}
		return false
	})
	return orders
}

func sortShortOrders(orders []LimitOrder) []LimitOrder {
	sort.SliceStable(orders, func(i, j int) bool {
		if orders[i].Price < orders[j].Price {
			return true
		}
		if orders[i].Price == orders[j].Price {
			if orders[i].BlockNumber < orders[j].BlockNumber {
				return true
			}
		}
		return false
	})
	return orders
}

func getNextHour() time.Time {
	now := time.Now().UTC()
	nextHour := now.Round(time.Hour)
	if time.Since(nextHour) >= 0 {
		nextHour = nextHour.Add(time.Hour)
	}
	return nextHour
}

func deleteOrder(db *InMemoryDatabase, signature []byte) {
	delete(db.orderMap, string(signature))
}
