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

To use the Pocket HTTP DB in other packages, [a Go client is available](https://github.com/pokt-foundation/db-client). This is a lightweight wrapper around the HTTP requests that PHD supports and should allow easy use of PHD from other backend services.

## Modifying Functionality

When adding or updating functionality to PHD, there is a order of operations that should be followed for best results.

**This may appear convoluted but if all these steps are completed we can have a high degree of confidence in any changes.**

1. **Add changes to the [Portal DB driver](https://github.com/pokt-foundation/portal-db) repository:**

- Start by updating the `postgresdrver/sqlc/schema.yml` and/or `postgresdrver/sqlc/query.yml` files
- Generate the SQLC Go code from these SQL files using `make`
- Add new functionality to the postgres driver methods and the Driver interface
- Update the driver integration tests in the `Portal DB` repository to cover all code changes
- Once merged to `main` a new version of the DB driver will be published, following semantic versioning standards based on the commit message(s)

2. **Add the new functionality to PHD:**

- Open a new feature branch in PHD, pointed to the `staging` branch
- Install the newly published `Portal DB driver` version into PHD
- Make any necessary changes to both the `cache.go` and `router.go`
- Update or create unit tests to cover all changes, ensuring they pass _(E2E tests will be added in **step 4**)_
- Push the feature branch to the PHD repository

3. **Add the new functionality to the [DB Client](https://github.com/pokt-foundation/db-client) repository:**

- Add or modify all of the relevant methods inside the `client.go` file and update the interface(s)
- Update the `docker-compose.test.yml` file **(line 9)** inside the `DB Client` repository to point to the PHD feature branch that was just pushed
- Update or create E2E tests inside the `DB Client` to cover all code changes
- As with the `Portal DB`, when these changes are merged to `main` a new version of the `DB Client` will be published

4. **Update PHD E2E tests using new `DB Client` version:**

- Install the newly published `DB Client` version into the PHD feature branch (this is only used for the E2E tests)
- The E2E tests that were written or modified in the `DB Client` in **step 3** may be reused in the main E2E test file
- Following the example of the already existing tests, add a check to confirm the second PHD test container also registered DB updates
- Once the E2E tests are passing the PR may be submitted for review

5. **Finally, update the DB Client `docker-compose.test.yml` file (**line 7**) back to the `master` branch of PHD**

By following these steps, you can ensure all changes are tested at every stage of development, from the database to the PHD API to the DB client used in other packages. This will give you a high degree of confidence in any modifications. 

In addition, the `DB Client` will be up to date and ready for use in any other repos that require these updates.

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
