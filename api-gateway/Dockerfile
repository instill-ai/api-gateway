ARG BASE_KRAKEND_VERSION

#
#
# Build KrakenD plugins
FROM golang:1.17.13-alpine AS build

RUN apk add make gcc musl-dev

COPY api-gateway/plugins /api-gateway/plugins

WORKDIR /api-gateway/plugins
RUN go env -w GO111MODULE=on
RUN go build -buildmode=plugin /api-gateway/plugins/handler/modifier.go
RUN go build -buildmode=plugin /api-gateway/plugins/client/grpc/grpc.go

RUN go test -v /api-gateway/plugins/handler

#
#
# Use KrakenD official image
FROM devopsfaith/krakend:${BASE_KRAKEND_VERSION} AS base

USER root

# Copy system packages
RUN apk update && apk add gettext jq

# Copy plugins
COPY --from=build /api-gateway/plugins /api-gateway/plugins

# Copy KrakenD flexible configuration files
COPY api-gateway/config /api-gateway/config

#
#
# dev image
FROM base AS dev

# Copy environment variables
COPY .env.dev /api-gateway/.env

RUN chown 1000:1000 -R /api-gateway
RUN apk add bash

USER krakend

WORKDIR /api-gateway

RUN bash config/envsubst.sh
RUN FC_ENABLE=1 \
    FC_SETTINGS="config/settings" \
    FC_PARTIALS="config/partials" \
    FC_TEMPLATES="config/templates" \
    FC_OUT="config/out.json" \
    krakend check -c config/base.json

RUN jq . config/out.json > config/krakend.json
RUN rm /api-gateway/.env && rm /api-gateway/config/out.json && rm -rf /api-gateway/config/settings

#
#
# prod image
FROM base AS prod

RUN chown 1000:1000 -R /api-gateway

USER krakend

WORKDIR /api-gateway
