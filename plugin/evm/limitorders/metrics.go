package limitorders

import (
	"github.com/ava-labs/subnet-evm/metrics"
)

var (
	transactionsPerBlockHistogram = metrics.NewRegisteredHistogram("transactions/rate", nil, metrics.NewExpDecaySample(1028, 0.015))

	gasUsedPerBlockHistogram      = metrics.NewRegisteredHistogram("gas_used_per_block", nil, metrics.NewExpDecaySample(1028, 0.015))
	blockGasCostPerBlockHistogram = metrics.NewRegisteredHistogram("block_gas_cost", nil, metrics.NewExpDecaySample(1028, 0.015))

	// only valid for OrderBook transactions send by this validator
	orderBookTransactionsSuccessTotalCounter = metrics.NewRegisteredCounter("orderbooktxs/total/success", nil)
	orderBookTransactionsFailureTotalCounter = metrics.NewRegisteredCounter("orderbooktxs/total/failure", nil)
)
