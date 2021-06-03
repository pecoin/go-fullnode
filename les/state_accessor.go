package les

import (
	"context"
	"errors"
	"fmt"

	"github.com/pecoin/go-fullnode/core"
	"github.com/pecoin/go-fullnode/core/state"
	"github.com/pecoin/go-fullnode/core/types"
	"github.com/pecoin/go-fullnode/core/vm"
	"github.com/pecoin/go-fullnode/light"
)

// stateAtBlock retrieves the state database associated with a certain block.
func (leth *LightEthereum) stateAtBlock(ctx context.Context, block *types.Block, reexec uint64) (*state.StateDB, func(), error) {
	return light.NewState(ctx, block.Header(), leth.odr), func() {}, nil
}

// statesInRange retrieves a batch of state databases associated with the specific
// block ranges.
func (leth *LightEthereum) statesInRange(ctx context.Context, fromBlock *types.Block, toBlock *types.Block, reexec uint64) ([]*state.StateDB, func(), error) {
	var states []*state.StateDB
	for number := fromBlock.NumberU64(); number <= toBlock.NumberU64(); number++ {
		header, err := leth.blockchain.GetHeaderByNumberOdr(ctx, number)
		if err != nil {
			return nil, nil, err
		}
		states = append(states, light.NewState(ctx, header, leth.odr))
	}
	return states, nil, nil
}

// stateAtTransaction returns the execution environment of a certain transaction.
func (leth *LightEthereum) stateAtTransaction(ctx context.Context, block *types.Block, txIndex int, reexec uint64) (core.Message, vm.BlockContext, *state.StateDB, func(), error) {
	// Short circuit if it's genesis block.
	if block.NumberU64() == 0 {
		return nil, vm.BlockContext{}, nil, nil, errors.New("no transaction in genesis")
	}
	// Create the parent state database
	parent, err := leth.blockchain.GetBlock(ctx, block.ParentHash(), block.NumberU64()-1)
	if err != nil {
		return nil, vm.BlockContext{}, nil, nil, err
	}
	statedb, _, err := leth.stateAtBlock(ctx, parent, reexec)
	if err != nil {
		return nil, vm.BlockContext{}, nil, nil, err
	}
	if txIndex == 0 && len(block.Transactions()) == 0 {
		return nil, vm.BlockContext{}, statedb, func() {}, nil
	}
	// Recompute transactions up to the target index.
	signer := types.MakeSigner(leth.blockchain.Config(), block.Number())
	for idx, tx := range block.Transactions() {
		// Assemble the transaction call message and return if the requested offset
		msg, _ := tx.AsMessage(signer)
		txContext := core.NewEVMTxContext(msg)
		context := core.NewEVMBlockContext(block.Header(), leth.blockchain, nil)
		statedb.Prepare(tx.Hash(), block.Hash(), idx)
		if idx == txIndex {
			return msg, context, statedb, func() {}, nil
		}
		// Not yet the searched for transaction, execute on top of the current state
		vmenv := vm.NewEVM(context, txContext, statedb, leth.blockchain.Config(), vm.Config{})
		if _, err := core.ApplyMessage(vmenv, msg, new(core.GasPool).AddGas(tx.Gas())); err != nil {
			return nil, vm.BlockContext{}, nil, nil, fmt.Errorf("transaction %#x failed: %v", tx.Hash(), err)
		}
		// Ensure any modifications are committed to the state
		// Only delete empty objects if EIP158/161 (a.k.a Spurious Dragon) is in effect
		statedb.Finalise(vmenv.ChainConfig().IsEIP158(block.Number()))
	}
	return nil, vm.BlockContext{}, nil, nil, fmt.Errorf("transaction index %d out of range for block %#x", txIndex, block.Hash())
}
