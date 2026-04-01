package resources

import (
	"context"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterBlueprints wires lynxprompt://blueprints into the server.
func RegisterBlueprints(s *server.MCPServer, lp *client.Client) {
	res := mcp.NewResource(
		"lynxprompt://blueprints",
		"Blueprint list",
		mcp.WithResourceDescription("All blueprints for the authenticated user"),
		mcp.WithMIMEType("application/json"),
	)

	s.AddResource(res, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		body, err := lp.ListBlueprints(ctx)
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "lynxprompt://blueprints",
				MIMEType: "application/json",
				Text:     string(body),
			},
		}, nil
	})
}
