#!/bin/bash

# (c) Copyright IBM Corp. 2024

# Function to extract the package url and other data from the markdown line
extract_info_from_markdown_line() {
        local markdown_line=$1
        # Extract target-package-url and local path from the markdown line using awk
        TARGET_PKG_URL=$(echo "$markdown_line" | awk -F '[(|)]' '{print $10}' | awk -F'https://pkg.go.dev/' '{print $2}')
        INSTANA_PKG_URL=$(echo "$markdown_line" | awk -F '[(|)]' '{print $13}')
        TARGET_PACKAGE_NAME=$(echo "$markdown_line" | awk -F '|' '{print $2}')
        LOCAL_PATH=$(echo "$INSTANA_PKG_URL" | awk -F 'github.com/instana/go-sensor/' '{print $2}')

        IS_STANDARD_LIBRARY="false"
        IS_DEPRECATED="false"

        if [[ $(echo "$markdown_line" | awk -F '|' '{print $11}') == " Standard library " ]]; then
            echo "Standard library"
            IS_STANDARD_LIBRARY="true"
        fi

        if [[ $(echo "$markdown_line" | awk -F '|' '{print $3}') == " Deprecated " ]]; then
            echo "Deprecated"
            IS_DEPRECATED="true"
        fi

        
}

# Function to query the latest released version of the package
find_latest_version() {
  local pkg=$1
  if [ -n "$pkg" ]; then
    # Query the latest version for the package
    local url="https://proxy.golang.org/${pkg}/@latest"
    local url1=$(echo "$url" | awk '{ print tolower($0) }')
    echo $url1
    LATEST_VERSION=$(curl -s "$url1" | awk -F '[:,]' '/Version/{print $2}' | tr -d '"' | tr -d 'v')
  else
      LATEST_VERSION=""
      echo "Invalid package location: $pkg"
  fi

}

# Function to find the current version of the package using go command.
find_current_version(){
    local pkg=$1
    CURRENT_VERSION=$(go list -m $pkg | awk '{print $NF}')
    if ! [[ "$CURRENT_VERSION" =~ ^"v" ]]; then
        echo "Invalid current version"
        CURRENT_VERSION="" 
    else
        CURRENT_VERSION=$(echo "$CURRENT_VERSION" | tr -d 'v')
    fi
}

# This script needs to be called from the go-sensor folder.
GO_TRACER_REPO_PATH=$(pwd)
TRACER_REPORTS_REPO_PATH=$(pwd)/../tracer-reports
GO_REPORTS_MD_PATH=$TRACER_REPORTS_REPO_PATH/automated/currency/go/report.md
GO_REPORTS_MD_PATH_COPY=$TRACER_REPORTS_REPO_PATH/automated/currency/go/report_copy.md
GO_REPORTS_MD_PATH_TMP=$TRACER_REPORTS_REPO_PATH/automated/currency/go/report_temp.md

# Copy the original file
cp $GO_REPORTS_MD_PATH $GO_REPORTS_MD_PATH_COPY

skip_execution=true
first_line=true
while IFS= read -r line; do

    if [ "$first_line" = true ]; then
        first_line=false
        changed_line="#### Note: This page is auto-generated. Any change will be overwritten after the next sync. For more details on Go Tracer SDK, please visit our [Github](https://github.com/instana/go-sensor) page. Last updated on $(date +'%d.%m.%Y')."
        # Update the markdown report file
        awk -v new_line="$changed_line" '{ if ($0 == old_line) print new_line; else print }' old_line="$line" $GO_REPORTS_MD_PATH > $GO_REPORTS_MD_PATH_TMP && mv $GO_REPORTS_MD_PATH_TMP $GO_REPORTS_MD_PATH
        continue
    fi

    # For skipping first few lines from the md file.
    if [ "$skip_execution" = true ]; then
        if [ "$line" = "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |" ]; then
            skip_execution=false
        fi
        continue
    fi

    # echo "Processing line: $line"
    extract_info_from_markdown_line "$line"

    if [ "$IS_STANDARD_LIBRARY" = "true" ] || [ "$IS_DEPRECATED" = "true" ]; then
        # skip execution
        continue
    fi
    
    folder=$GO_TRACER_REPO_PATH/$LOCAL_PATH

    cd $folder
    # Find the latest version of the instrumented package
    find_latest_version "$TARGET_PKG_URL"
    # echo "Latest version:" "$LATEST_VERSION"
    find_current_version "$TARGET_PKG_URL"
    # echo "Current version:" "$CURRENT_VERSION"

    # Replace supported and latest version in the markdown line
    changed_line=$(echo "$line" | awk -v new_val="$(printf ' %s ' "$CURRENT_VERSION")" 'BEGIN{OFS=FS="|"} {$5=new_val} 1')
    changed_line=$(echo "$changed_line" | awk -v new_val="$(printf ' %s ' "$LATEST_VERSION")" 'BEGIN{OFS=FS="|"} {$6=new_val} 1')

    if [ "$LATEST_VERSION" != "$CURRENT_VERSION" ]; then
        echo "Some difference!"
        echo "Latest version:" "$LATEST_VERSION"
        echo "Current version:" "$CURRENT_VERSION"

        changed_line=$(echo "$changed_line" | awk -v new_val=" No " 'BEGIN{OFS=FS="|"} {$7=new_val} 1')
    else
        changed_line=$(echo "$changed_line" | awk -v new_val=" Yes " 'BEGIN{OFS=FS="|"} {$7=new_val} 1')
    fi

    # Update the markdown report file
    awk -v new_line="$changed_line" '{ if ($0 == old_line) print new_line; else print }' old_line="$line" $GO_REPORTS_MD_PATH > $GO_REPORTS_MD_PATH_TMP && mv $GO_REPORTS_MD_PATH_TMP $GO_REPORTS_MD_PATH

    cd $GO_TRACER_REPO_PATH
done < "$GO_REPORTS_MD_PATH_COPY"