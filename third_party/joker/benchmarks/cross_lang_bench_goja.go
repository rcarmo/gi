//go:build ignore

// Cross-language benchmark: Goja (the JS engine used in gi)
// Run with: go run benchmarks/cross_lang_bench_goja.go

package main

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
)

const script = `
function arithmeticLoop() {
  var i = 0, s = 0;
  while (i < 100000) { s += (i * 7) % 11; i++; }
  return s;
}
function fib(n) { return n < 2 ? n : fib(n-1) + fib(n-2); }
function recursiveFib() { var s=0; for(var i=0;i<3;i++) s+=fib(24); return s; }
function tailRecursiveSum() { var n=100000,acc=0; while(n>0){acc+=n;n--;} return acc; }
function nbody() {
  var PI=3.141592653589793, SM=4*PI*PI, DPY=365.24;
  var b=[[0,0,0,0,0,0,SM],
    [4.84143144246472090,-1.16032004402742839,-0.103622044471123109,0.00166007664274403694*DPY,0.00769901118419740425*DPY,-0.0000690460016972063023*DPY,0.000954791938424326609*SM],
    [8.34336671824457987,4.12479856412430479,-0.403523417114321381,-0.00276742510726862411*DPY,0.00499852801234917238*DPY,0.0000230417297573763929*DPY,0.000285885980666130812*SM],
    [12.8943695621391310,-15.1111514016986312,-0.223307578892655734,0.00296460137564761618*DPY,0.00237847173959480950*DPY,-0.0000296589568540237556*DPY,0.0000436624404335156298*SM],
    [15.3796971148509165,-25.9193146099879641,0.179258772950371181,0.00268067772490389322*DPY,0.00162824170038242295*DPY,-0.0000951592254519715870*DPY,0.0000515138902046611451*SM]];
  var dt=0.01,n=5;
  for(var step=0;step<100;step++){
    for(var i=0;i<n;i++){for(var j=i+1;j<n;j++){
      var dx=b[i][0]-b[j][0],dy=b[i][1]-b[j][1],dz=b[i][2]-b[j][2];
      var d2=dx*dx+dy*dy+dz*dz,d=Math.sqrt(d2),mag=dt/(d2*d);
      b[i][3]-=dx*b[j][6]*mag;b[i][4]-=dy*b[j][6]*mag;b[i][5]-=dz*b[j][6]*mag;
      b[j][3]+=dx*b[i][6]*mag;b[j][4]+=dy*b[i][6]*mag;b[j][5]+=dz*b[i][6]*mag;
    }}
    for(var i2=0;i2<n;i2++){b[i2][0]+=dt*b[i2][3];b[i2][1]+=dt*b[i2][4];b[i2][2]+=dt*b[i2][5];}
  }
  var e=0;
  for(var i=0;i<n;i++){e+=0.5*b[i][6]*(b[i][3]*b[i][3]+b[i][4]*b[i][4]+b[i][5]*b[i][5]);
    for(var j=i+1;j<n;j++){var dx=b[i][0]-b[j][0],dy=b[i][1]-b[j][1],dz=b[i][2]-b[j][2];e-=b[i][6]*b[j][6]/Math.sqrt(dx*dx+dy*dy+dz*dz);}}
  return e;
}
function spectralNorm() {
  var n=50;
  function A(i,j){return 1.0/((i+j)*(i+j+1)/2+i+1);}
  function mulAv(v){var r=[];for(var i=0;i<n;i++){var s=0;for(var j=0;j<n;j++)s+=A(i,j)*v[j];r.push(s);}return r;}
  function mulAtv(v){var r=[];for(var i=0;i<n;i++){var s=0;for(var j=0;j<n;j++)s+=A(j,i)*v[j];r.push(s);}return r;}
  function mulAtAv(v){return mulAtv(mulAv(v));}
  var u=[];for(var i=0;i<n;i++)u.push(1.0);var v2;
  for(var it=0;it<10;it++){v2=mulAtAv(u);u=mulAtAv(v2);}
  var vBv=0,vv=0;for(var i=0;i<n;i++){vBv+=u[i]*v2[i];vv+=v2[i]*v2[i];}
  return Math.sqrt(vBv/vv);
}
function binaryTrees() {
  function make(d){return d===0?null:[make(d-1),make(d-1)];}
  function check(t){return t===null?1:1+check(t[0])+check(t[1]);}
  var total=0;
  for(var d=4;d<15;d++){var it=1<<(14-d),c=0;for(var i=0;i<it;i++)c+=check(make(d));total+=c;}
  return total;
}
function fannkuch(){var n=7,perm=[];for(var i=0;i<n;i++)perm.push(i);var mf=0,cs=0,sign=1,c=[];for(var i=0;i<n;i++)c.push(0);while(true){var p=perm.slice(),fl=0;while(p[0]!==0){var k=p[0];for(var lo=0,hi=k;lo<hi;lo++,hi--){var t=p[lo];p[lo]=p[hi];p[hi]=t;}fl++;}if(fl>mf)mf=fl;cs+=sign===1?fl:-fl;var i=1;sign=-sign;while(i<n){c[i]++;if(c[i]<i+1){if((i+1)%2===0){var t=perm[0];perm[0]=perm[i];perm[i]=t;}else{var t=perm[0];perm[0]=perm[1];perm[1]=t;}break;}c[i]=0;i++;}if(i>=n)break;}return mf*1000+cs;}
function mandelbrot(){var n=40,lsq=4.0,mi=50,count=0;for(var y=0;y<n;y++)for(var x=0;x<n;x++){var cr=2.0*x/n-1.5,ci=2.0*y/n-1.0,zr=0,zi=0,ins=1;for(var i=0;i<mi;i++){var zr2=zr*zr,zi2=zi*zi;if(zr2+zi2>lsq){ins=0;break;}zi=2*zr*zi+ci;zr=zr2-zi2+cr;}count+=ins;}return count;}
function fasta(){var im=139968,ia=3877,ic=29573,alu="GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGGGAGGCCGAGGCGGGCGGATCACCTGAGGTCAGGAGTTCGAGACCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAATACAAAAATTAGCCGGGCGTGGTGGCGCGCGCCTGTAATCCCAGCTACTCGGGAGGCTGAGGCAGGAGAATCGCTTGAACCCGGGAGGCGGAGGTTGCAGTGAGCCGAGATCGCGCCACTGCACTCCAGCCTGGGCGACAGAGCGAGACTCCGTCTCAAA",seed=42,cs=0;for(var i=0;i<1000;i++){seed=(seed*ia+ic)%im;cs+=seed%alu.length;}return cs+seed;}
function knucleotide(){var dna="GGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT TATTTTAATTTATAGTATTTTAATTTATAGT",total=0;for(var f=1;f<4;f++){var m={};for(var i=0;i<=dna.length-f;i++){var k=dna.substring(i,i+f);m[k]=(m[k]||0)+1;}var cnt=0;for(var k in m)cnt++;total+=cnt;}return total;}
function reverseComplement(){var dna="GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGGGAGGCCGAGGCGGGCGGATCACCTGAGGTCAGGAGTTCGAGACCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAATACAAAAATTAGCCGGGCGTGGTGGCGCGCGCCTGTAATCCCAGCTACTCGGGAGGCTGAGGCAGGAGAAT",comp={A:"T",T:"A",G:"C",C:"G"," ":" "},r="";for(var i=dna.length-1;i>=0;i--)r+=comp[dna[i]]||dna[i];return r.length;}
function regexRedux(){var inp="agggtaaa|tttaccct ggtattttaatttatagt aactatagtattttaatttatagt cattttaatttatagtaactatagtattttaatttatagt agggtaaa tttaccct agggtaaatttaccct agggtaaa|tttaccct",pats=["agggtaaa|tttaccct","[cgt]gggtaaa|tttaccc[acg]","a[act]ggtaaa|tttacc[agt]t","ag[act]gtaaa|tttac[agt]ct","agg[act]taaa|ttta[agt]cct","aggg[acg]aaa|ttt[cgt]ccct","agggt[cgt]aa|tt[acg]accct","agggta[cgt]a|t[acg]taccct","agggtaa[cgt]|[acg]ttaccct"],total=0;for(var i=0;i<pats.length;i++){var m=inp.match(new RegExp(pats[i],"g"));total+=m?m.length:0;}return total;}
function pidigits(){var q=1,r=0,t=1,k=1,n=3,l=3,d=0,cs=0;while(d<27){if(4*q+r-t<n*t){cs+=n;d++;var nr=10*(r-n*t);var nn=Math.floor((10*(3*q+r))/t)-10*n;q=q*10;r=nr;n=nn;}else{var q2=q*k,r2=(2*q+r)*l,t2=t*l,k2=k+1,n2=Math.floor((q*(7*k+2)+r*l)/t2),l2=l+2;q=q2;r=r2;t=t2;k=k2;n=n2;l=l2;}}return cs;}
`

func main() {
	vm := goja.New()
	_, err := vm.RunString(script)
	if err != nil {
		panic(err)
	}

	benchmarks := []string{
		"arithmeticLoop",
		"recursiveFib",
		"tailRecursiveSum",
		"nbody",
		"spectralNorm",
		"binaryTrees",
		"fannkuch",
		"mandelbrot",
		"fasta",
		"knucleotide",
		"reverseComplement",
		"regexRedux",
		"pidigits",
	}

	fmt.Println("Goja benchmarks (5 iterations each)")
	fmt.Println("============================================================")

	for _, name := range benchmarks {
		fn, ok := goja.AssertFunction(vm.Get(name))
		if !ok {
			fmt.Printf("%-30s SKIP (not a function)\n", name)
			continue
		}
		var totalNs int64
		var result goja.Value
		for i := 0; i < 5; i++ {
			start := time.Now()
			result, err = fn(goja.Undefined())
			elapsed := time.Since(start).Nanoseconds()
			if err != nil {
				fmt.Printf("%-30s ERROR: %v\n", name, err)
				break
			}
			totalNs += elapsed
		}
		avgMs := float64(totalNs) / 5.0 / 1_000_000.0
		fmt.Printf("%-30s %10.2f ms/op  (result: %v)\n", name, avgMs, result)
	}
}

