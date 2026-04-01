package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewDeleteHierarchy builds the Tool definition plus its handler.
func NewDeleteHierarchy(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("delete_hierarchy",
		mcp.WithDescription("Delete a hierarchy from LynxPrompt"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("ID of the hierarchy to delete"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := lp.DeleteHierarchy(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(
			fmt.Sprintf("Hierarchy deleted: %s", string(resp)),
		), nil
	}

	return tool, handler
}
