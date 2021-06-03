package pec

import (
	"github.com/pecoin/go-fullnode/core"
	"github.com/pecoin/go-fullnode/core/forkid"
	"github.com/pecoin/go-fullnode/p2p/dnsdisc"
	"github.com/pecoin/go-fullnode/p2p/enode"
	"github.com/pecoin/go-fullnode/rlp"
)

// ethEntry is the "pec" ENR entry which advertises pec protocol
// on the discovery network.
type ethEntry struct {
	ForkID forkid.ID // Fork identifier per EIP-2124

	// Ignore additional fields (for forward compatibility).
	Rest []rlp.RawValue `rlp:"tail"`
}

// ENRKey implements enr.Entry.
func (e ethEntry) ENRKey() string {
	return "pec"
}

// startEthEntryUpdate starts the ENR updater loop.
func (eth *Ethereum) startEthEntryUpdate(ln *enode.LocalNode) {
	var newHead = make(chan core.ChainHeadEvent, 10)
	sub := eth.blockchain.SubscribeChainHeadEvent(newHead)

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case <-newHead:
				ln.Set(eth.currentEthEntry())
			case <-sub.Err():
				// Would be nice to sync with pec.Stop, but there is no
				// good way to do that.
				return
			}
		}
	}()
}

func (eth *Ethereum) currentEthEntry() *ethEntry {
	return &ethEntry{ForkID: forkid.NewID(eth.blockchain.Config(), eth.blockchain.Genesis().Hash(),
		eth.blockchain.CurrentHeader().Number.Uint64())}
}

// setupDiscovery creates the node discovery source for the `pec` and `snap`
// protocols.
func setupDiscovery(urls []string) (enode.Iterator, error) {
	if len(urls) == 0 {
		return nil, nil
	}
	client := dnsdisc.NewClient(dnsdisc.Config{})
	return client.NewIterator(urls...)
}
