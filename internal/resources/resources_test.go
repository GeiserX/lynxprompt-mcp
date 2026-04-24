package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func fakeAPI(t *testing.T, routes map[string]string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body, ok := routes[r.URL.Path]; ok {
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

func fakeErrorAPI(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func readResourceMsg(id int, uri string) []byte {
	return []byte(fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": %d,
		"method": "resources/read",
		"params": {"uri": %q}
	}`, id, uri))
}

func extractText(t *testing.T, resp mcp.JSONRPCMessage) string {
	t.Helper()
	jsonResp, ok := resp.(mcp.JSONRPCResponse)
	if !ok {
		t.Fatalf("expected JSONRPCResponse, got %T: %+v", resp, resp)
	}
	b, err := json.Marshal(jsonResp.Result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	var result struct {
		Contents []struct {
			URI      string `json:"uri"`
			MIMEType string `json:"mimeType"`
			Text     string `json:"text"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(result.Contents) == 0 {
		t.Fatal("no contents in response")
	}
	return result.Contents[0].Text
}

func assertIsError(t *testing.T, resp mcp.JSONRPCMessage) {
	t.Helper()
	if _, ok := resp.(mcp.JSONRPCError); !ok {
		t.Fatalf("expected JSONRPCError, got %T: %+v", resp, resp)
	}
}

func newServer() *server.MCPServer {
	return server.NewMCPServer("test", "0.0.0", server.WithResourceCapabilities(true, true))
}

// ---------------------------------------------------------------------------
// RegisterBlueprints (static resource)
// ---------------------------------------------------------------------------

func TestRegisterBlueprints_returns_list_on_success(t *testing.T) {
	api := fakeAPI(t, map[string]string{
		"/api/v1/blueprints": `[{"id":"1","name":"test-bp"}]`,
	})
	lp := client.New(api.URL, "tok")
	s := newServer()
	RegisterBlueprints(s, lp)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "lynxprompt://blueprints"))
	text := extractText(t, resp)
	if !strings.Contains(text, "test-bp") {
		t.Errorf("expected test-bp in response, got %q", text)
	}
}

func TestRegisterBlueprints_returns_error_when_api_fails(t *testing.T) {
	api := fakeErrorAPI(t, 500, "server error")
	lp := client.New(api.URL, "tok")
	s := newServer()
	RegisterBlueprints(s, lp)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "lynxprompt://blueprints"))
	assertIsError(t, resp)
}

// ---------------------------------------------------------------------------
// RegisterBlueprint (template resource)
// ---------------------------------------------------------------------------

func TestRegisterBlueprint_returns_single_blueprint(t *testing.T) {
	api := fakeAPI(t, map[string]string{
		"/api/v1/blueprints/bp42": `{"id":"bp42","content":"hello world"}`,
	})
	lp := client.New(api.URL, "tok")
	s := newServer()
	RegisterBlueprint(s, lp)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "lynxprompt://blueprint/bp42"))
	text := extractText(t, resp)
	if !strings.Contains(text, "hello world") {
		t.Errorf("expected content, got %q", text)
	}
}

func TestRegisterBlueprint_returns_error_when_api_fails(t *testing.T) {
	api := fakeErrorAPI(t, 404, "not found")
	lp := client.New(api.URL, "tok")
	s := newServer()
	RegisterBlueprint(s, lp)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "lynxprompt://blueprint/missing"))
	assertIsError(t, resp)
}

// ---------------------------------------------------------------------------
// RegisterHierarchies (static resource)
// ---------------------------------------------------------------------------

func TestRegisterHierarchies_returns_list_on_success(t *testing.T) {
	api := fakeAPI(t, map[string]string{
		"/api/v1/hierarchies": `[{"id":"h1","name":"hier1"}]`,
	})
	lp := client.New(api.URL, "tok")
	s := newServer()
	RegisterHierarchies(s, lp)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "lynxprompt://hierarchies"))
	text := extractText(t, resp)
	if !strings.Contains(text, "hier1") {
		t.Errorf("expected hier1, got %q", text)
	}
}

func TestRegisterHierarchies_returns_error_when_api_fails(t *testing.T) {
	api := fakeErrorAPI(t, 500, "err")
	lp := client.New(api.URL, "tok")
	s := newServer()
	RegisterHierarchies(s, lp)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "lynxprompt://hierarchies"))
	assertIsError(t, resp)
}

// ---------------------------------------------------------------------------
// RegisterHierarchy (template resource)
// ---------------------------------------------------------------------------

func TestRegisterHierarchy_returns_single_hierarchy(t *testing.T) {
	api := fakeAPI(t, map[string]string{
		"/api/v1/hierarchies/h99": `{"id":"h99","tree":{"root":true}}`,
	})
	lp := client.New(api.URL, "tok")
	s := newServer()
	RegisterHierarchy(s, lp)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "lynxprompt://hierarchy/h99"))
	text := extractText(t, resp)
	if !strings.Contains(text, "h99") {
		t.Errorf("expected h99, got %q", text)
	}
}

func TestRegisterHierarchy_returns_error_when_api_fails(t *testing.T) {
	api := fakeErrorAPI(t, 403, "forbidden")
	lp := client.New(api.URL, "tok")
	s := newServer()
	RegisterHierarchy(s, lp)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "lynxprompt://hierarchy/nope"))
	assertIsError(t, resp)
}

// ---------------------------------------------------------------------------
// RegisterUser (static resource)
// ---------------------------------------------------------------------------

func TestRegisterUser_returns_user_info(t *testing.T) {
	api := fakeAPI(t, map[string]string{
		"/api/v1/user": `{"email":"test@example.com","name":"Test User"}`,
	})
	lp := client.New(api.URL, "tok")
	s := newServer()
	RegisterUser(s, lp)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "lynxprompt://user"))
	text := extractText(t, resp)
	if !strings.Contains(text, "test@example.com") {
		t.Errorf("expected email, got %q", text)
	}
}

func TestRegisterUser_returns_error_when_api_fails(t *testing.T) {
	api := fakeErrorAPI(t, 401, "unauthorized")
	lp := client.New(api.URL, "tok")
	s := newServer()
	RegisterUser(s, lp)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "lynxprompt://user"))
	assertIsError(t, resp)
}

// ---------------------------------------------------------------------------
// Empty ID paths for template resources
// ---------------------------------------------------------------------------

func TestRegisterBlueprint_returns_error_for_empty_id(t *testing.T) {
	api := fakeAPI(t, map[string]string{})
	lp := client.New(api.URL, "tok")
	s := newServer()
	RegisterBlueprint(s, lp)

	// URI with no ID after prefix
	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "lynxprompt://blueprint/"))
	assertIsError(t, resp)
}

func TestRegisterHierarchy_returns_error_for_empty_id(t *testing.T) {
	api := fakeAPI(t, map[string]string{})
	lp := client.New(api.URL, "tok")
	s := newServer()
	RegisterHierarchy(s, lp)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "lynxprompt://hierarchy/"))
	assertIsError(t, resp)
}
