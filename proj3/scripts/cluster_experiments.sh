#!/usr/bin/env bash
# Cluster sweep meant to run on fe.ai.cs.uchicago.edu's peanut-cpu
# partition via srun/sbatch. Same workload as experiments.sh but skips
# the local matplotlib plotting step (we plot on the dev machine after
# pulling results/timings.csv back).
#
# Usage:  srun --partition=peanut-cpu --cpus-per-task=16 --time=00:30:00 \
#             ./scripts/cluster_experiments.sh

set -euo pipefail
cd "$(dirname "$0")/.."

mkdir -p data results

echo "==> Build host: $(hostname)  Go: $(go version)"

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

echo "==> Done. ${#SEEDS[@]} seeds; timings in results/timings.csv"
wc -l results/timings.csv
