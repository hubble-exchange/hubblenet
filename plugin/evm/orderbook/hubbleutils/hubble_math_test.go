package hubbleutils

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestECRecovers(t *testing.T) {
	// 1. Test case from
	orderHash := "0xc03560e0135777a3e7f155dbe7edbda36b217d6d9817f61ac914ed7f1029b387"
	address, err := ECRecover(common.FromHex(orderHash), common.FromHex("0x4f47baaf7e2c447a3eaddb49e50908cea40811841077b2d0c53b7b496384c2ea783d1e53321e23b6f033750c46401b9e0705bde1b319bb461b55c752b9e980cc1c"))
	assert.Nil(t, err)
	assert.Equal(t, "0x70997970C51812dc3A010C7d01b50e0d17dc79C8", address.String())
}
