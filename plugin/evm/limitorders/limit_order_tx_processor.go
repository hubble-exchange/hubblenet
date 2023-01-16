package limitorders

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"math/rand"
	"time"

	"github.com/ava-labs/subnet-evm/accounts/abi"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/eth"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

// using multiple private keys to make executeMatchedOrders contract call.
// This will be replaced by validator's private key and address
var userAddress1 = "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"
var privateKey1 = "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"
var userAddress2 = "0x4Cf2eD3665F6bFA95cE6A11CFDb7A2EF5FC1C7E4"
var privateKey2 = "31b571bf6894a248831ff937bb49f7754509fe93bbd2517c9c73c4144c0e97dc"

var orderBookContractFileLocation = "contract-examples/artifacts/contracts/hubble-v2/OrderBook.sol/OrderBook.json"
var marginAccountContractFileLocation = "contract-examples/artifacts/contracts/hubble-v2/MarginAccount.sol/MarginAccount.json"
var clearingHouseContractFileLocation = "contract-examples/artifacts/contracts/hubble-v2/ClearingHouse.sol/ClearingHouse.json"
var OrderBookContractAddress = common.HexToAddress("0x0300000000000000000000000000000000000069")
var MarginAccountContractAddress = common.HexToAddress("0x0300000000000000000000000000000000000070")
var ClearingHouseContractAddress = common.HexToAddress("0x0300000000000000000000000000000000000071")

func SetOrderBookContractFileLocation(location string) {
	orderBookContractFileLocation = location
}

type LimitOrderTxProcessor interface {
	ExecuteMatchedOrdersTx(incomingOrder LimitOrder, matchedOrder LimitOrder, fillAmount uint) error
	PurgeLocalTx()
	CheckIfOrderBookContractCall(tx *types.Transaction) bool
	ExecuteFundingPaymentTx() error
	ExecuteLiquidation(trader common.Address, matchedOrder LimitOrder, fillAmount uint) error
	HandleOrderBookEvent(event *types.Log)
	HandleMarginAccountEvent(event *types.Log)
	HandleClearingHouseEvent(event *types.Log)
}

type limitOrderTxProcessor struct {
	txPool                       *core.TxPool
	memoryDb                     LimitOrderDatabase
	orderBookABI                 abi.ABI
	marginAccountABI             abi.ABI
	clearingHouseABI             abi.ABI
	marginAccountContractAddress common.Address
	clearingHouseContractAddress common.Address
	orderBookContractAddress     common.Address
	backend                      *eth.EthAPIBackend
}

// Order type is copy of Order struct defined in Orderbook contract
type Order struct {
	Trader            common.Address `json:"trader"`
	AmmIndex          *big.Int       `json:"ammIndex"`
	BaseAssetQuantity *big.Int       `json:"baseAssetQuantity"`
	Price             *big.Int       `json:"price"`
	Salt              *big.Int       `json:"salt"`
}

func NewLimitOrderTxProcessor(txPool *core.TxPool, memoryDb LimitOrderDatabase, backend *eth.EthAPIBackend) LimitOrderTxProcessor {
	jsonBytes, _ := ioutil.ReadFile(orderBookContractFileLocation)
	orderBookABI, err := abi.FromSolidityJson(string(jsonBytes))
	if err != nil {
		panic(err)
	}

	jsonBytes, _ = ioutil.ReadFile(marginAccountContractFileLocation)
	marginAccountABI, err := abi.FromSolidityJson(string(jsonBytes))
	if err != nil {
		panic(err)
	}

	jsonBytes, _ = ioutil.ReadFile(clearingHouseContractFileLocation)
	clearingHouseABI, err := abi.FromSolidityJson(string(jsonBytes))
	if err != nil {
		panic(err)
	}

	return &limitOrderTxProcessor{
		txPool:                       txPool,
		orderBookABI:                 orderBookABI,
		marginAccountABI:             marginAccountABI,
		clearingHouseABI:             clearingHouseABI,
		memoryDb:                     memoryDb,
		orderBookContractAddress:     OrderBookContractAddress,
		marginAccountContractAddress: MarginAccountContractAddress,
		clearingHouseContractAddress: ClearingHouseContractAddress,
		backend:                      backend,
	}
}

func (lotp *limitOrderTxProcessor) HandleOrderBookEvent(event *types.Log) {
	args := map[string]interface{}{}
	switch event.Topics[0] {
	case lotp.orderBookABI.Events["OrderPlaced"].ID:
		userAddress32 := event.Topics[1].String() // user's address in 32 bytes
		err := lotp.orderBookABI.UnpackIntoMap(args, "OrderPlaced", event.Data)
		if err != nil {
			log.Error("error in orderBookAbi.UnpackIntoMap", "method", "OrderPlaced", "err", err)
		}
		log.Info("HandleOrderBookEvent", "orderplaced args", args)
		order := getOrderFromRawOrder(args["order"])

		lotp.memoryDb.Add(&LimitOrder{
			Market:            Market(order.AmmIndex.Int64()),
			PositionType:      getPositionTypeBasedOnBaseAssetQuantity(int(order.BaseAssetQuantity.Int64())),
			UserAddress:       userAddress32[:2] + userAddress32[26:], // removes 0 padding
			BaseAssetQuantity: int(order.BaseAssetQuantity.Int64()),
			Price:             float64(order.Price.Int64()),
			Status:            Unfulfilled,
			RawOrder:          args["order"],
			RawSignature:      args["signature"],
			Signature:         args["signature"].([]byte),
			BlockNumber:       event.BlockNumber,
		})
	case lotp.orderBookABI.Events["OrdersMatched"].ID:
		log.Info("OrdersMatched event")
		err := lotp.orderBookABI.UnpackIntoMap(args, "OrdersMatched", event.Data)
		if err != nil {
			log.Error("error in orderBookAbi.UnpackIntoMap", "method", "OrdersMatched", "err", err)
		}
		log.Info("HandleOrderBookEvent", "OrdersMatched args", args)
		signature1 := args["signature1"].([]byte)
		signature2 := args["signature2"].([]byte)
		fillAmount := args["fillAmount"].(*big.Int).Int64()
		lotp.memoryDb.UpdateFilledBaseAssetQuantity(uint(fillAmount), signature1)
		lotp.memoryDb.UpdateFilledBaseAssetQuantity(uint(fillAmount), signature2)
	}
	log.Info("Log found", "log_.Address", event.Address.String(), "log_.BlockNumber", event.BlockNumber, "log_.Index", event.Index, "log_.TxHash", event.TxHash.String())

}

func (lotp *limitOrderTxProcessor) HandleMarginAccountEvent(event *types.Log) {
	args := map[string]interface{}{}
	switch event.Topics[0] {
	case lotp.marginAccountABI.Events["MarginAdded"].ID:
		userAddress32 := event.Topics[1].String() // user's address in 32 bytes
		err := lotp.marginAccountABI.UnpackIntoMap(args, "MarginAdded", event.Data)
		if err != nil {
			log.Error("error in marginAccountABI.UnpackIntoMap", "method", "MarginAdded", "err", err)
		}
		collateral := event.Topics[2].Big()
		userAddress := common.HexToAddress(userAddress32[:2] + userAddress32[26:])
		lotp.memoryDb.UpdateMargin(userAddress, Collateral(collateral.Int64()), float64(args["amount"].(int)))
	case lotp.marginAccountABI.Events["MarginRemoved"].ID:
		userAddress32 := event.Topics[1].String() // user's address in 32 bytes
		err := lotp.marginAccountABI.UnpackIntoMap(args, "MarginRemoved", event.Data)
		if err != nil {
			log.Error("error in marginAccountABI.UnpackIntoMap", "method", "MarginRemoved", "err", err)
		}
		collateral := event.Topics[2].Big()
		userAddress := common.HexToAddress(userAddress32[:2] + userAddress32[26:])
		lotp.memoryDb.UpdateMargin(userAddress, Collateral(collateral.Int64()), -1*float64(args["amount"].(int)))
	case lotp.marginAccountABI.Events["PnLRealized"].ID:
		userAddress32 := event.Topics[1].String() // user's address in 32 bytes
		err := lotp.marginAccountABI.UnpackIntoMap(args, "PnLRealized", event.Data)
		if err != nil {
			log.Error("error in marginAccountABI.UnpackIntoMap", "method", "PnLRealized", "err", err)
		}
		userAddress := common.HexToAddress(userAddress32[:2] + userAddress32[26:])
		realisedPnL := float64(args["realizedPnl"].(*big.Int).Int64())

		lotp.memoryDb.UpdateMargin(userAddress, USDC, realisedPnL)
	}
	log.Info("Log found", "log_.Address", event.Address.String(), "log_.BlockNumber", event.BlockNumber, "log_.Index", event.Index, "log_.TxHash", event.TxHash.String())
}

func (lotp *limitOrderTxProcessor) HandleClearingHouseEvent(event *types.Log) {
	args := map[string]interface{}{}
	switch event.Topics[0] {
	case lotp.clearingHouseABI.Events["FundingRateUpdated"].ID:
		log.Info("FundingRateUpdated event")
		err := lotp.clearingHouseABI.UnpackIntoMap(args, "FundingRateUpdated", event.Data)
		if err != nil {
			log.Error("error in clearingHouseABI.UnpackIntoMap", "method", "FundingRateUpdated", "err", err)
		}
		cumulativePremiumFraction := args["cumulativePremiumFraction"].(*big.Int)
		nextFundingTime := args["nextFundingTime"].(*big.Int)
		market := Market(int(event.Topics[1].Big().Int64()))
		lotp.memoryDb.UpdateUnrealisedFunding(Market(market), float64(cumulativePremiumFraction.Int64()))
		lotp.memoryDb.UpdateNextFundingTime(nextFundingTime.Uint64())

	case lotp.clearingHouseABI.Events["FundingPaid"].ID:
		log.Info("FundingPaid event")
		err := lotp.clearingHouseABI.UnpackIntoMap(args, "FundingPaid", event.Data)
		if err != nil {
			log.Error("error in clearingHouseABI.UnpackIntoMap", "method", "FundingPaid", "err", err)
		}
		userAddress32 := event.Topics[1].String() // user's address in 32 bytes
		userAddress := common.HexToAddress(userAddress32[:2] + userAddress32[26:])
		market := Market(int(event.Topics[2].Big().Int64()))
		cumulativePremiumFraction := args["cumulativePremiumFraction"].(*big.Int)
		lotp.memoryDb.ResetUnrealisedFunding(Market(market), userAddress, float64(cumulativePremiumFraction.Int64()))

	// both PositionModified and PositionLiquidated have the exact same signature
	case lotp.clearingHouseABI.Events["PositionModified"].ID, lotp.clearingHouseABI.Events["PositionLiquidated"].ID:
		log.Info("PositionModified event")
		err := lotp.clearingHouseABI.UnpackIntoMap(args, "PositionModified", event.Data)
		if err != nil {
			log.Error("error in clearingHouseABI.UnpackIntoMap", "method", "PositionModified", "err", err)
		}

		market := Market(int(event.Topics[2].Big().Int64()))
		baseAsset := args["baseAsset"].(*big.Int).Int64()
		quoteAsset := args["quoteAsset"].(*big.Int).Int64()
		lastPrice := float64(quoteAsset) / float64(baseAsset)
		lotp.memoryDb.UpdateLastPrice(market, lastPrice)

		userAddress32 := event.Topics[1].String() // user's address in 32 bytes
		userAddress := common.HexToAddress(userAddress32[:2] + userAddress32[26:])
		openNotional := float64(args["openNotional"].(*big.Int).Int64())
		size := float64(args["size"].(*big.Int).Int64())
		lotp.memoryDb.UpdatePosition(userAddress, market, size, openNotional)
	}
}

func (lotp *limitOrderTxProcessor) ExecuteLiquidation(trader common.Address, matchedOrder LimitOrder, fillAmount uint) error {
	nonce := lotp.txPool.Nonce(common.HexToAddress(userAddress1)) // admin address

	data, err := lotp.orderBookABI.Pack("liquidateAndExecuteOrder", trader.String(), matchedOrder.RawOrder, matchedOrder.Signature, fillAmount)
	if err != nil {
		log.Error("abi.Pack failed", "err", err)
		return err
	}
	key, err := crypto.HexToECDSA(privateKey1) // admin private key
	if err != nil {
		log.Error("HexToECDSA failed", "err", err)
		return err
	}
	executeMatchedOrdersTx := types.NewTransaction(nonce, lotp.orderBookContractAddress, big.NewInt(0), 5000000, big.NewInt(80000000000), data)
	signer := types.NewLondonSigner(big.NewInt(321123))
	signedTx, err := types.SignTx(executeMatchedOrdersTx, signer, key)
	if err != nil {
		log.Error("types.SignTx failed", "err", err)
	}
	err = lotp.txPool.AddLocal(signedTx)
	if err != nil {
		log.Error("lop.txPool.AddLocal failed", "err", err)
		return err
	}
	return nil
}

func (lotp *limitOrderTxProcessor) ExecuteFundingPaymentTx() error {

	nonce := lotp.txPool.Nonce(common.HexToAddress("0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")) // admin address

	data, err := lotp.orderBookABI.Pack("settleFunding")
	if err != nil {
		log.Error("abi.Pack failed", "err", err)
		return err
	}
	key, err := crypto.HexToECDSA("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027") // admin private key
	if err != nil {
		log.Error("HexToECDSA failed", "err", err)
		return err
	}
	settleFundingTx := types.NewTransaction(nonce, lotp.orderBookContractAddress, big.NewInt(0), 5000000, big.NewInt(80000000000), data)
	signer := types.NewLondonSigner(big.NewInt(321123))
	signedTx, err := types.SignTx(settleFundingTx, signer, key)
	if err != nil {
		log.Error("types.SignTx failed", "err", err)
		return err
	}
	err = lotp.txPool.AddLocal(signedTx)
	if err != nil {
		log.Error("types.SignTx failed", "err", err)
		return err
	}
	return nil

}

func (lotp *limitOrderTxProcessor) ExecuteMatchedOrdersTx(incomingOrder LimitOrder, matchedOrder LimitOrder, fillAmount uint) error {
	//randomly selecting private key to get different validator profile on different nodes
	rand.Seed(time.Now().UnixNano())
	var privateKey, userAddress string
	if rand.Intn(10000)%2 == 0 {
		privateKey = privateKey1
		userAddress = userAddress1
	} else {
		privateKey = privateKey2
		userAddress = userAddress2
	}

	nonce := lotp.txPool.Nonce(common.HexToAddress(userAddress)) // admin address
	ammID := big.NewInt(1)
	orders := make([]Order, 2)
	orders[0], orders[1] = getOrderFromRawOrder(incomingOrder.RawOrder), getOrderFromRawOrder(matchedOrder.RawOrder)
	signatures := make([][]byte, 2)
	signatures[0] = incomingOrder.Signature
	signatures[1] = matchedOrder.Signature

	data, err := lotp.orderBookABI.Pack("executeMatchedOrders", ammID, orders, signatures, big.NewInt(int64(fillAmount)))
	if err != nil {
		log.Error("abi.Pack failed", "err", err)
		return err
	}
	key, err := crypto.HexToECDSA(privateKey) // admin private key
	if err != nil {
		log.Error("HexToECDSA failed", "err", err)
		return err
	}
	executeMatchedOrdersTx := types.NewTransaction(nonce, lotp.orderBookContractAddress, big.NewInt(0), 5000000, big.NewInt(80000000000), data)
	signer := types.NewLondonSigner(big.NewInt(321123))
	signedTx, err := types.SignTx(executeMatchedOrdersTx, signer, key)
	if err != nil {
		log.Error("types.SignTx failed", "err", err)
	}
	err = lotp.txPool.AddLocal(signedTx)
	if err != nil {
		log.Error("lop.txPool.AddLocal failed", "err", err)
		return err
	}
	return nil
}

func (lotp *limitOrderTxProcessor) PurgeLocalTx() {
	pending := lotp.txPool.Pending(true)
	localAccounts := []common.Address{common.HexToAddress(userAddress1), common.HexToAddress(userAddress2)}

	for _, account := range localAccounts {
		if txs := pending[account]; len(txs) > 0 {
			for _, tx := range txs {
				m, err := getOrderBookContractCallMethod(tx, lotp.orderBookABI, lotp.orderBookContractAddress)
				if err == nil && m.Name == "executeMatchedOrders" {
					lotp.txPool.RemoveTx(tx.Hash())
				}
			}
		}
	}
}
func (lotp *limitOrderTxProcessor) CheckIfOrderBookContractCall(tx *types.Transaction) bool {
	return checkIfOrderBookContractCall(tx, lotp.orderBookABI, lotp.orderBookContractAddress)
}

func getPositionTypeBasedOnBaseAssetQuantity(baseAssetQuantity int) string {
	if baseAssetQuantity > 0 {
		return "long"
	}
	return "short"
}

func checkTxStatusSucess(backend eth.EthAPIBackend, hash common.Hash) bool {
	ctx := context.Background()
	defer ctx.Done()

	_, blockHash, _, index, err := backend.GetTransaction(ctx, hash)
	if err != nil {
		log.Error("err in lop.backend.GetTransaction", "err", err)
		return false
	}
	receipts, err := backend.GetReceipts(ctx, blockHash)
	if err != nil {
		log.Error("err in lop.backend.GetReceipts", "err", err)
		return false
	}
	if len(receipts) <= int(index) {
		return false
	}
	receipt := receipts[index]
	return receipt.Status == uint64(1)
}

func checkIfOrderBookContractCall(tx *types.Transaction, orderBookABI abi.ABI, orderBookContractAddress common.Address) bool {
	input := tx.Data()
	if tx.To() != nil && tx.To().Hash() == orderBookContractAddress.Hash() && len(input) > 3 {
		return true
	}
	return false
}

func getOrderBookContractCallMethod(tx *types.Transaction, orderBookABI abi.ABI, orderBookContractAddress common.Address) (*abi.Method, error) {
	if checkIfOrderBookContractCall(tx, orderBookABI, orderBookContractAddress) {
		input := tx.Data()
		method := input[:4]
		m, err := orderBookABI.MethodById(method)
		return m, err
	} else {
		err := errors.New("tx is not an orderbook contract call")
		return nil, err
	}
}

func getOrderFromRawOrder(rawOrder interface{}) Order {
	order := Order{}
	marshalledOrder, _ := json.Marshal(rawOrder)
	_ = json.Unmarshal(marshalledOrder, &order)
	return order
}
