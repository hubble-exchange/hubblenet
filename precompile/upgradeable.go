// (c) 2022 Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package precompile

import "github.com/ava-labs/subnet-evm/utils"

// Upgrade contains the timestamp for the upgrade along with
// a boolean [Disable]. If [Disable] is set, the upgrade deactivates
// the precompile and clears its storage.
type Upgrade struct {
	BlockTimestamp *uint64 `json:"blockTimestamp"`
	Disable        bool    `json:"disable,omitempty"`
}

// Timestamp returns the timestamp this network upgrade goes into effect.
func (u *Upgrade) Timestamp() *uint64 {
	return u.BlockTimestamp
}

// IsDisabled returns true if the network upgrade deactivates the precompile.
func (c *UpgradeableConfig) IsDisabled() bool {
	return c.Disable
}

// Equal returns true iff [other] has the same blockTimestamp and has the
// same on value for the Disable flag.
func (c *UpgradeableConfig) Equal(other *UpgradeableConfig) bool {
	if other == nil {
		return false
	}
	return u.Disable == other.Disable && utils.Uint64PtrEqual(u.BlockTimestamp, other.BlockTimestamp)
}
