package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pokt-foundation/portal-db/v2/testdata"
	"github.com/pokt-foundation/portal-db/v2/types"
)

var testCtx = context.Background()

func NewTestCache(t *testing.T) *Cache {
	c := require.New(t)

	readerMock := NewReaderMock(t)

	readerMock.On("ReadUserPermissions", testCtx).Return(testdata.UserPermissions, nil)
	readerMock.On("ReadUserIDsMap", testCtx).Return(testdata.UserIDs, nil)
	readerMock.On("ReadPlans", testCtx).Return(testdata.PayPlans, nil)
	readerMock.On("ReadAccounts", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(testdata.Accounts, nil)
	readerMock.On("ReadPortalApps", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(testdata.PortalApps, nil)
	readerMock.On("ReadChains", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(testdata.Chains, nil)
	readerMock.On("ReadBlockedContracts", testCtx).Return(testdata.GlobalBlockedContracts, nil)
	readerMock.On("ReadUserByUserID", mock.Anything, types.UserID("usertest")).Return(testdata.Users[types.UserID("usertest")], nil)

	cacheInstance, err := NewCache(testCtx, Config{
		Reader:              readerMock,
		Logger:              logrus.New(),
		CacheRefreshMinutes: 5 * time.Minute,
	})
	c.NoError(err)

	return cacheInstance
}

/* ---------- Cache Set and Read ---------- */

func Test_SetCache(t *testing.T) {
	cache := NewTestCache(t)

	tests := []struct {
		name             string
		actual, expected interface{}
	}{
		{
			name:     "Should return user permissions by user ID",
			actual:   cache.GetUserPermissionsByUserID("user_1"),
			expected: *testdata.UserPermissions["user_1"],
		},
		{
			name:     "Should return a user ID for a given provider user ID",
			actual:   cache.GetUserIDFromProviderUserID("auth0|chrisjen_avasarala"),
			expected: types.UserID("user_5"),
		},
		{
			name:     "Should return all accounts",
			actual:   accountsToMap(cache.GetAccounts(Options{})),
			expected: testdata.Accounts,
		},
		{
			name:     "Should return account by ID",
			actual:   cache.GetAccount("account_1"),
			expected: *testdata.Accounts["account_1"],
		},
		{
			name:   "Should return accounts by user ID",
			actual: accountsToMap(cache.GetAccountsByUserID("user_4")),
			expected: map[types.AccountID]*types.Account{
				"account_2": testdata.Accounts["account_2"],
				"account_4": testdata.Accounts["account_4"],
				"account_5": testdata.Accounts["account_5"],
			},
		},
		{
			name:     "Should return portal app by ID",
			actual:   cache.GetPortalApp("test_app_1"),
			expected: *testdata.PortalApps["test_app_1"],
		},
		{
			name:     "Should return portal app by protocol ID",
			actual:   cache.GetPortalAppByProtocolID("test_protocol_app_1"),
			expected: *testdata.PortalApps["test_app_1"],
		},
		{
			name:   "Should return portal app by protocol ID with multiple AATs",
			actual: cache.GetPortalAppByProtocolID("test_protocol_app_4"),
			expected: func() types.PortalApp {
				originalPortalApp := testdata.PortalApps["test_app_3"]
				expectedPortalApp := *originalPortalApp
				expectedPortalApp.AATs = map[types.ProtocolAppID]types.AAT{
					"test_protocol_app_4": {
						ID:              "test_protocol_app_4",
						Address:         "test_eb2e5bcba557cfe8fa76fd7fff54f9d1",
						PublicKey:       "test_f6a5d8690ecb669865bd752b7796a920",
						ClientPublicKey: "test_6ee5ea553408f0895923fd1569dc5072",
						PrivateKey:      "test_838d29d61a65401f7d56d084cb6e4783",
						Signature:       "test_cf05cf9bb26111c548e88fb6157af708",
						Version:         "0.0.1",
					},
				}
				return expectedPortalApp
			}(),
		},
		{
			name:     "Should return all portal apps",
			actual:   portalAppsToMap(cache.GetPortalApps(Options{})),
			expected: testdata.PortalApps,
		},
		{
			name:     "Should return chain by ID",
			actual:   cache.GetChain("0001", Options{}),
			expected: *testdata.Chains["0001"],
		},
		{
			name:     "Should return all chains",
			actual:   chainsToMap(cache.GetChains(Options{})),
			expected: testdata.Chains,
		},
		{
			name:     "Should return chain by ID with GigastakeRedirectAccount set",
			actual:   cache.GetChain("0001", Options{SetGigastakeRedirectApp: true}),
			expected: setGigastakeRedirectAccounts(*testdata.Chains["0001"], cache),
		},
		{
			name:     "Should return all chains with GigastakeRedirectAccount set",
			actual:   chainsToMap(cache.GetChains(Options{SetGigastakeRedirectApp: true})),
			expected: setGigastakeRedirectAccountsForChains(testdata.Chains, cache),
		},
		{
			name:     "Should return plan by type",
			actual:   cache.GetPlan("basic_plan"),
			expected: *testdata.PayPlans["basic_plan"],
		},
		{
			name:     "Should return all pay plans",
			actual:   payPlansToMap(cache.GetPlans()),
			expected: testdata.PayPlans,
		},
		{
			name:     "Should return all global blocked contracts",
			actual:   cache.GetBlockedContracts(),
			expected: testdata.GlobalBlockedContracts,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			c.Equal(test.expected, test.actual)
		})
	}
}

func TestCache_SetCache_Failure(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name         string
		failedMethod string
		err          error
	}{
		{
			name:         "should fail if ReadUserPermissions fails",
			failedMethod: "ReadUserPermissions",
			err:          errors.New("error on user permissions"),
		},
		{
			name:         "should fail if ReadUserIDsMap fails",
			failedMethod: "ReadUserIDsMap",
			err:          errors.New("error on user IDs"),
		},
		{
			name:         "should fail if ReadPlans fails",
			failedMethod: "ReadPlans",
			err:          errors.New("error on pay plans"),
		},
		{
			name:         "should fail if ReadAccounts fails",
			failedMethod: "ReadAccounts",
			err:          errors.New("error on Accounts"),
		},
		{
			name:         "should fail if ReadPortalApps fails",
			failedMethod: "ReadPortalApps",
			err:          errors.New("error on PortalApps"),
		},
		{
			name:         "should fail if ReadChains fails",
			failedMethod: "ReadChains",
			err:          errors.New("error on Chains"),
		},
		{
			name:         "should fail if ReadBlockedContracts fails",
			failedMethod: "ReadBlockedContracts",
			err:          errors.New("error on blocked contracts"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			readerMock := NewReaderMock(t)

			switch test.failedMethod {

			case "ReadUserPermissions":
				readerMock.On(test.failedMethod, testCtx).Return(nil, test.err)

			case "ReadUserIDsMap":
				readerMock.On("ReadUserPermissions", testCtx).Return(map[types.UserID]*types.UserPermissions{}, nil)
				readerMock.On(test.failedMethod, testCtx).Return(nil, test.err)

			case "ReadPlans":
				readerMock.On("ReadUserPermissions", testCtx).Return(map[types.UserID]*types.UserPermissions{}, nil)
				readerMock.On("ReadUserIDsMap", testCtx).Return(map[types.ProviderUserID]types.UserID{}, nil)
				readerMock.On(test.failedMethod, testCtx).Return(nil, test.err)

			case "ReadAccounts":
				readerMock.On("ReadUserPermissions", testCtx).Return(map[types.UserID]*types.UserPermissions{}, nil)
				readerMock.On("ReadUserIDsMap", testCtx).Return(map[types.ProviderUserID]types.UserID{}, nil)
				readerMock.On("ReadPlans", testCtx).Return(map[types.PayPlanType]*types.Plan{}, nil)
				readerMock.On(test.failedMethod, testCtx, types.DriverOptions{IncludeDeleted: true}).Return(nil, test.err)

			case "ReadPortalApps":
				readerMock.On("ReadUserPermissions", testCtx).Return(map[types.UserID]*types.UserPermissions{}, nil)
				readerMock.On("ReadUserIDsMap", testCtx).Return(map[types.ProviderUserID]types.UserID{}, nil)
				readerMock.On("ReadPlans", testCtx).Return(map[types.PayPlanType]*types.Plan{}, nil)
				readerMock.On("ReadAccounts", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(map[types.AccountID]*types.Account{}, nil)
				readerMock.On(test.failedMethod, testCtx, types.DriverOptions{IncludeDeleted: true}).Return(nil, test.err)

			case "ReadChains":
				readerMock.On("ReadUserPermissions", testCtx).Return(map[types.UserID]*types.UserPermissions{}, nil)
				readerMock.On("ReadUserIDsMap", testCtx).Return(map[types.ProviderUserID]types.UserID{}, nil)
				readerMock.On("ReadPlans", testCtx).Return(map[types.PayPlanType]*types.Plan{}, nil)
				readerMock.On("ReadAccounts", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(map[types.AccountID]*types.Account{}, nil)
				readerMock.On("ReadPortalApps", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(map[types.PortalAppID]*types.PortalApp{}, nil)
				readerMock.On(test.failedMethod, testCtx, types.DriverOptions{IncludeDeleted: true}).Return(nil, test.err)

			case "ReadBlockedContracts":
				readerMock.On("ReadUserPermissions", testCtx).Return(map[types.UserID]*types.UserPermissions{}, nil)
				readerMock.On("ReadUserIDsMap", testCtx).Return(map[types.ProviderUserID]types.UserID{}, nil)
				readerMock.On("ReadPlans", testCtx).Return(map[types.PayPlanType]*types.Plan{}, nil)
				readerMock.On("ReadAccounts", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(map[types.AccountID]*types.Account{}, nil)
				readerMock.On("ReadPortalApps", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(map[types.PortalAppID]*types.PortalApp{}, nil)
				readerMock.On("ReadChains", testCtx, types.DriverOptions{IncludeDeleted: true}).Return(map[types.RelayChainID]*types.Chain{}, nil)
				readerMock.On(test.failedMethod, testCtx).Return(types.GlobalBlockedContracts{}, test.err)

			}

			_, err := NewCache(testCtx, Config{
				Reader:              readerMock,
				Logger:              logrus.New(),
				CacheRefreshMinutes: 5 * time.Minute,
			})
			c.ErrorIs(err, test.err)
		})
	}
}

/* ---------- Accounts ---------- */

func Test_upsertAccount(t *testing.T) {
	tests := []struct {
		name         string
		account      types.Account
		accountUsers AccountsForUser
		pendingUsers map[types.AccountID][]types.AccountUserAccess
	}{
		{
			name:    "Should insert a new account",
			account: *testdata.TestCreateAccount,
		},
		{
			name: "Should update an existing account",
			account: types.Account{
				ID:       "account_1",
				PlanType: types.PayPlanType("enterprise_plan"),
				Users: map[types.UserID]types.AccountUserAccess{
					"user_1": testdata.AccountUserAccess[1],
					"user_2": testdata.AccountUserAccess[2],
					"user_8": testdata.AccountUserAccess[8],
				},
				PortalApps: map[types.PortalAppID]*types.PortalApp{
					"test_app_1": testdata.PortalApps["test_app_1"],
				},
				PartnerChainIDs:        map[types.RelayChainID]struct{}{"0001": {}, "0053": {}},
				PartnerThroughputLimit: 2_000,
				PartnerAppLimit:        1,
				CreatedAt:              testdata.MockTimestamp,
				UpdatedAt:              testdata.MockTimestamp,
			},
		},
		{
			name: "Should update an existing account when pending users found in cache",
			account: types.Account{
				ID:       "account_1",
				PlanType: types.PayPlanType("enterprise_plan"),
				Users: map[types.UserID]types.AccountUserAccess{
					"user_1": testdata.AccountUserAccess[1],
					"user_2": testdata.AccountUserAccess[2],
					"user_8": testdata.AccountUserAccess[8],
					"user_3": testdata.AccountUserAccess[3],
					"user_7": testdata.AccountUserAccess[7],
				},
				PortalApps: map[types.PortalAppID]*types.PortalApp{
					"test_app_1": testdata.PortalApps["test_app_1"],
				},
				PartnerChainIDs:        map[types.RelayChainID]struct{}{"0001": {}, "0053": {}},
				PartnerThroughputLimit: 2_000,
				PartnerAppLimit:        1,
				CreatedAt:              testdata.MockTimestamp,
				UpdatedAt:              testdata.MockTimestamp,
			},
			pendingUsers: map[types.AccountID][]types.AccountUserAccess{
				"account_1": {testdata.AccountUserAccess[7], testdata.AccountUserAccess[3]},
			},
		},
		{
			name: "Should update deleted account and remove from user ID accounts map",
			account: types.Account{
				ID:       "account_2",
				PlanType: types.PayPlanType("pro_plan"),
				Users: map[types.UserID]types.AccountUserAccess{
					"user_3": testdata.AccountUserAccess[3],
					"user_4": testdata.AccountUserAccess[4],
					"user_9": testdata.AccountUserAccess[9],
					"user_2": testdata.AccountUserAccess[10],
				},
				PortalApps: map[types.PortalAppID]*types.PortalApp{
					"test_app_2": testdata.PortalApps["test_app_2"],
				},
				Deleted: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			accountID := test.account.ID

			cache := NewTestCache(t)

			if len(test.pendingUsers) > 0 {
				cache.pendingAccountUserAccess = test.pendingUsers
			}

			// Upsert account
			cache.upsertAccount(test.account)

			if len(test.pendingUsers) > 0 {
				for _, user := range test.pendingUsers[test.account.ID] {
					test.account.Users[user.UserID] = user
				}
			}

			test.account.Plan = cache.plans[test.account.PlanType]
			c.Equal(test.account, cache.GetAccount(test.account.ID))

			accountFound := findAccountInAccounts(cache.GetAccounts(Options{IncludeDeleted: false}), accountID)
			if test.account.Deleted {
				c.False(accountFound, "Account should be removed from accounts map")
			} else {
				c.True(accountFound, "Account should be found in accounts map")
			}

			for _, user := range test.account.Users {
				accountsForUser := cache.GetAccountsByUserID(user.UserID)

				accountInUserAccounts := findAccountInAccounts(accountsForUser, accountID)

				if test.account.Deleted {
					c.False(accountInUserAccounts, "Account should be removed from accountsByUserID map")
				} else {
					c.True(accountInUserAccounts, "Account should be present in accountsByUserID map")
					c.Equal(test.account, getAccountFromAccounts(accountsForUser, accountID), "Account should be present in the cache and updated")
				}
			}
		})
	}
}

func Test_upsertAccountUser(t *testing.T) {
	tests := []struct {
		name        string
		accountID   types.AccountID
		accountUser types.AccountUserAccess
		updateRole  types.RoleName
	}{
		{
			name:        "Should add a new Accepted user to account and user permissions",
			accountID:   "account_2",
			accountUser: testdata.AccountUserAccess[1],
		},
		{
			name:        "Should add a new not-Accepted user to account",
			accountID:   "account_4",
			accountUser: testdata.AccountUserAccess[12],
		},
		{
			name:        "Should update an existing account User's role",
			accountID:   "account_1",
			accountUser: testdata.AccountUserAccess[2],
			updateRole:  types.RoleMember,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			if test.updateRole != "" {
				test.accountUser.RoleName = test.updateRole
			}

			test.accountUser.AccountID = test.accountID // will be set in the notification

			// Upsert account user
			cache.upsertAccountUser(test.accountUser)

			account := cache.GetAccount(test.accountUser.AccountID)
			test.accountUser.AccountID = "" // is unset in the cache TODO will be changed soon
			accountUser, ok := account.Users[test.accountUser.UserID]

			c.True(ok, "AccountUserAccess should be present in account Users map")
			c.Equal(test.accountUser, accountUser, "AccountUserAccess should be added or updated in the cache")

			accountFound := findAccountInAccounts(cache.GetAccountsByUserID(test.accountUser.UserID), test.accountID)
			c.True(accountFound, "Account should be present in accountsByUserID map")

			if test.accountUser.Accepted {
				userPermissions := cache.GetUserPermissionsByUserID(test.accountUser.UserID)
				accountPermission, ok := userPermissions.Accounts[test.accountID]

				c.True(ok, "AccountPermissions should be present in UserPermissions")
				c.Equal(test.accountUser.RoleName, accountPermission.RoleName, "AccountUserAccess role should be updated in UserPermissions")
			}
		})
	}
}

func Test_deleteAccountUser(t *testing.T) {
	tests := []struct {
		name            string
		accountID       types.AccountID
		accountUser     types.AccountUserAccess
		existingAccount bool
	}{
		{
			name:            "Should delete an existing user access and update permissions",
			accountID:       "account_1",
			accountUser:     testdata.AccountUserAccess[2],
			existingAccount: true,
		},
		{
			name:        "Should not fail when deleting a non-existing user access",
			accountID:   "account_5",
			accountUser: testdata.AccountUserAccess[12],
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			if test.existingAccount {
				_, ok := cache.GetAccount(test.accountID).Users[test.accountUser.UserID]
				c.True(ok, "AccountUserAccess should be found in account Users map before deletion")

				accountFound := findAccountInAccounts(cache.GetAccountsByUserID(test.accountUser.UserID), test.accountID)
				c.True(accountFound, "Account should be found in accountsByUserID map before deletion")

				userPermissions := cache.GetUserPermissionsByUserID(test.accountUser.UserID)
				_, ok = userPermissions.Accounts[test.accountID]
				c.True(ok, "UserPermissions should be found in UserPermissions before deletion")
			}

			test.accountUser.AccountID = test.accountID // will be set in the notification

			// Delete account user
			cache.deleteAccountUser(test.accountUser)

			account := cache.GetAccount(test.accountID)
			test.accountUser.AccountID = "" // is unset in the cache TODO will be changed soon

			_, ok := account.Users[test.accountUser.UserID]
			c.False(ok, "AccountUserAccess should be deleted from account Users map")

			accountFound := findAccountInAccounts(cache.GetAccountsByUserID(test.accountUser.UserID), test.accountID)
			c.False(accountFound, "Account should be removed from accountsByUserID map")

			userPermissions := cache.GetUserPermissionsByUserID(test.accountUser.UserID)
			_, ok = userPermissions.Accounts[test.accountID]
			c.False(ok, "UserPermissions should be removed from UserPermissions")
		})
	}
}

/* ---------- PortalApps ---------- */

func Test_upsertPortalApp(t *testing.T) {
	testApp := *testdata.TestCreatePortalApp
	testBaseApp := types.PortalApp{
		ID:        testApp.ID,
		AccountID: testApp.AccountID,
		Name:      testApp.Name,
		Gigastake: testApp.Gigastake,
		Staked:    testApp.Staked,
	}

	tests := []struct {
		name                    string
		portalApp               types.PortalApp
		pendingAAT              map[types.PortalAppID]types.AAT
		pendingSettings         map[types.PortalAppID]types.Settings
		pendingAppNotifications map[types.PortalAppID]types.AppNotification
		pendingWhitelists       map[types.PortalAppID][]types.Whitelist
		update, delete          bool
	}{
		{
			name:      "Should add a new PortalApp to cache with no pending side table data",
			portalApp: testBaseApp,
		},
		{
			name:      "Should add a new PortalApp to cache with pending AAT",
			portalApp: testBaseApp,
			pendingAAT: map[types.PortalAppID]types.AAT{
				testBaseApp.ID: testdata.TestCreatePortalAppAAT,
			},
		},
		{
			name:      "Should add a new PortalApp to cache with pending Settings",
			portalApp: testBaseApp,
			pendingSettings: map[types.PortalAppID]types.Settings{
				testBaseApp.ID: testApp.Settings,
			},
			pendingAAT: map[types.PortalAppID]types.AAT{
				testBaseApp.ID: testdata.TestCreatePortalAppAAT,
			},
		},
		{
			name:      "Should add a new PortalApp to cache with pending AppNotifications",
			portalApp: testBaseApp,
			pendingAppNotifications: map[types.PortalAppID]types.AppNotification{
				testBaseApp.ID: {
					Type:        types.NotificationTypeEmail,
					Active:      true,
					Destination: "email@pokt.network",
					Trigger:     "trigger456",
					Events:      map[types.NotificationEvent]bool{types.NotificationEventHalf: true},
				},
			},
			pendingSettings: map[types.PortalAppID]types.Settings{
				testBaseApp.ID: testApp.Settings,
			},
			pendingAAT: map[types.PortalAppID]types.AAT{
				testBaseApp.ID: testdata.TestCreatePortalAppAAT,
			},
		},
		{
			name:      "Should add a new PortalApp to cache with pending Whitelists",
			portalApp: testBaseApp,
			pendingSettings: map[types.PortalAppID]types.Settings{
				testBaseApp.ID: testApp.Settings,
			},
			pendingWhitelists: map[types.PortalAppID][]types.Whitelist{
				testBaseApp.ID: {
					{
						AppID: testBaseApp.ID,
						Type:  types.WhitelistTypeOrigins,
						Value: "https://example.com",
					},
					{
						AppID:   testBaseApp.ID,
						Type:    types.WhitelistTypeContracts,
						Value:   "0x0987654321abcdef",
						ChainID: "0064",
					},
				},
			},
			pendingAppNotifications: map[types.PortalAppID]types.AppNotification{
				testBaseApp.ID: {
					Type:        types.NotificationTypeEmail,
					Active:      true,
					Destination: "email@pokt.network",
					Trigger:     "trigger456",
					Events:      map[types.NotificationEvent]bool{types.NotificationEventHalf: true},
				},
			},
			pendingAAT: map[types.PortalAppID]types.AAT{
				testBaseApp.ID: testdata.TestCreatePortalAppAAT,
			},
		},
		{
			name:       "Should update an existing PortalApp in cache with all side tables",
			portalApp:  *testdata.PortalApps["test_app_1"],
			pendingAAT: nil, // AAT is never updated
			pendingSettings: map[types.PortalAppID]types.Settings{
				"test_app_1": testApp.Settings,
			},
			pendingAppNotifications: map[types.PortalAppID]types.AppNotification{
				"test_app_1": {
					Type:        types.NotificationTypeEmail,
					Active:      true,
					Destination: "user@example.network",
					Trigger:     "12345",
					Events:      map[types.NotificationEvent]bool{types.NotificationEventHalf: true},
				},
			},
			pendingWhitelists: map[types.PortalAppID][]types.Whitelist{
				"test_app_1": {
					{
						AppID: testBaseApp.ID,
						Type:  types.WhitelistTypeOrigins,
						Value: "https://portalgun.io",
					},
					{
						AppID:   testBaseApp.ID,
						Type:    types.WhitelistTypeContracts,
						Value:   "0xtest_2f78db6436527729929aaf6c616361de0f7",
						ChainID: "0001",
					},
				},
			},
		},
		{
			name:      "Should delete an existing PortalApp in cache",
			portalApp: *testdata.PortalApps["test_app_3"],
			delete:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			if test.update || test.delete {
				storedPortalApp := cache.GetPortalApp(test.portalApp.ID)
				c.Equal(test.portalApp, storedPortalApp, "PortalApp should be present in cache before update")
			}
			if test.delete {
				test.portalApp.Deleted = true
			}

			// Update the expected PortalApp with the pending data
			expectedPortalApp := test.portalApp
			if aat, ok := test.pendingAAT[test.portalApp.ID]; ok {
				expectedPortalApp.AATs = map[types.ProtocolAppID]types.AAT{
					aat.ID: aat,
				}
			}
			if settings, ok := test.pendingSettings[test.portalApp.ID]; ok {
				expectedPortalApp.Settings = settings
			}
			if appNotification, ok := test.pendingAppNotifications[test.portalApp.ID]; ok {
				if expectedPortalApp.Notifications == nil {
					expectedPortalApp.Notifications = make(map[types.NotificationType]types.AppNotification)
				}
				expectedPortalApp.Notifications[appNotification.Type] = appNotification
			}
			if whitelists, ok := test.pendingWhitelists[test.portalApp.ID]; ok {
				for _, whitelist := range whitelists {
					expectedPortalApp.AddWhitelist((whitelist))
				}
			}

			// Add pending data to cache
			cache.pendingAAT = test.pendingAAT
			cache.pendingSettings = test.pendingSettings
			cache.pendingAppNotifications = test.pendingAppNotifications
			cache.pendingWhitelists = test.pendingWhitelists

			// Add, update or delete PortalApp
			cache.upsertPortalApp(test.portalApp)

			portalAppFound := findPortalAppInPortalApps(cache.GetPortalApps(Options{IncludeDeleted: false}), test.portalApp.ID)
			if test.delete {
				c.False(portalAppFound, "PortalApp should be removed from portalApps map")
			} else {
				c.True(portalAppFound, "PortalApp should be found in portalApps map")

				// Check PortalApp in cache
				storedPortalApp := cache.GetPortalApp(test.portalApp.ID)
				c.Equal(expectedPortalApp, storedPortalApp, "PortalApp should be added to cache")
			}

			_, ok := cache.pendingAAT[test.portalApp.ID]
			c.False(ok, "Pending AAT should be removed from cache")
			_, ok = cache.pendingSettings[test.portalApp.ID]
			c.False(ok, "Pending Settings should be removed from cache")
			_, ok = cache.pendingAppNotifications[test.portalApp.ID]
			c.False(ok, "Pending AppNotifications should be removed from cache")
			_, ok = cache.pendingWhitelists[test.portalApp.ID]
			c.False(ok, "Pending Whitelists should be removed from cache")
		})
	}
}

func Test_addPortalAppAAT(t *testing.T) {
	tests := []struct {
		name        string
		aat         types.AAT
		portalAppID types.PortalAppID
		portalApp   *types.PortalApp
	}{
		{
			name: "Should add AAT to an existing PortalApp in cache",
			aat: types.AAT{
				ID:              "test_protocol_app_new",
				AppID:           "test_app_new_3450he46fhs7",
				Address:         "test_6bc837b1b0e526569dda9acbbcbb4f40",
				PublicKey:       "test_1e8f7cc2e1f812c7531fc020a75f4284",
				ClientPublicKey: "test_b8f50f694400f9b365ab5a498f5392dc",
				PrivateKey:      "test_3c28401b57940fbfd1c45da1a177d5e1",
				Signature:       "test_31f6d4a85f31e51f5b6d33dd2dd2fc2d",
				Version:         "0.0.1",
			},
			portalAppID: "test_app_new_3450he46fhs7",
			portalApp: &types.PortalApp{
				ID:        "test_app_new_3450he46fhs7",
				AccountID: "account_3",
				Name:      "new_pokt_app_1",
				Gigastake: true,
				Staked:    false,
			},
		},
		{
			name: "Should store AAT in pendingAAT map if PortalApp not in cache",
			aat: types.AAT{
				ID:              "test_protocol_app_new_2",
				AppID:           "test_app_nope_345he46fhs7",
				Address:         "test_8a75937963fee71e048c652a44491a53",
				PublicKey:       "test_ba65207ac55d99d1d190efb002ff3002",
				ClientPublicKey: "test_4d14919baae64635462597d587e1be43",
				PrivateKey:      "test_e02452030aa264af1d71bfa7ab021089",
				Signature:       "test_b77fc382440d5afe286e491bf86b7cad",
				Version:         "0.0.1",
			},
			portalAppID: "test_app_nope_345he46fhs7",
			portalApp:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			if test.portalApp != nil {
				cache.upsertPortalApp(*test.portalApp)
			}

			cache.addPortalAppAAT(test.aat)
			test.aat.AppID = "" // is unset in the cache TODO will be changed soon

			if test.portalApp != nil {
				storedPortalApp := cache.GetPortalApp(test.portalAppID)
				c.Equal(test.aat, storedPortalApp.AATs[test.aat.ID], "AAT should be added to the PortalApp in cache")
			} else {
				storedAAT, ok := cache.pendingAAT[test.portalAppID]
				c.True(ok, "AAT should be stored in pendingAAT map")
				c.Equal(test.aat, storedAAT, "AAT should be equal to the pending AAT")
			}
		})
	}
}

func Test_addPortalAppSettings(t *testing.T) {
	tests := []struct {
		name        string
		settings    types.Settings
		portalAppID types.PortalAppID
		portalApp   *types.PortalApp
	}{
		{
			name: "Should add Settings to an existing PortalApp in cache",
			settings: types.Settings{
				AppID:             "test_app_new_3450he46fhs7",
				SecretKey:         "test_secret_key_1",
				SecretKeyRequired: true,
			},
			portalAppID: "test_app_new_3450he46fhs7",
			portalApp: &types.PortalApp{
				ID:        "test_app_new_3450he46fhs7",
				AccountID: "account_3",
				Name:      "new_pokt_app_1",
				Gigastake: true,
				Staked:    false,
			},
		},
		{
			name: "Should store Settings in pendingSettings map if PortalApp not in cache",
			settings: types.Settings{
				AppID:             "test_app_nope_345he46fhs7",
				SecretKey:         "test_secret_key_2",
				SecretKeyRequired: false,
			},
			portalAppID: "test_app_nope_345he46fhs7",
			portalApp:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			if test.portalApp != nil {
				cache.upsertPortalApp(*test.portalApp)
			}

			cache.addPortalAppSettings(test.settings)
			test.settings.AppID = "" // is unset in the cache TODO will be changed soon

			if test.portalApp != nil {
				storedPortalApp := cache.GetPortalApp(test.portalAppID)
				c.Equal(test.settings.SecretKey, storedPortalApp.Settings.SecretKey, "SecretKey should be added to the PortalApp in cache")
				c.Equal(test.settings.SecretKeyRequired, storedPortalApp.Settings.SecretKeyRequired, "SecretKeyRequired should be added to the PortalApp in cache")
			} else {
				storedSettings, ok := cache.pendingSettings[test.portalAppID]
				c.True(ok, "Settings should be stored in pendingSettings map")
				c.Equal(test.settings, storedSettings, "Settings should be equal to the pending Settings")
			}
		})
	}
}

func Test_addAppNotification(t *testing.T) {
	tests := []struct {
		name         string
		notification types.AppNotification
		portalAppID  types.PortalAppID
		portalApp    *types.PortalApp
	}{
		{
			name: "Should add AppNotification to an existing PortalApp in cache",
			notification: types.AppNotification{
				AppID:       "test_app_new_3450he46fhs7",
				Type:        "test_type_1",
				Active:      true,
				Destination: "test_destination_1",
				Trigger:     "test_trigger_1",
				Events: map[types.NotificationEvent]bool{
					types.NotificationEventFull: true,
					types.NotificationEventHalf: false,
				},
			},
			portalAppID: "test_app_new_3450he46fhs7",
			portalApp: &types.PortalApp{
				ID:            "test_app_new_3450he46fhs7",
				Name:          "new_pokt_app_1",
				Gigastake:     true,
				Staked:        false,
				AccountID:     "account_3",
				Notifications: make(map[types.NotificationType]types.AppNotification),
			},
		},
		{
			name: "Should store AppNotification in pendingAppNotifications map if PortalApp not in cache",
			notification: types.AppNotification{
				AppID:       "test_app_nope_345he46fhs7",
				Type:        "test_type_2",
				Active:      false,
				Destination: "test_destination_2",
				Trigger:     "test_trigger_2",
				Events: map[types.NotificationEvent]bool{
					types.NotificationEventQuarter:  true,
					types.NotificationEventSignedUp: false,
				},
			},
			portalAppID: "test_app_nope_345he46fhs7",
			portalApp:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			if test.portalApp != nil {
				cache.upsertPortalApp(*test.portalApp)
			}

			cache.addAppNotification(test.notification)
			test.notification.AppID = "" // is unset in the cache TODO will be changed soon

			if test.portalApp != nil {
				storedPortalApp := cache.GetPortalApp(test.portalAppID)
				storedNotification := storedPortalApp.Notifications[test.notification.Type]
				c.Equal(test.notification, storedNotification, "AppNotification should be added to the PortalApp in cache")
			} else {
				storedNotification, ok := cache.pendingAppNotifications[test.portalAppID]
				c.True(ok, "AppNotification should be stored in pendingAppNotifications map")
				c.Equal(test.notification, storedNotification, "AppNotification should be equal to the pending AppNotification")
			}
		})
	}
}

func Test_addWhitelist(t *testing.T) {
	tests := []struct {
		name        string
		whitelist   types.Whitelist
		portalAppID types.PortalAppID
		portalApp   *types.PortalApp
	}{
		{
			name: "Should add Whitelist to an existing PortalApp in cache",
			whitelist: types.Whitelist{
				AppID:   "test_app_new_3450he46fhs7",
				Type:    types.WhitelistTypeOrigins,
				Value:   "https://example.com",
				ChainID: "test_chain_id_1",
			},
			portalAppID: "test_app_new_3450he46fhs7",
			portalApp: &types.PortalApp{
				ID:         "test_app_new_3450he46fhs7",
				Name:       "new_pokt_app_1",
				Gigastake:  true,
				Staked:     false,
				AccountID:  "account_3",
				Whitelists: types.Whitelists{},
			},
		},
		{
			name: "Should store Whitelist in pendingWhitelists map if PortalApp not in cache",
			whitelist: types.Whitelist{
				AppID:   "test_app_nope_345he46fhs7",
				Type:    types.WhitelistTypeUserAgents,
				Value:   "Mozilla/5.0",
				ChainID: "test_chain_id_2",
			},
			portalAppID: "test_app_nope_345he46fhs7",
			portalApp:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			if test.portalApp != nil {
				cache.upsertPortalApp(*test.portalApp)
			}

			cache.addWhitelist(test.whitelist)
			test.whitelist.AppID = "" // is unset in the cache TODO will be changed soon

			if test.portalApp != nil {
				storedPortalApp := cache.GetPortalApp(test.portalAppID)
				_, ok := storedPortalApp.Whitelists.Origins[types.Origin(test.whitelist.Value)]
				c.True(ok, "Whitelist should be added to the PortalApp in cache")
			} else {
				storedWhitelists, ok := cache.pendingWhitelists[test.portalAppID]
				c.True(ok, "Whitelist should be stored in pendingWhitelists map")
				c.Contains(storedWhitelists, test.whitelist, "Whitelist should be equal to the pending Whitelist")
			}
		})
	}
}

func Test_deleteWhitelist(t *testing.T) {
	tests := []struct {
		name        string
		whitelist   types.Whitelist
		portalAppID types.PortalAppID
		portalApp   *types.PortalApp
	}{
		{
			name: "Should remove Whitelist from an existing PortalApp in cache",
			whitelist: types.Whitelist{
				AppID:   "test_app_1",
				Type:    types.WhitelistTypeOrigins,
				Value:   "https://test.com",
				ChainID: "test_chain_id_1",
			},
			portalAppID: "test_app_new_3450he46fhs7",
			portalApp: &types.PortalApp{
				ID:         "test_app_1",
				Name:       "new_pokt_app_1",
				Gigastake:  true,
				Staked:     false,
				AccountID:  "account_3",
				Whitelists: types.Whitelists{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			storedPortalApp := cache.GetPortalApp(test.portalApp.ID)
			_, ok := storedPortalApp.Whitelists.Origins[types.Origin(test.whitelist.Value)]
			c.True(ok, "Whitelist should be present in PortalApp before update")

			if test.portalApp != nil {
				cache.upsertPortalApp(*test.portalApp)
			}

			// Add the whitelist before deleting it
			cache.addWhitelist(test.whitelist)

			cache.deleteWhitelist(test.whitelist)
			test.whitelist.AppID = "" // is unset in the cache TODO will be changed soon

			if test.portalApp != nil {
				storedPortalApp := cache.GetPortalApp(test.portalAppID)
				_, ok := storedPortalApp.Whitelists.Origins[types.Origin(test.whitelist.Value)]
				c.False(ok, "Whitelist should be removed from the PortalApp in cache")
			}
		})
	}
}

/* ---------- Chains ---------- */

func Test_upsertChain(t *testing.T) {
	tests := []struct {
		name                      string
		chain                     types.Chain
		pendingAltruists          map[types.RelayChainID][]types.Altruist
		pendingGigastakeRedirects map[types.RelayChainID][]types.GigastakeRedirect
		pendingChecks             map[types.RelayChainID][]types.Check
		update, delete            bool
	}{
		{
			name: "Should add a new Chain to cache with no pending side table data",
			chain: types.Chain{
				ID:          "0100",
				Blockchain:  "mango-mainnet",
				Description: "Mango Chain",
			},
		},
		{
			name: "Should add a new Chain to cache with pending Altruists",
			chain: types.Chain{
				ID:          "0016",
				Blockchain:  "greg-mainnet",
				Description: "Greg Chain",
			},
			pendingAltruists: map[types.RelayChainID][]types.Altruist{
				"0016": {
					{
						ChainID:  "0016",
						URL:      "https://test-rpc.greg.io:1234",
						Auth:     "auth_token",
						AuthType: types.ChainAuthTypeBasicAuth,
					},
					{
						ChainID:  "0016",
						URL:      "https://test-rpc.greg-1.io:5678",
						Auth:     "another_auth_token",
						AuthType: types.ChainAuthTypeBearerToken,
					},
				},
			},
		},
		{
			name: "Should add a new Chain to cache with pending GigastakeRedirects",
			chain: types.Chain{
				ID:          "0017",
				Blockchain:  "neptune-mainnet",
				Description: "Neptune Chain",
			},
			pendingGigastakeRedirects: map[types.RelayChainID][]types.GigastakeRedirect{
				"0017": {
					{
						ChainID:             "0017",
						PortalApplicationID: "test_app_1",
						Domain:              "nep-rpc.gateway.pokt.network",
						Alias:               "neptune-mainnet",
					},
				},
			},
		},
		{
			name: "Should add a new Chain to cache with pending Checks",
			chain: types.Chain{
				ID:          "0018",
				Blockchain:  "mercury-mainnet",
				Description: "Mercury Chain",
			},
			pendingChecks: map[types.RelayChainID][]types.Check{
				"0018": {
					{
						ChainID:   "0018",
						Type:      types.ChainCheckTypeSync,
						Payload:   `{"id":2,"jsonrpc":"2.0","method":"getSync"}`,
						ResultKey: "sync_status",
						Allowance: 3,
					},
				},
			},
		},

		{
			name:  "Should update an existing Chain in cache with all side tables",
			chain: *testdata.Chains["0021"],
			pendingAltruists: map[types.RelayChainID][]types.Altruist{
				"0021": {
					{
						ChainID:  "0021",
						URL:      "https://altruist-0021.com:1234",
						Auth:     "update:45678",
						AuthType: types.ChainAuthTypeBasicAuth,
					},
				},
			},
			pendingGigastakeRedirects: map[types.RelayChainID][]types.GigastakeRedirect{
				"0021": {
					{
						ChainID:             "0021",
						PortalApplicationID: "test_app_3",
						Domain:              "new-eth-rpc.gateway.pokt.network",
						Alias:               "eth-wow",
					},
				},
			},
			pendingChecks: map[types.RelayChainID][]types.Check{
				"0021": {
					{
						ChainID:   "0021",
						Type:      types.ChainCheckTypeSync,
						Payload:   `{"id":1,"jsonrpc":"2.0","method":"eth_suchWow","params":[]}`,
						ResultKey: "updated_result_key",
						Allowance: 3,
					},
					{
						ChainID:   "0021",
						Type:      types.ChainCheckTypeChain,
						Payload:   `{"method":"eth_Chonky","id":1,"jsonrpc":"2.0"}`,
						ResultKey: "id400",
					},
				},
			},
		},
		{
			name:   "Should delete an existing Chain in cache",
			chain:  *testdata.Chains["0064"],
			delete: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			chain := test.chain

			cache := NewTestCache(t)

			if test.update || test.delete {
				storedChain := cache.GetChain(chain.ID, Options{})
				c.Equal(chain, storedChain, "Chain should be present in cache before update")
			}
			if test.delete {
				chain.Deleted = true
			}

			// Update the expected Chain with the pending data
			expectedChain := chain
			if altruists, ok := test.pendingAltruists[chain.ID]; ok {
				expectedChain.Altruists = altruists
			}
			if gigastakeRedirects, ok := test.pendingGigastakeRedirects[chain.ID]; ok {
				expectedChain.Redirects = gigastakeRedirects
			}
			if checks, ok := test.pendingChecks[chain.ID]; ok {
				expectedChain.Checks = make(map[types.ChainCheckType]types.Check)
				for _, check := range checks {
					expectedChain.Checks[check.Type] = check
				}
			}

			// Add pending data to cache
			cache.pendingAltruists = test.pendingAltruists
			cache.pendingGigastakeRedirects = test.pendingGigastakeRedirects
			cache.pendingChecks = test.pendingChecks

			// Add, update or delete Chain
			cache.upsertChain(chain)

			chainFound := findChainInChains(cache.GetChains(Options{IncludeDeleted: false}), chain.ID)
			if test.delete {
				c.False(chainFound, "Chain should be removed from chains map")
			} else {
				c.True(chainFound, "Chain should be found in chains map")

				// Check Chain in cache
				storedChain := cache.GetChain(chain.ID, Options{})
				c.Equal(expectedChain, storedChain, "Chain should be added to cache")
			}

			_, ok := cache.pendingAltruists[chain.ID]
			c.False(ok, "Pending Altruists should be removed from cache")
			_, ok = cache.pendingGigastakeRedirects[chain.ID]
			c.False(ok, "Pending GigastakeRedirects should be removed from cache")
			_, ok = cache.pendingChecks[chain.ID]
			c.False(ok, "Pending Checks should be removed from cache")
		})
	}
}

func Test_addAltruist(t *testing.T) {
	tests := []struct {
		name          string
		altruist      types.Altruist
		chainID       types.RelayChainID
		existingChain bool
	}{
		{
			name: "Should add Altruist to an existing Chain in cache",
			altruist: types.Altruist{
				ChainID: "0053",
				URL:     "https://altruist-0053-1.com:1234",
			},
			chainID:       "0053",
			existingChain: true,
		},
		{
			name: "Should store Altruist in pendingAltruists map if Chain not in cache",
			altruist: types.Altruist{
				ChainID: "0981",
				URL:     "https://altruist-0981-1.com:1234",
			},
			chainID: "0981",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			cache.addAltruist(test.altruist)

			if test.existingChain {
				storedChain := cache.GetChain(test.chainID, Options{})
				found := false
				for _, storedAltruist := range storedChain.Altruists {
					if storedAltruist.URL == test.altruist.URL {
						found = true
						break
					}
				}
				c.True(found, "Altruist should be added to the Chain in cache")
			} else {
				storedAltruists, ok := cache.pendingAltruists[test.chainID]
				c.True(ok, "Altruist should be stored in pendingAltruists map")

				found := false
				for _, storedAltruist := range storedAltruists {
					if storedAltruist.URL == test.altruist.URL {
						found = true
						break
					}
				}
				c.True(found, "Altruist should be added to the pending Altruists")
			}
		})
	}
}

func Test_addGigastakeRedirect(t *testing.T) {
	tests := []struct {
		name          string
		redirect      types.GigastakeRedirect
		chainID       types.RelayChainID
		existingChain bool
	}{
		{
			name: "Should add GigastakeRedirect to an existing Chain in cache",
			redirect: types.GigastakeRedirect{
				ChainID:             "0053",
				PortalApplicationID: "test_app_2",
				Alias:               "altruist-0053",
				Domain:              "op-rpc-2.gateway.pokt.network",
			},
			chainID:       "0053",
			existingChain: true,
		},
		{
			name: "Should store GigastakeRedirect in pendingGigastakeRedirects map if Chain not in cache",
			redirect: types.GigastakeRedirect{
				ChainID:             "0981",
				PortalApplicationID: "test_app_2",
				Alias:               "altruist-0981",
				Domain:              "nope-rpc.gateway.pokt.network",
			},
			chainID: "0981",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			cache.addGigastakeRedirect(test.redirect)

			if test.existingChain {
				storedChain := cache.GetChain(test.chainID, Options{})
				found := false
				for _, storedRedirect := range storedChain.Redirects {
					if storedRedirect.Domain == test.redirect.Domain {
						found = true
						break
					}
				}
				c.True(found, "GigastakeRedirect should be added to the Chain in cache")
			} else {
				storedRedirects, ok := cache.pendingGigastakeRedirects[test.chainID]
				c.True(ok, "GigastakeRedirect should be stored in pendingGigastakeRedirects map")

				found := false
				for _, storedRedirect := range storedRedirects {
					if storedRedirect.Domain == test.redirect.Domain {
						found = true
						break
					}
				}
				c.True(found, "GigastakeRedirect should be added to the pending GigastakeRedirects")
			}
		})
	}
}

func Test_addCheck(t *testing.T) {
	tests := []struct {
		name          string
		check         types.Check
		chainID       types.RelayChainID
		existingChain bool
	}{
		{
			name: "Should add Check to an existing Chain in cache",
			check: types.Check{
				ChainID:   "0053",
				Type:      types.ChainCheckTypeSync,
				Payload:   `{"id":1,"jsonrpc":"2.0","method":"eth_protoMolecule","params":[]}`,
				ResultKey: "genesis",
				Allowance: 14,
			},
			chainID:       "0053",
			existingChain: true,
		},
		{
			name: "Should store Check in pendingChecks map if Chain not in cache",
			check: types.Check{
				ChainID:   "0981",
				Type:      types.ChainCheckTypeSync,
				Payload:   `{"id":1,"jsonrpc":"2.0","method":"eth_Gorbachev","params":[]}`,
				ResultKey: "nurdles",
				Allowance: 7,
			},
			chainID: "0981",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			cache.addCheck(test.check)

			if test.existingChain {
				storedChain := cache.GetChain(test.chainID, Options{})
				storedCheck, ok := storedChain.Checks[test.check.Type]
				c.True(ok, "Check should be added to the Chain in cache")
				c.Equal(test.check.Type, storedCheck.Type, "Check Type should match the stored Check Type")
			} else {
				storedChecks, ok := cache.pendingChecks[test.chainID]
				c.True(ok, "Check should be stored in pendingChecks map")

				found := false
				for _, storedCheck := range storedChecks {
					if storedCheck.Type == test.check.Type {
						found = true
						break
					}
				}
				c.True(found, "Check should be added to the pending Checks")
			}
		})
	}
}

/* ---------- Test Utils ---------- */
func accountsToMap(accounts []types.Account) map[types.AccountID]*types.Account {
	accountsMap := make(map[types.AccountID]*types.Account)
	for i := range accounts {
		accountsMap[accounts[i].ID] = &accounts[i]
	}
	return accountsMap
}

func portalAppsToMap(portalApps []types.PortalApp) map[types.PortalAppID]*types.PortalApp {
	portalAppsMap := make(map[types.PortalAppID]*types.PortalApp)
	for i := range portalApps {
		portalAppsMap[portalApps[i].ID] = &portalApps[i]
	}
	return portalAppsMap
}

func chainsToMap(chains []types.Chain) map[types.RelayChainID]*types.Chain {
	chainsMap := make(map[types.RelayChainID]*types.Chain)
	for i := range chains {
		chainsMap[chains[i].ID] = &chains[i]
	}
	return chainsMap
}

func payPlansToMap(payPlans []types.Plan) map[types.PayPlanType]*types.Plan {
	payPlansMap := make(map[types.PayPlanType]*types.Plan)
	for i := range payPlans {
		payPlansMap[payPlans[i].Type] = &payPlans[i]
	}
	return payPlansMap
}

func findAccountInAccounts(accounts []types.Account, accountID types.AccountID) bool {
	for _, account := range accounts {
		if account.ID == accountID {
			return true
		}
	}
	return false
}

func getAccountFromAccounts(accounts []types.Account, accountID types.AccountID) types.Account {
	for _, account := range accounts {
		if account.ID == accountID {
			return account
		}
	}
	return types.Account{}
}

func findPortalAppInPortalApps(portalApps []types.PortalApp, portalAppID types.PortalAppID) bool {
	for _, portalApp := range portalApps {
		if portalApp.ID == portalAppID {
			return true
		}
	}
	return false
}

func findChainInChains(chains []types.Chain, chainID types.RelayChainID) bool {
	for _, chain := range chains {
		if chain.ID == chainID {
			return true
		}
	}
	return false
}

func setGigastakeRedirectAccounts(chain types.Chain, cache *Cache) types.Chain {
	for i := range chain.Redirects {
		if portalApp, ok := cache.portalApps[chain.Redirects[i].PortalApplicationID]; ok {
			chain.Redirects[i].GigastakePortalApp = portalApp
		}
	}
	return chain
}

func setGigastakeRedirectAccountsForChains(chains map[types.RelayChainID]*types.Chain, cache *Cache) map[types.RelayChainID]*types.Chain {
	updatedChains := make(map[types.RelayChainID]*types.Chain)
	for id, chain := range chains {
		updatedChain := setGigastakeRedirectAccounts(*chain, cache)
		updatedChains[id] = &updatedChain
	}
	return updatedChains
}
