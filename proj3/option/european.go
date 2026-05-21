package option

import "math"

// priceEuropean prices a European call via direct sampling of S_T under GBM.
// No path is materialised: S_T = S_0 * exp((r - 0.5 sigma^2) T + sigma sqrt(T) Z).
func priceEuropean(s Spec, rng *XorShift64) Result {
	drift := (s.Rate - 0.5*s.Vol*s.Vol) * s.Maturity
	diff := s.Vol * math.Sqrt(s.Maturity)
	var sum, sumSq float64
	for i := 0; i < s.Paths; i++ {
		z := rng.NormFloat64()
		sT := s.Spot * math.Exp(drift+diff*z)
		payoff := math.Max(sT-s.Strike, 0)
		sum += payoff
		sumSq += payoff * payoff
	}
	n := float64(s.Paths)
	mean := sum / n
	variance := sumSq/n - mean*mean
	if variance < 0 {
		variance = 0
	}
	discount := math.Exp(-s.Rate * s.Maturity)
	return Result{
		ID:     s.ID,
		Price:  discount * mean,
		StdErr: discount * math.Sqrt(variance/n),
	}
}
