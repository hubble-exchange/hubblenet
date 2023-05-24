package evm

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/subnet-evm/accounts/abi"

	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/plugin/evm/limitorders"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	genesisJSON      string
	orderBookABI     abi.ABI
	alice, bob       common.Address
	aliceKey, bobKey *ecdsa.PrivateKey
	orderBookABIStr  string = `{"abi": [
		{
		  "inputs": [
			{
			  "internalType": "address",
			  "name": "_clearingHouse",
			  "type": "address"
			},
			{
			  "internalType": "address",
			  "name": "_marginAccount",
			  "type": "address"
			}
		  ],
		  "stateMutability": "nonpayable",
		  "type": "constructor"
		},
		{
		  "anonymous": false,
		  "inputs": [
			{
			  "indexed": false,
			  "internalType": "uint8",
			  "name": "version",
			  "type": "uint8"
			}
		  ],
		  "name": "Initialized",
		  "type": "event"
		},
		{
		  "anonymous": false,
		  "inputs": [
			{
			  "indexed": true,
			  "internalType": "address",
			  "name": "trader",
			  "type": "address"
			},
			{
			  "indexed": true,
			  "internalType": "bytes32",
			  "name": "orderHash",
			  "type": "bytes32"
			},
			{
			  "indexed": false,
			  "internalType": "string",
			  "name": "err",
			  "type": "string"
			},
			{
			  "indexed": false,
			  "internalType": "uint256",
			  "name": "toLiquidate",
			  "type": "uint256"
			}
		  ],
		  "name": "LiquidationError",
		  "type": "event"
		},
		{
		  "anonymous": false,
		  "inputs": [
			{
			  "indexed": true,
			  "internalType": "address",
			  "name": "trader",
			  "type": "address"
			},
			{
			  "indexed": true,
			  "internalType": "bytes32",
			  "name": "orderHash",
			  "type": "bytes32"
			},
			{
			  "indexed": false,
			  "internalType": "uint256",
			  "name": "fillAmount",
			  "type": "uint256"
			},
			{
			  "indexed": false,
			  "internalType": "uint256",
			  "name": "price",
			  "type": "uint256"
			},
			{
			  "indexed": false,
			  "internalType": "uint256",
			  "name": "openInterestNotional",
			  "type": "uint256"
			},
			{
			  "indexed": false,
			  "internalType": "address",
			  "name": "relayer",
			  "type": "address"
			},
			{
			  "indexed": false,
			  "internalType": "uint256",
			  "name": "timestamp",
			  "type": "uint256"
			}
		  ],
		  "name": "LiquidationOrderMatched",
		  "type": "event"
		},
		{
		  "anonymous": false,
		  "inputs": [
			{
			  "indexed": true,
			  "internalType": "address",
			  "name": "trader",
			  "type": "address"
			},
			{
			  "indexed": true,
			  "internalType": "bytes32",
			  "name": "orderHash",
			  "type": "bytes32"
			},
			{
			  "indexed": false,
			  "internalType": "uint256",
			  "name": "timestamp",
			  "type": "uint256"
			}
		  ],
		  "name": "OrderCancelled",
		  "type": "event"
		},
		{
		  "anonymous": false,
		  "inputs": [
			{
			  "indexed": true,
			  "internalType": "bytes32",
			  "name": "orderHash",
			  "type": "bytes32"
			},
			{
			  "indexed": false,
			  "internalType": "string",
			  "name": "err",
			  "type": "string"
			}
		  ],
		  "name": "OrderMatchingError",
		  "type": "event"
		},
		{
		  "anonymous": false,
		  "inputs": [
			{
			  "indexed": true,
			  "internalType": "address",
			  "name": "trader",
			  "type": "address"
			},
			{
			  "indexed": true,
			  "internalType": "bytes32",
			  "name": "orderHash",
			  "type": "bytes32"
			},
			{
			  "components": [
				{
				  "internalType": "uint256",
				  "name": "ammIndex",
				  "type": "uint256"
				},
				{
				  "internalType": "address",
				  "name": "trader",
				  "type": "address"
				},
				{
				  "internalType": "int256",
				  "name": "baseAssetQuantity",
				  "type": "int256"
				},
				{
				  "internalType": "uint256",
				  "name": "price",
				  "type": "uint256"
				},
				{
				  "internalType": "uint256",
				  "name": "salt",
				  "type": "uint256"
				},
				{
				  "internalType": "bool",
				  "name": "reduceOnly",
				  "type": "bool"
				}
			  ],
			  "indexed": false,
			  "internalType": "struct IOrderBook.Order",
			  "name": "order",
			  "type": "tuple"
			},
			{
			  "indexed": false,
			  "internalType": "uint256",
			  "name": "timestamp",
			  "type": "uint256"
			}
		  ],
		  "name": "OrderPlaced",
		  "type": "event"
		},
		{
		  "anonymous": false,
		  "inputs": [
			{
			  "indexed": true,
			  "internalType": "bytes32",
			  "name": "orderHash0",
			  "type": "bytes32"
			},
			{
			  "indexed": true,
			  "internalType": "bytes32",
			  "name": "orderHash1",
			  "type": "bytes32"
			},
			{
			  "indexed": false,
			  "internalType": "uint256",
			  "name": "fillAmount",
			  "type": "uint256"
			},
			{
			  "indexed": false,
			  "internalType": "uint256",
			  "name": "price",
			  "type": "uint256"
			},
			{
			  "indexed": false,
			  "internalType": "uint256",
			  "name": "openInterestNotional",
			  "type": "uint256"
			},
			{
			  "indexed": false,
			  "internalType": "address",
			  "name": "relayer",
			  "type": "address"
			},
			{
			  "indexed": false,
			  "internalType": "uint256",
			  "name": "timestamp",
			  "type": "uint256"
			}
		  ],
		  "name": "OrdersMatched",
		  "type": "event"
		},
		{
		  "anonymous": false,
		  "inputs": [
			{
			  "indexed": false,
			  "internalType": "address",
			  "name": "account",
			  "type": "address"
			}
		  ],
		  "name": "Paused",
		  "type": "event"
		},
		{
		  "anonymous": false,
		  "inputs": [
			{
			  "indexed": false,
			  "internalType": "address",
			  "name": "account",
			  "type": "address"
			}
		  ],
		  "name": "Unpaused",
		  "type": "event"
		},
		{
		  "inputs": [],
		  "name": "ORDER_TYPEHASH",
		  "outputs": [
			{
			  "internalType": "bytes32",
			  "name": "",
			  "type": "bytes32"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "bytes32[]",
			  "name": "orderHashes",
			  "type": "bytes32[]"
			}
		  ],
		  "name": "cancelMultipleOrders",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "bytes32",
			  "name": "orderHash",
			  "type": "bytes32"
			}
		  ],
		  "name": "cancelOrder",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [],
		  "name": "clearingHouse",
		  "outputs": [
			{
			  "internalType": "contract IClearingHouse",
			  "name": "",
			  "type": "address"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "bytes32",
			  "name": "orderHash0",
			  "type": "bytes32"
			},
			{
			  "internalType": "bytes32",
			  "name": "orderHash1",
			  "type": "bytes32"
			},
			{
			  "internalType": "int256",
			  "name": "fillAmount",
			  "type": "int256"
			}
		  ],
		  "name": "executeMatchedOrders",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "components": [
				{
				  "internalType": "uint256",
				  "name": "ammIndex",
				  "type": "uint256"
				},
				{
				  "internalType": "address",
				  "name": "trader",
				  "type": "address"
				},
				{
				  "internalType": "int256",
				  "name": "baseAssetQuantity",
				  "type": "int256"
				},
				{
				  "internalType": "uint256",
				  "name": "price",
				  "type": "uint256"
				},
				{
				  "internalType": "uint256",
				  "name": "salt",
				  "type": "uint256"
				},
				{
				  "internalType": "bool",
				  "name": "reduceOnly",
				  "type": "bool"
				}
			  ],
			  "internalType": "struct IOrderBook.Order[2]",
			  "name": "orders",
			  "type": "tuple[2]"
			},
			{
			  "internalType": "int256",
			  "name": "fillAmount",
			  "type": "int256"
			}
		  ],
		  "name": "executeMatchedOrders2",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [],
		  "name": "getLastTradePrices",
		  "outputs": [
			{
			  "internalType": "uint256[]",
			  "name": "lastTradePrices",
			  "type": "uint256[]"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "components": [
				{
				  "internalType": "uint256",
				  "name": "ammIndex",
				  "type": "uint256"
				},
				{
				  "internalType": "address",
				  "name": "trader",
				  "type": "address"
				},
				{
				  "internalType": "int256",
				  "name": "baseAssetQuantity",
				  "type": "int256"
				},
				{
				  "internalType": "uint256",
				  "name": "price",
				  "type": "uint256"
				},
				{
				  "internalType": "uint256",
				  "name": "salt",
				  "type": "uint256"
				},
				{
				  "internalType": "bool",
				  "name": "reduceOnly",
				  "type": "bool"
				}
			  ],
			  "internalType": "struct IOrderBook.Order",
			  "name": "order",
			  "type": "tuple"
			}
		  ],
		  "name": "getOrderHash",
		  "outputs": [
			{
			  "internalType": "bytes32",
			  "name": "",
			  "type": "bytes32"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [],
		  "name": "governance",
		  "outputs": [
			{
			  "internalType": "address",
			  "name": "",
			  "type": "address"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "string",
			  "name": "_name",
			  "type": "string"
			},
			{
			  "internalType": "string",
			  "name": "_version",
			  "type": "string"
			},
			{
			  "internalType": "address",
			  "name": "_governance",
			  "type": "address"
			}
		  ],
		  "name": "initialize",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "int256",
			  "name": "minSize",
			  "type": "int256"
			}
		  ],
		  "name": "initializeMinSize",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "address",
			  "name": "",
			  "type": "address"
			}
		  ],
		  "name": "isValidator",
		  "outputs": [
			{
			  "internalType": "bool",
			  "name": "",
			  "type": "bool"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "address",
			  "name": "trader",
			  "type": "address"
			},
			{
			  "internalType": "bytes32",
			  "name": "orderHash",
			  "type": "bytes32"
			},
			{
			  "internalType": "uint256",
			  "name": "liquidationAmount",
			  "type": "uint256"
			}
		  ],
		  "name": "liquidateAndExecuteOrder",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [],
		  "name": "marginAccount",
		  "outputs": [
			{
			  "internalType": "contract IMarginAccount",
			  "name": "",
			  "type": "address"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [],
		  "name": "minAllowableMargin",
		  "outputs": [
			{
			  "internalType": "uint256",
			  "name": "",
			  "type": "uint256"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "uint256",
			  "name": "",
			  "type": "uint256"
			}
		  ],
		  "name": "minSizes",
		  "outputs": [
			{
			  "internalType": "int256",
			  "name": "",
			  "type": "int256"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "bytes32",
			  "name": "",
			  "type": "bytes32"
			}
		  ],
		  "name": "orderInfo",
		  "outputs": [
			{
			  "components": [
				{
				  "internalType": "uint256",
				  "name": "ammIndex",
				  "type": "uint256"
				},
				{
				  "internalType": "address",
				  "name": "trader",
				  "type": "address"
				},
				{
				  "internalType": "int256",
				  "name": "baseAssetQuantity",
				  "type": "int256"
				},
				{
				  "internalType": "uint256",
				  "name": "price",
				  "type": "uint256"
				},
				{
				  "internalType": "uint256",
				  "name": "salt",
				  "type": "uint256"
				},
				{
				  "internalType": "bool",
				  "name": "reduceOnly",
				  "type": "bool"
				}
			  ],
			  "internalType": "struct IOrderBook.Order",
			  "name": "order",
			  "type": "tuple"
			},
			{
			  "internalType": "uint256",
			  "name": "blockPlaced",
			  "type": "uint256"
			},
			{
			  "internalType": "int256",
			  "name": "filledAmount",
			  "type": "int256"
			},
			{
			  "internalType": "uint256",
			  "name": "reservedMargin",
			  "type": "uint256"
			},
			{
			  "internalType": "enum IOrderBook.OrderStatus",
			  "name": "status",
			  "type": "uint8"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "string",
			  "name": "err",
			  "type": "string"
			}
		  ],
		  "name": "parseMatchingError",
		  "outputs": [
			{
			  "internalType": "bytes32",
			  "name": "orderHash",
			  "type": "bytes32"
			},
			{
			  "internalType": "string",
			  "name": "reason",
			  "type": "string"
			}
		  ],
		  "stateMutability": "pure",
		  "type": "function"
		},
		{
		  "inputs": [],
		  "name": "pause",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [],
		  "name": "paused",
		  "outputs": [
			{
			  "internalType": "bool",
			  "name": "",
			  "type": "bool"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "components": [
				{
				  "internalType": "uint256",
				  "name": "ammIndex",
				  "type": "uint256"
				},
				{
				  "internalType": "address",
				  "name": "trader",
				  "type": "address"
				},
				{
				  "internalType": "int256",
				  "name": "baseAssetQuantity",
				  "type": "int256"
				},
				{
				  "internalType": "uint256",
				  "name": "price",
				  "type": "uint256"
				},
				{
				  "internalType": "uint256",
				  "name": "salt",
				  "type": "uint256"
				},
				{
				  "internalType": "bool",
				  "name": "reduceOnly",
				  "type": "bool"
				}
			  ],
			  "internalType": "struct IOrderBook.Order[]",
			  "name": "orders",
			  "type": "tuple[]"
			}
		  ],
		  "name": "placeOrder",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "components": [
				{
				  "internalType": "uint256",
				  "name": "ammIndex",
				  "type": "uint256"
				},
				{
				  "internalType": "address",
				  "name": "trader",
				  "type": "address"
				},
				{
				  "internalType": "int256",
				  "name": "baseAssetQuantity",
				  "type": "int256"
				},
				{
				  "internalType": "uint256",
				  "name": "price",
				  "type": "uint256"
				},
				{
				  "internalType": "uint256",
				  "name": "salt",
				  "type": "uint256"
				},
				{
				  "internalType": "bool",
				  "name": "reduceOnly",
				  "type": "bool"
				}
			  ],
			  "internalType": "struct IOrderBook.Order[]",
			  "name": "orders",
			  "type": "tuple[]"
			}
		  ],
		  "name": "placeOrdersInSingleMarket",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "address",
			  "name": "",
			  "type": "address"
			},
			{
			  "internalType": "uint256",
			  "name": "",
			  "type": "uint256"
			}
		  ],
		  "name": "reduceOnlyAmount",
		  "outputs": [
			{
			  "internalType": "int256",
			  "name": "",
			  "type": "int256"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "address",
			  "name": "__governance",
			  "type": "address"
			}
		  ],
		  "name": "setGovernace",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "address",
			  "name": "validator",
			  "type": "address"
			},
			{
			  "internalType": "bool",
			  "name": "status",
			  "type": "bool"
			}
		  ],
		  "name": "setValidatorStatus",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [],
		  "name": "settleFunding",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [],
		  "name": "takerFee",
		  "outputs": [
			{
			  "internalType": "uint256",
			  "name": "",
			  "type": "uint256"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		},
		{
		  "inputs": [],
		  "name": "unpause",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "uint256",
			  "name": "ammIndex",
			  "type": "uint256"
			},
			{
			  "internalType": "int256",
			  "name": "minSize",
			  "type": "int256"
			}
		  ],
		  "name": "updateMinSize",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "internalType": "uint256",
			  "name": "_minAllowableMargin",
			  "type": "uint256"
			},
			{
			  "internalType": "uint256",
			  "name": "_takerFee",
			  "type": "uint256"
			}
		  ],
		  "name": "updateParams",
		  "outputs": [],
		  "stateMutability": "nonpayable",
		  "type": "function"
		},
		{
		  "inputs": [
			{
			  "components": [
				{
				  "internalType": "uint256",
				  "name": "ammIndex",
				  "type": "uint256"
				},
				{
				  "internalType": "address",
				  "name": "trader",
				  "type": "address"
				},
				{
				  "internalType": "int256",
				  "name": "baseAssetQuantity",
				  "type": "int256"
				},
				{
				  "internalType": "uint256",
				  "name": "price",
				  "type": "uint256"
				},
				{
				  "internalType": "uint256",
				  "name": "salt",
				  "type": "uint256"
				},
				{
				  "internalType": "bool",
				  "name": "reduceOnly",
				  "type": "bool"
				}
			  ],
			  "internalType": "struct IOrderBook.Order",
			  "name": "order",
			  "type": "tuple"
			},
			{
			  "internalType": "bytes",
			  "name": "signature",
			  "type": "bytes"
			}
		  ],
		  "name": "verifySigner",
		  "outputs": [
			{
			  "internalType": "address",
			  "name": "",
			  "type": "address"
			},
			{
			  "internalType": "bytes32",
			  "name": "",
			  "type": "bytes32"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		}
	  ]}`
	orderBookAddress common.Address = common.HexToAddress("0x0300000000000000000000000000000000000069")
	_1e18            *big.Int       = big.NewInt(1e18)
	_1e6             *big.Int       = big.NewInt(1e6)
)

func init() {
	var err error

	genesisJSON = `{"config":{"chainId":321123,"homesteadBlock":0,"eip150Block":0,"eip150Hash":"0x2086799aeebeae135c246c65021c82b4e15a2c451340993aacfd2751886514f0","eip155Block":0,"eip158Block":0,"byzantiumBlock":0,"constantinopleBlock":0,"petersburgBlock":0,"istanbulBlock":0,"muirGlacierBlock":0,"SubnetEVMTimestamp":0,"feeConfig":{"gasLimit":500000000,"targetBlockRate":1,"minBaseFee":60000000000,"targetGas":10000000,"baseFeeChangeDenominator":50,"minBlockGasCost":0,"maxBlockGasCost":0,"blockGasCostStep":10000}},"alloc":{"835cE0760387BC894E91039a88A00b6a69E65D94":{"balance":"0xD3C21BCECCEDA1000000"},"8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC":{"balance":"0xD3C21BCECCEDA1000000"},"55ee05dF718f1a5C1441e76190EB1a19eE2C9430":{"balance":"0xD3C21BCECCEDA1000000"},"4Cf2eD3665F6bFA95cE6A11CFDb7A2EF5FC1C7E4":{"balance":"0xD3C21BCECCEDA1000000"},"f39Fd6e51aad88F6F4ce6aB8827279cffFb92266":{"balance":"0xD3C21BCECCEDA1000000"},"70997970C51812dc3A010C7d01b50e0d17dc79C8":{"balance":"0xD3C21BCECCEDA1000000"},"3C44CdDdB6a900fa2b585dd299e03d12FA4293BC":{"balance":"0xD3C21BCECCEDA1000000"},"0x0300000000000000000000000000000000000069":{"balance":"0x0","code":"0x608060405234801561001057600080fd5b50600436106101005760003560e01c8063a48e6e5c11610097578063e684d71811610066578063e684d718146102ac578063e942ff80146102dd578063ed83d79c146102f9578063f973a2091461030357610100565b8063a48e6e5c14610228578063b76f0adf14610244578063bbee505314610260578063dbe648461461027c57610100565b80633245dea5116100d35780633245dea5146101a257806342c1f8a4146101d25780634cd88b76146101ee5780637114f7f81461020a57610100565b806322dae63714610105578063238e203f1461013657806327d57a9e146101685780632c82ce1714610186575b600080fd5b61011f600480360381019061011a9190611992565b610321565b60405161012d929190611a16565b60405180910390f35b610150600480360381019061014b9190611a6b565b610343565b60405161015f93929190611b2d565b60405180910390f35b61017061037a565b60405161017d9190611b64565b60405180910390f35b6101a0600480360381019061019b9190611b7f565b610380565b005b6101bc60048036038101906101b79190611bac565b61051f565b6040516101c99190611b64565b60405180910390f35b6101ec60048036038101906101e79190611bac565b610537565b005b61020860048036038101906102039190611c7a565b610541565b005b61021261068d565b60405161021f9190611db0565b60405180910390f35b610242600480360381019061023d9190611dd2565b610733565b005b61025e60048036038101906102599190611fdc565b6108f2565b005b61027a60048036038101906102759190611992565b610c17565b005b61029660048036038101906102919190611b7f565b610d0b565b6040516102a3919061204e565b60405180910390f35b6102c660048036038101906102c19190612069565b610d68565b6040516102d49291906120a9565b60405180910390f35b6102f760048036038101906102f29190611992565b610d99565b005b610301610dd8565b005b61030b610dda565b604051610318919061204e565b60405180910390f35b600080600061032f85610d0b565b905084602001518192509250509250929050565b60356020528060005260406000206000915090508060000154908060010154908060020160009054906101000a900460ff16905083565b60385481565b806020015173ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146103f2576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103e99061212f565b60405180910390fd5b60006103fd82610d0b565b90506001600381111561041357610412611ab6565b5b6035600083815260200190815260200160002060020160009054906101000a900460ff16600381111561044957610448611ab6565b5b14610489576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104809061219b565b60405180910390fd5b60036035600083815260200190815260200160002060020160006101000a81548160ff021916908360038111156104c3576104c2611ab6565b5b021790555080826020015173ffffffffffffffffffffffffffffffffffffffff167f26b214029d2b6a3a3bb2ae7cc0a5d4c9329a86381429e16dc45b3633cf83d369426040516105139190611b64565b60405180910390a35050565b60376020528060005260406000206000915090505481565b8060388190555050565b60008060019054906101000a900460ff161590508080156105725750600160008054906101000a900460ff1660ff16105b8061059f575061058130610e01565b15801561059e5750600160008054906101000a900460ff1660ff16145b5b6105de576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105d59061222d565b60405180910390fd5b60016000806101000a81548160ff021916908360ff160217905550801561061b576001600060016101000a81548160ff0219169083151502179055505b6106258383610e24565b61062f6001610537565b80156106885760008060016101000a81548160ff0219169083151502179055507f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498600160405161067f919061229f565b60405180910390a15b505050565b606060385467ffffffffffffffff8111156106ab576106aa6116bb565b5b6040519080825280602002602001820160405280156106d95781602001602082028036833780820191505090505b50905060005b60385481101561072f5760376000828152602001908152602001600020548282815181106107105761070f6122ba565b5b602002602001018181525050808061072790612318565b9150506106df565b5090565b670de0b6b3a764000081846060015161074c9190612361565b61075691906123ea565b603660008560000151815260200190815260200160002060008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060010160008282546107bc919061241b565b925050819055506107cc81610e81565b603660008560000151815260200190815260200160002060008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000016000828254610832919061244f565b92505081905550600061084e848461084985610e81565b610eee565b5090506108688161085e84610e81565b86604001516110d3565b61087f8461087584610e81565b8660600151611163565b808573ffffffffffffffffffffffffffffffffffffffff167fd7a2e338b47db7ba2c25b55a69d8eb13126b1ec669de521cd1985aae9ee32ca185858860600151878a606001516108cf9190612361565b33426040516108e39695949392919061256b565b60405180910390a35050505050565b600083600060028110610908576109076122ba565b5b60200201516040015113610951576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016109489061261f565b60405180910390fd5b600083600160028110610967576109666122ba565b5b602002015160400151126109b0576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016109a79061268b565b60405180910390fd5b600081136109f3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016109ea906126f7565b60405180910390fd5b82600160028110610a0757610a066122ba565b5b60200201516060015183600060028110610a2457610a236122ba565b5b6020020151606001511015610a6e576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610a6590612763565b60405180910390fd5b6000610aaa84600060028110610a8757610a866122ba565b5b602002015184600060028110610aa057610a9f6122ba565b5b6020020151610321565b9150506000610ae985600160028110610ac657610ac56122ba565b5b602002015185600160028110610adf57610ade6122ba565b5b6020020151610321565b915050610b13828487600060028110610b0557610b046122ba565b5b6020020151604001516110d3565b610b438184610b2190612783565b87600160028110610b3557610b346122ba565b5b6020020151604001516110d3565b600085600060028110610b5957610b586122ba565b5b6020020151606001519050610b8786600060028110610b7b57610b7a6122ba565b5b60200201518583611163565b610bb386600160028110610b9e57610b9d6122ba565b5b602002015185610bad90612783565b83611163565b81837faf4b403d9952e032974b549a4abad80faca307b0acc6e34d7e0b8c274d504590610bdf876114cc565b8485610bea8a6114cc565b610bf49190612361565b3342604051610c079594939291906127cc565b60405180910390a3505050505050565b6000610c238383610321565b91505060405180606001604052804381526020016000815260200160016003811115610c5257610c51611ab6565b5b81525060356000838152602001908152602001600020600082015181600001556020820151816001015560408201518160020160006101000a81548160ff02191690836003811115610ca757610ca6611ab6565b5b021790555090505080836020015173ffffffffffffffffffffffffffffffffffffffff167f7f274ad4fd1954f444e6bf0a812141d464edc592c19662dd5afa30d4f078d355858542604051610cfe939291906128c7565b60405180910390a3505050565b6000610d617f0a2e4d36552888a97d5a8975ad22b04e90efe5ea0a8abb97691b63b431eb25d260001b83604051602001610d46929190612906565b60405160208183030381529060405280519060200120611519565b9050919050565b6036602052816000526040600020602052806000526040600020600091509150508060000154908060010154905082565b6000610daa83838560400151610eee565b509050610dc081846040015185604001516110d3565b610dd38384604001518560600151611163565b505050565b565b7f0a2e4d36552888a97d5a8975ad22b04e90efe5ea0a8abb97691b63b431eb25d260001b81565b6000808273ffffffffffffffffffffffffffffffffffffffff163b119050919050565b600060019054906101000a900460ff16610e73576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610e6a906129a1565b60405180910390fd5b610e7d8282611533565b5050565b60007f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff821115610ee6576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610edd90612a33565b60405180910390fd5b819050919050565b6000806000610efd8686610321565b91505060016003811115610f1457610f13611ab6565b5b6035600083815260200190815260200160002060020160009054906101000a900460ff166003811115610f4a57610f49611ab6565b5b14610f8a576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610f8190612a9f565b60405180910390fd5b6000848760400151610f9c9190612abf565b13610fdc576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610fd390612c22565b60405180910390fd5b60008460356000848152602001908152602001600020600101546110009190612abf565b1215611041576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161103890612c8e565b60405180910390fd5b61104e86604001516115ae565b61106d60356000848152602001908152602001600020600101546115ae565b13156110ae576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016110a590612d20565b60405180910390fd5b8060356000838152602001908152602001600020600001549250925050935093915050565b816035600085815260200190815260200160002060010160008282546110f99190612d40565b92505081905550806035600085815260200190815260200160002060010154141561115e5760026035600085815260200190815260200160002060020160006101000a81548160ff0219169083600381111561115857611157611ab6565b5b02179055505b505050565b6000670de0b6b3a76400008261118061117b866115ae565b6114cc565b61118a9190612361565b61119491906123ea565b905060008460200151905060008560000151905060385481106111ec576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016111e390612e20565b60405180910390fd5b6000856036600084815260200190815260200160002060008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000015461124d9190612abf565b126112c157826036600083815260200190815260200160002060008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060010160008282546112b59190612e40565b92505081905550611442565b826036600083815260200190815260200160002060008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600101541061138a57826036600083815260200190815260200160002060008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600101600082825461137e919061241b565b92505081905550611441565b6036600082815260200190815260200160002060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060010154836113e9919061241b565b6036600083815260200190815260200160002060008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600101819055505b5b846036600083815260200190815260200160002060008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000160008282546114a59190612d40565b92505081905550836037600083815260200190815260200160002081905550505050505050565b600080821215611511576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161150890612ee2565b60405180910390fd5b819050919050565b600061152c6115266115d0565b83611610565b9050919050565b600060019054906101000a900460ff16611582576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401611579906129a1565b60405180910390fd5b600082805190602001209050600082805190602001209050816001819055508060028190555050505050565b6000808212156115c757816115c290612783565b6115c9565b815b9050919050565b600061160b7f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f6115fe611643565b61160661164d565b611657565b905090565b60008282604051602001611625929190612f7a565b60405160208183030381529060405280519060200120905092915050565b6000600154905090565b6000600254905090565b60008383834630604051602001611672959493929190612fb1565b6040516020818303038152906040528051906020012090509392505050565b6000604051905090565b600080fd5b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6116f3826116aa565b810181811067ffffffffffffffff82111715611712576117116116bb565b5b80604052505050565b6000611725611691565b905061173182826116ea565b919050565b6000819050919050565b61174981611736565b811461175457600080fd5b50565b60008135905061176681611740565b92915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006117978261176c565b9050919050565b6117a78161178c565b81146117b257600080fd5b50565b6000813590506117c48161179e565b92915050565b6000819050919050565b6117dd816117ca565b81146117e857600080fd5b50565b6000813590506117fa816117d4565b92915050565b60008115159050919050565b61181581611800565b811461182057600080fd5b50565b6000813590506118328161180c565b92915050565b600060c0828403121561184e5761184d6116a5565b5b61185860c061171b565b9050600061186884828501611757565b600083015250602061187c848285016117b5565b6020830152506040611890848285016117eb565b60408301525060606118a484828501611757565b60608301525060806118b884828501611757565b60808301525060a06118cc84828501611823565b60a08301525092915050565b600080fd5b600080fd5b600067ffffffffffffffff8211156118fd576118fc6116bb565b5b611906826116aa565b9050602081019050919050565b82818337600083830152505050565b6000611935611930846118e2565b61171b565b905082815260208101848484011115611951576119506118dd565b5b61195c848285611913565b509392505050565b600082601f830112611979576119786118d8565b5b8135611989848260208601611922565b91505092915050565b60008060e083850312156119a9576119a861169b565b5b60006119b785828601611838565b92505060c083013567ffffffffffffffff8111156119d8576119d76116a0565b5b6119e485828601611964565b9150509250929050565b6119f78161178c565b82525050565b6000819050919050565b611a10816119fd565b82525050565b6000604082019050611a2b60008301856119ee565b611a386020830184611a07565b9392505050565b611a48816119fd565b8114611a5357600080fd5b50565b600081359050611a6581611a3f565b92915050565b600060208284031215611a8157611a8061169b565b5b6000611a8f84828501611a56565b91505092915050565b611aa181611736565b82525050565b611ab0816117ca565b82525050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b60048110611af657611af5611ab6565b5b50565b6000819050611b0782611ae5565b919050565b6000611b1782611af9565b9050919050565b611b2781611b0c565b82525050565b6000606082019050611b426000830186611a98565b611b4f6020830185611aa7565b611b5c6040830184611b1e565b949350505050565b6000602082019050611b796000830184611a98565b92915050565b600060c08284031215611b9557611b9461169b565b5b6000611ba384828501611838565b91505092915050565b600060208284031215611bc257611bc161169b565b5b6000611bd084828501611757565b91505092915050565b600067ffffffffffffffff821115611bf457611bf36116bb565b5b611bfd826116aa565b9050602081019050919050565b6000611c1d611c1884611bd9565b61171b565b905082815260208101848484011115611c3957611c386118dd565b5b611c44848285611913565b509392505050565b600082601f830112611c6157611c606118d8565b5b8135611c71848260208601611c0a565b91505092915050565b60008060408385031215611c9157611c9061169b565b5b600083013567ffffffffffffffff811115611caf57611cae6116a0565b5b611cbb85828601611c4c565b925050602083013567ffffffffffffffff811115611cdc57611cdb6116a0565b5b611ce885828601611c4c565b9150509250929050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b611d2781611736565b82525050565b6000611d398383611d1e565b60208301905092915050565b6000602082019050919050565b6000611d5d82611cf2565b611d678185611cfd565b9350611d7283611d0e565b8060005b83811015611da3578151611d8a8882611d2d565b9750611d9583611d45565b925050600181019050611d76565b5085935050505092915050565b60006020820190508181036000830152611dca8184611d52565b905092915050565b6000806000806101208587031215611ded57611dec61169b565b5b6000611dfb878288016117b5565b9450506020611e0c87828801611838565b93505060e085013567ffffffffffffffff811115611e2d57611e2c6116a0565b5b611e3987828801611964565b925050610100611e4b87828801611757565b91505092959194509250565b600067ffffffffffffffff821115611e7257611e716116bb565b5b602082029050919050565b600080fd5b6000611e95611e9084611e57565b61171b565b90508060c08402830185811115611eaf57611eae611e7d565b5b835b81811015611ed85780611ec48882611838565b84526020840193505060c081019050611eb1565b5050509392505050565b600082601f830112611ef757611ef66118d8565b5b6002611f04848285611e82565b91505092915050565b600067ffffffffffffffff821115611f2857611f276116bb565b5b602082029050919050565b6000611f46611f4184611f0d565b61171b565b90508060208402830185811115611f6057611f5f611e7d565b5b835b81811015611fa757803567ffffffffffffffff811115611f8557611f846118d8565b5b808601611f928982611964565b85526020850194505050602081019050611f62565b5050509392505050565b600082601f830112611fc657611fc56118d8565b5b6002611fd3848285611f33565b91505092915050565b60008060006101c08486031215611ff657611ff561169b565b5b600061200486828701611ee2565b93505061018084013567ffffffffffffffff811115612026576120256116a0565b5b61203286828701611fb1565b9250506101a0612044868287016117eb565b9150509250925092565b60006020820190506120636000830184611a07565b92915050565b600080604083850312156120805761207f61169b565b5b600061208e85828601611757565b925050602061209f858286016117b5565b9150509250929050565b60006040820190506120be6000830185611aa7565b6120cb6020830184611a98565b9392505050565b600082825260208201905092915050565b7f4f425f73656e6465725f69735f6e6f745f747261646572000000000000000000600082015250565b60006121196017836120d2565b9150612124826120e3565b602082019050919050565b600060208201905081810360008301526121488161210c565b9050919050565b7f4f425f4f726465725f646f65735f6e6f745f6578697374000000000000000000600082015250565b60006121856017836120d2565b91506121908261214f565b602082019050919050565b600060208201905081810360008301526121b481612178565b9050919050565b7f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160008201527f647920696e697469616c697a6564000000000000000000000000000000000000602082015250565b6000612217602e836120d2565b9150612222826121bb565b604082019050919050565b600060208201905081810360008301526122468161220a565b9050919050565b6000819050919050565b600060ff82169050919050565b6000819050919050565b600061228961228461227f8461224d565b612264565b612257565b9050919050565b6122998161226e565b82525050565b60006020820190506122b46000830184612290565b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061232382611736565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff821415612356576123556122e9565b5b600182019050919050565b600061236c82611736565b915061237783611736565b9250817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff04831182151516156123b0576123af6122e9565b5b828202905092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b60006123f582611736565b915061240083611736565b9250826124105761240f6123bb565b5b828204905092915050565b600061242682611736565b915061243183611736565b925082821015612444576124436122e9565b5b828203905092915050565b600061245a826117ca565b9150612465836117ca565b9250827f8000000000000000000000000000000000000000000000000000000000000000018212600084121516156124a05761249f6122e9565b5b827f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0182136000841216156124d8576124d76122e9565b5b828203905092915050565b600081519050919050565b600082825260208201905092915050565b60005b8381101561251d578082015181840152602081019050612502565b8381111561252c576000848401525b50505050565b600061253d826124e3565b61254781856124ee565b93506125578185602086016124ff565b612560816116aa565b840191505092915050565b600060c08201905081810360008301526125858189612532565b90506125946020830188611a98565b6125a16040830187611a98565b6125ae6060830186611a98565b6125bb60808301856119ee565b6125c860a0830184611a98565b979650505050505050565b7f4f425f6f726465725f305f69735f6e6f745f6c6f6e6700000000000000000000600082015250565b60006126096016836120d2565b9150612614826125d3565b602082019050919050565b60006020820190508181036000830152612638816125fc565b9050919050565b7f4f425f6f726465725f315f69735f6e6f745f73686f7274000000000000000000600082015250565b60006126756017836120d2565b91506126808261263f565b602082019050919050565b600060208201905081810360008301526126a481612668565b9050919050565b7f4f425f66696c6c416d6f756e745f69735f6e6567000000000000000000000000600082015250565b60006126e16014836120d2565b91506126ec826126ab565b602082019050919050565b60006020820190508181036000830152612710816126d4565b9050919050565b7f4f425f6f72646572735f646f5f6e6f745f6d6174636800000000000000000000600082015250565b600061274d6016836120d2565b915061275882612717565b602082019050919050565b6000602082019050818103600083015261277c81612740565b9050919050565b600061278e826117ca565b91507f80000000000000000000000000000000000000000000000000000000000000008214156127c1576127c06122e9565b5b816000039050919050565b600060a0820190506127e16000830188611a98565b6127ee6020830187611a98565b6127fb6040830186611a98565b61280860608301856119ee565b6128156080830184611a98565b9695505050505050565b6128288161178c565b82525050565b612837816117ca565b82525050565b61284681611800565b82525050565b60c0820160008201516128626000850182611d1e565b506020820151612875602085018261281f565b506040820151612888604085018261282e565b50606082015161289b6060850182611d1e565b5060808201516128ae6080850182611d1e565b5060a08201516128c160a085018261283d565b50505050565b6000610100820190506128dd600083018661284c565b81810360c08301526128ef8185612532565b90506128fe60e0830184611a98565b949350505050565b600060e08201905061291b6000830185611a07565b612928602083018461284c565b9392505050565b7f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960008201527f6e697469616c697a696e67000000000000000000000000000000000000000000602082015250565b600061298b602b836120d2565b91506129968261292f565b604082019050919050565b600060208201905081810360008301526129ba8161297e565b9050919050565b7f53616665436173743a2076616c756520646f65736e27742066697420696e206160008201527f6e20696e74323536000000000000000000000000000000000000000000000000602082015250565b6000612a1d6028836120d2565b9150612a28826129c1565b604082019050919050565b60006020820190508181036000830152612a4c81612a10565b9050919050565b7f4f425f696e76616c69645f6f7264657200000000000000000000000000000000600082015250565b6000612a896010836120d2565b9150612a9482612a53565b602082019050919050565b60006020820190508181036000830152612ab881612a7c565b9050919050565b6000612aca826117ca565b9150612ad5836117ca565b9250827f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0482116000841360008413161615612b1457612b136122e9565b5b817f80000000000000000000000000000000000000000000000000000000000000000583126000841260008413161615612b5157612b506122e9565b5b827f80000000000000000000000000000000000000000000000000000000000000000582126000841360008412161615612b8e57612b8d6122e9565b5b827f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0582126000841260008412161615612bcb57612bca6122e9565b5b828202905092915050565b7f4f425f66696c6c5f616e645f626173655f7369676e5f6e6f745f6d6174636800600082015250565b6000612c0c601f836120d2565b9150612c1782612bd6565b602082019050919050565b60006020820190508181036000830152612c3b81612bff565b9050919050565b7f4f425f696e76616c69645f66696c6c416d6f756e740000000000000000000000600082015250565b6000612c786015836120d2565b9150612c8382612c42565b602082019050919050565b60006020820190508181036000830152612ca781612c6b565b9050919050565b7f4f425f66696c6c65645f616d6f756e745f6869676865725f7468616e5f6f726460008201527f65725f6261736500000000000000000000000000000000000000000000000000602082015250565b6000612d0a6027836120d2565b9150612d1582612cae565b604082019050919050565b60006020820190508181036000830152612d3981612cfd565b9050919050565b6000612d4b826117ca565b9150612d56836117ca565b9250817f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff03831360008312151615612d9157612d906122e9565b5b817f8000000000000000000000000000000000000000000000000000000000000000038312600083121615612dc957612dc86122e9565b5b828201905092915050565b7f4f425f706c656173655f77686974656c6973745f6e65775f616d6d0000000000600082015250565b6000612e0a601b836120d2565b9150612e1582612dd4565b602082019050919050565b60006020820190508181036000830152612e3981612dfd565b9050919050565b6000612e4b82611736565b9150612e5683611736565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff03821115612e8b57612e8a6122e9565b5b828201905092915050565b7f53616665436173743a2076616c7565206d75737420626520706f736974697665600082015250565b6000612ecc6020836120d2565b9150612ed782612e96565b602082019050919050565b60006020820190508181036000830152612efb81612ebf565b9050919050565b600081905092915050565b7f1901000000000000000000000000000000000000000000000000000000000000600082015250565b6000612f43600283612f02565b9150612f4e82612f0d565b600282019050919050565b6000819050919050565b612f74612f6f826119fd565b612f59565b82525050565b6000612f8582612f36565b9150612f918285612f63565b602082019150612fa18284612f63565b6020820191508190509392505050565b600060a082019050612fc66000830188611a07565b612fd36020830187611a07565b612fe06040830186611a07565b612fed6060830185611a98565b612ffa60808301846119ee565b969550505050505056fea264697066735822122096d90e98a54c642bd9dd81173852a06f31463aafa98e6e057569c6538ceafaf764736f6c63430008090033"},"0x0300000000000000000000000000000000000071":{"balance":"0x0","code":"0x608060405234801561001057600080fd5b506004361061002b5760003560e01c8063468f02d214610030575b600080fd5b61003861004e565b604051610045919061018b565b60405180910390f35b6060600167ffffffffffffffff81111561006b5761006a6101ad565b5b6040519080825280602002602001820160405280156100995781602001602082028036833780820191505090505b50905062989680816000815181106100b4576100b36101dc565b5b60200260200101818152505090565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b6000819050919050565b610102816100ef565b82525050565b600061011483836100f9565b60208301905092915050565b6000602082019050919050565b6000610138826100c3565b61014281856100ce565b935061014d836100df565b8060005b8381101561017e5781516101658882610108565b975061017083610120565b925050600181019050610151565b5085935050505092915050565b600060208201905081810360008301526101a5818461012d565b905092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fdfea264697066735822122039e97a6c3ee1092095bdb746678d98a91e90153cb1bec7ee19de94b59bf9e17e64736f6c63430008090033"}},"nonce":"0x0","timestamp":"0x0","extraData":"0x00","gasLimit":"500000000","difficulty":"0x0","mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000","coinbase":"0x0000000000000000000000000000000000000000","number":"0x0","gasUsed":"0x0","parentHash":"0x0000000000000000000000000000000000000000000000000000000000000000"}`

	orderBookABI, err = abi.FromSolidityJson(orderBookABIStr)
	if err != nil {
		panic(err)
	}

	aliceKey, _ = crypto.HexToECDSA("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
	bobKey, _ = crypto.HexToECDSA("31b571bf6894a248831ff937bb49f7754509fe93bbd2517c9c73c4144c0e97dc")
	alice = crypto.PubkeyToAddress(aliceKey.PublicKey)
	bob = crypto.PubkeyToAddress(bobKey.PublicKey)
}

func createPlaceOrderTx(t *testing.T, vm *VM, trader common.Address, privateKey *ecdsa.PrivateKey, size *big.Int, price *big.Int, salt *big.Int) common.Hash {
	nonce := vm.txPool.Nonce(trader)

	order := limitorders.Order{
		Trader:            trader,
		AmmIndex:          big.NewInt(0),
		BaseAssetQuantity: big.NewInt(0).Mul(size, _1e18),
		Price:             big.NewInt(0).Mul(price, _1e6),
		Salt:              salt,
		ReduceOnly:        false,
	}
	data, err := orderBookABI.Pack("placeOrder", order)
	if err != nil {
		t.Fatalf("orderBookABI.Pack failed: %v", err)
	}
	tx := types.NewTransaction(nonce, orderBookAddress, big.NewInt(0), 8000000, big.NewInt(500000000000), data)
	signer := types.NewLondonSigner(vm.chainConfig.ChainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		t.Fatalf("types.SignTx failed: %v", err)
	}
	errs := vm.txPool.AddRemotesSync([]*types.Transaction{signedTx})
	for _, err := range errs {
		if err != nil {
			t.Fatalf("lop.txPool.AddOrderBookTx failed: %v", err)
		}
	}
	return signedTx.Hash()
}

//	  A
//	 / \
//	B   C
//	    |
//	    D (matching tx of order 1 and 2)

// vm1 proposes block A containing order 1
// block A is accepted by vm1 and vm2
// vm1 proposes block B containing order 2
// vm1 and vm2 set preference to block B
// vm2 proposes block C containing order 2 & order 3
// vm1 and vm2 set preference to block C
// reorg happens when vm1 accepts block C
// vm2 proposes block D containing matching tx of order 1 and 2
// vm1 and vm2 set preference to block D
// vm1 accepts block D
// block D is important because an earlier bug caused vm1 to crash because order 2 didn't exist in vm1 memory DB after reorg
func TestHubbleLogs(t *testing.T) {
	// Create two VMs which will agree on block A and then
	// build the two distinct preferred chains above
	ctx := context.Background()
	issuer1, vm1, _, _ := GenesisVM(t, true, genesisJSON, "{\"pruning-enabled\":true}", "")
	issuer2, vm2, _, _ := GenesisVM(t, true, genesisJSON, "{\"pruning-enabled\":true}", "")

	defer func() {
		if err := vm1.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}

		if err := vm2.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	// long and short order
	createPlaceOrderTx(t, vm1, alice, aliceKey, big.NewInt(5), big.NewInt(10), big.NewInt(101))
	<-issuer1
	// include alice's long order
	blocksA := buildBlockAndSetPreference(t, vm1, vm2) // block A - both vms accept
	accept(t, blocksA...)

	createPlaceOrderTx(t, vm1, bob, bobKey, big.NewInt(-5), big.NewInt(10), big.NewInt(102))
	<-issuer1
	// bob's short order
	buildBlockAndSetPreference(t, vm1) // block B - vm1 only

	// build block C parallel to block B
	createPlaceOrderTx(t, vm2, bob, bobKey, big.NewInt(-5), big.NewInt(10), big.NewInt(102))
	createPlaceOrderTx(t, vm2, alice, aliceKey, big.NewInt(5), big.NewInt(11), big.NewInt(104))
	<-issuer2
	vm2BlockC := buildBlockAndSetPreference(t, vm2)[0] // block C - vm2 only for now

	vm1BlockC := parseBlock(t, vm1, vm2BlockC)
	setPreference(t, vm1BlockC, vm1)
	accept(t, vm1BlockC) // reorg happens here
	accept(t, vm2BlockC)

	// time.Sleep(2 * time.Second)
	detail1 := vm1.limitOrderProcesser.GetOrderBookAPI().GetDetailedOrderBookData(context.Background())
	detail2 := vm2.limitOrderProcesser.GetOrderBookAPI().GetDetailedOrderBookData(context.Background())
	t.Logf("VM1 Orders: %+v", detail1)
	t.Logf("VM2 Orders: %+v", detail2)

	if _, ok := detail1.OrderMap[common.HexToHash("0xdc30f1521636413ca875cde2bf0b4f0a756b7235af7638737b2279d6613b9540")]; !ok {
		t.Fatalf("Order 2 is not in VM1")
	}
	if _, ok := detail2.OrderMap[common.HexToHash("0xdc30f1521636413ca875cde2bf0b4f0a756b7235af7638737b2279d6613b9540")]; !ok {
		t.Fatalf("Order 2 is not in VM2")
	}

	// order matching tx
	vm2BlockD := buildBlockAndSetPreference(t, vm2)[0]
	vm1BlockD := parseBlock(t, vm1, vm2BlockD)
	setPreference(t, vm1BlockD, vm1)
	accept(t, vm1BlockD)
	accept(t, vm2BlockD)

	vm1LastAccepted, err := vm1.LastAccepted(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if vm1LastAccepted != vm1BlockD.ID() {
		t.Fatalf("VM1 last accepted block is not block D")
	}

	vm2LastAccepted, err := vm2.LastAccepted(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if vm2LastAccepted != vm2BlockD.ID() {
		t.Fatalf("VM2 last accepted block is not block D")
	}

	// Verify the Canonical Chain for Both VMs
	if err := vm2.blockChain.ValidateCanonicalChain(); err != nil {
		t.Fatalf("VM2 failed canonical chain verification due to: %s", err)
	}

	if err := vm1.blockChain.ValidateCanonicalChain(); err != nil {
		t.Fatalf("VM1 failed canonical chain verification due to: %s", err)
	}
}

func buildBlockAndSetPreference(t *testing.T, vms ...*VM) []snowman.Block {
	if len(vms) == 0 {
		t.Fatal("No VMs provided")
	}
	response := []snowman.Block{}
	vm1 := vms[0]
	vm1Blk, err := vm1.BuildBlock(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if err := vm1Blk.Verify(context.Background()); err != nil {
		t.Fatal(err)
	}

	if status := vm1Blk.Status(); status != choices.Processing {
		t.Fatalf("Expected status of built block to be %s, but found %s", choices.Processing, status)
	}

	if err := vm1.SetPreference(context.Background(), vm1Blk.ID()); err != nil {
		t.Fatal(err)
	}

	response = append(response, vm1Blk)

	for _, vm := range vms[1:] {

		vm2Blk, err := vm.ParseBlock(context.Background(), vm1Blk.Bytes())
		if err != nil {
			t.Fatalf("Unexpected error parsing block from vm2: %s", err)
		}
		if err := vm2Blk.Verify(context.Background()); err != nil {
			t.Fatalf("Block failed verification on VM2: %s", err)
		}
		if status := vm2Blk.Status(); status != choices.Processing {
			t.Fatalf("Expected status of block on VM2 to be %s, but found %s", choices.Processing, status)
		}
		if err := vm.SetPreference(context.Background(), vm2Blk.ID()); err != nil {
			t.Fatal(err)
		}
		response = append(response, vm2Blk)
	}

	return response
}

func buildBlock(t *testing.T, vm *VM) snowman.Block {
	vmBlk, err := vm.BuildBlock(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if err := vmBlk.Verify(context.Background()); err != nil {
		t.Fatal(err)
	}

	if status := vmBlk.Status(); status != choices.Processing {
		t.Fatalf("Expected status of built block to be %s, but found %s", choices.Processing, status)
	}

	return vmBlk
}

func parseBlock(t *testing.T, vm *VM, block snowman.Block) snowman.Block {
	newBlock, err := vm.ParseBlock(context.Background(), block.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error parsing block from vm: %s", err)
	}
	if err := newBlock.Verify(context.Background()); err != nil {
		t.Fatal(err)
	}

	if status := newBlock.Status(); status != choices.Processing {
		t.Fatalf("Expected status of built block to be %s, but found %s", choices.Processing, status)
	}

	return newBlock
}

func setPreference(t *testing.T, block snowman.Block, vms ...*VM) {
	for _, vm := range vms {
		if err := vm.SetPreference(context.Background(), block.ID()); err != nil {
			t.Fatal(err)
		}
	}
}

func accept(t *testing.T, blocks ...snowman.Block) {
	for _, block := range blocks {
		if err := block.Accept(context.Background()); err != nil {
			t.Fatalf("VM failed to accept block: %s", err)
		}
	}
}
