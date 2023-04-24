package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/pokt-foundation/pocket-http-db/cache"
	"github.com/pokt-foundation/pocket-http-db/router"
	postgresdriver "github.com/pokt-foundation/portal-db/postgres-driver"
	"github.com/pokt-foundation/utils-go/environment"
	"github.com/sirupsen/logrus"
)

const (
	pgHost        = "PG_HOST"
	pgPort        = "PG_PORT"
	pgUser        = "PG_USER"
	pgPassword    = "PG_PASSWORD"
	pgDatabase    = "PG_DATABASE"
	pgSSLRootCert = "PG_SSL_ROOT_CERT"
	pgSSLCert     = "PG_SSL_CERT"
	pgSSLKey      = "PG_SSL_KEY"
	apiKeys       = "API_KEYS"
	cacheRefresh  = "CACHE_REFRESH"
	port          = "PORT"

	defaultCacheRefreshMinutes = 5
	defaultPort                = "8080"
)

type options struct {
	pgHost, pgPort, pgUser, pgPassword, pgDatabase,
	pgSSLRootCert, pgSSLCert, pgSSLKey string
	apiKeys             map[string]bool
	cacheRefreshMinutes time.Duration
	port                string
}

// gatherOptions collects and returns the required environment variables.
func gatherOptions() options {
	return options{
		pgHost:        environment.MustGetString(pgHost),
		pgPort:        environment.MustGetString(pgPort),
		pgUser:        environment.MustGetString(pgUser),
		pgPassword:    environment.MustGetString(pgPassword),
		pgDatabase:    environment.MustGetString(pgDatabase),
		pgSSLRootCert: environment.GetString(pgSSLRootCert, ""),
		pgSSLCert:     environment.GetString(pgSSLCert, ""),
		pgSSLKey:      environment.GetString(pgSSLKey, ""),

		apiKeys:             environment.MustGetStringMap(apiKeys, ","),
		cacheRefreshMinutes: time.Duration(environment.GetInt64(cacheRefresh, defaultCacheRefreshMinutes)) * time.Minute,
		port:                environment.GetString(port, defaultPort),
	}
}

// createTempFile creates a temporary .pem certificate file for connecting to the database using SSL.
func createTempFile(content string, prefix string) (string, error) {
	tmpFile, err := ioutil.TempFile("", prefix+"-*.pem")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	content = strings.ReplaceAll(content, "\\n", "\n")

	_, err = tmpFile.WriteString(content)
	if err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

// buildConnectionString constructs the Postgre connection string using the given options.
// If the SSL env vars are present, it also creates temporary files for the SSL certificates
// and returns a cleanup function that removes the temporary files when called.
func buildConnectionString(options options) (string, func(), error) {
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s",
		options.pgHost,
		options.pgPort,
		options.pgUser,
		options.pgPassword,
		options.pgDatabase,
	)

	cleanup := func() {}

	// Only configure the connection string to use SSL if all the SSL variables are present.
	// This is to allow E2E testing of PHD using a local Docker Postgres DB without SSL.
	if options.pgSSLRootCert != "" && options.pgSSLCert != "" && options.pgSSLKey != "" {
		sslRootCertPath, err := createTempFile(options.pgSSLRootCert, "sslRootCert")
		if err != nil {
			return "", nil, err
		}
		sslCertPath, err := createTempFile(options.pgSSLCert, "sslCert")
		if err != nil {
			return "", nil, err
		}
		sslKeyPath, err := createTempFile(options.pgSSLKey, "sslKey")
		if err != nil {
			return "", nil, err
		}

		connectionString = fmt.Sprintf("%s sslmode=verify-ca sslrootcert=%s sslcert=%s sslkey=%s",
			connectionString,
			sslRootCertPath,
			sslCertPath,
			sslKeyPath,
		)

		cleanup = func() {
			os.Remove(sslRootCertPath)
			os.Remove(sslCertPath)
			os.Remove(sslKeyPath)
		}
	} else {
		connectionString = fmt.Sprintf("%s sslmode=disable", connectionString)
	}

	return connectionString, cleanup, nil
}

// httpHandler starts the HTTP server for the application using the provided router, port, and logger.
func httpHandler(router *router.Router, port string, log *logrus.Logger) {
	log.Printf("Postgres API v1 running in port: %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router.Router))
}

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	options := gatherOptions()

	// Connect to Postgres DB
	connectionString, cleanupTempCertFiles, err := buildConnectionString(options)
	if err != nil {
		panic(err)
	}

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Printf("Problem with listener, error: %s, event type: %d", err.Error(), ev)
		}
	}

	listener := pq.NewListener(connectionString, 10*time.Second, time.Minute, reportProblem)

	driver, err := postgresdriver.NewPostgresDriver(connectionString, listener)
	if err != nil {
		panic(err)
	}

	// Clean up the temporary SSL certificate files created from the .env vars
	cleanupTempCertFiles()

	// Create the cache and router
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
