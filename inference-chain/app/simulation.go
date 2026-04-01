package app

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// disabledOpsSimModule wraps an AppModuleSimulation but suppresses all weighted
// operations. Used for staking: genesis state (validators) still needs to be
// generated so InitGenesis has a non-empty validator set, but delegation /
// undelegation msgs are disabled in this PoC chain and must not be simulated.
type disabledOpsSimModule struct {
	module.AppModuleSimulation
}

func (disabledOpsSimModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
