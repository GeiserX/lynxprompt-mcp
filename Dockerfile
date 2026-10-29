# Base images are pinned by digest; Dependabot (docker ecosystem) bumps tag and digest together.
FROM golang:1.27.3@sha256:9048d5e7751706200eb1719c7dae0436950d362f29a1da91fa05e5c63867c60f AS builder
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w -X github.com/geiserx/lynxprompt-mcp/version.Version=${VERSION} -X github.com/geiserx/lynxprompt-mcp/version.Commit=${COMMIT} -X github.com/geiserx/lynxprompt-mcp/version.Date=${DATE}" \
    -o /out/lynxprompt-mcp ./cmd/server

FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a
# The base image can lag the CA bundle; take the current one, which drops
# the compromised eMudhra (emSign) intermediates.
RUN apk upgrade --no-cache ca-certificates-bundle
LABEL io.modelcontextprotocol.server.name="io.github.GeiserX/lynxprompt-mcp"
COPY --from=builder /out/lynxprompt-mcp /usr/local/bin/lynxprompt-mcp
# MCP clients and the MCP Registry start the image as `docker run -i`, so the
# default transport is stdio. For the HTTP server set TRANSPORT=http and an
# MCP_AUTH_TOKEN of 32+ characters: it listens on every interface in here.
ENV TRANSPORT=stdio
ENV LISTEN_ADDR=0.0.0.0:8080
ENV LYNXPROMPT_URL=https://lynxprompt.geiser.cloud
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/lynxprompt-mcp"]
