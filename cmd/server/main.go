package main

import (
	"context"
	"crypto/subtle"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/geiserx/lynxprompt-mcp/config"
	"github.com/geiserx/lynxprompt-mcp/internal/resources"
	"github.com/geiserx/lynxprompt-mcp/internal/tools"
	"github.com/geiserx/lynxprompt-mcp/version"
	"github.com/mark3labs/mcp-go/server"
)

func isLoopbackAddr(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "" || host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func bearerAuth(next http.Handler, token string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	log.Printf("LynxPrompt MCP %s starting...", version.String())

	// Load config & initialise LynxPrompt client
	cfg := config.LoadConfig()
	lp := client.New(cfg.BaseURL, cfg.Token)

	// Create MCP server
	s := server.NewMCPServer(
		"LynxPrompt MCP Bridge",
		version.Version,
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
	if transport == "stdio" {
		stdioSrv := server.NewStdioServer(s)
		log.Println("LynxPrompt MCP bridge running on stdio")
		if err := stdioSrv.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
			log.Fatalf("stdio server error: %v", err)
		}
	} else {
		httpSrv := server.NewStreamableHTTPServer(s)
		addr := os.Getenv("LISTEN_ADDR")
		if addr == "" {
			addr = "127.0.0.1:8080"
		}
		authToken := os.Getenv("MCP_AUTH_TOKEN")
		if authToken == "" && !isLoopbackAddr(addr) {
			log.Fatal("MCP_AUTH_TOKEN is required when LISTEN_ADDR is not loopback")
		}
		if authToken != "" {
			mux := http.NewServeMux()
			mux.Handle("/mcp", bearerAuth(httpSrv, authToken))
			log.Printf("LynxPrompt MCP bridge listening on %s (auth enabled)", addr)
			if err := http.ListenAndServe(addr, mux); err != nil {
				log.Fatalf("server error: %v", err)
			}
		} else {
			log.Printf("LynxPrompt MCP bridge listening on %s", addr)
			if err := httpSrv.Start(addr); err != nil {
				log.Fatalf("server error: %v", err)
			}
		}
	}
}
