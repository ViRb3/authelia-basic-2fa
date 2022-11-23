FROM golang:1.19.2-alpine AS builder

WORKDIR /src
COPY . .

RUN apk add --no-cache git && \
    go mod download && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "authelia-basic-2fa"

FROM alpine:3.17.0

WORKDIR /

COPY --from=builder "/src/authelia-basic-2fa" "/"

ENTRYPOINT ["/authelia-basic-2fa"]
EXPOSE 8081
