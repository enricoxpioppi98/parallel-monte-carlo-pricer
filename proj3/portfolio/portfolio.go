// Package portfolio handles JSON load/save and dataset generation for a
// collection of options that the runners price as a single batch.
package portfolio

import (
	"encoding/json"
	"fmt"
	"os"

	"proj3/option"
)

// Portfolio is a named collection of options.
type Portfolio struct {
	Name    string        `json:"name"`
	Options []option.Spec `json:"options"`
}

// Load reads a portfolio JSON file from disk.
func Load(path string) (Portfolio, error) {
	f, err := os.Open(path)
	if err != nil {
		return Portfolio{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	var p Portfolio
	if err := json.NewDecoder(f).Decode(&p); err != nil {
		return Portfolio{}, fmt.Errorf("decode %s: %w", path, err)
	}
	return p, nil
}

// Save writes a portfolio to disk as pretty-printed JSON.
func Save(path string, p Portfolio) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}
