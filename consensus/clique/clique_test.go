package clique

import (
	"math/big"
	"testing"

	"github.com/pecoin/go-fullnode/common"
	"github.com/pecoin/go-fullnode/core"
	"github.com/pecoin/go-fullnode/core/rawdb"
	"github.com/pecoin/go-fullnode/core/types"
	"github.com/pecoin/go-fullnode/core/vm"
	"github.com/pecoin/go-fullnode/crypto"
	"github.com/pecoin/go-fullnode/params"
)

// This test case is a repro of an annoying bug that took us forever to catch.
// In Clique PoA networks (Rinkeby, Görli, etc), consecutive blocks might have
// the same state root (no block subsidy, empty block). If a node crashes, the
// chain ends up losing the recent state and needs to regenerate it from blocks
// already in the database. The bug was that processing the block *prior* to an
// empty one **also completes** the empty one, ending up in a known-block error.
func TestReimportMirroredState(t *testing.T) {
	// Initialize a Clique chain with a single signer
	var (
		db     = rawdb.NewMemoryDatabase()
		key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		addr   = crypto.PubkeyToAddress(key.PublicKey)
		engine = New(params.AllCliqueProtocolChanges.Clique, db)
		signer = new(types.HomesteadSigner)
	)
	genspec := &core.Genesis{
		ExtraData: make([]byte, extraVanity+common.AddressLength+extraSeal),
		Alloc: map[common.Address]core.GenesisAccount{
			addr: {Balance: big.NewInt(1)},
		},
	}
	copy(genspec.ExtraData[extraVanity:], addr[:])
	genesis := genspec.MustCommit(db)

	// Generate a batch of blocks, each properly signed
	chain, _ := core.NewBlockChain(db, nil, params.AllCliqueProtocolChanges, engine, vm.Config{}, nil, nil)
	defer chain.Stop()

	blocks, _ := core.GenerateChain(params.AllCliqueProtocolChanges, genesis, engine, db, 3, func(i int, block *core.BlockGen) {
		// The chain maker doesn't have access to a chain, so the difficulty will be
		// lets unset (nil). Set it here to the correct value.
		block.SetDifficulty(diffInTurn)

		// We want to simulate an empty middle block, having the same state as the
		// first one. The last is needs a state change again to force a reorg.
		if i != 1 {
			tx, err := types.SignTx(types.NewTransaction(block.TxNonce(addr), common.Address{0x00}, new(big.Int), params.TxGas, nil, nil), signer, key)
			if err != nil {
				panic(err)
			}
			block.AddTxWithChain(chain, tx)
		}
	})
	for i, block := range blocks {
		header := block.Header()
		if i > 0 {
			header.ParentHash = blocks[i-1].Hash()
		}
		header.Extra = make([]byte, extraVanity+extraSeal)
		header.Difficulty = diffInTurn

		sig, _ := crypto.Sign(SealHash(header).Bytes(), key)
		copy(header.Extra[len(header.Extra)-extraSeal:], sig)
		blocks[i] = block.WithSeal(header)
	}
	// Insert the first two blocks and make sure the chain is valid
	db = rawdb.NewMemoryDatabase()
	genspec.MustCommit(db)

	chain, _ = core.NewBlockChain(db, nil, params.AllCliqueProtocolChanges, engine, vm.Config{}, nil, nil)
	defer chain.Stop()

	if _, err := chain.InsertChain(blocks[:2]); err != nil {
		t.Fatalf("failed to insert initial blocks: %v", err)
	}
	if head := chain.CurrentBlock().NumberU64(); head != 2 {
		t.Fatalf("chain head mismatch: have %d, want %d", head, 2)
	}

	// Simulate a crash by creating a new chain on top of the database, without
	// flushing the dirty states out. Insert the last block, triggering a sidechain
	// reimport.
	chain, _ = core.NewBlockChain(db, nil, params.AllCliqueProtocolChanges, engine, vm.Config{}, nil, nil)
	defer chain.Stop()

	if _, err := chain.InsertChain(blocks[2:]); err != nil {
		t.Fatalf("failed to insert final block: %v", err)
	}
	if head := chain.CurrentBlock().NumberU64(); head != 3 {
		t.Fatalf("chain head mismatch: have %d, want %d", head, 3)
	}
}
