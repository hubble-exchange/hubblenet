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
	id                uint64
	PositionType      string
	UserAddress       string
	BaseAssetQuantity int
	FilledBaseAssetQuantity int
	Price             float64
	Status            Status
	Salt              string
	Signature    []byte
	RawOrder     interface{}
	RawSignature interface{}
	BlockNumber  uint64
}

type Position struct {
	OpenNotional      float64
	Size              float64
	UnrealisedFunding float64
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
	Delete(signature []byte)
	GetLongOrders() []LimitOrder
	GetShortOrders() []LimitOrder
}

type InMemoryDatabase struct {
	orderMap        map[string]*LimitOrder     // signature => order
	traderMap       map[common.Address]*Trader // address => trader info
	nextFundingTime uint64
}

func NewInMemoryDatabase() *InMemoryDatabase {
	orderMap := map[string]*LimitOrder{}

	return &InMemoryDatabase{
		orderMap:        orderMap,
		traderMap:       map[common.Address]*Trader{},
		nextFundingTime: uint64(getNextHour().Unix()),
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

func (db *InMemoryDatabase) Delete(signature []byte) {
	deleteOrder(db, signature)
}

func (db *InMemoryDatabase) GetLongOrders() []LimitOrder {
	var longOrders []LimitOrder
	for _, order := range db.orderMap {
		if order.PositionType == "long" {
			longOrders = append(longOrders, *order)
		}
	}
	sortLongOrders(longOrders)
	return longOrders
}

func (db *InMemoryDatabase) GetShortOrders() []LimitOrder {
	var shortOrders []LimitOrder
	for _, order := range db.orderMap {
		if order.PositionType == "short" {
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

func (db *InMemoryDatabase) UpdatePositionForOrder(signature string, fillAmount float64) {
	order, ok := db.orderMap[signature]
	if !ok {

	}

	trader := common.HexToAddress(order.UserAddress)
	market := order.Market
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

	// increase/decrease depending on the order's original amount as fillAmount may be always positive
	positionSignBit := math.Signbit(float64(order.BaseAssetQuantity)) // returns true if qty is negative
	var positionSign float64
	if positionSignBit {
		positionSign = -1
	} else {
		positionSign = 1
	}
	position.Size += math.Abs(fillAmount) * positionSign
	db.traderMap[trader].Positions[market] = position

	// @todo: update notional position too
}

func (db *InMemoryDatabase) UpdateUnrealisedFunding(market Market, fundingRate float64) {
	for addr, trader := range db.traderMap {
		position := trader.Positions[market]
		newFunding := position.Size * fundingRate
		position.UnrealisedFunding = trader.Positions[market].UnrealisedFunding + newFunding
		db.traderMap[addr].Positions[market] = position
	}
}

func (db *InMemoryDatabase) ResetUnrealisedFunding(market Market, trader common.Address) {
	position := db.traderMap[trader].Positions[market]
	position.UnrealisedFunding = 0
	db.traderMap[trader].Positions[market] = position
}

func (db *InMemoryDatabase) GetAllTraders() map[common.Address]*Trader {
	return db.traderMap
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
