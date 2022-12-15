build:
	CGO_ENABLED=0 GOOS=linux go build -a -o bin/pocket-http-db ./main.go

test_env_up:
	docker-compose -f ./tests/docker-compose.yml up -d --remove-orphans --build
test_env_down:
	docker-compose -f ./tests/docker-compose.yml down --remove-orphans --rmi all -v
run_unit_tests:
	go test ./...  -short
run_e2e_tests:
	-go test ./... -run E2E -count=1;
run_all_tests:
	-go test ./... -count=1;

test_unit: run_unit_tests
test_e2e:  test_env_up run_e2e_tests  test_env_down
test:      test_env_up run_all_tests  test_env_down

init-pre-commit:
	wget https://github.com/pre-commit/pre-commit/releases/download/v2.20.0/pre-commit-2.20.0.pyz;
	python3 pre-commit-2.20.0.pyz install;
	python3 pre-commit-2.20.0.pyz autoupdate;
	go install golang.org/x/tools/cmd/goimports@latest;
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest;
	go install -v github.com/go-critic/go-critic/cmd/gocritic@latest;
	python3 pre-commit-2.20.0.pyz run --all-files;
	rm pre-commit-2.20.0.pyz;
