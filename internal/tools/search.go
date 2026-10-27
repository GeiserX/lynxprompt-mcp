package tools

import (
	"context"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	searchTypes = []string{"configs", "commands"}
	searchSorts = []string{"popular", "recent", "downloads", "favorites"}
)

// NewSearchBlueprints builds the Tool definition plus its handler.
func NewSearchBlueprints(lp *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("search_blueprints",
		mcp.WithDescription("Search the public LynxPrompt marketplace of AI agent instruction files "+
			"(AGENTS.md, CLAUDE.md, Cursor rules, slash commands and so on). Use it to find "+
			"blueprints other people published. To read or change your own blueprints use "+
			"list_blueprints and get_blueprint instead. Needs no token. Returns one page of "+
			"results (templates[] with id, name, description, author, type, tags) plus total and hasMore."),
		annotations("Search marketplace blueprints", true, false, true),
		mcp.WithString("query",
			mcp.Description("Words to match in name, description or tags, e.g. \"nextjs\". Leave out to list the most popular blueprints."),
			mcp.MaxLength(200),
		),
		mcp.WithString("type",
			mcp.Description("\"configs\" for instruction files, \"commands\" for slash commands and workflows. Leave out for both."),
			mcp.Enum(searchTypes...),
		),
		mcp.WithString("category",
			mcp.Description("Category slug, e.g. web, devops, saas, mobile, data or other."),
			mcp.MaxLength(50),
		),
		mcp.WithArray("tags",
			mcp.Description("Match blueprints carrying any of these tags, e.g. [\"react\", \"typescript\"]."),
			mcp.WithStringItems(mcp.MaxLength(30)),
			mcp.MaxItems(10),
		),
		mcp.WithString("sort",
			mcp.Description("Result order. Default popular."),
			mcp.Enum(searchSorts...),
		),
		mcp.WithInteger("page",
			mcp.Description("Page number, starting at 1. Default 1."),
			mcp.Min(1), mcp.Max(1000),
		),
		mcp.WithInteger("limit",
			mcp.Description("Results per page, 1 to 50. Default 12."),
			mcp.Min(1), mcp.Max(50),
		),
		searchOutput,
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		p := client.SearchParams{}
		p.Query, _ = optString(req, "query")
		p.Category, _ = optString(req, "category")
		if v, ok := optString(req, "type"); ok {
			if err := oneOf("type", v, searchTypes); err != nil {
				return errorResult(err), nil
			}
			p.Type = v
		}
		if v, ok := optString(req, "sort"); ok {
			if err := oneOf("sort", v, searchSorts); err != nil {
				return errorResult(err), nil
			}
			p.Sort = v
		}
		var err error
		if p.Tags, err = optStringList(req, "tags"); err != nil {
			return errorResult(err), nil
		}
		if p.Page, err = optInt(req, "page", 0, 1, 1000); err != nil {
			return errorResult(err), nil
		}
		if p.Limit, err = optInt(req, "limit", 0, 1, 50); err != nil {
			return errorResult(err), nil
		}

		resp, err := lp.SearchBlueprints(ctx, p)
		if err != nil {
			return errorResult(err), nil
		}
		return jsonResult(resp), nil
	}

	return tool, handler
}
