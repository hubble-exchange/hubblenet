package limitorders

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ava-labs/subnet-evm/accounts/abi"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/eth"
	"github.com/ava-labs/subnet-evm/internal/ethapi"
	"github.com/ava-labs/subnet-evm/rpc"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

var OrderBookContractAddress = common.HexToAddress("0x0300000000000000000000000000000000000069")
var MarginAccountContractAddress = common.HexToAddress("0x0300000000000000000000000000000000000070")
var ClearingHouseContractAddress = common.HexToAddress("0x0300000000000000000000000000000000000071")

type LimitOrderTxProcessor interface {
	ExecuteMatchedOrdersTx(incomingOrder LimitOrder, matchedOrder LimitOrder, fillAmount *big.Int) error
	PurgeLocalTx()
	CheckIfOrderBookContractCall(tx *types.Transaction) bool
	ExecuteFundingPaymentTx() error
	ExecuteLiquidation(trader common.Address, matchedOrder LimitOrder, fillAmount *big.Int) error
	GetUnderlyingPrice() ([]*big.Int, error)
}

type limitOrderTxProcessor struct {
	txPool                       *core.TxPool
	memoryDb                     LimitOrderDatabase
	orderBookABI                 abi.ABI
	clearingHouseABI             abi.ABI
	marginAccountABI             abi.ABI
	orderBookContractAddress     common.Address
	clearingHouseContractAddress common.Address
	marginAccountContractAddress common.Address
	backend                      *eth.EthAPIBackend
	validatorAddress             common.Address
	validatorPrivateKey          string
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
	orderBookABI, err := abi.FromSolidityJson(string(orderBookAbi))
	if err != nil {
		panic(err)
	}

	clearingHouseABI, err := abi.FromSolidityJson(string(clearingHouseAbi))
	if err != nil {
		panic(err)
	}

	marginAccountABI, err := abi.FromSolidityJson(string(marginAccountAbi))
	if err != nil {
		panic(err)
	}
	validatorPrivateKey := os.Getenv("GRPC_VALIDATOR_PRIVATE_KEY")
	if validatorPrivateKey == "" {
		panic("private key is not supplied")
	}
	validatorAddress, err := getAddressFromPrivateKey(validatorPrivateKey)
	if err != nil {
		panic(fmt.Sprint("unable to get address from private key with error", err.Error()))
	}

	return &limitOrderTxProcessor{
		txPool:                       txPool,
		orderBookABI:                 orderBookABI,
		clearingHouseABI:             clearingHouseABI,
		marginAccountABI:             marginAccountABI,
		memoryDb:                     memoryDb,
		orderBookContractAddress:     OrderBookContractAddress,
		clearingHouseContractAddress: ClearingHouseContractAddress,
		marginAccountContractAddress: MarginAccountContractAddress,
		backend:                      backend,
		validatorAddress:             validatorAddress,
		validatorPrivateKey:          validatorPrivateKey,
	}
}

func (lotp *limitOrderTxProcessor) ExecuteLiquidation(trader common.Address, matchedOrder LimitOrder, fillAmount *big.Int) error {
	log.Info("ExecuteLiquidation", "trader", trader, "matchedOrder", matchedOrder, "fillAmount", prettifyScaledBigInt(fillAmount, 18))
	return lotp.executeLocalTx(lotp.orderBookContractAddress, lotp.orderBookABI, "liquidateAndExecuteOrder", trader, getOrderFromRawOrder(matchedOrder.RawOrder), matchedOrder.Signature, fillAmount)
}

func (lotp *limitOrderTxProcessor) ExecuteFundingPaymentTx() error {
	log.Info("ExecuteFundingPaymentTx")
	return lotp.executeLocalTx(lotp.orderBookContractAddress, lotp.orderBookABI, "settleFunding")
}

func (lotp *limitOrderTxProcessor) ExecuteMatchedOrdersTx(incomingOrder LimitOrder, matchedOrder LimitOrder, fillAmount *big.Int) error {
	log.Info("ExecuteMatchedOrdersTx", "LongOrder", incomingOrder, "ShortOrder", matchedOrder, "fillAmount", prettifyScaledBigInt(fillAmount, 18))

	orders := make([]Order, 2)
	orders[0], orders[1] = getOrderFromRawOrder(incomingOrder.RawOrder), getOrderFromRawOrder(matchedOrder.RawOrder)

	signatures := make([][]byte, 2)
	signatures[0] = incomingOrder.Signature
	signatures[1] = matchedOrder.Signature

	return lotp.executeLocalTx(lotp.orderBookContractAddress, lotp.orderBookABI, "executeMatchedOrders", orders, signatures, fillAmount)
}

func (lotp *limitOrderTxProcessor) executeLocalTx(contract common.Address, contractABI abi.ABI, method string, args ...interface{}) error {
	nonce := lotp.txPool.GetOrderBookTxNonce(common.HexToAddress(lotp.validatorAddress.Hex())) // admin address

	data, err := contractABI.Pack(method, args...)
	if err != nil {
		log.Error("abi.Pack failed", "method", method, "args", args, "err", err)
		return err
	}
	key, err := crypto.HexToECDSA(lotp.validatorPrivateKey) // admin private key
	if err != nil {
		log.Error("HexToECDSA failed", "err", err)
		return err
	}
	tx := types.NewTransaction(nonce, contract, big.NewInt(0), 5000000, big.NewInt(70000000000), data)
	signer := types.NewLondonSigner(lotp.backend.ChainConfig().ChainID)
	signedTx, err := types.SignTx(tx, signer, key)
	if err != nil {
		log.Error("types.SignTx failed", "err", err)
	}
	err = lotp.txPool.AddOrderBookTx(signedTx)
	if err != nil {
		log.Error("lop.txPool.AddOrderBookTx failed", "err", err, "tx", signedTx.Hash().String(), "nonce", nonce)
		return err
	}
	log.Info("executeLocalTx - AddOrderBookTx success", "tx", signedTx.Hash().String(), "nonce", nonce)

	return nil
}

func (lotp *limitOrderTxProcessor) PurgeLocalTx() {
	pending := lotp.txPool.Pending(true)
	for _, txs := range pending {
		for _, tx := range txs {
			method, err := getOrderBookContractCallMethod(tx, lotp.orderBookABI, lotp.orderBookContractAddress)
			if err == nil {
				if method.Name == "executeMatchedOrders" || method.Name == "settleFunding" || method.Name == "liquidateAndExecuteOrder" {
					lotp.txPool.RemoveTx(tx.Hash())
				}
			}
		}
	}
	lotp.txPool.PurgeOrderBookTxs()
}

func (lotp *limitOrderTxProcessor) GetUnderlyingPrice() ([]*big.Int, error) {
	data, err := lotp.clearingHouseABI.Pack("getUnderlyingPrice")
	if err != nil {
		log.Error("abi.Pack failed", "method", "getUnderlyingPrice", "err", err)
		return nil, err
	}
	args := ethapi.TransactionArgs{
		To:      &lotp.clearingHouseContractAddress,
		Input:   (*hexutil.Bytes)(&data),
		ChainID: (*hexutil.Big)(lotp.backend.ChainConfig().ChainID),
	}
	blockNumber := rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(lotp.backend.LastAcceptedBlock().Number().Int64()))
	res, err := ethapi.DoCall(context.Background(), lotp.backend, args, blockNumber, nil, time.Minute, 5000000)
	if err != nil {
		return nil, err
	}
	rawData, err := hexutil.Decode("0x" + hex.EncodeToString(res.ReturnData))
	if err != nil {
		return nil, err
	}
	uintArray, err := lotp.clearingHouseABI.Unpack("getUnderlyingPrice", rawData)
	if err != nil {
		return nil, err
	}

	if len(uintArray) != 0 {
		underlyingPrices := uintArray[0].([]*big.Int)
		if len(underlyingPrices) != 0 {
			return underlyingPrices, nil
		}
	}
	return nil, fmt.Errorf("Contracts have not yet initialized")
}

func (lotp *limitOrderTxProcessor) CheckIfOrderBookContractCall(tx *types.Transaction) bool {
	return checkIfOrderBookContractCall(tx, lotp.orderBookABI, lotp.orderBookContractAddress)
}

func getPositionTypeBasedOnBaseAssetQuantity(baseAssetQuantity *big.Int) string {
	if baseAssetQuantity.Sign() == 1 {
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

func getAddressFromPrivateKey(key string) (common.Address, error) {
	privateKey, err := crypto.HexToECDSA(key) // admin private key
	if err != nil {
		return common.Address{}, err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, errors.New("unable to get address from private key")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address, nil
}
