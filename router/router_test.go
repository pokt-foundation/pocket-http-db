package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pokt-foundation/pocket-http-db/cache"
	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type writerMock struct {
	mock.Mock
}

func (w *writerMock) WriteLoadBalancer(loadBalancer *repository.LoadBalancer) (*repository.LoadBalancer, error) {
	args := w.Called()

	return args.Get(0).(*repository.LoadBalancer), args.Error(1)
}

func (w *writerMock) UpdateLoadBalancer(id string, options *repository.UpdateLoadBalancer) error {
	args := w.Called()

	return args.Error(0)
}

func (w *writerMock) WriteApplication(app *repository.Application) (*repository.Application, error) {
	args := w.Called()

	return args.Get(0).(*repository.Application), args.Error(1)
}

func (w *writerMock) UpdateApplication(id string, options *repository.UpdateApplication) error {
	args := w.Called()

	return args.Error(0)
}

func newTestRouter() (*Router, error) {
	readerMock := &cache.ReaderMock{}

	readerMock.On("ReadApplications").Return([]*repository.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
			Limits: repository.AppLimits{
				DailyLimit: 1000000,
			},
		},
		{
			ID:     "5f62b7d8be3591c4dea8566a",
			UserID: "60ecb2bf67774900350d9c43",
		},
		{
			ID:     "5f62b7d8be3591c4dea8566f",
			UserID: "60ecb2bf67774900350d9c44",
		},
	}, nil)

	readerMock.On("ReadBlockchains").Return([]*repository.Blockchain{
		{
			ID: "0021",
		},
		{
			ID: "0022",
		},
	}, nil)

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{
		{
			ID: "60ecb2bf67774900350d9c42",
			ApplicationIDs: []string{
				"5f62b7d8be3591c4dea8566d",
				"5f62b7d8be3591c4dea8566a",
			},
			UserID: "60ecb2bf67774900350d9c43",
		},
		{
			ID: "60ecb2bf67774900350d9c43",
			ApplicationIDs: []string{
				"5f62b7d8be3591c4dea8566d",
				"5f62b7d8be3591c4dea8566a",
			},
		},
	}, nil)

	readerMock.On("ReadUsers").Return([]*repository.User{
		{
			ID: "60ecb2bf67774900350d9c43",
		},
		{
			ID: "60ecb2bf67774900350d9c44",
		},
	}, nil)

	return NewRouter(readerMock, nil)
}

func TestRouter_HealthCheck(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)
}

func TestRouter_GetApplications(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/application", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*repository.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
			Limits: repository.AppLimits{
				DailyLimit: 1000000,
			},
		},
		{
			ID:     "5f62b7d8be3591c4dea8566a",
			UserID: "60ecb2bf67774900350d9c43",
		},
		{
			ID:     "5f62b7d8be3591c4dea8566f",
			UserID: "60ecb2bf67774900350d9c44",
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())

	req.Header.Set("Authorization", "wrong")

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusUnauthorized, rr.Code)
}

func TestRouter_GetApplicationsLimits(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/application/limits", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*repository.AppLimits{
		{
			AppID:      "5f62b7d8be3591c4dea8566d",
			DailyLimit: 1000000,
		},
		{
			AppID:      "5f62b7d8be3591c4dea8566a",
			DailyLimit: 0,
		},
		{
			AppID:      "5f62b7d8be3591c4dea8566f",
			DailyLimit: 0,
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func TestRouter_GetApplication(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/application/5f62b7d8be3591c4dea8566d", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal(&repository.Application{
		ID:     "5f62b7d8be3591c4dea8566d",
		UserID: "60ecb2bf67774900350d9c43",
		Limits: repository.AppLimits{
			DailyLimit: 1000000,
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func TestRouter_CreateApplication(t *testing.T) {
	c := require.New(t)

	rawAppToSend := &repository.Application{
		UserID: "60ddc61b6e29c3003378361D",
	}

	appToSend, err := json.Marshal(rawAppToSend)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/application", bytes.NewBuffer(appToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	appToReturn := &repository.Application{
		ID:     "60ddc61b6e29c3003378361E",
		UserID: "60ddc61b6e29c3003378361D",
	}

	writerMock := &writerMock{}

	writerMock.On("WriteApplication", mock.Anything).Return(appToReturn, nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	marshaledReturnApp, err := json.Marshal(appToReturn)
	c.NoError(err)

	c.Equal(marshaledReturnApp, rr.Body.Bytes())

	req, err = http.NewRequest(http.MethodPost, "/application", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPost, "/application", bytes.NewBuffer(appToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("WriteApplication", mock.Anything).Return(appToReturn, errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func TestRouter_UpdateApplication(t *testing.T) {
	c := require.New(t)

	rawUpdateInput := &repository.UpdateApplication{
		Name:   "pablo",
		UserID: "6025be31e1261e00308bfa3a",
		Status: repository.Orphaned,
		GatewaySettings: &repository.GatewaySettings{
			SecretKey: "1234",
		},
	}

	updateInputToSend, err := json.Marshal(rawUpdateInput)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPut, "/application/5f62b7d8be3591c4dea8566d", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	writerMock := &writerMock{}

	writerMock.On("UpdateApplication", mock.Anything).Return(nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/application/5f62b7d8be3591c4dea85664", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/application/5f62b7d8be3591c4dea8566d", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/application/5f62b7d8be3591c4dea8566d", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("UpdateApplication", mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func TestRouter_GetApplicationsByUserID(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/user/60ecb2bf67774900350d9c43/application", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*repository.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
			Limits: repository.AppLimits{
				DailyLimit: 1000000,
			},
		},
		{
			ID:     "5f62b7d8be3591c4dea8566a",
			UserID: "60ecb2bf67774900350d9c43",
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func TestRouter_GetLoadbalancerByUserID(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/user/60ecb2bf67774900350d9c43/load_balancer", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	var marshaledBody []*repository.LoadBalancer

	err = json.Unmarshal(rr.Body.Bytes(), &marshaledBody)
	c.NoError(err)

	c.Equal("60ecb2bf67774900350d9c42", marshaledBody[0].ID)
}

func TestRouter_GetBlockchains(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/blockchain", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*repository.Blockchain{
		{
			ID: "0021",
		},
		{
			ID: "0022",
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func TestRouter_GetBlockchain(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/blockchain/0021", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal(&repository.Blockchain{
		ID: "0021",
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func TestRouter_GetLoadBalancers(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/load_balancer", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	var marshaledBody []*repository.LoadBalancer

	err = json.Unmarshal(rr.Body.Bytes(), &marshaledBody)
	c.NoError(err)

	c.Len(marshaledBody, 2)
	c.Equal("60ecb2bf67774900350d9c42", marshaledBody[0].ID)
	c.Equal("60ecb2bf67774900350d9c43", marshaledBody[1].ID)
}

func TestRouter_GetLoadBalancer(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/load_balancer/60ecb2bf67774900350d9c42", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	var marshaledBody repository.LoadBalancer

	err = json.Unmarshal(rr.Body.Bytes(), &marshaledBody)
	c.NoError(err)

	c.Equal("60ecb2bf67774900350d9c42", marshaledBody.ID)
}

func TestRouter_CreateLoadBalancer(t *testing.T) {
	c := require.New(t)

	rawLBToSend := &repository.LoadBalancer{
		UserID: "60ddc61b6e29c3003378361D",
	}

	lbToSend, err := json.Marshal(rawLBToSend)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/load_balancer", bytes.NewBuffer(lbToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	lbToReturn := &repository.LoadBalancer{
		ID:     "60ddc61b6e29c3003378361E",
		UserID: "60ddc61b6e29c3003378361D",
	}

	writerMock := &writerMock{}

	writerMock.On("WriteLoadBalancer", mock.Anything).Return(lbToReturn, nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	marshaledReturnLB, err := json.Marshal(lbToReturn)
	c.NoError(err)

	c.Equal(marshaledReturnLB, rr.Body.Bytes())

	req, err = http.NewRequest(http.MethodPost, "/load_balancer", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPost, "/load_balancer", bytes.NewBuffer(lbToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("WriteLoadBalancer", mock.Anything).Return(lbToReturn, errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func TestRouter_UpdateLoadBalancer(t *testing.T) {
	c := require.New(t)

	rawUpdateInput := &repository.UpdateLoadBalancer{
		Name:   "pablo",
		UserID: "60ddc61b6e29c3003378361D",
	}

	updateInputToSend, err := json.Marshal(rawUpdateInput)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPut, "/load_balancer/60ecb2bf67774900350d9c42", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	writerMock := &writerMock{}

	writerMock.On("UpdateLoadBalancer", mock.Anything).Return(nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/load_balancer/5f62b7d8be3591c4dea85664", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/load_balancer/60ecb2bf67774900350d9c42", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/load_balancer/60ecb2bf67774900350d9c42", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("UpdateLoadBalancer", mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func TestRouter_GetUsers(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/user", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*repository.User{
		{
			ID: "60ecb2bf67774900350d9c43",
		},
		{
			ID: "60ecb2bf67774900350d9c44",
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func TestRouter_GetUser(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/user/60ecb2bf67774900350d9c43", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal(&repository.User{
		ID: "60ecb2bf67774900350d9c43",
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}
