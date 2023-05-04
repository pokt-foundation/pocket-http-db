package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	v1Types "github.com/pokt-foundation/portal-db/types"

	"github.com/pokt-foundation/pocket-http-db/cache"
	"github.com/pokt-foundation/portal-db/v2/driver"
	legacyadapters "github.com/pokt-foundation/portal-db/v2/legacy-adapters"
	"github.com/pokt-foundation/portal-db/v2/testdata"
	"github.com/pokt-foundation/portal-db/v2/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testCtx = context.Background()

func newTestRouter(t *testing.T) (*Router, error) {
	c := require.New(t)

	readerMock := cache.NewReaderMock(t)

	readerMock.On("ReadUserPermissions", testCtx).Return(testdata.UserPermissions, nil)
	readerMock.On("ReadUserIDsMap", testCtx).Return(testdata.UserIDs, nil)
	readerMock.On("ReadPlans", testCtx).Return(testdata.PayPlans, nil)
	readerMock.On("ReadAccounts", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(testdata.Accounts, nil)
	readerMock.On("ReadAccounts", mock.Anything, types.DriverOptions{IncludeDeleted: false}).Return(testdata.Accounts, nil)
	readerMock.On("ReadPortalApps", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(testdata.PortalApps, nil)
	readerMock.On("ReadChains", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(testdata.Chains, nil)
	readerMock.On("ReadBlockedContracts", testCtx).Return(testdata.GlobalBlockedContracts, nil)
	readerMock.On("ReadUserByUserID", mock.Anything, types.UserID("usertest")).Return(testUsers[types.UserID("usertest")], nil)

	testLog := logrus.New()

	cacheInstance, err := cache.NewCache(testCtx, cache.Config{
		Reader:              readerMock,
		Logger:              testLog,
		CacheRefreshMinutes: 5 * time.Minute,
	})
	c.NoError(err)

	return NewRouter(Config{
		Cache:   cacheInstance,
		Writer:  nil, // mock set in tests
		APIKeys: map[string]bool{"": true},
		Log:     testLog,
	})
}

func TestRouter_GetUserPermissionsByUserID(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		expectedPerms  *v1Types.UserPermissions
		expectedStatus int
	}{
		{
			name:   "Should fetch user permissions for a given user ID",
			userID: "auth0|ellen_ripley",
			expectedPerms: &v1Types.UserPermissions{
				UserID: "auth0|ellen_ripley",
				LoadBalancers: map[v1Types.LoadBalancerID]v1Types.LoadBalancerPermissions{
					"test_app_2": {
						RoleName:    "OWNER",
						Permissions: []v1Types.PermissionsEnum{"read:endpoint", "write:endpoint", "delete:endpoint", "transfer:endpoint"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Should return 404 status for non-existent user ID",
			userID:         "auth0|sir_robin",
			expectedPerms:  nil,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			url := fmt.Sprintf("/v1/user/%s/permission", test.userID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedPerms != nil {
				var permsResp *v1Types.UserPermissions
				err := json.Unmarshal(rr.Body.Bytes(), &permsResp)

				c.NoError(err)
				c.Equal(test.expectedPerms, permsResp)
			}
		})
	}
}

func TestRouter_GetApplicationsByUserID(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		expectedApps   map[string]*v1Types.Application
		expectedStatus int
	}{
		{
			name:   "Should fetch applications for a given user ID",
			userID: "auth0|james_holden",
			expectedApps: map[string]*v1Types.Application{
				"test_protocol_app_1": testApps["test_protocol_app_1"],
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Should return 404 status for non-existent user ID",
			userID:         "auth0|lord_farquad",
			expectedApps:   nil,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			url := fmt.Sprintf("/v1/user/%s/application", test.userID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedApps != nil {
				var appsResp []v1Types.Application
				err := json.Unmarshal(rr.Body.Bytes(), &appsResp)
				c.NoError(err)
				c.Equal(test.expectedApps, convertApplicationsToMap(appsResp))
			}
		})
	}
}

func TestRouter_GetApplications(t *testing.T) {
	tests := []struct {
		name           string
		expectedApps   map[string]*v1Types.Application
		expectedStatus int
	}{
		{
			name:           "Should fetch all applications excluding deleted ones",
			expectedApps:   testApps,
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/v1/application", nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedApps != nil {
				var appsResp []v1Types.Application
				err := json.Unmarshal(rr.Body.Bytes(), &appsResp)

				c.NoError(err)
				c.Equal(test.expectedApps, convertApplicationsToMap(appsResp))
			}
		})
	}
}

func TestRouter_GetApplication(t *testing.T) {
	tests := []struct {
		name           string
		appID          string
		expectedApp    v1Types.Application
		expectedStatus int
	}{
		{
			name:           "Should fetch the application with the provided ID",
			appID:          "test_protocol_app_1",
			expectedApp:    *testApps["test_protocol_app_1"],
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Should return 404 for non-existent application",
			appID:          "test_not_real_app_112233",
			expectedApp:    v1Types.Application{},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/application/%s", test.appID), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedApp.ID != "" {
				var appResp v1Types.Application
				err := json.Unmarshal(rr.Body.Bytes(), &appResp)
				c.NoError(err)
				c.Equal(test.expectedApp, appResp)
			}
		})
	}
}

func TestRouter_GetLoadBalancersByUserID(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		roleNameFilter types.RoleName
		expectedLBs    map[string]v1Types.LoadBalancer
		expectedStatus int
	}{
		{
			name:   "Should fetch load balancers for the provided user ID with no role name filter",
			userID: "github|james_holden",
			expectedLBs: map[string]v1Types.LoadBalancer{
				"test_app_1": testLoadBalancers["test_app_1"],
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Should fetch load balancers for the provided user ID and role name filter",
			userID:         "github|james_holden",
			roleNameFilter: types.RoleOwner,
			expectedLBs: map[string]v1Types.LoadBalancer{
				"test_app_1": testLoadBalancers["test_app_1"],
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Should return 400 for invalid role name filter",
			userID:         "1",
			roleNameFilter: types.RoleName("EMPEROR"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Should return 404 for non-existent user",
			userID:         "auth0|rickon_stark",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/user/%s/load_balancer", test.userID), nil)
			if test.roleNameFilter != "" {
				q := req.URL.Query()
				q.Add("filter", string(test.roleNameFilter))
				req.URL.RawQuery = q.Encode()
			}
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedLBs != nil {
				var lbsResp []v1Types.LoadBalancer
				err := json.Unmarshal(rr.Body.Bytes(), &lbsResp)
				c.NoError(err)
				c.Equal(test.expectedLBs, convertLoadBalancersToMap(lbsResp))
			}
		})
	}
}

func TestRouter_GetLoadBalancers(t *testing.T) {
	tests := []struct {
		name           string
		expectedLBs    map[string]v1Types.LoadBalancer
		expectedStatus int
	}{
		{
			name:           "Should fetch all load balancers",
			expectedLBs:    testLoadBalancers,
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/v1/load_balancer", nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedLBs != nil {
				var lbsResp []v1Types.LoadBalancer
				err := json.Unmarshal(rr.Body.Bytes(), &lbsResp)
				c.NoError(err)
				c.Equal(test.expectedLBs, convertLoadBalancersToMap(lbsResp))
			}
		})
	}
}

func TestRouter_GetLoadBalancer(t *testing.T) {
	tests := []struct {
		name           string
		loadBalancerID string
		expectedLB     v1Types.LoadBalancer
		expectedStatus int
	}{
		{
			name:           "Should fetch the load balancer with the provided ID",
			loadBalancerID: "test_app_1",
			expectedLB:     testLoadBalancers["test_app_1"],
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Should return 404 for non-existent load balancer ID",
			loadBalancerID: "test_lb_who_am_i_777fhwn",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/load_balancer/%s", test.loadBalancerID), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedLB.ID != "" {
				var lbResp v1Types.LoadBalancer
				err := json.Unmarshal(rr.Body.Bytes(), &lbResp)
				c.NoError(err)
				c.Equal(test.expectedLB, lbResp)
			}
		})
	}
}

func TestRouter_GetBlockchains(t *testing.T) {
	tests := []struct {
		name           string
		expectedChains map[string]v1Types.Blockchain
		expectedStatus int
	}{
		{
			name:           "Should fetch all blockchains",
			expectedChains: testBlockchains,
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/v1/blockchain", nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedChains != nil {
				var chainsResp []v1Types.Blockchain
				err := json.Unmarshal(rr.Body.Bytes(), &chainsResp)
				c.NoError(err)
				c.Equal(test.expectedChains, convertBlockchainsToMap(chainsResp))
			}
		})
	}
}

func TestRouter_GetBlockchain(t *testing.T) {
	tests := []struct {
		name           string
		blockchainID   string
		expectedChain  v1Types.Blockchain
		expectedStatus int
	}{
		{
			name:           "Should fetch a blockchain by its ID",
			blockchainID:   "0001",
			expectedChain:  testBlockchains["0001"],
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Should return 404 for a non-existent blockchain ID",
			blockchainID:   "9000",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/blockchain/%s", test.blockchainID), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedChain.ID != "" {
				var chainResp v1Types.Blockchain
				err := json.Unmarshal(rr.Body.Bytes(), &chainResp)
				c.NoError(err)
				c.Equal(test.expectedChain, chainResp)
			}
		})
	}
}

func TestRouter_GetPayPlans(t *testing.T) {
	tests := []struct {
		name             string
		expectedPayPlans map[string]v1Types.PayPlan
		expectedStatus   int
	}{
		{
			name:             "Should fetch all pay plans",
			expectedPayPlans: testPayPlans,
			expectedStatus:   http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/v1/pay_plan", nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedPayPlans != nil {
				var payPlansResp []v1Types.PayPlan
				err := json.Unmarshal(rr.Body.Bytes(), &payPlansResp)
				c.NoError(err)
				c.Equal(test.expectedPayPlans, convertPayPlansToMap(payPlansResp))
			}
		})
	}
}

func TestRouter_GetPayPlan(t *testing.T) {
	tests := []struct {
		name            string
		payPlanType     string
		expectedPayPlan v1Types.PayPlan
		expectedStatus  int
	}{
		{
			name:            "Should fetch the basic_plan pay plan",
			payPlanType:     "basic_plan",
			expectedPayPlan: testPayPlans["basic_plan"],
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "Should fetch the developer_plan pay plan",
			payPlanType:     "developer_plan",
			expectedPayPlan: testPayPlans["developer_plan"],
			expectedStatus:  http.StatusOK,
		},
		{
			name:           "Should return 404 for non-existent pay plan",
			payPlanType:    "non_existent_plan",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/pay_plan/%s", test.payPlanType), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedPayPlan.Type != "" {
				var ppResp v1Types.PayPlan
				err := json.Unmarshal(rr.Body.Bytes(), &ppResp)
				c.NoError(err)
				c.Equal(test.expectedPayPlan, ppResp)
			}
		})
	}
}

func TestRouter_CreateLoadBalancer(t *testing.T) {
	tests := []struct {
		name              string
		loadBalancer      v1Types.LoadBalancer
		expectedStatus    int
		expectedPortalApp types.PortalApp
	}{
		{
			name:              "Should create a new load balancer",
			loadBalancer:      testdata.LegacyLoadBalancer,
			expectedStatus:    http.StatusOK,
			expectedPortalApp: testdata.V2CreatePortalApp,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			testdata.V2CreateAccount.ID = ""
			writerMock.On("WritePortalApp", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
				&testdata.V2CreatePortalApp, nil,
			).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			lbBytes, err := json.Marshal(test.loadBalancer)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, "/v1/load_balancer", bytes.NewBuffer(lbBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedPortalApp.ID != "" {
				var appResp types.PortalApp
				err := json.Unmarshal(rr.Body.Bytes(), &appResp)
				c.NoError(err)
				c.Equal(test.expectedPortalApp, appResp)
			}

			// TODO - update this if/when listener channel to router methods added
			// if test.expectAccount {
			// 	var lbAccount types.Account
			// 	for _, account := range router.Cache.GetAccounts(cache.Options{}) {
			// 		if account.LegacyLoadBalancerID == testdata.LegacyLoadBalancer.ID {
			// 			lbAccount = account
			// 			break
			// 		}
			// 	}
			// 	c.True(lbAccount.ID != 0)

			// 	account := router.Cache.GetAccount(lbAccount.ID)
			// 	c.NotEmpty(account.ID)
			// 	c.Equal(test.loadBalancer.ID, string(account.ID))
			// }

			// if test.expectPortalApp {
			// 	portalApp := router.Cache.GetPortalApp(types.PortalAppID(test.loadBalancer.Applications[0].ID))
			// 	c.NotEmpty(portalApp.ID)
			// 	c.Equal(test.loadBalancer.Applications[0].ID, string(portalApp.ID))
			// }
		})
	}
}

func TestRouter_CreateLoadBalancerUser(t *testing.T) {
	tests := []struct {
		name                      string
		loadBalancerID            string
		userAccess                v1Types.UserAccess
		expectedStatus            int
		expectedLoadBalancerUsers []v1Types.UserAccess
	}{
		{
			name:                      "Should create a new load balancer user",
			loadBalancerID:            "test_app_3",
			userAccess:                testdata.LegacyUserAccess,
			expectedStatus:            http.StatusOK,
			expectedLoadBalancerUsers: []v1Types.UserAccess{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			v2AccountUserAccess := legacyadapters.ConvertToV2AccountUserAccess(test.userAccess)
			writerMock.On("WriteAccountUser", mock.Anything, mock.Anything, mock.Anything).Return(&v2AccountUserAccess, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			userAccessBytes, err := json.Marshal(test.userAccess)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/v1/load_balancer/%s/user", test.loadBalancerID), bytes.NewBuffer(userAccessBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if len(test.expectedLoadBalancerUsers) > 0 {
				var lbResp v1Types.LoadBalancer
				err := json.Unmarshal(rr.Body.Bytes(), &lbResp)
				c.NoError(err)
				c.Equal(test.expectedLoadBalancerUsers, lbResp.Users)
			}
		})
	}
}

func TestRouter_CreateBlockchain(t *testing.T) {
	tests := []struct {
		name             string
		blockchain       v1Types.Blockchain
		expectedStatus   int
		expectedResponse v1Types.Blockchain
	}{
		{
			name:             "Should create a new blockchain",
			blockchain:       testdata.LegacyBlockchain,
			expectedStatus:   http.StatusOK,
			expectedResponse: testdata.LegacyBlockchain,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			v2Chain := legacyadapters.ConvertToV2Chain(test.blockchain)
			writerMock.On("WriteChain", mock.Anything, mock.Anything, mock.Anything).Return(&v2Chain, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			blockchainBytes, err := json.Marshal(test.blockchain)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, "/v1/blockchain", bytes.NewBuffer(blockchainBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var blockchainResp v1Types.Blockchain
			err = json.Unmarshal(rr.Body.Bytes(), &blockchainResp)
			c.NoError(err)
			c.Equal(test.expectedResponse, blockchainResp)
		})
	}
}

func TestRouter_CreateRedirect(t *testing.T) {
	tests := []struct {
		name             string
		redirect         v1Types.Redirect
		expectedStatus   int
		expectedResponse v1Types.Redirect
	}{
		{
			name:             "Should create a new redirect",
			redirect:         testdata.LegacyRedirect,
			expectedStatus:   http.StatusOK,
			expectedResponse: testdata.LegacyRedirect,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("UpdateChain", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			redirectBytes, err := json.Marshal(test.redirect)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, "/v1/blockchain/redirect", bytes.NewBuffer(redirectBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var redirectResp v1Types.Redirect
			err = json.Unmarshal(rr.Body.Bytes(), &redirectResp)
			c.NoError(err)
			c.Equal(test.expectedResponse, redirectResp)
		})
	}
}

func TestRouter_CreatePortalUser(t *testing.T) {
	tests := []struct {
		name             string
		userInput        types.CreateUser
		expectedStatus   int
		expectedResponse types.CreateUserResponse
	}{
		{
			name: "Should create a new user",
			userInput: types.CreateUser{
				Email:            "test@test.com",
				AuthProviderType: types.AuthTypeAuth0Github,
				ProviderUserID:   "auth0_username|test",
			},
			expectedStatus: http.StatusOK,
			expectedResponse: types.CreateUserResponse{
				User: types.User{
					Email:    "test@test.com",
					SignedUp: true,
					AuthProviders: map[types.AuthType]types.UserAuthProvider{
						types.AuthTypeAuth0Github: {ProviderUserID: "auth0_username|test", Type: types.AuthTypeAuth0Github, Provider: types.AuthProviderAuth0},
					},
				},
				AccountID: types.AccountID("accountTestID"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("WriteUserNewSignUp", mock.Anything, mock.Anything, mock.Anything).Return(&types.User{
				Email:    "test@test.com",
				SignedUp: true,
				AuthProviders: map[types.AuthType]types.UserAuthProvider{
					types.AuthTypeAuth0Github: {ProviderUserID: "auth0_username|test", Type: types.AuthTypeAuth0Github, Provider: types.AuthProviderAuth0},
				},
			}, types.AccountID("accountTestID"), nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			userInputBytes, err := json.Marshal(test.userInput)
			c.NoError(err)

			req, err := http.NewRequestWithContext(testCtx, http.MethodPost, "/v1/user", bytes.NewBuffer(userInputBytes))
			c.NoError(err)

			fmt.Println(req.Context())

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var resp types.CreateUserResponse
			err = json.Unmarshal(rr.Body.Bytes(), &resp)
			c.NoError(err)
			c.Equal(test.expectedResponse, resp)
		})
	}
}

func TestRouter_UpdateFirstDateSurpassed(t *testing.T) {
	tests := []struct {
		name           string
		updateInput    types.UpdateFirstDateSurpassed
		expectedStatus int
	}{
		{
			name: "Should update first date surpassed",
			updateInput: types.UpdateFirstDateSurpassed{
				PortalAppIDs:       []string{"test_app_1", "test_app_2", "test_app_3"},
				FirstDateSurpassed: testdata.MockTimestamp,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("UpdatePortalAppsFirstDateSurpassed", mock.Anything, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			updateInputBytes, err := json.Marshal(test.updateInput)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, "/v1/application/first_date_surpassed", bytes.NewBuffer(updateInputBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var appsResp []v1Types.Application
			err = json.Unmarshal(rr.Body.Bytes(), &appsResp)
			c.NoError(err)
			for _, app := range appsResp {
				c.Equal(test.updateInput.FirstDateSurpassed, app.FirstDateSurpassed)
			}
		})
	}
}

func TestRouter_UpdateLoadBalancerUserRole(t *testing.T) {
	tests := []struct {
		name           string
		lbID           string
		userID         string
		updateUser     v1Types.UpdateUserAccess
		expectedStatus int
	}{
		{
			name: "Should update load balancer user role",
			lbID: "test_app_2",
			updateUser: v1Types.UpdateUserAccess{
				UserID:   "user_9",
				RoleName: v1Types.RoleAdmin,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("SetAccountUserRole", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			updateUserBytes, err := json.Marshal(test.updateUser)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/v1/load_balancer/%s/user", test.lbID), bytes.NewBuffer(updateUserBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var lbResp v1Types.LoadBalancer
			err = json.Unmarshal(rr.Body.Bytes(), &lbResp)
			c.NoError(err)

			updated := false
			for _, user := range lbResp.Users {
				if user.UserID == test.updateUser.UserID {
					c.Equal(test.updateUser.RoleName, user.RoleName)
					updated = true
					break
				}
			}
			c.True(updated, "User role should be updated")
		})
	}
}

func TestRouter_AcceptLoadBalancerUser(t *testing.T) {
	tests := []struct {
		name           string
		lbID           string
		acceptUser     v1Types.UpdateUserAccess
		expectedStatus int
	}{
		{
			name: "Should accept load balancer user",
			lbID: "test_app_2",
			acceptUser: v1Types.UpdateUserAccess{
				UserID:   "user_9",
				RoleName: v1Types.RoleAdmin,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("UpdateAcceptAccountUser", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			acceptUserBytes, err := json.Marshal(test.acceptUser)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/v1/load_balancer/%s/user/accept", test.lbID), bytes.NewBuffer(acceptUserBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var lbResp v1Types.LoadBalancer
			err = json.Unmarshal(rr.Body.Bytes(), &lbResp)
			c.NoError(err)
			accepted := false
			for _, user := range lbResp.Users {
				if user.UserID == test.acceptUser.UserID {
					c.Equal(test.acceptUser.RoleName, user.RoleName)
					c.True(user.Accepted, "User should be accepted")
					accepted = true
					break
				}
			}
			c.True(accepted, "User should be accepted")
		})
	}
}

func TestRouter_DeleteLoadBalancerUser(t *testing.T) {
	tests := []struct {
		name           string
		lbID           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "Should delete load balancer user",
			lbID:           "test_app_3",
			userID:         "user_7",
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("RemoveAccountUser", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/load_balancer/%s/user/%s", test.lbID, test.userID), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var lbResp v1Types.LoadBalancer
			err = json.Unmarshal(rr.Body.Bytes(), &lbResp)
			c.NoError(err)

			deleted := true
			for _, user := range lbResp.Users {
				if user.UserID == test.userID {
					deleted = false
					break
				}
			}
			c.True(deleted, "User should be deleted")
		})
	}
}

func TestRouter_UpdateBlockchain(t *testing.T) {
	tests := []struct {
		name             string
		blockchainID     string
		updateBlockchain v1Types.UpdateBlockchain
		expectedStatus   int
	}{
		{
			name:             "Should update blockchain",
			blockchainID:     "0001",
			updateBlockchain: testdata.LegacyUpdateBlockchain,
			expectedStatus:   http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("UpdateChain", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			updateBlockchainBytes, err := json.Marshal(test.updateBlockchain)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/v1/blockchain/%s", test.blockchainID), bytes.NewBuffer(updateBlockchainBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var updatedChainResp v1Types.Blockchain
			err = json.Unmarshal(rr.Body.Bytes(), &updatedChainResp)
			c.NoError(err)

			c.Equal(test.updateBlockchain.Blockchain, updatedChainResp.Blockchain)
			c.Equal(test.updateBlockchain.Description, updatedChainResp.Description)
			c.Equal(test.updateBlockchain.EnforceResult, updatedChainResp.EnforceResult)
			c.Equal(test.updateBlockchain.Path, updatedChainResp.Path)
			c.Equal(test.updateBlockchain.Ticker, updatedChainResp.Ticker)
			c.ElementsMatch(test.updateBlockchain.BlockchainAliases, updatedChainResp.BlockchainAliases)
			c.Equal(test.updateBlockchain.Body, updatedChainResp.SyncCheckOptions.Body)
			c.Equal(test.updateBlockchain.ResultKey, updatedChainResp.SyncCheckOptions.ResultKey)
			c.Equal(*test.updateBlockchain.Allowance, updatedChainResp.SyncCheckOptions.Allowance)
			c.Equal(test.updateBlockchain.Altruist, updatedChainResp.Altruist)
		})
	}
}

func TestRouter_ActivateBlockchain(t *testing.T) {
	tests := []struct {
		name           string
		blockchainID   string
		active         bool
		expectedStatus int
	}{
		{
			name:           "Should deactivate blockchain",
			blockchainID:   "0001",
			active:         false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Should activate blockchain",
			blockchainID:   "0001",
			active:         true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("SetChainActiveStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.active, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			activeBytes, err := json.Marshal(test.active)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/v1/blockchain/%s/activate", test.blockchainID), bytes.NewBuffer(activeBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var activatedResp bool
			err = json.Unmarshal(rr.Body.Bytes(), &activatedResp)
			c.NoError(err)

			c.Equal(test.active, activatedResp)
		})
	}
}

/* ---------- Test Utils ---------- */

func convertApplicationsToMap(apps []v1Types.Application) map[string]*v1Types.Application {
	appMap := make(map[string]*v1Types.Application)
	for i := range apps {
		appMap[apps[i].ID] = &apps[i]
	}
	return appMap
}

func convertLoadBalancersToMap(lbs []v1Types.LoadBalancer) map[string]v1Types.LoadBalancer {
	lbMap := make(map[string]v1Types.LoadBalancer)
	for i := range lbs {
		if len(lbs[i].Applications) != 0 {
			sort.Slice(lbs[i].Applications, func(j, k int) bool {
				return lbs[i].Applications[j].ID < lbs[i].Applications[k].ID
			})
		}
		lbMap[lbs[i].ID] = lbs[i]
	}
	return lbMap
}

func convertBlockchainsToMap(blockchains []v1Types.Blockchain) map[string]v1Types.Blockchain {
	result := make(map[string]v1Types.Blockchain)
	for _, blockchain := range blockchains {
		result[blockchain.ID] = blockchain
	}
	return result
}

func convertPayPlansToMap(payPlans []v1Types.PayPlan) map[string]v1Types.PayPlan {
	payPlansMap := make(map[string]v1Types.PayPlan)
	for _, pp := range payPlans {
		payPlansMap[string(pp.Type)] = pp
	}
	return payPlansMap
}

var (
	testApps = map[string]*v1Types.Application{
		"test_protocol_app_1": {
			ID:                 "test_protocol_app_1",
			Name:               "pokt_app_123",
			UserID:             "auth0|james_holden",
			FirstDateSurpassed: testdata.MockTimestamp,
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_34715cae753e67c75fbb340442e7de8e",
				ApplicationPublicKey: "test_34715cae753e67c75fbb340442e7de8e",
				ApplicationSignature: "test_1dc39a2e5a84a35bf030969a0b3231f7",
				ClientPublicKey:      "test_89a3af6a587aec02cfade6f5000424c2",
				PrivateKey:           "test_11b8d394ca331d7c7a71ca1896d630f6",
				Version:              "0.0.1",
			},
			GatewaySettings: v1Types.GatewaySettings{
				SecretKey:            "test_40f482d91a5ef2300ebb4e2308c",
				SecretKeyRequired:    true,
				WhitelistOrigins:     []string{"https://test.com"},
				WhitelistUserAgents:  []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"},
				WhitelistContracts:   []v1Types.WhitelistContracts{{BlockchainID: "0001", Contracts: []string{"0x1234567890abcdef"}}},
				WhitelistMethods:     []v1Types.WhitelistMethods{{BlockchainID: "0001", Methods: []string{"GET"}}},
				WhitelistBlockchains: []string{"0053"},
			},
			Limit:                v1Types.AppLimit{PayPlan: v1Types.PayPlan{Type: "basic_plan", Limit: 1000}},
			NotificationSettings: v1Types.NotificationSettings{SignedUp: false, Quarter: true, Half: false, ThreeQuarters: true, Full: true},
			CreatedAt:            testdata.MockTimestamp,
			UpdatedAt:            testdata.MockTimestamp,
		},
		"test_protocol_app_2": {
			ID:                 "test_protocol_app_2",
			Name:               "pokt_app_456",
			UserID:             "auth0|ellen_ripley",
			FirstDateSurpassed: testdata.MockTimestamp,
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_8237c72345f12d1b1a8b64a1a7f66fa4",
				ApplicationPublicKey: "test_8237c72345f12d1b1a8b64a1a7f66fa4",
				ApplicationSignature: "test_f48d33b30ddaf60a1e5bb50d2ba8da5a",
				ClientPublicKey:      "test_04c71d90a92f40416b6f1d7d8af17e02",
				PrivateKey:           "test_2e83c836a29b423a47d8e18c779fd422",
				Version:              "0.0.1",
			},
			GatewaySettings: v1Types.GatewaySettings{
				SecretKey:            "test_9c9e3b193cfba5348f93bb2f3e3fb794",
				WhitelistOrigins:     []string{"https://example.com"},
				WhitelistUserAgents:  []string{"Mozilla/5.0 (Linux; Android 10; SM-A205U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36"},
				WhitelistContracts:   []v1Types.WhitelistContracts{{BlockchainID: "0064", Contracts: []string{"0x0987654321abcdef"}}},
				WhitelistMethods:     []v1Types.WhitelistMethods{{BlockchainID: "0064", Methods: []string{"POST"}}},
				WhitelistBlockchains: []string{"0021"},
			},
			Limit:                v1Types.AppLimit{PayPlan: v1Types.PayPlan{Type: "pro_plan", Limit: 5_000}},
			NotificationSettings: v1Types.NotificationSettings{Half: true, Full: true},
			CreatedAt:            testdata.MockTimestamp,
			UpdatedAt:            testdata.MockTimestamp,
		},
		"test_protocol_app_3": {
			ID:                 "test_protocol_app_3",
			Name:               "pokt_app_789",
			UserID:             "auth0|chrisjen_avasarala",
			FirstDateSurpassed: testdata.MockTimestamp,
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_b5e07928fc80083c13ad0201b81bae9b",
				ApplicationPublicKey: "test_f608500e4fe3e09014fe2411b4a560b5",
				ApplicationSignature: "test_c3cd8be16ba32e24dd49fdb0247fc9b8",
				ClientPublicKey:      "test_328a9cf1b35085eeaa669aa858f6fba9",
				PrivateKey:           "test_8663e187c19f3c6e27317eab4ed6d7d5",
				Version:              "0.0.1",
			},
			GatewaySettings: v1Types.GatewaySettings{
				SecretKey: "test_9f48b13e2bc5fd31ab367841f11495c1",
			},
			Limit:     v1Types.AppLimit{PayPlan: v1Types.PayPlan{Type: "startup_plan", Limit: 500}},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
		"test_protocol_app_4": {
			ID:                 "test_protocol_app_4",
			Name:               "pokt_app_789",
			UserID:             "auth0|chrisjen_avasarala",
			FirstDateSurpassed: testdata.MockTimestamp,
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_eb2e5bcba557cfe8fa76fd7fff54f9d1",
				ApplicationPublicKey: "test_f6a5d8690ecb669865bd752b7796a920",
				ApplicationSignature: "test_cf05cf9bb26111c548e88fb6157af708",
				ClientPublicKey:      "test_6ee5ea553408f0895923fd1569dc5072",
				PrivateKey:           "test_838d29d61a65401f7d56d084cb6e4783",
				Version:              "0.0.1",
			},
			GatewaySettings: v1Types.GatewaySettings{
				SecretKey: "test_9f48b13e2bc5fd31ab367841f11495c1",
			},
			Limit:     v1Types.AppLimit{PayPlan: v1Types.PayPlan{Type: "startup_plan", Limit: 500}},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
	}

	testLoadBalancers = map[string]v1Types.LoadBalancer{
		"test_app_1": {
			ID:                "test_app_1",
			Name:              "pokt_app_123",
			UserID:            "auth0|james_holden",
			Gigastake:         true,
			GigastakeRedirect: true,
			RequestTimeout:    5_000,
			Applications: []*v1Types.Application{
				testApps["test_protocol_app_1"],
			},
			Users: []v1Types.UserAccess{
				{
					UserID:   "user_1",
					RoleName: "OWNER",
					Email:    "james.holden123@test.com",
					Accepted: true,
				},
				{
					UserID:   "user_2",
					RoleName: "ADMIN",
					Email:    "paul.atreides456@test.com",
					Accepted: true,
				},
				{
					UserID:   "user_8",
					RoleName: "ADMIN",
					Email:    "rick.deckard456@test.com",
					Accepted: false,
				},
			},
			Integrations: v1Types.AccountIntegrations{
				CovalentAPIKeyFree: "covalent_api_key_1",
			},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
		"test_app_2": {
			ID:             "test_app_2",
			Name:           "pokt_app_456",
			UserID:         "auth0|ellen_ripley",
			RequestTimeout: 10_000,
			Applications: []*v1Types.Application{
				testApps["test_protocol_app_2"],
			},
			Users: []v1Types.UserAccess{
				{
					UserID:   "user_3",
					RoleName: "OWNER",
					Email:    "ellen.ripley789@test.com",
					Accepted: true,
				},
				{
					UserID:   "user_2",
					RoleName: "MEMBER",
					Email:    "paul.atreides456@test.com",
					Accepted: true,
				},
				{
					UserID:   "user_4",
					RoleName: "MEMBER",
					Email:    "ulfric.stormcloak123@test.com",
					Accepted: true,
				},
				{
					UserID:   "user_9",
					RoleName: "MEMBER",
					Email:    "tyrion.lannister789@test.com",
					Accepted: false,
				},
			},
			Integrations: v1Types.AccountIntegrations{
				CovalentAPIKeyFree: "covalent_api_key_2",
			},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
		"test_app_3": {
			ID:                "test_app_3",
			Name:              "pokt_app_789",
			UserID:            "auth0|chrisjen_avasarala",
			GigastakeRedirect: false,
			RequestTimeout:    10_000,
			Applications: []*v1Types.Application{
				testApps["test_protocol_app_3"],
				testApps["test_protocol_app_4"],
			},
			Users: []v1Types.UserAccess{
				{
					UserID:   "user_5",
					RoleName: "OWNER",
					Email:    "chrisjen.avasarala1@test.com",
					Accepted: true,
				},
				{
					UserID:   "user_6",
					RoleName: "ADMIN",
					Email:    "amos.burton789@test.com",
					Accepted: true,
				},
				{
					UserID:   "user_10",
					RoleName: "MEMBER",
					Email:    "daenerys.targaryen123@test.com",
					Accepted: false,
				},
				{
					UserID:   "user_7",
					RoleName: "MEMBER",
					Email:    "frodo.baggins123@test.com",
					Accepted: true,
				},
			},
			Integrations: v1Types.AccountIntegrations{
				CovalentAPIKeyFree: "covalent_api_key_3",
			},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
		// "test_app_4": {
		// 	ID:     "test_app_4",
		// 	Name:   "test_account_4",
		// 	UserID: "auth0|ulfric_stormcloak",
		// 	Users: []types.UserAccess{
		// 		{
		// 			UserID:   "user_4",
		// 			RoleName: "OWNER",
		// 			Email:    "ulfric.stormcloak123@test.com",
		// 			Accepted: true,
		// 		},
		// 	},
		// 	CreatedAt: testdata.MockTimestamp,
		// 	UpdatedAt: testdata.MockTimestamp,
		// },
		// "test_app_5": {
		// 	ID:     "test_app_5",
		// 	Name:   "test_account_5",
		// 	UserID: "auth0|ulfric_stormcloak",
		// 	Users: []types.UserAccess{
		// 		{
		// 			UserID:   "user_4",
		// 			RoleName: "OWNER",
		// 			Email:    "ulfric.stormcloak123@test.com",
		// 			Accepted: true,
		// 		},
		// 	},
		// 	CreatedAt: testdata.MockTimestamp,
		// 	UpdatedAt: testdata.MockTimestamp,
		// },
	}

	testBlockchains = map[string]v1Types.Blockchain{
		"0001": {
			ID:                "0001",
			Altruist:          "https://test_pocket:auth123456@altruist-0001.com:1234", // pragma: allowlist secret
			Blockchain:        "mainnet",
			Description:       "Pocket Network Mainnet",
			EnforceResult:     "JSON",
			Path:              "/v1/query/height",
			Ticker:            "POKT",
			BlockchainAliases: []string{"mainnet"},
			Active:            true,
			Redirects: []v1Types.Redirect{
				{Alias: "altruist-0001", Domain: "pokt-rpc.gateway.pokt.network", LoadBalancerID: "test_app_1"},
			},
			SyncCheckOptions: v1Types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"query\"}",
				ResultKey: "result.sync_info", Allowance: 1,
			},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
		"0021": {
			ID:                "0021",
			Altruist:          "https://test_pocket:auth123456@altruist-0021.com:1234", // pragma: allowlist secret
			Blockchain:        "eth-mainnet",
			ChainID:           "1",
			ChainIDCheck:      "{\"method\":\"eth_chainId\",\"id\":1,\"jsonrpc\":\"2.0\"}",
			Description:       "Ethereum Mainnet",
			EnforceResult:     "JSON",
			Ticker:            "ETH",
			BlockchainAliases: []string{"eth-mainnet"},
			LogLimitBlocks:    100000,
			Active:            true,
			Redirects: []v1Types.Redirect{
				{Alias: "altruist-0021", Domain: "eth-rpc.gateway.pokt.network", LoadBalancerID: "test_app_3"},
			},
			SyncCheckOptions: v1Types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[]}",
				ResultKey: "result", Allowance: 5,
			},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
		"0040": {
			ID:                "0040",
			Altruist:          "https://test_pocket:auth123456@altruist-0040.com:1234", // pragma: allowlist secret
			Blockchain:        "harmony-0",
			Description:       "Harmony Shard 0",
			EnforceResult:     "JSON",
			Ticker:            "HMY",
			BlockchainAliases: []string{"harmony-0"},
			Active:            true,
			Redirects: []v1Types.Redirect{
				{Alias: "altruist-0040", Domain: "hmy-rpc.gateway.pokt.network", LoadBalancerID: "test_app_3"},
			},
			SyncCheckOptions: v1Types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"hmy_blockNumber\",\"params\":[]}",
				ResultKey: "result",
				Allowance: 8,
			},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
		"0053": {
			ID:                "0053",
			Altruist:          "https://test_pocket:auth123456@altruist-0053.com:1234", // pragma: allowlist secret
			Blockchain:        "optimism-mainnet",
			Description:       "Optimism Mainnet",
			EnforceResult:     "JSON",
			Ticker:            "OP",
			BlockchainAliases: []string{"optimism-mainnet"},
			LogLimitBlocks:    100000,
			Active:            true,
			Redirects: []v1Types.Redirect{
				{Alias: "altruist-0053", Domain: "op-rpc.gateway.pokt.network", LoadBalancerID: "test_app_2"},
			},
			SyncCheckOptions: v1Types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[]}",
				ResultKey: "result",
				Allowance: 2,
			},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
		"0064": {
			ID:                "0064",
			Altruist:          "https://test_pocket:auth123456@altruist-0064.com:1234", // pragma: allowlist secret
			Blockchain:        "sui-testnet",
			Description:       "Sui Testnet",
			EnforceResult:     "JSON",
			Ticker:            "SUI-TESTNET",
			BlockchainAliases: []string{"sui-testnet"},
			LogLimitBlocks:    100000,
			RequestTimeout:    60000,
			SyncCheckOptions: v1Types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"sui_blockNumber\",\"params\":[]}",
				ResultKey: "result",
				Allowance: 7,
			},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
	}

	testPayPlans = map[string]v1Types.PayPlan{
		"basic_plan":       {Type: "basic_plan", Limit: 1_000},
		"developer_plan":   {Type: "developer_plan", Limit: 100},
		"enterprise_plan":  {Type: "enterprise_plan", Limit: 10_000},
		"pro_plan":         {Type: "pro_plan", Limit: 5_000},
		"startup_plan":     {Type: "startup_plan", Limit: 500},
		"ENTERPRISE":       {Type: "ENTERPRISE", Limit: 5_000},
		"FREETIER_V0":      {Type: "FREETIER_V0", Limit: 5_000},
		"PAY_AS_YOU_GO_V0": {Type: "PAY_AS_YOU_GO_V0", Limit: 5_000},
		"TEST_PLAN_10K":    {Type: "TEST_PLAN_10K", Limit: 5_000},
		"TEST_PLAN_90K":    {Type: "TEST_PLAN_90K", Limit: 5_000},
		"TEST_PLAN_V0":     {Type: "TEST_PLAN_V0", Limit: 5_000},
	}

	testUsers = map[types.UserID]*types.User{
		types.UserID("usertest"): {
			Email:    "test@test.com",
			SignedUp: true,
			AuthProviders: map[types.AuthType]types.UserAuthProvider{
				types.AuthTypeAuth0Github: {ProviderUserID: "auth0_username|test", Type: types.AuthTypeAuth0Github, Provider: types.AuthProviderAuth0},
			},
		},
	}
)
