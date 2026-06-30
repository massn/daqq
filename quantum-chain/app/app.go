package app

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"sort"
	"strings"

	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/keeper"
	icahostkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"

	"quantum-chain/app/gui"
	"quantum-chain/docs"
	beaconmodulekeeper "quantum-chain/x/beacon/keeper"
	problemsmodulekeeper "quantum-chain/x/problems/keeper"
	quantumchainmodulekeeper "quantum-chain/x/quantumchain/keeper"
	randomcircuitmodulekeeper "quantum-chain/x/random_circuit/keeper"
	randomcircuittypes "quantum-chain/x/random_circuit/types"
)

const (
	// Name is the name of the application.
	Name = "quantum-chain"
	// AccountAddressPrefix is the prefix for accounts addresses.
	AccountAddressPrefix = "qc"
	// ChainCoinType is the coin type of the chain.
	ChainCoinType = 118
)

// DefaultNodeHome default home directories for the application daemon
var DefaultNodeHome string

var (
	_ runtime.AppI            = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
)

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*runtime.App
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	// keepers
	// only keepers required by the app are exposed
	// the list of all modules is available in the app_config
	AuthKeeper            authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             *govkeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	ConsensusParamsKeeper consensuskeeper.Keeper
	CircuitBreakerKeeper  circuitkeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper

	// ibc keepers
	IBCKeeper           *ibckeeper.Keeper
	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper
	TransferKeeper      ibctransferkeeper.Keeper

	// simulation manager
	sm                  *module.SimulationManager
	QuantumchainKeeper  quantumchainmodulekeeper.Keeper
	BeaconKeeper        beaconmodulekeeper.Keeper
	RandomCircuitKeeper randomcircuitmodulekeeper.Keeper
	ProblemsKeeper      problemsmodulekeeper.Keeper
}

func init() {
	var err error
	clienthelpers.EnvPrefix = Name
	DefaultNodeHome, err = clienthelpers.GetNodeHomeDirectory("." + Name)
	if err != nil {
		panic(err)
	}
}

// AppConfig returns the default app config.
func AppConfig() depinject.Config {
	return depinject.Configs(
		appConfig,
		depinject.Supply(
			// supply custom module basics
			map[string]module.AppModuleBasic{
				genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			},
		),
	)
}

// New returns a reference to an initialized App.
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	var (
		app        = &App{}
		appBuilder *runtime.AppBuilder

		// merge the AppConfig and other configuration in one config
		appConfig = depinject.Configs(
			AppConfig(),
			depinject.Supply(
				appOpts, // supply app options
				logger,  // supply logger

				// Supply with IBC keeper getter for the IBC modules with App Wiring.
				// The IBC Keeper cannot be passed because it has not been initiated yet.
				// Passing the getter, the app IBC Keeper will always be accessible.
				// This needs to be removed after IBC supports App Wiring.
				app.GetIBCKeeper,

				// here alternative options can be supplied to the DI container.
				// those options can be used f.e to override the default behavior of some modules.
				// for instance supplying a custom address codec for not using bech32 addresses.
				// read the depinject documentation and depinject module wiring for more information
				// on available options and how to use them.
			),
		)
	)

	var appModules map[string]appmodule.AppModule
	if err := depinject.Inject(appConfig,
		&appBuilder,
		&appModules,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		&app.AuthKeeper,
		&app.BankKeeper,
		&app.StakingKeeper,
		&app.SlashingKeeper,
		&app.MintKeeper,
		&app.DistrKeeper,
		&app.GovKeeper,
		&app.UpgradeKeeper,
		&app.AuthzKeeper,
		&app.ConsensusParamsKeeper,
		&app.CircuitBreakerKeeper,
		&app.ParamsKeeper,
		&app.QuantumchainKeeper,
		&app.BeaconKeeper,
		&app.RandomCircuitKeeper,
		&app.ProblemsKeeper,
	); err != nil {
		panic(err)
	}

	// add to default baseapp options
	// enable optimistic execution
	baseAppOptions = append(baseAppOptions, baseapp.SetOptimisticExecution())

	// build app
	app.App = appBuilder.Build(db, traceStore, baseAppOptions...)

	// register legacy modules
	if err := app.registerIBCModules(appOpts); err != nil {
		panic(err)
	}

	/****  Module Options ****/

	// create the simulation manager and define the order of the modules for deterministic simulations
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AuthKeeper, authsims.RandomGenesisAccounts, nil),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, overrideModules)

	app.sm.RegisterStoreDecoders()

	// A custom InitChainer sets if extra pre-init-genesis logic is required.
	// This is necessary for manually registered modules that do not support app wiring.
	// Manually set the module version map as shown below.
	// The upgrade module will automatically handle de-duplication of the module version map.
	app.SetInitChainer(func(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
		if err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap()); err != nil {
			return nil, err
		}
		return app.App.InitChainer(ctx, req)
	})

	// Register software-upgrade handlers so a governance-approved upgrade runs
	// module migrations when Cosmovisor swaps in the new binary at its height,
	// and apply any pending upgrade's store changes before state is loaded.
	app.setupUpgradeHandlers()
	app.setupStoreLoaders()

	if err := app.Load(loadLatest); err != nil {
		panic(err)
	}

	return app
}

// GetSubspace returns a param subspace for a given module name.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// LegacyAmino returns App's amino codec.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns App's app codec.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns App's InterfaceRegistry.
func (app *App) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

// TxConfig returns App's TxConfig
func (app *App) TxConfig() client.TxConfig {
	return app.txConfig
}

// GetKey returns the KVStoreKey for the provided store key.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	kvStoreKey, ok := app.UnsafeFindStoreKey(storeKey).(*storetypes.KVStoreKey)
	if !ok {
		return nil
	}
	return kvStoreKey
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	app.App.RegisterAPIRoutes(apiSvr, apiConfig)
	// register swagger API in app.go so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}

	// register app's OpenAPI routes.
	docs.RegisterOpenAPIService(Name, apiSvr.Router)

	// register the embedded web visualizer (served same-origin from the API
	// server, so it needs no CORS configuration).
	gui.RegisterGUIService(apiSvr.ClientCtx, app.guiSeeds, app.guiProblems, app.guiResults, apiSvr.Router)
}

// guiResults reads submitted random-circuit results grouped by round, for the
// embedded web visualizer to show per-round cross-validation. Each submission's
// distribution is summarized by a hash; a round "agrees" when every
// participant's hash is identical.
func (app *App) guiResults() ([]gui.ResultRound, error) {
	ctx, err := app.CreateQueryContext(0, false)
	if err != nil {
		return nil, err
	}

	iter, err := app.RandomCircuitKeeper.Results.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	byRound := map[uint64][]gui.ResultSubmission{}
	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return nil, err
		}
		roundID := kv.Key.K1()
		rd := kv.Value
		byRound[roundID] = append(byRound[roundID], gui.ResultSubmission{
			Address:          rd.Address,
			BlockHeight:      rd.BlockHeight,
			NumStates:        len(rd.Distribution.Entries),
			DistributionHash: distributionHash(rd.Distribution),
		})
	}

	results := make([]gui.ResultRound, 0, len(byRound))
	for roundID, subs := range byRound {
		sort.Slice(subs, func(i, j int) bool { return subs[i].Address < subs[j].Address })
		agreement := true
		for _, s := range subs {
			if s.DistributionHash != subs[0].DistributionHash {
				agreement = false
				break
			}
		}
		results = append(results, gui.ResultRound{RoundID: roundID, Submissions: subs, Agreement: agreement})
	}
	return results, nil
}

// distributionHash returns a canonical SHA-256 (hex) of a distribution: entries
// are sorted by basis state and joined, so two participants that computed the
// same distribution hash to the same value regardless of submission order.
func distributionHash(d randomcircuittypes.Distribution) string {
	entries := make([]randomcircuittypes.ProbabilityEntry, len(d.Entries))
	copy(entries, d.Entries)
	sort.Slice(entries, func(i, j int) bool { return entries[i].State < entries[j].State })
	var b strings.Builder
	for _, e := range entries {
		b.WriteString(e.State)
		b.WriteByte('=')
		b.WriteString(e.Probability)
		b.WriteByte(';')
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

// guiSeeds reads every finalized network-shared random seed from the beacon
// module, for the embedded web visualizer.
func (app *App) guiSeeds() ([]gui.SeedEntry, error) {
	ctx, err := app.CreateQueryContext(0, false)
	if err != nil {
		return nil, err
	}

	iter, err := app.BeaconKeeper.Seeds.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var seeds []gui.SeedEntry
	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return nil, err
		}
		seeds = append(seeds, gui.SeedEntry{RoundID: kv.Key, Seed: kv.Value})
	}
	return seeds, nil
}

// guiProblems reads the registered problems from the problems module registry,
// for the embedded web visualizer to show what the network is solving.
func (app *App) guiProblems() ([]gui.ProblemEntry, error) {
	ctx, err := app.CreateQueryContext(0, false)
	if err != nil {
		return nil, err
	}

	iter, err := app.ProblemsKeeper.Problems.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var problems []gui.ProblemEntry
	for ; iter.Valid(); iter.Next() {
		p, err := iter.Value()
		if err != nil {
			return nil, err
		}
		problems = append(problems, gui.ProblemEntry{
			ID:          p.Id,
			Name:        p.Name,
			Enabled:     p.Enabled,
			Description: p.Description,
		})
	}
	return problems, nil
}

// GetMaccPerms returns a copy of the module account permissions
//
// NOTE: This is solely to be used for testing purposes.
func GetMaccPerms() map[string][]string {
	dup := make(map[string][]string)
	for _, perms := range moduleAccPerms {
		dup[perms.GetAccount()] = perms.GetPermissions()
	}

	return dup
}

// BlockedAddresses returns all the app's blocked account addresses.
func BlockedAddresses() map[string]bool {
	result := make(map[string]bool)

	if len(blockAccAddrs) > 0 {
		for _, addr := range blockAccAddrs {
			result[addr] = true
		}
	} else {
		for addr := range GetMaccPerms() {
			result[addr] = true
		}
	}

	return result
}
