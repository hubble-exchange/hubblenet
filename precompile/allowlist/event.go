// Code generated
// This file is a generated precompile contract config with stubbed abstract functions.
// The file is generated by a template. Please inspect every code and comment in this file before use.

package allowlist

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/contract"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// AllowListEventGasCost is the gas cost of a call to the AllowList contract's event.
	// It is the base gas cost + the gas cost of the topics (signature, role, account, caller)
	// and the gas cost of the non-indexed data (oldRole).
	AllowListEventGasCost = contract.LogGas + contract.LogTopicGas*4 + contract.LogDataGas*common.HashLength
)

// PackRoleSetEvent packs the event into the appropriate arguments for RoleSet.
// It returns topic hashes and the encoded non-indexed data.
func PackRoleSetEvent(role Role, account common.Address, caller common.Address, oldRole Role) ([]common.Hash, []byte, error) {
	return AllowListABI.PackEvent("RoleSet", role.Big(), account, caller, oldRole.Big())
}

// UnpackRoleSetEventData attempts to unpack non-indexed [dataBytes].
func UnpackRoleSetEventData(dataBytes []byte) (Role, error) {
	eventData := struct {
		OldRole *big.Int
	}{}
	err := AllowListABI.UnpackIntoInterface(&eventData, "RoleSet", dataBytes)
	if err != nil {
		return Role{}, err
	}
	return FromBig(eventData.OldRole)
}