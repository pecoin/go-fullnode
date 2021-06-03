package utils

import "testing"

func TestExecQueue(t *testing.T) {
	var (
		N        = 10000
		q        = NewExecQueue(N)
		counter  int
		execd    = make(chan int)
		testexit = make(chan struct{})
	)
	defer q.Quit()
	defer close(testexit)

	check := func(state string, wantOK bool) {
		c := counter
		counter++
		qf := func() {
			select {
			case execd <- c:
			case <-testexit:
			}
		}
		if q.CanQueue() != wantOK {
			t.Fatalf("CanQueue() == %t for %s", !wantOK, state)
		}
		if q.Queue(qf) != wantOK {
			t.Fatalf("Queue() == %t for %s", !wantOK, state)
		}
	}

	for i := 0; i < N; i++ {
		check("queue below cap", true)
	}
	check("full queue", false)
	for i := 0; i < N; i++ {
		if c := <-execd; c != i {
			t.Fatal("execution out of order")
		}
	}
	q.Quit()
	check("closed queue", false)
}
