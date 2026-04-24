package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// newTestServer creates an httptest.Server that records the last request and
// responds with the given status and body.
func newTestServer(t *testing.T, status int, body string) (*httptest.Server, *http.Request) {
	t.Helper()
	var lastReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the body so callers can inspect it later.
		b, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(strings.NewReader(string(b)))
		lastReq = r.Clone(r.Context())
		lastReq.Body = io.NopCloser(strings.NewReader(string(b)))
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, lastReq
}

// routingServer responds based on path for multi-endpoint tests.
func routingServer(t *testing.T, routes map[string]string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if body, ok := routes[path]; ok {
			w.WriteHeader(200)
			w.Write([]byte(body))
			return
		}
		w.WriteHeader(404)
		w.Write([]byte("not found"))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// ---------------------------------------------------------------------------
// New
// ---------------------------------------------------------------------------

func TestNew_trims_trailing_slash_from_base(t *testing.T) {
	c := New("https://example.com/", "tok")
	if c.base != "https://example.com" {
		t.Errorf("base = %q, want trailing slash trimmed", c.base)
	}
}

func TestNew_keeps_base_without_trailing_slash(t *testing.T) {
	c := New("https://example.com", "tok")
	if c.base != "https://example.com" {
		t.Errorf("base = %q", c.base)
	}
}

func TestNew_sets_token(t *testing.T) {
	c := New("https://example.com", "my-token")
	if c.token != "my-token" {
		t.Errorf("token = %q, want %q", c.token, "my-token")
	}
}

// ---------------------------------------------------------------------------
// buildURL
// ---------------------------------------------------------------------------

func TestBuildURL_without_query_params(t *testing.T) {
	c := New("https://example.com", "")
	got := c.buildURL("/api/v1/test", nil)
	want := "https://example.com/api/v1/test"
	if got != want {
		t.Errorf("buildURL = %q, want %q", got, want)
	}
}

func TestBuildURL_with_query_params(t *testing.T) {
	c := New("https://example.com", "")
	q := make(map[string][]string)
	q["q"] = []string{"hello"}
	got := c.buildURL("/api/blueprints", q)
	want := "https://example.com/api/blueprints?q=hello"
	if got != want {
		t.Errorf("buildURL = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// do — auth header
// ---------------------------------------------------------------------------

func TestDo_sets_auth_header_when_token_present(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "secret-token")
	req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
	_, err := c.do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "Bearer secret-token" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer secret-token")
	}
}

func TestDo_omits_auth_header_when_token_empty(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "")
	req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
	_, err := c.do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "" {
		t.Errorf("Authorization should be empty, got %q", gotAuth)
	}
}

func TestDo_returns_error_on_4xx_status(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte("forbidden"))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "tok")
	req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
	_, err := c.do(req)
	if err == nil {
		t.Fatal("expected error for 403 status")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error should mention status code 403, got: %v", err)
	}
	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("error should contain body, got: %v", err)
	}
}

func TestDo_returns_error_on_5xx_status(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal error"))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "")
	req, _ := http.NewRequest("GET", srv.URL+"/x", nil)
	_, err := c.do(req)
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should mention 500, got: %v", err)
	}
}

func TestDo_returns_body_on_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "")
	req, _ := http.NewRequest("GET", srv.URL+"/ok", nil)
	body, err := c.do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Errorf("body = %q", string(body))
	}
}

// ---------------------------------------------------------------------------
// ListBlueprints
// ---------------------------------------------------------------------------

func TestListBlueprints_returns_response_body(t *testing.T) {
	srv := routingServer(t, map[string]string{
		"/api/v1/blueprints": `[{"id":"1","name":"bp1"}]`,
	})
	c := New(srv.URL, "tok")
	body, err := c.ListBlueprints(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(body), "bp1") {
		t.Errorf("expected bp1 in body, got %s", body)
	}
}

func TestListBlueprints_propagates_error_on_failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte("unauthorized"))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "")
	_, err := c.ListBlueprints(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// GetBlueprint
// ---------------------------------------------------------------------------

func TestGetBlueprint_includes_id_in_path(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"abc"}`))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "tok")
	_, err := c.GetBlueprint(context.Background(), "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/blueprints/abc" {
		t.Errorf("path = %q, want /api/v1/blueprints/abc", gotPath)
	}
}

func TestGetBlueprint_escapes_special_characters_in_id(t *testing.T) {
	var gotRawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRawPath = r.URL.RawPath
		if gotRawPath == "" {
			gotRawPath = r.URL.Path
		}
		w.WriteHeader(200)
		w.Write([]byte("{}"))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "tok")
	_, err := c.GetBlueprint(context.Background(), "a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(gotRawPath, "a/b") {
		t.Errorf("id should be escaped, got raw path %q", gotRawPath)
	}
	if !strings.Contains(gotRawPath, "a%2Fb") {
		t.Errorf("expected escaped id a%%2Fb in path, got %q", gotRawPath)
	}
}

// ---------------------------------------------------------------------------
// ListHierarchies
// ---------------------------------------------------------------------------

func TestListHierarchies_returns_response_body(t *testing.T) {
	srv := routingServer(t, map[string]string{
		"/api/v1/hierarchies": `[{"id":"h1"}]`,
	})
	c := New(srv.URL, "tok")
	body, err := c.ListHierarchies(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(body), "h1") {
		t.Errorf("expected h1, got %s", body)
	}
}

// ---------------------------------------------------------------------------
// GetHierarchy
// ---------------------------------------------------------------------------

func TestGetHierarchy_includes_id_in_path(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "tok")
	_, err := c.GetHierarchy(context.Background(), "h99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/hierarchies/h99" {
		t.Errorf("path = %q", gotPath)
	}
}

// ---------------------------------------------------------------------------
// GetUser
// ---------------------------------------------------------------------------

func TestGetUser_hits_correct_endpoint(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"email":"u@x.com"}`))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "tok")
	body, err := c.GetUser(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/user" {
		t.Errorf("path = %q", gotPath)
	}
	if !strings.Contains(string(body), "u@x.com") {
		t.Errorf("body = %s", body)
	}
}

// ---------------------------------------------------------------------------
// SearchBlueprints
// ---------------------------------------------------------------------------

func TestSearchBlueprints_sends_all_query_params(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(200)
		w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "")
	_, err := c.SearchBlueprints(context.Background(), "test", "cat1", "AGENTS_MD", "go,mcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"q=test", "category=cat1", "type=AGENTS_MD", "tags=go"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query %q missing %q", gotQuery, want)
		}
	}
}

func TestSearchBlueprints_omits_empty_optional_params(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(200)
		w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "")
	_, err := c.SearchBlueprints(context.Background(), "test", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(gotQuery, "category") {
		t.Errorf("empty category should not appear: %q", gotQuery)
	}
	if strings.Contains(gotQuery, "type") {
		t.Errorf("empty type should not appear: %q", gotQuery)
	}
	if strings.Contains(gotQuery, "tags") {
		t.Errorf("empty tags should not appear: %q", gotQuery)
	}
}

// ---------------------------------------------------------------------------
// CreateBlueprint
// ---------------------------------------------------------------------------

func TestCreateBlueprint_sends_POST_with_json_body(t *testing.T) {
	var gotMethod, gotCT string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotCT = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":"new1"}`))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "tok")
	body := map[string]any{"name": "bp", "content": "hello", "type": "CLAUDE_MD"}
	resp, err := c.CreateBlueprint(context.Background(), body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q", gotMethod)
	}
	if gotCT != "application/json" {
		t.Errorf("content-type = %q", gotCT)
	}
	if !strings.Contains(string(gotBody), `"name":"bp"`) {
		t.Errorf("body = %s", gotBody)
	}
	if !strings.Contains(string(resp), "new1") {
		t.Errorf("resp = %s", resp)
	}
}

// ---------------------------------------------------------------------------
// UpdateBlueprint
// ---------------------------------------------------------------------------

func TestUpdateBlueprint_sends_PUT_with_id_in_path(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"updated":true}`))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "tok")
	_, err := c.UpdateBlueprint(context.Background(), "bp123", map[string]any{"name": "new"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "PUT" {
		t.Errorf("method = %q", gotMethod)
	}
	if gotPath != "/api/v1/blueprints/bp123" {
		t.Errorf("path = %q", gotPath)
	}
}

// ---------------------------------------------------------------------------
// DeleteBlueprint
// ---------------------------------------------------------------------------

func TestDeleteBlueprint_sends_DELETE_with_id_in_path(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"deleted":true}`))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "tok")
	_, err := c.DeleteBlueprint(context.Background(), "del1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "DELETE" {
		t.Errorf("method = %q", gotMethod)
	}
	if gotPath != "/api/v1/blueprints/del1" {
		t.Errorf("path = %q", gotPath)
	}
}

// ---------------------------------------------------------------------------
// CreateHierarchy
// ---------------------------------------------------------------------------

func TestCreateHierarchy_sends_POST_with_json_body(t *testing.T) {
	var gotMethod string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":"h1"}`))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "tok")
	body := map[string]any{"name": "hier", "repository_root": "/repo"}
	_, err := c.CreateHierarchy(context.Background(), body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q", gotMethod)
	}
	if !strings.Contains(string(gotBody), `"name":"hier"`) {
		t.Errorf("body = %s", gotBody)
	}
}

// ---------------------------------------------------------------------------
// DeleteHierarchy
// ---------------------------------------------------------------------------

func TestDeleteHierarchy_sends_DELETE_with_id_in_path(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "tok")
	_, err := c.DeleteHierarchy(context.Background(), "hd1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "DELETE" {
		t.Errorf("method = %q", gotMethod)
	}
	if gotPath != "/api/v1/hierarchies/hd1" {
		t.Errorf("path = %q", gotPath)
	}
}

// ---------------------------------------------------------------------------
// Error propagation from unreachable server
// ---------------------------------------------------------------------------

func TestDo_returns_error_when_server_unreachable(t *testing.T) {
	c := New("http://127.0.0.1:1", "tok") // port 1 should be unreachable
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://127.0.0.1:1/x", nil)
	_, err := c.do(req)
	if err == nil {
		t.Fatal("expected connection error")
	}
}

// ---------------------------------------------------------------------------
// NewRequestWithContext error paths (invalid base URL with control char)
// ---------------------------------------------------------------------------

func TestListBlueprints_returns_error_with_invalid_url(t *testing.T) {
	c := New("http://\x00invalid", "tok")
	_, err := c.ListBlueprints(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestGetBlueprint_returns_error_with_invalid_url(t *testing.T) {
	c := New("http://\x00invalid", "tok")
	_, err := c.GetBlueprint(context.Background(), "id")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestListHierarchies_returns_error_with_invalid_url(t *testing.T) {
	c := New("http://\x00invalid", "tok")
	_, err := c.ListHierarchies(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestGetHierarchy_returns_error_with_invalid_url(t *testing.T) {
	c := New("http://\x00invalid", "tok")
	_, err := c.GetHierarchy(context.Background(), "id")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestGetUser_returns_error_with_invalid_url(t *testing.T) {
	c := New("http://\x00invalid", "tok")
	_, err := c.GetUser(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestSearchBlueprints_returns_error_with_invalid_url(t *testing.T) {
	c := New("http://\x00invalid", "tok")
	_, err := c.SearchBlueprints(context.Background(), "q", "", "", "")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestCreateBlueprint_returns_error_with_invalid_url(t *testing.T) {
	c := New("http://\x00invalid", "tok")
	_, err := c.CreateBlueprint(context.Background(), map[string]any{"name": "x"})
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestUpdateBlueprint_returns_error_with_invalid_url(t *testing.T) {
	c := New("http://\x00invalid", "tok")
	_, err := c.UpdateBlueprint(context.Background(), "id", map[string]any{"name": "x"})
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestDeleteBlueprint_returns_error_with_invalid_url(t *testing.T) {
	c := New("http://\x00invalid", "tok")
	_, err := c.DeleteBlueprint(context.Background(), "id")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestCreateHierarchy_returns_error_with_invalid_url(t *testing.T) {
	c := New("http://\x00invalid", "tok")
	_, err := c.CreateHierarchy(context.Background(), map[string]any{"name": "x"})
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestDeleteHierarchy_returns_error_with_invalid_url(t *testing.T) {
	c := New("http://\x00invalid", "tok")
	_, err := c.DeleteHierarchy(context.Background(), "id")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}
