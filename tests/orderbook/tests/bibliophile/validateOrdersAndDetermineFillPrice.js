const { ethers, BigNumber } = require("ethers");
const { expect } = require("chai");
const utils = require("../utils")

const {
    alice,
    bob,
    governance,
    orderBook,
    clearingHouse,
    hubblebibliophile,
    getOrder,
    multiplySize,
    multiplyPrice,
    placeOrder,
    cancelOrder,
    _1e6,
    provider,
    addMargin,
    removeAllAvailableMargin,
} = utils

// Testing hubblebibliophile precompile contract 
describe("Test validateOrdersAndDetermineFillPrice", function () {
    beforeEach(async function () {
        market = 0
        longOrderBaseAssetQuantity = multiplySize(0.1) // 0.1 ether
        shortOrderBaseAssetQuantity = multiplySize(-0.1) // 0.1 ether
        longOrderPrice = multiplyPrice(1800)
        shortOrderPrice = multiplyPrice(1800)
        initialMargin = multiplyPrice(BigNumber.from(150000))
        await addMargin(alice, initialMargin)
        await addMargin(bob, initialMargin)
    });

    afterEach(async function () {
        await removeAllAvailableMargin(alice)
        await removeAllAvailableMargin(bob)
    })

    context("Basic validations", function () {
        it("returns error if longOrder's baseAssetQuantity is negative", async function () {
            orderPrice = longOrderPrice
            longOrder = getOrder(market, alice.address, shortOrderBaseAssetQuantity, orderPrice, getRandomSalt(), false)
            longOrderHash = await orderBook.getOrderHash(longOrder)

            shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, orderPrice, getRandomSalt(), false) 
            shortOrderHash = await orderBook.getOrderHash(shortOrder)

            try {
                await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], longOrderBaseAssetQuantity)
            } catch (error) {
                expect(error.error.body).to.match(/OB_order_0_is_not_long/)
                return
            }
            expect.fail('Expected throw not received');
        })

        it("returns error if shortOrder's baseAssetQuantity is positive", async function () {
            orderPrice = longOrderPrice
            longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, orderPrice, getRandomSalt(), false)
            longOrderHash = await orderBook.getOrderHash(longOrder)

            shortOrder = getOrder(market, bob.address, longOrderBaseAssetQuantity, orderPrice, getRandomSalt(), false) 
            shortOrderHash = await orderBook.getOrderHash(shortOrder)

            try {
                await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], longOrderBaseAssetQuantity)
            } catch (error) {
                expect(error.error.body).to.match(/OB_order_1_is_not_short/)
                return
            }
            expect.fail('Expected throw not received');
        })

        it("returns error if amm is different for long and short orders", async function () {
            longOrderMarket = 0
            shortOrderMarket = 1
            orderPrice = longOrderPrice

            longOrder = getOrder(longOrderMarket, alice.address, longOrderBaseAssetQuantity, orderPrice, getRandomSalt(), false)
            longOrderHash = await orderBook.getOrderHash(longOrder)

            shortOrder = getOrder(shortOrderMarket, bob.address, shortOrderBaseAssetQuantity, orderPrice, getRandomSalt(), false) 
            shortOrderHash = await orderBook.getOrderHash(shortOrder)

            try {
                await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], longOrderBaseAssetQuantity)
            } catch (error) {
                expect(error.error.body).to.match(/OB_orders_for_different_amms/)
                return
            }
            expect.fail('Expected throw not received');
        });

        it("returns error if longOrder's price is less than shortOrder's price", async function () {
            longOrderPrice = multiplyPrice(1800)
            shortOrderPrice = multiplyPrice(1900)

            longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
            longOrderHash = await orderBook.getOrderHash(longOrder)

            shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false) 
            shortOrderHash = await orderBook.getOrderHash(shortOrder)

            try {
                await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], longOrderBaseAssetQuantity)
            } catch (error) {
                expect(error.error.body).to.match(/OB_orders_do_not_match/)
                return
            }
            expect.fail('Expected throw not received');
        });
    });
    context("when either one or both order's status is not placed", function () {
        context("when both longOrder's and shortOrder's status is not placed", function () {
            it("returns error if both orders were never placed", async function () {
                longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false) 
                longOrderHash = await orderBook.getOrderHash(longOrder)
                shortOrderHash = await orderBook.getOrderHash(shortOrder)
                try {
                    await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], longOrderBaseAssetQuantity)
                } catch (error) {
                    expect(error.error.body).to.match(/OB_invalid_order/)
                    return
                }
                expect.fail('Expected throw not received');
            });
            it.skip("returns error if both orders were cancelled", async function () {
            });
            it.skip("returns error if both orders were filled", async function () {
            });
        });

        context("when longOrder's status is not placed and shortOrder's status is placed", async function () {
            it("returns error if longOrder does not exist", async function () {
                longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false) 
                await placeOrderFromOrder(shortOrder, bob)
                longOrderHash = await orderBook.getOrderHash(longOrder)
                shortOrderHash = await orderBook.getOrderHash(shortOrder)
                try {
                    await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], longOrderBaseAssetQuantity)
                } catch (error) {
                    expect(error.error.body).to.match(/OB_invalid_order/)
                    await cancelOrderFromOrder(shortOrder, bob)
                    return
                }
                expect.fail('Expected throw not received');
            });

            it("returns error if longOrder's status is cancelled", async function () {
                longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                await placeOrderFromOrder(longOrder, alice)
                await cancelOrderFromOrder(longOrder, alice)

                shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false) 
                await placeOrderFromOrder(shortOrder, bob)

                longOrderHash = await orderBook.getOrderHash(longOrder)
                shortOrderHash = await orderBook.getOrderHash(shortOrder)
                try {
                    await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], longOrderBaseAssetQuantity)
                } catch (error) {
                    expect(error.error.body).to.match(/OB_invalid_order/)
                    await cancelOrderFromOrder(shortOrder, bob)
                    return
                }
                expect.fail('Expected throw not received');
            });

            it("returns error if longOrder's status is filled", async function () {
                longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                await placeOrderFromOrder(longOrder, alice)

                shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false)
                await placeOrderFromOrder(shortOrder, bob)

                shortOrder2 = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false)
                await placeOrderFromOrder(shortOrder2, bob)

                longOrderHash = await orderBook.getOrderHash(longOrder)
                shortOrder2Hash = await orderBook.getOrderHash(shortOrder2)

                try {
                    await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder2], [longOrderHash, shortOrder2Hash], longOrderBaseAssetQuantity)
                } catch (error) {
                    //cleanup
                    await cancelOrderFromOrder(shortOrder2, bob)
                    aliceOppositeOrder = getOrder(market, alice.address, shortOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), true)
                    bobOppositeOrder = getOrder(market, bob.address, longOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), true)
                    await placeOrderFromOrder(aliceOppositeOrder, alice)
                    await placeOrderFromOrder(bobOppositeOrder, bob)

                    expect(error.error.body).to.match(/OB_invalid_order/)
                    return
                }
                expect.fail('Expected throw not received');
            });
        });

        context("when longOrder's status is placed and shortOrder's status is not placed", async function () {
            it("returns error if shortOrder does not exist", async function () {
                longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                await placeOrderFromOrder(longOrder, alice)

                shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false) 
                longOrderHash = await orderBook.getOrderHash(longOrder)
                shortOrderHash = await orderBook.getOrderHash(shortOrder)
                try {
                    await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], longOrderBaseAssetQuantity)
                } catch (error) {
                    expect(error.error.body).to.match(/OB_invalid_order/)
                    await cancelOrderFromOrder(longOrder, alice)
                    return
                }
                expect.fail('Expected throw not received');
            });

            it("returns error if shortOrder's status is cancelled", async function () {
                shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false) 
                await placeOrderFromOrder(shortOrder, bob)
                await cancelOrderFromOrder(shortOrder, bob)

                // placing it after else orders will be filled
                longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                await placeOrderFromOrder(longOrder, alice)

                longOrderHash = await orderBook.getOrderHash(longOrder)
                shortOrderHash = await orderBook.getOrderHash(shortOrder)
                try {
                    await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], longOrderBaseAssetQuantity)
                } catch (error) {
                    expect(error.error.body).to.match(/OB_invalid_order/)
                    await cancelOrderFromOrder(longOrder, alice)
                    return
                }
                expect.fail('Expected throw not received');
            });

            it("returns error if shortOrder's status is filled", async function () {
                shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false)
                await placeOrderFromOrder(shortOrder, bob)

                longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                await placeOrderFromOrder(longOrder, alice)

                longOrder2 = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                await placeOrderFromOrder(longOrder2, alice)

                longOrder2Hash = await orderBook.getOrderHash(longOrder2)
                shortOrderHash = await orderBook.getOrderHash(shortOrder)

                try {
                    await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder2, shortOrder], [longOrder2Hash, shortOrderHash], longOrderBaseAssetQuantity)
                } catch (error) {
                    //cleanup
                    await cancelOrderFromOrder(longOrder2, alice)
                    aliceOppositeOrder = getOrder(market, alice.address, shortOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), true)
                    bobOppositeOrder = getOrder(market, bob.address, longOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), true)
                    await placeOrderFromOrder(aliceOppositeOrder, alice)
                    await placeOrderFromOrder(bobOppositeOrder, bob)

                    expect(error.error.body).to.match(/OB_invalid_order/)
                    return
                }
                expect.fail('Expected throw not received');
            });
        });
    });

    context('When both longOrder and shortOrder status is placed', async function () {
        beforeEach(async function () {
            await disableValidatorMatching()
        });
        afterEach(async function () {
            await enableValidatorMatching()
        });

        context("When fillAmount is not multiple of minSizeRequirement", async function () {
            it.skip("returns error if fillAmount is zero", async function () {
                longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                longOrderHash = await orderBook.getOrderHash(longOrder)
                await placeOrderFromOrder(longOrder, alice)

                shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false)
                shortOrderHash = await orderBook.getOrderHash(shortOrder)
                await placeOrderFromOrder(shortOrder, bob)

                fillAmount = BigNumber.from(0)

                try {
                    await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], fillAmount)
                } catch (error) {
                    expect(error.error.body).to.match(/OB.not_multiple/)
                    await cancelOrderFromOrder(longOrder, alice)
                    await cancelOrderFromOrder(shortOrder, bob)
                    return
                }
                expect.fail('Expected throw not received');
            });

            it("returns error if fillAmount is not zero but less than minSizeRequirement", async function () {
                const ammAddress = await clearingHouse.amms(market)
                const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
                minSizeRequirement = await amm.minSizeRequirement()
                fillAmount = minSizeRequirement.sub(1)

                longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                longOrderHash = await orderBook.getOrderHash(longOrder)
                await placeOrderFromOrder(longOrder, alice)

                shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false)
                shortOrderHash = await orderBook.getOrderHash(shortOrder)
                await placeOrderFromOrder(shortOrder, bob)

                try {
                    await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], fillAmount)
                } catch (error) {
                    expect(error.error.body).to.match(/OB.not_multiple/)
                    await cancelOrderFromOrder(longOrder, alice)
                    await cancelOrderFromOrder(shortOrder, bob)
                    return
                }     
                expect.fail('Expected throw not received');
            });

            it("returns error if fillAmount is > minSizeRequirement but not multiple of minSizeRequirement", async function () {
                const ammAddress = await clearingHouse.amms(market)
                const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
                minSizeRequirement = await amm.minSizeRequirement()
                fillAmount = minSizeRequirement.mul(3).sub(1)

                longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                longOrderHash = await orderBook.getOrderHash(longOrder)
                await placeOrderFromOrder(longOrder, alice)

                shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false)
                shortOrderHash = await orderBook.getOrderHash(shortOrder)
                await placeOrderFromOrder(shortOrder, bob)

                try {
                    await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], fillAmount)
                } catch (error) {
                    expect(error.error.body).to.match(/OB.not_multiple/)
                    await cancelOrderFromOrder(longOrder, alice)
                    await cancelOrderFromOrder(shortOrder, bob)
                    return
                }     
                expect.fail('Expected throw not received');
            });
        });

        context("when fillAmount is multiple of minSizeRequirement", async function () {
            it("returns error if longOrder price is less than lowerBoundPrice", async function () {
                const ammAddress = await clearingHouse.amms(market)
                const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
                minSizeRequirement = await amm.minSizeRequirement()
                fillAmount = minSizeRequirement.mul(3)
                maxOracleSpreadRatio = await amm.maxOracleSpreadRatio()
                oraclePrice = await amm.getUnderlyingPrice()
                lowerBoundPrice = oraclePrice.mul(_1e6.sub(maxOracleSpreadRatio)).div(_1e6)

                longOrderPrice = lowerBoundPrice.sub(1)
                longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                await placeOrderFromOrder(longOrder, alice)
                longOrderHash = await orderBook.getOrderHash(longOrder)
                
                shortOrderPrice = longOrderPrice
                shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false)
                shortOrderHash = await orderBook.getOrderHash(shortOrder)
                await placeOrderFromOrder(shortOrder, bob)

                try {
                    await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], fillAmount)
                } catch (error) {
                    expect(error.error.body).to.match(/OB_long_order_price_too_low/)
                    await cancelOrderFromOrder(longOrder, alice)
                    await cancelOrderFromOrder(shortOrder, bob)
                    return
                }
            });

            it("returns error if shortPrice is greater than upperBoundPrice", async function () {
                const ammAddress = await clearingHouse.amms(market)
                const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
                minSizeRequirement = await amm.minSizeRequirement()
                fillAmount = minSizeRequirement.mul(3)
                maxOracleSpreadRatio = await amm.maxOracleSpreadRatio()
                oraclePrice = await amm.getUnderlyingPrice()
                lowerBoundPrice = oraclePrice.mul(_1e6.sub(maxOracleSpreadRatio)).div(_1e6)
                upperBoundPrice = oraclePrice.mul(_1e6.add(maxOracleSpreadRatio)).div(_1e6)

                longOrderPrice = upperBoundPrice.add(1)
                longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                await placeOrderFromOrder(longOrder, alice)
                longOrderHash = await orderBook.getOrderHash(longOrder)
                
                shortOrderPrice = upperBoundPrice.add(1)
                shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false)
                shortOrderHash = await orderBook.getOrderHash(shortOrder)
                await placeOrderFromOrder(shortOrder, bob)

                try {
                    await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], fillAmount)
                } catch (error) {
                    expect(error.error.body).to.match(/OB_short_order_price_too_high/)
                    await cancelOrderFromOrder(longOrder, alice)
                    await cancelOrderFromOrder(shortOrder, bob)
                    return
                }
            });

            context("when longOrder price is > lowerBound and shortOrder price is < upperBound", async function () {
                context("When longOrder was placed in earlier block than shortOrder", async function () {
                    it("returns longOrder's price as fillPrice if longOrder price is greater than lowerBoundPrice but less than upperBoundPrice", async function () {
                        const ammAddress = await clearingHouse.amms(market)
                        const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
                        minSizeRequirement = await amm.minSizeRequirement()
                        fillAmount = minSizeRequirement.mul(3)
                        maxOracleSpreadRatio = await amm.maxOracleSpreadRatio()
                        oraclePrice = await amm.getUnderlyingPrice()
                        lowerBoundPrice = oraclePrice.mul(_1e6.sub(maxOracleSpreadRatio)).div(_1e6)
                        upperBoundPrice = oraclePrice.mul(_1e6.add(maxOracleSpreadRatio)).div(_1e6)
                        longOrderPrice = upperBoundPrice.sub(1)

                        longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                        await placeOrderFromOrder(longOrder, alice)
                        longOrderHash = await orderBook.getOrderHash(longOrder)
                        
                        shortOrderPrice = longOrderPrice
                        shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false)
                        shortOrderHash = await orderBook.getOrderHash(shortOrder)
                        await placeOrderFromOrder(shortOrder, bob)

                        response = await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], fillAmount)
                        expect(response.fillPrice.toString()).to.equal(longOrderPrice.toString())
                        expect(response.mode0).to.equal(1)
                        expect(response.mode1).to.equal(0)
                        await cancelOrderFromOrder(longOrder, alice)
                        await cancelOrderFromOrder(shortOrder, bob)
                    });

                    it("returns upperBound as fillPrice if longOrder price is greater than upperBoundPrice", async function () {
                        const ammAddress = await clearingHouse.amms(market)
                        const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
                        minSizeRequirement = await amm.minSizeRequirement()
                        fillAmount = minSizeRequirement.mul(3)
                        maxOracleSpreadRatio = await amm.maxOracleSpreadRatio()
                        oraclePrice = await amm.getUnderlyingPrice()
                        lowerBoundPrice = oraclePrice.mul(_1e6.sub(maxOracleSpreadRatio)).div(_1e6)
                        upperBoundPrice = oraclePrice.mul(_1e6.add(maxOracleSpreadRatio)).div(_1e6)

                        longOrderPrice = upperBoundPrice.add(1)
                        longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                        await placeOrderFromOrder(longOrder, alice)
                        longOrderHash = await orderBook.getOrderHash(longOrder)
                        
                        shortOrderPrice = upperBoundPrice.sub(1)
                        shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false)
                        shortOrderHash = await orderBook.getOrderHash(shortOrder)
                        await placeOrderFromOrder(shortOrder, bob)

                        response = await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], fillAmount)
                        expect(response.fillPrice.toString()).to.equal(upperBoundPrice.toString())
                        expect(response.mode0).to.equal(1)
                        expect(response.mode1).to.equal(0)
                        await cancelOrderFromOrder(longOrder, alice)
                        await cancelOrderFromOrder(shortOrder, bob)
                    })
                });

                context("When shortOrder was placed in same or earlier block than longOrder", async function () {
                    it("returns shortOrder's price as fillPrice if shortOrder price is less than upperBoundPrice greater than lowerBoundPrice", async function () {
                        const ammAddress = await clearingHouse.amms(market)
                        const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
                        minSizeRequirement = await amm.minSizeRequirement()
                        fillAmount = minSizeRequirement.mul(3)
                        maxOracleSpreadRatio = await amm.maxOracleSpreadRatio()
                        oraclePrice = await amm.getUnderlyingPrice()
                        lowerBoundPrice = oraclePrice.mul(_1e6.sub(maxOracleSpreadRatio)).div(_1e6)
                        upperBoundPrice = oraclePrice.mul(_1e6.add(maxOracleSpreadRatio)).div(_1e6)

                        shortOrderPrice = upperBoundPrice.sub(1)
                        shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false)
                        shortOrderHash = await orderBook.getOrderHash(shortOrder)
                        await placeOrderFromOrder(shortOrder, bob)

                        longOrderPrice = shortOrderPrice
                        longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                        await placeOrderFromOrder(longOrder, alice)
                        longOrderHash = await orderBook.getOrderHash(longOrder)
                        
                        response = await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], fillAmount)
                        expect(response.fillPrice.toString()).to.equal(shortOrderPrice.toString())
                        expect(response.mode0).to.equal(0)
                        expect(response.mode1).to.equal(1)
                        await cancelOrderFromOrder(longOrder, alice)
                        await cancelOrderFromOrder(shortOrder, bob)
                    });

                    it("returns lowerBoundPrice price as fillPrice if shortOrder's price is less than lowerBoundPrice", async function () {
                        const ammAddress = await clearingHouse.amms(market)
                        const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
                        minSizeRequirement = await amm.minSizeRequirement()
                        fillAmount = minSizeRequirement.mul(3)
                        maxOracleSpreadRatio = await amm.maxOracleSpreadRatio()
                        oraclePrice = await amm.getUnderlyingPrice()
                        lowerBoundPrice = oraclePrice.mul(_1e6.sub(maxOracleSpreadRatio)).div(_1e6)
                        upperBoundPrice = oraclePrice.mul(_1e6.add(maxOracleSpreadRatio)).div(_1e6)

                        shortOrderPrice = lowerBoundPrice.sub(1)
                        shortOrder = getOrder(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false)
                        shortOrderHash = await orderBook.getOrderHash(shortOrder)
                        await placeOrderFromOrder(shortOrder, bob)

                        longOrderPrice = shortOrderPrice.add(1)
                        longOrder = getOrder(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false)
                        await placeOrderFromOrder(longOrder, alice)
                        longOrderHash = await orderBook.getOrderHash(longOrder)
                        
                        response = await hubblebibliophile.validateOrdersAndDetermineFillPrice([longOrder, shortOrder], [longOrderHash, shortOrderHash], fillAmount)
                        expect(response.fillPrice.toString()).to.equal(lowerBoundPrice.toString())
                        expect(response.mode0).to.equal(0)
                        expect(response.mode1).to.equal(1)
                        await cancelOrderFromOrder(longOrder, alice)
                        await cancelOrderFromOrder(shortOrder, bob)
                    });
                });
            });
        });
    })
});


async function placeOrderFromOrder(order, trader) {
    return placeOrder(order.ammIndex, trader, order.baseAssetQuantity, order.price, order.salt, order.reduceOnly)
}

function getRandomSalt() {
    return BigNumber.from(Date.now())
}

async function cancelOrderFromOrder(order, trader) {
    return cancelOrder(order.ammIndex, trader, order.baseAssetQuantity, order.price, order.salt, order.reduceOnly)
}

async function enableValidatorMatching() {
    const tx = await orderBook.connect(governance).setValidatorStatus(ethers.utils.getAddress('0x4Cf2eD3665F6bFA95cE6A11CFDb7A2EF5FC1C7E4'), true)
    await tx.wait()
}

async function disableValidatorMatching() {
    const tx = await orderBook.connect(governance).setValidatorStatus(ethers.utils.getAddress('0x4Cf2eD3665F6bFA95cE6A11CFDb7A2EF5FC1C7E4'), false)
    await tx.wait()
}

async function waitForOrdersToMatch() {
    await sleep(10)
}