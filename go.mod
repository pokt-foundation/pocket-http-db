module github.com/pokt-foundation/pocket-http-driver

go 1.18

require github.com/gorilla/mux v1.8.0

require github.com/pokt-foundation/portal-api-go v0.0.1

require (
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/lib/pq v1.10.6 // indirect
)

replace github.com/pokt-foundation/portal-api-go v0.0.1 => ../portal-api-go
