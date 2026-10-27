package tools

import (
	"context"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var listVisibilities = []string{"all", "PRIVATE", "TEAM", "PUBLIC"}

// NewListBlueprints builds the Tool definition plus its handler.
func NewListBlueprints(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("list_blueprints",
		mcp.WithDescription("List the blueprints owned by the user of LYNXPROMPT_TOKEN, newest change first, "+
			"without their content. Use it to find a blueprint id (bp_...) for get_blueprint, "+
			"update_blueprint or delete_blueprint. Paged: when has_more is true, call again with "+
			"offset increased by limit."),
		annotations("List my blueprints", true, false, true),
		mcp.WithInteger("limit",
			mcp.Description("Blueprints per page, 1 to 100. Default 50."),
			mcp.Min(1), mcp.Max(100),
		),
		mcp.WithInteger("offset",
			mcp.Description("How many blueprints to skip. Default 0."),
			mcp.Min(0), mcp.Max(100000),
		),
		mcp.WithString("visibility",
			mcp.Description("Only blueprints with this visibility. Default all."),
			mcp.Enum(listVisibilities...),
		),
		listBlueprintsOutput,
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		opts, err := listOptions(req)
		if err != nil {
			return errorResult(err), nil
		}
		if v, ok := optString(req, "visibility"); ok {
			if err := oneOf("visibility", v, listVisibilities); err != nil {
				return errorResult(err), nil
			}
			if v != "all" {
				opts.Visibility = v
			}
		}
		resp, err := lp.ListBlueprints(ctx, opts)
		if err != nil {
			return errorResult(err), nil
		}
		return jsonResult(resp), nil
	}

	return tool, handler
}

func listOptions(req mcp.CallToolRequest) (client.ListOptions, error) {
	var opts client.ListOptions
	var err error
	if opts.Limit, err = optInt(req, "limit", 0, 1, 100); err != nil {
		return opts, err
	}
	if opts.Offset, err = optInt(req, "offset", 0, 0, 100000); err != nil {
		return opts, err
	}
	return opts, nil
}
