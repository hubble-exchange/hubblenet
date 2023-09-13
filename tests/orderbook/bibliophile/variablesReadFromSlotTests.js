const { ethers, BigNumber } = require('ethers');
const utils = require('../utils.js');
const chai = require('chai');
const { assert, expect } = chai;
let chaiHttp = require('chai-http');

chai.use(chaiHttp);

const {
    _1e6,
    _1e18,
    addMargin,
    alice,
    bob,
    cancelOrderFromLimitOrderV2,
    charlie,
    clearingHouse,
    getAMMContract,
    getIOCOrder,
    getOrderV2,
    getRequiredMarginForLongOrder,
    getRequiredMarginForShortOrder,
    governance,
    ioc,
    multiplyPrice,
    multiplySize,
    orderBook,
    placeOrderFromLimitOrderV2,
    placeIOCOrder,
    provider,
    removeAllAvailableMargin,
    url,
} = utils;



describe('Testing variables read from slots by precompile', function () {
    context("Clearing house contract variables", function () {
        // vars read from slot
        // minAllowableMargin, maintenanceMargin, takerFee, amms
        it("should read the correct value from contracts", async function () {
            method = "testing_getClearingHouseVars"
            params =[ charlie.address ]
            response = await makehttpCall(method, params)
            result = response.body.result

            actualMaintenanceMargin = await clearingHouse.maintenanceMargin()
            actualMinAllowableMargin = await clearingHouse.minAllowableMargin()
            actualTakerFee = await clearingHouse.takerFee()
            actualAmms = await clearingHouse.getAMMs()

            expect(result.maintenance_margin).to.equal(actualMaintenanceMargin.toNumber())
            expect(result.min_allowable_margin).to.equal(actualMinAllowableMargin.toNumber())
            expect(result.taker_fee).to.equal(actualTakerFee.toNumber())
            expect(result.amms.length).to.equal(actualAmms.length)
            for(let i = 0; i < result.amms.length; i++) {
                expect(result.amms[i].toLowerCase()).to.equal(actualAmms[i].toLowerCase())
            }
            newMaintenanceMargin = BigNumber.from(20000)
            newMinAllowableMargin = BigNumber.from(40000)
            newTakerFee = BigNumber.from(10000)
            makerFee = await clearingHouse.makerFee()
            referralShare = await clearingHouse.referralShare()
            tradingFeeDiscount = await clearingHouse.tradingFeeDiscount()
            liquidationPenalty = await clearingHouse.liquidationPenalty()
            tx = await clearingHouse.connect(governance).setParams(
                newMaintenanceMargin,
                newMinAllowableMargin,
                newTakerFee,
                makerFee,
                referralShare,
                tradingFeeDiscount,
                liquidationPenalty
            )
            await tx.wait()

            response = await makehttpCall(method, params)
            result = response.body.result

            expect(result.maintenance_margin).to.equal(newMaintenanceMargin.toNumber())
            expect(result.min_allowable_margin).to.equal(newMinAllowableMargin.toNumber())
            expect(result.taker_fee).to.equal(newTakerFee.toNumber())

            // revert config
            tx = await clearingHouse.connect(governance).setParams(
                actualMaintenanceMargin,
                actualMinAllowableMargin,
                actualTakerFee,
                makerFee,
                referralShare,
                tradingFeeDiscount,
                liquidationPenalty
            )
            await tx.wait()
        })
    })

    context("Margin account contract variables", function () {
        // vars read from slot
        // margin, reservedMargin
        it("should read the correct value from contracts", async function () {
            // zero balance
            method ="testing_getMarginAccountVars"
            params =[ 0, charlie.address ]
            response = await makehttpCall(method, params)
            expect(response.body.result.margin).to.equal(0)
            expect(response.body.result.reserved_margin).to.equal(0)

            // add balance for order and then place
            longOrder = getOrderV2(0, charlie.address, multiplySize(0.1), multiplyPrice(1800), BigNumber.from(Date.now()), false)
            requiredMargin = await getRequiredMarginForLongOrder(longOrder)
            await addMargin(charlie, requiredMargin)
            await placeOrderFromLimitOrderV2(longOrder, charlie)

            method ="testing_getMarginAccountVars"
            params =[ 0, charlie.address ]
            response = await makehttpCall(method, params)

            //cleanup
            await cancelOrderFromLimitOrderV2(longOrder, charlie)
            await removeAllAvailableMargin(charlie)

            expect(response.body.result.margin).to.equal(requiredMargin.toNumber())
            expect(response.body.result.reserved_margin).to.equal(requiredMargin.toNumber())
        })
    })

    context.only("AMM contract variables", function () {
        // vars read from slot
        // positions, cumulativePremiumFraction, maxOracleSpreadRatio, maxLiquidationRatio, minSizeRequirement, oracle, underlyingAsset, 
        // maxLiquidationPriceSpread, redStoneAdapter, redStoneFeedId, impactMarginNotional, lastTradePrice, bids, asks, bidsHead, asksHead
        let ammIndex = 0
        it.skip("should read the correct value of variables from contracts which have default value after setup", async function () {
            // maxOracleSpreadRatio, maxLiquidationRatio, minSizeRequirement, oracle, underlyingAsset, maxLiquidationPriceSpread
            amms = await clearingHouse.getAMMs()
            ammAddress = amms[ammIndex]
            method ="testing_getAMMVars"
            params =[ ammAddress, ammIndex, charlie.address ]
            response = await makehttpCall(method, params)

            amm = new ethers.Contract(ammAddress, require('../abi/AMM.json'), provider)
            actualMaxOracleSpreadRatio = await amm.maxOracleSpreadRatio()
            actualOracleAddress = await amm.oracle()
            actualMaxLiquidationRatio = await amm.maxLiquidationRatio()
            actualMinSizeRequirement = await amm.minSizeRequirement()
            actualUnderlyingAssetAddress = await amm.underlyingAsset()
            actualMaxLiquidationPriceSpread = await amm.maxLiquidationPriceSpread()

            result = response.body.result
            expect(result.max_oracle_spread_ratio).to.equal(actualMaxOracleSpreadRatio.toNumber())
            expect(result.oracle_address.toLowerCase()).to.equal(actualOracleAddress.toString().toLowerCase())
            expect(result.max_liquidation_ratio).to.equal(actualMaxLiquidationRatio.toNumber())
            expect(String(result.min_size_requirement)).to.equal(actualMinSizeRequirement.toString())
            expect(result.underlying_asset_address.toLowerCase()).to.equal(actualUnderlyingAssetAddress.toString().toLowerCase())
            expect(result.max_liquidation_price_spread).to.equal(actualMaxLiquidationPriceSpread.toNumber())
        })
        context.only("should read the correct value of variables from contracts which dont have default value after setup", async function () {
            // positions, cumulativePremiumFraction, redStoneAdapter, redStoneFeedId, impactMarginNotional, lastTradePrice, bids, asks, bidsHead, asksHead
            it("variables which need set config before reading", async function () {
                // redStoneAdapter, redStoneFeedId, impactMarginNotional
                // impactMarginNotional
                amm = await getAMMContract(ammIndex)
                oracleAddress = await amm.oracle()
                redStoneAdapterAddress = await amm.redStoneAdapter()
                redStoneFeedId = await amm.redStoneFeedId()
                impactMarginNotional = await amm.impactMarginNotional()

                newOracleAddress = alice.address 
                newRedStoneAdapterAddress = bob.address 
                newRedStoneFeedId = ethers.utils.formatBytes32String("redStoneFeedId") 
                tx = await amm.connect(governance).setOracleConfig(newOracleAddress, newRedStoneAdapterAddress, newRedStoneFeedId)
                await tx.wait()

                newImpactMarginNotional = BigNumber.from(100000)
                tx = await amm.connect(governance).setImpactMarginNotional(newImpactMarginNotional)
                await tx.wait()

                amms = await clearingHouse.getAMMs()
                ammAddress = amms[ammIndex]
                method ="testing_getAMMVars"
                params =[ ammAddress, ammIndex, charlie.address ]
                response = await makehttpCall(method, params)
                result = response.body.result

                expect(result.oracle_address.toLowerCase()).to.equal(newOracleAddress.toLowerCase())
                expect(result.red_stone_adapter_address.toLowerCase()).to.equal(newRedStoneAdapterAddress.toLowerCase())
                expect(result.red_stone_feed_id).to.equal(newRedStoneFeedId)
                expect(result.impact_margin_notional).to.equal(newImpactMarginNotional.toNumber())

                // revert config
                await amm.connect(governance).setOracleConfig(oracleAddress, redStoneAdapterAddress, redStoneFeedId)
                await amm.connect(governance).setImpactMarginNotional(impactMarginNotional)
            })
            it("variables which need place order before reading", async function () {
                //bids, asks, bidsHead, asksHead
                await removeAllAvailableMargin(alice)
                await removeAllAvailableMargin(bob)
                longOrderBaseAssetQuantity = multiplySize(0.1) // 0.1 ether
                shortOrderBaseAssetQuantity = multiplySize(-0.1) // 0.1 ether
                longOrderPrice = multiplyPrice(1799)
                shortOrderPrice = multiplyPrice(1801)
                longOrder = getOrderV2(ammIndex, alice.address, longOrderBaseAssetQuantity, longOrderPrice, BigNumber.from(Date.now()), false)
                requiredMarginAlice = await getRequiredMarginForLongOrder(longOrder)
                await addMargin(alice, requiredMarginAlice)
                await placeOrderFromLimitOrderV2(longOrder, alice)

                shortOrder = getOrderV2(ammIndex, bob.address, shortOrderBaseAssetQuantity, shortOrderPrice, BigNumber.from(Date.now()), false)
                requiredMarginBob = await getRequiredMarginForShortOrder(shortOrder)
                await addMargin(bob, requiredMarginBob)
                await placeOrderFromLimitOrderV2(shortOrder, bob)

                amms = await clearingHouse.getAMMs()
                ammAddress = amms[ammIndex]
                method ="testing_getAMMVars"
                params =[ ammAddress, ammIndex, alice.address ]
                response = await makehttpCall(method, params)
                result = response.body.result

                //cleanup
                await cancelOrderFromLimitOrderV2(longOrder, alice)
                await cancelOrderFromLimitOrderV2(shortOrder, bob)
                await removeAllAvailableMargin(alice)
                await removeAllAvailableMargin(bob)

                expect(result.asks_head).to.equal(shortOrderPrice.toNumber())
                expect(result.bids_head).to.equal(longOrderPrice.toNumber())
                expect(String(result.bids_head_size)).to.equal(longOrderBaseAssetQuantity.toString())
                expect(String(result.asks_head_size)).to.equal(shortOrderBaseAssetQuantity.abs().toString())
            })
            it.only("variable which need position before reading", async function () {
                longOrderBaseAssetQuantity = multiplySize(0.1) // 0.1 ether
                shortOrderBaseAssetQuantity = multiplySize(-0.1) // 0.1 ether
                orderPrice = multiplyPrice(1800)
                longOrder = getOrderV2(ammIndex, alice.address, longOrderBaseAssetQuantity, orderPrice, BigNumber.from(Date.now()), false)
                requiredMarginAlice = await getRequiredMarginForLongOrder(longOrder)
                await addMargin(alice, requiredMarginAlice)
                await placeOrderFromLimitOrderV2(longOrder, alice)

                shortOrder = getOrderV2(ammIndex, bob.address, shortOrderBaseAssetQuantity, orderPrice, BigNumber.from(Date.now()), false)
                requiredMarginBob = await getRequiredMarginForShortOrder(shortOrder)
                await addMargin(bob, requiredMarginBob)
                await placeOrderFromLimitOrderV2(shortOrder, bob)

                amms = await clearingHouse.getAMMs()
                ammAddress = amms[ammIndex]
                method ="testing_getAMMVars"

                params =[ ammAddress, ammIndex, alice.address ]
                resultAlice = (await makehttpCall(method, params)).body.result
                params =[ ammAddress, ammIndex, bob.address ]
                resultBob = (await makehttpCall(method, params)).body.result

                //cleanup
                oppositeLongOrder = getOrderV2(ammIndex, bob.address, longOrderBaseAssetQuantity, orderPrice, BigNumber.from(Date.now()), true)
                await placeOrderFromLimitOrderV2(oppositeLongOrder, bob)
                oppositeShortOrder = getOrderV2(ammIndex, alice.address, shortOrderBaseAssetQuantity, orderPrice, BigNumber.from(Date.now()), true)
                await placeOrderFromLimitOrderV2(oppositeShortOrder, alice)
                await utils.waitForOrdersToMatch()
                await removeAllAvailableMargin(alice)
                await removeAllAvailableMargin(bob)

                expect(String(resultAlice.position.size)).to.equal(longOrderBaseAssetQuantity.toString())
                expect(String(resultAlice.position.open_notional)).to.equal(longOrderBaseAssetQuantity.mul(orderPrice).div(_1e18).toString())
                expect(String(resultBob.position.size)).to.equal(shortOrderBaseAssetQuantity.toString())
                expect(String(resultBob.position.open_notional)).to.equal(shortOrderBaseAssetQuantity.mul(orderPrice).div(_1e18).toString())
                expect(result.last_price).to.equal(orderPrice.toNumber())
            })
        })
    })

    context("IOC order contract variables", function () {
        it("should read the correct value from contracts", async function () {
            let charlieBalance = _1e6.mul(150)
            await addMargin(charlie, charlieBalance)

            longOrderBaseAssetQuantity = multiplySize(0.1) // 0.1 ether
            orderPrice = multiplyPrice(1800)
            salt = BigNumber.from(Date.now())
            market = BigNumber.from(0)

            latestBlockNumber = await provider.getBlockNumber()
            lastTimestamp = (await provider.getBlock(latestBlockNumber)).timestamp
            expireAt = lastTimestamp + 6
            IOCOrder = getIOCOrder(expireAt, market, charlie.address, longOrderBaseAssetQuantity, orderPrice, salt, false)
            orderHash = await ioc.getOrderHash(IOCOrder)
            params = [ orderHash ]
            method ="testing_getIOCOrdersVars"

            // before placing order
            result = (await makehttpCall(method, params)).body.result

            actualExpirationCap = await ioc.expirationCap()
            expectedExpirationCap = result.ioc_expiration_cap

            expect(expectedExpirationCap).to.equal(actualExpirationCap.toNumber())
            expect(result.order_details.block_placed).to.eq(0)
            expect(result.order_details.filled_amount).to.eq(0)
            expect(result.order_details.order_status).to.eq(0)

            //placing order
            txDetails = await placeIOCOrder(IOCOrder, charlie) 
            result = (await makehttpCall(method, params)).body.result

            actualBlockPlaced = txDetails.txReceipt.blockNumber
            expect(result.order_details.block_placed).to.eq(actualBlockPlaced)
            expect(result.order_details.filled_amount).to.eq(0)
            expect(result.order_details.order_status).to.eq(1)

            //cleanup
            await removeAllAvailableMargin(charlie)
        })
    })
    context("order book contract variables", function () {
        it("should read the correct value from contracts", async function () {
            let charlieBalance = _1e6.mul(150)
            await addMargin(charlie, charlieBalance)

            longOrderBaseAssetQuantity = multiplySize(0.1) // 0.1 ether
            orderPrice = multiplyPrice(1800)
            salt = BigNumber.from(Date.now())
            market = BigNumber.from(0)

            latestBlockNumber = await provider.getBlockNumber()
            lastTimestamp = (await provider.getBlock(latestBlockNumber)).timestamp
            expireAt = lastTimestamp + 6
            order = getOrder(market, charlie.address, longOrderBaseAssetQuantity, orderPrice, salt, false)
            orderHash = await orderBook.getOrderHash(order)
            params=[charlie.address, alice.address, orderHash]

            // before placing order
            result = (await makehttpCall(method, params)).body.result

            actualResult = await orderBook.isTradingAuthority(charlie.address, alice.address)
            expect(result.is_trading_authority).to.equal(actualResult)

            expect(result.order_details.block_placed).to.eq(0)
            expect(result.order_details.filled_amount).to.eq(0)
            expect(result.order_details.order_status).to.eq(0)

            //placing order
            txDetails = await placeOrderFromLimitOrder(order, charlie)
            result = (await makehttpCall(method, params)).body.result
            // cleanup
            await cancelOrderFromLimitOrder(order, charlie)
            await removeAllAvailableMargin(charlie)

            actualBlockPlaced = txDetails.txReceipt.blockNumber
            expect(result.order_details.block_placed).to.eq(actualBlockPlaced)
            expect(result.order_details.filled_amount).to.eq(0)
            expect(result.order_details.order_status).to.eq(1)

        })
    })
})

async function makehttpCall(method, params=[]) {
    body = {
        "jsonrpc":"2.0",
        "id" :1,
        "method" : method,
        "params" : params
    }

    const serverUrl = url.split("/").slice(0, 3).join("/")
    path = "/".concat(url.split("/").slice(3).join("/"))
    return chai.request(serverUrl)
        .post(path)
        .send(body)
}
