package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pokt-foundation/pocket-http-db/cache"
	"github.com/pokt-foundation/pocket-http-db/router"
	postgresdriver "github.com/pokt-foundation/portal-db/v2/postgres-driver"
	"github.com/pokt-foundation/portal-db/v2/types"
	"github.com/pokt-foundation/utils-go/environment"
	"github.com/sirupsen/logrus"
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
	// Image tag used to return Docker image version in health check response body.
	imageTag        = "IMAGE_TAG"
	defaultImageTag = "development"

	// TODO: add a map to take strings and converts to their logrus equivalent
	// Valid Values are 0-6 going from fatal to trace
	logLevel = "LOG_LEVEL"
)

type (
	options struct {
		appEnv types.AppEnv

		pgHost, pgPort, pgUser, pgPassword, pgDatabase,
		dbInstanceConnectionName, dbPublicIP, dbPrivateIP string

		apiKeys map[string]bool

		cacheRefreshMinutes time.Duration
		port, imageTag      string

		logLevel int64
	}
	DBConfig interface {
		GetConn(ctx context.Context) (mainConnPool *pgxpool.Pool, err error)
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
	appEnv := types.AppEnv(environment.MustGetString(appEnv))
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
		logLevel:            environment.GetInt64(logLevel, 4), // defaults to INFO
	}

	switch appEnv {
	case types.AppEnvProduction:
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
func (c *CloudSQLConfig) GetConn(ctx context.Context) (mainConnPool *pgxpool.Pool, err error) {
	driverConfig := postgresdriver.CloudSQLConfig{
		DBUser:                 c.options.pgUser,
		DBPassword:             c.options.pgPassword,
		DBName:                 c.options.pgDatabase,
		InstanceConnectionName: c.options.dbInstanceConnectionName,
		PublicIP:               c.options.dbPublicIP,
		PrivateIP:              c.options.dbPrivateIP,
	}

	mainConnPool, err = postgresdriver.NewCloudSQLPGXPool(ctx, driverConfig)
	if err != nil {
		return nil, err
	}

	return mainConnPool, nil
}

// TestDBConfig.GetConn connects to a Postgres database using standard connection string and user/PW.
// Intended to be used for running tests on a local Docker container. Will be used if APP_ENV is 'test' or 'development'.
func (c *TestDBConfig) GetConn(ctx context.Context) (mainConnPool *pgxpool.Pool, err error) {
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.options.pgHost,
		c.options.pgPort,
		c.options.pgUser,
		c.options.pgPassword,
		c.options.pgDatabase,
	)

	mainConnPool, err = postgresdriver.NewPGXPool(ctx, connectionString)
	if err != nil {
		return nil, err
	}

	return mainConnPool, nil
}

// runServer starts the HTTP server and handles shutdown signals.
func runServer(ctx context.Context, router *router.Router, options options, log *logrus.Logger) {
	// Run HTTP server in a goroutine so that it doesn't block the shutdown signal listener.
	server := &http.Server{
		Addr:    ":" + options.port,
		Handler: router.Router,
	}

	go func() {
		log.Printf("Pocket HTTP DB %s running on port: %s", options.imageTag, options.port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Create a channel to listen for an interrupt or termination signal from the system.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block here until we receive the interrupt signal
	<-shutdown

	// The context is used to give the server a deadline for shutting down.
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)

	if err := server.Shutdown(ctxWithTimeout); err != nil {
		log.Fatalf("Server shutdown failed:%+v", err)
	}

	defer cancel()

	log.Println("Server stopped")
}

func main() {
	ctx := context.Background()

	// Gather env vars
	options := gatherOptions()

	// Start pprof server only if in development mode
	if options.appEnv == types.AppEnvDevelopment {
		go func() {
			log.Println("pprof:", http.ListenAndServe("localhost:6060", nil))
		}()
	}

	// Configure logger
	// TODO: use new logger
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.Level(options.logLevel))
	log.Printf("Pocket HTTP DB %s starting...", options.imageTag)

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
	connPool, err := dbConfig.GetConn(ctx)
	if err != nil {
		panic(err)
	}

	notificationChannel := make(chan *types.Notification, 32)

	listener := postgresdriver.NewPGXPoolListener(connPool, notificationChannel)

	driver, errChan, err := postgresdriver.NewPostgresDriver(connPool, listener, notificationChannel)
	if err != nil {
		panic(err)
	}

	// Pass DB reader to cache and instantiate
	cache, err := cache.NewCache(ctx, cache.Config{
		Reader:              driver,
		Logger:              log,
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
		Log:      log,
		ImageTag: options.imageTag,
	})
	if err != nil {
		panic(err)
	}

	// This goroutine continuously monitors the error channel and logs any errors that come in.
	go func() {
		for err := range errChan {
			log.Printf("error from listener: %v", err)
		}
	}()

	// Run the server and handle its lifecycle (including graceful shutdown)
	runServer(ctx, router, options, log)
}
