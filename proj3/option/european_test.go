package option

import (
	"math"
	"testing"
)

// blackScholesCall is the closed-form Black-Scholes price of a European
// call on a non-dividend asset. Used here as an analytic reference to
// sanity-check the Monte Carlo pricer.
func blackScholesCall(s, k, r, sigma, t float64) float64 {
	d1 := (math.Log(s/k) + (r+0.5*sigma*sigma)*t) / (sigma * math.Sqrt(t))
	d2 := d1 - sigma*math.Sqrt(t)
	nd1 := 0.5 * (1 + math.Erf(d1/math.Sqrt2))
	nd2 := 0.5 * (1 + math.Erf(d2/math.Sqrt2))
	return s*nd1 - k*math.Exp(-r*t)*nd2
}

// TestEuropeanAgainstBlackScholes prices an at-the-money European call
// by Monte Carlo and checks it lies within 3 standard errors of the
// closed-form value - a 99.7% confidence interval for a correct pricer.
func TestEuropeanAgainstBlackScholes(t *testing.T) {
	spec := Spec{
		ID:       0,
		Kind:     KindEuropean,
		Spot:     100,
		Strike:   100,
		Rate:     0.05,
		Vol:      0.2,
		Maturity: 1.0,
		Paths:    50_000,
		Steps:    1,
	}
	rng := NewRNG(42, 0)
	got := Price(spec, rng)
	want := blackScholesCall(spec.Spot, spec.Strike, spec.Rate, spec.Vol, spec.Maturity)
	tol := 3 * got.StdErr
	if math.Abs(got.Price-want) > tol {
		t.Fatalf("priceEuropean = %.4f, Black-Scholes = %.4f, |diff| = %.4f > 3*SE = %.4f",
			got.Price, want, math.Abs(got.Price-want), tol)
	}
}

// TestRNGDeterminism asserts that the (seed, taskIndex) -> RNG mapping
// is deterministic, which is what makes pricing reproducible across
// runners. Two RNGs built from the same arguments must produce
// identical draws.
func TestRNGDeterminism(t *testing.T) {
	a := NewRNG(12345, 7)
	b := NewRNG(12345, 7)
	for i := 0; i < 1000; i++ {
		if x, y := a.Uint64(), b.Uint64(); x != y {
			t.Fatalf("rng diverged at draw %d: %d vs %d", i, x, y)
		}
	}
}
