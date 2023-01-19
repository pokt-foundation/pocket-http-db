package cache

import (
	"context"
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

func (r *ReaderMock) ReadApplications(ctx context.Context) ([]*types.Application, error) {
	args := r.Called(ctx)

	return args.Get(0).([]*types.Application), args.Error(1)
}

func (r *ReaderMock) ReadBlockchains(ctx context.Context) ([]*types.Blockchain, error) {
	args := r.Called(ctx)

	return args.Get(0).([]*types.Blockchain), args.Error(1)
}

func (r *ReaderMock) ReadLoadBalancers(ctx context.Context) ([]*types.LoadBalancer, error) {
	args := r.Called(ctx)

	return args.Get(0).([]*types.LoadBalancer), args.Error(1)
}

func (r *ReaderMock) ReadPayPlans(ctx context.Context) ([]*types.PayPlan, error) {
	args := r.Called(ctx)

	return args.Get(0).([]*types.PayPlan), args.Error(1)
}

func (r *ReaderMock) ReadRedirects(ctx context.Context) ([]*types.Redirect, error) {
	args := r.Called(ctx)

	return args.Get(0).([]*types.Redirect), args.Error(1)
}
