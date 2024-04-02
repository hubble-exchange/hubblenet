package orderbook

import (
	"encoding/json"
	"math/big"

	"github.com/ava-labs/subnet-evm/accounts/abi"
	"github.com/ava-labs/subnet-evm/core/state"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/metrics"
	"github.com/ava-labs/subnet-evm/plugin/evm/orderbook/abis"
	"github.com/ava-labs/subnet-evm/precompile/contracts/bibliophile"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

var (
	getMatchingTxsErrorCounter   = metrics.NewRegisteredCounter("GetMatchingTxs_errors", nil)
	getMatchingTxsWarningCounter = metrics.NewRegisteredCounter("GetMatchingTxs_warnings", nil)
)

type TempMatcher struct {
	db                LimitOrderDatabase
	tempDB            LimitOrderDatabase
	lotp              LimitOrderTxProcessor
	orderBookABI      abi.ABI
	limitOrderBookABI abi.ABI
	iocOrderBookABI   abi.ABI
}

func NewTempMatcher(db LimitOrderDatabase, lotp LimitOrderTxProcessor) *TempMatcher {
	orderBookABI, err := abi.FromSolidityJson(string(abis.OrderBookAbi))
	if err != nil {
		panic(err)
	}

	limitOrderBookABI, err := abi.FromSolidityJson(string(abis.LimitOrderBookAbi))
	if err != nil {
		panic(err)
	}

	iocOrderBookABI, err := abi.FromSolidityJson(string(abis.IOCOrderBookAbi))
	if err != nil {
		panic(err)
	}

	return &TempMatcher{
		db:                db,
		lotp:              lotp,
		orderBookABI:      orderBookABI,
		limitOrderBookABI: limitOrderBookABI,
		iocOrderBookABI:   iocOrderBookABI,
	}
}

func (matcher *TempMatcher) GetMatchingTxs(tx *types.Transaction, stateDB *state.StateDB, blockNumber *big.Int) map[common.Address]types.Transactions {
	var isError bool
	defer func() {
		if isError {
			getMatchingTxsErrorCounter.Inc(1)
		}
	}()

	to := tx.To()

	if to == nil || len(tx.Data()) < 4 {
		return nil
	}

	method := tx.Data()[:4]
	methodData := tx.Data()[4:]

	var err error
	var markets []Market
	if matcher.tempDB == nil {
		matcher.tempDB, err = matcher.db.GetOrderBookDataCopy()
		if err != nil {
			log.Error("GetMatchingTxs: error in fetching tempDB", "err", err)
			isError = true
			return nil
		}
	}
	switch *to {
	case LimitOrderBookContractAddress:
		abiMethod, err := matcher.limitOrderBookABI.MethodById(method)
		if err != nil {
			log.Error("GetMatchingTxs: error in fetching abiMethod", "err", err)
			isError = true
			return nil
		}

		// check for placeOrders and cancelOrders txs
		switch abiMethod.Name {
		case "placeOrders":
			orders, err := getLimitOrdersFromMethodData(abiMethod, methodData, blockNumber)
			if err != nil {
				log.Error("GetMatchingTxs: error in fetching orders from placeOrders tx data", "err", err)
				isError = true
				return nil
			}
			marketsMap := make(map[Market]struct{})
			for _, order := range orders {
				// the transaction in the args is supposed to be already committed in the db, so the status should be placed
				status := bibliophile.GetOrderStatus(stateDB, order.Id)
				if status != 1 { // placed
					log.Warn("GetMatchingTxs: invalid limit order status", "status", status, "order", order.Id.String())
					getMatchingTxsWarningCounter.Inc(1)
					continue
				}

				matcher.tempDB.Add(order)
				marketsMap[order.Market] = struct{}{}
			}

			markets = make([]Market, 0, len(marketsMap))
			for market := range marketsMap {
				markets = append(markets, market)
			}

		case "cancelOrders":
			orders, err := getLimitOrdersFromMethodData(abiMethod, methodData, blockNumber)
			if err != nil {
				log.Error("GetMatchingTxs: error in fetching orders from cancelOrders tx data", "err", err)
				isError = true
				return nil
			}
			for _, order := range orders {
				if err := matcher.tempDB.SetOrderStatus(order.Id, Cancelled, "", blockNumber.Uint64()); err != nil {
					log.Error("GetMatchingTxs: error in SetOrderStatus", "orderId", order.Id.String(), "err", err)
					return nil
				}
			}
			// no need to run matching
			return nil
		default:
			return nil
		}

	case IOCOrderBookContractAddress:
		abiMethod, err := matcher.iocOrderBookABI.MethodById(method)
		if err != nil {
			log.Error("Error in fetching abiMethod", "err", err)
			isError = true
			return nil
		}

		switch abiMethod.Name {
		case "placeOrders":
			orders, err := getIOCOrdersFromMethodData(abiMethod, methodData, blockNumber)
			if err != nil {
				log.Error("Error in fetching orders", "err", err)
				isError = true
				return nil
			}
			marketsMap := make(map[Market]struct{})
			for _, order := range orders {
				// the transaction in the args is supposed to be already committed in the db, so the status should be placed
				status := bibliophile.IOCGetOrderStatus(stateDB, order.Id)
				if status != 1 { // placed
					log.Warn("GetMatchingTxs: invalid ioc order status", "status", status, "order", order.Id.String())
					getMatchingTxsWarningCounter.Inc(1)
					continue
				}

				matcher.tempDB.Add(order)
				marketsMap[order.Market] = struct{}{}
			}

			markets = make([]Market, 0, len(marketsMap))
			for market := range marketsMap {
				markets = append(markets, market)
			}
		default:
			return nil
		}
	default:
		// tx is not related to orderbook
		return nil
	}

	configService := NewConfigServiceFromStateDB(stateDB)
	tempMatchingPipeline := NewTemporaryMatchingPipeline(matcher.tempDB, matcher.lotp, configService)

	return tempMatchingPipeline.GetOrderMatchingTransactions(blockNumber, markets)
}

func (matcher *TempMatcher) ResetMemoryDB() {
	matcher.tempDB = nil
}

func getLimitOrdersFromMethodData(abiMethod *abi.Method, methodData []byte, blockNumber *big.Int) ([]*Order, error) {
	unpackedData, err := abiMethod.Inputs.Unpack(methodData)
	if err != nil {
		log.Error("Error in unpacking data", "err", err)
		return nil, err
	}

	limitOrders := []*LimitOrder{}
	ordersInterface := unpackedData[0]

	marshalledOrders, _ := json.Marshal(ordersInterface)
	err = json.Unmarshal(marshalledOrders, &limitOrders)
	if err != nil {
		log.Error("Error in unmarshalling orders", "err", err)
		return nil, err
	}

	orders := []*Order{}
	for _, limitOrder := range limitOrders {
		orderId, err := limitOrder.Hash()
		if err != nil {
			log.Error("Error in hashing order", "err", err)
			// @todo: send to metrics
			return nil, err
		}

		order := &Order{
			Id:                      orderId,
			Market:                  Market(limitOrder.AmmIndex.Int64()),
			PositionType:            getPositionTypeBasedOnBaseAssetQuantity(limitOrder.BaseAssetQuantity),
			Trader:                  limitOrder.Trader,
			BaseAssetQuantity:       limitOrder.BaseAssetQuantity,
			FilledBaseAssetQuantity: big.NewInt(0),
			Price:                   limitOrder.Price,
			RawOrder:                limitOrder,
			Salt:                    limitOrder.Salt,
			ReduceOnly:              limitOrder.ReduceOnly,
			BlockNumber:             blockNumber,
			OrderType:               Limit,
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func getIOCOrdersFromMethodData(abiMethod *abi.Method, methodData []byte, blockNumber *big.Int) ([]*Order, error) {
	unpackedData, err := abiMethod.Inputs.Unpack(methodData)
	if err != nil {
		log.Error("Error in unpacking data", "err", err)
		return nil, err
	}

	iocOrders := []*IOCOrder{}
	ordersInterface := unpackedData[0]

	marshalledOrders, _ := json.Marshal(ordersInterface)
	err = json.Unmarshal(marshalledOrders, &iocOrders)
	if err != nil {
		log.Error("Error in unmarshalling orders", "err", err)
		return nil, err
	}

	orders := []*Order{}
	for _, iocOrder := range iocOrders {
		orderId, err := iocOrder.Hash()
		if err != nil {
			log.Error("Error in hashing order", "err", err)
			// @todo: send to metrics
			return nil, err
		}

		order := &Order{
			Id:                      orderId,
			Market:                  Market(iocOrder.AmmIndex.Int64()),
			PositionType:            getPositionTypeBasedOnBaseAssetQuantity(iocOrder.BaseAssetQuantity),
			Trader:                  iocOrder.Trader,
			BaseAssetQuantity:       iocOrder.BaseAssetQuantity,
			FilledBaseAssetQuantity: big.NewInt(0),
			Price:                   iocOrder.Price,
			RawOrder:                iocOrder,
			Salt:                    iocOrder.Salt,
			ReduceOnly:              iocOrder.ReduceOnly,
			BlockNumber:             blockNumber,
			OrderType:               Limit,
		}
		orders = append(orders, order)
	}

	return orders, nil
}
