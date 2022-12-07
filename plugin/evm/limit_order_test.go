package evm

import (
	"testing"

	"gotest.tools/assert"
)

func TestSetOrderBookContractFileLocation(t *testing.T) {
	newFileLocation := "new/location"
	SetOrderBookContractFileLocation(newFileLocation)
	assert.Equal(t, newFileLocation, orderBookContractFileLocation)
}
