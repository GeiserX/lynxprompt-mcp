<p align="center">
  <img src="https://raw.githubusercontent.com/GeiserX/lynxprompt-mcp/main/docs/images/banner.svg" alt="LynxPrompt MCP banner" width="900"/>
</p>

<h1 align="center">LynxPrompt-MCP</h1>

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

- Read-only resources for blueprints, hierarchies and the current user (`lynxprompt://blueprints`, `lynxprompt://hierarchies`, `lynxprompt://user`).
- Tools to search, create, update and delete blueprints, and to create and delete hierarchies.
- One JSON-RPC endpoint (`/mcp`) over HTTP, or stdio with `TRANSPORT=stdio`.
- Listens on loopback by default; `MCP_AUTH_TOKEN` adds bearer auth when you expose it.
- Ships as a Docker image, an npm package (`npx lynxprompt-mcp`) and multi-arch Go binaries.

## Quick start

```sh
npx lynxprompt-mcp
```

Set `LYNXPROMPT_TOKEN` (and `LYNXPROMPT_URL` for a self-hosted instance) first. Docker Compose and local builds are in [Installation](https://github.com/GeiserX/lynxprompt-mcp/blob/main/docs/installation.md).

## Documentation

- [Installation](https://github.com/GeiserX/lynxprompt-mcp/blob/main/docs/installation.md): Docker Compose, npm, local build
- [Configuration](https://github.com/GeiserX/lynxprompt-mcp/blob/main/docs/configuration.md): environment variables and an example client config
- [Resources and tools](https://github.com/GeiserX/lynxprompt-mcp/blob/main/docs/usage.md)
- [Development](https://github.com/GeiserX/lynxprompt-mcp/blob/main/docs/development.md): testing, contributing, credits
- [Related projects and listings](https://github.com/GeiserX/lynxprompt-mcp/blob/main/docs/related.md)

Related: [LynxPrompt](https://github.com/GeiserX/LynxPrompt), the self-hosted platform this server talks to.

## License

[GPL-3.0](https://github.com/GeiserX/lynxprompt-mcp/blob/main/LICENSE)
