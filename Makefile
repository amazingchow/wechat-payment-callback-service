include .env.local .env.secret .env.shared
export

VERSION      := v1.0.0
GIT_HASH     := $(shell git rev-parse --short HEAD)
SERVICE      := wechat-payment-callback-service
SRC          := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
TARGETS      := wechat-payment-callback-service
TEST_TARGETS :=
ALL_TARGETS  := $(TARGETS) $(TEST_TARGETS)
CURR_DIR     := $(shell pwd)

.PHONY: help
help: ### Display this help screen.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: init
init: ### Initialize the project.
	@go mod tidy
	@go mod vendor

.PHONY: pb-fmt
pb-fmt: ### Format the proto files using clang-format (sudo apt install clang-format).
	@clang-format -i ./protos/*.proto

ifeq ($(race), 1)
	BUILD_FLAGS := -race
endif

ifeq ($(gc_debug), 1)
	BUILD_FLAGS += -gcflags=all="-N -l"
endif

.PHONY: build
build: clean pb-fmt init $(ALL_TARGETS) ### Build the service.

$(TARGETS): $(SRC)
	@GOOS=linux GOARCH=amd64 go build -mod vendor $(BUILD_FLAGS) $(PWD)/cmd/$@

$(TEST_TARGETS): $(SRC)
	@GOOS=linux GOARCH=amd64 go build -mod vendor $(BUILD_FLAGS) $(PWD)/test/$@

.PHONY: clean
clean: ### Clean the service.
	@rm -f $(ALL_TARGETS)

.PHONY: test
test: ### Run the tests.
	@env CI="true" go test -count=1 -v -p 1 $(shell go list ./... | grep -v /igspb | grep -v /cmd) -coverprofile unit_test_coverage.txt || true

.PHONY: local_run
local_run: build ### Run your service locally.
	@./${SERVICE} -conf ./etc/${SERVICE}-dev.json 2>&1 | tee dev.log

IMAGE_VERSION := ${VERSION}-${GIT_HASH}

.PHONY: image
image: build ### Build your service image.
	@docker build -f ./devops/docker/Dockerfile -t infra-${SERVICE}:${IMAGE_VERSION} .

.PHONY: check_compose
check_compose: ### Check the docker-compose configuration.
	@docker-compose -f "${CURR_DIR}/docker-compose.yml" config

.PHONY: run_compose
run_compose: image check_compose ### Run the application with docker-compose.
	@mkdir -p ~/.infra-config/${SERVICE}
	@cp -f ./etc/${SERVICE}-prod.json ~/.infra-config/${SERVICE}/${SERVICE}.json
	@mkdir -p ${CURR_DIR}/.logs
	@mkdir -p ${CURR_DIR}/.persistent
	@mkdir -p ${CURR_DIR}/.locks
	@mkdir -p ${CURR_DIR}/.shares
	@docker-compose -f "${CURR_DIR}/docker-compose.yml" up -d --build

.PHONY: shutdown_compose
shutdown_compose: ### Shutdown the application with docker-compose.
	@docker-compose -f "${CURR_DIR}/docker-compose.yml" down

now=$(shell date "+%Y%m%d%H%M%S")
.PHONY: logs
logs: ### Show the logs of the running service.
	@docker logs -f infra-${SERVICE} 2>&1 | tee prod_${now}.log
