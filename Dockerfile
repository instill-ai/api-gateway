FROM --platform=amd64 golang:1.19.1-alpine3.16 AS build

ARG SERVICE_NAME

USER root

RUN apk update && apk add make gcc musl-dev

WORKDIR /${SERVICE_NAME}

COPY plugins plugins

WORKDIR /${SERVICE_NAME}/plugins

RUN go env -w GO111MODULE=on
RUN go build -buildmode=plugin /${SERVICE_NAME}/plugins/handler/modifier.go
RUN go build -buildmode=plugin /${SERVICE_NAME}/plugins/client/grpc/grpc.go

RUN go test -v plugins/handler

FROM --platform=$BUILDPLATFORM instill/krakend:2.1.0

ARG SERVICE_NAME

USER root

RUN apk update && apk add bash gettext jq

WORKDIR /${SERVICE_NAME}

COPY .env .env

COPY config config

RUN bash /${SERVICE_NAME}/config/envsubst.sh && \
    FC_ENABLE=1 \
    FC_SETTINGS="/${SERVICE_NAME}/config/settings" \
    FC_PARTIALS="/${SERVICE_NAME}/config/partials" \
    FC_TEMPLATES="/${SERVICE_NAME}/config/templates" \
    FC_OUT="/${SERVICE_NAME}/config/out.json" \
    krakend check -c /${SERVICE_NAME}/config/base.json

RUN jq . config/out.json > krakend.json
RUN rm config/out.json && rm -rf config/settings

# Copy plugins
COPY --from=build /${SERVICE_NAME}/plugins /${SERVICE_NAME}/plugins
