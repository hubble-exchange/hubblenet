package bibliophile

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/contract"
	"github.com/ethereum/go-ethereum/common"
)

type BibliophileClient interface {
	GetSize(market common.Address, trader *common.Address) *big.Int
	GetMarketAddressFromMarketID(marketID int64) common.Address
	DetermineFillPrice(marketId int64, fillAmount *big.Int, longOrderPrice, shortOrderPrice, blockPlaced0, blockPlaced1 *big.Int) (*ValidateOrdersAndDetermineFillPriceOutput, error)

	GetBlockPlaced(orderHash [32]byte) *big.Int
	GetOrderFilledAmount(orderHash [32]byte) *big.Int
	GetOrderStatus(orderHash [32]byte) int64

	IOC_GetBlockPlaced(orderHash [32]byte) *big.Int
	IOC_GetOrderFilledAmount(orderHash [32]byte) *big.Int
	IOC_GetOrderStatus(orderHash [32]byte) int64

	GetAccessibleState() contract.AccessibleState
}

// Define a structure that will implement the Bibliophile interface
type bibliophileClient struct {
	accessibleState contract.AccessibleState
}

func NewBibliophileClient(accessibleState contract.AccessibleState) BibliophileClient {
	return &bibliophileClient{
		accessibleState: accessibleState,
	}
}

func (b *bibliophileClient) GetAccessibleState() contract.AccessibleState {
	return b.accessibleState
}
func (b *bibliophileClient) GetSize(market common.Address, trader *common.Address) *big.Int {
	return GetSize(b.accessibleState.GetStateDB(), market, trader)
}

func (b *bibliophileClient) GetMarketAddressFromMarketID(marketID int64) common.Address {
	return GetMarketAddressFromMarketID(marketID, b.accessibleState.GetStateDB())
}

func (b *bibliophileClient) DetermineFillPrice(marketId int64, fillAmount *big.Int, longOrderPrice, shortOrderPrice, blockPlaced0, blockPlaced1 *big.Int) (*ValidateOrdersAndDetermineFillPriceOutput, error) {
	return DetermineFillPrice(b.accessibleState.GetStateDB(), marketId, fillAmount, longOrderPrice, shortOrderPrice, blockPlaced0, blockPlaced1)
}

func (b *bibliophileClient) GetBlockPlaced(orderHash [32]byte) *big.Int {
	return GetBlockPlaced(b.accessibleState.GetStateDB(), orderHash)
}

func (b *bibliophileClient) GetOrderFilledAmount(orderHash [32]byte) *big.Int {
	return GetOrderFilledAmount(b.accessibleState.GetStateDB(), orderHash)
}

func (b *bibliophileClient) GetOrderStatus(orderHash [32]byte) int64 {
	return GetOrderStatus(b.accessibleState.GetStateDB(), orderHash)
}

func (b *bibliophileClient) IOC_GetBlockPlaced(orderHash [32]byte) *big.Int {
	return IOC_GetBlockPlaced(b.accessibleState.GetStateDB(), orderHash)
}

func (b *bibliophileClient) IOC_GetOrderFilledAmount(orderHash [32]byte) *big.Int {
	return IOC_GetOrderFilledAmount(b.accessibleState.GetStateDB(), orderHash)
}

func (b *bibliophileClient) IOC_GetOrderStatus(orderHash [32]byte) int64 {
	return IOC_GetOrderStatus(b.accessibleState.GetStateDB(), orderHash)
}
