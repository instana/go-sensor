#!/usr/bin/env bash

# (c) Copyright IBM Corp. 2024

debug_log() {
    if [[ "$DEBUG_LOG" = true ]]; then
        echo "$@"
    fi
}

# Function to extract the package url and other data from the markdown line
extract_info_from_markdown_line() {
        local markdown_line=$1
        # Extract target-package-url and local path from the markdown line using awk
        TARGET_PKG_URL=$(echo "$markdown_line" | awk -F '[(|)]' '{print $12}' | awk -F'https://pkg.go.dev/' '{print $2}')
        INSTANA_PKG_URL=$(echo "$markdown_line" | awk -F '[(|)]' '{print $15}')
        TARGET_PACKAGE_NAME=$(echo "$markdown_line" | awk -F '|' '{print $2}')
        LOCAL_PATH=$(echo "$INSTANA_PKG_URL" | awk -F 'github.com/instana/go-sensor/' '{print $2}')

        IS_STANDARD_LIBRARY="false"
        IS_DEPRECATED="false"

        if [[ $(echo "$markdown_line" | awk -F '|' '{print $13}') == " Standard library " ]]; then
            debug_log "Standard library"
            IS_STANDARD_LIBRARY="true"
        fi

        if [[ $(echo "$markdown_line" | awk -F '|' '{print $3}') == " Deprecated " ]]; then
            debug_log "Deprecated"
            IS_DEPRECATED="true"
        fi
}

# Function to query the latest released version of the package
find_latest_version() {
  local pkg=$1
  if [[ -n "$pkg" ]]; then
    # Query the latest version for the package
    local url="https://proxy.golang.org/${pkg}/@latest"
    local url1=$(echo "$url" | awk '{ print tolower($0) }')
    debug_log $url1
    LATEST_VERSION=$(curl -s "$url1" | awk -F '[:,]' '{print $2}' | tr -d '"' | tr -d 'v')
    LATEST_VERSION_DATE=$(curl -s "$url1" | awk -F '[:,]' '{print $4}' | tr -d '"' | cut -d'T' -f1 | xargs -I {} date -d {} "+%a %b %d %Y")
  else
      LATEST_VERSION=""
      LATEST_VERSION_DATE=""
      debug_log "Invalid package location: $pkg"
      echo "Invalid package location: $pkg" >> $OUTPUT_TO_SLACK
  fi

}

# Function to find the current version of the package using go command.
find_current_version() {
    local pkg=$1
    CURRENT_VERSION=$(go list -m $pkg | awk '{print $NF}')
    if ! [[ "$CURRENT_VERSION" =~ ^"v" ]]; then
        debug_log "Invalid current version"
        CURRENT_VERSION="" 
    else
        CURRENT_VERSION=$(echo "$CURRENT_VERSION" | tr -d 'v')
    fi
}

# Function to find the next immediate released version of the current version
# This helps in accurately calculating the number of days delayed
find_immediate_next_version(){
    local pkg=$1
    local curr_version="v${2}"
    debug_log "Current version: ", $curr_version

    if [[ -n "$pkg" ]]; then
        local url="https://proxy.golang.org/${pkg}/@v/list"
        local url1=$(echo "$url" | awk '{ print tolower($0) }')
        debug_log $url1

        local response=$(curl -s "$url1")
        IFS=$'\n' read -rd '' -a versions <<< "$response"
        sorted_versions=($(printf "%s\n" "${versions[@]}" | sort -V))
        local regex_pattern="^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$"
        local found=false
        
        debug_log "Versions: " "${versions[*]}"
        debug_log "Sorted versions: " "${sorted_versions[*]}"

        for i in "${!sorted_versions[@]}"; do
            if [[ "${sorted_versions[i]}" == "$curr_version" ]]; then
                for ((j=i+1; j<${#sorted_versions[@]}; j++)); do
                    if [[ "${sorted_versions[j]}" =~ $regex_pattern ]]; then
                        IMMEDIATE_NEXT_VERSION=${sorted_versions[j]#v}
                        debug_log $IMMEDIATE_NEXT_VERSION
                        found=true
                        break 2
                    fi
                done
            fi
        done

        if [[ "$found" = true ]]; then
            # Query the full info of that package version to get the release darte
            local url="https://proxy.golang.org/${pkg}/@v/v${IMMEDIATE_NEXT_VERSION}.info"
            local url1=$(echo "$url" | awk '{ print tolower($0) }')
            debug_log $url1
            IMMEDIATE_NEXT_VERSION_DATE=$(curl -s "$url1" | awk -F '[:,]' '{print $4}' | tr -d '"' | cut -d'T' -f1 | xargs -I {} date -d {} "+%a %b %d %Y")
        else
            IMMEDIATE_NEXT_VERSION=""
            IMMEDIATE_NEXT_VERSION_DATE=""
            echo "Something wrong with querying immediate next version - $pkg" >> $OUTPUT_TO_SLACK
        fi

    else
        IMMEDIATE_NEXT_VERSION=""
        IMMEDIATE_NEXT_VERSION_DATE=""
        debug_log "Invalid package location: $pkg"
        echo "Invalid package location: $pkg" >> $OUTPUT_TO_SLACK
        echo "Something wrong with finding immediate next version - $pkg" >> $OUTPUT_TO_SLACK
  fi

}

find_days_behind_last_support() {

    local date1=$(date -d "$1" +%s)
    local date2=$(date +%s) # today

    local difference=$((date2 - date1))
    DAYS_BEHIND=$((difference / 86400))
}



# This script needs to be called from the go-sensor folder.
GO_TRACER_REPO_PATH=$(pwd)
TRACER_REPORTS_REPO_PATH=$(pwd)/../tracer-reports
GO_REPORTS_MD_PATH=$TRACER_REPORTS_REPO_PATH/automated/currency/go/report.md
GO_REPORTS_MD_PATH_COPY=$TRACER_REPORTS_REPO_PATH/automated/currency/go/report_copy.md
GO_REPORTS_MD_PATH_TMP=$TRACER_REPORTS_REPO_PATH/automated/currency/go/report_temp.md
OUTPUT_TO_SLACK=$GO_TRACER_REPO_PATH/output.txt

# Set this to true if you need logs
DEBUG_LOG=false

# Copy the original file
cp $GO_REPORTS_MD_PATH $GO_REPORTS_MD_PATH_COPY
rm $OUTPUT_TO_SLACK

skip_execution=true
first_line=true
while IFS= read -r line; do

    if [[ "$first_line" = true ]]; then
        first_line=false
        changed_line="#### Note: This page is auto-generated. Any change will be overwritten after the next sync. For more details on Go Tracer SDK, please visit our [Github](https://github.com/instana/go-sensor) page. Last updated on $(date +'%d.%m.%Y')."
        # Update the markdown report file
        awk -v new_line="$changed_line" '{ if ($0 == old_line) print new_line; else print }' old_line="$line" $GO_REPORTS_MD_PATH > $GO_REPORTS_MD_PATH_TMP && mv $GO_REPORTS_MD_PATH_TMP $GO_REPORTS_MD_PATH
        continue
    fi

    # For skipping first few lines from the md file.
    if [[ "$skip_execution" = true ]]; then
        if [[ "$line" = "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |" ]]; then
            skip_execution=false
        fi
        continue
    fi

    extract_info_from_markdown_line "$line"


    if [[ "$IS_STANDARD_LIBRARY" = "true" ]] || [[ "$IS_DEPRECATED" = "true" ]]; then
        # skip execution
        continue
    fi
    
    folder=$GO_TRACER_REPO_PATH/$LOCAL_PATH

    cd $folder
    # Find the latest version of the instrumented package
    find_latest_version "$TARGET_PKG_URL"
    find_current_version "$TARGET_PKG_URL"

    # Replace supported, latest version and latest version published date in the markdown line
    changed_line=$(echo "$line" | awk -v new_val="$(printf ' %s ' "$CURRENT_VERSION")" 'BEGIN{OFS=FS="|"} {$5=new_val} 1')
    changed_line=$(echo "$changed_line" | awk -v new_val="$(printf ' %s ' "$LATEST_VERSION")" 'BEGIN{OFS=FS="|"} {$6=new_val} 1')
    changed_line=$(echo "$changed_line" | awk -v new_val="$(printf ' %s ' "$LATEST_VERSION_DATE")" 'BEGIN{OFS=FS="|"} {$7=new_val} 1')

    if [[ "$LATEST_VERSION" != "$CURRENT_VERSION" ]]; then
        find_immediate_next_version "$TARGET_PKG_URL" "$CURRENT_VERSION"

        debug_log "Latest version:" "$LATEST_VERSION"
        debug_log "Current version:" "$CURRENT_VERSION"
        debug_log "Immediate next version:" "$IMMEDIATE_NEXT_VERSION"
        debug_log "Immediate next version date:" "$IMMEDIATE_NEXT_VERSION_DATE"

        echo "$TARGET_PACKAGE_NAME - $CURRENT_VERSION - $LATEST_VERSION - $IMMEDIATE_NEXT_VERSION" >> $OUTPUT_TO_SLACK

        find_days_behind_last_support "$IMMEDIATE_NEXT_VERSION_DATE"

        debug_log "$DAYS_BEHIND"

        changed_line=$(echo "$changed_line" | awk -v new_val="$(printf ' %s day/s ' "$DAYS_BEHIND")" 'BEGIN{OFS=FS="|"} {$8=new_val} 1')
        changed_line=$(echo "$changed_line" | awk -v new_val=" No " 'BEGIN{OFS=FS="|"} {$9=new_val} 1')

        # If the latest version and the immediate next version differ, indicating that support has been pending for a long time, add a note in the notes section to specify the immediate next version
        if [[ "$LATEST_VERSION" != "$IMMEDIATE_NEXT_VERSION" ]]; then
            changed_line=$(echo "$changed_line" | awk -v new_val="$(printf ' Support is pending from v%s onwards. ' "$IMMEDIATE_NEXT_VERSION")" 'BEGIN{OFS=FS="|"} {$13=new_val} 1')
        else
            changed_line=$(echo "$changed_line" | awk -v new_val=" " 'BEGIN{OFS=FS="|"} {$13=new_val} 1')
        fi
    else
        changed_line=$(echo "$changed_line" | awk -v new_val=" 0 day/s " 'BEGIN{OFS=FS="|"} {$8=new_val} 1')
        changed_line=$(echo "$changed_line" | awk -v new_val=" Yes " 'BEGIN{OFS=FS="|"} {$9=new_val} 1')
        changed_line=$(echo "$changed_line" | awk -v new_val=" " 'BEGIN{OFS=FS="|"} {$13=new_val} 1')
    fi

    # Update the markdown report file
    awk -v new_line="$changed_line" '{ if ($0 == old_line) print new_line; else print }' old_line="$line" $GO_REPORTS_MD_PATH > $GO_REPORTS_MD_PATH_TMP && mv $GO_REPORTS_MD_PATH_TMP $GO_REPORTS_MD_PATH

    cd $GO_TRACER_REPO_PATH
done < "$GO_REPORTS_MD_PATH_COPY"

rm $GO_REPORTS_MD_PATH_COPY
