package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/geiserx/lynxprompt-mcp/config"
	"github.com/geiserx/lynxprompt-mcp/internal/mcpserver"
	"github.com/geiserx/lynxprompt-mcp/version"
	"github.com/mark3labs/mcp-go/server"
)

const (
	// maxRequestBytes caps one HTTP request body. A blueprint is text, so
	// 4 MiB leaves plenty of room; the SDK would otherwise read any size.
	maxRequestBytes = 4 << 20
	// minAuthTokenLen is enforced when the HTTP server is reachable from
	// other hosts.
	minAuthTokenLen = 32
	endpointPath    = "/mcp"
)

func isLoopbackAddr(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// bearerAuth requires "Authorization: Bearer <token>" (scheme matched
// case-insensitively) and compares the token in constant time.
func bearerAuth(next http.Handler, token string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme, got, ok := strings.Cut(r.Header.Get("Authorization"), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") ||
			subtle.ConstantTimeCompare([]byte(strings.TrimSpace(got)), []byte(token)) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="lynxprompt-mcp"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// limitBody answers 413 for a declared oversize body and caps the rest.
func limitBody(next http.Handler, max int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > max {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, max)
		next.ServeHTTP(w, r)
	})
}

// httpConfig is everything the HTTP transport reads from the environment.
type httpConfig struct {
	Addr           string
	AuthToken      string
	AllowedOrigins []string
}

func loadHTTPConfig() httpConfig {
	c := httpConfig{
		Addr:      os.Getenv("LISTEN_ADDR"),
		AuthToken: os.Getenv("MCP_AUTH_TOKEN"),
	}
	if c.Addr == "" {
		c.Addr = "127.0.0.1:8080"
	}
	for _, o := range strings.Split(os.Getenv("MCP_ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			c.AllowedOrigins = append(c.AllowedOrigins, o)
		}
	}
	return c
}

func (c httpConfig) validate() error {
	if isLoopbackAddr(c.Addr) {
		return nil
	}
	if c.AuthToken == "" {
		return errors.New("MCP_AUTH_TOKEN is required when LISTEN_ADDR is not loopback")
	}
	if len(c.AuthToken) < minAuthTokenLen {
		return fmt.Errorf("MCP_AUTH_TOKEN must be at least %d characters when LISTEN_ADDR is not loopback (try: openssl rand -hex 32)", minAuthTokenLen)
	}
	return nil
}

// newHTTPHandler wraps the stateless Streamable HTTP transport in the checks
// the SDK does not make: Origin validation, a body cap and, when configured,
// bearer auth.
func newHTTPHandler(s *server.MCPServer, c httpConfig) (http.Handler, error) {
	mcpHTTP := server.NewStreamableHTTPServer(s,
		server.WithEndpointPath(endpointPath),
		server.WithStateLess(true),
	)
	cop := http.NewCrossOriginProtection()
	for _, o := range c.AllowedOrigins {
		if err := cop.AddTrustedOrigin(o); err != nil {
			return nil, fmt.Errorf("MCP_ALLOWED_ORIGINS: %w", err)
		}
	}
	var h http.Handler = limitBody(mcpHTTP, maxRequestBytes)
	h = cop.Handler(h)
	if c.AuthToken != "" {
		h = bearerAuth(h, c.AuthToken)
	}
	mux := http.NewServeMux()
	mux.Handle(endpointPath, h)
	return mux, nil
}

func newHTTPServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}
}

func main() {
	log.Printf("LynxPrompt MCP %s starting...", version.String())

	cfg := config.LoadConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	lp := client.New(cfg.BaseURL, cfg.Token)
	s := mcpserver.New(lp, mcpserver.Options{ReadOnly: cfg.ReadOnly})
	if cfg.ReadOnly {
		log.Println("read-only mode: create, update and delete tools are not registered")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if strings.ToLower(os.Getenv("TRANSPORT")) == "stdio" {
		log.Println("LynxPrompt MCP bridge running on stdio")
		if err := server.NewStdioServer(s).Listen(ctx, os.Stdin, os.Stdout); err != nil && !errors.Is(err, context.Canceled) {
			log.Fatalf("stdio server error: %v", err)
		}
		return
	}

	hc := loadHTTPConfig()
	if err := hc.validate(); err != nil {
		log.Fatal(err)
	}
	h, err := newHTTPHandler(s, hc)
	if err != nil {
		log.Fatal(err)
	}
	srv := newHTTPServer(hc.Addr, h)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	auth := ""
	if hc.AuthToken != "" {
		auth = " (auth enabled)"
	}
	log.Printf("LynxPrompt MCP bridge listening on %s%s%s", hc.Addr, endpointPath, auth)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
