package cache

import (
	"context"
	"testing"

	"github.com/pokt-foundation/portal-db/v2/driver"
	postgresdriver "github.com/pokt-foundation/portal-db/v2/postgres-driver"
	"github.com/pokt-foundation/portal-db/v2/types"
)

type ReaderMock struct {
	driver.MockDriver
	listener     *postgresdriver.ListenerMock
	notification chan *types.Notification
}

func NewReaderMock(t *testing.T) *ReaderMock {
	notificationChannel := make(chan *types.Notification, 32)

	listenerMock := postgresdriver.NewListenerMock(notificationChannel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := listenerMock.Listen(ctx)
	if err != nil {
		panic(err)
	}

	mock := &ReaderMock{
		MockDriver:   *driver.NewMockDriver(t),
		listener:     listenerMock,
		notification: notificationChannel,
	}

	err = listenerMock.Listen(ctx)
	if err != nil {
		panic(err)
	}

	return mock
}

func (r *ReaderMock) NotificationChannel() <-chan *types.Notification {
	return r.notification
}
