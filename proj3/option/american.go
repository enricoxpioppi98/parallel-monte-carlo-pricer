package option

import "math"

// priceAmerican prices an American put via Longstaff-Schwartz Monte Carlo.
// All P paths are simulated forward; backward induction then regresses the
// discounted continuation value on a quadratic basis {1, S, S^2} at each
// in-the-money node and updates the optimal exercise decision per path.
// This is the heaviest pricer in the system by a wide margin and is what
// drives the load imbalance the work-stealing runner has to recover from.
//
// Two micro-optimisations matter here, since this function dominates the
// unbalanced workload:
//
//  1. A single flat []float64 buffer is used for all paths (indexed as
//     flat[i*(N+1)+t]) instead of [][]float64, cutting allocations from
//     P+1 to 2 per call. Reduces GC pressure under high T concurrency.
//  2. Discount powers discountStep^k are precomputed once for k in 0..N,
//     replacing the math.Pow calls inside the inner loops with array
//     lookups.
func priceAmerican(s Spec, rng *XorShift64) Result {
	N := s.Steps
	P := s.Paths
	dt := s.Maturity / float64(N)
	drift := (s.Rate - 0.5*s.Vol*s.Vol) * dt
	diff := s.Vol * math.Sqrt(dt)
	discountStep := math.Exp(-s.Rate * dt)

	// Precompute discountStep^k for k = 0..N so the inner-loop discount
	// factors are array lookups instead of math.Pow calls.
	discPow := make([]float64, N+1)
	discPow[0] = 1.0
	for k := 1; k <= N; k++ {
		discPow[k] = discPow[k-1] * discountStep
	}

	// Flat path matrix: path i at step t is flat[i*(N+1)+t].
	stride := N + 1
	flat := make([]float64, P*stride)
	for i := 0; i < P; i++ {
		base := i * stride
		flat[base] = s.Spot
		for t := 1; t <= N; t++ {
			z := rng.NormFloat64()
			flat[base+t] = flat[base+t-1] * math.Exp(drift+diff*z)
		}
	}

	cf := make([]float64, P)
	exerciseStep := make([]int, P)
	for i := 0; i < P; i++ {
		cf[i] = math.Max(s.Strike-flat[i*stride+N], 0)
		exerciseStep[i] = N
	}

	itmIdx := make([]int, 0, P)
	for t := N - 1; t >= 1; t-- {
		itmIdx = itmIdx[:0]
		for i := 0; i < P; i++ {
			if s.Strike-flat[i*stride+t] > 0 {
				itmIdx = append(itmIdx, i)
			}
		}
		if len(itmIdx) < 3 {
			continue
		}

		// Build the normal equations X^T X b = X^T y for X = [1, S, S^2].
		var m00, m01, m02, m11, m12, m22 float64
		var y0, y1, y2 float64
		for _, i := range itmIdx {
			S := flat[i*stride+t]
			df := discPow[exerciseStep[i]-t]
			y := cf[i] * df
			S2 := S * S
			m00 += 1
			m01 += S
			m02 += S2
			m11 += S2
			m12 += S2 * S
			m22 += S2 * S2
			y0 += y
			y1 += y * S
			y2 += y * S2
		}
		a, b, c, ok := solveSym3x3(m00, m01, m02, m11, m12, m22, y0, y1, y2)
		if !ok {
			continue
		}

		for _, i := range itmIdx {
			S := flat[i*stride+t]
			continuation := a + b*S + c*S*S
			immediate := s.Strike - S
			if immediate > continuation {
				cf[i] = immediate
				exerciseStep[i] = t
			}
		}
	}

	var sum, sumSq float64
	for i := 0; i < P; i++ {
		df := discPow[exerciseStep[i]]
		v := cf[i] * df
		sum += v
		sumSq += v * v
	}
	n := float64(P)
	mean := sum / n
	variance := sumSq/n - mean*mean
	if variance < 0 {
		variance = 0
	}
	return Result{
		ID:     s.ID,
		Price:  mean,
		StdErr: math.Sqrt(variance / n),
	}
}

// solveSym3x3 solves the symmetric 3x3 system M b = y by Cramer's rule, where
//
//	M = [[m00, m01, m02], [m01, m11, m12], [m02, m12, m22]]
//	y = [y0, y1, y2]
//
// Returns (b0, b1, b2, true) on success, or (0, 0, 0, false) if the matrix
// is near-singular (which can happen if there are too few distinct S values).
func solveSym3x3(m00, m01, m02, m11, m12, m22, y0, y1, y2 float64) (float64, float64, float64, bool) {
	det := m00*(m11*m22-m12*m12) - m01*(m01*m22-m12*m02) + m02*(m01*m12-m11*m02)
	if math.Abs(det) < 1e-12 {
		return 0, 0, 0, false
	}
	d0 := y0*(m11*m22-m12*m12) - m01*(y1*m22-m12*y2) + m02*(y1*m12-m11*y2)
	d1 := m00*(y1*m22-m12*y2) - y0*(m01*m22-m12*m02) + m02*(m01*y2-y1*m02)
	d2 := m00*(m11*y2-y1*m12) - m01*(m01*y2-y1*m02) + y0*(m01*m12-m11*m02)
	return d0 / det, d1 / det, d2 / det, true
}
