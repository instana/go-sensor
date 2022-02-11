include instrumentation/instapgx/Makefile

MODULES = $(filter-out $(EXCLUDE_DIRS), $(shell find . -name go.mod -exec dirname {} \;))
LINTER ?= $(shell go env GOPATH)/bin/golangci-lint

# The list of Go build tags as they are specified in respective integration test files
INTEGRATION_TESTS = fargate gcr lambda

# the Go version to vendor dependencies listed in go.mod
VENDOR_GO_VERSION ?= go1.15
VENDOR_GO = $(shell go env GOPATH)/bin/$(VENDOR_GO_VERSION)
MODULES_VENDOR = $(addsuffix /vendor,$(MODULES))

ifeq ($(RUN_LINTER),yes)
test: $(LINTER)
endif

ifeq ($(VENDOR_DEPS),yes)
# We need to vendor all dependencies at once before running test, so the `go get -t -d ./...`
# on go1.9 and go1.10 does not try to re-download them
test: $(MODULES_VENDOR)
endif

test: $(MODULES) legal

$(MODULES):
	cd $@ && go get -d -t ./... && go test $(GOFLAGS) ./...
ifeq ($(RUN_LINTER),yes)
	cd $@ && $(LINTER) run
endif

integration: $(INTEGRATION_TESTS) pgx_integration_test

$(INTEGRATION_TESTS):
	go test $(GOFLAGS) -tags "$@ integration" $(shell grep --exclude-dir=instapgx -lR '^// +build \($@,\)\?integration\(,$@\)\?' .)

$(LINTER):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/a2bc9b7a99e3280805309d71036e8c2106853250/install.sh \
	| sh -s -- -b $(basename $(GOPATH))/bin v1.23.8

$(MODULES_VENDOR): $(VENDOR_GO)
	cd $(shell dirname $@) && $(VENDOR_GO) mod vendor
	find $@ -name go.mod -delete

$(VENDOR_GO):
	go get golang.org/dl/$(VENDOR_GO_VERSION)
	$(VENDOR_GO) download

install:
	cd .git/hooks && ln -fs ../../.githooks/* .
	brew install gh

# Make sure there is a copyright at the first line of each .go file
legal:
	awk 'FNR==1 { if (tolower($$0) !~ "^//.+copyright") { print FILENAME" does not contain copyright header"; rc=1 } }; END { exit rc }' $$(find . -name '*.go' -type f | grep -v "/vendor/")

instrumentation/% :
	mkdir -p $@
	cd $@ && go mod init github.com/instana/go-sensor/$@
	sed "s~Copyright (c) [0-9]*~Copyright (c) $(shell date +%Y)~" LICENSE.md > $@/LICENSE.md
	printf "VERSION_TAG_PREFIX ?= $@/v\nGO_MODULE_NAME ?= github.com/instana/go-sensor/$@\n\ninclude ../../Makefile.release\n" > $@/Makefile
	printf '// (c) Copyright IBM Corp. %s\n// (c) Copyright Instana Inc. %s\n\npackage %s\n\nconst Version = "0.0.0"\n' $(shell date +%Y) $(shell date +%Y) $(notdir $@) > $@/version.go

.PHONY: test vendor install legal $(MODULES) $(INTEGRATION_TESTS)

# Release targets
include Makefile.release
