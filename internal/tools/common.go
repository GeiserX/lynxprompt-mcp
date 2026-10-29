// Package tools defines the MCP tools of the LynxPrompt server.
package tools

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Def is one tool plus its handler, as the server registers it.
type Def struct {
	Tool    mcp.Tool
	Handler server.ToolHandlerFunc
}

// ReadOnly reports whether the tool only reads data.
func (d Def) ReadOnly() bool {
	h := d.Tool.Annotations.ReadOnlyHint
	return h != nil && *h
}

// All returns every tool in a fixed order: reads first, then writes.
func All(lp *client.Client) []Def {
	ctors := []func(*client.Client) (mcp.Tool, server.ToolHandlerFunc){
		NewSearchBlueprints,
		NewListBlueprints,
		NewGetBlueprint,
		NewListHierarchies,
		NewGetHierarchy,
		NewCreateBlueprint,
		NewUpdateBlueprint,
		NewDeleteBlueprint,
		NewCreateHierarchy,
		NewDeleteHierarchy,
	}
	defs := make([]Def, 0, len(ctors))
	for _, ctor := range ctors {
		t, h := ctor(lp)
		defs = append(defs, Def{Tool: t, Handler: h})
	}
	return defs
}

// Register adds the tools to s. With readOnly set, tools that change data are
// not registered at all, so neither the client nor the model ever sees them.
func Register(s *server.MCPServer, lp *client.Client, readOnly bool) {
	for _, d := range All(lp) {
		if readOnly && !d.ReadOnly() {
			continue
		}
		s.AddTool(d.Tool, d.Handler)
	}
}

// annotations sets the title and all four behaviour hints explicitly, so no
// client falls back to the spec defaults (destructive, open world).
func annotations(title string, readOnly, destructive, idempotent bool) mcp.ToolOption {
	return func(t *mcp.Tool) {
		t.Title = title
		t.Annotations = mcp.ToolAnnotation{
			Title:           title,
			ReadOnlyHint:    mcp.ToBoolPtr(readOnly),
			DestructiveHint: mcp.ToBoolPtr(destructive),
			IdempotentHint:  mcp.ToBoolPtr(idempotent),
			// Every tool talks to a LynxPrompt instance over the network.
			OpenWorldHint: mcp.ToBoolPtr(true),
		}
	}
}

// jsonResult returns an upstream JSON object as structuredContent, with the
// same JSON as text for clients that ignore structured output.
func jsonResult(body []byte) *mcp.CallToolResult {
	var v map[string]any
	if err := json.Unmarshal(body, &v); err != nil {
		return mcp.NewToolResultError("LynxPrompt API returned a response that is not a JSON object")
	}
	return mcp.NewToolResultStructured(v, string(body))
}

func errorResult(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}

// optString returns a non-empty string argument.
func optString(req mcp.CallToolRequest, key string) (string, bool) {
	s, ok := req.GetArguments()[key].(string)
	s = strings.TrimSpace(s)
	return s, ok && s != ""
}

// optStringList accepts an array of strings or, for older callers, one
// comma-separated string. Empty items are dropped.
func optStringList(req mcp.CallToolRequest, key string) ([]string, error) {
	raw, ok := req.GetArguments()[key]
	if !ok || raw == nil {
		return nil, nil
	}
	var items []string
	switch v := raw.(type) {
	case string:
		items = strings.Split(v, ",")
	case []string:
		items = v
	case []any:
		for _, it := range v {
			s, ok := it.(string)
			if !ok {
				return nil, fmt.Errorf("%s must be a list of strings", key)
			}
			items = append(items, s)
		}
	default:
		return nil, fmt.Errorf("%s must be a list of strings", key)
	}
	out := make([]string, 0, len(items))
	for _, s := range items {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out, nil
}

// optInt returns an integer argument within [lo, hi], or def when absent.
func optInt(req mcp.CallToolRequest, key string, def, lo, hi int) (int, error) {
	raw, ok := req.GetArguments()[key]
	if !ok || raw == nil {
		return def, nil
	}
	var n int
	switch v := raw.(type) {
	case float64:
		if v != float64(int(v)) {
			return 0, fmt.Errorf("%s must be a whole number", key)
		}
		n = int(v)
	case int:
		n = v
	default:
		return 0, fmt.Errorf("%s must be a whole number", key)
	}
	if n < lo || n > hi {
		return 0, fmt.Errorf("%s must be between %d and %d (got %d)", key, lo, hi, n)
	}
	return n, nil
}

// oneOf checks an enum argument the handler received directly.
func oneOf(key, val string, allowed []string) error {
	for _, a := range allowed {
		if val == a {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of %s (got %q)", key, strings.Join(allowed, ", "), val)
}

// encodeContent sends blueprint text as content_base64. Cloudflare's WAF in
// front of the hosted instance rejects raw request bodies that contain shell or
// config snippets, which is what blueprints are made of; the API decodes
// content_base64 back to text, and the official CLI does the same.
func encodeContent(body map[string]any, content string) {
	body["content_base64"] = base64.StdEncoding.EncodeToString([]byte(content))
}

var (
	visibilities = []string{"PRIVATE", "TEAM", "PUBLIC"}

	// blueprintTypes are the file formats LynxPrompt knows, from its API.
	blueprintTypes = []string{
		"AGENTS_MD", "CLAUDE_MD", "COPILOT_INSTRUCTIONS", "WINDSURF_RULES",
		"CLINE_RULES", "CODEX_MD", "CURSOR_RULES", "CURSOR_MDC", "GEMINI_MD",
		"ZED_INSTRUCTIONS", "VOID_CONFIG", "TRAE_RULES", "IDX_CONFIG",
		"ROO_RULES", "CONTINUE_CONFIG", "CODY_CONFIG", "TABNINE_CONFIG",
		"AMAZONQ_RULES", "AUGMENT_RULES", "KILOCODE_RULES", "JUNIE_GUIDELINES",
		"KIRO_STEERING", "AIDER_MD", "GOOSEHINTS",
		"CURSOR_COMMAND", "CLAUDE_COMMAND", "WINDSURF_WORKFLOW",
		"COPILOT_PROMPT", "CONTINUE_PROMPT", "OPENCODE_COMMAND", "CUSTOM",
	}
)

// idProperty is the schema for a bp_/ha_ id argument.
func idProperty(name, desc string) mcp.ToolOption {
	return mcp.WithString(name,
		mcp.Required(),
		mcp.Description(desc),
		mcp.Pattern(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`),
		mcp.MaxLength(128),
	)
}

// requireID reads and validates an id argument.
func requireID(req mcp.CallToolRequest, key string) (string, error) {
	id, err := req.RequireString(key)
	if err != nil {
		return "", err
	}
	if err := client.ValidateID(id); err != nil {
		return "", err
	}
	return id, nil
}
