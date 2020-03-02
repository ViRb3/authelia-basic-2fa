FROM alpine:latest
LABEL Name=local/authelia-basic-2fa

ARG NAMESPACE=ViRb3
ARG PROJECT=authelia-basic-2fa
ARG VERSION=<put_version_here>
ARG ARCH=arm64

RUN apk --no-cache add wget

WORKDIR /

RUN wget -q -O "/start" "https://github.com/${NAMESPACE}/${PROJECT}/releases/download/${VERSION}/${PROJECT}-${ARCH}" && \
    chmod +x "/start"

ENTRYPOINT ["/start"] 
EXPOSE 8081
