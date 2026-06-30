package app

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	v1_1 "quantum-chain/app/upgrades/v1_1"
)

// setupUpgradeHandlers registers an upgrade handler for every named daqq
// software upgrade. Cosmovisor swaps in the new binary at the planned height;
// the handler then runs module migrations so on-chain state matches the new
// binary before the chain resumes. Called from New before app.Load.
func (app *App) setupUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		v1_1.UpgradeName,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			// v1-1 is a deliberately no-op protocol bump: RunMigrations is an
			// identity here, and establishes the pattern real future protocol
			// versions will follow.
			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		},
	)
}

// setupStoreLoaders configures baseapp's store loader from the upgrade-info
// file Cosmovisor leaves on disk. When the node restarts at an upgrade height,
// this applies that upgrade's StoreUpgrades (added/renamed/deleted module
// stores) before state is loaded. On a normal start no file exists, so the
// plan is empty and the default loader is kept. Called from New before app.Load.
func (app *App) setupStoreLoaders() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}
	if upgradeInfo.Name == "" || app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	switch upgradeInfo.Name {
	case v1_1.UpgradeName:
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &v1_1.StoreUpgrades))
	}
}
