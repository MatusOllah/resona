#!/usr/bin/env python3

"""
Generates golden window function test data as JSON.
Do not edit the output JSON manually!
"""

import numpy as np
from scipy.signal import windows
import json

def welch(n: int):
    return (1 - ((np.arange(n) - (n - 1)/2) / (n / 2))**2)

def generate(n: int):
    print(f"[*] Generating golden data for n = {n}")

    data = {
        "Rectangular": windows.boxcar(n).tolist(),
        "Welch": welch(n).tolist(),
        "Hann": windows.hann(n, sym=True).tolist(),
        "Hamming": windows.hamming(n, sym=True).tolist(),
        "Blackman": windows.blackman(n, sym=True).tolist(),
    }

    with open(f"window{n}.json", "w") as f:
        json.dump(data, f, indent=4)

def main():
    lengths = [0, 1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024]

    for n in lengths:
        generate(n)

if __name__ == "__main__": main()
