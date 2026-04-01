package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/geiserx/lynxprompt-mcp/config"
	"github.com/geiserx/lynxprompt-mcp/internal/resources"
	"github.com/geiserx/lynxprompt-mcp/internal/tools"
	"github.com/geiserx/lynxprompt-mcp/version"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	log.Printf("LynxPrompt MCP %s starting...", version.String())

	// Load config & initialise LynxPrompt client
	cfg := config.LoadConfig()
	lp := client.New(cfg.BaseURL, cfg.Token)

	// Create MCP server
	s := server.NewMCPServer(
		"LynxPrompt MCP Bridge",
		"0.0.1",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	// -----------------------------------------------------------------
	// RESOURCES
	// -----------------------------------------------------------------

	// lynxprompt://blueprints
	resources.RegisterBlueprints(s, lp)

	// lynxprompt://blueprint/{id}
	resources.RegisterBlueprint(s, lp)

	// lynxprompt://hierarchies
	resources.RegisterHierarchies(s, lp)

	// lynxprompt://hierarchy/{id}
	resources.RegisterHierarchy(s, lp)

	// lynxprompt://user
	resources.RegisterUser(s, lp)

	// -----------------------------------------------------------------
	// TOOLS
	// -----------------------------------------------------------------

	// search_blueprints
	tool, handler := tools.NewSearchBlueprints(lp)
	s.AddTool(tool, handler)

	// create_blueprint
	tool, handler = tools.NewCreateBlueprint(lp)
	s.AddTool(tool, handler)

	// update_blueprint
	tool, handler = tools.NewUpdateBlueprint(lp)
	s.AddTool(tool, handler)

	// delete_blueprint
	tool, handler = tools.NewDeleteBlueprint(lp)
	s.AddTool(tool, handler)

	// create_hierarchy
	tool, handler = tools.NewCreateHierarchy(lp)
	s.AddTool(tool, handler)

	// delete_hierarchy
	tool, handler = tools.NewDeleteHierarchy(lp)
	s.AddTool(tool, handler)

	// -----------------------------------------------------------------
	// TRANSPORT
	// -----------------------------------------------------------------
	transport := strings.ToLower(os.Getenv("TRANSPORT"))
	if transport == "http" {
		httpSrv := server.NewStreamableHTTPServer(s)
		log.Println("LynxPrompt MCP bridge listening on :8080")
		if err := httpSrv.Start(":8080"); err != nil {
			log.Fatalf("server error: %v", err)
		}
	} else {
		stdioSrv := server.NewStdioServer(s)
		log.Println("LynxPrompt MCP bridge running on stdio")
		if err := stdioSrv.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
			log.Fatalf("stdio server error: %v", err)
		}
	}
}
