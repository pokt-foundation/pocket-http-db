package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GetOwner(t *testing.T) {
	tests := []struct {
		name        string
		account     Account
		expected    AccountUser
		expectedErr error
	}{
		{
			name:     "Should return owner when account has an owner",
			account:  testAccount,
			expected: testAccount.AccountUsers["user_1"],
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner, err := test.account.GetOwner()
			assert.Equal(t, test.expected, owner)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func Test_GetOwnerID(t *testing.T) {
	tests := []struct {
		name        string
		account     Account
		expected    UserID
		expectedErr error
	}{
		{
			name:     "Should return owner when account has an owner",
			account:  testAccount,
			expected: "user_1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner, err := test.account.GetOwnerID()
			assert.Equal(t, test.expected, owner)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func Test_GetAcceptedPortalApps(t *testing.T) {
	tests := []struct {
		name     string
		account  Account
		userID   UserID
		accepted bool
		expected map[PortalAppID]*PortalApp
	}{
		{
			name:     "Should return accepted apps when accepted is true",
			account:  testAccount,
			userID:   "user_1",
			accepted: true,
			expected: map[PortalAppID]*PortalApp{
				"test_app_1": testAccount.PortalApps["test_app_1"],
			},
		},
		{
			name:     "Should return non-accepted apps when accepted is false",
			account:  testAccount,
			userID:   "user_8",
			accepted: false,
			expected: map[PortalAppID]*PortalApp{
				"test_app_1": testAccount.PortalApps["test_app_1"],
			},
		},
		{
			name:     "Should return empty map when no apps are accepted",
			account:  testAccount,
			userID:   "user_9",
			accepted: true,
			expected: map[PortalAppID]*PortalApp{},
		},
		{
			name:     "Should return empty map when all apps are accepted",
			account:  testAccount,
			userID:   "user_2",
			accepted: false,
			expected: map[PortalAppID]*PortalApp{},
		},
		{
			name:     "Should return empty map when userID does not exist",
			account:  testAccount,
			userID:   "user_nonexistent",
			accepted: true,
			expected: map[PortalAppID]*PortalApp{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.account.GetAcceptedPortalApps(test.userID, test.accepted)
			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_HasUserAcceptedInvite(t *testing.T) {
	tests := []struct {
		name       string
		account    Account
		userID     UserID
		appID      PortalAppID
		expected   bool
		expectedOk bool
	}{
		{
			name:       "Should return true if user has accepted the PortalApp",
			account:    testAccount,
			userID:     "user_1",
			appID:      "test_app_1",
			expected:   true,
			expectedOk: true,
		},
		{
			name:       "Should return false if user has not accepted the PortalApp",
			account:    testAccount,
			userID:     "user_8",
			appID:      "test_app_1",
			expected:   false,
			expectedOk: true,
		},
		{
			name:       "Should return false if user is not in AccountUsers map",
			account:    testAccount,
			userID:     "user_nonexistent",
			appID:      "test_app_1",
			expected:   false,
			expectedOk: false,
		},
		{
			name:       "Should return false if PortalApp is not in PortalAppsAccepted map",
			account:    testAccount,
			userID:     "user_1",
			appID:      "test_app_nonexistent",
			expected:   false,
			expectedOk: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, actualOk := test.account.HasUserAcceptedInvite(test.userID, test.appID)
			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedOk, actualOk)
		})
	}
}

var testAccount = Account{
	ID:       "account_1",
	Name:     "The Brave Voyager",
	IconURL:  "https://picsum.photos/200",
	PlanType: PayPlanType("basic_plan"),
	AccountUsers: map[UserID]AccountUser{
		"user_1": {
			AccountID: "",
			UserID:    "user_1",
			Email:     "james.holden123@test.com",
			Owner:     true,
			PortalAppRoles: map[PortalAppID]RoleName{
				"test_app_1": RoleOwner,
			},
			PortalAppsAccepted: map[PortalAppID]bool{
				"test_app_1": true,
			},
		},
		"user_2": {
			AccountID: "",
			UserID:    "user_2",
			Email:     "paul.atreides456@test.com",
			PortalAppRoles: map[PortalAppID]RoleName{
				"test_app_2": RoleAdmin,
			},
			PortalAppsAccepted: map[PortalAppID]bool{
				"test_app_2": true,
			},
		},
		"user_8": {
			AccountID: "",
			UserID:    "user_8",
			Email:     "rick.deckard456@test.com",
			PortalAppRoles: map[PortalAppID]RoleName{
				"test_app_1": RoleAdmin,
			},
			PortalAppsAccepted: map[PortalAppID]bool{
				"test_app_1": false,
			},
		},
		"user_9": {
			AccountID: "",
			UserID:    "user_9",
			Email:     "tyrion.lannister789@test.com",
			PortalAppRoles: map[PortalAppID]RoleName{
				"test_app_2": RoleMember,
			},
			PortalAppsAccepted: map[PortalAppID]bool{
				"test_app_2": false,
			},
		},
	},
	PortalApps: map[PortalAppID]*PortalApp{
		"test_app_1": {
			ID:          "test_app_1",
			AccountID:   "account_1",
			Name:        "pokt_app_123",
			AppEmoji:    "1F336",
			Description: "Embark on an interstellar journey with our powerful application.",
			AATs: map[ProtocolAppID]AAT{
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
			Settings: Settings{
				Environment:       EnvironmentProduction,
				SecretKey:         "test_40f482d91a5ef2300ebb4e2308c",
				SecretKeyRequired: true,
				MonthlyRelayLimit: 2_500_000,
				FavoritedChainIDs: map[RelayChainID]struct{}{"0001": {}, "0053": {}},
			},
			Whitelists: Whitelists{
				Origins:     map[Origin]struct{}{"https://test.com": {}},
				UserAgents:  map[UserAgent]struct{}{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)": {}},
				Blockchains: map[RelayChainID]struct{}{"0053": {}},
				Contracts: map[RelayChainID]map[Contract]struct{}{
					"0001": {"0x1234567890abcdef": {}},
				},
				Methods: map[RelayChainID]map[Method]struct{}{
					"0001": {"GET": {}},
				},
			},
			Notifications: map[NotificationType]AppNotification{
				NotificationTypeEmail: {
					Type:        NotificationTypeEmail,
					Active:      true,
					Destination: "test@test.com",
					Trigger:     "trigger123",
					Events: map[NotificationEvent]bool{
						NotificationEventFull:          true,
						NotificationEventQuarter:       true,
						NotificationEventThreeQuarters: true,
					},
				},
			},
			FirstDateSurpassed: timestamp,
			CreatedAt:          timestamp,
			UpdatedAt:          timestamp,
			// TODO remove legacy fields when migration to V2 schema complete
			LegacyFields: LegacyFields{
				PlanType:             FreetierV0,
				DailyLimit:           250_000,
				CustomLimit:          0,
				RequestTimeout:       5_000,
				StripeSubscriptionID: "stripe_id_1",
			},
			PortalAppUsers: map[UserID]PortalAppUser{
				"user_1": {
					ID:       "user_1",
					Email:    "james.holden123@test.com",
					RoleName: RoleOwner,
					Accepted: true,
				},
				"user_8": {
					ID:       "user_8",
					Email:    "rick.deckard456@test.com",
					RoleName: RoleAdmin,
					Accepted: false,
				},
			},
		},
		"test_app_2": {
			ID:          "test_app_2",
			AccountID:   "account_2",
			Name:        "pokt_app_456",
			AppEmoji:    "1F336",
			Description: "Unveil the mysteries of the multiverse with our enchanting application.",
			AATs: map[ProtocolAppID]AAT{
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
			Settings: Settings{
				Environment:       EnvironmentProduction,
				SecretKey:         "test_9c9e3b193cfba5348f93bb2f3e3fb794",
				SecretKeyRequired: false,
				MonthlyRelayLimit: 1_500_000,
				FavoritedChainIDs: map[RelayChainID]struct{}{"0021": {}, "0064": {}},
			},
			Whitelists: Whitelists{
				Origins:     map[Origin]struct{}{"https://example.com": {}},
				UserAgents:  map[UserAgent]struct{}{"Mozilla/5.0 (Linux; Android 10; SM-A205U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36": {}},
				Blockchains: map[RelayChainID]struct{}{"0021": {}},
				Contracts: map[RelayChainID]map[Contract]struct{}{
					"0064": {"0x0987654321abcdef": {}},
				},
				Methods: map[RelayChainID]map[Method]struct{}{
					"0064": {"POST": {}},
				},
			},
			Notifications: map[NotificationType]AppNotification{
				NotificationTypeEmail: {
					Type:        NotificationTypeEmail,
					Active:      true,
					Destination: "email@pokt.network",
					Trigger:     "trigger456",
					Events: map[NotificationEvent]bool{
						NotificationEventHalf: true,
						NotificationEventFull: true,
					},
				},
			},
			FirstDateSurpassed: timestamp,
			CreatedAt:          timestamp,
			UpdatedAt:          timestamp,
			// TODO remove legacy fields when migration to V2 schema complete
			LegacyFields: LegacyFields{
				PlanType:             PayAsYouGoV0,
				DailyLimit:           0,
				CustomLimit:          0,
				RequestTimeout:       10_000,
				StripeSubscriptionID: "stripe_id_2",
			},
			PortalAppUsers: map[UserID]PortalAppUser{
				"user_2": {
					ID:       "user_3",
					Email:    "ellen.ripley789@test.com",
					RoleName: RoleOwner,
					Accepted: true,
				},
				"user_9": {
					ID:       "user_9",
					Email:    "tyrion.lannister789@test.com",
					RoleName: RoleMember,
					Accepted: false,
				},
			},
		},
	},
	Integrations: AccountIntegrations{
		CovalentAPIKeyFree: "covalent_api_key_1",
	},
	PartnerChainIDs:        map[RelayChainID]struct{}{"0001": {}, "0053": {}},
	PartnerThroughputLimit: 2_000,
	PartnerAppLimit:        1,
	CreatedAt:              timestamp,
	UpdatedAt:              timestamp,
}

var timestamp = time.Date(2022, time.November, 11, 11, 11, 11, 0, time.UTC)
