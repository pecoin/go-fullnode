package server

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/pecoin/go-fullnode/common/mclock"
	"github.com/pecoin/go-fullnode/p2p/enode"
	"github.com/pecoin/go-fullnode/p2p/enr"
	"github.com/pecoin/go-fullnode/p2p/nodestate"
)

var (
	testSetup         = &nodestate.Setup{}
	ppTestClientFlag  = testSetup.NewFlag("ppTestClientFlag")
	ppTestClientField = testSetup.NewField("ppTestClient", reflect.TypeOf(&ppTestClient{}))
	ppUpdateFlag      = testSetup.NewFlag("ppUpdateFlag")
	ppTestSetup       = NewPriorityPoolSetup(testSetup)
)

func init() {
	ppTestSetup.Connect(ppTestClientField, ppUpdateFlag)
}

const (
	testCapacityStepDiv      = 100
	testCapacityToleranceDiv = 10
	testMinCap               = 100
)

type ppTestClient struct {
	node         *enode.Node
	balance, cap uint64
}

func (c *ppTestClient) Priority(cap uint64) int64 {
	return int64(c.balance / cap)
}

func (c *ppTestClient) EstimatePriority(cap uint64, addBalance int64, future, bias time.Duration, update bool) int64 {
	return int64(c.balance / cap)
}

func TestPriorityPool(t *testing.T) {
	clock := &mclock.Simulated{}
	ns := nodestate.NewNodeStateMachine(nil, nil, clock, testSetup)

	ns.SubscribeField(ppTestSetup.CapacityField, func(node *enode.Node, state nodestate.Flags, oldValue, newValue interface{}) {
		if n := ns.GetField(node, ppTestSetup.priorityField); n != nil {
			c := n.(*ppTestClient)
			c.cap = newValue.(uint64)
		}
	})
	pp := NewPriorityPool(ns, ppTestSetup, clock, testMinCap, 0, testCapacityStepDiv)
	ns.Start()
	pp.SetLimits(100, 1000000)
	clients := make([]*ppTestClient, 100)
	raise := func(c *ppTestClient) {
		for {
			var ok bool
			ns.Operation(func() {
				_, ok = pp.RequestCapacity(c.node, c.cap+c.cap/testCapacityStepDiv, 0, true)
			})
			if !ok {
				return
			}
		}
	}
	var sumBalance uint64
	check := func(c *ppTestClient) {
		expCap := 1000000 * c.balance / sumBalance
		capTol := expCap / testCapacityToleranceDiv
		if c.cap < expCap-capTol || c.cap > expCap+capTol {
			t.Errorf("Wrong node capacity (expected %d, got %d)", expCap, c.cap)
		}
	}

	for i := range clients {
		c := &ppTestClient{
			node:    enode.SignNull(&enr.Record{}, enode.ID{byte(i)}),
			balance: 100000000000,
			cap:     1000,
		}
		sumBalance += c.balance
		clients[i] = c
		ns.SetState(c.node, ppTestClientFlag, nodestate.Flags{}, 0)
		ns.SetField(c.node, ppTestSetup.priorityField, c)
		ns.SetState(c.node, ppTestSetup.InactiveFlag, nodestate.Flags{}, 0)
		raise(c)
		check(c)
	}

	for count := 0; count < 100; count++ {
		c := clients[rand.Intn(len(clients))]
		oldBalance := c.balance
		c.balance = uint64(rand.Int63n(100000000000) + 100000000000)
		sumBalance += c.balance - oldBalance
		pp.ns.SetState(c.node, ppUpdateFlag, nodestate.Flags{}, 0)
		pp.ns.SetState(c.node, nodestate.Flags{}, ppUpdateFlag, 0)
		if c.balance > oldBalance {
			raise(c)
		} else {
			for _, c := range clients {
				raise(c)
			}
		}
		// check whether capacities are proportional to balances
		for _, c := range clients {
			check(c)
		}
		if count%10 == 0 {
			// test available capacity calculation with capacity curve
			c = clients[rand.Intn(len(clients))]
			curve := pp.GetCapacityCurve().Exclude(c.node.ID())

			add := uint64(rand.Int63n(10000000000000))
			c.balance += add
			sumBalance += add
			expCap := curve.MaxCapacity(func(cap uint64) int64 {
				return int64(c.balance / cap)
			})
			//fmt.Println(expCap, c.balance, sumBalance)
			/*for i, cp := range curve.points {
				fmt.Println("cp", i, cp, "ex", curve.getPoint(i))
			}*/
			var ok bool
			expFail := expCap + 1
			if expFail < testMinCap {
				expFail = testMinCap
			}
			ns.Operation(func() {
				_, ok = pp.RequestCapacity(c.node, expFail, 0, true)
			})
			if ok {
				t.Errorf("Request for more than expected available capacity succeeded")
			}
			if expCap >= testMinCap {
				ns.Operation(func() {
					_, ok = pp.RequestCapacity(c.node, expCap, 0, true)
				})
				if !ok {
					t.Errorf("Request for expected available capacity failed")
				}
			}
			c.balance -= add
			sumBalance -= add
			pp.ns.SetState(c.node, ppUpdateFlag, nodestate.Flags{}, 0)
			pp.ns.SetState(c.node, nodestate.Flags{}, ppUpdateFlag, 0)
			for _, c := range clients {
				raise(c)
			}
		}
	}

	ns.Stop()
}

func TestCapacityCurve(t *testing.T) {
	clock := &mclock.Simulated{}
	ns := nodestate.NewNodeStateMachine(nil, nil, clock, testSetup)
	pp := NewPriorityPool(ns, ppTestSetup, clock, 400000, 0, 2)
	ns.Start()
	pp.SetLimits(10, 10000000)
	clients := make([]*ppTestClient, 10)

	for i := range clients {
		c := &ppTestClient{
			node:    enode.SignNull(&enr.Record{}, enode.ID{byte(i)}),
			balance: 100000000000 * uint64(i+1),
			cap:     1000000,
		}
		clients[i] = c
		ns.SetState(c.node, ppTestClientFlag, nodestate.Flags{}, 0)
		ns.SetField(c.node, ppTestSetup.priorityField, c)
		ns.SetState(c.node, ppTestSetup.InactiveFlag, nodestate.Flags{}, 0)
		ns.Operation(func() {
			pp.RequestCapacity(c.node, c.cap, 0, true)
		})
	}

	curve := pp.GetCapacityCurve()
	check := func(balance, expCap uint64) {
		cap := curve.MaxCapacity(func(cap uint64) int64 {
			return int64(balance / cap)
		})
		var fail bool
		if cap == 0 || expCap == 0 {
			fail = cap != expCap
		} else {
			pri := balance / cap
			expPri := balance / expCap
			fail = pri != expPri && pri != expPri+1
		}
		if fail {
			t.Errorf("Incorrect capacity for %d balance (got %d, expected %d)", balance, cap, expCap)
		}
	}

	check(0, 0)
	check(10000000000, 100000)
	check(50000000000, 500000)
	check(100000000000, 1000000)
	check(200000000000, 1000000)
	check(300000000000, 1500000)
	check(450000000000, 1500000)
	check(600000000000, 2000000)
	check(800000000000, 2000000)
	check(1000000000000, 2500000)

	pp.SetLimits(11, 10000000)
	curve = pp.GetCapacityCurve()

	check(0, 0)
	check(10000000000, 100000)
	check(50000000000, 500000)
	check(150000000000, 750000)
	check(200000000000, 1000000)
	check(220000000000, 1100000)
	check(275000000000, 1100000)
	check(375000000000, 1500000)
	check(450000000000, 1500000)
	check(600000000000, 2000000)
	check(800000000000, 2000000)
	check(1000000000000, 2500000)

	ns.Stop()
}
