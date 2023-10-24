SHELL := /bin/bash

build:
	CGO_ENABLED=0 GOOS=linux go build -a -o bin/portal-http-db ./main.go

gen:
	@echo "🧪 Generating SQLC code and Mocks ..."
	@make gen_sql gen_mocks 1>/dev/null
	@echo "✅ SQLC and Mocks generated successfully!"

gen_sql:
	sqlc generate -f ./postgres-driver/sqlc/sqlc.yaml
gen_mocks: gen_driver_mocks gen_client gen_reader gen_writer

gen_driver_mocks:
	mockery --name=Driver --filename=mock_driver.go --recursive --inpackage
gen_client:
	mockery --name=IDBClient --filename=mock_client.go --recursive --inpackage
gen_reader:
	mockery --name=IDBReader --filename=mock_reader.go --recursive --inpackage
gen_writer:
	mockery --name=IDBWriter --filename=mock_writer.go --recursive --inpackage

# These targets spin up and shut down the E2E test env in docker.
test_env_up:
	@echo "🧪 Starting up Portal HTTP DB test environment ..."
	@docker-compose -f ./testdata/docker-compose.test.yml up -d --remove-orphans --build
	@echo "⏳ Waiting for test DB to be ready ..."
	@attempts=0; until pg_isready -h localhost -p 5432 -U postgres -d postgres 1>/dev/null || [[ $$attempts -eq 5 ]]; do sleep 2; ((attempts++)); done
	@[[ $$attempts -lt 5 ]] && echo "🐘 Test Portal DB is up ..." || (echo "❌ Test Portal DB failed to start" && make test_env_down 1>/dev/null && exit 1)
	@sleep 2;
	@echo "⏳ Performing health check on portal-http-db-1 ..."
	@attempts=0; until curl -s http://localhost:8080/healthz 1>/dev/null || [[ $$attempts -eq 5 ]]; do sleep 2; ((attempts++)); done
	@[[ $$attempts -lt 5 ]] && echo "🖥️  portal-http-db-1 is online ..." || (echo "❌ portal-http-db-1 failed health check" && make test_env_down 1>/dev/null && exit 1)
	@echo "⏳ Performing health check on portal-http-db-2 ..."
	@attempts=0; until curl -s http://localhost:8081/healthz 1>/dev/null || [[ $$attempts -eq 5 ]]; do sleep 2; ((attempts++)); done
	@[[ $$attempts -lt 5 ]] && echo "🖥️  portal-http-db-2 is online ..." || (echo "❌ portal-http-db-2 failed health check" && make test_env_down 1>/dev/null && exit 1)
	@echo "🚀 Test environment is up!"
test_db_up:
	@echo "🧪 Starting up Portal HTTP DB test database only ..."
	@docker-compose -f ./testdata/docker-compose.test.yml up -d --remove-orphans --build portal-db
	@echo "⏳ Waiting for test DB to be ready ..."
	@attempts=0; until pg_isready -h localhost -p 5432 -U postgres -d postgres 1>/dev/null || [[ $$attempts -eq 5 ]]; do sleep 2; ((attempts++)); done
	@[[ $$attempts -lt 5 ]] && echo "🐘 Test Portal DB is up ..." || (echo "❌ Test Portal DB failed to start" && make test_env_down 1>/dev/null && exit 1)
	@echo "🚀 Test environment is up ..."
test_env_down:
	@echo "🧪 Shutting down Portal HTTP DB test environment ..."
	@docker-compose -f ./testdata/docker-compose.test.yml down --remove-orphans 1>/dev/null
	@echo "✅ Test environment is down."
reset_test_env: stop_test_phd reset_test_db start_test_phd reset_test_cache
reset_test_env_ci: stop_test_phd reset_test_db_ci start_test_phd reset_test_cache

run_all_tests:
	go test ./... -count=1

# These targets run all variations of the test suite.
run_driver_tests:
	@echo "🚗 Running PGDriver tests..."
	@go test ./... -run PGDriver -count=1 || true
	@echo "✅ PGDriver tests completed!"
run_e2e_tests:
	@echo "🌐 Running E2E tests..."
	@go test ./... -run E2E -count=1 || true
	@echo "✅ E2E tests completed!"
test_unit:
	@echo "🛠 Running Unit tests..."
	@go test ./... -short || true
	@echo "✅ Unit tests completed!"

# These targets run all variations of the test suite in CI.
run_driver_tests_ci:
	go test ./... -run PGDriver -count=1
run_e2e_tests_ci:
	go test ./... -run E2E -count=1
test_unit_ci:
	go test ./... -short

# This target runs all tests, unit and E2E, which includes spinning up the Docker test env.
# For local development see below for how to run the E2E tests locally without spinning up the test env each run.
test: test_unit test_env_up run_driver_tests reset_test_env run_e2e_tests test_env_down

test_driver: test_db_up run_driver_tests test_env_down

# These targets are for running E2E tests locally. Once the test env is up and running via `make test_env_up`
# you can run the E2E tests use `make test_dev`. This will reset the test DB and cache before running the tests.
# This allows for running the E2E tests multiple times quickly without having to restart the test env.
reset_test_db_ci:
	@echo "🧨 Resetting the test database ..."
	@docker exec portal-db psql -U postgres -d postgres -f /scripts/reset_test_db.sql >/dev/null || (echo "❌ Database reset operation failed:" && docker exec portal-db psql -U postgres -d postgres -f /docker-entrypoint-initdb.d/reset_test_db.sql)
	@echo "✅ Test database reset operation completed successfully!"
reset_test_db:
	@echo "🧨 Resetting the test database ..."
	@docker exec -it portal-db psql -U postgres -d postgres -f /scripts/reset_test_db.sql >/dev/null || (echo "❌ Database reset operation failed:" && docker exec -it portal-db psql -U postgres -d postgres -f /docker-entrypoint-initdb.d/reset_test_db.sql)
	@echo "✅ Test database reset operation completed successfully!"
reset_test_cache:
	@echo "💡 Resetting the PHD cache ..."
	@sleep 1
	@curl -X POST -H 'authorization: test_api_key_6789' http://localhost:8080/reset_cache >/dev/null 2>&1
	@curl -X POST -H 'authorization: test_api_key_6789' http://localhost:8081/reset_cache >/dev/null 2>&1
	@echo "✅ Test PHD cache reset completed successfully!"
stop_test_phd:
	@echo "⏹️ Stopping the test PHD containers ..."
	@docker-compose -f ./testdata/docker-compose.test.yml stop portal-http-db-1 portal-http-db-2 1>/dev/null
	@echo "✅ Test PHD containers stopped successfully!"
start_test_phd:
	@echo "▶️ Restarting the test PHD containers ..."
	@docker-compose -f ./testdata/docker-compose.test.yml start portal-http-db-1 portal-http-db-2 1>/dev/null
	@echo "✅ Test PHD containers restarted successfully!"

test_dev: reset_test_db run_driver_tests reset_test_env run_e2e_tests
test_driver_dev: rest_test_db run_driver_tests

# This target install pre-commit to the repo and should be run only once, after cloning the repo for the first time.
init-pre-commit:
	wget https://github.com/pre-commit/pre-commit/releases/download/v2.20.0/pre-commit-2.20.0.pyz;
	python3 pre-commit-2.20.0.pyz install;
	go install golang.org/x/tools/cmd/goimports@v0.6.0;
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.0;
	go install -v github.com/go-critic/go-critic/cmd/gocritic@v0.6.5;
	python3 pre-commit-2.20.0.pyz run --all-files;
	rm pre-commit-2.20.0.pyz;
