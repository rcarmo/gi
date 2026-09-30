// Package codemode implements the opt-in Joker/MCP CLI prototype. It does not
// reuse the privileged scripting bridge or register model-visible tools.
package codemode

import (
 "encoding/json"
 "fmt"
 "io"
 "os"
 "path/filepath"
 "regexp"
 "strings"
)

const (
 MaxSource = 64 << 10
 MaxFrame = 1 << 20
 MaxOutput = 32 << 10
 MaxOperations = 32
)

type Config struct {
 Servers map[string]ServerConfig `json:"servers"`
}
type ServerConfig struct {
 Command string `json:"command"`
 Args []string `json:"args,omitempty"`
 // Env is an explicit environment, NOT an overlay on Gi's credentials.
 Env map[string]string `json:"env,omitempty"`
 Dir string `json:"dir,omitempty"`
 AllowedTools []string `json:"allowedTools"`
}
var serverName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

func LoadConfig(path string) (Config, error) {
 var cfg Config
 raw, err := ReadBounded(path, MaxSource)
 if err != nil { return cfg, err }
 dec := json.NewDecoder(strings.NewReader(string(raw)))
 dec.DisallowUnknownFields()
 if err := dec.Decode(&cfg); err != nil { return cfg, fmt.Errorf("codemode config: %w", err) }
 if err := dec.Decode(new(any)); err != io.EOF { return cfg, fmt.Errorf("codemode config: trailing data") }
 return cfg, cfg.Validate()
}
func (c Config) Validate() error {
 if len(c.Servers) > 16 { return fmt.Errorf("at most 16 MCP servers allowed") }
 for name, s := range c.Servers {
  if !serverName.MatchString(name) || strings.Contains(name, "__") { return fmt.Errorf("invalid server name %q", name) }
  if !filepath.IsAbs(s.Command) { return fmt.Errorf("server %s requires an absolute command path", name) }
  if s.Dir != "" && !filepath.IsAbs(s.Dir) { return fmt.Errorf("server %s requires an absolute dir", name) }
  if len(s.AllowedTools) == 0 { return fmt.Errorf("server %s requires an explicit nonempty allowedTools list", name) }
  seen := map[string]bool{}
  for _, n := range s.AllowedTools {
   if n == "" || n == "*" || seen[n] { return fmt.Errorf("server %s has invalid/duplicate allowedTools entry", name) }
   seen[n] = true
  }
  for k, v := range s.Env {
   if k == "" || strings.ContainsAny(k, "=\x00") || strings.ContainsRune(v, 0) { return fmt.Errorf("server %s has invalid environment", name) }
  }
 }
 return nil
}
func ReadBounded(path string, max int) ([]byte, error) {
 f, err := os.Open(path)
 if err != nil { return nil, err }; defer f.Close()
 b, err := io.ReadAll(io.LimitReader(f, int64(max)+1))
 if len(b) > max { return nil, fmt.Errorf("file exceeds %d bytes", max) }
 return b, err
}
