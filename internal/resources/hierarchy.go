package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterHierarchy wires lynxprompt://hierarchy/{id} into the server.
func RegisterHierarchy(s *server.MCPServer, lp *client.Client) {
	tpl := mcp.NewResourceTemplate(
		"lynxprompt://hierarchy/{id}",
		"Single hierarchy with tree",
		mcp.WithTemplateDescription("Full hierarchy document including tree structure"),
		mcp.WithTemplateMIMEType("application/json"),
	)

	s.AddResourceTemplate(tpl, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		id := strings.TrimPrefix(req.Params.URI, "lynxprompt://hierarchy/")
		if id == "" {
			return nil, fmt.Errorf("missing hierarchy id")
		}

		body, err := lp.GetHierarchy(ctx, id)
		if err != nil {
			return nil, err
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      fmt.Sprintf("lynxprompt://hierarchy/%s", id),
				MIMEType: "application/json",
				Text:     string(body),
			},
		}, nil
	})
}
