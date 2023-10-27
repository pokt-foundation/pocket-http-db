package testdata

import (
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/types"
)

/*
This file contains the Go struct equivalent to the seed test data for the test database container defined in `seed_test_db.sql`.

This mock data can be used anywhere that used the test database Docker container `pocketfoundation/test-portal-postgres`.

If/when adding new data to the `seed_test_db.sql` file it should also be updated here.
This means all E2E tests that use `pocketfoundation/test-portal-postgres` only need to be kept updated in one place.
*/

var MockTimestamp = time.Date(2022, time.November, 11, 11, 11, 11, 0, time.UTC)

func Users() map[types.UserID]*types.User {
	return map[types.UserID]*types.User{
		"user_1": {
			ID:               "user_1",
			Email:            "james.holden123@test.com",
			SignedUp:         true,
			IconURL:          "https://picsum.photos/200",
			UpdatesProduct:   true,
			UpdatesMarketing: false,
			BetaTester:       true,
			AuthProviders: map[types.AuthType]types.UserAuthProvider{
				types.AuthTypeAuth0Username: {
					ProviderUserID: "auth0|james_holden",
					Type:           types.AuthTypeAuth0Username,
					Provider:       "auth0",
					Federated:      false,
				},
				types.AuthTypeAuth0Github: {
					ProviderUserID: "github|james_holden",
					Type:           types.AuthTypeAuth0Github,
					Provider:       "auth0",
					Federated:      true,
				},
			},
			Permissions: map[types.AccountID]types.AccountPermissions{
				"account_1": types.AccountPermissions{
					Accepted: true,
					RoleName: types.RoleOwner,
					Permissions: []types.Permissions{
						types.PermReadEndpoint,
						types.PermWriteEndpoint,
						types.PermDeleteEndpoint,
						types.PermTransferEndpoint,
					},
				},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"user_2": {
			ID:               "user_2",
			Email:            "paul.atreides456@test.com",
			SignedUp:         true,
			IconURL:          "https://picsum.photos/200",
			UpdatesProduct:   false,
			UpdatesMarketing: true,
			BetaTester:       false,
			AuthProviders: map[types.AuthType]types.UserAuthProvider{
				types.AuthTypeAuth0Username: {
					ProviderUserID: "auth0|paul_atreides",
					Type:           types.AuthTypeAuth0Username,
					Provider:       "auth0",
					Federated:      false,
				},
				types.AuthTypeAuth0Github: {
					ProviderUserID: "github|paul_atreides",
					Type:           types.AuthTypeAuth0Github,
					Provider:       "auth0",
					Federated:      true,
				},
			},
			Permissions: map[types.AccountID]types.AccountPermissions{
				"account_1": types.AccountPermissions{
					Accepted: true,
					RoleName: types.RoleAdmin,
					Permissions: []types.Permissions{
						types.PermReadEndpoint,
						types.PermWriteEndpoint,
					},
				},
				"account_2": types.AccountPermissions{
					Accepted: true,
					RoleName: types.RoleMember,
					Permissions: []types.Permissions{
						types.PermReadEndpoint,
					},
				},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"user_3": {
			ID:               "user_3",
			Email:            "ellen.ripley789@test.com",
			SignedUp:         true,
			IconURL:          "https://picsum.photos/200",
			UpdatesProduct:   true,
			UpdatesMarketing: true,
			BetaTester:       false,
			AuthProviders: map[types.AuthType]types.UserAuthProvider{
				types.AuthTypeAuth0Username: {
					ProviderUserID: "auth0|ellen_ripley",
					Type:           types.AuthTypeAuth0Username,
					Provider:       "auth0",
					Federated:      false,
				},
			},
			Permissions: map[types.AccountID]types.AccountPermissions{
				"account_2": types.AccountPermissions{
					Accepted: true,
					RoleName: types.RoleOwner,
					Permissions: []types.Permissions{
						types.PermReadEndpoint,
						types.PermWriteEndpoint,
						types.PermDeleteEndpoint,
						types.PermTransferEndpoint,
					},
				},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"user_4": {
			ID:               "user_4",
			Email:            "ulfric.stormcloak123@test.com",
			SignedUp:         true,
			IconURL:          "https://picsum.photos/200",
			UpdatesProduct:   false,
			UpdatesMarketing: false,
			BetaTester:       true,
			AuthProviders: map[types.AuthType]types.UserAuthProvider{
				types.AuthTypeAuth0Username: {
					ProviderUserID: "auth0|ulfric_stormcloak",
					Type:           types.AuthTypeAuth0Username,
					Provider:       "auth0",
					Federated:      false,
				},
			},
			Permissions: map[types.AccountID]types.AccountPermissions{
				"account_2": types.AccountPermissions{
					Accepted: true,
					RoleName: types.RoleMember,
					Permissions: []types.Permissions{
						types.PermReadEndpoint,
					},
				},
				"account_4": types.AccountPermissions{
					Accepted: true,
					RoleName: types.RoleOwner,
					Permissions: []types.Permissions{
						types.PermReadEndpoint,
						types.PermWriteEndpoint,
						types.PermDeleteEndpoint,
						types.PermTransferEndpoint,
					},
				},
				"account_5": types.AccountPermissions{
					Accepted: true,
					RoleName: types.RoleOwner,
					Permissions: []types.Permissions{
						types.PermReadEndpoint,
						types.PermWriteEndpoint,
						types.PermDeleteEndpoint,
						types.PermTransferEndpoint,
					},
				},
			},

			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"user_5": {
			ID:               "user_5",
			Email:            "chrisjen.avasarala1@test.com",
			SignedUp:         true,
			IconURL:          "https://picsum.photos/200",
			UpdatesProduct:   true,
			UpdatesMarketing: false,
			BetaTester:       false,
			AuthProviders: map[types.AuthType]types.UserAuthProvider{
				types.AuthTypeAuth0Username: {
					ProviderUserID: "auth0|chrisjen_avasarala",
					Type:           types.AuthTypeAuth0Username,
					Provider:       "auth0",
					Federated:      false,
				},
			},
			Permissions: map[types.AccountID]types.AccountPermissions{
				"account_3": types.AccountPermissions{
					Accepted: true,
					RoleName: types.RoleOwner,
					Permissions: []types.Permissions{
						types.PermReadEndpoint,
						types.PermWriteEndpoint,
						types.PermDeleteEndpoint,
						types.PermTransferEndpoint,
					},
				},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"user_6": {
			ID:               "user_6",
			Email:            "amos.burton789@test.com",
			SignedUp:         true,
			IconURL:          "https://picsum.photos/200",
			UpdatesProduct:   false,
			UpdatesMarketing: true,
			BetaTester:       true,
			AuthProviders: map[types.AuthType]types.UserAuthProvider{
				types.AuthTypeAuth0Username: {
					ProviderUserID: "auth0|amos_burton",
					Type:           types.AuthTypeAuth0Username,
					Provider:       "auth0",
					Federated:      false,
				},
			},
			Permissions: map[types.AccountID]types.AccountPermissions{
				"account_3": types.AccountPermissions{
					Accepted: true,
					RoleName: types.RoleAdmin,
					Permissions: []types.Permissions{
						types.PermReadEndpoint,
						types.PermWriteEndpoint,
					},
				},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"user_7": {
			ID:               "user_7",
			Email:            "frodo.baggins123@test.com",
			SignedUp:         true,
			IconURL:          "https://picsum.photos/200",
			UpdatesProduct:   true,
			UpdatesMarketing: true,
			BetaTester:       true,
			AuthProviders: map[types.AuthType]types.UserAuthProvider{
				types.AuthTypeAuth0Username: {
					ProviderUserID: "auth0|frodo_baggins",
					Type:           types.AuthTypeAuth0Username,
					Provider:       "auth0",
					Federated:      false,
				},
			},
			Permissions: map[types.AccountID]types.AccountPermissions{
				"account_3": types.AccountPermissions{
					Accepted: true,
					RoleName: types.RoleMember,
					Permissions: []types.Permissions{
						types.PermReadEndpoint,
					},
				},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"user_8": {
			ID:               "user_8",
			Email:            "rick.deckard456@test.com",
			SignedUp:         true,
			IconURL:          "https://picsum.photos/200",
			UpdatesProduct:   true,
			UpdatesMarketing: true,
			BetaTester:       false,
			AuthProviders: map[types.AuthType]types.UserAuthProvider{
				types.AuthTypeAuth0Username: {
					ProviderUserID: "auth0|rick_deckard",
					Type:           types.AuthTypeAuth0Username,
					Provider:       "auth0",
					Federated:      false,
				},
			},
			Permissions: map[types.AccountID]types.AccountPermissions{
				"account_1": types.AccountPermissions{
					Accepted: false,
					RoleName: types.RoleAdmin,
					Permissions: []types.Permissions{
						types.PermReadEndpoint,
						types.PermWriteEndpoint,
					},
				},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"user_9": {
			ID:               "user_9",
			Email:            "tyrion.lannister789@test.com",
			SignedUp:         true,
			IconURL:          "https://picsum.photos/200",
			UpdatesProduct:   true,
			UpdatesMarketing: true,
			BetaTester:       false,
			AuthProviders: map[types.AuthType]types.UserAuthProvider{
				types.AuthTypeAuth0Username: {
					ProviderUserID: "auth0|tyrion_lannister",
					Type:           types.AuthTypeAuth0Username,
					Provider:       "auth0",
					Federated:      false,
				},
			},
			Permissions: map[types.AccountID]types.AccountPermissions{
				"account_2": types.AccountPermissions{
					Accepted: false,
					RoleName: types.RoleMember,
					Permissions: []types.Permissions{
						types.PermReadEndpoint,
					},
				},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"user_10": {
			ID:               "user_10",
			Email:            "daenerys.targaryen123@test.com",
			SignedUp:         false,
			IconURL:          "https://picsum.photos/200",
			UpdatesProduct:   false,
			UpdatesMarketing: false,
			BetaTester:       true,
			AuthProviders:    map[types.AuthType]types.UserAuthProvider{},
			Permissions: map[types.AccountID]types.AccountPermissions{
				"account_3": types.AccountPermissions{
					Accepted: false,
					RoleName: types.RoleMember,
					Permissions: []types.Permissions{
						types.PermReadEndpoint,
					},
				},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"user_11": {
			ID:               "user_11",
			Email:            "bernard.marx@test.com",
			SignedUp:         true,
			IconURL:          "https://picsum.photos/200",
			UpdatesProduct:   true,
			UpdatesMarketing: true,
			BetaTester:       false,
			AuthProviders: map[types.AuthType]types.UserAuthProvider{
				types.AuthTypeAuth0Username: {
					ProviderUserID: "auth0|bernard_marx",
					Type:           types.AuthTypeAuth0Username,
					Provider:       "auth0",
					Federated:      false,
				},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"user_12": {
			ID:               "user_12",
			Email:            "george.foreman@test.com",
			SignedUp:         false,
			IconURL:          "https://picsum.photos/200",
			UpdatesProduct:   false,
			UpdatesMarketing: true,
			BetaTester:       false,
			AuthProviders:    map[types.AuthType]types.UserAuthProvider{},
			CreatedAt:        MockTimestamp,
			UpdatedAt:        MockTimestamp,
		},
	}
}

func UserIDs() map[types.ProviderUserID]types.UserID {
	return map[types.ProviderUserID]types.UserID{
		"auth0|amos_burton":        "user_6",
		"auth0|bernard_marx":       "user_11",
		"auth0|chrisjen_avasarala": "user_5",
		"auth0|ellen_ripley":       "user_3",
		"auth0|frodo_baggins":      "user_7",
		"auth0|james_holden":       "user_1",
		"auth0|paul_atreides":      "user_2",
		"auth0|rick_deckard":       "user_8",
		"auth0|tyrion_lannister":   "user_9",
		"auth0|ulfric_stormcloak":  "user_4",
		"github|james_holden":      "user_1",
		"github|paul_atreides":     "user_2",
	}
}

func AccountUsers() map[int]types.AccountUser {
	return map[int]types.AccountUser{
		1: {ID: "user_1", Email: "james.holden123@test.com", RoleName: types.RoleOwner, Accepted: true},
		2: {ID: "user_2", Email: "paul.atreides456@test.com", RoleName: types.RoleAdmin, Accepted: true},
		3: {ID: "user_3", Email: "ellen.ripley789@test.com", RoleName: types.RoleOwner, Accepted: true},
		4: {ID: "user_3", Email: "ellen.ripley789@test.com", RoleName: types.RoleOwner, Accepted: true},
		5: {ID: "user_5", Email: "chrisjen.avasarala1@test.com", RoleName: types.RoleOwner, Accepted: true},
		6: {ID: "user_6", Email: "amos.burton789@test.com", RoleName: types.RoleAdmin, Accepted: true},
		7: {ID: "user_7", Email: "frodo.baggins123@test.com", RoleName: types.RoleMember, Accepted: true},
		8: {ID: "user_8", Email: "rick.deckard456@test.com", RoleName: types.RoleAdmin, Accepted: true},
		9: {ID: "user_9", Email: "tyrion.lannister789@test.com", RoleName: types.RoleMember, Accepted: false},
		// Paul is an admin of Account 1 as well as a member of Account 2
		10: {ID: "user_2", Email: "paul.atreides456@test.com", RoleName: types.RoleMember, Accepted: true},
		// Ulfric is an admin of Account 2 as well as the owner of Accounts 4 and 5
		11: {ID: "user_4", Email: "ulfric.stormcloak123@test.com", RoleName: types.RoleOwner, Accepted: true},
		// Daenerys has not signed up with an auth provider yet and is a member of Account 3
		12: {ID: "user_10", Email: "daenerys.targaryen123@test.com", RoleName: types.RoleMember, Accepted: false},
		// Bernard is an existing user and is used to create a new AccountUser row
		13: {ID: "user_11", Email: "bernard.marx@test.com"},
		// Winston has not signed up yet and is used to create a new AccountUser row
		14: {Email: "winston.smith@test.com"},
	}
}

type AccountOptions struct {
	IncludePortalApps bool
	IncludePlans      bool
	AssignUsers       bool
}

func Accounts(opts ...AccountOptions) map[types.AccountID]*types.Account {
	accounts := map[types.AccountID]*types.Account{
		"account_1": {
			ID:       "account_1",
			Name:     "The Brave Voyager",
			IconURL:  "https://picsum.photos/200",
			PlanType: types.PayPlanType("basic_plan"),
			Integrations: types.AccountIntegrations{
				StripeSubscriptionID: "stripe_id_1",
				CovalentAPIKeyFree:   "covalent_api_key_1",
			},
			Notifications: map[types.NotificationType]types.AccountNotification{
				types.NotificationTypeEmail: {
					Type:        types.NotificationTypeEmail,
					Active:      true,
					Destination: "james.holden123@test.com",
					Trigger:     "trigger123",
					Events: map[types.NotificationEvent]bool{
						types.NotificationEventFull:          true,
						types.NotificationEventQuarter:       true,
						types.NotificationEventThreeQuarters: true,
					},
				},
			},
			Users:                  map[types.UserID]types.AccountUser{},
			PartnerChainIDs:        map[types.RelayChainID]struct{}{"0001": {}, "0053": {}},
			PartnerThroughputLimit: 2_000,
			PartnerAppLimit:        1,
			PortalApps:             map[types.PortalAppID]*types.PortalApp{},
			CreatedAt:              MockTimestamp,
			UpdatedAt:              MockTimestamp,
		},
		"account_2": {
			ID:       "account_2",
			Name:     "Dragonborn Explorer",
			IconURL:  "https://picsum.photos/200",
			PlanType: types.PayPlanType("pro_plan"),
			Integrations: types.AccountIntegrations{
				StripeSubscriptionID: "stripe_id_2",
				CovalentAPIKeyFree:   "covalent_api_key_2",
			},
			Notifications: map[types.NotificationType]types.AccountNotification{
				types.NotificationTypeEmail: {
					Type:        types.NotificationTypeEmail,
					Active:      true,
					Destination: "ellen.ripley789@test.com",
					Trigger:     "trigger456",
					Events: map[types.NotificationEvent]bool{
						types.NotificationEventHalf: true,
						types.NotificationEventFull: true,
					},
				},
			},
			Users:                  map[types.UserID]types.AccountUser{},
			PartnerChainIDs:        map[types.RelayChainID]struct{}{"0001": {}, "0053": {}, "0021": {}, "0064": {}},
			PartnerThroughputLimit: 5_000,
			PartnerAppLimit:        3,
			PortalApps:             map[types.PortalAppID]*types.PortalApp{},
			CreatedAt:              MockTimestamp,
			UpdatedAt:              MockTimestamp,
		},
		"account_3": {
			ID:       "account_3",
			Name:     "Fellowship Startup",
			IconURL:  "https://picsum.photos/200",
			PlanType: types.PayPlanType("startup_plan"),
			Integrations: types.AccountIntegrations{
				StripeSubscriptionID: "stripe_id_3",
				CovalentAPIKeyFree:   "covalent_api_key_3",
			},
			Notifications: map[types.NotificationType]types.AccountNotification{
				types.NotificationTypeEmail: {
					Type:        types.NotificationTypeEmail,
					Active:      true,
					Destination: "chrisjen.avasarala1@test.com",
					Trigger:     "trigger789",
					Events: map[types.NotificationEvent]bool{
						types.NotificationEventHalf:    true,
						types.NotificationEventFull:    true,
						types.NotificationEventQuarter: true,
					},
				},
			},
			Users:                  map[types.UserID]types.AccountUser{},
			PartnerChainIDs:        map[types.RelayChainID]struct{}{"0001": {}, "0053": {}, "0064": {}, "0034": {}},
			PartnerThroughputLimit: 1_000,
			PartnerAppLimit:        2,
			PortalApps:             map[types.PortalAppID]*types.PortalApp{},
			CreatedAt:              MockTimestamp,
			UpdatedAt:              MockTimestamp,
		},
		"account_4": {
			ID:       "account_4",
			Name:     "Iron Throne Enterprise",
			IconURL:  "https://picsum.photos/200",
			PlanType: types.PayPlanType("enterprise_plan"),
			Integrations: types.AccountIntegrations{
				StripeSubscriptionID: "stripe_id_4",
				CovalentAPIKeyFree:   "covalent_api_key_4",
			},
			Users:                  map[types.UserID]types.AccountUser{},
			PartnerChainIDs:        map[types.RelayChainID]struct{}{"0001": {}},
			PartnerThroughputLimit: 1_000,
			PartnerAppLimit:        2,
			PortalApps:             map[types.PortalAppID]*types.PortalApp{},
			CreatedAt:              MockTimestamp,
			UpdatedAt:              MockTimestamp,
		},
		"account_5": {
			ID:                     "account_5",
			Name:                   "Nomad Wanderer",
			IconURL:                "https://picsum.photos/200",
			PlanType:               types.PayPlanType("basic_plan"),
			Users:                  map[types.UserID]types.AccountUser{},
			PartnerChainIDs:        map[types.RelayChainID]struct{}{"0006": {}, "0040": {}},
			PartnerThroughputLimit: 6_000,
			PartnerAppLimit:        1,
			PortalApps:             map[types.PortalAppID]*types.PortalApp{},
			CreatedAt:              MockTimestamp,
			UpdatedAt:              MockTimestamp,
		},
	}

	if len(opts) > 0 {
		if opts[0].IncludePortalApps {
			for _, app := range PortalApps() {
				accounts[app.AccountID].PortalApps[app.ID] = app
			}
		}
		if opts[0].IncludePlans {
			for _, account := range accounts {
				account.Plan = PayPlans()[account.PlanType]
			}
		}
		if opts[0].AssignUsers {
			usersForAccountIDs := map[types.AccountID]map[types.UserID]*types.User{
				"account_1": {
					"user_1": Users()["user_1"],
					"user_2": Users()["user_2"],
					"user_8": Users()["user_8"],
				},
				"account_2": {
					"user_3": Users()["user_3"],
					"user_4": Users()["user_4"],
					"user_9": Users()["user_9"],
					"user_2": Users()["user_2"],
				},
				"account_3": {
					"user_5":  Users()["user_5"],
					"user_6":  Users()["user_6"],
					"user_7":  Users()["user_7"],
					"user_10": Users()["user_10"],
				},
				"account_4": {
					"user_4": Users()["user_4"],
				},
				"account_5": {
					"user_4": Users()["user_4"],
				},
			}

			for _, account := range accounts {
				if usersForAccount, ok := usersForAccountIDs[account.ID]; ok {
					account.AssignAccountUsers(usersForAccount)
				}
			}
		}
	}

	return accounts
}

func PortalApps() map[types.PortalAppID]*types.PortalApp {
	return map[types.PortalAppID]*types.PortalApp{
		"test_app_1": {
			ID:          "test_app_1",
			AccountID:   "account_1",
			Name:        "pokt_app_123",
			AppEmoji:    "1F336",
			Description: "Embark on an interstellar journey with our powerful application.",
			AATs: map[types.ProtocolAppID]types.AAT{
				"test_protocol_app_1": {
					ID:              "test_protocol_app_1",
					Address:         "test_34715cae753e67c75fbb340442e7de8e",
					PublicKey:       "test_34715cae753e67c75fbb340442e7de8e",
					ClientPublicKey: "test_89a3af6a587aec02cfade6f5000424c2",
					PrivateKey:      "",
					Signature:       "test_1dc39a2e5a84a35bf030969a0b3231f7",
					Version:         "0.0.1",
				},
			},
			Settings: types.Settings{
				Environment:       types.EnvironmentProduction,
				SecretKey:         "test_40f482d91a5ef2300ebb4e2308c",
				SecretKeyRequired: true,
				MonthlyRelayLimit: 2_500_000,
				FavoritedChainIDs: map[types.RelayChainID]struct{}{"0001": {}, "0053": {}},
			},
			Whitelists: types.Whitelists{
				Origins:     map[types.Origin]struct{}{"https://test.com": {}},
				UserAgents:  map[types.UserAgent]struct{}{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)": {}},
				Blockchains: map[types.RelayChainID]struct{}{"0053": {}},
				Contracts: map[types.RelayChainID]map[types.Contract]struct{}{
					"0001": {"0x1234567890abcdef": {}},
				},
				Methods: map[types.RelayChainID]map[types.Method]struct{}{
					"0001": {"GET": {}},
				},
			},
			FirstDateSurpassed: MockTimestamp,
			CreatedAt:          MockTimestamp,
			UpdatedAt:          MockTimestamp,
		},
		"test_app_2": {
			ID:          "test_app_2",
			AccountID:   "account_2",
			Name:        "pokt_app_456",
			AppEmoji:    "1F336",
			Description: "Unveil the mysteries of the multiverse with our enchanting application.",
			AATs: map[types.ProtocolAppID]types.AAT{
				"test_protocol_app_2": {
					ID:              "test_protocol_app_2",
					Address:         "test_8237c72345f12d1b1a8b64a1a7f66fa4",
					PublicKey:       "test_8237c72345f12d1b1a8b64a1a7f66fa4",
					ClientPublicKey: "test_04c71d90a92f40416b6f1d7d8af17e02",
					PrivateKey:      "",
					Signature:       "test_f48d33b30ddaf60a1e5bb50d2ba8da5a",
					Version:         "0.0.1",
				},
			},
			Settings: types.Settings{
				Environment:       types.EnvironmentProduction,
				SecretKey:         "test_9c9e3b193cfba5348f93bb2f3e3fb794",
				SecretKeyRequired: false,
				MonthlyRelayLimit: 1_500_000,
				FavoritedChainIDs: map[types.RelayChainID]struct{}{"0021": {}, "0064": {}},
			},
			Whitelists: types.Whitelists{
				Origins:     map[types.Origin]struct{}{"https://example.com": {}},
				UserAgents:  map[types.UserAgent]struct{}{"Mozilla/5.0 (Linux; Android 10; SM-A205U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36": {}},
				Blockchains: map[types.RelayChainID]struct{}{"0021": {}},
				Contracts: map[types.RelayChainID]map[types.Contract]struct{}{
					"0064": {"0x0987654321abcdef": {}},
				},
				Methods: map[types.RelayChainID]map[types.Method]struct{}{
					"0064": {"POST": {}},
				},
			},
			FirstDateSurpassed: MockTimestamp,
			CreatedAt:          MockTimestamp,
			UpdatedAt:          MockTimestamp,
		},
		"test_app_3": {
			ID:          "test_app_3",
			AccountID:   "account_3",
			Name:        "pokt_app_789",
			AppEmoji:    "1F336",
			Description: "Harness the power of a connected world with our revolutionary application.",
			AATs: map[types.ProtocolAppID]types.AAT{
				"test_protocol_app_3": {
					ID:              "test_protocol_app_3",
					Address:         "test_b5e07928fc80083c13ad0201b81bae9b",
					PublicKey:       "test_f608500e4fe3e09014fe2411b4a560b5",
					ClientPublicKey: "test_328a9cf1b35085eeaa669aa858f6fba9",
					PrivateKey:      "",
					Signature:       "test_c3cd8be16ba32e24dd49fdb0247fc9b8",
					Version:         "0.0.1",
				},
				"test_protocol_app_4": {
					ID:              "test_protocol_app_4",
					Address:         "test_eb2e5bcba557cfe8fa76fd7fff54f9d1",
					PublicKey:       "test_f6a5d8690ecb669865bd752b7796a920",
					ClientPublicKey: "test_6ee5ea553408f0895923fd1569dc5072",
					PrivateKey:      "",
					Signature:       "test_cf05cf9bb26111c548e88fb6157af708",
					Version:         "0.0.1",
				},
			},
			Settings: types.Settings{
				Environment:       types.EnvironmentProduction,
				SecretKey:         "test_9f48b13e2bc5fd31ab367841f11495c1",
				SecretKeyRequired: false,
				MonthlyRelayLimit: 4_500_000,
				FavoritedChainIDs: map[types.RelayChainID]struct{}{"0001": {}, "0034": {}},
			},
			FirstDateSurpassed: MockTimestamp,
			CreatedAt:          MockTimestamp,
			UpdatedAt:          MockTimestamp,
		},
	}
}

// This is used for testing the converted values of the above Portal Apps
// Does not exist in the test DB seed data as it is never stored in the DB.
func PortalAppLites() map[types.PortalAppID]*types.PortalAppLite {
	return map[types.PortalAppID]*types.PortalAppLite{
		"test_app_1": {
			ID: "test_app_1",
			PublicKeys: []types.PortalAppPublicKey{
				"test_34715cae753e67c75fbb340442e7de8e",
			},
			Settings: types.SettingsLite{
				SecretKey:         "test_40f482d91a5ef2300ebb4e2308c",
				SecretKeyRequired: true,
			},
			Whitelists: types.Whitelists{
				Origins: map[types.Origin]struct{}{
					"https://test.com": {},
				},
				UserAgents: map[types.UserAgent]struct{}{
					"Mozilla/5.0 (Windows NT 10.0; Win64; x64)": {},
				},
				Blockchains: map[types.RelayChainID]struct{}{
					"0053": {},
				},
				Contracts: map[types.RelayChainID]map[types.Contract]struct{}{
					"0001": {
						"0x1234567890abcdef": {},
					},
				},
				Methods: map[types.RelayChainID]map[types.Method]struct{}{
					"0001": {
						"GET": {},
					},
				},
			},
			Plan: types.PlanLite{
				PlanType:        types.PayPlanType("basic_plan"),
				ChainIDs:        map[types.RelayChainID]struct{}{"0001": {}, "0053": {}},
				ThroughputLimit: 5_000,
			},
		},
		"test_app_2": {
			ID: "test_app_2",
			PublicKeys: []types.PortalAppPublicKey{
				"test_8237c72345f12d1b1a8b64a1a7f66fa4",
			},
			Whitelists: types.Whitelists{
				Origins: map[types.Origin]struct{}{
					"https://example.com": {},
				},
				UserAgents: map[types.UserAgent]struct{}{
					"Mozilla/5.0 (Linux; Android 10; SM-A205U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36": {},
				},
				Blockchains: map[types.RelayChainID]struct{}{
					"0021": {},
				},
				Contracts: map[types.RelayChainID]map[types.Contract]struct{}{
					"0064": {
						"0x0987654321abcdef": {},
					},
				},
				Methods: map[types.RelayChainID]map[types.Method]struct{}{
					"0064": {
						"POST": {},
					},
				},
			},
			Plan: types.PlanLite{
				PlanType:        types.PayPlanType("pro_plan"),
				ChainIDs:        map[types.RelayChainID]struct{}{"0001": {}, "0053": {}, "0021": {}, "0064": {}},
				ThroughputLimit: 10_000,
			},
		},
		"test_app_3": {
			ID: "test_app_3",
			PublicKeys: []types.PortalAppPublicKey{
				"test_f608500e4fe3e09014fe2411b4a560b5",
				"test_f6a5d8690ecb669865bd752b7796a920",
			},
			Plan: types.PlanLite{
				PlanType:        types.PayPlanType("startup_plan"),
				ChainIDs:        map[types.RelayChainID]struct{}{"0001": {}, "0053": {}, "0064": {}, "0034": {}},
				ThroughputLimit: 1_000,
			},
		},
	}
}

func PayPlans() map[types.PayPlanType]*types.Plan {
	return map[types.PayPlanType]*types.Plan{
		types.PayPlanType("FREETIER_V0"): {
			Type:              types.PayPlanType("FREETIER_V0"),
			Name:              "Free Tier",
			Description:       "Ideal for small projects and testing",
			ChainIDs:          map[types.RelayChainID]struct{}{"0001": {}, "0053": {}},
			MonthlyRelayLimit: 5_000_000,
			ThroughputLimit:   5_000,
			AppLimit:          5,
			LegacyDailyLimit:  250_000,
			CreatedAt:         MockTimestamp,
		},
		types.PayPlanType("PAY_AS_YOU_GO_V0"): {
			Type:              types.PayPlanType("PAY_AS_YOU_GO_V0"),
			Name:              "Pay As You Go",
			Description:       "Pay only for what you use",
			ChainIDs:          map[types.RelayChainID]struct{}{"0001": {}, "0053": {}},
			MonthlyRelayLimit: 10_000_000,
			ThroughputLimit:   10_000,
			AppLimit:          5,
			LegacyDailyLimit:  0,
			CreatedAt:         MockTimestamp,
		},
		types.PayPlanType("ENTERPRISE"): {
			Type:              types.PayPlanType("ENTERPRISE"),
			Name:              "Enterprise",
			Description:       "Premium plan for large businesses",
			ChainIDs:          map[types.RelayChainID]struct{}{"0001": {}, "0053": {}},
			MonthlyRelayLimit: 10_000_000,
			ThroughputLimit:   10_000,
			AppLimit:          5,
			LegacyDailyLimit:  0,
			CreatedAt:         MockTimestamp,
		},
		types.PayPlanType("TEST_PLAN_V0"): {
			Type:              types.PayPlanType("TEST_PLAN_V0"),
			Name:              "Test Plan",
			Description:       "For testing purposes",
			ChainIDs:          map[types.RelayChainID]struct{}{"0001": {}, "0053": {}},
			MonthlyRelayLimit: 10_000_000,
			ThroughputLimit:   10_000,
			AppLimit:          5,
			LegacyDailyLimit:  0,
			CreatedAt:         MockTimestamp,
		},
		types.PayPlanType("TEST_PLAN_10K"): {
			Type:              types.PayPlanType("TEST_PLAN_10K"),
			Name:              "Test Plan 10K",
			Description:       "Test plan with 10K daily limit for testing purposes",
			ChainIDs:          map[types.RelayChainID]struct{}{"0001": {}, "0053": {}},
			MonthlyRelayLimit: 10_000_000,
			ThroughputLimit:   10_000,
			AppLimit:          5,
			LegacyDailyLimit:  10_000,
			CreatedAt:         MockTimestamp,
		},
		types.PayPlanType("TEST_PLAN_90K"): {
			Type:              types.PayPlanType("TEST_PLAN_90K"),
			Name:              "Test Plan 90K",
			Description:       "Test plan with 90K daily limit for testing purposes",
			ChainIDs:          map[types.RelayChainID]struct{}{"0001": {}, "0053": {}},
			MonthlyRelayLimit: 90_000_000,
			ThroughputLimit:   10_000,
			AppLimit:          5,
			LegacyDailyLimit:  90_000,
			CreatedAt:         MockTimestamp,
		},
		types.PayPlanType("basic_plan"): {
			Type:              types.PayPlanType("basic_plan"),
			Name:              "Basic Plan",
			Description:       "Basic plan for testing purposes",
			ChainIDs:          map[types.RelayChainID]struct{}{"0001": {}, "0053": {}},
			MonthlyRelayLimit: 5_000_000,
			ThroughputLimit:   5_000,
			AppLimit:          2,
			LegacyDailyLimit:  1_000,
			CreatedAt:         MockTimestamp,
		},
		types.PayPlanType("pro_plan"): {
			Type:              types.PayPlanType("pro_plan"),
			Name:              "Pro Plan",
			Description:       "Pro plan for testing purposes",
			ChainIDs:          map[types.RelayChainID]struct{}{"0001": {}, "0053": {}, "0021": {}, "0064": {}},
			MonthlyRelayLimit: 10_000_000,
			ThroughputLimit:   10_000,
			AppLimit:          5,
			LegacyDailyLimit:  5_000,
			CreatedAt:         MockTimestamp,
		},
		types.PayPlanType("enterprise_plan"): {
			Type:              types.PayPlanType("enterprise_plan"),
			Name:              "Enterprise Plan",
			Description:       "Enterprise plan for testing purposes",
			ChainIDs:          map[types.RelayChainID]struct{}{"0001": {}, "0053": {}, "0021": {}, "0064": {}, "0034": {}},
			MonthlyRelayLimit: 20_000_000,
			ThroughputLimit:   20_000,
			AppLimit:          10,
			LegacyDailyLimit:  10_000,
			CreatedAt:         MockTimestamp,
		},
		types.PayPlanType("developer_plan"): {
			Type:              types.PayPlanType("developer_plan"),
			Name:              "Developer Plan",
			Description:       "Developer plan for testing purposes",
			ChainIDs:          map[types.RelayChainID]struct{}{"0001": {}, "0053": {}, "0021": {}, "0034": {}},
			MonthlyRelayLimit: 500_000,
			ThroughputLimit:   500,
			AppLimit:          1,
			LegacyDailyLimit:  100,
			CreatedAt:         MockTimestamp,
		},
		types.PayPlanType("startup_plan"): {
			Type:              types.PayPlanType("startup_plan"),
			Name:              "Startup Plan",
			Description:       "Startup plan for testing purposes",
			ChainIDs:          map[types.RelayChainID]struct{}{"0001": {}, "0053": {}, "0064": {}, "0034": {}},
			MonthlyRelayLimit: 1_000_000,
			ThroughputLimit:   1_000,
			AppLimit:          5,
			LegacyDailyLimit:  500,
			CreatedAt:         MockTimestamp,
		},
	}
}

type ChainOptions struct {
	IncludeGigastakeApps bool
}

func Chains(opts ...ChainOptions) map[types.RelayChainID]*types.Chain {
	chains := map[types.RelayChainID]*types.Chain{
		"0001": {
			ID:            "0001",
			IconURL:       "https://picsum.photos/200",
			Blockchain:    "pokt-mainnet",
			Description:   "Pocket Network Mainnet",
			EnforceResult: "JSON",
			Path:          "/v1/query/height",
			Ticker:        "POKT",
			Active:        true,
			Altruists: map[types.AltruistURL]types.Altruist{
				"https://altruist-0001.com:1234": {
					URL:      "https://altruist-0001.com:1234",
					AuthType: types.ChainAuthTypeBasicAuth,
					Auth:     "test_pocket:auth123456",
				},
			},
			Checks: map[types.ChainCheckType]types.Check{
				types.ChainCheckTypeSync: {
					Type:      types.ChainCheckTypeSync,
					Payload:   `{"id":1,"jsonrpc":"2.0","method":"query"}`,
					ResultKey: "result.sync_info",
					Allowance: 1,
				},
			},
			// DEPRECATED - TODO remove when move to only store aliases is complete
			AliasDomains: map[types.ChainAlias][]types.ChainDomain{
				"pokt-mainnet": {"pokt-rpc.gateway.pokt.network"},
			},
			Aliases: map[types.ChainAlias]struct{}{
				"pokt-mainnet": {},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"0053": {
			ID:             "0053",
			IconURL:        "https://picsum.photos/200",
			Blockchain:     "optimism-mainnet",
			Description:    "Optimism Mainnet",
			EnforceResult:  "JSON",
			Ticker:         "OP",
			LogLimitBlocks: 100000,
			RequestTimeout: 0,
			Active:         true,
			Altruists: map[types.AltruistURL]types.Altruist{
				"https://altruist-0053.com:1234": {
					URL:      "https://altruist-0053.com:1234",
					AuthType: types.ChainAuthTypeBasicAuth,
					Auth:     "test_pocket:auth123456",
				},
			},
			Checks: map[types.ChainCheckType]types.Check{
				types.ChainCheckTypeSync: {
					Type:      types.ChainCheckTypeSync,
					Payload:   `{"id":1,"jsonrpc":"2.0","method":"eth_blockNumber","params":[]}`,
					ResultKey: "result",
					Allowance: 2,
				},
			},
			// DEPRECATED - TODO remove when move to only store aliases is complete
			AliasDomains: map[types.ChainAlias][]types.ChainDomain{
				"optimism-mainnet": {"op-rpc.gateway.pokt.network"},
			},
			Aliases: map[types.ChainAlias]struct{}{
				"optimism-mainnet": {},
				"optimism-rpc":     {},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"0021": {
			ID:             "0021",
			IconURL:        "https://picsum.photos/200",
			Blockchain:     "eth-mainnet",
			Description:    "Ethereum Mainnet",
			EnforceResult:  "JSON",
			Ticker:         "ETH",
			LogLimitBlocks: 100000,
			Active:         true,
			Altruists: map[types.AltruistURL]types.Altruist{
				"https://altruist-0021.com:1234": {
					URL:      "https://altruist-0021.com:1234",
					AuthType: types.ChainAuthTypeBasicAuth,
					Auth:     "test_pocket:auth123456",
				},
			},
			Checks: map[types.ChainCheckType]types.Check{
				types.ChainCheckTypeSync: {
					Type:      types.ChainCheckTypeSync,
					Payload:   `{"id":1,"jsonrpc":"2.0","method":"eth_blockNumber","params":[]}`,
					ResultKey: "result",
					Allowance: 5,
				},
				types.ChainCheckTypeChain: {
					Type:       types.ChainCheckTypeChain,
					Payload:    `{"method":"eth_chainId","id":1,"jsonrpc":"2.0"}`,
					ResultKey:  "id",
					EVMChainID: 1,
				},
			},
			// DEPRECATED - TODO remove when move to only store aliases is complete
			AliasDomains: map[types.ChainAlias][]types.ChainDomain{
				"eth-mainnet": {"eth-rpc.gateway.pokt.network"},
			},
			Aliases: map[types.ChainAlias]struct{}{
				"eth-mainnet": {},
				"eth-rpc":     {},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"0064": {
			ID:             "0064",
			IconURL:        "https://picsum.photos/200",
			Blockchain:     "sui-testnet",
			Description:    "Sui Testnet",
			EnforceResult:  "JSON",
			Ticker:         "SUI-TESTNET",
			LogLimitBlocks: 100000,
			RequestTimeout: 60000,
			Active:         false,
			Altruists: map[types.AltruistURL]types.Altruist{
				"https://altruist-0064.com:1234": {
					URL:      "https://altruist-0064.com:1234",
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
				"sui-testnet": {},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"0040": {
			ID:            "0040",
			IconURL:       "https://picsum.photos/200",
			Blockchain:    "harmony-0",
			Description:   "Harmony Shard 0",
			EnforceResult: "JSON",
			Ticker:        "HMY",
			Active:        true,
			Altruists: map[types.AltruistURL]types.Altruist{
				"https://altruist-0040.com:1234": {
					URL:      "https://altruist-0040.com:1234",
					AuthType: types.ChainAuthTypeBasicAuth,
					Auth:     "test_pocket:auth123456",
				},
			},
			Checks: map[types.ChainCheckType]types.Check{
				types.ChainCheckTypeSync: {
					Type:      types.ChainCheckTypeSync,
					Payload:   `{"id":1,"jsonrpc":"2.0","method":"hmy_blockNumber","params":[]}`,
					ResultKey: "result",
					Allowance: 8,
				},
			},
			// DEPRECATED - TODO remove when move to only store aliases is complete
			AliasDomains: map[types.ChainAlias][]types.ChainDomain{
				"harmony-0": {"hmy-rpc.gateway.pokt.network"},
			},
			Aliases: map[types.ChainAlias]struct{}{
				"harmony-0": {},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
	}

	if len(opts) > 0 && opts[0].IncludeGigastakeApps {
		for _, gigastakeApp := range GigastakeApps() {
			for chainID := range gigastakeApp.ChainIDs {
				if chains[chainID].GigastakeApps == nil {
					chains[chainID].GigastakeApps = make(map[types.GigastakeAppID]*types.GigastakeApp)
				}
				chains[chainID].GigastakeApps[gigastakeApp.ID] = gigastakeApp
			}
		}
	}

	return chains
}

func GigastakeApps() map[types.GigastakeAppID]*types.GigastakeApp {
	return map[types.GigastakeAppID]*types.GigastakeApp{
		"test_gigastake_app_1": {
			ID:              "test_gigastake_app_1",
			ChainIDs:        map[types.RelayChainID]struct{}{"0001": {}},
			Name:            "pokt_gigastake",
			Address:         "test_8d4f6a5b0c6e9f1db12c1f662e5ec8c5",
			PublicKey:       types.GigastakeAppPublicKey("test_37a0e8437f5149dc98a9a5b207efc2d0"),
			ClientPublicKey: "test_65c29f0cc82e418b81a528a0c0682a9f",
			Signature:       "test_f22651fb566346fca30b605e5f46e3ca",
			Version:         "0.0.1",
			CreatedAt:       MockTimestamp,
			UpdatedAt:       MockTimestamp,
		},
		"test_gigastake_app_2": {
			ID:              "test_gigastake_app_2",
			ChainIDs:        map[types.RelayChainID]struct{}{"0053": {}},
			Name:            "optimism_gigastake",
			Address:         "test_5c60d434db4e42d2b5d2ea6eeb8933c4",
			PublicKey:       types.GigastakeAppPublicKey("test_a7e28f8d716541a0a332a5dc6b7e4e6e"),
			ClientPublicKey: "test_ba4e53dada8f4f939048e56dc8f88f37",
			Signature:       "test_52e991c26da841bc882ad3a3ee9ee964",
			Version:         "0.0.1",
			CreatedAt:       MockTimestamp,
			UpdatedAt:       MockTimestamp,
		},
		"test_gigastake_app_3": {
			ID:              "test_gigastake_app_3",
			ChainIDs:        map[types.RelayChainID]struct{}{"0040": {}},
			Name:            "harmony_gigastake",
			Address:         "test_e570c841d5cd4f6197e0428ed7c517fd",
			PublicKey:       types.GigastakeAppPublicKey("test_4f805bbbf96c4a649efc3f4f95616f2e"),
			ClientPublicKey: "test_789f9d6adcc846f1a079bf68237b5f5c",
			Signature:       "test_01eac46efc9242a2be73879f1d09f1dc",
			Version:         "0.0.1",
			CreatedAt:       MockTimestamp,
			UpdatedAt:       MockTimestamp,
		},
	}
}

func GlobalBlockedContracts() types.GlobalBlockedContracts {
	return types.GlobalBlockedContracts{
		BlockedAddresses: map[types.BlockedAddress]struct{}{
			"0xtest_6789abcdef0123456789abcdef01234567":   {},
			"0xtest_f0123456789abcdef0123456789abcdef01":  {},
			"0xtest_cdef0123456789abcdef0123456789abcdef": {},
			"0xtest_56789abcdef0123456789abcdef01234567":  {},
			"0xtest_789abcdef0123456789abcdef0123456789":  {},
		},
	}
}

var (
	/* ----- Create Data ----- */

	TestCreateAccount = &types.Account{
		ID:         "create_account_id",
		Name:       "Protogen Corp",
		IconURL:    "https://picsum.photos/200",
		PlanType:   types.PayPlanType("developer_plan"),
		PortalApps: map[types.PortalAppID]*types.PortalApp{},
		CreatedAt:  MockTimestamp,
		UpdatedAt:  MockTimestamp,
	}

	TestCreateUser = types.CreateUser{
		Email:          "commander.data@example.com",
		ProviderUserID: "auth0|commander_data",
	}

	TestCreatePortalAppByOwner = &types.PortalApp{
		AccountID:   "account_4",
		Name:        "create_pokt_app_1",
		Description: "Embark on a journey across the enchanting realm of Middle-earth.",
		AppEmoji:    "1F336",
		Settings: types.Settings{
			Environment:       types.EnvironmentProduction,
			SecretKey:         "test_3e3fb7949c9e3b193cfba5348f93bb2f",
			SecretKeyRequired: true,
		},
		FirstDateSurpassed: MockTimestamp,
		CreatedAt:          MockTimestamp,
		UpdatedAt:          MockTimestamp,
	}

	TestCreatePortalAppByNonOwner = &types.PortalApp{
		AccountID:   "account_1",
		Name:        "create_pokt_app_2",
		Description: "Travel the mystic sands of Arrakis",
		AppEmoji:    "1F336",
		Settings: types.Settings{
			Environment:       types.EnvironmentProduction,
			SecretKey:         "test_83756e546134df7933bfd748b27bde51",
			SecretKeyRequired: true,
		},
		FirstDateSurpassed: MockTimestamp,
		CreatedAt:          MockTimestamp,
		UpdatedAt:          MockTimestamp,
	}

	TestCreatePortalAppAAT = types.AAT{
		Address:         "test_d24a1810570804902ffdf8328d338ccc663",
		PublicKey:       "test_7137982b1e7dc54f55bb4f658724bc1eece31aa39058dc6d45c79dca2ba",
		ClientPublicKey: "test_ba7b889d6c96dfa1b3838672439ce8dfd4451bc3cf6ce0e942e347e3762",
		PrivateKey:      "test_e42879ebe7c1dabc291240b8c1830656ea21b94336fbc044a5b6c70e876e3c34132501323e769eccfb00f8e88ccfc6ee2808a7b8a8e97dd646b2fd26eb300719d87668ec7f8698f38bbe02734fc8c50b72d917b7047af8106d2624b0eb036570d17aef2bafc489befd2811e2774227ec7b70ed2140c5d6f2dd7b5af7502d8a96cd137de68fe0f9ef2541df11eaaeb7997aae357ab025ac269c3",
		Signature:       "test_15f627b3e92893431e8a59711c6d169a724f53f7dbad89110bd21a4c34315786e0bc9f943bdb73e3bd660cf2bf5a10a71b85a3c9062cb3413c29367ea72",
		Version:         "0.0.1",
	}

	TestCreateGigastakeApp = types.GigastakeApp{
		ID:              "test_create_gigastake_1",
		ChainIDs:        map[types.RelayChainID]struct{}{"0001": {}},
		Name:            "pokt_create_gigastake",
		Address:         "test_f9b26cd01e10b2054ef4f69e64edbdbfb0d",
		PublicKey:       types.GigastakeAppPublicKey("test_bc795ac9141ec75166ae11f3217bd04a1a3a62492efc8579e34819cfbdf"),
		ClientPublicKey: "test_94389a8890dc6ab49a74a37b7140413b4c6bb0640689cc2dc628d49b17b",
		Signature:       "test_ed37b11a1c4456a3e9b4ce0d69fc3ea962887924c6cfca969170d6c4a552bf13c8cd73649fd27929730b4229a834b090bf7a0ed7f506a98f315f961fda4",
		Version:         "0.0.1",
	}

	TestCreateChain = &types.Chain{
		ID:            "0006",
		IconURL:       "https://picsum.photos/200",
		Blockchain:    "solana-mainnet",
		Description:   "Solana",
		EnforceResult: "JSON",
		Ticker:        "SOL",
		Altruists: map[types.AltruistURL]types.Altruist{
			"https://test-rpc.solana-1.io:1234": {
				URL:      "https://test-rpc.solana-1.io:1234",
				AuthType: types.ChainAuthTypeBasicAuth,
				Auth:     "test_pocket:auth123456",
			},
			"https://test-rpc.solana-2.io:1234": {
				URL:      "https://test-rpc.solana-2.io:1234",
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
			types.ChainCheckTypeChain: {
				Type:       types.ChainCheckTypeChain,
				Payload:    `{"id":1,"jsonrpc":"2.0","method":"getChain"}`,
				ResultKey:  "chain",
				EVMChainID: 5,
			},
		},
		Aliases: map[types.ChainAlias]struct{}{
			"solana-mainnet": {},
			"sol-mainnet":    {},
		},
		// DEPRECATED - TODO remove when move to only store aliases is complete
		AliasDomains: map[types.ChainAlias][]types.ChainDomain{
			"sol-mainnet": {"solana-rpc.gateway.pokt.network"},
		},
		CreatedAt: MockTimestamp,
		UpdatedAt: MockTimestamp,
	}

	TestCreateNewChainInput = types.NewChainInput{
		Chain: &types.Chain{
			ID:            "0007",
			Blockchain:    "cardano-mainnet",
			Description:   "Cardano",
			EnforceResult: "JSON",
			Ticker:        "ADA",
			Altruists: map[types.AltruistURL]types.Altruist{
				"https://test-rpc.cardano-1.io:1234": {
					URL:      "https://test-rpc.cardano-1.io:1234",
					AuthType: types.ChainAuthTypeBasicAuth,
					Auth:     "test_cardano:auth123456",
				},
				"https://test-rpc.cardano-2.io:1234": {
					URL:      "https://test-rpc.cardano-2.io:1234",
					AuthType: types.ChainAuthTypeBasicAuth,
					Auth:     "test_cardano:auth123456",
				},
			},
			Checks: map[types.ChainCheckType]types.Check{
				types.ChainCheckTypeSync: {
					Type:      types.ChainCheckTypeSync,
					Payload:   `{"id":1,"jsonrpc":"2.0","method":"getSync"}`,
					ResultKey: "sync",
					Allowance: 2,
				},
				types.ChainCheckTypeChain: {
					Type:       types.ChainCheckTypeChain,
					Payload:    `{"id":1,"jsonrpc":"2.0","method":"getChain"}`,
					ResultKey:  "chain",
					EVMChainID: 7,
				},
			},
			Aliases: map[types.ChainAlias]struct{}{
				"cardano-mainnet": {},
				"ada-mainnet":     {},
			},
			// DEPRECATED - TODO remove when move to only store aliases is complete
			AliasDomains: map[types.ChainAlias][]types.ChainDomain{
				"ada-mainnet": {"cardano-rpc.gateway.pokt.network"},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		GigastakeApps: []*types.GigastakeApp{
			{
				ID:              "test_create_gigastake_ada",
				Name:            "pokt_create_gigastake_ada",
				Address:         "test_c3e9d2a1a3214bc7b364f51362a8a8e4",
				PublicKey:       "test_8a4b6d3f48274d8988d0f5b4866efce1",
				ClientPublicKey: "test_c3f9e2a7b2f74bc799d0f3b4962efce1",
				Signature:       "test_4ef8e2a7b4f74bc989d0f3b4962efce1",
				Version:         "0.0.1",
			},
		},
	}

	/* ----- Update Data ----- */

	// TestUpdatePortalApp app used to test updates of PortalApps
	TestUpdatePortalApp = &types.PortalApp{
		ID:          "test_app_update_b03ca84c",
		AccountID:   "account_1",
		Name:        "", // name set in test
		Description: "Embark on a journey across the solar system with Pur'n'Kleen",
		AppEmoji:    "1F336",
		AATs:        map[types.ProtocolAppID]types.AAT{"test_protocol_app_1": {}},
		Settings: types.Settings{
			Environment:       types.EnvironmentProduction,
			SecretKey:         "test_849c1397586f9fb6f902576120d0d10f",
			SecretKeyRequired: true,
		},
		FirstDateSurpassed: MockTimestamp,
		CreatedAt:          MockTimestamp,
		UpdatedAt:          MockTimestamp,
	}

	UpdatePortalAppName        = "portal-app-updated"
	UpdatePortalAppDescription = "Updating the application name like the shifting sands of Arrakis."
	UpdatePortalAppEmoji       = types.AppEmoji("1F336")
	UpdatePortalAppSettings    = &types.UpdateAppSettings{
		Environment:       types.EnvironmentProduction,
		SecretKey:         "test_9d07c8a96ad53e7c288b0e86f37c5680",
		SecretKeyRequired: true,
		MonthlyRelayLimit: 2_500_000,
		FavoritedChainIDs: []string{"0003", "0009", "00H3"},
	}
	UpdatePortalAppWhitelists = &types.WhitelistsObject{
		AppWhitelists: [3]types.ApplicationWhitelists{
			{Type: "origins", Values: []string{"https://portalgun.io", "https://subdomain.example.com", "https://www.example.com"}},
			{Type: "userAgents", Values: []string{"Brave", "Google Chrome", "Mozilla Firefox", "Netscape Navigator", "Safari"}},
			{Type: "blockchains", Values: []string{"0001", "0002", "003E", "0056"}},
		},
		ChainWhitelists: [2]types.ChainWhitelists{
			{Type: "contracts", Values: []types.ChainIDWhitelists{
				{ChainID: "0001", Values: []string{"0xtest_2f78db6436527729929aaf6c616361de0f7", "0xtest_5fbfe3e9af3971dd833d26ba9b5c936f0be"}},
				{ChainID: "0002", Values: []string{"0xtest_1111117dc0aa78b770fa6a738034120c302", "0xtest_a39b223fe8d0a0e5c4f27ead9083c756cc2"}},
				{ChainID: "003E", Values: []string{"0xtest_0a85d5af5bf1d1762f925bdaddc4201f984", "0xtest_f958d2ee523a2206206994597c13d831ec7"}},
				{ChainID: "0056", Values: []string{"0xtest_00000f279d81a1d3cc75430faa017fa5a2e", "0xtest_5068778dd592e39a122f4f5a5cf09c90fe2"}},
			},
			},
			{Type: "methods", Values: []types.ChainIDWhitelists{
				{ChainID: "0001", Values: []string{"GET", "POST", "PUT"}},
				{ChainID: "0002", Values: []string{"DELETE", "GET", "POST", "PUT"}},
				{ChainID: "003E", Values: []string{"GET"}},
				{ChainID: "0056", Values: []string{"GET", "POST"}},
			},
			},
		},
	}

	UpdateChainOne = types.UpdateChain{
		ID:            "0001",
		IconURL:       newString("https://picsum.photos/246"),
		Blockchain:    newChainAlias("mainnet-NEW"),
		Description:   newString("Pocket Network Mainnet Update"),
		EnforceResult: newString("JSON"),
		Path:          newString("/v1/query/height/wow"),
		Ticker:        newString("POKT-123"),
		Active:        newBool(true),
		Altruists: &map[types.AltruistURL]types.Altruist{
			"https://altruist-0001.com:1234": {
				URL:      "https://altruist-0001.com:1234",
				AuthType: types.ChainAuthTypeBasicAuth,
				Auth:     "test_pocket:auth123456",
			},
			"https://altruist-0001-2.com:1234": {
				URL:      "https://altruist-0001-2.com:1234",
				AuthType: types.ChainAuthTypeBasicAuth,
				Auth:     "test_pocket:auth123456",
			},
			"https://altruist-0001-3.com:1234": {
				URL:      "https://altruist-0001-3.com:1234",
				AuthType: types.ChainAuthTypeBasicAuth,
				Auth:     "test_pocket:auth123456",
			},
		},
		Checks: &map[types.ChainCheckType]types.Check{
			types.ChainCheckTypeSync: {
				Type:      types.ChainCheckTypeSync,
				Payload:   `{"id":1,"jsonrpc":"2.0","method":"query"}`,
				ResultKey: "result.sync_info",
				Allowance: 1,
			},
			types.ChainCheckTypeChain: {
				Type:       types.ChainCheckTypeChain,
				Payload:    `{"id":1,"jsonrpc":"2.0","method":"chain"}`,
				ResultKey:  "result.sync_info",
				EVMChainID: 3,
			},
			types.ChainCheckTypeMerge: {
				Type:      types.ChainCheckTypeMerge,
				Payload:   `{"id":1,"jsonrpc":"2.0","method":"merge"}`,
				ResultKey: "result.sync_info",
			},
		},
		Aliases: &map[types.ChainAlias]struct{}{
			"pokt-mainnet": {},
			"pokt-main":    {},
		},
	}
	UpdateChainTwo = types.UpdateChain{
		ID:            "0001",
		IconURL:       newString("https://picsum.photos/200"),
		Blockchain:    newChainAlias("mainnet-ULTRA"),
		Description:   newString("Pocket Network Mainnet Original"),
		EnforceResult: newString("JSON"),
		Path:          newString("/v1/query/height/wow"),
		Ticker:        newString("POKT-456"),
		Active:        newBool(true),
		Altruists: &map[types.AltruistURL]types.Altruist{
			"https://altruist-0001-3.com:1234": {
				URL:      "https://altruist-0001-3.com:1234",
				AuthType: types.ChainAuthTypeBasicAuth,
				Auth:     "test_pocket:auth123456",
			},
		},
		Checks: &map[types.ChainCheckType]types.Check{
			types.ChainCheckTypeSync: {
				Type:      types.ChainCheckTypeSync,
				Payload:   `{"id":1,"jsonrpc":"2.0","method":"query"}`,
				ResultKey: "result.sync_info",
				Allowance: 1,
			},
		},
		Aliases: &map[types.ChainAlias]struct{}{
			"pokt-mainnet": {},
			"pokt-main":    {},
			"pokt-rpc-2":   {},
		},
	}
	UpdateChainThree = types.UpdateChain{
		ID:            "0001",
		Blockchain:    newChainAlias("mainnet-ULTRA"),
		Description:   newString("Pocket Network Mainnet Original"),
		EnforceResult: newString("JSON"),
		Path:          newString("/v1/query/height/wow"),
		Ticker:        newString("POKT-456"),
		Active:        newBool(true),
	}

	// UpdateChainNotExists used to test updating a Chain that doesn't exist
	UpdateChainNotExists      = types.UpdateChain{ID: "0073"}
	UpdateChainInvalidIconURL = types.UpdateChain{ID: "0001", IconURL: newString("what_is_555")}
	UpdateChainInvalidURL     = types.UpdateChain{ID: "0073",
		Altruists: &map[types.AltruistURL]types.Altruist{
			"htz:/bad-domain2": {URL: "htz:/bad-domain2"},
		},
	}
	UpdateChainInvalidDomain = types.UpdateChain{ID: "0073",
		AliasDomains: &map[types.ChainAlias][]types.ChainDomain{
			"sol-mainnet": {"im-not-a-domain"},
		},
	}

	UpdateUserOne = types.UpdateUser{
		ID:               "user_5",
		IconURL:          newString("https://picsum.photos/227"),
		UpdatesProduct:   newBool(false),
		UpdatesMarketing: newBool(false),
		BetaTester:       newBool(true),
	}
	UpdateUserTwo = types.UpdateUser{
		ID:               "user_5",
		IconURL:          newString("https://picsum.photos/200"),
		UpdatesProduct:   newBool(true),
		UpdatesMarketing: newBool(false),
		BetaTester:       newBool(false),
	}
	UpdateUserInvalidURL = types.UpdateUser{
		ID:      "user_5",
		IconURL: newString("i-am-not-a-url"),
	}
)

func newString(s string) *string {
	return &s
}

func newChainAlias(a types.ChainAlias) *types.ChainAlias {
	return &a
}

func newBool(b bool) *bool {
	return &b
}
