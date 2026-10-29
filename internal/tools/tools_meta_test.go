package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// wantHints is the contract for every tool: readOnly, destructive, idempotent.
var wantHints = map[string][3]bool{
	"search_blueprints": {true, false, true},
	"list_blueprints":   {true, false, true},
	"get_blueprint":     {true, false, true},
	"list_hierarchies":  {true, false, true},
	"get_hierarchy":     {true, false, true},
	"create_blueprint":  {false, false, false},
	"update_blueprint":  {false, true, true},
	"delete_blueprint":  {false, true, true},
	"create_hierarchy":  {false, false, true},
	"delete_hierarchy":  {false, true, true},
}

func TestEveryTool_has_title_all_four_hints_and_an_output_schema(t *testing.T) {
	defs := All(client.New("http://localhost", ""))
	if len(defs) != len(wantHints) {
		t.Fatalf("got %d tools, want %d", len(defs), len(wantHints))
	}
	for _, d := range defs {
		tool := d.Tool
		want, ok := wantHints[tool.Name]
		if !ok {
			t.Errorf("unexpected tool %q", tool.Name)
			continue
		}
		a := tool.Annotations
		if tool.Title == "" || a.Title == "" {
			t.Errorf("%s: missing title", tool.Name)
		}
		if a.ReadOnlyHint == nil || a.DestructiveHint == nil || a.IdempotentHint == nil || a.OpenWorldHint == nil {
			t.Errorf("%s: every hint must be set explicitly: %+v", tool.Name, a)
			continue
		}
		got := [3]bool{*a.ReadOnlyHint, *a.DestructiveHint, *a.IdempotentHint}
		if got != want {
			t.Errorf("%s: hints (readOnly, destructive, idempotent) = %v, want %v", tool.Name, got, want)
		}
		if !*a.OpenWorldHint {
			t.Errorf("%s: openWorldHint must be true (it calls a remote API)", tool.Name)
		}
		if len(tool.RawOutputSchema) == 0 || !json.Valid(tool.RawOutputSchema) {
			t.Errorf("%s: missing or invalid output schema", tool.Name)
		}
		if len(tool.Description) < 60 {
			t.Errorf("%s: description too thin: %q", tool.Name, tool.Description)
		}
	}
}

func TestRegister_read_only_hides_every_mutating_tool(t *testing.T) {
	lp := client.New("http://localhost", "")
	for _, readOnly := range []bool{false, true} {
		s := server.NewMCPServer("t", "0")
		Register(s, lp, readOnly)
		got := s.ListTools()
		for name, h := range wantHints {
			_, present := got[name]
			wantPresent := !readOnly || h[0]
			if present != wantPresent {
				t.Errorf("readOnly=%v: %s present=%v, want %v", readOnly, name, present, wantPresent)
			}
		}
	}
}

func TestListBlueprints_sends_paging_and_returns_structured_content(t *testing.T) {
	srv, rec := recordingAPI(t, 200, `{"blueprints":[{"id":"bp_1","name":"one"}],"total":1,"limit":10,"offset":5,"has_more":false}`)
	_, handler := NewListBlueprints(client.New(srv.URL, "tok"))

	result, _ := handler(context.Background(), makeCallToolRequest(map[string]any{
		"limit": float64(10), "offset": float64(5), "visibility": "PRIVATE",
	}))
	if result.IsError {
		t.Fatalf("unexpected error: %+v", result.Content)
	}
	if rec.Path != "/api/v1/blueprints" || rec.Query != "limit=10&offset=5&visibility=PRIVATE" {
		t.Errorf("request = %s ?%s", rec.Path, rec.Query)
	}
	sc, ok := result.StructuredContent.(map[string]any)
	if !ok || sc["total"] != float64(1) {
		t.Errorf("structuredContent = %#v", result.StructuredContent)
	}
	if tc := result.Content[0].(mcp.TextContent); !strings.HasPrefix(tc.Text, "{") {
		t.Errorf("text fallback must be the JSON itself, got %q", tc.Text)
	}
}

func TestListBlueprints_rejects_out_of_range_limit(t *testing.T) {
	srv, rec := recordingAPI(t, 200, `{}`)
	_, handler := NewListBlueprints(client.New(srv.URL, "tok"))
	for _, args := range []map[string]any{
		{"limit": float64(0)}, {"limit": float64(101)}, {"limit": 2.5}, {"offset": float64(-1)}, {"visibility": "SECRET"},
	} {
		if result, _ := handler(context.Background(), makeCallToolRequest(args)); !result.IsError {
			t.Errorf("%v: expected an error", args)
		}
	}
	if rec.Method != "" {
		t.Error("the API was called with invalid arguments")
	}
}

func TestReadTools_hit_the_right_endpoints(t *testing.T) {
	tests := []struct {
		ctor func(*client.Client) (mcp.Tool, server.ToolHandlerFunc)
		args map[string]any
		path string
	}{
		{NewGetBlueprint, map[string]any{"id": "bp_abc"}, "/api/v1/blueprints/bp_abc"},
		{NewListHierarchies, map[string]any{"limit": float64(3)}, "/api/v1/hierarchies"},
		{NewGetHierarchy, map[string]any{"id": "ha_abc"}, "/api/v1/hierarchies/ha_abc"},
	}
	for _, tc := range tests {
		srv, rec := recordingAPI(t, 200, `{"ok":true}`)
		tool, handler := tc.ctor(client.New(srv.URL, "tok"))
		result, _ := handler(context.Background(), makeCallToolRequest(tc.args))
		if result.IsError {
			t.Errorf("%s: %+v", tool.Name, result.Content)
		}
		if rec.Method != "GET" || rec.Path != tc.path {
			t.Errorf("%s: %s %s, want GET %s", tool.Name, rec.Method, rec.Path, tc.path)
		}
	}
}

func TestIDTools_reject_path_tricks_before_calling_the_api(t *testing.T) {
	for _, ctor := range []func(*client.Client) (mcp.Tool, server.ToolHandlerFunc){
		NewGetBlueprint, NewUpdateBlueprint, NewDeleteBlueprint, NewGetHierarchy, NewDeleteHierarchy,
	} {
		srv, rec := recordingAPI(t, 200, `{}`)
		tool, handler := ctor(client.New(srv.URL, "tok"))
		for _, id := range []string{"..", ".", "a/b", "x?y=1"} {
			result, _ := handler(context.Background(), makeCallToolRequest(map[string]any{"id": id, "name": "valid name"}))
			if !result.IsError {
				t.Errorf("%s accepted id %q", tool.Name, id)
			}
		}
		if rec.Method != "" {
			t.Errorf("%s called the API with a bad id", tool.Name)
		}
	}
}

func TestUpdateBlueprint_sends_expected_checksum_and_base64_content(t *testing.T) {
	srv, rec := recordingAPI(t, 200, `{"success":true,"blueprint":{"id":"bp_1"}}`)
	_, handler := NewUpdateBlueprint(client.New(srv.URL, "tok"))

	content := "ssh deploy@host 'cat /etc/passwd'"
	result, _ := handler(context.Background(), makeCallToolRequest(map[string]any{
		"id": "bp_1", "content": content, "expected_checksum": "abc123", "tags": []any{"ops", "ssh"}, "type": "claude_md",
	}))
	if result.IsError {
		t.Fatalf("unexpected error: %+v", result.Content)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(rec.Body), &body); err != nil {
		t.Fatal(err)
	}
	if body["expected_checksum"] != "abc123" {
		t.Errorf("expected_checksum not sent: %s", rec.Body)
	}
	if body["content_base64"] != base64.StdEncoding.EncodeToString([]byte(content)) || body["content"] != nil {
		t.Errorf("content must travel only as content_base64: %s", rec.Body)
	}
	if body["type"] != "CLAUDE_MD" {
		t.Errorf("type should be upper-cased: %s", rec.Body)
	}
	if tags, _ := body["tags"].([]any); len(tags) != 2 {
		t.Errorf("tags = %v", body["tags"])
	}
}

func TestUpdateBlueprint_turns_a_409_into_an_actionable_error(t *testing.T) {
	srv, _ := recordingAPI(t, 409, `{"error":"Conflict - blueprint has been modified","current_checksum":"def456"}`)
	_, handler := NewUpdateBlueprint(client.New(srv.URL, "tok"))

	result, _ := handler(context.Background(), makeCallToolRequest(map[string]any{
		"id": "bp_1", "content": "a long enough body", "expected_checksum": "abc123",
	}))
	if !result.IsError {
		t.Fatal("expected a tool error")
	}
	text := result.Content[0].(mcp.TextContent).Text
	for _, want := range []string{"409", "def456", "get_blueprint"} {
		if !strings.Contains(text, want) {
			t.Errorf("error %q lacks %q", text, want)
		}
	}
}

func TestCreateBlueprint_rejects_short_content_and_bad_enums(t *testing.T) {
	srv, rec := recordingAPI(t, 201, `{}`)
	_, handler := NewCreateBlueprint(client.New(srv.URL, "tok"))
	for _, args := range []map[string]any{
		{"name": "valid", "content": "short"},
		{"name": "ab", "content": "long enough content"},
		{"name": "valid", "content": "long enough content", "visibility": "WORLD"},
		{"name": "valid", "content": "long enough content", "type": "README"},
		{"name": "valid", "content": "long enough content", "hierarchy_id": "../x"},
		{"name": "valid", "content": "long enough content", "tags": []any{strings.Repeat("t", 31)}},
	} {
		if result, _ := handler(context.Background(), makeCallToolRequest(args)); !result.IsError {
			t.Errorf("%v: expected an error", args)
		}
	}
	if rec.Method != "" {
		t.Error("the API was called with invalid arguments")
	}
}

func TestCreateBlueprint_sends_hierarchy_fields(t *testing.T) {
	srv, rec := recordingAPI(t, 201, `{"success":true,"blueprint":{"id":"bp_2"}}`)
	_, handler := NewCreateBlueprint(client.New(srv.URL, "tok"))
	result, _ := handler(context.Background(), makeCallToolRequest(map[string]any{
		"name": "api agents", "content": "long enough content", "hierarchy_id": "ha_1",
		"parent_id": "bp_root", "repository_path": "packages/api/AGENTS.md", "category": "web",
	}))
	if result.IsError {
		t.Fatalf("unexpected error: %+v", result.Content)
	}
	for _, want := range []string{`"hierarchy_id":"ha_1"`, `"parent_id":"bp_root"`, `"repository_path":"packages/api/AGENTS.md"`, `"category":"web"`} {
		if !strings.Contains(rec.Body, want) {
			t.Errorf("body %s lacks %s", rec.Body, want)
		}
	}
}

// get_blueprint reads any blueprint the token can see, which is the only way
// to fetch a marketplace result's content; its description must not say
// otherwise, or a model never tries.
func TestGetBlueprint_description_covers_marketplace_and_team_blueprints(t *testing.T) {
	tool, _ := NewGetBlueprint(client.New("http://127.0.0.1:1", ""))
	d := tool.Description
	if strings.Contains(d, "Only works for blueprints owned") {
		t.Errorf("description still limits get_blueprint to owned blueprints: %q", d)
	}
	for _, want := range []string{"search_blueprints", "team", "Only the owner can update or delete"} {
		if !strings.Contains(d, want) {
			t.Errorf("description lacks %q: %q", want, d)
		}
	}
}
