// Package runner contains the three execution strategies (sequential,
// static-partition map-reduce, work-stealing) that all price the same
// portfolio of options. The pricing primitive `runOne` and the shared
// Config are defined here so each strategy file stays focused on its
// own scheduling logic.
package runner

import (
	"proj3/option"
	"proj3/portfolio"
)

// Config is the shared runtime state for all runners.
// Results is pre-allocated with one slot per option; workers write to
// disjoint indices, so no synchronisation is needed on the slice itself.
type Config struct {
	Portfolio portfolio.Portfolio
	Threads   int
	Seed      uint64
	Results   []option.Result
}

// NewConfig builds a Config with the results slice pre-allocated.
func NewConfig(p portfolio.Portfolio, threads int, seed uint64) *Config {
	return &Config{
		Portfolio: p,
		Threads:   threads,
		Seed:      seed,
		Results:   make([]option.Result, len(p.Options)),
	}
}

// PortfolioValue is the reduce step shared by all runners: the sum of
// per-option prices. Kept on Config so callers can re-run reduce after
// any execution mode.
func (c *Config) PortfolioValue() float64 {
	var sum float64
	for _, r := range c.Results {
		sum += r.Price
	}
	return sum
}

// runOne prices the option at the given index. Each call creates its own
// RNG seeded from (c.Seed, idx), making prices independent of execution
// order or thread count.
func runOne(c *Config, idx int) {
	spec := c.Portfolio.Options[idx]
	rng := option.NewRNG(c.Seed, idx)
	c.Results[idx] = option.Price(spec, rng)
}

// clampThreads bounds the requested thread count to [1, n]. Used by the
// parallel runners so the rest of the code can assume 1 <= T <= n.
func clampThreads(T, n int) int {
	if T < 1 {
		return 1
	}
	if T > n {
		return n
	}
	return T
}
