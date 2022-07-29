package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pokt-foundation/pocket-http-db/environment"
	"github.com/pokt-foundation/pocket-http-db/router"
	postgresdriver "github.com/pokt-foundation/portal-api-go/postgres-driver"
)

var (
	port             = environment.GetString("PORT", "8080")
	connectionString = environment.GetString("CONNECTION_STRING", "")
)

func main() {
	driver, err := postgresdriver.NewPostgresDriverFromConnectionString(connectionString)
	if err != nil {
		panic(err)
	}

	router, err := router.NewRouter(driver)
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", router.HealthCheck)
	r.HandleFunc("/blockchain", router.GetBlockchains)
	r.HandleFunc("/blockchain/{id}", router.GetBlockchain)
	r.HandleFunc("/application", router.GetApplications)
	r.HandleFunc("/application/{id}", router.GetApplication)
	r.HandleFunc("/load_balancer", router.GetLoadBalancers)
	r.HandleFunc("/load_balancer/{id}", router.GetLoadBalancer)
	r.HandleFunc("/user", router.GetUsers)
	r.HandleFunc("/user/{id}", router.GetUser)
	r.HandleFunc("/user/{id}/application", router.GetApplicationByUserID)

	http.Handle("/", r)

	log.Printf("Postgres API running in port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
