// Contains all the wrappers from the params package.

package geth

import (
	"encoding/json"

	"github.com/pecoin/go-fullnode/core"
	"github.com/pecoin/go-fullnode/p2p/enode"
	"github.com/pecoin/go-fullnode/params"
)

// MainnetGenesis returns the JSON spec to use for the main Ethereum network. It
// is actually empty since that defaults to the hard coded binary genesis block.
func MainnetGenesis() string {
	return ""
}

// TestnetGenesis returns the JSON spec to use for the Testnet test network
func TestnetGenesis() string {
	enc, err := json.Marshal(core.DefaultTestnetGenesisBlock())
	if err != nil {
		panic(err)
	}
	return string(enc)
}

// FoundationBootnodes returns the enode URLs of the P2P bootstrap nodes operated
// by the foundation running the V5 discovery protocol.
func FoundationBootnodes() *Enodes {
	nodes := &Enodes{nodes: make([]*enode.Node, len(params.MainnetBootnodes))}
	for i, url := range params.MainnetBootnodes {
		var err error
		nodes.nodes[i], err = enode.Parse(enode.ValidSchemes, url)
		if err != nil {
			panic("invalid node URL: " + err.Error())
		}
	}
	return nodes
}
