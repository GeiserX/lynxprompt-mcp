package tools

import (
	"context"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewCreateBlueprint builds the Tool definition plus its handler.
func NewCreateBlueprint(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("create_blueprint",
		mcp.WithDescription("Save a new blueprint (an AI agent instruction file such as AGENTS.md or "+
			"CLAUDE.md) to the token user's LynxPrompt account. Each call creates a new blueprint, "+
			"even with the same name; to change an existing one use update_blueprint. Private by "+
			"default; visibility PUBLIC publishes it to the public marketplace. To place it in a "+
			"monorepo hierarchy pass hierarchy_id and repository_path. Returns the new blueprint with its bp_ id."),
		annotations("Create blueprint", false, false, false),
		nameProperty("Blueprint name, at least 3 characters, e.g. \"Next.js app AGENTS.md\".", mcp.Required()),
		contentProperty("The full text of the instruction file, at least 10 characters.", mcp.Required()),
		typeProperty,
		descriptionProperty,
		visibilityProperty,
		teamIDProperty,
		categoryProperty,
		tagsProperty,
		mcp.WithString("hierarchy_id",
			mcp.Description("Hierarchy to add the blueprint to (ha_..., from list_hierarchies or create_hierarchy)."),
			mcp.Pattern(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`),
		),
		mcp.WithString("parent_id",
			mcp.Description("Parent blueprint inside the hierarchy (bp_...), e.g. the root AGENTS.md of the repo."),
			mcp.Pattern(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`),
		),
		mcp.WithString("repository_path",
			mcp.Description("Where the file lives in the repository, relative to its root, e.g. packages/api/AGENTS.md."),
			mcp.MaxLength(500),
		),
		blueprintOutput,
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := req.RequireString("name"); err != nil {
			return errorResult(err), nil
		}
		content, err := req.RequireString("content")
		if err != nil {
			return errorResult(err), nil
		}

		body := map[string]any{}
		if err := commonBlueprintFields(req, body); err != nil {
			return errorResult(err), nil
		}
		if _, ok := body["name"]; !ok {
			return mcp.NewToolResultError("name must not be empty"), nil
		}
		if err := contentField(content, body); err != nil {
			return errorResult(err), nil
		}
		for _, key := range []string{"hierarchy_id", "parent_id"} {
			if v, ok := optString(req, key); ok {
				if err := client.ValidateID(v); err != nil {
					return errorResult(err), nil
				}
				body[key] = v
			}
		}
		if v, ok := optString(req, "repository_path"); ok {
			body["repository_path"] = v
		}
		if err := teamIDField(req, body); err != nil {
			return errorResult(err), nil
		}

		resp, err := lp.CreateBlueprint(ctx, body)
		if err != nil {
			return errorResult(err), nil
		}
		return jsonResult(resp), nil
	}

	return tool, handler
}
