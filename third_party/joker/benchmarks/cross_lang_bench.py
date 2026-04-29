#!/usr/bin/env python3
"""Cross-language benchmark suite matching the Joker CLBG benchmarks."""
import time, math

def bench(name, fn, iterations=5):
    times = []
    for _ in range(iterations):
        start = time.perf_counter_ns()
        result = fn()
        elapsed = time.perf_counter_ns() - start
        times.append(elapsed)
    avg_ms = sum(times) / len(times) / 1_000_000
    print(f"{name:30s} {avg_ms:10.2f} ms/op  (result: {result})")

# --- Arithmetic loop (matches Joker benchmark) ---
def arithmetic_loop():
    i, s = 0, 0
    while i < 100000:
        s += (i * 7) % 11
        i += 1
    return s

# --- Recursive fib ---
def fib(n):
    if n < 2: return n
    return fib(n-1) + fib(n-2)

def recursive_fib():
    s = 0
    for _ in range(3):
        s += fib(24)
    return s

# --- Tail-recursive sum (iterative in Python) ---
def tail_recursive_sum():
    n, acc = 100000, 0
    while n > 0:
        acc += n
        n -= 1
    return acc

# --- N-body (100 steps) ---
def nbody():
    PI = 3.141592653589793
    SOLAR_MASS = 4 * PI * PI
    DAYS_PER_YEAR = 365.24

    bodies = [
        [0.0, 0.0, 0.0, 0.0, 0.0, 0.0, SOLAR_MASS],
        [4.84143144246472090, -1.16032004402742839, -0.103622044471123109,
         0.00166007664274403694*DAYS_PER_YEAR, 0.00769901118419740425*DAYS_PER_YEAR, -0.0000690460016972063023*DAYS_PER_YEAR,
         0.000954791938424326609*SOLAR_MASS],
        [8.34336671824457987, 4.12479856412430479, -0.403523417114321381,
         -0.00276742510726862411*DAYS_PER_YEAR, 0.00499852801234917238*DAYS_PER_YEAR, 0.0000230417297573763929*DAYS_PER_YEAR,
         0.000285885980666130812*SOLAR_MASS],
        [12.8943695621391310, -15.1111514016986312, -0.223307578892655734,
         0.00296460137564761618*DAYS_PER_YEAR, 0.00237847173959480950*DAYS_PER_YEAR, -0.0000296589568540237556*DAYS_PER_YEAR,
         0.0000436624404335156298*SOLAR_MASS],
        [15.3796971148509165, -25.9193146099879641, 0.179258772950371181,
         0.00268067772490389322*DAYS_PER_YEAR, 0.00162824170038242295*DAYS_PER_YEAR, -0.0000951592254519715870*DAYS_PER_YEAR,
         0.0000515138902046611451*SOLAR_MASS],
    ]
    dt = 0.01
    n = len(bodies)
    for _ in range(100):
        for i in range(n):
            bi = bodies[i]
            for j in range(i+1, n):
                bj = bodies[j]
                dx = bi[0]-bj[0]; dy = bi[1]-bj[1]; dz = bi[2]-bj[2]
                dist2 = dx*dx + dy*dy + dz*dz
                dist = math.sqrt(dist2)
                mag = dt / (dist2 * dist)
                bi[3] -= dx * bj[6] * mag
                bi[4] -= dy * bj[6] * mag
                bi[5] -= dz * bj[6] * mag
                bj[3] += dx * bi[6] * mag
                bj[4] += dy * bi[6] * mag
                bj[5] += dz * bi[6] * mag
        for bi in bodies:
            bi[0] += dt * bi[3]
            bi[1] += dt * bi[4]
            bi[2] += dt * bi[5]
    # energy
    e = 0.0
    for i in range(n):
        bi = bodies[i]
        e += 0.5 * bi[6] * (bi[3]*bi[3] + bi[4]*bi[4] + bi[5]*bi[5])
        for j in range(i+1, n):
            bj = bodies[j]
            dx = bi[0]-bj[0]; dy = bi[1]-bj[1]; dz = bi[2]-bj[2]
            e -= bi[6]*bj[6] / math.sqrt(dx*dx+dy*dy+dz*dz)
    return round(e, 6)

# --- Spectral norm (N=50) ---
def spectral_norm():
    n = 50
    def A(i,j): return 1.0 / ((i+j)*(i+j+1)//2 + i + 1)
    def mul_Av(v):
        return [sum(A(i,j)*v[j] for j in range(n)) for i in range(n)]
    def mul_Atv(v):
        return [sum(A(j,i)*v[j] for j in range(n)) for i in range(n)]
    def mul_AtAv(v): return mul_Atv(mul_Av(v))
    u = [1.0]*n
    for _ in range(10):
        v = mul_AtAv(u)
        u = mul_AtAv(v)
    vBv = sum(u[i]*v[i] for i in range(n))
    vv = sum(v[i]*v[i] for i in range(n))
    return round(math.sqrt(vBv/vv), 9)

# --- Binary trees (depth 14) ---
def binary_trees():
    def make_tree(d):
        if d == 0: return None
        return (make_tree(d-1), make_tree(d-1))
    def check(t):
        if t is None: return 1
        return 1 + check(t[0]) + check(t[1])
    total = 0
    for d in range(4, 15):
        iters = 1 << (14 - d)
        c = sum(check(make_tree(d)) for _ in range(iters))
        total += c
    return total

if __name__ == "__main__":
    print("Python 3.13 benchmarks (5 iterations each)")
    print("=" * 60)
    bench("arithmetic_loop", arithmetic_loop)
    bench("recursive_fib", recursive_fib)
    bench("tail_recursive_sum", tail_recursive_sum)
    bench("nbody_100steps", nbody)
    bench("spectral_norm_50", spectral_norm)
    bench("binary_trees_14", binary_trees)

# --- Fannkuch-redux (N=7) ---
def fannkuch():
    n = 7
    perm = list(range(n))
    max_flips = 0
    checksum = 0
    c = [0] * n
    sign = 1
    while True:
        # count flips
        p = perm[:]
        flips = 0
        while p[0] != 0:
            k = p[0]
            p[:k+1] = p[k::-1]
            flips += 1
        if flips > max_flips: max_flips = flips
        checksum += flips if sign == 1 else -flips
        # next permutation (Heap's algorithm)
        i = 1
        sign = -sign
        while i < n:
            c[i] += 1
            if c[i] < i + 1:
                if (i + 1) % 2 == 0:
                    perm[0], perm[i] = perm[i], perm[0]
                else:
                    perm[0], perm[1] = perm[1], perm[0]
                break
            c[i] = 0
            i += 1
        else:
            break
    return max_flips * 1000 + checksum

# --- Mandelbrot (N=200, max_iter=50) ---
def mandelbrot():
    n, limit_sq, max_iter = 40, 4.0, 50
    count = 0
    for y in range(n):
        for x in range(n):
            cr = 2.0 * x / n - 1.5
            ci = 2.0 * y / n - 1.0
            zr, zi = 0.0, 0.0
            inside = 1
            for _ in range(max_iter):
                zr2, zi2 = zr*zr, zi*zi
                if zr2 + zi2 > limit_sq:
                    inside = 0
                    break
                zi = 2.0*zr*zi + ci
                zr = zr2 - zi2 + cr
            count += inside
    return count

# --- Fasta (N=1000) ---
def fasta():
    im, ia, ic = 139968, 3877, 29573
    alu = "GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGGGAGGCCGAGGCGGGCGGATCACCTGAGGTCAGGAGTTCGAGACCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAATACAAAAATTAGCCGGGCGTGGTGGCGCGCGCCTGTAATCCCAGCTACTCGGGAGGCTGAGGCAGGAGAATCGCTTGAACCCGGGAGGCGGAGGTTGCAGTGAGCCGAGATCGCGCCACTGCACTCCAGCCTGGGCGACAGAGCGAGACTCCGTCTCAAA"
    seed, checksum = 42, 0
    for _ in range(1000):
        seed = (seed * ia + ic) % im
        checksum += seed % len(alu)
    return checksum + seed

# --- K-nucleotide (simplified) ---
def knucleotide():
    dna = "GGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT"
    total = 0
    for frame in range(1, 4):
        freq = {}
        for i in range(len(dna) - frame + 1):
            k = dna[i:i+frame]
            freq[k] = freq.get(k, 0) + 1
        total += len(freq)
    return total

# --- Reverse-complement (simplified) ---
def reverse_complement():
    dna = "GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGGGAGGCCGAGGCGGGCGGATCACCTGAGGTCAGGAGTTCGAGACCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAATACAAAAATTAGCCGGGCGTGGTGGCGCGCGCCTGTAATCCCAGCTACTCGGGAGGCTGAGGCAGGAGAAT"
    comp = {"A":"T","T":"A","G":"C","C":"G"," ":" "}
    return len("".join(comp.get(c, c) for c in reversed(dna)))

bench("fannkuch_7", fannkuch)
bench("mandelbrot_200", mandelbrot)
bench("fasta_1000", fasta)
bench("knucleotide", knucleotide)
bench("reverse_complement", reverse_complement)

# --- Regex-redux (simplified) ---
import re
def regex_redux():
    inp = "agggtaaa|tttaccct ggtattttaatttatagt aactatagtattttaatttatagt cattttaatttatagtaactatagtattttaatttatagt agggtaaa tttaccct agggtaaatttaccct agggtaaa|tttaccct"
    patterns = [
        "agggtaaa|tttaccct",
        "[cgt]gggtaaa|tttaccc[acg]",
        "a[act]ggtaaa|tttacc[agt]t",
        "ag[act]gtaaa|tttac[agt]ct",
        "agg[act]taaa|ttta[agt]cct",
        "aggg[acg]aaa|ttt[cgt]ccct",
        "agggt[cgt]aa|tt[acg]accct",
        "agggta[cgt]a|t[acg]taccct",
        "agggtaa[cgt]|[acg]ttaccct",
    ]
    total = 0
    for pat in patterns:
        total += len(re.findall(pat, inp))
    return total

bench("regex_redux", regex_redux)

# --- Pidigits (N=27) ---
def pidigits():
    q, r, t, k, n, l = 1, 0, 1, 1, 3, 3
    digits, checksum = 0, 0
    while digits < 27:
        if 4*q + r - t < n*t:
            checksum += n
            digits += 1
            q, r, n = q*10, 10*(r - n*t), (10*(3*q + r))//t - 10*n
        else:
            q2 = q*k
            r2 = (2*q + r)*l
            t2 = t*l
            k2 = k + 1
            n2 = (q*(7*k + 2) + r*l) // t2
            l2 = l + 2
            q, r, t, k, n, l = q2, r2, t2, k2, n2, l2
    return checksum

bench("pidigits_27", pidigits)
