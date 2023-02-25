SHELL := /bin/bash

ifndef VERBOSE
.SILENT:
endif

# version info
VERSION := $(shell cat VERSION)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_USER := $(USER)@$(HOSTNAME)
BUILD_DATE := $(shell date +"%FT%T")

# go command flags
export GO111MODULE=on
GOFLAGS := -v

# linker flags
LDFLAGS += -X main.version=$(VERSION)
LDFLAGS += -X main.revision=$(GIT_COMMIT)
LDFLAGS += -X main.branch=$(GIT_BRANCH)
LDFLAGS += -X main.buildUser=$(BUILD_USER)
LDFLAGS += -X main.buildDate=$(BUILD_DATE)

SRC_PACKAGES := $(shell go list ./...)

all: test build

build:
	echo ">> $@"
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" .

test:
	echo ">> $@"
	go test -race $(GOFLAGS) ./...

vet:
	echo ">> $@"
	go vet $(SRC_PACKAGES)

staticcheck:
	echo ">> $@"
	staticcheck $(SRC_PACKAGES)

ci: vet staticcheck test

prereq:
	go install honnef.co/go/tools/cmd/staticcheck@2023.1.2

clean:
	rm -f $(BINARY)
