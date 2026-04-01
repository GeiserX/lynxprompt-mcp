package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewDeleteBlueprint builds the Tool definition plus its handler.
func NewDeleteBlueprint(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("delete_blueprint",
		mcp.WithDescription("Delete a blueprint from LynxPrompt"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("ID of the blueprint to delete"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := lp.DeleteBlueprint(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(
			fmt.Sprintf("Blueprint deleted: %s", string(resp)),
		), nil
	}

	return tool, handler
}
