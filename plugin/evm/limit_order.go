package evm

import (
	"context"
	"math"
	"sort"
	"sync"

	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/eth"
	"github.com/ava-labs/subnet-evm/eth/filters"
	"github.com/ava-labs/subnet-evm/plugin/evm/limitorders"

	"github.com/ava-labs/avalanchego/snow"
	"github.com/ethereum/go-ethereum/log"
)

type LimitOrderProcesser interface {
	ListenAndProcessTransactions()
	RunLiquidationsAndMatching()
	IsFundingPaymentTime(lastBlockTime uint64) bool
	ExecuteFundingPayment() error
}

type limitOrderProcesser struct {
	ctx                   *snow.Context
	txPool                *core.TxPool
	shutdownChan          <-chan struct{}
	shutdownWg            *sync.WaitGroup
	backend               *eth.EthAPIBackend
	blockChain            *core.BlockChain
	memoryDb              limitorders.LimitOrderDatabase
	limitOrderTxProcessor limitorders.LimitOrderTxProcessor
}

func NewLimitOrderProcesser(ctx *snow.Context, txPool *core.TxPool, shutdownChan <-chan struct{}, shutdownWg *sync.WaitGroup, backend *eth.EthAPIBackend, blockChain *core.BlockChain, memoryDb limitorders.LimitOrderDatabase, lotp limitorders.LimitOrderTxProcessor) LimitOrderProcesser {
	log.Info("**** NewLimitOrderProcesser")
	return &limitOrderProcesser{
		ctx:                   ctx,
		txPool:                txPool,
		shutdownChan:          shutdownChan,
		shutdownWg:            shutdownWg,
		backend:               backend,
		memoryDb:              memoryDb,
		blockChain:            blockChain,
		limitOrderTxProcessor: lotp,
	}
}

func (lop *limitOrderProcesser) ListenAndProcessTransactions() {
	lastAccepted := lop.blockChain.LastAcceptedBlock().NumberU64()
	if lastAccepted > 0 {
		log.Info("ListenAndProcessTransactions - beginning sync", " till block number", lastAccepted)
		ctx := context.Background()

		allTxs := types.Transactions{}
		for i := uint64(0); i <= lastAccepted; i++ {
			block := lop.blockChain.GetBlockByNumber(i)
			if block != nil {
				logs, err := lop.backend.GetLogs(ctx, block.Hash(), block.NumberU64())
				if err != nil {
					log.Error("lop.backend.GetLogs Failed", "err", err)
					continue
				}
				flatLogs := []*types.Log{}
				for _, logsArr := range logs {
					flatLogs = append(flatLogs, logsArr...)
				}
				log.Info("ListenAndProcessTransactions", "block number", i, "logs", flatLogs)
				processEvents(flatLogs, lop)
			} else {
				log.Error("Nil block found", "block number", i)
			}
		}

		log.Info("ListenAndProcessTransactions - sync complete", "till block number", lastAccepted, "total transactions", len(allTxs))
	}

	lop.listenAndStoreLimitOrderTransactions()
}

func (lop *limitOrderProcesser) IsFundingPaymentTime(lastBlockTime uint64) bool {
	return lastBlockTime >= lop.memoryDb.GetNextFundingTime()
}

func (lop *limitOrderProcesser) ExecuteFundingPayment() error {
	// @todo get index twap for each market with warp msging

	return lop.limitOrderTxProcessor.ExecuteFundingPaymentTx()
}

func (lop *limitOrderProcesser) RunLiquidationsAndMatching() {
	lop.limitOrderTxProcessor.PurgeLocalTx()
	for _, market := range limitorders.GetActiveMarkets() {
		longOrders := lop.memoryDb.GetLongOrders(market)
		shortOrders := lop.memoryDb.GetShortOrders(market)
		longOrders, shortOrders = lop.runLiquidations(market, longOrders, shortOrders)
		lop.runMatchingEngine(longOrders, shortOrders)
	}
}

func (lop *limitOrderProcesser) runMatchingEngine(longOrders []limitorders.LimitOrder, shortOrders []limitorders.LimitOrder) {

	if len(longOrders) == 0 || len(shortOrders) == 0 {
		return
	}
	for i := 0; i < len(longOrders); i++ {
		for j := 0; j < len(shortOrders); j++ {
			if getUnFilledBaseAssetQuantity(longOrders[i]) == 0 {
				break
			}
			if getUnFilledBaseAssetQuantity(shortOrders[j]) == 0 {
				continue
			}
			if longOrders[i].Price == shortOrders[j].Price {
				fillAmount := math.Abs(math.Min(float64(getUnFilledBaseAssetQuantity(longOrders[i])), float64(-(getUnFilledBaseAssetQuantity(shortOrders[j])))))
				err := lop.limitOrderTxProcessor.ExecuteMatchedOrdersTx(longOrders[i], shortOrders[j], uint(fillAmount))
				if err == nil {
					longOrders[i].FilledBaseAssetQuantity = longOrders[i].FilledBaseAssetQuantity + int(fillAmount)
					shortOrders[j].FilledBaseAssetQuantity = shortOrders[j].FilledBaseAssetQuantity - int(fillAmount)
				} else {
					log.Error("Error while executing order", "err", err, "longOrder", longOrders[i], "shortOrder", shortOrders[i], "fillAmount", fillAmount)
				}
			}
		}
	}
}

func (lop *limitOrderProcesser) runLiquidations(market limitorders.Market, longOrders []limitorders.LimitOrder, shortOrders []limitorders.LimitOrder) (filteredLongOrder []limitorders.LimitOrder, filteredShortOrder []limitorders.LimitOrder) {
	markPrice := lop.memoryDb.GetLastPrice(market)
	var oraclePrice float64 = 20 // @todo: change this

	liquidableTraders := lop.memoryDb.GetLiquidableTraders(market, markPrice, oraclePrice)

	for _, liquidable := range liquidableTraders {
		var matchedOrders []limitorders.LimitOrder
		if liquidable.Size > 0 {
			// we'll need to sell, so we need a long order to match
			matchedOrders = longOrders
		} else {
			matchedOrders = shortOrders
		}
		if len(matchedOrders) == 0 {
			log.Error("no matching order found for liquidation", "trader", liquidable.Address.String(), "size", liquidable.Size)
		} else {
			// remove the selected matched order from long and short order arrays so that the same order is not matched with multiple liquidations
			if liquidable.Size > 0 {
				longOrders = append(longOrders[:0], longOrders[1:]...) // remove 0th index element
			} else {
				shortOrders = append(shortOrders[:0], shortOrders[1:]...) // remove 0th index element
			}
			matchedOrder := matchedOrders[0]
			lop.limitOrderTxProcessor.ExecuteLiquidation(liquidable.Address, matchedOrder)
		}
	}

	return longOrders, shortOrders
}

func (lop *limitOrderProcesser) listenAndStoreLimitOrderTransactions() {
	newChainChan := make(chan core.ChainEvent)
	chainAcceptedEventSubscription := lop.backend.SubscribeChainAcceptedEvent(newChainChan)

	// lop.limitOrderTxProcessor.GetLastPrice(limitorders.AvaxPerp)
	lop.shutdownWg.Add(1)
	go lop.ctx.Log.RecoverAndPanic(func() {
		defer lop.shutdownWg.Done()
		defer chainAcceptedEventSubscription.Unsubscribe()

		for {
			select {
			case newChainAcceptedEvent := <-newChainChan:
				tsHashes := []string{}
				// blockNumber := newChainAcceptedEvent.Block.Number().Uint64()
				for _, tx := range newChainAcceptedEvent.Block.Transactions() {
					tsHashes = append(tsHashes, tx.Hash().String())
					// if lop.limitOrderTxProcessor.CheckIfOrderBookContractCall(tx) {
					// 	lop.limitOrderTxProcessor.HandleOrderBookTx(tx, blockNumber, *lop.backend)
					// }
				}
				log.Info("$$$$$ New head event", "number", newChainAcceptedEvent.Block.Header().Number, "tx hashes", tsHashes,
					"miner", newChainAcceptedEvent.Block.Coinbase().String(),
					"root", newChainAcceptedEvent.Block.Header().Root.String(), "gas used", newChainAcceptedEvent.Block.Header().GasUsed,
					"nonce", newChainAcceptedEvent.Block.Header().Nonce)

			case <-lop.shutdownChan:
				return
			}
		}
	})

	logsCh := make(chan []*types.Log)
	logsSubscription := lop.backend.SubscribeAcceptedLogsEvent(logsCh)
	lop.shutdownWg.Add(1)
	go lop.ctx.Log.RecoverAndPanic(func() {
		defer lop.shutdownWg.Done()
		defer logsSubscription.Unsubscribe()

		for {
			select {
			case logs := <-logsCh:
				processEvents(logs, lop)
			case <-lop.shutdownChan:
				return
			}
		}
	})

	filterSystem := filters.NewFilterSystem(lop.backend, filters.Config{})
	filters.NewEventSystem(filterSystem, false)
}

func processEvents(logs []*types.Log, lop *limitOrderProcesser) {
	// sort by log index
	sort.SliceStable(logs, func(i, j int) bool {
		return logs[i].Index < logs[j].Index
	})
	for _, event := range logs {
		if event.Removed {
			// skip removed logs
			continue
		}
		switch event.Address {
		case orderBookContractAddress:
			// lop.limitOrderTxProcessor.HandleOrderBookEvent(event)
		case marginAccountContractAddress:
			lop.limitOrderTxProcessor.HandleMarginAccountEvent(event)
		case clearingHouseContractAddress:
			lop.limitOrderTxProcessor.HandleClearingHouseEvent(event)
		}
	}
}

func getUnFilledBaseAssetQuantity(order limitorders.LimitOrder) int {
	return order.BaseAssetQuantity - order.FilledBaseAssetQuantity
}
