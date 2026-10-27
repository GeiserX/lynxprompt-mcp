package tools

import (
	"context"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewListHierarchies builds the Tool definition plus its handler.
func NewListHierarchies(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("list_hierarchies",
		mcp.WithDescription("List your hierarchies. A hierarchy groups the blueprints of one repository "+
			"(for example the AGENTS.md files of a monorepo) under one repository_root. Returns ids "+
			"(ha_...) for get_hierarchy and delete_hierarchy. Paged like list_blueprints."),
		annotations("List my hierarchies", true, false, true),
		mcp.WithInteger("limit",
			mcp.Description("Hierarchies per page, 1 to 100. Default 50."),
			mcp.Min(1), mcp.Max(100),
		),
		mcp.WithInteger("offset",
			mcp.Description("How many hierarchies to skip. Default 0."),
			mcp.Min(0), mcp.Max(100000),
		),
		listHierarchiesOutput,
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		opts, err := listOptions(req)
		if err != nil {
			return errorResult(err), nil
		}
		resp, err := lp.ListHierarchies(ctx, opts)
		if err != nil {
			return errorResult(err), nil
		}
		return jsonResult(resp), nil
	}

	return tool, handler
}
