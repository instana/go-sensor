#!/bin/sh

# This script provides utilities to update all instrumentations to refer to the latest core module.
#
# Usage
#
# Update instrumentation phase
#
# 1. Run `./instrumentations.sh update`. This will update all instrumentations to reference the latest core module.
# 2. Run `make test` to ensure that all instrumentations work with the latest core.
# 3. If any errors are detected in some instrumentation, fix them.
# 4. If everything is fine, commit your changes, open a PR and get it merged into the main branch.
#
# Release phase
# 1. Run `./instrumentations.sh release` to create tags for each instrumentation with a new minor version, and update all version.go files.

# Checks if gh is installed, otherwise stop the script
if ! [ -x "$(command -v gh)" ]; then
  echo "Error: gh is not installed." >&2
  exit 1
fi

# Checks if the user is logged into Github, otherwise stop the script
if gh auth status 2>&1 | grep -i "You are not logged"; then
  echo "Error: You must log into Github." >&2
  exit 1
fi

CORE_VERSION=latest

# List of folders to be excluded from the instrumentation list
# If new matches must be added, use regular expressions. eg:
# EXCLUDED_DIRS="\/.*\/example|\/new_match"
EXCLUDED_DIRS="\/.*\/example"

# List of instrumentation folders
LIB_LIST=$(find ./instrumentation -name go.mod -exec dirname {} \; | grep -E -v "$EXCLUDED_DIRS")

# Updates all instrumentations to use the @latest version of the core module
run_update() {
  for lib in $LIB_LIST
    do cd "$lib" && go get github.com/instana/go-sensor@$CORE_VERSION && go mod tidy && cd -;
  done
}

# Updates version.go and creates a new tag for every instrumentation, incrementing the minor version
run_release() {
  TAGS=""
  for lib in $LIB_LIST
    do LIB_PATH="$(echo "$lib" | sed 's/\.\///')"

    # Expected to find something like: instrumentation/instaredis/v1.5.0
    # This option will be used if the instrumentation has no v2 subfolder
    TAG_TO_SEARCH="$LIB_PATH/v[0-1].*"

    # Expected to identify packages with subfolders. eg: instrumentation/instaredis/v2
    NEW_VERSION_FOLDER=$(echo "$lib" | grep -E "v[2-9].*")

    # If NEW_VERSION_FOLDER has something we update TAG_TO_SEARCH
    if [ -n "$NEW_VERSION_FOLDER" ]; then
      # Expected to be a version. eg: 1.5.0
      NEW_MAJOR_VERSION=$(echo "$NEW_VERSION_FOLDER" | sed "s/.*v//")

      # Expected to be tag name with major version higher than 1. eg: instrumentation/instaredis/v2.1.0
      TAG_TO_SEARCH="$LIB_PATH/v$NEW_MAJOR_VERSION.*"
    fi

    VERSION=$(git tag -l "$TAG_TO_SEARCH" | sort -V | tail -n1 | sed "s/.*v//")

    if [ -z "$VERSION" ]; then
      VERSION="0.0.0"
    fi

    MINOR_VERSION=$(echo "$VERSION" | sed -En 's/[0-9]+\.([0-9]+)\.[0-9]+/\1/p')
    MAJOR_VERSION=$(echo "$VERSION" | sed -En 's/([0-9]+)\.[0-9]+\.[0-9]+/\1/p')
    MINOR_VERSION=$((MINOR_VERSION+1))
    NEW_VERSION="$MAJOR_VERSION.$MINOR_VERSION.0"

    # Updates the minor version in version.go
    sed -i '' -E "s/[0-9]+\.[0-9]+\.[0-9]+/${NEW_VERSION}/" "$lib"/version.go | tail -1

    # Tags to be created after version.go is merged to the main branch with the new version
    PATH_WITHOUT_V=$(echo "$LIB_PATH" | sed "s/\/v[0-9]*//")
    TAGS="$TAGS $PATH_WITHOUT_V/v$MAJOR_VERSION.$MINOR_VERSION.0"
  done

  # Commit all version.go files to the main branch
  git add ./instrumentation/**/version.go
  git add ./instrumentation/**/**/version.go
  git commit -m "Bumping new version of the instrumentation"
  git push origin main

  echo "Creating tags for each instrumentation"

  for t in $TAGS
    do git tag "$t" && git push origin "$t"
  done

  # Release every instrumentation
  for t in $TAGS
    do gh release create "$t" \
		--title "$t" \
		--notes "Updated instrumentation with the latest version of go-sensor core module.<br/><br/> --auto-generated--"
  done
}

run_replace() {
  ROOT_DIR="$PWD"

  for lib in $LIB_LIST
    do cd "$lib" && go mod edit -replace=github.com/instana/go-sensor="$ROOT_DIR" && cd -;
  done
}

run_dropreplace() {
  ROOT_DIR="$PWD"

  for lib in $LIB_LIST
    do cd "$lib" && go mod edit -dropreplace=github.com/instana/go-sensor && cd -;
  done
}

if [ "$1" = "update" ]; then
  run_update
  exit 0
fi

if [ "$1" = "release" ]; then
  run_release
  exit 0
fi

if [ "$1" = "replace" ]; then
  run_replace
  exit 0
fi

if [ "$1" = "dropreplace" ]; then
  run_dropreplace
  exit 0
fi

echo "---------------------------------------------------------"
echo "Usage: $0 COMMAND"
echo ""
echo "Where COMMAND can be:"
echo "- update: Updates every instrumentation to reference the latest version of the core"
echo "- release: Releases all instrumentations by increasing a minor version"
echo "- replace: Adds the 'replace' directive into the go.mod file of each instrumentation changing github.com/instana/go-sensor to the local path"
echo "- dropreplace: Removes existing 'replace' directives that changes github.com/instana/go-sensor to a different path"
echo ""
echo "Example: $0 update"
echo "---------------------------------------------------------"
