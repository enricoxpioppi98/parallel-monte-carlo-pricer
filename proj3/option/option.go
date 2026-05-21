// Package option defines the Spec / Result types and Monte Carlo pricers
// for European, Asian, and American options under geometric Brownian motion.
package option

// Kind enumerates the supported option types.
type Kind string

const (
	KindEuropean Kind = "european"
	KindAsian    Kind = "asian"
	KindAmerican Kind = "american"
)

// Spec describes a single option to price.
type Spec struct {
	ID       int     `json:"id"`
	Kind     Kind    `json:"kind"`
	Spot     float64 `json:"spot"`     // S_0
	Strike   float64 `json:"strike"`   // K
	Rate     float64 `json:"rate"`     // r, risk-free rate
	Vol      float64 `json:"vol"`      // sigma, annualized volatility
	Maturity float64 `json:"maturity"` // T, in years
	Paths    int     `json:"paths"`
	Steps    int     `json:"steps"` // 1 is fine for European
}

// Result holds the pricer output for one option.
type Result struct {
	ID     int     `json:"id"`
	Price  float64 `json:"price"`
	StdErr float64 `json:"std_err"`
}

// Price dispatches to the type-specific Monte Carlo pricer.
func Price(spec Spec, rng *XorShift64) Result {
	switch spec.Kind {
	case KindEuropean:
		return priceEuropean(spec, rng)
	case KindAsian:
		return priceAsian(spec, rng)
	case KindAmerican:
		return priceAmerican(spec, rng)
	default:
		panic("option: unknown kind " + string(spec.Kind))
	}
}
