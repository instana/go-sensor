test: linux

GOVERSION ?= latest
linux: *.go
	docker run --rm -v `pwd`/..:/go/src/github.com/instana/go-sensor golang:$(GOVERSION) \
		/bin/bash -c "cd /go/src/github.com/instana/go-sensor/process && go test"

.PHONY: test
