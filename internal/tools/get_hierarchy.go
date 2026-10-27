package tools

import (
	"context"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewGetHierarchy builds the Tool definition plus its handler.
func NewGetHierarchy(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("get_hierarchy",
		mcp.WithDescription("Read one of your hierarchies with every blueprint in it, both as a flat list "+
			"(blueprints[], with repository_path and content_checksum) and as a tree nested by parent_id. "+
			"Blueprint content is not included; use get_blueprint for that."),
		annotations("Get my hierarchy", true, false, true),
		idProperty("id", "Hierarchy id from list_hierarchies, e.g. ha_clx1abc."),
		hierarchyOutput,
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := requireID(req, "id")
		if err != nil {
			return errorResult(err), nil
		}
		resp, err := lp.GetHierarchy(ctx, id)
		if err != nil {
			return errorResult(err), nil
		}
		return jsonResult(resp), nil
	}

	return tool, handler
}
