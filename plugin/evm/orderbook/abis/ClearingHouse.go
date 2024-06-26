package abis

var ClearingHouseAbi = []byte(`{"abi": [
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
        "indexed": false,
        "internalType": "uint256",
        "name": "nextSampleTime",
        "type": "uint256"
      }
    ],
    "name": "NotifyNextPISample",
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
    "name": "PISampleSkipped",
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
        "name": "premiumIndex",
        "type": "int256"
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
    "name": "PISampledUpdated",
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
        "internalType": "enum IClearingHouse.OrderExecutionMode",
        "name": "mode",
        "type": "uint8"
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
    "name": "LIQUIDATION_FAILED",
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
        "internalType": "uint256",
        "name": "",
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
    "inputs": [
      {
        "internalType": "uint256[]",
        "name": "impactBids",
        "type": "uint256[]"
      },
      {
        "internalType": "uint256[]",
        "name": "impactAsks",
        "type": "uint256[]"
      },
      {
        "internalType": "uint256[]",
        "name": "midPrice",
        "type": "uint256[]"
      }
    ],
    "name": "commitLiquiditySample",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "defaultOrderBook",
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
    "name": "getAMMs",
    "outputs": [
      {
        "internalType": "contract IAMM[]",
        "name": "",
        "type": "address[]"
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
    "name": "getNotionalPositionAndMarginVanilla",
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
    "inputs": [],
    "name": "hubbleReferral",
    "outputs": [
      {
        "internalType": "contract IHubbleReferral",
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
        "name": "_governance",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "_feeSink",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "_marginAccount",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "_defaultOrderBook",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "_vusd",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "_hubbleReferral",
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
        "internalType": "address",
        "name": "",
        "type": "address"
      }
    ],
    "name": "isWhitelistedOrderBook",
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
    "inputs": [],
    "name": "juror",
    "outputs": [
      {
        "internalType": "contract IJuror",
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
        "name": "",
        "type": "address"
      }
    ],
    "name": "lastFundingPaid",
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
    "name": "lastFundingTime",
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
            "internalType": "bytes32",
            "name": "orderHash",
            "type": "bytes32"
          },
          {
            "internalType": "enum IClearingHouse.OrderExecutionMode",
            "name": "mode",
            "type": "uint8"
          }
        ],
        "internalType": "struct IClearingHouse.Instruction",
        "name": "instruction",
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
    "outputs": [
      {
        "internalType": "uint256",
        "name": "openInterest",
        "type": "uint256"
      }
    ],
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
        "name": "ammIndex",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "price",
        "type": "uint256"
      },
      {
        "internalType": "int256",
        "name": "toLiquidate",
        "type": "int256"
      }
    ],
    "name": "liquidateSingleAmm",
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
    "name": "nextSampleTime",
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
            "internalType": "bytes32",
            "name": "orderHash",
            "type": "bytes32"
          },
          {
            "internalType": "enum IClearingHouse.OrderExecutionMode",
            "name": "mode",
            "type": "uint8"
          }
        ],
        "internalType": "struct IClearingHouse.Instruction[2]",
        "name": "orders",
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
    "outputs": [
      {
        "internalType": "uint256",
        "name": "openInterest",
        "type": "uint256"
      }
    ],
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
            "internalType": "bytes32",
            "name": "orderHash",
            "type": "bytes32"
          },
          {
            "internalType": "enum IClearingHouse.OrderExecutionMode",
            "name": "mode",
            "type": "uint8"
          }
        ],
        "internalType": "struct IClearingHouse.Instruction",
        "name": "order",
        "type": "tuple"
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
      },
      {
        "internalType": "bool",
        "name": "is2ndTrade",
        "type": "bool"
      }
    ],
    "name": "openPosition",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "openInterest",
        "type": "uint256"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "orderBook",
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
    "inputs": [],
    "name": "referralShare",
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
        "name": "_feeSink",
        "type": "address"
      }
    ],
    "name": "setFeeSink",
    "outputs": [],
    "stateMutability": "nonpayable",
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
        "name": "_juror",
        "type": "address"
      }
    ],
    "name": "setJuror",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "_orderBook",
        "type": "address"
      },
      {
        "internalType": "bool",
        "name": "_status",
        "type": "bool"
      }
    ],
    "name": "setOrderBookWhitelist",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "int256",
        "name": "_maintenanceMargin",
        "type": "int256"
      },
      {
        "internalType": "int256",
        "name": "_minAllowableMargin",
        "type": "int256"
      },
      {
        "internalType": "int256",
        "name": "_takerFee",
        "type": "int256"
      },
      {
        "internalType": "int256",
        "name": "_makerFee",
        "type": "int256"
      },
      {
        "internalType": "uint256",
        "name": "_referralShare",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "_tradingFeeDiscount",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "_liquidationPenalty",
        "type": "uint256"
      }
    ],
    "name": "setParams",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "_referral",
        "type": "address"
      }
    ],
    "name": "setReferral",
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
    "name": "tradingFeeDiscount",
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
        "internalType": "address",
        "name": "trader",
        "type": "address"
      }
    ],
    "name": "updatePositions",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "vusd",
    "outputs": [
      {
        "internalType": "contract VUSD",
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
        "name": "_amm",
        "type": "address"
      }
    ],
    "name": "whitelistAmm",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  }
]}`)
