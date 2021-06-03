package les

import (
	"github.com/pecoin/go-fullnode/core/forkid"
	"github.com/pecoin/go-fullnode/p2p/dnsdisc"
	"github.com/pecoin/go-fullnode/p2p/enode"
	"github.com/pecoin/go-fullnode/rlp"
)

// lesEntry is the "les" ENR entry. This is set for LES servers only.
type lesEntry struct {
	// Ignore additional fields (for forward compatibility).
	VfxVersion uint
	Rest       []rlp.RawValue `rlp:"tail"`
}

func (lesEntry) ENRKey() string { return "les" }

// ethEntry is the "pec" ENR entry. This is redeclared here to avoid depending on package pec.
type ethEntry struct {
	ForkID forkid.ID
	_      []rlp.RawValue `rlp:"tail"`
}

func (ethEntry) ENRKey() string { return "pec" }

// setupDiscovery creates the node discovery source for the pec protocol.
func (eth *LightEthereum) setupDiscovery() (enode.Iterator, error) {
	it := enode.NewFairMix(0)

	// Enable DNS discovery.
	if len(eth.config.EthDiscoveryURLs) != 0 {
		client := dnsdisc.NewClient(dnsdisc.Config{})
		dns, err := client.NewIterator(eth.config.EthDiscoveryURLs...)
		if err != nil {
			return nil, err
		}
		it.AddSource(dns)
	}

	// Enable DHT.
	if eth.udpEnabled {
		it.AddSource(eth.p2pServer.DiscV5.RandomNodes())
	}

	forkFilter := forkid.NewFilter(eth.blockchain)
	iterator := enode.Filter(it, func(n *enode.Node) bool { return nodeIsServer(forkFilter, n) })
	return iterator, nil
}

// nodeIsServer checks whether n is an LES server node.
func nodeIsServer(forkFilter forkid.Filter, n *enode.Node) bool {
	var les lesEntry
	var eth ethEntry
	return n.Load(&les) == nil && n.Load(&eth) == nil && forkFilter(eth.ForkID) == nil
}
