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
	readerMock.On("ReadGigastakeApps", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(testdata.GigastakeApps, nil)
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

/* ---------- V2 Chain Handler Method Unit Tests ---------- */

func TestRouter_getChainByID(t *testing.T) {
	tests := []struct {
		name                   string
		chainID                types.RelayChainID
		expectedStatus         int
		expectedChain          types.Chain
		expectedChainGigastake *types.GigastakeApp
	}{
		{
			name:                   "Should fetch a single chain",
			chainID:                "0001",
			expectedStatus:         http.StatusOK,
			expectedChain:          *testdata.Chains["0001"],
			expectedChainGigastake: testdata.GigastakeApps["test_gigastake_app_1"],
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v2/chain/%s", test.chainID), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var chainResp types.Chain
			err = json.Unmarshal(rr.Body.Bytes(), &chainResp)
			c.NoError(err)
			test.expectedChain.GigastakeApps = map[types.GigastakeAppID]*types.GigastakeApp{
				test.expectedChainGigastake.ID: test.expectedChainGigastake,
			}
			c.Equal(test.expectedChain, chainResp)
		})
	}
}

func TestRouter_getAllChains(t *testing.T) {
	tests := []struct {
		name           string
		chainID        types.RelayChainID
		expectedStatus int
		expectedChains map[types.RelayChainID]*types.Chain
	}{
		{
			name:           "Should fetch a single chain",
			chainID:        "0001",
			expectedStatus: http.StatusOK,
			expectedChains: testdata.Chains,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/v2/chain", nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var chainsResp []types.Chain
			err = json.Unmarshal(rr.Body.Bytes(), &chainsResp)
			c.NoError(err)
			chainsRespMap := map[types.RelayChainID]*types.Chain{}
			for i := range chainsResp {
				chainsRespMap[chainsResp[i].ID] = &chainsResp[i]
			}
			c.Equal(test.expectedChains, chainsRespMap)
		})
	}
}

func TestRouter_createNewChainAndGigastakeApps(t *testing.T) {
	tests := []struct {
		name             string
		newChainInput    types.NewChainInput
		expectedStatus   int
		expectedNewChain *types.NewChainInput
	}{
		{
			name:             "Should create a new chain and its GigastakeApps",
			newChainInput:    testdata.TestCreateNewChainInput,
			expectedStatus:   http.StatusOK,
			expectedNewChain: &testdata.TestCreateNewChainInput,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("WriteChainAndGigastakeApps", mock.Anything, mock.Anything, mock.Anything).Return(test.expectedNewChain, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			newChainInputBytes, err := json.Marshal(test.newChainInput)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, "/v2/chain", bytes.NewBuffer(newChainInputBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var chainResp *types.NewChainInput
			err = json.Unmarshal(rr.Body.Bytes(), &chainResp)
			c.NoError(err)
			c.Equal(test.expectedNewChain, chainResp)
		})
	}
}

func TestRouter_createNewGigastakeApp(t *testing.T) {
	tests := []struct {
		name                 string
		gigastakeAppInput    types.GigastakeApp
		expectedStatus       int
		expectedGigastakeApp *types.GigastakeApp // assume you have this type that matches the output structure
	}{
		{
			name:                 "Should create a new GigastakeApp",
			gigastakeAppInput:    testdata.TestCreateGigastakeApp,
			expectedStatus:       http.StatusOK,
			expectedGigastakeApp: &testdata.TestCreateGigastakeApp,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("WriteGigastakeApp", mock.Anything, mock.Anything, mock.Anything).Return(test.expectedGigastakeApp, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			gigastakeAppInputBytes, err := json.Marshal(test.gigastakeAppInput)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, "/v2/chain/gigastake", bytes.NewBuffer(gigastakeAppInputBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var gigastakeAppResp *types.GigastakeApp
			err = json.Unmarshal(rr.Body.Bytes(), &gigastakeAppResp)
			c.NoError(err)
			c.Equal(test.expectedGigastakeApp, gigastakeAppResp)
		})
	}
}

func TestRouter_updateChain(t *testing.T) {
	tests := []struct {
		name           string
		chainUpdate    types.Chain
		expectedStatus int
		expectedChain  *types.Chain
	}{
		{
			name:           "Should update a chain",
			chainUpdate:    testdata.UpdateChainOne,
			expectedStatus: http.StatusOK,
			expectedChain:  &testdata.UpdateChainOne,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("UpdateChain", mock.Anything, test.chainUpdate, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			chainUpdateBytes, err := json.Marshal(test.chainUpdate)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, "/v2/chain", bytes.NewBuffer(chainUpdateBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var chainResp *types.Chain
			err = json.Unmarshal(rr.Body.Bytes(), &chainResp)
			c.NoError(err)
			c.Equal(test.expectedChain, chainResp)
		})
	}
}

func TestRouter_activateChain(t *testing.T) {
	tests := []struct {
		name           string
		chainID        types.RelayChainID
		active         bool
		expectedStatus int
	}{
		{
			name:           "Should activate a chain",
			chainID:        "test_chain_id",
			active:         true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Should deactivate a chain",
			chainID:        "0064",
			active:         false,
			expectedStatus: http.StatusOK,
		},
	}

	c := require.New(t)
	router, err := newTestRouter(t)
	c.NoError(err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			writerMock := driver.NewMockDriver(t)
			writerMock.On("SetChainActiveStatus", mock.Anything, test.chainID, test.active, mock.Anything).Return(test.active, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			activeBytes, err := json.Marshal(test.active)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/v2/chain/%s/activate", test.chainID), bytes.NewBuffer(activeBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var activated bool
			err = json.Unmarshal(rr.Body.Bytes(), &activated)
			c.NoError(err)
			c.Equal(test.active, activated)
		})
	}
}

/* ---------- V1 Unit Tests ---------- */

func TestRouter_getUserPermissionsByUserID(t *testing.T) {
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

func TestRouter_getApplicationsByUserID(t *testing.T) {
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

func TestRouter_getApplications(t *testing.T) {
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

func TestRouter_getLoadBalancersByUserID(t *testing.T) {
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

func TestRouter_getLoadBalancers(t *testing.T) {
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

func TestRouter_getLoadBalancer(t *testing.T) {
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

func TestRouter_getBlockchains(t *testing.T) {
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

func TestRouter_getBlockchain(t *testing.T) {
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

func TestRouter_getPayPlans(t *testing.T) {
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

func TestRouter_getPayPlan(t *testing.T) {
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

func TestRouter_createLoadBalancer(t *testing.T) {
	tests := []struct {
		name              string
		loadBalancer      v1Types.LoadBalancer
		providerUserID    string
		expectedStatus    int
		expectedPortalApp types.PortalApp
	}{
		{
			name:              "Should create a new load balancer",
			loadBalancer:      testdata.LegacyLoadBalancer,
			providerUserID:    "auth0|james_holden",
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

			test.loadBalancer.UserID = test.providerUserID
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
		})
	}
}

func TestRouter_createLoadBalancerUser(t *testing.T) {
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
			writerMock.On("WriteAccountUser", mock.Anything, mock.Anything, mock.Anything).Return(types.UserID("dummy_user_id"), nil).Once()
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

func TestRouter_createPortalUser(t *testing.T) {
	tests := []struct {
		name             string
		userInput        types.CreateUser
		expectedStatus   int
		expectedResponse types.CreateUserResponse
	}{
		{
			name: "Should create a new user",
			userInput: types.CreateUser{
				Email:          "test@test.com",
				ProviderUserID: "auth0|test",
			},
			expectedStatus: http.StatusOK,
			expectedResponse: types.CreateUserResponse{
				User: types.User{
					Email:    "test@test.com",
					SignedUp: true,
					AuthProviders: map[types.AuthType]types.UserAuthProvider{
						types.AuthTypeAuth0Github: {ProviderUserID: "auth0|test", Type: types.AuthTypeAuth0Github, Provider: types.AuthProviderAuth0},
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
					types.AuthTypeAuth0Github: {ProviderUserID: "auth0|test", Type: types.AuthTypeAuth0Github, Provider: types.AuthProviderAuth0},
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

func TestRouter_updateFirstDateSurpassed(t *testing.T) {
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

func TestRouter_updateLoadBalancer(t *testing.T) {
	tests := []struct {
		name           string
		lbID           string
		updateInput    v1Types.UpdateApplication
		expectedStatus int
	}{
		{
			name: "Should update load balancer",
			lbID: "test_app_1",
			updateInput: v1Types.UpdateApplication{
				Name: "Updated Test App",
				GatewaySettings: &v1Types.UpdateGatewaySettings{
					ID:                "gateway_1",
					SecretKey:         "new_secret_key",
					SecretKeyRequired: nil,
					WhitelistOrigins:  []string{"https://example.com"},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Should update load balancer plan to Enterprise",
			lbID: "test_app_1",
			updateInput: v1Types.UpdateApplication{
				Limit: &v1Types.AppLimit{
					PayPlan: v1Types.PayPlan{
						Type: v1Types.Enterprise,
					},
					CustomLimit: 10_000,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Should update load balancer plan to PayAsYouGoV0",
			lbID: "test_app_1",
			updateInput: v1Types.UpdateApplication{
				Limit: &v1Types.AppLimit{
					PayPlan: v1Types.PayPlan{
						Type: v1Types.PayAsYouGoV0,
					},
				},
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
			writerMock.On("UpdatePortalApp", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			if test.updateInput.Limit != nil {
				writerMock.On("UpdateAccount", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
			}
			router.writer = writerMock

			rr := httptest.NewRecorder()

			updateInputBytes, err := json.Marshal(test.updateInput)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/v1/load_balancer/%s", test.lbID), bytes.NewBuffer(updateInputBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var lbResp v1Types.LoadBalancer
			err = json.Unmarshal(rr.Body.Bytes(), &lbResp)
			c.NoError(err)

			if test.updateInput.Name != "" {
				c.Equal(test.updateInput.Name, lbResp.Name)
			}
			if test.updateInput.GatewaySettings != nil && test.updateInput.GatewaySettings.SecretKey != "" {
				c.Equal(test.updateInput.GatewaySettings.SecretKey, lbResp.Applications[0].GatewaySettings.SecretKey)
			}
			if test.updateInput.GatewaySettings != nil && len(test.updateInput.GatewaySettings.WhitelistOrigins) > 0 {
				c.Equal(test.updateInput.GatewaySettings.WhitelistOrigins, lbResp.Applications[0].GatewaySettings.WhitelistOrigins)
			}
			if test.updateInput.Limit != nil && test.updateInput.Limit.PayPlan.Type != "" {
				switch test.updateInput.Limit.PayPlan.Type {
				case v1Types.Enterprise:
					c.Equal(test.updateInput.Limit.CustomLimit, lbResp.Applications[0].DailyLimit())
				case v1Types.PayAsYouGoV0:
					c.Equal(0, lbResp.Applications[0].DailyLimit())
				}
			}
		})
	}
}

func TestRouter_updateLoadBalancerUserRole(t *testing.T) {
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
		{
			name: "Should update load balancer user role back",
			lbID: "test_app_2",
			updateUser: v1Types.UpdateUserAccess{
				UserID:   "user_9",
				RoleName: v1Types.RoleMember,
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

func TestRouter_acceptLoadBalancerUser(t *testing.T) {
	tests := []struct {
		name           string
		lbID           string
		portalUserID   string
		acceptUser     v1Types.UpdateUserAccess
		expectedStatus int
	}{
		{
			name:         "Should accept load balancer user",
			lbID:         "test_app_2",
			portalUserID: "user_9",
			acceptUser: v1Types.UpdateUserAccess{
				UserID:   "auth0|tyrion_lannister",
				RoleName: v1Types.RoleMember,
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
				if user.UserID == test.portalUserID {
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

func TestRouter_deleteLoadBalancerUser(t *testing.T) {
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
			writerMock.On("RemoveAccountUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
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
			ID:     "test_protocol_app_1",
			UserID: "user_1",
			Name:   "pokt_app_123",
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_34715cae753e67c75fbb340442e7de8e",
				ApplicationPublicKey: "test_34715cae753e67c75fbb340442e7de8e",
				ApplicationSignature: "test_1dc39a2e5a84a35bf030969a0b3231f7",
				ClientPublicKey:      "test_89a3af6a587aec02cfade6f5000424c2",
				PrivateKey:           "",
				Version:              "0.0.1",
			},
			GatewaySettings: v1Types.GatewaySettings{
				SecretKey:           "test_40f482d91a5ef2300ebb4e2308c",
				SecretKeyRequired:   true,
				WhitelistOrigins:    []string{"https://test.com"},
				WhitelistUserAgents: []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"},
				WhitelistContracts: []v1Types.WhitelistContracts{
					{BlockchainID: "0001", Contracts: []string{"0x1234567890abcdef"}},
				},
				WhitelistMethods: []v1Types.WhitelistMethods{
					{BlockchainID: "0001", Methods: []string{"GET"}},
				},
				WhitelistBlockchains: []string{"0053"},
			},
			Limit: v1Types.AppLimit{
				PayPlan: v1Types.PayPlan{
					Type:  v1Types.FreetierV0,
					Limit: 250_000,
				},
			},
			NotificationSettings: v1Types.NotificationSettings{
				Quarter:       true,
				ThreeQuarters: true,
				Full:          true,
			},
			FirstDateSurpassed: testdata.MockTimestamp,
			CreatedAt:          testdata.MockTimestamp,
			UpdatedAt:          testdata.MockTimestamp,
		},
		"test_protocol_app_2": {
			ID:     "test_protocol_app_2",
			UserID: "user_3",
			Name:   "pokt_app_456",
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_8237c72345f12d1b1a8b64a1a7f66fa4",
				ApplicationPublicKey: "test_8237c72345f12d1b1a8b64a1a7f66fa4",
				ApplicationSignature: "test_f48d33b30ddaf60a1e5bb50d2ba8da5a",
				ClientPublicKey:      "test_04c71d90a92f40416b6f1d7d8af17e02",
				PrivateKey:           "",
				Version:              "0.0.1",
			},
			Limit: v1Types.AppLimit{
				PayPlan: v1Types.PayPlan{
					Type:  v1Types.PayAsYouGoV0,
					Limit: 0,
				},
			},
			GatewaySettings: v1Types.GatewaySettings{
				SecretKey:           "test_9c9e3b193cfba5348f93bb2f3e3fb794",
				WhitelistOrigins:    []string{"https://example.com"},
				WhitelistUserAgents: []string{"Mozilla/5.0 (Linux; Android 10; SM-A205U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36"},
				WhitelistContracts: []v1Types.WhitelistContracts{
					{BlockchainID: "0064", Contracts: []string{"0x0987654321abcdef"}},
				},
				WhitelistMethods: []v1Types.WhitelistMethods{
					{BlockchainID: "0064", Methods: []string{"POST"}},
				},
				WhitelistBlockchains: []string{"0021"},
			},
			NotificationSettings: v1Types.NotificationSettings{
				Half: true,
				Full: true,
			},
			FirstDateSurpassed: testdata.MockTimestamp,
			CreatedAt:          testdata.MockTimestamp,
			UpdatedAt:          testdata.MockTimestamp,
		},
		"test_protocol_app_3": {
			ID:     "test_protocol_app_3",
			UserID: "user_5",
			Name:   "pokt_app_789",
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_b5e07928fc80083c13ad0201b81bae9b",
				ApplicationPublicKey: "test_f608500e4fe3e09014fe2411b4a560b5",
				ApplicationSignature: "test_c3cd8be16ba32e24dd49fdb0247fc9b8",
				ClientPublicKey:      "test_328a9cf1b35085eeaa669aa858f6fba9",
				PrivateKey:           "",
				Version:              "0.0.1",
			},
			Limit: v1Types.AppLimit{
				PayPlan: v1Types.PayPlan{
					Type:  v1Types.Enterprise,
					Limit: 0,
				},
				CustomLimit: 4_200_000,
			},
			GatewaySettings: v1Types.GatewaySettings{
				SecretKey: "test_9f48b13e2bc5fd31ab367841f11495c1",
			},
			FirstDateSurpassed: testdata.MockTimestamp,
			CreatedAt:          testdata.MockTimestamp,
			UpdatedAt:          testdata.MockTimestamp,
		},
		"test_protocol_app_4": {
			ID:     "test_protocol_app_4",
			UserID: "user_5",
			Name:   "pokt_app_789",
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_eb2e5bcba557cfe8fa76fd7fff54f9d1",
				ApplicationPublicKey: "test_f6a5d8690ecb669865bd752b7796a920",
				ApplicationSignature: "test_cf05cf9bb26111c548e88fb6157af708",
				ClientPublicKey:      "test_6ee5ea553408f0895923fd1569dc5072",
				PrivateKey:           "",
				Version:              "0.0.1",
			},
			Limit: v1Types.AppLimit{
				PayPlan: v1Types.PayPlan{
					Type:  v1Types.Enterprise,
					Limit: 0,
				},
				CustomLimit: 4200000,
			},
			GatewaySettings: v1Types.GatewaySettings{
				SecretKey: "test_9f48b13e2bc5fd31ab367841f11495c1",
			},
			FirstDateSurpassed: testdata.MockTimestamp,
			CreatedAt:          testdata.MockTimestamp,
			UpdatedAt:          testdata.MockTimestamp,
		},
		"test_gigastake_app_1": {
			ID:   "test_gigastake_app_1",
			Name: "pokt_gigastake",
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_8d4f6a5b0c6e9f1db12c1f662e5ec8c5",
				ApplicationPublicKey: "test_37a0e8437f5149dc98a9a5b207efc2d0",
				ApplicationSignature: "test_f22651fb566346fca30b605e5f46e3ca",
				ClientPublicKey:      "test_65c29f0cc82e418b81a528a0c0682a9f",
				PrivateKey:           "",
				Version:              "0.0.1",
			},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
		"test_gigastake_app_2": {
			ID:   "test_gigastake_app_2",
			Name: "optimism_gigastake",
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_5c60d434db4e42d2b5d2ea6eeb8933c4",
				ApplicationPublicKey: "test_a7e28f8d716541a0a332a5dc6b7e4e6e",
				ApplicationSignature: "test_52e991c26da841bc882ad3a3ee9ee964",
				ClientPublicKey:      "test_ba4e53dada8f4f939048e56dc8f88f37",
				PrivateKey:           "",
				Version:              "0.0.1",
			},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
		"test_gigastake_app_3": {
			ID:   "test_gigastake_app_3",
			Name: "harmony_gigastake",
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_e570c841d5cd4f6197e0428ed7c517fd",
				ApplicationPublicKey: "test_4f805bbbf96c4a649efc3f4f95616f2e",
				ApplicationSignature: "test_01eac46efc9242a2be73879f1d09f1dc",
				ClientPublicKey:      "test_789f9d6adcc846f1a079bf68237b5f5c",
				PrivateKey:           "",
				Version:              "0.0.1",
			},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
	}

	testLoadBalancers = map[string]v1Types.LoadBalancer{
		"test_app_1": {
			ID:                "test_app_1",
			AccountID:         "account_1",
			Name:              "pokt_app_123",
			UserID:            "user_1",
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
			ID:                "test_app_2",
			AccountID:         "account_2",
			Name:              "pokt_app_456",
			UserID:            "user_3",
			GigastakeRedirect: true,
			RequestTimeout:    10_000,
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
			AccountID:         "account_3",
			Name:              "pokt_app_789",
			UserID:            "user_5",
			GigastakeRedirect: true,
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
		"legacy_lb_1": {
			ID:        "legacy_lb_1",
			Gigastake: true,
			Applications: []*v1Types.Application{
				{
					ID:   "test_gigastake_app_1",
					Name: "pokt_gigastake",
					GatewayAAT: v1Types.GatewayAAT{
						Address:              "test_8d4f6a5b0c6e9f1db12c1f662e5ec8c5",
						ApplicationPublicKey: "test_37a0e8437f5149dc98a9a5b207efc2d0",
						ApplicationSignature: "test_f22651fb566346fca30b605e5f46e3ca",
						ClientPublicKey:      "test_65c29f0cc82e418b81a528a0c0682a9f",
						PrivateKey:           "",
						Version:              "0.0.1",
					},
					CreatedAt: testdata.MockTimestamp,
					UpdatedAt: testdata.MockTimestamp,
				},
			},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
		"legacy_lb_2": {
			ID:        "legacy_lb_2",
			Gigastake: true,
			Applications: []*v1Types.Application{
				{
					ID:   "test_gigastake_app_2",
					Name: "optimism_gigastake",
					GatewayAAT: v1Types.GatewayAAT{
						Address:              "test_5c60d434db4e42d2b5d2ea6eeb8933c4",
						ApplicationPublicKey: "test_a7e28f8d716541a0a332a5dc6b7e4e6e",
						ApplicationSignature: "test_52e991c26da841bc882ad3a3ee9ee964",
						ClientPublicKey:      "test_ba4e53dada8f4f939048e56dc8f88f37",
						PrivateKey:           "",
						Version:              "0.0.1",
					},
					CreatedAt: testdata.MockTimestamp,
					UpdatedAt: testdata.MockTimestamp,
				},
			},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
		"legacy_lb_3": {
			ID:        "legacy_lb_3",
			Gigastake: true,
			Applications: []*v1Types.Application{
				{
					ID:   "test_gigastake_app_3",
					Name: "harmony_gigastake",
					GatewayAAT: v1Types.GatewayAAT{
						Address:              "test_e570c841d5cd4f6197e0428ed7c517fd",
						ApplicationPublicKey: "test_4f805bbbf96c4a649efc3f4f95616f2e",
						ApplicationSignature: "test_01eac46efc9242a2be73879f1d09f1dc",
						ClientPublicKey:      "test_789f9d6adcc846f1a079bf68237b5f5c",
						PrivateKey:           "",
						Version:              "0.0.1",
					},
					CreatedAt: testdata.MockTimestamp,
					UpdatedAt: testdata.MockTimestamp,
				},
			},
			CreatedAt: testdata.MockTimestamp,
			UpdatedAt: testdata.MockTimestamp,
		},
	}

	testBlockchains = map[string]v1Types.Blockchain{
		"0001": {
			ID:                "0001",
			Altruist:          "https://test_pocket:auth123456@altruist-0001.com:1234", // pragma: allowlist secret
			Blockchain:        "pokt-mainnet",
			Description:       "Pocket Network Mainnet",
			EnforceResult:     "JSON",
			Path:              "/v1/query/height",
			Ticker:            "POKT",
			BlockchainAliases: []string{"pokt-mainnet"},
			Active:            true,
			Redirects: []v1Types.Redirect{
				{Alias: "pokt-mainnet", Domain: "pokt-rpc.gateway.pokt.network", LoadBalancerID: "legacy_lb_1"},
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
				{Alias: "eth-mainnet", Domain: "eth-rpc.gateway.pokt.network", LoadBalancerID: ""},
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
				{Alias: "harmony-0", Domain: "hmy-rpc.gateway.pokt.network", LoadBalancerID: "legacy_lb_3"},
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
				{Alias: "optimism-mainnet", Domain: "op-rpc.gateway.pokt.network", LoadBalancerID: "legacy_lb_2"},
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
			ID:             "0064",
			Altruist:       "https://test_pocket:auth123456@altruist-0064.com:1234", // pragma: allowlist secret
			Blockchain:     "sui-testnet",
			Description:    "Sui Testnet",
			EnforceResult:  "JSON",
			Ticker:         "SUI-TESTNET",
			LogLimitBlocks: 100000,
			RequestTimeout: 60000,
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
		"ENTERPRISE":       {Type: "ENTERPRISE", Limit: 0},
		"FREETIER_V0":      {Type: "FREETIER_V0", Limit: 250_000},
		"PAY_AS_YOU_GO_V0": {Type: "PAY_AS_YOU_GO_V0", Limit: 0},
		"TEST_PLAN_10K":    {Type: "TEST_PLAN_10K", Limit: 10_000},
		"TEST_PLAN_90K":    {Type: "TEST_PLAN_90K", Limit: 90_000},
		"TEST_PLAN_V0":     {Type: "TEST_PLAN_V0", Limit: 0},
	}

	testUsers = map[types.UserID]*types.User{
		types.UserID("usertest"): {
			Email:    "test@test.com",
			SignedUp: true,
			AuthProviders: map[types.AuthType]types.UserAuthProvider{
				types.AuthTypeAuth0Github: {ProviderUserID: "auth0|test", Type: types.AuthTypeAuth0Github, Provider: types.AuthProviderAuth0},
			},
		},
	}
)
