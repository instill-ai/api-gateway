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

FROM --platform=$BUILDPLATFORM devopsfaith/krakend:${KRAKEND_CE_VERSION}

ARG SERVICE_NAME

RUN apk update && apk add make bash gettext jq curl

WORKDIR /${SERVICE_NAME}

COPY . .

COPY --from=build /${SERVICE_NAME}/plugin/grpc-proxy.so /${SERVICE_NAME}/plugin/grpc-proxy.so

ARG TARGETARCH
RUN curl -JLO "https://dl.filippo.io/mkcert/latest?for=linux/$TARGETARCH" && \
    chmod +x mkcert-v*-linux-$TARGETARCH && \
    cp mkcert-v*-linux-$TARGETARCH /usr/local/bin/mkcert
