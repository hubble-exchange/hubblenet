package evm

import (
	"fmt"
	"math"
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
	RunMatchingEngine()
	RunLiquidations() []error
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

		allTxs := types.Transactions{}
		for i := uint64(0); i <= lastAccepted; i++ {
			block := lop.blockChain.GetBlockByNumber(i)
			if block != nil {
				for _, tx := range block.Transactions() {
					if lop.limitOrderTxProcessor.CheckIfOrderBookContractCall(tx) {
						lop.limitOrderTxProcessor.HandleOrderBookTx(tx, i, *lop.backend)
					}
				}
			}
		}

		log.Info("ListenAndProcessTransactions - sync complete", "till block number", lastAccepted, "total transactions", len(allTxs))
	}

	// @todo maintain margin amounts, open position size, open notionals for all users in memory
	lop.listenAndStoreLimitOrderTransactions()
}

func (lop *limitOrderProcesser) IsFundingPaymentTime(lastBlockTime uint64) bool {
	return lastBlockTime >= lop.memoryDb.GetNextFundingTime()
}

func (lop *limitOrderProcesser) ExecuteFundingPayment() error {
	// @todo get index twap for each market with warp msging

	return lop.limitOrderTxProcessor.ExecuteFundingPaymentTx()
}

func (lop *limitOrderProcesser) RunMatchingEngine() {
	lop.limitOrderTxProcessor.PurgeLocalTx()
	longOrders := lop.memoryDb.GetLongOrders()
	shortOrders := lop.memoryDb.GetShortOrders()
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
				}
			}
		}
	}
}

func (lop *limitOrderProcesser) RunLiquidations() (errors []error) {
	longOrders := lop.memoryDb.GetLongOrders()
	shortOrders := lop.memoryDb.GetShortOrders()

	for _, market := range limitorders.GetActiveMarkets() {
		var markPrice float64 = 20 // last traded price
		var oraclePrice float64 = 19

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
				errors = append(errors, fmt.Errorf("no matching order found for trader %s, size = %f", liquidable.Address.String(), liquidable.Size))
				continue
			} else {
				// remove the selected matched order from long and short order arrays so that the same order is not matched with multiple liquidations
				if liquidable.Size > 0 {
					longOrders = append(longOrders[:0], longOrders[1:]...) // remove 0th index element
				} else {
					shortOrders = append(shortOrders[:0], shortOrders[1:]...) // remove 0th index element
				}
				matchedOrder := matchedOrders[0]
				err := lop.limitOrderTxProcessor.ExecuteLiquidation(liquidable.Address, matchedOrder)
				if err != nil {
					errors = append(errors, err)
					continue
				}
			}
		}
	}

	return errors
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
				blockNumber := newChainAcceptedEvent.Block.Number().Uint64()
				for _, tx := range newChainAcceptedEvent.Block.Transactions() {
					tsHashes = append(tsHashes, tx.Hash().String())
					if lop.limitOrderTxProcessor.CheckIfOrderBookContractCall(tx) {
						lop.limitOrderTxProcessor.HandleOrderBookTx(tx, blockNumber, *lop.backend)
					}
				}
				// @todo maintain margin amounts, open position size, open notionals for all users in memory
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
				for _, event := range logs {
					switch event.Address {
					case orderBookContractAddress:
						lop.limitOrderTxProcessor.HandleOrderBookEvent(event)
					case marginAccountContractAddress:
						lop.limitOrderTxProcessor.HandleMarginAccountEvent(event)
					case clearingHouseContractAddress:
						lop.limitOrderTxProcessor.HandleClearingHouseEvent(event)
					}
				}

			case <-lop.shutdownChan:
				return
			}
		}
	})

	filterSystem := filters.NewFilterSystem(lop.backend, filters.Config{})
	filters.NewEventSystem(filterSystem, false)
}

func getUnFilledBaseAssetQuantity(order limitorders.LimitOrder) int {
	return order.BaseAssetQuantity - order.FilledBaseAssetQuantity
}
