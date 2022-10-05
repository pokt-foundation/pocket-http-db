run:
	make build
	./bin/pocket-http-db

build:
	go build -o bin/pocket-http-db .

t_unit:
	go test ./... -short

t_e2e:
	go test ./... -run E2E

t_all:
	go test ./... 

t_verbose:
	go test ./... -v

init-pre-commit:
	python3 pre-commit-2.20.0.pyz install;
	python3 pre-commit-2.20.0.pyz autoupdate;
	go install golang.org/x/tools/cmd/goimports@latest;
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0;
	go install -v github.com/go-critic/go-critic/cmd/gocritic@latest;
	python3 pre-commit-2.20.0.pyz run --all-files;