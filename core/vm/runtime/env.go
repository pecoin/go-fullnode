package runtime

import (
	"github.com/pecoin/go-fullnode/core"
	"github.com/pecoin/go-fullnode/core/vm"
)

func NewEnv(cfg *Config) *vm.EVM {
	txContext := vm.TxContext{
		Origin:   cfg.Origin,
		GasPrice: cfg.GasPrice,
	}
	blockContext := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     cfg.GetHashFn,
		Coinbase:    cfg.Coinbase,
		BlockNumber: cfg.BlockNumber,
		Time:        cfg.Time,
		Difficulty:  cfg.Difficulty,
		GasLimit:    cfg.GasLimit,
	}

	return vm.NewEVM(blockContext, txContext, cfg.State, cfg.ChainConfig, cfg.EVMConfig)
}
