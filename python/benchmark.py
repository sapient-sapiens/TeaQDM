import os
import sys
import time
import csv

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from tqdm import tqdm as real_tqdm
from python.wrapper import TeaQDM


class RawBar:
    # I guess this doesn't have to render anything so it would be very hard to beat this. 
    def __init__(self, iterable, desc="Raw", total=None, width=40):
        self.iterable = iterable
        self.desc = desc
        if total is None and hasattr(iterable, "__len__"):
            total = len(iterable)
        self.total = total
        self.n = 0
        self.width = width

    def __iter__(self):
        for x in self.iterable:
            self.n += 1
            yield x


def fast_work(i: int) -> int: # time.sleep() can only measure starting from microseconds. 
    return i * 2


def medium_work(i: int) -> int:
    s = 0
    for _ in range(100):
        s += i
    return s


def slow_work(i: int) -> int:
    s = 0
    for _ in range(1000):
        s += i * i
    return s


def time_loop(label: str, iterable, fn):
    start = time.perf_counter()
    for x in iterable:
        fn(x)
    return time.perf_counter() - start


def run_one_test(name: str, n: int, fn, writer):
    baseline = time_loop("baseline", range(n), fn)
    teaqdm_t = time_loop("teaqdm", TeaQDM(range(n), desc=name), fn)
    raw_t = time_loop("raw", RawBar(range(n), desc=name), fn)
    tqdm_t = time_loop("tqdm", real_tqdm(range(n), desc=name, ncols=80, disable=True), fn)

    writer.writerow({
        'workload': name,
        'iterations': n,
        'baseline_s': f"{baseline:.6f}",
        'teaqdm_s': f"{teaqdm_t:.6f}",
        'teaqdm_overhead_s': f"{teaqdm_t - baseline:.6f}",
        'teaqdm_percent': f"{teaqdm_t / baseline * 100:.2f}",
        'raw_s': f"{raw_t:.6f}",
        'raw_overhead_s': f"{raw_t - baseline:.6f}",
        'raw_percent': f"{raw_t / baseline * 100:.2f}",
        'tqdm_s': f"{tqdm_t:.6f}",
        'tqdm_overhead_s': f"{tqdm_t - baseline:.6f}",
        'tqdm_percent': f"{tqdm_t / baseline * 100:.2f}",
    })


def main():
    os.environ["TEAQDM_NO_ALTSCREEN"] = "1"

    with open("benchmark_results.csv", mode="w", newline="") as csvfile:
        fieldnames = [
            'workload', 'iterations',
            'baseline_s',
            'teaqdm_s', 'teaqdm_overhead_s', 'teaqdm_percent',
            'raw_s', 'raw_overhead_s', 'raw_percent',
            'tqdm_s', 'tqdm_overhead_s', 'tqdm_percent'
        ]
        writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
        writer.writeheader()
        run_one_test("fast", 1000, fast_work, writer)
        run_one_test("medium", 500, medium_work, writer)
        run_one_test("slow", 200, slow_work, writer)


if __name__ == "__main__":
    main()
