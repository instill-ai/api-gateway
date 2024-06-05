.DEFAULT_GOAL:=help

#============================================================================

# load environment variables for local development
include .env
export

#============================================================================

.PHONY: dev
dev:							## Run dev container
	@docker inspect --type container ${SERVICE_NAME} >/dev/null 2>&1 && echo "A container named ${SERVICE_NAME} is already running." || \
		echo "Run latest container ${SERVICE_NAME}. To stop it, run \"make stop\"."
	@docker run -d --rm \
		-v $(PWD):/${SERVICE_NAME} \
		-p ${SERVICE_PORT}:${SERVICE_PORT} \
		-p ${STATES_PORT}:${STATES_PORT} \
		-p ${METRICS_PORT}:${METRICS_PORT} \
		--network instill-network \
		--name ${SERVICE_NAME} \
		instill/${SERVICE_NAME}:dev >/dev/null 2>&1

.PHONY: logs
logs:							## Tail container logs with -n 10
	@docker logs ${SERVICE_NAME} --follow --tail=10

.PHONY: stop
stop:							## Stop container
	@docker stop -t 1 ${SERVICE_NAME}

.PHONY: rm
rm:								## Remove container
	@docker rm -f ${SERVICE_NAME}

.PHONY: top
top:							## Display all running service processes
	@docker top ${SERVICE_NAME}

.PHONY: build
build:							## Build dev docker image
	@docker build \
		--build-arg SERVICE_NAME=${SERVICE_NAME} \
		--build-arg GOLANG_VERSION=${GOLANG_VERSION} \
		--build-arg ALPINE_VERSION=${ALPINE_VERSION} \
		--build-arg KRAKEND_CE_VERSION=${KRAKEND_CE_VERSION} \
		-f Dockerfile.dev -t instill/${SERVICE_NAME}:dev .

.PHONY: plugin
plugin:							# Compile the KrakenD plugins and copy them to /usr/local/lib/krakend/plugins
	@bash -c "cd plugins/grpc-proxy && go build -buildmode=plugin -buildvcs=false -o /usr/local/lib/krakend/plugins/grpc-proxy.so ./..."
	@bash -c "cd plugins/multi-auth && go build -buildmode=plugin -buildvcs=false -o /usr/local/lib/krakend/plugins/multi-auth.so ./..."
	@bash -c "cd plugins/registry && go build -buildmode=plugin -buildvcs=false -o /usr/local/lib/krakend/plugins/registry.so ./..."

.PHONY: config
config:							## Output the composed KrakenD configuration
	@bash config/envsubst.sh
	@bash scripts/generate_jwks.sh
	@FC_ENABLE=1 \
		FC_SETTINGS="config/settings" \
		FC_PARTIALS="config/share/partials" \
		FC_TEMPLATES="config/share/templates" \
		FC_OUT="config/out.json" \
		krakend check -c config/base.json
	@jq . config/out.json > krakend.json
	@rm config/out.json && rm -rf config/settings

.PHONY: run
run:
	@krakend run -c krakend.json

help:       					## Show this help.
	@echo "\nMake Application using Docker-Compose files."
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m (default: help)\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
