package core

// clbg_bench_test.go — Computer Language Benchmarks Game programs ported to Joker.
// These provide realistic, well-known workloads for measuring interpreter performance.

import (
	"math"
	"sync"
	"testing"
)

var clbgInitOnce sync.Once

func clbgInit() {
	clbgInitOnce.Do(func() {
		sqrtProc := Proc{Fn: func(args []Object) Object {
			x := EnsureArgIsNumber(args, 0).Double().D
			return Double{D: math.Sqrt(x)}
		}, Name: "procSqrt"}
		// Register in core namespace
		vr := GLOBAL_ENV.CoreNamespace.Intern(MakeSymbol("sqrt"))
		vr.Value = sqrtProc
		vr.meta = EmptyArrayMap()
		// Also map into user namespace so the parser can resolve it
		ns := GLOBAL_ENV.CurrentNamespace()
		uv := ns.Intern(MakeSymbol("sqrt"))
		uv.Value = sqrtProc
		uv.meta = EmptyArrayMap()
	})
}

// BenchmarkCLBGNBody runs the n-body planetary simulation.
// Reference: https://benchmarksgame-team.pages.debian.net/benchmarksgame/description/nbody.html
func BenchmarkCLBGNBody(b *testing.B) {
	clbgInit()
	expr := compileBenchExpr(b, nbodyScript)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Eval(expr, nil)
	}
}

func BenchmarkCLBGSpectralNorm(b *testing.B) {
	clbgInit()
	expr := compileBenchExpr(b, spectralNormScript)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Eval(expr, nil)
	}
}

func BenchmarkCLBGBinaryTrees(b *testing.B) {
	clbgInit()
	expr := compileBenchExpr(b, binaryTreesScript)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Eval(expr, nil)
	}
}

// n-body: simulate 5 bodies for 1000 steps, return final energy.
// Scaled down from CLBG's 50M steps for benchmark harness practicality.
const nbodyScript = `
(let [pi 3.141592653589793
      solar-mass (* 4.0 (* pi pi))
      days-per-year 365.24
      ;; bodies: [x y z vx vy vz mass] as flat vector
      ;; Jupiter, Saturn, Uranus, Neptune + Sun
      initial-bodies
      [;; Sun
       0.0 0.0 0.0 0.0 0.0 0.0 solar-mass
       ;; Jupiter
       4.84143144246472090 -1.16032004402742839 -0.103622044471123109
       (* 0.00166007664274403694 days-per-year)
       (* 0.00769901118419740425 days-per-year)
       (* -0.0000690460016972063023 days-per-year)
       (* 0.000954791938424326609 solar-mass)
       ;; Saturn
       8.34336671824457987 4.12479856412430479 -0.403523417114321381
       (* -0.00276742510726862411 days-per-year)
       (* 0.00499852801234917238 days-per-year)
       (* 0.0000230417297573763929 days-per-year)
       (* 0.000285885980666130812 solar-mass)
       ;; Uranus
       12.8943695621391310 -15.1111514016986312 -0.223307578892655734
       (* 0.00296460137564761618 days-per-year)
       (* 0.00237847173959480950 days-per-year)
       (* -0.0000296589568540237556 days-per-year)
       (* 0.0000436624404335156298 solar-mass)
       ;; Neptune
       15.3796971148509165 -25.9193146099879641 0.179258772950371181
       (* 0.00268067772490389322 days-per-year)
       (* 0.00162824170038242295 days-per-year)
       (* -0.0000951592254519715870 days-per-year)
       (* 0.0000515138902046611451 solar-mass)]
      n-bodies 5
      body-size 7
      get-x (fn [bs i] (nth bs (* i body-size)))
      get-y (fn [bs i] (nth bs (+ (* i body-size) 1)))
      get-z (fn [bs i] (nth bs (+ (* i body-size) 2)))
      get-vx (fn [bs i] (nth bs (+ (* i body-size) 3)))
      get-vy (fn [bs i] (nth bs (+ (* i body-size) 4)))
      get-vz (fn [bs i] (nth bs (+ (* i body-size) 5)))
      get-m (fn [bs i] (nth bs (+ (* i body-size) 6)))
      set-at (fn [bs idx val] (assoc bs idx val))
      advance
      (fn [bs dt]
        (loop [i 0 b bs]
          (if (= i n-bodies)
            b
            (let [ix (get-x b i) iy (get-y b i) iz (get-z b i)
                  im (get-m b i)
                  b2 (loop [j (+ i 1) bj b vxi (get-vx b i) vyi (get-vy b i) vzi (get-vz b i)]
                       (if (= j n-bodies)
                         (let [base (* i body-size)]
                           (set-at (set-at (set-at bj (+ base 3) vxi) (+ base 4) vyi) (+ base 5) vzi))
                         (let [jx (get-x bj j) jy (get-y bj j) jz (get-z bj j)
                               jm (get-m bj j)
                               dx (- ix jx) dy (- iy jy) dz (- iz jz)
                               dist2 (+ (* dx dx) (+ (* dy dy) (* dz dz)))
                               dist (sqrt dist2)
                               mag (/ dt (* dist2 dist))
                               vxi2 (- vxi (* dx (* jm mag)))
                               vyi2 (- vyi (* dy (* jm mag)))
                               vzi2 (- vzi (* dz (* jm mag)))
                               jvx (+ (get-vx bj j) (* dx (* im mag)))
                               jvy (+ (get-vy bj j) (* dy (* im mag)))
                               jvz (+ (get-vz bj j) (* dz (* im mag)))
                               jbase (* j body-size)
                               bj2 (set-at (set-at (set-at bj (+ jbase 3) jvx) (+ jbase 4) jvy) (+ jbase 5) jvz)]
                           (recur (+ j 1) bj2 vxi2 vyi2 vzi2))))]
              (let [base (* i body-size)
                    vx (get-vx b2 i) vy (get-vy b2 i) vz (get-vz b2 i)
                    b3 (set-at (set-at (set-at b2 base (+ ix (* dt vx)))
                                       (+ base 1) (+ iy (* dt vy)))
                               (+ base 2) (+ iz (* dt vz)))]
                (recur (+ i 1) b3))))))
      energy
      (fn [bs]
        (loop [i 0 e 0.0]
          (if (= i n-bodies)
            e
            (let [vx (get-vx bs i) vy (get-vy bs i) vz (get-vz bs i)
                  m (get-m bs i)
                  ke (* 0.5 (* m (+ (* vx vx) (+ (* vy vy) (* vz vz)))))
                  pe (loop [j (+ i 1) pe2 0.0]
                       (if (= j n-bodies)
                         pe2
                         (let [dx (- (get-x bs i) (get-x bs j))
                               dy (- (get-y bs i) (get-y bs j))
                               dz (- (get-z bs i) (get-z bs j))
                               dist (sqrt (+ (* dx dx) (+ (* dy dy) (* dz dz))))]
                           (recur (+ j 1) (- pe2 (/ (* m (get-m bs j)) dist))))))]
              (recur (+ i 1) (+ e (- ke pe)))))))]
  (loop [step 0 bodies initial-bodies]
    (if (= step 100)
      (energy bodies)
      (recur (+ step 1) (advance bodies 0.01)))))
`

// spectral-norm: compute spectral norm of an infinite matrix, N=50.
// Scaled down from CLBG's N=5500 for benchmark harness practicality.
const spectralNormScript = `
(let [n 50
      A (fn [i j] (/ 1.0 (+ (/ (* (+ i j) (+ (+ i j) 1)) 2) (+ i 1))))
      mul-Av (fn [v n]
        (loop [i 0 result []]
          (if (= i n)
            result
            (let [s (loop [j 0 s 0.0]
                      (if (= j n)
                        s
                        (recur (+ j 1) (+ s (* (A i j) (nth v j))))))]
              (recur (+ i 1) (conj result s))))))
      mul-Atv (fn [v n]
        (loop [i 0 result []]
          (if (= i n)
            result
            (let [s (loop [j 0 s 0.0]
                      (if (= j n)
                        s
                        (recur (+ j 1) (+ s (* (A j i) (nth v j))))))]
              (recur (+ i 1) (conj result s))))))
      mul-AtAv (fn [v n] (mul-Atv (mul-Av v n) n))
      initial-u (loop [i 0 v []]
                  (if (= i n) v (recur (+ i 1) (conj v 1.0))))]
  (loop [iter 0 u initial-u v []]
    (if (= iter 10)
      (let [vBv (loop [i 0 s 0.0]
                  (if (= i n) s (recur (+ i 1) (+ s (* (nth u i) (nth v i))))))
            vv (loop [i 0 s 0.0]
                 (if (= i n) s (recur (+ i 1) (+ s (* (nth v i) (nth v i))))))]
        (sqrt (/ vBv vv)))
      (let [v2 (mul-AtAv u n)
            u2 (mul-AtAv v2 n)]
        (recur (+ iter 1) u2 v2)))))
`

// binary-trees: allocate and check binary trees, depth=14.
// Scaled down from CLBG's depth=21 for benchmark harness practicality.
const binaryTreesScript = `
(letfn [(make-tree [depth]
          (if (= depth 0)
            [:leaf]
            (let [d (- depth 1)]
              [:node (make-tree d) (make-tree d)])))
        (check-tree [tree]
          (if (= (first tree) :leaf)
            1
            (+ 1 (+ (check-tree (nth tree 1)) (check-tree (nth tree 2))))))]
  (loop [d 4 total 0]
    (if (= d 15)
      total
      (let [iterations (loop [i 0 n 1] (if (= i (- 14 d)) n (recur (+ i 1) (* n 2))))
            check (loop [i 0 c 0]
                    (if (= i iterations)
                      c
                      (recur (+ i 1) (+ c (check-tree (make-tree d))))))]
        (recur (+ d 1) (+ total check))))))
`
