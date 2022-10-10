test:
	go test ./... -count=1

test_unit:
	go test ./... -short

test_e2e:
	go test ./... -run E2E

test_verbose:
	go test ./... -v

init-pre-commit:
	curl https://github.com/pre-commit/pre-commit/releases/download/v2.20.0/pre-commit-2.20.0.pyz > pre-commit-2.20.0.pyz;
	python3 pre-commit-2.20.0.pyz install;
	python3 pre-commit-2.20.0.pyz autoupdate;
	go install golang.org/x/tools/cmd/goimports@latest;
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest;
	go install -v github.com/go-critic/go-critic/cmd/gocritic@latest;
	python3 pre-commit-2.20.0.pyz run --all-files
	rm pre-commit-2.20.0.pyz;