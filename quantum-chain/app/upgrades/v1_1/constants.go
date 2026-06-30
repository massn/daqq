// Package v1_1 defines the first daqq software upgrade. It is a deliberately
// no-op protocol bump used to exercise the Cosmovisor hot-swap path end to end
// (governance proposal -> swap at the upgrade height -> resume) before any real
// state-machine or proto v2 change ships.
package v1_1

import (
	storetypes "cosmossdk.io/store/types"
)

// UpgradeName is the on-chain name referenced both by the software-upgrade
// governance proposal and by the Cosmovisor `upgrades/<name>/` directory that
// holds the swapped-in binary. It must match in both places exactly.
const UpgradeName = "v1-1"

// StoreUpgrades lists the store key changes applied atomically at the upgrade
// height (added / renamed / deleted modules). This first upgrade is a no-op, so
// it changes no stores; later protocol versions that introduce a new module
// add its key to Added here.
var StoreUpgrades = storetypes.StoreUpgrades{}
