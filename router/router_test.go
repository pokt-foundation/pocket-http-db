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
	"github.com/pokt-foundation/utils-go/logger"

	"github.com/pokt-foundation/portal-db/v2/driver"
	"github.com/pokt-foundation/portal-db/v2/testdata"
	"github.com/pokt-foundation/portal-db/v2/types"
	"github.com/pokt-foundation/portal-http-db/cache"
	phdtestdata "github.com/pokt-foundation/portal-http-db/testdata"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testCtx = context.Background()

func newTestRouter(t *testing.T) (*Router, error) {
	c := require.New(t)

	readerMock, _ := cache.NewReaderMock(t)

	readerMock.On("ReadUserIDsMap", testCtx).Return(testdata.UserIDs, nil)
	readerMock.On("ReadAllUsers", testCtx).Return(testdata.Users, nil)
	readerMock.On("ReadPlans", testCtx).Return(testdata.PayPlans, nil)
	readerMock.On("ReadAccounts", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(testdata.Accounts, nil)
	readerMock.On("ReadAccounts", mock.Anything, types.DriverOptions{IncludeDeleted: false}).Return(testdata.Accounts, nil)
	readerMock.On("ReadPortalApps", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(testdata.PortalApps, nil)
	readerMock.On("ReadGigastakeApps", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(testdata.GigastakeApps, nil)
	readerMock.On("ReadChains", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(testdata.Chains, nil)
	readerMock.On("ReadBlockedContracts", testCtx).Return(testdata.GlobalBlockedContracts, nil)
	readerMock.On("ReadUserByUserID", mock.Anything, types.UserID("usertest")).Return(testUsers[types.UserID("usertest")], nil)

	logger := logger.New()

	cacheInstance, err := cache.NewCache(testCtx, cache.Config{
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
				Blockchain:     *testdata.UpdateChainOne.Blockchain,
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
		portalAppUsers     map[types.PortalAppID]map[types.UserID]*types.AccountUserAccess
	}{
		// {
		// 	name:               "Should fetch all portal applications for a user",
		// 	userID:             "user_1",
		// 	roleFilter:         "",
		// 	expectedStatus:     http.StatusOK,
		// 	expectedPortalApps: []types.PortalApp{*testdata.PortalApps["test_app_1"]},
		// 	portalAppUsers:     testdata.PortalAppUsers,
		// },
		{
			name:               "Should fetch all portal applications for a user with a specific role",
			userID:             "user_4",
			roleFilter:         types.RoleMember,
			expectedStatus:     http.StatusOK,
			expectedPortalApps: []types.PortalApp{*testdata.PortalApps["test_app_2"]},
			portalAppUsers:     testdata.PortalAppUsers,
		},
		// {
		// 	name:               "Should fail if no portal apps are found for a given user",
		// 	userID:             "user_4",
		// 	roleFilter:         types.RoleAdmin,
		// 	expectedStatus:     http.StatusNotFound,
		// 	expectedPortalApps: []types.PortalApp{},
		// },
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
				for i, portalApp := range test.expectedPortalApps {
					updatedPortalApp := portalApp
					updatedPortalApp.Users = test.portalAppUsers[portalApp.ID]
					test.expectedPortalApps[i] = updatedPortalApp
				}
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
				PlanType: testdata.UpdatePortalAppPlan.PlanType,
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

func TestRouter_getUserAccounts(t *testing.T) {
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

func TestRouter_getUserAccount(t *testing.T) {
	tests := []struct {
		name             string
		accountID        types.AccountID
		userID           types.UserID
		expectedStatus   int
		expectedAccount  types.Account
		accountPlan      *types.Plan
		accountPortalApp *types.PortalApp
	}{
		{
			name:             "Should get one user account by ID",
			accountID:        "account_1",
			userID:           "user_1",
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

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v2/user/%s/account/%s", test.userID, test.accountID), nil)
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

func TestRouter_updateUser(t *testing.T) {
	tests := []struct {
		name            string
		updateUserInput types.UpdateUser
		updatedUser     *types.User
		expectedStatus  int
	}{
		{
			name:            "Should update an existing user",
			updateUserInput: testdata.UpdateUserOne,
			updatedUser: &types.User{
				ID:               "user_5",
				IconURL:          "https://picsum.photos/227",
				UpdatesProduct:   false,
				UpdatesMarketing: false,
				BetaTester:       true,
				UpdatedAt:        testdata.MockTimestamp,
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
			writerMock.On("UpdateUser", mock.Anything, test.updateUserInput, mock.Anything).Return(test.updatedUser, nil).Once()
			router.writer = writerMock

			rr := httptest.NewRecorder()

			updateUserInputBytes, err := json.Marshal(test.updateUserInput)
			c.NoError(err)

			req, err := http.NewRequest(http.MethodPut, "/v2/user", bytes.NewBuffer(updateUserInputBytes))
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			var updatedUser *types.User
			err = json.Unmarshal(rr.Body.Bytes(), &updatedUser)
			c.NoError(err)
			c.Equal(test.updatedUser, updatedUser)
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

func TestRouter_getPortalUser(t *testing.T) {
	tests := []struct {
		name           string
		providerUserID types.ProviderUserID
		portalUserID   types.UserID
		expectedUser   types.User
		expectedUserID types.UserID
		fullDetails    bool
		expectedStatus int
	}{
		{
			name:           "Should fetch Portal User for a given auth provider user ID if `full_details` is true",
			providerUserID: "auth0|ellen_ripley",
			expectedUser:   *testdata.Users["user_3"],
			fullDetails:    true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Should fetch Portal User ID for a given portal user ID if `full_details` is true",
			portalUserID:   "user_3",
			expectedUser:   *testdata.Users["user_3"],
			fullDetails:    true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Should fetch Portal User ID for a given auth provider user ID if `full_details` is false",
			providerUserID: "auth0|ellen_ripley",
			expectedUserID: "user_3",
			fullDetails:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Should fetch Portal User ID for a given portal user ID if `full_details` is false",
			portalUserID:   "user_3",
			expectedUserID: "user_3",
			fullDetails:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Should return 404 status for non-existent user ID",
			providerUserID: "auth0|sir_robin",
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

			var userID string
			switch test.providerUserID {
			case "":
				userID = string(test.portalUserID)
			default:
				userID = string(test.providerUserID)
			}

			endpoint := fmt.Sprintf("/v2/user/%s", userID)
			if test.fullDetails {
				endpoint = fmt.Sprintf("%s?full_details=true", endpoint)
			}

			req, err := http.NewRequest(http.MethodGet, endpoint, nil)
			c.NoError(err)

			router.Router.ServeHTTP(rr, req)
			c.Equal(test.expectedStatus, rr.Code)

			if test.expectedStatus == http.StatusOK {
				switch test.fullDetails {
				case true:
					var portalUserResp types.User
					err = json.Unmarshal(rr.Body.Bytes(), &portalUserResp)
					c.NoError(err)
					c.Equal(test.expectedUser, portalUserResp)
				default:
					var portalUserIDResp types.UserID
					err = json.Unmarshal(rr.Body.Bytes(), &portalUserIDResp)
					c.NoError(err)
					c.Equal(test.expectedUserID, portalUserIDResp)
				}
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
						RoleName:    v1Types.RoleOwner,
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
		name            string
		includeInactive string
		expectedChains  map[string]v1Types.Blockchain
		expectedStatus  int
	}{
		{
			name:            "Should fetch all active blockchains",
			includeInactive: "false",
			expectedChains:  phdtestdata.FilterActiveBlockchains(testBlockchains),
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "Should fetch all blockchains including inactive",
			includeInactive: "true",
			expectedChains:  testBlockchains,
			expectedStatus:  http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			router, err := newTestRouter(t)
			c.NoError(err)

			rr := httptest.NewRecorder()

			url := fmt.Sprintf("/v1/blockchain?include_inactive=%s", test.includeInactive)
			req, err := http.NewRequest(http.MethodGet, url, nil)
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
				PortalAppIDs:       []types.PortalAppID{"test_app_1", "test_app_2", "test_app_3"},
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
					RoleName: v1Types.RoleOwner,
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
					RoleName: v1Types.RoleOwner,
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
					RoleName: v1Types.RoleOwner,
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
		"0001-POKT-pokt-mainnet": {
			ID:        "0001-POKT-pokt-mainnet",
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
		"0053-OP-optimism-mainnet": {
			ID:        "0053-OP-optimism-mainnet",
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
		"0040-HMY-harmony-0": {
			ID:        "0040-HMY-harmony-0",
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
				{Alias: "pokt-mainnet", Domain: "pokt-rpc.gateway.pokt.network", LoadBalancerID: "0001-POKT-pokt-mainnet"},
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
				{Alias: "eth-mainnet", Domain: "eth-rpc.gateway.pokt.network", LoadBalancerID: "0021-ETH-eth-mainnet"},
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
				{Alias: "harmony-0", Domain: "hmy-rpc.gateway.pokt.network", LoadBalancerID: "0040-HMY-harmony-0"},
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
				{Alias: "optimism-mainnet", Domain: "op-rpc.gateway.pokt.network", LoadBalancerID: "0053-OP-optimism-mainnet"},
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
