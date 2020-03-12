MODULES = $(shell find . -name go.mod -exec dirname {} \;)

test: $(MODULES)

$(MODULES):
	cd $@ && go get -d -t ./... && go test $(GOFLAGS) ./...

.PHONY: test $(MODULES)
