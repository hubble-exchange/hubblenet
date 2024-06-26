package orderbook

import (
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/ava-labs/subnet-evm/core/types"
	hu "github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
	"github.com/ava-labs/subnet-evm/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

const (
	// ticker frequency for calling signalTxsReady
	matchingTickerDuration = 5 * time.Second
	sanitaryTickerDuration = 1 * time.Second
)

type MatchingPipeline struct {
	mu             sync.Mutex
	db             LimitOrderDatabase
	lotp           LimitOrderTxProcessor
	configService  IConfigService
	MatchingTicker *time.Ticker
	SanitaryTicker *time.Ticker
}

func NewMatchingPipeline(
	db LimitOrderDatabase,
	lotp LimitOrderTxProcessor,
	configService IConfigService) *MatchingPipeline {

	return &MatchingPipeline{
		db:             db,
		lotp:           lotp,
		configService:  configService,
		MatchingTicker: time.NewTicker(matchingTickerDuration),
		SanitaryTicker: time.NewTicker(sanitaryTickerDuration),
	}
}

func NewTemporaryMatchingPipeline(
	db LimitOrderDatabase,
	lotp LimitOrderTxProcessor,
	configService IConfigService) *MatchingPipeline {

	return &MatchingPipeline{
		db:            db,
		lotp:          lotp,
		configService: configService,
	}
}

func (pipeline *MatchingPipeline) RunSanitization() {
	pipeline.db.RemoveExpiredSignedOrders()
}

func (pipeline *MatchingPipeline) Run(blockNumber *big.Int) bool {
	pipeline.mu.Lock()
	defer pipeline.mu.Unlock()

	// reset ticker
	pipeline.MatchingTicker.Reset(matchingTickerDuration)
	// SUNSET: this is ok, we can skip matching, liquidation, settleFunding, commitSampleLiquidity when markets are settled
	markets := pipeline.GetActiveMarkets()
	log.Info("MatchingPipeline:Run", "blockNumber", blockNumber)

	if len(markets) == 0 {
		return false
	}

	// start fresh and purge all local transactions
	pipeline.lotp.PurgeOrderBookTxs()

	if isFundingPaymentTime(pipeline.db.GetNextFundingTime()) {
		log.Info("MatchingPipeline:isFundingPaymentTime")
		err := executeFundingPayment(pipeline.lotp)
		if err != nil {
			log.Error("Funding payment job failed", "err", err)
		}
	}

	// check nextSamplePITime
	if isSamplePITime(pipeline.db.GetNextSamplePITime(), pipeline.db.GetSamplePIAttemptedTime()) {
		log.Info("MatchingPipeline:isSamplePITime")
		err := pipeline.lotp.ExecuteSamplePITx()
		if err != nil {
			log.Error("Sample PI job failed", "err", err)
		}
	}

	// fetch various hubble market params and run the matching engine
	hState := GetHubbleState(pipeline.configService)

	// build trader map
	liquidablePositions, ordersToCancel, marginMap := pipeline.db.GetNaughtyTraders(hState)
	cancellableOrderIds := pipeline.cancelLimitOrders(ordersToCancel)
	orderMap := make(map[Market]*Orders)
	for _, market := range markets {
		orderMap[market] = pipeline.fetchOrders(market, hState.OraclePrices[market], cancellableOrderIds, blockNumber)
	}
	pipeline.runLiquidations(liquidablePositions, orderMap, hState.OraclePrices, marginMap)
	for _, market := range markets {
		// @todo should we prioritize matching in any particular market?
		upperBound, _ := pipeline.configService.GetAcceptableBounds(market)
		pipeline.runMatchingEngine(pipeline.lotp, orderMap[market].longOrders, orderMap[market].shortOrders, marginMap, hState.MinAllowableMargin, hState.TakerFee, upperBound)
	}

	orderBookTxsCount := pipeline.lotp.GetOrderBookTxsCount()
	log.Info("MatchingPipeline:Complete", "orderBookTxsCount", orderBookTxsCount)
	if orderBookTxsCount > 0 {
		pipeline.lotp.SetOrderBookTxsBlockNumber(blockNumber.Uint64())
		return true
	}
	return false
}

func (pipeline *MatchingPipeline) GetOrderMatchingTransactions(blockNumber *big.Int, markets []Market) map[common.Address]types.Transactions {
	pipeline.mu.Lock()
	defer pipeline.mu.Unlock()

	// SUNSET: ok to skip when markets are settled
	activeMarkets := pipeline.GetActiveMarkets()
	log.Info("MatchingPipeline:GetOrderMatchingTransactions")

	if len(activeMarkets) == 0 {
		return nil
	}

	// start fresh and purge all local transactions
	pipeline.lotp.PurgeOrderBookTxs()

	// fetch various hubble market params and run the matching engine
	hState := GetHubbleState(pipeline.configService)
	hState.OraclePrices = hu.ArrayToMap(pipeline.configService.GetUnderlyingPrices())

	marginMap := make(map[common.Address]*big.Int)
	for addr, trader := range pipeline.db.GetAllTraders() {
		userState := &hu.UserState{
			Positions:      translatePositions(trader.Positions),
			Margins:        getMargins(&trader, len(hState.Assets)),
			PendingFunding: getTotalFunding(&trader, hState.ActiveMarkets),
			ReservedMargin: new(big.Int).Set(trader.Margin.Reserved),
			// this is the only leveldb read, others above are in-memory reads
			ReduceOnlyAmounts: pipeline.configService.GetReduceOnlyAmounts(addr),
		}
		marginMap[addr] = hu.GetAvailableMargin(hState, userState)
	}
	for _, market := range markets {
		orders := pipeline.fetchOrders(market, hState.OraclePrices[market], map[common.Hash]struct{}{}, blockNumber)
		upperBound, _ := pipeline.configService.GetAcceptableBounds(market)
		pipeline.runMatchingEngine(pipeline.lotp, orders.longOrders, orders.shortOrders, marginMap, hState.MinAllowableMargin, hState.TakerFee, upperBound)
	}

	orderbookTxs := pipeline.lotp.GetOrderBookTxs()
	pipeline.lotp.PurgeOrderBookTxs()
	return orderbookTxs
}

type Orders struct {
	longOrders  []Order
	shortOrders []Order
}

func (pipeline *MatchingPipeline) GetActiveMarkets() []Market {
	count := pipeline.configService.GetActiveMarketsCount()
	markets := make([]Market, count)
	for i := int64(0); i < count; i++ {
		markets[i] = Market(i)
	}
	return markets
}

func (pipeline *MatchingPipeline) GetCollaterals() []hu.Collateral {
	return pipeline.configService.GetCollaterals()
}

func (pipeline *MatchingPipeline) cancelLimitOrders(cancellableOrders map[common.Address][]Order) map[common.Hash]struct{} {
	cancellableOrderIds := map[common.Hash]struct{}{}
	// @todo: if there are too many cancellable orders, they might not fit in a single block. Need to adjust for that.
	for _, orders := range cancellableOrders {
		if len(orders) == 0 {
			continue
		}
		rawOrders := make([]LimitOrder, 0)
		for _, order := range orders {
			rawOrders = append(rawOrders, *order.RawOrder.(*LimitOrder))
			cancellableOrderIds[order.Id] = struct{}{} // do not attempt to match these orders
		}

		log.Info("orders to cancel", "num", len(orders))
		// cancel max of 5 orders. change this if the tx gas limit (1.5m) is changed
		if err := pipeline.lotp.ExecuteLimitOrderCancel(rawOrders[0:int(math.Min(float64(len(rawOrders)), 5))]); err != nil {
			log.Error("Error in ExecuteOrderCancel", "orders", orders, "err", err)
		}
	}
	return cancellableOrderIds
}

func (pipeline *MatchingPipeline) fetchOrders(market Market, underlyingPrice *big.Int, cancellableOrderIds map[common.Hash]struct{}, blockNumber *big.Int) *Orders {
	_, lowerBoundForLongs := pipeline.configService.GetAcceptableBounds(market)
	// any long orders below the permissible lowerbound are irrelevant, because they won't be matched no matter what.
	// this assumes that all above cancelOrder transactions got executed successfully (or atleast they are not meant to be executed anyway if they passed the cancellation criteria)
	longOrders := removeOrdersWithIds(pipeline.db.GetLongOrders(market, lowerBoundForLongs, blockNumber), cancellableOrderIds)

	// say if there were no long orders, then shord orders above liquidation upper bound are irrelevant, because they won't be matched no matter what
	// note that this assumes that permissible liquidation spread <= oracle spread
	upperBoundforShorts, _ := pipeline.configService.GetAcceptableBoundsForLiquidation(market)

	// however, if long orders exist, then
	if len(longOrders) != 0 {
		oracleUpperBound, _ := pipeline.configService.GetAcceptableBounds(market)
		// take the max of price of highest long and liq upper bound. But say longOrders[0].Price > oracleUpperBound ? - then we discard orders above oracleUpperBound, because they won't be matched no matter what
		upperBoundforShorts = utils.BigIntMin(utils.BigIntMax(longOrders[0].Price, upperBoundforShorts), oracleUpperBound)
	}
	shortOrders := removeOrdersWithIds(pipeline.db.GetShortOrders(market, upperBoundforShorts, blockNumber), cancellableOrderIds)
	return &Orders{longOrders, shortOrders}
}

func (pipeline *MatchingPipeline) runLiquidations(liquidablePositions []LiquidablePosition, orderMap map[Market]*Orders, underlyingPrices map[Market]*big.Int, marginMap map[common.Address]*big.Int) {
	if len(liquidablePositions) == 0 {
		return
	}

	log.Info("found positions to liquidate", "num", len(liquidablePositions))

	// we need to retreive permissible bounds for liquidations in each market
	// SUNSET: this is ok, we can skip liquidations when markets are settled
	markets := pipeline.GetActiveMarkets()
	type S struct {
		Upperbound *big.Int
		Lowerbound *big.Int
	}
	if len(markets) == 0 {
		return
	}
	liquidationBounds := make([]S, len(markets))
	for _, market := range markets {
		upperbound, lowerbound := pipeline.configService.GetAcceptableBoundsForLiquidation(market)
		liquidationBounds[market] = S{Upperbound: upperbound, Lowerbound: lowerbound}
	}

	minAllowableMargin := pipeline.configService.GetMinAllowableMargin()
	takerFee := pipeline.configService.GetTakerFee()
	for _, liquidable := range liquidablePositions {
		market := liquidable.Market
		numOrdersExhausted := 0
		switch liquidable.PositionType {
		case LONG:
			for _, order := range orderMap[market].longOrders {
				if order.Price.Cmp(liquidationBounds[market].Lowerbound) == -1 {
					// further orders are not not eligible to liquidate with
					break
				}
				fillAmount := utils.BigIntMinAbs(liquidable.GetUnfilledSize(), order.GetUnFilledBaseAssetQuantity())
				if marginMap[order.Trader] == nil {
					// compatibility with existing tests
					marginMap[order.Trader] = big.NewInt(0)
				}
				requiredMargin, err := isExecutable(&order, fillAmount, minAllowableMargin, takerFee, liquidationBounds[market].Upperbound, marginMap[order.Trader])
				if err != nil {
					log.Error("order is not executable", "order", order, "err", err)
					numOrdersExhausted++
					continue
				}
				marginMap[order.Trader].Sub(marginMap[order.Trader], requiredMargin) // deduct available margin for this run
				pipeline.lotp.ExecuteLiquidation(liquidable.Address, order, fillAmount)
				order.FilledBaseAssetQuantity.Add(order.FilledBaseAssetQuantity, fillAmount)
				liquidable.FilledSize.Add(liquidable.FilledSize, fillAmount)
				if order.GetUnFilledBaseAssetQuantity().Sign() == 0 {
					numOrdersExhausted++
				}
				if liquidable.GetUnfilledSize().Sign() == 0 {
					break // partial/full liquidation for this position slated for this run is complete
				}
			}
			orderMap[market].longOrders = orderMap[market].longOrders[numOrdersExhausted:]
		case SHORT:
			for _, order := range orderMap[market].shortOrders {
				if order.Price.Cmp(liquidationBounds[market].Upperbound) == 1 {
					// further orders are not not eligible to liquidate with
					break
				}
				fillAmount := utils.BigIntMinAbs(liquidable.GetUnfilledSize(), order.GetUnFilledBaseAssetQuantity())
				if marginMap[order.Trader] == nil {
					marginMap[order.Trader] = big.NewInt(0)
				}
				requiredMargin, err := isExecutable(&order, fillAmount, minAllowableMargin, takerFee, liquidationBounds[market].Upperbound, marginMap[order.Trader])
				if err != nil {
					log.Error("order is not executable", "order", order, "err", err)
					numOrdersExhausted++
					continue
				}
				marginMap[order.Trader].Sub(marginMap[order.Trader], requiredMargin) // deduct available margin for this run
				pipeline.lotp.ExecuteLiquidation(liquidable.Address, order, fillAmount)
				order.FilledBaseAssetQuantity.Sub(order.FilledBaseAssetQuantity, fillAmount)
				liquidable.FilledSize.Sub(liquidable.FilledSize, fillAmount)
				if order.GetUnFilledBaseAssetQuantity().Sign() == 0 {
					numOrdersExhausted++
				}
				if liquidable.GetUnfilledSize().Sign() == 0 {
					break // partial/full liquidation for this position slated for this run is complete
				}
			}
			orderMap[market].shortOrders = orderMap[market].shortOrders[numOrdersExhausted:]
		}
		if liquidable.GetUnfilledSize().Sign() != 0 {
			unquenchedLiquidationsCounter.Inc(1)
			log.Info("unquenched liquidation", "liquidable", liquidable)
		}
	}
}

func (pipeline *MatchingPipeline) runMatchingEngine(lotp LimitOrderTxProcessor, longOrders []Order, shortOrders []Order, marginMap map[common.Address]*big.Int, minAllowableMargin, takerFee, upperBound *big.Int) {
	for i := 0; i < len(longOrders); i++ {
		// if there are no short orders or if the price of the first long order is < the price of the first short order, then we can stop matching
		if len(shortOrders) == 0 || longOrders[i].Price.Cmp(shortOrders[0].Price) == -1 {
			break
		}
		numOrdersExhausted := 0
		for j := 0; j < len(shortOrders); j++ {
			fillAmount, err := areMatchingOrders(longOrders[i], shortOrders[j], marginMap, minAllowableMargin, takerFee, upperBound)
			if err != nil {
				log.Error("orders not matcheable", "longOrder", longOrders[i], "shortOrder", shortOrders[i], "err", err)
				continue
			}
			longOrders[i], shortOrders[j] = ExecuteMatchedOrders(lotp, longOrders[i], shortOrders[j], fillAmount)
			if shortOrders[j].GetUnFilledBaseAssetQuantity().Sign() == 0 {
				numOrdersExhausted++
			}
			if longOrders[i].GetUnFilledBaseAssetQuantity().Sign() == 0 {
				break
			}
		}
		shortOrders = shortOrders[numOrdersExhausted:]
	}
}

func areMatchingOrders(longOrder, shortOrder Order, marginMap map[common.Address]*big.Int, minAllowableMargin, takerFee, upperBound *big.Int) (*big.Int, error) {
	if longOrder.Price.Cmp(shortOrder.Price) == -1 {
		return nil, fmt.Errorf("long order price %s is less than short order price %s", longOrder.Price, shortOrder.Price)
	}
	blockDiff := longOrder.BlockNumber.Cmp(shortOrder.BlockNumber)
	if blockDiff == -1 && (longOrder.OrderType == IOC || shortOrder.isPostOnly()) ||
		blockDiff == 1 && (shortOrder.OrderType == IOC || longOrder.isPostOnly()) {
		return nil, fmt.Errorf("resting order semantics mismatch")
	}
	fillAmount := utils.BigIntMinAbs(longOrder.GetUnFilledBaseAssetQuantity(), shortOrder.GetUnFilledBaseAssetQuantity())
	if fillAmount.Sign() == 0 {
		return nil, fmt.Errorf("no fill amount")
	}

	longMargin, err := isExecutable(&longOrder, fillAmount, minAllowableMargin, takerFee, upperBound, marginMap[longOrder.Trader])
	if err != nil {
		return nil, err
	}

	shortMargin, err := isExecutable(&shortOrder, fillAmount, minAllowableMargin, takerFee, upperBound, marginMap[shortOrder.Trader])
	if err != nil {
		return nil, err
	}
	marginMap[longOrder.Trader].Sub(marginMap[longOrder.Trader], longMargin)
	marginMap[shortOrder.Trader].Sub(marginMap[shortOrder.Trader], shortMargin)
	return fillAmount, nil
}

func isExecutable(order *Order, fillAmount, minAllowableMargin, takerFee, upperBound, availableMargin *big.Int) (*big.Int, error) {
	if order.OrderType == Limit || order.ReduceOnly {
		return big.NewInt(0), nil // no extra margin required because for limit orders it is already reserved
	}
	requiredMargin := big.NewInt(0)
	if order.OrderType == IOC {
		requiredMargin = getRequiredMargin(order, fillAmount, minAllowableMargin, takerFee, upperBound)
	}
	if order.OrderType == Signed {
		requiredMargin = getRequiredMargin(order, fillAmount, minAllowableMargin, big.NewInt(0) /* signed orders are always maker */, upperBound)
	}
	if requiredMargin.Cmp(availableMargin) > 0 {
		return nil, fmt.Errorf("insufficient margin. trader %s, required: %s, available: %s", order.Trader, requiredMargin, availableMargin)
	}
	return requiredMargin, nil
}

func getRequiredMargin(order *Order, fillAmount, minAllowableMargin, takerFee, upperBound *big.Int) *big.Int {
	price := order.Price
	if order.BaseAssetQuantity.Sign() == -1 && order.Price.Cmp(upperBound) == -1 {
		price = upperBound
	}
	return hu.GetRequiredMargin(price, fillAmount, minAllowableMargin, takerFee)
}

func ExecuteMatchedOrders(lotp LimitOrderTxProcessor, longOrder, shortOrder Order, fillAmount *big.Int) (Order, Order) {
	lotp.ExecuteMatchedOrdersTx(longOrder, shortOrder, fillAmount)
	longOrder.FilledBaseAssetQuantity = big.NewInt(0).Add(longOrder.FilledBaseAssetQuantity, fillAmount)
	shortOrder.FilledBaseAssetQuantity = big.NewInt(0).Sub(shortOrder.FilledBaseAssetQuantity, fillAmount)
	return longOrder, shortOrder
}

func matchLongAndShortOrder(lotp LimitOrderTxProcessor, longOrder, shortOrder Order) (Order, Order, bool) {
	fillAmount := utils.BigIntMinAbs(longOrder.GetUnFilledBaseAssetQuantity(), shortOrder.GetUnFilledBaseAssetQuantity())
	if longOrder.Price.Cmp(shortOrder.Price) == -1 || fillAmount.Sign() == 0 {
		return longOrder, shortOrder, false
	}
	if longOrder.BlockNumber.Cmp(shortOrder.BlockNumber) > 0 && longOrder.isPostOnly() {
		log.Warn("post only long order matched with a resting order", "longOrder", longOrder, "shortOrder", shortOrder)
		return longOrder, shortOrder, false
	}
	if shortOrder.BlockNumber.Cmp(longOrder.BlockNumber) > 0 && shortOrder.isPostOnly() {
		log.Warn("post only short order matched with a resting order", "longOrder", longOrder, "shortOrder", shortOrder)
		return longOrder, shortOrder, false
	}
	if err := lotp.ExecuteMatchedOrdersTx(longOrder, shortOrder, fillAmount); err != nil {
		return longOrder, shortOrder, false
	}
	longOrder.FilledBaseAssetQuantity = big.NewInt(0).Add(longOrder.FilledBaseAssetQuantity, fillAmount)
	shortOrder.FilledBaseAssetQuantity = big.NewInt(0).Sub(shortOrder.FilledBaseAssetQuantity, fillAmount)
	return longOrder, shortOrder, true
}

func isFundingPaymentTime(nextFundingTime uint64) bool {
	if nextFundingTime == 0 {
		return false
	}

	now := uint64(time.Now().Unix())
	return now >= nextFundingTime
}

func isSamplePITime(nextSamplePITime, lastAttempt uint64) bool {
	if nextSamplePITime == 0 {
		return false
	}

	now := uint64(time.Now().Unix())
	return now >= nextSamplePITime && now >= lastAttempt+5 // give 5 secs for the tx to be mined
}

func executeFundingPayment(lotp LimitOrderTxProcessor) error {
	// @todo get index twap for each market with warp msging

	return lotp.ExecuteFundingPaymentTx()
}

func removeOrdersWithIds(orders []Order, orderIds map[common.Hash]struct{}) []Order {
	var filteredOrders []Order
	for _, order := range orders {
		if _, ok := orderIds[order.Id]; !ok {
			filteredOrders = append(filteredOrders, order)
		}
	}
	return filteredOrders
}
