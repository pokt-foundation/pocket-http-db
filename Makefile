test_unit:
	go test ./... -short

test_env_up:
	docker-compose -f ./tests/docker-compose.yml up -d --remove-orphans --build

test_env_down:
	docker-compose -f ./tests/docker-compose.yml down --remove-orphans --rmi all -v

test:
	make test_env_up;
	sleep 1;
	-go test ./... -count=1 -v;
	make test_env_down;

test_e2e:
	make test_env_up;
	sleep 1;
	-go test ./... -run E2E -count=1 -v;
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