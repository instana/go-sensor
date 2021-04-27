Release Steps
=============

1. Ensure tests are passing on [Circle CI](https://app.circleci.com/pipelines/github/instana/go-sensor)
2. Make sure your local `master` branch is up-to-date:

   ```bash
   git checkout master && git pull --rebase
   ```
3. If you're about to release a new version of the main module, bump the package version in [`version.go`](./version.go), commit & push the
   version change to the `master` branch
4. Create a git tag for the new version:
   - for the main module the tag should be `v1.X.Y`, where `X` and `Y` correspond to the version in the `version.go`
   - for instrumentation submodules tags should be prefixed with the relative path to the submodule, for example `instrumentation/instalambda/v1.2.0`.
     Sumodule versions are independent from the version of `github.com/instana/go-sensor`
5. Push the tag to GitHub with `git push --tag`

   **Never ever update version tags that were already pushed**, this will cause errors on the user side because of different checksum. If a tag has been
   created accidentally pointing to a wrong commit, create a new patch version with a fix instead of updating it.
7. Go to the [Releases](https://github.com/instana/go-sensor/releases) page and create a new release for the version tag. Use the [appropriate template](#release-templates)
   depending on the release type and module.
8. Make sure that the [Renew pkg.go.dev documentation](https://github.com/instana/go-sensor/actions/workflows/documentation.yml) workflow has been executed. This
   makes [GoDoc](https://pkg.go.dev/instana/go-sensor) aware of the new version.

Release templates
-----------------

* <details>
  <summary>Main module</summary>
  This {minor,patch} release includes the following fixes & improvements:

  * ...
  </details>
* <details>
  <summary>Instrumentation submodule</summary>
  This {minor,patch} release of &lt;submodule name&gt; includes the following fixes & improvements:

  * ...
  </details>
