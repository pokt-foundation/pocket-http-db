package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	dbclient "github.com/pokt-foundation/portal-http-db/v2/client"
	"github.com/pokt-foundation/portal-http-db/v2/testdata"
	"github.com/pokt-foundation/portal-http-db/v2/types"
	"github.com/pokt-foundation/utils-go/logger"
)

var testCtx = context.Background()

func NewTestCache(t *testing.T) *Cache {
	c := require.New(t)

	readerMock, _ := NewReaderMock(t)

	readerMock.On("ReadAllUsers", testCtx).Return(testdata.Users(), nil)
	readerMock.On("ReadPlans", testCtx).Return(testdata.PayPlans(), nil)
	readerMock.On("ReadAccounts", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(testdata.Accounts(), nil)
	readerMock.On("ReadPortalApps", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(testdata.PortalApps(), nil)
	readerMock.On("ReadGigastakeApps", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(testdata.GigastakeApps(), nil)
	readerMock.On("ReadChains", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(testdata.Chains(), nil)
	readerMock.On("ReadBlockedContracts", testCtx).Return(testdata.GlobalBlockedContracts(), nil)
	readerMock.On("ReadUserByUserID", mock.Anything, types.UserID("usertest")).Return(testdata.Users()[types.UserID("usertest")], nil)
	readerMock.On("ReadValidatorConfig", mock.Anything).Return(testdata.ValidatorPluginConfig(), nil)

	cacheInstance, err := NewCache(testCtx, Config{
		Reader:              readerMock,
		Logger:              logger.New(),
		CacheRefreshMinutes: 5 * time.Minute,
	})
	c.NoError(err)

	return cacheInstance
}

/* ---------- Cache Set and Read ---------- */

func Test_SetCache(t *testing.T) {
	cache := NewTestCache(t)

	setAccountOpts := testdata.AccountOptions{IncludePortalApps: true, IncludePlans: true, AssignUsers: true}
	setChainOpts := testdata.ChainOptions{IncludeGigastakeApps: true}

	tests := []struct {
		name             string
		actual, expected interface{}
	}{
		{
			name:     "Should return a user for a given provider user ID",
			actual:   cache.GetUserByProviderID("auth0|chrisjen_avasarala"),
			expected: *testdata.Users()["user_5"],
		},
		{
			name:     "Should return a user for a given portal user ID",
			actual:   cache.GetUserByID("user_5"),
			expected: *testdata.Users()["user_5"],
		},
		{
			name:     "Should return all accounts",
			actual:   testdata.AccountsToMap(cache.GetAccounts(Options{})),
			expected: testdata.Accounts(setAccountOpts),
		},
		{
			name:     "Should return one account",
			actual:   cache.GetAccount("account_1"),
			expected: *testdata.Accounts(setAccountOpts)["account_1"],
		},
		{
			name:     "Should return account by ID",
			actual:   cache.GetUserAccount("account_1", "user_1", Options{}),
			expected: *testdata.Accounts(setAccountOpts)["account_1"],
		},
		{
			name:   "Should return accounts by user ID where accepted param not provided",
			actual: testdata.AccountsToMap(cache.GetUserAccounts("user_4", Options{})),
			expected: map[types.AccountID]*types.Account{
				"account_2": testdata.Accounts(setAccountOpts)["account_2"],
				"account_4": testdata.Accounts(setAccountOpts)["account_4"],
				"account_5": testdata.Accounts(setAccountOpts)["account_5"],
			},
		},
		{
			name: "Should return accounts by user ID where accepted = true",
			actual: testdata.AccountsToMap(cache.GetUserAccounts("user_4", Options{
				Accepted: dbclient.BoolPtr(true),
			})),
			expected: map[types.AccountID]*types.Account{
				"account_2": testdata.Accounts(setAccountOpts)["account_2"],
				"account_4": testdata.Accounts(setAccountOpts)["account_4"],
				"account_5": testdata.Accounts(setAccountOpts)["account_5"],
			},
		},
		{
			name:     "Should return portal app by ID",
			actual:   cache.GetPortalApp("test_app_1"),
			expected: *testdata.PortalApps()["test_app_1"],
		},
		{
			name:     "Should return all portal apps",
			actual:   testdata.PortalAppsToMap(cache.GetPortalApps(Options{})),
			expected: testdata.PortalApps(),
		},
		{
			name:     "Should return all portal app lites for middleware",
			actual:   testdata.PortalAppLitesToMap(cache.GetPortalAppsForMiddleware()),
			expected: testdata.PortalAppLites(),
		},
		{
			name:     "Should return gigastake app by AAT ID",
			actual:   cache.GetGigastakeAppByID("test_gigastake_app_1"),
			expected: *testdata.GigastakeApps()["test_gigastake_app_1"],
		},
		{
			name:     "Should return all gigastake apps",
			actual:   testdata.GigastakeAppsToMap(cache.GetGigastakeApps(Options{})),
			expected: testdata.GigastakeApps(),
		},
		{
			name:     "Should return chain by ID",
			actual:   cache.GetChain("0001", Options{}),
			expected: *testdata.Chains(setChainOpts)["0001"],
		},
		{
			name:     "Should return all active chains",
			actual:   testdata.ChainsToMap(cache.GetChains(Options{})),
			expected: testdata.FilterActiveChains(testdata.Chains(setChainOpts)),
		},
		{
			name:     "Should return all chains inclouding inactive",
			actual:   testdata.ChainsToMap(cache.GetChains(Options{IncludeInactiveChains: true})),
			expected: testdata.Chains(setChainOpts),
		},
		{
			name:     "Should return plan by type",
			actual:   cache.GetPlan(types.FreetierV0),
			expected: *testdata.PayPlans()[types.FreetierV0],
		},
		{
			name:     "Should return all pay plans",
			actual:   testdata.PayPlansToMap(cache.GetPlans()),
			expected: testdata.PayPlans(),
		},
		{
			name:     "Should return all global blocked contracts",
			actual:   cache.GetBlockedContracts(),
			expected: testdata.GlobalBlockedContracts(),
		},
		{
			name:     "Should return the Validator Plugin Config",
			actual:   cache.GetValidatorConfig(),
			expected: *testdata.ValidatorPluginConfig(),
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
			name:         "should fail if ReadAllUsers fails",
			failedMethod: "ReadAllUsers",
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
			name:         "should fail if ReadGigastakeApps fails",
			failedMethod: "ReadGigastakeApps",
			err:          errors.New("error on GigastakeApps"),
		},
		{
			name:         "should fail if ReadBlockedContracts fails",
			failedMethod: "ReadBlockedContracts",
			err:          errors.New("error on blocked contracts"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			readerMock, _ := NewReaderMock(t)

			switch test.failedMethod {

			case "ReadAllUsers":
				readerMock.On(test.failedMethod, testCtx).Return(nil, test.err)

			case "ReadPlans":
				readerMock.On("ReadAllUsers", testCtx).Return(testdata.Users(), nil)
				readerMock.On(test.failedMethod, testCtx).Return(nil, test.err)

			case "ReadAccounts":
				readerMock.On("ReadAllUsers", testCtx).Return(testdata.Users(), nil)
				readerMock.On("ReadPlans", testCtx).Return(map[types.PayPlanType]*types.Plan{}, nil)
				readerMock.On(test.failedMethod, testCtx, types.DriverOptions{IncludeDeleted: false}).Return(nil, test.err)

			case "ReadPortalApps":
				readerMock.On("ReadAllUsers", testCtx).Return(testdata.Users(), nil)
				readerMock.On("ReadPlans", testCtx).Return(map[types.PayPlanType]*types.Plan{}, nil)
				readerMock.On("ReadAccounts", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(map[types.AccountID]*types.Account{}, nil)
				readerMock.On(test.failedMethod, testCtx, types.DriverOptions{IncludeDeleted: false}).Return(nil, test.err)

			case "ReadChains":
				readerMock.On("ReadAllUsers", testCtx).Return(testdata.Users(), nil)
				readerMock.On("ReadPlans", testCtx).Return(map[types.PayPlanType]*types.Plan{}, nil)
				readerMock.On("ReadAccounts", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(map[types.AccountID]*types.Account{}, nil)
				readerMock.On("ReadPortalApps", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(map[types.PortalAppID]*types.PortalApp{}, nil)
				readerMock.On(test.failedMethod, testCtx, types.DriverOptions{IncludeDeleted: false}).Return(nil, test.err)

			case "ReadGigastakeApps":
				readerMock.On("ReadAllUsers", testCtx).Return(testdata.Users(), nil)
				readerMock.On("ReadPlans", testCtx).Return(map[types.PayPlanType]*types.Plan{}, nil)
				readerMock.On("ReadAccounts", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(map[types.AccountID]*types.Account{}, nil)
				readerMock.On("ReadPortalApps", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(map[types.PortalAppID]*types.PortalApp{}, nil)
				readerMock.On("ReadChains", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(map[types.RelayChainID]*types.Chain{}, test.err)

			case "ReadBlockedContracts":
				readerMock.On("ReadAllUsers", testCtx).Return(testdata.Users(), nil)
				readerMock.On("ReadPlans", testCtx).Return(map[types.PayPlanType]*types.Plan{}, nil)
				readerMock.On("ReadAccounts", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(map[types.AccountID]*types.Account{}, nil)
				readerMock.On("ReadPortalApps", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(map[types.PortalAppID]*types.PortalApp{}, nil)
				readerMock.On("ReadChains", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(map[types.RelayChainID]*types.Chain{}, nil)
				readerMock.On("ReadGigastakeApps", testCtx, types.DriverOptions{IncludeDeleted: false}).Return(map[types.GigastakeAppID]*types.GigastakeApp{}, nil)
				readerMock.On(test.failedMethod, testCtx).Return(types.GlobalBlockedContracts{}, test.err)
			}

			_, err := NewCache(testCtx, Config{
				Reader:              readerMock,
				Logger:              logger.New(),
				CacheRefreshMinutes: 5 * time.Minute,
			})
			c.ErrorIs(err, test.err)
		})
	}
}

/* ---------- Accounts ---------- */

func Test_upsertAccountUser(t *testing.T) {
	tests := []struct {
		name        string
		accountID   types.AccountID
		accountUser types.AccountUser
		updateRoles map[types.PortalAppID]types.RoleName
	}{
		{
			name:      "Should add a new Accepted user to account and user permissions",
			accountID: "account_2",
			accountUser: types.AccountUser{
				ID:       "user_1",
				Email:    "james.holden123@test.com",
				RoleName: types.RoleMember,
				Accepted: true,
			},
		},
		{
			name:      "Should add a new not-Accepted user to account",
			accountID: "account_3",
			accountUser: types.AccountUser{
				ID:       "user_10",
				Email:    "daenerys.targaryen123@test.com",
				RoleName: types.RoleMember,
				Accepted: true,
			},
		},
		{
			name:        "Should update an existing account User's roles",
			accountID:   "account_1",
			accountUser: testdata.AccountUsers()[2],
			updateRoles: map[types.PortalAppID]types.RoleName{
				"test_app_1": types.RoleMember,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			// Upsert account user
			testAccountUser := test.accountUser
			testAccountUser.AccountID = test.accountID
			cache.upsertAccountUser(testAccountUser)

			account := cache.GetUserAccount(test.accountID, test.accountUser.ID, Options{})
			accountUser, ok := account.Users[test.accountUser.ID]
			c.True(ok, "AccountUser should be present in account Users map")
			c.Equal(test.accountUser, accountUser, "AccountUser should be added or updated in the cache")

			accountFound := testdata.FindAccountInAccounts(cache.GetUserAccounts(test.accountUser.ID, Options{}), test.accountID)
			c.True(accountFound, "Account should be present in accountsByUserID map")

			user := cache.GetUserByID(test.accountUser.ID)
			accountPermissions, ok := user.Permissions[test.accountID]
			c.True(ok, "PortalAppPermissions should be present in UserPermissions")
			c.Equal(test.accountUser.RoleName, accountPermissions.RoleName, "Role should be updated in UserPermissions")
			c.Equal(test.accountUser.Accepted, accountPermissions.Accepted, "Accepted should be updated in UserPermissions")
		})
	}
}

func Test_upsertAccount(t *testing.T) {
	tests := []struct {
		name                string
		account             types.Account
		accountUser         types.AccountUser
		accountNotification types.AccountNotification
		userID              types.UserID
	}{
		{
			name:    "Should insert a new account",
			account: *testdata.TestCreateAccount,
			accountUser: types.AccountUser{
				AccountID: "create_account_id",
				ID:        "user_1",
				Email:     "james.holden123@test.com",
				RoleName:  types.RoleOwner,
				Accepted:  true,
			},
		},
		{
			name:    "Should add a new Account to cache with pending AppNotifications",
			account: *testdata.TestCreateAccount,
			accountUser: types.AccountUser{
				AccountID: "create_account_id",
				ID:        "user_1",
				Email:     "james.holden123@test.com",
				RoleName:  types.RoleOwner,
				Accepted:  true,
			},
			accountNotification: types.AccountNotification{
				AccountID:   "create_account_id",
				Type:        types.NotificationTypeEmail,
				Active:      true,
				Destination: "email@pokt.network",
				Trigger:     "trigger456",
				Events:      map[types.NotificationEvent]bool{types.NotificationEventHalf: true},
			},
		},
		{
			name: "Should update an existing account",
			account: types.Account{
				ID:       "account_1",
				PlanType: types.PayPlanType("enterprise_plan"),
				Users: map[types.UserID]types.AccountUser{
					"user_1": testdata.AccountUsers()[1],
					"user_2": testdata.AccountUsers()[2],
					"user_8": testdata.AccountUsers()[8],
				},
				PortalApps: map[types.PortalAppID]*types.PortalApp{
					"test_app_1": testdata.PortalApps()["test_app_1"],
				},
				Integrations: types.AccountIntegrations{
					CovalentAPIKeyFree: "covalent_api_key_1",
				},
				Notifications:          map[types.NotificationType]types.AccountNotification{"email": types.AccountNotification{AccountID: "", Type: "email", Active: true, Destination: "james.holden123@test.com", Trigger: "trigger123", Events: map[types.NotificationEvent]bool{"full": true, "quarter": true, "threeQuarters": true}}},
				PartnerChainIDs:        map[types.RelayChainID]struct{}{"0001": {}, "0053": {}},
				PartnerThroughputLimit: 2_000,
				PartnerAppLimit:        1,
				CreatedAt:              testdata.MockTimestamp,
				UpdatedAt:              testdata.MockTimestamp,
			},
			userID: "user_1",
		},
		{
			name: "Should update deleted account and remove from user ID accounts map",
			account: types.Account{
				ID:       "account_2",
				PlanType: types.PayPlanType("pro_plan"),
				Users: map[types.UserID]types.AccountUser{
					"user_3": testdata.AccountUsers()[3],
					"user_4": testdata.AccountUsers()[4],
					"user_9": testdata.AccountUsers()[9],
					"user_2": testdata.AccountUsers()[10],
				},
				PortalApps: map[types.PortalAppID]*types.PortalApp{
					"test_app_2": testdata.PortalApps()["test_app_2"],
				},
				Integrations: types.AccountIntegrations{CovalentAPIKeyFree: "covalent_api_key_1"},
				Deleted:      true,
			},
			userID: "user_3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			accountID := test.account.ID

			cache := NewTestCache(t)

			// Upsert account
			cache.upsertAccount(test.account)

			expectedAccount := test.account

			userID := test.userID
			if test.accountUser.ID != "" {
				userID = test.accountUser.ID
				cache.upsertAccountUser(test.accountUser)
			}
			if test.accountNotification.Type != "" {
				cache.addAccountNotification(test.accountNotification)
				if expectedAccount.Notifications == nil {
					expectedAccount.Notifications = make(map[types.NotificationType]types.AccountNotification)
				}
				test.accountNotification.AccountID = "" // is unset in the cache TODO will be changed soon
				expectedAccount.Notifications[test.accountNotification.Type] = test.accountNotification
			}

			if test.account.Deleted {
				c.Empty(cache.GetUserAccount(accountID, test.accountUser.ID, Options{}), "Account should be removed from cache")
			} else {
				expectedAccount.Plan = cache.plans[expectedAccount.PlanType]
				if test.accountUser.ID != "" {
					test.accountUser.AccountID = ""
					expectedAccount.Users = map[types.UserID]types.AccountUser{test.accountUser.ID: test.accountUser}
				}

				c.Equal(expectedAccount, cache.GetUserAccount(expectedAccount.ID, userID, Options{}), "Account should be present in cache and updated")

				accountFound := testdata.FindAccountInAccounts(cache.GetAccounts(Options{IncludeDeleted: false}), accountID)
				if expectedAccount.Deleted {
					c.False(accountFound, "Account should be removed from accounts map")
				} else {
					c.True(accountFound, "Account should be found in accounts map")
				}

				for _, user := range expectedAccount.Users {
					accountsForUser := cache.GetUserAccounts(user.ID, Options{
						Accepted: dbclient.BoolPtr(true),
					})

					accountInUserAccounts := testdata.FindAccountInAccounts(accountsForUser, accountID)
					if expectedAccount.Deleted {
						c.False(accountInUserAccounts, "Account should be removed from accountsByUserID map")
					} else {
						if user.Accepted {
							c.True(accountInUserAccounts, "Account should be present in accountsByUserID map", user.ID, accountID)
							c.Equal(expectedAccount, testdata.GetAccountFromAccounts(accountsForUser, accountID), "Account should be present in the cache and updated")
						} else {
							c.False(accountInUserAccounts, "Account should not be present in accountsByUserID map if user not accepted and options param set")
						}
					}
				}

				_, ok := cache.pendingAccountNotifications[expectedAccount.ID]
				c.False(ok, "Pending AccountNotifications should be removed from cache")
			}
		})
	}
}

func Test_addAccountNotification(t *testing.T) {
	tests := []struct {
		name         string
		notification types.AccountNotification
		userID       types.UserID
		accountID    types.AccountID
		account      *types.Account
	}{
		{
			name: "Should add AccountNotification to an existing Account in cache",
			notification: types.AccountNotification{
				AccountID:   "account_3",
				Type:        "test_type_1",
				Active:      true,
				Destination: "test_destination_1",
				Trigger:     "test_trigger_1",
				Events: map[types.NotificationEvent]bool{
					types.NotificationEventFull: true,
					types.NotificationEventHalf: false,
				},
			},
			userID:    "user_5",
			accountID: "account_3",
			account: &types.Account{
				ID:            "account_3",
				Name:          "new_pokt_app_1",
				Notifications: make(map[types.NotificationType]types.AccountNotification),
			},
		},
		{
			name: "Should store AccountNotification in pendingAppNotifications map if Account not in cache",
			notification: types.AccountNotification{
				AccountID:   "test_account_nope",
				Type:        "test_type_2",
				Active:      false,
				Destination: "test_destination_2",
				Trigger:     "test_trigger_2",
				Events: map[types.NotificationEvent]bool{
					types.NotificationEventQuarter:  true,
					types.NotificationEventSignedUp: false,
				},
			},
			accountID: "test_account_nope",
			account:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			if test.account != nil {
				cache.upsertAccount(*test.account)
			}

			cache.addAccountNotification(test.notification)
			test.notification.AccountID = "" // is unset in the cache TODO will be changed soon

			if test.account != nil {
				storedAccount := cache.GetUserAccount(test.accountID, test.userID, Options{})
				storedNotification := storedAccount.Notifications[test.notification.Type]
				c.Equal(test.notification, storedNotification, "AccountNotification should be added to the Account in cache")
			} else {
				storedNotification, ok := cache.pendingAccountNotifications[test.accountID]
				c.True(ok, "AccountNotification should be stored in pendingAccountNotifications map")
				c.Equal(test.notification, storedNotification, "AccountNotification should be equal to the pending AccountNotification")
			}
		})
	}
}

func Test_deleteAccountUser(t *testing.T) {
	tests := []struct {
		name            string
		accountID       types.AccountID
		accountUser     types.AccountUser
		existingAccount bool
	}{
		{
			name:            "Should delete an existing user access and update permissions",
			accountID:       "account_1",
			accountUser:     testdata.AccountUsers()[2],
			existingAccount: true,
		},
		{
			name:        "Should not fail when deleting a non-existing user access",
			accountID:   "account_5",
			accountUser: testdata.AccountUsers()[12],
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			if test.existingAccount {
				_, ok := cache.GetUserAccount(test.accountID, test.accountUser.ID, Options{}).Users[test.accountUser.ID]
				c.True(ok, "AccountUser should be found in account Users map before deletion")

				accountFound := testdata.FindAccountInAccounts(cache.GetUserAccounts(test.accountUser.ID, Options{}), test.accountID)
				c.True(accountFound, "Account should be found in accountsByUserID map before deletion")

				user := cache.GetUserByID(test.accountUser.ID)
				_, ok = user.Permissions[test.accountID]
				c.True(ok, "PortalAppPermissions should be found in UserPermissions for each PortalAppRole before deletion")
			}

			// Delete account user
			testAccountUser := test.accountUser
			testAccountUser.AccountID = test.accountID
			cache.deleteAccountUser(testAccountUser)

			account := cache.GetUserAccount(test.accountID, testAccountUser.ID, Options{})

			_, ok := account.Users[test.accountUser.ID]
			c.False(ok, "AccountUser should be deleted from account Users map")

			accountFound := testdata.FindAccountInAccounts(cache.GetUserAccounts(test.accountUser.ID, Options{}), test.accountID)
			c.False(accountFound, "Account should be removed from accountsByUserID map")

			user := cache.GetUserByID(test.accountUser.ID)
			_, ok = user.Permissions[test.accountID]
			c.False(ok, "PortalAppPermissions should be removed from UserPermissions for each PortalAppRole")
		})
	}
}

/* ---------- PortalApps ---------- */

func Test_upsertPortalApp(t *testing.T) {
	testApp := *testdata.TestCreatePortalAppByOwner
	testBaseApp := types.PortalApp{
		ID:        testApp.ID,
		AccountID: testApp.AccountID,
		Name:      testApp.Name,
	}

	tests := []struct {
		name              string
		portalApp         types.PortalApp
		pendingAAT        map[types.PortalAppID]types.AAT
		pendingSettings   map[types.PortalAppID]types.Settings
		pendingWhitelists map[types.PortalAppID][]types.Whitelist
		update, delete    bool
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
			pendingAAT: map[types.PortalAppID]types.AAT{
				testBaseApp.ID: testdata.TestCreatePortalAppAAT,
			},
		},
		{
			name:       "Should update an existing PortalApp in cache with all side tables",
			portalApp:  *testdata.PortalApps()["test_app_1"],
			pendingAAT: nil, // AAT is never updated
			pendingSettings: map[types.PortalAppID]types.Settings{
				"test_app_1": testApp.Settings,
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
			portalApp: *testdata.PortalApps()["test_app_3"],
			delete:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			if test.update || test.delete {
				storedPortalApp := cache.GetPortalApp(test.portalApp.ID)
				c.Equal(test.portalApp.ID, storedPortalApp.ID, "PortalApp should be present in cache before update")
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
			if whitelists, ok := test.pendingWhitelists[test.portalApp.ID]; ok {
				for _, whitelist := range whitelists {
					expectedPortalApp.AddWhitelist((whitelist))
				}
			}

			// Add pending data to cache
			cache.pendingPortalAppAATs = test.pendingAAT
			cache.pendingSettings = test.pendingSettings
			cache.pendingWhitelists = test.pendingWhitelists

			// Add, update or delete PortalApp
			cache.upsertPortalApp(test.portalApp)

			portalAppFound := testdata.FindPortalAppInPortalApps(cache.GetPortalApps(Options{IncludeDeleted: false}), test.portalApp.ID)
			if test.delete {
				c.False(portalAppFound, "PortalApp should be removed from portalApps map")
			} else {
				c.True(portalAppFound, "PortalApp should be found in portalApps map")

				// Check PortalApp in cache
				storedPortalApp := cache.GetPortalApp(test.portalApp.ID)
				c.Equal(expectedPortalApp.ID, storedPortalApp.ID, "PortalApp should be added to cache")
			}

			_, ok := cache.pendingPortalAppAATs[test.portalApp.ID]
			c.False(ok, "Pending AAT should be removed from cache")
			_, ok = cache.pendingSettings[test.portalApp.ID]
			c.False(ok, "Pending Settings should be removed from cache")
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
				ID:              "test_protocol_app_7",
				AppID:           "test_app_new_3450he46fhs7",
				Address:         "test_6bc837b1b0e526569dda9acbbcbb4f40",
				PublicKey:       "test_1e8f7cc2e1f812c7531fc020a75f4284",
				ClientPublicKey: "test_b8f50f694400f9b365ab5a498f5392dc",
				PrivateKey:      "",
				Signature:       "test_31f6d4a85f31e51f5b6d33dd2dd2fc2d",
				Version:         "0.0.1",
			},
			portalAppID: "test_app_new_3450he46fhs7",
			portalApp: &types.PortalApp{
				ID:        "test_app_new_3450he46fhs7",
				AccountID: "account_3",
				Name:      "new_pokt_app_1",
			},
		},
		{
			name: "Should store AAT in pendingAAT map if PortalApp not in cache",
			aat: types.AAT{
				ID:              "test_protocol_app_8",
				AppID:           "test_app_nope_345he46fhs7",
				Address:         "test_8a75937963fee71e048c652a44491a53",
				PublicKey:       "test_ba65207ac55d99d1d190efb002ff3002",
				ClientPublicKey: "test_4d14919baae64635462597d587e1be43",
				PrivateKey:      "",
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
				storedAAT, ok := cache.pendingPortalAppAATs[test.portalAppID]
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

/* ---------- GigastakeApps ---------- */

func Test_upsertGigastakeApp(t *testing.T) {
	tests := []struct {
		name                       string
		gigastakeApp               types.GigastakeApp
		pendingChainsGigastakeApps map[types.GigastakeAppID]types.ChainGigastakeApp
		update, delete             bool
	}{
		{
			name:                       "Should add a new GigastakeApp to cache with no pending AAT",
			gigastakeApp:               testdata.TestCreateGigastakeApp,
			pendingChainsGigastakeApps: nil,
			update:                     false,
			delete:                     false,
		},
		{
			name:         "Should add a new GigastakeApp to cache with pending AAT",
			gigastakeApp: testdata.TestCreateGigastakeApp,
			pendingChainsGigastakeApps: map[types.GigastakeAppID]types.ChainGigastakeApp{
				testdata.TestCreateGigastakeApp.ID: {
					ChainID:        "0001",
					GigastakeAppID: testdata.TestCreateGigastakeApp.ID,
				},
			},
			update: false,
			delete: false,
		},
		{
			name:         "Should update an existing GigastakeApp in cache with pending AAT",
			gigastakeApp: *testdata.GigastakeApps()["test_gigastake_app_1"],
			pendingChainsGigastakeApps: map[types.GigastakeAppID]types.ChainGigastakeApp{
				testdata.TestCreateGigastakeApp.ID: {
					ChainID:        "0001",
					GigastakeAppID: testdata.TestCreateGigastakeApp.ID,
				},
			},
			update: true,
			delete: false,
		},
		{
			name:                       "Should delete an existing GigastakeApp in cache",
			gigastakeApp:               *testdata.GigastakeApps()["test_gigastake_app_3"],
			pendingChainsGigastakeApps: nil,
			update:                     false,
			delete:                     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			if test.update || test.delete {
				storedGigastakeApp := cache.GetGigastakeAppByID(test.gigastakeApp.ID)
				c.Equal(test.gigastakeApp, storedGigastakeApp, "GigastakeApp should be present in cache before update")
			}
			if test.delete {
				test.gigastakeApp.Deleted = true
			}

			// Update the expected GigastakeApp with the pending AAT
			expectedGigastakeApp := test.gigastakeApp
			if chainGigastake, ok := test.pendingChainsGigastakeApps[test.gigastakeApp.ID]; ok {
				expectedGigastakeApp.ChainIDs[chainGigastake.ChainID] = struct{}{}
			}

			// Add pending ChainGigastake to cache
			cache.pendingChainsGigastakeApps = test.pendingChainsGigastakeApps

			// Add, update, or delete GigastakeApp
			cache.upsertGigastakeApp(test.gigastakeApp)

			gigastakeAppFound := testdata.FindGigastakeAppInGigastakeApps(cache.GetGigastakeApps(Options{}), test.gigastakeApp.ID)
			if test.delete {
				c.False(gigastakeAppFound, "GigastakeApp should be removed from gigastakeApps map")
			} else {
				c.True(gigastakeAppFound, "GigastakeApp should be found in gigastakeApps map")

				// Check GigastakeApp in cache
				storedGigastakeApp := cache.GetGigastakeAppByID(test.gigastakeApp.ID)
				c.Equal(expectedGigastakeApp, storedGigastakeApp, "GigastakeApp should be added to cache")
			}

			_, ok := cache.pendingChainsGigastakeApps[test.gigastakeApp.ID]
			c.False(ok, "Pending AAT should be removed from cache")
		})
	}
}

/* ---------- Chains ---------- */

func Test_upsertChain(t *testing.T) {
	tests := []struct {
		name                string
		chain               types.Chain
		pendingAltruists    map[types.RelayChainID][]types.Altruist
		pendingAliasDomains map[types.RelayChainID][]types.AliasDomains
		pendingChecks       map[types.RelayChainID][]types.Check
		update, delete      bool
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
			name: "Should add a new Chain to cache with pending AliasDomains",
			chain: types.Chain{
				ID:          "0017",
				Blockchain:  "neptune-mainnet",
				Description: "Neptune Chain",
			},
			pendingAliasDomains: map[types.RelayChainID][]types.AliasDomains{
				"0017": {
					{
						ChainID: "0017",
						Alias:   "neptune-mainnet",
						Domains: []types.ChainDomain{"nep-rpc.gateway.pokt.network"},
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
			chain: *testdata.Chains()["0021"],
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
			pendingAliasDomains: map[types.RelayChainID][]types.AliasDomains{
				"0021": {
					{
						ChainID: "0021",
						Alias:   "eth-wow",
						Domains: []types.ChainDomain{"new-eth-rpc.gateway.pokt.network"},
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
			chain:  *testdata.Chains()["0064"],
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
				expectedChain.Altruists = make(map[types.AltruistURL]types.Altruist)
				for _, altruist := range altruists {
					expectedChain.Altruists[altruist.URL] = altruist
				}
			}
			if aliasDomains, ok := test.pendingAliasDomains[chain.ID]; ok {
				if expectedChain.AliasDomains == nil {
					expectedChain.AliasDomains = make(map[types.ChainAlias][]types.ChainDomain)
				}
				for _, aliasDomain := range aliasDomains {
					expectedChain.AliasDomains[aliasDomain.Alias] = aliasDomain.Domains
				}
			}
			if checks, ok := test.pendingChecks[chain.ID]; ok {
				expectedChain.Checks = make(map[types.ChainCheckType]types.Check)
				for _, check := range checks {
					expectedChain.Checks[check.Type] = check
				}
			}

			// Add pending data to cache
			cache.pendingAltruists = test.pendingAltruists
			cache.pendingAliasDomains = test.pendingAliasDomains
			cache.pendingChecks = test.pendingChecks

			// Add or update Chain
			cache.upsertChain(chain)

			chainFound := testdata.FindChainInChains(cache.GetChains(Options{IncludeInactiveChains: true}), chain.ID)
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
			_, ok = cache.pendingAliasDomains[chain.ID]
			c.False(ok, "Pending AliasDomains should be removed from cache")
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

func Test_addAliasDomains(t *testing.T) {
	tests := []struct {
		name          string
		aliasDomains  types.AliasDomains
		chainID       types.RelayChainID
		existingChain bool
	}{
		{
			name: "Should add AliasDomains to an existing Chain in cache",
			aliasDomains: types.AliasDomains{
				ChainID: "0053",
				Alias:   "altruist-0053",
				Domains: []types.ChainDomain{"op-rpc-2.gateway.pokt.network"},
			},
			chainID:       "0053",
			existingChain: true,
		},
		{
			name: "Should store AliasDomains in pendingAliasDomains map if Chain not in cache",
			aliasDomains: types.AliasDomains{
				ChainID: "0981",
				Alias:   "altruist-0981",
				Domains: []types.ChainDomain{"nope-rpc.gateway.pokt.network"},
			},
			chainID: "0981",
		},
	}

	domainSliceEqual := func(a, b []types.ChainDomain) bool {
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
		return true
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			cache := NewTestCache(t)

			cache.addAliasDomains(test.aliasDomains)

			if test.existingChain {
				storedChain := cache.GetChain(test.chainID, Options{})
				found := false
				for alias, domains := range storedChain.AliasDomains {
					if alias == test.aliasDomains.Alias && domainSliceEqual(domains, test.aliasDomains.Domains) {
						found = true
						break
					}
				}
				c.True(found, "AliasDomains should be added to the Chain in cache")
			} else {
				storedAliasDomains, ok := cache.pendingAliasDomains[test.chainID]
				c.True(ok, "AliasDomains should be stored in pendingAliasDomains map")

				found := false
				for _, storedAliasDomain := range storedAliasDomains {
					if storedAliasDomain.Alias == test.aliasDomains.Alias && domainSliceEqual(storedAliasDomain.Domains, test.aliasDomains.Domains) {
						found = true
						break
					}
				}
				c.True(found, "AliasDomains should be added to the pending AliasDomains")
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

/* ---------- Users ---------- */

func Test_upsertUser(t *testing.T) {
	tests := []struct {
		name               string
		user, expectedUser types.User
		pendingAccounts    bool
		err                error
	}{
		{
			name: "Should insert a new Portal User with no pending accounts",
			user: types.User{
				ID:       "user_22", // user ID set in test case
				Email:    "geralt.of.rivia623@example.com",
				SignedUp: true,
				AuthProviders: map[types.AuthType]types.UserAuthProvider{
					types.AuthTypeAuth0Username: {
						ProviderUserID: "auth0|geralt_of_rivia",
						Type:           types.AuthTypeAuth0Username,
						Provider:       types.AuthProviderAuth0,
						Federated:      false,
					},
				},
				CreatedAt: testdata.MockTimestamp,
				UpdatedAt: testdata.MockTimestamp,
			},
			expectedUser: types.User{
				ID:       "user_22", // user ID set in test case
				Email:    "geralt.of.rivia623@example.com",
				SignedUp: true,
				AuthProviders: map[types.AuthType]types.UserAuthProvider{
					types.AuthTypeAuth0Username: {
						ProviderUserID: "auth0|geralt_of_rivia",
						Type:           types.AuthTypeAuth0Username,
						Provider:       types.AuthProviderAuth0,
						Federated:      false,
					},
				},
				CreatedAt: testdata.MockTimestamp,
				UpdatedAt: testdata.MockTimestamp,
			},
			err: nil,
		},
		{
			name:            "Should insert a new Portal User with pending accounts",
			user:            *testdata.Users()["user_10"],
			expectedUser:    *testdata.Users()["user_10"],
			pendingAccounts: true,
			err:             nil,
		},
		{
			name: "Should update an existing user",
			user: types.User{
				ID:               "user_1",
				Email:            "james.holden752@test.com",
				SignedUp:         true,
				IconURL:          "https://picsum.photos/666",
				UpdatesProduct:   true,
				UpdatesMarketing: true,
				BetaTester:       false,
			},
			expectedUser: types.User{
				ID:               "user_1",
				Email:            "james.holden752@test.com",
				IconURL:          "https://picsum.photos/666",
				SignedUp:         true,
				UpdatesProduct:   true,
				UpdatesMarketing: true,
				Permissions: map[types.AccountID]types.AccountPermissions{
					"account_1": {Accepted: true, RoleName: "OWNER", Permissions: []types.Permissions{"read:endpoint", "write:endpoint", "delete:endpoint", "transfer:endpoint"}},
				},
				AuthProviders: map[types.AuthType]types.UserAuthProvider{
					"auth0_github": {
						ProviderUserID: "github|james_holden",
						Type:           "auth0_github",
						Provider:       "auth0",
						Federated:      true,
					},
					"auth0_username": {
						ProviderUserID: "auth0|james_holden",
						Type:           "auth0_username",
						Provider:       "auth0",
					},
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			cache := NewTestCache(t)

			cache.upsertUser(test.user)

			user := cache.GetUserByID(test.user.ID)
			c.Equal(test.expectedUser, user)

			if test.pendingAccounts {
				accounts := cache.GetUserAccounts(test.user.ID, Options{})
				for _, account := range accounts {
					c.NotEmpty(account.Users[test.user.ID])
				}
			}
		})
	}
}

func Test_addUserAuthProvider(t *testing.T) {
	tests := []struct {
		name         string
		userAuth     types.UserAuthProvider
		initialUser  types.User
		expectedUser types.User
	}{
		{
			name: "Should add a new AuthProvider to an existing user",
			userAuth: types.UserAuthProvider{
				UserID:         "user_1",
				Type:           types.AuthTypeAuth0Github,
				ProviderUserID: "github|james_holden",
				Provider:       "auth0",
				Federated:      true,
			},
			initialUser: types.User{
				ID:    "user_1",
				Email: "james.holden752@test.com",
				AuthProviders: map[types.AuthType]types.UserAuthProvider{
					types.AuthTypeAuth0Username: {
						ProviderUserID: "auth0|james_holden",
						Type:           types.AuthTypeAuth0Username,
						Provider:       "auth0",
						Federated:      false,
					},
				},
			},
			expectedUser: types.User{
				ID:    "user_1",
				Email: "james.holden752@test.com",
				Permissions: map[types.AccountID]types.AccountPermissions{
					"account_1": {Accepted: true, RoleName: "OWNER", Permissions: []types.Permissions{"read:endpoint", "write:endpoint", "delete:endpoint", "transfer:endpoint"}},
				},
				AuthProviders: map[types.AuthType]types.UserAuthProvider{
					"auth0_github": {
						ProviderUserID: "github|james_holden",
						Type:           "auth0_github",
						Provider:       "auth0",
						Federated:      true,
					},
					"auth0_username": {
						ProviderUserID: "auth0|james_holden",
						Type:           "auth0_username",
						Provider:       "auth0",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			cache := NewTestCache(t)

			cache.upsertUser(test.initialUser)

			cache.addUserAuthProvider(test.userAuth)

			actualUser := cache.GetUserByID(test.userAuth.UserID)
			c.Equal(test.expectedUser, actualUser)
		})
	}
}

func Test_deleteUser(t *testing.T) {
	tests := []struct {
		name        string
		initialUser types.User
	}{
		{
			name: "Should delete an existing user",
			initialUser: types.User{
				ID:    "user_1",
				Email: "james.holden752@test.com",
				AuthProviders: map[types.AuthType]types.UserAuthProvider{
					types.AuthTypeAuth0Username: {
						ProviderUserID: "auth0|james_holden",
						Type:           types.AuthTypeAuth0Username,
						Provider:       "auth0",
						Federated:      false,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			cache := NewTestCache(t)

			cache.upsertUser(test.initialUser)

			actualUser := cache.GetUserByID(test.initialUser.ID)
			c.NotEmpty(actualUser)

			cache.deleteUser(test.initialUser)

			deletedUser := cache.GetUserByID(test.initialUser.ID)
			c.Empty(deletedUser)

			deletedUser = cache.GetUserByProviderID(test.initialUser.AuthProviders[types.AuthTypeAuth0Username].ProviderUserID)
			c.Empty(deletedUser)
		})
	}
}
