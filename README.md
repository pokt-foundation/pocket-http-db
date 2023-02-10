<div align="center">
    <img src=".github/banner.png" alt="Pocket Network logo" width="600"/>
    <h1>Pocket HTTP Database</h1>
    <big>The Pocket HTTP DB (PHD) provides access to the Portal database using a RESTful API.</big>
    <div>
       <br/>
        <a href="https://github.com/pokt-foundation/pocket-http-db/pulse"><img src="https://img.shields.io/github/last-commit/pokt-foundation/pocket-http-db.svg"/></a>
        <a href="https://github.com/pokt-foundation/pocket-http-db/pulls"><img src="https://img.shields.io/github/issues-pr/pokt-foundation/pocket-http-db.svg"/></a>
        <a href="https://github.com/pokt-foundation/pocket-http-db/issues"><img src="https://img.shields.io/github/issues-closed/pokt-foundation/pocket-http-db.svg"/></a>
    </div>
</div>
<br/>

# Development

## DB Client

To use the Pocket HTTP DB in other packages, [a Go client is available](https://github.com/pokt-foundation/db-client). This is a lightweight wrapper around the HTTP requests that PHd supports and should allow easy use of PHD from other backend services.

## Adding New Functionality

When adding new functionality, there is a order of operations that should be followed for best results.

**This may seem convoluted but if all these steps are completed we can have an extremely high degree of confidence in any modifications as they are all fully tested at every stage of development from the database to the PHD API to the DB client used in other packages.**

1. Add new functionality to the [Portal DB](https://github.com/pokt-foundation/db-client) driver repository. This should start with the `sqlc/schema.yml` and/or `sqlc/query.yml` files, followed by generating the Go code using `make` and moving to the postgres driver methods (as well as updating the `Driver` interface). Finally, the tests in the `Portal DB` repo should be updated to cover any and all new changes. Once merged to `main` a new version of the DB driver will be published, following semantic versioning standards based on the commit message(s).

2. Open a new feature branch in PHD pointed to the `staging` branch. The newly published `Portal DB` driver version should be installed into PHD and all necessary changes made to both the `cache` and `router`. At this stage, unit tests should be updated or created to cover any and all changes and be passing; E2E tests will be added in step 4. At this point, the feature branch should be pushed to the PHD repo.

3. The new functionality should be added to the [DB Client](https://github.com/pokt-foundation/db-client). The `docker-compose.test.yml` file inside this repo should be updated on `line 9` to point to the PHD feature branch that was just pushed to PHD. Once this is done, the E2E tests inside the `DB Client` should be updated or created to cover any and all changes. As with the `Portal DB`, when these changes are merged to `main` a new version of the `DB Client` will be published.

4. The newly published `DB Client` version should be installed into PHD and the E2E tests that were written or modified in the DB driver in step 3 may be reused in the main E2E test file for PHD (with the caveat that we add an additional check for the second test PHD instance that runs in the PHD E2E; follow the examples that already exist in the test file). Once the E2E tests pass with the new changes the PR can be submitted for review.

5. Once the PR has been merged to PHD `master`, update `line 7` of the `DB Client` `docker-compose.test.yml` file back to `master` branch.

5. Profit.

Again, this may seem like a lot of steps but if they are all followed then by the time the changes reach `staging` we can be very confident they work and as a side benefit the `DB Client` is ready to be used in all services.

When adding or modifying routes, the corresponding method should be created or updated in the DB client. Once the E2E test for the new or updated method in the DB client is completed, the latest DB client  version can be installed in PHD and the E2E test used to verify the changes in PHD as well.

## Pre-Commit Installation

Before starting development work on this repo, `pre-commit` must be installed.

In order to do so, run the command **`make init-pre-commit`** from the repository root.

Once this is done, the following checks will be performed on every commit to the repo and must pass before the commit is allowed:

### 1. Basic checks

- **check-yaml** - Checks YAML files for errors
- **check-merge-conflict** - Ensures there are no merge conflict markers
- **end-of-file-fixer** - Adds a newline to end of files
- **trailing-whitespace** - Trims trailing whitespace
- **no-commit-to-branch** - Ensures commits are not made directly to `main`

### 2. Go-specific checks

- **go-fmt** - Runs `gofmt`
- **go-imports** - Runs `goimports`
- **golangci-lint** - run `golangci-lint run ./...`
- **go-critic** - run `gocritic check ./...`
- **go-build** - run `go build`
- **go-mod-tidy** - run `go mod tidy -v`

### 3. Detect Secrets

Will detect any potential secrets or sensitive information before allowing a commit.

- Test variables that may resemble secrets (random hex strings, etc.) should be prefixed with `test_`
- The inline comment `pragma: allowlist secret` may be added to a line to force acceptance of a false positive
