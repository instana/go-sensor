
MODULES = $(filter-out $(EXCLUDE_DIRS), $(shell find . -name go.mod -exec dirname {} \;))
LINTER ?= $(shell go env GOPATH)/bin/golangci-lint
INTEGRATION_TESTS = fargate_integration

# the Go version to vendor dependencies listed in go.mod
VENDOR_GO_VERSION ?= go1.15
VENDOR_GO = $(shell go env GOPATH)/bin/$(VENDOR_GO_VERSION)
MODULES_VENDOR = $(addsuffix /vendor,$(MODULES))

ifdef RUN_LINTER
test: $(LINTER)
endif

ifdef VENDOR_DEPS
# We need to vendor all dependencies at once before running test, so the `go get -t -d ./...`
# on go1.9 and go1.10 does not try to re-download them
test: $(MODULES_VENDOR)
endif

test: $(MODULES)

$(MODULES):
	cd $@ && go get -d -t ./... && go test $(GOFLAGS) ./...
ifdef RUN_LINTER
	cd $@ && $(LINTER) run
endif

integration: $(INTEGRATION_TESTS)

$(INTEGRATION_TESTS):
	go test $(GOFLAGS) -tags $@ $(shell grep -lR "// +build $@" .)

$(LINTER):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/a2bc9b7a99e3280805309d71036e8c2106853250/install.sh \
	| sh -s -- -b $(basename $(GOPATH))/bin v1.23.8


$(MODULES_VENDOR): $(VENDOR_GO)
	cd $(shell dirname $@) && $(VENDOR_GO) mod vendor 
	find $@ -name go.mod -delete

$(VENDOR_GO):
	go get golang.org/dl/$(VENDOR_GO_VERSION)
	$(VENDOR_GO) download

.PHONY: test vendor $(MODULES) $(INTEGRATION_TESTS)
