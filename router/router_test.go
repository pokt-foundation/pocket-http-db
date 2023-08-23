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

	"github.com/pokt-foundation/utils-go/logger"

	"github.com/pokt-foundation/pocket-http-db/cache"
	phdtestdata "github.com/pokt-foundation/pocket-http-db/testdata"
	"github.com/pokt-foundation/portal-db/v2/driver"
	"github.com/pokt-foundation/portal-db/v2/testdata"
	"github.com/pokt-foundation/portal-db/v2/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testUsers = map[types.UserID]*types.User{
	types.UserID("usertest"): {
		Email:    "test@test.com",
		SignedUp: true,
		AuthProviders: map[types.AuthType]types.UserAuthProvider{
			types.AuthTypeAuth0Github: {ProviderUserID: "auth0|test", Type: types.AuthTypeAuth0Github, Provider: types.AuthProviderAuth0},
		},
	},
}

func newTestRouter(t *testing.T) (*Router, error) {
	c := require.New(t)

	readerMock, _ := cache.NewReaderMock(t)

	readerMock.On("ReadUserPermissions", context.Background()).Return(testdata.UserPermissions, nil)
	readerMock.On("ReadUserIDsMap", context.Background()).Return(testdata.UserIDs, nil)
	readerMock.On("ReadPlans", context.Background()).Return(testdata.PayPlans, nil)
	readerMock.On("ReadAccounts", context.Background(), types.DriverOptions{IncludeDeleted: true}).Return(testdata.Accounts, nil)
	readerMock.On("ReadAccounts", mock.Anything, types.DriverOptions{IncludeDeleted: false}).Return(testdata.Accounts, nil)
	readerMock.On("ReadPortalApps", context.Background(), types.DriverOptions{IncludeDeleted: true}).Return(testdata.PortalApps, nil)
	readerMock.On("ReadGigastakeApps", context.Background(), types.DriverOptions{IncludeDeleted: true}).Return(testdata.GigastakeApps, nil)
	readerMock.On("ReadChains", context.Background(), types.DriverOptions{IncludeDeleted: true}).Return(testdata.Chains, nil)
	readerMock.On("ReadBlockedContracts", context.Background()).Return(testdata.GlobalBlockedContracts, nil)
	readerMock.On("ReadUserByUserID", mock.Anything, types.UserID("usertest")).Return(testUsers[types.UserID("usertest")], nil)

	logger := logger.New()

	cacheInstance, err := cache.NewCache(context.Background(), cache.Config{
		Reader:              readerMock,
		Logger:              logger,
		CacheRefreshMinutes: 5 * time.Minute,
	})
	c.NoError(err)

	return NewRouter(Config{
		Cache:   cacheInstance,
		Writer:  nil, // mock set in tests
		APIKeys: map[string]bool{"": true},
		Logger:  logger,
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

func TestRouter_getGigastakeAppByID(t *testing.T) {
	tests := []struct {
		name                 string
		gigastakeAppID       types.GigastakeAppID
		expectedStatus       int
		expectedGigastakeApp types.GigastakeApp
	}{
		{
			name:                 "Should fetch a single gigastake app",
			gigastakeAppID:       "test_gigastake_app_1",
			expectedStatus:       http.StatusOK,
			expectedGigastakeApp: *testdata.GigastakeApps["test_gigastake_app_1"],
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v2/gigastake/%s", test.gigastakeAppID), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var gigastakeAppResp types.GigastakeApp
			err = json.Unmarshal(rr.Body.Bytes(), &gigastakeAppResp)
			c.NoError(err)

			c.Equal(test.expectedGigastakeApp, gigastakeAppResp)
		})
	}
}

func TestRouter_getAllChains(t *testing.T) {
	tests := []struct {
		name            string
		chainID         types.RelayChainID
		includeInactive string
		expectedStatus  int
		expectedChains  map[types.RelayChainID]*types.Chain
	}{
		{
			name:            "Should fetch all active chains",
			chainID:         "0001",
			includeInactive: "false",
			expectedStatus:  http.StatusOK,
			expectedChains:  phdtestdata.FilterActiveChains(testdata.Chains),
		},
		{
			name:            "Should fetch all chains including inactive",
			chainID:         "0001",
			includeInactive: "true",
			expectedStatus:  http.StatusOK,
			expectedChains:  testdata.Chains,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			// Add the includeInactive parameter to the URL
			url := fmt.Sprintf("/v2/chain?include_inactive=%s", test.includeInactive)
			req, err := http.NewRequest(http.MethodGet, url, nil)
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

func TestRouter_getAllGigastakeApps(t *testing.T) {
	tests := []struct {
		name           string
		includeDeleted bool
		expectedStatus int
		expectedApps   map[types.GigastakeAppID]*types.GigastakeApp
	}{
		{
			name:           "Should fetch all gigastake apps",
			includeDeleted: false,
			expectedStatus: http.StatusOK,
			expectedApps:   testdata.GigastakeApps,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/v2/gigastake", nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var gigastakeAppsResp []types.GigastakeApp
			err = json.Unmarshal(rr.Body.Bytes(), &gigastakeAppsResp)
			c.NoError(err)
			gigastakeApps := map[types.GigastakeAppID]*types.GigastakeApp{}
			for i := range gigastakeAppsResp {
				gigastakeApps[gigastakeAppsResp[i].ID] = &gigastakeAppsResp[i]
			}
			c.Equal(test.expectedApps, gigastakeApps)
		})
	}
}

func TestRouter_getAllGigastakeAppsByChain(t *testing.T) {
	tests := []struct {
		name           string
		chainID        types.RelayChainID
		expectedStatus int
		expectedApps   map[types.GigastakeAppID]*types.GigastakeApp
	}{
		{
			name:           "Should fetch all gigastake apps for a single chain ID",
			chainID:        "0001",
			expectedStatus: http.StatusOK,
			expectedApps: map[types.GigastakeAppID]*types.GigastakeApp{
				"test_gigastake_app_1": testdata.GigastakeApps["test_gigastake_app_1"],
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v2/chain/%s/gigastake", test.chainID), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var gigastakeAppsByChainResp []types.GigastakeApp
			err = json.Unmarshal(rr.Body.Bytes(), &gigastakeAppsByChainResp)
			c.NoError(err)
			gigastakeAppsByChain := map[types.GigastakeAppID]*types.GigastakeApp{}
			for i := range gigastakeAppsByChainResp {
				gigastakeAppsByChain[gigastakeAppsByChainResp[i].ID] = &gigastakeAppsByChainResp[i]
			}
			c.Equal(test.expectedApps, gigastakeAppsByChain)
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
		expectedGigastakeApp *types.GigastakeApp
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
		chainID        types.RelayChainID
		chainUpdate    types.UpdateChain
		expectedStatus int
		expectedChain  *types.Chain
	}{
		{
			name:           "Should update a chain",
			chainID:        "0001",
			chainUpdate:    testdata.UpdateChainOne,
			expectedStatus: http.StatusOK,
			expectedChain: &types.Chain{
				ID:             "0001",
				Description:    *testdata.UpdateChainOne.Description,
				EnforceResult:  *testdata.UpdateChainOne.EnforceResult,
				Path:           *testdata.UpdateChainOne.Path,
				Ticker:         *testdata.UpdateChainOne.Ticker,
				AllowedMethods: testdata.UpdateChainOne.AllowedMethods,
				Active:         *testdata.UpdateChainOne.Active,
				Altruists:      *testdata.UpdateChainOne.Altruists,
				Checks:         *testdata.UpdateChainOne.Checks,
				Aliases:        *testdata.UpdateChainOne.Aliases,
				CreatedAt:      testdata.MockTimestamp,
				UpdatedAt:      testdata.MockTimestamp,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("UpdateChain", mock.Anything, test.chainUpdate, mock.Anything).Return(test.expectedChain, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			chainUpdateBytes, err := json.Marshal(test.chainUpdate)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/v2/chain/%s", test.chainID), bytes.NewBuffer(chainUpdateBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var chainResp *types.Chain
			err = json.Unmarshal(rr.Body.Bytes(), &chainResp)
			c.NoError(err)

			// Assert that the fields in the updated chain match the expected values
			// We compare the whole Chain object now instead of each field
			c.Equal(test.expectedChain, chainResp)
		})
	}
}

func TestRouter_updateGigastakeApp(t *testing.T) {
	tests := []struct {
		name               string
		gigastakeAppUpdate types.UpdateGigastakeApp
		expectedStatus     int
	}{
		{
			name: "Should update a Gigastake App",
			gigastakeAppUpdate: types.UpdateGigastakeApp{
				ID:       "test_gigastake_app_1",
				Name:     "pokt_gigastake_updated",
				ChainIDs: []types.RelayChainID{"0001", "0040", "0053"},
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
			writerMock.On("UpdateGigastakeApp", mock.Anything, test.gigastakeAppUpdate, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			gigastakeAppUpdateBytes, err := json.Marshal(test.gigastakeAppUpdate)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/v2/chain/gigastake/%s", test.gigastakeAppUpdate.ID), bytes.NewBuffer(gigastakeAppUpdateBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var gigastakeAppResp types.UpdateGigastakeApp
			err = json.Unmarshal(rr.Body.Bytes(), &gigastakeAppResp)
			c.NoError(err)
			c.Equal(test.gigastakeAppUpdate, gigastakeAppResp)
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

/* ---------- V2 Portal App Handler Method Unit Tests ---------- */

func TestRouter_getPortalAppByID(t *testing.T) {
	tests := []struct {
		name              string
		portalAppID       types.PortalAppID
		expectedStatus    int
		expectedPortalApp types.PortalApp
	}{
		{
			name:              "Should fetch a single portal application",
			portalAppID:       "test_app_1",
			expectedStatus:    http.StatusOK,
			expectedPortalApp: *testdata.PortalApps["test_app_1"],
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v2/portal_app/%s", test.portalAppID), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var portalAppResp types.PortalApp
			err = json.Unmarshal(rr.Body.Bytes(), &portalAppResp)
			c.NoError(err)
			c.Equal(test.expectedPortalApp, portalAppResp)
		})
	}
}

func TestRouter_getAllPortalApps(t *testing.T) {
	tests := []struct {
		name               string
		expectedStatus     int
		expectedPortalApps map[types.PortalAppID]*types.PortalApp
	}{
		{
			name:               "Should fetch all portal applications",
			expectedStatus:     http.StatusOK,
			expectedPortalApps: testdata.PortalApps,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/v2/portal_app", nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var portalAppsResp []types.PortalApp
			err = json.Unmarshal(rr.Body.Bytes(), &portalAppsResp)
			c.NoError(err)
			portalAppsRespMap := map[types.PortalAppID]*types.PortalApp{}
			for i := range portalAppsResp {
				portalAppsRespMap[portalAppsResp[i].ID] = &portalAppsResp[i]
			}
			c.Equal(test.expectedPortalApps, portalAppsRespMap)
		})
	}
}

func TestRouter_getPortalAppsForMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		expectedStatus     int
		expectedPortalApps map[types.PortalAppID]*types.PortalAppLite
	}{
		{
			name:               "Should fetch all portal applications for middleware",
			expectedStatus:     http.StatusOK,
			expectedPortalApps: testdata.PortalAppLites,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/v2/middleware/portal_app", nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var portalAppsResp []*types.PortalAppLite
			err = json.Unmarshal(rr.Body.Bytes(), &portalAppsResp)
			c.NoError(err)
			c.Equal(test.expectedPortalApps, portalAppLitesToMap(portalAppsResp))
		})
	}
}

func TestRouter_getPortalAppsByUser(t *testing.T) {
	tests := []struct {
		name               string
		userID             types.UserID
		roleFilter         types.RoleName
		expectedStatus     int
		expectedPortalApps []types.PortalApp
	}{
		{
			name:               "Should fetch all portal applications for a user",
			userID:             "user_1",
			roleFilter:         "",
			expectedStatus:     http.StatusOK,
			expectedPortalApps: []types.PortalApp{*testdata.PortalApps["test_app_1"]},
		},
		{
			name:               "Should fetch all portal applications for a user with a specific role",
			userID:             "user_4",
			roleFilter:         types.RoleMember,
			expectedStatus:     http.StatusOK,
			expectedPortalApps: []types.PortalApp{*testdata.PortalApps["test_app_2"]},
		},
		{
			name:               "Should fail if no portal apps are found for a given user",
			userID:             "user_4",
			roleFilter:         types.RoleAdmin,
			expectedStatus:     http.StatusNotFound,
			expectedPortalApps: []types.PortalApp{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			endpoint := fmt.Sprintf("/v2/user/%s/portal_app", test.userID)
			if test.roleFilter != "" {
				endpoint += fmt.Sprintf("?filters=%s", test.roleFilter)
			}

			req, err := http.NewRequest(http.MethodGet, endpoint, nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedStatus == http.StatusOK {
				var portalAppsResp []types.PortalApp
				err = json.Unmarshal(rr.Body.Bytes(), &portalAppsResp)
				c.NoError(err)
				c.Equal(test.expectedPortalApps, portalAppsResp)
			}
		})
	}
}

func TestRouter_createNewPortalApp(t *testing.T) {
	tests := []struct {
		name              string
		portalAppInput    types.PortalApp
		expectedStatus    int
		expectedPortalApp *types.PortalApp
	}{
		{
			name:              "Should create a new PortalApp",
			portalAppInput:    *testdata.TestCreatePortalApp,
			expectedStatus:    http.StatusOK,
			expectedPortalApp: testdata.TestCreatePortalApp,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("WritePortalApp", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.expectedPortalApp, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			portalAppInputBytes, err := json.Marshal(test.portalAppInput)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, "/v2/portal_app", bytes.NewBuffer(portalAppInputBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var portalAppResp *types.PortalApp
			err = json.Unmarshal(rr.Body.Bytes(), &portalAppResp)
			c.NoError(err)
			c.Equal(test.expectedPortalApp, portalAppResp)
		})
	}
}

func TestRouter_updatePortalApp(t *testing.T) {
	tests := []struct {
		name            string
		portalAppID     types.PortalAppID
		portalAppUpdate types.UpdatePortalApp
		expectedStatus  int
	}{
		{
			name:        "Should update PortalApp name only",
			portalAppID: "test_app_1",
			portalAppUpdate: types.UpdatePortalApp{
				Name: testdata.UpdatePortalAppName,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Should update PortalApp settings only",
			portalAppID: "test_app_1",
			portalAppUpdate: types.UpdatePortalApp{
				Settings: testdata.UpdatePortalAppSettings,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Should update PortalApp notifications only",
			portalAppID: "test_app_1",
			portalAppUpdate: types.UpdatePortalApp{
				Notifications: testdata.UpdatePortalAppNotifications,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Should update PortalApp whitelists only",
			portalAppID: "test_app_1",
			portalAppUpdate: types.UpdatePortalApp{
				Whitelists: testdata.UpdatePortalAppWhitelists,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Should update PortalApp plan only",
			portalAppID: "test_app_1",
			portalAppUpdate: types.UpdatePortalApp{
				PlanType:   testdata.UpdatePortalAppPlan.PlanType,
				DailyLimit: testdata.UpdatePortalAppPlan.DailyLimit,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Should update PortalApp enterprise plan only",
			portalAppID: "test_app_1",
			portalAppUpdate: types.UpdatePortalApp{
				PlanType:    testdata.UpdatePortalAppEnterprisePlan.PlanType,
				CustomLimit: testdata.UpdatePortalAppEnterprisePlan.CustomLimit,
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
			writerMock.On("UpdatePortalApp", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			portalAppUpdateBytes, err := json.Marshal(test.portalAppUpdate)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/v2/portal_app/%s", test.portalAppID), bytes.NewBuffer(portalAppUpdateBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var portalAppResp types.UpdatePortalApp
			err = json.Unmarshal(rr.Body.Bytes(), &portalAppResp)
			c.NoError(err)
			test.portalAppUpdate.AppID = test.portalAppID
			c.Equal(test.portalAppUpdate, portalAppResp)
		})
	}
}

func TestRouter_deletePortalApp(t *testing.T) {
	tests := []struct {
		name             string
		portalAppID      types.PortalAppID
		expectedStatus   int
		expectedResponse map[string]string
	}{
		{
			name:             "Should delete a Portal App",
			portalAppID:      "test_portal_app_1",
			expectedStatus:   http.StatusOK,
			expectedResponse: map[string]string{"status": "deleted"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("SetPortalAppDeleted", mock.Anything, test.portalAppID, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/portal_app/%s", string(test.portalAppID)), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var portalAppResp map[string]string
			err = json.Unmarshal(rr.Body.Bytes(), &portalAppResp)
			c.NoError(err)
			c.Equal(test.expectedResponse, portalAppResp)
		})
	}
}

func TestRouter_updatePortalAppFirstDateSurpassed(t *testing.T) {
	tests := []struct {
		name                     string
		firstDateSurpassedUpdate types.UpdateFirstDateSurpassed
		expectedStatus           int
	}{
		{
			name: "Should update PortalApp's FirstDateSurpassed field",
			firstDateSurpassedUpdate: types.UpdateFirstDateSurpassed{
				PortalAppIDs: []types.PortalAppID{
					"test_protocol_app_1",
					"test_protocol_app_2",
				},
				FirstDateSurpassed: testdata.MockTimestamp,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Should handle no AppIDs provided",
			firstDateSurpassedUpdate: types.UpdateFirstDateSurpassed{
				PortalAppIDs:       []types.PortalAppID{},
				FirstDateSurpassed: testdata.MockTimestamp,
			},
			expectedStatus: http.StatusBadRequest,
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

			firstDateSurpassedUpdateBytes, err := json.Marshal(test.firstDateSurpassedUpdate)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, "/v2/portal_app/first_date_surpassed", bytes.NewBuffer(firstDateSurpassedUpdateBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedStatus == http.StatusOK {
				var response map[string]string
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				c.NoError(err)
				c.Equal("updated", response["status"])
			}
		})
	}
}

/* ---------- V2 Account Handler Method Unit Tests ---------- */

func TestRouter_getAccounts(t *testing.T) {
	tests := []struct {
		name             string
		accountID        types.AccountID
		expectedStatus   int
		expectedAccounts int
	}{
		{
			name:             "Should get all accounts",
			expectedStatus:   http.StatusOK,
			expectedAccounts: 5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/v2/account", nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var accountsResp []types.Account
			err = json.Unmarshal(rr.Body.Bytes(), &accountsResp)
			c.NoError(err)
			c.Len(accountsResp, test.expectedAccounts)
		})
	}
}

func TestRouter_getAccountByID(t *testing.T) {
	tests := []struct {
		name             string
		accountID        types.AccountID
		expectedStatus   int
		expectedAccount  types.Account
		accountPlan      *types.Plan
		accountPortalApp *types.PortalApp
	}{
		{
			name:             "Should get one account by ID",
			accountID:        "account_1",
			expectedStatus:   http.StatusOK,
			expectedAccount:  *testdata.Accounts["account_1"],
			accountPlan:      testdata.PayPlans["basic_plan"],
			accountPortalApp: testdata.PortalApps["test_app_1"],
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v2/account/%s", test.accountID), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var accountResp types.Account
			err = json.Unmarshal(rr.Body.Bytes(), &accountResp)
			c.NoError(err)
			test.expectedAccount.Plan = test.accountPlan
			test.expectedAccount.PortalApps = make(map[types.PortalAppID]*types.PortalApp)
			test.expectedAccount.PortalApps[test.accountPortalApp.ID] = test.accountPortalApp
			c.Equal(test.expectedAccount, accountResp)
		})
	}
}

func TestRouter_getAccountsByUser(t *testing.T) {
	tests := []struct {
		name             string
		userID           types.UserID
		expectedStatus   int
		expectedData     []types.Account
		accountPlan      *types.Plan
		accountPortalApp *types.PortalApp
	}{
		{
			name:           "Should fetch all accounts for a given user ID",
			userID:         "user_1",
			expectedStatus: http.StatusOK,
			expectedData: []types.Account{
				*testdata.Accounts["account_1"],
			},
			accountPlan:      testdata.PayPlans["basic_plan"],
			accountPortalApp: testdata.PortalApps["test_app_1"],
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v2/user/%s/account", test.userID), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var accountsResp []types.Account
			err = json.Unmarshal(rr.Body.Bytes(), &accountsResp)
			c.NoError(err)
			test.expectedData[0].Plan = test.accountPlan
			test.expectedData[0].PortalApps = make(map[types.PortalAppID]*types.PortalApp)
			test.expectedData[0].PortalApps[test.accountPortalApp.ID] = test.accountPortalApp
			c.Equal(test.expectedData, accountsResp)
		})
	}
}

func TestRouter_createAccount(t *testing.T) {
	tests := []struct {
		name            string
		userID          types.UserID
		accountInput    types.Account
		expectedStatus  int
		expectedAccount *types.Account
	}{
		{
			name:           "Should create a new account",
			userID:         "user_2",
			accountInput:   *testdata.TestCreateAccount,
			expectedStatus: http.StatusOK,
			expectedAccount: &types.Account{
				ID:        "account_77",
				PlanType:  types.PayPlanType("developer_plan"),
				CreatedAt: testdata.MockTimestamp,
				UpdatedAt: testdata.MockTimestamp,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("WriteAccount", mock.Anything, test.userID, test.accountInput, mock.Anything).Return(test.expectedAccount, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			accountInputBytes, err := json.Marshal(test.accountInput)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/v2/user/%s/account", test.userID), bytes.NewBuffer(accountInputBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var accountResp *types.Account
			err = json.Unmarshal(rr.Body.Bytes(), &accountResp)
			c.NoError(err)
			c.Equal(test.expectedAccount, accountResp)
		})
	}
}

func TestRouter_updateAccount(t *testing.T) {
	tests := []struct {
		name               string
		accountID          types.AccountID
		updateAccountInput types.UpdateAccount
		expectedStatus     int
		expectedAccount    *types.Account
	}{
		{
			name:      "Should update an account",
			accountID: "account_1",
			updateAccountInput: types.UpdateAccount{
				AccountID: "account_1",
				PlanType:  types.TestPlan10K,
			},
			expectedStatus: http.StatusOK,
			expectedAccount: &types.Account{
				ID:       "account_1",
				PlanType: types.TestPlan10K,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("UpdateAccount", mock.Anything, test.updateAccountInput, mock.Anything).Return(test.expectedAccount, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			updateAccountInputBytes, err := json.Marshal(test.updateAccountInput)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/v2/account/%s", test.accountID), bytes.NewBuffer(updateAccountInputBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var accountResp *types.Account
			err = json.Unmarshal(rr.Body.Bytes(), &accountResp)
			c.NoError(err)
			c.Equal(test.expectedAccount, accountResp)
		})
	}
}

func TestRouter_createAccountIntegration(t *testing.T) {
	tests := []struct {
		name                string
		accountID           types.AccountID
		accountIntegrations types.AccountIntegrations
		expectedStatus      int
		expectedIntegration *types.AccountIntegrations
	}{
		{
			name:      "Should create an AccountIntegration",
			accountID: "account_1",
			accountIntegrations: types.AccountIntegrations{
				AccountID:          "account_1",
				CovalentAPIKeyFree: "covalent_free_key_1",
				CovalentAPIKeyPaid: "covalent_paid_key_1",
			},
			expectedStatus: http.StatusOK,
			expectedIntegration: &types.AccountIntegrations{
				AccountID:          "account_1",
				CovalentAPIKeyFree: "covalent_free_key_1",
				CovalentAPIKeyPaid: "covalent_paid_key_1",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("UpsertAccountIntegration", mock.Anything, test.accountIntegrations).Return(test.expectedIntegration, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			accountIntegrationsBytes, err := json.Marshal(test.accountIntegrations)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/v2/account/%s/integration", test.accountID), bytes.NewBuffer(accountIntegrationsBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var accountIntegrationResp *types.AccountIntegrations
			err = json.Unmarshal(rr.Body.Bytes(), &accountIntegrationResp)
			c.NoError(err)
			c.Equal(test.expectedIntegration, accountIntegrationResp)
		})
	}
}

func TestRouter_updateAccountIntegration(t *testing.T) {
	tests := []struct {
		name                string
		accountID           types.AccountID
		accountIntegrations types.AccountIntegrations
		expectedStatus      int
		expectedIntegration *types.AccountIntegrations
	}{
		{
			name:      "Should update an AccountIntegration",
			accountID: "account_1",
			accountIntegrations: types.AccountIntegrations{
				AccountID:          "account_1",
				CovalentAPIKeyFree: "covalent_free_key_1_updated",
				CovalentAPIKeyPaid: "covalent_paid_key_1_updated",
			},
			expectedStatus: http.StatusOK,
			expectedIntegration: &types.AccountIntegrations{
				AccountID:          "account_1",
				CovalentAPIKeyFree: "covalent_free_key_1_updated",
				CovalentAPIKeyPaid: "covalent_paid_key_1_updated",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("UpsertAccountIntegration", mock.Anything, test.accountIntegrations).Return(test.expectedIntegration, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			accountIntegrationsBytes, err := json.Marshal(test.accountIntegrations)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/v2/account/%s/integration", test.accountID), bytes.NewBuffer(accountIntegrationsBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var accountIntegrationResp *types.AccountIntegrations
			err = json.Unmarshal(rr.Body.Bytes(), &accountIntegrationResp)
			c.NoError(err)
			c.Equal(test.expectedIntegration, accountIntegrationResp)
		})
	}
}

func TestRouter_deleteAccount(t *testing.T) {
	tests := []struct {
		name           string
		accountID      types.AccountID
		expectedStatus int
	}{
		{
			name:           "Should set account's Deleted field to true",
			accountID:      "account_1",
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("SetAccountDeleted", mock.Anything, test.accountID, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/account/%s", test.accountID), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var deleteStatus map[string]string
			err = json.Unmarshal(rr.Body.Bytes(), &deleteStatus)
			c.NoError(err)
			c.Equal(map[string]string{"status": "deleted"}, deleteStatus)
		})
	}
}

/* ---------- V2 Account User Handler Method Unit Tests ---------- */

func TestRouter_writeAccountUser(t *testing.T) {
	tests := []struct {
		name            string
		createUserInput types.CreateAccountUserAccess
		expectedStatus  int
		expectedUserID  types.UserID
	}{
		{
			name: "Should create a new account user",
			createUserInput: types.CreateAccountUserAccess{
				AccountID:   "account_1",
				PortalAppID: "app_1",
				Email:       "test@example.com",
				RoleName:    "admin",
			},
			expectedStatus: http.StatusOK,
			expectedUserID: "user_1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("WriteAccountUser", mock.Anything, test.createUserInput, mock.Anything).Return(test.expectedUserID, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			createUserInputBytes, err := json.Marshal(test.createUserInput)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, "/v2/account/user", bytes.NewBuffer(createUserInputBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var userIDResp map[string]types.UserID
			err = json.Unmarshal(rr.Body.Bytes(), &userIDResp)
			c.NoError(err)
			c.Equal(test.expectedUserID, userIDResp["userID"])
		})
	}
}

func TestRouter_updateAccountUserRole(t *testing.T) {
	tests := []struct {
		name            string
		updateUserInput types.UpdateAccountUserRole
		expectedStatus  int
	}{
		{
			name: "Should update the role of an account user",
			updateUserInput: types.UpdateAccountUserRole{
				PortalAppID: "app_1",
				UserID:      "user_1",
				AccountID:   "account_1",
				RoleName:    "admin",
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
			writerMock.On("SetAccountUserRole", mock.Anything, test.updateUserInput, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			updateUserInputBytes, err := json.Marshal(test.updateUserInput)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, "/v2/account/user/update_role", bytes.NewBuffer(updateUserInputBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var statusResp map[string]string
			err = json.Unmarshal(rr.Body.Bytes(), &statusResp)
			c.NoError(err)
			c.Equal("updated", statusResp["status"])
		})
	}
}

func TestRouter_acceptAccountUser(t *testing.T) {
	tests := []struct {
		name            string
		acceptUserInput types.UpdateAcceptAccountUser
		expectedStatus  int
	}{
		{
			name: "Should accept an account user",
			acceptUserInput: types.UpdateAcceptAccountUser{
				PortalAppID:      "app_1",
				UserID:           "user_1",
				AuthProviderType: types.AuthType("auth_type_1"),
				ProviderUserID:   "providerUser_1",
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
			writerMock.On("UpdateAcceptAccountUser", mock.Anything, test.acceptUserInput, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			acceptUserInputBytes, err := json.Marshal(test.acceptUserInput)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, "/v2/account/user/accept", bytes.NewBuffer(acceptUserInputBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var statusResp map[string]string
			err = json.Unmarshal(rr.Body.Bytes(), &statusResp)
			c.NoError(err)
			c.Equal("accepted", statusResp["status"])
		})
	}
}

func TestRouter_removeAccountUser(t *testing.T) {
	tests := []struct {
		name            string
		removeUserInput types.UpdateRemoveAccountUser
		expectedStatus  int
	}{
		{
			name: "Should remove an account user",
			removeUserInput: types.UpdateRemoveAccountUser{
				UserID:      "user_1",
				PortalAppID: "app_1",
				AccountID:   "account_1",
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
			writerMock.On("RemoveAccountUser", mock.Anything, test.removeUserInput.UserID, test.removeUserInput.PortalAppID, test.removeUserInput.AccountID).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			removeUserInputBytes, err := json.Marshal(test.removeUserInput)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, "/v2/account/user/remove", bytes.NewBuffer(removeUserInputBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var statusResp map[string]string
			err = json.Unmarshal(rr.Body.Bytes(), &statusResp)
			c.NoError(err)
			c.Equal("removed", statusResp["status"])
		})
	}
}

/* ---------- V2 User Handler Method Unit Tests ---------- */

func TestRouter_getUserPermissionByUserID(t *testing.T) {
	tests := []struct {
		name                   string
		userID                 types.ProviderUserID
		expectedUserPermission *types.UserPermissions
		expectedStatus         int
	}{
		{
			name:   "Should fetch user permissions for a given user ID",
			userID: "auth0|ellen_ripley",
			expectedUserPermission: &types.UserPermissions{
				UserID: "user_3",
				PortalApps: map[types.PortalAppID]types.PortalAppPermissions{
					"test_app_2": {
						RoleName: types.RoleOwner,
						Permissions: []types.Permissions{
							"read:endpoint", "write:endpoint", "delete:endpoint", "transfer:endpoint",
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:                   "Should return 404 status for non-existent user ID",
			userID:                 "auth0|sir_robin",
			expectedUserPermission: nil,
			expectedStatus:         http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v2/user/%s/permission", test.userID), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedUserPermission != nil {
				var userPermissionResp *types.UserPermissions
				err = json.Unmarshal(rr.Body.Bytes(), &userPermissionResp)
				c.NoError(err)
				c.Equal(test.expectedUserPermission, userPermissionResp)
			}
		})
	}
}

func TestRouter_getPortalUserIDFromProviderUserID(t *testing.T) {
	tests := []struct {
		name           string
		userID         types.ProviderUserID
		expectedUserID types.UserID
		expectedStatus int
	}{
		{
			name:           "Should fetch Portal User ID for a given auth provider user ID",
			userID:         "auth0|ellen_ripley",
			expectedUserID: "user_3",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Should return 404 status for non-existent user ID",
			userID:         "auth0|sir_robin",
			expectedUserID: "",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v2/user/%s", test.userID), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedStatus == http.StatusOK {
				var portalUserIDResp types.UserID
				err = json.Unmarshal(rr.Body.Bytes(), &portalUserIDResp)
				c.NoError(err)
				c.Equal(test.expectedUserID, portalUserIDResp)
			}
		})
	}
}

func TestRouter_createUser(t *testing.T) {
	tests := []struct {
		name                 string
		createUserInput      types.CreateUser
		expectedStatus       int
		expectedUserResponse *types.CreateUserResponse
	}{
		{
			name: "Should create a new user",
			createUserInput: types.CreateUser{
				Email:          "ellen_ripley@nostromo.space",
				ProviderUserID: "auth0|ellen_ripley",
			},
			expectedStatus: http.StatusOK,
			expectedUserResponse: &types.CreateUserResponse{
				User: types.User{
					ID:       "user_3",
					Email:    "ellen_ripley@nostromo.space",
					SignedUp: true,
					AuthProviders: map[types.AuthType]types.UserAuthProvider{
						"auth0": {
							UserID:         "user_3",
							ProviderUserID: "auth0|ellen_ripley",
							Type:           "auth0",
							Provider:       "auth0",
							Federated:      true,
						},
					},
					CreatedAt: testdata.MockTimestamp,
					UpdatedAt: testdata.MockTimestamp,
				},
				AccountID: "account_42",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("WriteUserNewSignUp", mock.Anything, test.createUserInput, mock.Anything).Return(&test.expectedUserResponse.User, test.expectedUserResponse.AccountID, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			userInputBytes, err := json.Marshal(test.createUserInput)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, "/v2/user", bytes.NewBuffer(userInputBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var userResp *types.CreateUserResponse
			err = json.Unmarshal(rr.Body.Bytes(), &userResp)
			c.NoError(err)
			c.Equal(test.expectedUserResponse, userResp)
		})
	}
}

func TestRouter_deleteUser(t *testing.T) {
	tests := []struct {
		name            string
		userID          types.UserID
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "Should delete a user",
			userID:          "user_3",
			expectedStatus:  http.StatusOK,
			expectedMessage: "deleted",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("DeletePortalUser", mock.Anything, test.userID).Return(types.UserID(""), nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/user/%s", test.userID), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var message map[string]string
			err = json.Unmarshal(rr.Body.Bytes(), &message)
			c.NoError(err)
			c.Equal(test.expectedMessage, message["status"])
		})
	}
}

/* ---------- V2 Global Blocked Contracts Handler Method Unit Tests ---------- */

func TestRouter_getPlans(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectedPlans  map[types.PayPlanType]*types.Plan
	}{
		{
			name:           "Should fetch all pay plans",
			expectedStatus: http.StatusOK,
			expectedPlans:  testdata.PayPlans,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/v2/plan", nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var plans []types.Plan
			err = json.Unmarshal(rr.Body.Bytes(), &plans)
			c.NoError(err)
			expectedPlansMap := make(map[types.PayPlanType]types.Plan)
			plansMap := make(map[types.PayPlanType]types.Plan)
			for _, plan := range test.expectedPlans {
				expectedPlansMap[plan.Type] = *plan
			}
			for _, plan := range plans {
				plansMap[plan.Type] = plan
			}
			c.Equal(expectedPlansMap, plansMap)
		})
	}
}

/* ---------- V2 Global Blocked Contracts Handler Method Unit Tests ---------- */

func TestRouter_getBlockedContracts(t *testing.T) {
	tests := []struct {
		name                string
		expectedStatus      int
		expectedBlockedCons map[types.BlockedAddress]struct{}
	}{
		{
			name:                "Should fetch all blocked contracts",
			expectedStatus:      http.StatusOK,
			expectedBlockedCons: testdata.GlobalBlockedContracts.BlockedAddresses,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/v2/blocked_contract", nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var blockedContracts types.GlobalBlockedContracts
			err = json.Unmarshal(rr.Body.Bytes(), &blockedContracts)
			c.NoError(err)
			c.Equal(test.expectedBlockedCons, blockedContracts.BlockedAddresses)
		})
	}
}

func TestRouter_writeBlockedContract(t *testing.T) {
	tests := []struct {
		name             string
		blockedContract  types.BlockedContract
		expectedStatus   int
		expectedResponse map[string]string
	}{
		{
			name: "Should add a new blocked contract",
			blockedContract: types.BlockedContract{
				BlockedAddress: "0xtest_0123456789abcdef0123456789abcdef012345",
				Active:         true,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: map[string]string{"status": "added"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("WriteBlockedContract", mock.Anything, test.blockedContract.BlockedAddress, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			blockedContractBytes, err := json.Marshal(test.blockedContract)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPost, "/v2/blocked_contract", bytes.NewBuffer(blockedContractBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var response map[string]string
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			c.NoError(err)
			c.Equal(test.expectedResponse, response)
		})
	}
}

func TestRouter_updateBlockedContractActive(t *testing.T) {
	tests := []struct {
		name             string
		blockedAddress   types.BlockedAddress
		blockedInput     types.BlockedContract
		expectedStatus   int
		expectedResponse map[string]bool
	}{
		{
			name:           "Should update blocked contract active status to true",
			blockedAddress: "0xtest_6789abcdef0123456789abcdef01234567",
			blockedInput: types.BlockedContract{
				BlockedAddress: "0xtest_6789abcdef0123456789abcdef01234567",
				Active:         true,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: map[string]bool{"contract_active": true},
		},
		{
			name:           "Should update blocked contract active status to false",
			blockedAddress: "0xtest_6789abcdef0123456789abcdef01234567",
			blockedInput: types.BlockedContract{
				BlockedAddress: "0xtest_6789abcdef0123456789abcdef01234567",
				Active:         false,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: map[string]bool{"contract_active": false},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("UpdateBlockedContractActive", mock.Anything, test.blockedAddress, test.blockedInput.Active, mock.Anything).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			blockedInputBytes, err := json.Marshal(test.blockedInput)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/v2/blocked_contract/%s/active", test.blockedAddress), bytes.NewBuffer(blockedInputBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var response map[string]bool
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			c.NoError(err)
			c.Equal(test.expectedResponse, response)
		})
	}
}

func TestRouter_removeBlockedContract(t *testing.T) {
	tests := []struct {
		name             string
		blockedAddress   types.BlockedAddress
		expectedStatus   int
		expectedResponse map[string]string
	}{
		{
			name:             "Should remove a blocked contract",
			blockedAddress:   "0xtest_6789abcdef0123456789abcdef01234567",
			expectedStatus:   http.StatusOK,
			expectedResponse: map[string]string{"status": "deleted"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			writerMock := driver.NewMockDriver(t)
			writerMock.On("RemoveBlockedContract", mock.Anything, test.blockedAddress).Return(nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/blocked_contract/%s", test.blockedAddress), nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var response map[string]string
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			c.NoError(err)
			c.Equal(test.expectedResponse, response)
		})
	}
}

func portalAppLitesToMap(portalAppLites []*types.PortalAppLite) map[types.PortalAppID]*types.PortalAppLite {
	portalAppLitesMap := make(map[types.PortalAppID]*types.PortalAppLite)
	for i := range portalAppLites {
		strKeys := make([]string, len(portalAppLites[i].PublicKeys))
		for j, key := range portalAppLites[i].PublicKeys {
			strKeys[j] = string(key)
		}

		sort.Strings(strKeys)

		for j, key := range strKeys {
			portalAppLites[i].PublicKeys[j] = types.PortalAppPublicKey(key)
		}

		portalAppLitesMap[portalAppLites[i].ID] = portalAppLites[i]
	}
	return portalAppLitesMap
}
