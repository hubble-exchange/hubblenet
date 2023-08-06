const { BigNumber } = require('ethers');
const { expect } = require('chai');

const utils = require('../utils')

const {
    _1e6,
    _1e18,
    addMargin,
    alice,
    charlie,
    clearingHouse,
    hubblebibliophile,
    multiplyPrice,
    multiplySize,
    placeOrder,
    removeAllAvailableMargin,
    waitForOrdersToMatch
} = utils

// Testing hubblebibliophile precompile contract 

describe('Testing getNotionalPositionAndMargin',async function () {
    charlieInitialMargin = multiplyPrice(BigNumber.from(150000))
    market = BigNumber.from(0)

    context('When notional position and margin are 0', async function () {
        it('should return 0 as notionalPosition and 0 as margin', async function () {
            result = await hubblebibliophile.getNotionalPositionAndMargin(charlie.address, false, 0)
            expect(result.notionalPosition.toString()).to.equal("0")
            expect(result.margin.toString()).to.equal("0")
        })
    })

    context('When notional position is zero but margin is non zero', async function () {
        it('should return 0 as notionalPosition and amount deposited as margin for trader', async function () {
            await addMargin(charlie, charlieInitialMargin)

            // Test without any open positions
            result = await hubblebibliophile.getNotionalPositionAndMargin(charlie.address, false, 0)

            //cleanup before assertions as any failure in assertion will not clean the state
            await removeAllAvailableMargin(charlie) 

            expect(result.notionalPosition.toString()).to.equal("0")
            expect(result.margin.toString()).to.equal(charlieInitialMargin.toString())
        })
    })

    context('When notional position and margin are both non zero', async function () {
        let aliceOrderPrice = multiplyPrice(1800)
        let charlieOrderPrice = multiplyPrice(1800)
        market = BigNumber.from(0)

        context('when user creates a position', async function () {
            it('returns the notional position and margin', async function () {
                let aliceOrderSize = multiplySize(0.1)
                let charlieOrderSize = multiplySize(-0.1)

                await addMargin(charlie, charlieInitialMargin)
                await addMargin(alice, charlieInitialMargin)
                // charlie places a short order 
                await placeOrder(market, charlie, charlieOrderSize, charlieOrderPrice)
                // alice places a long order
                await placeOrder(market, alice, aliceOrderSize, aliceOrderPrice)
                await waitForOrdersToMatch()

                result = await hubblebibliophile.getNotionalPositionAndMargin(charlie.address, false, 0)

                // got the response before cleanup
                // cleanup before assertions as any failure in assertion will not clean the state
                // charlie places a long order 
                await placeOrder(market, charlie, aliceOrderSize, charlieOrderPrice)
                // alice places a short order
                await placeOrder(market, alice, charlieOrderSize, aliceOrderPrice)
                await waitForOrdersToMatch() 
                await removeAllAvailableMargin(charlie)
                await removeAllAvailableMargin(alice)

                // tests
                takerFee = await clearingHouse.takerFee() // in 1e6 units
                aliceOrderFee = takerFee.mul(aliceOrderSize).mul(aliceOrderPrice).div(_1e18).div(_1e6)
                charlieOrderFee = takerFee.mul(charlieOrderSize.abs()).mul(charlieOrderPrice).div(_1e18).div(_1e6)
                expectedCharlieMargin = charlieInitialMargin.sub(charlieOrderFee)
                expectedNotionalPosition = charlieOrderSize.abs().mul(charlieOrderPrice).div(_1e18)
                expect(result.notionalPosition.toString()).to.equal(expectedNotionalPosition.toString())
                expect(result.margin.toString()).to.equal(expectedCharlieMargin.toString())
            })
        })

        context('when user increases the position', async function () {
            it('returns the notional position and margin', async function () {
                aliceOrder1Size = multiplySize(0.1)
                charlieOrder1Size = multiplySize(-0.1)
                aliceOrder2Size = multiplySize(0.2)
                charlieOrder2Size = multiplySize(-0.2)

                await addMargin(charlie, charlieInitialMargin)
                await addMargin(alice, charlieInitialMargin)
                //charlie is a maker for 1st order
                await placeOrder(market, charlie, charlieOrder1Size, charlieOrderPrice)
                await placeOrder(market, alice, aliceOrder1Size, aliceOrderPrice)
                await waitForOrdersToMatch()

                // increase position
                await placeOrder(market, alice, aliceOrder2Size, aliceOrderPrice)
                // charlie is taker for 2nd order
                await placeOrder(market, charlie, charlieOrder2Size, charlieOrderPrice)
                await waitForOrdersToMatch()

                result = await hubblebibliophile.getNotionalPositionAndMargin(charlie.address, false, 0)

                //cleanup
                totalAliceOrderSize = aliceOrder1Size.add(aliceOrder2Size)
                totalCharlieOrderSize = charlieOrder1Size.add(charlieOrder2Size)
                await addMargin(charlie, charlieInitialMargin)
                await placeOrder(market, charlie, totalAliceOrderSize, charlieOrderPrice)
                await placeOrder(market, alice, totalCharlieOrderSize, aliceOrderPrice)
                await waitForOrdersToMatch()
                await removeAllAvailableMargin(charlie)
                await removeAllAvailableMargin(alice)

                // tests
                makerFee = await clearingHouse.makerFee() // in 1e6 units
                charlieOrder1Fee = makerFee.mul(charlieOrder1Size.abs()).mul(charlieOrderPrice).div(_1e18).div(_1e6)
                takerFee = await clearingHouse.takerFee()
                charlieOrder2Fee = takerFee.mul(charlieOrder2Size.abs()).mul(charlieOrderPrice).div(_1e18).div(_1e6)
                expectedCharlieMargin = charlieInitialMargin.sub(charlieOrder1Fee).sub(charlieOrder2Fee)
                order1Notional = charlieOrder1Size.mul(charlieOrderPrice).div(_1e18)
                order2Notional = charlieOrder2Size.mul(charlieOrderPrice).div(_1e18)
                expectedNotionalPosition = order1Notional.add(order2Notional).abs()
                expect(result.notionalPosition.toString()).to.equal(expectedNotionalPosition.toString())
                expect(result.margin.toString()).to.equal(expectedCharlieMargin.toString())
            })
        })

        context('when user decreases the position', async function () {
            it('returns the notional position and margin', async function () {
                aliceOrder1Size = multiplySize(0.1)
                charlieOrder1Size = multiplySize(-0.1)
                aliceOrder2Size = multiplySize(-0.2)
                charlieOrder2Size = multiplySize(0.2)

                await addMargin(charlie, charlieInitialMargin)
                await addMargin(alice, charlieInitialMargin)
                //charlie is a maker for 1st order
                await placeOrder(market, charlie, charlieOrder1Size, charlieOrderPrice)
                await placeOrder(market, alice, aliceOrder1Size, aliceOrderPrice)
                await waitForOrdersToMatch()

                // increase position
                await placeOrder(market, alice, aliceOrder2Size, aliceOrderPrice)
                // charlie is taker for 2nd order
                await placeOrder(market, charlie, charlieOrder2Size, charlieOrderPrice)
                await waitForOrdersToMatch()

                result = await hubblebibliophile.getNotionalPositionAndMargin(charlie.address, false, 0)

                //cleanup
                totalAliceOrderSize = aliceOrder1Size.add(aliceOrder2Size)
                totalCharlieOrderSize = charlieOrder1Size.add(charlieOrder2Size)
                await addMargin(charlie, charlieInitialMargin)
                await placeOrder(market, charlie, totalAliceOrderSize, charlieOrderPrice)
                await placeOrder(market, alice, totalCharlieOrderSize, aliceOrderPrice)
                await waitForOrdersToMatch()
                await removeAllAvailableMargin(charlie)
                await removeAllAvailableMargin(alice)

                // tests
                makerFee = await clearingHouse.makerFee() // in 1e6 units
                charlieOrder1Fee = makerFee.mul(charlieOrder1Size.abs()).mul(charlieOrderPrice).div(_1e18).div(_1e6)
                takerFee = await clearingHouse.takerFee()
                charlieOrder2Fee = takerFee.mul(charlieOrder2Size.abs()).mul(charlieOrderPrice).div(_1e18).div(_1e6)
                expectedCharlieMargin = charlieInitialMargin.sub(charlieOrder1Fee).sub(charlieOrder2Fee)
                order1Notional = charlieOrder1Size.mul(charlieOrderPrice).div(_1e18)
                order2Notional = charlieOrder2Size.mul(charlieOrderPrice).div(_1e18)
                expectedNotionalPosition = order1Notional.add(order2Notional).abs()
                expect(result.notionalPosition.toString()).to.equal(expectedNotionalPosition.toString())
                expect(result.margin.toString()).to.equal(expectedCharlieMargin.toString())
            })
        })

        context('when user closes whole position', async function () {
            it('returns the notional position and margin', async function () {
                aliceOrder1Size = multiplySize(0.3)
                charlieOrder1Size = multiplySize(-0.3)
                aliceOrder2Size = multiplySize(-0.3)
                charlieOrder2Size = multiplySize(0.3)

                await addMargin(charlie, charlieInitialMargin)
                await addMargin(alice, charlieInitialMargin)
                //charlie is a maker for 1st order
                await placeOrder(market, charlie, charlieOrder1Size, charlieOrderPrice)
                await placeOrder(market, alice, aliceOrder1Size, aliceOrderPrice)
                await waitForOrdersToMatch()

                // close position
                await placeOrder(market, alice, aliceOrder2Size, aliceOrderPrice)
                // charlie is taker for 2nd order
                await placeOrder(market, charlie, charlieOrder2Size, charlieOrderPrice)
                await waitForOrdersToMatch()

                result = await hubblebibliophile.getNotionalPositionAndMargin(charlie.address, false, 0)
                
                //cleanup
                await removeAllAvailableMargin(charlie)
                await removeAllAvailableMargin(alice)

                // tests
                makerFee = await clearingHouse.makerFee() // in 1e6 units
                charlieOrder1Fee = makerFee.mul(charlieOrder1Size.abs()).mul(charlieOrderPrice).div(_1e18).div(_1e6)
                takerFee = await clearingHouse.takerFee()
                charlieOrder2Fee = takerFee.mul(charlieOrder2Size.abs()).mul(charlieOrderPrice).div(_1e18).div(_1e6)
                expectedCharlieMargin = charlieInitialMargin.sub(charlieOrder1Fee).sub(charlieOrder2Fee)
                order1Notional = charlieOrder1Size.mul(charlieOrderPrice).div(_1e18)
                order2Notional = charlieOrder2Size.mul(charlieOrderPrice).div(_1e18)
                expectedNotionalPosition = order1Notional.add(order2Notional).abs()
                expect(result.notionalPosition.toString()).to.equal(expectedNotionalPosition.toString())
                expect(result.margin.toString()).to.equal(expectedCharlieMargin.toString())
            })
        })
    })
})
