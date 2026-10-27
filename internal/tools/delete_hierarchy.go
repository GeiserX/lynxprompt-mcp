package tools

import (
	"context"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewDeleteHierarchy builds the Tool definition plus its handler.
func NewDeleteHierarchy(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("delete_hierarchy",
		mcp.WithDescription("Permanently delete one of your hierarchies. The blueprints in it are not "+
			"deleted: they stay in your account, unlinked from any hierarchy."),
		annotations("Delete hierarchy", false, true, true),
		idProperty("id", "Hierarchy id, e.g. ha_clx1abc."),
		deleteOutput,
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := requireID(req, "id")
		if err != nil {
			return errorResult(err), nil
		}
		resp, err := lp.DeleteHierarchy(ctx, id)
		if err != nil {
			return errorResult(err), nil
		}
		return jsonResult(resp), nil
	}

	return tool, handler
}
