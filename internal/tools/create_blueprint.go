package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewCreateBlueprint builds the Tool definition plus its handler.
func NewCreateBlueprint(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("create_blueprint",
		mcp.WithDescription("Create a new blueprint in LynxPrompt"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the blueprint"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Blueprint content (the actual prompt/instructions text)"),
		),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Blueprint type, e.g. AGENTS_MD, CLAUDE_MD"),
		),
		mcp.WithString("description",
			mcp.Description("Short description of the blueprint (optional)"),
		),
		mcp.WithString("visibility",
			mcp.Description("Visibility: PRIVATE, TEAM, or PUBLIC (optional, defaults to PRIVATE)"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		content, err := req.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		bpType, err := req.RequireString("type")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"name":    name,
			"content": content,
			"type":    bpType,
		}

		if desc, ok := req.GetArguments()["description"].(string); ok && desc != "" {
			body["description"] = desc
		}
		if vis, ok := req.GetArguments()["visibility"].(string); ok && vis != "" {
			body["visibility"] = vis
		}

		resp, err := lp.CreateBlueprint(body)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(
			fmt.Sprintf("Blueprint created: %s", string(resp)),
		), nil
	}

	return tool, handler
}
