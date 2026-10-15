// Package mcp is gi's Model Context Protocol client, compatible with Pi's
// MCP support (docs/mcp.md): the same mcp.json format and locations, the
// same validation and value expansion, stdio and streamable-HTTP transports,
// and Pi's server lifecycle. Tool exposure, tool_search and codemode build on
// it; see docs/implementation/plans/mcp-codemode-plan.md.
package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/config"
)

// Exposure values, as in Pi. "codemode-deferred" is an alias of codemode.
const (
	ExposureCodemode = "codemode"
	ExposureDeferred = "deferred"
	ExposureDirect   = "direct"
	ExposureHidden   = "hidden"
)

// DefaultTimeout is Pi's default per-request timeout.
const DefaultTimeout = 60 * time.Second

// ServerConfig is one mcpServers entry, validated but not yet expanded:
// ${VAR} and !command values are resolved only when connecting.
type ServerConfig struct {
	Name         string            `json:"-"`
	Transport    string            `json:"-"` // "stdio" or "http"
	Command      string            `json:"command,omitempty"`
	Args         []string          `json:"args,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	Cwd          string            `json:"cwd,omitempty"`
	URL          string            `json:"url,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	OAuth        json.RawMessage   `json:"oauth,omitempty"`
	Auth         *ProviderAuth     `json:"auth,omitempty"`
	Timeout      time.Duration     `json:"-"`
	Enabled      bool              `json:"-"`
	Description  string            `json:"description,omitempty"`
	Exposure     string            `json:"-"`
	ToolExposure map[string]string `json:"toolExposure,omitempty"`
	// toolExposureOrder keeps toolExposure keys in file order: among *
	// patterns the first match wins (Pi), and Go maps do not keep order.
	toolExposureOrder []string
	Source            string `json:"-"` // file that defined the entry
	Scope             string `json:"-"` // "global" (user mcp.json) or "project", as Pi labels them
	// Override is the project mcp.json whose entry overrides enabled,
	// exposure or toolExposure of this global server (Pi's override).
	Override string          `json:"-"`
	raw      json.RawMessage // the entry as written, for overrides
}

// ProviderAuth reuses provider credentials without an MCP OAuth flow.
type ProviderAuth struct {
	Provider string `json:"provider"`
}

func validateProviderURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return fmt.Errorf("auth.provider requires a valid HTTPS or loopback HTTP URL")
	}
	if u.Scheme == "https" {
		return nil
	}
	if u.Scheme == "http" {
		host := strings.ToLower(u.Hostname())
		if host == "localhost" {
			return nil
		}
		if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
			return nil
		}
	}
	return fmt.Errorf("auth.provider requires HTTPS or loopback HTTP")
}

// Config is the merged user and (trusted) project configuration.
type Config struct {
	Servers            map[string]ServerConfig
	AutoEnableCodemode bool
	// Errors lists invalid entries; they are skipped, never fatal (Pi).
	Errors []error
	// ProjectConfig is the trusted project's mcp.json, where /mcp saves
	// project overrides of global servers; "" when the project is not read.
	ProjectConfig string
}

// overrideKeys are what a project entry without command, url or type may
// set on the global server of the same name (Pi's OVERRIDE_KEYS).
var overrideKeys = []string{"enabled", "exposure", "toolExposure"}

// isOverride is Pi's isOverride: an entry without command, url or type
// overrides a server defined elsewhere instead of defining one.
func isOverride(entry map[string]json.RawMessage) bool {
	_, command := entry["command"]
	_, url := entry["url"]
	_, typ := entry["type"]
	return !command && !url && !typ
}

// Names returns server names in sorted order.
func (c Config) Names() []string {
	names := make([]string, 0, len(c.Servers))
	for name := range c.Servers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

type rawFile struct {
	MCPServers         map[string]json.RawMessage `json:"mcpServers"`
	AutoEnableCodemode *bool                      `json:"autoEnableCodemode"`
}

type rawServer struct {
	Type         string            `json:"type"`
	Command      string            `json:"command"`
	Args         []string          `json:"args"`
	Env          map[string]string `json:"env"`
	Cwd          string            `json:"cwd"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers"`
	OAuth        json.RawMessage   `json:"oauth"`
	Auth         json.RawMessage   `json:"auth"`
	Timeout      *float64          `json:"timeout"`
	Enabled      *bool             `json:"enabled"`
	Description  string            `json:"description"`
	Exposure     string            `json:"exposure"`
	ToolExposure map[string]string `json:"toolExposure"`
}

var serverNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// UserConfigPath is the user-level MCP configuration: ~/.gi/agent/mcp.json if
// it exists, else Pi's ~/.pi/agent/mcp.json.
func UserConfigPath() string { return config.UserConfigFile("mcp.json") }

// ProjectConfigPath is the project MCP configuration: .gi/mcp.json if it
// exists, else Pi's .pi/mcp.json.
func ProjectConfigPath(workspace string) string {
	return config.ProjectConfigFile(workspace, "mcp.json")
}

// LoadConfig reads the user configuration and, only when projectTrusted, the
// project one; a project entry replaces a user entry with the same name and a
// project autoEnableCodemode overrides the user value (Pi). gi has no project
// trust model yet, so callers pass false until one exists (issue #16).
func LoadConfig(userPath, projectPath string, projectTrusted bool) Config {
	cfg := Config{Servers: map[string]ServerConfig{}, AutoEnableCodemode: true}
	sources := []string{userPath}
	if projectTrusted && projectPath != "" {
		sources = append(sources, projectPath)
		cfg.ProjectConfig = projectPath
	}
	for _, path := range sources {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				cfg.Errors = append(cfg.Errors, fmt.Errorf("%s: %w", path, err))
			}
			continue
		}
		var file rawFile
		if err := json.Unmarshal(data, &file); err != nil {
			cfg.Errors = append(cfg.Errors, fmt.Errorf("%s: invalid JSON: %w", path, err))
			continue
		}
		if file.AutoEnableCodemode != nil {
			cfg.AutoEnableCodemode = *file.AutoEnableCodemode
		}
		names := make([]string, 0, len(file.MCPServers))
		for name := range file.MCPServers {
			names = append(names, name)
		}
		sort.Strings(names)
		seen := map[string]string{} // canonical name -> name, within this file
		for _, name := range names {
			var entry map[string]json.RawMessage
			if path != userPath && json.Unmarshal(file.MCPServers[name], &entry) == nil && entry != nil && isOverride(entry) {
				cfg.applyOverride(path, name, file.MCPServers[name])
				continue
			}
			server, err := parseServer(name, file.MCPServers[name], path)
			if err != nil {
				cfg.Errors = append(cfg.Errors, fmt.Errorf("%s: server %q: %w", path, name, err))
				continue
			}
			canon := canonicalName(name)
			if other, dup := seen[canon]; dup {
				cfg.Errors = append(cfg.Errors, fmt.Errorf("%s: server %q duplicates %q (names differing only in - and _ are the same server)", path, name, other))
				continue
			}
			seen[canon] = name
			server.Scope = "global"
			if path != userPath {
				server.Scope = "project"
			}
			// A later file (the project) replaces an earlier entry with the
			// same canonical name.
			for existing := range cfg.Servers {
				if canonicalName(existing) == canon {
					delete(cfg.Servers, existing)
				}
			}
			cfg.Servers[name] = server
		}
	}
	return cfg
}

// applyOverride is Pi's project override: the global server keeps its
// definition (and credentials the project could not set) and takes
// enabled, exposure and toolExposure from the project entry, validated as
// one merged entry.
func (c *Config) applyOverride(path, name string, raw json.RawMessage) {
	base, ok := c.Servers[name]
	if !ok {
		c.Errors = append(c.Errors, fmt.Errorf("%s: server %q needs \"command\" or \"url\", or a global server to override", path, name))
		return
	}
	entry, _ := parseOrderedObject(raw)
	for _, key := range entry.keys {
		if !slices.Contains(overrideKeys, key) {
			c.Errors = append(c.Errors, fmt.Errorf("%s: server %q: an override can only set %s", path, name, strings.Join(overrideKeys, ", ")))
			return
		}
	}
	merged, _ := parseOrderedObject(base.raw)
	for _, key := range entry.keys {
		merged.set(key, entry.values[key])
	}
	encoded, _ := merged.MarshalJSON()
	server, err := parseServer(name, encoded, base.Source)
	if err != nil {
		c.Errors = append(c.Errors, fmt.Errorf("%s: server %q: %w", path, name, err))
		return
	}
	server.Scope, server.Override = base.Scope, path
	c.Servers[name] = server
}

// canonicalName treats names that differ only in - and _ as the same server.
func canonicalName(name string) string { return strings.ReplaceAll(name, "-", "_") }

func parseServer(name string, raw json.RawMessage, source string) (ServerConfig, error) {
	if !serverNamePattern.MatchString(name) {
		return ServerConfig{}, fmt.Errorf("name may contain only letters, digits, _ and -")
	}
	var r rawServer
	dec := json.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&r); err != nil {
		return ServerConfig{}, fmt.Errorf("invalid entry: %w", err)
	}
	s := ServerConfig{Name: name, Command: r.Command, Args: r.Args, Env: r.Env, Cwd: r.Cwd, URL: r.URL,
		Headers: r.Headers, OAuth: r.OAuth, Description: strings.TrimSpace(r.Description), ToolExposure: r.ToolExposure,
		Timeout: DefaultTimeout, Enabled: true, Source: source, raw: raw}
	switch strings.ToLower(strings.TrimSpace(r.Type)) {
	case "":
	case "stdio":
		s.Transport = "stdio"
	case "http", "streamable-http":
		s.Transport = "http"
	case "sse":
		return ServerConfig{}, fmt.Errorf("legacy SSE transport is not supported; use the streamable HTTP URL")
	default:
		return ServerConfig{}, fmt.Errorf("type must be stdio, http or streamable-http")
	}
	hasCommand, hasURL := strings.TrimSpace(r.Command) != "", strings.TrimSpace(r.URL) != ""
	switch {
	case hasCommand && hasURL:
		return ServerConfig{}, fmt.Errorf("use either command or url, not both")
	case !hasCommand && !hasURL:
		return ServerConfig{}, fmt.Errorf("command or url is required")
	case hasCommand:
		if s.Transport == "http" {
			return ServerConfig{}, fmt.Errorf("type %q needs a url", r.Type)
		}
		s.Transport = "stdio"
		if len(r.Headers) > 0 || len(r.OAuth) > 0 || len(r.Auth) > 0 {
			return ServerConfig{}, fmt.Errorf("headers, oauth and auth apply only to url servers")
		}
	default:
		if s.Transport == "stdio" {
			return ServerConfig{}, fmt.Errorf("type stdio needs a command")
		}
		s.Transport = "http"
		if !strings.HasPrefix(r.URL, "http://") && !strings.HasPrefix(r.URL, "https://") {
			return ServerConfig{}, fmt.Errorf("url must be an http or https URL")
		}
		if len(r.Args) > 0 || len(r.Env) > 0 || r.Cwd != "" {
			return ServerConfig{}, fmt.Errorf("args, env and cwd apply only to command servers")
		}
		if err := validateOAuth(r.OAuth); err != nil {
			return ServerConfig{}, err
		}
		if len(r.Auth) > 0 {
			var auth ProviderAuth
			decoder := json.NewDecoder(bytes.NewReader(r.Auth))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&auth); err != nil || strings.TrimSpace(auth.Provider) == "" {
				return ServerConfig{}, fmt.Errorf("auth must contain a non-empty provider")
			}
			auth.Provider = strings.TrimSpace(auth.Provider)
			if !serverNamePattern.MatchString(auth.Provider) {
				return ServerConfig{}, fmt.Errorf("invalid auth provider name")
			}
			if len(r.OAuth) > 0 && string(r.OAuth) != "null" {
				return ServerConfig{}, fmt.Errorf("auth.provider and oauth are mutually exclusive")
			}
			for header := range r.Headers {
				if strings.EqualFold(header, "Authorization") {
					return ServerConfig{}, fmt.Errorf("auth.provider and Authorization header are mutually exclusive")
				}
			}
			if err := validateProviderURL(s.URL); err != nil {
				return ServerConfig{}, err
			}
			s.Auth = &auth
		}
	}
	if r.Timeout != nil {
		if *r.Timeout <= 0 {
			return ServerConfig{}, fmt.Errorf("timeout must be a positive number of seconds")
		}
		s.Timeout = time.Duration(*r.Timeout * float64(time.Second))
	}
	if r.Enabled != nil {
		s.Enabled = *r.Enabled
	}
	exposure, err := normalizeExposure(r.Exposure, ExposureCodemode)
	if err != nil {
		return ServerConfig{}, err
	}
	s.Exposure = exposure
	s.toolExposureOrder = orderedKeys(raw, "toolExposure")
	for pattern, value := range r.ToolExposure {
		if strings.TrimSpace(pattern) == "" {
			return ServerConfig{}, fmt.Errorf("toolExposure keys must be tool names or * patterns")
		}
		if _, err := normalizeExposure(value, ""); err != nil {
			return ServerConfig{}, fmt.Errorf("toolExposure %q: %w", pattern, err)
		}
	}
	return s, nil
}

func normalizeExposure(value, fallback string) (string, error) {
	switch v := strings.TrimSpace(value); v {
	case "":
		if fallback == "" {
			return "", fmt.Errorf("exposure must be codemode, deferred, direct or hidden")
		}
		return fallback, nil
	case ExposureCodemode, "codemode-deferred":
		return ExposureCodemode, nil
	case ExposureDeferred, ExposureDirect, ExposureHidden:
		return v, nil
	default:
		return "", fmt.Errorf(`exposure must be one of "codemode", "deferred", "direct", "hidden"`)
	}
}

// ToolExposureFor resolves a tool's exposure: an exact toolExposure name wins,
// then the first matching * pattern (in key order), then the server exposure.
func (s ServerConfig) ToolExposureFor(tool string) string {
	if v, ok := s.ToolExposure[tool]; ok {
		e, _ := normalizeExposure(v, s.Exposure)
		return e
	}
	order := s.toolExposureOrder
	if len(order) != len(s.ToolExposure) { // built in code rather than parsed
		order = order[:0:0]
		for p := range s.ToolExposure {
			order = append(order, p)
		}
		sort.Strings(order)
	}
	for _, p := range order {
		if !strings.Contains(p, "*") {
			continue
		}
		if globMatch(p, tool) {
			e, _ := normalizeExposure(s.ToolExposure[p], s.Exposure)
			return e
		}
	}
	return s.Exposure
}

func globMatch(pattern, name string) bool {
	re := "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), `\*`, ".*") + "$"
	ok, _ := regexp.MatchString(re, name)
	return ok
}

var envReference = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// commandTimeout bounds !command value expansion.
const commandTimeout = 10 * time.Second

// expandValue resolves Pi's value syntax for env and headers: a whole-value
// "!command" runs the command and uses its trimmed stdout; otherwise
// ${VAR} references are replaced from the environment (unset: empty).
func expandValue(ctx context.Context, value string, lookup func(string) (string, bool)) (string, error) {
	if strings.HasPrefix(value, "!") {
		ctx, cancel := context.WithTimeout(ctx, commandTimeout)
		defer cancel()
		out, err := exec.CommandContext(ctx, "sh", "-c", strings.TrimPrefix(value, "!")).Output()
		if err != nil {
			return "", fmt.Errorf("value command failed: %w", err)
		}
		return strings.TrimSpace(string(out)), nil
	}
	return envReference.ReplaceAllStringFunc(value, func(ref string) string {
		v, _ := lookup(envReference.FindStringSubmatch(ref)[1])
		return v
	}), nil
}

// expandHome turns a leading "~/" into the home directory (command, args, cwd).
func expandHome(value string) string {
	if strings.HasPrefix(value, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, value[2:])
		}
	}
	return value
}

// orderedKeys returns the keys of the object stored under field in raw, in
// file order (empty when absent or not an object).
func orderedKeys(raw json.RawMessage, field string) []string {
	var outer map[string]json.RawMessage
	if json.Unmarshal(raw, &outer) != nil {
		return nil
	}
	inner, ok := outer[field]
	if !ok {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(inner))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return nil
	}
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return keys
		}
		key, _ := tok.(string)
		keys = append(keys, key)
		var skip json.RawMessage
		if dec.Decode(&skip) != nil {
			return keys
		}
	}
	return keys
}

var loopbackHosts = []string{"localhost", "127.0.0.1", "[::1]"}

// isLoopbackRedirectURI is Pi's isLoopbackRedirectUri: an http URI that the
// loopback callback server can serve.
func isLoopbackRedirectURI(value string) bool {
	u, err := url.Parse(value)
	if err != nil {
		return false
	}
	host := u.Hostname()
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return u.Scheme == "http" && slices.Contains(loopbackHosts, strings.ToLower(host)) && u.RawQuery == "" && u.Fragment == ""
}

// validateOAuth is Pi's validateOAuth (core/mcp-servers.js): the "oauth"
// settings of a url server, with Pi's messages.
func validateOAuth(raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	var v map[string]json.RawMessage
	if json.Unmarshal(raw, &v) != nil || v == nil {
		return errors.New("oauth must be an object")
	}
	str := func(key string) (string, bool, bool) { // value, present, is a string
		field, ok := v[key]
		if !ok {
			return "", false, true
		}
		var s string
		err := json.Unmarshal(field, &s)
		return s, true, err == nil
	}
	if _, _, ok := str("clientId"); !ok {
		return errors.New("oauth.clientId must be a string")
	}
	if _, _, ok := str("clientSecret"); !ok {
		return errors.New("oauth.clientSecret must be a string")
	}
	var port *float64
	if field, ok := v["callbackPort"]; ok {
		var n float64
		if json.Unmarshal(field, &n) != nil || n != float64(int(n)) || n < 1 || n > 65535 {
			return errors.New("oauth.callbackPort must be a port number")
		}
		port = &n
	}
	callbackURL, hasCallback, ok := str("callbackUrl")
	if hasCallback {
		if !ok || !isLoopbackRedirectURI(callbackURL) {
			return errors.New("oauth.callbackUrl must be an http URI on localhost, 127.0.0.1, or [::1] without query or fragment")
		}
		u, _ := url.Parse(callbackURL)
		if p := u.Port(); p != "" && port != nil && p != strconv.Itoa(int(*port)) {
			return errors.New("oauth.callbackUrl and oauth.callbackPort name different ports")
		}
	}
	if _, _, ok := str("scope"); !ok {
		return errors.New("oauth.scope must be a string")
	}
	if name, present, ok := str("clientName"); present && (!ok || strings.TrimSpace(name) == "") {
		return errors.New("oauth.clientName must be a non-empty string")
	}
	if registration, present, ok := str("clientRegistration"); present && !(ok && registration == "dcr") {
		if !ok || registration != "cimd" {
			return errors.New(`oauth.clientRegistration must be "dcr" or "cimd"`)
		}
		_, hasID := v["clientId"]
		_, hasName := v["clientName"]
		if hasID || hasName {
			return errors.New(`oauth.clientRegistration "cimd" cannot be combined with oauth.clientId or oauth.clientName`)
		}
		if hasCallback {
			u, _ := url.Parse(callbackURL)
			if u.Hostname() == "::1" || urlPath(u) != "/callback" {
				return errors.New(`oauth.clientRegistration "cimd" requires oauth.callbackUrl on localhost or 127.0.0.1 with path /callback`)
			}
		}
	}
	if metadataURL, present, ok := str("authServerMetadataUrl"); present {
		u, err := url.Parse(metadataURL)
		host := ""
		if err == nil {
			host = u.Hostname()
			if strings.Contains(host, ":") {
				host = "[" + host + "]"
			}
		}
		if !ok || err != nil || u.Host == "" || !(u.Scheme == "https" || (u.Scheme == "http" && slices.Contains(loopbackHosts, host))) {
			return errors.New("oauth.authServerMetadataUrl must be an https URL, or http on localhost, 127.0.0.1, or [::1]")
		}
	}
	return nil
}
