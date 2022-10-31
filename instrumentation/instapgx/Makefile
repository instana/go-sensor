GO=go
GO_MINOR_VERSION = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
MINIMUM_SUPPORTED_GO_MINOR_VERSION = 13
GO_MODULE_NAME ?= github.com/instana/go-sensor/instrumentation/instapgx
VERSION_TAG_PREFIX ?= instrumentation/instapgx/v

pgx_integration_test:
	if [ $(GO_MINOR_VERSION) -gt $(MINIMUM_SUPPORTED_GO_MINOR_VERSION) ]; then \
		cd instrumentation/instapgx && go test -tags=integration; \
	elif [ $(GO_MINOR_VERSION) -lt $(MINIMUM_SUPPORTED_GO_MINOR_VERSION) ] ; then \
		echo "pgx_integration_test is ignored"; \
		exit 0; \
	fi


include ../../Makefile.release