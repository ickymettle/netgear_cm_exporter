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

# cross compile settings
XC_OUTPUT := _builds/{{.Dir}}_v$(VERSION)_{{.OS}}_{{.Arch}}
XC_OS     := linux darwin freebsd
XC_ARCH   := amd64 386
XC_OSARCH := linux/arm linux/mipsle

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

build:
	echo ">> $@"
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" .

test:
	echo ">> $@"
	go test -race $(GOFLAGS) ./...

lint:
	echo ">> $@"
	golint -set_exit_status $(SRC_PACKAGES)

vet:
	echo ">> $@"
	go vet $(SRC_PACKAGES)

staticcheck:
	echo ">> $@"
	# staticcheck doesn't quite support modules yet
	#staticcheck $(SRC_PACKAGES)

ci: lint vet staticcheck test

cross-compile:
	gox -verbose -output $(XC_OUTPUT) -os "$(XC_OS)" -arch "$(XC_ARCH)" -osarch "$(XC_OSARCH)"

prereq:
	go get -u golang.org/x/lint/golint
	go get -u honnef.co/go/tools/cmd/staticcheck
	go get -u github.com/mitchellh/gox
	go get -u github.com/tcnksm/ghr

clean:
	rm -rf netgear_cm_exporter _builds
