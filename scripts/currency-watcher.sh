#!/bin/bash

# (c) Copyright IBM Corp. 2025

# ==== Configuration ====
MY_REPO="instana/go-sensor"
REPOS=( "rabbitmq/amqp091-go" "beego/beego" \
        "labstack/echo" "valyala/fasthttp" "gofiber/fiber" "gin-gonic/gin" "couchbase/gocb" "go-gorm/gorm" \
        "graphql-go/graphql" "grpc/grpc-go" "julienschmidt/httprouter" "aws/aws-lambda-go" "sirupsen/logrus" \
        "mongodb/mongo-go-driver" "gorilla/mux" "jackc/pgx" "gomodule/redigo" "go-redis/redis" "IBM/sarama")
INSTRUMENTATIONS=("instaamqp091" "instabeego" \
                  "instaecho" "instafasthttp" "instafiber" "instagin" "instagocb" "instagorm" \
                  "instagraphql" "instagrpc" "instahttprouter" "instalambda" "instalogrus" \
                  "instamongo" "instamux" "instapgx" "instaredigo" "instaredis" "instasarama")
NUM_PRS=100
DRY_RUN=false  # Set to false to actually create PRs
INFO=true
DEBUG=true


# ==== Function: Get top N version titles from RSS ====
get_versions_from_rss() {
  months=1
  # Compute cutoff date in YYYY-MM-DD format
  CUTOFF_DATE=$(date -v-"$months"m +%Y-%m-%d 2>/dev/null || date --date="-$months months" +%Y-%m-%d)

  repo=$1
  data=$(curl -s "https://github.com/$repo/releases.atom")
  echo "$data" | awk -v cutoff="$CUTOFF_DATE" '

  BEGIN {
    RS="</entry>"
    FS="\n"
    count=0
  }
  {
    title = ""; updated = ""
    for (i = 1; i <= NF; i++) {
      if ($i ~ /<title>/) {
        gsub(/.*<title>|<\/title>.*/, "", $i)
        title = $i
      } else if ($i ~ /<updated>/) {
        gsub(/.*<updated>|<\/updated>.*/, "", $i)
        updated_raw = $i
        split(updated_raw, dt, "T")
        updated = dt[1]
      }
    }

    # Compare dates
    if (title != "" && updated != "") {
      # If updated is older than cutoff, skip
      if (updated < cutoff) {
        next
      }
      # Extract version from title (e.g., v1.9.2)
      version = title
      # Convert updated to dd-mm-yyyy
      split(updated, d, "-")
      formatted_date = d[3] "-" d[2] "-" d[1]
      #print version"::" formatted_date
      print version
    }

    count++
    if (count == 3) exit
  }
  '
}

# ==== Function: Get PR titles from our repo ====
  get_pr_titles() {
      pr_titles=$(curl -s -H "Authorization: token $GH_TOKEN" \
          "https://api.github.com/repos/$MY_REPO/pulls?state=all&sort=created&direction=desc&per_page=$NUM_PRS")
      echo "$pr_titles" | tr -d '\000-\037' | jq -r '.[].title'
  }

# ==== Function: Notify Slack ====
notify_slack() {
    local repo=$1
    local version=$2
    local message="🚀 *New Release detected for $repo v$version*\nRepository: \`$repo\`\nVersion: \`$version\`"
    curl -X POST -H 'Content-type: application/json' --data "{\"text\":\"$message\"}" "$SLACK_HOOK"
}

# ==== Function: Create a PR for untracked release ====
create_pr() {
        local repo=$1
        local version=$2
        local instrumentation=$3
        local repo_name=""
        repo_name=$(dirname "$repo")
        local branch_name=""
        branch_name="$(date +%Y%m%d)-$repo_name-$version"

        # Create a branch and commit the changes
        git config user.name "tracer-team-go"
        git config user.email "github-actions@github.com"

        git remote set-url origin "https://${GITHUB_USERNAME}:${GH_TOKEN}@github.com/instana/go-sensor.git"

        git checkout main
        git pull origin main
        git checkout -b "$branch_name"

        echo "Tracking release $version of $repo" > "release-tracker-${branch_name}.md"
        git add .
        git commit -m "feat(currency): adding support for v$version of $repo"
        git push origin "$branch_name"
        CURRENT_TIME_UNIX=$(date '+%s')
        pr_url=$(gh pr create --title "feat(currency): updated instrumentation of $instrumentation for new version v$version. Id: $CURRENT_TIME_UNIX" \
                              --body "Auto-created PR to track version $version from $repo" \
                              --base main --head "$branch_name")
        [ "$INFO" = "true" ] && echo "[CREATED] $repo: version $version — PR URL: $pr_url"
        git checkout main
        git branch -D "$branch_name"
}

# ==== Function: Create PR and push notifications in Slack for untracked release ====
create_pr_for_untracked_release() {
    local repo=$1
    local version=$2
    local instrumentation=$3

    [ "$INFO" = "true" ] && echo "[INFO] Preparing PR for $repo version $version..."

    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] Would push branch $branch_name and create PR."
        echo "[DRY RUN] Would send Slack message for $repo version $version"
    else
        # Disabling this for the time being as it is not adding much value currently
        #create_pr "$repo" "$version" "$instrumentation"
        notify_slack "$repo" "$version"
    fi
}


# ==== Main Script ====
[ "$INFO" = "true" ] && echo "[INFO] Checking repositories for untracked versions..."
PR_TITLES=$(get_pr_titles)

echo "Last 100 PR titles:"
echo "$PR_TITLES"
