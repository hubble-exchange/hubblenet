const { expect } = require("chai")
const { BigNumber } = require("ethers")

const gasLimit = 5e6 // subnet genesis file only allows for this much

const {
    _1e6,
    _1e18,
    alice,
    bob,
    charlie,
    getEventsFromOrderBookTx,
    getOrderV2,
    getRandomSalt,
    getRequiredMarginForLongOrder,
    getRequiredMarginForShortOrder,
    juror,
    multiplyPrice,
    multiplySize,
    removeAllAvailableMargin,
    orderBook,
    addMargin,
    clearingHouse,
    placeOrderFromLimitOrderV2,
    cancelOrderFromLimitOrderV2,
    getOrderBookEvents,
    marginAccount,
    waitForOrdersToMatch,
} = require("../utils")

describe.only("Juror tests", async function() {
    context("Alice is a new user and tries to place a valid longOrder", async function() {
        // Alice is a new user and tries to place a valid longOrder - should fail
        // After user adds margin and tries to place a valid order - should succeed
        // check if margin is reserved
        // User tries to place same order again - should fail
        // Cancel order - should succeed
        // try cancel same order again - should fail
        // available margin should be amount deposited
        let longOrderBaseAssetQuantity = multiplySize(0.1) // 0.1 ether
        let orderPrice = multiplyPrice(1800)
        let market = BigNumber.from(0)
        let longOrder = getOrderV2(market, alice.address, longOrderBaseAssetQuantity, orderPrice, getRandomSalt()) 

        it("should fail as trader has not margin", async function() {
            await removeAllAvailableMargin(alice)
            output = await juror.validatePlaceLimitOrder(longOrder, alice.address)
            expect(output.err).to.equal("insufficient margin")
            expectedOrderHash = await orderBook.getOrderHashV2(longOrder)
            expect(output.orderHash).to.equal(expectedOrderHash)
            expect(output.res.reserveAmount.toNumber()).to.equal(0)
            expectedAmmAddress = await clearingHouse.amms(market)
            expect(output.res.amm).to.equal(expectedAmmAddress)
        })
        it("should succeed after trader deposits margin and return reserve margin", async function() {
            totalRequiredMargin = await getRequiredMarginForLongOrder(longOrder.price, longOrder.baseAssetQuantity)
            await addMargin(alice, totalRequiredMargin)
            output = await juror.validatePlaceLimitOrder(longOrder, alice.address)
            expect(output.err).to.equal("")
            expectedOrderHash = await orderBook.getOrderHashV2(longOrder)
            expect(output.orderHash).to.equal(expectedOrderHash)
            expect(output.res.reserveAmount.toNumber()).to.equal(totalRequiredMargin.toNumber())
            expectedAmmAddress = await clearingHouse.amms(market)
            expect(output.res.amm).to.equal(expectedAmmAddress)

            // place the order
            output = await placeOrderFromLimitOrderV2(longOrder, alice)
            orderStatus = await orderBook.orderStatus(expectedOrderHash)
            expect(orderStatus.status).to.equal(1)
            expect(orderStatus.reservedMargin.toNumber()).to.equal(totalRequiredMargin.toNumber())
            expect(orderStatus.blockPlaced.toNumber()).to.equal(output.txReceipt.blockNumber)
            expect(orderStatus.filledAmount.toNumber()).to.equal(0)
        })
        it("should emit OrderRejected if trader tries to place same order again", async function() {
            output = await juror.validatePlaceLimitOrder(longOrder, alice.address)
            expect(output.err).to.equal("order already exists")
            expectedOrderHash = await orderBook.getOrderHashV2(longOrder)
            expect(output.orderHash).to.equal(expectedOrderHash)
            expect(output.res.reserveAmount.toNumber()).to.equal(0)
            expectedAmmAddress = await clearingHouse.amms(market)
            expect(output.res.amm).to.equal(expectedAmmAddress)

            output = await placeOrderFromLimitOrderV2(longOrder, alice)
            events = await getOrderBookEvents(output.txReceipt.blockNumber)
            var orderBookLogWithEvent
            for(i = 0; i < events.length; i++) {
                if(events[i].transactionHash == output.txReceipt.transactionHash) {
                    orderBookLogWithEvent = events[i]
                    break
                }
            }
            expect(orderBookLogWithEvent.event).to.equal("OrderRejected")
            expect(orderBookLogWithEvent.args.err).to.equal("order already exists")
            expect(orderBookLogWithEvent.args.orderHash).to.equal(expectedOrderHash)
            expect(orderBookLogWithEvent.args.trader).to.equal(alice.address)
        })
        it("should succeed if trader cancels order", async function() {
            output = await juror.validateCancelLimitOrder(longOrder, alice.address, false)
            expect(output.err).to.equal("")
            expectedOrderHash = await orderBook.getOrderHashV2(longOrder)
            expect(output.orderHash).to.equal(expectedOrderHash)
            expect(output.res.unfilledAmount.toString()).to.equal(longOrder.baseAssetQuantity.toString())
            expect(output.res.amm).to.equal(await clearingHouse.amms(market))

            await cancelOrderFromLimitOrderV2(longOrder, alice)
            orderStatus = await orderBook.orderStatus(expectedOrderHash)
            expect(orderStatus.status).to.equal(3)
            expect(orderStatus.reservedMargin.toNumber()).to.equal(0)
            expect(orderStatus.blockPlaced.toNumber()).to.equal(0)
            expect(orderStatus.filledAmount.toNumber()).to.equal(0)
        })
        it("should fail if trader tries to cancel same order again", async function() {
            output = await juror.validateCancelLimitOrder(longOrder, alice.address, false)
            expect(output.err).to.equal("Cancelled")
            expectedOrderHash = await orderBook.getOrderHashV2(longOrder)
            expect(output.orderHash).to.equal(expectedOrderHash)
            expect(output.res.unfilledAmount.toString()).to.equal("0")
            expect(output.res.amm).to.equal("0x0000000000000000000000000000000000000000")

            output = await cancelOrderFromLimitOrderV2(longOrder, alice)
            events = await getOrderBookEvents(output.txReceipt.blockNumber)
            var orderBookLogWithEvent
            for(i = 0; i < events.length; i++) {
                if(events[i].transactionHash == output.txReceipt.transactionHash) {
                    orderBookLogWithEvent = events[i]
                    break
                }
            }
            expect(orderBookLogWithEvent.event).to.equal("OrderCancelRejected")
            expect(orderBookLogWithEvent.args.err).to.equal("Cancelled")
            expect(orderBookLogWithEvent.args.orderHash).to.equal(expectedOrderHash)
            expect(orderBookLogWithEvent.args.trader).to.equal(alice.address)
        })
        it("should have available margin equal to amount deposited", async function() {
            margin = await marginAccount.getAvailableMargin(alice.address)
            expect(margin.toNumber()).to.equal(totalRequiredMargin.toNumber())
        })
    })
    context("Bob is a new user and trades via a trading authority", async function() {
        // Bob is also a new user and trades via a trading authority
        // Trading authority tries to place a valid shortOrder from bob without authorization - should fail
        // bob authorizes trading authority to place orders on his behalf
        // trading authority tries to place a valid shortOrder from bob with authorization - should succeed
        // Place same order again via trading authority - should fail
        // Cancel order via trading authority - should succeed
        // Cancel same order again via trading authority - should fail
        // available margin should be amount deposited
        let shortOrderBaseAssetQuantity = multiplySize(-0.1) // 0.1 ether
        let orderPrice = multiplyPrice(1800)
        let market = BigNumber.from(0)
        let shortOrder = getOrderV2(market, bob.address, shortOrderBaseAssetQuantity, orderPrice, getRandomSalt()) 
        let tradingAuthority = charlie

        it("should fail as trader has no margin", async function() {
            await removeAllAvailableMargin(bob)
            output = await juror.validatePlaceLimitOrder(shortOrder, bob.address)
            expect(output.err).to.equal("insufficient margin")
            expectedOrderHash = await orderBook.getOrderHashV2(shortOrder)
            expect(output.orderHash).to.equal(expectedOrderHash)
            expect(output.res.reserveAmount.toNumber()).to.equal(0)
            expectedAmmAddress = await clearingHouse.amms(market)
            expect(output.res.amm).to.equal(expectedAmmAddress)
        })
        it("after depositing margin, it should fail if trading authority tries to place order without authorization", async function() {
            totalRequiredMargin = await getRequiredMarginForShortOrder(shortOrder.price, shortOrder.baseAssetQuantity) 
            await addMargin(bob, totalRequiredMargin)
            const tx = await orderBook.connect(bob).revokeTradingAuthority(tradingAuthority.address)
            await tx.wait()

            output = await juror.validatePlaceLimitOrder(shortOrder, tradingAuthority.address)
            expect(output.err).to.equal("no trading authority")
            expectedOrderHash = await orderBook.getOrderHashV2(shortOrder)
            expect(output.orderHash).to.equal(expectedOrderHash)
            expect(output.res.reserveAmount.toNumber()).to.equal(0)
            expectedAmmAddress = await clearingHouse.amms(market)
            expect(output.res.amm).to.equal("0x0000000000000000000000000000000000000000")
        })
        it("should succeed if trading authority tries to place order with authorization", async function() {
            const tx  = await orderBook.connect(bob).whitelistTradingAuthority(tradingAuthority.address)
            await tx.wait()

            output = await juror.validatePlaceLimitOrder(shortOrder, tradingAuthority.address)
            expect(output.err).to.equal("")
            expectedOrderHash = await orderBook.getOrderHashV2(shortOrder)
            expect(output.orderHash).to.equal(expectedOrderHash)
            expect(output.res.reserveAmount.toNumber()).to.equal(totalRequiredMargin.toNumber())
            expectedAmmAddress = await clearingHouse.amms(market)
            expect(output.res.amm).to.equal(expectedAmmAddress)

            // place the order
            output = await placeOrderFromLimitOrderV2(shortOrder, tradingAuthority)
            orderStatus = await orderBook.orderStatus(expectedOrderHash)
            expect(orderStatus.status).to.equal(1)
            expect(orderStatus.reservedMargin.toNumber()).to.equal(totalRequiredMargin.toNumber())
            expect(orderStatus.blockPlaced.toNumber()).to.equal(output.txReceipt.blockNumber)
            expect(orderStatus.filledAmount.toNumber()).to.equal(0)
        })
        it("should emit OrderRejected if trading authority tries to place same order again", async function() {
            output = await juror.validatePlaceLimitOrder(shortOrder, tradingAuthority.address)
            expect(output.err).to.equal("order already exists")
            expectedOrderHash = await orderBook.getOrderHashV2(shortOrder)
            expect(output.orderHash).to.equal(expectedOrderHash)
            expect(output.res.reserveAmount.toNumber()).to.equal(0)
            expectedAmmAddress = await clearingHouse.amms(market)
            expect(output.res.amm).to.equal(expectedAmmAddress)

            output = await placeOrderFromLimitOrderV2(shortOrder, tradingAuthority)
            events = await getOrderBookEvents(output.txReceipt.blockNumber)
            var orderBookLogWithEvent
            for(i = 0; i < events.length; i++) {
                if(events[i].transactionHash == output.txReceipt.transactionHash) {
                    orderBookLogWithEvent = events[i]
                    break
                }
            }
            expect(orderBookLogWithEvent.event).to.equal("OrderRejected")
            expect(orderBookLogWithEvent.args.err).to.equal("order already exists")
            expect(orderBookLogWithEvent.args.orderHash).to.equal(expectedOrderHash)
            expect(orderBookLogWithEvent.args.trader).to.equal(shortOrder.trader)
        })
        it("should succeed if trading authority cancels order", async function() {
            output = await juror.validateCancelLimitOrder(shortOrder, tradingAuthority.address, false)
            expect(output.err).to.equal("")
            expectedOrderHash = await orderBook.getOrderHashV2(shortOrder)
            expect(output.orderHash).to.equal(expectedOrderHash)
            expect(output.res.unfilledAmount.toString()).to.equal(shortOrder.baseAssetQuantity.toString())
            expect(output.res.amm).to.equal(await clearingHouse.amms(market))
            
            await cancelOrderFromLimitOrderV2(shortOrder, tradingAuthority)
            orderStatus = await orderBook.orderStatus(expectedOrderHash)
            expect(orderStatus.status).to.equal(3)
            expect(orderStatus.reservedMargin.toNumber()).to.equal(0)
            expect(orderStatus.blockPlaced.toNumber()).to.equal(0)
            expect(orderStatus.filledAmount.toNumber()).to.equal(0)
        })
        it("should fail if trading authority tries to cancel same order again", async function() {
            output = await juror.validateCancelLimitOrder(shortOrder, tradingAuthority.address, false)
            expect(output.err).to.equal("Cancelled")
            expectedOrderHash = await orderBook.getOrderHashV2(shortOrder)
            expect(output.orderHash).to.equal(expectedOrderHash)
            expect(output.res.unfilledAmount.toString()).to.equal("0")
            expect(output.res.amm).to.equal("0x0000000000000000000000000000000000000000")

            output = await cancelOrderFromLimitOrderV2(shortOrder, tradingAuthority)
            events = await getOrderBookEvents(output.txReceipt.blockNumber)
            var orderBookLogWithEvent
            for(i = 0; i < events.length; i++) {
                if(events[i].transactionHash == output.txReceipt.transactionHash) {
                    orderBookLogWithEvent = events[i]
                    break
                }
            }
            expect(orderBookLogWithEvent.event).to.equal("OrderCancelRejected")
            expect(orderBookLogWithEvent.args.err).to.equal("Cancelled")
            expect(orderBookLogWithEvent.args.orderHash).to.equal(expectedOrderHash)
            expect(orderBookLogWithEvent.args.trader).to.equal(shortOrder.trader)
        })
        it("should have available margin equal to amount deposited", async function() {
            margin = await marginAccount.getAvailableMargin(bob.address)
            expect(margin.toNumber()).to.equal(totalRequiredMargin.toNumber())
        })
    })

    context("Market maker is trying to place/cancel orders", async function() {
        // Market maker tries to place a valid postonly longOrder 1 - should pass
        // Market maker tries to place a valid postonly shortOrder1 - should pass
        // Market maker tries to place same order again - should fail
        // Market maker tries to place postonly longOrder2 with higher or same price - should fail
        // Market maker tries to place postonly longOrder2 with lower price - should succeed
        // Market maker tries to cancel longOrder1 and longOrder2 - should pass
        // Market maker tries to cancel same longOrders - should fail

        // Market maker tries to place same order again - should fail
        // Market maker tries to place postonly shortOrder2 with lower or same price - should fail
        // Market maker tries to place postonly shortOrder2 with higher price - should succeed(cancel order for cleanup)
        // Market maker tries to cancel shortOrder1 and shortOrder2 - should pass
        // Market maker tries to cancel same shortOrders - should fail
        let marketMaker = alice
        let shortOrderBaseAssetQuantity = multiplySize(-0.1) // 0.1 ether
        let longOrderBaseAssetQuantity = multiplySize(0.1) // 0.1 ether
        let longOrderPrice = multiplyPrice(1799)
        let shortOrderPrice = multiplyPrice(1801)
        let market = BigNumber.from(0)
        let longOrder = getOrderV2(market, marketMaker.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false, true)
        let shortOrder = getOrderV2(market, marketMaker.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false, true)


        this.beforeAll(async function() {
            await addMargin(marketMaker, multiplyPrice(150000))
        })
        this.afterAll(async function() {
            await removeAllAvailableMargin(marketMaker)
        })

        context("should succeed when market maker tries to place valid postonly orders in blank orderbook", async function() {
            it("should succeed if market maker tries to place a valid postonly longOrder", async function() {
                totalRequiredMargin = await getRequiredMarginForLongOrder(longOrder.price, longOrder.baseAssetQuantity) 
                output = await juror.validatePlaceLimitOrder(longOrder, marketMaker.address)
                expect(output.err).to.equal("")
                expectedOrderHash = await orderBook.getOrderHashV2(longOrder)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.reserveAmount.toNumber()).to.equal(totalRequiredMargin.toNumber())
                expectedAmmAddress = await clearingHouse.amms(market)
                expect(output.res.amm).to.equal(expectedAmmAddress)

                // place the order
                output = await placeOrderFromLimitOrderV2(longOrder, marketMaker)
                orderStatus = await orderBook.orderStatus(expectedOrderHash)
                expect(orderStatus.status).to.equal(1)
                expect(orderStatus.reservedMargin.toNumber()).to.equal(totalRequiredMargin.toNumber())
                expect(orderStatus.blockPlaced.toNumber()).to.equal(output.txReceipt.blockNumber)
                expect(orderStatus.filledAmount.toNumber()).to.equal(0)
            })
            it("should succeed if market maker tries to place a valid postonly shortOrder", async function() {
                totalRequiredMargin = await getRequiredMarginForShortOrder(shortOrder.price, shortOrder.baseAssetQuantity)
                output = await juror.validatePlaceLimitOrder(shortOrder, marketMaker.address)
                expect(output.err).to.equal("")
                expectedOrderHash = await orderBook.getOrderHashV2(shortOrder)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.reserveAmount.toNumber()).to.equal(totalRequiredMargin.toNumber())
                expectedAmmAddress = await clearingHouse.amms(market)
                expect(output.res.amm).to.equal(expectedAmmAddress)

                // place the order
                output = await placeOrderFromLimitOrderV2(shortOrder, marketMaker)
                orderStatus = await orderBook.orderStatus(expectedOrderHash)
                expect(orderStatus.status).to.equal(1)
                expect(orderStatus.reservedMargin.toNumber()).to.equal(totalRequiredMargin.toNumber())
                expect(orderStatus.blockPlaced.toNumber()).to.equal(output.txReceipt.blockNumber)
                expect(orderStatus.filledAmount.toNumber()).to.equal(0)
            })
        })
        context("should emit OrderRejected if market maker tries to place same orders again", async function() {
            it("should emit OrderRejected if market maker tries to place same longOrder again", async function() {
                output = await juror.validatePlaceLimitOrder(longOrder, marketMaker.address)
                expect(output.err).to.equal("order already exists")
                expectedOrderHash = await orderBook.getOrderHashV2(longOrder)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.reserveAmount.toNumber()).to.equal(0)
                expectedAmmAddress = await clearingHouse.amms(market)
                expect(output.res.amm).to.equal(expectedAmmAddress)

                output = await placeOrderFromLimitOrderV2(longOrder, marketMaker)
                events = await getOrderBookEvents(output.txReceipt.blockNumber)
                orderBookLogWithEvent = (await getEventsFromOrderBookTx(output.txReceipt.transactionHash))[0]
                expect(orderBookLogWithEvent.event).to.equal("OrderRejected")
                expect(orderBookLogWithEvent.args.err).to.equal("order already exists")
                expect(orderBookLogWithEvent.args.orderHash).to.equal(expectedOrderHash)
                expect(orderBookLogWithEvent.args.trader).to.equal(longOrder.trader)
            })

            it("should emit OrderRejected if market maker tries to place same shortOrder again", async function() {
                output = await juror.validatePlaceLimitOrder(shortOrder, marketMaker.address)
                expect(output.err).to.equal("order already exists")
                expectedOrderHash = await orderBook.getOrderHashV2(shortOrder)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.reserveAmount.toNumber()).to.equal(0)
                expectedAmmAddress = await clearingHouse.amms(market)
                expect(output.res.amm).to.equal(expectedAmmAddress)

                output = await placeOrderFromLimitOrderV2(shortOrder, marketMaker)
                events = await getOrderBookEvents(output.txReceipt.blockNumber)
                orderBookLogWithEvent = (await getEventsFromOrderBookTx(output.txReceipt.transactionHash))[0]
                expect(orderBookLogWithEvent.event).to.equal("OrderRejected")
                expect(orderBookLogWithEvent.args.err).to.equal("order already exists")
                expect(orderBookLogWithEvent.args.orderHash).to.equal(expectedOrderHash)
                expect(orderBookLogWithEvent.args.trader).to.equal(shortOrder.trader)
            })
        })
        context("when postonly order have potential matches in orderbook", async function() {
            // longOrder and shortOrder are present in orderbook.
            // asksHead = 1801 * 1e6
            // bidsHead = 1799 * 1e6
            it("should fail if market maker tries to place a postonly longOrder2 with higher or same price as shortOrder", async function() {
                samePrice = shortOrder.price
                longOrder2 = getOrderV2(market, marketMaker.address, longOrderBaseAssetQuantity, samePrice, getRandomSalt(), false, true)
                output = await juror.validatePlaceLimitOrder(longOrder2, marketMaker.address)
                expect(output.err).to.equal("crossing market")
                expectedOrderHash = await orderBook.getOrderHashV2(longOrder2)
                expect(output.orderHash).to.equal(expectedOrderHash)
                totalRequiredMarginForLongOrder2 = await getRequiredMarginForLongOrder(longOrder2.price, longOrder2.baseAssetQuantity)
                expect(output.res.reserveAmount.toNumber()).to.equal(totalRequiredMarginForLongOrder2.toNumber())
                expectedAmmAddress = await clearingHouse.amms(market)
                expect(output.res.amm).to.equal(expectedAmmAddress)

                // place the order
                output = await placeOrderFromLimitOrderV2(longOrder2, marketMaker)
                orderBookLogWithEvent = (await getEventsFromOrderBookTx(output.txReceipt.transactionHash))[0]
                expect(orderBookLogWithEvent.event).to.equal("OrderRejected")
                expect(orderBookLogWithEvent.args.err).to.equal("crossing market")
                expect(orderBookLogWithEvent.args.orderHash).to.equal(expectedOrderHash)
                expect(orderBookLogWithEvent.args.trader).to.equal(longOrder2.trader)

                higherPrice = shortOrderPrice.add(1)
                longOrder3 = getOrderV2(market, marketMaker.address, longOrderBaseAssetQuantity, higherPrice, getRandomSalt(), false, true)
                output = await juror.validatePlaceLimitOrder(longOrder3, marketMaker.address)
                expect(output.err).to.equal("crossing market")
                expectedOrderHash = await orderBook.getOrderHashV2(longOrder3)
                expect(output.orderHash).to.equal(expectedOrderHash)
                totalRequiredMarginForLongOrder3 = await getRequiredMarginForLongOrder(longOrder3.price, longOrder3.baseAssetQuantity)
                expect(output.res.reserveAmount.toNumber()).to.equal(totalRequiredMarginForLongOrder3.toNumber())
                expectedAmmAddress = await clearingHouse.amms(market)
                expect(output.res.amm).to.equal(expectedAmmAddress)

                // place the order
                output = await placeOrderFromLimitOrderV2(longOrder3, marketMaker)
                orderBookLogWithEvent = (await getEventsFromOrderBookTx(output.txReceipt.transactionHash))[0]
                expect(orderBookLogWithEvent.event).to.equal("OrderRejected")
                expect(orderBookLogWithEvent.args.err).to.equal("crossing market")
                expect(orderBookLogWithEvent.args.orderHash).to.equal(expectedOrderHash)
                expect(orderBookLogWithEvent.args.trader).to.equal(longOrder3.trader)
            })
            it("should fail if market maker tries to place a postonly shortOrder2 with lower or same price as longOrder", async function() {
                samePrice = longOrder.price
                shortOrder2 = getOrderV2(market, marketMaker.address, shortOrderBaseAssetQuantity, samePrice, getRandomSalt(), false, true)
                output = await juror.validatePlaceLimitOrder(shortOrder2, marketMaker.address)
                expect(output.err).to.equal("crossing market")
                expectedOrderHash = await orderBook.getOrderHashV2(shortOrder2)
                expect(output.orderHash).to.equal(expectedOrderHash)
                totalRequiredMarginForShortOrder2 = await getRequiredMarginForShortOrder(shortOrder2.price, shortOrder2.baseAssetQuantity) 
                expect(output.res.reserveAmount.toNumber()).to.equal(totalRequiredMarginForShortOrder2.toNumber())
                expectedAmmAddress = await clearingHouse.amms(market)
                expect(output.res.amm).to.equal(expectedAmmAddress)

                // place the order
                output = await placeOrderFromLimitOrderV2(shortOrder2, marketMaker)
                orderBookLogWithEvent = (await getEventsFromOrderBookTx(output.txReceipt.transactionHash))[0]
                expect(orderBookLogWithEvent.event).to.equal("OrderRejected")
                expect(orderBookLogWithEvent.args.err).to.equal("crossing market")
                expect(orderBookLogWithEvent.args.orderHash).to.equal(expectedOrderHash)
                expect(orderBookLogWithEvent.args.trader).to.equal(shortOrder2.trader)


                lowerPrice = longOrderPrice.sub(1)
                shortOrder3 = getOrderV2(market, marketMaker.address, shortOrderBaseAssetQuantity, lowerPrice, getRandomSalt(), false, true)
                output = await juror.validatePlaceLimitOrder(shortOrder3, marketMaker.address)
                expect(output.err).to.equal("crossing market")
                expectedOrderHash = await orderBook.getOrderHashV2(shortOrder3)
                expect(output.orderHash).to.equal(expectedOrderHash)
                totalRequiredMarginForShortOrder3 = await getRequiredMarginForShortOrder(shortOrder3.price, shortOrder3.baseAssetQuantity)
                expect(output.res.reserveAmount.toNumber()).to.equal(totalRequiredMarginForShortOrder3.toNumber())
                expectedAmmAddress = await clearingHouse.amms(market)
                
                // place the order
                output = await placeOrderFromLimitOrderV2(shortOrder3, marketMaker)
                orderBookLogWithEvent = (await getEventsFromOrderBookTx(output.txReceipt.transactionHash))[0]
                expect(orderBookLogWithEvent.event).to.equal("OrderRejected")
                expect(orderBookLogWithEvent.args.err).to.equal("crossing market")
                expect(orderBookLogWithEvent.args.orderHash).to.equal(expectedOrderHash)
                expect(orderBookLogWithEvent.args.trader).to.equal(shortOrder3.trader)
            })
        })
        context("when postonly order does not have potential matches in orderbook", async function() {
            it("should succeed if market maker tries to place another postonly longOrder with lower price than all shortOrders", async function() {
                lowerPrice = shortOrder.price.sub(1)
                longOrder4 = getOrderV2(market, marketMaker.address, longOrderBaseAssetQuantity, lowerPrice, getRandomSalt(), false, true)
                totalRequiredMargin = await getRequiredMarginForLongOrder(longOrder4.price, longOrder4.baseAssetQuantity)
                output = await juror.validatePlaceLimitOrder(longOrder4, marketMaker.address)
                expect(output.err).to.equal("")
                expectedOrderHash = await orderBook.getOrderHashV2(longOrder4)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.reserveAmount.toNumber()).to.equal(totalRequiredMargin.toNumber())
                expectedAmmAddress = await clearingHouse.amms(market)
                expect(output.res.amm).to.equal(expectedAmmAddress)

                // place the order
                output = await placeOrderFromLimitOrderV2(longOrder4, marketMaker)
                orderStatus = await orderBook.orderStatus(expectedOrderHash)
                expect(orderStatus.status).to.equal(1)
                expect(orderStatus.reservedMargin.toNumber()).to.equal(totalRequiredMargin.toNumber())
                expect(orderStatus.blockPlaced.toNumber()).to.equal(output.txReceipt.blockNumber)
                expect(orderStatus.filledAmount.toNumber()).to.equal(0)
            })
            it("should succeed if market maker tries to place another postonly shortOrder with higher price than all longOrders", async function() {
                higherPrice = longOrder4.price.add(1)
                shortOrder4 = getOrderV2(market, marketMaker.address, shortOrderBaseAssetQuantity, higherPrice, getRandomSalt(), false, true)
                totalRequiredMargin = await getRequiredMarginForShortOrder(shortOrder4.price, shortOrder4.baseAssetQuantity)
                output = await juror.validatePlaceLimitOrder(shortOrder4, marketMaker.address)
                expect(output.err).to.equal("")
                expectedOrderHash = await orderBook.getOrderHashV2(shortOrder4)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.reserveAmount.toNumber()).to.equal(totalRequiredMargin.toNumber())
                expectedAmmAddress = await clearingHouse.amms(market)

                // place the order
                output = await placeOrderFromLimitOrderV2(shortOrder4, marketMaker)
                orderStatus = await orderBook.orderStatus(expectedOrderHash)
                expect(orderStatus.status).to.equal(1)
                expect(orderStatus.reservedMargin.toNumber()).to.equal(totalRequiredMargin.toNumber())
                expect(orderStatus.blockPlaced.toNumber()).to.equal(output.txReceipt.blockNumber)
                expect(orderStatus.filledAmount.toNumber()).to.equal(0)
            })
        })

        context("should succeed when market maker tries to cancel postonly orders", async function() {
            it("should succeed if market maker tries to cancel longOrder", async function() {
                // cancel longOrder
                output = await juror.validateCancelLimitOrder(longOrder, marketMaker.address, false)
                expect(output.err).to.equal("")
                expectedOrderHash = await orderBook.getOrderHashV2(longOrder)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.unfilledAmount.toString()).to.equal(longOrder.baseAssetQuantity.toString())
                expect(output.res.amm).to.equal(await clearingHouse.amms(market))
                
                await cancelOrderFromLimitOrderV2(longOrder, marketMaker)
                orderStatus = await orderBook.orderStatus(expectedOrderHash)
                expect(orderStatus.status).to.equal(3)
                expect(orderStatus.reservedMargin.toNumber()).to.equal(0)
                expect(orderStatus.blockPlaced.toNumber()).to.equal(0)
                expect(orderStatus.filledAmount.toNumber()).to.equal(0)

                // cancel longOrder4
                output = await juror.validateCancelLimitOrder(longOrder4, marketMaker.address, false)
                expect(output.err).to.equal("")
                expectedOrderHash = await orderBook.getOrderHashV2(longOrder4)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.unfilledAmount.toString()).to.equal(longOrder4.baseAssetQuantity.toString())
                expect(output.res.amm).to.equal(await clearingHouse.amms(market))

                await cancelOrderFromLimitOrderV2(longOrder4, marketMaker)
                orderStatus = await orderBook.orderStatus(expectedOrderHash)
                expect(orderStatus.status).to.equal(3)
                expect(orderStatus.reservedMargin.toNumber()).to.equal(0)
                expect(orderStatus.blockPlaced.toNumber()).to.equal(0)
                expect(orderStatus.filledAmount.toNumber()).to.equal(0)
            })
            it("should succeed if market maker tries to cancel shortOrder", async function() {
                // cancel shortOrder
                output = await juror.validateCancelLimitOrder(shortOrder, marketMaker.address, false)
                expect(output.err).to.equal("")
                expectedOrderHash = await orderBook.getOrderHashV2(shortOrder)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.unfilledAmount.toString()).to.equal(shortOrder.baseAssetQuantity.toString())
                expect(output.res.amm).to.equal(await clearingHouse.amms(market))

                await cancelOrderFromLimitOrderV2(shortOrder, marketMaker)
                orderStatus = await orderBook.orderStatus(expectedOrderHash)
                expect(orderStatus.status).to.equal(3)
                expect(orderStatus.reservedMargin.toNumber()).to.equal(0)
                expect(orderStatus.blockPlaced.toNumber()).to.equal(0)
                expect(orderStatus.filledAmount.toNumber()).to.equal(0)

                // cancel shortOrder4
                output = await juror.validateCancelLimitOrder(shortOrder4, marketMaker.address, false)
                expect(output.err).to.equal("")
                expectedOrderHash = await orderBook.getOrderHashV2(shortOrder4)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.unfilledAmount.toString()).to.equal(shortOrder4.baseAssetQuantity.toString())
                expect(output.res.amm).to.equal(await clearingHouse.amms(market))

                await cancelOrderFromLimitOrderV2(shortOrder4, marketMaker)
                orderStatus = await orderBook.orderStatus(expectedOrderHash)
                expect(orderStatus.status).to.equal(3)
                expect(orderStatus.reservedMargin.toNumber()).to.equal(0)
                expect(orderStatus.blockPlaced.toNumber()).to.equal(0)
                expect(orderStatus.filledAmount.toNumber()).to.equal(0)
            })
        })
        context("should fail if market maker tries to cancel same orders again", async function() {
            it("should fail if market maker tries to cancel same longOrders again", async function() {
                // cancel longOrder
                output = await juror.validateCancelLimitOrder(longOrder, marketMaker.address, false)
                expect(output.err).to.equal("Cancelled")
                expectedOrderHash = await orderBook.getOrderHashV2(longOrder)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.unfilledAmount.toString()).to.equal("0")
                expect(output.res.amm).to.equal("0x0000000000000000000000000000000000000000")

                // cancel longOrder4
                output = await juror.validateCancelLimitOrder(longOrder4, marketMaker.address, false)
                expect(output.err).to.equal("Cancelled")
                expectedOrderHash = await orderBook.getOrderHashV2(longOrder4)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.unfilledAmount.toString()).to.equal("0")
                expect(output.res.amm).to.equal("0x0000000000000000000000000000000000000000")
            })
            it("should fail if market maker tries to cancel same shortOrders again", async function() {
                // cancel shortOrder
                output = await juror.validateCancelLimitOrder(shortOrder, marketMaker.address, false)
                expect(output.err).to.equal("Cancelled")
                expectedOrderHash = await orderBook.getOrderHashV2(shortOrder)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.unfilledAmount.toString()).to.equal("0")

                // cancel shortOrder4
                output = await juror.validateCancelLimitOrder(shortOrder4, marketMaker.address, false)
                expect(output.err).to.equal("Cancelled")
                expectedOrderHash = await orderBook.getOrderHashV2(shortOrder4)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.unfilledAmount.toString()).to.equal("0")
            })
        })
    })
    context.only("When users have positions and then try to place/cancel orders", async function() {
        // Alice has long Position and bob has short position
        // Alice tries to close half of her position via ui(so places reduceOnly order)
        // If reduceOnly order is longOrder - it should fail
        // If reduceOnly order is shortOrder - it should succeed
        // if there are open shortOrder for alice reduceOnly order should fail
        // if order"s size + openReduceOnlyAmount > posSize of alice - it should fail
        let shortOrderBaseAssetQuantity = multiplySize(-0.1) // 0.1 ether
        let longOrderBaseAssetQuantity = multiplySize(0.1) // 0.1 ether
        let longOrderPrice = multiplyPrice(1800)
        let shortOrderPrice = multiplyPrice(1800)
        let market = BigNumber.from(0)
        let longOrder = getOrderV2(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false, false)
        let shortOrder = getOrderV2(market, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false, false)

        this.beforeAll(async function() {
            await addMargin(alice, multiplyPrice(150000))
            await addMargin(bob, multiplyPrice(150000))
            await placeOrderFromLimitOrderV2(longOrder, alice)
            await placeOrderFromLimitOrderV2(shortOrder, bob)
        })
        this.afterAll(async function() {
            let oppositeShortOrder = getOrderV2(market, alice.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false, false)
            let oppositeLongOrder = getOrderV2(market, bob.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false, false)
            await placeOrderFromLimitOrderV2(oppositeShortOrder, alice)
            await placeOrderFromLimitOrderV2(oppositeLongOrder, bob)
            await removeAllAvailableMargin(alice)
            await removeAllAvailableMargin(bob)
        })

        context("alice has long position", async function() {
            it("should fail if alice tries to place a long reduceOnly order", async function() {
                //ensure position is created for alice
                orderStatus = await orderBook.orderStatus(await orderBook.getOrderHashV2(longOrder))
                expect(orderStatus.status).to.equal(2)
                expect(orderStatus.filledAmount.toString()).to.equal(longOrder.baseAssetQuantity.toString())
                expect(orderStatus.reservedMargin.toNumber()).to.equal(0)
                expect(orderStatus.blockPlaced.toNumber()).to.equal(0)

                orderSize = longOrderBaseAssetQuantity.div(2)
                let reduceOnlyLongOrder = getOrderV2(market, alice.address, orderSize, longOrderPrice, getRandomSalt(), true, false)
                output = await juror.validatePlaceLimitOrder(reduceOnlyLongOrder, alice.address)
                expect(output.err).to.equal("reduce only order must reduce position")
                expectedOrderHash = await orderBook.getOrderHashV2(reduceOnlyLongOrder)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.reserveAmount.toNumber()).to.equal(0)
                expectedAmmAddress = await clearingHouse.amms(market)
                expect(output.res.amm).to.equal(expectedAmmAddress)

                // place the order
                output = await placeOrderFromLimitOrderV2(reduceOnlyLongOrder, alice)
                orderBookLogWithEvent = (await getEventsFromOrderBookTx(output.txReceipt.transactionHash))[0]
                expect(orderBookLogWithEvent.event).to.equal("OrderRejected")
                expect(orderBookLogWithEvent.args.err).to.equal("reduce only order must reduce position")
                expect(orderBookLogWithEvent.args.orderHash).to.equal(expectedOrderHash)
                expect(orderBookLogWithEvent.args.trader).to.equal(reduceOnlyLongOrder.trader)
            })
            it("should fail when alice has a open shortOrder and tries to place a short reduceOnly order", async function() {
                let shortOrderBaseAssetQuantity = longOrderBaseAssetQuantity.div(2).mul(-1)
                let shortOrder = getOrderV2(market, alice.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), false, false)
                requiredMargin = await getRequiredMarginForShortOrder(shortOrder.price, shortOrder.baseAssetQuantity)
                await addMargin(alice, requiredMargin)
                await placeOrderFromLimitOrderV2(shortOrder, alice)

                let reduceOnlyShortOrder = getOrderV2(market, alice.address, shortOrderBaseAssetQuantity, shortOrderPrice, getRandomSalt(), true, false)
                output = await juror.validatePlaceLimitOrder(reduceOnlyShortOrder, alice.address)
                expect(output.err).to.equal("open orders")
                expectedOrderHash = await orderBook.getOrderHashV2(reduceOnlyShortOrder)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.reserveAmount.toNumber()).to.equal(0)
                expectedAmmAddress = await clearingHouse.amms(market)
                expect(output.res.amm).to.equal(expectedAmmAddress)

                // place the order
                output = await placeOrderFromLimitOrderV2(reduceOnlyShortOrder, alice)
                orderBookLogWithEvent = (await getEventsFromOrderBookTx(output.txReceipt.transactionHash))[0]
                expect(orderBookLogWithEvent.event).to.equal("OrderRejected")
                expect(orderBookLogWithEvent.args.err).to.equal("open orders")
                expect(orderBookLogWithEvent.args.orderHash).to.equal(expectedOrderHash)
                expect(orderBookLogWithEvent.args.trader).to.equal(reduceOnlyShortOrder.trader)

                await cancelOrderFromLimitOrderV2(shortOrder, alice)
            })
            it("should succeed if alice tries to place a short reduceOnly order", async function() {
                orderSize = longOrderBaseAssetQuantity.div(2).mul(-1)
                let reduceOnlyShortOrder = getOrderV2(market, alice.address, orderSize, shortOrderPrice, getRandomSalt(), true, false)
                output = await juror.validatePlaceLimitOrder(reduceOnlyShortOrder, alice.address)
                expect(output.err).to.equal("")
                expectedOrderHash = await orderBook.getOrderHashV2(reduceOnlyShortOrder)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.reserveAmount.toNumber()).to.equal(0)
                expectedAmmAddress = await clearingHouse.amms(market)
                expect(output.res.amm).to.equal(expectedAmmAddress)

                // place the order
                output = await placeOrderFromLimitOrderV2(reduceOnlyShortOrder, alice)
                orderStatus = await orderBook.orderStatus(expectedOrderHash)
                expect(orderStatus.status).to.equal(1)
                expect(orderStatus.reservedMargin.toNumber()).to.equal(0)
                expect(orderStatus.blockPlaced.toNumber()).to.equal(output.txReceipt.blockNumber)
                expect(orderStatus.filledAmount.toNumber()).to.equal(0)

            })
            it("should fail if alice tries to place a longOrder(reduceOnly=false) to increase her position after placing a short reduceOnly order", async function() {
                let longOrder2 = getOrderV2(market, alice.address, longOrderBaseAssetQuantity, longOrderPrice, getRandomSalt(), false, false)
                output = await juror.validatePlaceLimitOrder(longOrder2, alice.address)
                expect(output.err).to.equal("open reduce only orders")
                expectedOrderHash = await orderBook.getOrderHashV2(longOrder2)
                expect(output.orderHash).to.equal(expectedOrderHash)
                expect(output.res.reserveAmount.toNumber()).to.equal(0)
                expectedAmmAddress = await clearingHouse.amms(market)
                expect(output.res.amm).to.equal(expectedAmmAddress)

                // place the order
                output = await placeOrderFromLimitOrderV2(longOrder2, alice)
                orderBookLogWithEvent = (await getEventsFromOrderBookTx(output.txReceipt.transactionHash))[0]
                expect(orderBookLogWithEvent.event).to.equal("OrderRejected")
                expect(orderBookLogWithEvent.args.err).to.equal("open reduce only orders")
                expect(orderBookLogWithEvent.args.orderHash).to.equal(expectedOrderHash)
                expect(orderBookLogWithEvent.args.trader).to.equal(longOrder2.trader)

                // temporary, for cleanup - cancel reduceOnlyShortOrder
                await cancelOrderFromLimitOrderV2(reduceOnlyShortOrder, alice)
            })
            it("should fail if alice tries to place a short reduceOnly order with size >  posSize - reduceOnlyShortOrder.baseAssetQuantity", async function() {
            })
        })
    })
})
