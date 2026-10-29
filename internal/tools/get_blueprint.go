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
		mcp.WithDescription("Read one blueprint the token can see, including its full content and "+
			"its content_checksum: your own, your team's, or a public one found with "+
			"search_blueprints. Call it before update_blueprint and pass content_checksum as "+
			"expected_checksum, so an edit made elsewhere in the meantime is not overwritten. "+
			"Only the owner can update or delete a blueprint. For a public blueprint of someone "+
			"outside your team, repository_path is only the file name."),
		annotations("Get blueprint", true, false, true),
		idProperty("id", "Blueprint id from list_blueprints or search_blueprints, e.g. bp_clx1abc."),
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
