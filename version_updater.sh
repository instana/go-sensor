#!/bin/bash

# (c) Copyright IBM Corp. 2023

# ---------------------------------------------------------------------- #
# Script to monitor the packages and update the instrumentation packages #
# ---------------------------------------------------------------------- #

# Function to extract the package url, current version and local path from the markdown line

debug_log() {
    if [[ "$DEBUG_LOG" = true ]]; then
        echo "$@"
    fi
    return 0
}

extract_info_from_markdown_line() {
  local markdown_line=$1
  # Extract target-package-url, current version and local path from the markdown line using awk
  TARGET_PKG_URL=$(echo "$markdown_line" | awk -F '[(|)]' '{print $5}' | awk -F'https://pkg.go.dev/' '{print $2}')
  INSTANA_PKG_URL=$(echo "$markdown_line" | awk -F '[(|)]' '{print $8}')
  TARGET_PACKAGE_NAME=$(echo "$markdown_line" | awk -F '[][]' '{print $2}' | tr -d '[:space:]' | tr -d '()')
  INSTANA_PACKAGE_NAME=$(echo "$markdown_line" | awk -F '[][]' '{print $4}' | tr -d '[:space:]' | tr -d '()')
  LOCAL_PATH=$(echo "$INSTANA_PKG_URL" | awk -F 'github.com/instana/go-sensor/' '{print $2}')
  CURRENT_VERSION=$(echo "$markdown_line" | awk -F '|' '{print $7}' | tr -d '[:space:]')
}

# Function to find all the pending version updates for the instrumented packages.
find_versions_to_update(){
    local pkg=$1
    local curr_version="${2}"
    VERSIONS_TO_UPDATE=()
    debug_log "Current version: ", $curr_version

    if [[ -n "$pkg" ]]; then
        local url="https://proxy.golang.org/${pkg}/@v/list"
        local url1=$(echo "$url" | awk '{ print tolower($0) }')
        debug_log $url1

        local response=$(curl -s "$url1")
        IFS=$'\n' read -rd '' -a versions <<< "$response"
        sorted_versions=($(printf "%s\n" "${versions[@]}" | sort -V))
        local regex_pattern="^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$"
        
        debug_log "Versions: " "${versions[*]}"
        debug_log "Sorted versions: " "${sorted_versions[*]}"

        for i in "${!sorted_versions[@]}"; do
            if [[ "${sorted_versions[i]}" == "$curr_version" ]]; then
                for ((j=i+1; j<${#sorted_versions[@]}; j++)); do
                    if [[ "${sorted_versions[j]}" =~ $regex_pattern ]]; then

                        # Query the full info of that package version to get the release date
                        local url2="https://proxy.golang.org/${pkg}/@v/${sorted_versions[j]}.info"
                        local url3=$(echo "$url2" | awk '{ print tolower($0) }')
                        debug_log $url3
                        RELEASE_DATE=$(curl -s "$url3" | awk -F '[:,]' '{print $4}' | tr -d '"' | cut -d'T' -f1 | xargs -I {} date -d {} "+%a %b %d %Y")

                        local date1=$(date -d "$RELEASE_DATE" +%s)
                        local date2=$(date +%s) # today

                        local difference=$((date2 - date1))
                        DAYS_BEHIND_RELEASE=$((difference / 86400))

                        debug_log "Days behind release(${sorted_versions[j]}): $DAYS_BEHIND_RELEASE"

                        # if the package released 7 days ago, add it to the list of versions to update
                        if [[ $DAYS_BEHIND_RELEASE -gt 7 ]]; then
                            VERSIONS_TO_UPDATE+=("${sorted_versions[j]}")   
                        fi               
                    fi
                done
            fi
        done

    else
      echo "Invalid package location: $pkg"
      VERSIONS_TO_UPDATE=()      
  fi

  echo "Versions to update: " "${VERSIONS_TO_UPDATE[*]}"
  return 0

}

DEBUG_LOG=true
TRACER_PATH=$(pwd)
LIBRARY_INFO_MD_PATH=$(pwd)/supported_versions.md
LIBRARY_INFO_MD_TMP=$(pwd)/supported_versions_temp.md
LIBRARY_INFO_MD_PATH_COPY=$(pwd)/supported_versions_copy.md

# Check if the file exists
if [[ ! -f "$LIBRARY_INFO_MD_PATH" ]]; then
  echo "Error: File '$LIBRARY_INFO_MD_PATH' not found."
  exit 1
fi

# Copy the original file
cp $LIBRARY_INFO_MD_PATH $LIBRARY_INFO_MD_PATH_COPY

# Open the file and read it line by line
first_line=true
while IFS= read -r line; do
  # Skip the first line
  # As it only contains the markdown headers
  if [[ "$first_line" = true ]]; then
    first_line=false
    continue
  fi

  echo "Processing line: $line"
  extract_info_from_markdown_line "$line"

  # Create a branch and commit the changes
  git config user.name "tracer-team-go"
  git config user.email "github-actions@github.com"


  folder=$TRACER_PATH/$LOCAL_PATH

  INSTRUMENTATION=$INSTANA_PACKAGE_NAME

  echo "--------------$INSTRUMENTATION-----------------"

  # Print the extracted values
  echo "Target Package URL: $TARGET_PKG_URL"
  echo "Instana Package URL: $INSTANA_PKG_URL"
  echo "Target Package Text: $TARGET_PACKAGE_NAME"
  echo "Instana Package Text: $INSTANA_PACKAGE_NAME"
  echo "Local Path: $LOCAL_PATH"
  echo "Current version: $CURRENT_VERSION"

  if [[ -z "$TARGET_PKG_URL" ]]; then
    continue
  fi

  # Find all the pending version updates for the instrumented packages.
  find_versions_to_update "$TARGET_PKG_URL" "$CURRENT_VERSION"

  for i in "${!VERSIONS_TO_UPDATE[@]}"; do
    LATEST_VERSION="${VERSIONS_TO_UPDATE[i]}"
    echo "Latest version:" "$LATEST_VERSION"

    git checkout main

    PR_TITLE="feat(currency): updated instrumentation of $INSTRUMENTATION for new version $LATEST_VERSION"
    COMMIT_MSG="feat(currency): updated go.mod, go.sum files, README.md for $INSTRUMENTATION"

    if gh pr list | grep -q "$PR_TITLE"; then
      echo "PR for $INSTRUMENTATION newer version:$LATEST_VERSION already exists. Skipping to next iteration"
      continue
    fi

    echo "Update needed for this package. Update process starting..."

    cd $folder

    # For some packages, eg : https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/cloud.google.com/go/storage
    # The go.mod file will be in the previous directory
    # Need this check here to proceed to the correct directory containing go.mod
    LOCAL_PATH_2=$(go list -m | awk -F 'github.com/instana/go-sensor/' '{print $2}')

    if [[ "$LOCAL_PATH" = "$LOCAL_PATH_2" ]]; then
      echo "No need to change working directory!"
    else
      # change working folder to the correct path
      folder=$TRACER_PATH/$LOCAL_PATH_2
      cd "$folder" || continue
    fi

    go get "$TARGET_PKG_URL@$LATEST_VERSION"
    go mod edit -toolchain=none
    go mod tidy
    go test ./... || echo "Continuing the operation even if the test fails. This needs manual intervention"

    # Need to update the current version in the supported_versions.md file
    new_line=$(echo "$line" | awk -v old_value="$CURRENT_VERSION" -v new_value="$LATEST_VERSION" '{ for (i=NF; i>0; i--) if ($i == old_value) { $i = new_value; break } }1')
    awk -v new_line="$new_line" '{ if ($0 == old_line) print new_line; else print }' old_line="$line" $LIBRARY_INFO_MD_PATH >$LIBRARY_INFO_MD_TMP && mv $LIBRARY_INFO_MD_TMP $LIBRARY_INFO_MD_PATH

    CURRENT_TIME_UNIX=$(date '+%s')
    git checkout -b "update-instrumentations-$INSTRUMENTATION-id-$CURRENT_TIME_UNIX"

    git add go.mod go.sum $LIBRARY_INFO_MD_PATH
    git commit -m "$COMMIT_MSG"
    git push origin @

    # Create a PR request for the changes
    # shellcheck disable=SC2046
    gh pr create --title "$PR_TITLE. Id: $CURRENT_TIME_UNIX" \
      --body "This PR adds changes for the newer version $LATEST_VERSION for the instrumented package" --head $(git branch --show-current) --label "tekton_ci"

    # Back to working directory
    cd $TRACER_PATH
  done

done <"$LIBRARY_INFO_MD_PATH_COPY"
