package deque

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func TestPushPopLIFO(t *testing.T) {
	d := New(16)
	for i := 0; i < 10; i++ {
		d.PushBottom(Task{Index: i})
	}
	for i := 9; i >= 0; i-- {
		task, ok := d.PopBottom()
		if !ok || task.Index != i {
			t.Fatalf("expected %d, got %+v ok=%v", i, task, ok)
		}
	}
	if _, ok := d.PopBottom(); ok {
		t.Fatal("expected empty after draining")
	}
}

func TestStealFIFO(t *testing.T) {
	d := New(16)
	for i := 0; i < 5; i++ {
		d.PushBottom(Task{Index: i})
	}
	for i := 0; i < 5; i++ {
		task, ok := d.Steal()
		if !ok || task.Index != i {
			t.Fatalf("expected %d, got %+v ok=%v", i, task, ok)
		}
	}
	if _, ok := d.Steal(); ok {
		t.Fatal("expected empty after draining")
	}
}

func TestPopOnEmpty(t *testing.T) {
	d := New(8)
	if _, ok := d.PopBottom(); ok {
		t.Fatal("expected empty")
	}
	if _, ok := d.Steal(); ok {
		t.Fatal("expected empty")
	}
}

// TestConcurrentOwnerAndThieves stresses the deque with the owner popping
// from bottom while several thieves steal from top. Every task must be
// consumed exactly once.
func TestConcurrentOwnerAndThieves(t *testing.T) {
	const (
		N       = 50000
		Thieves = 4
	)
	cap := 1
	for cap < N {
		cap <<= 1
	}
	d := New(cap)
	for i := 0; i < N; i++ {
		d.PushBottom(Task{Index: i})
	}

	seen := make([]atomic.Bool, N)
	var consumed atomic.Int64
	var wg sync.WaitGroup
	wg.Add(Thieves + 1)

	// Thieves.
	for i := 0; i < Thieves; i++ {
		go func() {
			defer wg.Done()
			for {
				task, ok := d.Steal()
				if !ok {
					if d.Len() == 0 {
						return
					}
					runtime.Gosched()
					continue
				}
				if seen[task.Index].Swap(true) {
					t.Errorf("task %d stolen twice", task.Index)
					return
				}
				consumed.Add(1)
			}
		}()
	}

	// Owner.
	go func() {
		defer wg.Done()
		for {
			task, ok := d.PopBottom()
			if !ok {
				if d.Len() == 0 {
					return
				}
				runtime.Gosched()
				continue
			}
			if seen[task.Index].Swap(true) {
				t.Errorf("task %d popped twice", task.Index)
				return
			}
			consumed.Add(1)
		}
	}()

	wg.Wait()
	if got := consumed.Load(); got != N {
		t.Fatalf("expected %d consumed, got %d", N, got)
	}
	for i := 0; i < N; i++ {
		if !seen[i].Load() {
			t.Fatalf("task %d never consumed", i)
		}
	}
}
