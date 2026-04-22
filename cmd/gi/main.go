package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
	giweb "github.com/rcarmo/gi/internal/web"
)

func main() {
	listen := flag.String("listen", ":8081", "HTTP listen address")
	dbPath := flag.String("db", "./gi.db", "SQLite database path")
	flag.Parse()

	s, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	engine := turn.New(s)
	server := giweb.New(s, engine)

	log.Printf("Gi web listening on %s using %s", *listen, *dbPath)
	if err := http.ListenAndServe(*listen, server.Handler()); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
