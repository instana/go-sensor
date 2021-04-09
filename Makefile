MODULES = $(filter-out $(EXCLUDE_DIRS), $(shell find . -name go.mod -exec dirname {} \;))
# print filtered-out modules, for easy debugging
$(info MODULES are: $(MODULES))

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

test: $(MODULES)

$(MODULES):
	go list ./...
	cd $@ && pwd && go get -d -t ./... && go test $(GOFLAGS) ./...
ifeq ($(RUN_LINTER),yes)
	cd $@ && $(LINTER) run
endif

integration: $(INTEGRATION_TESTS)

$(INTEGRATION_TESTS):
	go test $(GOFLAGS) -tags "$@ integration" $(shell grep -lR '^// +build \($@,\)\?integration\(,$@\)\?' .)

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

.PHONY: test vendor $(MODULES) $(INTEGRATION_TESTS)
