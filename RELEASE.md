Release Steps
=============

**Never ever update version tags that were already pushed**, this will cause errors on the user side because of different checksum. If a tag has been created accidentally pointing to a wrong commit, create a new patch version with a fix instead of updating it.

1. Ensure tests are passing on [Circle CI](https://app.circleci.com/pipelines/github/instana/go-sensor)
2. From the module directory run `make (minor|patch) release`
   - This would create a new commit updating the module version constant in `version.go`, tag and push it to GitHub.
   - If you have [GitHub CLI](https://cli.github.com/) installed, it will also create a draft release with the changelog
3. Go to [Releases](https://github.com/instana/go-sensor/releases) page, review and publish the draft release created by `make`
