# Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `LYNXPROMPT_TOKEN` | _(empty)_ | API token (`lp_...`) from Settings, API tokens. Every tool except `search_blueprints` needs it. Secret. |
| `LYNXPROMPT_URL` | `https://lynxprompt.com` | LynxPrompt instance URL. Must be `https://` unless the host is loopback. |
| `LYNXPROMPT_ALLOW_HTTP` | `false` | Allow an `http://` `LYNXPROMPT_URL` on a remote host. The token then travels unencrypted. |
| `LYNXPROMPT_READ_ONLY` | `false` | `true` registers only the read tools: no create, update or delete. |
| `TRANSPORT` | _(empty = HTTP)_ | `stdio` for stdio. The npm package always uses stdio; the Docker image defaults to stdio. |
| `LISTEN_ADDR` | `127.0.0.1:8080` (Docker: `0.0.0.0:8080`) | HTTP listen address. |
| `MCP_AUTH_TOKEN` | _(empty)_ | Bearer token for the HTTP endpoint. Required, and at least 32 characters, when `LISTEN_ADDR` is not loopback. Generate one with `openssl rand -hex 32`. |
| `MCP_ALLOWED_ORIGINS` | _(empty)_ | Comma-separated browser origins allowed to call the HTTP endpoint cross-site, e.g. `https://app.example.com`. Clients that are not browsers send no Origin and are unaffected. |

In HTTP mode they can also come from a `.env` file (see `.env.example`).

## HTTP endpoint

The HTTP server is stateless Streamable HTTP on `/mcp`. It rejects cross-site browser requests (403), bodies over 4 MiB (413) and, when `MCP_AUTH_TOKEN` is set, any request without `Authorization: Bearer <token>` (401). Point a client that accepts remote servers at it:

```json
{
  "mcpServers": {
    "lynxprompt": {
      "type": "http",
      "url": "http://127.0.0.1:8080/mcp",
      "headers": { "Authorization": "Bearer <your MCP_AUTH_TOKEN>" }
    }
  }
}
```

Claude Code: `claude mcp add --transport http lynxprompt http://127.0.0.1:8080/mcp --header "Authorization: Bearer <your MCP_AUTH_TOKEN>"`.

The stdio client snippets for Claude Code, Claude Desktop, Cursor, VS Code and Codex are in the [README](https://github.com/GeiserX/lynxprompt-mcp#install).
