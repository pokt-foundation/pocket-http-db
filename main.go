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

var (
	connectionString = environment.MustGetString("CONNECTION_STRING")
	apiKeys          = environment.MustGetStringMap("API_KEYS", ",")

	cacheRefresh = environment.GetInt64("CACHE_REFRESH", 10)
	port         = environment.GetString("PORT", "8080")

	log = logrus.New()
)

func init() {
	// log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&logrus.JSONFormatter{})
}

func logError(msg string, err error) {
	fields := logrus.Fields{
		"err": err.Error(),
	}

	log.WithFields(fields).Error(err)
}

func cacheHandler(router *router.Router) {
	for {
		time.Sleep(time.Duration(cacheRefresh) * time.Minute)

		err := router.Cache.SetCache()
		if err != nil {
			logError("Cache refresh failed", err)
		}
	}
}

func httpHandler(router *router.Router) {
	http.Handle("/", router.Router)

	log.Printf("Postgres API running in port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func main() {
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Printf("Problem with listener, error: %s, event type: %d", err.Error(), ev)
		}
	}

	listener := pq.NewListener(connectionString, 10*time.Second, time.Minute, reportProblem)

	driver, err := postgresdriver.NewPostgresDriverFromConnectionString(connectionString, listener)
	if err != nil {
		panic(err)
	}

	router, err := router.NewRouter(driver, driver, apiKeys, log)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)

	go httpHandler(router)
	go cacheHandler(router)

	wg.Wait()
}
