.DEFAULT_GOAL:=help

COMPOSE_DEV := -f docker-compose.dev.yml
COMPOSE_MONITORING := -f docker-compose.monitoring.yml
COMPOSE_LOGGING := -f docker-compose.logging.yml
COMPOSE_ALL := ${COMPOSE_DEV} ${COMPOSE_MONITORING} ${COMPOSE_LOGGING}

INSTILL_SERVICES := api_gateway mgmt_backend_migrate mgmt_backend_init mgmt_backend model_backend_migrate model_backend_init model_backend pipeline_backend_migrate pipeline_backend connector_backend_migrate connector_backend_init connector_backend_worker connector_backend
3RD_PARTY_SERVICES := pg_sql hydra hydra_migrate kratos kratos_migrate influxdb cassandra temporal temporal_admin_tools temporal_ui redis
MONITORING_SERVICES := influxdb grafana prometheus alertmanager
LOGGING_SERVICES := elasticsearch kibana filebeat
ALL_SERVICES := ${INSTILL_SERVICES} ${3RD_PARTY_SERVICES} ${MONITORING_SERVICES} ${LOGGING_SERVICES}

#============================================================================

# load environment variables for local development
include .env.dev
export

TRITONSERVER_IMAGE_TAG := $(if $(filter arm64,$(shell uname -m)),instill/tritonserver:${TRITON_SERVER_VERSION}-py3-cpu-arm64,nvcr.io/nvidia/tritonserver:${TRITON_SERVER_VERSION}-py3)
TRITONCONDAENV_IMAGE_TAG := $(if $(filter arm64,$(shell uname -m)),instill/triton-conda-env:${TRITON_CONDA_ENV_VERSION}-m1,instill/triton-conda-env:${TRITON_CONDA_ENV_VERSION}-cpu)
REDIS_IMAGE_TAG := $(if $(filter arm64,$(shell uname -m)),arm64v8/redis:${REDIS_VERSION}-alpine,amd64/redis:${REDIS_VERSION}-alpine)

NVIDIA_SMI := $(shell nvidia-smi 2>/dev/null 1>&2; echo $$?)
ifeq ($(NVIDIA_SMI),0)
	TRITONSERVER_RUNTIME := nvidia
	TRITONCONDAENV_IMAGE_TAG := instill/triton-conda-env:${TRITON_CONDA_ENV_VERSION}-gpu
endif

#============================================================================

keystore:		## Setup Elasticsearch Keystore, by initializing passwords, and add credentials defined in `keystore.sh`.
	docker-compose -f docker-compose.setup.yml run --rm keystore
.PHONY: keystore

certs:		    ## Generate Elasticsearch SSL Certs.
	docker-compose -f docker-compose.setup.yml run --rm certs
.PHONY: certs

setup:		    ## Generate Elasticsearch SSL Certs and Keystore.
	@make certs
	@make keystore
.PHONY: setup

#============================================================================

dev:		    	## Start development stack
	@docker-compose ${COMPOSE_DEV} up -d --build
.PHONY: dev

monitoring:		    ## Start monitoring stack
	@docker-compose ${COMPOSE_MONITORING} up -d --build
.PHONY: monitoring

logging:		    ## Start logging stack
	@docker-compose ${COMPOSE_LOGGING} up -d --build
.PHONY: logging

all:		    	## Start all components including application, monitoring and logging stacks.
	@docker-compose ${COMPOSE_ALL} up -d
.PHONY: all

logs:			## Tail all logs with -n 10.
	@docker-compose $(COMPOSE_ALL) logs --follow --tail=10
.PHONY: logs

stop:			## Stop all components.
	@docker-compose ${COMPOSE_ALL} stop ${ALL_SERVICES}
.PHONY: stop

start:			## Start all stopped components.
	@docker-compose ${COMPOSE_ALL} start ${ALL_SERVICES}
.PHONY: start

restart:		## Restart all components.
	@docker-compose ${COMPOSE_ALL} restart ${ALL_SERVICES}
.PHONY: restart

rm:				## Remove all stopped components containers.
	@docker-compose $(COMPOSE_ALL) rm -f ${ALL_SERVICES}
.PHONY: rm

down:			## Down all components.
	@docker-compose ${COMPOSE_ALL} down
.PHONY: down

images:			## List all images of components.
	@docker-compose $(COMPOSE_ALL) images ${ALL_SERVICES}
.PHONY: images

ps:			## List all component containers.
	@docker-compose $(COMPOSE_ALL) ps ${ALL_SERVICES}
.PHONY: ps

prune:			## Remove all containers and delete volume
	@make stop && make rm
	@docker volume prune -f
.PHONY: prune

#============================================================================

api:				## Run api-gateway container
	@docker run -d --name api-gateway -p 8080:8080 -p 8090:8090 -p 9091:9091 \
		-v ${PWD}/api-gateway/config:/api-gateway/config \
		harbor.instill.tech/api-gateway/api-gateway:dev \
		run --debug --config /api-gateway/config/api-gateway.json
.PHONY: api

api-logs:			## Follow logs for only the api-gateway container
	@docker logs api-gateway --follow
.PHONY: api-logs

api-stop:			## Stop api-gateway container
	@docker stop api-gateway
.PHONY: api-stop

api-start:			## Start stopped api-gateway container
	@docker start api-gateway
.PHONY: api-start

api-restart:		## Restart api-gateway container
	@docker restart api-gateway
.PHONY: api-restart

api-rm:				## Remove api-gateway container
	@docker rm api-gateway
.PHONY: api-rm

api-ps:				## List api-gateway container
	@docker container ps --filter "name=api-gateway"
.PHONY: api-ps

#============================================================================

build-dev:			## Build KrakenD plugins and build a harbor.instill.tech/api-gateway/api-gateway:dev image for local development
	@DOCKER_BUILDKIT=1 \
		docker build --build-arg BASE_KRAKEND_VERSION=${BASE_KRAKEND_VERSION} \
		--target dev \
		-f api-gateway/Dockerfile \
		-t harbor.instill.tech/api-gateway/api-gateway:dev .
.PHONY: build-dev

#============================================================================

config:				## Output the composed KrakenD configuration for debugging
	@cp .env.dev api-gateway/.env
	@bash api-gateway/config/envsubst.sh
	@FC_ENABLE=1 \
		FC_SETTINGS="api-gateway/config/settings" \
		FC_PARTIALS="api-gateway/config/partials" \
		FC_TEMPLATES="api-gateway/config/templates" \
		FC_OUT="api-gateway/config/out.json" \
		krakend check -c api-gateway/config/base.json
	@jq . api-gateway/config/out.json > krakend.json
	@rm api-gateway/.env && rm api-gateway/config/out.json && rm -rf api-gateway/config/settings
.PHONY: config

#============================================================================
test:
	@go version
	@go install go.k6.io/xk6/cmd/xk6@latest
	@xk6 build --with github.com/szkiba/xk6-jose@latest
	# @TEST_FOLDER_ABS_PATH=${PWD}/tests ./k6 run tests/mgmt-backend.js --no-usage-report
	@TEST_FOLDER_ABS_PATH=${PWD}/tests ./k6 run tests/rest-pipeline-backend.js --no-usage-report
	@TEST_FOLDER_ABS_PATH=${PWD}/tests ./k6 run tests/rest-model-backend.js --no-usage-report
	@TEST_FOLDER_ABS_PATH=${PWD}/tests ./k6 run tests/rest-connector-backend.js --no-usage-report
	@rm k6
.PHONY: test

help:       		## Show this help.
	@echo "\nMake Application using Docker-Compose files."
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m (default: help)\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
