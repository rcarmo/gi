package core

// --- fannkuch-redux (N=7) ---

var fannkuchScript = `(let [n 7
      init-perm (loop [i 0 v []] (if (= i n) v (recur (+ i 1) (conj v i))))
      flip (fn [perm]
        (let [k (nth perm 0)]
          (loop [lo 0 hi k p perm]
            (if (< lo hi)
              (let [tmp (nth p lo)]
                (recur (+ lo 1) (- hi 1)
                  (assoc (assoc p lo (nth p hi)) hi tmp)))
              p))))
      count-flips (fn [perm]
        (loop [p perm c 0]
          (if (= (nth p 0) 0)
            c
            (recur (flip p) (+ c 1)))))]
  (letfn [(heap-perm [perm c nn max-flips checksum sign]
            (if (= nn 1)
              (let [flips (count-flips perm)
                    new-max (if (< max-flips flips) flips max-flips)
                    new-cs (if (= sign 1) (+ checksum flips) (- checksum flips))]
                [new-max new-cs])
              (loop [i 0 p perm mf max-flips cs checksum sg sign]
                (if (= i nn)
                  [mf cs]
                  (let [result (heap-perm p c (- nn 1) mf cs sg)
                        mf2 (nth result 0)
                        cs2 (nth result 1)
                        p2 (if (= (rem nn 2) 0)
                             (let [tmp (nth p 0)]
                               (assoc (assoc p 0 (nth p (- nn 1))) (- nn 1) tmp))
                             (let [tmp (nth p 0)]
                               (assoc (assoc p 0 (nth p i)) i tmp)))]
                    (recur (+ i 1) p2 mf2 cs2 (* sg -1)))))))]
    (let [result (heap-perm init-perm
                   (loop [i 0 v []] (if (= i n) v (recur (+ i 1) (conj v 0))))
                   n 0 0 1)]
      (+ (* (nth result 0) 1000) (nth result 1)))))`

// --- mandelbrot (N=200, max-iter=50) ---

var mandelbrotScript = `(letfn [(pixel [cr ci]
    (loop [zr 0.0 zi 0.0 i 0]
      (if (= i 50)
        1
        (let [zr2 (* zr zr)
              zi2 (* zi zi)]
          (if (< 4.0 (+ zr2 zi2))
            0
            (recur (+ (- zr2 zi2) cr)
                   (+ (* 2.0 (* zr zi)) ci)
                   (+ i 1)))))))]
  (loop [y 0 count 0]
    (if (= y 40)
      count
      (let [rc (loop [x 0 rc 0]
                 (if (= x 40)
                   rc
                   (recur (+ x 1) (+ rc
                     (pixel (- (/ (* 2.0 x) 40) 1.5)
                            (- (/ (* 2.0 y) 40) 1.0))))))]
        (recur (+ y 1) (+ count rc))))))`

// --- fasta (N=1000) — sequence generation ---

var fastaScript = `(let [im 139968
      ia 3877
      ic 29573
      alu "GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGGGAGGCCGAGGCGGGCGGATCACCTGAGGTCAGGAGTTCGAGACCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAATACAAAAATTAGCCGGGCGTGGTGGCGCGCGCCTGTAATCCCAGCTACTCGGGAGGCTGAGGCAGGAGAATCGCTTGAACCCGGGAGGCGGAGGTTGCAGTGAGCCGAGATCGCGCCACTGCACTCCAGCCTGGGCGACAGAGCGAGACTCCGTCTCAAA"]
  (loop [i 0 seed 42 checksum 0]
    (if (= i 1000)
      (+ checksum seed)
      (let [new-seed (rem (+ (* seed ia) ic) im)
            idx (rem new-seed (count alu))]
        (recur (+ i 1) new-seed (+ checksum idx))))))`

// --- pidigits (N=100) — big-integer-free approximation using Machin's formula terms ---
// (This is a simplified version since Joker doesn't have arbitrary precision by default;
//  we compute a bounded digit extraction instead.)

var pidigitsScript = `(loop [i 0
       q 1 r 0 t 1 k 1 n 3 l 3
       digits 0 checksum 0]
  (if (= digits 27)
    checksum
    (if (< (- (+ (* 4 q) r) t) (* n t))
      (recur (+ i 1) (* q 10) (* 10 (- r (* n t))) t k
             (- (/ (* 10 (+ (* 3 q) r)) t) (* 10 n)) l
             (+ digits 1) (+ checksum n))
      (let [q2 (* q k)
            r2 (* (+ (* 2 q) r) l)
            t2 (* t l)
            k2 (+ k 1)
            n2 (/ (+ (* q (+ (* 7 k) 2)) (* r l)) t2)
            l2 (+ l 2)]
        (recur i q2 r2 t2 k2 n2 l2 digits checksum)))))`

// --- k-nucleotide (simplified) — hash map frequency counting over a string ---

var knucleotideScript = `(let [dna "GGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT"
      len-dna (count dna)]
  (loop [frame 1 total 0]
    (if (= frame 4)
      total
      (let [freq (loop [i 0 m {}]
                   (if (< (- len-dna frame) i)
                     m
                     (let [k (loop [j 0 s ""]
                               (if (= j frame)
                                 s
                                 (recur (+ j 1) (str s (nth dna (+ i j))))))]
                       (recur (+ i 1) (assoc m k (+ 1 (get m k 0)))))))]
        (recur (+ frame 1) (+ total (count freq)))))))`

// --- reverse-complement (simplified) — reverse and complement a DNA string ---

var reverseComplementScript = `(let [dna "GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGGGAGGCCGAGGCGGGCGGATCACCTGAGGTCAGGAGTTCGAGACCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAATACAAAAATTAGCCGGGCGTGGTGGCGCGCGCCTGTAATCCCAGCTACTCGGGAGGCTGAGGCAGGAGAAT"
      comp (fn [c]
        (if (= c "A") "T"
        (if (= c "T") "A"
        (if (= c "G") "C"
        (if (= c "C") "G" c)))))
      len-dna (count dna)]
  (loop [i 0 result ""]
    (if (= i len-dna)
      (count result)
      (recur (+ i 1) (str (comp (str (nth dna (- (- len-dna 1) i)))) result)))))`

// suppress unused
var _ = fannkuchScript
var _ = mandelbrotScript
var _ = fastaScript
var _ = pidigitsScript
var _ = knucleotideScript
var _ = reverseComplementScript

// --- regex-redux (simplified) — regex match and replace on a DNA-like string ---

var regexReduxScript = `(let [input "agggtaaa|tttaccct ggtattttaatttatagt aactatagtattttaatttatagtagtattttaatttatagt cattttaatttatagtaactatagtattttaatttatagt agggtaaa tttaccct agggtaaatttaccct agggtaaa|tttaccct"
      patterns ["agggtaaa|tttaccct"
                "[cgt]gggtaaa|tttaccc[acg]"
                "a[act]ggtaaa|tttacc[agt]t"
                "ag[act]gtaaa|tttac[agt]ct"
                "agg[act]taaa|ttta[agt]cct"
                "aggg[acg]aaa|ttt[cgt]ccct"
                "agggt[cgt]aa|tt[acg]accct"
                "agggta[cgt]a|t[acg]taccct"
                "agggtaa[cgt]|[acg]ttaccct"]]
  (loop [i 0 total 0]
    (if (= i (count patterns))
      total
      (let [pat (nth patterns i)
            matches (re-seq (re-pattern pat) input)
            c (count matches)]
        (recur (+ i 1) (+ total c))))))`
