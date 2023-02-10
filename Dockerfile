ARG GOLANG_VERSION
ARG ALPINE_VERSION
ARG KRAKEND_CE_VERSION
FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} AS build

ARG SERVICE_NAME

RUN apk --no-cache --virtual .build-deps add tar make gcc musl-dev binutils-gold build-base curl

WORKDIR /${SERVICE_NAME}

COPY plugin plugin

ARG TARGETARCH
ARG BUILDARCH
RUN if [[ "$BUILDARCH" = "amd64" && "$TARGETARCH" = "arm64" ]] ; \
    then \
    curl -sL http://musl.cc/aarch64-linux-musl-cross.tgz | \
    tar zx && \
    export PATH="$PATH:/${SERVICE_NAME}/aarch64-linux-musl-cross/bin" && \
    cd plugin && go mod download && \
    CGO_ENABLED=1 ARCH=$TARGETARCH GOARCH=$TARGETARCH GOHOSTARCH=$BUILDARCH \
    CC=aarch64-linux-musl-gcc EXTRA_LDFLAGS='-extld=aarch64-linux-musl-gcc' \
    go build -buildmode=plugin -o grpc-proxy.so ./server/grpc; \
    else \
    cd plugin && go mod download && \
    CGO_ENABLED=1 go build -buildmode=plugin -o grpc-proxy.so ./server/grpc; fi

FROM devopsfaith/krakend:${KRAKEND_CE_VERSION}

RUN apk update && apk add make bash gettext jq curl

ARG SERVICE_NAME

WORKDIR /${SERVICE_NAME}

COPY --from=build --chown=krakend:nogroup /${SERVICE_NAME}/plugin/grpc-proxy.so /${SERVICE_NAME}/plugin/grpc-proxy.so
COPY .env .env
COPY Makefile Makefile
COPY config config

RUN chown krakend:nogroup -R .

USER krakend
