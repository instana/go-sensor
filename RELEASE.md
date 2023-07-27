Release Steps
=============

All release steps rely on Github Actions.
If you do not have permission to run Github actions into this repository, please request it so.


## Release Types

The Instana Go Tracer consists of three distinguished release types:

1. Release of the core module, aka `go-sensor`
1. Release of an instrumented package. Eg: `instrumentation/instagin`
1. Release of all instrumented packages updated for using the latest core

Each of these releases are described below:

### Core Module Release

The core module needs to be released in a couple of cases:

1. When a new feature or improvement is made into the core. This basically means any change made in the root of the project that affects Go files.
1. When a new span is created. This will be needed if a new instrumentation is on the way and the span type is new, and needs to reside in the core module, according to the current design of the tracer.

Steps to release the core module:

1. Go to the [repository actions](https://github.com/instana/go-sensor/actions)
1. Click on [Go Tracer Release](https://github.com/instana/go-sensor/actions/workflows/release.yml)
1. On the right side of the page, click on `Run workflow`
1. Keep the default branch `main`
1. Keep the default package as `.`
1. Select `minor` or `patch` according to the type of version you want to release
1. If you want to review the release and manually release it, keep the checkbox `Release as a draft?`
   a. If you keep it as a draft, you will have to go to the [releases page](https://github.com/instana/go-sensor/releases) and publish the release
   b. If you uncheck the `Release as a draft?` box, the release will take place

### Package Release

An instrumented package needs to be released when a new instrumentation is introduced or if an existing package is updated.

The steps to release an instrumented package is nearly the same as the core module:

1. Go to the [repository actions](https://github.com/instana/go-sensor/actions)
1. Click on [Go Tracer Release](https://github.com/instana/go-sensor/actions/workflows/release.yml)
1. On the right side of the page, click on `Run workflow`
1. Keep the default branch `main`
1. Type the name of the package you want to release. Eg: `instagin`. You only add extra information when you want to release a different major version other than v1. In that case, suppose you want to release v2 of instaredigo, then write `instaredigo/v2`. If an existing package is provided the workflow will fail.
1. Select `major`, `minor` or `patch` according to the type of version you want to release
1. If you want to review the release and manually release it, keep the checkbox `Release as a draft?`
   a. If you keep it as a draft, you will have to go to the releases page and publish the release
   b. If you uncheck the `Release as a draft?` box, the release will take place

> For releases done by the `Go Tracer Release` action that are not drafts, the workflow will automatically post the release in the Slack channel and will update https://pkg.go.dev website with the latest version of the packages
>
> For releases done by the `Go Tracer Release` action that are drafts, both the Slack post and https://pkg.go.dev website update will happen after the release is manually published.

### Updated Packages Release

When a new package is instrumented, it will import the latest core by the time the instrumentation was done. This means that when new versions of the core module is released, this package will be outdated with latest changes from the core.

We want to keep instrumented packages up to date with the latest core module as much as we can, so a Github Action was created for this.

Every time the core module is released via the `Go Tracer Release` action, as a draft or not, it will create a pull request with all the instrumented packages updated to import the new version of the core module.

It is your responsibility to check the success of this pull request and manually fix any potential issues in it.
Once the pull request is successful and reviewed by the team, it can be merged into the main branch.

When the pull request is merged into the main branch an action `sss` will be automatically triggered to release all packages.
It will also update the https://pkg.go.dev website, but it **won't** post each package into the Slack channel
