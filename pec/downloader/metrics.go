// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/pecoin/go-fullnode/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("pec/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("pec/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("pec/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("pec/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("pec/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("pec/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("pec/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("pec/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("pec/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("pec/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("pec/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("pec/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("pec/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("pec/downloader/states/drop", nil)

	throttleCounter = metrics.NewRegisteredCounter("pec/downloader/throttle", nil)
)
