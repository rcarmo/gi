package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/acme/autocert"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	gitui "github.com/rcarmo/gi/internal/tui"
	"github.com/rcarmo/gi/internal/turn"
	giweb "github.com/rcarmo/gi/internal/web"
)

func main() {
	listen := flag.String("listen", "", "HTTP listen address (overrides -bind/-port)")
	bind := flag.String("bind", "127.0.0.1", "Bind address / interface host")
	port := flag.Int("port", 8081, "HTTP port")
	certFile := flag.String("tls-cert", "", "TLS certificate file for HTTPS")
	keyFile := flag.String("tls-key", "", "TLS private key file for HTTPS")
	acmeDomains := flag.String("acme-domains", "", "Comma-separated domains for ACME/Let's Encrypt HTTPS")
	acmeEmail := flag.String("acme-email", "", "Contact email for ACME registration")
	acmeCache := flag.String("acme-cache", "sqlite", "ACME certificate cache: sqlite, vfs, or filesystem directory path")
	acmeAcceptTOS := flag.Bool("acme-accept-tos", false, "Accept the ACME CA terms of service")
	acmeHTTPListen := flag.String("acme-http-listen", ":http", "HTTP listen address for ACME HTTP-01 challenges and redirects; empty disables")
	dbPath := flag.String("db", "./gi.db", "SQLite database path")
	workspace := flag.String("workspace", "/workspace", "Workspace root")
	model := flag.String("model", "", "Override default model (e.g. gemma4:latest)")
	logFile := flag.String("log-file", "", "Optional log file path")
	pidFile := flag.String("pid-file", "", "Optional pid file path")
	tuiMode := flag.Bool("tui", false, "Run the terminal UI instead of the web server")
	flag.Parse()

	if *tuiMode {
		if err := gitui.Run(*dbPath, *workspace, *model); err != nil {
			log.Fatalf("tui: %v", err)
		}
		return
	}

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
	if *model != "" {
		runtimeCfg.DefaultModel = *model
	}
	engine := turn.NewWithRuntimeConfig(s, runtimeCfg, runtimeCfg.SystemPrompt)
	server := giweb.New(s, engine, runtimeCfg)
	server.StartInboundWorkDispatcher(context.Background())

	handler := server.Handler()
	if *acmeDomains != "" {
		domains := splitCSV(*acmeDomains)
		if len(domains) == 0 {
			log.Fatalf("acme-domains must include at least one domain")
		}
		cache, cacheLabel, err := acmeCacheFor(*acmeCache, s)
		if err != nil {
			log.Fatalf("acme cache: %v", err)
		}
		if !*acmeAcceptTOS {
			log.Fatalf("ACME requires -acme-accept-tos")
		}
		manager := &autocert.Manager{
			Cache:      cache,
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domains...),
			Email:      *acmeEmail,
		}
		if *acmeHTTPListen != "" {
			go func() {
				log.Printf("Gi ACME HTTP-01/redirect listener on %s for %s", *acmeHTTPListen, strings.Join(domains, ","))
				if err := http.ListenAndServe(*acmeHTTPListen, manager.HTTPHandler(nil)); err != nil {
					log.Printf("acme http listener: %v", err)
				}
			}()
		}
		log.Printf("Gi HTTPS listening on %s using ACME domains=%s db=%s cache=%s", effectiveListen, strings.Join(domains, ","), *dbPath, cacheLabel)
		srv := &http.Server{Addr: effectiveListen, Handler: handler, TLSConfig: manager.TLSConfig()}
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			log.Fatalf("listen https/acme: %v", err)
		}
		return
	}
	if *certFile != "" || *keyFile != "" {
		if *certFile == "" || *keyFile == "" {
			log.Fatalf("both -tls-cert and -tls-key are required for static HTTPS")
		}
		log.Printf("Gi HTTPS listening on %s using %s", effectiveListen, *dbPath)
		srv := &http.Server{Addr: effectiveListen, Handler: handler, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
		if err := srv.ListenAndServeTLS(*certFile, *keyFile); err != nil {
			log.Fatalf("listen https: %v", err)
		}
		return
	}
	log.Printf("Gi HTTP listening on %s using %s", effectiveListen, *dbPath)
	if err := http.ListenAndServe(effectiveListen, handler); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func acmeCacheFor(value string, s *store.Store) (autocert.Cache, string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "sqlite") || strings.EqualFold(value, "kv") {
		return store.NewACMESQLiteCache(s), "sqlite:kv_store/acme/autocert", nil
	}
	if strings.EqualFold(value, "vfs") {
		return store.NewACMEVFSCache(s), "sqlite:vfs_files/acme-autocert", nil
	}
	return autocert.DirCache(value), value, nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
