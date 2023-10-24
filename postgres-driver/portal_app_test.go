package postgresdriver

import (
	"context"
	"fmt"
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/testdata"
	"github.com/pokt-foundation/portal-http-db/v2/types"
)

func (ts *PGDriverTestSuite) Test_PGDriver_ReadPortalApps() {
	tests := []struct {
		name       string
		portalApps map[types.PortalAppID]*types.PortalApp
		options    types.DriverOptions
		err        error
	}{
		{
			name: "Should return all non-deleted PortalApps from the database",
			portalApps: map[types.PortalAppID]*types.PortalApp{
				testdata.PortalApps["test_app_1"].ID: testdata.PortalApps["test_app_1"],
				testdata.PortalApps["test_app_2"].ID: testdata.PortalApps["test_app_2"],
				testdata.PortalApps["test_app_3"].ID: testdata.PortalApps["test_app_3"],
			},
			options: types.DriverOptions{},
			err:     nil,
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			portalApps, err := ts.driver.ReadPortalApps(context.Background(), test.options)
			ts.Equal(test.err, err)
			ts.Equal(test.portalApps, portalApps)
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_Z_SetPortalAppDeleted() {
	tests := []struct {
		name                                          string
		deleteParams                                  DeletePortalAppParams
		portalAppsBeforeDelete, portalAppsAfterDelete map[types.PortalAppID]*types.PortalApp
		err                                           error
	}{
		{
			name: "Should set a PortalApp's deleted field to true, causing it to not appear in the ReadPortalApps query",
			deleteParams: DeletePortalAppParams{
				ID: testdata.PortalApps["test_app_3"].ID, DeletedAt: newTimestamptz(testdata.MockTimestamp),
			},
			portalAppsBeforeDelete: map[types.PortalAppID]*types.PortalApp{
				testdata.PortalApps["test_app_1"].ID: testdata.PortalApps["test_app_1"],
				testdata.PortalApps["test_app_2"].ID: testdata.PortalApps["test_app_2"],
				testdata.PortalApps["test_app_3"].ID: testdata.PortalApps["test_app_3"],
			},
			portalAppsAfterDelete: map[types.PortalAppID]*types.PortalApp{
				testdata.PortalApps["test_app_1"].ID: testdata.PortalApps["test_app_1"],
				testdata.PortalApps["test_app_2"].ID: testdata.PortalApps["test_app_2"],
			},
			err: nil,
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			// Check all PortalApps exist before delete
			portalApps, err := ts.driver.ReadPortalApps(context.Background(), types.DriverOptions{IncludeDeleted: false})
			ts.Equal(test.err, err)
			ts.Equal(test.portalAppsBeforeDelete, filterAppsCreatedInTest(portalApps))

			// Delete PortalApp
			err = ts.driver.SetPortalAppDeleted(context.Background(), test.deleteParams.ID, test.deleteParams.DeletedAt.Time)
			ts.Equal(test.err, err)

			// Check PortalApp was deleted
			portalApps, err = ts.driver.ReadPortalApps(context.Background(), types.DriverOptions{IncludeDeleted: false})
			ts.Equal(test.err, err)
			ts.Equal(test.portalAppsAfterDelete, filterAppsCreatedInTest(portalApps))

			// Check PortalApp still appears if IncludeDeleted: true
			portalApps, err = ts.driver.ReadPortalApps(context.Background(), types.DriverOptions{IncludeDeleted: true})
			ts.Equal(test.err, err)
			testDeletedApp, ok := test.portalAppsBeforeDelete[test.deleteParams.ID]
			ts.True(ok)
			testDeletedApp.Deleted = true
			ts.Equal(test.portalAppsBeforeDelete, filterAppsCreatedInTest(portalApps))
		})
	}
}

// filterAppsCreatedInTest filters out PortalApps created in this test suite
func filterAppsCreatedInTest(portalApps map[types.PortalAppID]*types.PortalApp) map[types.PortalAppID]*types.PortalApp {
	filtered := map[types.PortalAppID]*types.PortalApp{}
	for id, app := range portalApps {
		if len(app.ID) != 8 {
			app.CreatedAt = testdata.MockTimestamp
			app.UpdatedAt = testdata.MockTimestamp
			app.FirstDateSurpassed = testdata.MockTimestamp
			filtered[id] = app
		}
	}
	return filtered
}

func (ts *PGDriverTestSuite) Test_PGDriver_WritePortalApp() {
	tests := []struct {
		name            string
		portalApp       types.PortalApp
		aat             types.AAT
		portalAppUsers  map[types.UserID]types.AccountUser
		testCreatedTime time.Time
		err             error
	}{
		{
			name:            "Should create a new PortalApp in the database when created by owner",
			portalApp:       *testdata.TestCreatePortalAppByOwner,
			aat:             testdata.TestCreatePortalAppAAT,
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name:            "Should create a new PortalApp in the database when created by non-owner",
			portalApp:       *testdata.TestCreatePortalAppByNonOwner,
			aat:             testdata.TestCreatePortalAppAAT,
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should fail when name not provided",
			portalApp: types.PortalApp{
				AppEmoji: "1F31B",
				Name:     "",
			},
			aat: testdata.TestCreatePortalAppAAT,
			err: errEmptyPortalAppName,
		},
		{
			name: "Should fail when invalid environment provided",
			portalApp: types.PortalApp{
				AppEmoji: "1F31B",
				Name:     "sebastian",
				Settings: types.Settings{Environment: types.Environment("under da sea")},
			},
			aat: testdata.TestCreatePortalAppAAT,
			err: fmt.Errorf(errInvalidEnvironment.Error(), "under da sea"),
		},
		{
			name: "Should fail when plan does not exist",
			portalApp: types.PortalApp{
				AppEmoji: "1F31B",
				Name:     "whatever",
				LegacyFields: types.LegacyFields{
					PlanType: "nonexistent-plan",
				},
				Settings: types.Settings{Environment: types.EnvironmentProduction},
			},
			aat: testdata.TestCreatePortalAppAAT,
			err: fmt.Errorf(errPayPlanDoesntExist.Error(), "nonexistent-plan"),
		},
		{
			name: "Should fail when account does not exist",
			portalApp: types.PortalApp{
				AppEmoji: "1F31B",
				Name:     "whatever",
				LegacyFields: types.LegacyFields{
					PlanType: types.FreetierV0,
				},
				Settings:  types.Settings{Environment: types.EnvironmentProduction},
				AccountID: "nonexistent-account",
			},
			aat: testdata.TestCreatePortalAppAAT,
			err: fmt.Errorf(errAccountDoesntExist.Error(), "nonexistent-account"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			createdPortalApp, err := ts.driver.WritePortalApp(context.Background(), test.portalApp, test.aat, test.testCreatedTime)
			ts.Equal(test.err, err)

			if test.err == nil {
				test.portalApp.ID = createdPortalApp.ID
				test.portalApp.AATs = createdPortalApp.AATs
				ts.Equal(&test.portalApp, createdPortalApp)

				// Check portal app can be read from driver
				portalApps, err := ts.driver.ReadPortalApps(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				for appID, aat := range test.portalApp.AATs {
					aat.PrivateKey = "" // PrivateKey is never read from the DB
					test.portalApp.AATs[appID] = aat
				}
				ts.Equal(&test.portalApp, portalApps[createdPortalApp.ID])

			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_UpdatePortalApp() {
	tests := []struct {
		name                     string
		updatePortalApp          types.UpdatePortalApp
		testUpdateTime           time.Time
		testUpdatedName          string
		testUpdatedDescription   string
		testUpdatedAppEmoji      types.AppEmoji
		testUpdatedSettings      types.Settings
		testUpdatedNotifications map[types.NotificationType]types.AppNotification
		testUpdatedWhitelists    types.Whitelists
		testUpdatedLegacyFields  types.LegacyFields
		err                      error
	}{
		{
			name: "Should update a new PortalApp in the database with all fields",
			updatePortalApp: types.UpdatePortalApp{
				Name:                 testdata.UpdatePortalAppName,
				Description:          testdata.UpdatePortalAppDescription,
				AppEmoji:             testdata.UpdatePortalAppEmoji,
				Settings:             testdata.UpdatePortalAppSettings,
				Notifications:        testdata.UpdatePortalAppNotifications,
				Whitelists:           testdata.UpdatePortalAppWhitelists,
				PlanType:             testdata.UpdatePortalAppPlan.PlanType,
				StripeSubscriptionID: testdata.UpdatePortalAppStripeSubscriptionID.StripeSubscriptionID,
			},
			testUpdateTime:         testdata.MockTimestamp,
			testUpdatedName:        "portal-app-updated",
			testUpdatedDescription: testdata.UpdatePortalAppDescription,
			testUpdatedAppEmoji:    testdata.UpdatePortalAppEmoji,
			testUpdatedSettings: types.Settings{
				Environment:       types.EnvironmentProduction,
				SecretKey:         "test_9d07c8a96ad53e7c288b0e86f37c5680",
				SecretKeyRequired: true,
				MonthlyRelayLimit: 2_500_000,
				FavoritedChainIDs: map[types.RelayChainID]struct{}{"0003": {}, "0009": {}, "00H3": {}},
			},
			testUpdatedNotifications: map[types.NotificationType]types.AppNotification{
				types.NotificationTypeEmail: {
					Type:        types.NotificationTypeEmail,
					Active:      true,
					Destination: "user@example.com",
					Trigger:     "daily",
					Events: map[types.NotificationEvent]bool{
						types.NotificationEventSignedUp:      true,
						types.NotificationEventHalf:          true,
						types.NotificationEventQuarter:       true,
						types.NotificationEventThreeQuarters: true,
						types.NotificationEventFull:          true,
					},
				},
				types.NotificationTypeWebhook: {
					Type:        types.NotificationTypeWebhook,
					Active:      true,
					Destination: "https://example.com/webhook",
					Trigger:     "hourly",
					Events: map[types.NotificationEvent]bool{
						types.NotificationEventHalf: true,
						types.NotificationEventFull: true,
					},
				},
			},
			testUpdatedWhitelists: types.Whitelists{
				Origins:     map[types.Origin]struct{}{"https://portalgun.io": {}, "https://subdomain.example.com": {}, "https://www.example.com": {}},
				UserAgents:  map[types.UserAgent]struct{}{"Brave": {}, "Google Chrome": {}, "Mozilla Firefox": {}, "Netscape Navigator": {}, "Safari": {}},
				Blockchains: map[types.RelayChainID]struct{}{"0001": {}, "0002": {}, "003E": {}, "0056": {}},
				Contracts: map[types.RelayChainID]map[types.Contract]struct{}{
					"0001": {"0xtest_2f78db6436527729929aaf6c616361de0f7": {}, "0xtest_5fbfe3e9af3971dd833d26ba9b5c936f0be": {}},
					"0002": {"0xtest_1111117dc0aa78b770fa6a738034120c302": {}, "0xtest_a39b223fe8d0a0e5c4f27ead9083c756cc2": {}},
					"003E": {"0xtest_0a85d5af5bf1d1762f925bdaddc4201f984": {}, "0xtest_f958d2ee523a2206206994597c13d831ec7": {}},
					"0056": {"0xtest_00000f279d81a1d3cc75430faa017fa5a2e": {}, "0xtest_5068778dd592e39a122f4f5a5cf09c90fe2": {}},
				},
				Methods: map[types.RelayChainID]map[types.Method]struct{}{
					"0001": {"GET": {}, "POST": {}, "PUT": {}},
					"0002": {"DELETE": {}, "GET": {}, "POST": {}, "PUT": {}},
					"003E": {"GET": {}},
					"0056": {"GET": {}, "POST": {}},
				},
			},
			testUpdatedLegacyFields: types.LegacyFields{
				PlanType:             types.FreetierV0,
				DailyLimit:           250_000,
				StripeSubscriptionID: "update_stripe_subscription_id",
			},
			err: nil,
		},
		{
			name: "Should update a new PortalApp in the database with only a new Name",
			updatePortalApp: types.UpdatePortalApp{
				Name: testdata.UpdatePortalAppName,
			},
			testUpdateTime:  testdata.MockTimestamp,
			testUpdatedName: "portal-app-updated",
			err:             nil,
		},
		{
			name: "Should update a new PortalApp in the database with only a new Description",
			updatePortalApp: types.UpdatePortalApp{
				Description: "Updating the application name like altering memories in the neon-lit streets of futuristic Los Angeles.",
			},
			testUpdateTime:         testdata.MockTimestamp,
			testUpdatedDescription: "Updating the application name like altering memories in the neon-lit streets of futuristic Los Angeles.",
			err:                    nil,
		},
		{
			name: "Should update a new PortalApp in the database with only a new App Emoji",
			updatePortalApp: types.UpdatePortalApp{
				AppEmoji: "1F336",
			},
			testUpdateTime:      testdata.MockTimestamp,
			testUpdatedAppEmoji: "1F336",
			err:                 nil,
		},
		{
			name: "Should update a new PortalApp in the database with only new Settings",
			updatePortalApp: types.UpdatePortalApp{
				Settings: testdata.UpdatePortalAppSettings,
			},
			testUpdateTime: testdata.MockTimestamp,
			testUpdatedSettings: types.Settings{
				Environment:       types.EnvironmentProduction,
				SecretKey:         "test_9d07c8a96ad53e7c288b0e86f37c5680",
				SecretKeyRequired: true,
				MonthlyRelayLimit: 2_500_000,
				FavoritedChainIDs: map[types.RelayChainID]struct{}{"0003": {}, "0009": {}, "00H3": {}},
			},
			err: nil,
		},
		{
			name: "Should update a new PortalApp in the database with only new Notifications",
			updatePortalApp: types.UpdatePortalApp{
				Notifications: testdata.UpdatePortalAppNotifications,
			},
			testUpdateTime: testdata.MockTimestamp,
			testUpdatedNotifications: map[types.NotificationType]types.AppNotification{
				types.NotificationTypeEmail: {
					Type:        types.NotificationTypeEmail,
					Active:      true,
					Destination: "user@example.com",
					Trigger:     "daily",
					Events: map[types.NotificationEvent]bool{
						types.NotificationEventSignedUp:      true,
						types.NotificationEventHalf:          true,
						types.NotificationEventQuarter:       true,
						types.NotificationEventThreeQuarters: true,
						types.NotificationEventFull:          true,
					},
				},
				types.NotificationTypeWebhook: {
					Type:        types.NotificationTypeWebhook,
					Active:      true,
					Destination: "https://example.com/webhook",
					Trigger:     "hourly",
					Events: map[types.NotificationEvent]bool{
						types.NotificationEventHalf: true,
						types.NotificationEventFull: true,
					},
				},
			},
			err: nil,
		},
		{
			name: "Should update a new PortalApp in the database with only new Whitelists",
			updatePortalApp: types.UpdatePortalApp{
				Whitelists: testdata.UpdatePortalAppWhitelists,
			},
			testUpdateTime: testdata.MockTimestamp,
			testUpdatedWhitelists: types.Whitelists{
				Origins:     map[types.Origin]struct{}{"https://portalgun.io": {}, "https://subdomain.example.com": {}, "https://www.example.com": {}},
				UserAgents:  map[types.UserAgent]struct{}{"Brave": {}, "Google Chrome": {}, "Mozilla Firefox": {}, "Netscape Navigator": {}, "Safari": {}},
				Blockchains: map[types.RelayChainID]struct{}{"0001": {}, "0002": {}, "003E": {}, "0056": {}},
				Contracts: map[types.RelayChainID]map[types.Contract]struct{}{
					"0001": {"0xtest_2f78db6436527729929aaf6c616361de0f7": {}, "0xtest_5fbfe3e9af3971dd833d26ba9b5c936f0be": {}},
					"0002": {"0xtest_1111117dc0aa78b770fa6a738034120c302": {}, "0xtest_a39b223fe8d0a0e5c4f27ead9083c756cc2": {}},
					"003E": {"0xtest_0a85d5af5bf1d1762f925bdaddc4201f984": {}, "0xtest_f958d2ee523a2206206994597c13d831ec7": {}},
					"0056": {"0xtest_00000f279d81a1d3cc75430faa017fa5a2e": {}, "0xtest_5068778dd592e39a122f4f5a5cf09c90fe2": {}},
				},
				Methods: map[types.RelayChainID]map[types.Method]struct{}{
					"0001": {"GET": {}, "POST": {}, "PUT": {}},
					"0002": {"DELETE": {}, "GET": {}, "POST": {}, "PUT": {}},
					"003E": {"GET": {}},
					"0056": {"GET": {}, "POST": {}},
				},
			},
			err: nil,
		},
		{
			name: "Should update a new PortalApp in the database with a new stripe subscription ID",
			updatePortalApp: types.UpdatePortalApp{
				StripeSubscriptionID: testdata.UpdatePortalAppStripeSubscriptionID.StripeSubscriptionID,
			},
			testUpdateTime: testdata.MockTimestamp,
			testUpdatedLegacyFields: types.LegacyFields{
				PlanType:             types.FreetierV0,
				DailyLimit:           250_000,
				StripeSubscriptionID: "update_stripe_subscription_id",
			},
			err: nil,
		},
		{
			name: "Should update a new PortalApp in the database with a new plan",
			updatePortalApp: types.UpdatePortalApp{
				PlanType: testdata.UpdatePortalAppPlan.PlanType,
			},
			testUpdateTime: testdata.MockTimestamp,
			testUpdatedLegacyFields: types.LegacyFields{
				PlanType:             types.FreetierV0,
				DailyLimit:           250_000,
				StripeSubscriptionID: "update_stripe_subscription_id",
			},
			err: nil,
		},
		{
			name: "Should update a new PortalApp in the database with another new plan",
			updatePortalApp: types.UpdatePortalApp{
				PlanType: testdata.UpdatePortalAppPlanTwo.PlanType,
			},
			testUpdateTime: testdata.MockTimestamp,
			testUpdatedLegacyFields: types.LegacyFields{
				PlanType:             types.TestPlan90k,
				DailyLimit:           90_000,
				StripeSubscriptionID: "update_stripe_subscription_id",
			},
			err: nil,
		},
		{
			name: "Should update a new PortalApp in the database with an Enterprise plan",
			updatePortalApp: types.UpdatePortalApp{
				PlanType: testdata.UpdatePortalAppEnterprisePlan.PlanType,
			},
			testUpdateTime: testdata.MockTimestamp,
			testUpdatedLegacyFields: types.LegacyFields{
				PlanType:             types.Enterprise,
				CustomLimit:          5_600_000,
				StripeSubscriptionID: "update_stripe_subscription_id",
			},
			err: nil,
		},
		{
			name: "Should fail if an invalid plan type provided",
			updatePortalApp: types.UpdatePortalApp{
				PlanType: types.PayPlanType("what_am_i_doing"),
			},
			err: fmt.Errorf(errPayPlanDoesntExist.Error(), "what_am_i_doing"),
		},
	}

	for i, test := range tests {
		ts.Run(test.name, func() {
			// Create new portal app for test case
			createApp := *testdata.TestUpdatePortalApp
			createApp.Name = fmt.Sprintf("test-update-portal-app-%d", i+1)
			createdPortalApp, err := ts.driver.WritePortalApp(context.Background(), createApp, testdata.TestCreatePortalAppAAT, testdata.MockTimestamp)
			ts.NoError(err)

			// Update created portal app for test case
			updateApp := test.updatePortalApp
			updateApp.AppID = createdPortalApp.ID
			err = ts.driver.UpdatePortalApp(context.Background(), updateApp, test.testUpdateTime)
			ts.Equal(test.err, err)

			if test.err == nil {
				// Get all portal apps from DB
				portalApps, err := ts.driver.ReadPortalApps(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				updatedPortalApp, ok := portalApps[createdPortalApp.ID]
				ts.True(ok)

				// Check update changes
				if test.testUpdatedName != "" {
					ts.Equal(test.testUpdatedName, updatedPortalApp.Name)
				} else {
					ts.Equal(createdPortalApp.Name, updatedPortalApp.Name)
				}

				if test.testUpdatedDescription != "" {
					ts.Equal(test.testUpdatedDescription, updatedPortalApp.Description)
				} else {
					ts.Equal(createdPortalApp.Description, updatedPortalApp.Description)
				}

				if test.testUpdatedAppEmoji != "" {
					ts.Equal(test.testUpdatedAppEmoji, updatedPortalApp.AppEmoji)
				} else {
					ts.Equal(createdPortalApp.AppEmoji, updatedPortalApp.AppEmoji)
				}

				if test.testUpdatedSettings.Environment != "" {
					ts.Equal(test.testUpdatedWhitelists, updatedPortalApp.Whitelists)
				} else {
					ts.Equal(createdPortalApp.Settings, updatedPortalApp.Settings)
				}

				if len(test.testUpdatedNotifications) != 0 {
					ts.Equal(test.testUpdatedNotifications, updatedPortalApp.Notifications)
				} else {
					ts.Equal(createdPortalApp.Notifications, updatedPortalApp.Notifications)
				}

				if len(test.testUpdatedWhitelists.Origins) != 0 {
					ts.Equal(test.testUpdatedWhitelists, updatedPortalApp.Whitelists)
				} else {
					ts.Equal(createdPortalApp.Whitelists, updatedPortalApp.Whitelists)
				}

				if test.updatePortalApp.PlanType != "" {
					ts.Equal(test.testUpdatedLegacyFields.PlanType, updatedPortalApp.LegacyFields.PlanType)
					if test.updatePortalApp.PlanType != types.Enterprise {
						dailyLimit, err := ts.driver.GetPlanDailyLimit(context.Background(), test.testUpdatedLegacyFields.PlanType)
						ts.NoError(err)
						ts.Equal(test.testUpdatedLegacyFields.DailyLimit, dailyLimit.Int32)
					}
				}
				if test.updatePortalApp.CustomLimit != 0 {
					ts.Equal(test.testUpdatedLegacyFields.CustomLimit, updatedPortalApp.LegacyFields.CustomLimit)
				}
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_UpdatePortalAppsFirstDateSurpassed() {
	tests := []struct {
		name   string
		update *types.UpdateFirstDateSurpassed
		err    error
	}{
		{
			name: "Should update multiple PortalApps in the database with a LegacyFields.FirstDateSurpassed timestamp ",
			update: &types.UpdateFirstDateSurpassed{
				PortalAppIDs:       []types.PortalAppID{testdata.PortalApps["test_app_1"].ID, testdata.PortalApps["test_app_2"].ID},
				FirstDateSurpassed: time.Date(2023, time.February, 14, 0, 0, 0, 0, time.UTC),
			},
			err: nil,
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			// Update FirstDateSurpassed fields for PortalApps
			err := ts.driver.UpdatePortalAppsFirstDateSurpassed(context.Background(), test.update)
			ts.Equal(test.err, err)

			// Check FirstDateSurpassed fields for PortalApps in the DB
			portalApps, err := ts.driver.ReadPortalApps(context.Background(), types.DriverOptions{})
			ts.Equal(test.err, err)
			for _, appID := range test.update.PortalAppIDs {
				portalApp, ok := portalApps[types.PortalAppID(appID)]
				ts.True(ok)
				ts.Equal(test.update.FirstDateSurpassed, portalApp.FirstDateSurpassed)
			}
		})
	}
}
