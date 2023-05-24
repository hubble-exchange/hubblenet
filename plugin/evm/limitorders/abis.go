package limitorders

var orderBookAbi = []byte(`{"abi": [
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
        "internalType": "struct IOrderBook.Order",
        "name": "order",
        "type": "tuple"
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
    "name": "placeOrders",
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
]}`)

var marginAccountAbi = []byte(`{"abi": [
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
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "seizeAmount",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "repayAmount",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "timestamp",
        "type": "uint256"
      }
    ],
    "name": "MarginAccountLiquidated",
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
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "timestamp",
        "type": "uint256"
      }
    ],
    "name": "MarginAdded",
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
        "indexed": false,
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      }
    ],
    "name": "MarginReleased",
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
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "timestamp",
        "type": "uint256"
      }
    ],
    "name": "MarginRemoved",
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
        "indexed": false,
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      }
    ],
    "name": "MarginReserved",
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
        "indexed": false,
        "internalType": "int256",
        "name": "realizedPnl",
        "type": "int256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "timestamp",
        "type": "uint256"
      }
    ],
    "name": "PnLRealized",
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
        "indexed": false,
        "internalType": "uint256[]",
        "name": "seized",
        "type": "uint256[]"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "repayAmount",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "timestamp",
        "type": "uint256"
      }
    ],
    "name": "SettledBadDebt",
    "type": "event"
  },
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      }
    ],
    "name": "addMargin",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      },
      {
        "internalType": "address",
        "name": "to",
        "type": "address"
      }
    ],
    "name": "addMarginFor",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "trader",
        "type": "address"
      }
    ],
    "name": "getAvailableMargin",
    "outputs": [
      {
        "internalType": "int256",
        "name": "availableMargin",
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
        "name": "trader",
        "type": "address"
      }
    ],
    "name": "getNormalizedMargin",
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
        "name": "trader",
        "type": "address"
      }
    ],
    "name": "getSpotCollateralValue",
    "outputs": [
      {
        "internalType": "int256",
        "name": "spot",
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
        "name": "trader",
        "type": "address"
      },
      {
        "internalType": "bool",
        "name": "includeFunding",
        "type": "bool"
      }
    ],
    "name": "isLiquidatable",
    "outputs": [
      {
        "internalType": "enum IMarginAccount.LiquidationStatus",
        "name": "",
        "type": "uint8"
      },
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      },
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
        "internalType": "address",
        "name": "trader",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "repay",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "minSeizeAmount",
        "type": "uint256"
      }
    ],
    "name": "liquidateExactRepay",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      },
      {
        "internalType": "address",
        "name": "trader",
        "type": "address"
      }
    ],
    "name": "margin",
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
    "inputs": [],
    "name": "oracle",
    "outputs": [
      {
        "internalType": "contract IOracle",
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
        "internalType": "address",
        "name": "trader",
        "type": "address"
      },
      {
        "internalType": "int256",
        "name": "realizedPnl",
        "type": "int256"
      }
    ],
    "name": "realizePnL",
    "outputs": [],
    "stateMutability": "nonpayable",
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
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      }
    ],
    "name": "releaseMargin",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      }
    ],
    "name": "removeMargin",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      },
      {
        "internalType": "address",
        "name": "trader",
        "type": "address"
      }
    ],
    "name": "removeMarginFor",
    "outputs": [],
    "stateMutability": "nonpayable",
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
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      }
    ],
    "name": "reserveMargin",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "trader",
        "type": "address"
      }
    ],
    "name": "reservedMargin",
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
    "name": "supportedAssets",
    "outputs": [
      {
        "components": [
          {
            "internalType": "contract IERC20",
            "name": "token",
            "type": "address"
          },
          {
            "internalType": "uint256",
            "name": "weight",
            "type": "uint256"
          },
          {
            "internalType": "uint8",
            "name": "decimals",
            "type": "uint8"
          }
        ],
        "internalType": "struct IMarginAccount.Collateral[]",
        "name": "",
        "type": "tuple[]"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "supportedAssetsLen",
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
        "internalType": "address",
        "name": "recipient",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      }
    ],
    "name": "transferOutVusd",
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
        "internalType": "address",
        "name": "trader",
        "type": "address"
      }
    ],
    "name": "weightedAndSpotCollateral",
    "outputs": [
      {
        "internalType": "int256",
        "name": "",
        "type": "int256"
      },
      {
        "internalType": "int256",
        "name": "",
        "type": "int256"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }
]}`)

var clearingHouseAbi = []byte(`{"abi": [
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
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "int256",
        "name": "takerFundingPayment",
        "type": "int256"
      },
      {
        "indexed": false,
        "internalType": "int256",
        "name": "cumulativePremiumFraction",
        "type": "int256"
      }
    ],
    "name": "FundingPaid",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "int256",
        "name": "premiumFraction",
        "type": "int256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "underlyingPrice",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "int256",
        "name": "cumulativePremiumFraction",
        "type": "int256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "nextFundingTime",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "timestamp",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "blockNumber",
        "type": "uint256"
      }
    ],
    "name": "FundingRateUpdated",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "amm",
        "type": "address"
      }
    ],
    "name": "MarketAdded",
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
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "int256",
        "name": "baseAsset",
        "type": "int256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "price",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "int256",
        "name": "realizedPnl",
        "type": "int256"
      },
      {
        "indexed": false,
        "internalType": "int256",
        "name": "size",
        "type": "int256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "openNotional",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "int256",
        "name": "fee",
        "type": "int256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "timestamp",
        "type": "uint256"
      }
    ],
    "name": "PositionLiquidated",
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
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "int256",
        "name": "baseAsset",
        "type": "int256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "price",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "int256",
        "name": "realizedPnl",
        "type": "int256"
      },
      {
        "indexed": false,
        "internalType": "int256",
        "name": "size",
        "type": "int256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "openNotional",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "int256",
        "name": "fee",
        "type": "int256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "timestamp",
        "type": "uint256"
      }
    ],
    "name": "PositionModified",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "referrer",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "referralBonus",
        "type": "uint256"
      }
    ],
    "name": "ReferralBonusAdded",
    "type": "event"
  },
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "idx",
        "type": "uint256"
      }
    ],
    "name": "amms",
    "outputs": [
      {
        "internalType": "contract IAMM",
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
        "internalType": "address",
        "name": "trader",
        "type": "address"
      }
    ],
    "name": "assertMarginRequirement",
    "outputs": [],
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
        "internalType": "bool",
        "name": "includeFundingPayments",
        "type": "bool"
      },
      {
        "internalType": "enum IClearingHouse.Mode",
        "name": "mode",
        "type": "uint8"
      }
    ],
    "name": "calcMarginFraction",
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
    "inputs": [],
    "name": "feeSink",
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
    "inputs": [],
    "name": "getAmmsLength",
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
        "internalType": "address",
        "name": "trader",
        "type": "address"
      },
      {
        "internalType": "bool",
        "name": "includeFundingPayments",
        "type": "bool"
      },
      {
        "internalType": "enum IClearingHouse.Mode",
        "name": "mode",
        "type": "uint8"
      }
    ],
    "name": "getNotionalPositionAndMargin",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "notionalPosition",
        "type": "uint256"
      },
      {
        "internalType": "int256",
        "name": "margin",
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
        "name": "trader",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "ammIndex",
        "type": "uint256"
      }
    ],
    "name": "getPositionSize",
    "outputs": [
      {
        "internalType": "int256",
        "name": "posSizes",
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
        "name": "trader",
        "type": "address"
      }
    ],
    "name": "getPositionSizes",
    "outputs": [
      {
        "internalType": "int256[]",
        "name": "posSizes",
        "type": "int256[]"
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
      }
    ],
    "name": "getTotalFunding",
    "outputs": [
      {
        "internalType": "int256",
        "name": "totalFunding",
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
        "name": "trader",
        "type": "address"
      },
      {
        "internalType": "int256",
        "name": "margin",
        "type": "int256"
      },
      {
        "internalType": "enum IClearingHouse.Mode",
        "name": "mode",
        "type": "uint8"
      }
    ],
    "name": "getTotalNotionalPositionAndUnrealizedPnl",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "notionalPosition",
        "type": "uint256"
      },
      {
        "internalType": "int256",
        "name": "unrealizedPnl",
        "type": "int256"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "getUnderlyingPrice",
    "outputs": [
      {
        "internalType": "uint256[]",
        "name": "prices",
        "type": "uint256[]"
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
      }
    ],
    "name": "isAboveMaintenanceMargin",
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
        "internalType": "struct IOrderBook.Order",
        "name": "order",
        "type": "tuple"
      },
      {
        "components": [
          {
            "internalType": "bytes32",
            "name": "orderHash",
            "type": "bytes32"
          },
          {
            "internalType": "uint256",
            "name": "blockPlaced",
            "type": "uint256"
          },
          {
            "internalType": "enum IOrderBook.OrderExecutionMode",
            "name": "mode",
            "type": "uint8"
          }
        ],
        "internalType": "struct IOrderBook.MatchInfo",
        "name": "matchInfo",
        "type": "tuple"
      },
      {
        "internalType": "int256",
        "name": "liquidationAmount",
        "type": "int256"
      },
      {
        "internalType": "uint256",
        "name": "price",
        "type": "uint256"
      },
      {
        "internalType": "address",
        "name": "trader",
        "type": "address"
      }
    ],
    "name": "liquidate",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "liquidationPenalty",
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
    "name": "maintenanceMargin",
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
    "inputs": [],
    "name": "makerFee",
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
    "inputs": [],
    "name": "minAllowableMargin",
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
        "components": [
          {
            "internalType": "bytes32",
            "name": "orderHash",
            "type": "bytes32"
          },
          {
            "internalType": "uint256",
            "name": "blockPlaced",
            "type": "uint256"
          },
          {
            "internalType": "enum IOrderBook.OrderExecutionMode",
            "name": "mode",
            "type": "uint8"
          }
        ],
        "internalType": "struct IOrderBook.MatchInfo[2]",
        "name": "matchInfo",
        "type": "tuple[2]"
      },
      {
        "internalType": "int256",
        "name": "fillAmount",
        "type": "int256"
      },
      {
        "internalType": "uint256",
        "name": "fulfillPrice",
        "type": "uint256"
      }
    ],
    "name": "openComplementaryPositions",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "orderBook",
    "outputs": [
      {
        "internalType": "contract IOrderBook",
        "name": "",
        "type": "address"
      }
    ],
    "stateMutability": "view",
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
        "name": "trader",
        "type": "address"
      }
    ],
    "name": "updatePositions",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  }
]}`)
