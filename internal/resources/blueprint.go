package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterBlueprint wires lynxprompt://blueprint/{id} into the server.
func RegisterBlueprint(s *server.MCPServer, lp *client.Client) {
	tpl := mcp.NewResourceTemplate(
		"lynxprompt://blueprint/{id}",
		"Single blueprint with content",
		mcp.WithTemplateDescription("Full blueprint document including content"),
		mcp.WithTemplateMIMEType("application/json"),
	)

	s.AddResourceTemplate(tpl, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		id := strings.TrimPrefix(req.Params.URI, "lynxprompt://blueprint/")
		if id == "" {
			return nil, fmt.Errorf("missing blueprint id")
		}

		body, err := lp.GetBlueprint(ctx, id)
		if err != nil {
			return nil, err
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      fmt.Sprintf("lynxprompt://blueprint/%s", id),
				MIMEType: "application/json",
				Text:     string(body),
			},
		}, nil
	})
}
