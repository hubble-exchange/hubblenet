package limitorders

import (
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/utils"
	"github.com/ethereum/go-ethereum/common"
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
	Placed    = "placed"
	Filled    = "filled"
	Cancelled = "cancelled"
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
	// InProgressBaseAssetQuantity map[common.Hash]*big.Int `json:"in_progress_base_asset_quantity"` // block hash => quantity
	RawOrder interface{} `json:"-"`
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
	Add(order *LimitOrder)
	Delete(id string)
	UpdateFilledBaseAssetQuantity(quantity *big.Int, orderId string)
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
	GetInProgressBlocks() []*types.Block
	UpdateInProgressState(block *types.Block, quantityMap map[string]*big.Int)
	RemoveInProgressState(block *types.Block, quantityMap map[string]*big.Int)
	Accept(block *types.Block)
}

type StateSnapshot struct {
	OrderMap        map[string]*LimitOrder     `json:"order_map"`  // ID => order
	TraderMap       map[common.Address]*Trader `json:"trader_map"` // address => trader info
	NextFundingTime uint64                     `json:"next_funding_time"`
	LastPrice       map[Market]*big.Int        `json:"last_price"`
}

type InMemoryDatabase struct {
	mu                     sync.Mutex `json:"-"`
	stateSnapshot          map[common.Hash]*StateSnapshot
	acceptedHeadBlockHash  common.Hash
	preferredHeadBlockHash common.Hash
	// InProgressBlocks []*types.Block `json:"in_progress_blocks"` // block number => block hash
	// InProgressBlockOrders  map[common.Hash][]string   `json:"in_progress_block_orders"`  // block hash => list of order ids
}

func NewInMemoryDatabase() *InMemoryDatabase {
	orderMap := map[string]*LimitOrder{}
	lastPrice := map[Market]*big.Int{AvaxPerp: big.NewInt(0)}
	traderMap := map[common.Address]*Trader{}
	// inProgressBlockOrders := map[common.Hash][]string{}
	inProgressBlockNumbers := []*types.Block{}

	return &InMemoryDatabase{
		OrderMap:        orderMap,
		TraderMap:       traderMap,
		NextFundingTime: 0,
		LastPrice:       lastPrice,
		// InProgressBlockOrders: inProgressBlockOrders,
		InProgressBlocks: inProgressBlockNumbers,
	}
}

func (db *InMemoryDatabase) Accept(block *types.Block) {
	if db.acceptedHeadBlockHash != block.Header().ParentHash {
		// @todo throw error
	}
	delete(db.stateSnapshot, db.acceptedHeadBlockHash)
	db.acceptedHeadBlockHash = block.Hash()
}

func (db *InMemoryDatabase) Discover(block *types.Block) {
	block.
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

func (db *InMemoryDatabase) Add(order *LimitOrder) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.OrderMap[getIdFromLimitOrder(*order)] = order
}

func (db *InMemoryDatabase) Delete(orderId string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	deleteOrder(db, orderId)
}

func (db *InMemoryDatabase) UpdateFilledBaseAssetQuantity(quantity *big.Int, orderId string) {
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
		deleteOrder(db, orderId)
	}
}

func (db *InMemoryDatabase) GetNextFundingTime() uint64 {
	return db.NextFundingTime
}

func (db *InMemoryDatabase) UpdateNextFundingTime(nextFundingTime uint64) {
	db.NextFundingTime = nextFundingTime
}

func (db *InMemoryDatabase) getPreferredHeadStateSnapshot() *StateSnapshot {
	return db.stateSnapshot[db.preferredHeadBlockHash]
}

func (db *InMemoryDatabase) GetLongOrders(market Market) []LimitOrder {
	var longOrders []LimitOrder
	var snapshot = db.getPreferredHeadStateSnapshot()
	for _, order := range snapshot.OrderMap {
		if order.PositionType == "long" && order.Market == market {
			longOrders = append(longOrders, *order)
		}
	}
	sortLongOrders(longOrders)
	return longOrders
}

func (db *InMemoryDatabase) GetShortOrders(market Market) []LimitOrder {
	var shortOrders []LimitOrder
	var snapshot = db.getPreferredHeadStateSnapshot()
	for _, order := range snapshot.OrderMap {
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
func (db *InMemoryDatabase) GetInProgressBlocks() []*types.Block {
	// sort descending
	sort.Slice(db.InProgressBlocks, func(i, j int) bool {
		return db.InProgressBlocks[i].NumberU64() > db.InProgressBlocks[j].NumberU64()
	})
	return db.InProgressBlocks
}

func (db *InMemoryDatabase) UpdateInProgressState(block *types.Block, quantityMap map[string]*big.Int) {
	db.mu.Lock()
	defer db.mu.Unlock()

	log.Info("#### updating state for block", "number", block.NumberU64(), "hash", block.Hash().String(), "orders", len(quantityMap))
	for orderId, quantity := range quantityMap {
		if _, ok := db.OrderMap[orderId]; !ok {
			log.Info("#### order id not found; ignoring", "orderId", orderId, "block", block.Hash().String())
			continue
		}

		log.Info("#### setting InProgress in order", "orderId", orderId, "block", block.Hash().String(), "quantity", quantity.Uint64())
		db.OrderMap[orderId].InProgressBaseAssetQuantity[block.Hash()] = quantity
	}

	db.InProgressBlocks = append(db.InProgressBlocks, block)

	traderMap := map[common.Address]Trader{}
	for address, trader := range db.TraderMap {
		traderMap[address] = *trader
	}
}

func (db *InMemoryDatabase) RemoveInProgressState(block *types.Block, quantityMap map[string]*big.Int) {
	db.mu.Lock()
	defer db.mu.Unlock()

	log.Info("#### removing state for block", "number", block.NumberU64(), "hash", block.Hash().String(), "orders", len(quantityMap))
	for orderId, quantity := range quantityMap {
		if _, ok := db.OrderMap[orderId]; !ok {
			log.Info("#### order id not found; ignoring", "orderId", orderId, "block", block.Hash().String())
			continue
		}

		log.Info("#### removing InProgress in order", "orderId", orderId, "block", block.Hash().String(), "quantity", quantity.Uint64())
		delete(db.OrderMap[orderId].InProgressBaseAssetQuantity, block.Hash())
	}

	db.InProgressBlocks = deleteBlockFromSlice(db.InProgressBlocks, block)
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

func deleteOrder(db *InMemoryDatabase, id string) {
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

func getIdFromLimitOrder(order LimitOrder) string {
	return order.UserAddress + order.Salt.String()
}
