#!/bin/bash

# (c) Copyright IBM Corp. 2023

# ---------------------------------------------------------------------- #
# Script to monitor the packages and update the instrumentation packages #
# ---------------------------------------------------------------------- #

# Function to extract the package url and current version from the metadata
extract_info_from_markdown() {
    if [ -e "$1" ]; then
        local markdown_text=$(<"$1")
        # Extract target-package-url and current-version using awk
        TARGET_PKG_URL=$(echo "$markdown_text" | awk -F: '/target-pkg-url:/ {print $2}' | tr -d '[:space:]')
        CURRENT_VERSION=$(echo "$markdown_text" | awk -F: '/current-version:/ {print $2}' | tr -d '[:space:]')
    else
        echo "Error: File not found - $1"
        TARGET_PKG_URL=""
        CURRENT_VERSION=""
    fi
}

# Function to query the latest released version of the package
find_latest_version() {
  local pkg=$1
  if [ -n "$pkg" ]; then
      # Query the latest version for the package
      local url="https://proxy.golang.org/${pkg}/@latest"
      LATEST_VERSION=$(curl -s "$url" | jq .Version | tr -d '"')
  else
      LATEST_VERSION=""
      echo "Invalid package location: $pkg"
  fi

}

# Function to compare versions
version_compare() {
    local version1=$1
    local version2=$2

    local major_version1=$(echo "$version1" | sed -E 's/v([0-9]+)\.([0-9]+)\..*/\1/')
    local minor_version1=$(echo "$version1" | sed -E 's/v([0-9]+)\.([0-9]+)\..*/\2/')
    local major_version2=$(echo "$version2" | sed -E 's/v([0-9]+)\.([0-9]+)\..*/\1/')
    local minor_version2=$(echo "$version2" | sed -E 's/v([0-9]+)\.([0-9]+)\..*/\2/')

    # We are checking the changes in minor versions for automation purpose
    if [ "$major_version1" = "$major_version2" ] && [ "$minor_version1" -gt "$minor_version2" ]; then
      echo "true"
    else
      echo "false"
    fi
}

# Function to update the metadata with the latest version information
replace_version_in_file() {
    local version=$1
    local file_path=$2

    # Read the content of the file
    local file_content=$(<"$file_path")

    # Replace current-version with the new version
    # shellcheck disable=SC2001
    local updated_content=$(echo "$file_content" | sed "s/current-version: [^ ]*/current-version: $version/")

    # Write the updated content back to the file
    echo "$updated_content" > "$file_path"
    echo "Version in file $file_path updated to $version"
}

DIRECTORY_PATH=$(pwd)/instrumentation
echo "$DIRECTORY_PATH"

if [ -d "$DIRECTORY_PATH" ]; then
    for folder in "$DIRECTORY_PATH"/*/; do
        # Create a branch and commit the changes
        git config user.name "IBM/Instana/Team Go"
        git config user.email "github-actions@github.com"

        git checkout main

        INSTRUMENTATION=$(basename "$folder")
        # Identify the path to the README file
        README_PATH="${folder}README.md"

        echo "--------------$INSTRUMENTATION-----------------"
        if [ ! -e "$README_PATH" ]; then
          continue
        fi

        # Extract the metadata from the README file
        extract_info_from_markdown "$README_PATH"

        if [ -z "$TARGET_PKG_URL" ]; then
          continue
        fi

        # Print the extracted values
        #echo "Target Package URL: $TARGET_PKG_URL"
        echo "Current Version: $CURRENT_VERSION"

        # Find the latest version of the instrumented package
        find_latest_version "$TARGET_PKG_URL"
        echo "Latest version:" "$LATEST_VERSION"

        version_compare "$LATEST_VERSION" "$CURRENT_VERSION"
        update_needed=$( version_compare "$LATEST_VERSION" "$CURRENT_VERSION" )

        if [ "$update_needed" != true ]; then
          continue
        fi

        if gh pr list | grep -q "instrumentation $INSTRUMENTATION for new version $LATEST_VERSION"; then
          echo "PR for $INSTRUMENTATION newer version:$LATEST_VERSION already exists. Skipping to next iteration"
          continue
        fi

        echo "Update needed for this package. Update process starting..."
        cd "$folder" || continue
        go get "$TARGET_PKG_URL"
        go mod tidy
        go test ./... || echo "Continuing the operation even if the test fails. This needs manual intervention"


        # Need to update the current version in the README file
        replace_version_in_file "$LATEST_VERSION" "$README_PATH"

        CURRENT_TIME_UNIX=$(date '+%s')
        git checkout -b "update-instrumentations-$INSTRUMENTATION-id-$CURRENT_TIME_UNIX"

        git add go.mod go.sum README.md
        git commit -m "Updated go.mod, go.sum files, README.md for $INSTRUMENTATION"
        git push origin @

        # Create a PR request for the changes
        # shellcheck disable=SC2046
        gh pr create --title "Updating instrumentation $INSTRUMENTATION for new version $LATEST_VERSION. Id: $CURRENT_TIME_UNIX" \
        --body "This PR adds changes for the newer version $LATEST_VERSION for the instrumented package" --head $(git branch --show-current)

    done
else
    echo "Error: The specified path is not a directory."
fi
