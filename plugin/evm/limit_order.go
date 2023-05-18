package evm

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"math/big"
	"sync"

	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/eth"
	"github.com/ava-labs/subnet-evm/eth/filters"
	"github.com/ava-labs/subnet-evm/plugin/evm/limitorders"
	"github.com/ava-labs/subnet-evm/utils"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

const (
	memoryDBSnapshotKey string = "memoryDBSnapshot"
	snapshotInterval    uint64 = 1000 // save snapshot every 1000 blocks
)

type LimitOrderProcesser interface {
	ListenAndProcessTransactions()
	RunBuildBlockPipeline()
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
	filterAPI              *filters.FilterAPI
	hubbleDB               database.Database
}

func NewLimitOrderProcesser(ctx *snow.Context, txPool *core.TxPool, shutdownChan <-chan struct{}, shutdownWg *sync.WaitGroup, backend *eth.EthAPIBackend, blockChain *core.BlockChain, memoryDb limitorders.LimitOrderDatabase, hubbleDB database.Database, lotp limitorders.LimitOrderTxProcessor) LimitOrderProcesser {
	log.Info("**** NewLimitOrderProcesser")
	contractEventProcessor := limitorders.NewContractEventsProcessor(memoryDb)
	buildBlockPipeline := limitorders.NewBuildBlockPipeline(memoryDb, lotp)
	filterSystem := filters.NewFilterSystem(backend, filters.Config{})
	filterAPI := filters.NewFilterAPI(filterSystem, true)
	return &limitOrderProcesser{
		ctx:                    ctx,
		txPool:                 txPool,
		shutdownChan:           shutdownChan,
		shutdownWg:             shutdownWg,
		backend:                backend,
		memoryDb:               memoryDb,
		hubbleDB:               hubbleDB,
		blockChain:             blockChain,
		limitOrderTxProcessor:  lotp,
		contractEventProcessor: contractEventProcessor,
		buildBlockPipeline:     buildBlockPipeline,
		filterAPI:              filterAPI,
	}
}

func (lop *limitOrderProcesser) ListenAndProcessTransactions() {
	lastAccepted := lop.blockChain.LastAcceptedBlock().Number()
	if lastAccepted.Sign() > 0 {
		fromBlock := big.NewInt(0)

		// first load the last snapshot containing data till block x, then parse all the logs since block x + 1
		blockNumber, err := lop.loadMemoryDBSnapshot()
		if err != nil {
			log.Error("ListenAndProcessTransactions - Error in loading snapshot", "err", err)
		}
		if blockNumber > 0 {
			log.Info("ListenAndProcessTransactions - Memory DB snapshot loaded", "blockNumber", blockNumber)
			fromBlock = big.NewInt(int64(blockNumber + 1))
		}

		log.Info("ListenAndProcessTransactions - beginning sync", " till block number", lastAccepted)
		ctx := context.Background()

		var toBlock *big.Int
		toBlock = utils.BigIntMin(lastAccepted, big.NewInt(0).Add(fromBlock, big.NewInt(10000)))
		for toBlock.Cmp(fromBlock) > 0 {
			logs, err := lop.filterAPI.GetLogs(ctx, filters.FilterCriteria{
				FromBlock: fromBlock,
				ToBlock:   toBlock, // check that this is inclusive...
				Addresses: []common.Address{limitorders.OrderBookContractAddress, limitorders.ClearingHouseContractAddress, limitorders.MarginAccountContractAddress},
			})
			log.Info("ListenAndProcessTransactions", "fromBlock", fromBlock.String(), "toBlock", toBlock.String(), "number of logs", len(logs), "err", err)
			if err != nil {
				log.Error("ListenAndProcessTransactions - GetLogs failed", "err", err)
				panic(err)
			}
			lop.contractEventProcessor.ProcessEvents(logs)
			lop.contractEventProcessor.ProcessAcceptedEvents(logs)

			fromBlock = fromBlock.Add(toBlock, big.NewInt(1))
			toBlock = utils.BigIntMin(lastAccepted, big.NewInt(0).Add(fromBlock, big.NewInt(10000)))
		}
		lop.memoryDb.Accept(lastAccepted.Uint64()) // will delete stale orders from the memorydb
	}

	lop.listenAndStoreLimitOrderTransactions()
}

func (lop *limitOrderProcesser) RunBuildBlockPipeline() {
	lop.buildBlockPipeline.Run()
}

func (lop *limitOrderProcesser) GetOrderBookAPI() *limitorders.OrderBookAPI {
	return limitorders.NewOrderBookAPI(lop.memoryDb, lop.backend)
}

func (lop *limitOrderProcesser) listenAndStoreLimitOrderTransactions() {
	logsCh := make(chan []*types.Log)
	logsSubscription := lop.backend.SubscribeHubbleLogsEvent(logsCh)
	lop.shutdownWg.Add(1)
	go lop.ctx.Log.RecoverAndPanic(func() {
		defer lop.shutdownWg.Done()
		defer logsSubscription.Unsubscribe()

		for {
			select {
			case logs := <-logsCh:
				lop.contractEventProcessor.ProcessEvents(logs)
			case <-lop.shutdownChan:
				return
			}
		}
	})

	acceptedLogsCh := make(chan []*types.Log)
	acceptedLogsSubscription := lop.backend.SubscribeAcceptedLogsEvent(acceptedLogsCh)
	lop.shutdownWg.Add(1)
	go lop.ctx.Log.RecoverAndPanic(func() {
		defer lop.shutdownWg.Done()
		defer acceptedLogsSubscription.Unsubscribe()

		for {
			select {
			case logs := <-acceptedLogsCh:
				lop.contractEventProcessor.ProcessAcceptedEvents(logs)
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

func (lop *limitOrderProcesser) handleChainAcceptedEvent(event core.ChainEvent) {
	block := event.Block
	log.Info("#### received ChainAcceptedEvent", "number", block.NumberU64(), "hash", block.Hash().String())
	lop.memoryDb.Accept(block.NumberU64())

	if block.NumberU64()%snapshotInterval == 0 {
		err := lop.saveMemoryDBSnapshot(block.Number())
		if err != nil {
			log.Error("Error in saving memory DB snapshot", "err", err)
		} else {
			log.Info("Saved memory DB snapshot successfully")
		}

	}
}

func (lop *limitOrderProcesser) loadMemoryDBSnapshot() (blockNumber uint64, err error) {
	snapshotFound, err := lop.hubbleDB.Has([]byte(memoryDBSnapshotKey))
	if err != nil {
		return blockNumber, fmt.Errorf("Error in checking snapshot: err=%v", err)
	}

	if !snapshotFound {
		return blockNumber, nil
	}

	memorySnapshotBytes, err := lop.hubbleDB.Get([]byte(memoryDBSnapshotKey))
	if err != nil {
		return blockNumber, fmt.Errorf("Error in fetching snapshot: err=%v", err)
	}

	buf := bytes.NewBuffer(memorySnapshotBytes)
	var snapshot limitorders.Snapshot
	err = gob.NewDecoder(buf).Decode(&snapshot)
	if err != nil {
		log.Error("Error in gob parsing", "snapshotbytes", string(memorySnapshotBytes))
		return blockNumber, fmt.Errorf("Error in json parsing: err=%v", err)
	}

	err = lop.memoryDb.LoadFromSnapshot(snapshot)
	if err != nil {
		return blockNumber, fmt.Errorf("Error in loading from snapshot: err=%v", err)
	}

	return snapshot.BlockNumber.Uint64(), nil
}

func (lop *limitOrderProcesser) saveMemoryDBSnapshot(blockNumber *big.Int) error {
	memoryDb := lop.memoryDb.GetOrderBookData()

	snapshot := limitorders.Snapshot{
		Data: limitorders.InMemoryDatabase{
			OrderMap:        memoryDb.OrderMap,
			TraderMap:       memoryDb.TraderMap,
			LastPrice:       memoryDb.LastPrice,
			NextFundingTime: memoryDb.NextFundingTime,
		},
		BlockNumber: blockNumber,
	}

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(&snapshot)
	if err != nil {
		return fmt.Errorf("Error in gob encoding: err=%v", err)
	}

	err = lop.hubbleDB.Put([]byte(memoryDBSnapshotKey), buf.Bytes())
	if err != nil {
		return fmt.Errorf("Error in saving to DB: err=%v", err)
	}

	return nil
}
