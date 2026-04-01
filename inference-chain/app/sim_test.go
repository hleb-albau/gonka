//go:build sims

package app_test

import (
	"encoding/json"
	"io"
	"testing"

	"cosmossdk.io/log"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simsx "github.com/cosmos/cosmos-sdk/testutil/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"

	"github.com/productscience/inference/app"
	inferencemodule "github.com/productscience/inference/x/inference/module"
	inferencetypes "github.com/productscience/inference/x/inference/types"
)

func init() {
	simcli.GetSimulatorFlags()
	app.InitSDKConfig()
	inferencemodule.IgnoreDuplicateDenomRegistration = true
}

// TestFullAppSimulation runs the full app simulation against all default seeds in parallel.
// Short smoke run: go test -tags sims -run TestFullAppSimulation -v -timeout 10m -NumBlocks=50 -BlockSize=20
// Full run:        go test -tags sims -run TestFullAppSimulation -v -timeout 60m
func TestFullAppSimulation(t *testing.T) {
	simsx.Run(t, newSimApp, setupStateFactory)
}

func newSimApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *app.App {
	bApp, err := app.New(logger, db, traceStore, loadLatest, appOpts, []wasmkeeper.Option{}, baseAppOptions...)
	if err != nil {
		panic(err)
	}
	return bApp
}

func fixBankGenesisState(bApp *app.App, rawState map[string]json.RawMessage) {
	bankStateBz, ok := rawState[banktypes.ModuleName]
	if !ok {
		panic("bank genesis state missing from randomized state")
	}
	var bankState banktypes.GenesisState
	bApp.AppCodec().MustUnmarshalJSON(bankStateBz, &bankState)

	// Add Gonka denoms, required by inference module initialization
	bankState.DenomMetadata = append(bankState.DenomMetadata, banktypes.Metadata{
		Description: "Coins for the Gonka network.",
		Base:        inferencetypes.BaseCoin,
		Display:     inferencetypes.NativeCoin,
		Name:        "Gonka",
		Symbol:      "GNK",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: inferencetypes.BaseCoin, Exponent: 0, Aliases: []string{"nanogonka"}},
			{Denom: "ugonka", Exponent: 3, Aliases: []string{"microgonka"}},
			{Denom: "mgonka", Exponent: 6, Aliases: []string{"milligonka"}},
			{Denom: inferencetypes.NativeCoin, Exponent: 9},
		},
	})
	// Bank's randomized supply includes tokens for NumBonded staking validators,
	// but since staking ops are disabled (noopSimModule), the bonded pool is never
	// funded. Recompute Supply from actual Balances to keep genesis consistent.
	var actualSupply sdk.Coins
	for _, balance := range bankState.Balances {
		actualSupply = actualSupply.Add(balance.Coins...)
	}
	bankState.Supply = actualSupply
	rawState[banktypes.ModuleName] = bApp.AppCodec().MustMarshalJSON(&bankState)
}

func setupStateFactory(bApp *app.App) simsx.SimStateFactory {
	return simsx.SimStateFactory{
		Codec: bApp.AppCodec(),
		AppStateFn: simtestutil.AppStateFnWithExtendedCb(
			bApp.AppCodec(),
			bApp.SimulationManager(),
			bApp.DefaultGenesis(),
			func(rawState map[string]json.RawMessage) {
				fixBankGenesisState(bApp, rawState)
			},
		),
		BlockedAddr:   app.BlockedAddresses(),
		AccountSource: bApp.AccountKeeper,
		BalanceSource: bApp.BankKeeper,
	}
}
