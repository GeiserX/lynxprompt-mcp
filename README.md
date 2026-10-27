<p align="center">
  <img src="https://raw.githubusercontent.com/GeiserX/lynxprompt-mcp/main/docs/images/banner.svg" alt="lynxprompt-mcp" width="900"/>
</p>

<h1 align="center">lynxprompt-mcp</h1>

<p align="center">
  <a href="https://www.npmjs.com/package/lynxprompt-mcp"><img src="https://img.shields.io/npm/v/lynxprompt-mcp?style=flat-square&logo=npm" alt="npm"/></a>
  <a href="https://github.com/GeiserX/lynxprompt-mcp/actions/workflows/ci.yml"><img src="https://github.com/GeiserX/lynxprompt-mcp/actions/workflows/ci.yml/badge.svg" alt="CI"/></a>
  <a href="https://codecov.io/gh/GeiserX/lynxprompt-mcp"><img src="https://codecov.io/gh/GeiserX/lynxprompt-mcp/graph/badge.svg" alt="codecov"/></a>
  <a href="https://hub.docker.com/r/drumsergio/lynxprompt-mcp"><img src="https://img.shields.io/docker/pulls/drumsergio/lynxprompt-mcp?style=flat-square&logo=docker" alt="Docker Pulls"/></a>
  <a href="https://github.com/GeiserX/lynxprompt-mcp/blob/main/LICENSE"><img src="https://img.shields.io/github/license/GeiserX/lynxprompt-mcp?style=flat-square" alt="License"/></a>
</p>

<p align="center"><strong>An MCP server for any LynxPrompt instance. LLMs use it to browse, search and manage AI configuration blueprints.</strong></p>

It works with [lynxprompt.com](https://lynxprompt.com) or your own self-hosted LynxPrompt, over HTTP or stdio.

## Features

- Tools to search the public marketplace, list and read your own blueprints and hierarchies, and create, update and delete them. Every tool carries a title, read-only/destructive hints and an output schema.
- `update_blueprint` takes the `content_checksum` from `get_blueprint` as `expected_checksum`, so it never silently overwrites an edit made elsewhere.
- `LYNXPROMPT_READ_ONLY=true` registers only the read tools.
- Read-only resources too: `lynxprompt://blueprints`, `lynxprompt://blueprint/{id}`, `lynxprompt://hierarchies`, `lynxprompt://hierarchy/{id}`, `lynxprompt://user`.
- Speaks MCP 2026-07-28 and every earlier revision down to 2024-11-05 from the same binary, so current clients (which still use the `initialize` handshake) and 2026-07-28 clients both work.
- stdio, or stateless Streamable HTTP on `/mcp` with Origin checks, a body limit and bearer auth when exposed.
- Ships as an npm package (`npx lynxprompt-mcp`), a Docker image and multi-arch Go binaries.

## Tools

| Tool | What it does | Changes data |
|------|--------------|--------------|
| `search_blueprints` | Search the public marketplace (query, configs/commands, category, tags, sort, page) | no |
| `list_blueprints` | Your blueprints, paged, filter by visibility | no |
| `get_blueprint` | One of your blueprints with content and `content_checksum` | no |
| `list_hierarchies` | Your hierarchies (monorepo groupings), paged | no |
| `get_hierarchy` | One hierarchy with its blueprints, flat and as a tree | no |
| `create_blueprint` | New blueprint; `visibility: PUBLIC` publishes it | yes |
| `update_blueprint` | Change fields; pass `expected_checksum` | yes, overwrites |
| `delete_blueprint` | Permanent delete | yes, destructive |
| `create_hierarchy` | New hierarchy; returns the existing one for the same `repository_root` | yes |
| `delete_hierarchy` | Delete a hierarchy; its blueprints are kept, unlinked | yes, destructive |

## Install

Create an API token on your LynxPrompt instance under Settings, API tokens. For a self-hosted instance also set `LYNXPROMPT_URL`. Add `LYNXPROMPT_READ_ONLY=true` to hide the write tools.

**Claude Code**

```sh
claude mcp add lynxprompt --scope user --env LYNXPROMPT_TOKEN=lp_... -- npx -y lynxprompt-mcp
```

**Claude Desktop** (`claude_desktop_config.json`) and **Cursor** (`~/.cursor/mcp.json`)

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

**VS Code** (`.vscode/mcp.json`, asks for the token once and stores it encrypted)

```json
{
  "inputs": [
    { "type": "promptString", "id": "lynxprompt-token", "description": "LynxPrompt API token", "password": true }
  ],
  "servers": {
    "lynxprompt": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "lynxprompt-mcp"],
      "env": { "LYNXPROMPT_TOKEN": "${input:lynxprompt-token}" }
    }
  }
}
```

**Codex** (`~/.codex/config.toml`)

```toml
[mcp_servers.lynxprompt]
command = "npx"
args = ["-y", "lynxprompt-mcp"]
env = { LYNXPROMPT_TOKEN = "lp_..." }
```

**Docker** (stdio, any client): command `docker`, args `["run", "-i", "--rm", "-e", "TRANSPORT=stdio", "-e", "LYNXPROMPT_TOKEN", "drumsergio/lynxprompt-mcp:v0.2.0"]`.

The npm package downloads the release binary on first run and checks its sha256 against the checksums shipped inside the package. The HTTP server, Docker Compose and local builds are in [Getting started](https://github.com/GeiserX/lynxprompt-mcp/blob/main/docs/getting-started.md).

## Documentation

- [Getting started](https://github.com/GeiserX/lynxprompt-mcp/blob/main/docs/getting-started.md): Docker Compose, npm, local build
- [Configuration](https://github.com/GeiserX/lynxprompt-mcp/blob/main/docs/configuration.md): environment variables and MCP client configuration
- [Usage](https://github.com/GeiserX/lynxprompt-mcp/blob/main/docs/usage.md): resources, tools and protocol versions
- [Development](https://github.com/GeiserX/lynxprompt-mcp/blob/main/docs/development.md): testing, contributing, credits
- [Related projects](https://github.com/GeiserX/lynxprompt-mcp/blob/main/docs/related.md): the family and registry listings

## Related projects

[LynxPrompt](https://github.com/GeiserX/LynxPrompt), the self-hosted platform this server talks to.

## License

[GPL-3.0-or-later](https://github.com/GeiserX/lynxprompt-mcp/blob/main/LICENSE)
