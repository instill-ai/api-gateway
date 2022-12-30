ARG GOLANG_VERSION
ARG APLINE_VERSION
ARG KRAKEND_VERSION
FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION}-alpine${APLINE_VERSION} AS build

ARG SERVICE_NAME

RUN apk --no-cache --virtual .build-deps add tar make gcc musl-dev binutils-gold

RUN cd / && wget http://musl.cc/aarch64-linux-musl-cross.tgz && \
    tar zxf aarch64-linux-musl-cross.tgz && rm -f aarch64-linux-musl-cross.tgz

WORKDIR /${SERVICE_NAME}

COPY plugin plugin

WORKDIR /${SERVICE_NAME}/plugin

RUN go build -buildmode=plugin -o grpc-proxy.so /api-gateway/plugin/server/grpc

FROM --platform=$BUILDPLATFORM devopsfaith/krakend:${KRAKEND_VERSION}

ARG SERVICE_NAME

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

RUN jq . config/out.json > /etc/krakend/krakend.json
RUN rm config/out.json && rm -rf config/settings

# Copy plugin
COPY --from=build /${SERVICE_NAME}/plugin /${SERVICE_NAME}/plugin

ENTRYPOINT [ "/usr/bin/krakend" ]
CMD [ "run", "-c", "/etc/krakend/krakend.json" ]
