VERSION := $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)
PROJECTNAME := $(shell basename "$(PWD)")

# Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

GOCMD=go
GOBUILD=$(GOCMD) build -mod=vendor
GOCLEAN=$(GOCMD) clean
GOVET=$(GOCMD) vet
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOVEND=$(GOCMD) mod vendor
GO111MODULE=on

export GO111MODULE

all: build
build: govet gotest build-linux build-darwin build-windows

clean:
	$(GOCLEAN) ./...
	rm -rf bin/*

get:
	$(GOGET) ./...

vend:
	$(GOVEND)

gotest:
	$(GOTEST) ./...

govet:
	$(GOVET) ./...

verbosetest:
	$(GOTEST) -v ./...
	$(GOTEST) -cover ./...

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -v $(LDFLAGS) -o "bin/$(PROJECTNAME)-$(VERSION)-linux-amd64" main.go
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -v $(LDFLAGS) -o "bin/$(PROJECTNAME)-$(VERSION)-darwin-amd64" main.go
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -v $(LDFLAGS) -o "bin/$(PROJECTNAME)-$(VERSION)-windows-amd64" main.go
