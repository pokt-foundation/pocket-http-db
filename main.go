package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pokt-foundation/pocket-http-db/cache"
	"github.com/pokt-foundation/pocket-http-db/router"
	postgresdriver "github.com/pokt-foundation/portal-db/v2/postgres-driver"
	"github.com/pokt-foundation/portal-db/v2/types"
	"github.com/pokt-foundation/utils-go/environment"
	"github.com/sirupsen/logrus"
)

const (
	// App env determines type of DB connection
	appEnv = "APP_ENV"

	// DB vars - Required for all envs
	pgUser     = "PG_USER"
	pgPassword = "PG_PASSWORD"
	pgDatabase = "PG_DATABASE"
	// DB vars - Required for development/test env
	pgHost = "PG_HOST"
	pgPort = "PG_PORT"
	// DB vars - Required for production env
	dbInstanceConnectionName = "DB_INSTANCE_CONNECTION_NAME"
	dbPrivateIP              = "DB_PRIVATE_IP"
	// PHD config vars
	apiKeys = "API_KEYS"
	// Optional vars (will default if not set)
	cacheRefresh               = "CACHE_REFRESH"
	defaultCacheRefreshMinutes = 5
	port                       = "PORT"
	defaultPort                = "8080"
)

type (
	options struct {
		appEnv types.AppEnv

		pgHost, pgPort, pgUser, pgPassword, pgDatabase,
		dbInstanceConnectionName, dbPrivateIP string

		apiKeys             map[string]bool
		cacheRefreshMinutes time.Duration
		port                string
	}

	DBConfig interface {
		GetConn() (mainConn, listenerConn *pgx.Conn, err error)
	}

	CloudSQLConfig struct {
		options options
	}
	TestDBConfig struct {
		options options
	}
)

// gatherOptions collects and returns the required environment variables.
func gatherOptions() options {
	appEnv := types.AppEnv(environment.MustGetString(appEnv))
	if !appEnv.IsValid() {
		panic(fmt.Sprintf("invalid APP_ENV: %s", appEnv))
	}

	options := options{
		appEnv:              appEnv,
		pgHost:              environment.MustGetString(pgHost),
		pgPort:              environment.MustGetString(pgPort),
		pgUser:              environment.MustGetString(pgUser),
		apiKeys:             environment.MustGetStringMap(apiKeys, ","),
		cacheRefreshMinutes: time.Duration(environment.GetInt64(cacheRefresh, defaultCacheRefreshMinutes)) * time.Minute,
		port:                environment.GetString(port, defaultPort),
	}

	switch appEnv {
	case types.AppEnvProduction:
		// For CloudSQL DB
		options.dbInstanceConnectionName = environment.MustGetString(dbInstanceConnectionName)
		options.dbPrivateIP = environment.GetString(dbPrivateIP, "")
	default:
		// For local DB
		options.pgPassword = environment.MustGetString(pgPassword)
		options.pgDatabase = environment.MustGetString(pgDatabase)
	}

	return options
}

// CloudSQLConfig.GetConn connects to a GCP CloudSQL instance using the cloudsqlconn lib.
// Intended for production use. Will be used if APP_ENV is 'production'.
func (c *CloudSQLConfig) GetConn() (mainConn, listenerConn *pgx.Conn, err error) {
	driverConfig := postgresdriver.CloudSQLConfig{
		DBUser:                 c.options.pgUser,
		DBPassword:             c.options.pgPassword,
		DBName:                 c.options.pgDatabase,
		InstanceConnectionName: c.options.dbInstanceConnectionName,
		PrivateIP:              c.options.dbPrivateIP,
	}

	mainConn, err = postgresdriver.NewCloudSQLPGXConnection(driverConfig)
	if err != nil {
		return nil, nil, err
	}
	listenerConn, err = postgresdriver.NewCloudSQLPGXConnection(driverConfig)
	if err != nil {
		return nil, nil, err
	}

	return mainConn, listenerConn, nil
}

// TestDBConfig.GetConn connects to a Postgres database using standard connection string and user/PW.
// Intended to be used for running tests on a local Docker container. Will be used if APP_ENV is 'test' or 'development'.
func (c *TestDBConfig) GetConn() (mainConn, listenerConn *pgx.Conn, err error) {
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.options.pgHost,
		c.options.pgPort,
		c.options.pgUser,
		c.options.pgPassword,
		c.options.pgDatabase,
	)

	mainConn, err = postgresdriver.NewPGXConnection(connectionString)
	if err != nil {
		return nil, nil, err
	}
	listenerConn, err = postgresdriver.NewPGXConnection(connectionString)
	if err != nil {
		return nil, nil, err
	}

	return mainConn, listenerConn, nil
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

	// Choose DB configuration based on APP_ENV
	var dbConfig DBConfig
	switch options.appEnv {
	case types.AppEnvProduction:
		// For CloudSQL DB
		dbConfig = &CloudSQLConfig{options: options}
	default:
		// For local DB
		dbConfig = &TestDBConfig{options: options}
	}

	// Connect to Postgres DB, initialize listener and instantiate driver
	mainConn, listenerConn, err := dbConfig.GetConn()
	if err != nil {
		panic(err)
	}

	notificationChannel := make(chan *types.Notification, 32)

	listener := postgresdriver.NewPGXListener(listenerConn, notificationChannel)

	driver, err := postgresdriver.NewPostgresDriver(mainConn, listener, notificationChannel)
	if err != nil {
		panic(err)
	}

	// Pass DB reader to cache and instantiate
	cache, err := cache.NewCache(ctx, cache.Config{
		Reader:              driver,
		Logger:              logger,
		CacheRefreshMinutes: options.cacheRefreshMinutes,
	})
	if err != nil {
		panic(err)
	}

	// Pass cache and DB writer to router and instantiate
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
