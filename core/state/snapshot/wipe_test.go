package snapshot

import (
	"math/rand"
	"testing"

	"github.com/pecoin/go-fullnode/common"
	"github.com/pecoin/go-fullnode/core/rawdb"
	"github.com/pecoin/go-fullnode/pecdb/memorydb"
)

// Tests that given a database with random data content, all parts of a snapshot
// can be crrectly wiped without touching anything else.
func TestWipe(t *testing.T) {
	// Create a database with some random snapshot data
	db := memorydb.New()

	for i := 0; i < 128; i++ {
		account := randomHash()
		rawdb.WriteAccountSnapshot(db, account, randomHash().Bytes())
		for j := 0; j < 1024; j++ {
			rawdb.WriteStorageSnapshot(db, account, randomHash(), randomHash().Bytes())
		}
	}
	rawdb.WriteSnapshotRoot(db, randomHash())

	// Add some random non-snapshot data too to make wiping harder
	for i := 0; i < 65536; i++ {
		// Generate a key that's the wrong length for a state snapshot item
		var keysize int
		for keysize == 0 || keysize == 32 || keysize == 64 {
			keysize = 8 + rand.Intn(64) // +8 to ensure we will "never" randomize duplicates
		}
		// Randomize the suffix, dedup and inject it under the snapshot namespace
		keysuffix := make([]byte, keysize)
		rand.Read(keysuffix)

		if rand.Int31n(2) == 0 {
			db.Put(append(rawdb.SnapshotAccountPrefix, keysuffix...), randomHash().Bytes())
		} else {
			db.Put(append(rawdb.SnapshotStoragePrefix, keysuffix...), randomHash().Bytes())
		}
	}
	// Sanity check that all the keys are present
	var items int

	it := db.NewIterator(rawdb.SnapshotAccountPrefix, nil)
	defer it.Release()

	for it.Next() {
		key := it.Key()
		if len(key) == len(rawdb.SnapshotAccountPrefix)+common.HashLength {
			items++
		}
	}
	it = db.NewIterator(rawdb.SnapshotStoragePrefix, nil)
	defer it.Release()

	for it.Next() {
		key := it.Key()
		if len(key) == len(rawdb.SnapshotStoragePrefix)+2*common.HashLength {
			items++
		}
	}
	if items != 128+128*1024 {
		t.Fatalf("snapshot size mismatch: have %d, want %d", items, 128+128*1024)
	}
	if hash := rawdb.ReadSnapshotRoot(db); hash == (common.Hash{}) {
		t.Errorf("snapshot block marker mismatch: have %#x, want <not-nil>", hash)
	}
	// Wipe all snapshot entries from the database
	<-wipeSnapshot(db, true)

	// Iterate over the database end ensure no snapshot information remains
	it = db.NewIterator(rawdb.SnapshotAccountPrefix, nil)
	defer it.Release()

	for it.Next() {
		key := it.Key()
		if len(key) == len(rawdb.SnapshotAccountPrefix)+common.HashLength {
			t.Errorf("snapshot entry remained after wipe: %x", key)
		}
	}
	it = db.NewIterator(rawdb.SnapshotStoragePrefix, nil)
	defer it.Release()

	for it.Next() {
		key := it.Key()
		if len(key) == len(rawdb.SnapshotStoragePrefix)+2*common.HashLength {
			t.Errorf("snapshot entry remained after wipe: %x", key)
		}
	}
	if hash := rawdb.ReadSnapshotRoot(db); hash != (common.Hash{}) {
		t.Errorf("snapshot block marker remained after wipe: %#x", hash)
	}
	// Iterate over the database and ensure miscellaneous items are present
	items = 0

	it = db.NewIterator(nil, nil)
	defer it.Release()

	for it.Next() {
		items++
	}
	if items != 65536 {
		t.Fatalf("misc item count mismatch: have %d, want %d", items, 65536)
	}
}
