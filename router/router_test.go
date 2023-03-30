package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pokt-foundation/pocket-http-db/cache"
	"github.com/pokt-foundation/portal-db/v2/driver"
	"github.com/pokt-foundation/portal-db/v2/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestRouter(t *testing.T) (*Router, error) {
	readerMock := cache.NewReaderMock(t)

	readerMock.On("ReadPayPlans", mock.Anything).Return([]*types.PayPlan{
		{
			Type:  types.FreetierV0,
			Limit: 250000,
		},
		{
			Type:  types.PayAsYouGoV0,
			Limit: 0,
		},
	}, nil)

	readerMock.On("ReadApplications", mock.Anything).Return([]*types.Application{
		{
			ID:     "test_app_e3591c4dea8566d",
			UserID: "test_user_1dbffbdfeeb225",
			Limit: types.AppLimit{
				PayPlan: types.PayPlan{
					Type:  types.FreetierV0,
					Limit: 250000,
				},
			},
			FirstDateSurpassed: time.Date(2022, time.July, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:     "test_app_e3591c4dea8566a",
			UserID: "test_user_1dbffbdfeeb225",
		},
		{
			ID:     "test_app_9f69710adbd9a8a",
			UserID: "test_user_969710adbd9a8a",
		},
		{
			ID:     "test_app_616dfbe503c2975",
			UserID: "test_user_969710adbd9a8a",
		},
		{
			ID:     "test_app_f2dcb22a518d2ce",
			UserID: "test_user_04228205bd261a",
		},
		{
			ID:     "test_app_2b21b25b12ad62d",
			UserID: "test_user_04228205bd261a",
		},
	}, nil)

	readerMock.On("ReadBlockchains", mock.Anything).Return([]*types.Blockchain{
		{
			ID:          "0021",
			Description: "test-chain-21",
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
			ID:          "0022",
			Description: "test-chain-22",
			Redirects: []types.Redirect{
				{
					BlockchainID:   "0022",
					Alias:          "eth-mainnet",
					Domain:         "eth-mainnet.gateway.network",
					LoadBalancerID: "45678",
				},
			},
		},
	}, nil)

	readerMock.On("ReadLoadBalancers", mock.Anything).Return([]*types.LoadBalancer{
		{
			ID: "test_lb_15a3f7accb30783",
			ApplicationIDs: []string{
				"test_app_e3591c4dea8566d",
				"test_app_e3591c4dea8566a",
			},
			UserID: "test_user_1dbffbdfeeb225",
			Users: []types.UserAccess{
				{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "user1@test.com", Accepted: true},
			},
		},
		{
			ID: "test_lb_34987u329rfn23f",
			ApplicationIDs: []string{
				"test_app_9f69710adbd9a8a",
				"test_app_616dfbe503c2975",
			},
			UserID: "test_user_969710adbd9a8a",
			Users: []types.UserAccess{
				{RoleName: types.RoleOwner, UserID: "test_user_969710adbd9a8a", Email: "user2@test.com", Accepted: true},
				{RoleName: types.RoleMember, UserID: "", Email: "user1@test.com", Accepted: false},
			},
		},
		{
			ID: "test_lb_3890ru23jfi32fj",
			ApplicationIDs: []string{
				"test_app_f2dcb22a518d2ce",
				"test_app_2b21b25b12ad62d",
			},
			UserID: "test_user_04228205bd261a",
			Users: []types.UserAccess{
				{RoleName: types.RoleOwner, UserID: "test_user_04228205bd261a", Email: "user3@test.com", Accepted: true},
				{RoleName: types.RoleAdmin, UserID: "test_user_1dbffbdfeeb225", Email: "user1@test.com", Accepted: true},
			},
		},
		{
			ID: "test_lb_3890ru23jfi32fj2",
			ApplicationIDs: []string{
				"test_app_f2dcb22a518d2ce",
				"test_app_2b21b25b12ad62d",
			},
			UserID: "test_user_04228205bd261a",
			Users: []types.UserAccess{
				{RoleName: types.RoleOwner, UserID: "test_user_04228205bd261a", Email: "user3@test.com", Accepted: true},
				{RoleName: types.RoleMember, UserID: "test_user_1dbffbdfeeb225", Email: "user1@test.com", Accepted: false},
			},
		},
	}, nil)

	readerMock.On("ReadUserPermissions", mock.Anything).Return(map[types.UserID]*types.UserPermissions{
		"test_user_1dbffbdfeeb225": {
			UserID: "test_user_1dbffbdfeeb225",
			LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{
				"test_lb_15a3f7accb30783": {
					RoleName:    types.RoleOwner,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint},
				},
				"test_lb_34987u329rfn23f": {
					RoleName:    types.RoleMember,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint},
				},
				"test_lb_3890ru23jfi32fj": {
					RoleName:    types.RoleAdmin,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint},
				},
				"test_lb_3890ru23jfi32fj2": {
					RoleName:    types.RoleMember,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint},
				},
			},
		},
		"test_user_969710adbd9a8a": {
			UserID: "test_user_47fhf7rwejf943",
			LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{
				"test_lb_34987u329rfn23f": {
					RoleName:    types.RoleOwner,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint},
				},
			},
		},
		"test_user_04228205bd261a": {
			UserID: "test_user_774900350d9c43",
			LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{
				"test_lb_3890ru23jfi32fj": {
					RoleName:    types.RoleOwner,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint},
				},
			},
		},
	}, nil)

	testLog := logrus.New()

	cache, err := cache.NewCache(readerMock, time.Minute*5, testLog)
	if err != nil {
		return nil, err
	}

	return NewRouter(cache, readerMock, nil, map[string]bool{"": true}, testLog)
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

	req, err := http.NewRequest(http.MethodGet, "/v1/application", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*types.Application{
		{
			ID:     "test_app_e3591c4dea8566d",
			UserID: "test_user_1dbffbdfeeb225",
			Limit: types.AppLimit{
				PayPlan: types.PayPlan{
					Type:  types.FreetierV0,
					Limit: 250000,
				},
			},
			FirstDateSurpassed: time.Date(2022, time.July, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:     "test_app_e3591c4dea8566a",
			UserID: "test_user_1dbffbdfeeb225",
		},
		{
			ID:     "test_app_9f69710adbd9a8a",
			UserID: "test_user_969710adbd9a8a",
		},
		{
			ID:     "test_app_616dfbe503c2975",
			UserID: "test_user_969710adbd9a8a",
		},
		{
			ID:     "test_app_f2dcb22a518d2ce",
			UserID: "test_user_04228205bd261a",
		},
		{
			ID:     "test_app_2b21b25b12ad62d",
			UserID: "test_user_04228205bd261a",
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

	req, err := http.NewRequest(http.MethodGet, "/v1/application/test_app_e3591c4dea8566d", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal(&types.Application{
		ID:     "test_app_e3591c4dea8566d",
		UserID: "test_user_1dbffbdfeeb225",
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

	req, err = http.NewRequest(http.MethodGet, "/v1/application/5f62b7d8be3591c3dea8566d", nil)
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func TestRouter_CreateApplication(t *testing.T) {
	c := require.New(t)

	rawAppToSend := &types.Application{
		UserID: "test_user_29c3003378361D",
		Limit: types.AppLimit{
			PayPlan: types.PayPlan{
				Type:  types.FreetierV0,
				Limit: 250000,
			},
		},
	}

	appToSend, err := json.Marshal(rawAppToSend)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/v1/application", bytes.NewBuffer(appToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	appToReturn := &types.Application{
		ID:     "test_app_e29c3003378361E",
		UserID: "test_user_29c3003378361D",
		Limit: types.AppLimit{
			PayPlan: types.PayPlan{
				Type:  types.FreetierV0,
				Limit: 250000,
			},
		},
	}

	writerMock := driver.NewMockDriver(t)

	writerMock.On("WriteApplication", mock.Anything, mock.Anything).Return(appToReturn, nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	marshaledReturnApp, err := json.Marshal(&types.Application{
		ID:     "test_app_e29c3003378361E",
		UserID: "test_user_29c3003378361D",
		Limit: types.AppLimit{
			PayPlan: types.PayPlan{
				Type:  types.FreetierV0,
				Limit: 250000,
			},
		},
	})
	c.NoError(err)

	c.Equal(marshaledReturnApp, rr.Body.Bytes())

	req, err = http.NewRequest(http.MethodPost, "/v1/application", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPost, "/v1/application", bytes.NewBuffer(appToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("WriteApplication", mock.Anything, mock.Anything).Return(appToReturn, errors.New("dummy error")).Once()

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

	req, err := http.NewRequest(http.MethodPut, "/v1/application/test_app_e3591c4dea8566d", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	writerMock := driver.NewMockDriver(t)

	writerMock.On("UpdateApplication", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/v1/application/test_app_3591c4dea85664", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/v1/application/test_app_e3591c4dea8566d", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/v1/application/test_app_e3591c4dea8566d", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("UpdateApplication", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusUnprocessableEntity, rr.Code)
}

func TestRouter_UpdateFirstDateSurpassed(t *testing.T) {
	c := require.New(t)

	rawUpdateInput := &types.UpdateFirstDateSurpassed{
		ApplicationIDs:     []string{"test_app_e3591c4dea8566d"},
		FirstDateSurpassed: time.Now(),
	}

	updateInputToSend, err := json.Marshal(rawUpdateInput)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/v1/application/first_date_surpassed", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	writerMock := driver.NewMockDriver(t)

	writerMock.On("UpdateAppFirstDateSurpassed", mock.Anything, mock.Anything).Return(nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	writerMock.On("UpdateAppFirstDateSurpassed", mock.Anything, mock.Anything).Return(errors.New("dummy error")).Once()

	req, err = http.NewRequest(http.MethodPost, "/v1/application/first_date_surpassed", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)

	req, err = http.NewRequest(http.MethodPost, "/v1/application/first_date_surpassed", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	rawUpdateInput.ApplicationIDs = []string{"test_app_3591c4dea85664"}

	updateInputToSend, err = json.Marshal(rawUpdateInput)
	c.NoError(err)

	req, err = http.NewRequest(http.MethodPost, "/v1/application/first_date_surpassed", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)

	rawUpdateInput.ApplicationIDs = nil

	updateInputToSend, err = json.Marshal(rawUpdateInput)
	c.NoError(err)

	req, err = http.NewRequest(http.MethodPost, "/v1/application/first_date_surpassed", bytes.NewBuffer(updateInputToSend))
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

	req, err := http.NewRequest(http.MethodPut, "/v1/application/test_app_e3591c4dea8566d", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	writerMock := driver.NewMockDriver(t)

	writerMock.On("RemoveApplication", mock.Anything, mock.Anything).Return(nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/v1/application/test_app_3591c4dea85664", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/v1/application/test_app_e3591c4dea8566d", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/v1/application/test_app_e3591c4dea8566d", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("RemoveApplication", mock.Anything, mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func TestRouter_GetApplicationsByUserID(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/v1/user/test_user_1dbffbdfeeb225/application", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*types.Application{
		{
			ID:     "test_app_e3591c4dea8566d",
			UserID: "test_user_1dbffbdfeeb225",
			Limit: types.AppLimit{
				PayPlan: types.PayPlan{
					Type:  types.FreetierV0,
					Limit: 250000,
				},
			},
			FirstDateSurpassed: time.Date(2022, time.July, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:     "test_app_e3591c4dea8566a",
			UserID: "test_user_1dbffbdfeeb225",
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())

	req, err = http.NewRequest(http.MethodGet, "/v1/user/61ecb2bf67774900350d9c43/application", nil)
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

// TODO - rewrite all Router & Cache tests as table-driven like the test below
func TestRouter_GetLoadBalancersByUserID(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		roleNameFilter types.RoleName
		expectedLBs    []*types.LoadBalancer
		expectedStatus int
		err            error
	}{
		{
			name:   "Should fetch all load balancers for a given user when no filter provided",
			userID: "test_user_1dbffbdfeeb225",
			expectedLBs: []*types.LoadBalancer{
				{
					ID:     "test_lb_15a3f7accb30783",
					UserID: "test_user_1dbffbdfeeb225",
					Applications: []*types.Application{
						{ID: "test_app_e3591c4dea8566d", UserID: "test_user_1dbffbdfeeb225", Limit: types.AppLimit{PayPlan: types.PayPlan{Type: types.FreetierV0, Limit: 250000}}, FirstDateSurpassed: time.Date(2022, time.July, 21, 0, 0, 0, 0, time.UTC)},
						{ID: "test_app_e3591c4dea8566a", UserID: "test_user_1dbffbdfeeb225"},
					},
					Users: []types.UserAccess{
						{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "user1@test.com", Accepted: true},
					},
				},
				{
					ID:     "test_lb_3890ru23jfi32fj",
					UserID: "test_user_04228205bd261a",
					Applications: []*types.Application{
						{
							ID:     "test_app_f2dcb22a518d2ce",
							UserID: "test_user_04228205bd261a",
						},
						{
							ID:     "test_app_2b21b25b12ad62d",
							UserID: "test_user_04228205bd261a",
						},
					},
					Users: []types.UserAccess{
						{RoleName: types.RoleOwner, UserID: "test_user_04228205bd261a", Email: "user3@test.com", Accepted: true},
						{RoleName: types.RoleAdmin, UserID: "test_user_1dbffbdfeeb225", Email: "user1@test.com", Accepted: true},
					},
				},
			},
			expectedStatus: http.StatusOK,
			err:            nil,
		},
		{
			name:           "Should fetch only load balancers a given user is the OWNER of when OWNER filter provided",
			userID:         "test_user_1dbffbdfeeb225",
			roleNameFilter: types.RoleOwner,
			expectedLBs: []*types.LoadBalancer{
				{
					ID:     "test_lb_15a3f7accb30783",
					UserID: "test_user_1dbffbdfeeb225",
					Applications: []*types.Application{
						{ID: "test_app_e3591c4dea8566d", UserID: "test_user_1dbffbdfeeb225", Limit: types.AppLimit{PayPlan: types.PayPlan{Type: types.FreetierV0, Limit: 250000}}, FirstDateSurpassed: time.Date(2022, time.July, 21, 0, 0, 0, 0, time.UTC)},
						{ID: "test_app_e3591c4dea8566a", UserID: "test_user_1dbffbdfeeb225"},
					},
					Users: []types.UserAccess{
						{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "user1@test.com", Accepted: true},
					},
				},
			},
			expectedStatus: http.StatusOK,
			err:            nil,
		},
		{
			name:           "Should fetch only load balancers a given user is an ADMIN of when ADMIN filter provided",
			userID:         "test_user_1dbffbdfeeb225",
			roleNameFilter: types.RoleAdmin,
			expectedLBs: []*types.LoadBalancer{
				{
					ID:     "test_lb_3890ru23jfi32fj",
					UserID: "test_user_04228205bd261a",
					Applications: []*types.Application{
						{ID: "test_app_f2dcb22a518d2ce", UserID: "test_user_04228205bd261a"},
						{ID: "test_app_2b21b25b12ad62d", UserID: "test_user_04228205bd261a"},
					},
					Users: []types.UserAccess{
						{RoleName: types.RoleOwner, UserID: "test_user_04228205bd261a", Email: "user3@test.com", Accepted: true},
						{RoleName: types.RoleAdmin, UserID: "test_user_1dbffbdfeeb225", Email: "user1@test.com", Accepted: true},
					},
				},
			},
			expectedStatus: http.StatusOK,
			err:            nil,
		},
		{
			name:           "Should not fetch a load balancer for a given user if they have not accepted their invite",
			userID:         "test_user_1dbffbdfeeb225",
			roleNameFilter: types.RoleMember,
			expectedLBs:    []*types.LoadBalancer{},
			expectedStatus: http.StatusOK,
			err:            nil,
		},
		{
			name:           "Should fail if an invalid role name filter query param is provided",
			userID:         "test_user_969710adbd9a8a",
			roleNameFilter: types.RoleName("test_sir_not_appearing_in_this_film"),
			expectedStatus: http.StatusBadRequest,
			err:            errors.New("dummy error"),
		},
		{
			name:           "Should fail if the user has no load balancers",
			userID:         "test_sir_not_appearing_in_this_film",
			expectedStatus: http.StatusNotFound,
			err:            errors.New("dummy error"),
		},
	}

	for _, test := range tests {
		c := require.New(t)

		router, err := newTestRouter(t)
		c.NoError(err)

		rr := httptest.NewRecorder()

		url := fmt.Sprintf("/v1/user/%s/load_balancer", test.userID)
		if test.roleNameFilter != types.RoleName("") {
			url = fmt.Sprintf("%s?filter=%s", url, string(test.roleNameFilter))
		}
		req, err := http.NewRequest(http.MethodGet, url, nil)
		c.NoError(err)

		router.Router.ServeHTTP(rr, req)
		c.Equal(test.expectedStatus, rr.Code)

		if test.err == nil {
			var lbResp []*types.LoadBalancer
			err := json.Unmarshal(rr.Body.Bytes(), &lbResp)

			c.NoError(err)
			c.Equal(test.expectedLBs, lbResp)
		}
	}
}

func TestRouter_GetPendingLoadBalancersByEmail(t *testing.T) {
	tests := []struct {
		name           string
		userEmail      string
		expectedLBs    []*types.LoadBalancer
		expectedStatus int
		err            error
	}{
		{
			name:      "Should fetch all pending load balancers for a given email",
			userEmail: "user1@test.com",
			expectedLBs: []*types.LoadBalancer{
				{
					ID: "test_lb_34987u329rfn23f",
					Applications: []*types.Application{
						{ID: "test_app_9f69710adbd9a8a", UserID: "test_user_969710adbd9a8a"},
						{ID: "test_app_616dfbe503c2975", UserID: "test_user_969710adbd9a8a"},
					},
					UserID: "test_user_969710adbd9a8a",
					Users: []types.UserAccess{
						{RoleName: types.RoleOwner, UserID: "test_user_969710adbd9a8a", Email: "user2@test.com", Accepted: true},
						{RoleName: types.RoleMember, UserID: "", Email: "user1@test.com", Accepted: false},
					},
				},
				{
					ID: "test_lb_3890ru23jfi32fj2",

					Applications: []*types.Application{
						{ID: "test_app_f2dcb22a518d2ce", UserID: "test_user_04228205bd261a"},
						{ID: "test_app_2b21b25b12ad62d", UserID: "test_user_04228205bd261a"},
					},
					UserID: "test_user_04228205bd261a",
					Users: []types.UserAccess{
						{RoleName: types.RoleOwner, UserID: "test_user_04228205bd261a", Email: "user3@test.com", Accepted: true},
						{RoleName: types.RoleMember, UserID: "test_user_1dbffbdfeeb225", Email: "user1@test.com", Accepted: false},
					},
				},
			},
			expectedStatus: http.StatusOK,
			err:            nil,
		},
		{
			name:           "Should not fetch a load balancer for a given email if they have already accepted all their invites",
			userEmail:      "test_user_969710adbd9a8a",
			expectedStatus: http.StatusNotFound,
			err:            nil,
		},
		{
			name:           "Should fail if the email has no load balancers",
			userEmail:      "test_sir@not_appearing_in_this_film.com",
			expectedStatus: http.StatusNotFound,
			err:            errors.New("dummy error"),
		},
	}

	for _, test := range tests {
		c := require.New(t)

		router, err := newTestRouter(t)
		c.NoError(err)

		rr := httptest.NewRecorder()

		url := fmt.Sprintf("/v1/user/%s/load_balancer/pending", test.userEmail)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		c.NoError(err)

		router.Router.ServeHTTP(rr, req)
		c.Equal(test.expectedStatus, rr.Code)

		if rr.Code == http.StatusOK {
			var lbResp []*types.LoadBalancer
			err := json.Unmarshal(rr.Body.Bytes(), &lbResp)

			c.NoError(err)
			c.Equal(test.expectedLBs, lbResp)
		}
	}
}

// TODO - rewrite all Router & Cache tests as table-driven like the test below
func TestRouter_GetUserPermissionsByUserID(t *testing.T) {
	tests := []struct {
		name                string
		userID              string
		expectedPermissions *types.UserPermissions
		expectedStatus      int
		err                 error
	}{
		{
			name:   "Should fetch all load balancer permissions for a given user",
			userID: "test_user_1dbffbdfeeb225",
			expectedPermissions: &types.UserPermissions{
				UserID: "test_user_1dbffbdfeeb225",
				LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{
					"test_lb_15a3f7accb30783": {
						RoleName: types.RoleOwner,
						Permissions: []types.PermissionsEnum{
							"read:endpoint",
							"write:endpoint",
						},
					},
					"test_lb_34987u329rfn23f": {
						RoleName: types.RoleMember,
						Permissions: []types.PermissionsEnum{
							"read:endpoint",
						},
					},
					"test_lb_3890ru23jfi32fj": {
						RoleName: types.RoleAdmin,
						Permissions: []types.PermissionsEnum{
							"read:endpoint",
							"write:endpoint",
						},
					},
					"test_lb_3890ru23jfi32fj2": {
						RoleName: types.RoleMember,
						Permissions: []types.PermissionsEnum{
							"read:endpoint",
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			err:            nil,
		},
		{
			name:   "Should fetch all load balancer permissions for a given user",
			userID: "test_user_969710adbd9a8a",
			expectedPermissions: &types.UserPermissions{
				UserID: "test_user_47fhf7rwejf943",
				LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{
					"test_lb_34987u329rfn23f": {
						RoleName:    types.RoleOwner,
						Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint},
					},
				},
			},
			expectedStatus: http.StatusOK,
			err:            nil,
		},
		{
			name:           "Should fail if a user has no permissions in the cache",
			userID:         "test_not_a_real_user",
			expectedStatus: http.StatusNotFound,
			err:            errPermissionsNotFound,
		},
	}

	for _, test := range tests {
		c := require.New(t)

		router, err := newTestRouter(t)
		c.NoError(err)

		rr := httptest.NewRecorder()

		url := fmt.Sprintf("/v1/user/%s/permission", test.userID)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		c.NoError(err)

		router.Router.ServeHTTP(rr, req)
		c.Equal(test.expectedStatus, rr.Code)

		if test.err == nil {
			var permsResponse *types.UserPermissions
			err := json.Unmarshal(rr.Body.Bytes(), &permsResponse)
			c.NoError(err)
			c.Equal(test.expectedPermissions, permsResponse)
		}
	}
}

func TestRouter_GetBlockchains(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/v1/blockchain", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*types.Blockchain{
		{
			ID:          "0021",
			Description: "test-chain-21",
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
			ID:          "0022",
			Description: "test-chain-22",
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

	req, err := http.NewRequest(http.MethodGet, "/v1/blockchain/0021", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal(&types.Blockchain{
		ID:          "0021",
		Description: "test-chain-21",
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

	req, err = http.NewRequest(http.MethodGet, "/v1/blockchain/00210", nil)
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func TestRouter_GetLoadBalancers(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/v1/load_balancer", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	var marshaledBody []*types.LoadBalancer

	err = json.Unmarshal(rr.Body.Bytes(), &marshaledBody)
	c.NoError(err)

	c.Len(marshaledBody, 4)
	c.Equal("test_lb_15a3f7accb30783", marshaledBody[0].ID)
	c.Equal("test_lb_34987u329rfn23f", marshaledBody[1].ID)
	c.Equal("test_lb_3890ru23jfi32fj", marshaledBody[2].ID)
}

func TestRouter_GetLoadBalancer(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/v1/load_balancer/test_lb_34987u329rfn23f", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	var marshaledBody types.LoadBalancer

	err = json.Unmarshal(rr.Body.Bytes(), &marshaledBody)
	c.NoError(err)

	c.Equal("test_lb_34987u329rfn23f", marshaledBody.ID)

	req, err = http.NewRequest(http.MethodGet, "/v1/load_balancer/60fcb2bf67774900350d9c42", nil)
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func TestRouter_CreateLoadBalancer(t *testing.T) {
	c := require.New(t)

	rawLBToSend := &types.LoadBalancer{
		UserID: "test_user_29c3003378361D",
		Users: []types.UserAccess{
			{
				UserID:   "test_user_29c3003378361D",
				RoleName: types.RoleOwner,
				Email:    "user1@test.com",
			},
		},
	}

	lbToSend, err := json.Marshal(rawLBToSend)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/v1/load_balancer", bytes.NewBuffer(lbToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	lbToReturn := &types.LoadBalancer{
		ID:     "test_app_e29c3003378361E",
		UserID: "test_user_29c3003378361D",
		Users: []types.UserAccess{
			{
				UserID:   "test_user_29c3003378361D",
				RoleName: types.RoleOwner,
				Email:    "user1@test.com",
			},
		},
	}

	writerMock := driver.NewMockDriver(t)

	writerMock.On("WriteLoadBalancer", mock.Anything, mock.Anything).Return(lbToReturn, nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	marshaledReturnLB, err := json.Marshal(lbToReturn)
	c.NoError(err)

	c.Equal(marshaledReturnLB, rr.Body.Bytes())

	req, err = http.NewRequest(http.MethodPost, "/v1/load_balancer", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPost, "/v1/load_balancer", bytes.NewBuffer(lbToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("WriteLoadBalancer", mock.Anything, mock.Anything).Return(lbToReturn, errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

// TODO - rewrite all Router & Cache tests as table-driven like the test below
func TestRouter_CreateLoadBalancerUser(t *testing.T) {
	tests := []struct {
		name                  string
		lbID, ownerID         string
		userInput             types.UserAccess
		invalidJSON           bool
		expectedLBAfterCreate *types.LoadBalancer
		expectedStatus        int
		err                   error
	}{
		{
			name:    "Should create a new User for a LoadBalancer with correct input",
			lbID:    "test_lb_34987u329rfn23f",
			ownerID: "test_user_969710adbd9a8a",
			userInput: types.UserAccess{
				UserID:   "test_user_member1234",
				RoleName: types.RoleMember,
				Email:    "member1@test.com",
			},
			expectedLBAfterCreate: &types.LoadBalancer{
				ID:     "test_lb_34987u329rfn23f",
				UserID: "test_user_969710adbd9a8a",
				Applications: []*types.Application{
					{ID: "test_app_9f69710adbd9a8a", UserID: "test_user_969710adbd9a8a"},
					{ID: "test_app_616dfbe503c2975", UserID: "test_user_969710adbd9a8a"},
				},
				Users: []types.UserAccess{
					{RoleName: types.RoleOwner, UserID: "test_user_969710adbd9a8a", Email: "user2@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "", Email: "user1@test.com", Accepted: false},
					{RoleName: types.RoleMember, UserID: "", Email: "member1@test.com", Accepted: false},
				},
			},
			expectedStatus: http.StatusOK,
			err:            nil,
		},
		{
			name:           "Should fail if the load balancer does not exist",
			lbID:           "test_sir_not_appearing_in_this_film",
			expectedStatus: http.StatusNotFound,
			err:            errors.New("dummy error"),
		},
		{
			name:           "Should fail if the data cannot be marshaled to JSON",
			lbID:           "test_lb_34987u329rfn23f",
			invalidJSON:    true,
			expectedStatus: http.StatusBadRequest,
			err:            errors.New("dummy error"),
		},
		{
			name: "Should fail if DB driver throws an error",
			lbID: "test_lb_34987u329rfn23f",
			userInput: types.UserAccess{
				UserID:   "test_user_member1234",
				RoleName: types.RoleMember,
				Email:    "member1@test.com",
			},
			expectedStatus: http.StatusInternalServerError,
			err:            errors.New("dummy database driver error"),
		},
	}

	for _, test := range tests {
		c := require.New(t)

		router, err := newTestRouter(t)
		c.NoError(err)
		writerMock := driver.NewMockDriver(t)
		router.Writer = writerMock

		if (test.userInput != types.UserAccess{}) {
			writerMock.On("WriteLoadBalancerUser", mock.Anything, mock.Anything, mock.Anything).Return(test.err)
		}

		rr := httptest.NewRecorder()

		userToSend, err := json.Marshal(test.userInput)
		c.NoError(err)

		var testBuffer *bytes.Buffer
		switch {
		case !test.invalidJSON:
			testBuffer = bytes.NewBuffer(userToSend)
		case test.invalidJSON:
			testBuffer = bytes.NewBuffer([]byte("invalid JSON"))
		}
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/v1/load_balancer/%s/user", test.lbID), testBuffer)
		c.NoError(err)

		router.Router.ServeHTTP(rr, req)
		c.Equal(test.expectedStatus, rr.Code)

		if test.err == nil {
			var lbResp types.LoadBalancer
			err := json.Unmarshal(rr.Body.Bytes(), &lbResp)
			c.NoError(err)
			c.Equal(test.expectedLBAfterCreate, &lbResp)
		}
	}
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

	req, err := http.NewRequest(http.MethodPut, "/v1/load_balancer/test_lb_34987u329rfn23f", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	writerMock := driver.NewMockDriver(t)

	writerMock.On("UpdateLoadBalancer", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/v1/load_balancer/test_app_3591c4dea85664", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/v1/load_balancer/test_lb_34987u329rfn23f", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/v1/load_balancer/test_lb_34987u329rfn23f", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("UpdateLoadBalancer", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

// TODO - rewrite all Router & Cache tests as table-driven like the test below
func TestRouter_UpdateLoadBalancerUserRole(t *testing.T) {
	tests := []struct {
		name                  string
		lbID                  string
		userUpdate            types.UpdateUserAccess
		invalidJSON           bool
		expectedLBAfterUpdate *types.LoadBalancer
		expectedStatus        int
		err                   error
	}{
		{
			name: "Should update User for a LoadBalancer with a new role with correct input",
			lbID: "test_lb_3890ru23jfi32fj",
			userUpdate: types.UpdateUserAccess{
				Email:    "user1@test.com",
				RoleName: types.RoleMember,
			},
			expectedLBAfterUpdate: &types.LoadBalancer{
				ID: "test_lb_3890ru23jfi32fj",
				Applications: []*types.Application{
					{ID: "test_app_f2dcb22a518d2ce", UserID: "test_user_04228205bd261a"},
					{ID: "test_app_2b21b25b12ad62d", UserID: "test_user_04228205bd261a"},
				},
				UserID: "test_user_04228205bd261a",
				Users: []types.UserAccess{
					{RoleName: types.RoleOwner, UserID: "test_user_04228205bd261a", Email: "user3@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "test_user_1dbffbdfeeb225", Email: "user1@test.com", Accepted: true},
				},
			},
			expectedStatus: http.StatusOK,
			err:            nil,
		},
		{
			name:           "Should fail if the load balancer does not exist",
			lbID:           "test_sir_not_appearing_in_this_film",
			expectedStatus: http.StatusNotFound,
			err:            errors.New("dummy error"),
		},
		{
			name:           "Should fail if the data cannot be marshaled to JSON",
			lbID:           "test_lb_34987u329rfn23f",
			invalidJSON:    true,
			expectedStatus: http.StatusBadRequest,
			err:            errors.New("dummy error"),
		},
		{
			name: "Should fail if DB driver throws an error",
			lbID: "test_lb_34987u329rfn23f",
			userUpdate: types.UpdateUserAccess{
				UserID:   "test_user_969710adbd9a8a",
				RoleName: types.RoleAdmin,
			},
			expectedStatus: http.StatusInternalServerError,
			err:            errors.New("dummy database driver error"),
		},
	}

	for _, test := range tests {
		c := require.New(t)

		router, err := newTestRouter(t)
		c.NoError(err)
		writerMock := driver.NewMockDriver(t)
		router.Writer = writerMock

		if (test.userUpdate != types.UpdateUserAccess{}) {
			writerMock.On("UpdateUserAccessRole", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.err)
		}

		rr := httptest.NewRecorder()

		userToSend, err := json.Marshal(test.userUpdate)
		c.NoError(err)

		var testBuffer *bytes.Buffer
		switch {
		case !test.invalidJSON:
			testBuffer = bytes.NewBuffer(userToSend)
		case test.invalidJSON:
			testBuffer = bytes.NewBuffer([]byte("invalid JSON"))
		}
		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/v1/load_balancer/%s/user", test.lbID), testBuffer)
		c.NoError(err)

		router.Router.ServeHTTP(rr, req)
		c.Equal(test.expectedStatus, rr.Code)

		if test.err == nil {
			var lbResp types.LoadBalancer
			err := json.Unmarshal(rr.Body.Bytes(), &lbResp)
			c.NoError(err)
			c.Equal(test.expectedLBAfterUpdate, &lbResp)
		}
	}
}

// TODO - rewrite all Router & Cache tests as table-driven like the test below
func TestRouter_AcceptLoadBalancerUser(t *testing.T) {
	tests := []struct {
		name                         string
		lbID                         string
		acceptUserUpdate             types.UpdateUserAccess
		expectedUsersAfterUpdate     []types.UserAccess
		expectedRemainingInvitations []*types.LoadBalancer
		expectedStatus               int
		runMock                      bool
		err                          error
	}{
		{
			name: "Should update User ID and Accepted field for a LoadBalancer",
			lbID: "test_lb_34987u329rfn23f",
			acceptUserUpdate: types.UpdateUserAccess{
				Email:  "user1@test.com",
				UserID: "test_user_1dbffbdfeeb225",
			},
			expectedUsersAfterUpdate: []types.UserAccess{
				{RoleName: types.RoleOwner, UserID: "test_user_969710adbd9a8a", Email: "user2@test.com", Accepted: true},
				{RoleName: types.RoleMember, UserID: "test_user_1dbffbdfeeb225", Email: "user1@test.com", Accepted: true},
			},
			expectedRemainingInvitations: []*types.LoadBalancer{
				{
					ID: "test_lb_3890ru23jfi32fj2",
					Applications: []*types.Application{
						{ID: "test_app_f2dcb22a518d2ce", UserID: "test_user_04228205bd261a"},
						{ID: "test_app_2b21b25b12ad62d", UserID: "test_user_04228205bd261a"},
					},
					UserID: "test_user_04228205bd261a",
					Users: []types.UserAccess{
						{RoleName: types.RoleOwner, UserID: "test_user_04228205bd261a", Email: "user3@test.com", Accepted: true},
						{RoleName: types.RoleMember, UserID: "test_user_1dbffbdfeeb225", Email: "user1@test.com", Accepted: false},
					},
				},
			},
			expectedStatus: http.StatusOK,
			runMock:        true,
			err:            nil,
		},
		{
			name:           "Should fail if the load balancer does not exist",
			lbID:           "test_not_real_lb",
			expectedStatus: http.StatusNotFound,
			err:            errors.New("dummy error"),
		},
		{
			name: "Should fail if DB driver throws an error",
			lbID: "test_lb_34987u329rfn23f",
			acceptUserUpdate: types.UpdateUserAccess{
				Email:  "user1@test.com",
				UserID: "test_user_1dbffbdfeeb225",
			},
			expectedStatus: http.StatusInternalServerError,
			runMock:        true,
			err:            errors.New("dummy database driver error"),
		},
	}

	for _, test := range tests {
		c := require.New(t)

		router, err := newTestRouter(t)
		c.NoError(err)
		writerMock := driver.NewMockDriver(t)
		router.Writer = writerMock

		if test.runMock {
			writerMock.On("AcceptUserAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.err)
		}

		rr := httptest.NewRecorder()

		acceptUserUpdate, err := json.Marshal(test.acceptUserUpdate)
		c.NoError(err)

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/v1/load_balancer/%s/user/accept", test.lbID), bytes.NewBuffer(acceptUserUpdate))
		c.NoError(err)

		router.Router.ServeHTTP(rr, req)
		c.Equal(test.expectedStatus, rr.Code)

		if test.err == nil {
			var lbResp types.LoadBalancer
			err := json.Unmarshal(rr.Body.Bytes(), &lbResp)
			c.NoError(err)
			c.Equal(test.expectedUsersAfterUpdate, lbResp.Users)

			var invitationsResponse []*types.LoadBalancer
			req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/user/%s/load_balancer/pending", test.acceptUserUpdate.Email), nil)
			c.NoError(err)

			rr = httptest.NewRecorder()
			router.Router.ServeHTTP(rr, req)

			err = json.Unmarshal(rr.Body.Bytes(), &invitationsResponse)
			c.NoError(err)
			c.Equal(test.expectedRemainingInvitations, invitationsResponse)
		}
	}
}

// TODO - rewrite all Router & Cache tests as table-driven like the test below
func TestRouter_DeleteLoadBalancerUser(t *testing.T) {
	tests := []struct {
		name                  string
		lbID                  string
		emailOfUserToDelete   string
		invalidJSON           bool
		expectedLBAfterDelete *types.LoadBalancer
		expectedStatus        int
		runMock               bool
		err                   error
	}{
		{
			name:                "Should delete User for a LoadBalancer with correct input",
			lbID:                "test_lb_3890ru23jfi32fj",
			emailOfUserToDelete: "user1@test.com",
			expectedLBAfterDelete: &types.LoadBalancer{
				ID:     "test_lb_3890ru23jfi32fj",
				UserID: "test_user_04228205bd261a",
				Applications: []*types.Application{
					{ID: "test_app_f2dcb22a518d2ce", UserID: "test_user_04228205bd261a"},
					{ID: "test_app_2b21b25b12ad62d", UserID: "test_user_04228205bd261a"},
				},
				Users: []types.UserAccess{
					{RoleName: types.RoleOwner, UserID: "test_user_04228205bd261a", Email: "user3@test.com", Accepted: true},
				},
			},
			expectedStatus: http.StatusOK,
			runMock:        true,
			err:            nil,
		},
		{
			name:           "Should fail if the load balancer does not exist",
			lbID:           "test_sir_not_appearing_in_this_film",
			expectedStatus: http.StatusNotFound,
			err:            errors.New("dummy error"),
		},
		{
			name:                "Should fail if DB driver throws an error",
			lbID:                "test_lb_3890ru23jfi32fj",
			emailOfUserToDelete: "user1@test.com",
			expectedStatus:      http.StatusInternalServerError,
			runMock:             true,
			err:                 errors.New("dummy database driver error"),
		},
	}

	for _, test := range tests {
		c := require.New(t)

		router, err := newTestRouter(t)
		c.NoError(err)
		writerMock := driver.NewMockDriver(t)
		router.Writer = writerMock

		if test.runMock {
			writerMock.On("RemoveUserAccess", mock.Anything, mock.Anything, mock.Anything).Return(test.err)
		}

		rr := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/load_balancer/%s/user/%s", test.lbID, test.emailOfUserToDelete), nil)
		c.NoError(err)

		router.Router.ServeHTTP(rr, req)
		c.Equal(test.expectedStatus, rr.Code)

		if test.err == nil {
			var lbResp types.LoadBalancer
			err := json.Unmarshal(rr.Body.Bytes(), &lbResp)
			c.NoError(err)
			c.Equal(test.expectedLBAfterDelete, &lbResp)
		}
	}
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

	req, err := http.NewRequest(http.MethodPut, "/v1/load_balancer/test_lb_34987u329rfn23f", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	writerMock.On("RemoveLoadBalancer", mock.Anything, mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/v1/load_balancer/test_lb_34987u329rfn23f", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("RemoveLoadBalancer", mock.Anything, mock.Anything).Return(nil).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	req, err = http.NewRequest(http.MethodPut, "/v1/load_balancer/test_app_3591c4dea85664", bytes.NewBuffer(updateInputToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusNotFound, rr.Code)
}

func TestRouter_GetPayPlans(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/v1/pay_plan", nil)
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

	req, err := http.NewRequest(http.MethodGet, "/v1/pay_plan/freetier_v0", nil)
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

	req, err = http.NewRequest(http.MethodGet, "/v1/pay_plan/freetier_v21", nil)
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

	req, err := http.NewRequest(http.MethodPost, "/v1/blockchain", bytes.NewBuffer(chainToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	chainToReturn := &types.Blockchain{
		ID:     "test_app_e29c3003378361E",
		Ticker: "POKT",
	}

	writerMock := driver.NewMockDriver(t)

	writerMock.On("WriteBlockchain", mock.Anything, mock.Anything).Return(chainToReturn, nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	marshaledReturnChain, err := json.Marshal(chainToReturn)
	c.NoError(err)

	c.Equal(marshaledReturnChain, rr.Body.Bytes())

	req, err = http.NewRequest(http.MethodPost, "/v1/blockchain", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPost, "/v1/blockchain", bytes.NewBuffer(chainToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("WriteBlockchain", mock.Anything, mock.Anything).Return(chainToReturn, errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func TestRouter_CreateRedirect(t *testing.T) {
	c := require.New(t)

	rawRedirectsToSend := &types.Redirect{BlockchainID: "0021"}

	redirectToSend, err := json.Marshal(rawRedirectsToSend)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/v1/blockchain/redirect", bytes.NewBuffer(redirectToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	redirectToReturn := &types.Redirect{BlockchainID: "0021"}

	writerMock := driver.NewMockDriver(t)

	writerMock.On("WriteRedirect", mock.Anything, mock.Anything).Return(redirectToReturn, nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	marshaledReturnRedirect, err := json.Marshal(redirectToReturn)
	c.NoError(err)

	c.Equal(marshaledReturnRedirect, rr.Body.Bytes())

	req, err = http.NewRequest(http.MethodPost, "/v1/blockchain/redirect", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPost, "/v1/blockchain/redirect", bytes.NewBuffer(redirectToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("WriteRedirect", mock.Anything, mock.Anything).Return(redirectToReturn, errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

// TODO - rewrite all Router & Cache tests as table-driven like the test below
func TestRouter_UpdateBlockchain(t *testing.T) {
	tests := []struct {
		name                          string
		blockchainID                  string
		blockchainUpdate              types.UpdateBlockchain
		expectedBlockchainAfterUpdate types.Blockchain
		expectedStatus                int
		runMock                       bool
		err                           error
	}{
		{
			name:         "Should update User ID and Accepted field for a LoadBalancer",
			blockchainID: "0021",
			blockchainUpdate: types.UpdateBlockchain{
				Description: "test-chain-21-updated",
				Ticker:      "TEST-123",
				Altruist:    "https://wwww.whatever.com/rpc",
				ResultKey:   "test-key",
			},
			expectedBlockchainAfterUpdate: types.Blockchain{
				ID:          "0021",
				Description: "test-chain-21-updated",
				Ticker:      "TEST-123",
				Altruist:    "https://wwww.whatever.com/rpc",
				Redirects: []types.Redirect{
					{
						BlockchainID:   "0021",
						Alias:          "pokt-mainnet",
						Domain:         "pokt-mainnet.gateway.network",
						LoadBalancerID: "12345",
					},
				},
				SyncCheckOptions: types.SyncCheckOptions{
					ResultKey: "test-key",
				},
			},
			expectedStatus: http.StatusOK,
			runMock:        true,
			err:            nil,
		},
		{
			name:           "Should fail if the blockchain does not exist",
			blockchainID:   "test_not_real_chain",
			expectedStatus: http.StatusNotFound,
			err:            errors.New("dummy error"),
		},
		{
			name:         "Should fail if DB driver throws an error",
			blockchainID: "0021",
			blockchainUpdate: types.UpdateBlockchain{
				Description: "test-chain-21-updated",
				Ticker:      "TEST-123",
				Altruist:    "https://wwww.whatever.com/rpc",
				ResultKey:   "test-key",
			},
			expectedStatus: http.StatusInternalServerError,
			runMock:        true,
			err:            errors.New("dummy database driver error"),
		},
	}

	for _, test := range tests {
		c := require.New(t)

		router, err := newTestRouter(t)
		c.NoError(err)
		writerMock := driver.NewMockDriver(t)
		router.Writer = writerMock

		if test.runMock {
			writerMock.On("UpdateChain", mock.Anything, mock.Anything, mock.Anything).Return(test.err)
		}

		rr := httptest.NewRecorder()

		chainUpdateJSON, err := json.Marshal(test.blockchainUpdate)
		c.NoError(err)

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/v1/blockchain/%s", test.blockchainID), bytes.NewBuffer(chainUpdateJSON))
		c.NoError(err)

		router.Router.ServeHTTP(rr, req)
		c.Equal(test.expectedStatus, rr.Code)

		if test.err == nil {
			var chainResp types.Blockchain
			err := json.Unmarshal(rr.Body.Bytes(), &chainResp)
			c.NoError(err)
			c.Equal(test.expectedBlockchainAfterUpdate, chainResp)
		}
	}
}

func TestRouter_ActivateBlockchain(t *testing.T) {
	c := require.New(t)

	activeStatusToSend, err := json.Marshal(true)
	c.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/v1/blockchain/0021/activate", bytes.NewBuffer(activeStatusToSend))
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter(t)
	c.NoError(err)

	writerMock := driver.NewMockDriver(t)

	writerMock.On("ActivateChain", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	router.Writer = writerMock

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	req, err = http.NewRequest(http.MethodPost, "/v1/blockchain/0021/activate", bytes.NewBuffer([]byte("wrong")))
	c.NoError(err)

	rr = httptest.NewRecorder()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusBadRequest, rr.Code)

	req, err = http.NewRequest(http.MethodPost, "/v1/blockchain/0021/activate", bytes.NewBuffer(activeStatusToSend))
	c.NoError(err)

	rr = httptest.NewRecorder()

	writerMock.On("ActivateChain", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("dummy error")).Once()

	router.Router.ServeHTTP(rr, req)

	c.Equal(http.StatusInternalServerError, rr.Code)
}

func TestRouter_GetLoadBalancersCountByEmail(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		expectedCount  int
		expectedStatus int
		err            error
	}{
		{
			name:           "Should return the exact number of load balancers owned by an email",
			email:          "user1@test.com",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			err:            nil,
		},
		{
			name:           "Should return 0 if there's not load balanced binded with this email",
			email:          "fakeuser@test.com",
			expectedStatus: http.StatusOK,
			err:            nil,
		},
	}

	for _, test := range tests {
		c := require.New(t)

		router, err := newTestRouter(t)
		c.NoError(err)

		rr := httptest.NewRecorder()

		url := fmt.Sprintf("/v1/user/%s/load_balancer/count", test.email)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		c.NoError(err)

		router.Router.ServeHTTP(rr, req)
		c.Equal(test.expectedStatus, rr.Code)

		if test.err == nil {
			var lbCount int
			err := json.Unmarshal(rr.Body.Bytes(), &lbCount)
			c.NoError(err)
			c.Equal(test.expectedCount, lbCount)
		}
	}
}
