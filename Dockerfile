ARG GOLANG_VERSION
ARG ALPINE_VERSION
ARG KRAKEND_CE_VERSION

FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} as krakend_builder

RUN apk --no-cache --virtual .build-deps add git make gcc musl-dev binutils-gold

ARG KRAKEND_CE_VERSION
RUN git clone -b v${KRAKEND_CE_VERSION} https://github.com/krakendio/krakend-ce.git /krakend && cd /krakend && make build && cp krakend /usr/bin


FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} AS build

ARG SERVICE_NAME

RUN apk --no-cache --virtual .build-deps add tar make gcc musl-dev binutils-gold build-base curl git bash

WORKDIR /${SERVICE_NAME}

COPY plugins/grpc-proxy plugins/grpc-proxy
COPY plugins/multi-auth plugins/multi-auth
COPY plugins/registry plugins/registry
COPY plugins/sse-streaming plugins/sse-streaming

ARG TARGETARCH
ARG BUILDARCH
RUN if [[ "$BUILDARCH" = "amd64" && "$TARGETARCH" = "arm64" ]] ; \
    then \
    curl -sL http://musl.cc/aarch64-linux-musl-cross.tgz | \
    tar zx && \
    export PATH="$PATH:/${SERVICE_NAME}/aarch64-linux-musl-cross/bin" && \
    cd /${SERVICE_NAME}/plugins/grpc-proxy && go mod download && \
    CGO_ENABLED=1 ARCH=$TARGETARCH GOARCH=$TARGETARCH GOHOSTARCH=$BUILDARCH \
    CC=aarch64-linux-musl-gcc EXTRA_LDFLAGS='-extld=aarch64-linux-musl-gcc' \
    go build -buildmode=plugin -buildvcs=false -o grpc-proxy.so ./ ; \
    else \
    cd /${SERVICE_NAME}/plugins/grpc-proxy && go mod download && \
    CGO_ENABLED=1 go build -buildmode=plugin -buildvcs=false -o grpc-proxy.so ./ ; fi

RUN if [[ "$BUILDARCH" = "amd64" && "$TARGETARCH" = "arm64" ]] ; \
    then \
    curl -sL http://musl.cc/aarch64-linux-musl-cross.tgz | \
    tar zx && \
    export PATH="$PATH:/${SERVICE_NAME}/aarch64-linux-musl-cross/bin" && \
    cd /${SERVICE_NAME}/plugins/multi-auth && go mod download && \
    CGO_ENABLED=1 ARCH=$TARGETARCH GOARCH=$TARGETARCH GOHOSTARCH=$BUILDARCH \
    CC=aarch64-linux-musl-gcc EXTRA_LDFLAGS='-extld=aarch64-linux-musl-gcc' \
    go build -buildmode=plugin -buildvcs=false -o multi-auth.so ./; \
    else \
    cd /${SERVICE_NAME}/plugins/multi-auth && go mod download && \
    CGO_ENABLED=1 go build -buildmode=plugin -buildvcs=false -o multi-auth.so ./ ; fi

RUN if [[ "$BUILDARCH" = "amd64" && "$TARGETARCH" = "arm64" ]] ; \
    then \
    curl -sL http://musl.cc/aarch64-linux-musl-cross.tgz | \
    tar zx && \
    export PATH="$PATH:/${SERVICE_NAME}/aarch64-linux-musl-cross/bin" && \
    cd /${SERVICE_NAME}/plugins/registry && go mod download && \
    CGO_ENABLED=1 ARCH=$TARGETARCH GOARCH=$TARGETARCH GOHOSTARCH=$BUILDARCH \
    CC=aarch64-linux-musl-gcc EXTRA_LDFLAGS='-extld=aarch64-linux-musl-gcc' \
    go build -buildmode=plugin -buildvcs=false -o registry.so ./ ; \
    else \
    cd /${SERVICE_NAME}/plugins/registry && go mod download && \
    CGO_ENABLED=1 go build -buildmode=plugin -buildvcs=false -o registry.so ./ ; fi

RUN if [[ "$BUILDARCH" = "amd64" && "$TARGETARCH" = "arm64" ]] ; \
    then \
    curl -sL http://musl.cc/aarch64-linux-musl-cross.tgz | \
    tar zx && \
    export PATH="$PATH:/${SERVICE_NAME}/aarch64-linux-musl-cross/bin" && \
    cd /${SERVICE_NAME}/plugins/sse-streaming && go mod download && \
    CGO_ENABLED=1 ARCH=$TARGETARCH GOARCH=$TARGETARCH GOHOSTARCH=$BUILDARCH \
    CC=aarch64-linux-musl-gcc EXTRA_LDFLAGS='-extld=aarch64-linux-musl-gcc' \
    go build -buildmode=plugin -buildvcs=false -o sse-streaming.so ./ ; \
    else \
    cd /${SERVICE_NAME}/plugins/sse-streaming && go mod download && \
    CGO_ENABLED=1 go build -buildmode=plugin -buildvcs=false -o sse-streaming.so ./ ; fi

RUN cd /${SERVICE_NAME} && \
    git clone -b v2.0.12 https://github.com/lestrrat-go/jwx.git && \
    if [[ "$BUILDARCH" = "amd64" && "$TARGETARCH" = "arm64" ]] ; \
    then \
    curl -sL http://musl.cc/aarch64-linux-musl-cross.tgz | \
    tar zx && \
    export PATH="$PATH:/${SERVICE_NAME}/aarch64-linux-musl-cross/bin" && \
    cd /${SERVICE_NAME}/jwx/cmd/jwx && go mod download && \
    CGO_ENABLED=1 ARCH=$TARGETARCH GOARCH=$TARGETARCH GOHOSTARCH=$BUILDARCH \
    CC=aarch64-linux-musl-gcc EXTRA_LDFLAGS='-extld=aarch64-linux-musl-gcc' \
    go build -o /go/bin/jwx . ; \
    else \
    cd /${SERVICE_NAME}/jwx/cmd/jwx && go mod download && \
    go build -o /go/bin/jwx . ; \
    fi

FROM alpine:${ALPINE_VERSION}

RUN apk update && apk add make bash gettext jq curl
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -u 1000 -S -D -H krakend && \
    mkdir /etc/krakend && \
    echo '{ "version": 3 }' > /etc/krakend/krakend.json

COPY --from=krakend_builder /usr/bin/krakend /usr/bin/krakend

ARG SERVICE_NAME

WORKDIR /${SERVICE_NAME}

RUN mkdir -p /usr/local/lib/krakend/plugins && chmod 777 /usr/local/lib/krakend/plugins

COPY --from=build --chown=krakend:nogroup /${SERVICE_NAME}/plugins/grpc-proxy/grpc-proxy.so /usr/local/lib/krakend/plugins
COPY --from=build --chown=krakend:nogroup /${SERVICE_NAME}/plugins/multi-auth/multi-auth.so /usr/local/lib/krakend/plugins
COPY --from=build --chown=krakend:nogroup /${SERVICE_NAME}/plugins/registry/registry.so /usr/local/lib/krakend/plugins
COPY --from=build --chown=krakend:nogroup /${SERVICE_NAME}/plugins/sse-streaming/sse-streaming.so /usr/local/lib/krakend/plugins
COPY --from=build --chown=krakend:nogroup /go/bin/jwx /go/bin/jwx
RUN mkdir -p /instill && chmod 777 /instill

COPY .env .env
COPY Makefile Makefile
COPY config config
COPY scripts scripts

RUN chown krakend:nogroup -R .

USER krakend
