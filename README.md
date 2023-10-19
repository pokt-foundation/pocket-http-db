<div align="center">
    <img src=".github/grove_logo.png" alt="Portal Network logo" width="600"/>
    <h1>Portal HTTP Database</h1>
    <big>The Portal HTTP DB (PHD) provides access to the Portal database using a RESTful API.</big>
    <div>
       <br/>
        <a href="https://github.com/pokt-foundation/portal-http-db/pulse"><img src="https://img.shields.io/github/last-commit/pokt-foundation/portal-http-db.svg"/></a>
        <a href="https://github.com/pokt-foundation/portal-http-db/pulls"><img src="https://img.shields.io/github/issues-pr/pokt-foundation/portal-http-db.svg"/></a>
        <a href="https://github.com/pokt-foundation/portal-http-db/issues"><img src="https://img.shields.io/github/issues-closed/pokt-foundation/portal-http-db.svg"/></a>
    </div>
</div>
<br/>

# PHD Client

To use the Portal HTTP DB in other packages, import the `github.com/pokt-foundation/portal-http-db/client` package. 

This is a lightweight wrapper around the HTTP requests that PHD supports and allows easy use of PHD from other backend services.

### PHD Client Usage

This client should be installed in any Go backend repo that needs to interact with PHD. 

The interface to be used depends on the interactions required by the repo:

- `IDBReader`: read-only
- `IDBWrite`: write-only
- `IDBClient`: read & write

### PHD Client Mocks

Mocks of the above interfaces are generated using [Mockery](https://github.com/vektra/mockery). 

To generate them from the interfaces run `make gen_client_mocks`.

# Modifying PHD Functionality

When adding or updating functionality to PHD, there is a order of operations that should be followed for best results.

1. *(Optional)* **Add changes to the [Portal DB Driver](https://github.com/pokt-foundation/portal-db) repository if necessary**

- Start by updating the `postgresdrver/sqlc/schema.yml` and/or `postgresdrver/sqlc/query.yml` files
- Generate the SQLC Go code from these SQL files using `make`
- Add new functionality to the postgres driver methods and the Driver interface
- Update the driver integration tests in the `Portal DB` repository to cover all code changes
- Once merged to `main` a new version of the DB driver will be published, following semantic versioning standards based on the commit message(s)

2. **Add the new functionality to PHD**

- Open a new feature branch in PHD, pointed to the `staging` branch
- Install the newly published `Portal DB driver` version into PHD
- Make any necessary changes to both the `cache.go` and/or `router.go`
- Update or create unit tests in `cache_Test.go` and/or `router_test.go` to cover all changes, ensuring they pass

3. **Update the PHD Client and PHD E2E tests**

- Add or modify all of the relevant methods inside the `client/client.go` file and update the interface
- Run `make gen_client_mocks` to update the PHD Client mocks
- Update or create E2E tests inside `main_e2e_test.go` to cover all code changes

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
- **go-unit-tests** - run `CGO_ENABLED=0 go test ./... -short || fail`

### 3. Detect Secrets

Will detect any potential secrets or sensitive information before allowing a commit.

- Test variables that may resemble secrets (random hex strings, etc.) should be prefixed with `test_`
- The inline comment `pragma: allowlist secret` may be added to a line to force acceptance of a false positive
