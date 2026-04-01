package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewCreateHierarchy builds the Tool definition plus its handler.
func NewCreateHierarchy(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("create_hierarchy",
		mcp.WithDescription("Create a new hierarchy in LynxPrompt"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the hierarchy"),
		),
		mcp.WithString("repository_root",
			mcp.Required(),
			mcp.Description("Root path of the repository this hierarchy maps to"),
		),
		mcp.WithString("description",
			mcp.Description("Short description of the hierarchy (optional)"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		repoRoot, err := req.RequireString("repository_root")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"name":            name,
			"repository_root": repoRoot,
		}

		if desc, ok := req.GetArguments()["description"].(string); ok && desc != "" {
			body["description"] = desc
		}

		resp, err := lp.CreateHierarchy(body)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(
			fmt.Sprintf("Hierarchy created: %s", string(resp)),
		), nil
	}

	return tool, handler
}
