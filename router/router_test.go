package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pokt-foundation/portal-db/driver"
	"github.com/pokt-foundation/portal-db/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testCtx = context.Background()

func newTestRouter(t *testing.T) (*Router, error) {
	readerMock := driver.NewMockDriver(t)

	readerMock.On("ReadPayPlans", testCtx).Return([]*types.PayPlan{
		{
			Type:  types.FreetierV0,
			Limit: 250000,
		},
		{
			Type:  types.PayAsYouGoV0,
			Limit: 0,
		},
	}, nil)

	readerMock.On("ReadApplications", testCtx).Return([]*types.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
			Limit: types.AppLimit{
				PayPlan: types.PayPlan{
					Type:  types.FreetierV0,
					Limit: 250000,
				},
			},
			FirstDateSurpassed: time.Date(2022, time.July, 21, 0, 0, 0, 0, time.UTC),
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

	readerMock.On("ReadBlockchains", testCtx).Return([]*types.Blockchain{
		{
			ID: "0021",
		},
		{
			ID: "0022",
		},
	}, nil)

	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{
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

	return NewRouter(readerMock, nil, map[string]bool{"": true}, logrus.New())
}

func TestRouter_HealthCheck(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)
}

func TestRouter_GetApplications(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/application", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*types.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
			Limit: types.AppLimit{
				PayPlan: types.PayPlan{
					Type:  types.FreetierV0,
					Limit: 250000,
				},
			},
			FirstDateSurpassed: time.Date(2022, time.July, 21, 0, 0, 0, 0, time.UTC),
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

func TestRouter_GetApplication(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/application/5f62b7d8be3591c4dea8566d", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal(&types.Application{
		ID:     "5f62b7d8be3591c4dea8566d",
		UserID: "60ecb2bf67774900350d9c43",
		Limit: types.AppLimit{
			PayPlan: types.PayPlan{
				Type:  types.FreetierV0,
				Limit: 250000,
			},
		},
		FirstDateSurpassed: time.Date(2022, time.July, 21, 0, 0, 0, 0, time.UTC),
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())

	req, err = http.NewRequest(http.MethodGet, "/application/5f62b7d8be3591c3dea8566d", nil)
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func TestRouter_CreateApplication(t *testing.T) {
	c := require.New(t)

	rawAppToSend := &types.Application{
		UserID: "60ddc61b6e29c3003378361D",
		Limit: types.AppLimit{
			PayPlan: types.PayPlan{
				Type:  types.FreetierV0,
				Limit: 250000,
			},
		},
	}

	appToSend, err := json.Marshal(rawAppToSend)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/application", bytes.NewBuffer(appToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	appToReturn := &types.Application{
		ID:     "60ddc61b6e29c3003378361E",
		UserID: "60ddc61b6e29c3003378361D",
		Limit: types.AppLimit{
			PayPlan: types.PayPlan{
				Type:  types.FreetierV0,
				Limit: 250000,
			},
		},
	}

	writerMock := driver.NewMockDriver(t)

	writerMock.On("WriteApplication", testCtx, mock.Anything).Return(appToReturn, nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	marshaledReturnApp, err := json.Marshal(&types.Application{
		ID:     "60ddc61b6e29c3003378361E",
		UserID: "60ddc61b6e29c3003378361D",
		Limit: types.AppLimit{
			PayPlan: types.PayPlan{
				Type:  types.FreetierV0,
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

	writerMock.On("WriteApplication", testCtx, mock.Anything).Return(appToReturn, errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func TestRouter_UpdateApplication(t *testing.T) {
	c := require.New(t)

	trueBool := true
	rawUpdateInput := &types.UpdateApplication{
		Name:   "pablo",
		Status: types.Orphaned,
		Limit: &types.AppLimit{
			PayPlan: types.PayPlan{
				Type:  types.PayAsYouGoV0,
				Limit: 0,
			},
		},
		FirstDateSurpassed: time.Now(),
		GatewaySettings: &types.UpdateGatewaySettings{
			SecretKey: "1234",
		},
		NotificationSettings: &types.UpdateNotificationSettings{
			Half: &trueBool,
		},
	}

	updateInputToSend, err := json.Marshal(rawUpdateInput)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPut, "/application/5f62b7d8be3591c4dea8566d", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	writerMock := driver.NewMockDriver(t)

	writerMock.On("UpdateApplication", testCtx, mock.Anything).Return(nil).Once()

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

	writerMock.On("UpdateApplication", testCtx, mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusUnprocessableEntity, rr.Code)
}

func TestRouter_UpdateFirstDateSurpassed(t *testing.T) {
	c := require.New(t)

	rawUpdateInput := &types.UpdateFirstDateSurpassed{
		ApplicationIDs:     []string{"5f62b7d8be3591c4dea8566d"},
		FirstDateSurpassed: time.Now(),
	}

	updateInputToSend, err := json.Marshal(rawUpdateInput)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/application/first_date_surpassed", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	writerMock := driver.NewMockDriver(t)

	writerMock.On("UpdateFirstDateSurpassed", testCtx, mock.Anything).Return(nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	writerMock.On("UpdateFirstDateSurpassed", testCtx, mock.Anything).Return(errors.New("dummy error")).Once()

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

	rawUpdateInput.ApplicationIDs = []string{"5f62b7d8be3591c4dea85664"}

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

func TestRouter_RemoveApplication(t *testing.T) {
	c := require.New(t)

	rawRemoveInput := &types.UpdateApplication{
		Remove: true,
	}

	updateInputToSend, err := json.Marshal(rawRemoveInput)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPut, "/application/5f62b7d8be3591c4dea8566d", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	writerMock := driver.NewMockDriver(t)

	writerMock.On("RemoveApplication", testCtx, mock.Anything).Return(nil).Once()

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

	writerMock.On("RemoveApplication", testCtx, mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func TestRouter_GetApplicationsByUserID(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/user/60ecb2bf67774900350d9c43/application", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*types.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
			Limit: types.AppLimit{
				PayPlan: types.PayPlan{
					Type:  types.FreetierV0,
					Limit: 250000,
				},
			},
			FirstDateSurpassed: time.Date(2022, time.July, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:     "5f62b7d8be3591c4dea8566a",
			UserID: "60ecb2bf67774900350d9c43",
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

func TestRouter_GetLoadbalancerByUserID(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/user/60ecb2bf67774900350d9c43/load_balancer", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	var marshaledBody []*types.LoadBalancer

	err = json.Unmarshal(rr.Body.Bytes(), &marshaledBody)
	c.NoError(err)

	c.Equal("60ecb2bf67774900350d9c42", marshaledBody[0].ID)

	req, err = http.NewRequest(http.MethodGet, "/user/60ecb2bf67774900350d9d43/load_balancer", nil)
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func TestRouter_GetBlockchains(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/blockchain", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*types.Blockchain{
		{
			ID: "0021",
			Redirects: []types.Redirect{
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
			Redirects: []types.Redirect{
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

func TestRouter_GetBlockchain(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/blockchain/0021", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal(&types.Blockchain{
		ID: "0021",
		Redirects: []types.Redirect{
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

func TestRouter_GetLoadBalancers(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/load_balancer", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	var marshaledBody []*types.LoadBalancer

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

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	var marshaledBody types.LoadBalancer

	err = json.Unmarshal(rr.Body.Bytes(), &marshaledBody)
	c.NoError(err)

	c.Equal("60ecb2bf67774900350d9c42", marshaledBody.ID)

	req, err = http.NewRequest(http.MethodGet, "/load_balancer/60fcb2bf67774900350d9c42", nil)
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func TestRouter_CreateLoadBalancer(t *testing.T) {
	c := require.New(t)

	rawLBToSend := &types.LoadBalancer{
		UserID: "60ddc61b6e29c3003378361D",
	}

	lbToSend, err := json.Marshal(rawLBToSend)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/load_balancer", bytes.NewBuffer(lbToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	lbToReturn := &types.LoadBalancer{
		ID:     "60ddc61b6e29c3003378361E",
		UserID: "60ddc61b6e29c3003378361D",
	}

	writerMock := driver.NewMockDriver(t)

	writerMock.On("WriteLoadBalancer", testCtx, mock.Anything).Return(lbToReturn, nil).Once()

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

	writerMock.On("WriteLoadBalancer", testCtx, mock.Anything).Return(lbToReturn, errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func TestRouter_UpdateLoadBalancer(t *testing.T) {
	c := require.New(t)

	trueBool := true
	rawUpdateInput := &types.UpdateLoadBalancer{
		Name: "pablo",
		StickyOptions: &types.UpdateStickyOptions{
			Duration:      "21",
			StickyOrigins: []string{"pjog"},
			StickyMax:     21,
			Stickiness:    &trueBool,
		},
	}

	updateInputToSend, err := json.Marshal(rawUpdateInput)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPut, "/load_balancer/60ecb2bf67774900350d9c42", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	writerMock := driver.NewMockDriver(t)

	writerMock.On("UpdateLoadBalancer", testCtx, mock.Anything).Return(nil).Once()

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

	writerMock.On("UpdateLoadBalancer", testCtx, mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func TestRouter_RemoveLoadBalancer(t *testing.T) {
	c := require.New(t)

	rawUpdateInput := &types.UpdateLoadBalancer{
		Remove: true,
	}

	updateInputToSend, err := json.Marshal(rawUpdateInput)
	c.NoError(err)

	router, err := newTestRouter(t)
	c.NoError(err)

	writerMock := driver.NewMockDriver(t)

	router.Writer = writerMock

	req, err := http.NewRequest(http.MethodPut, "/load_balancer/60ecb2bf67774900350d9c42", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	writerMock.On("RemoveLoadBalancer", testCtx, mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/load_balancer/60ecb2bf67774900350d9c42", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("RemoveLoadBalancer", testCtx, mock.Anything).Return(nil).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/load_balancer/5f62b7d8be3591c4dea85664", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func TestRouter_GetPayPlans(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/pay_plan", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*types.PayPlan{
		{
			Type:  types.FreetierV0,
			Limit: 250000,
		},
		{
			Type:  types.PayAsYouGoV0,
			Limit: 0,
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func TestRouter_GetPayPlan(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/pay_plan/freetier_v0", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal(&types.PayPlan{
		Type:  types.FreetierV0,
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

func TestRouter_CreateBlockchain(t *testing.T) {
	c := require.New(t)

	rawChainToSend := &types.Blockchain{
		Ticker: "POKT",
	}

	chainToSend, err := json.Marshal(rawChainToSend)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/blockchain", bytes.NewBuffer(chainToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	chainToReturn := &types.Blockchain{
		ID:     "60ddc61b6e29c3003378361E",
		Ticker: "POKT",
	}

	writerMock := driver.NewMockDriver(t)

	writerMock.On("WriteBlockchain", testCtx, mock.Anything).Return(chainToReturn, nil).Once()

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

	writerMock.On("WriteBlockchain", testCtx, mock.Anything).Return(chainToReturn, errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func TestRouter_CreateRedirect(t *testing.T) {
	c := require.New(t)

	rawRedirectsToSend := &types.Redirect{BlockchainID: "0021"}

	redirectToSend, err := json.Marshal(rawRedirectsToSend)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/redirect", bytes.NewBuffer(redirectToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	redirectToReturn := &types.Redirect{BlockchainID: "0021"}

	writerMock := driver.NewMockDriver(t)

	writerMock.On("WriteRedirect", testCtx, mock.Anything).Return(redirectToReturn, nil).Once()

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

	writerMock.On("WriteRedirect", testCtx, mock.Anything).Return(redirectToReturn, errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func TestRouter_ActivateBlockchain(t *testing.T) {
	c := require.New(t)

	activeStatusToSend, err := json.Marshal(true)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/blockchain/0021/activate", bytes.NewBuffer(activeStatusToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	writerMock := driver.NewMockDriver(t)

	writerMock.On("ActivateBlockchain", testCtx, mock.Anything).Return(nil).Once()

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

	writerMock.On("ActivateBlockchain", testCtx, mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}
