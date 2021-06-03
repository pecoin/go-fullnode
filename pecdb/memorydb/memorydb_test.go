package memorydb

import (
	"testing"

	"github.com/pecoin/go-fullnode/pecdb"
	"github.com/pecoin/go-fullnode/pecdb/dbtest"
)

func TestMemoryDB(t *testing.T) {
	t.Run("DatabaseSuite", func(t *testing.T) {
		dbtest.TestDatabaseSuite(t, func() pecdb.KeyValueStore {
			return New()
		})
	})
}
