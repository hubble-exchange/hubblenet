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

		// first load the last snapshot containing finalised data till block x and unfinalised data will block y
		acceptedBlockNumber, headBlockNumber, err := lop.loadMemoryDBSnapshot()
		if err != nil {
			log.Error("ListenAndProcessTransactions - error in loading snapshot", "err", err)
		} else {
			if acceptedBlockNumber > 0 && headBlockNumber > 0 {
				log.Info("ListenAndProcessTransactions - memory DB snapshot loaded", "acceptedBlockNumber", acceptedBlockNumber, "headBlockNumber", headBlockNumber)
			} else {
				// not an error, but unlikely after the blockchain is running for some time
				log.Warn("ListenAndProcessTransactions - no snapshot found")
			}
		}

		if acceptedBlockNumber == 0 && headBlockNumber == 0 {
			// snapshot was not loaded, start from the beginnning and fetch the logs in chunks
			log.Info("ListenAndProcessTransactions - beginning sync", " till block number", lastAccepted)
			toBlock := utils.BigIntMin(lastAccepted, big.NewInt(0).Add(fromBlock, big.NewInt(10000)))
			for toBlock.Cmp(fromBlock) > 0 {
				logs := lop.getLogs(fromBlock, toBlock)
				log.Info("ListenAndProcessTransactions - fetching log chunk", "fromBlock", fromBlock.String(), "toBlock", toBlock.String(), "number of logs", len(logs), "err", err)
				lop.contractEventProcessor.ProcessEvents(logs)
				lop.contractEventProcessor.ProcessAcceptedEvents(logs)

				fromBlock = fromBlock.Add(toBlock, big.NewInt(1))
				toBlock = utils.BigIntMin(lastAccepted, big.NewInt(0).Add(fromBlock, big.NewInt(10000)))
			}
		} else {
			// snapshot was loaded; fetch only the logs after snapshot blocks(separately for both accepted and head block)
			// no need to break it into chunks cuz the remaining blocks should be few
			fromBlock = big.NewInt(int64(acceptedBlockNumber) + 1)
			logs := lop.getLogs(fromBlock, lastAccepted)
			log.Info("ListenAndProcessTransactions - fetched logs since accepted block", "fromBlock", fromBlock.String(), "toBlock", lastAccepted.String(), "number of logs", len(logs), "err", err)
			lop.contractEventProcessor.ProcessAcceptedEvents(logs)

			fromBlock = big.NewInt(int64(headBlockNumber) + 1)
			logs = lop.getLogs(fromBlock, lastAccepted)
			log.Info("ListenAndProcessTransactions - fetched logs since head block", "fromBlock", fromBlock.String(), "toBlock", lastAccepted.String(), "number of logs", len(logs), "err", err)
			lop.contractEventProcessor.ProcessEvents(logs)
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
		}

	}
}

func (lop *limitOrderProcesser) loadMemoryDBSnapshot() (acceptedBlockNumber uint64, headBlockNumber uint64, err error) {
	snapshotFound, err := lop.hubbleDB.Has([]byte(memoryDBSnapshotKey))
	if err != nil {
		return acceptedBlockNumber, headBlockNumber, fmt.Errorf("Error in checking snapshot in hubbleDB: err=%v", err)
	}

	if !snapshotFound {
		return acceptedBlockNumber, headBlockNumber, nil
	}

	memorySnapshotBytes, err := lop.hubbleDB.Get([]byte(memoryDBSnapshotKey))
	if err != nil {
		return acceptedBlockNumber, headBlockNumber, fmt.Errorf("Error in fetching snapshot from hubbleDB; err=%v", err)
	}

	buf := bytes.NewBuffer(memorySnapshotBytes)
	var snapshot limitorders.Snapshot
	err = gob.NewDecoder(buf).Decode(&snapshot)
	if err != nil {
		return acceptedBlockNumber, headBlockNumber, fmt.Errorf("Error in snapshot parsing; err=%v", err)
	}

	if snapshot.HeadBlockNumber.Uint64() == 0 || snapshot.AcceptedBlockNumber.Uint64() == 0 {
		return acceptedBlockNumber, headBlockNumber, fmt.Errorf("Invalid snapshot; accepted block number =%d, head block number =%v",
			snapshot.AcceptedBlockNumber.Uint64(), snapshot.HeadBlockNumber.Uint64())
	}

	headBlock := lop.blockChain.GetBlockByNumber(snapshot.HeadBlockNumber.Uint64())
	if headBlock.Hash() != snapshot.HeadBlockHash {
		// if the head block at the time of saving the snapshot is not part of the finalised blockchain now,
		// it means there was a reorg and the snapshot has to be discarded
		// This should happen very rarely
		return acceptedBlockNumber, headBlockNumber, fmt.Errorf("HeadBlock mismatch; block number =%d, snapshot head block=%v, blockchain head block=%v",
			snapshot.HeadBlockNumber.Uint64(), snapshot.HeadBlockHash.String(), headBlock.Hash().String())
	}

	err = lop.memoryDb.LoadFromSnapshot(snapshot)
	if err != nil {
		return acceptedBlockNumber, headBlockNumber, fmt.Errorf("Error in loading from snapshot: err=%v", err)
	}

	return snapshot.AcceptedBlockNumber.Uint64(), snapshot.HeadBlockNumber.Uint64(), nil
}

func (lop *limitOrderProcesser) saveMemoryDBSnapshot(acceptedBlockNumber *big.Int) error {
	currentHeadBlock := lop.blockChain.CurrentBlock()

	snapshotBytes, err := lop.memoryDb.GenerateSnapshot(acceptedBlockNumber, currentHeadBlock.Number(), currentHeadBlock.Hash())
	if err != nil {
		return fmt.Errorf("Error in generating snapshot: err=%v", err)
	}

	err = lop.hubbleDB.Put([]byte(memoryDBSnapshotKey), snapshotBytes)
	if err != nil {
		return fmt.Errorf("Error in saving to DB: err=%v", err)
	}

	log.Info("Saved memory DB snapshot successfully", "accepted block", acceptedBlockNumber, "head block number", currentHeadBlock.Number(), "head block hash", currentHeadBlock.Hash())

	return nil
}

func (lop *limitOrderProcesser) getLogs(fromBlock, toBlock *big.Int) []*types.Log {
	ctx := context.Background()
	logs, err := lop.filterAPI.GetLogs(ctx, filters.FilterCriteria{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: []common.Address{limitorders.OrderBookContractAddress, limitorders.ClearingHouseContractAddress, limitorders.MarginAccountContractAddress},
	})

	if err != nil {
		log.Error("ListenAndProcessTransactions - GetLogs failed", "err", err)
		panic(err)
	}
	return logs
}
