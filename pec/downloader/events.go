package downloader

import "github.com/pecoin/go-fullnode/core/types"

type DoneEvent struct {
	Latest *types.Header
}
type StartEvent struct{}
type FailedEvent struct{ Err error }
