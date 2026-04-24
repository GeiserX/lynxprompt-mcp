package tools

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func fakeAPI(t *testing.T, routes map[string]string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		if body, ok := routes[key]; ok {
			w.WriteHeader(200)
			w.Write([]byte(body))
			return
		}
		// Fallback: match path only
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

// recordingAPI captures the last request method, path, and body.
type recordedReq struct {
	Method string
	Path   string
	Body   string
	Query  string
}

func recordingAPI(t *testing.T, status int, respBody string) (*httptest.Server, *recordedReq) {
	t.Helper()
	rec := &recordedReq{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		rec.Method = r.Method
		rec.Path = r.URL.Path
		rec.Body = string(b)
		rec.Query = r.URL.RawQuery
		w.WriteHeader(status)
		w.Write([]byte(respBody))
	}))
	t.Cleanup(srv.Close)
	return srv, rec
}

func makeCallToolRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

// ---------------------------------------------------------------------------
// SearchBlueprints
// ---------------------------------------------------------------------------

func TestSearchBlueprints_returns_results_on_success(t *testing.T) {
	srv, rec := recordingAPI(t, 200, `[{"id":"1","name":"found"}]`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewSearchBlueprints(lp)

	req := makeCallToolRequest(map[string]any{
		"query":    "test",
		"category": "dev",
		"type":     "AGENTS_MD",
		"tags":     "go,mcp",
	})

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Method != "GET" {
		t.Errorf("method = %q", rec.Method)
	}
	if !strings.Contains(rec.Query, "q=test") {
		t.Errorf("query missing q param: %q", rec.Query)
	}
	if !strings.Contains(rec.Query, "category=dev") {
		t.Errorf("query missing category: %q", rec.Query)
	}
	// Check result text
	if len(result.Content) == 0 {
		t.Fatal("empty result content")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("result is not TextContent")
	}
	if !strings.Contains(tc.Text, "found") {
		t.Errorf("result text = %q", tc.Text)
	}
}

func TestSearchBlueprints_returns_error_when_query_missing(t *testing.T) {
	srv, _ := recordingAPI(t, 200, `[]`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewSearchBlueprints(lp)

	req := makeCallToolRequest(map[string]any{})

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true when query is missing")
	}
}

func TestSearchBlueprints_returns_error_when_api_fails(t *testing.T) {
	srv := fakeErrorAPI(t, 500, "boom")
	lp := client.New(srv.URL, "tok")
	_, handler := NewSearchBlueprints(lp)

	req := makeCallToolRequest(map[string]any{"query": "test"})
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError when API returns 500")
	}
}

func TestSearchBlueprints_omits_empty_optional_params(t *testing.T) {
	srv, rec := recordingAPI(t, 200, `[]`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewSearchBlueprints(lp)

	req := makeCallToolRequest(map[string]any{"query": "test"})
	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(rec.Query, "category") {
		t.Errorf("should not have category in query: %q", rec.Query)
	}
}

// ---------------------------------------------------------------------------
// CreateBlueprint
// ---------------------------------------------------------------------------

func TestCreateBlueprint_sends_correct_request(t *testing.T) {
	srv, rec := recordingAPI(t, 201, `{"id":"new-bp"}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewCreateBlueprint(lp)

	req := makeCallToolRequest(map[string]any{
		"name":        "My Blueprint",
		"content":     "blueprint content here",
		"type":        "CLAUDE_MD",
		"description": "A description",
		"visibility":  "PUBLIC",
	})

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Method != "POST" {
		t.Errorf("method = %q", rec.Method)
	}
	if rec.Path != "/api/v1/blueprints" {
		t.Errorf("path = %q", rec.Path)
	}
	if !strings.Contains(rec.Body, `"name":"My Blueprint"`) {
		t.Errorf("body missing name: %s", rec.Body)
	}
	if !strings.Contains(rec.Body, `"description":"A description"`) {
		t.Errorf("body missing description: %s", rec.Body)
	}
	if !strings.Contains(rec.Body, `"visibility":"PUBLIC"`) {
		t.Errorf("body missing visibility: %s", rec.Body)
	}
	tc := result.Content[0].(mcp.TextContent)
	if !strings.Contains(tc.Text, "new-bp") {
		t.Errorf("result = %q", tc.Text)
	}
}

func TestCreateBlueprint_returns_error_when_name_missing(t *testing.T) {
	srv, _ := recordingAPI(t, 201, `{}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewCreateBlueprint(lp)

	req := makeCallToolRequest(map[string]any{
		"content": "x",
		"type":    "AGENTS_MD",
	})
	result, _ := handler(context.Background(), req)
	if !result.IsError {
		t.Error("expected error when name missing")
	}
}

func TestCreateBlueprint_returns_error_when_content_missing(t *testing.T) {
	srv, _ := recordingAPI(t, 201, `{}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewCreateBlueprint(lp)

	req := makeCallToolRequest(map[string]any{
		"name": "test",
		"type": "AGENTS_MD",
	})
	result, _ := handler(context.Background(), req)
	if !result.IsError {
		t.Error("expected error when content missing")
	}
}

func TestCreateBlueprint_returns_error_when_type_missing(t *testing.T) {
	srv, _ := recordingAPI(t, 201, `{}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewCreateBlueprint(lp)

	req := makeCallToolRequest(map[string]any{
		"name":    "test",
		"content": "x",
	})
	result, _ := handler(context.Background(), req)
	if !result.IsError {
		t.Error("expected error when type missing")
	}
}

func TestCreateBlueprint_returns_error_when_api_fails(t *testing.T) {
	srv := fakeErrorAPI(t, 500, "fail")
	lp := client.New(srv.URL, "tok")
	_, handler := NewCreateBlueprint(lp)

	req := makeCallToolRequest(map[string]any{
		"name":    "test",
		"content": "x",
		"type":    "AGENTS_MD",
	})
	result, _ := handler(context.Background(), req)
	if !result.IsError {
		t.Error("expected error when API fails")
	}
}

func TestCreateBlueprint_without_optional_fields(t *testing.T) {
	srv, rec := recordingAPI(t, 201, `{"id":"bp1"}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewCreateBlueprint(lp)

	req := makeCallToolRequest(map[string]any{
		"name":    "test",
		"content": "x",
		"type":    "AGENTS_MD",
	})
	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(rec.Body, "description") {
		t.Errorf("body should not contain description: %s", rec.Body)
	}
	if strings.Contains(rec.Body, "visibility") {
		t.Errorf("body should not contain visibility: %s", rec.Body)
	}
}

// ---------------------------------------------------------------------------
// UpdateBlueprint
// ---------------------------------------------------------------------------

func TestUpdateBlueprint_sends_correct_request(t *testing.T) {
	srv, rec := recordingAPI(t, 200, `{"updated":true}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewUpdateBlueprint(lp)

	req := makeCallToolRequest(map[string]any{
		"id":          "bp42",
		"name":        "Updated Name",
		"content":     "new content",
		"description": "new desc",
		"visibility":  "TEAM",
	})

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Method != "PUT" {
		t.Errorf("method = %q", rec.Method)
	}
	if rec.Path != "/api/v1/blueprints/bp42" {
		t.Errorf("path = %q", rec.Path)
	}
	if !strings.Contains(rec.Body, `"name":"Updated Name"`) {
		t.Errorf("body = %s", rec.Body)
	}
	tc := result.Content[0].(mcp.TextContent)
	if !strings.Contains(tc.Text, "updated") {
		t.Errorf("result = %q", tc.Text)
	}
}

func TestUpdateBlueprint_returns_error_when_id_missing(t *testing.T) {
	srv, _ := recordingAPI(t, 200, `{}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewUpdateBlueprint(lp)

	req := makeCallToolRequest(map[string]any{
		"name": "x",
	})
	result, _ := handler(context.Background(), req)
	if !result.IsError {
		t.Error("expected error when id missing")
	}
}

func TestUpdateBlueprint_returns_error_when_no_fields_to_update(t *testing.T) {
	srv, _ := recordingAPI(t, 200, `{}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewUpdateBlueprint(lp)

	req := makeCallToolRequest(map[string]any{
		"id": "bp42",
	})
	result, _ := handler(context.Background(), req)
	if !result.IsError {
		t.Error("expected error when nothing to update")
	}
	tc := result.Content[0].(mcp.TextContent)
	if !strings.Contains(tc.Text, "nothing to update") {
		t.Errorf("error message = %q", tc.Text)
	}
}

func TestUpdateBlueprint_returns_error_when_api_fails(t *testing.T) {
	srv := fakeErrorAPI(t, 500, "fail")
	lp := client.New(srv.URL, "tok")
	_, handler := NewUpdateBlueprint(lp)

	req := makeCallToolRequest(map[string]any{
		"id":   "bp42",
		"name": "new",
	})
	result, _ := handler(context.Background(), req)
	if !result.IsError {
		t.Error("expected error when API fails")
	}
}

func TestUpdateBlueprint_sends_only_provided_fields(t *testing.T) {
	srv, rec := recordingAPI(t, 200, `{"ok":true}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewUpdateBlueprint(lp)

	req := makeCallToolRequest(map[string]any{
		"id":   "bp42",
		"name": "only-name",
	})
	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(rec.Body, `"name":"only-name"`) {
		t.Errorf("body should have name: %s", rec.Body)
	}
	if strings.Contains(rec.Body, "content") {
		t.Errorf("body should not have content: %s", rec.Body)
	}
	if strings.Contains(rec.Body, "description") {
		t.Errorf("body should not have description: %s", rec.Body)
	}
}

// ---------------------------------------------------------------------------
// DeleteBlueprint
// ---------------------------------------------------------------------------

func TestDeleteBlueprint_sends_correct_request(t *testing.T) {
	srv, rec := recordingAPI(t, 200, `{"deleted":true}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewDeleteBlueprint(lp)

	req := makeCallToolRequest(map[string]any{"id": "bp-del"})

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Method != "DELETE" {
		t.Errorf("method = %q", rec.Method)
	}
	if rec.Path != "/api/v1/blueprints/bp-del" {
		t.Errorf("path = %q", rec.Path)
	}
	tc := result.Content[0].(mcp.TextContent)
	if !strings.Contains(tc.Text, "deleted") {
		t.Errorf("result = %q", tc.Text)
	}
}

func TestDeleteBlueprint_returns_error_when_id_missing(t *testing.T) {
	srv, _ := recordingAPI(t, 200, `{}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewDeleteBlueprint(lp)

	req := makeCallToolRequest(map[string]any{})
	result, _ := handler(context.Background(), req)
	if !result.IsError {
		t.Error("expected error when id missing")
	}
}

func TestDeleteBlueprint_returns_error_when_api_fails(t *testing.T) {
	srv := fakeErrorAPI(t, 500, "fail")
	lp := client.New(srv.URL, "tok")
	_, handler := NewDeleteBlueprint(lp)

	req := makeCallToolRequest(map[string]any{"id": "x"})
	result, _ := handler(context.Background(), req)
	if !result.IsError {
		t.Error("expected error when API fails")
	}
}

// ---------------------------------------------------------------------------
// CreateHierarchy
// ---------------------------------------------------------------------------

func TestCreateHierarchy_sends_correct_request(t *testing.T) {
	srv, rec := recordingAPI(t, 201, `{"id":"h-new"}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewCreateHierarchy(lp)

	req := makeCallToolRequest(map[string]any{
		"name":            "My Hierarchy",
		"repository_root": "/home/repo",
		"description":     "test desc",
	})

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Method != "POST" {
		t.Errorf("method = %q", rec.Method)
	}
	if rec.Path != "/api/v1/hierarchies" {
		t.Errorf("path = %q", rec.Path)
	}
	if !strings.Contains(rec.Body, `"name":"My Hierarchy"`) {
		t.Errorf("body missing name: %s", rec.Body)
	}
	if !strings.Contains(rec.Body, `"repository_root":"/home/repo"`) {
		t.Errorf("body missing repo root: %s", rec.Body)
	}
	if !strings.Contains(rec.Body, `"description":"test desc"`) {
		t.Errorf("body missing description: %s", rec.Body)
	}
	tc := result.Content[0].(mcp.TextContent)
	if !strings.Contains(tc.Text, "h-new") {
		t.Errorf("result = %q", tc.Text)
	}
}

func TestCreateHierarchy_returns_error_when_name_missing(t *testing.T) {
	srv, _ := recordingAPI(t, 201, `{}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewCreateHierarchy(lp)

	req := makeCallToolRequest(map[string]any{
		"repository_root": "/repo",
	})
	result, _ := handler(context.Background(), req)
	if !result.IsError {
		t.Error("expected error when name missing")
	}
}

func TestCreateHierarchy_returns_error_when_repository_root_missing(t *testing.T) {
	srv, _ := recordingAPI(t, 201, `{}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewCreateHierarchy(lp)

	req := makeCallToolRequest(map[string]any{
		"name": "test",
	})
	result, _ := handler(context.Background(), req)
	if !result.IsError {
		t.Error("expected error when repository_root missing")
	}
}

func TestCreateHierarchy_returns_error_when_api_fails(t *testing.T) {
	srv := fakeErrorAPI(t, 500, "fail")
	lp := client.New(srv.URL, "tok")
	_, handler := NewCreateHierarchy(lp)

	req := makeCallToolRequest(map[string]any{
		"name":            "test",
		"repository_root": "/repo",
	})
	result, _ := handler(context.Background(), req)
	if !result.IsError {
		t.Error("expected error when API fails")
	}
}

func TestCreateHierarchy_without_optional_description(t *testing.T) {
	srv, rec := recordingAPI(t, 201, `{"id":"h1"}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewCreateHierarchy(lp)

	req := makeCallToolRequest(map[string]any{
		"name":            "test",
		"repository_root": "/repo",
	})
	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(rec.Body, "description") {
		t.Errorf("body should not contain description: %s", rec.Body)
	}
}

// ---------------------------------------------------------------------------
// DeleteHierarchy
// ---------------------------------------------------------------------------

func TestDeleteHierarchy_sends_correct_request(t *testing.T) {
	srv, rec := recordingAPI(t, 200, `{"deleted":true}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewDeleteHierarchy(lp)

	req := makeCallToolRequest(map[string]any{"id": "h-del"})

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Method != "DELETE" {
		t.Errorf("method = %q", rec.Method)
	}
	if rec.Path != "/api/v1/hierarchies/h-del" {
		t.Errorf("path = %q", rec.Path)
	}
	tc := result.Content[0].(mcp.TextContent)
	if !strings.Contains(tc.Text, "deleted") {
		t.Errorf("result = %q", tc.Text)
	}
}

func TestDeleteHierarchy_returns_error_when_id_missing(t *testing.T) {
	srv, _ := recordingAPI(t, 200, `{}`)
	lp := client.New(srv.URL, "tok")
	_, handler := NewDeleteHierarchy(lp)

	req := makeCallToolRequest(map[string]any{})
	result, _ := handler(context.Background(), req)
	if !result.IsError {
		t.Error("expected error when id missing")
	}
}

func TestDeleteHierarchy_returns_error_when_api_fails(t *testing.T) {
	srv := fakeErrorAPI(t, 500, "fail")
	lp := client.New(srv.URL, "tok")
	_, handler := NewDeleteHierarchy(lp)

	req := makeCallToolRequest(map[string]any{"id": "x"})
	result, _ := handler(context.Background(), req)
	if !result.IsError {
		t.Error("expected error when API fails")
	}
}

// ---------------------------------------------------------------------------
// Tool definitions: verify names are correct
// ---------------------------------------------------------------------------

func TestToolDefinitions_have_correct_names(t *testing.T) {
	lp := client.New("http://localhost", "")
	tests := []struct {
		name     string
		toolFunc func(*client.Client) (mcp.Tool, server.ToolHandlerFunc)
		wantName string
	}{
		{"search", NewSearchBlueprints, "search_blueprints"},
		{"create_bp", NewCreateBlueprint, "create_blueprint"},
		{"update_bp", NewUpdateBlueprint, "update_blueprint"},
		{"delete_bp", NewDeleteBlueprint, "delete_blueprint"},
		{"create_h", NewCreateHierarchy, "create_hierarchy"},
		{"delete_h", NewDeleteHierarchy, "delete_hierarchy"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tool, _ := tc.toolFunc(lp)
			if tool.Name != tc.wantName {
				t.Errorf("tool name = %q, want %q", tool.Name, tc.wantName)
			}
		})
	}
}
