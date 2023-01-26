package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lib/pq"
	v0router "github.com/pokt-foundation/pocket-http-db/router"
	v1router "github.com/pokt-foundation/pocket-http-db/v1_router"
	v0postgresdriver "github.com/pokt-foundation/portal-api-go/postgres-driver"
	v1postgresdriver "github.com/pokt-foundation/portal-db/postgres-driver"
	"github.com/pokt-foundation/utils-go/environment"
	"github.com/sirupsen/logrus"
)

const (
	connectionString = "CONNECTION_STRING"
	apiKeys          = "API_KEYS"
	cacheRefresh     = "CACHE_REFRESH"
	port             = "PORT"

	defaultCacheRefreshMinutes = 10
	defaultPort                = "8080"
)

type options struct {
	connectionString string
	apiKeys          map[string]bool
	cacheRefresh     int64
	port             string
}

func gatherOptions() options {
	return options{
		connectionString: environment.MustGetString(connectionString),
		apiKeys:          environment.MustGetStringMap(apiKeys, ","),
		cacheRefresh:     environment.GetInt64(cacheRefresh, defaultCacheRefreshMinutes),
		port:             environment.GetString(port, defaultPort),
	}
}

/* DEPRECATED -- v0 cache handler using portal-api-go driver */
func v0CacheHandler(router *v0router.Router, cacheRefresh int64, log *logrus.Logger) {
	for {
		time.Sleep(time.Duration(cacheRefresh) * time.Minute)

		err := router.Cache.SetCache()
		if err != nil {
			log.WithFields(logrus.Fields{"err": err.Error()}).Error(err)
		}
	}
}

/* NEW -- v1 cache handler using portal-db driver */
func v1CacheHandler(router *v1router.Router, cacheRefresh int64, log *logrus.Logger) {
	for {
		time.Sleep(time.Duration(cacheRefresh) * time.Minute)

		err := router.Cache.SetCache()
		if err != nil {
			log.WithFields(logrus.Fields{"err": err.Error()}).Error(err)
		}
	}
}

func httpHandler(v0Router *v0router.Router, v1Router *v1router.Router, port string, log *logrus.Logger) {
	// TODO - remove this line and replace v0Router.Router with router.Router
	// inside ListenAndServe below once v0 routes are ready to be removed
	v0Router.Router.PathPrefix("/v1").Handler(v1Router.Router)

	log.Printf("Postgres API v0/v1 running in port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, v0Router.Router))
}

func main() {
	log := logrus.New()
	// log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&logrus.JSONFormatter{})

	options := gatherOptions()

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Printf("Problem with listener, error: %s, event type: %d", err.Error(), ev)
		}
	}

	/* DEPRECATED -- v0 DB connection, listener and router using portal-api-go driver */
	v0Listener := pq.NewListener(options.connectionString, 10*time.Second, time.Minute, reportProblem)

	v0Driver, err := v0postgresdriver.NewPostgresDriverFromConnectionString(options.connectionString, v0Listener)
	if err != nil {
		panic(err)
	}

	v0Router, err := v0router.NewRouter(v0Driver, v0Driver, options.apiKeys, log)
	if err != nil {
		panic(err)
	}

	/* NEW -- v1 DB connection, listener and router using portal-db driver */
	listener := pq.NewListener(options.connectionString, 10*time.Second, time.Minute, reportProblem)

	v1Driver, err := v1postgresdriver.NewPostgresDriver(options.connectionString, listener)
	if err != nil {
		panic(err)
	}

	router, err := v1router.NewRouter(v1Driver, v1Driver, options.apiKeys, log)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)

	go httpHandler(v0Router, router, options.port, log)

	/* DEPRECATED -- cache handler using portal-api-go driver */
	go v0CacheHandler(v0Router, options.cacheRefresh, log)

	/* NEW -- cache handler using portal-db driver */
	go v1CacheHandler(router, options.cacheRefresh, log)

	wg.Wait()
}
