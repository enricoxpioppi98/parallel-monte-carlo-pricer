// Package deque implements an array-based lock-free work-stealing deque,
// the data structure required by the project brief for the work-stealing
// runner.
//
// The owner pushes and pops from the bottom (LIFO); thieves steal from
// the top (FIFO). All synchronisation is done with atomics only - no
// mutexes - which is what the assignment requires.
//
// The deque uses a fixed-power-of-two backing array. Callers size it to
// accommodate the largest initial workload; growth is not implemented
// because the work-stealing runner only pushes during init.
package deque

import "sync/atomic"

// Task is the unit of work stored in the deque. Index references the
// option's position in the shared Portfolio.Options slice.
type Task struct {
	Index int
}

// emptyTask is what a failed Pop/Steal returns. Callers should look at
// the ok bool, not the index value.
var emptyTask = Task{Index: -1}

// cachelinePad is sized so a single atomic.Int64 (8 bytes) plus padding
// fills one 64-byte cache line. This isolates `top` (hammered by thieves)
// from `bottom` (hammered by the owner) so the two cache lines bouncing
// between cores never collide on the same line.
const cachelinePad = 56

// Deque is a single-producer / multi-consumer lock-free deque.
// "Producer" here is the owning worker - it alone calls PushBottom and
// PopBottom. Any number of thieves can call Steal concurrently.
//
// Cache-line padding around `top` and `bottom` is a deliberate
// micro-optimisation: without it, the two atomics share one 64-byte line
// and the line bounces between every thief's core and the owner's core
// on every Steal, costing measurable throughput at T >= 8.
type Deque struct {
	_      [cachelinePad]byte
	top    atomic.Int64 // thief end; only increases
	_      [cachelinePad]byte
	bottom atomic.Int64 // owner end; both grows and shrinks
	_      [cachelinePad]byte
	mask   int64 // capacity - 1; capacity is always a power of two
	buf    []atomic.Int64
}

// New allocates a deque whose backing array holds at most `capacity`
// elements. capacity must be a positive power of two.
func New(capacity int) *Deque {
	if capacity <= 0 || capacity&(capacity-1) != 0 {
		panic("deque: capacity must be a positive power of two")
	}
	return &Deque{
		mask: int64(capacity - 1),
		buf:  make([]atomic.Int64, capacity),
	}
}

// PushBottom enqueues a task at the owner's end. Owner-only.
// Panics if the deque is full, which our runner avoids by sizing the
// deque to the full task count up front.
func (d *Deque) PushBottom(t Task) {
	b := d.bottom.Load()
	if b-d.top.Load() >= int64(len(d.buf)) {
		panic("deque: overflow")
	}
	d.buf[b&d.mask].Store(int64(t.Index))
	d.bottom.Store(b + 1)
}

// PopBottom removes a task from the owner's end (LIFO). Owner-only.
// Returns (task, true) on success, (emptyTask, false) if empty or if a
// thief raced and won the last element.
func (d *Deque) PopBottom() (Task, bool) {
	b := d.bottom.Load() - 1
	d.bottom.Store(b)
	// Sequentially consistent load on top gives the fence the algorithm
	// requires between the bottom-decrement and the top-read.
	t := d.top.Load()
	if t > b {
		// Deque was already empty; restore bottom so further pushes work.
		d.bottom.Store(t)
		return emptyTask, false
	}
	idx := d.buf[b&d.mask].Load()
	if t < b {
		// More than one element; the owner's read is uncontested.
		return Task{Index: int(idx)}, true
	}
	// Exactly one element - race with thieves for it.
	if !d.top.CompareAndSwap(t, t+1) {
		// Thief beat us to it.
		d.bottom.Store(t + 1)
		return emptyTask, false
	}
	d.bottom.Store(t + 1)
	return Task{Index: int(idx)}, true
}

// Steal removes a task from the thief end (FIFO). Safe to call from any
// goroutine that is not the owner. Returns (emptyTask, false) if the
// deque appears empty or another thief won the CAS.
func (d *Deque) Steal() (Task, bool) {
	t := d.top.Load()
	// Sequentially consistent load on bottom orders this read after the
	// top read - the fence the algorithm needs.
	b := d.bottom.Load()
	if t >= b {
		return emptyTask, false
	}
	idx := d.buf[t&d.mask].Load()
	if !d.top.CompareAndSwap(t, t+1) {
		return emptyTask, false
	}
	return Task{Index: int(idx)}, true
}

// Len returns an approximate count of remaining elements. The value may
// be stale by the time the caller reads it; useful for diagnostics only.
func (d *Deque) Len() int {
	b := d.bottom.Load()
	t := d.top.Load()
	if b < t {
		return 0
	}
	return int(b - t)
}
