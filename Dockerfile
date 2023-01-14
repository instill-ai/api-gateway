ARG GOLANG_VERSION
ARG ALPINE_VERSION
ARG KRAKEND_CE_VERSION
FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} AS build

ARG SERVICE_NAME

RUN apk --no-cache --virtual .build-deps add tar make gcc musl-dev binutils-gold curl

WORKDIR /${SERVICE_NAME}

COPY plugin plugin

ARG TARGETARCH
ARG BUILDARCH
RUN if [[ "$BUILDARCH" = "amd64" && "$TARGETARCH" = "arm64" ]] ; \
    then \
    curl -sL http://musl.cc/aarch64-linux-musl-cross.tgz | \
    tar zxv && \
    export PATH="$PATH:/${SERVICE_NAME}/aarch64-linux-musl-cross/bin" && \
    cd plugin && \
    ARCH=$TARGETARCH GOARCH=$TARGETARCH GOHOSTARCH=$BUILDARCH \
    CC=aarch64-linux-musl-gcc EXTRA_LDFLAGS='-extld=aarch64-linux-musl-gcc' \
    go build -buildmode=plugin -o grpc-proxy.so ./server/grpc; \
    else \
    cd plugin && \
    go build -buildmode=plugin -o grpc-proxy.so ./server/grpc; fi

FROM devopsfaith/krakend:${KRAKEND_CE_VERSION}

ARG SERVICE_NAME

RUN apk update && apk add make bash gettext jq curl

WORKDIR /${SERVICE_NAME}

COPY . .

COPY --from=build /${SERVICE_NAME}/plugin/grpc-proxy.so /${SERVICE_NAME}/plugin/grpc-proxy.so

ARG TARGETARCH
RUN curl -sJLO "https://dl.filippo.io/mkcert/latest?for=linux/$TARGETARCH" && \
    chmod +x mkcert-v*-linux-$TARGETARCH && \
    cp mkcert-v*-linux-$TARGETARCH /usr/local/bin/mkcert
