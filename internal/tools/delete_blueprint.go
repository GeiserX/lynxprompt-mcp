package tools

import (
	"context"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewDeleteBlueprint builds the Tool definition plus its handler.
func NewDeleteBlueprint(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("delete_blueprint",
		mcp.WithDescription("Permanently delete one of your blueprints. It cannot be restored. "+
			"Confirm the id with get_blueprint or list_blueprints first."),
		annotations("Delete blueprint", false, true, true),
		idProperty("id", "Blueprint id, e.g. bp_clx1abc."),
		deleteOutput,
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := requireID(req, "id")
		if err != nil {
			return errorResult(err), nil
		}
		resp, err := lp.DeleteBlueprint(ctx, id)
		if err != nil {
			return errorResult(err), nil
		}
		return jsonResult(resp), nil
	}

	return tool, handler
}
