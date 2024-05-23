package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pokt-foundation/portal-http-db/v2/cache"
	postgresdriver "github.com/pokt-foundation/portal-http-db/v2/postgres-driver"
	"github.com/pokt-foundation/portal-http-db/v2/router"
	"github.com/pokt-foundation/portal-http-db/v2/types"
	"github.com/pokt-foundation/utils-go/environment"
	"github.com/pokt-foundation/utils-go/logger"
)

const (
	// App Env determines type of DB connection.
	appEnv = "APP_ENV"
	// Postgres DB vars - Required for all Envs.
	pgUser     = "PG_USER"
	pgPassword = "PG_PASSWORD"
	pgDatabase = "PG_DATABASE"
	// Local DB vars - Required for development/test Env.
	pgHost = "PG_HOST"
	pgPort = "PG_PORT"
	// CloudSQL DB vars - Required for production Env.
	// May only set one of DB_PUBLIC_IP or DB_PRIVATE_IP; setting both will error.
	dbInstanceConnectionName = "DB_INSTANCE_CONNECTION_NAME"
	dbPublicIP               = "DB_PUBLIC_IP"
	dbPrivateIP              = "DB_PRIVATE_IP"
	// PHD config vars
	apiKeys = "API_KEYS"
	// Optional vars. Will use defaults if not set.
	cacheRefresh               = "CACHE_REFRESH"
	defaultCacheRefreshMinutes = 5
	port                       = "PORT"
	defaultPort                = "8080"
	profiler                   = "PROFILER"
	defaultProfiler            = false
	// Image tag used to return Docker image version in health check response body.
	imageTag        = "IMAGE_TAG"
	defaultImageTag = "development"
)

type (
	// Environment Options
	options struct {
		appEnv types.Environment

		pgHost, pgPort, pgUser, pgPassword, pgDatabase,
		dbInstanceConnectionName, dbPublicIP, dbPrivateIP,
		port, imageTag string

		profiler bool

		apiKeys map[string]bool

		cacheRefreshMinutes time.Duration
	}

	// DB config structs
	DBConfig interface {
		GetConn(ctx context.Context) (mainConnPool *pgxpool.Pool, cleanup func() error, err error)
	}
	CloudSQLConfig struct {
		options
	}
	TestDBConfig struct {
		options
	}
)

// gatherOptions collects and returns the required environment variables.
func gatherOptions() options {
	appEnv := types.Environment(environment.MustGetString(appEnv))
	if !appEnv.IsValid() {
		panic(fmt.Sprintf("invalid APP_ENV: %s", appEnv))
	}

	options := options{
		appEnv:              appEnv,
		pgUser:              environment.MustGetString(pgUser),
		pgPassword:          environment.MustGetString(pgPassword),
		pgDatabase:          environment.MustGetString(pgDatabase),
		apiKeys:             environment.MustGetStringMap(apiKeys, ","),
		cacheRefreshMinutes: time.Duration(environment.GetInt64(cacheRefresh, defaultCacheRefreshMinutes)) * time.Minute,
		port:                environment.GetString(port, defaultPort),
		imageTag:            environment.GetString(imageTag, defaultImageTag),
		profiler:            environment.GetBool(profiler, defaultProfiler),
	}

	switch appEnv {
	case types.EnvProduction:
		// For CloudSQL DB
		options.dbInstanceConnectionName = environment.MustGetString(dbInstanceConnectionName)
		options.dbPublicIP = environment.GetString(dbPublicIP, "")
		options.dbPrivateIP = environment.GetString(dbPrivateIP, "")
	default:
		// For local DB
		options.pgHost = environment.MustGetString(pgHost)
		options.pgPort = environment.MustGetString(pgPort)
	}

	return options
}

// CloudSQLConfig.GetConn connects to a GCP CloudSQL instance using the cloudsqlconn lib.
// Intended for production use. Will be used if APP_ENV is 'production'.
func (c *CloudSQLConfig) GetConn(ctx context.Context) (mainConnPool *pgxpool.Pool, cleanup func() error, err error) {
	driverConfig := postgresdriver.CloudSQLConfig{
		DBUser:                 c.options.pgUser,
		DBPassword:             c.options.pgPassword,
		DBName:                 c.options.pgDatabase,
		InstanceConnectionName: c.options.dbInstanceConnectionName,
		PublicIP:               c.options.dbPublicIP,
		PrivateIP:              c.options.dbPrivateIP,
	}

	mainConnPool, cleanup, err = postgresdriver.NewCloudSQLPGXPool(driverConfig)
	if err != nil {
		return nil, nil, err
	}

	return mainConnPool, cleanup, nil
}

// TestDBConfig.GetConn connects to a Postgres database using standard connection string and user/PW.
// Intended to be used for running tests on a local Docker container. Will be used if APP_ENV is 'test' or 'development'.
func (c *TestDBConfig) GetConn(ctx context.Context) (mainConnPool *pgxpool.Pool, cleanup func() error, err error) {
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.options.pgHost,
		c.options.pgPort,
		c.options.pgUser,
		c.options.pgPassword,
		c.options.pgDatabase,
	)

	mainConnPool, cleanup, err = postgresdriver.NewPGXPool(connectionString)
	if err != nil {
		return nil, nil, err
	}

	return mainConnPool, cleanup, nil
}

// httpHandler starts the HTTP server for the application using the provided router, port, and logger.
func httpHandler(router *router.Router, options options, logger *logger.Logger) {
	logger.Info(fmt.Sprintf("Portal HTTP DB %s running on port: %s", options.imageTag, options.port))
	logger.Info(fmt.Sprintf("Log level is: %s", logger.LogLevel()))
	if options.appEnv != types.EnvProduction {
		logger.Info(fmt.Sprintf("Running in %s mode. Connected to DB at %s:%s", options.appEnv, options.pgHost, options.pgPort))
	}
	// Run the HTTP server
	log.Fatal(http.ListenAndServe(":"+options.port, router.Router))
}

func main() {
	// Gather env vars
	options := gatherOptions()

	// Instantiate logger
	logger := logger.New()

	// Choose DB configuration based on APP_ENV
	var dbConfig DBConfig
	switch options.appEnv {
	case types.EnvProduction:
		// For CloudSQL DB
		dbConfig = &CloudSQLConfig{options: options}
	default:
		// For local DB
		dbConfig = &TestDBConfig{options: options}
	}

	// Connect to Postgres DB, initialize listener and instantiate driver
	connPool, cleanup, err := dbConfig.GetConn(context.Background())
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			logger.Error(fmt.Sprintf("Failed to clean up: %v", err))
		}
	}()

	notificationChannel := make(chan *types.Notification, 32)
	listener := postgresdriver.NewPGXPoolListener(connPool, notificationChannel)
	driver, errChan, err := postgresdriver.NewPostgresDriver(connPool, listener, notificationChannel)
	if err != nil {
		panic(err)
	}

	// Pass DB reader to cache and instantiate
	cache, err := cache.NewCache(context.Background(), cache.Config{
		Reader:              driver,
		Logger:              logger,
		CacheRefreshMinutes: options.cacheRefreshMinutes,
	})
	if err != nil {
		panic(err)
	}

	// Pass cache and DB writer to router and instantiate
	router, err := router.NewRouter(router.Config{
		Cache:    cache,
		Writer:   driver,
		APIKeys:  options.apiKeys,
		Logger:   logger,
		ImageTag: options.imageTag,
	})
	if err != nil {
		panic(err)
	}

	// This goroutine continuously monitors the error channel and logs any errors that come in.
	go func() {
		for err := range errChan {
			logger.Error(fmt.Sprintf("error from listener: %v", err))
		}
	}()

	// Start pprof server only if profiler flag set
	if options.profiler {
		go func() {
			logger.Info(fmt.Sprintf("pprof: %v", http.ListenAndServe("localhost:6060", nil)))
		}()
	}

	// Start the HTTP server
	var wg sync.WaitGroup
	wg.Add(1)
	go httpHandler(router, options, logger)
	wg.Wait()
}
