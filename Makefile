# path to actual config - the one that is copied to the docker container
CONFIG_PATH:=resources/config/config.yaml

# path to docker compose file
DCOMPOSE:=docker-compose.yaml

NETWORK:=api

# path to external config which will copied to CONFIG_PATH
CONFIG_SOURCE_PATH=resources/config/config_default.yaml

# improve build time
DOCKER_BUILD_KIT:=COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1
DCOMPOSE_BUILD_ARGS:=--build-arg CONFIG_PATH=${CONFIG_PATH}


-include .env

DB_PORT ?= 5432
LOCAL_PG_DSN=postgresql://postgres:postgres@localhost:${DB_PORT}/diplom?sslmode=disable

all: down build up

c: down clean build up

rebuild-backend:
	cp ${CONFIG_SOURCE_PATH} ${CONFIG_PATH}
	docker-compose -f ${DCOMPOSE} stop backend
	${DOCKER_BUILD_KIT} docker-compose build ${DCOMPOSE_BUILD_ARGS} backend
	docker-compose -f ${DCOMPOSE} up backend -d

down:
	docker-compose -f ${DCOMPOSE} down --remove-orphans

build: network
	cp ${CONFIG_SOURCE_PATH} ${CONFIG_PATH}
	${DOCKER_BUILD_KIT} docker-compose build ${DCOMPOSE_BUILD_ARGS}

up:
	docker-compose -f ${DCOMPOSE} up --remove-orphans -d
	go run ./cmd

# Vendoring is useful for local debugging since you don't have to
# reinstall all packages again and again in docker
mod:
	go mod tidy && go mod vendor && go install ./...

clean:
	rm -rf postgres-data

network:
	docker network inspect ${NETWORK} >/dev/null 2>&1 || docker network create --driver bridge ${NETWORK}