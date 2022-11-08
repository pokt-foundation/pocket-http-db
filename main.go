package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/pokt-foundation/pocket-http-db/router"
	postgresdriver "github.com/pokt-foundation/portal-api-go/postgres-driver"
	"github.com/pokt-foundation/utils-go/environment"
	"github.com/sirupsen/logrus"
)

const (
	CONNECTION_STRING = "CONNECTION_STRING"
	API_KEYS          = "API_KEYS"
	CACHE_REFRESH     = "CACHE_REFRESH"
	PORT              = "PORT"

	CACHE_REFRESH_DEFAULT_MINUTES = 10
	DEFAULT_PORT                  = "8080"
)

type options struct {
	connectionString string
	apiKeys          map[string]bool
	cacheRefresh     int64
	port             string
}

func gatherOptions() options {
	return options{
		connectionString: environment.MustGetString(CONNECTION_STRING),
		apiKeys:          environment.MustGetStringMap(API_KEYS, ","),
		cacheRefresh:     environment.GetInt64(CACHE_REFRESH, CACHE_REFRESH_DEFAULT_MINUTES),
		port:             environment.GetString(PORT, DEFAULT_PORT),
	}
}

func cacheHandler(router *router.Router, cacheRefresh int64, log *logrus.Logger) {
	for {
		time.Sleep(time.Duration(cacheRefresh) * time.Minute)

		err := router.Cache.SetCache()
		if err != nil {
			log.WithFields(logrus.Fields{"err": err.Error()}).Error(err)
		}
	}
}

func httpHandler(router *router.Router, port string, log *logrus.Logger) {
	http.Handle("/", router.Router)

	log.Printf("Postgres API running in port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
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

	listener := pq.NewListener(options.connectionString, 10*time.Second, time.Minute, reportProblem)

	driver, err := postgresdriver.NewPostgresDriverFromConnectionString(options.connectionString, listener)
	if err != nil {
		panic(err)
	}

	router, err := router.NewRouter(driver, driver, options.apiKeys, log)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)

	go httpHandler(router, options.port, log)
	go cacheHandler(router, options.cacheRefresh, log)

	wg.Wait()
}
