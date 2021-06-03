package les

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/pecoin/go-fullnode/common"
	"github.com/pecoin/go-fullnode/core"
	"github.com/pecoin/go-fullnode/core/rawdb"
	"github.com/pecoin/go-fullnode/core/types"
	"github.com/pecoin/go-fullnode/les/checkpointoracle"
	"github.com/pecoin/go-fullnode/light"
	"github.com/pecoin/go-fullnode/log"
	"github.com/pecoin/go-fullnode/node"
	"github.com/pecoin/go-fullnode/p2p"
	"github.com/pecoin/go-fullnode/p2p/enode"
	"github.com/pecoin/go-fullnode/params"
	"github.com/pecoin/go-fullnode/pec/ethconfig"
	"github.com/pecoin/go-fullnode/pecdb"
	"github.com/pecoin/go-fullnode/peclient"
)

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

type chainReader interface {
	CurrentHeader() *types.Header
}

// lesCommons contains fields needed by both server and client.
type lesCommons struct {
	genesis                      common.Hash
	config                       *ethconfig.Config
	chainConfig                  *params.ChainConfig
	iConfig                      *light.IndexerConfig
	chainDb, lesDb               pecdb.Database
	chainReader                  chainReader
	chtIndexer, bloomTrieIndexer *core.ChainIndexer
	oracle                       *checkpointoracle.CheckpointOracle

	closeCh chan struct{}
	wg      sync.WaitGroup
}

// NodeInfo represents a short summary of the Ethereum sub-protocol metadata
// known about the host peer.
type NodeInfo struct {
	Network    uint64                   `json:"network"`    // Ethereum network ID (1=Frontier, 2=Testnet)
	Difficulty *big.Int                 `json:"difficulty"` // Total difficulty of the host's blockchain
	Genesis    common.Hash              `json:"genesis"`    // SHA3 hash of the host's genesis block
	Config     *params.ChainConfig      `json:"config"`     // Chain configuration for the fork rules
	Head       common.Hash              `json:"head"`       // SHA3 hash of the host's best owned block
	CHT        params.TrustedCheckpoint `json:"cht"`        // Trused CHT checkpoint for fast catchup
}

// makeProtocols creates protocol descriptors for the given LES versions.
func (c *lesCommons) makeProtocols(versions []uint, runPeer func(version uint, p *p2p.Peer, rw p2p.MsgReadWriter) error, peerInfo func(id enode.ID) interface{}, dialCandidates enode.Iterator) []p2p.Protocol {
	protos := make([]p2p.Protocol, len(versions))
	for i, version := range versions {
		version := version
		protos[i] = p2p.Protocol{
			Name:     "les",
			Version:  version,
			Length:   ProtocolLengths[version],
			NodeInfo: c.nodeInfo,
			Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
				return runPeer(version, peer, rw)
			},
			PeerInfo:       peerInfo,
			DialCandidates: dialCandidates,
		}
	}
	return protos
}

// nodeInfo retrieves some protocol metadata about the running host node.
func (c *lesCommons) nodeInfo() interface{} {
	head := c.chainReader.CurrentHeader()
	hash := head.Hash()
	return &NodeInfo{
		Network:    c.config.NetworkId,
		Difficulty: rawdb.ReadTd(c.chainDb, hash, head.Number.Uint64()),
		Genesis:    c.genesis,
		Config:     c.chainConfig,
		Head:       hash,
		CHT:        c.latestLocalCheckpoint(),
	}
}

// latestLocalCheckpoint finds the common stored section index and returns a set
// of post-processed trie roots (CHT and BloomTrie) associated with the appropriate
// section index and head hash as a local checkpoint package.
func (c *lesCommons) latestLocalCheckpoint() params.TrustedCheckpoint {
	sections, _, _ := c.chtIndexer.Sections()
	sections2, _, _ := c.bloomTrieIndexer.Sections()
	// Cap the section index if the two sections are not consistent.
	if sections > sections2 {
		sections = sections2
	}
	if sections == 0 {
		// No checkpoint information can be provided.
		return params.TrustedCheckpoint{}
	}
	return c.localCheckpoint(sections - 1)
}

// localCheckpoint returns a set of post-processed trie roots (CHT and BloomTrie)
// associated with the appropriate head hash by specific section index.
//
// The returned checkpoint is only the checkpoint generated by the local indexers,
// not the stable checkpoint registered in the registrar contract.
func (c *lesCommons) localCheckpoint(index uint64) params.TrustedCheckpoint {
	sectionHead := c.chtIndexer.SectionHead(index)
	return params.TrustedCheckpoint{
		SectionIndex: index,
		SectionHead:  sectionHead,
		CHTRoot:      light.GetChtRoot(c.chainDb, index, sectionHead),
		BloomRoot:    light.GetBloomTrieRoot(c.chainDb, index, sectionHead),
	}
}

// setupOracle sets up the checkpoint oracle contract client.
func (c *lesCommons) setupOracle(node *node.Node, genesis common.Hash, ethconfig *ethconfig.Config) *checkpointoracle.CheckpointOracle {
	config := ethconfig.CheckpointOracle
	if config == nil {
		// Try loading default config.
		config = params.CheckpointOracles[genesis]
	}
	if config == nil {
		log.Info("Checkpoint oracle is not enabled")
		return nil
	}
	if config.Address == (common.Address{}) || uint64(len(config.Signers)) < config.Threshold {
		log.Warn("Invalid checkpoint oracle config")
		return nil
	}
	oracle := checkpointoracle.New(config, c.localCheckpoint)
	rpcClient, _ := node.Attach()
	client := peclient.NewClient(rpcClient)
	oracle.Start(client)
	log.Info("Configured checkpoint oracle", "address", config.Address, "signers", len(config.Signers), "threshold", config.Threshold)
	return oracle
}
