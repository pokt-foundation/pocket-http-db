package postgresdriver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pokt-foundation/portal-http-db/v2/types"
	"github.com/pokt-foundation/utils-go/id"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The PostgresDriver struct satisfies the Driver interface which defines all database driver methods
type (
	PostgresDriver struct {
		*Queries
		DB           *pgxpool.Pool
		notification chan *types.Notification
		listener     Listener
	}

	CloudSQLConfig struct {
		DBUser                 string
		DBPassword             string
		DBName                 string
		InstanceConnectionName string
		PublicIP               string
		PrivateIP              string
	}
)

var (
	errCantSetPrivateandPublicIP error = errors.New("cannot set both private and public IP; must set one or the other")
)

/* ---------- Postgres Connection Funcs ---------- */

/* <--------- PGX Pool Connection ---------> */

/*
NewCloudSQLPGXPool
- Creates a pool of connections to a Cloud SQL instance using the provided CloudSQLConfig.
- Uses the Cloud SQL Connector for Go and pgx for the connections.
- Establishes a dialer with the desired options (like using private IP).
- For each acquired connection from the pool, custom enum types are registered.
- Returns the established connection pool.
- It is important to note that this function will return an error if both PublicIP and PrivateIP are provided in the CloudSQLConfig.
*/
func NewCloudSQLPGXPool(options CloudSQLConfig) (*pgxpool.Pool, func() error, error) {
	if options.PublicIP != "" && options.PrivateIP != "" {
		return nil, nil, errCantSetPrivateandPublicIP
	}

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s", options.DBUser, options.DBPassword, options.DBName)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, nil, err
	}

	var opts []cloudsqlconn.Option
	if options.PublicIP != "" {
		opts = append(opts, cloudsqlconn.WithDefaultDialOptions(cloudsqlconn.WithPublicIP()))
	}
	if options.PrivateIP != "" {
		opts = append(opts, cloudsqlconn.WithDefaultDialOptions(cloudsqlconn.WithPrivateIP()))
	}

	d, err := cloudsqlconn.NewDialer(context.Background(), opts...)
	if err != nil {
		return nil, nil, err
	}

	config.ConnConfig.DialFunc = func(ctx context.Context, _, _ string) (net.Conn, error) {
		return d.Dial(ctx, options.InstanceConnectionName)
	}
	pool, err := createAndConfigurePool(config)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() error {
		pool.Close()
		return d.Close()
	}

	return pool, cleanup, nil
}

/*
NewPGXPool
- Creates a pool of connections to a PostgreSQL database using the provided connection string.
- Parses the connection string into a pgx pool configuration object.
- For each acquired connection from the pool, custom enum types are registered.
- Returns the established connection pool.
- This function is ideal for creating multiple reusable connections to a PostgreSQL database, particularly useful for handling multiple concurrent database operations.
*/
func NewPGXPool(connectionString string) (*pgxpool.Pool, func() error, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, nil, err
	}

	pool, err := createAndConfigurePool(config)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() error {
		pool.Close()
		return nil
	}

	return pool, cleanup, nil
}

// Configures the connection pool with custom enum types.
func createAndConfigurePool(config *pgxpool.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig: %v", err)
	}

	// Collect the custom data types once, store them in memory, and register them for every future connection.
	customTypes, err := getCustomDataTypes(context.Background(), pool)
	if err != nil {
		return nil, err
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		for _, t := range customTypes {
			conn.TypeMap().RegisterType(t)
		}
		return nil
	}

	// Immediately close the old pool and open a new one with the new config.
	pool.Close()
	pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

// Any custom DB types made with CREATE TYPE need to be registered with pgx.
// https://github.com/kyleconroy/sqlc/issues/2116
// https://stackoverflow.com/questions/75658429/need-to-update-psql-row-of-a-composite-type-in-golang-with-jack-pgx
// https://pkg.go.dev/github.com/jackc/pgx/v5/pgtype
func getCustomDataTypes(ctx context.Context, pool *pgxpool.Pool) ([]*pgtype.Type, error) {
	// Get a single connection just to load type information.
	conn, err := pool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, err
	}

	dataTypeNames := []string{
		"auth_provider",
		"_auth_provider",
		"auth_type",
		"_auth_type",
		"chain_auth_type",
		"_chain_auth_type",
		"chain_check_type",
		"_chain_check_type",
		"environment",
		"_environment",
		"notification_event",
		"_notification_event",
		"notification_type",
		"_notification_type",
		"permissions",
		"_permissions",
		"whitelist_type",
		"_whitelist_type",
	}

	var typesToRegister []*pgtype.Type
	for _, typeName := range dataTypeNames {
		dataType, err := conn.Conn().LoadType(ctx, typeName)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to load type %s: %v", typeName, err)
		}
		// You need to register only for this connection too, otherwise the array type will look for the register element type.
		conn.Conn().TypeMap().RegisterType(dataType)
		typesToRegister = append(typesToRegister, dataType)
	}
	return typesToRegister, nil
}

/*
NewPostgresDriver
- Creates an instance of PostgresDriver using the provided pgx connection, Listener, and notification channel.
- Initiates the listening of notifications in a separate goroutine.
- Returns the created PostgresDriver instance.
*/
func NewPostgresDriver(pool *pgxpool.Pool, listener Listener, notificationChannel chan *types.Notification) (*PostgresDriver, chan error, error) {
	errCh := make(chan error, 100)

	driver := &PostgresDriver{
		Queries:      New(pool),
		DB:           pool,
		listener:     listener,
		notification: notificationChannel,
	}

	go func() {
		err := driver.listener.Listen(context.Background())
		if err != nil {
			errCh <- fmt.Errorf("error listening for notifications: %v", err)
		}
	}()

	return driver, errCh, nil
}

// Ping ensures the database connection is healthy
func (d *PostgresDriver) Ping(ctx context.Context) error {
	return d.DB.Ping(ctx)
}

// NotificationChannel returns receiver Notification channel
func (pg *PostgresDriver) NotificationChannel() <-chan *types.Notification {
	return pg.notification
}

/* ---------- Postgres Driver Util Methods ---------- */
const idLength = 8

var (
	errCheckingID = errors.New("error checking if ID %s exists")
)

// generateID generates a new random ID with the specified length and checks that it doesn't already exist in the DB.
// If it already exists in the DB a new random ID is generated until one is created that does not already exist.
// Checks the following tables for existing IDs: `portal_applications`, `portal_application_aats`, `gigastake_applications`, `accounts` & `users`
func (pg *PostgresDriver) generateID(ctx context.Context) (string, error) {
	var generatedID string
	var err error
	idExists := true

	for idExists {
		generatedID = id.GenerateID(idLength)
		idExists, err = pg.CheckIDExists(ctx, generatedID)
		if err != nil {
			return "", fmt.Errorf(errCheckingID.Error(), generatedID)
		}
	}

	return generatedID, nil
}

func newText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: value,
		Valid:  true,
	}
}

func newInt4(value int32, allowZero bool) pgtype.Int4 {
	if !allowZero && value == 0 {
		return pgtype.Int4{}
	}

	return pgtype.Int4{
		Int32: value,
		Valid: true,
	}
}

func newBool(value *bool) pgtype.Bool {
	if value == nil {
		return pgtype.Bool{Valid: false}
	}

	return pgtype.Bool{
		Bool:  *value,
		Valid: true,
	}
}

func newTimestamptz(value time.Time) pgtype.Timestamptz {
	if value.IsZero() {
		return pgtype.Timestamptz{}
	}

	return pgtype.Timestamptz{
		Time:  value,
		Valid: true,
	}
}

func newNullString(s *string) pgtype.Text {
	if s == nil || *s == "" {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: *s,
		Valid:  true,
	}
}

func newNullInt(i *int32, allowZero bool) pgtype.Int4 {
	if i == nil || (!allowZero && *i == 0) {
		return pgtype.Int4{}
	}

	return pgtype.Int4{
		Int32: *i,
		Valid: true,
	}
}

func errNoRows(err error) bool {
	return strings.Contains(err.Error(), "no rows in result set")
}
