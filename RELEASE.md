Release Steps
=============

**Never ever update version tags that were already pushed**, this will cause errors on the user side because of different checksum. If a tag has been created accidentally pointing to a wrong commit, create a new patch version with a fix instead of updating it.

1. Make sure to checkout the master branch and pull the latest changes
2. Make sure to have the [GitHub CLI](https://cli.github.com/) installed and that you are logged in
3. Ensure that tests are passing on [Circle CI](https://app.circleci.com/pipelines/github/instana/go-sensor)
4. From the module directory run `make (minor|patch) release`. If you want to release the core module, run the command from the root directory
   - This creates a new commit updating the module version constant in `version.go`, tag and push it to GitHub.
   - If you are properly logged into Github via [GitHub CLI](https://cli.github.com/), it will create a draft release with the changelog
5. Go to the [Releases](https://github.com/instana/go-sensor/releases) page, review and publish the draft release created by `make`
6. When releasing a new version of the core module - that is - releasing from the root directory, make sure to follow the steps in the section below to update all instrumentations to reference the new core module release.

## Updating instrumentations after a core module release

Releasing a new version of the core module doesn't automatically update and release a new version of each instrumentation, which we will cover in this section.

Follow the steps below to release the updated instrumentations:

1. Make sure to have the [GitHub CLI](https://cli.github.com/) installed and that you are logged in.
1. Once a new version of the core module is released, update the master branch and create a new branch.
1. Make sure that the latest core module release is updated in the [Go Package manager](https://pkg.go.dev/github.com/instana/go-sensor). If not, manually update it until it shows up.
1. In the new branch, run `./instrumentations.sh update` to update all instrumentations to the latest core module version.
1. Run `make test` to assure that no instrumentation is broken after the update. If there are any issues, fix them accordingly.
1. Run `make integration` to assure that no instrumentation is broken after the update. If there are any issues, fix them accordingly.
1. Once all tests pass, create a pull request with the changes and get it merged into the master branch.
1. Switch to the master branch and pull the new changes.
1. Run `./instrumentations.sh release` to release every instrumentation.

If everything goes well, you should be able to see the instrumentations released in the [release page](https://github.com/instana/go-sensor/releases).
