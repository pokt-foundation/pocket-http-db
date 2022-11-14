package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pokt-foundation/pocket-http-db/cache"
	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/sirupsen/logrus"
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

func (w *writerMock) RemoveLoadBalancer(id string) error {
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

func (w *writerMock) UpdateFirstDateSurpassed(firstDateSurpassed *repository.UpdateFirstDateSurpassed) error {
	args := w.Called()

	return args.Error(0)
}

func (w *writerMock) RemoveApplication(id string) error {
	args := w.Called()

	return args.Error(0)
}

func (w *writerMock) WriteBlockchain(blockchain *repository.Blockchain) (*repository.Blockchain, error) {
	args := w.Called()

	return args.Get(0).(*repository.Blockchain), args.Error(1)
}

func (w *writerMock) WriteRedirect(redirect *repository.Redirect) (*repository.Redirect, error) {
	args := w.Called()

	return args.Get(0).(*repository.Redirect), args.Error(1)
}

func (w *writerMock) ActivateBlockchain(id string, active bool) error {
	args := w.Called()

	return args.Error(0)
}

func newTestRouter() (*Router, error) {
	readerMock := &cache.ReaderMock{}

	readerMock.On("ReadPayPlans").Return([]*repository.PayPlan{
		{
			Type:  repository.FreetierV0,
			Limit: 250000,
		},
		{
			Type:  repository.PayAsYouGoV0,
			Limit: 0,
		},
	}, nil)

	readerMock.On("ReadRedirects").Return([]*repository.Redirect{
		{
			BlockchainID:   "0021",
			Alias:          "pokt-mainnet",
			Domain:         "pokt-mainnet.gateway.network",
			LoadBalancerID: "12345",
		},
		{
			BlockchainID:   "0022",
			Alias:          "eth-mainnet",
			Domain:         "eth-mainnet.gateway.network",
			LoadBalancerID: "45678",
		},
	}, nil)

	readerMock.On("ReadApplications").Return([]*repository.Application{
		{
			ID:     "test_app_e3591c4dea8566d",
			UserID: "test_user_774900350d9c43",
			Limit: repository.AppLimit{
				PayPlan: repository.PayPlan{
					Type:  repository.FreetierV0,
					Limit: 250000,
				},
			},
			FirstDateSurpassed: time.Date(2022, time.July, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:     "test_app_e3591c4dea8566a",
			UserID: "test_user_774900350d9c43",
		},
		{
			ID:     "test_app_e3591c4dea8566f",
			UserID: "test_user_774900350d9c44",
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
			ID: "test_lb_67774900350d9c42",
			ApplicationIDs: []string{
				"test_app_e3591c4dea8566d",
				"test_app_e3591c4dea8566a",
			},
			UserID: "test_user_774900350d9c43",
		},
		{
			ID: "test_user_774900350d9c43",
			ApplicationIDs: []string{
				"test_app_e3591c4dea8566d",
				"test_app_e3591c4dea8566a",
			},
		},
	}, nil)

	return NewRouter(readerMock, nil, map[string]bool{"": true}, logrus.New())
}

func Test_V0_Router_HealthCheck(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)
}

func Test_V0_Router_GetApplications(t *testing.T) {
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
			ID:     "test_app_e3591c4dea8566d",
			UserID: "test_user_774900350d9c43",
			Limit: repository.AppLimit{
				PayPlan: repository.PayPlan{
					Type:  repository.FreetierV0,
					Limit: 250000,
				},
			},
			FirstDateSurpassed: time.Date(2022, time.July, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:     "test_app_e3591c4dea8566a",
			UserID: "test_user_774900350d9c43",
		},
		{
			ID:     "test_app_e3591c4dea8566f",
			UserID: "test_user_774900350d9c44",
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())

	req.Header.Set("Authorization", "wrong")

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusUnauthorized, rr.Code)
}

func Test_V0_Router_GetApplicationsLimits(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/application/limits", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	dateSurpassed := time.Date(2022, time.July, 21, 0, 0, 0, 0, time.UTC)

	expectedBody, err := json.Marshal([]*repository.AppLimits{
		{
			AppID:                "test_app_e3591c4dea8566d",
			AppUserID:            "test_user_774900350d9c43",
			PlanType:             repository.FreetierV0,
			DailyLimit:           250000,
			FirstDateSurpassed:   &dateSurpassed,
			NotificationSettings: &repository.NotificationSettings{},
		},
		{
			AppID:                "test_app_e3591c4dea8566a",
			AppUserID:            "test_user_774900350d9c43",
			DailyLimit:           0,
			NotificationSettings: &repository.NotificationSettings{},
		},
		{
			AppID:                "test_app_e3591c4dea8566f",
			AppUserID:            "test_user_774900350d9c44",
			DailyLimit:           0,
			NotificationSettings: &repository.NotificationSettings{},
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func Test_V0_Router_GetApplication(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/application/test_app_e3591c4dea8566d", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal(&repository.Application{
		ID:     "test_app_e3591c4dea8566d",
		UserID: "test_user_774900350d9c43",
		Limit: repository.AppLimit{
			PayPlan: repository.PayPlan{
				Type:  repository.FreetierV0,
				Limit: 250000,
			},
		},
		FirstDateSurpassed: time.Date(2022, time.July, 21, 0, 0, 0, 0, time.UTC),
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())

	req, err = http.NewRequest(http.MethodGet, "/application/test_app_e3591c3dea8566d", nil)
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func Test_V0_Router_CreateApplication(t *testing.T) {
	c := require.New(t)

	rawAppToSend := &repository.Application{
		UserID: "test_app_e29c3003378361D",
		Limit: repository.AppLimit{
			PayPlan: repository.PayPlan{
				Type:  repository.FreetierV0,
				Limit: 250000,
			},
		},
	}

	appToSend, err := json.Marshal(rawAppToSend)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/application", bytes.NewBuffer(appToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	appToReturn := &repository.Application{
		ID:     "test_app_e29c3003378361E",
		UserID: "test_app_e29c3003378361D",
		Limit: repository.AppLimit{
			PayPlan: repository.PayPlan{
				Type:  repository.FreetierV0,
				Limit: 250000,
			},
		},
	}

	writerMock := &writerMock{}

	writerMock.On("WriteApplication", mock.Anything).Return(appToReturn, nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	marshaledReturnApp, err := json.Marshal(&repository.Application{
		ID:     "test_app_e29c3003378361E",
		UserID: "test_app_e29c3003378361D",
		Limit: repository.AppLimit{
			PayPlan: repository.PayPlan{
				Type:  repository.FreetierV0,
				Limit: 250000,
			},
		},
	})
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

func Test_V0_Router_UpdateApplication(t *testing.T) {
	c := require.New(t)

	rawUpdateInput := &repository.UpdateApplication{
		Name:   "pablo",
		Status: repository.Orphaned,
		Limit: &repository.AppLimit{
			PayPlan: repository.PayPlan{
				Type:  repository.PayAsYouGoV0,
				Limit: 0,
			},
		},
		FirstDateSurpassed: time.Now(),
		GatewaySettings: &repository.GatewaySettings{
			SecretKey: "1234",
		},
		NotificationSettings: &repository.NotificationSettings{
			Half: true,
		},
	}

	updateInputToSend, err := json.Marshal(rawUpdateInput)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPut, "/application/test_app_e3591c4dea8566d", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	writerMock := &writerMock{}

	writerMock.On("UpdateApplication", mock.Anything).Return(nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/application/test_app_e3591c4dea85664", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/application/test_app_e3591c4dea8566d", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/application/test_app_e3591c4dea8566d", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("UpdateApplication", mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusUnprocessableEntity, rr.Code)
}

func Test_V0_Router_UpdateFirstDateSurpassed(t *testing.T) {
	c := require.New(t)

	rawUpdateInput := &repository.UpdateFirstDateSurpassed{
		ApplicationIDs:     []string{"test_app_e3591c4dea8566d"},
		FirstDateSurpassed: time.Now(),
	}

	updateInputToSend, err := json.Marshal(rawUpdateInput)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/application/first_date_surpassed", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	writerMock := &writerMock{}

	writerMock.On("UpdateFirstDateSurpassed", mock.Anything).Return(nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	writerMock.On("UpdateFirstDateSurpassed", mock.Anything).Return(errors.New("dummy error")).Once()

	req, err = http.NewRequest(http.MethodPost, "/application/first_date_surpassed", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)

	req, err = http.NewRequest(http.MethodPost, "/application/first_date_surpassed", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	rawUpdateInput.ApplicationIDs = []string{"test_app_e3591c4dea85664"}

	updateInputToSend, err = json.Marshal(rawUpdateInput)
	c.NoError(err)

	req, err = http.NewRequest(http.MethodPost, "/application/first_date_surpassed", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)

	rawUpdateInput.ApplicationIDs = nil

	updateInputToSend, err = json.Marshal(rawUpdateInput)
	c.NoError(err)

	req, err = http.NewRequest(http.MethodPost, "/application/first_date_surpassed", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)
}

func Test_V0_Router_RemoveApplication(t *testing.T) {
	c := require.New(t)

	rawRemoveInput := &repository.UpdateApplication{
		Remove: true,
	}

	updateInputToSend, err := json.Marshal(rawRemoveInput)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPut, "/application/test_app_e3591c4dea8566d", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	writerMock := &writerMock{}

	writerMock.On("RemoveApplication", mock.Anything).Return(nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/application/test_app_e3591c4dea85664", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/application/test_app_e3591c4dea8566d", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/application/test_app_e3591c4dea8566d", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("RemoveApplication", mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func Test_V0_Router_GetApplicationsByUserID(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/user/test_user_774900350d9c43/application", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*repository.Application{
		{
			ID:     "test_app_e3591c4dea8566d",
			UserID: "test_user_774900350d9c43",
			Limit: repository.AppLimit{
				PayPlan: repository.PayPlan{
					Type:  repository.FreetierV0,
					Limit: 250000,
				},
			},
			FirstDateSurpassed: time.Date(2022, time.July, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:     "test_app_e3591c4dea8566a",
			UserID: "test_user_774900350d9c43",
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())

	req, err = http.NewRequest(http.MethodGet, "/user/61ecb2bf67774900350d9c43/application", nil)
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func Test_V0_Router_GetLoadbalancerByUserID(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/user/test_user_774900350d9c43/load_balancer", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	var marshaledBody []*repository.LoadBalancer

	err = json.Unmarshal(rr.Body.Bytes(), &marshaledBody)
	c.NoError(err)

	c.Equal("test_lb_67774900350d9c42", marshaledBody[0].ID)

	req, err = http.NewRequest(http.MethodGet, "/user/60ecb2bf67774900350d9d43/load_balancer", nil)
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func Test_V0_Router_GetBlockchains(t *testing.T) {
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
			Redirects: []repository.Redirect{
				{
					BlockchainID:   "0021",
					Alias:          "pokt-mainnet",
					Domain:         "pokt-mainnet.gateway.network",
					LoadBalancerID: "12345",
				},
			},
		},
		{
			ID: "0022",
			Redirects: []repository.Redirect{
				{
					BlockchainID:   "0022",
					Alias:          "eth-mainnet",
					Domain:         "eth-mainnet.gateway.network",
					LoadBalancerID: "45678",
				},
			},
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func Test_V0_Router_GetBlockchain(t *testing.T) {
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
		Redirects: []repository.Redirect{
			{
				BlockchainID:   "0021",
				Alias:          "pokt-mainnet",
				Domain:         "pokt-mainnet.gateway.network",
				LoadBalancerID: "12345",
			},
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())

	req, err = http.NewRequest(http.MethodGet, "/blockchain/00210", nil)
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func Test_V0_Router_GetLoadBalancers(t *testing.T) {
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
	c.Equal("test_lb_67774900350d9c42", marshaledBody[0].ID)
	c.Equal("test_user_774900350d9c43", marshaledBody[1].ID)
}

func Test_V0_Router_GetLoadBalancer(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/load_balancer/test_lb_67774900350d9c42", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	var marshaledBody repository.LoadBalancer

	err = json.Unmarshal(rr.Body.Bytes(), &marshaledBody)
	c.NoError(err)

	c.Equal("test_lb_67774900350d9c42", marshaledBody.ID)

	req, err = http.NewRequest(http.MethodGet, "/load_balancer/60fcb2bf67774900350d9c42", nil)
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func Test_V0_Router_CreateLoadBalancer(t *testing.T) {
	c := require.New(t)

	rawLBToSend := &repository.LoadBalancer{
		UserID: "test_app_e29c3003378361D",
	}

	lbToSend, err := json.Marshal(rawLBToSend)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/load_balancer", bytes.NewBuffer(lbToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	lbToReturn := &repository.LoadBalancer{
		ID:     "test_app_e29c3003378361E",
		UserID: "test_app_e29c3003378361D",
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

func Test_V0_Router_UpdateLoadBalancer(t *testing.T) {
	c := require.New(t)

	rawUpdateInput := &repository.UpdateLoadBalancer{
		Name: "pablo",
		StickyOptions: &repository.StickyOptions{
			Duration:      "21",
			StickyOrigins: []string{"pjog"},
			StickyMax:     21,
			Stickiness:    true,
		},
	}

	updateInputToSend, err := json.Marshal(rawUpdateInput)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPut, "/load_balancer/test_lb_67774900350d9c42", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	writerMock := &writerMock{}

	writerMock.On("UpdateLoadBalancer", mock.Anything).Return(nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/load_balancer/test_app_e3591c4dea85664", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/load_balancer/test_lb_67774900350d9c42", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/load_balancer/test_lb_67774900350d9c42", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("UpdateLoadBalancer", mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func Test_V0_Router_RemoveLoadBalancer(t *testing.T) {
	c := require.New(t)

	rawUpdateInput := &repository.UpdateLoadBalancer{
		Remove: true,
	}

	updateInputToSend, err := json.Marshal(rawUpdateInput)
	c.NoError(err)

	router, err := newTestRouter()
	c.NoError(err)

	writerMock := &writerMock{}

	router.Writer = writerMock

	req, err := http.NewRequest(http.MethodPut, "/load_balancer/test_lb_67774900350d9c42", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	writerMock.On("RemoveLoadBalancer", mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/load_balancer/test_lb_67774900350d9c42", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("RemoveLoadBalancer", mock.Anything).Return(nil).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/load_balancer/test_app_e3591c4dea85664", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func Test_V0_Router_GetPayPlans(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/pay_plan", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*repository.PayPlan{
		{
			Type:  repository.FreetierV0,
			Limit: 250000,
		},
		{
			Type:  repository.PayAsYouGoV0,
			Limit: 0,
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func Test_V0_Router_GetPayPlan(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/pay_plan/freetier_v0", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal(&repository.PayPlan{
		Type:  repository.FreetierV0,
		Limit: 250000,
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())

	req, err = http.NewRequest(http.MethodGet, "/pay_plan/freetier_v21", nil)
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func Test_V0_Router_CreateBlockchain(t *testing.T) {
	c := require.New(t)

	rawChainToSend := &repository.Blockchain{
		Ticker: "POKT",
	}

	chainToSend, err := json.Marshal(rawChainToSend)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/blockchain", bytes.NewBuffer(chainToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	chainToReturn := &repository.Blockchain{
		ID:     "test_app_e29c3003378361E",
		Ticker: "POKT",
	}

	writerMock := &writerMock{}

	writerMock.On("WriteBlockchain", mock.Anything).Return(chainToReturn, nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	marshaledReturnChain, err := json.Marshal(chainToReturn)
	c.NoError(err)

	c.Equal(marshaledReturnChain, rr.Body.Bytes())

	req, err = http.NewRequest(http.MethodPost, "/blockchain", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPost, "/blockchain", bytes.NewBuffer(chainToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("WriteBlockchain", mock.Anything).Return(chainToReturn, errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func Test_V0_Router_CreateRedirect(t *testing.T) {
	c := require.New(t)

	rawRedirectsToSend := &repository.Redirect{BlockchainID: "0021"}

	redirectToSend, err := json.Marshal(rawRedirectsToSend)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/redirect", bytes.NewBuffer(redirectToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	redirectToReturn := &repository.Redirect{ID: "60ddc61ew3h4rn4nfnkkdf93", BlockchainID: "0021"}

	writerMock := &writerMock{}

	writerMock.On("WriteRedirect", mock.Anything).Return(redirectToReturn, nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	marshaledReturnRedirect, err := json.Marshal(redirectToReturn)
	c.NoError(err)

	c.Equal(marshaledReturnRedirect, rr.Body.Bytes())

	req, err = http.NewRequest(http.MethodPost, "/redirect", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPost, "/redirect", bytes.NewBuffer(redirectToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("WriteRedirect", mock.Anything).Return(redirectToReturn, errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func Test_V0_Router_ActivateBlockchain(t *testing.T) {
	c := require.New(t)

	activeStatusToSend, err := json.Marshal(true)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/blockchain/0021/activate", bytes.NewBuffer(activeStatusToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	writerMock := &writerMock{}

	writerMock.On("ActivateBlockchain", mock.Anything).Return(nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	req, err = http.NewRequest(http.MethodPost, "/blockchain/0021/activate", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPost, "/blockchain/0021/activate", bytes.NewBuffer(activeStatusToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("ActivateBlockchain", mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}
