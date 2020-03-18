
MODULES = $(filter-out $(EXCLUDE_DIRS), $(shell find . -name go.mod -exec dirname {} \;))
LINTER ?= $(shell go env GOPATH)/bin/golangci-lint

ifdef RUN_LINTER
test: $(LINTER)
endif

test: $(MODULES)

$(MODULES):
	cd $@ && go get -d -t ./... && go test $(GOFLAGS) ./...
ifdef RUN_LINTER
	cd $@ && $(LINTER) run
endif

$(LINTER):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/a2bc9b7a99e3280805309d71036e8c2106853250/install.sh \
	| sh -s -- -b $(basename $(GOPATH))/bin v1.23.8

.PHONY: test $(MODULES)
