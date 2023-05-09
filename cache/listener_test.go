package cache

import (
	"testing"
	"time"

	"github.com/pokt-foundation/portal-db/v2/testdata"
	"github.com/pokt-foundation/portal-db/v2/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func newTestListenerCache(t *testing.T, readerMock *ReaderMock) *Cache {
	c := require.New(t)

	readerMock.On("ReadUserPermissions", testCtx).Return(testdata.UserPermissions, nil)
	readerMock.On("ReadUserIDsMap", testCtx).Return(testdata.UserIDs, nil)
	readerMock.On("ReadPlans", testCtx).Return(testdata.PayPlans, nil)
	readerMock.On("ReadAccounts", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(testdata.Accounts, nil)
	readerMock.On("ReadPortalApps", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(testdata.PortalApps, nil)
	readerMock.On("ReadChains", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(testdata.Chains, nil)
	readerMock.On("ReadBlockedContracts", testCtx).Return(testdata.GlobalBlockedContracts, nil)

	cacheInstance, err := NewCache(testCtx, Config{
		Reader:              readerMock,
		Logger:              logrus.New(),
		CacheRefreshMinutes: 5 * time.Minute,
	})
	c.NoError(err)

	return cacheInstance
}

/* ---------- Accounts ---------- */

func Test_parseAccountNotification(t *testing.T) {
	tests := []struct {
		name             string
		mainTableAction  types.Action
		sideTablesAction types.Action
		accountID        types.AccountID
		account          *types.Account
	}{
		{
			name:             "Should insert Account in cache when receiving an insert notification",
			mainTableAction:  types.ActionInsert,
			sideTablesAction: types.ActionInsert,
			accountID:        "account_6",
			account:          testdata.TestCreateAccount,
		},
		{
			name:             "Should update Account in cache when receiving an update notification",
			mainTableAction:  types.ActionUpdate,
			sideTablesAction: types.ActionUpdate,
			accountID:        "account_1",
			account: &types.Account{
				PlanType:               types.PayPlanType("enterprise_plan"),
				PartnerChainIDs:        map[types.RelayChainID]struct{}{"0001": {}},
				PartnerThroughputLimit: 1_000,
				PartnerAppLimit:        2,
			},
		},
		{
			name:             "Should delete Account in cache when receiving an update notification with Account.Deleted = true",
			mainTableAction:  types.ActionUpdate,
			sideTablesAction: types.ActionUpdate,
			accountID:        "account_1",
			account: &types.Account{
				PlanType:               types.PayPlanType("enterprise_plan"),
				PartnerChainIDs:        map[types.RelayChainID]struct{}{"0001": {}},
				PartnerThroughputLimit: 1_000,
				PartnerAppLimit:        2,
				Deleted:                true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			readerMock := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)

			test.account.ID = test.accountID
			readerMock.listener.MockEvent(test.mainTableAction, test.sideTablesAction, test.account)

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			account := cache.GetAccount(test.accountID)
			c.Equal(test.account.PlanType, account.PlanType)
			c.Equal(test.account.PartnerChainIDs, account.PartnerChainIDs)
			c.Equal(test.account.PartnerThroughputLimit, account.PartnerThroughputLimit)
			c.Equal(test.account.PartnerAppLimit, account.PartnerAppLimit)

			accountFound := findAccountInAccounts(cache.GetAccounts(Options{IncludeDeleted: false}), test.accountID)
			if test.account.Deleted {
				c.False(accountFound, "Account should be removed from accounts map")
			} else {
				c.True(accountFound, "Account should be found in accounts map")
			}
		})
	}
}

func Test_parseAccountUserAccessNotification(t *testing.T) {
	tests := []struct {
		name            string
		sideTableAction types.Action
		userID          types.UserID
		accountID       types.AccountID
		account         *types.Account
	}{
		{
			name:            "Should upsert AccountUserAccess in cache when receiving an insert notification",
			sideTableAction: types.ActionInsert,
			userID:          "user_3",
			accountID:       "account_1",
			account: &types.Account{
				Users: map[types.UserID]types.AccountUserAccess{
					"user_3": testdata.AccountUserAccess[3],
				},
			},
		},
		{
			name:            "Should upsert AccountUserAccess in cache when receiving an update notification",
			sideTableAction: types.ActionUpdate,
			userID:          "user_2",
			accountID:       "account_1",
			account: &types.Account{
				Users: map[types.UserID]types.AccountUserAccess{
					"user_3": {
						UserID:   "user_2",
						Email:    "paul.atreides456@test.com",
						RoleName: types.RoleMember,
						Accepted: true,
						ProviderUserIDs: map[types.AuthType]string{
							types.AuthTypeAuth0Username: "auth0|paul_atreides",
							types.AuthTypeAuth0Github:   "github|paul_atreides",
						},
					},
				},
			},
		},
		{
			name:            "Should delete AccountUserAccess from cache when receiving a delete notification",
			sideTableAction: types.ActionDelete,
			userID:          "user_8",
			accountID:       "account_1",
			account: &types.Account{
				Users: map[types.UserID]types.AccountUserAccess{
					"user_8": testdata.AccountUserAccess[8],
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			readerMock := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)

			test.account.ID = test.accountID
			readerMock.listener.MockEvent("", test.sideTableAction, test.account)

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			if test.sideTableAction == types.ActionDelete {
				accountFound := findAccountInAccounts(cache.GetAccountsByUserID(test.userID), test.accountID)
				c.False(accountFound, "Account should be removed from accountsByUserID map")
			} else {
				accountFound := findAccountInAccounts(cache.GetAccountsByUserID(test.userID), test.accountID)
				c.True(accountFound, "Account should be found in accountsByUserID map before deletion")
			}
		})
	}
}

func Test_parsePortalAppNotification(t *testing.T) {
	tests := []struct {
		name             string
		mainTableAction  types.Action
		sideTablesAction types.Action
		portalAppID      types.PortalAppID
		portalApp        types.PortalApp
	}{
		{
			name:             "Should insert PortalApp in cache when receiving an insert notification",
			mainTableAction:  types.ActionInsert,
			sideTablesAction: types.ActionInsert,
			portalAppID:      "test_app_create_e2887s4",
			portalApp: types.PortalApp{
				ID:        "test_app_create_e2887s4",
				AccountID: "account_1",
				Name:      "create_pokt_app_2",
				Gigastake: true,
				Staked:    false,
				AATs: map[types.ProtocolAppID]types.AAT{
					"test_protocol_app_1": {
						ID:              "test_protocol_app_1",
						Address:         "test_162789bdf3fec1da21e3b28a757bb648",
						PublicKey:       "test_360c8eb90ef9536a888a843689a5a20c",
						ClientPublicKey: "test_97ae5cb13fa6c67cb3e5904ba79e842e",
						PrivateKey:      "test_dceabfd53f9b9c90552ba1eef2c49c78",
						Signature:       "test_6505b12fc22692a4d794005f7e253b2e",
						Version:         "0.0.1",
					},
				},
				Settings: types.Settings{
					Environment:       types.EnvironmentProduction,
					SecretKey:         "test_59a1cc884196e66cdfabe1f130159124",
					SecretKeyRequired: true,
				},
			},
		},
		{
			name:             "Should update PortalApp in cache when receiving an update notification",
			mainTableAction:  types.ActionUpdate,
			sideTablesAction: types.ActionUpdate,
			portalAppID:      "test_app_update_b03ca84c",
			portalApp:        *testdata.TestUpdatePortalApp,
		},
		{
			name:             "Should delete Account in cache when receiving an update notification with Account.Deleted = true",
			mainTableAction:  types.ActionUpdate,
			sideTablesAction: types.ActionUpdate,
			portalAppID:      "test_app_update_b03ca84c",
			portalApp: types.PortalApp{
				ID:        "test_app_create_e2887s4",
				AccountID: "account_1",
				Name:      "create_pokt_app_2",
				Deleted:   true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			readerMock := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)

			test.portalApp.ID = test.portalAppID
			readerMock.listener.MockEvent(test.mainTableAction, test.sideTablesAction, &test.portalApp)

			time.Sleep(200 * time.Millisecond) // need time for cache refresh

			if !test.portalApp.Deleted {
				c.Equal(test.portalApp, cache.GetPortalApp(test.portalAppID))
			}

			portalAppFound := findPortalAppInPortalApps(cache.GetPortalApps(Options{IncludeDeleted: false}), test.portalAppID)
			if test.portalApp.Deleted {
				c.False(portalAppFound, "PortalApp should be removed from portal apps map")
			} else {
				c.True(portalAppFound, "PortalApp should be found in portal apps map")
			}
		})
	}
}

func Test_parseAATNotification(t *testing.T) {
	tests := []struct {
		name            string
		sideTableAction types.Action
		portalApp       *types.PortalApp
	}{
		{
			name:            "Should add AAT to PortalApp in cache when receiving an insert notification",
			sideTableAction: types.ActionInsert,
			portalApp: &types.PortalApp{
				AATs: map[types.ProtocolAppID]types.AAT{
					"test_protocol_app_3": {
						ID:              "test_protocol_app_3",
						Address:         "test_162789bdf3fec1da21e3b28a757bb648",
						PublicKey:       "test_360c8eb90ef9536a888a843689a5a20c",
						ClientPublicKey: "test_97ae5cb13fa6c67cb3e5904ba79e842e",
						PrivateKey:      "test_dceabfd53f9b9c90552ba1eef2c49c78",
						Signature:       "test_6505b12fc22692a4d794005f7e253b2e",
						Version:         "0.0.1",
					},
				},
			},
		},
		{
			name:            "Should update AAT in PortalApp in cache when receiving an update notification",
			sideTableAction: types.ActionUpdate,
			portalApp: &types.PortalApp{
				AATs: map[types.ProtocolAppID]types.AAT{
					"test_protocol_app_3": {
						ID:              "test_protocol_app_3",
						Address:         "test_80bd695bd00e788aefb90931beeb4354",
						PublicKey:       "test_d2b20869b2dad0c0e1f87fb0e291aeac",
						ClientPublicKey: "test_97ba2aaf626c3bbb90f0d3331867a075",
						PrivateKey:      "test_fd8b15e5d309e2837bc563b82ef71a9c",
						Signature:       "test_bef7b8a4562c500d9a16802364d4eb57",
						Version:         "0.0.1",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			readerMock := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)

			// Add the PortalApp to the cache
			cache.upsertPortalApp(*test.portalApp)

			// test.portalApp.AATs.AppID = "" // is unset in the cache TODO will be changed soon

			readerMock.listener.MockEvent("", test.sideTableAction, test.portalApp)

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			portalApp := cache.GetPortalApp(test.portalApp.ID)
			c.Equal(test.portalApp.AATs, portalApp.AATs)
		})
	}
}

func Test_parseSettingsNotification(t *testing.T) {
	tests := []struct {
		name            string
		sideTableAction types.Action
		portalApp       *types.PortalApp
	}{
		{
			name:            "Should add Settings to PortalApp in cache when receiving an insert notification",
			sideTableAction: types.ActionInsert,
			portalApp: &types.PortalApp{
				Settings: types.Settings{
					AppID:             "test_app_3",
					SecretKey:         "test_3eb65b62717d9e062610d92abfe98873",
					MonthlyRelayLimit: 50_000_000,
				},
			},
		},
		{
			name:            "Should update Settings in PortalApp in cache when receiving an update notification",
			sideTableAction: types.ActionUpdate,
			portalApp: &types.PortalApp{
				Settings: types.Settings{
					AppID:             "test_app_3",
					SecretKey:         "test_6faf7a9f573e9f7a1ee44f85faae1ffd",
					MonthlyRelayLimit: 120_000_000,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			readerMock := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)

			// Add the PortalApp to the cache
			cache.upsertPortalApp(*test.portalApp)

			test.portalApp.Settings.AppID = "" // is unset in the cache TODO will be changed soon

			readerMock.listener.MockEvent("", test.sideTableAction, test.portalApp)

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			portalApp := cache.GetPortalApp(test.portalApp.ID)
			c.Equal(test.portalApp.Settings, portalApp.Settings)
		})
	}
}

func Test_parseWhitelistNotification(t *testing.T) {
	tests := []struct {
		name            string
		sideTableAction types.Action
		chain           types.RelayChainID
		portalApp       *types.PortalApp
	}{
		{
			name:            "Should add Whitelist to PortalApp in cache when receiving an insert notification",
			sideTableAction: types.ActionInsert,
			chain:           "0082",
			portalApp: &types.PortalApp{
				ID: "test_app_2",
				Whitelists: types.Whitelists{
					Blockchains: map[types.RelayChainID]struct{}{"0082": {}},
				},
			},
		},
		{
			name:            "Should delete Whitelist from PortalApp in cache when receiving a delete notification",
			sideTableAction: types.ActionDelete,
			chain:           "0021",
			portalApp: &types.PortalApp{
				ID: "test_app_2",
				Whitelists: types.Whitelists{
					Blockchains: map[types.RelayChainID]struct{}{"0021": {}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			readerMock := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)

			if test.sideTableAction == types.ActionDelete {
				storedPortalApp := cache.GetPortalApp(test.portalApp.ID)
				ok := storedPortalApp.IsBlockchainWhitelisted(test.chain)
				c.True(ok, "Whitelist should be present in the PortalApp in cache before deletion")
			}

			readerMock.listener.MockEvent("", test.sideTableAction, test.portalApp)

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			storedPortalApp := cache.GetPortalApp(test.portalApp.ID)
			ok := storedPortalApp.IsBlockchainWhitelisted(test.chain)
			if test.sideTableAction == types.ActionInsert {
				c.True(ok, "Whitelist should be added to the PortalApp in cache")
			} else if test.sideTableAction == types.ActionDelete {
				c.False(ok, "Whitelist should be removed from the PortalApp in cache")
			}
		})
	}
}

func Test_parseChainNotification(t *testing.T) {
	tests := []struct {
		name             string
		mainTableAction  types.Action
		sideTablesAction types.Action
		chain            types.Chain
	}{
		{
			name:             "Should insert Chain to cache when receiving an insert notification",
			mainTableAction:  types.ActionInsert,
			sideTablesAction: types.ActionInsert,
			chain: types.Chain{
				ID:            "0008",
				Blockchain:    "boba-mainnet",
				Description:   "Boba",
				EnforceResult: "JSON",
				Ticker:        "BOBA",
				ChainAliases:  []string{"boba-mainnet"},
				Altruists: []types.Altruist{
					{
						URL:      "https://test-rpc.boba-1.io:1234",
						AuthType: types.ChainAuthTypeBasicAuth,
						Auth:     "test_pocket:auth123456",
					},
				},
				Redirects: []types.GigastakeRedirect{
					{PortalApplicationID: "test_app_6", Alias: "boba-mainnet", Domain: "boba-rpc.gateway.pokt.network"},
				},
				Checks: map[types.ChainCheckType]types.Check{
					types.ChainCheckTypeSync: {
						Type:      types.ChainCheckTypeSync,
						Payload:   `{"id":1,"jsonrpc":"2.0","method":"getSync"}`,
						ResultKey: "sync",
						Allowance: 2,
					},
				},
				CreatedAt: testdata.MockTimestamp,
				UpdatedAt: testdata.MockTimestamp,
			},
		},
		{
			name:             "Should update Chain in cache when receiving an update notification",
			mainTableAction:  types.ActionUpdate,
			sideTablesAction: types.ActionUpdate,
			chain: types.Chain{
				ID:             "0064",
				Blockchain:     "sui-testnet",
				Description:    "Sui Testnet",
				EnforceResult:  "JSON",
				Ticker:         "SUI-TESTNET-123",
				ChainAliases:   []string{"sui-testnet"},
				LogLimitBlocks: 100000,
				RequestTimeout: 60000,
				Active:         false,
				Altruists: []types.Altruist{
					{
						URL:      "https://altruist-0064.com:1234",
						AuthType: types.ChainAuthTypeBasicAuth,
						Auth:     "test_pocket:auth123456",
					},
					{
						URL:      "https://altruist-0064-2.com:1234",
						AuthType: types.ChainAuthTypeBasicAuth,
						Auth:     "test_pocket:auth123456",
					},
				},
				Redirects: []types.GigastakeRedirect{},
				Checks: map[types.ChainCheckType]types.Check{
					types.ChainCheckTypeSync: {
						Type:      types.ChainCheckTypeSync,
						Payload:   `{"id":1,"jsonrpc":"2.0","method":"sui_blockNumber","params":[]}`,
						ResultKey: "result",
						Allowance: 7,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			readerMock := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)

			readerMock.listener.MockEvent(test.mainTableAction, test.sideTablesAction, &test.chain)

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			chain := cache.GetChain(test.chain.ID, Options{})
			c.Equal(test.chain, chain)
		})
	}
}

func Test_parseChainAltruistNotification(t *testing.T) {
	tests := []struct {
		name            string
		sideTableAction types.Action
		chainID         types.RelayChainID
		chain           types.Chain
	}{
		{
			name:            "Should add Altruist to Chain in cache when receiving an insert notification",
			sideTableAction: types.ActionInsert,
			chainID:         "0017",
			chain: types.Chain{
				ID: "0017",
				Altruists: []types.Altruist{
					{
						ChainID:  "0017",
						URL:      "https://test-rpc.neptune.io:1234",
						Auth:     "auth_token",
						AuthType: types.ChainAuthTypeBasicAuth,
					},
				},
			},
		},
		{
			name:            "Should update Altruist in Chain in cache when receiving an update notification",
			sideTableAction: types.ActionUpdate,
			chainID:         "0017",
			chain: types.Chain{
				ID: "0017",
				Altruists: []types.Altruist{
					{
						ChainID:  "0017",
						URL:      "https://test-rpc.neptune.io:1234",
						Auth:     "new_auth_token",
						AuthType: types.ChainAuthTypeBearerToken,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			readerMock := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)

			cache.upsertChain(test.chain)

			readerMock.listener.MockEvent("", test.sideTableAction, &test.chain)

			test.chain.Altruists[0].ChainID = "" // is unset in the cache TODO will be changed soon

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			chain := cache.GetChain(test.chainID, Options{})
			c.Equal(test.chain.Altruists, chain.Altruists)
		})
	}
}

func Test_parseGigastakeRedirectNotification(t *testing.T) {
	tests := []struct {
		name            string
		sideTableAction types.Action
		chainID         types.RelayChainID
		chain           types.Chain
	}{
		{
			name:            "Should add GigastakeRedirect to Chain in cache when receiving an insert notification",
			sideTableAction: types.ActionInsert,
			chainID:         "9826",
			chain: types.Chain{
				ID: "9826",
				Redirects: []types.GigastakeRedirect{
					{
						ChainID:             "9826",
						PortalApplicationID: "test_app_3",
						Domain:              "sol-rpc.gateway.pokt.network",
						Alias:               "solana-mainnet",
					},
				},
			},
		},
		{
			name:            "Should update GigastakeRedirect in Chain in cache when receiving an update notification",
			sideTableAction: types.ActionUpdate,
			chainID:         "9826",
			chain: types.Chain{
				ID: "9826",
				Redirects: []types.GigastakeRedirect{
					{
						ChainID:             "9826",
						PortalApplicationID: "test_app_3",
						Domain:              "sol-rpc.gateway.pokt.network",
						Alias:               "updated-solana-mainnet",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			readerMock := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)

			cache.upsertChain(test.chain)

			test.chain.Redirects[0].ChainID = "" // is unset in the cache TODO will be changed soon

			readerMock.listener.MockEvent("", test.sideTableAction, &test.chain)

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			chain := cache.GetChain(test.chainID, Options{})
			c.Equal(test.chain.Redirects, chain.Redirects)
		})
	}
}

func Test_parseChainCheckNotification(t *testing.T) {
	tests := []struct {
		name            string
		sideTableAction types.Action
		chainID         types.RelayChainID
		checkType       types.ChainCheckType
		chain           types.Chain
	}{
		{
			name:            "Should add Check to Chain in cache when receiving an insert notification",
			sideTableAction: types.ActionInsert,
			chainID:         "7811",
			checkType:       types.ChainCheckTypeChain,
			chain: types.Chain{
				ID: "7811",
				Checks: map[types.ChainCheckType]types.Check{
					types.ChainCheckTypeChain: {
						ChainID:   "7811",
						Type:      types.ChainCheckTypeChain,
						ResultKey: "response",
						Allowance: 1_234,
					},
				},
			},
		},
		{
			name:            "Should update Check in Chain in cache when receiving an update notification",
			sideTableAction: types.ActionUpdate,
			chainID:         "7811",
			checkType:       types.ChainCheckTypeChain,
			chain: types.Chain{
				ID: "7811",
				Checks: map[types.ChainCheckType]types.Check{
					types.ChainCheckTypeChain: {
						ChainID:   "7811",
						Type:      types.ChainCheckTypeChain,
						ResultKey: "good_response",
						Allowance: 5_678,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			readerMock := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)

			cache.upsertChain(test.chain)

			testCheck := test.chain.Checks[test.checkType]
			testCheck.ChainID = "" // is unset in the cache TODO will be changed soon

			readerMock.listener.MockEvent("", test.sideTableAction, &test.chain)

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			chain := cache.GetChain(test.chainID, Options{})
			c.Equal(chain.Checks, chain.Checks)
		})
	}
}
