.DEFAULT_GOAL:=help

#============================================================================

# load environment variables for local development
include .env
export

#============================================================================

.PHONY: dev
dev:							## Run dev container
	@docker compose ls -q | grep -q "instill-vdp" && true || \
		(echo "Error: Run \"make dev PROFILE=api-gateway\" in vdp repository (https://github.com/instill-ai/vdp) in your local machine first." && exit 1)
	@docker inspect --type container ${SERVICE_NAME} >/dev/null 2>&1 && echo "A container named ${SERVICE_NAME} is already running." || \
		echo "Run dev container ${SERVICE_NAME}. To stop it, run \"make stop\"."
	@docker run -d --rm \
		-v $(PWD)/plugin:/${SERVICE_NAME}/plugin \
		-v $(PWD)/config:/${SERVICE_NAME}/config \
		-v ${PWD}/cert:/${SERVICE_NAME}/cert \
		-v ${PWD}/Makefile:/${SERVICE_NAME}/Makefile \
		-v ${PWD}/.env:/${SERVICE_NAME}/.env \
		-v ${PWD}/cert/rootCA.pem:/etc/ssl/cert/rootCA.pem \
		-p ${SERVICE_PORT}:${SERVICE_PORT} \
		--network instill-network \
		--name ${SERVICE_NAME} \
		instill/${SERVICE_NAME}:dev

.PHONY: cert
cert:							## Run mkcert to (re-)generate TLS files
	@rm -rf cert && mkdir cert
	@mkcert -client -key-file cert/dev-key.pem -cert-file cert/dev-cert.pem localhost api-gateway
	@cp "$(shell mkcert -CAROOT)"/rootCA.pem cert/

.PHONY: logs
logs:							## Tail container logs with -n 10
	@docker logs ${SERVICE_NAME} --follow --tail=10

.PHONY: stop
stop:							## Stop container
	@docker stop -t 1 ${SERVICE_NAME}

.PHONY: rm
rm:							## Remove container
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

.PHONY: config
config:				## Output the composed KrakenD configuration
	@bash config/envsubst.sh
	@FC_ENABLE=1 \
		FC_SETTINGS="config/settings" \
		FC_PARTIALS="config/partials" \
		FC_TEMPLATES="config/templates" \
		FC_OUT="config/out.json" \
		krakend check -c config/base.json
	@jq . config/out.json > krakend.json
	@rm config/out.json && rm -rf config/settings

help:       		## Show this help.
	@echo "\nMake Application using Docker-Compose files."
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m (default: help)\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
