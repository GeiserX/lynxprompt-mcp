package tools

import (
	"context"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewCreateHierarchy builds the Tool definition plus its handler.
func NewCreateHierarchy(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("create_hierarchy",
		mcp.WithDescription("Create a hierarchy: a group that ties the blueprints of one repository "+
			"together, for example every AGENTS.md of a monorepo. repository_root identifies the "+
			"repository; if you already have a hierarchy with that repository_root, it is returned "+
			"unchanged instead of a duplicate. Add blueprints to it with create_blueprint (hierarchy_id "+
			"plus repository_path). Returns the hierarchy with its ha_ id."),
		annotations("Create hierarchy", false, false, true),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Display name, e.g. \"acme monorepo\"."),
			mcp.MinLength(1), mcp.MaxLength(200),
		),
		mcp.WithString("repository_root",
			mcp.Required(),
			mcp.Description("Unique identifier of the repository, e.g. github.com/acme/monorepo."),
			mcp.MinLength(1), mcp.MaxLength(500),
		),
		mcp.WithString("description",
			mcp.Description("Short description of the hierarchy."),
			mcp.MaxLength(1000),
		),
		hierarchyOutput,
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return errorResult(err), nil
		}
		repoRoot, err := req.RequireString("repository_root")
		if err != nil {
			return errorResult(err), nil
		}

		body := map[string]any{
			"name":            name,
			"repository_root": repoRoot,
		}
		if desc, ok := optString(req, "description"); ok {
			body["description"] = desc
		}

		resp, err := lp.CreateHierarchy(ctx, body)
		if err != nil {
			return errorResult(err), nil
		}
		return jsonResult(resp), nil
	}

	return tool, handler
}
