package evm

import (
	"context"
	"math/big"
	"sync"

	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/core/rawdb"
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
	RunBuildBlockPipeline(lastBlockTime uint64)
	GetOrderBookAPI() *limitorders.OrderBookAPI
}

type limitOrderProcesser struct {
	ctx                    *snow.Context
	txPool                 *core.TxPool
	shutdownChan           <-chan struct{}
	shutdownWg             *sync.WaitGroup
	backend                *eth.EthAPIBackend
	blockChain             *core.BlockChain
	memoryDb               limitorders.LimitOrderDatabase
	limitOrderTxProcessor  limitorders.LimitOrderTxProcessor
	contractEventProcessor *limitorders.ContractEventsProcessor
	buildBlockPipeline     *limitorders.BuildBlockPipeline
	mu                     sync.Mutex
}

func NewLimitOrderProcesser(ctx *snow.Context, txPool *core.TxPool, shutdownChan <-chan struct{}, shutdownWg *sync.WaitGroup, backend *eth.EthAPIBackend, blockChain *core.BlockChain, memoryDb limitorders.LimitOrderDatabase, lotp limitorders.LimitOrderTxProcessor) LimitOrderProcesser {
	log.Info("**** NewLimitOrderProcesser")
	contractEventProcessor := limitorders.NewContractEventsProcessor(memoryDb)
	buildBlockPipeline := limitorders.NewBuildBlockPipeline(memoryDb, lotp)
	return &limitOrderProcesser{
		ctx:                    ctx,
		txPool:                 txPool,
		shutdownChan:           shutdownChan,
		shutdownWg:             shutdownWg,
		backend:                backend,
		memoryDb:               memoryDb,
		blockChain:             blockChain,
		limitOrderTxProcessor:  lotp,
		contractEventProcessor: contractEventProcessor,
		buildBlockPipeline:     buildBlockPipeline,
	}
}

func (lop *limitOrderProcesser) ListenAndProcessTransactions() {
	lastAccepted := lop.blockChain.LastAcceptedBlock().Number()
	if lastAccepted.Sign() > 0 {
		log.Info("ListenAndProcessTransactions - beginning sync", " till block number", lastAccepted)
		ctx := context.Background()

		filterSystem := filters.NewFilterSystem(lop.backend, filters.Config{})
		filterAPI := filters.NewFilterAPI(filterSystem, true)

		var fromBlock, toBlock *big.Int
		fromBlock = big.NewInt(0)
		toBlock = utils.BigIntMin(lastAccepted, big.NewInt(0).Add(fromBlock, big.NewInt(10000)))
		for toBlock.Cmp(fromBlock) >= 0 {
			logs, err := filterAPI.GetLogs(ctx, filters.FilterCriteria{
				FromBlock: fromBlock,
				ToBlock:   toBlock,
				Addresses: []common.Address{limitorders.OrderBookContractAddress, limitorders.ClearingHouseContractAddress, limitorders.MarginAccountContractAddress},
			})
			if err != nil {
				log.Error("ListenAndProcessTransactions - GetLogs failed", "err", err)
				panic(err)
			}
			lop.contractEventProcessor.ProcessEvents(logs)
			log.Info("ListenAndProcessTransactions", "fromBlock", fromBlock.String(), "toBlock", toBlock.String(), "number of logs", len(logs), "err", err)

			toBlock = utils.BigIntMin(lastAccepted, big.NewInt(0).Add(fromBlock, big.NewInt(10000)))
			fromBlock = fromBlock.Add(toBlock, big.NewInt(1))
		}
	}

	lop.listenAndStoreLimitOrderTransactions()
}

func (lop *limitOrderProcesser) RunBuildBlockPipeline(lastBlockTime uint64) {
	lop.buildBlockPipeline.Run(lastBlockTime)
}

func (lop *limitOrderProcesser) GetOrderBookAPI() *limitorders.OrderBookAPI {
	return limitorders.NewOrderBookAPI(lop.memoryDb)
}

func (lop *limitOrderProcesser) listenAndStoreLimitOrderTransactions() {
	chainHeadEventCh := make(chan core.ChainHeadEvent)
	chainHeadEventSubscription := lop.backend.SubscribeChainHeadEvent(chainHeadEventCh)
	lop.shutdownWg.Add(1)
	go lop.ctx.Log.RecoverAndPanic(func() {
		defer lop.shutdownWg.Done()
		defer chainHeadEventSubscription.Unsubscribe()

		for {
			select {
			case chainHeadEvent := <-chainHeadEventCh:
				lop.handleChainHeadEvent(chainHeadEvent.Block)
			case <-lop.shutdownChan:
				return
			}
		}
	})

	chainAcceptedEventCh := make(chan core.ChainEvent)
	chainAcceptedEventSubscription := lop.backend.SubscribeChainAcceptedEvent(chainAcceptedEventCh)
	lop.shutdownWg.Add(1)
	go lop.ctx.Log.RecoverAndPanic(func() {
		defer lop.shutdownWg.Done()
		defer chainAcceptedEventSubscription.Unsubscribe()

		for {
			select {
			case chainAcceptedEvent := <-chainAcceptedEventCh:
				lop.handleChainAcceptedEvent(chainAcceptedEvent)
			case <-lop.shutdownChan:
				return
			}
		}
	})
}

func (lop *limitOrderProcesser) handleChainHeadEvent(block *types.Block) {
	lop.mu.Lock()
	defer lop.mu.Unlock()

	log.Info("#### received ChainHeadEvent", "number", block.NumberU64(), "hash", block.Hash().String())
	inProgressBlocks := lop.memoryDb.GetInProgressBlocks()
	blocksToApply := []*types.Block{}
	blocksToRemove := []*types.Block{}
	if len(inProgressBlocks) > 0 {
		// figure out which blocks to apply and which blocks to remove from in-progress state
		commonHeader := rawdb.FindCommonAncestor(lop.backend.ChainDb(), block.Header(), inProgressBlocks[0].Header())
		currentBlock := block
		for {
			blocksToApply = append(blocksToApply, currentBlock)
			currentBlock = lop.blockChain.GetBlockByHash(currentBlock.ParentHash())
			if currentBlock.Hash() == commonHeader.Hash() {
				break
			}
		}
		for _, inProgressBlock := range inProgressBlocks {
			if inProgressBlock.Hash() == commonHeader.Hash() {
				break
			}
			blocksToRemove = append(blocksToRemove, inProgressBlock)
		}
		log.Info("#### inprogress blocks found", "blocks", len(inProgressBlocks), "commonHeader", commonHeader.Number.Uint64(), "common hash", commonHeader.Hash().String(), "blocksToApply", blockHashes(blocksToApply), "blocksToRemove", blockHashes(blocksToRemove))
	} else {
		// assume only one block needs to be added
		log.Info("#### inprogress blocks not found; applying single block", "number", block.NumberU64(), "hash", block.Hash().String())
		blocksToApply = append(blocksToApply, block)
	}

	// iterate in straight order so that blocks are removed in order of decreasing block number
	for _, blockToRemove := range blocksToRemove {
		lop.memoryDb.RemoveInProgressState(blockToRemove, lop.getOrderQuantityMap(blockToRemove))
	}

	// iterate in reverse order so that blocks are applied in order of increasing block number
	if len(blocksToApply) > 0 {
		for i := len(blocksToApply) - 1; i >= 0; i-- {
			blockToApply := blocksToApply[i]
			lop.memoryDb.UpdateInProgressState(blockToApply, lop.getOrderQuantityMap(blockToApply))
		}
	}
}

func (lop *limitOrderProcesser) handleChainAcceptedEvent(event core.ChainEvent) {
	lop.mu.Lock()
	defer lop.mu.Unlock()

	block := event.Block

	log.Info("#### received ChainAcceptedEvent", "number", block.NumberU64(), "hash", block.Hash().String())

	lop.memoryDb.RemoveInProgressState(block, lop.getOrderQuantityMap(block))

	lop.contractEventProcessor.ProcessEvents(event.Logs)
}

func blockHashes(blocks []*types.Block) []string {
	hashes := []string{}
	for _, block := range blocks {
		hashes = append(hashes, block.Hash().String())
	}
	return hashes
}

func (lop *limitOrderProcesser) getOrderQuantityMap(block *types.Block) map[string]*big.Int {
	logs, _ := lop.backend.GetLogs(context.Background(), block.Hash(), block.NumberU64())
	flatLogs := types.FlattenLogs(logs)

	return lop.contractEventProcessor.GetMatchedOrderQuantity(flatLogs)
}
