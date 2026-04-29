// Cross-language benchmark suite matching the Joker CLBG benchmarks.
// Run with: bun benchmarks/cross_lang_bench.js

function bench(name, fn, iterations = 5) {
  const times = [];
  let result;
  for (let i = 0; i < iterations; i++) {
    const start = Bun.nanoseconds();
    result = fn();
    times.push(Bun.nanoseconds() - start);
  }
  const avg = times.reduce((a, b) => a + b, 0) / times.length / 1_000_000;
  console.log(`${name.padEnd(30)} ${avg.toFixed(2).padStart(10)} ms/op  (result: ${result})`);
}

// --- Arithmetic loop ---
function arithmeticLoop() {
  let i = 0, s = 0;
  while (i < 100000) {
    s += (i * 7) % 11;
    i++;
  }
  return s;
}

// --- Recursive fib ---
function fib(n) {
  if (n < 2) return n;
  return fib(n - 1) + fib(n - 2);
}
function recursiveFib() {
  let s = 0;
  for (let i = 0; i < 3; i++) s += fib(24);
  return s;
}

// --- Tail-recursive sum ---
function tailRecursiveSum() {
  let n = 100000, acc = 0;
  while (n > 0) { acc += n; n--; }
  return acc;
}

// --- N-body (100 steps) ---
function nbody() {
  const PI = 3.141592653589793;
  const SOLAR_MASS = 4 * PI * PI;
  const DAYS_PER_YEAR = 365.24;
  const bodies = [
    [0,0,0,0,0,0,SOLAR_MASS],
    [4.84143144246472090,-1.16032004402742839,-0.103622044471123109,
     0.00166007664274403694*DAYS_PER_YEAR,0.00769901118419740425*DAYS_PER_YEAR,-0.0000690460016972063023*DAYS_PER_YEAR,
     0.000954791938424326609*SOLAR_MASS],
    [8.34336671824457987,4.12479856412430479,-0.403523417114321381,
     -0.00276742510726862411*DAYS_PER_YEAR,0.00499852801234917238*DAYS_PER_YEAR,0.0000230417297573763929*DAYS_PER_YEAR,
     0.000285885980666130812*SOLAR_MASS],
    [12.8943695621391310,-15.1111514016986312,-0.223307578892655734,
     0.00296460137564761618*DAYS_PER_YEAR,0.00237847173959480950*DAYS_PER_YEAR,-0.0000296589568540237556*DAYS_PER_YEAR,
     0.0000436624404335156298*SOLAR_MASS],
    [15.3796971148509165,-25.9193146099879641,0.179258772950371181,
     0.00268067772490389322*DAYS_PER_YEAR,0.00162824170038242295*DAYS_PER_YEAR,-0.0000951592254519715870*DAYS_PER_YEAR,
     0.0000515138902046611451*SOLAR_MASS],
  ];
  const dt = 0.01, n = bodies.length;
  for (let step = 0; step < 100; step++) {
    for (let i = 0; i < n; i++) {
      const bi = bodies[i];
      for (let j = i+1; j < n; j++) {
        const bj = bodies[j];
        const dx = bi[0]-bj[0], dy = bi[1]-bj[1], dz = bi[2]-bj[2];
        const dist2 = dx*dx + dy*dy + dz*dz;
        const dist = Math.sqrt(dist2);
        const mag = dt / (dist2 * dist);
        bi[3] -= dx*bj[6]*mag; bi[4] -= dy*bj[6]*mag; bi[5] -= dz*bj[6]*mag;
        bj[3] += dx*bi[6]*mag; bj[4] += dy*bi[6]*mag; bj[5] += dz*bi[6]*mag;
      }
    }
    for (const bi of bodies) { bi[0]+=dt*bi[3]; bi[1]+=dt*bi[4]; bi[2]+=dt*bi[5]; }
  }
  let e = 0;
  for (let i = 0; i < n; i++) {
    const bi = bodies[i];
    e += 0.5*bi[6]*(bi[3]*bi[3]+bi[4]*bi[4]+bi[5]*bi[5]);
    for (let j = i+1; j < n; j++) {
      const bj = bodies[j];
      const dx=bi[0]-bj[0], dy=bi[1]-bj[1], dz=bi[2]-bj[2];
      e -= bi[6]*bj[6] / Math.sqrt(dx*dx+dy*dy+dz*dz);
    }
  }
  return +e.toFixed(6);
}

// --- Spectral norm (N=50) ---
function spectralNorm() {
  const n = 50;
  function A(i,j) { return 1.0 / ((i+j)*(i+j+1)/2 + i + 1); }
  function mulAv(v) {
    const r = new Array(n);
    for (let i=0;i<n;i++){let s=0;for(let j=0;j<n;j++) s+=A(i,j)*v[j]; r[i]=s;}
    return r;
  }
  function mulAtv(v) {
    const r = new Array(n);
    for (let i=0;i<n;i++){let s=0;for(let j=0;j<n;j++) s+=A(j,i)*v[j]; r[i]=s;}
    return r;
  }
  function mulAtAv(v) { return mulAtv(mulAv(v)); }
  let u = new Array(n).fill(1.0), v;
  for (let i = 0; i < 10; i++) { v = mulAtAv(u); u = mulAtAv(v); }
  let vBv = 0, vv = 0;
  for (let i = 0; i < n; i++) { vBv += u[i]*v[i]; vv += v[i]*v[i]; }
  return +Math.sqrt(vBv/vv).toFixed(9);
}

// --- Binary trees (depth 14) ---
function binaryTrees() {
  function make(d) { return d === 0 ? null : [make(d-1), make(d-1)]; }
  function check(t) { return t === null ? 1 : 1 + check(t[0]) + check(t[1]); }
  let total = 0;
  for (let d = 4; d < 15; d++) {
    const iters = 1 << (14 - d);
    let c = 0;
    for (let i = 0; i < iters; i++) c += check(make(d));
    total += c;
  }
  return total;
}

console.log("Bun/JSC benchmarks (5 iterations each)");
console.log("=".repeat(60));
bench("arithmetic_loop", arithmeticLoop);
bench("recursive_fib", recursiveFib);
bench("tail_recursive_sum", tailRecursiveSum);
bench("nbody_100steps", nbody);
bench("spectral_norm_50", spectralNorm);
bench("binary_trees_14", binaryTrees);

// --- Fannkuch-redux (N=7) ---
function fannkuch() {
  const n = 7;
  const perm = Array.from({length:n},(_,i)=>i);
  let maxFlips=0, checksum=0, sign=1;
  const c = new Array(n).fill(0);
  while (true) {
    const p = perm.slice();
    let flips = 0;
    while (p[0]!==0) { const k=p[0]; for(let lo=0,hi=k;lo<hi;lo++,hi--){const t=p[lo];p[lo]=p[hi];p[hi]=t;} flips++; }
    if (flips>maxFlips) maxFlips=flips;
    checksum += sign===1?flips:-flips;
    let i=1; sign=-sign;
    while (i<n) {
      c[i]++;
      if (c[i]<i+1) { if((i+1)%2===0){const t=perm[0];perm[0]=perm[i];perm[i]=t;}else{const t=perm[0];perm[0]=perm[1];perm[1]=t;} break; }
      c[i]=0; i++;
    }
    if (i>=n) break;
  }
  return maxFlips*1000+checksum;
}

// --- Mandelbrot (N=200, max_iter=50) ---
function mandelbrot() {
  const n=40, lsq=4.0, mi=50;
  let count=0;
  for(let y=0;y<n;y++) for(let x=0;x<n;x++) {
    const cr=2.0*x/n-1.5, ci=2.0*y/n-1.0;
    let zr=0,zi=0,inside=1;
    for(let i=0;i<mi;i++){const zr2=zr*zr,zi2=zi*zi;if(zr2+zi2>lsq){inside=0;break;}zi=2*zr*zi+ci;zr=zr2-zi2+cr;}
    count+=inside;
  }
  return count;
}

// --- Fasta (N=1000) ---
function fasta() {
  const im=139968,ia=3877,ic=29573;
  const alu="GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGGGAGGCCGAGGCGGGCGGATCACCTGAGGTCAGGAGTTCGAGACCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAATACAAAAATTAGCCGGGCGTGGTGGCGCGCGCCTGTAATCCCAGCTACTCGGGAGGCTGAGGCAGGAGAATCGCTTGAACCCGGGAGGCGGAGGTTGCAGTGAGCCGAGATCGCGCCACTGCACTCCAGCCTGGGCGACAGAGCGAGACTCCGTCTCAAA";
  let seed=42,cs=0;
  for(let i=0;i<1000;i++){seed=(seed*ia+ic)%im;cs+=seed%alu.length;}
  return cs+seed;
}

// --- K-nucleotide ---
function knucleotide() {
  const dna="GGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT";
  let total=0;
  for(let f=1;f<4;f++){const m=new Map();for(let i=0;i<=dna.length-f;i++){const k=dna.substring(i,i+f);m.set(k,(m.get(k)||0)+1);}total+=m.size;}
  return total;
}

// --- Reverse complement ---
function reverseComplement() {
  const dna="GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGGGAGGCCGAGGCGGGCGGATCACCTGAGGTCAGGAGTTCGAGACCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAATACAAAAATTAGCCGGGCGTGGTGGCGCGCGCCTGTAATCCCAGCTACTCGGGAGGCTGAGGCAGGAGAAT";
  const comp={A:"T",T:"A",G:"C",C:"G"," ":" "};
  let r=""; for(let i=dna.length-1;i>=0;i--) r+=comp[dna[i]]||dna[i];
  return r.length;
}

bench("fannkuch_7", fannkuch);
bench("mandelbrot_200", mandelbrot);
bench("fasta_1000", fasta);
bench("knucleotide", knucleotide);
bench("reverse_complement", reverseComplement);

// --- Regex-redux ---
function regexRedux() {
  const inp = "agggtaaa|tttaccct ggtattttaatttatagt aactatagtattttaatttatagt cattttaatttatagtaactatagtattttaatttatagt agggtaaa tttaccct agggtaaatttaccct agggtaaa|tttaccct";
  const patterns = [
    /agggtaaa|tttaccct/g, /[cgt]gggtaaa|tttaccc[acg]/g, /a[act]ggtaaa|tttacc[agt]t/g,
    /ag[act]gtaaa|tttac[agt]ct/g, /agg[act]taaa|ttta[agt]cct/g, /aggg[acg]aaa|ttt[cgt]ccct/g,
    /agggt[cgt]aa|tt[acg]accct/g, /agggta[cgt]a|t[acg]taccct/g, /agggtaa[cgt]|[acg]ttaccct/g
  ];
  let total = 0;
  for (const p of patterns) { const m = inp.match(p); total += m ? m.length : 0; }
  return total;
}
bench("regex_redux", regexRedux);

// --- Pidigits (N=27) ---
function pidigits() {
  let q=1,r=0,t=1,k=1,n=3,l=3,digits=0,cs=0;
  while (digits < 27) {
    if (4*q+r-t < n*t) {
      cs += n; digits++;
      const nr = 10*(r-n*t); const nn = Math.floor((10*(3*q+r))/t) - 10*n;
      q = q*10; r = nr; n = nn;
    } else {
      const q2=q*k, r2=(2*q+r)*l, t2=t*l, k2=k+1;
      const n2=Math.floor((q*(7*k+2)+r*l)/t2), l2=l+2;
      q=q2;r=r2;t=t2;k=k2;n=n2;l=l2;
    }
  }
  return cs;
}
bench("pidigits_27", pidigits);
