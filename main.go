package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/pokt-foundation/pocket-http-db/router"
	v0Router "github.com/pokt-foundation/pocket-http-db/v0_router"
	v0PostgresDriver "github.com/pokt-foundation/portal-api-go/postgres-driver"
	postgresdriver "github.com/pokt-foundation/portal-db/postgres-driver"
	"github.com/pokt-foundation/utils-go/environment"
	"github.com/sirupsen/logrus"

	_ "net/http/pprof"
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
func v0CacheHandler(v0Router *v0Router.Router, cacheRefresh int64, log *logrus.Logger) {
	for {
		time.Sleep(time.Duration(cacheRefresh) * time.Minute)

		err := v0Router.Cache.SetCache()
		if err != nil {
			log.WithFields(logrus.Fields{"err": err.Error()}).Error(err)
		}
	}
}

/* NEW -- v1 cache handler using portal-db driver */
func cacheHandler(router *router.Router, cacheRefresh int64, log *logrus.Logger) {
	for {
		time.Sleep(time.Duration(cacheRefresh) * time.Minute)

		err := router.Cache.SetCache()
		if err != nil {
			log.WithFields(logrus.Fields{"err": err.Error()}).Error(err)
		}
	}
}

func httpHandler(v0Router *v0Router.Router, router *router.Router, port string, log *logrus.Logger) {
	// TODO - remove this line and replace v0Router.Router with router.Router
	// inside ListenAndServe below once v0 routes are ready to be removed
	v0Router.Router.PathPrefix("/v1").Handler(router.Router)

	log.Printf("Postgres API running in port: %s\n", port)
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

	v0Driver, err := v0PostgresDriver.NewPostgresDriverFromConnectionString(options.connectionString, v0Listener)
	if err != nil {
		panic(err)
	}

	v0Router, err := v0Router.NewRouter(v0Driver, v0Driver, options.apiKeys, log)
	if err != nil {
		panic(err)
	}

	/* NEW -- v1 DB connection, listener and router using portal-db driver */
	listener := pq.NewListener(options.connectionString, 10*time.Second, time.Minute, reportProblem)

	driver, err := postgresdriver.NewPostgresDriver(options.connectionString, listener)
	if err != nil {
		panic(err)
	}

	router, err := router.NewRouter(driver, driver, options.apiKeys, log)
	if err != nil {
		panic(err)
	}

	// MEMORY PROFILING
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	var wg sync.WaitGroup

	wg.Add(1)

	go httpHandler(v0Router, router, options.port, log)

	/* DEPRECATED -- cache handler using portal-api-go driver */
	go v0CacheHandler(v0Router, options.cacheRefresh, log)

	/* NEW -- cache handler using portal-db driver */
	go cacheHandler(router, options.cacheRefresh, log)

	wg.Wait()
}
