package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/geiserx/lynxprompt-mcp/internal/mcpserver"
)

const testToken = "0123456789abcdef0123456789abcdef"

func newTestHTTP(t *testing.T, c httpConfig) *httptest.Server {
	t.Helper()
	s := mcpserver.New(client.New("http://127.0.0.1:1", ""), mcpserver.Options{})
	h, err := newHTTPHandler(s, c)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

const legacyInit = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`

const modernToolsList = `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientCapabilities":{},"io.modelcontextprotocol/clientInfo":{"name":"t","version":"1"}}}}`

func post(t *testing.T, url, body string, hdr map[string]string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var sb strings.Builder
	buf := make([]byte, 32<<10)
	for {
		n, err := resp.Body.Read(buf)
		sb.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return resp, sb.String()
}

func modernHeaders() map[string]string {
	return map[string]string{"MCP-Protocol-Version": "2026-07-28", "Mcp-Method": "tools/list"}
}

func TestHTTP_legacy_initialize_and_modern_tools_list_on_one_stateless_endpoint(t *testing.T) {
	srv := newTestHTTP(t, httpConfig{})
	resp, body := post(t, srv.URL+"/mcp", legacyInit, nil)
	if resp.StatusCode != 200 || !strings.Contains(body, `"protocolVersion":"2025-11-25"`) {
		t.Fatalf("legacy initialize: %d %s", resp.StatusCode, body)
	}
	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		t.Errorf("stateless server minted a session id %q", sid)
	}
	resp, body = post(t, srv.URL+"/mcp", modernToolsList, modernHeaders())
	if resp.StatusCode != 200 || !strings.Contains(body, `"resultType":"complete"`) || !strings.Contains(body, "search_blueprints") {
		t.Fatalf("modern tools/list: %d %s", resp.StatusCode, body)
	}
}

func TestHTTP_rejects_a_cross_site_origin(t *testing.T) {
	srv := newTestHTTP(t, httpConfig{})
	resp, _ := post(t, srv.URL+"/mcp", legacyInit, map[string]string{"Origin": "https://evil.example", "Sec-Fetch-Site": "cross-site"})
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("cross-site request: status %d, want 403", resp.StatusCode)
	}
	// Positive control: a trusted origin passes.
	srv = newTestHTTP(t, httpConfig{AllowedOrigins: []string{"https://app.example"}})
	resp, _ = post(t, srv.URL+"/mcp", legacyInit, map[string]string{"Origin": "https://app.example", "Sec-Fetch-Site": "cross-site"})
	if resp.StatusCode != 200 {
		t.Errorf("trusted origin: status %d, want 200", resp.StatusCode)
	}
}

func TestHTTP_rejects_an_oversized_body(t *testing.T) {
	srv := newTestHTTP(t, httpConfig{})
	big := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"pad":"` + strings.Repeat("a", maxRequestBytes) + `"}}`
	resp, _ := post(t, srv.URL+"/mcp", big, nil)
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("oversized body: status %d, want 413", resp.StatusCode)
	}
}

func TestHTTP_bearer_auth(t *testing.T) {
	srv := newTestHTTP(t, httpConfig{AuthToken: testToken})
	cases := []struct {
		header string
		want   int
	}{
		{"", 401},
		{testToken, 401},                   // no scheme
		{"Basic " + testToken, 401},        // wrong scheme
		{"Bearer wrong", 401},              // wrong token
		{"Bearer " + testToken + "x", 401}, // longer token
		{"Bearer " + testToken, 200},       // correct
		{"bearer " + testToken, 200},       // scheme is case-insensitive
	}
	for _, tc := range cases {
		hdr := map[string]string{}
		if tc.header != "" {
			hdr["Authorization"] = tc.header
		}
		resp, _ := post(t, srv.URL+"/mcp", legacyInit, hdr)
		if resp.StatusCode != tc.want {
			t.Errorf("Authorization %q: status %d, want %d", tc.header, resp.StatusCode, tc.want)
		}
	}
}

func TestIsLoopbackAddr(t *testing.T) {
	for addr, want := range map[string]bool{
		"127.0.0.1:8080": true,
		"localhost:8080": true,
		"[::1]:8080":     true,
		":8080":          false, // binds every interface
		"0.0.0.0:8080":   false,
		"10.0.0.5:8080":  false,
		"example.com:80": false,
		"garbage":        false,
	} {
		if got := isLoopbackAddr(addr); got != want {
			t.Errorf("isLoopbackAddr(%q) = %v, want %v", addr, got, want)
		}
	}
}

func TestHTTPConfig_validate(t *testing.T) {
	cases := []struct {
		c  httpConfig
		ok bool
	}{
		{httpConfig{Addr: "127.0.0.1:8080"}, true},
		{httpConfig{Addr: "0.0.0.0:8080"}, false},
		{httpConfig{Addr: "0.0.0.0:8080", AuthToken: "short"}, false},
		{httpConfig{Addr: "0.0.0.0:8080", AuthToken: testToken}, true},
		{httpConfig{Addr: ":8080"}, false},
	}
	for _, tc := range cases {
		if err := tc.c.validate(); (err == nil) != tc.ok {
			t.Errorf("validate(%+v) = %v, want ok=%v", tc.c, err, tc.ok)
		}
	}
}

func TestNewHTTPServer_sets_timeouts(t *testing.T) {
	s := newHTTPServer("127.0.0.1:0", http.NotFoundHandler())
	if s.ReadHeaderTimeout == 0 || s.ReadTimeout == 0 || s.IdleTimeout == 0 || s.MaxHeaderBytes == 0 {
		t.Errorf("http.Server is missing limits: %+v", s)
	}
}

func TestNewHTTPHandler_rejects_a_bad_allowed_origin(t *testing.T) {
	s := mcpserver.New(client.New("http://127.0.0.1:1", ""), mcpserver.Options{})
	if _, err := newHTTPHandler(s, httpConfig{AllowedOrigins: []string{"not a url"}}); err == nil {
		t.Error("expected an error for a malformed origin")
	}
}
