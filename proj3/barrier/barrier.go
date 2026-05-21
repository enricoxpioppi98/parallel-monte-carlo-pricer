// Package barrier provides a reusable n-thread synchronization barrier
// implemented with sync.Mutex + sync.Cond, as required by the project
// brief for the basic-parallel runner.
package barrier

import "sync"

// Barrier blocks Wait callers until exactly n of them have called Wait.
// Once the n'th caller arrives, all are released and the barrier resets,
// so the same Barrier can be used across multiple supersteps. A generation
// counter is used to handle spurious wakeups and reuse correctly.
type Barrier struct {
	mu      sync.Mutex
	cond    *sync.Cond
	n       int
	waiting int
	gen     uint64
}

// New constructs a Barrier that releases when n participants have arrived.
func New(n int) *Barrier {
	b := &Barrier{n: n}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// Wait blocks until n total callers have invoked Wait on this barrier
// for the current generation. The last arrival broadcasts to release
// everyone and bumps the generation so the barrier is ready for reuse.
func (b *Barrier) Wait() {
	b.mu.Lock()
	gen := b.gen
	b.waiting++
	if b.waiting == b.n {
		b.gen++
		b.waiting = 0
		b.cond.Broadcast()
		b.mu.Unlock()
		return
	}
	for gen == b.gen {
		b.cond.Wait()
	}
	b.mu.Unlock()
}
