<p align="center">
  <img src="docs/images/banner.svg" alt="LynxPrompt MCP banner" width="900"/>
</p>

<h1 align="center">LynxPrompt-MCP</h1>

<p align="center">
  <a href="https://codecov.io/gh/GeiserX/lynxprompt-mcp"><img src="https://codecov.io/gh/GeiserX/lynxprompt-mcp/graph/badge.svg" alt="codecov"/></a>
  <a href="https://www.npmjs.com/package/lynxprompt-mcp"><img src="https://img.shields.io/npm/v/lynxprompt-mcp?style=flat-square&logo=npm" alt="npm"/></a>
  <img src="https://img.shields.io/badge/Go-1.24-blue?style=flat-square&logo=go&logoColor=white" alt="Go"/>
  <a href="https://hub.docker.com/r/drumsergio/lynxprompt-mcp"><img src="https://img.shields.io/docker/pulls/drumsergio/lynxprompt-mcp?style=flat-square&logo=docker" alt="Docker Pulls"/></a>
  <a href="https://github.com/GeiserX/lynxprompt-mcp/stargazers"><img src="https://img.shields.io/github/stars/GeiserX/lynxprompt-mcp?style=flat-square&logo=github" alt="GitHub Stars"/></a>
  <a href="https://github.com/GeiserX/lynxprompt-mcp/blob/main/LICENSE"><img src="https://img.shields.io/github/license/GeiserX/lynxprompt-mcp?style=flat-square" alt="License"/></a>
</p>
<p align="center">
  <a href="https://registry.modelcontextprotocol.io"><img src="https://img.shields.io/badge/MCP-Official%20Registry-E6522C?style=flat-square" alt="Official MCP Registry"/></a>
  <a href="https://glama.ai/mcp/servers/GeiserX/lynxprompt-mcp"><img src="https://glama.ai/mcp/servers/GeiserX/lynxprompt-mcp/badges/score.svg" alt="Glama MCP Server" /></a>
  <a href="https://mcpservers.org/servers/geiserx/lynxprompt-mcp"><img src="https://img.shields.io/badge/MCPServers.org-listed-green?style=flat-square" alt="MCPServers.org"/></a>
  <a href="https://mcp.so/server/lynxprompt-mcp"><img src="https://img.shields.io/badge/mcp.so-listed-blue?style=flat-square" alt="mcp.so"/></a>
  <a href="https://github.com/toolsdk-ai/toolsdk-mcp-registry"><img src="https://img.shields.io/badge/ToolSDK-Registry-orange?style=flat-square" alt="ToolSDK Registry"/></a>
</p>

<p align="center"><strong>A tiny bridge that exposes any LynxPrompt instance as an MCP server, enabling LLMs to browse, search, and manage AI configuration blueprints.</strong></p>

---

## What you get

| Type          | What for                                                   | MCP URI / Tool id                |
|---------------|------------------------------------------------------------|----------------------------------|
| **Resources** | Browse blueprints, hierarchies, and user info read-only    | `lynxprompt://blueprints`<br>`lynxprompt://blueprint/{id}`<br>`lynxprompt://hierarchies`<br>`lynxprompt://hierarchy/{id}`<br>`lynxprompt://user` |
| **Tools**     | Create, update, delete blueprints and manage hierarchies   | `search_blueprints`<br>`create_blueprint`<br>`update_blueprint`<br>`delete_blueprint`<br>`create_hierarchy`<br>`delete_hierarchy` |

Everything is exposed over a single JSON-RPC endpoint (`/mcp`).
LLMs / Agents can: `initialize` -> `readResource` -> `listTools` -> `callTool` ... and so on.

---

## Quick-start (Docker Compose)

```yaml
services:
  lynxprompt-mcp:
    image: drumsergio/lynxprompt-mcp:latest
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      - LYNXPROMPT_URL=https://lynxprompt.com
      - LYNXPROMPT_TOKEN=lp_xxx
```

> **Security note:** The HTTP transport listens on `127.0.0.1:8080` by default. If you need to expose it on a network, place it behind a reverse proxy with authentication.

## Install via npm (stdio transport)

```sh
npx lynxprompt-mcp
```

Or install globally:

```sh
npm install -g lynxprompt-mcp
lynxprompt-mcp
```

This downloads the pre-built Go binary from GitHub Releases for your platform and runs it with stdio transport. Requires at least one [published release](https://github.com/GeiserX/lynxprompt-mcp/releases).

## Local build

```sh
git clone https://github.com/GeiserX/lynxprompt-mcp
cd lynxprompt-mcp

# (optional) create .env from the sample
cp .env.example .env && $EDITOR .env

go run ./cmd/server
```

## Configuration

| Variable          | Default                  | Description                                      |
|-------------------|--------------------------|--------------------------------------------------|
| `LYNXPROMPT_URL`  | `https://lynxprompt.com` | LynxPrompt instance URL (without trailing /)     |
| `LYNXPROMPT_TOKEN`| _(required)_             | API token in `lp_xxx` format                     |
| `LISTEN_ADDR`     | `127.0.0.1:8080`         | HTTP listen address (Docker sets `0.0.0.0:8080`) |
| `TRANSPORT`       | _(empty = HTTP)_         | Set to `stdio` for stdio transport               |

Put them in a `.env` file (from `.env.example`) or set them in the environment.

## Testing

Tested with [Inspector](https://modelcontextprotocol.io/docs/tools/inspector) and it is currently fully working. Before making a PR, make sure this MCP server behaves well via this medium.

## Example configuration for client LLMs

```json
{
  "schema_version": "v1",
  "name_for_human": "LynxPrompt-MCP",
  "name_for_model": "lynxprompt_mcp",
  "description_for_human": "Browse, search, and manage AI configuration blueprints from LynxPrompt.",
  "description_for_model": "Interact with a LynxPrompt instance that stores AI configuration blueprints. First call initialize, then reuse the returned session id in header \"Mcp-Session-Id\" for every other call. Use readResource to fetch URIs that begin with lynxprompt://. Use listTools to discover available actions and callTool to execute them.",
  "auth": { "type": "none" },
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

## Credits

[LynxPrompt](https://lynxprompt.com) -- AI configuration blueprint management

[MCP-GO](https://github.com/mark3labs/mcp-go) -- modern MCP implementation

[GoReleaser](https://goreleaser.com/) -- painless multi-arch releases

## Maintainers

[@GeiserX](https://github.com/GeiserX).

## Contributing

Feel free to dive in! [Open an issue](https://github.com/GeiserX/lynxprompt-mcp/issues/new) or submit PRs.

LynxPrompt-MCP follows the [Contributor Covenant](http://contributor-covenant.org/version/2/1/) Code of Conduct.

## Other MCP Servers by GeiserX

- [cashpilot-mcp](https://github.com/GeiserX/cashpilot-mcp) — Passive income monitoring
- [duplicacy-mcp](https://github.com/GeiserX/duplicacy-mcp) — Backup health monitoring
- [genieacs-mcp](https://github.com/GeiserX/genieacs-mcp) — TR-069 device management
- [pumperly-mcp](https://github.com/GeiserX/pumperly-mcp) — Fuel and EV charging prices
- [telegram-archive-mcp](https://github.com/GeiserX/telegram-archive-mcp) — Telegram message archive
