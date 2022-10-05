package cache

import (
	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/stretchr/testify/mock"
)

// ReaderMock struct handler for mocking reader interface
type ReaderMock struct {
	mock.Mock
}

func (r *ReaderMock) ReadApplications() ([]repository.Application, error) {
	args := r.Called()

	return args.Get(0).([]repository.Application), args.Error(1)
}

func (r *ReaderMock) ReadBlockchains() ([]repository.Blockchain, error) {
	args := r.Called()

	return args.Get(0).([]repository.Blockchain), args.Error(1)
}

func (r *ReaderMock) ReadLoadBalancers() ([]repository.LoadBalancer, error) {
	args := r.Called()

	return args.Get(0).([]repository.LoadBalancer), args.Error(1)
}

func (r *ReaderMock) ReadPayPlans() ([]repository.PayPlan, error) {
	args := r.Called()

	return args.Get(0).([]repository.PayPlan), args.Error(1)
}

func (r *ReaderMock) ReadRedirects() ([]repository.Redirect, error) {
	args := r.Called()

	return args.Get(0).([]repository.Redirect), args.Error(1)
}
