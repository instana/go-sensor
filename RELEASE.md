Release Steps
=============

**Never ever update version tags that were already pushed**, this will cause errors on the user side because of different checksum. If a tag has been created accidentally pointing to a wrong commit, create a new patch version with a fix instead of updating it.

1. Ensure tests are passing on [Circle CI](https://app.circleci.com/pipelines/github/instana/go-sensor)
2. From the module directory run `make (minor|patch) release`
   - This would create a new commit updating the module version constant in `version.go`, tag and push it to GitHub.
   - If you have [GitHub CLI](https://cli.github.com/) installed, it will also create a draft release with the changelog
3. Go to [Releases](https://github.com/instana/go-sensor/releases) page, review and publish the draft release created by `make`
4. When releasing a new version of the core module - that is - releasing from the root directory only, make sure to follow the steps in the section below to update all instrumentations to reference to the new core module release.

## Updating instrumentations after a core module release

Releasing a new version of the core module doesn't automatically update and release a new version of each instrumentation, which we will cover in this section.

Follow the steps below to release the updated instrumentations:

1. Once a new version of the core module is released, update the master branch and create a new branch.
1. In the new branch, run `./instrumentations.sh update` to update all instrumentations to the latest core module version.
1. Run `make test` to assure that no instrumentation is broken after the update. If there are any issues, fix them accordingly.
1. Once all tests pass, create a pull request with the changes and get it merged into the master branch.
1. Switch to the master branch and pull the new changes.
1. Make sure to have [gh](https://cli.github.com/) installed.
1. Make sure to be [logged](https://cli.github.com/manual/gh_auth_login) into Github via `gh` command.
1. Run `./instrumentations.sh release` to release every instrumentation.

If everything goes well, you should be able to see the instrumentations released in the [release page](https://github.com/instana/go-sensor/releases).
