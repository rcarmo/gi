package codemode

import (
 "context"
 "encoding/json"
 "fmt"
 "io"
 "log/slog"
 "os/exec"
 "sort"
 "strings"
 "time"

 "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Broker is per-invocation and sequential. Config is a frozen operator policy;
// no user source can supply commands, change policy, or access server env values.
type Broker struct {
 config Config
 sessions map[string]*serverSession
 operations int
}
type serverSession struct { session *mcp.ClientSession; cmd *exec.Cmd; stop context.CancelFunc }
func NewBroker(cfg Config) (*Broker, error) {
 // Copy caller-owned maps/slices so authorization cannot change under a call.
 b, err := json.Marshal(cfg); if err != nil { return nil, err }
 var frozen Config
 if err := json.Unmarshal(b, &frozen); err != nil { return nil, err }
 if err := frozen.Validate(); err != nil { return nil, err }
 return &Broker{config: frozen, sessions: map[string]*serverSession{}}, nil
}
func (b *Broker) Close() {
 for name, s := range b.sessions {
  s.stop(); killProcess(s.cmd)
  _ = s.session.Close(); _ = s.cmd.Wait()
  delete(b.sessions, name)
 }
}
func (b *Broker) connect(ctx context.Context, name string) (*mcp.ClientSession, error) {
 if s := b.sessions[name]; s != nil { return s.session, nil }
 cfg, ok := b.config.Servers[name]; if !ok { return nil, fmt.Errorf("unknown MCP server %q", name) }
 lifetime, stop := context.WithCancel(ctx)
 cmd := exec.CommandContext(lifetime, cfg.Command, cfg.Args...)
 cmd.Env = []string{} // Deliberately never inherit Gi's environment/credentials.
 for k, v := range cfg.Env { cmd.Env = append(cmd.Env, k+"="+v) }
 cmd.Dir = cfg.Dir
 if cmd.Dir == "" { cmd.Dir = "/" }
 prepareProcess(cmd)
 cmd.Stderr = io.Discard // Server logs may contain credentials; never relay them.
 in, err := cmd.StdinPipe(); if err != nil { stop(); return nil, err }
 out, err := cmd.StdoutPipe(); if err != nil { stop(); in.Close(); return nil, err }
 if err := cmd.Start(); err != nil { stop(); in.Close(); out.Close(); return nil, fmt.Errorf("MCP server %s failed to start", name) }
 client := mcp.NewClient(&mcp.Implementation{Name:"gi-codemode-prototype", Version:"0.1"}, &mcp.ClientOptions{
  Capabilities: &mcp.ClientCapabilities{}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
 })
 handshake, cancel := context.WithTimeout(ctx, 5*time.Second); defer cancel()
 // Pin the well-supported protocol without SDK multi-round-trip call replay.
 session, err := client.Connect(handshake, &mcp.IOTransport{Reader:out, Writer:in, MaxLineLength:MaxFrame}, &mcp.ClientSessionOptions{ProtocolVersion:"2025-11-25"})
 if err != nil { stop(); killProcess(cmd); in.Close(); out.Close(); _ = cmd.Wait(); return nil, fmt.Errorf("MCP server %s initialization failed", name) }
 b.sessions[name] = &serverSession{session, cmd, stop}
 return session, nil
}
func (b *Broker) allowed(server, tool string) bool {
 for _, n := range b.config.Servers[server].AllowedTools { if n == tool { return true } }; return false
}
func toolName(server, name string) string { return "mcp__"+server+"__"+name }
func splitName(name string) (string,string,error) {
 p := strings.SplitN(name, "__", 3)
 if len(p)!=3 || p[0]!="mcp" || p[1]=="" || p[2]=="" { return "","",fmt.Errorf("expected mcp__server__tool") }
 return p[1],p[2],nil
}
// Refresh on each operation instead of caching stale tools. All pages and total
// catalogue bytes are bounded. SDK list-change support is not needed for this
// short-lived, pinned-protocol prototype.
func (b *Broker) list(ctx context.Context, server string) ([]*mcp.Tool,error) {
 s, err := b.connect(ctx,server); if err != nil { return nil,err }
 var out []*mcp.Tool; cursor := ""; seen := map[string]bool{}; bytes := 0
 for page:=0; page<16; page++ {
  r, err := s.ListTools(ctx,&mcp.ListToolsParams{Cursor:cursor}); if err != nil { return nil,fmt.Errorf("MCP server %s tools/list failed",server) }
  raw, err := json.Marshal(r); if err != nil { return nil,err }; bytes+=len(raw)
  if bytes>MaxFrame || len(out)+len(r.Tools)>2048 { return nil,fmt.Errorf("MCP catalogue budget exceeded") }
  for _, t := range r.Tools {
   if t==nil || t.Name=="" || seen[t.Name] { return nil,fmt.Errorf("invalid or duplicate MCP tool") }
   seen[t.Name]=true
   out=append(out,t)
  }
  if r.NextCursor=="" { return out,nil }; cursor=r.NextCursor
 }
 return nil,fmt.Errorf("MCP catalogue page limit exceeded")
}

type operation struct {
 Kind string `json:"kind"`
 Query string `json:"query,omitempty"`
 Server string `json:"server,omitempty"`
 Name string `json:"name,omitempty"`
 Arguments json.RawMessage `json:"arguments,omitempty"`
 Limit int `json:"limit,omitempty"`
 Offset int `json:"offset,omitempty"`
}
type summary struct { Name string `json:"name"`; Description string `json:"description"` }
func (b *Broker) handle(ctx context.Context, op operation) (json.RawMessage,error) {
 if err:=ctx.Err(); err!=nil { return nil,err }
 b.operations++; if b.operations>MaxOperations { return nil,fmt.Errorf("MCP operation budget exceeded") }
 var result any
 switch op.Kind {
 case "search":
  if op.Limit==0 { op.Limit=10 }
  if op.Limit<1 || op.Limit>50 || op.Offset<0 || op.Offset>32768 { return nil,fmt.Errorf("invalid search limit/offset") }
  names:=[]string{}
  if op.Server!="" { if _,ok:=b.config.Servers[op.Server];!ok { return nil,fmt.Errorf("unknown MCP server") }; names=append(names,op.Server) } else { for n:=range b.config.Servers { names=append(names,n) }; sort.Strings(names) }
  matches:=[]summary{}
  for _,server:=range names {
   list,err:=b.list(ctx,server); if err!=nil { return nil,err }
   for _,t:=range list {
    name:=toolName(server,t.Name)
    if b.allowed(server,t.Name) && strings.Contains(strings.ToLower(name+" "+t.Description),strings.ToLower(op.Query)) { matches=append(matches,summary{name,clipUTF8(t.Description,240)}) }
   }
  }
  sort.Slice(matches,func(i,j int)bool{return matches[i].Name<matches[j].Name})
  start:=min(op.Offset,len(matches)); end:=min(start+op.Limit,len(matches))
  result=struct { Tools []summary `json:"tools"`; Total int `json:"total"`; NextOffset *int `json:"nextOffset,omitempty"` }{Tools:matches[start:end],Total:len(matches)}
  if end<len(matches) { result=struct { Tools []summary `json:"tools"`; Total int `json:"total"`; NextOffset int `json:"nextOffset"` }{matches[start:end],len(matches),end} }
 case "describe","call":
  server,name,err:=splitName(op.Name); if err!=nil { return nil,err }
  if !b.allowed(server,name) { return nil,fmt.Errorf("MCP tool not allowed: %s",op.Name) }
  list,err:=b.list(ctx,server); if err!=nil { return nil,err }
  var tool *mcp.Tool
  for _,t:=range list { if t.Name==name { tool=t;break } }
  if tool==nil { return nil,fmt.Errorf("MCP tool no longer available: %s",op.Name) }
  if op.Kind=="describe" {
   result=struct{Name string `json:"name"`; Description string `json:"description"`; InputSchema any `json:"inputSchema"`; OutputSchema any `json:"outputSchema,omitempty"`}{op.Name,tool.Description,tool.InputSchema,tool.OutputSchema}
  } else {
   args:=op.Arguments; if len(args)==0 { args=json.RawMessage(`{}`) }
   var obj map[string]json.RawMessage
   if err:=json.Unmarshal(args,&obj); err!=nil || obj==nil { return nil,fmt.Errorf("MCP arguments must be an object") }
   // Authorization is checked again immediately before dispatch. No automatic
   // retry: cancellation/transport errors leave side-effect outcome unknown.
   if !b.allowed(server,name) { return nil,fmt.Errorf("MCP tool not allowed") }
   r,err:=b.sessions[server].session.CallTool(ctx,&mcp.CallToolParams{Name:name,Arguments:args})
   if err!=nil { return nil,fmt.Errorf("MCP tool call failed; outcome may be unknown (not retried)") }
   result=r // Preserve content, structuredContent and isError; never auto-emit.
  }
 default: return nil,fmt.Errorf("unknown MCP operation")
 }
 raw,err:=json.Marshal(result)
 if len(raw)>MaxFrame { return nil,fmt.Errorf("MCP result exceeds byte limit") }
 return raw,err
}
func clipUTF8(s string,n int) string {
 if len(s)<=n { return s }
 for n>0 && s[n]&0xc0==0x80 { n-- }
 return s[:n]+"…"
}
