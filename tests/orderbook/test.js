
const { ethers } = require('ethers');
const { BigNumber } = require('ethers')
const axios = require('axios');
const { expect } = require('chai');
const { randomInt } = require('crypto');

const OrderBookContractAddress = "0x0300000000000000000000000000000000000069"
const MarginAccountContractAddress = "0x0300000000000000000000000000000000000070"
const ClearingHouseContractAddress = "0x0300000000000000000000000000000000000071"

let domain, orderType, orderBook, alice
const orderBookAbi = [{ "inputs": [{ "internalType": "address", "name": "_clearingHouse", "type": "address" }], "stateMutability": "nonpayable", "type": "constructor" }, { "anonymous": false, "inputs": [{ "indexed": false, "internalType": "uint8", "name": "version", "type": "uint8" }], "name": "Initialized", "type": "event" }, { "anonymous": false, "inputs": [{ "indexed": true, "internalType": "address", "name": "trader", "type": "address" }, { "indexed": true, "internalType": "bytes32", "name": "orderHash", "type": "bytes32" }, { "indexed": false, "internalType": "string", "name": "err", "type": "string" }, { "indexed": false, "internalType": "uint256", "name": "toLiquidate", "type": "uint256" }], "name": "LiquidationError", "type": "event" }, { "anonymous": false, "inputs": [{ "indexed": true, "internalType": "address", "name": "trader", "type": "address" }, { "indexed": true, "internalType": "bytes32", "name": "orderHash", "type": "bytes32" }, { "indexed": false, "internalType": "bytes", "name": "signature", "type": "bytes" }, { "indexed": false, "internalType": "uint256", "name": "fillAmount", "type": "uint256" }, { "indexed": false, "internalType": "uint256", "name": "openInterestNotional", "type": "uint256" }, { "indexed": false, "internalType": "address", "name": "relayer", "type": "address" }], "name": "LiquidationOrderMatched", "type": "event" }, { "anonymous": false, "inputs": [{ "indexed": true, "internalType": "address", "name": "trader", "type": "address" }, { "indexed": true, "internalType": "bytes32", "name": "orderHash", "type": "bytes32" }], "name": "OrderCancelled", "type": "event" }, { "anonymous": false, "inputs": [{ "indexed": true, "internalType": "bytes32", "name": "orderHash", "type": "bytes32" }, { "indexed": false, "internalType": "string", "name": "err", "type": "string" }], "name": "OrderMatchingError", "type": "event" }, { "anonymous": false, "inputs": [{ "indexed": true, "internalType": "address", "name": "trader", "type": "address" }, { "indexed": true, "internalType": "bytes32", "name": "orderHash", "type": "bytes32" }, { "components": [{ "internalType": "uint256", "name": "ammIndex", "type": "uint256" }, { "internalType": "address", "name": "trader", "type": "address" }, { "internalType": "int256", "name": "baseAssetQuantity", "type": "int256" }, { "internalType": "uint256", "name": "price", "type": "uint256" }, { "internalType": "uint256", "name": "salt", "type": "uint256" }], "indexed": false, "internalType": "struct IOrderBook.Order", "name": "order", "type": "tuple" }, { "indexed": false, "internalType": "bytes", "name": "signature", "type": "bytes" }], "name": "OrderPlaced", "type": "event" }, { "anonymous": false, "inputs": [{ "indexed": true, "internalType": "bytes32", "name": "orderHash0", "type": "bytes32" }, { "indexed": true, "internalType": "bytes32", "name": "orderHash1", "type": "bytes32" }, { "indexed": false, "internalType": "uint256", "name": "fillAmount", "type": "uint256" }, { "indexed": false, "internalType": "uint256", "name": "price", "type": "uint256" }, { "indexed": false, "internalType": "uint256", "name": "openInterestNotional", "type": "uint256" }, { "indexed": false, "internalType": "address", "name": "relayer", "type": "address" }], "name": "OrdersMatched", "type": "event" }, { "anonymous": false, "inputs": [{ "indexed": false, "internalType": "address", "name": "account", "type": "address" }], "name": "Paused", "type": "event" }, { "anonymous": false, "inputs": [{ "indexed": false, "internalType": "address", "name": "account", "type": "address" }], "name": "Unpaused", "type": "event" }, { "inputs": [], "name": "ORDER_TYPEHASH", "outputs": [{ "internalType": "bytes32", "name": "", "type": "bytes32" }], "stateMutability": "view", "type": "function" }, { "inputs": [{ "components": [{ "internalType": "uint256", "name": "ammIndex", "type": "uint256" }, { "internalType": "address", "name": "trader", "type": "address" }, { "internalType": "int256", "name": "baseAssetQuantity", "type": "int256" }, { "internalType": "uint256", "name": "price", "type": "uint256" }, { "internalType": "uint256", "name": "salt", "type": "uint256" }], "internalType": "struct IOrderBook.Order[]", "name": "orders", "type": "tuple[]" }], "name": "cancelMultipleOrders", "outputs": [], "stateMutability": "nonpayable", "type": "function" }, { "inputs": [{ "components": [{ "internalType": "uint256", "name": "ammIndex", "type": "uint256" }, { "internalType": "address", "name": "trader", "type": "address" }, { "internalType": "int256", "name": "baseAssetQuantity", "type": "int256" }, { "internalType": "uint256", "name": "price", "type": "uint256" }, { "internalType": "uint256", "name": "salt", "type": "uint256" }], "internalType": "struct IOrderBook.Order", "name": "order", "type": "tuple" }], "name": "cancelOrder", "outputs": [], "stateMutability": "nonpayable", "type": "function" }, { "inputs": [], "name": "clearingHouse", "outputs": [{ "internalType": "contract IClearingHouse", "name": "", "type": "address" }], "stateMutability": "view", "type": "function" }, { "inputs": [{ "components": [{ "internalType": "uint256", "name": "ammIndex", "type": "uint256" }, { "internalType": "address", "name": "trader", "type": "address" }, { "internalType": "int256", "name": "baseAssetQuantity", "type": "int256" }, { "internalType": "uint256", "name": "price", "type": "uint256" }, { "internalType": "uint256", "name": "salt", "type": "uint256" }], "internalType": "struct IOrderBook.Order[2]", "name": "orders", "type": "tuple[2]" }, { "internalType": "bytes[2]", "name": "signatures", "type": "bytes[2]" }, { "internalType": "int256", "name": "fillAmount", "type": "int256" }], "name": "executeMatchedOrders", "outputs": [], "stateMutability": "nonpayable", "type": "function" }, { "inputs": [], "name": "getLastTradePrices", "outputs": [{ "internalType": "uint256[]", "name": "lastTradePrices", "type": "uint256[]" }], "stateMutability": "view", "type": "function" }, { "inputs": [{ "components": [{ "internalType": "uint256", "name": "ammIndex", "type": "uint256" }, { "internalType": "address", "name": "trader", "type": "address" }, { "internalType": "int256", "name": "baseAssetQuantity", "type": "int256" }, { "internalType": "uint256", "name": "price", "type": "uint256" }, { "internalType": "uint256", "name": "salt", "type": "uint256" }], "internalType": "struct IOrderBook.Order", "name": "order", "type": "tuple" }], "name": "getOrderHash", "outputs": [{ "internalType": "bytes32", "name": "", "type": "bytes32" }], "stateMutability": "view", "type": "function" }, { "inputs": [], "name": "governance", "outputs": [{ "internalType": "address", "name": "", "type": "address" }], "stateMutability": "view", "type": "function" }, { "inputs": [{ "internalType": "string", "name": "_name", "type": "string" }, { "internalType": "string", "name": "_version", "type": "string" }, { "internalType": "address", "name": "_governance", "type": "address" }], "name": "initialize", "outputs": [], "stateMutability": "nonpayable", "type": "function" }, { "inputs": [{ "internalType": "address", "name": "trader", "type": "address" }, { "components": [{ "internalType": "uint256", "name": "ammIndex", "type": "uint256" }, { "internalType": "address", "name": "trader", "type": "address" }, { "internalType": "int256", "name": "baseAssetQuantity", "type": "int256" }, { "internalType": "uint256", "name": "price", "type": "uint256" }, { "internalType": "uint256", "name": "salt", "type": "uint256" }], "internalType": "struct IOrderBook.Order", "name": "order", "type": "tuple" }, { "internalType": "bytes", "name": "signature", "type": "bytes" }, { "internalType": "uint256", "name": "liquidationAmount", "type": "uint256" }], "name": "liquidateAndExecuteOrder", "outputs": [], "stateMutability": "nonpayable", "type": "function" }, { "inputs": [{ "internalType": "bytes32", "name": "", "type": "bytes32" }], "name": "orderInfo", "outputs": [{ "internalType": "uint256", "name": "blockPlaced", "type": "uint256" }, { "internalType": "int256", "name": "filledAmount", "type": "int256" }, { "internalType": "enum IOrderBook.OrderStatus", "name": "status", "type": "uint8" }], "stateMutability": "view", "type": "function" }, { "inputs": [{ "internalType": "string", "name": "err", "type": "string" }], "name": "parseMatchingError", "outputs": [{ "internalType": "bytes32", "name": "orderHash", "type": "bytes32" }, { "internalType": "string", "name": "reason", "type": "string" }], "stateMutability": "pure", "type": "function" }, { "inputs": [], "name": "pause", "outputs": [], "stateMutability": "nonpayable", "type": "function" }, { "inputs": [], "name": "paused", "outputs": [{ "internalType": "bool", "name": "", "type": "bool" }], "stateMutability": "view", "type": "function" }, { "inputs": [{ "components": [{ "internalType": "uint256", "name": "ammIndex", "type": "uint256" }, { "internalType": "address", "name": "trader", "type": "address" }, { "internalType": "int256", "name": "baseAssetQuantity", "type": "int256" }, { "internalType": "uint256", "name": "price", "type": "uint256" }, { "internalType": "uint256", "name": "salt", "type": "uint256" }], "internalType": "struct IOrderBook.Order", "name": "order", "type": "tuple" }, { "internalType": "bytes", "name": "signature", "type": "bytes" }], "name": "placeOrder", "outputs": [], "stateMutability": "nonpayable", "type": "function" }, { "inputs": [{ "internalType": "address", "name": "_governance", "type": "address" }], "name": "setGovernace", "outputs": [], "stateMutability": "nonpayable", "type": "function" }, { "inputs": [{ "internalType": "address", "name": "validator", "type": "address" }, { "internalType": "bool", "name": "status", "type": "bool" }], "name": "setValidatorStatus", "outputs": [], "stateMutability": "nonpayable", "type": "function" }, { "inputs": [], "name": "settleFunding", "outputs": [], "stateMutability": "nonpayable", "type": "function" }, { "inputs": [], "name": "unpause", "outputs": [], "stateMutability": "nonpayable", "type": "function" }, { "inputs": [{ "components": [{ "internalType": "uint256", "name": "ammIndex", "type": "uint256" }, { "internalType": "address", "name": "trader", "type": "address" }, { "internalType": "int256", "name": "baseAssetQuantity", "type": "int256" }, { "internalType": "uint256", "name": "price", "type": "uint256" }, { "internalType": "uint256", "name": "salt", "type": "uint256" }], "internalType": "struct IOrderBook.Order", "name": "order", "type": "tuple" }, { "internalType": "bytes", "name": "signature", "type": "bytes" }], "name": "verifySigner", "outputs": [{ "internalType": "address", "name": "", "type": "address" }, { "internalType": "bytes32", "name": "", "type": "bytes32" }], "stateMutability": "view", "type": "function" }]
const url = "http://127.0.0.1:9650/ext/bc/21B8FKbEjSpmoC3VP1j1Cy22s1Ja6JikbbU2kCcw3Wv6EitnaM/rpc"

describe('Submit transaction and compare with EVM state', function () {
    before('', async function () {
        const provider = new ethers.providers.JsonRpcProvider(url);

        // Set up signer
        alice = new ethers.Wallet('56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027', provider);

        // Set up contract interface
        orderBook = new ethers.Contract(OrderBookContractAddress, orderBookAbi);
        domain = {
            name: 'Hubble',
            version: '2.0',
            chainId: (await provider.getNetwork()).chainId,
            verifyingContract: orderBook.address
        }

        orderType = {
            Order: [
                // field ordering must be the same as LIMIT_ORDER_TYPEHASH
                { name: "ammIndex", type: "uint256" },
                { name: "trader", type: "address" },
                { name: "baseAssetQuantity", type: "int256" },
                { name: "price", type: "uint256" },
                { name: "salt", type: "uint256" },
            ]
        }

    })

    it('Place order', async function () {

        const {tx, hash, order} = await placeOrder(alice, 5, 10)
        console.log({ tx });

        // Wait for transaction to be mined
        await tx.wait();

        const evmState = await getEVMState()
        const evmOrder = evmState.order_map[hash]
        console.log(evmOrder);
        console.log({order})
        // console.log(bnToFloat(order.baseAssetQuantity, 0))
        // console.log(bnToFloat(order.price, 0))
        // console.log(bnToFloat(order.salt, 0))

        expect(evmOrder).to.not.null
        expect(evmOrder.user_address).to.eq(order.trader)
        expect(evmOrder.salt).to.eq(bnToFloat(order.salt, 0))
        expect(evmOrder.price).to.eq(bnToFloat(order.price, 0))
        expect(evmOrder.base_asset_quantity).to.eq(bnToFloat(order.baseAssetQuantity, 0))

    });
});

async function placeOrder(trader, size, price) {
    const order = {
        ammIndex: BigNumber.from(0),
        trader: trader.address,
        baseAssetQuantity: ethers.utils.parseEther(size.toString()),
        price: ethers.utils.parseUnits(price.toString(), 6),
        salt: BigNumber.from(Date.now() + randomInt(100))
    }
    const signature = await trader._signTypedData(domain, orderType, order)
    const hash = await orderBook.connect(trader).getOrderHash(order)
    console.log({signature, hash});

    const tx = await orderBook.connect(trader).placeOrder(order, signature)
    return {tx, hash, order}
}

async function getEVMState() {
    const response = await axios.post(url, {
        jsonrpc: '2.0',
        id: 1,
        method: 'orderbook_getDetailedOrderBookData',
        params: []
    }, {
        headers: {
            'Content-Type': 'application/json'
        }
    });

    return response.data.result
}

function bnToFloat(num, decimals = 6) {
    return parseFloat(ethers.utils.formatUnits(num.toString(), decimals))
}
