ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

RUN apt update && apt install -y ca-certificates

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -v -o /app ./discordbot/


FROM debian:bookworm

ENV TZ=America/Sao_Paulo

EXPOSE 8080

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app /usr/local/bin/
CMD ["app"]
