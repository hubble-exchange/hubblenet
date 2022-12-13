package evm

import (
	"io/ioutil"
	"math/big"

	"sync"

	"github.com/ava-labs/subnet-evm/accounts/abi"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/eth"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/ava-labs/subnet-evm/plugin/evm/limitorders"

	"github.com/ava-labs/avalanchego/snow"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

var orderBookContractFileLocation = "contract-examples/artifacts/contracts/hubble-v2/OrderBook.sol/OrderBook.json"

type LimitOrderProcesser interface {
	ListenAndProcessTransactions()
}

type limitOrderProcesser struct {
	ctx          *snow.Context
	chainConfig  *params.ChainConfig
	txPool       *core.TxPool
	shutdownChan <-chan struct{}
	shutdownWg   *sync.WaitGroup
	backend      *eth.EthAPIBackend
	memoryDb     limitorders.InMemoryDatabase
	orderBookABI abi.ABI
}

func SetOrderBookContractFileLocation(location string) {
	orderBookContractFileLocation = location
}

func NewLimitOrderProcesser(ctx *snow.Context, chainConfig *params.ChainConfig, txPool *core.TxPool, shutdownChan <-chan struct{}, shutdownWg *sync.WaitGroup, backend *eth.EthAPIBackend) LimitOrderProcesser {
	jsonBytes, _ := ioutil.ReadFile(orderBookContractFileLocation)
	orderBookAbi, err := abi.FromSolidityJson(string(jsonBytes))
	if err != nil {
		panic(err)
	}

	return &limitOrderProcesser{
		ctx:          ctx,
		chainConfig:  chainConfig,
		txPool:       txPool,
		shutdownChan: shutdownChan,
		shutdownWg:   shutdownWg,
		backend:      backend,
		memoryDb:     limitorders.NewInMemoryDatabase(),
		orderBookABI: orderBookAbi,
	}
}

func (lop *limitOrderProcesser) ListenAndProcessTransactions() {
	lop.listenAndStoreLimitOrderTransactions()
}

func (lop *limitOrderProcesser) listenAndStoreLimitOrderTransactions() {
	newHeadChan := make(chan core.NewTxPoolHeadEvent)
	lop.txPool.SubscribeNewHeadEvent(newHeadChan)

	lop.shutdownWg.Add(1)
	go lop.ctx.Log.RecoverAndPanic(func() {
		defer lop.shutdownWg.Done()

		for {
			select {
			case newHeadEvent := <-newHeadChan:
				tsHashes := []string{}
				for _, tx := range newHeadEvent.Block.Transactions() {
					tsHashes = append(tsHashes, tx.Hash().String())
					parseTx(lop.txPool, lop.orderBookABI, lop.memoryDb, tx) // parse update in memory db
				}
				log.Info("$$$$$ New head event", "number", newHeadEvent.Block.Header().Number, "tx hashes", tsHashes,
					"miner", newHeadEvent.Block.Coinbase().String(),
					"root", newHeadEvent.Block.Header().Root.String(), "gas used", newHeadEvent.Block.Header().GasUsed,
					"nonce", newHeadEvent.Block.Header().Nonce)
			case <-lop.shutdownChan:
				return
			}
		}
	})
}

func parseTx(txPool *core.TxPool, orderBookABI abi.ABI, memoryDb limitorders.InMemoryDatabase, tx *types.Transaction) {
	input := tx.Data()
	if len(input) < 4 {
		log.Info("transaction data has less than 3 fields")
		return
	}
	method := input[:4]
	m, err := orderBookABI.MethodById(method)
	if err == nil {
		in := make(map[string]interface{})
		_ = m.Inputs.UnpackIntoMap(in, input[4:])
		if m.Name == "placeOrder" {
			log.Info("##### in ParseTx", "placeOrder tx hash", tx.Hash().String())
			order, _ := in["order"].(struct {
				Trader            common.Address `json:"trader"`
				BaseAssetQuantity *big.Int       `json:"baseAssetQuantity"`
				Price             *big.Int       `json:"price"`
				Salt              *big.Int       `json:"salt"`
			})
			signature := in["signature"].([]byte)

			baseAssetQuantity := int(order.BaseAssetQuantity.Int64())
			if baseAssetQuantity == 0 {
				log.Error("order not saved because baseAssetQuantity is zero")
				return
			}
			positionType := getPositionTypeBasedOnBaseAssetQuantity(baseAssetQuantity)
			price, _ := new(big.Float).SetInt(order.Price).Float64()
			limitOrder := limitorders.LimitOrder{
				PositionType:      positionType,
				UserAddress:       order.Trader.Hash().String(),
				BaseAssetQuantity: baseAssetQuantity,
				Price:             price,
				Salt:              order.Salt.String(),
				Status:            "unfulfilled",
				Signature:         signature,
				RawOrder:          in["order"],
				RawSignature:      in["signature"],
			}
			memoryDb.Add(limitOrder)
			matchLimitOrderAgainstStoredLimitOrders(txPool, orderBookABI, memoryDb, limitOrder)
		}
		if m.Name == "executeMatchedOrders" {
			signature1 := in["signature1"].([]byte)
			memoryDb.Delete(signature1)
			signature2 := in["signature2"].([]byte)
			memoryDb.Delete(signature2)
		}
	}
}

func matchLimitOrderAgainstStoredLimitOrders(txPool *core.TxPool, orderBookABI abi.ABI, memoryDb limitorders.InMemoryDatabase, limitOrder limitorders.LimitOrder) {
	oppositePositionType := getOppositePositionType(limitOrder.PositionType)
	potentialMatchingOrders := memoryDb.GetOrdersByPriceAndPositionType(oppositePositionType, limitOrder.Price)

	for _, order := range potentialMatchingOrders {
		if order.BaseAssetQuantity == -(limitOrder.BaseAssetQuantity) && order.Price == limitOrder.Price {
			callExecuteMatchedOrders(txPool, orderBookABI, limitOrder, *order)
		}
	}
}

func callExecuteMatchedOrders(txPool *core.TxPool, orderBookABI abi.ABI, incomingOrder limitorders.LimitOrder, matchedOrder limitorders.LimitOrder) {
	nonce := txPool.Nonce(common.HexToAddress("0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")) // admin address

	data, err := orderBookABI.Pack("executeMatchedOrders", incomingOrder.RawOrder, incomingOrder.Signature, matchedOrder.RawOrder, matchedOrder.Signature)
	if err != nil {
		log.Error("abi.Pack failed", "err", err)
	}
	key, err := crypto.HexToECDSA("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027") // admin private key
	if err != nil {
		log.Error("HexToECDSA failed", "err", err)
	}
	executeMatchedOrdersTx := types.NewTransaction(nonce, common.HexToAddress("0x0300000000000000000000000000000000000069"), big.NewInt(0), 5000000, big.NewInt(80000000000), data)
	signer := types.NewLondonSigner(big.NewInt(321123))
	signedTx, err := types.SignTx(executeMatchedOrdersTx, signer, key)
	if err != nil {
		log.Error("types.SignTx failed", "err", err)
	}
	err = txPool.AddLocal(signedTx)
	if err != nil {
		log.Error("lop.txPool.AddLocal failed", "err", err)
	}
}

func getOppositePositionType(positionType string) string {
	if positionType == "long" {
		return "short"
	}
	return "long"
}

func getPositionTypeBasedOnBaseAssetQuantity(baseAssetQuantity int) string {
	if baseAssetQuantity > 0 {
		return "long"
	}
	return "short"
}
