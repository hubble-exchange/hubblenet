const { ethers, BigNumber } = require("ethers");
const { expect } = require("chai");
const utils = require("../utils")

const {
    alice,
    clearingHouse,
    hubblebibliophile,
    getOrder,
    multiplySize,
    multiplyPrice,
    _1e6,
    provider,
} = utils

// Testing hubblebibliophile precompile contract 
describe('Testing validateLiquidationOrderAndDetermineFillPrice',async function () {
    market = 0
    salt = BigNumber.from(101)

    context('When liquidation amount is not multiple of minSizeRequirement', async function () {
        it.skip('returns error if liquidationAmount is zero', async function () {
            const ammAddress = await clearingHouse.amms(market)
            const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
            minSizeRequirement = await amm.minSizeRequirement()
            liquidationAmount = BigNumber.from(0) 

            orderPrice = multiplyPrice(1800)
            longOrderBaseAssetQuantity = multiplySize(0.1) // 0.1 ether
            shortOrderBaseAssetQuantity = multiplySize(-0.1) // short 0.1 ether
            longOrder = getOrder(BigNumber.from(market), alice.address, longOrderBaseAssetQuantity, orderPrice, salt, false)
            shortOrder = getOrder(BigNumber.from(market), alice.address, shortOrderBaseAssetQuantity, orderPrice, salt, false)

            // try long order
            try {
                await hubblebibliophile.validateLiquidationOrderAndDetermineFillPrice(longOrder, liquidationAmount)
            } catch (error) {
                expect(error.error.body).to.match(/OB.not_multiple/)
            }

            // try short order
            try {
                await hubblebibliophile.validateLiquidationOrderAndDetermineFillPrice(shortOrder, liquidationAmount)
            } catch (error) {
                expect(error.error.body).to.match(/OB.not_multiple/)
                return
            }
            expect.fail('Expected throw not received');
        })

        it('returns error if liquidationAmount is greater than zero less than minSizeRequirement', async function () {
            const ammAddress = await clearingHouse.amms(market)
            const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
            minSizeRequirement = await amm.minSizeRequirement()
            liquidationAmount = minSizeRequirement.div(BigNumber.from(2))

            orderPrice = multiplyPrice(1800)
            longOrderBaseAssetQuantity = multiplySize(0.1) // 0.1 ether
            shortOrderBaseAssetQuantity = multiplySize(-0.1) // short 0.1 ether
            longOrder = getOrder(BigNumber.from(market), alice.address, longOrderBaseAssetQuantity, orderPrice, salt, false)
            shortOrder = getOrder(BigNumber.from(market), alice.address, shortOrderBaseAssetQuantity, orderPrice, salt, false)

            // try long order
            try {
                await hubblebibliophile.validateLiquidationOrderAndDetermineFillPrice(longOrder, liquidationAmount)
            } catch (error) {
                expect(error.error.body).to.match(/OB.not_multiple/)
            }

            // try short order
            try {
                await hubblebibliophile.validateLiquidationOrderAndDetermineFillPrice(shortOrder, liquidationAmount)
            } catch (error) {
                expect(error.error.body).to.match(/OB.not_multiple/)
                return
            }
            expect.fail('Expected throw not received');
        })

        it('returns error if liquidationAmount is greater than minSizeRequirement but not a multiple', async function () {
            const ammAddress = await clearingHouse.amms(market)
            const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
            minSizeRequirement = await amm.minSizeRequirement()
            liquidationAmount = minSizeRequirement.mul(BigNumber.from(3)).div(BigNumber.from(2))

            console.log("minSizeRequirement", minSizeRequirement.toString(), "liquidationAmount", liquidationAmount.toString())

            orderPrice = multiplyPrice(1800)
            longOrderBaseAssetQuantity = multiplySize(0.1) // 0.1 ether
            shortOrderBaseAssetQuantity = multiplySize(-0.1) // short 0.1 ether
            longOrder = getOrder(BigNumber.from(market), alice.address, longOrderBaseAssetQuantity, orderPrice, salt, false)
            shortOrder = getOrder(BigNumber.from(market), alice.address, shortOrderBaseAssetQuantity, orderPrice, salt, false)

            // try long order
            try {
                response = await hubblebibliophile.validateLiquidationOrderAndDetermineFillPrice(longOrder, liquidationAmount)
                console.log("response for long order", response.toString())
            } catch (error) {
                expect(error.error.body).to.match(/OB.not_multiple/)
            }

            // try short order
            try {
                await hubblebibliophile.validateLiquidationOrderAndDetermineFillPrice(shortOrder, liquidationAmount)
            } catch (error) {
                expect(error.error.body).to.match(/OB.not_multiple/)
                return
            }
            expect.fail('Expected throw not received');
        })
    })

    context('When liquidationAmount is multiple of minSizeRequirement', async function () {
        context('For a long order', async function () {
            it('returns error if price is less than liquidation lower bound price', async function () {
                longOrderBaseAssetQuantity = multiplySize(0.3) // long 0.3 ether
                liquidationAmount = multiplySize(0.2) // 0.2 ether
                const ammAddress = await clearingHouse.amms(market)
                const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
                oraclePrice = (await amm.getUnderlyingPrice())
                maxLiquidationPriceSpread = await amm.maxLiquidationPriceSpread()
                // liqLowerBound = oraclePrice*(1e6 - liquidationPriceSpread)/1e6
                liqLowerBound = oraclePrice.mul(_1e6.sub(maxLiquidationPriceSpread)).div(_1e6)
                longOrderPrice = liqLowerBound.sub(1)

                longOrder = getOrder(BigNumber.from(market), alice.address, longOrderBaseAssetQuantity, longOrderPrice, salt, false)
                try {
                    await hubblebibliophile.validateLiquidationOrderAndDetermineFillPrice(longOrder, liquidationAmount)
                } catch (error) {
                    expect(error.error.body).to.match(/OB_long_order_price_too_low/)
                    return
                }
                expect.fail('Expected throw not received');
            })

            it('returns upperBound as fillPrice is price is more than upperBound', async function () {
                longOrderBaseAssetQuantity = multiplySize(0.3) // long 0.3 ether
                liquidationAmount = multiplySize(0.2) // 0.2 ether
                const ammAddress = await clearingHouse.amms(market)
                const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
                oraclePrice = (await amm.getUnderlyingPrice())
                oraclePriceSpreadThreshold = (await amm.maxOracleSpreadRatio())
                // upperBound = (oraclePrice*(1e6 + oraclePriceSpreadThreshold))/1e6
                upperBound = oraclePrice.mul(_1e6.add(oraclePriceSpreadThreshold)).div(_1e6)

                longOrderPrice1 = upperBound.add(BigNumber.from(1))
                longOrder1 = getOrder(BigNumber.from(market), alice.address, longOrderBaseAssetQuantity, longOrderPrice1, salt, false)
                responseLongOrder1 =  await hubblebibliophile.validateLiquidationOrderAndDetermineFillPrice(longOrder1, liquidationAmount)
                expect(responseLongOrder1.toString()).to.equal(upperBound.toString())

                longOrderPrice2 = upperBound.add(BigNumber.from(1000))
                longOrder2 = getOrder(BigNumber.from(market), alice.address, longOrderBaseAssetQuantity, longOrderPrice2, salt, false)
                responseLongOrder2 = await hubblebibliophile.validateLiquidationOrderAndDetermineFillPrice(longOrder2, liquidationAmount)
                expect(responseLongOrder2.toString()).to.equal(upperBound.toString())

                longOrderPrice3 = upperBound.add(BigNumber.from(_1e6))
                longOrder3 = getOrder(BigNumber.from(market), alice.address, longOrderBaseAssetQuantity, longOrderPrice3, salt, false)
                responseLongOrder3 = await hubblebibliophile.validateLiquidationOrderAndDetermineFillPrice(longOrder3, liquidationAmount)
                expect(responseLongOrder3.toString()).to.equal(upperBound.toString())
            })
        })

        context('Testing short order', async function () {
            it('returns lower bound as fillPrice if shortPrice is less than lowerBound', async function () {
                shortOrderBaseAssetQuantity = multiplySize(-0.4) // short 0.4 ether
                liquidationAmount = multiplySize(0.2) // 0.2 ether
                const ammAddress = await clearingHouse.amms(market)
                const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
                oraclePrice = (await amm.getUnderlyingPrice())
                oraclePriceSpreadThreshold = (await amm.maxOracleSpreadRatio())
                // upperBound = (oraclePrice*(1e6 - oraclePriceSpreadThreshold))/1e6
                lowerBound = oraclePrice.mul(_1e6.sub(oraclePriceSpreadThreshold)).div(_1e6)

                shortOrderPrice1 = lowerBound.sub(BigNumber.from(1))
                shortOrder1 = getOrder(BigNumber.from(market), alice.address, shortOrderBaseAssetQuantity, shortOrderPrice1, salt, false)
                responseShortOrder1 = await hubblebibliophile.validateLiquidationOrderAndDetermineFillPrice(shortOrder1, liquidationAmount)
                expect(responseShortOrder1.toString()).to.equal(lowerBound.toString())

                shortOrderPrice2 = lowerBound.sub(BigNumber.from(1000))
                shortOrder2 = getOrder(BigNumber.from(market), alice.address, shortOrderBaseAssetQuantity, shortOrderPrice2, salt, false)
                responseShortOrder2 = await hubblebibliophile.validateLiquidationOrderAndDetermineFillPrice(shortOrder2, liquidationAmount)
                expect(responseShortOrder2.toString()).to.equal(lowerBound.toString())

                shortOrderPrice3 = lowerBound.sub(BigNumber.from(_1e6))
                shortOrder3 = getOrder(BigNumber.from(market), alice.address, shortOrderBaseAssetQuantity, shortOrderPrice3, salt, false)
                responseShortOrder3 = await hubblebibliophile.validateLiquidationOrderAndDetermineFillPrice(shortOrder3, liquidationAmount)
                expect(responseShortOrder3.toString()).to.equal(lowerBound.toString())
            })

            it('returns error if price is more than liquidation upperBound', async function () {
                const ammAddress = await clearingHouse.amms(market)
                const amm = new ethers.Contract(ammAddress, require('../../abi/AMM.json'), provider);
                oraclePrice = (await amm.getUnderlyingPrice())
                maxLiquidationPriceSpread = (await amm.maxLiquidationPriceSpread())
                // liqUpperBound = oraclePrice*(1e6 + maxLiquidationPriceSpread))
                liqUpperBound = oraclePrice.mul(_1e6.add(maxLiquidationPriceSpread)).div(_1e6)
                shortOrderPrice = liqUpperBound.add(BigNumber.from(1))

                shortOrder = getOrder(BigNumber.from(market), alice.address, shortOrderBaseAssetQuantity, shortOrderPrice, salt, false)
                try {
                    await hubblebibliophile.validateLiquidationOrderAndDetermineFillPrice(shortOrder, liquidationAmount)
                } catch (error) {
                    expect(error.error.body).to.match(/OB_short_order_price_too_high/)
                    return
                }
                expect.fail('Expected throw not received');
            })
        })
    })
})