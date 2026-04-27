//go:build ignore

package main

import (
	"context"
	"fmt"
	"github.com/rcarmo/gi/internal/scripting"
	"strings"
)

func main() {
	base := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta"}
	words := make([]string, 0, 16)
	for i := 0; i < 16; i++ {
		words = append(words, base[i%len(base)])
	}
	text := strings.Join(words, " ")
	script := fmt.Sprintf(`(def text %q)
(def words (re-seq #"\\S+" text))
(loop [i 0 counts {}]
  (if (= i (count words))
    (str (+ (get counts "theta" 0) (get counts "alpha" 0)))
    (let [w (nth words i)]
      (recur (inc i) (assoc counts w (inc (get counts w 0)))))))`, text)
	out, err := scripting.NewJokerRunner().Execute(context.Background(), script, &scripting.Bridge{})
	fmt.Printf("out=%q err=%v\nscript=%s\n", out, err, script)
}
