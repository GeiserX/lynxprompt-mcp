// Package mcpserver assembles the LynxPrompt MCP server: identity,
// capabilities, resources and tools. cmd/server only picks the transport.
package mcpserver

import (
	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/geiserx/lynxprompt-mcp/internal/resources"
	"github.com/geiserx/lynxprompt-mcp/internal/tools"
	"github.com/geiserx/lynxprompt-mcp/version"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	// Name is the server name reported in serverInfo.
	Name = "LynxPrompt MCP Bridge"

	instructions = `LynxPrompt stores AI agent instruction files ("blueprints": AGENTS.md, CLAUDE.md, Cursor rules, slash commands).
- Blueprints have bp_ ids. list_blueprints lists your own; search_blueprints searches the public marketplace of other people's.
- get_blueprint reads any blueprint you can see (yours, your team's, public ones) with its content and content_checksum. Only the owner can update or delete one.
- Before update_blueprint, call get_blueprint and pass its content_checksum as expected_checksum; a 409 means someone else changed it, so re-read and merge.
- visibility PUBLIC publishes a blueprint to everyone. delete_blueprint is permanent.
- A hierarchy (ha_ id) groups the blueprints of one repository; delete_hierarchy only unlinks its blueprints.`

	instructionsReadOnly = "\n- This server runs read-only: the create, update and delete tools are not available."

	// Static lists can be cached for five minutes; anything read from the
	// user's account is private and never cached.
	listCacheTTLMs = 5 * 60 * 1000

	// toolResultLimit bounds one tool result (structured plus text copy).
	toolResultLimit = 10 << 20
)

// Options configure New.
type Options struct {
	ReadOnly bool
	// ValidateOutput checks every structured tool result against its output
	// schema. Tests turn it on; production leaves it off so a new upstream
	// field type never turns a good answer into an error.
	ValidateOutput bool
}

// New builds the MCP server around a LynxPrompt API client.
func New(lp *client.Client, opts Options) *server.MCPServer {
	instr := instructions
	if opts.ReadOnly {
		instr += instructionsReadOnly
	}
	serverOpts := []server.ServerOption{
		// The tool list never changes while the process runs.
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, false),
		server.WithRecovery(),
		server.WithResourceRecovery(),
		server.WithInstructions(instr),
		server.WithTitle("LynxPrompt"),
		server.WithDescription("Search, read and manage AI agent instruction blueprints on a LynxPrompt instance."),
		server.WithWebsiteURL("https://github.com/GeiserX/lynxprompt-mcp"),
		server.WithIcons(
			mcp.Icon{Src: "https://lynxprompt.geiser.cloud/favicon.svg", MIMEType: "image/svg+xml", Sizes: []string{"any"}},
			mcp.Icon{Src: "https://lynxprompt.geiser.cloud/icon-192.png", MIMEType: "image/png", Sizes: []string{"192x192"}},
		),
		server.WithCacheHints(listCacheTTLMs, mcp.CacheScopePublic),
		server.WithMethodCacheHints(mcp.MethodResourcesRead, 0, mcp.CacheScopePrivate),
		server.WithInputSchemaValidation(),
		server.WithStrictInputSchemaDefault(),
		server.WithToolResultSizeLimit(toolResultLimit),
	}
	if opts.ValidateOutput {
		serverOpts = append(serverOpts, server.WithOutputSchemaValidation())
	}
	s := server.NewMCPServer(Name, version.Version, serverOpts...)

	resources.RegisterBlueprints(s, lp)
	resources.RegisterBlueprint(s, lp)
	resources.RegisterHierarchies(s, lp)
	resources.RegisterHierarchy(s, lp)
	resources.RegisterUser(s, lp)

	tools.Register(s, lp, opts.ReadOnly)
	return s
}
