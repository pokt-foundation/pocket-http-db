# Pocket HTTP Database

The Pocket HTTP Database performs all database interaction with the Postgres database that is used by the Portal API.

## Pre-Commit Installation

Run the command `make init-pre-commit` from the repository root.

Once this is done, the following commands will be performed on every commit to the repo and must pass before the commit is allowed:

- **go-fmt** - Runs `gofmt`
- **go-imports** - Runs `goimports`
- **golangci-lint** - run `golangci-lint run ./...`
- **go-critic** - run `gocritic check ./...`
- **go-build** - run `go build`
- **go-mod-tidy** - run `go mod tidy -v`
