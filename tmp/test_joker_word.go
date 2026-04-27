//go:build ignore

package main

import (
	"context"
	"fmt"
	"github.com/rcarmo/gi/internal/scripting"
)

func main() {
	script := `(def text "alpha beta gamma alpha")
(def words (re-seq #"[^\s]+" text))
(loop [i 0 counts {}]
  (if (= i (count words))
    (str (+ (get counts "alpha" 0) (get counts "beta" 0)))
    (let [w (nth words i)]
      (recur (inc i) (assoc counts w (inc (get counts w 0)))))))`
	out, err := scripting.NewJokerRunner().Execute(context.Background(), script, &scripting.Bridge{})
	fmt.Printf("out=%q err=%v\n", out, err)
}
