package cache

import (
	postgresdriver "github.com/pokt-foundation/portal-api-go/postgres-driver"
	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/stretchr/testify/mock"
)

// ReaderMock struct handler for mocking reader interface
type ReaderMock struct {
	mock.Mock
	lMock        *postgresdriver.ListenerMock
	notification chan *repository.Notification
}

func NewReaderMock() *ReaderMock {
	mock := &ReaderMock{
		lMock:        postgresdriver.NewListenerMock(),
		notification: make(chan *repository.Notification, 32),
	}

	go postgresdriver.Listen(mock.lMock.NotificationChannel(), mock.notification)

	return mock
}

func (r *ReaderMock) ReadApplications() ([]*repository.Application, error) {
	args := r.Called()

	return args.Get(0).([]*repository.Application), args.Error(1)
}

func (r *ReaderMock) ReadBlockchains() ([]*repository.Blockchain, error) {
	args := r.Called()

	return args.Get(0).([]*repository.Blockchain), args.Error(1)
}

func (r *ReaderMock) ReadLoadBalancers() ([]*repository.LoadBalancer, error) {
	args := r.Called()

	return args.Get(0).([]*repository.LoadBalancer), args.Error(1)
}

func (r *ReaderMock) ReadPayPlans() ([]*repository.PayPlan, error) {
	args := r.Called()

	return args.Get(0).([]*repository.PayPlan), args.Error(1)
}

func (r *ReaderMock) ReadRedirects() ([]*repository.Redirect, error) {
	args := r.Called()

	return args.Get(0).([]*repository.Redirect), args.Error(1)
}

func (r *ReaderMock) NotificationChannel() <-chan *repository.Notification {
	return r.notification
}
