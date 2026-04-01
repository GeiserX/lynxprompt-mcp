FROM golang:1.26 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /out/lynxprompt-mcp ./cmd/server

FROM alpine:3.23
COPY --from=builder /out/lynxprompt-mcp /usr/local/bin/lynxprompt-mcp
EXPOSE 8080
ENV TRANSPORT=stdio
ENV LYNXPROMPT_URL=https://lynxprompt.com
ENV LYNXPROMPT_TOKEN=""
ENTRYPOINT ["/usr/local/bin/lynxprompt-mcp"]
