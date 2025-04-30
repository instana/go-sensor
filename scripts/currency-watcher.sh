#!/bin/bash

# ==== Configuration ====
MY_REPO="instana/go-sensor"
#LOCAL_REPO_PATH="/path/to/your/local/repo"
REPOS=( "rabbitmq/amqp091-go" "aws/aws-sdk-go" "aws/aws-sdk-go-v2" "beego/beego" "Azure/azure-sdk-for-go" \
        "labstack/echo" "valyala/fasthttp" "gofiber/fiber" "gin-gonic/gin" "couchbase/gocb" "go-gorm/gorm" \
        "graphql-go/graphql" "grpc/grpc-go" "julienschmidt/httprouter" "aws/aws-lambda-go" "sirupsen/logrus" \
        "mongodb/mongo-go-driver" "gorilla/mux" "jackc/pgx" "gomodule/redigo" "go-redis/redis" "IBM/sarama")
INSTRUMENTATIONS=("instaamqp091" "instaawssdk" "instaawsv2" "instabeego" "instaazurefunction" \
                  "instaecho" "instafasthttp" "instafiber" "instagin" "instagocb" "instagorm" \
                  "instagraphql" "instagrpc" "instahttprouter" "instalambda" "instalogrus" \
                  "instamongo" "instamux" "instapgx" "instaredigo" "instaredis" "instasarama")
NUM_PRS=100
DRY_RUN=false  # Set to false to actually create PRs


# ==== Function: Get top N version titles from RSS ====
get_versions_from_rss() {
  local repo=$1
    local num_tags=${2:-3}
    curl -s "https://github.com/$repo/releases.atom" |
    sed -n 's:.*<title>\(.*\)</title>.*:\1:p' | sed -n '2,'$((num_tags+1))'p'
}

# ==== Function: Get PR titles from our repo ====
  get_pr_titles() {
      curl -s -H "Authorization: token $GH_TOKEN" \
          "https://api.github.com/repos/$MY_REPO/pulls?state=all&sort=created&direction=desc&per_page=$NUM_PRS" |
      jq -r '.[].title'
  }

# ==== Function: Notify Slack ====
notify_slack() {
    local repo=$1
    local version=$2
    local pr_url=$3
    local message="🚀 *New Release detected for $repo v$version*\nRepository: \`$repo\`\nVersion: \`$version\`\nPR Created: <$pr_url|View PR>"

    curl -X POST https://slack.com/api/chat.postMessage \
             -H "Authorization: Bearer $SLACK_TOKEN" \
             -H "Content-type: application/json" \
             --data "{\"channel\": \"$SLACK_CHANNEL_ID\",\"text\": \"$message\" }"
}

# ==== Function: Create PR for untracked release ====
create_pr_for_untracked_release() {
    local repo=$1
    local version=$2
    local instrumentation=$3
    local repo_name=$(dirname "$repo")
    local branch_name
    branch_name="$(date +%Y%m%d)-$repo_name-$version"

    echo "[INFO] Preparing PR for $repo version $version..."

    #cd "$LOCAL_REPO_PATH" || { echo "[ERROR] Repo path not found!"; return 1; }

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

    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] Would push branch $branch_name and create PR."
        echo "[DRY RUN] Would send Slack message for $repo version $version"
    else
        git push origin "$branch_name"
        CURRENT_TIME_UNIX=$(date '+%s')
        pr_url=$(gh pr create --title "feat(currency): updated instrumentation of $instrumentation for new version v$version. Id: $CURRENT_TIME_UNIX" \
                              --body "Auto-created PR to track version $version from $repo" \
                              --base main --head "$branch_name")
        echo "[CREATED] $repo: version $version — PR URL: $pr_url"
        git checkout main
        git branch -D "$branch_name"
        notify_slack "$repo" "$version" "$pr_url"
    fi
}

# ==== Main Script ====
echo "[INFO] Checking repositories for untracked versions..."
PR_TITLES=$(get_pr_titles)

#echo $PR_TITLES >> pr_data.txt
for i in "${!REPOS[@]}"; do
    repo="${REPOS[$i]}"
    instrumentation="${INSTRUMENTATIONS[$i]}"
    echo "++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++"
    echo "[INFO] Processing: $repo for instrumentation: $instrumentation..."
    versions=$(get_versions_from_rss "$repo")

    while IFS= read -r version; do
        #echo "-------------------------------------------------------------------------------"
        #echo "[INFO] version extracted: $version"
        version_clean=$(echo "$version" | grep -oE '[vV]?[0-9]+\.[0-9]+\.[0-9]+' | grep -vE '^[12][0-9]{3}\.[01]?[0-9]\.[0-9]{2}$' | sed 's/^v//' | head -n1)
        #echo "[INFO] cleaned version: $version_clean"

        [[ -z "$version_clean" ]] && continue

        pattern="^feat\(currency\): updated instrumentation of ${instrumentation}[a-zA-Z0-9_\/-]* for new version v${version_clean}\. Id: [a-zA-Z0-9_-]+$"

        if echo "$PR_TITLES" | grep -qE "$pattern"; then
            pr=$(echo "$PR_TITLES" | grep -oE "$pattern")
            echo "Matched pull request: [$pr]"
            echo "[INFO] The latest version has been instrumented. Aborting the check for lower versions"
            break
        else
            echo "[INFO] There is a need to create a PR for $version_clean"
            create_pr_for_untracked_release "$repo" "$version_clean" "$instrumentation"
            break
        fi
    done <<< "$versions"
done
