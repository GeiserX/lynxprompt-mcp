package resources

import (
	"context"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterHierarchies wires lynxprompt://hierarchies into the server.
func RegisterHierarchies(s *server.MCPServer, lp *client.Client) {
	res := mcp.NewResource(
		"lynxprompt://hierarchies",
		"Hierarchy list",
		mcp.WithResourceDescription("Up to 100 hierarchies of the authenticated user (the list_hierarchies tool pages past that)"),
		mcp.WithMIMEType("application/json"),
	)

	s.AddResource(res, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		body, err := lp.ListHierarchies(ctx, client.ListOptions{Limit: 100})
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "lynxprompt://hierarchies",
				MIMEType: "application/json",
				Text:     string(body),
			},
		}, nil
	})
}
