package portfolio

import (
	"fmt"
	"math/rand"

	"proj3/option"
)

// Generate builds a portfolio of n options according to the given kind:
//
//   - "balanced":   100% European calls. All options are roughly the same
//     compute cost, so static partitioning should already
//     give near-ideal speedup.
//   - "unbalanced": 70% European, 20% Asian, 10% American.  Asian and
//     American options are ~100x more expensive than European,
//     so a static-partition runner gets bottlenecked on whatever
//     workers happened to receive the heavy options.  This is the
//     scenario where the work-stealing runner should win.
//
// Options are placed in the slice grouped by type (Europeans first, then
// Asians, then Americans) so contiguous-chunk static partitioning bunches
// the heavy options on the last few workers. This is the worst case for
// static partitioning and the strongest case for work-stealing - exactly
// the contrast the writeup leans on.
func Generate(n int, kind string, seed int64) (Portfolio, error) {
	r := rand.New(rand.NewSource(seed))
	opts := make([]option.Spec, 0, n)

	switch kind {
	case "balanced":
		for i := 0; i < n; i++ {
			opts = append(opts, mkEuropean(i, r))
		}
	case "unbalanced":
		nEur := (7 * n) / 10
		nAsian := (2 * n) / 10
		nAmer := n - nEur - nAsian
		for i := 0; i < nEur; i++ {
			opts = append(opts, mkEuropean(len(opts), r))
		}
		for i := 0; i < nAsian; i++ {
			opts = append(opts, mkAsian(len(opts), r))
		}
		for i := 0; i < nAmer; i++ {
			opts = append(opts, mkAmerican(len(opts), r))
		}
	default:
		return Portfolio{}, fmt.Errorf("unknown portfolio kind: %s", kind)
	}
	return Portfolio{Name: kind, Options: opts}, nil
}

func mkEuropean(id int, r *rand.Rand) option.Spec {
	return option.Spec{
		ID:       id,
		Kind:     option.KindEuropean,
		Spot:     100.0,
		Strike:   80.0 + 40.0*r.Float64(),
		Rate:     0.05,
		Vol:      0.20 + 0.10*r.Float64(),
		Maturity: 0.5 + 1.5*r.Float64(),
		Paths:    20000,
		Steps:    1,
	}
}

func mkAsian(id int, r *rand.Rand) option.Spec {
	return option.Spec{
		ID:       id,
		Kind:     option.KindAsian,
		Spot:     100.0,
		Strike:   80.0 + 40.0*r.Float64(),
		Rate:     0.05,
		Vol:      0.20 + 0.10*r.Float64(),
		Maturity: 0.5 + 1.5*r.Float64(),
		Paths:    20000,
		Steps:    100,
	}
}

func mkAmerican(id int, r *rand.Rand) option.Spec {
	return option.Spec{
		ID:       id,
		Kind:     option.KindAmerican,
		Spot:     100.0,
		Strike:   80.0 + 40.0*r.Float64(),
		Rate:     0.05,
		Vol:      0.20 + 0.10*r.Float64(),
		Maturity: 0.5 + 1.5*r.Float64(),
		Paths:    20000,
		Steps:    50,
	}
}
