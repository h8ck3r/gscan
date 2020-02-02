DOCKER_GOOS ?= $(GOOS)
DOCKER_GOARCH ?= $(GOARCH)
DOCKER_GO111MODULE := on
DOCKER_CGO_ENABLED ?= 0

PROJECT_NAME := gscan
DOCKER_IMAGE_NAME := gscan

ifeq ($(DOCKER_GOOS),)
	DOCKER_GOOS := linux
endif

ifeq ($(DOCKER_GOARCH),)
	DOCKER_GOARCH := amd64
endif

ifeq ($(DOCKER_GO111MODULE),)
	DOCKER_GO111MODULE := on
endif

DOCKERGOCOMMAND=@docker run --gpus all -v $(CURDIR):/go/src/github.com/h8ck3r/$(PROJECT_NAME) --userns host -w /go/src/github.com/h8ck3r/$(PROJECT_NAME) -e GOOS=$(DOCKER_GOOS) -e GOARCH=$(DOCKER_GOARCH) -e GO111MODULE=$(DOCKER_GO111MODULE) -e CGO_ENABLED=$(DOCKER_CGO_ENABLED) --rm -it $(DOCKER_IMAGE_NAME) go

PREFIX = $(CURDIR)
BIN = $(PREFIX)/bin
CONFDIR = $(PREFIX)/etc/$(PROJECT_NAME)
LIBDIR = $(PREFIX)/lib/$(PROJECT_NAME)

OUT = $(PROJECT_NAME)
SRC = $(PROJECT_NAME).go

all: build

dockerimage:
	@docker build -t $(DOCKER_IMAGE_NAME) .

$(BIN):
	@mkdir -p "$@"

$(BIN)/%:

$(CONFDIR):
	@mkdir -p "$@"

$(CONFDIR)/%:
	@touch $(CONFDIR)/test.txt

config: $(CONFDIR)/$(PROJECT_NAME)

build: dockerimage
	@$(DOCKERGOCOMMAND) build -x -v -o $(OUT) $(SRC)

test: dockerimage
	@docker run --gpus all -v $(CURDIR):/go/src/github.com/h8ck3r/$(PROJECT_NAME) --userns host -w /go/src/github.com/h8ck3r/$(PROJECT_NAME) -e GOOS=$(DOCKER_GOOS) -e GOARCH=$(DOCKER_GOARCH) -e GO111MODULE=$(DOCKER_GO111MODULE) -e CGO_ENABLED=1 --rm -it $(DOCKER_IMAGE_NAME) go test -race -v ./... | sed -e '/PASS/ s//$(shell printf "\033[32mPASS\033[0m")/' -e '/FAIL/ s//$(shell printf "\033[31mFAIL\033[0m")/' -e '/SKIP/ s//$(shell printf "\033[93mSKIP\033[0m")/'

bench: dockerimage
	@$(DOCKERGOCOMMAND) test -bench ./...

clean:
	@docker rmi $(DOCKER_IMAGE_NAME)
	@rm -rvf $(CURDIR)/bin
	@go clean -x -v
	@go clean -x -v -cache
	@go clean -x -v -testcache
