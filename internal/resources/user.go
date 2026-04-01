package resources

import (
	"context"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterUser wires lynxprompt://user into the server.
func RegisterUser(s *server.MCPServer, lp *client.Client) {
	res := mcp.NewResource(
		"lynxprompt://user",
		"Authenticated user info",
		mcp.WithResourceDescription("Profile of the currently authenticated LynxPrompt user"),
		mcp.WithMIMEType("application/json"),
	)

	s.AddResource(res, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		body, err := lp.GetUser()
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "lynxprompt://user",
				MIMEType: "application/json",
				Text:     string(body),
			},
		}, nil
	})
}
