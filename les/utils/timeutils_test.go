package utils

import (
	"testing"
	"time"

	"github.com/pecoin/go-fullnode/common/mclock"
)

func TestUpdateTimer(t *testing.T) {
	timer := NewUpdateTimer(mclock.System{}, -1)
	if timer != nil {
		t.Fatalf("Create update timer with negative threshold")
	}
	sim := &mclock.Simulated{}
	timer = NewUpdateTimer(sim, time.Second)
	if updated := timer.Update(func(diff time.Duration) bool { return true }); updated {
		t.Fatalf("Update the clock without reaching the threshold")
	}
	sim.Run(time.Second)
	if updated := timer.Update(func(diff time.Duration) bool { return true }); !updated {
		t.Fatalf("Doesn't update the clock when reaching the threshold")
	}
	if updated := timer.UpdateAt(sim.Now()+mclock.AbsTime(time.Second), func(diff time.Duration) bool { return true }); !updated {
		t.Fatalf("Doesn't update the clock when reaching the threshold")
	}
	timer = NewUpdateTimer(sim, 0)
	if updated := timer.Update(func(diff time.Duration) bool { return true }); !updated {
		t.Fatalf("Doesn't update the clock without threshold limitaion")
	}
}
