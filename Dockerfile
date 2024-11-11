ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

RUN apt update && apt install -y ca-certificates

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /app ./discordbot/


FROM debian:bookworm

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app /usr/local/bin/
CMD ["app"]
