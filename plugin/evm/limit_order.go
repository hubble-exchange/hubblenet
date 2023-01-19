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
	"github.com/ethereum/go-ethereum/common"
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

		filterSystem := filters.NewFilterSystem(lop.backend, filters.Config{})
		filterAPI := filters.NewFilterAPI(filterSystem, true)
		logs, err := filterAPI.GetLogs(ctx, filters.FilterCriteria{
			Addresses: []common.Address{limitorders.OrderBookContractAddress, limitorders.ClearingHouseContractAddress, limitorders.MarginAccountContractAddress},
		})
		if err != nil {
			log.Error("ListenAndProcessTransactions - GetLogs failed", "err", err)
			panic(err)
		}
		log.Info("ListenAndProcessTransactions", "total logs", len(logs), "err", err)
		processEvents(logs, lop)
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

	liquidablePositions := lop.memoryDb.GetLiquidableTraders(market, oraclePrice)

	for i, liquidable := range liquidablePositions {
		var oppositeOrders []limitorders.LimitOrder
		switch liquidable.PositionType {
		case "long":
			oppositeOrders = shortOrders
		case "short":
			oppositeOrders = longOrders
		}
		if len(oppositeOrders) == 0 {
			log.Error("no matching order found for liquidation", "trader", liquidable.Address.String(), "size", liquidable.Size)
			continue // so that all other liquidable positions get logged
		}
		for j, oppositeOrder := range oppositeOrders {
			if liquidable.GetUnfilledSize().Sign() == 0 {
				break
			}
			fillAmount := utils.BigIntMinAbs(liquidable.GetUnfilledSize(), oppositeOrder.GetUnFilledBaseAssetQuantity())
			if fillAmount.Sign() == 0 {
				continue
			}
			lop.limitOrderTxProcessor.ExecuteLiquidation(liquidable.Address, oppositeOrder, fillAmount)

			switch liquidable.PositionType {
			case "long":
				oppositeOrders[j].FilledBaseAssetQuantity.Sub(shortOrders[j].FilledBaseAssetQuantity, fillAmount)
				liquidablePositions[i].FilledSize.Add(liquidablePositions[i].FilledSize, fillAmount)
			case "short":
				oppositeOrders[j].FilledBaseAssetQuantity.Add(shortOrders[j].FilledBaseAssetQuantity, fillAmount)
				liquidablePositions[i].FilledSize.Sub(liquidablePositions[i].FilledSize, fillAmount)
			}
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
	// sort by block number & log index
	sort.SliceStable(logs, func(i, j int) bool {
		if logs[i].BlockNumber == logs[j].BlockNumber {
			return logs[i].Index < logs[j].Index
		}
		return logs[i].BlockNumber < logs[j].BlockNumber
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
