DOCKER_GOOS			?= $(GOOS)
DOCKER_GOARCH		?= $(GOARCH)
DOCKER_GO111MODULE	:= on
DOCKER_CGO_ENABLED	?= 0

PROJECT_NAME		:= gscan
DOCKER_IMAGE_NAME	:= gscan
GOLINT				:= $(BIN)/golint

ifeq ($(DOCKER_GOOS),)
	DOCKER_GOOS := linux
endif

ifeq ($(DOCKER_GOARCH),)
	DOCKER_GOARCH := amd64
endif

ifeq ($(DOCKER_GO111MODULE),)
	DOCKER_GO111MODULE := on
endif

DOCKERGOCOMMAND := @docker run --gpus all -v $(CURDIR):/go/src/github.com/h8ck3r/$(PROJECT_NAME) --userns host -w /go/src/github.com/h8ck3r/$(PROJECT_NAME) -e GOOS=$(DOCKER_GOOS) -e GOARCH=$(DOCKER_GOARCH) -e GO111MODULE=$(DOCKER_GO111MODULE) -e CGO_ENABLED=$(DOCKER_CGO_ENABLED) --rm -it $(DOCKER_IMAGE_NAME) go

PREFIX = $(CURDIR)
BIN = $(PREFIX)/bin
CONFDIR = $(PREFIX)/etc/$(PROJECT_NAME)
LIBDIR = $(PREFIX)/lib/$(PROJECT_NAME)

OUT = $(PROJECT_NAME)
SRC = $(PROJECT_NAME).go

.PHONY: all
all: build ## run target build

.PHONY: dockerimage
dockerimage: ## build the golang docker image for this project
	@docker build -t $(DOCKER_IMAGE_NAME) .

$(BIN):
	@mkdir -p "$@"

$(BIN)/%:

$(CONFDIR):
	@mkdir -p "$@"

$(CONFDIR)/%:
	@touch $(CONFDIR)/test.txt

.PHONY: config
config: $(CONFDIR)/$(PROJECT_NAME) ## create the configuration directory

.PHONY: build
build: dockerimage ## compile the binary
	@$(DOCKERGOCOMMAND) build -x -v -o $(OUT) $(SRC)

.PHONY: test
test: dockerimage ## run all tests
	@docker run --gpus all -v $(CURDIR):/go/src/github.com/h8ck3r/$(PROJECT_NAME) --userns host -w /go/src/github.com/h8ck3r/$(PROJECT_NAME) -e GOOS=$(DOCKER_GOOS) -e GOARCH=$(DOCKER_GOARCH) -e GO111MODULE=$(DOCKER_GO111MODULE) -e CGO_ENABLED=1 --rm -it $(DOCKER_IMAGE_NAME) go test -race -v ./... | sed -e '/PASS/ s//$(shell printf "\033[32mPASS\033[0m")/' -e '/FAIL/ s//$(shell printf "\033[31mFAIL\033[0m")/' -e '/SKIP/ s//$(shell printf "\033[93mSKIP\033[0m")/'

.PHONY: bench
bench: dockerimage ## run all benchmarks
	@$(DOCKERGOCOMMAND) test -bench ./...

.PHONY: fmt
fmt: ## format all go files
	@go fmt ./...

.PHONY: lint
lint: | $(BIN)/golint ## run golint
	@$(GOLINT) -set_exit_status ./...

.PHONY: clean
clean: ## cleanup the build and docker cache
	@docker rmi $(DOCKER_IMAGE_NAME)
	@rm -rvf $(CURDIR)/bin
	@go clean -x -v
	@go clean -x -v -cache
	@go clean -x -v -testcache

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<configurations> <target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@awk 'BEGIN {FS = "[ :?=##]"; printf "\nConfigurations:\n"}; length($0)>1 && /^[A-Z]/ {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)