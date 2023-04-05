
const { ethers } = require('ethers');
const { BigNumber } = require('ethers')
const axios = require('axios');
const { expect } = require('chai');
const { randomInt } = require('crypto');

const OrderBookContractAddress = "0x0300000000000000000000000000000000000069"
const MarginAccountContractAddress = "0x0300000000000000000000000000000000000070"
const MarginAccountHelperContractAddress = "0x610178dA211FEF7D417bC0e6FeD39F05609AD788"
const ClearingHouseContractAddress = "0x0300000000000000000000000000000000000071"

let domain, orderType, orderBook, marginAccount, marginAccountHelper
let alice, bob, charlie, aliceAddress, bobAddress, charlieAddress

const ZERO = BigNumber.from(0)
const _1e6 = BigNumber.from(10).pow(6)
const _1e8 = BigNumber.from(10).pow(8)
const _1e12 = BigNumber.from(10).pow(12)
const _1e18 = ethers.constants.WeiPerEther

const homedir = require('os').homedir()
let conf = require(`${homedir}/.hubblenet.json`)
const url = `http://127.0.0.1:9650/ext/bc/${conf.chain_id}/rpc`

describe('Submit transaction and compare with EVM state', function () {
    before('', async function () {
        const provider = new ethers.providers.JsonRpcProvider(url);

        // Set up signer
        alice = new ethers.Wallet('56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027', provider);
        bob = new ethers.Wallet('31b571bf6894a248831ff937bb49f7754509fe93bbd2517c9c73c4144c0e97dc', provider);
        charlie = new ethers.Wallet('15614556be13730e9e8d6eacc1603143e7b96987429df8726384c2ec4502ef6e', provider);
        aliceAddress = alice.address.toLowerCase()
        bobAddress = bob.address.toLowerCase()
        charlieAddress = charlie.address.toLowerCase()
        console.log({ alice: aliceAddress, bob: bobAddress, charlie: charlieAddress });

        // Set up contract interface
        orderBook = new ethers.Contract(OrderBookContractAddress, require('./abi/OrderBook.json'));
        marginAccount = new ethers.Contract(MarginAccountContractAddress, require('./abi/MarginAccount.json'));
        marginAccountHelper = new ethers.Contract(MarginAccountHelperContractAddress, require('./abi/MarginAccountHelper.json'));
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

    it('Add margin', async function () {
        tx = await addMargin(alice, _1e6.mul(40))
        await tx.wait();

        tx = await addMargin(bob, _1e6.mul(40))
        await tx.wait();

        const evmState = await getEVMState()
        expect(evmState.trader_map[aliceAddress].margins["0"], 40000000)
        expect(evmState.trader_map[bobAddress].margins["0"], 40000000)
    });

    it('Remove margin', async function () {
        const tx = await marginAccount.connect(alice).removeMargin(0, _1e6.mul(10))
        await tx.wait();

        const evmState = await getEVMState()
        expect(evmState.trader_map[aliceAddress].margins["0"], 30000000)
        expect(evmState.trader_map[bobAddress].margins["0"], 40000000)
    });

    it('Place order', async function () {
        const { hash, order } = await placeOrder(alice, 5, 10)

        const evmState = await getEVMState()
        const evmOrder = evmState.order_map[hash]

        expectedOrder = {
            "market": 0,
            "position_type": "long",
            "user_address": "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC",
            "base_asset_quantity": 5000000000000000000,
            "filled_base_asset_quantity": 0,
            "price": 10000000,
        }
        expect(evmOrder).to.deep.include(expectedOrder)
        expect(evmOrder.lifecycle_list[0].Status).to.eq(0)
    });

    it('Match order', async function () {
        await placeOrder(bob, -5, 10)

        const evmState = await getEVMState()

        const expectedAlice = {
            "positions": {
                "0": {
                    "open_notional": 50000000,
                    "size": 5000000000000000000,
                    "unrealised_funding": null,
                    "last_premium_fraction": 0,
                    "liquidation_threshold": 5000000000000000000
                }
            },
            "margins": {
                "0": 29975000
            }
        }
        const expectedBob = {
            "positions": {
                "0": {
                    "open_notional": 50000000,
                    "size": -5000000000000000000,
                    "unrealised_funding": null,
                    "last_premium_fraction": 0,
                    "liquidation_threshold": -5000000000000000000
                }
            },
            "margins": {
                "0": 39975000
            }
        }
        expect(Object.keys(evmState.order_map).length).to.eq(0) // no open orders left
        expect(evmState.trader_map[aliceAddress]).to.deep.include(expectedAlice)
        expect(evmState.trader_map[bobAddress]).to.deep.include(expectedBob)

        expect(evmState.last_price["0"]).to.eq(10 * 1e6)
    });

    it('Partially match order', async function () {
        const {hash: longHash} = await placeOrder(alice, 5, 10)
        const {hash: shortHash} = await placeOrder(bob, -2, 10)

        const evmState = await getEVMState()

        const expectOrder = {
                "market": 0,
                "position_type": "long",
                "user_address": "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC",
                "base_asset_quantity": 5000000000000000000,
                "filled_base_asset_quantity": 2000000000000000000,
                "price": 10000000,
        }

        const expectAlice = {
            "positions": {
                "0": {
                    "open_notional": 70000000,
                    "size": 7000000000000000000,
                    "unrealised_funding": null,
                    "last_premium_fraction": 0,
                    "liquidation_threshold": 5000000000000000000
                }
            },
            "margins": {
                "0": 29965000
            }
        }
        const expectedBob = {
            "positions": {
                "0": {
                    "open_notional": 70000000,
                    "size": -7000000000000000000,
                    "unrealised_funding": null,
                    "last_premium_fraction": 0,
                    "liquidation_threshold": -5000000000000000000
                }
            },
            "margins": {
                "0": 39965000
            }
        }

        expect(evmState.order_map[longHash]).to.deep.include(expectOrder) // 1 partially filled open order left
        expect(evmState.order_map[shortHash]).to.eq(undefined)
        expect(evmState.trader_map[aliceAddress]).to.deep.eq(expectAlice)
        expect(evmState.trader_map[bobAddress]).to.deep.eq(expectedBob)
    });

    it('Order cancel', async function () {
        const { hash, order } = await placeOrder(alice, 2, 14)

        tx = await orderBook.connect(alice).cancelOrder(order)
        await tx.wait()

        const evmState = await getEVMState()

        expect(evmState.order_map[hash]).to.eq(undefined)
    });

    it('Order match error', async function () {
        const {hash: charlieHash} = await placeOrder(charlie, 50, 12)
        const {hash: bobHash} = await placeOrder(bob, -10, 12)

        const expectedBobOrder = {
            "market": 0,
            "position_type": "short",
            "user_address": "0x4Cf2eD3665F6bFA95cE6A11CFDb7A2EF5FC1C7E4",
            "base_asset_quantity": -10000000000000000000,
            "filled_base_asset_quantity": 0,
            "price": 12000000,
          }
        const evmState = await getEVMState()
        expect(evmState.order_map[charlieHash]).to.eq(undefined) // should be deleted
        expect(evmState.order_map[bobHash]).to.deep.contain(expectedBobOrder) // should be deleted
        expect(evmState.order_map[bobHash].lifecycle_list[0].Status).to.eq(0)
    });

    it('Liquidate trader', async function () {
        await addMargin(charlie, _1e6.mul(100))
        await addMargin(alice, _1e6.mul(200))
        await addMargin(bob, _1e6.mul(200))

        await sleep(3)

        // large position by charlie
        const {hash: charlieHash} = await placeOrder(charlie, 49, 10) // 46 + 3 is fulfilled
        const {hash: bobHash1} = await placeOrder(bob, -49, 10) // 46 + 3

        evmState = await getEVMState()
        // reduce the price
        const {hash: aliceHash} =  await placeOrder(alice, 10, 8) // 7 matched; 3 used for liquidation
        const {hash: bobHash2} = await placeOrder(bob, -10, 8) // 3 + 7

        // long order so that liquidation can run
        const {hash} = await placeOrder(alice, 10, 8) // 10 used for liquidation

        const expectedCharlie = {
            "positions": {
              "0": {
                "open_notional": 360000000, // 49 - 10 - 3(from a previous order)
                "size": 36000000000000000000,
                "unrealised_funding": null,
                "last_premium_fraction": 0,
                "liquidation_threshold": 12250000000000000000 // 49/4
              }
            },
            "margins": {
              "0": 68555000
            }
          }

        evmState = await getEVMState()
        expect(evmState.trader_map[charlieAddress]).to.deep.include(expectedCharlie)
        expect(evmState.order_map[hash]).to.eq(undefined) // should be completely fulfilled
        expect(evmState.order_map[charlieHash]).to.eq(undefined)
        expect(evmState.order_map[bobHash1]).to.eq(undefined)
        expect(evmState.order_map[bobHash2]).to.eq(undefined)
        expect(evmState.order_map[aliceHash]).to.eq(undefined)
    });
});

async function placeOrder(trader, size, price) {
    const order = {
        ammIndex: ZERO,
        trader: trader.address,
        baseAssetQuantity: ethers.utils.parseEther(size.toString()),
        price: ethers.utils.parseUnits(price.toString(), 6),
        salt: BigNumber.from(Date.now() + randomInt(100))
    }
    const signature = await trader._signTypedData(domain, orderType, order)
    const hash = await orderBook.connect(trader).getOrderHash(order)

    const tx = await orderBook.connect(trader).placeOrder(order, signature)
    await tx.wait()
    return { tx, hash, order }
}

function addMargin(trader, amount) {
    const hgtAmount = _1e12.mul(amount)
    return marginAccountHelper.connect(trader).addVUSDMarginWithReserve(amount, { value: hgtAmount })
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

function sleep(s) {
    console.log(`Requested a sleep of ${s} seconds...`)
    return new Promise(resolve => setTimeout(resolve, s * 1000));
}
