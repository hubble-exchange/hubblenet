package limitorders

import (
	"context"
	"errors"
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

type LimitOrderTxProcessor struct {
	txPool                   *core.TxPool
	orderBookABI             abi.ABI
	memoryDb                 *InMemoryDatabase
	orderBookContractAddress common.Address
	backend                  *eth.EthAPIBackend
}

func NewLimitOrderTxProcessor(txPool *core.TxPool, orderBookABI abi.ABI, memoryDb *InMemoryDatabase, orderBookContractAddress common.Address, backend *eth.EthAPIBackend) *LimitOrderTxProcessor {
	return &LimitOrderTxProcessor{
		txPool:                   txPool,
		orderBookABI:             orderBookABI,
		memoryDb:                 memoryDb,
		orderBookContractAddress: orderBookContractAddress,
		backend:                  backend,
	}
}

func (lotp *LimitOrderTxProcessor) HandleOrderBookEvent(event *types.Log) {
	args := map[string]interface{}{}
	switch event.Topics[0] {
	case lotp.orderBookABI.Events["OrderPlaced"].ID:
		userAddress32 := event.Topics[1].String() // user's address in 32 bytes
		err := lotp.orderBookABI.UnpackIntoMap(args, "OrderPlaced", event.Data)
		if err != nil {
			log.Error("error in orderBookAbi.UnpackIntoMap", "method", "OrderPlaced", "err", err)
		}
		log.Info("HandleOrderBookEvent", "orderplaced args", args)
		order, _ := args["order"].(struct {
			Trader            common.Address `json:"trader"`
			BaseAssetQuantity *big.Int       `json:"baseAssetQuantity"`
			Price             *big.Int       `json:"price"`
			Salt              *big.Int       `json:"salt"`
		})

		lotp.memoryDb.Add(&LimitOrder{
			Market:            AvaxPerp, // @todo: get this from event
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
		lotp.memoryDb.UpdatePositionForOrder(string(signature1), args["fillAmount"].(float64))
		lotp.memoryDb.UpdatePositionForOrder(string(signature2), args["fillAmount"].(float64))
		lotp.memoryDb.Delete(signature1)
		lotp.memoryDb.Delete(signature2) // @todo: check this method after partiall fill code
	}
	log.Info("Log found", "log_.Address", event.Address.String(), "log_.BlockNumber", event.BlockNumber, "log_.Index", event.Index, "log_.TxHash", event.TxHash.String())

}

func (lotp *LimitOrderTxProcessor) HandleMarginAccountEvent(event *types.Log) {
	args := map[string]interface{}{}
	switch event.Topics[0] {
	case lotp.orderBookABI.Events["MarginAdded"].ID:
		userAddress32 := event.Topics[1].String() // user's address in 32 bytes
		err := lotp.orderBookABI.UnpackIntoMap(args, "MarginAdded", event.Data)
		if err != nil {
			log.Error("error in orderBookAbi.UnpackIntoMap", "method", "MarginAdded", "err", err)
		}
		collateral := event.Topics[2].Big()
		userAddress := common.HexToAddress(userAddress32[:2] + userAddress32[26:])
		lotp.memoryDb.UpdateMargin(userAddress, Collateral(collateral.Int64()), float64(args["amount"].(int)))
	case lotp.orderBookABI.Events["MarginRemoved"].ID:
		userAddress32 := event.Topics[1].String() // user's address in 32 bytes
		err := lotp.orderBookABI.UnpackIntoMap(args, "MarginRemoved", event.Data)
		if err != nil {
			log.Error("error in orderBookAbi.UnpackIntoMap", "method", "MarginRemoved", "err", err)
		}
		collateral := event.Topics[2].Big()
		userAddress := common.HexToAddress(userAddress32[:2] + userAddress32[26:])
		lotp.memoryDb.UpdateMargin(userAddress, Collateral(collateral.Int64()), -1*float64(args["amount"].(int)))

	}
	log.Info("Log found", "log_.Address", event.Address.String(), "log_.BlockNumber", event.BlockNumber, "log_.Index", event.Index, "log_.TxHash", event.TxHash.String())
}

func (lotp *LimitOrderTxProcessor) HandleClearingHouseEvent(event *types.Log) {
	args := map[string]interface{}{}
	switch event.Topics[0] {
	case lotp.orderBookABI.Events["FundingRateUpdated"].ID:
		log.Info("FundingRateUpdated event")
		err := lotp.orderBookABI.UnpackIntoMap(args, "FundingRateUpdated", event.Data)
		if err != nil {
			log.Error("error in orderBookAbi.UnpackIntoMap", "method", "FundingRateUpdated", "err", err)
		}
		premiumFraction := args["premiumFraction"].(int64)
		market := args["market"].(int)
		lotp.memoryDb.UpdateUnrealisedFunding(Market(market), float64(premiumFraction))
	case lotp.orderBookABI.Events["FundingPaid"].ID:
		log.Info("FundingPaid event")
		err := lotp.orderBookABI.UnpackIntoMap(args, "FundingPaid", event.Data)
		if err != nil {
			log.Error("error in orderBookAbi.UnpackIntoMap", "method", "FundingPaid", "err", err)
		}
		userAddress32 := event.Topics[1].String() // user's address in 32 bytes
		userAddress := common.HexToAddress(userAddress32[:2] + userAddress32[26:])
		market := args["market"].(int)
		lotp.memoryDb.ResetUnrealisedFunding(Market(market), userAddress)
	}
}

func (lotp *LimitOrderTxProcessor) HandleOrderBookTx(tx *types.Transaction, blockNumber uint64, backend eth.EthAPIBackend) {
	m, err := getOrderBookContractCallMethod(tx, lotp.orderBookABI, lotp.orderBookContractAddress)
	if !checkTxStatusSucess(backend, tx.Hash()) {
		// no need to parse if tx was not successful
		return
	}
	if err == nil {
		input := tx.Data()
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
			limitOrder := &LimitOrder{
				PositionType:      positionType,
				UserAddress:       order.Trader.Hash().String(),
				BaseAssetQuantity: baseAssetQuantity,
				Price:             price,
				Status: "unfulfilled",
				Signature:         signature,
				BlockNumber:  blockNumber,
				RawOrder:     in["order"],
				RawSignature: in["signature"],
			}
			lotp.memoryDb.Add(limitOrder)
		}
		if m.Name == "executeMatchedOrders" {
			// @todo: change args once the contract is updated
			signature1 := in["signature1"].([]byte)
			signature2 := in["signature2"].([]byte)
			// fillAmount := in["fillAmount"].(int64)

			// lotp.memoryDb.UpdatePositionForOrder(string(signature1), float64(fillAmount))
			// lotp.memoryDb.UpdatePositionForOrder(string(signature2), float64(fillAmount))
			lotp.memoryDb.Delete(signature1)
			lotp.memoryDb.Delete(signature2)
		}
		if m.Name == "settleFunding" {
			// funding payment was successful, so update the nnext funding time now
			lotp.memoryDb.UpdateNextFundingTime()
		}
		if m.Name == "addMargin" {
			// funding payment was successful, so update the nnext funding time now
			// lotp.memoryDb.UpdateMargin()
		}
	}
}

func (lotp *LimitOrderTxProcessor) ExecuteLiquidation(trader common.Address, matchedOrder LimitOrder) error {
	nonce := lotp.txPool.Nonce(common.HexToAddress(userAddress1)) // admin address

	data, err := lotp.orderBookABI.Pack("liquidateAndExecuteOrder", trader.String(), matchedOrder.RawOrder, matchedOrder.Signature)
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

func (lotp *LimitOrderTxProcessor) ExecuteFundingPaymentTx() error {

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

func (lotp *LimitOrderTxProcessor) ExecuteMatchedOrdersTx(incomingOrder LimitOrder, matchedOrder LimitOrder) error {
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

	data, err := lotp.orderBookABI.Pack("executeMatchedOrders", incomingOrder.RawOrder, incomingOrder.Signature, matchedOrder.RawOrder, matchedOrder.Signature)
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

func (lotp *LimitOrderTxProcessor) PurgeLocalTx() {
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
func (lotp *LimitOrderTxProcessor) CheckIfOrderBookContractCall(tx *types.Transaction) bool {
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
