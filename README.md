<div align="center">
    <img src=".github/grove_logo.png" alt="Portal Network logo" width="600"/>
    <h1>Portal HTTP Database</h1>
    <big>The Portal HTTP DB (PHD) provides access to the Portal database through a REST API.</big>
</div>
<br/>

# Table of Contents
  - [PHD API](#phd-api)
  - [PHD Client](#phd-client)
  - [DB Driver](#db-driver)
    - [Postgres Driver](#2-postgres-driver)
    - [Types Package](#3-types)
  - [Development](#development)
    - [Code Generation](#code-generation)
    - [Modifying PHD Functionality](#modifying-phd-functionality)
    - [Pre-Commit](#pre-commit-installation)

# PHD API

The PHD API is a REST interface over top of a Postgres database connection. It has 3 key components:

## 1. [Cache](https://github.com/pokt-foundation/portal-http-db/blob/master/cache/cache.go)

The `cache` contains all the data contained in the database, in the form of the Go structs used across the backend repos.

The `cache` is set up to refresh at a specified interval (default is 5 minutes).

## 2. [Listener](https://github.com/pokt-foundation/portal-http-db/blob/master/cache/listener.go)

The listener (located in the `cache` package) receives Postgres notification payloads triggered by the Postgres database and updates the cache in response to any writes performed on the DB.

All `cache` updates in PHD happen in response to one of these listener events.

## 3. [Router](https://github.com/pokt-foundation/portal-http-db/blob/master/router/router.go)

The `router` contains all endpoints that may be used for interacting with the Portal DB:

# PHD Client

To use the Portal HTTP DB Client in other backend repos, run:

```go
go get github.com/pokt-foundation/portal-http-db/v2@latest
```

and import the module:

```go
phdClient github.com/pokt-foundation/portal-http-db/v2/phd-client
```

This client is a lightweight wrapper around the HTTP requests that PHD supports and allows easy use of PHD from other backend services.

The interface to be used depends on the interactions required by the repo:

- `IDBReader`: read-only
- `IDBWrite`: write-only
- `IDBClient`: read & write

# DB Driver

### 1. Driver

`github.com/pokt-foundation/portal-http-db/v2/driver`

Contains the following interfaces:

- **Driver**: contains all Read & Write methods.
- **Reader**: contains only Read methods and the Notification channel.
- **Writer**: contains only Write methods.

### 2. Postgres Driver

`github.com/pokt-foundation/portal-http-db/v2/postgresdriver`

Contains all functionality to interact with Postgres.

- Provides a struct that satisfies the Driver interface.
- Type-safe Go code is generated from SQL schema by SQLC.

### 3. Types

`github.com/pokt-foundation/portal-http-db/v2/types`

Contains all database structs and their associated methods which are used across the Portal API backend Go repos.

# Development

## Code Generation
⚠️ IMPORTANT - Before any commit the `make gen` target **MUST** be run. ⚠️

In order for this target to properly generate the Go code and mocks the following dependencies **MUST** be installed:

- [SQLC](https://docs.sqlc.dev/en/stable/overview/install.html)
- [Mockery](https://vektra.github.io/mockery/latest/installation/)

# Modifying PHD Functionality

When adding or updating functionality to PHD, there is a order of operations that must be followed for best results.

1. **Add changes to the `types` package *(if necessary)***

- All structs used throughout the Portal backend are located inside `./types`
- Unit tests for the the `types` package must also be updated

2. **Add changes to the `driver` and/or `postgresdriver` packages *(if necessary)***

- Start by updating the `postgresdrver/sqlc/schema.yml` and/or `postgresdrver/sqlc/query.yml` files
- Generate the SQLC Go code from these SQL files using `make gen`
- Add new functionality to the postgres driver methods and the Driver interface
- Update the driver integration tests in the `postgresdriver` package to cover all code changes

3. **Add the new functionality to PHD `cache` and/or `router` packages**

- Open a new feature branch in PHD, pointed to the `staging` branch
- Install the newly published `Portal DB driver` version into PHD
- Make any necessary changes to both the `cache.go` and/or `router.go`
- Update or create unit tests in `cache_Test.go` and/or `router_test.go` to cover all changes, ensuring they pass

4. **Update the `phd-client` package and PHD E2E tests**

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
