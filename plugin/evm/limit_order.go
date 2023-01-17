package evm

import (
	"context"
	"math/big"
	"sort"
	"sync"

	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/eth"
	"github.com/ava-labs/subnet-evm/eth/filters"
	"github.com/ava-labs/subnet-evm/plugin/evm/limitorders"
	"github.com/ava-labs/subnet-evm/utils"

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
			if longOrders[i].GetUnFilledBaseAssetQuantity().Sign() == 0 {
				break
			}
			if shortOrders[j].GetUnFilledBaseAssetQuantity().Sign() == 0 {
				continue
			}
			if longOrders[i].Price == shortOrders[j].Price {
				fillAmount := utils.BigIntMinAbs(longOrders[i].GetUnFilledBaseAssetQuantity(), shortOrders[j].GetUnFilledBaseAssetQuantity())
				err := lop.limitOrderTxProcessor.ExecuteMatchedOrdersTx(longOrders[i], shortOrders[j], fillAmount)
				if err == nil {
					longOrders[i].FilledBaseAssetQuantity = big.NewInt(0).Add(longOrders[i].FilledBaseAssetQuantity, fillAmount)
					shortOrders[j].FilledBaseAssetQuantity = big.NewInt(0).Sub(shortOrders[j].FilledBaseAssetQuantity, fillAmount)
				} else {
					log.Error("Error while executing order", "err", err, "longOrder", longOrders[i], "shortOrder", shortOrders[i], "fillAmount", fillAmount)
				}
			}
		}
	}
}

func (lop *limitOrderProcesser) runLiquidations(market limitorders.Market, longOrders []limitorders.LimitOrder, shortOrders []limitorders.LimitOrder) (filteredLongOrder []limitorders.LimitOrder, filteredShortOrder []limitorders.LimitOrder) {
	oraclePrice := big.NewInt(20 * 10e6) // @todo: get it from the oracle

	longPositions, shortPositions := lop.memoryDb.GetLiquidableTraders(market, oraclePrice)

	for i, liquidable := range longPositions {
		if len(shortOrders) == 0 {
			log.Error("no matching order found for liquidation", "trader", liquidable.Address.String(), "size", liquidable.Size)
			continue // so that all other liquidable positions get logged
		}
		for j, shortOrder := range shortOrders {
			if liquidable.GetUnfilledSize().Sign() == 0 {
				break
			}
			fillAmount := utils.BigIntMinAbs(liquidable.GetUnfilledSize(), shortOrder.GetUnFilledBaseAssetQuantity())
			if fillAmount.Sign() == 0 {
				continue
			}
			lop.limitOrderTxProcessor.ExecuteLiquidation(liquidable.Address, shortOrder, fillAmount)
			shortOrders[j].FilledBaseAssetQuantity.Sub(shortOrders[j].FilledBaseAssetQuantity, fillAmount)
			longPositions[i].FilledSize.Add(longPositions[i].FilledSize, fillAmount)
		}
	}

	for i, liquidable := range shortPositions {
		if len(longOrders) == 0 {
			log.Error("no matching order found for liquidation", "trader", liquidable.Address.String(), "size", liquidable.Size)
			continue // so that all other liquidable positions get logged
		}
		for j, longOrder := range longOrders {
			if liquidable.GetUnfilledSize().Sign() == 0 {
				break
			}
			fillAmount := utils.BigIntMinAbs(liquidable.GetUnfilledSize(), longOrder.GetUnFilledBaseAssetQuantity())
			if fillAmount.Sign() == 0 {
				continue
			}
			lop.limitOrderTxProcessor.ExecuteLiquidation(liquidable.Address, longOrder, fillAmount)
			longOrders[j].FilledBaseAssetQuantity.Add(longOrders[j].FilledBaseAssetQuantity, fillAmount)
			shortPositions[i].FilledSize.Sub(shortPositions[i].FilledSize, fillAmount)
		}
	}

	return longOrders, shortOrders
}

func (lop *limitOrderProcesser) listenAndStoreLimitOrderTransactions() {
	newChainChan := make(chan core.ChainEvent)
	chainAcceptedEventSubscription := lop.backend.SubscribeChainAcceptedEvent(newChainChan)

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
		case limitorders.OrderBookContractAddress:
			lop.limitOrderTxProcessor.HandleOrderBookEvent(event)
		case limitorders.MarginAccountContractAddress:
			lop.limitOrderTxProcessor.HandleMarginAccountEvent(event)
		case limitorders.ClearingHouseContractAddress:
			lop.limitOrderTxProcessor.HandleClearingHouseEvent(event)
		}
	}
}
