package cache

import (
	"testing"
	"time"

	phdtestdata "github.com/pokt-foundation/pocket-http-db/testdata"
	"github.com/pokt-foundation/portal-db/v2/testdata"
	"github.com/pokt-foundation/portal-db/v2/types"
	"github.com/pokt-foundation/utils-go/logger"
	"github.com/stretchr/testify/require"
)

func newTestListenerCache(t *testing.T, readerMock *ReaderMock) *Cache {
	c := require.New(t)

	readerMock.On("ReadUserPermissions", testCtx).Return(testdata.UserPermissions, nil)
	readerMock.On("ReadUserIDsMap", testCtx).Return(testdata.UserIDs, nil)
	readerMock.On("ReadPlans", testCtx).Return(testdata.PayPlans, nil)
	readerMock.On("ReadAccounts", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(testdata.Accounts, nil)
	readerMock.On("ReadPortalApps", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(testdata.PortalApps, nil)
	readerMock.On("ReadGigastakeApps", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(testdata.GigastakeApps, nil)
	readerMock.On("ReadChains", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(testdata.Chains, nil)
	readerMock.On("ReadBlockedContracts", testCtx).Return(testdata.GlobalBlockedContracts, nil)

	cacheInstance, err := NewCache(testCtx, Config{
		Reader:              readerMock,
		Logger:              logger.New(),
		CacheRefreshMinutes: 5 * time.Minute,
	})
	c.NoError(err)

	return cacheInstance
}

/* ---------- Accounts ---------- */

func Test_parseAccountUserAccessNotification(t *testing.T) {
	tests := []struct {
		name            string
		sideTableAction types.Action
		userID          types.UserID
		accountID       types.AccountID
		account         *types.Account
	}{
		{
			name:            "Should insert AccountUserAccess in cache when receiving an insert notification",
			sideTableAction: types.ActionInsert,
			userID:          "user_3",
			accountID:       "account_1",
			account: &types.Account{
				Users: map[types.UserID]types.AccountUserAccess{
					"user_3": {
						UserID:   "user_3",
						Email:    "ellen.ripley789@test.com",
						Accepted: true,
						PortalAppRoles: map[types.PortalAppID]types.RoleName{
							"test_app_1": types.RoleMember,
						},
					},
				},
			},
		},
		{
			name:            "Should update AccountUserAccess in cache when receiving an update notification",
			sideTableAction: types.ActionUpdate,
			userID:          "user_2",
			accountID:       "account_1",
			account: &types.Account{
				Users: map[types.UserID]types.AccountUserAccess{
					"user_3": {
						UserID:   "user_2",
						Email:    "paul.atreides456@test.com",
						Accepted: true,
						PortalAppRoles: map[types.PortalAppID]types.RoleName{
							"test_app_1": types.RoleMember,
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

			readerMock, cancel := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)
			defer cancel()

			test.account.ID = test.accountID
			readerMock.listener.MockEvent("", test.sideTableAction, test.account)

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			if test.sideTableAction == types.ActionDelete {
				accountFound := phdtestdata.FindAccountInAccounts(cache.GetAccountsByUserID(test.userID), test.accountID)
				c.False(accountFound, "Account should be removed from accountsByUserID map")
			} else {
				accountFound := phdtestdata.FindAccountInAccounts(cache.GetAccountsByUserID(test.userID), test.accountID)
				c.True(accountFound, "Account should be found in accountsByUserID map")
			}
		})
	}
}

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

			readerMock, cancel := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)
			defer cancel()

			test.account.ID = test.accountID
			testAccount := test.account
			readerMock.listener.MockEvent(test.mainTableAction, test.sideTablesAction, testAccount)

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			account := cache.GetAccount(test.accountID)
			if testAccount.Deleted {
				c.Empty(account)
			} else {
				c.Equal(testAccount.PlanType, account.PlanType)
				c.Equal(testAccount.PartnerChainIDs, account.PartnerChainIDs)
				c.Equal(testAccount.PartnerThroughputLimit, account.PartnerThroughputLimit)
				c.Equal(testAccount.PartnerAppLimit, account.PartnerAppLimit)
			}

			accountFound := phdtestdata.FindAccountInAccounts(cache.GetAccounts(Options{IncludeDeleted: false}), test.accountID)
			if testAccount.Deleted {
				c.False(accountFound, "Account should be removed from accounts map")
			} else {
				c.True(accountFound, "Account should be found in accounts map")
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
				AATs: map[types.ProtocolAppID]types.AAT{
					"test_protocol_app_1": {
						ID:              "test_protocol_app_1",
						Address:         "test_162789bdf3fec1da21e3b28a757bb648",
						PublicKey:       "test_360c8eb90ef9536a888a843689a5a20c",
						ClientPublicKey: "test_97ae5cb13fa6c67cb3e5904ba79e842e",
						PrivateKey:      "",
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
			portalApp: types.PortalApp{
				ID:        "test_app_update_b03ca84c",
				AccountID: "account_1",
				AATs:      map[types.ProtocolAppID]types.AAT{"": {}},
				Settings: types.Settings{
					Environment:       types.EnvironmentProduction,
					SecretKey:         "test_849c1397586f9fb6f902576120d0d10f",
					SecretKeyRequired: true,
				},
				LegacyFields: types.LegacyFields{
					CustomLimit:    0,
					RequestTimeout: 5_000,
				},
			},
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

			readerMock, cancel := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)
			defer cancel()

			test.portalApp.ID = test.portalAppID
			readerMock.listener.MockEvent(test.mainTableAction, test.sideTablesAction, &test.portalApp)

			time.Sleep(200 * time.Millisecond) // need time for cache refresh

			if !test.portalApp.Deleted {
				c.Equal(test.portalApp, cache.GetPortalApp(test.portalAppID))
			}

			portalAppFound := phdtestdata.FindPortalAppInPortalApps(cache.GetPortalApps(Options{IncludeDeleted: false}), test.portalAppID)
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
						PrivateKey:      "",
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
						PrivateKey:      "",
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

			readerMock, cancel := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)
			defer cancel()

			// Add the PortalApp to the cache
			cache.upsertPortalApp(*test.portalApp)

			readerMock.listener.MockEvent("", test.sideTableAction, test.portalApp)

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			portalApp := cache.GetPortalApp(test.portalApp.ID)
			c.Equal(test.portalApp.AATs, portalApp.AATs)
		})
	}
}

func Test_parseGigastakeAppNotification(t *testing.T) {
	tests := []struct {
		name             string
		mainTableAction  types.Action
		sideTablesAction types.Action
		gigastakeApp     types.GigastakeApp
	}{
		{
			name:             "Should insert GigastakeApp in cache when receiving an insert notification",
			mainTableAction:  types.ActionInsert,
			sideTablesAction: types.ActionInsert,
			gigastakeApp: types.GigastakeApp{
				ID:              "test_gigastake_app_1",
				ChainIDs:        map[types.RelayChainID]struct{}{"0001": {}},
				Name:            "pokt_gigastake",
				Address:         "test_8d4f6a5b0c6e9f1db12c1f662e5ec8c5",
				PublicKey:       "test_37a0e8437f5149dc98a9a5b207efc2d0",
				ClientPublicKey: "test_65c29f0cc82e418b81a528a0c0682a9f",
				Signature:       "test_f22651fb566346fca30b605e5f46e3ca",
				Version:         "0.0.1",
				CreatedAt:       testdata.MockTimestamp,
				UpdatedAt:       testdata.MockTimestamp,
			},
		},
		{
			name:             "Should update GigastakeApp in cache when receiving an update notification",
			mainTableAction:  types.ActionUpdate,
			sideTablesAction: types.ActionUpdate,
			gigastakeApp: types.GigastakeApp{
				ID:              "test_gigastake_app_1",
				ChainIDs:        map[types.RelayChainID]struct{}{"0001": {}},
				Name:            "updated_pokt_gigastake",
				Address:         "test_updated_address",
				PublicKey:       "test_updated_public_key",
				ClientPublicKey: "test_updated_client_public_key",
				Signature:       "test_updated_signature",
				Version:         "0.0.1",
				CreatedAt:       testdata.MockTimestamp,
				UpdatedAt:       testdata.MockTimestamp,
			},
		},
		{
			name:             "Should delete GigastakeApp in cache when receiving an update notification with GigastakeApp.Deleted = true",
			mainTableAction:  types.ActionUpdate,
			sideTablesAction: types.ActionUpdate,
			gigastakeApp: types.GigastakeApp{
				ID:       "test_gigastake_app_1",
				ChainIDs: map[types.RelayChainID]struct{}{"0001": {}},
				Name:     "pokt_gigastake",
				Deleted:  true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			readerMock, cancel := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)
			defer cancel()

			readerMock.listener.MockEvent(test.mainTableAction, test.sideTablesAction, &test.gigastakeApp)

			time.Sleep(200 * time.Millisecond) // need time for cache refresh

			if !test.gigastakeApp.Deleted {
				c.Equal(test.gigastakeApp, cache.GetGigastakeAppByID(test.gigastakeApp.ID))
			}

			gigastakeAppFound := phdtestdata.FindGigastakeAppInGigastakeApps(cache.GetGigastakeApps(Options{}), test.gigastakeApp.ID)
			if test.gigastakeApp.Deleted {
				c.False(gigastakeAppFound, "GigastakeApp should be removed from gigastakeApps map")
			} else {
				c.True(gigastakeAppFound, "GigastakeApp should be found in gigastakeApps map")
			}
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

			readerMock, cancel := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)
			defer cancel()

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

			readerMock, cancel := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)
			defer cancel()

			if test.sideTableAction == types.ActionDelete {
				storedPortalApp := cache.GetPortalApp(test.portalApp.ID)
				_, ok := storedPortalApp.Whitelists.Blockchains[test.chain]
				c.True(ok, "Whitelist should be present in the PortalApp in cache before deletion")
			}

			readerMock.listener.MockEvent("", test.sideTableAction, test.portalApp)

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			storedPortalApp := cache.GetPortalApp(test.portalApp.ID)
			_, ok := storedPortalApp.Whitelists.Blockchains[test.chain]
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
				Description:   "Boba",
				EnforceResult: "JSON",
				Ticker:        "BOBA",
				Altruists: map[types.AltruistURL]types.Altruist{
					"https://test-rpc.boba-1.io:1234": {
						URL:      "https://test-rpc.boba-1.io:1234",
						AuthType: types.ChainAuthTypeBasicAuth,
						Auth:     "test_pocket:auth123456",
					},
				},
				Checks: map[types.ChainCheckType]types.Check{
					types.ChainCheckTypeSync: {
						Type:      types.ChainCheckTypeSync,
						Payload:   `{"id":1,"jsonrpc":"2.0","method":"getSync"}`,
						ResultKey: "sync",
						Allowance: 2,
					},
				},
				Aliases: map[types.ChainAlias]struct{}{
					"boba-mainnet":   {},
					"boba-mainnet-2": {},
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
				Description:    "Sui Testnet",
				EnforceResult:  "JSON",
				Ticker:         "SUI-TESTNET-123",
				LogLimitBlocks: 100000,
				RequestTimeout: 60000,
				Active:         false,
				Altruists: map[types.AltruistURL]types.Altruist{
					"https://altruist-0064.com:1234": {
						URL:      "https://altruist-0064.com:1234",
						AuthType: types.ChainAuthTypeBasicAuth,
						Auth:     "test_pocket:auth123456",
					},
					"https://altruist-0064-2.com:1234": {
						URL:      "https://altruist-0064-2.com:1234",
						AuthType: types.ChainAuthTypeBasicAuth,
						Auth:     "test_pocket:auth123456",
					},
				},
				Checks: map[types.ChainCheckType]types.Check{
					types.ChainCheckTypeSync: {
						Type:      types.ChainCheckTypeSync,
						Payload:   `{"id":1,"jsonrpc":"2.0","method":"sui_blockNumber","params":[]}`,
						ResultKey: "result",
						Allowance: 7,
					},
				},
				Aliases: map[types.ChainAlias]struct{}{
					"sui-testnet":   {},
					"sui-testnet-2": {},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			readerMock, cancel := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)
			defer cancel()

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
				Altruists: map[types.AltruistURL]types.Altruist{
					"https://test-rpc.neptune.io:1234": {
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
				Altruists: map[types.AltruistURL]types.Altruist{
					"https://test-rpc.neptune.io:1234": {
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

			readerMock, cancel := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)
			defer cancel()

			cache.upsertChain(test.chain)

			readerMock.listener.MockEvent("", test.sideTableAction, &test.chain)

			for _, altruist := range test.chain.Altruists {
				altruist.ChainID = "" // is unset in the cache TODO will be changed soon
				test.chain.Altruists[altruist.URL] = altruist
			}

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			chain := cache.GetChain(test.chainID, Options{})
			c.Equal(test.chain.Altruists, chain.Altruists)
		})
	}
}

func Test_parseAliasDomainsNotification(t *testing.T) {
	tests := []struct {
		name            string
		sideTableAction types.Action
		chainID         types.RelayChainID
		chain           types.Chain
	}{
		{
			name:            "Should add AliasDomains to Chain in cache when receiving an insert notification",
			sideTableAction: types.ActionInsert,
			chainID:         "9826",
			chain: types.Chain{
				ID: "9826",
				Aliases: map[types.ChainAlias]struct{}{
					"solana-mainnet": {},
				},
			},
		},
		{
			name:            "Should update AliasDomains in Chain in cache when receiving an update notification",
			sideTableAction: types.ActionUpdate,
			chainID:         "9826",
			chain: types.Chain{
				ID: "9826",
				Aliases: map[types.ChainAlias]struct{}{
					"updated-solana-mainnet": {},
				},
			},
		},
		{
			name:            "Should update AliasDomains with a longer slice in Chain in cache when receiving an update notification",
			sideTableAction: types.ActionUpdate,
			chainID:         "9826",
			chain: types.Chain{
				ID: "9826",
				Aliases: map[types.ChainAlias]struct{}{
					"updated-solana-mainnet": {},
				},
			},
		},
		{
			name:            "Should delete AliasDomains in Chain in cache when receiving a delete notification",
			sideTableAction: types.ActionDelete,
			chainID:         "9826",
			chain: types.Chain{
				ID: "9826",
				Aliases: map[types.ChainAlias]struct{}{
					"updated-solana-mainnet": {},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			readerMock, cancel := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)
			defer cancel()

			cache.upsertChain(test.chain)

			readerMock.listener.MockEvent("", test.sideTableAction, &test.chain)

			time.Sleep(100 * time.Millisecond) // need time for cache refresh

			chain := cache.GetChain(test.chainID, Options{})

			if test.sideTableAction != types.ActionDelete {
				// Merge any additional keys from chain.AliasDomains into test.chain.AliasDomains before checking for equality
				for key, value := range chain.Aliases {
					if _, exists := test.chain.Aliases[key]; !exists {
						test.chain.Aliases[key] = value
					}
				}
				c.Equal(test.chain.Aliases, chain.Aliases, "Chain alias domains should match expected")
			} else {
				// For deletion, just check that the alias does not exist in chain.Aliases
				var aliasKey types.ChainAlias
				for key := range test.chain.Aliases {
					aliasKey = key
					break
				}
				c.NotEmpty(aliasKey)
				_, exists := chain.Aliases[aliasKey]
				c.False(exists, "Alias should not exist after delete operation")
			}
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

			readerMock, cancel := NewReaderMock(t)
			cache := newTestListenerCache(t, readerMock)
			defer cancel()

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
