# Configuration

| Variable          | Default                  | Description                                      |
|-------------------|--------------------------|--------------------------------------------------|
| `LYNXPROMPT_URL`  | `https://lynxprompt.com` | LynxPrompt instance URL (without trailing /)     |
| `LYNXPROMPT_TOKEN`| _(required)_             | API token in `lp_xxx` format                     |
| `LISTEN_ADDR`     | `127.0.0.1:8080`         | HTTP listen address (Docker sets `127.0.0.1:8080`) |
| `MCP_AUTH_TOKEN`  | _(empty)_                | Bearer token for HTTP auth (required if LISTEN_ADDR is not loopback) |
| `TRANSPORT`       | _(empty = HTTP)_         | Set to `stdio` for stdio transport               |

Put them in a `.env` file (from `.env.example`) or set them in the environment.

## MCP client configuration

The npm package always runs over stdio. Add this to your client's `mcpServers` (Claude Desktop, Claude Code, Cursor); add `LYNXPROMPT_URL` to `env` for a self-hosted instance:

```json
{
  "mcpServers": {
    "lynxprompt": {
      "command": "npx",
      "args": ["-y", "lynxprompt-mcp"],
      "env": { "LYNXPROMPT_TOKEN": "lp_..." }
    }
  }
}
```

For the HTTP server (Docker or `go run`), point a client that accepts remote servers at `/mcp`, with the bearer token when `MCP_AUTH_TOKEN` is set:

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
