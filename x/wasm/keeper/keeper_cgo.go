//go:build cgo

package keeper

import (
	"path/filepath"

	wasmvm "github.com/CosmWasm/wasmvm/v2"

	"cosmossdk.io/collections"
	corestoretypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/airchains-network/junction/x/wasm/types"
)

// NewKeeper creates a new contract Keeper instance
// If customEncoders is non-nil, we can use this to override some of the message handler, especially custom
func NewKeeper(
	cdc codec.Codec,
	storeService corestoretypes.KVStoreService,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
	distrKeeper types.DistributionKeeper,
	ics4Wrapper types.ICS4Wrapper,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	capabilityKeeper types.CapabilityKeeper,
	portSource types.ICS20TransferPortSource,
	router MessageRouter,
	_ GRPCQueryRouter,
	homeDir string,
	wasmConfig types.WasmConfig,
	availableCapabilities []string,
	authority string,
	opts ...Option,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	keeper := &Keeper{
		StoreService:         storeService,
		Cdc:                  cdc,
		WasmVM:               nil,
		AccountKeeper:        accountKeeper,
		Bank:                 NewBankCoinTransferrer(bankKeeper),
		accountPruner:        NewVestingCoinBurner(bankKeeper),
		PortKeeper:           portKeeper,
		CapabilityKeeper:     capabilityKeeper,
		Messenger:            NewDefaultMessageHandler(router, ics4Wrapper, channelKeeper, capabilityKeeper, bankKeeper, cdc, portSource),
		QueryGasLimit:        wasmConfig.SmartQueryGasLimit,
		GasRegister:          types.NewDefaultWasmGasRegister(),
		MaxQueryStackSize:    types.DefaultMaxQueryStackSize,
		AcceptedAccountTypes: defaultAcceptedAccountTypes,
		Params:               collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		PropagateGovAuthorization: map[types.AuthorizationPolicyAction]struct{}{
			types.AuthZActionInstantiate: {},
		},
		Authority: authority,
	}
	keeper.wasmVMQueryHandler = DefaultQueryPlugins(bankKeeper, stakingKeeper, distrKeeper, channelKeeper, keeper)
	preOpts, postOpts := splitOpts(opts)
	for _, o := range preOpts {
		o.apply(keeper)
	}
	// only set the wasmvm if no one set this in the options
	// NewVM does a lot, so better not to create it and silently drop it.
	if keeper.WasmVM == nil {
		var err error
		keeper.WasmVM, err = wasmvm.NewVM(filepath.Join(homeDir, "wasm"), availableCapabilities, contractMemoryLimit, wasmConfig.ContractDebugMode, wasmConfig.MemoryCacheSize)
		if err != nil {
			panic(err)
		}
	}

	for _, o := range postOpts {
		o.apply(keeper)
	}
	// not updatable, yet
	keeper.wasmVMResponseHandler = NewDefaultWasmVMContractResponseHandler(NewMessageDispatcher(keeper.Messenger, keeper))
	return *keeper
}
