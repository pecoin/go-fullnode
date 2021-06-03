package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/docker/docker/pkg/reexec"
	"github.com/pecoin/go-fullnode/internal/cmdtest"
)

type testPeckey struct {
	*cmdtest.TestCmd
}

// spawns peckey with the given command line args.
func runPeckey(t *testing.T, args ...string) *testPeckey {
	tt := new(testPeckey)
	tt.TestCmd = cmdtest.NewTestCmd(t, tt)
	tt.Run("peckey-test", args...)
	return tt
}

func TestMain(m *testing.M) {
	// Run the app if we've been exec'd as "peckey-test" in runPeckey.
	reexec.Register("peckey-test", func() {
		if err := app.Run(os.Args); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	})
	// check if we have been reexec'd
	if reexec.Init() {
		return
	}
	os.Exit(m.Run())
}
