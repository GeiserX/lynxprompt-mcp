# Configuration

| Variable          | Default                  | Description                                      |
|-------------------|--------------------------|--------------------------------------------------|
| `LYNXPROMPT_URL`  | `https://lynxprompt.com` | LynxPrompt instance URL (without trailing /)     |
| `LYNXPROMPT_TOKEN`| _(required)_             | API token in `lp_xxx` format                     |
| `LISTEN_ADDR`     | `127.0.0.1:8080`         | HTTP listen address (Docker sets `127.0.0.1:8080`) |
| `MCP_AUTH_TOKEN`  | _(empty)_                | Bearer token for HTTP auth (required if LISTEN_ADDR is not loopback) |
| `TRANSPORT`       | _(empty = HTTP)_         | Set to `stdio` for stdio transport               |

Put them in a `.env` file (from `.env.example`) or set them in the environment.

## Example configuration for client LLMs

```json
{
  "schema_version": "v1",
  "name_for_human": "LynxPrompt-MCP",
  "name_for_model": "lynxprompt_mcp",
  "description_for_human": "Browse, search, and manage AI configuration blueprints from LynxPrompt.",
  "description_for_model": "Interact with a LynxPrompt instance that stores AI configuration blueprints. First call initialize, then reuse the returned session id in header \"Mcp-Session-Id\" for every other call. Use readResource to fetch URIs that begin with lynxprompt://. Use listTools to discover available actions and callTool to execute them.",
  "auth": { "type": "bearer", "token": "your-secret-token" },
  "api": {
    "type": "jsonrpc-mcp",
    "url":  "http://localhost:8080/mcp",
    "init_method": "initialize",
    "session_header": "Mcp-Session-Id"
  },
  "logo_url": "https://lynxprompt.com/logo.png",
  "contact_email": "acsdesk@protonmail.com",
  "legal_info_url": "https://github.com/GeiserX/lynxprompt-mcp/blob/main/LICENSE"
}
```
