const { BigNumber } = require("ethers");
const { expect } = require("chai");

const utils = require("../utils")

const {
    alice,
    hubblebibliophile,
    placeOrder,
    addMargin,
    removeAllAvailableMargin,
    multiplyPrice,
    multiplySize,
    sleep,
} = utils

// Testing hubblebibliophile precompile contract 
describe("Testing getPositionSizes", function () {
    market = BigNumber.from(0)

    context("when there are no open positions", async function () {
        context("when trader never opened a position", async function () {
            it("should return 0", async function () {
                let bobMargin = multiplyPrice(15000)
                await addMargin(bob, bobMargin)
                
                alicePositionSize = await hubblebibliophile.getPositionSizes(alice.address)
                bobPositionSize = await hubblebibliophile.getPositionSizes(bob.address)

                // cleanup
                removeAllAvailableMargin(bob)

                console.log("alicePositionSize", alicePositionSize[0].toString(), "bobPositionSize", bobPositionSize[0].toString())
                expect(alicePositionSize[0].toString()).to.equal("0")
                expect(bobPositionSize[0].toString()).to.equal("0")
            })
        });

        context("when trader opened and closed a position", async function () {
            it("should return 0", async function () {
                let bobOrderPrice = multiplyPrice(1800)
                let aliceOrderPrice = multiplyPrice(1800)
                let aliceMargin = multiplyPrice(15000)
                let bobMargin = multiplyPrice(15000)
                await addMargin(alice, aliceMargin)
                await addMargin(bob, bobMargin)

                aliceOrder1Size = multiplySize(0.3)
                bobOrder1Size = multiplySize(-0.3)
                await placeOrder(market, alice, aliceOrder1Size, aliceOrderPrice)
                await placeOrder(market, bob, bobOrder1Size, bobOrderPrice)
                await sleep(10)

                alicePositionSize = await hubblebibliophile.getPositionSizes(alice.address)
                expect(alicePositionSize[0].toString()).to.equal((aliceOrder1Size).toString())
                bobPositionSize = await hubblebibliophile.getPositionSizes(bob.address)
                expect(bobPositionSize[0].toString()).to.equal((bobOrder1Size).toString())

                aliceOrder2Size = multiplySize(-0.3)
                bobOrder2Size = multiplySize(0.3)
                await placeOrder(market, alice, aliceOrder2Size, aliceOrderPrice)
                await placeOrder(market, bob, bobOrder2Size, bobOrderPrice)
                await sleep(10)

                alicePositionSizeFinal = await hubblebibliophile.getPositionSizes(alice.address)
                expect(alicePositionSizeFinal[0].toString()).to.equal(aliceOrder1Size.add(aliceOrder2Size).toString())
                bobPositionSizeFinal = await hubblebibliophile.getPositionSizes(bob.address)
                expect(bobPositionSizeFinal[0].toString()).to.equal(bobOrder1Size.add(bobOrder2Size).toString())

                // cleanup
                await removeAllAvailableMargin(alice)
                await removeAllAvailableMargin(bob)
            });
        });
    })

    context("when there are open positions", async function () {
        it("it returns positionSize by aggregating the positions", async function () {
            let bobOrderPrice = multiplyPrice(1800)
            let aliceOrderPrice = multiplyPrice(1800)
            let aliceMargin = multiplyPrice(15000)
            let bobMargin = multiplyPrice(15000)
            await addMargin(alice, aliceMargin)
            await addMargin(bob, bobMargin)

            aliceOrder1Size = multiplySize(0.1)
            bobOrder1Size = multiplySize(-0.1)
            await placeOrder(market, alice, aliceOrder1Size, aliceOrderPrice)
            await placeOrder(market, bob, bobOrder1Size, bobOrderPrice)
            await sleep(10)

            alicePositionSize = await hubblebibliophile.getPositionSizes(alice.address)
            expect(alicePositionSize[0].toString()).to.equal((aliceOrder1Size).toString())
            bobPositionSize = await hubblebibliophile.getPositionSizes(bob.address)
            expect(bobPositionSize[0].toString()).to.equal((bobOrder1Size).toString())

            aliceOrder2Size = multiplySize(0.2)
            bobOrder2Size = multiplySize(-0.2)
            await placeOrder(market, alice, aliceOrder2Size, aliceOrderPrice)
            await placeOrder(market, bob, bobOrder2Size, bobOrderPrice)
            await sleep(10)

            alicePositionSizeFinal = await hubblebibliophile.getPositionSizes(alice.address)
            bobPositionSizeFinal = await hubblebibliophile.getPositionSizes(bob.address)
            expect(alicePositionSizeFinal[0].toString()).to.equal(aliceOrder1Size.add(aliceOrder2Size).toString())
            expect(bobPositionSizeFinal[0].toString()).to.equal(bobOrder1Size.add(bobOrder2Size).toString())

            // cleanup
            totalAliceOrderSize = aliceOrder1Size.add(aliceOrder2Size)
            totalBobOrderSize = bobOrder1Size.add(bobOrder2Size)
            await addMargin(alice, aliceMargin)
            await addMargin(bob, bobMargin)
            await placeOrder(market, bob, totalAliceOrderSize, bobOrderPrice)
            await placeOrder(market, alice, totalBobOrderSize, aliceOrderPrice)
            await sleep(10)
            await removeAllAvailableMargin(alice)
            await removeAllAvailableMargin(bob)
        })
    });
});


