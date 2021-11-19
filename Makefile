MODULES = $(filter-out $(EXCLUDE_DIRS), $(shell find . -name go.mod -exec dirname {} \;))
LINTER ?= $(shell go env GOPATH)/bin/golangci-lint

# The list of Go build tags as they are specified in respective integration test files
INTEGRATION_TESTS = fargate gcr lambda

# the Go version to vendor dependencies listed in go.mod
VENDOR_GO_VERSION ?= go1.15
VENDOR_GO = $(shell go env GOPATH)/bin/$(VENDOR_GO_VERSION)
MODULES_VENDOR = $(addsuffix /vendor,$(MODULES))

VERSION_TAG_PREFIX ?= v
VERSION_GO_FILE ?= $(shell grep -l '^var Version = ".\+"' *.go | head -1)

GIT_REMOTE ?= origin
GIT_MAIN_BRANCH ?= $(shell git remote show $(GIT_REMOTE) | sed -n '/HEAD branch/s/.*: //p')
GIT_TREE_STATE = $(shell (git status --porcelain | grep -q .) && echo dirty || echo clean)

GITHUB_CLI = $(shell which gh 2>/dev/null)

# Search for the latest vX.Y.Z tag ignoring all others, and use the X.Y.Z part as a version
GIT_VERSION = $(shell git tag | grep "^$(VERSION_TAG_PREFIX)[0-9]\+\(\.[0-9]\+\)\{2\}" | sort -V | tail -n1 | sed "s~^$(VERSION_TAG_PREFIX)~~" )
ifeq ($(GIT_VERSION),)
	GIT_VERSION = 0.0.0
endif

# Parse semantic version X.Y.Z found in git tags
GIT_MAJOR_VERSION = $(word 1,$(subst ., ,$(GIT_VERSION)))
GIT_MINOR_VERSION = $(word 2,$(subst ., ,$(GIT_VERSION)))
GIT_PATCH_VERSION = $(word 3,$(subst ., ,$(GIT_VERSION)))

# Parse version from version.go file
VERSION = $(shell grep -hm1 '^var Version = ".\+"' ${VERSION_GO_FILE} | head -1 | sed -E 's/var Version = "([0-9\.]+)"/\1/')

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

# Make sure there is a copyright at the first line of each .go file
legal:
	awk 'FNR==1 { if (tolower($$0) !~ "^//.+copyright") { print FILENAME" does not contain copyright header"; rc=1 } }; END { exit rc }' $$(find . -name '*.go' -type f | grep -v "/vendor/")

.PHONY: test vendor install legal $(MODULES) $(INTEGRATION_TESTS)

# Create changelog if GitHub CLI is installed
ifneq ($(GITHUB_CLI),)
release : CHANGELOG.txt
endif

# Create new release tag and publish a release from it
release :
	git tag $(VERSION_TAG_PREFIX)$(VERSION)
	git push --tags
ifneq ($(GITHUB_CLI),)
	$(GITHUB_CLI) release create $(VERSION_TAG_PREFIX)$(VERSION) \
		--draft \
		--title $(VERSION_TAG_PREFIX)$(VERSION) \
		--notes-file CHANGELOG.txt
	rm CHANGELOG.txt
else
	@echo "GitHub CLI is not installed. Please proceed with creating the release on https://github.com"
endif

# Calculate the next major version from the latest one found in git tags, update version.go, commit and push changes to remote
major : ensure-git-clean update-local-copy
	$(eval VERSION = $(shell echo $$(($(GIT_MAJOR_VERSION)+1))).0.0)
	@printf "You are about to increase the major version to $(VERSION), are you sure? [y/N]: " && read ans && [ $${ans:-N} = y ]
	sed -i '' 's/^var Version = ".*"/var Version = "$(VERSION)"/' ${VERSION_GO_FILE}
	git commit -m "Bump version to v$(VERSION)" ${VERSION_GO_FILE}
	git push $(GIT_REMOTE) HEAD

# Calculate the next minor version from the latest one found in git tags, update version.go, commit and push changes to remote
minor : ensure-git-clean update-local-copy
	$(eval VERSION = $(GIT_MAJOR_VERSION).$(shell echo $$(($(GIT_MINOR_VERSION)+1))).0)
	sed -i '' 's/^var Version = ".*"/var Version = "$(VERSION)"/' ${VERSION_GO_FILE}
	git commit -m "Bump version to v$(VERSION)" ${VERSION_GO_FILE}
	git push $(GIT_REMOTE) HEAD

# Calculate the next patch version from the latest one found in git tags, update version.go, commit and push changes to remote
patch : ensure-git-clean update-local-copy
	$(eval VERSION = $(GIT_MAJOR_VERSION).$(GIT_MINOR_VERSION).$(shell echo $$(($(GIT_PATCH_VERSION)+1))))
	sed -i '' 's/^var Version = ".*"/var Version = "$(VERSION)"/' ${VERSION_GO_FILE}
	git commit -m "Bump version to v$(VERSION)" ${VERSION_GO_FILE}
	git push $(GIT_REMOTE) HEAD

CHANGELOG.txt :
	echo 'This release includes the following fixes & improvements:' > CHANGELOG.txt
	git log --merges $(VERSION_TAG_PREFIX)$(GIT_VERSION).. --pretty='format:%s' | \
		sed 's~Merge pull request #\([0-9]*\).*~\1~' | \
		xargs -n1 $(GITHUB_CLI) pr view | \
		grep '^title:' | \
		sed 's~^title:~*~' >> CHANGELOG.txt
	echo >> CHANGELOG.txt

# Make sure there are no uncommitted changes in the working tree
ensure-git-clean :
ifeq ($(GIT_TREE_STATE),dirty)
	$(error "There are uncommitted changes in current working tree. Please stash or commit them before release.")
endif

# Switch to the main branch, then pull and rebase to the latest changes including tags
update-local-copy :
	git fetch --tags $(GIT_REMOTE)
	git checkout $(GIT_MAIN_BRANCH)
	git rebase $(GIT_REMOTE)/$(GIT_MAIN_BRANCH)

.PHONY : major minor patch release ensure-git-clean update-local-copy
