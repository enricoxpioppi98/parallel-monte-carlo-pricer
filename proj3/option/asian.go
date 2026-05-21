package option

import "math"

// priceAsian prices an arithmetic-average Asian call. The payoff at maturity
// is max(mean(S_t) - K, 0), so each path must be walked step-by-step.
func priceAsian(s Spec, rng *XorShift64) Result {
	dt := s.Maturity / float64(s.Steps)
	drift := (s.Rate - 0.5*s.Vol*s.Vol) * dt
	diff := s.Vol * math.Sqrt(dt)
	var sum, sumSq float64
	for i := 0; i < s.Paths; i++ {
		st := s.Spot
		avg := 0.0
		for j := 0; j < s.Steps; j++ {
			z := rng.NormFloat64()
			st = st * math.Exp(drift+diff*z)
			avg += st
		}
		avg /= float64(s.Steps)
		payoff := math.Max(avg-s.Strike, 0)
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
