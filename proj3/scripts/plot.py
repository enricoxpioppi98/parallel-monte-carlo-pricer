#!/usr/bin/env python3
"""Generate speedup and runtime plots from results/timings.csv.

Reads the CSV the Go binary appends to and writes PNGs into plots/.
If the CSV contains multiple seeds per (mode, dataset, threads) tuple,
runtimes are averaged before computing speedup.
"""

import csv
import os
import sys
from collections import defaultdict

try:
    import matplotlib.pyplot as plt
except ImportError:
    sys.stderr.write(
        "matplotlib not installed. Install with: python3 -m pip install matplotlib\n"
    )
    sys.exit(1)


ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
TIMINGS = os.path.join(ROOT, "results", "timings.csv")
PLOTS = os.path.join(ROOT, "plots")
os.makedirs(PLOTS, exist_ok=True)


def load_timings():
    """Return {(mode, dataset, threads): mean_elapsed_ms}."""
    runs = defaultdict(list)
    with open(TIMINGS) as f:
        reader = csv.DictReader(f)
        for row in reader:
            key = (row["mode"], row["dataset"], int(row["threads"]))
            runs[key].append(int(row["elapsed_ms"]))
    return {k: sum(v) / len(v) for k, v in runs.items()}


def plot_speedup(times, dataset, outpath):
    base = times.get(("seq", dataset, 1))
    if base is None:
        print(f"warn: no seq baseline for {dataset}; skipping")
        return
    thread_counts = sorted(
        t
        for (m, d, t) in times
        if d == dataset and m in ("par", "steal")
    )
    if not thread_counts:
        return
    par = [base / times[("par", dataset, T)] for T in thread_counts]
    steal = [base / times[("steal", dataset, T)] for T in thread_counts]

    plt.figure(figsize=(7, 5))
    plt.plot(thread_counts, par, "o-", label="Map-reduce (static partition)")
    plt.plot(thread_counts, steal, "s-", label="Work-stealing (Chase-Lev deque)")
    plt.plot(thread_counts, thread_counts, "k--", alpha=0.35, label="Ideal speedup")
    plt.xlabel("Threads")
    plt.ylabel("Speedup vs sequential")
    plt.title(f"Speedup - {dataset} portfolio")
    plt.grid(alpha=0.3)
    plt.legend()
    plt.tight_layout()
    plt.savefig(outpath, dpi=140)
    plt.close()
    print(f"wrote {outpath}")


def plot_runtime_compare(times, outpath):
    """Bar chart contrasting balanced vs unbalanced sequential runtime."""
    datasets = ["balanced", "unbalanced"]
    seq_times = [times.get(("seq", d, 1), 0) / 1000.0 for d in datasets]
    plt.figure(figsize=(6, 4))
    bars = plt.bar(datasets, seq_times, color=["#5b9bd5", "#ed7d31"])
    plt.ylabel("Sequential runtime (s)")
    plt.title("Sequential runtime: balanced vs unbalanced")
    for bar, v in zip(bars, seq_times):
        plt.text(
            bar.get_x() + bar.get_width() / 2,
            v,
            f"{v:.2f}s",
            ha="center",
            va="bottom",
        )
    plt.tight_layout()
    plt.savefig(outpath, dpi=140)
    plt.close()
    print(f"wrote {outpath}")


def plot_unbalanced_zoom(times, outpath):
    """Bar chart of elapsed time at T=8 across the three modes on the
    unbalanced portfolio - the most useful single-glance summary."""
    modes = ["seq", "par", "steal"]
    labels = ["Sequential", "Map-reduce (T=8)", "Work-stealing (T=8)"]
    vals = []
    for m in modes:
        T = 1 if m == "seq" else 8
        ms = times.get((m, "unbalanced", T))
        vals.append((ms or 0) / 1000.0)
    plt.figure(figsize=(7, 4.5))
    bars = plt.bar(labels, vals, color=["#999999", "#5b9bd5", "#70ad47"])
    plt.ylabel("Elapsed (s)")
    plt.title("Unbalanced portfolio - elapsed time at T=8")
    for bar, v in zip(bars, vals):
        plt.text(
            bar.get_x() + bar.get_width() / 2,
            v,
            f"{v:.2f}s",
            ha="center",
            va="bottom",
        )
    plt.tight_layout()
    plt.savefig(outpath, dpi=140)
    plt.close()
    print(f"wrote {outpath}")


MODE_LABEL = {
    "par": "Map-reduce (static partition)",
    "steal": "Work-stealing (Chase-Lev deque)",
}


def plot_speedup_per_implementation(times, mode, outpath):
    """One speedup curve per dataset, for a single parallel implementation.

    Satisfies the assignment's "one speedup graph per parallel
    implementation" requirement explicitly: this chart isolates `mode`
    and overlays balanced + unbalanced datasets so the reader can see
    how that single implementation scales as the workload changes.
    """
    thread_counts = sorted(
        t for (m, d, t) in times if m == mode and d == "balanced"
    )
    if not thread_counts:
        return
    bal_base = times.get(("seq", "balanced", 1))
    unb_base = times.get(("seq", "unbalanced", 1))
    if bal_base is None or unb_base is None:
        return
    bal = [bal_base / times[(mode, "balanced", T)] for T in thread_counts]
    unb = [unb_base / times[(mode, "unbalanced", T)] for T in thread_counts]

    plt.figure(figsize=(7, 5))
    plt.plot(thread_counts, bal, "o-", label="Balanced portfolio")
    plt.plot(thread_counts, unb, "s-", label="Unbalanced portfolio")
    plt.plot(thread_counts, thread_counts, "k--", alpha=0.35, label="Ideal speedup")
    plt.xlabel("Threads")
    plt.ylabel("Speedup vs sequential")
    plt.title(f"Speedup - {MODE_LABEL[mode]}")
    plt.grid(alpha=0.3)
    plt.legend()
    plt.tight_layout()
    plt.savefig(outpath, dpi=140)
    plt.close()
    print(f"wrote {outpath}")


def main():
    if not os.path.exists(TIMINGS):
        sys.stderr.write(f"no timings at {TIMINGS}; run the binary first\n")
        sys.exit(1)
    times = load_timings()
    # Per-dataset plots: both implementations overlaid (best at-a-glance view).
    plot_speedup(times, "balanced", os.path.join(PLOTS, "speedup_balanced.png"))
    plot_speedup(times, "unbalanced", os.path.join(PLOTS, "speedup_unbalanced.png"))
    # Per-implementation plots: both datasets overlaid (explicitly satisfies
    # the assignment's "one speedup graph per parallel implementation" line).
    plot_speedup_per_implementation(times, "par", os.path.join(PLOTS, "speedup_par.png"))
    plot_speedup_per_implementation(times, "steal", os.path.join(PLOTS, "speedup_steal.png"))
    plot_runtime_compare(times, os.path.join(PLOTS, "runtime_seq.png"))
    plot_unbalanced_zoom(times, os.path.join(PLOTS, "unbalanced_t8.png"))


if __name__ == "__main__":
    main()
