package limitorders

import (
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

var maxLiquidationRatio *big.Int = big.NewInt(25 * 10e4)
var minSizeRequirement *big.Int = big.NewInt(0).Mul(big.NewInt(5), big.NewInt(1e18))

type Market int

const (
	AvaxPerp Market = iota
)

func GetActiveMarkets() []Market {
	return []Market{AvaxPerp}
}

type Collateral int

const (
	HUSD Collateral = iota
)

var collateralWeightMap map[Collateral]float64 = map[Collateral]float64{HUSD: 1}

type Status string

const (
	Placed Status = "placed"
	// Deleted   = "deleted"
	Filled    Status = "filled"
	Cancelled Status = "cancelled"
)

type LimitOrder struct {
	Id     uint64 `json:"id"`
	Market Market `json:"market"`
	// @todo make this an enum
	PositionType            string   `json:"position_type"`
	UserAddress             string   `json:"user_address"`
	BaseAssetQuantity       *big.Int `json:"base_asset_quantity"`
	FilledBaseAssetQuantity *big.Int `json:"filled_base_asset_quantity"`
	Salt                    *big.Int `json:"salt"`
	Price                   *big.Int `json:"price"`
	Status                  Status   `json:"status"`
	Signature               []byte   `json:"signature"`
	BlockNumber             *big.Int `json:"block_number"` // block number order was placed on
	FulfilmentBlock         uint64
	RawOrder                interface{} `json:"-"`
}

func (order LimitOrder) GetUnFilledBaseAssetQuantity() *big.Int {
	return big.NewInt(0).Sub(order.BaseAssetQuantity, order.FilledBaseAssetQuantity)
}

type Position struct {
	OpenNotional         *big.Int `json:"open_notional"`
	Size                 *big.Int `json:"size"`
	UnrealisedFunding    *big.Int `json:"unrealised_funding"`
	LastPremiumFraction  *big.Int `json:"last_premium_fraction"`
	LiquidationThreshold *big.Int `json:"liquidation_threshold"`
}

type Trader struct {
	Positions   map[Market]*Position    `json:"positions"` // position for every market
	Margins     map[Collateral]*big.Int `json:"margins"`   // available margin/balance for every market
	BlockNumber *big.Int                `json:"block_number"`
}

type LimitOrderDatabase interface {
	GetAllOrders() []LimitOrder
	Add(orderId common.Hash, order *LimitOrder)
	Delete(orderId common.Hash)
	UpdateFilledBaseAssetQuantity(quantity *big.Int, orderId common.Hash, blockNumber uint64)
	GetLongOrders(market Market) []LimitOrder
	GetShortOrders(market Market) []LimitOrder
	UpdatePosition(trader common.Address, market Market, size *big.Int, openNotional *big.Int, isLiquidation bool)
	UpdateMargin(trader common.Address, collateral Collateral, addAmount *big.Int)
	UpdateUnrealisedFunding(market Market, cumulativePremiumFraction *big.Int)
	ResetUnrealisedFunding(market Market, trader common.Address, cumulativePremiumFraction *big.Int)
	UpdateNextFundingTime(nextFundingTime uint64)
	GetNextFundingTime() uint64
	UpdateLastPrice(market Market, lastPrice *big.Int)
	GetLastPrice(market Market) *big.Int
	GetAllTraders() map[common.Address]Trader
	GetOrderBookData() InMemoryDatabase
	Accept(blockNumber uint64)
}

type InMemoryDatabase struct {
	mu              sync.Mutex                  `json:"-"`
	OrderMap        map[common.Hash]*LimitOrder `json:"order_map"`  // ID => order
	TraderMap       map[common.Address]*Trader  `json:"trader_map"` // address => trader info
	NextFundingTime uint64                      `json:"next_funding_time"`
	LastPrice       map[Market]*big.Int         `json:"last_price"`
}

func NewInMemoryDatabase() *InMemoryDatabase {
	orderMap := map[common.Hash]*LimitOrder{}
	lastPrice := map[Market]*big.Int{AvaxPerp: big.NewInt(0)}
	traderMap := map[common.Address]*Trader{}

	return &InMemoryDatabase{
		OrderMap:        orderMap,
		TraderMap:       traderMap,
		NextFundingTime: 0,
		LastPrice:       lastPrice,
	}
}

func (db *InMemoryDatabase) Accept(blockNumber uint64) {
	for orderId, order := range db.OrderMap {
		if order.FulfilmentBlock > 0 && order.FulfilmentBlock <= blockNumber {
			delete(db.OrderMap, orderId)
		}
	}
}

func (db *InMemoryDatabase) GetAllOrders() []LimitOrder {
	db.mu.Lock()
	defer db.mu.Unlock()

	allOrders := []LimitOrder{}
	for _, order := range db.OrderMap {
		allOrders = append(allOrders, *order)
	}
	return allOrders
}

func (db *InMemoryDatabase) Add(orderId common.Hash, order *LimitOrder) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.OrderMap[orderId] = order
}

func (db *InMemoryDatabase) Delete(orderId common.Hash) {
	db.mu.Lock()
	defer db.mu.Unlock()

	delete(db.OrderMap, orderId)
}

func (db *InMemoryDatabase) UpdateFilledBaseAssetQuantity(quantity *big.Int, orderId common.Hash, block uint64) {
	db.mu.Lock()
	defer db.mu.Unlock()

	limitOrder := db.OrderMap[orderId]
	if limitOrder.PositionType == "long" {
		limitOrder.FilledBaseAssetQuantity.Add(limitOrder.FilledBaseAssetQuantity, quantity) // filled = filled + quantity
	}
	if limitOrder.PositionType == "short" {
		limitOrder.FilledBaseAssetQuantity.Sub(limitOrder.FilledBaseAssetQuantity, quantity) // filled = filled - quantity
	}

	if limitOrder.BaseAssetQuantity.Cmp(limitOrder.FilledBaseAssetQuantity) == 0 {
		limitOrder.FulfilmentBlock = block
	}
}

func (db *InMemoryDatabase) GetNextFundingTime() uint64 {
	return db.NextFundingTime
}

func (db *InMemoryDatabase) UpdateNextFundingTime(nextFundingTime uint64) {
	db.NextFundingTime = nextFundingTime
}

func (db *InMemoryDatabase) GetLongOrders(market Market) []LimitOrder {
	var longOrders []LimitOrder
	for _, order := range db.OrderMap {
		if order.PositionType == "long" && order.Market == market {
			longOrders = append(longOrders, *order)
		}
	}
	sortLongOrders(longOrders)
	return longOrders
}

func (db *InMemoryDatabase) GetShortOrders(market Market) []LimitOrder {
	var shortOrders []LimitOrder
	for _, order := range db.OrderMap {
		if order.PositionType == "short" && order.Market == market {
			shortOrders = append(shortOrders, *order)
		}
	}
	sortShortOrders(shortOrders)
	return shortOrders
}

func (db *InMemoryDatabase) UpdateMargin(trader common.Address, collateral Collateral, addAmount *big.Int) {
	if _, ok := db.TraderMap[trader]; !ok {
		db.TraderMap[trader] = &Trader{
			Positions: map[Market]*Position{},
			Margins:   map[Collateral]*big.Int{},
		}
	}

	if _, ok := db.TraderMap[trader].Margins[collateral]; !ok {
		db.TraderMap[trader].Margins[collateral] = big.NewInt(0)
	}

	db.TraderMap[trader].Margins[collateral].Add(db.TraderMap[trader].Margins[collateral], addAmount)
}

func (db *InMemoryDatabase) UpdatePosition(trader common.Address, market Market, size *big.Int, openNotional *big.Int, isLiquidation bool) {
	if _, ok := db.TraderMap[trader]; !ok {
		db.TraderMap[trader] = &Trader{
			Positions: map[Market]*Position{},
			Margins:   map[Collateral]*big.Int{},
		}
	}

	if _, ok := db.TraderMap[trader].Positions[market]; !ok {
		db.TraderMap[trader].Positions[market] = &Position{}
	}

	db.TraderMap[trader].Positions[market].Size = size
	db.TraderMap[trader].Positions[market].OpenNotional = openNotional
	db.TraderMap[trader].Positions[market].LastPremiumFraction = big.NewInt(0)

	if !isLiquidation {
		db.TraderMap[trader].Positions[market].LiquidationThreshold = getLiquidationThreshold(size)
	}
}

func (db *InMemoryDatabase) UpdateUnrealisedFunding(market Market, cumulativePremiumFraction *big.Int) {
	for _, trader := range db.TraderMap {
		position := trader.Positions[market]
		if position != nil {
			position.UnrealisedFunding = dividePrecisionSize(big.NewInt(0).Mul(big.NewInt(0).Sub(cumulativePremiumFraction, position.LastPremiumFraction), position.Size))
		}
	}
}

func (db *InMemoryDatabase) ResetUnrealisedFunding(market Market, trader common.Address, cumulativePremiumFraction *big.Int) {
	if db.TraderMap[trader] != nil {
		if _, ok := db.TraderMap[trader].Positions[market]; ok {
			db.TraderMap[trader].Positions[market].UnrealisedFunding = big.NewInt(0)
			db.TraderMap[trader].Positions[market].LastPremiumFraction = cumulativePremiumFraction
		}
	}
}

func (db *InMemoryDatabase) UpdateLastPrice(market Market, lastPrice *big.Int) {
	db.LastPrice[market] = lastPrice
}

func (db *InMemoryDatabase) GetLastPrice(market Market) *big.Int {
	return db.LastPrice[market]
}
func (db *InMemoryDatabase) GetAllTraders() map[common.Address]Trader {
	traderMap := map[common.Address]Trader{}
	for address, trader := range db.TraderMap {
		traderMap[address] = *trader
	}
	return traderMap
}

func deleteBlockFromSlice(slice []*types.Block, block *types.Block) []*types.Block {
	for i, other := range slice {
		if other.Hash() == block.Hash() {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func sortLongOrders(orders []LimitOrder) []LimitOrder {
	sort.SliceStable(orders, func(i, j int) bool {
		if orders[i].Price.Cmp(orders[j].Price) == 1 {
			return true
		}
		if orders[i].Price.Cmp(orders[j].Price) == 0 {
			if orders[i].BlockNumber.Cmp(orders[j].BlockNumber) == -1 {
				return true
			}
		}
		return false
	})
	return orders
}

func sortShortOrders(orders []LimitOrder) []LimitOrder {
	sort.SliceStable(orders, func(i, j int) bool {
		if orders[i].Price.Cmp(orders[j].Price) == -1 {
			return true
		}
		if orders[i].Price.Cmp(orders[j].Price) == 0 {
			if orders[i].BlockNumber.Cmp(orders[j].BlockNumber) == -1 {
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

func deleteOrder(db *InMemoryDatabase, id common.Hash) {
	log.Info("#### deleting order", "orderId", id)
	delete(db.OrderMap, id)
}

func (db *InMemoryDatabase) GetOrderBookData() InMemoryDatabase {
	return *db
}

func getLiquidationThreshold(size *big.Int) *big.Int {
	absSize := big.NewInt(0).Abs(size)
	maxLiquidationSize := divideByBasePrecision(big.NewInt(0).Mul(absSize, maxLiquidationRatio))
	threshold := big.NewInt(0).Add(maxLiquidationSize, big.NewInt(1))
	liquidationThreshold := utils.BigIntMax(threshold, minSizeRequirement)
	return big.NewInt(0).Mul(liquidationThreshold, big.NewInt(int64(size.Sign()))) // same sign as size
}

// @todo change this to return the EIP712 hash instead
func getIdFromLimitOrder(order LimitOrder) common.Hash {
	return crypto.Keccak256Hash([]byte(order.UserAddress + order.Salt.String()))
}
