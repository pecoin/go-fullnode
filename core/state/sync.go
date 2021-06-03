package state

import (
	"bytes"

	"github.com/pecoin/go-fullnode/common"
	"github.com/pecoin/go-fullnode/pecdb"
	"github.com/pecoin/go-fullnode/rlp"
	"github.com/pecoin/go-fullnode/trie"
)

// NewStateSync create a new state trie download scheduler.
func NewStateSync(root common.Hash, database pecdb.KeyValueReader, bloom *trie.SyncBloom) *trie.Sync {
	var syncer *trie.Sync
	callback := func(path []byte, leaf []byte, parent common.Hash) error {
		var obj Account
		if err := rlp.Decode(bytes.NewReader(leaf), &obj); err != nil {
			return err
		}
		syncer.AddSubTrie(obj.Root, path, parent, nil)
		syncer.AddCodeEntry(common.BytesToHash(obj.CodeHash), path, parent)
		return nil
	}
	syncer = trie.NewSync(root, database, callback, bloom)
	return syncer
}
