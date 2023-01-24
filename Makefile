build:
	CGO_ENABLED=0 GOOS=linux go build -a -o bin/pocket-http-db ./main.go

test_env_up:
	docker-compose -f ./docker-compose.test.yml up -d --remove-orphans --build;
	sleep 5;
test_env_down:
	docker-compose -f ./docker-compose.test.yml down --remove-orphans -v

# v0 tests - TODO remove when v0 dropped
run_e2e_tests_v0:
	-go test -run V0_E2E -count=1 -v;
run_all_tests_v0:
	-go test ./... -run V0 -count=1;
test_unit_v0:
	go test ./...  -run V0 -short -count=1
test_e2e_v0: test_env_up run_e2e_tests_v0 test_env_down
test_v0:     test_env_up run_all_tests_v0 test_env_down

# v1 tests
run_e2e_tests:
	-go test -run V1_E2E -count=1 -v;
run_all_tests:
	-go test ./... -run V1 -count=1;
test_unit:
	go test ./...  -run V1 -short -count=1
test_e2e: test_env_up run_e2e_tests test_env_down
# TODO replace `make test` target once v0 dropped
# test: test_env_up run_all_tests test_env_down
test_v1:  test_env_up run_all_tests test_env_down

# all tests - TODO remove when v0 dropped
test:
	make test_env_up;
	make run_all_tests_v0;
	make test_env_down;
	sleep 2;
	make test_env_up;
	make run_all_tests;
	make test_env_down;

init-pre-commit:
	wget https://github.com/pre-commit/pre-commit/releases/download/v2.20.0/pre-commit-2.20.0.pyz;
	python3 pre-commit-2.20.0.pyz install;
	python3 pre-commit-2.20.0.pyz autoupdate;
	go install golang.org/x/tools/cmd/goimports@latest;
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest;
	go install -v github.com/go-critic/go-critic/cmd/gocritic@latest;
	python3 pre-commit-2.20.0.pyz run --all-files;
	rm pre-commit-2.20.0.pyz;
