package limitorders

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ava-labs/subnet-evm/accounts/abi"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

type ContractEventsProcessor struct {
	orderBookABI     abi.ABI
	marginAccountABI abi.ABI
	clearingHouseABI abi.ABI
	database         LimitOrderDatabase
}

func NewContractEventsProcessor(database LimitOrderDatabase) *ContractEventsProcessor {
	orderBookABI, err := abi.FromSolidityJson(string(orderBookAbi))
	if err != nil {
		panic(err)
	}

	marginAccountABI, err := abi.FromSolidityJson(string(marginAccountAbi))
	if err != nil {
		panic(err)
	}

	clearingHouseABI, err := abi.FromSolidityJson(string(clearingHouseAbi))
	if err != nil {
		panic(err)
	}
	return &ContractEventsProcessor{
		orderBookABI:     orderBookABI,
		marginAccountABI: marginAccountABI,
		clearingHouseABI: clearingHouseABI,
		database:         database,
	}
}

func (cep *ContractEventsProcessor) ProcessEvents(logs []*types.Log, removed bool) {
	// removed logs are received in increasing order of block number
	// but the way that we have written our logic they are best processed in the opposite order
	if removed {
		reversedLogs := make([]*types.Log, 0, len(logs))
		for i := len(logs) - 1; i >= 0; i-- {
			reversedLogs = append(reversedLogs, logs[i])
		}
		logs = reversedLogs
	}
	for _, event := range logs {
		switch event.Address {
		case OrderBookContractAddress:
			cep.handleOrderBookEvent(event, removed)
		case MarginAccountContractAddress:
			cep.handleMarginAccountEvent(event)
		case ClearingHouseContractAddress:
			cep.handleClearingHouseEvent(event)
		}
	}
}

func parseOrderId(orderHash interface{}) common.Hash {
	_orderId, _ := orderHash.([32]byte)
	return common.BytesToHash(_orderId[:])
}

func (cep *ContractEventsProcessor) handleOrderBookEvent(event *types.Log, removed bool) {
	args := map[string]interface{}{}
	switch event.Topics[0] {
	case cep.orderBookABI.Events["OrderPlaced"].ID:
		err := cep.orderBookABI.UnpackIntoMap(args, "OrderPlaced", event.Data)
		if err != nil {
			log.Error("error in orderBookAbi.UnpackIntoMap", "method", "OrderPlaced", "err", err)
			return
		}
		log.Info("HandleOrderBookEvent", "orderplaced args", args, "removed", removed)
		orderId := parseOrderId(args["orderHash"])
		if !removed {
			order := getOrderFromRawOrder(args["order"])
			log.Info("#### adding order", "orderId", orderId.String(), "block", event.BlockHash.String(), "number", event.BlockNumber)
			cep.database.Add(orderId, &LimitOrder{
				Market:                  Market(order.AmmIndex.Int64()),
				PositionType:            getPositionTypeBasedOnBaseAssetQuantity(order.BaseAssetQuantity),
				UserAddress:             getAddressFromTopicHash(event.Topics[1]).String(),
				BaseAssetQuantity:       order.BaseAssetQuantity,
				FilledBaseAssetQuantity: big.NewInt(0),
				Price:                   order.Price,
				Status:                  Placed,
				RawOrder:                args["order"],
				Signature:               args["signature"].([]byte),
				Salt:                    order.Salt,
				BlockNumber:             big.NewInt(int64(event.BlockNumber)),
			})
		} else {
			log.Info("#### deleting order", "orderId", orderId, "block", event.BlockHash.String(), "number", event.BlockNumber)
			cep.database.Delete(orderId)
		}
		SendTxReadySignal() // what does this do?
	case cep.orderBookABI.Events["OrderCancelled"].ID:
		err := cep.orderBookABI.UnpackIntoMap(args, "OrderCancelled", event.Data)
		if err != nil {
			log.Error("error in orderBookAbi.UnpackIntoMap", "method", "OrderCancelled", "err", err)
			return
		}
		log.Info("HandleOrderBookEvent", "OrderCancelled args", args, "removed", removed)
		orderId := parseOrderId(args["orderHash"])
		if !removed {
			cep.database.GetOrderBookData().OrderMap[orderId].Status = Cancelled
		} else {
			// orders that are already fulfilled will be marked Placed as well;
			// however they will not be used for matching as long as we filter by unfilled base asset quantity for that order
			cep.database.GetOrderBookData().OrderMap[orderId].Status = Placed
		}
	case cep.orderBookABI.Events["OrdersMatched"].ID:
		err := cep.orderBookABI.UnpackIntoMap(args, "OrdersMatched", event.Data)
		if err != nil {
			log.Error("error in orderBookAbi.UnpackIntoMap", "method", "OrdersMatched", "err", err)
			return
		}

		order0Id := parseOrderId(args["orderHash"].([2][32]byte)[0])
		order1Id := parseOrderId(args["orderHash"].([2][32]byte)[1])
		fmt.Printf("matching order %s and %s", order0Id.String(), order1Id.String())
		// order1Id := args["orderHash"].([]common.Hash)[1]
		fillAmount := args["fillAmount"].(*big.Int)
		if !removed {
			log.Info("#### matched orders", "orderId_0", order0Id.String(), "orderId_1", order1Id, "block", event.BlockHash.String(), "number", event.BlockNumber)
			cep.database.UpdateFilledBaseAssetQuantity(fillAmount, order0Id, event.BlockNumber)
			cep.database.UpdateFilledBaseAssetQuantity(fillAmount, order1Id, event.BlockNumber)
		} else {
			fillAmount.Neg(fillAmount)
			log.Info("#### removed matched orders", "orderId_0", order0Id.String(), "orderId_1", order1Id, "block", event.BlockHash.String(), "number", event.BlockNumber)
			cep.database.UpdateFilledBaseAssetQuantity(fillAmount, order0Id, event.BlockNumber)
			cep.database.UpdateFilledBaseAssetQuantity(fillAmount, order1Id, event.BlockNumber)
		}
	case cep.orderBookABI.Events["LiquidationOrderMatched"].ID:
		log.Info("LiquidationOrderMatched event")
		err := cep.orderBookABI.UnpackIntoMap(args, "LiquidationOrderMatched", event.Data)
		if err != nil {
			log.Error("error in orderBookAbi.UnpackIntoMap", "method", "LiquidationOrderMatched", "err", err)
			return
		}
		log.Info("HandleOrderBookEvent", "LiquidationOrderMatched args", args)
		fillAmount := args["fillAmount"].(*big.Int)

		orderId := parseOrderId(args["orderHash"])
		// @todo update liquidable position info
		if !removed {
			cep.database.UpdateFilledBaseAssetQuantity(fillAmount, orderId, event.BlockNumber)
		} else {
			cep.database.UpdateFilledBaseAssetQuantity(fillAmount.Neg(fillAmount), orderId, event.BlockNumber)
		}
	}
}

func (cep *ContractEventsProcessor) handleMarginAccountEvent(event *types.Log) {
	args := map[string]interface{}{}
	switch event.Topics[0] {
	case cep.marginAccountABI.Events["MarginAdded"].ID:
		err := cep.marginAccountABI.UnpackIntoMap(args, "MarginAdded", event.Data)
		if err != nil {
			log.Error("error in marginAccountABI.UnpackIntoMap", "method", "MarginAdded", "err", err)
			return
		}
		collateral := event.Topics[2].Big().Int64()
		cep.database.UpdateMargin(getAddressFromTopicHash(event.Topics[1]), Collateral(collateral), args["amount"].(*big.Int))
	case cep.marginAccountABI.Events["MarginRemoved"].ID:
		err := cep.marginAccountABI.UnpackIntoMap(args, "MarginRemoved", event.Data)
		if err != nil {
			log.Error("error in marginAccountABI.UnpackIntoMap", "method", "MarginRemoved", "err", err)
			return
		}
		collateral := event.Topics[2].Big().Int64()
		cep.database.UpdateMargin(getAddressFromTopicHash(event.Topics[1]), Collateral(collateral), big.NewInt(0).Neg(args["amount"].(*big.Int)))
	case cep.marginAccountABI.Events["PnLRealized"].ID:
		err := cep.marginAccountABI.UnpackIntoMap(args, "PnLRealized", event.Data)
		if err != nil {
			log.Error("error in marginAccountABI.UnpackIntoMap", "method", "PnLRealized", "err", err)
			return
		}
		realisedPnL := args["realizedPnl"].(*big.Int)

		cep.database.UpdateMargin(getAddressFromTopicHash(event.Topics[1]), HUSD, realisedPnL)
	}
	log.Info("Log found", "log_.Address", event.Address.String(), "log_.BlockNumber", event.BlockNumber, "log_.Index", event.Index, "log_.TxHash", event.TxHash.String())
}

func (cep *ContractEventsProcessor) handleClearingHouseEvent(event *types.Log) {
	args := map[string]interface{}{}
	switch event.Topics[0] {
	case cep.clearingHouseABI.Events["FundingRateUpdated"].ID:
		err := cep.clearingHouseABI.UnpackIntoMap(args, "FundingRateUpdated", event.Data)
		if err != nil {
			log.Error("error in clearingHouseABI.UnpackIntoMap", "method", "FundingRateUpdated", "err", err)
			return
		}
		cumulativePremiumFraction := args["cumulativePremiumFraction"].(*big.Int)
		nextFundingTime := args["nextFundingTime"].(*big.Int)
		market := Market(int(event.Topics[1].Big().Int64()))
		log.Info("FundingRateUpdated event", "args", args, "cumulativePremiumFraction", cumulativePremiumFraction, "market", market)
		cep.database.UpdateUnrealisedFunding(Market(market), cumulativePremiumFraction)
		cep.database.UpdateNextFundingTime(nextFundingTime.Uint64())

	case cep.clearingHouseABI.Events["FundingPaid"].ID:
		log.Info("FundingPaid event")
		err := cep.clearingHouseABI.UnpackIntoMap(args, "FundingPaid", event.Data)
		if err != nil {
			log.Error("error in clearingHouseABI.UnpackIntoMap", "method", "FundingPaid", "err", err)
			return
		}
		market := Market(int(event.Topics[2].Big().Int64()))
		cumulativePremiumFraction := args["cumulativePremiumFraction"].(*big.Int)
		cep.database.ResetUnrealisedFunding(Market(market), getAddressFromTopicHash(event.Topics[1]), cumulativePremiumFraction)

	// both PositionModified and PositionLiquidated have the exact same signature
	case cep.clearingHouseABI.Events["PositionModified"].ID:
		log.Info("PositionModified event")
		err := cep.clearingHouseABI.UnpackIntoMap(args, "PositionModified", event.Data)
		if err != nil {
			log.Error("error in clearingHouseABI.UnpackIntoMap", "method", "PositionModified", "err", err)
			return
		}

		market := Market(int(event.Topics[2].Big().Int64()))
		baseAsset := args["baseAsset"].(*big.Int)
		quoteAsset := args["quoteAsset"].(*big.Int)
		lastPrice := big.NewInt(0).Div(big.NewInt(0).Mul(quoteAsset, big.NewInt(1e18)), baseAsset)
		lastPrice.Abs(lastPrice)
		cep.database.UpdateLastPrice(market, lastPrice)

		openNotional := args["openNotional"].(*big.Int)
		size := args["size"].(*big.Int)
		cep.database.UpdatePosition(getAddressFromTopicHash(event.Topics[1]), market, size, openNotional, false)
	case cep.clearingHouseABI.Events["PositionLiquidated"].ID:
		log.Info("PositionLiquidated event")
		err := cep.clearingHouseABI.UnpackIntoMap(args, "PositionLiquidated", event.Data)
		if err != nil {
			log.Error("error in clearingHouseABI.UnpackIntoMap", "method", "PositionLiquidated", "err", err)
			return
		}

		market := Market(int(event.Topics[2].Big().Int64()))
		baseAsset := args["baseAsset"].(*big.Int)
		quoteAsset := args["quoteAsset"].(*big.Int)
		lastPrice := big.NewInt(0).Div(big.NewInt(0).Mul(quoteAsset, big.NewInt(1e18)), baseAsset)
		cep.database.UpdateLastPrice(market, lastPrice)

		openNotional := args["openNotional"].(*big.Int)
		size := args["size"].(*big.Int)
		cep.database.UpdatePosition(getAddressFromTopicHash(event.Topics[1]), market, size, openNotional, true)
	}
}

func getAddressFromTopicHash(topicHash common.Hash) common.Address {
	address32 := topicHash.String() // address in 32 bytes with 0 padding
	return common.HexToAddress(address32[:2] + address32[26:])
}

func getOrderFromRawOrder(rawOrder interface{}) Order {
	order := Order{}
	marshalledOrder, _ := json.Marshal(rawOrder)
	_ = json.Unmarshal(marshalledOrder, &order)
	return order
}

func getOrdersFromRawOrderList(rawOrders interface{}) [2]Order {
	orders := [2]Order{}
	marshalledOrders, _ := json.Marshal(rawOrders)
	_ = json.Unmarshal(marshalledOrders, &orders)
	return orders
}

// @todo change this to return the EIP712 hash instead
func getIdFromOrder(order Order) common.Hash {
	return crypto.Keccak256Hash([]byte(order.Trader.String() + order.Salt.String()))
}
