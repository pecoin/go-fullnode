package client

import (
	"testing"
	"time"

	"github.com/pecoin/go-fullnode/common/mclock"
	"github.com/pecoin/go-fullnode/p2p/enode"
	"github.com/pecoin/go-fullnode/p2p/enr"
	"github.com/pecoin/go-fullnode/p2p/nodestate"
)

func testNode(i int) *enode.Node {
	return enode.SignNull(new(enr.Record), testNodeID(i))
}

func TestQueueIteratorFIFO(t *testing.T) {
	testQueueIterator(t, true)
}

func TestQueueIteratorLIFO(t *testing.T) {
	testQueueIterator(t, false)
}

func testQueueIterator(t *testing.T, fifo bool) {
	ns := nodestate.NewNodeStateMachine(nil, nil, &mclock.Simulated{}, testSetup)
	qi := NewQueueIterator(ns, sfTest2, sfTest3.Or(sfTest4), fifo, nil)
	ns.Start()
	for i := 1; i <= iterTestNodeCount; i++ {
		ns.SetState(testNode(i), sfTest1, nodestate.Flags{}, 0)
	}
	next := func() int {
		ch := make(chan struct{})
		go func() {
			qi.Next()
			close(ch)
		}()
		select {
		case <-ch:
		case <-time.After(time.Second * 5):
			t.Fatalf("Iterator.Next() timeout")
		}
		node := qi.Node()
		ns.SetState(node, sfTest4, nodestate.Flags{}, 0)
		return testNodeIndex(node.ID())
	}
	exp := func(i int) {
		n := next()
		if n != i {
			t.Errorf("Wrong item returned by iterator (expected %d, got %d)", i, n)
		}
	}
	explist := func(list []int) {
		for i := range list {
			if fifo {
				exp(list[i])
			} else {
				exp(list[len(list)-1-i])
			}
		}
	}

	ns.SetState(testNode(1), sfTest2, nodestate.Flags{}, 0)
	ns.SetState(testNode(2), sfTest2, nodestate.Flags{}, 0)
	ns.SetState(testNode(3), sfTest2, nodestate.Flags{}, 0)
	explist([]int{1, 2, 3})
	ns.SetState(testNode(4), sfTest2, nodestate.Flags{}, 0)
	ns.SetState(testNode(5), sfTest2, nodestate.Flags{}, 0)
	ns.SetState(testNode(6), sfTest2, nodestate.Flags{}, 0)
	ns.SetState(testNode(5), sfTest3, nodestate.Flags{}, 0)
	explist([]int{4, 6})
	ns.SetState(testNode(1), nodestate.Flags{}, sfTest4, 0)
	ns.SetState(testNode(2), nodestate.Flags{}, sfTest4, 0)
	ns.SetState(testNode(3), nodestate.Flags{}, sfTest4, 0)
	ns.SetState(testNode(2), sfTest3, nodestate.Flags{}, 0)
	ns.SetState(testNode(2), nodestate.Flags{}, sfTest3, 0)
	explist([]int{1, 3, 2})
	ns.Stop()
}
