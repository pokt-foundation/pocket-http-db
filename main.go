package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
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
	pgHost                     = "PG_HOST"
	pgPort                     = "PG_PORT"
	pgUser                     = "PG_USER"
	pgPassword                 = "PG_PASSWORD"
	pgDatabase                 = "PG_DATABASE"
	pgSSLRootCert              = "PG_SSL_ROOT_CERT"
	pgSSLCert                  = "PG_SSL_CERT"
	pgSSLKey                   = "PG_SSL_KEY"
	apiKeys                    = "API_KEYS"
	cacheRefresh               = "CACHE_REFRESH"
	port                       = "PORT"
	defaultCacheRefreshMinutes = 5
	defaultPort                = "8080"

	sslRootCertFilename = "ssl_root_cert.pem"
	sslCertFilename     = "ssl_cert.pem"
	sslKeyFilename      = "ssl_key.pem"
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
		pgHost:              environment.MustGetString(pgHost),
		pgPort:              environment.MustGetString(pgPort),
		pgUser:              environment.MustGetString(pgUser),
		pgPassword:          environment.MustGetString(pgPassword),
		pgDatabase:          environment.MustGetString(pgDatabase),
		pgSSLRootCert:       environment.GetString(pgSSLRootCert, ""),
		pgSSLCert:           environment.GetString(pgSSLCert, ""),
		pgSSLKey:            environment.GetString(pgSSLKey, ""),
		apiKeys:             environment.MustGetStringMap(apiKeys, ","),
		cacheRefreshMinutes: time.Duration(environment.GetInt64(cacheRefresh, defaultCacheRefreshMinutes)) * time.Minute,
		port:                environment.GetString(port, defaultPort),
	}
}

// createCertFile creates a .pem certificate file for connecting to the database using SSL.
func createCertFile(content string, filename string) (string, error) {
	// Remove the file if it exists
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		err := os.Remove(filename)
		if err != nil {
			return "", err
		}
	}

	certFile, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer certFile.Close()

	if err := certFile.Chmod(0600); err != nil {
		return "", err
	}

	content = strings.ReplaceAll(content, "\\n", "\n")

	_, err = certFile.WriteString(content)
	if err != nil {
		return "", err
	}

	return certFile.Name(), nil
}

// buildConnectionString constructs the Postgre connection string using the given options.
// If the SSL env vars are present, it also creates certificate files for the SSL certificates.
func buildConnectionString(options options) (string, error) {
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s",
		options.pgHost,
		options.pgPort,
		options.pgUser,
		options.pgPassword,
		options.pgDatabase,
	)

	// Only configure the connection string to use SSL if all the SSL variables are present.
	// This is to allow E2E testing of PHD using a local Docker Postgres DB without SSL.
	if options.pgSSLRootCert != "" && options.pgSSLCert != "" && options.pgSSLKey != "" {
		sslRootCertPath, err := createCertFile(options.pgSSLRootCert, sslRootCertFilename)
		if err != nil {
			return "", err
		}
		sslCertPath, err := createCertFile(options.pgSSLCert, sslCertFilename)
		if err != nil {
			return "", err
		}
		sslKeyPath, err := createCertFile(options.pgSSLKey, sslKeyFilename)
		if err != nil {
			return "", err
		}

		connectionString = fmt.Sprintf("%s sslmode=verify-ca sslrootcert=%s sslcert=%s sslkey=%s",
			connectionString,
			sslRootCertPath,
			sslCertPath,
			sslKeyPath,
		)
	} else {
		connectionString = fmt.Sprintf("%s sslmode=disable", connectionString)
	}

	return connectionString, nil
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
	connectionString, err := buildConnectionString(options)
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
