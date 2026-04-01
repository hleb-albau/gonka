//go:build simsbench

package app_test

import (
	"testing"

	simsx "github.com/cosmos/cosmos-sdk/testutil/simsx"

	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
)

// Profile with:
// go test -tags simsbench -benchmem -run=^$ github.com/productscience/inference/app -bench ^BenchmarkFullAppSimulation$ -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	b.ReportAllocs()

	config := simcli.NewConfigFromFlags()
	config.ChainID = simsx.SimAppChainID

	simsx.RunWithSeed(b, config, newSimApp, setupStateFactory, 1, nil)
}
