package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/pokt-foundation/pocket-http-db/cache"
	"github.com/pokt-foundation/pocket-http-db/router"
	postgresdriver "github.com/pokt-foundation/portal-db/v2/postgres-driver"
	"github.com/pokt-foundation/utils-go/environment"
	"github.com/sirupsen/logrus"
)

const (
	connectionString = "CONNECTION_STRING"
	apiKeys          = "API_KEYS"
	cacheRefresh     = "CACHE_REFRESH"
	port             = "PORT"

	defaultCacheRefreshMinutes = 5
	defaultPort                = "8080"
)

type options struct {
	connectionString    string
	apiKeys             map[string]bool
	cacheRefreshMinutes time.Duration
	port                string
}

func gatherOptions() options {
	return options{
		connectionString:    environment.MustGetString(connectionString),
		apiKeys:             environment.MustGetStringMap(apiKeys, ","),
		cacheRefreshMinutes: time.Duration(environment.GetInt64(cacheRefresh, defaultCacheRefreshMinutes)) * time.Minute,
		port:                environment.GetString(port, defaultPort),
	}
}

func httpHandler(router *router.Router, port string, log *logrus.Logger) {
	log.Printf("Postgres API v1 running in port: %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router.Router))
}

func main() {
	logger := logrus.New()
	// log as JSON instead of the default ASCII formatter.
	logger.SetFormatter(&logrus.JSONFormatter{})

	options := gatherOptions()

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Printf("Problem with listener, error: %s, event type: %d", err.Error(), ev)
		}
	}

	listener := pq.NewListener(options.connectionString, 10*time.Second, time.Minute, reportProblem)

	driver, err := postgresdriver.NewPostgresDriver(options.connectionString, listener)
	if err != nil {
		panic(err)
	}

	cache, err := cache.NewCache(driver, options.cacheRefreshMinutes, logger)
	if err != nil {
		panic(err)
	}

	router, err := router.NewRouter(cache, driver, driver, options.apiKeys, logger)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)

	go httpHandler(router, options.port, logger)

	wg.Wait()
}
