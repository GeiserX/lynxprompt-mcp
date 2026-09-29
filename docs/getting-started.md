# Getting started

## Docker Compose

```yaml
services:
  lynxprompt-mcp:
    image: drumsergio/lynxprompt-mcp:v0.1.0
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      - LYNXPROMPT_URL=https://lynxprompt.com
      - LYNXPROMPT_TOKEN=lp_xxx
```

> **Security note:** The HTTP transport listens on `127.0.0.1:8080` by default. If you need to expose it on a network, place it behind a reverse proxy with authentication.

## npm (stdio transport)

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

See [Configuration](configuration.md) for the environment variables.
