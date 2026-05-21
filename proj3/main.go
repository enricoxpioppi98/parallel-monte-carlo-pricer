// Command proj3 is the CLI front-end for the parallel Monte Carlo
// option pricer. It exposes two subcommands:
//
//	proj3 gen     - generate a portfolio JSON file (balanced or unbalanced)
//	proj3 price   - price a portfolio with seq / par / steal runners,
//	                appending one timing row to a CSV file
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"proj3/portfolio"
	"proj3/runner"
)

const usageText = `Parallel Monte Carlo option pricer.

Usage:
  proj3 gen   --kind {balanced|unbalanced} --n N --out FILE [--seed S]
  proj3 price --mode {seq|par|steal} --in FILE --out CSV [--threads T] [--seed S]

Commands:
  gen      Generate a portfolio dataset.
  price    Price a portfolio. Appends one row to OUT:
           mode,threads,dataset,seed,elapsed_ms,portfolio_value

Examples:
  proj3 gen --kind=unbalanced --n=200 --out=data/unbalanced.json
  proj3 price --mode=seq   --in=data/balanced.json   --threads=1 --out=results/timings.csv
  proj3 price --mode=par   --in=data/unbalanced.json --threads=8 --out=results/timings.csv
  proj3 price --mode=steal --in=data/unbalanced.json --threads=8 --out=results/timings.csv
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "gen":
		cmdGen(os.Args[2:])
	case "price":
		cmdPrice(os.Args[2:])
	case "-h", "--help", "help":
		fmt.Print(usageText)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(2)
	}
}

func cmdGen(args []string) {
	fs := flag.NewFlagSet("gen", flag.ExitOnError)
	kind := fs.String("kind", "balanced", "balanced | unbalanced")
	n := fs.Int("n", 200, "number of options")
	out := fs.String("out", "", "output JSON path")
	seed := fs.Int64("seed", 42, "RNG seed for option parameter generation")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if *out == "" {
		fmt.Fprintln(os.Stderr, "gen: --out is required")
		os.Exit(2)
	}
	p, err := portfolio.Generate(*n, *kind, *seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gen: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "gen: mkdir: %v\n", err)
		os.Exit(1)
	}
	if err := portfolio.Save(*out, p); err != nil {
		fmt.Fprintf(os.Stderr, "gen: save: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %d %s options to %s\n", len(p.Options), *kind, *out)
}

func cmdPrice(args []string) {
	fs := flag.NewFlagSet("price", flag.ExitOnError)
	mode := fs.String("mode", "seq", "seq | par | steal")
	in := fs.String("in", "", "input portfolio JSON path")
	outCSV := fs.String("out", "", "output CSV path (appended to)")
	threads := fs.Int("threads", runtime.NumCPU(), "number of worker threads (par/steal only)")
	seed := fs.Uint64("seed", 42, "RNG seed for pricing")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if *in == "" || *outCSV == "" {
		fmt.Fprintln(os.Stderr, "price: --in and --out are required")
		os.Exit(2)
	}

	p, err := portfolio.Load(*in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "price: %v\n", err)
		os.Exit(1)
	}

	// Sequential always runs single-threaded; record threads=1 in the CSV
	// so plotting scripts can key off (mode, dataset, threads) cleanly.
	threadCount := *threads
	if *mode == "seq" {
		threadCount = 1
	}

	cfg := runner.NewConfig(p, threadCount, *seed)

	start := time.Now()
	switch *mode {
	case "seq":
		runner.RunSequential(cfg)
	case "par":
		runner.RunMapReduce(cfg)
	case "steal":
		runner.RunWorkStealing(cfg)
	default:
		fmt.Fprintf(os.Stderr, "price: unknown mode %q\n", *mode)
		os.Exit(2)
	}
	elapsed := time.Since(start)
	total := cfg.PortfolioValue()

	if err := os.MkdirAll(filepath.Dir(*outCSV), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "price: mkdir: %v\n", err)
		os.Exit(1)
	}
	if err := appendCSV(*outCSV, []string{
		*mode,
		strconv.Itoa(threadCount),
		p.Name,
		strconv.FormatUint(*seed, 10),
		strconv.FormatInt(elapsed.Milliseconds(), 10),
		strconv.FormatFloat(total, 'f', 4, 64),
	}); err != nil {
		fmt.Fprintf(os.Stderr, "price: csv: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("mode=%s threads=%d dataset=%s elapsed=%dms total=%.4f\n",
		*mode, threadCount, p.Name, elapsed.Milliseconds(), total)
}

func appendCSV(path string, row []string) error {
	// Header is needed when the file doesn't exist yet OR when it exists
	// but is empty (e.g. truncated by `: > file.csv` in the run script).
	info, statErr := os.Stat(path)
	needsHeader := os.IsNotExist(statErr) || (statErr == nil && info.Size() == 0)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if needsHeader {
		if err := w.Write([]string{"mode", "threads", "dataset", "seed", "elapsed_ms", "portfolio_value"}); err != nil {
			return err
		}
	}
	if err := w.Write(row); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}
