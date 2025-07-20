#!/bin/bash

build_major() {
  # Only 2 scenarios:
  # - v1 was never released, so FOUND_VERSION_IN_TAG=0.0.0 and NEW_MAJOR_VERSION is empty. NEW_VERSION should be 1.0.0
  # - v2 or higher was never released, so FOUND_VERSION_IN_TAG=0.0.0 and NEW_MAJOR_VERSION is 2, 3, 4... NEW_VERSION should be 2.0.0, or higher

  MAJOR_VERSION=$(echo "$FOUND_VERSION_IN_TAG" | sed -En 's/([0-9]+)\.[0-9]+\.[0-9]+.*/\1/p')

  if [ "$MAJOR_VERSION" != "0" ]; then
    echo "Cannot release new major version '$NEW_MAJOR_VERSION' with existing tag $FOUND_VERSION_IN_TAG"
    exit 1
  fi

  if [ -z "$NEW_MAJOR_VERSION" ]; then
    NEW_VERSION="1.0.0"
  else
    NEW_VERSION="$NEW_MAJOR_VERSION.0.0"
  fi
}

build_minor() {
  # We abort the release if there is an attempt to release a minor version of a major v2 or higher that doesn't exist.
  # This may happen if the v2 folder exists, but the major v2 hasn't been released yet.
  if [ "$FOUND_VERSION_IN_TAG" = "0.0.0" ] && [ -n "$NEW_MAJOR_VERSION" ]; then
    echo "Cannot release a minor version of a major v2 or higher that doesn't exist"
    exit 1
  fi

  MINOR_VERSION=$(echo "$FOUND_VERSION_IN_TAG" | sed -En 's/[0-9]+\.([0-9]+)\.[0-9]+.*/\1/p')
  MAJOR_VERSION=$(echo "$FOUND_VERSION_IN_TAG" | sed -En 's/([0-9]+)\.[0-9]+\.[0-9]+.*/\1/p')
  MINOR_VERSION=$((MINOR_VERSION+1))
  NEW_VERSION="$MAJOR_VERSION.$MINOR_VERSION.0"
}

build_patch() {
  # We abort the release if there is an attempt to release a patch version of a major v2 or higher that doesn't exist.
  # This may happen if the v2 folder exists, but the major v2 hasn't been released yet.
  if [ "$FOUND_VERSION_IN_TAG" = "0.0.0" ] && [ -n "$NEW_MAJOR_VERSION" ]; then
    echo "Cannot release a patch version of a major v2 or higher that doesn't exist"
    exit 1
  fi

  PATCH_VERSION=$(echo "$FOUND_VERSION_IN_TAG" | sed -En 's/[0-9]+\.[0-9]+\.([0-9]+).*/\1/p')
  MINOR_VERSION=$(echo "$FOUND_VERSION_IN_TAG" | sed -En 's/[0-9]+\.([0-9]+)\.[0-9]+.*/\1/p')
  MAJOR_VERSION=$(echo "$FOUND_VERSION_IN_TAG" | sed -En 's/([0-9]+)\.[0-9]+\.[0-9]+.*/\1/p')
  PATCH_VERSION=$((PATCH_VERSION+1))
  NEW_VERSION="$MAJOR_VERSION.$MINOR_VERSION.$PATCH_VERSION"
}

IS_CORE="false"

LIB_PATH=.

if [ "$INSTANA_PACKAGE_NAME" = "." ]; then
  IS_CORE="true"
  echo "Releasing core module"
else
  echo "Releasing $INSTANA_PACKAGE_NAME"
fi

if [ "$IS_CORE" = "false" ]; then
  cd instrumentation/"$INSTANA_PACKAGE_NAME" || exit
  LIB_PATH=instrumentation/$INSTANA_PACKAGE_NAME
fi

echo "lib path: $LIB_PATH"

# Expected to find something like: instrumentation/instaredis/v1.5.0
# This option will be used if the instrumentation has no v2 subfolder
OPTIONAL_GREP_STR="v[0-9]+\.[0-9]+\.[0-9]+$"
TAG_TO_SEARCH="v[0-1].[0-9]*.[0-9]*"

if [ "$IS_CORE" = "false" ]; then
  TAG_TO_SEARCH="$LIB_PATH/$TAG_TO_SEARCH"
fi

# Only relevant for instrumentations
if [ "$IS_CORE" = "false" ]; then
  # Expected to identify packages with subfolders. eg: instrumentation/instaredis/v2
  NEW_VERSION_FOLDER=$(echo "$LIB_PATH" | grep -E "/v[2-9].*")

  echo "New version folder. eg: v2, v3...: $NEW_VERSION_FOLDER"

  # If NEW_VERSION_FOLDER has something we update TAG_TO_SEARCH
  if [ -n "$NEW_VERSION_FOLDER" ]; then
    # Expected to parse a version. eg: 2.5.0. Will extract 2 in the case of 2.5.0, which is the new major version
    NEW_MAJOR_VERSION=$(echo "$NEW_VERSION_FOLDER" | sed "s/.*v//")

    echo "New major version: $NEW_MAJOR_VERSION"

    # Expected to be tag name with major version higher than 1. eg: instrumentation/instaredis/v2.1.0
    TAG_TO_SEARCH="$LIB_PATH.[0-9]*.[0-9]*"
  fi
fi

echo "Tag to search: $TAG_TO_SEARCH"

# git fetch --unshallow --tags
FOUND_VERSION_IN_TAG=$(git tag -l "$TAG_TO_SEARCH" | grep -E "$OPTIONAL_GREP_STR" | sort -V | tail -n1 | sed "s/.*v//")

echo "Version found in tags: $FOUND_VERSION_IN_TAG"

# This works if the lib is new, and it was never instrumented.
# But if it's a new version, eg: /v2, this will fail later. we need to fix it
if [ -z "$FOUND_VERSION_IN_TAG" ]; then
  FOUND_VERSION_IN_TAG="0.0.0"
  echo "Version updated to: $FOUND_VERSION_IN_TAG, and new major version is $NEW_MAJOR_VERSION"
fi

if [ "$LIB_VERSION_TYPE" = "major" ]; then
  build_major
elif [ "$LIB_VERSION_TYPE" = "minor" ]; then
  build_minor
else
  build_patch
fi

echo "New version to release is: $NEW_VERSION"

# Updates the minor version in version.go
sed -i -E "s/[0-9]+\.[0-9]+\.[0-9]+/${NEW_VERSION}/" version.go | tail -1

git remote set-url --push origin https://$GIT_COMMITTER_NAME:$GITHUB_TOKEN@github.com/instana/go-sensor

git add version.go
git commit -m "Updated version.go to $NEW_VERSION"
git push origin @

# Tags to be created after version.go is merged to the main branch with the new version
PATH_WITHOUT_V=$(echo "$LIB_PATH" | sed "s/\/v[0-9]*//")

if [ "$IS_CORE" = "false" ]; then
  NEW_VERSION_TAG="$PATH_WITHOUT_V/v$NEW_VERSION"
else
  NEW_VERSION_TAG="v$NEW_VERSION"
fi

echo "Releasing as draft: $RELEASE_AS_DRAFT"

AS_DRAFT="-d"

if [ "$RELEASE_AS_DRAFT" != "true" ]; then
  AS_DRAFT=""
fi

echo "RELEASE_VERSION=v$NEW_VERSION" >> "$GITHUB_OUTPUT"

if [ "$IS_CORE" = "false" ]; then
  echo "RELEASE_PACKAGE=$INSTANA_PACKAGE_NAME" >> "$GITHUB_OUTPUT"
else
  echo "RELEASE_PACKAGE=go-sensor" >> "$GITHUB_OUTPUT"
fi

echo "$GITHUB_TOKEN" > gh_token.txt
gh auth login --with-token < gh_token.txt
rm gh_token.txt

git tag "$NEW_VERSION_TAG"
git push origin "$NEW_VERSION_TAG"
gh release create "$NEW_VERSION_TAG" $AS_DRAFT --title "$NEW_VERSION_TAG" --notes "New release $NEW_VERSION_TAG."
