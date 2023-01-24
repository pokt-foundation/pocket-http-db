build:
	CGO_ENABLED=0 GOOS=linux go build -a -o bin/pocket-http-db ./main.go

test_env_up:
<<<<<<< HEAD
	docker-compose -f ./docker-compose.test.yml up -d --remove-orphans --build;
=======
	docker-compose -f ./tests/docker-compose.yml up -d --remove-orphans --build;
>>>>>>> 0947908 (Yelp detect secrets (#89))
	sleep 2;
test_env_down:
	docker-compose -f ./docker-compose.test.yml down --remove-orphans -v

run_e2e_tests_v0:
	-go test -run V0_E2E -count=1 -v;
run_all_tests_v0:
	-go test ./... -run V0 -count=1;
test_unit_v0:
	go test ./...  -run V0 -short -count=1
test_e2e_v0: test_env_up run_e2e_tests_v0 test_env_down
test_v0:     test_env_up run_all_tests_v0 test_env_down

test:
	make test_env_up;
	make run_all_tests_v0;
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
