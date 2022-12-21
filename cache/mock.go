package cache

import (
	"testing"

	"github.com/pokt-foundation/portal-db/driver"
	postgresdriver "github.com/pokt-foundation/portal-db/postgres-driver"
	"github.com/pokt-foundation/portal-db/types"
)

type ReaderMock struct {
	driver.MockDriver
	lMock        *postgresdriver.ListenerMock
	notification chan *types.Notification
}

func NewReaderMock(t *testing.T) *ReaderMock {
	mock := &ReaderMock{
		MockDriver:   *driver.NewMockDriver(t),
		lMock:        postgresdriver.NewListenerMock(),
		notification: make(chan *types.Notification, 32),
	}

	go postgresdriver.Listen(mock.lMock.NotificationChannel(), mock.notification)

	return mock
}

func (r *ReaderMock) NotificationChannel() <-chan *types.Notification {
	return r.notification
}
