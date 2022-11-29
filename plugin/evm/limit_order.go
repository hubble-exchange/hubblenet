package evm

import (
	"io/ioutil"
	"math/big"
	"reflect"
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

type LimitOrderProcesser interface {
	ListenAndProcessLimitOrderTransactions()
}

type limitOrderProcesser struct {
	ctx          *snow.Context
	chainConfig  *params.ChainConfig
	txPool       *core.TxPool
	shutdownChan <-chan struct{}
	shutdownWg   *sync.WaitGroup
	backend      *eth.EthAPIBackend
	database     limitorders.LimitOrderDatabase
}

func NewLimitOrderProcesser(ctx *snow.Context, chainConfig *params.ChainConfig, txPool *core.TxPool, shutdownChan <-chan struct{}, shutdownWg *sync.WaitGroup, backend *eth.EthAPIBackend) LimitOrderProcesser {
	database, err := limitorders.InitializeDatabase()
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
		database:     database,
	}
}

func (lop *limitOrderProcesser) ListenAndProcessLimitOrderTransactions() {
	lop.listenAndStoreLimitOrderTransactions()
}

func (lop *limitOrderProcesser) listenAndStoreLimitOrderTransactions() {

	type Order struct {
		Trader            common.Address `json:"trader"`
		BaseAssetQuantity *big.Int       `json:"baseAssetQuantity"`
		Price             *big.Int       `json:"price"`
		Salt              *big.Int       `json:"salt"`
	}
	jsonBytes, _ := ioutil.ReadFile("contract-examples/artifacts/contracts/OrderBook.sol/OrderBook.json")
	orderBookAbi, err := abi.FromSolidityJson(string(jsonBytes))
	if err != nil {
		panic(err)
	}

	txSubmitChan := make(chan core.NewTxsEvent)
	lop.txPool.SubscribeNewTxsEvent(txSubmitChan)
	lop.shutdownWg.Add(1)
	go lop.ctx.Log.RecoverAndPanic(func() {
		defer lop.shutdownWg.Done()

		for {
			select {
			case txsEvent := <-txSubmitChan:
				log.Info("New transaction event detected")

				for i := 0; i < len(txsEvent.Txs); i++ {
					tx := txsEvent.Txs[i]
					input := tx.Data() // "input" field above
					if len(input) < 4 {
						log.Info("transaction data has less than 3 fields")
						continue
					}
					method := input[:4]
					m, err := orderBookAbi.MethodById(method)
					if err == nil {
						log.Info("transaction was made by OrderBook contract")
						log.Info("transaction", "method name", m.Name)
						in := make(map[string]interface{})
						_ = m.Inputs.UnpackIntoMap(in, input[4:])
						// m.Inputs[3].UnmarshalJSON()
						if m.Name == "placeOrder" {
							log.Info("transaction", "input", in)
							order, ok := in["order"].(struct {
								Trader            common.Address `json:"trader"`
								BaseAssetQuantity *big.Int       `json:"baseAssetQuantity"`
								Price             *big.Int       `json:"price"`
								Salt              *big.Int       `json:"salt"`
							})
							signature := in["signature"].([]byte)
							log.Info("####", "type of in[order]", reflect.TypeOf(in["order"]))
							log.Info("transaction", "order", order, "ok", ok)

							var positionType string
							if order.BaseAssetQuantity.Int64() > 0 {
								positionType = "long"
							} else {
								positionType = "short"
							}
							baseAssetQuantity, _ := new(big.Float).SetInt(order.BaseAssetQuantity).Float64()
							err := lop.database.InsertLimitOrder(positionType, order.Trader.Hash().String(), int(order.BaseAssetQuantity.Uint64()), baseAssetQuantity, order.Salt.String(), signature)
							if err != nil {
								log.Error("######", "err in database.InsertLimitOrder", err)
							}
							log.Info("######", "inserted!!")
						}
						if m.Name == "executeTestOrder" {
							log.Info("transaction", "input", in)
							order, ok := in["order"].(struct {
								Trader            common.Address `json:"trader"`
								BaseAssetQuantity *big.Int       `json:"baseAssetQuantity"`
								Price             *big.Int       `json:"price"`
								Salt              *big.Int       `json:"salt"`
							})
							log.Info("####", "type of in[order]", reflect.TypeOf(in["order"]))
							log.Info("transaction", "order", order, "ok", ok)

							err := lop.database.UpdateLimitOrderStatus(order.Trader.Hash().String(), order.Salt.String(), "fulfilled")
							if err != nil {
								log.Error("######", "err in database.UpdateLimitOrderStatus", err)
							}
						}
					}
				}
			case <-lop.shutdownChan:
				return
			}
		}
	})
}

func CheckMatchingOrders(txs types.Transactions, txPool *core.TxPool) []*types.Transaction {
	jsonBytes, _ := ioutil.ReadFile("contract-examples/artifacts/contracts/OrderBook.sol/OrderBook.json")
	orderBookAbi, err := abi.FromSolidityJson(string(jsonBytes))
	if err != nil {
		panic(err)
	}

	matchingOrderTxs := []*types.Transaction{}

	for i := 0; i < len(txs); i++ {
		tx := txs[i]
		if tx.To() != nil && tx.Data() != nil && len(tx.Data()) != 0 {
			log.Info("transaction", "to is", tx.To().String())
			input := tx.Data() // "input" field above
			log.Info("transaction", "data is", input)
			if len(input) < 4 {
				log.Info("transaction data has less than 3 fields")
				continue
			}
			method := input[:4]
			m, err := orderBookAbi.MethodById(method)
			if err == nil {
				log.Info("transaction was made by OrderBook contract")
				log.Info("transaction", "method name", m.Name)
				// log.Info("transaction", "amount in is: %+v\n", in["amount"])
				// log.Info("transaction", "to is: %+v\n", in["to"])
				in := make(map[string]interface{})
				_ = m.Inputs.UnpackIntoMap(in, input[4:])
				if m.Name == "placeOrder" {
					log.Info("transaction", "input", in)
					nonce := txPool.Nonce(common.HexToAddress("0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")) // admin address

					data, err := orderBookAbi.Pack("executeTestOrder", in["order"], in["signature"])
					if err != nil {
						log.Error("abi.Pack failed", "err", err)
					}
					// log.Info("####", "data", data)
					key, err := crypto.HexToECDSA("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027") // admin private key
					if err != nil {
						log.Error("HexToECDSA failed", "err", err)
					}
					tx := types.NewTransaction(nonce, common.HexToAddress("0x52C84043CD9c865236f11d9Fc9F56aa003c1f922"), big.NewInt(0), 8000000, big.NewInt(250000000), data)
					signer := types.NewLondonSigner(big.NewInt(99999))
					signedTx, err := types.SignTx(tx, signer, key)
					if err != nil {
						log.Error("types.SignTx failed", "err", err)
					}
					matchingOrderTxs = append(matchingOrderTxs, signedTx)
				}
			}
		}
	}
	return matchingOrderTxs
}
