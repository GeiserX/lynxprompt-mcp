package tools

import (
	"context"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewUpdateBlueprint builds the Tool definition plus its handler.
func NewUpdateBlueprint(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("update_blueprint",
		mcp.WithDescription("Change one of your blueprints. Only the fields you pass change; content "+
			"replaces the whole text and the old text is not kept. Read the blueprint with "+
			"get_blueprint first and pass its content_checksum as expected_checksum: if someone "+
			"edited it since (web UI, CLI, another agent) the call fails with 409 instead of "+
			"overwriting their change. Setting visibility PUBLIC publishes it to the marketplace."),
		annotations("Update blueprint", false, true, true),
		idProperty("id", "Blueprint id, e.g. bp_clx1abc."),
		nameProperty("New name, at least 3 characters."),
		contentProperty("New full text; replaces the current content."),
		mcp.WithString("expected_checksum",
			mcp.Description("content_checksum from get_blueprint. Strongly recommended whenever content changes."),
			mcp.MaxLength(128),
		),
		typeProperty,
		descriptionProperty,
		visibilityProperty,
		categoryProperty,
		tagsProperty,
		blueprintOutput,
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := requireID(req, "id")
		if err != nil {
			return errorResult(err), nil
		}

		body := map[string]any{}
		if err := commonBlueprintFields(req, body); err != nil {
			return errorResult(err), nil
		}
		if content, ok := req.GetArguments()["content"].(string); ok && content != "" {
			if err := contentField(content, body); err != nil {
				return errorResult(err), nil
			}
		}
		if len(body) == 0 {
			return mcp.NewToolResultError("nothing to update: provide at least one of name, content, description, visibility, type, category or tags"), nil
		}
		if v, ok := optString(req, "expected_checksum"); ok {
			body["expected_checksum"] = v
		}

		resp, err := lp.UpdateBlueprint(ctx, id, body)
		if err != nil {
			return errorResult(err), nil
		}
		return jsonResult(resp), nil
	}

	return tool, handler
}
