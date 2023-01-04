MODULES = $(filter-out $(EXCLUDE_DIRS), $(shell find . -name go.mod -exec dirname {} \;))
LINTER ?= $(shell go env GOPATH)/bin/golangci-lint

# The list of Go build tags as they are specified in respective integration test files
INTEGRATION_TESTS = fargate gcr lambda azure

ifeq ($(RUN_LINTER),yes)
test: $(LINTER)
endif

test: $(MODULES) legal

$(MODULES):
	cd $@ && go get -d -t ./... && go test $(GOFLAGS) ./...
ifeq ($(RUN_LINTER),yes)
	cd $@ && $(LINTER) run
endif

integration: $(INTEGRATION_TESTS)
	cd instrumentation/instapgx && go test -tags=integration

$(INTEGRATION_TESTS):
	go test $(GOFLAGS) -tags "$@ integration" $(shell grep --exclude-dir=instapgx -lR '^// +build \($@,\)\?integration\(,$@\)\?' .)

$(LINTER):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/a2bc9b7a99e3280805309d71036e8c2106853250/install.sh \
	| sh -s -- -b $(basename $(GOPATH))/bin v1.23.8

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

fmtcheck:
	@gofmt -l .
	@test -z $(shell gofmt -l . && exit 1)

importcheck:
	@test -z $(shell goimports -l . && exit 1)

.PHONY: test install legal fmtcheck importcheck $(MODULES) $(INTEGRATION_TESTS)

# Release targets
include Makefile.release
