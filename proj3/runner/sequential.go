package runner

// RunSequential prices every option in a single goroutine, in input order.
// It is the reference implementation: every parallel runner must produce
// the same per-option prices (within float tolerance) when given the same
// seed.
func RunSequential(c *Config) {
	for i := range c.Portfolio.Options {
		runOne(c, i)
	}
}
