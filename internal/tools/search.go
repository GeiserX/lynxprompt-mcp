package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewSearchBlueprints builds the Tool definition plus its handler.
func NewSearchBlueprints(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("search_blueprints",
		mcp.WithDescription("Search the public LynxPrompt blueprint marketplace"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query string"),
		),
		mcp.WithString("category",
			mcp.Description("Filter by category (optional)"),
		),
		mcp.WithString("type",
			mcp.Description("Filter by type, e.g. AGENTS_MD, CLAUDE_MD (optional)"),
		),
		mcp.WithString("tags",
			mcp.Description("Comma-separated tags to filter by (optional)"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		category, _ := req.GetArguments()["category"].(string)
		bpType, _ := req.GetArguments()["type"].(string)
		tags, _ := req.GetArguments()["tags"].(string)

		resp, err := lp.SearchBlueprints(query, category, bpType, tags)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(
			fmt.Sprintf("Search results: %s", string(resp)),
		), nil
	}

	return tool, handler
}
