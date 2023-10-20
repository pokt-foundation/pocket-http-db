# This Dockerfile used to build the pocketfoundation/test-portal-postgres image used for testing PHD
FROM postgres:14.7

COPY ./postgres-driver/sqlc/schema.sql /docker-entrypoint-initdb.d/
COPY ./testdata/seed_test_db.sql /docker-entrypoint-initdb.d/
COPY ./testdata/reset_test_db.sql /scripts/
