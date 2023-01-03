ARG GOLANG_VERSION
ARG ALPINE_VERSION
ARG KRAKEND_CE_VERSION
FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} AS build

ARG SERVICE_NAME

RUN apk --no-cache --virtual .build-deps add tar make gcc musl-dev binutils-gold

WORKDIR /${SERVICE_NAME}

COPY plugin plugin

ARG TARGETARCH
ARG BUILDARCH
RUN if [ "$BUILDARCH" = "amd64" && "$TARGETARCH" = "arm64" ] ; \
    then \
    cd / && wget http://musl.cc/aarch64-linux-musl-cross.tgz && \
    tar zxf aarch64-linux-musl-cross.tgz && rm -f aarch64-linux-musl-cross.tgz && \
    export PATH="$PATH:/aarch64-linux-musl-cross/bin" && \
    cd /${SERVICE_NAME}/plugin && \
    ARCH=$TARGETARCH GOARCH=$TARGETARCH GOHOSTARCH=$BUILDARCH \
    CC=aarch64-linux-musl-gcc EXTRA_LDFLAGS='-extld=aarch64-linux-musl-gcc' \
    go build -buildmode=plugin -o grpc-proxy.so /api-gateway/plugin/server/grpc; \
    else \
    cd plugin && \
    go build -buildmode=plugin -o grpc-proxy.so /api-gateway/plugin/server/grpc; fi

FROM --platform=$BUILDPLATFORM devopsfaith/krakend:${KRAKEND_CE_VERSION}

ARG SERVICE_NAME

RUN apk update && apk add make bash gettext jq

WORKDIR /${SERVICE_NAME}

COPY . .

COPY --from=build /${SERVICE_NAME}/plugin/grpc-proxy.so /${SERVICE_NAME}/plugin/grpc-proxy.so
