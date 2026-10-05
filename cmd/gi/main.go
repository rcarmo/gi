package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/httpserver"
	"github.com/rcarmo/gi/internal/inference"
	gimcp "github.com/rcarmo/gi/internal/mcp"
	_ "github.com/rcarmo/gi/internal/peering/tsnetbackend" // tsnet, linked only into gi
	"github.com/rcarmo/gi/internal/store"
	storecache "github.com/rcarmo/gi/internal/store/cache"
	gitui "github.com/rcarmo/gi/internal/tui"
	"github.com/rcarmo/gi/internal/turn"
	"github.com/rcarmo/gi/internal/version"
	giweb "github.com/rcarmo/gi/internal/web"
)

func configureTUILogging(logFile string) {
	logFile = strings.TrimSpace(logFile)
	if logFile == "" {
		logFile = filepath.Join(config.DefaultStateDir("gi"), "gi-tui.log")
	}
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		log.SetOutput(io.Discard)
		return
	}
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.SetOutput(io.Discard)
		return
	}
	log.SetOutput(f)
}

// webOnlyFlagsSet lists explicitly set flags that only affect the web server.
func webOnlyFlagsSet() []string {
	webOnly := map[string]bool{"listen": true, "bind": true, "port": true, "tls-cert": true, "tls-key": true, "acme-domains": true, "acme-email": true, "acme-cache": true, "acme-accept-tos": true, "acme-http-listen": true, "pid-file": true}
	var set []string
	flag.Visit(func(f *flag.Flag) {
		if webOnly[f.Name] {
			set = append(set, "-"+f.Name)
		}
	})
	return set
}

func main() {
	defer finishFixtureProfiles()
	startProfiling()
	if len(os.Args) > 1 && (os.Args[1] == "-version" || os.Args[1] == "--version") {
		fmt.Println("gi " + version.String())
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "mcp" {
		os.Exit(runMCPCommand(os.Args[2:]))
	}
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
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
	dbPath := flag.String("db", config.DefaultTUIDBPath(), "SQLite database path")
	workspace := flag.String("workspace", config.DefaultWorkspaceRoot(), "Workspace root")
	model := flag.String("model", "", "Override default model (e.g. gemma4:latest)")
	logFile := flag.String("log-file", "", "Optional log file path")
	pidFile := flag.String("pid-file", "", "Optional pid file path")
	webMode := flag.Bool("web", false, "Run the web UI server instead of the terminal UI")
	_ = flag.Bool("tui", true, "Run the terminal UI (default; kept for compatibility)")
	tuiLayout := flag.String("tui-mode", "", "Terminal rendering: fullscreen or regular (native scrollback); default: tuiMode setting, else fullscreen")
	flag.Parse()
	startModelCatalogRefresh(*workspace)

	if !*webMode {
		// Server-only flags without -web are almost certainly a mistake;
		// fail loudly rather than silently opening a terminal UI.
		if flags := webOnlyFlagsSet(); len(flags) > 0 {
			return fmt.Errorf("web UI flags %s need -web (gi runs the terminal UI by default)", strings.Join(flags, ", "))
		}
		configureTUILogging(*logFile)
		if err := os.MkdirAll(filepath.Dir(*dbPath), 0o755); err != nil {
			return fmt.Errorf("create tui db dir: %w", err)
		}
		if err := gitui.RunMode(*dbPath, *workspace, *model, *tuiLayout); err != nil {
			log.Fatalf("tui: %v", err)
		}
		return nil
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
	processCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	engine := turn.NewWithRuntimeConfig(s, runtimeCfg, runtimeCfg.SystemPrompt)
	engine.EnableMCP() // Pi mcp.json servers (.gi first, then .pi); connects in the background
	defer engine.Close()
	server := giweb.New(s, engine, runtimeCfg)
	server.StartInboundWorkDispatcher(processCtx)
	if err := server.StartWorkspaceIndex(processCtx); err != nil {
		log.Printf("workspace index disabled: %v", err)
	}
	defer server.CloseWorkspaceIndex()
	defer server.CloseTerminals()
	defer server.CloseVNC()

	handler := server.Handler()
	var listeners []httpserver.Listener
	runHTTPServer := func(srv *http.Server, serve func() error, label string) error {
		listeners = append(listeners, httpserver.Listener{Server: srv, Serve: serve, Label: label})
		return httpserver.Run(processCtx, stop, 5*time.Second, listeners...)
	}
	if *acmeDomains != "" {
		domains := splitCSV(*acmeDomains)
		if len(domains) == 0 {
			return fmt.Errorf("acme-domains must include at least one domain")
		}
		cache, cacheLabel, err := acmeCacheFor(*acmeCache, s)
		if err != nil {
			return fmt.Errorf("acme cache: %w", err)
		}
		if !*acmeAcceptTOS {
			return fmt.Errorf("ACME requires -acme-accept-tos")
		}
		manager := &autocert.Manager{
			Cache:      cache,
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domains...),
			Email:      *acmeEmail,
		}
		if *acmeHTTPListen != "" {
			acmeSrv := &http.Server{Addr: *acmeHTTPListen, Handler: manager.HTTPHandler(nil)}
			log.Printf("Gi ACME HTTP-01/redirect listener on %s for %s", *acmeHTTPListen, strings.Join(domains, ","))
			listeners = append(listeners, httpserver.Listener{Server: acmeSrv, Serve: acmeSrv.ListenAndServe, Label: "acme-http"})
		}
		log.Printf("Gi HTTPS listening on %s using ACME domains=%s db=%s cache=%s", effectiveListen, strings.Join(domains, ","), *dbPath, cacheLabel)
		srv := &http.Server{Addr: effectiveListen, Handler: handler, TLSConfig: manager.TLSConfig()}
		return runHTTPServer(srv, func() error { return srv.ListenAndServeTLS("", "") }, "https/acme")
	}
	if *certFile != "" || *keyFile != "" {
		if *certFile == "" || *keyFile == "" {
			return fmt.Errorf("both -tls-cert and -tls-key are required for static HTTPS")
		}
		log.Printf("Gi HTTPS listening on %s using %s", effectiveListen, *dbPath)
		srv := &http.Server{Addr: effectiveListen, Handler: handler, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
		return runHTTPServer(srv, func() error { return srv.ListenAndServeTLS(*certFile, *keyFile) }, "https")
	}
	log.Printf("Gi HTTP listening on %s using %s", effectiveListen, *dbPath)
	srv := &http.Server{Addr: effectiveListen, Handler: handler}
	return runHTTPServer(srv, func() error { return srv.ListenAndServe() }, "http")
}

func acmeCacheFor(value string, s *store.Store) (autocert.Cache, string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "sqlite") || strings.EqualFold(value, "kv") {
		return storecache.NewSQLiteCache(s), "sqlite:kv_store/acme/autocert", nil
	}
	if strings.EqualFold(value, "vfs") {
		return storecache.NewVFSCache(s), "sqlite:vfs_files/acme-autocert", nil
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

// runMCPCommand runs `gi mcp ...` (Pi's `pi mcp`) without starting a session.
func runMCPCommand(args []string) int {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = config.DefaultWorkspaceRoot()
	}
	return gimcp.RunCommand(args, gimcp.CLIOptions{
		Cwd: cwd, UserPath: gimcp.UserConfigPath(), ProjectPath: gimcp.ProjectConfigPath(cwd),
		LogPath:         config.UserConfigCandidates("mcp.log")[0],
		CredentialsPath: config.UserConfigFile("mcp-auth.json"),
	})
}

// startProfiling serves net/http/pprof on GI_PPROF (e.g. 127.0.0.1:6060)
// when set, for CPU, allocation and goroutine profiles of a running gi. It
// listens on its own server, never on the web UI's handler.
func startProfiling() {
	addr := strings.TrimSpace(os.Getenv("GI_PPROF"))
	if addr == "" {
		return
	}
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(100000) // sample blocking events of 100µs and longer
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("pprof: %v", err)
		}
	}()
}

// startModelCatalogRefresh is Pi's startup model refresh, in the
// background and bounded to 15 seconds, unless PI_OFFLINE is set: expired
// Copilot credentials are refreshed (updating the account's model list), and
// with modelCatalogUrl set, models-store.json catalogues that are due.
func startModelCatalogRefresh(workspace string) {
	if os.Getenv("PI_OFFLINE") != "" {
		return
	}
	base := config.Load(workspace).ModelCatalogURL
	go func() {
		if err := inference.RefreshExpiredCopilotCredentials(); err != nil {
			log.Printf("models: copilot credentials: %v", err)
		}
		if base == "" {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := inference.RefreshModelCatalog(ctx, base, false); err != nil {
			log.Printf("models: catalogue refresh: %v", err)
		}
	}()
}
