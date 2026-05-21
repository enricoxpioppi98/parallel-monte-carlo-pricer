package option

import "math"

// XorShift64 is a per-task xorshift64* PRNG. It is not safe for concurrent
// use; the runners build one RNG per task so workers never share state.
type XorShift64 struct {
	state     uint64
	cached    float64
	hasCached bool
}

// NewRNG derives a per-task seed deterministically from (seed, taskIndex).
// Using the same seed across runners therefore yields the same prices
// regardless of mode or thread count, which makes correctness checks
// trivial. The 0x9E37... multiplier is 2^64 / phi (the "golden ratio
// hash" constant from Knuth), chosen because it disperses sequential
// taskIndex values across the 64-bit state space.
func NewRNG(seed uint64, taskIndex int) *XorShift64 {
	s := seed ^ (uint64(taskIndex+1) * 0x9E3779B97F4A7C15)
	if s == 0 {
		s = 1
	}
	return &XorShift64{state: s}
}

// Uint64 returns the next pseudo-random uint64 using Vigna's xorshift64*
// (xorshift64 followed by multiplication by an odd constant; see Vigna,
// "An experimental exploration of Marsaglia's xorshift generators,
// scrambled", 2014). The shift triple (13, 7, 17) is the canonical
// xorshift64 set; 0x2545F4914F6CDD1D is Vigna's recommended scrambler.
func (r *XorShift64) Uint64() uint64 {
	x := r.state
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	r.state = x
	return x * 0x2545F4914F6CDD1D
}

func (r *XorShift64) Float64() float64 {
	return float64(r.Uint64()>>11) / (1 << 53)
}

// NormFloat64 returns a standard normal sample via Box-Muller, caching the
// second sample of each pair to halve the math.Log/Sqrt cost.
func (r *XorShift64) NormFloat64() float64 {
	if r.hasCached {
		r.hasCached = false
		return r.cached
	}
	var u1 float64
	for {
		u1 = r.Float64()
		if u1 > 1e-300 {
			break
		}
	}
	u2 := r.Float64()
	mag := math.Sqrt(-2.0 * math.Log(u1))
	r.cached = mag * math.Sin(2*math.Pi*u2)
	r.hasCached = true
	return mag * math.Cos(2*math.Pi*u2)
}
