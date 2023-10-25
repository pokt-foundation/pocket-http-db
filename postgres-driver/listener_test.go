package postgresdriver

import (
	"context"
	"testing"
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/testdata"
	"github.com/pokt-foundation/portal-http-db/v2/types"
	"github.com/stretchr/testify/require"
)

func Test_PGDriver_Listen(t *testing.T) {
	tests := []struct {
		name                  string
		content               types.SavedOnDB
		expectedNotifications map[types.Table]*types.Notification
		wantPanic             bool
	}{
		{
			name:    "Should process Account notifications",
			content: testdata.Accounts["account_4"],
			expectedNotifications: map[types.Table]*types.Notification{
				types.TableAccounts: {
					Table:  types.TableAccounts,
					Action: types.ActionInsert,
					Data: &types.Account{
						ID:                     "account_4",
						PlanType:               "enterprise_plan",
						PartnerChainIDs:        map[types.RelayChainID]struct{}{"0001": {}},
						PartnerThroughputLimit: 1_000,
						PartnerAppLimit:        2,
						CreatedAt:              testdata.MockTimestamp,
						UpdatedAt:              testdata.MockTimestamp,
					},
				},
				types.TableAccountUserAccess: {
					Table:  types.TableAccountUserAccess,
					Action: types.ActionUpdate,
					Data: &types.AccountUser{
						AccountID: "account_4",
						ID:        "user_4",
						RoleName:  types.RoleOwner,
						Accepted:  true,
					},
				},
			},
		},
		{
			name:    "Should process PortalApp notifications",
			content: testdata.PortalApps["test_app_1"],
			expectedNotifications: map[types.Table]*types.Notification{
				types.TablePortalApps: {
					Table:  types.TablePortalApps,
					Action: types.ActionInsert,
					Data: &types.PortalApp{
						ID:                 "test_app_1",
						AccountID:          "account_1",
						Name:               "pokt_app_123",
						FirstDateSurpassed: testdata.MockTimestamp,
						CreatedAt:          testdata.MockTimestamp,
						UpdatedAt:          testdata.MockTimestamp,
						Deleted:            false,
						LegacyFields: types.LegacyFields{
							StripeSubscriptionID: "stripe_id_1",
						},
					},
				},
				types.TablePortalAppAATs: {
					Table:  types.TablePortalAppAATs,
					Action: types.ActionUpdate,
					Data: &types.AAT{
						ID:              "test_protocol_app_1",
						AppID:           "test_app_1",
						Address:         "test_34715cae753e67c75fbb340442e7de8e",
						PublicKey:       "test_34715cae753e67c75fbb340442e7de8e",
						ClientPublicKey: "test_89a3af6a587aec02cfade6f5000424c2",
						PrivateKey:      "",
						Signature:       "test_1dc39a2e5a84a35bf030969a0b3231f7",
						Version:         "0.0.1",
					},
				},
				types.TableAppSettings: {
					Table:  types.TableAppSettings,
					Action: types.ActionUpdate,
					Data: &types.Settings{
						AppID:             "test_app_1",
						Environment:       "production",
						SecretKey:         "test_40f482d91a5ef2300ebb4e2308c",
						SecretKeyRequired: true,
						MonthlyRelayLimit: 2_500_000,
						FavoritedChainIDs: map[types.RelayChainID]struct{}{"0001": {}, "0053": {}},
					},
				},
				types.TableAppWhitelists: {
					Table:  types.TableAppWhitelists,
					Action: types.ActionUpdate,
					Data: &types.Whitelist{
						AppID: "test_app_1",
						Type:  "blockchains",
						Value: "0053",
					},
				},
				types.TableAppNotifications: {
					Table:  types.TableAppNotifications,
					Action: types.ActionUpdate,
					Data: &types.AppNotification{
						AppID:       "test_app_1",
						Type:        types.NotificationTypeEmail,
						Active:      true,
						Destination: "test@test.com",
						Trigger:     "trigger123",
						Events:      map[types.NotificationEvent]bool{"full": true, "quarter": true, "threeQuarters": true},
					},
				},
			},
		},
		{
			name:    "Should process GigastakeApp notifications",
			content: testdata.GigastakeApps["test_gigastake_app_1"],
			expectedNotifications: map[types.Table]*types.Notification{
				types.TableGigastakeApps: {
					Table:  types.TableGigastakeApps,
					Action: types.ActionInsert,
					Data: &types.GigastakeApp{
						ID:              "test_gigastake_app_1",
						Name:            "pokt_gigastake",
						Address:         "test_8d4f6a5b0c6e9f1db12c1f662e5ec8c5",
						PublicKey:       "test_37a0e8437f5149dc98a9a5b207efc2d0",
						ClientPublicKey: "test_65c29f0cc82e418b81a528a0c0682a9f",
						Signature:       "test_f22651fb566346fca30b605e5f46e3ca",
						Version:         "0.0.1",
						CreatedAt:       testdata.MockTimestamp,
						UpdatedAt:       testdata.MockTimestamp,
						Deleted:         false,
					},
				},
				types.TableChainGigastakeApps: {
					Table:  types.TableChainGigastakeApps,
					Action: types.ActionUpdate,
					Data: &types.ChainGigastakeApp{
						ChainID:        "0001",
						GigastakeAppID: "test_gigastake_app_1",
					},
				},
			},
		},
		{
			name:    "Should process Chain notifications",
			content: testdata.Chains["0001"],
			expectedNotifications: map[types.Table]*types.Notification{
				types.TableChains: {
					Table:  types.TableChains,
					Action: types.ActionInsert,
					Data: &types.Chain{
						ID:            "0001",
						Blockchain:    "pokt-mainnet",
						Description:   "Pocket Network Mainnet",
						EnforceResult: "JSON",
						Path:          "/v1/query/height",
						Ticker:        "POKT",
						Active:        true,
						CreatedAt:     time.Date(2022, time.November, 11, 11, 11, 11, 0, time.UTC),
						UpdatedAt:     time.Date(2022, time.November, 11, 11, 11, 11, 0, time.UTC),
					},
				},
				types.TableChainAltruists: {
					Table:  types.TableChainAltruists,
					Action: types.ActionUpdate,
					Data: &types.Altruist{
						ChainID:  "0001",
						URL:      "https://altruist-0001.com:1234",
						Auth:     "test_pocket:auth123456",
						AuthType: types.ChainAuthTypeBasicAuth,
					},
				},
				types.TableChainChecks: {
					Table:  types.TableChainChecks,
					Action: types.ActionUpdate,
					Data: &types.Check{
						Type:      types.ChainCheckTypeSync,
						ChainID:   "0001",
						Payload:   "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"query\"}",
						ResultKey: "result.sync_info",
						Allowance: 1,
					},
				},
				types.TableChainAliases: {
					Table:  types.TableChainAliases,
					Action: types.ActionUpdate,
					Data: &types.Alias{
						ChainID: "0001",
						Alias:   "pokt-mainnet",
					},
				},
				types.TableChainAliasDomains: {
					Table:  types.TableChainAliasDomains,
					Action: types.ActionUpdate,
					Data: &types.AliasDomains{
						ChainID: "0001",
						Alias:   "pokt-mainnet",
						Domains: []types.ChainDomain{"pokt-rpc.gateway.pokt.network"},
					},
				},
			},
		},
		{
			name:                  "Should return error for unsupported content",
			content:               &types.User{},
			expectedNotifications: map[types.Table]*types.Notification{},
			wantPanic:             true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != test.wantPanic {
					t.Errorf("recover = %v, wantPanic = %v", r, test.wantPanic)
				}
			}()

			c := require.New(t)

			channel := make(chan *types.Notification, 32)

			listenerMock := NewListenerMock(channel)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err := listenerMock.Listen(ctx)
			c.NoError(err)

			driver, errCh, _ := NewPostgresDriver(nil, listenerMock, channel)

			go func() {
				for err := range errCh {
					t.Errorf("error in listener: %v", err)
				}
			}()

			listenerMock.MockEvent(types.ActionInsert, types.ActionUpdate, test.content)

			<-time.After(1 * time.Second)
			driver.CloseListener()

			notificationsMap := make(map[types.Table]*types.Notification)

			for n := range driver.NotificationChannel() {
				notificationsMap[n.Table] = n
			}

			c.Equal(test.expectedNotifications, notificationsMap)
		})
	}

}
