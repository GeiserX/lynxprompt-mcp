# Getting started

## npm (stdio transport)

```sh
npx -y lynxprompt-mcp
```

On first run it downloads the Go binary for your platform from the matching GitHub release and checks its sha256 against the `checksums.txt` shipped inside the npm package. npm 12 skips install scripts by default, so the download happens on first run rather than at install time.

## Docker, stdio

MCP clients start the image with `docker run -i`:

```sh
docker run -i --rm -e TRANSPORT=stdio -e LYNXPROMPT_TOKEN=lp_xxx drumsergio/lynxprompt-mcp:v0.2.0
```

## Docker Compose, HTTP

```yaml
services:
  lynxprompt-mcp:
    image: drumsergio/lynxprompt-mcp:v0.2.0
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      - TRANSPORT=http
      - LISTEN_ADDR=0.0.0.0:8080
      - MCP_AUTH_TOKEN=${MCP_AUTH_TOKEN}   # 32+ characters, e.g. openssl rand -hex 32
      - LYNXPROMPT_URL=https://lynxprompt.geiser.cloud
      - LYNXPROMPT_TOKEN=${LYNXPROMPT_TOKEN}
```

Inside the container the server listens on every interface, so it refuses to start without `MCP_AUTH_TOKEN`. The compose file above still publishes the port on the host's loopback only; publish it more widely only behind TLS.

## Local build

```sh
git clone https://github.com/GeiserX/lynxprompt-mcp
cd lynxprompt-mcp

# (optional) create .env from the sample
cp .env.example .env && $EDITOR .env

go run ./cmd/server
```

See [Configuration](configuration.md) for the environment variables.
