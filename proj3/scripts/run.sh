#!/usr/bin/env bash
# One-command runner: build the binary, generate both datasets, run the
# full sweep (seq + par/steal across the assignment's required thread
# counts), and produce plots. This is the script graders run.
#
# Usage:  ./scripts/run.sh
# Output: results/timings.csv, plots/*.png

set -euo pipefail
cd "$(dirname "$0")/.."

mkdir -p data results plots

echo "==> Building..."
go build -o proj3 .

echo "==> Generating datasets..."
./proj3 gen --kind=balanced   --n=200 --seed=42 --out=data/balanced.json
./proj3 gen --kind=unbalanced --n=200 --seed=42 --out=data/unbalanced.json

# Reset timings file so plot.py sees only this sweep.
: > results/timings.csv

echo "==> Sequential baselines..."
./proj3 price --mode=seq --in=data/balanced.json   --threads=1 --out=results/timings.csv
./proj3 price --mode=seq --in=data/unbalanced.json --threads=1 --out=results/timings.csv

echo "==> Parallel sweep (T = 2 4 6 8 12)..."
for T in 2 4 6 8 12; do
  for DATA in balanced unbalanced; do
    ./proj3 price --mode=par   --in=data/$DATA.json --threads=$T --out=results/timings.csv
    ./proj3 price --mode=steal --in=data/$DATA.json --threads=$T --out=results/timings.csv
  done
done

echo "==> Plotting..."
python3 scripts/plot.py

echo "Done. Plots in plots/ ; timings in results/timings.csv"
