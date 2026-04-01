package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewUpdateBlueprint builds the Tool definition plus its handler.
func NewUpdateBlueprint(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("update_blueprint",
		mcp.WithDescription("Update an existing blueprint in LynxPrompt"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("ID of the blueprint to update"),
		),
		mcp.WithString("name",
			mcp.Description("New name for the blueprint (optional)"),
		),
		mcp.WithString("content",
			mcp.Description("New content for the blueprint (optional)"),
		),
		mcp.WithString("description",
			mcp.Description("New description for the blueprint (optional)"),
		),
		mcp.WithString("visibility",
			mcp.Description("New visibility: PRIVATE, TEAM, or PUBLIC (optional)"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{}

		if name, ok := req.GetArguments()["name"].(string); ok && name != "" {
			body["name"] = name
		}
		if content, ok := req.GetArguments()["content"].(string); ok && content != "" {
			body["content"] = content
		}
		if desc, ok := req.GetArguments()["description"].(string); ok && desc != "" {
			body["description"] = desc
		}
		if vis, ok := req.GetArguments()["visibility"].(string); ok && vis != "" {
			body["visibility"] = vis
		}

		if len(body) == 0 {
			return mcp.NewToolResultError("nothing to update: provide at least one of name, content, description, or visibility"), nil
		}

		resp, err := lp.UpdateBlueprint(id, body)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(
			fmt.Sprintf("Blueprint updated: %s", string(resp)),
		), nil
	}

	return tool, handler
}
