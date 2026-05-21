# proj3 - Parallel Monte Carlo Option Pricer

Three runners that price the same portfolio of exotic options
(European, Asian, American) and let you compare:

  - **`seq`**    - single-goroutine baseline
  - **`par`**    - static-partition map-reduce with a `sync.Cond` barrier
  - **`steal`**  - work-stealing across per-worker Chase-Lev lock-free deques

See `REPORT.md` for the full write-up.

## Quick start

```bash
# Build, generate datasets, run the full sweep, produce plots.
./scripts/run.sh
```

The script writes timing rows to `results/timings.csv` and PNGs to
`plots/`. The whole sweep finishes in well under a minute locally.

## Running by hand

```bash
go build -o proj3 .

# Generate a portfolio
./proj3 gen --kind=unbalanced --n=200 --out=data/unbalanced.json

# Price it
./proj3 price --mode=seq   --in=data/unbalanced.json --threads=1 --out=results/timings.csv
./proj3 price --mode=par   --in=data/unbalanced.json --threads=8 --out=results/timings.csv
./proj3 price --mode=steal --in=data/unbalanced.json --threads=8 --out=results/timings.csv

# Plot
python3 scripts/plot.py
```

## Cluster experiments (Peanut, the graded path)

Follow the course's **Project Cluster Notes** literally:

```bash
# 1. On fe.ai.cs.uchicago.edu, install Go 1.19.13 in $HOME (one time):
cd ~ && wget -qO- https://go.dev/dl/go1.19.13.linux-amd64.tar.gz | tar xz

# 2. Clone this repo into your home directory.

# 3. Configure the sbatch script:
cd path/to/proj3/benchmark
#    edit benchmark-proj3.sh:
#      --mail-user=YOUR_CNETID@cs.uchicago.edu
#      --chdir=$(pwd)        # paste the absolute path printed by `pwd`
mkdir -p slurm/out            # for sbatch stdout/stderr logs

# 4. Submit:
sbatch benchmark-proj3.sh
```

The job writes timing rows to `proj3/results/timings.csv` and sbatch's
own logs to `proj3/benchmark/slurm/out/<jobid>.{out,err}`. Plotting
happens off-cluster (the compute nodes may not have matplotlib):

```bash
# Back on any machine with matplotlib:
python3 scripts/plot.py
```

The numbers in `REPORT.md` come from a five-seed `peanut-cpu` run via
the same `cluster_experiments.sh` that `benchmark-proj3.sh` invokes.

## CLI reference

```
proj3 gen   --kind {balanced|unbalanced} --n N --out FILE [--seed S]
proj3 price --mode {seq|par|steal} --in FILE --out CSV [--threads T] [--seed S]
```

`price` appends one row to its `--out` CSV:

```
mode,threads,dataset,seed,elapsed_ms,portfolio_value
```

## Tests

```bash
go test -race ./...
```

Two test files ship with the project:

  - `deque/deque_test.go` - stress test exercising one owner against
    several thieves over 50,000 tasks; uses `-race` to catch any
    forgotten atomic.
  - `option/european_test.go` - Black-Scholes parity (the Monte Carlo
    European pricer must agree with the closed-form value within 3
    standard errors) plus an RNG determinism test that guarantees
    the seq / par / steal modes price every option identically.

## Directory layout

```
proj3/
  main.go              CLI entry
  option/              GBM, RNG, European/Asian/American pricers (+ tests)
  portfolio/           JSON load/save, dataset generator
  barrier/             Reusable sync.Cond barrier
  deque/               Chase-Lev lock-free deque (+ -race stress test)
  runner/              sequential, map-reduce, work-stealing
  scripts/             run.sh, experiments.sh, cluster_experiments.sh, plot.py
  benchmark/           Sbatch script + Slurm log directory (peanut-cpu)
  data/                Generated portfolios
  results/             Timing CSVs
  plots/               Generated PNGs
  REPORT.md            Full write-up
```

## Requirements

  - Go 1.19+
  - Python 3 with `matplotlib` for plotting
