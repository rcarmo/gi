package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
	giweb "github.com/rcarmo/gi/internal/web"
)

func main() {
	listen := flag.String("listen", "", "HTTP listen address (overrides -bind/-port)")
	bind := flag.String("bind", "127.0.0.1", "Bind address / interface host")
	port := flag.Int("port", 8081, "HTTP port")
	dbPath := flag.String("db", "./gi.db", "SQLite database path")
	workspace := flag.String("workspace", "/workspace", "Workspace root")
	logFile := flag.String("log-file", "", "Optional log file path")
	pidFile := flag.String("pid-file", "", "Optional pid file path")
	flag.Parse()

	effectiveListen := *listen
	if effectiveListen == "" {
		effectiveListen = net.JoinHostPort(*bind, fmt.Sprintf("%d", *port))
	}

	if *logFile != "" {
		if err := os.MkdirAll(filepath.Dir(*logFile), 0o755); err != nil {
			log.Fatalf("create log dir: %v", err)
		}
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			log.Fatalf("open log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}
	if *pidFile != "" {
		if err := os.MkdirAll(filepath.Dir(*pidFile), 0o755); err != nil {
			log.Fatalf("create pid dir: %v", err)
		}
		if err := os.WriteFile(*pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0o644); err != nil {
			log.Fatalf("write pid file: %v", err)
		}
		defer os.Remove(*pidFile)
	}

	s, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	runtimeCfg := config.Load(*workspace)
	engine := turn.New(s)
	server := giweb.New(s, engine, runtimeCfg)

	log.Printf("Gi web listening on %s using %s", effectiveListen, *dbPath)
	if err := http.ListenAndServe(effectiveListen, server.Handler()); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
