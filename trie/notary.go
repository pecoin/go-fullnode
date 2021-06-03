package trie

import (
	"github.com/pecoin/go-fullnode/pecdb"
	"github.com/pecoin/go-fullnode/pecdb/memorydb"
)

// KeyValueNotary tracks which keys have been accessed through a key-value reader
// with te scope of verifying if certain proof datasets are maliciously bloated.
type KeyValueNotary struct {
	pecdb.KeyValueReader
	reads map[string]struct{}
}

// NewKeyValueNotary wraps a key-value database with an access notary to track
// which items have bene accessed.
func NewKeyValueNotary(db pecdb.KeyValueReader) *KeyValueNotary {
	return &KeyValueNotary{
		KeyValueReader: db,
		reads:          make(map[string]struct{}),
	}
}

// Get retrieves an item from the underlying database, but also tracks it as an
// accessed slot for bloat checks.
func (k *KeyValueNotary) Get(key []byte) ([]byte, error) {
	k.reads[string(key)] = struct{}{}
	return k.KeyValueReader.Get(key)
}

// Accessed returns s snapshot of the original key-value store containing only the
// data accessed through the notary.
func (k *KeyValueNotary) Accessed() pecdb.KeyValueStore {
	db := memorydb.New()
	for keystr := range k.reads {
		key := []byte(keystr)
		val, _ := k.KeyValueReader.Get(key)
		db.Put(key, val)
	}
	return db
}
