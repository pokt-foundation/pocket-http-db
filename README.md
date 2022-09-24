# Pocket HTTP Database

The Pocket HTTP Database performs all database interaction with the Postgres database that is used by the Portal API.

## Pre-Commit Installation

[Full Instructions here](https://pre-commit.com/#install)

Before you can run the pre-commit hooks, you need to have the **pre-commit** package manager installed.

Using pip:

`pip install pre-commit`

Using homebrew:

`brew install pre-commit`

Then, inside this repo, run the following commands to install **pre-commit** and the required Go packages:

```bash
pre-commit install;
pre-commit autoupdate;
go install golang.org/x/tools/cmd/goimports@latest;
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.49.0;
go install -v github.com/go-critic/go-critic/cmd/gocritic@latest
```

Once installed, it's a good idea to run `pre-commit run --all-files` once.

Once this is done, the following commands will be performed on every commit to the repo and must pass before the commit is allowed:

- **go-fmt** - Runs `gofmt`
- **go-imports** - Runs `goimports`
- **golangci-lint** - run `golangci-lint run ./...`
- **go-critic** - run `gocritic check ./...`
- **go-build** - run `go build`
- **go-mod-tidy** - run `go mod tidy -v`
