package main

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/pokt-foundation/pocket-http-db/cache"
	"github.com/pokt-foundation/pocket-http-db/router"
	postgresdriver "github.com/pokt-foundation/portal-db/v2/postgres-driver"
	"github.com/pokt-foundation/portal-db/v2/types"
	"github.com/pokt-foundation/utils-go/environment"
	"github.com/sirupsen/logrus"
)

const (
	pgUser                   = "PG_USER"
	pgPassword               = "PG_PASSWORD"
	pgDatabase               = "PG_DATABASE"
	dbInstanceConnectionName = "DB_INSTANCE_CONNECTION_NAME"
	dbPrivateIP              = "DB_PRIVATE_IP"

	apiKeys = "API_KEYS"

	cacheRefresh               = "CACHE_REFRESH"
	port                       = "PORT"
	defaultCacheRefreshMinutes = 5
	defaultPort                = "8080"
)

type options struct {
	pgUser, pgPassword, pgDatabase,
	dbInstanceConnectionName, dbPrivateIP string
	apiKeys             map[string]bool
	cacheRefreshMinutes time.Duration
	port                string
}

// gatherOptions collects and returns the required environment variables.
func gatherOptions() options {
	return options{
		pgUser:                   environment.MustGetString(pgUser),
		pgPassword:               environment.MustGetString(pgPassword),
		pgDatabase:               environment.MustGetString(pgDatabase),
		dbInstanceConnectionName: environment.MustGetString(dbInstanceConnectionName),
		dbPrivateIP:              environment.GetString(dbPrivateIP, ""),

		apiKeys:             environment.MustGetStringMap(apiKeys, ","),
		cacheRefreshMinutes: time.Duration(environment.GetInt64(cacheRefresh, defaultCacheRefreshMinutes)) * time.Minute,
		port:                environment.GetString(port, defaultPort),
	}
}

// httpHandler starts the HTTP server for the application using the provided router, port, and logger.
func httpHandler(router *router.Router, port string, log *logrus.Logger) {
	log.Printf("Pocket HTTP DB v2 running on port: %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router.Router))
}

func main() {
	// Configure logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	logger.Printf("Pocket HTTP DB v2 starting...")

	// Gather env vars
	options := gatherOptions()

	ctx := context.Background()

	// Connect to Postgres DB
	driverConfig := postgresdriver.CloudSQLConfig{
		DBUser:                 options.pgUser,
		DBPassword:             options.pgPassword,
		DBName:                 options.pgDatabase,
		InstanceConnectionName: options.dbInstanceConnectionName,
		PrivateIP:              options.dbPrivateIP,
	}

	mainConn, err := postgresdriver.NewCloudSQLPGXConnection(driverConfig)
	if err != nil {
		panic(err)
	}
	listenerConn, err := postgresdriver.NewCloudSQLPGXConnection(driverConfig)
	if err != nil {
		panic(err)
	}

	notificationChannel := make(chan *types.Notification, 32)

	listener := postgresdriver.NewPGXListener(listenerConn, notificationChannel)

	driver, err := postgresdriver.NewPostgresDriver(mainConn, listener, notificationChannel)
	if err != nil {
		panic(err)
	}

	// Create the cache and router
	cache, err := cache.NewCache(ctx, cache.Config{
		Reader:              driver,
		Logger:              logger,
		CacheRefreshMinutes: options.cacheRefreshMinutes,
	})
	if err != nil {
		panic(err)
	}

	router, err := router.NewRouter(router.Config{
		Cache:   cache,
		Writer:  driver,
		APIKeys: options.apiKeys,
		Log:     logger,
	})
	if err != nil {
		panic(err)
	}

	// Start the HTTP server
	var wg sync.WaitGroup
	wg.Add(1)
	go httpHandler(router, options.port, logger)
	wg.Wait()
}
