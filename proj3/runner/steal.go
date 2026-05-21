package runner

import (
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"

	"proj3/deque"
)

// RunWorkStealing executes pricing tasks on per-worker Chase-Lev deques.
//
// Init:   Tasks are distributed round-robin across the T deques so the
//
//	heavy options (Asians/Americans, grouped at the end of the
//	options slice) are spread across all workers, not piled onto
//	one.
//
// Worker: Pops from its own deque (LIFO). When empty, picks a random
//
//	victim and tries to steal from the top (FIFO). Loops until
//	every task has been consumed.
//
// Termination is by global counter: `remaining` starts at n and is
// decremented after each runOne. Workers exit when remaining hits 0.
// Because no worker pushes tasks after init, no fancy termination
// detection is required.
//
// Result writes are race-free: each goroutine writes c.Results[idx]
// for an idx it just popped/stole, so two goroutines can never touch
// the same slot.
func RunWorkStealing(c *Config) {
	n := len(c.Portfolio.Options)
	T := clampThreads(c.Threads, n)

	capacity := 1
	for capacity < n {
		capacity <<= 1
	}
	if capacity < 16 {
		capacity = 16
	}

	deques := make([]*deque.Deque, T)
	for t := range deques {
		deques[t] = deque.New(capacity)
	}
	for i := 0; i < n; i++ {
		deques[i%T].PushBottom(deque.Task{Index: i})
	}

	var remaining atomic.Int64
	remaining.Store(int64(n))

	var wg sync.WaitGroup
	wg.Add(T)
	for tID := 0; tID < T; tID++ {
		go func(self int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(int64(self) + 1))
			myDeque := deques[self]
			for {
				if task, ok := myDeque.PopBottom(); ok {
					runOne(c, task.Index)
					remaining.Add(-1)
					continue
				}
				// Own deque empty: see if we're done first, else try
				// stealing from a random victim.
				if remaining.Load() == 0 {
					return
				}
				if T > 1 {
					victim := r.Intn(T)
					if victim == self {
						victim = (victim + 1) % T
					}
					if task, ok := deques[victim].Steal(); ok {
						runOne(c, task.Index)
						remaining.Add(-1)
						continue
					}
				}
				// No work to steal right now; yield so the runtime can
				// schedule other goroutines instead of spinning.
				runtime.Gosched()
			}
		}(tID)
	}
	wg.Wait()
	// Reduce happens in main.go via Config.PortfolioValue().
}
