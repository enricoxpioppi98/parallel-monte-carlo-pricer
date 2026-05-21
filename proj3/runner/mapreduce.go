package runner

import (
	"sync"

	"proj3/barrier"
)

// RunMapReduce is the basic-parallel implementation: a map-reduce with a
// static contiguous partition in the map phase and a sync.Cond barrier
// (required by the assignment) between map and reduce.
//
// Map:    Workers each price a contiguous slice of the options. Results
//
//	are written into the pre-allocated c.Results at disjoint
//	indices, so the map phase is contention-free.
//
// Barrier: c.Threads workers + the main goroutine arrive at the barrier
//
//	(T+1 participants), guaranteeing that the reduce only starts
//	after every map task has finished.
//
// Reduce: The main goroutine sums per-option prices.
//
// The static partition is deliberate: with the unbalanced portfolio
// (heavy options grouped at the end) some workers will finish almost
// instantly while one is still grinding through a chunk of Americans.
// That tail latency is the gap the work-stealing runner closes.
func RunMapReduce(c *Config) {
	n := len(c.Portfolio.Options)
	T := clampThreads(c.Threads, n)

	bar := barrier.New(T + 1) // workers + main goroutine
	var wg sync.WaitGroup
	chunk := (n + T - 1) / T

	for t := 0; t < T; t++ {
		start := t * chunk
		end := start + chunk
		if end > n {
			end = n
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				runOne(c, i)
			}
			bar.Wait() // map phase done
		}(start, end)
	}

	bar.Wait() // main goroutine releases when all workers arrive
	wg.Wait()
	// Reduce happens in main.go via Config.PortfolioValue(), so the timed
	// region of each runner is the map phase only and the three modes
	// stay comparable.
}
