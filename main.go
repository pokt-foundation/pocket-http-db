package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/pokt-foundation/pocket-http-db/environment"
	"github.com/pokt-foundation/pocket-http-db/router"
	postgresdriver "github.com/pokt-foundation/portal-api-go/postgres-driver"
)

var (
	port             = environment.GetString("PORT", "8080")
	cacheRefresh     = environment.GetInt64("CACHE_REFRESH", 10)
	connectionString = environment.GetString("CONNECTION_STRING", "")
)

func cacheHandler(router *router.Router) {
	for {
		time.Sleep(time.Duration(cacheRefresh) * time.Minute)

		err := router.Cache.SetCache()
		if err != nil {
			fmt.Printf("Cache refresh failed with error: %s", err.Error())
		}
	}
}

func httpHandler(router *router.Router) {
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

	r.Use(router.AuthorizationHandler)

	http.Handle("/", r)

	log.Printf("Postgres API running in port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func main() {
	driver, err := postgresdriver.NewPostgresDriverFromConnectionString(connectionString)
	if err != nil {
		panic(err)
	}

	router, err := router.NewRouter(driver)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)

	go httpHandler(router)
	go cacheHandler(router)

	wg.Wait()
}
