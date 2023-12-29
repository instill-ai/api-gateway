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

COPY grpc_proxy_plugin grpc_proxy_plugin
COPY multi_auth_plugin multi_auth_plugin

ARG TARGETARCH
ARG BUILDARCH
RUN if [[ "$BUILDARCH" = "amd64" && "$TARGETARCH" = "arm64" ]] ; \
    then \
    curl -sL http://musl.cc/aarch64-linux-musl-cross.tgz | \
    tar zx && \
    export PATH="$PATH:/${SERVICE_NAME}/aarch64-linux-musl-cross/bin" && \
    cd /${SERVICE_NAME}/grpc_proxy_plugin && go mod download && \
    CGO_ENABLED=1 ARCH=$TARGETARCH GOARCH=$TARGETARCH GOHOSTARCH=$BUILDARCH \
    CC=aarch64-linux-musl-gcc EXTRA_LDFLAGS='-extld=aarch64-linux-musl-gcc' \
    go build -buildmode=plugin -buildvcs=false -o grpc-proxy.so ./pkg; \
    else \
    cd /${SERVICE_NAME}/grpc_proxy_plugin && go mod download && \
    CGO_ENABLED=1 go build -buildmode=plugin -buildvcs=false -o grpc-proxy.so ./pkg; fi

RUN if [[ "$BUILDARCH" = "amd64" && "$TARGETARCH" = "arm64" ]] ; \
    then \
    curl -sL http://musl.cc/aarch64-linux-musl-cross.tgz | \
    tar zx && \
    export PATH="$PATH:/${SERVICE_NAME}/aarch64-linux-musl-cross/bin" && \
    cd /${SERVICE_NAME}/multi_auth_plugin && go mod download && \
    CGO_ENABLED=1 ARCH=$TARGETARCH GOARCH=$TARGETARCH GOHOSTARCH=$BUILDARCH \
    CC=aarch64-linux-musl-gcc EXTRA_LDFLAGS='-extld=aarch64-linux-musl-gcc' \
    go build -buildmode=plugin -buildvcs=false -o multi-auth.so ./server; \
    else \
    cd /${SERVICE_NAME}/multi_auth_plugin && go mod download && \
    CGO_ENABLED=1 go build -buildmode=plugin -buildvcs=false -o multi-auth.so ./server; fi

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

RUN mkdir -p /usr/local/lib/krakend/plugin && chmod 777 /usr/local/lib/krakend/plugin

COPY --from=build --chown=krakend:nogroup /${SERVICE_NAME}/grpc_proxy_plugin/grpc-proxy.so /usr/local/lib/krakend/plugin
COPY --from=build --chown=krakend:nogroup /${SERVICE_NAME}/multi_auth_plugin/multi-auth.so /usr/local/lib/krakend/plugin
COPY --from=build --chown=krakend:nogroup /go/bin/jwx /go/bin/jwx
RUN mkdir -p /instill && chmod 777 /instill

COPY .env .env
COPY Makefile Makefile
COPY config config
COPY scripts scripts

RUN chown krakend:nogroup -R .

USER krakend
