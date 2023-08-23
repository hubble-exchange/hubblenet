const { expect } = require("chai");
const { BigNumber } = require("ethers");
const utils = require("../utils")

const {
    addMargin,
    alice,
    cancelOrderFromLimitOrder,
    getOrderV2,
    getRandomSalt,
    juror,
    multiplyPrice,
    multiplySize,
    placeOrderFromLimitOrder,
    removeAllAvailableMargin,
} = utils

describe("Testing ValidateCancelLimitOrder", async function() {
    market = BigNumber.from(0)
    longBaseAssetQuantity = multiplySize(0.1) 
    shortBaseAssetQuantity = multiplySize("-0.1") 
    price = multiplyPrice(1800)
    salt = getRandomSalt()
    initialMargin = multiplyPrice(500000)

    context("when order's status is not placed", async function() {
        context("when order's status is invalid", async function() {
            it("should return error", async function() {
                assertLowMargin = false
                longOrder = getOrderV2(market, longBaseAssetQuantity, price, salt)
                try {
                    await juror.validateCancelLimitOrder(longOrder, alice.address, assertLowMargin)
                } catch (error) {
                    error_message = JSON.parse(error.error.body).error.message
                    expect(error_message).to.equal("invalid order")
                }
                shortOrder = getOrderV2(market, shortBaseAssetQuantity, price, salt, true)
                try {
                    await juror.validateCancelLimitOrder(shortOrder, alice.address, assertLowMargin)
                } catch (error) {
                    error_message = JSON.parse(error.error.body).error.message
                    expect(error_message).to.equal("invalid order")
                    return
                }
                expect.fail("Expected throw not received");
            })
        })
        context("when order's status is cancelled", async function() {
            this.beforeEach(async function() {
                await addMargin(alice, initialMargin)
            })
            this.afterEach(async function() {
                await removeAllAvailableMargin(alice)
            })

            it("should return error", async function() {
                longOrder = getOrderV2(market, longBaseAssetQuantity, price, salt)
                console.log("placing order")
                await placeOrderFromLimitOrder(longOrder, alice)
                console.log("cancelling order")
                await cancelOrderFromLimitOrder(longOrder, alice)
                try {
                    await juror.validateCancelLimitOrder(longOrder, alice.address, assertLowMargin)
                } catch (error) {
                    error_message = JSON.parse(error.error.body).error.message
                    expect(error_message).to.equal("cancelled order")
                }
            })
        })
        it("should return error when order's status is filled", async function() {
        })
    })
})
