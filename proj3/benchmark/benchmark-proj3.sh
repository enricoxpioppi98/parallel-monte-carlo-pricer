#!/usr/bin/env bash
# Sbatch job script for the Peanut CPU partition, following the project
# cluster notes verbatim:
#   - Run on partition peanut-cpu
#   - Use a Go install in $HOME/go (set up via the wget command in the
#     cluster notes), NOT `module load`
#   - Output / error logs land in benchmark/slurm/out/
#
# Usage (on fe.ai.cs.uchicago.edu):
#   cd path/to/proj3/benchmark
#   # 1. Edit the two TODO lines below: --mail-user and --chdir
#   # 2. Make sure $HOME/go/bin/go exists (see the cluster notes setup step)
#   # 3. Submit:
#   sbatch benchmark-proj3.sh
#
# After completion the timing rows land in proj3/results/timings.csv and the
# raw sbatch output in proj3/benchmark/slurm/out/<jobid>.out.

#SBATCH --job-name=proj3-bench
#SBATCH --partition=peanut-cpu
#SBATCH --cpus-per-task=16
#SBATCH --time=00:30:00
#SBATCH --mail-type=END,FAIL
#SBATCH --mail-user=__EDIT_ME__CNETID@cs.uchicago.edu
#SBATCH --chdir=__EDIT_ME__ABSOLUTE_PATH_TO_PROJ3_BENCHMARK_DIR
#SBATCH --output=slurm/out/%j.out
#SBATCH --error=slurm/out/%j.err

set -euo pipefail

# Fail loud if the two TODO placeholders above weren't filled in. These
# values are syntactically valid for sbatch (so the directives parse) but
# the script body refuses to proceed until the grader has edited them.
if [[ "${SLURM_SUBMIT_DIR:-}" == *__EDIT_ME__* ]] || \
   grep -q '__EDIT_ME__' "$0" 2>/dev/null; then
  echo "ERROR: edit --mail-user and --chdir at the top of $(basename "$0") first." >&2
  echo "       Both currently contain the placeholder marker '__EDIT_ME__'." >&2
  exit 2
fi

# Cluster notes are explicit: do NOT `module load`; use the $HOME/go
# install that the setup step put in place.
export PATH=$HOME/go/bin:$PATH

echo "===================================="
echo "host       : $(hostname)"
echo "go version : $(go version 2>/dev/null || echo 'MISSING - install via cluster-notes wget step')"
echo "cwd        : $(pwd)"
echo "date       : $(date)"
echo "===================================="

# This script lives in proj3/benchmark/. The sweep script lives in
# proj3/scripts/. Step up one directory so the relative paths inside
# cluster_experiments.sh (data/, results/, etc.) resolve.
cd ..

./scripts/cluster_experiments.sh
