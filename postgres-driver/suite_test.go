package postgresdriver

import (
	"context"
	"fmt"
	"testing"

	"github.com/pokt-foundation/portal-http-db/v2/types"
	"github.com/stretchr/testify/suite"
)

const (
	connectionString = "postgres://postgres:pgpassword@localhost:5432/postgres?sslmode=disable" // pragma: allowlist secret
)

type (
	PGDriverTestSuite struct {
		suite.Suite
		connectionString string
		driver           *PostgresDriver
	}
)

func Test_PGDriver_RunPGDriverSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping driver integration test")
	}

	testSuite := new(PGDriverTestSuite)
	testSuite.connectionString = connectionString

	suite.Run(t, testSuite)
}

// SetupSuite runs before each test suite run
func (ts *PGDriverTestSuite) SetupSuite() {
	_, err := ts.initPostgresDriver()
	ts.NoError(err)
}

// Initializes a real instance of the Postgres driver that connects to the test Postgres Docker container
func (ts *PGDriverTestSuite) initPostgresDriver() (func() error, error) {
	ctx := context.Background()

	mainConnPool, cleanup, err := NewPGXPool(connectionString)
	if err != nil {
		return nil, err
	}

	notificationChannel := make(chan *types.Notification, 32)

	listener := NewPGXPoolListener(mainConnPool, notificationChannel)

	driver, errCh, err := NewPostgresDriver(mainConnPool, listener, notificationChannel)
	if err != nil {
		panic(err)
	}
	ts.driver = driver

	// This goroutine continuously monitors the error channel and logs any errors that come in.
	go func() {
		for err := range errCh {
			fmt.Printf("error from listener: %v", err)
		}
	}()

	if err := ts.driver.Ping(ctx); err != nil {
		return nil, err
	}

	return cleanup, nil
}
