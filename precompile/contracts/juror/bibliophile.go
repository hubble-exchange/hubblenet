package juror

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/contract"
	"github.com/ava-labs/subnet-evm/precompile/contracts/bibliophile"
	"github.com/ethereum/go-ethereum/common"
)

type Bibliophile interface {
	GetSize(market common.Address, trader *common.Address) *big.Int
	GetMarketAddressFromMarketID(marketID int64) common.Address
	DetermineFillPrice(marketId int64, fillAmount *big.Int, longOrderPrice, shortOrderPrice, blockPlaced0, blockPlaced1 *big.Int) (*bibliophile.ValidateOrdersAndDetermineFillPriceOutput, error)
	GetBlockPlaced(orderHash [32]byte) *big.Int
	GetOrderFilledAmount(orderHash [32]byte) *big.Int
	GetOrderStatus(orderHash [32]byte) int64
}

// Define a structure that will implement the Bibliophile interface
type BibliophileImpl struct {
	stateDB contract.StateDB
}

func NewBibliophile(stateDB contract.StateDB) Bibliophile {
	return &BibliophileImpl{
		stateDB: stateDB,
	}
}

// Implement the methods of the Bibliophile interface in bibliophile
func (b *BibliophileImpl) GetSize(market common.Address, trader *common.Address) *big.Int {
	return bibliophile.GetSize(b.stateDB, market, trader)
}

func (b *BibliophileImpl) GetMarketAddressFromMarketID(marketID int64) common.Address {
	return bibliophile.GetMarketAddressFromMarketID(marketID, b.stateDB)
}

func (b *BibliophileImpl) DetermineFillPrice(marketId int64, fillAmount *big.Int, longOrderPrice, shortOrderPrice, blockPlaced0, blockPlaced1 *big.Int) (*bibliophile.ValidateOrdersAndDetermineFillPriceOutput, error) {
	return bibliophile.DetermineFillPrice(b.stateDB, marketId, fillAmount, longOrderPrice, shortOrderPrice, blockPlaced0, blockPlaced1)
}

func (b *BibliophileImpl) GetBlockPlaced(orderHash [32]byte) *big.Int {
	return bibliophile.GetBlockPlaced(b.stateDB, orderHash)
}

func (b *BibliophileImpl) GetOrderFilledAmount(orderHash [32]byte) *big.Int {
	return bibliophile.GetOrderFilledAmount(b.stateDB, orderHash)
}

func (b *BibliophileImpl) GetOrderStatus(orderHash [32]byte) int64 {
	return bibliophile.GetOrderStatus(b.stateDB, orderHash)
}

func fromTwosComplement(b []byte) *big.Int {
	t := new(big.Int).SetBytes(b)
	if b[0]&0x80 != 0 {
		t.Sub(t, new(big.Int).Lsh(big.NewInt(1), uint(len(b)*8)))
	}
	return t
}
