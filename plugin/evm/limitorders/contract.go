package limitorders

import (
	"context"
	"math/big"

	"github.com/ava-labs/subnet-evm/internal/ethapi"
	"github.com/ava-labs/subnet-evm/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
)

const gasCap = 5000000
const maxFeePerGas = 70000000000

func (lotp *limitOrderTxProcessor) GetLastPrice(market Market) []byte {
	from := common.HexToAddress("0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	nonce := hexutil.Uint64(lotp.txPool.Nonce(from))
	// gasLimit := hexutil.Uint64(5000000)
	// gasPrice := big.NewInt(5000000)
	data, err := lotp.orderBookABI.Pack("getLastPrice", int(market))
	if err != nil {
		panic(err)
	}
	// gasPrice := hexutil.Uint64(5000000)
	args := ethapi.TransactionArgs{
		From:         &from,
		To:           &lotp.orderBookContractAddress,
		GasPrice:     nil,
		MaxFeePerGas: (*hexutil.Big)(big.NewInt(maxFeePerGas)),
		Nonce:        &nonce,
		Input:        (*hexutil.Bytes)(&data),
		ChainID:      (*hexutil.Big)(big.NewInt(321123)),
	}
	blockNumber := rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(lotp.backend.LastAcceptedBlock().Number().Int64()))
	// res, err := ethapi.DoCall(context.Background(), lotp.backend, args, blockNumber, nil, time.Minute, 5000000)

	bcApi := ethapi.NewBlockChainAPI(lotp.backend)
	res, err := bcApi.Call(context.Background(), args, blockNumber, nil)
	log.Info("GetLastPrice ethapi.DoCall", "res", res, "err", err)
	if err == nil {
		log.Info("GetLastPrice ethapi.DoCall result", "res", res.String())
		mapp := map[string]interface{}{}
		bytes, _ := res.MarshalText()
		lotp.orderBookABI.UnpackIntoMap(mapp, "getLastPrice", bytes)
		log.Info("GetLastPrice ethapi.DoCall mapp", "mapp", mapp)
	}
	return nil
}
