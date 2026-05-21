#!/usr/bin/env bash
# Cluster sweep: same as run.sh but with multiple seeds so the speedup
# plots can average across runs and reduce timing noise. Intended for
# the Peanut cluster.
#
# Usage:  ./scripts/experiments.sh

set -euo pipefail
cd "$(dirname "$0")/.."

mkdir -p data results plots

echo "==> Building..."
go build -o proj3 .

echo "==> Generating datasets..."
./proj3 gen --kind=balanced   --n=200 --seed=42 --out=data/balanced.json
./proj3 gen --kind=unbalanced --n=200 --seed=42 --out=data/unbalanced.json

: > results/timings.csv

SEEDS=(42 100 7 1234 9999)

for SEED in "${SEEDS[@]}"; do
  echo "==> Seed $SEED ..."
  for DATA in balanced unbalanced; do
    ./proj3 price --mode=seq --in=data/$DATA.json --threads=1 --seed=$SEED --out=results/timings.csv
  done
  for T in 2 4 6 8 12; do
    for DATA in balanced unbalanced; do
      ./proj3 price --mode=par   --in=data/$DATA.json --threads=$T --seed=$SEED --out=results/timings.csv
      ./proj3 price --mode=steal --in=data/$DATA.json --threads=$T --seed=$SEED --out=results/timings.csv
    done
  done
done

echo "==> Plotting..."
python3 scripts/plot.py

echo "Done. ${#SEEDS[@]} seeds; plots in plots/ ; raw timings in results/timings.csv"
