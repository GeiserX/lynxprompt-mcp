package mcpserver

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/server"
)

// stdioPeer drives the real server through the SDK's stdio transport, one
// newline-delimited JSON-RPC message at a time, exactly as a client process would.
type stdioPeer struct {
	t      *testing.T
	in     *io.PipeWriter
	out    *bufio.Reader
	nextID int
}

func startStdio(t *testing.T, opts Options, apiURL string) *stdioPeer {
	t.Helper()
	if apiURL == "" {
		apiURL = "http://127.0.0.1:1"
	}
	s := New(client.New(apiURL, "lp_test"), opts)
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = server.NewStdioServer(s).Listen(ctx, inR, outW)
		outW.Close()
	}()
	t.Cleanup(func() {
		inW.Close()
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("stdio server did not stop")
		}
	})
	return &stdioPeer{t: t, in: inW, out: bufio.NewReader(outR)}
}

func modernMeta() map[string]any {
	return map[string]any{
		"io.modelcontextprotocol/protocolVersion":    "2026-07-28",
		"io.modelcontextprotocol/clientCapabilities": map[string]any{},
		"io.modelcontextprotocol/clientInfo":         map[string]any{"name": "protocol-test", "version": "1.0.0"},
	}
}

func (p *stdioPeer) send(msg map[string]any) {
	p.t.Helper()
	b, err := json.Marshal(msg)
	if err != nil {
		p.t.Fatal(err)
	}
	if _, err := p.in.Write(append(b, '\n')); err != nil {
		p.t.Fatalf("write: %v", err)
	}
}

// call sends a request and returns the response with the same id.
func (p *stdioPeer) call(method string, params map[string]any) map[string]any {
	p.t.Helper()
	p.nextID++
	id := p.nextID
	msg := map[string]any{"jsonrpc": "2.0", "id": id, "method": method}
	if params != nil {
		msg["params"] = params
	}
	p.send(msg)
	type line struct {
		b   []byte
		err error
	}
	for {
		ch := make(chan line, 1)
		go func() {
			b, err := p.out.ReadBytes('\n')
			ch <- line{b, err}
		}()
		var l line
		select {
		case l = <-ch:
		case <-time.After(10 * time.Second):
			p.t.Fatalf("%s: no response within 10s", method)
		}
		if l.err != nil {
			p.t.Fatalf("%s: read: %v", method, l.err)
		}
		var resp map[string]any
		if err := json.Unmarshal(l.b, &resp); err != nil {
			p.t.Fatalf("%s: stdout carried a non-JSON line %q", method, l.b)
		}
		if got, ok := resp["id"].(float64); ok && int(got) == id {
			return resp
		}
	}
}

func modernCall(p *stdioPeer, method string, params map[string]any) map[string]any {
	if params == nil {
		params = map[string]any{}
	}
	params["_meta"] = modernMeta()
	return p.call(method, params)
}

func result(t *testing.T, resp map[string]any) map[string]any {
	t.Helper()
	r, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected a result, got %v", resp)
	}
	return r
}

func toolNames(t *testing.T, r map[string]any) []string {
	t.Helper()
	list, _ := r["tools"].([]any)
	var names []string
	for _, it := range list {
		names = append(names, it.(map[string]any)["name"].(string))
	}
	return names
}

var allTools = []string{
	"create_blueprint", "create_hierarchy", "delete_blueprint", "delete_hierarchy", "get_blueprint",
	"get_hierarchy", "list_blueprints", "list_hierarchies", "search_blueprints", "update_blueprint",
}

func TestProtocol_modern_discover_advertises_2026_07_28(t *testing.T) {
	p := startStdio(t, Options{}, "")
	r := result(t, modernCall(p, "server/discover", nil))
	var versions []string
	for _, v := range r["supportedVersions"].([]any) {
		versions = append(versions, v.(string))
	}
	for _, want := range []string{"2026-07-28", "2025-11-25", "2024-11-05"} {
		if !slices.Contains(versions, want) {
			t.Errorf("supportedVersions %v lacks %s", versions, want)
		}
	}
	if !strings.Contains(fmt.Sprint(r["instructions"]), "expected_checksum") {
		t.Errorf("discover should carry the server instructions, got %v", r["instructions"])
	}
}

func TestProtocol_modern_tools_list_is_complete_and_cacheable(t *testing.T) {
	p := startStdio(t, Options{}, "")
	r := result(t, modernCall(p, "tools/list", nil))
	if r["resultType"] != "complete" {
		t.Errorf("resultType = %v, want complete", r["resultType"])
	}
	if ttl, ok := r["ttlMs"].(float64); !ok || ttl <= 0 {
		t.Errorf("ttlMs = %v, want a positive cache lifetime", r["ttlMs"])
	}
	if r["cacheScope"] != "public" {
		t.Errorf("cacheScope = %v, want public", r["cacheScope"])
	}
	names := toolNames(t, r)
	if !slices.Equal(names, allTools) {
		t.Errorf("tools = %v, want %v (sorted, deterministic)", names, allTools)
	}
}

func TestProtocol_modern_ping_is_removed(t *testing.T) {
	p := startStdio(t, Options{}, "")
	resp := modernCall(p, "ping", nil)
	e, ok := resp["error"].(map[string]any)
	if !ok || e["code"] != float64(-32601) {
		t.Errorf("modern ping: want error -32601, got %v", resp)
	}
}

func TestProtocol_legacy_initialize_then_tools_list_still_works(t *testing.T) {
	p := startStdio(t, Options{}, "")
	init := result(t, p.call("initialize", map[string]any{
		"protocolVersion": "2025-11-25",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "legacy-client", "version": "1.0.0"},
	}))
	if init["protocolVersion"] != "2025-11-25" {
		t.Errorf("negotiated %v, want 2025-11-25", init["protocolVersion"])
	}
	info, _ := init["serverInfo"].(map[string]any)
	if info["name"] != Name || info["title"] != "LynxPrompt" {
		t.Errorf("serverInfo = %v", info)
	}
	p.send(map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})

	r := result(t, p.call("tools/list", nil))
	if names := toolNames(t, r); !slices.Equal(names, allTools) {
		t.Errorf("legacy tools/list = %v", names)
	}
	if _, has := r["resultType"]; has {
		t.Error("legacy responses must keep the old shape (no resultType)")
	}
}

func TestProtocol_read_only_lists_only_read_tools(t *testing.T) {
	p := startStdio(t, Options{ReadOnly: true}, "")
	names := toolNames(t, result(t, modernCall(p, "tools/list", nil)))
	want := []string{"get_blueprint", "get_hierarchy", "list_blueprints", "list_hierarchies", "search_blueprints"}
	if !slices.Equal(names, want) {
		t.Errorf("read-only tools = %v, want %v", names, want)
	}
}

// fixtures are shaped like real LynxPrompt API answers, nulls included.
var fixtures = map[string]string{
	"/api/blueprints":          `{"templates":[{"id":"bp_a","name":"Next.js","description":"","author":"Ana","downloads":3,"likes":1,"tags":["web"],"tier":"SHORT","type":"AGENTS_MD","category":"web","isOfficial":false,"aiAssisted":false,"isOwner":false}],"popularTags":[],"total":1,"hasMore":false}`,
	"/api/v1/blueprints":       `{"blueprints":[{"id":"bp_a","name":"Mine","description":null,"type":"CLAUDE_MD","tier":"SHORT","category":"other","tags":[],"visibility":"PRIVATE","downloads":0,"favorites":0,"created_at":"2026-01-01T00:00:00.000Z","updated_at":"2026-01-02T00:00:00.000Z","hierarchy_id":null,"parent_id":null,"repository_path":null,"content_checksum":"0123456789abcdef"}],"total":1,"limit":50,"offset":0,"has_more":false}`,
	"/api/v1/blueprints/bp_a":  `{"blueprint":{"id":"bp_a","name":"Mine","description":null,"content":"# Rules\nbe nice","type":"CLAUDE_MD","tier":"SHORT","category":"other","tags":[],"visibility":"PRIVATE","variables":null,"downloads":0,"favorites":0,"ai_assisted":false,"showcase_url":null,"created_at":"2026-01-01T00:00:00.000Z","updated_at":"2026-01-02T00:00:00.000Z","hierarchy_id":"ha_h","parent_id":null,"repository_path":"AGENTS.md","content_checksum":"0123456789abcdef"}}`,
	"/api/v1/hierarchies":      `{"hierarchies":[{"id":"ha_h","name":"mono","description":null,"repository_root":"github.com/acme/mono","blueprint_count":2,"created_at":"2026-01-01T00:00:00.000Z","updated_at":"2026-01-01T00:00:00.000Z"}],"total":1,"limit":50,"offset":0,"has_more":false}`,
	"/api/v1/hierarchies/ha_h": `{"hierarchy":{"id":"ha_h","name":"mono","description":null,"repository_root":"github.com/acme/mono","created_at":"2026-01-01T00:00:00.000Z","updated_at":"2026-01-01T00:00:00.000Z"},"blueprints":[{"id":"bp_a","name":"Mine","description":null,"type":"AGENTS_MD","tier":"SHORT","repository_path":"AGENTS.md","parent_id":null,"content_checksum":"0123456789abcdef","visibility":"PRIVATE","downloads":0,"favorites":0,"created_at":"2026-01-01T00:00:00.000Z","updated_at":"2026-01-01T00:00:00.000Z"}],"tree":[{"id":"bp_a","children":[]}],"total_blueprints":1}`,
}

func TestProtocol_read_tools_return_structured_content_that_matches_their_schema(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body, ok := fixtures[r.URL.Path]; ok {
			w.Write([]byte(body))
			return
		}
		w.WriteHeader(404)
		w.Write([]byte(`{"error":"Blueprint not found"}`))
	}))
	t.Cleanup(api.Close)

	p := startStdio(t, Options{ValidateOutput: true}, api.URL)
	calls := map[string]map[string]any{
		"search_blueprints": {"query": "next"},
		"list_blueprints":   {},
		"get_blueprint":     {"id": "bp_a"},
		"list_hierarchies":  {},
		"get_hierarchy":     {"id": "ha_h"},
	}
	names := make([]string, 0, len(calls))
	for n := range calls {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		r := result(t, modernCall(p, "tools/call", map[string]any{"name": name, "arguments": calls[name]}))
		if r["isError"] == true {
			t.Errorf("%s: %v", name, r["content"])
			continue
		}
		if _, ok := r["structuredContent"].(map[string]any); !ok {
			t.Errorf("%s: no structuredContent in %v", name, r)
		}
	}

	// Input validation happens on the server, before any API call.
	r := result(t, modernCall(p, "tools/call", map[string]any{"name": "list_blueprints", "arguments": map[string]any{"limit": 1000}}))
	if r["isError"] != true {
		t.Errorf("limit=1000 should fail schema validation, got %v", r)
	}
	r = result(t, modernCall(p, "tools/call", map[string]any{"name": "get_blueprint", "arguments": map[string]any{"id": "bp_a", "bogus": 1}}))
	if r["isError"] != true {
		t.Errorf("unknown arguments should be rejected, got %v", r)
	}
}
