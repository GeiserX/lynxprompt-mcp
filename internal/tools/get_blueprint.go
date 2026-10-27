package tools

import (
	"context"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewGetBlueprint builds the Tool definition plus its handler.
func NewGetBlueprint(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("get_blueprint",
		mcp.WithDescription("Read one of your own blueprints, including its full content and its "+
			"content_checksum. Call it before update_blueprint and pass content_checksum as "+
			"expected_checksum, so an edit made elsewhere in the meantime is not overwritten. "+
			"Only works for blueprints owned by the token's user."),
		annotations("Get my blueprint", true, false, true),
		idProperty("id", "Blueprint id from list_blueprints, e.g. bp_clx1abc."),
		blueprintOutput,
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := requireID(req, "id")
		if err != nil {
			return errorResult(err), nil
		}
		resp, err := lp.GetBlueprint(ctx, id)
		if err != nil {
			return errorResult(err), nil
		}
		return jsonResult(resp), nil
	}

	return tool, handler
}
