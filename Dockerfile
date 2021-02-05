FROM golang:1.15.8-alpine AS builder

WORKDIR /src
COPY . .

RUN go mod download && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "bin-release"

FROM alpine:3.11

WORKDIR /

COPY --from=builder "/src/bin-release" "/authelia-basic-2fa"

ENTRYPOINT ["/authelia-basic-2fa"]
EXPOSE 8081
