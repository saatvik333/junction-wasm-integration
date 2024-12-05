package app

import (
	wasmkeeper "github.com/airchains-network/junction/x/wasm/keeper"
)

// Deprecated: Use BuiltInCapabilities from github.com/CosmWasm/wasmd/x/wasm/keeper
func AllCapabilities() []string {
	return wasmkeeper.BuiltInCapabilities()
}
