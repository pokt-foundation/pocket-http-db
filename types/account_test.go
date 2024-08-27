package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var timestamp = time.Date(2022, time.November, 11, 11, 11, 11, 0, time.UTC)

var testAccount = Account{
	ID:       "account_1",
	Name:     "The Brave Voyager",
	IconURL:  "https://picsum.photos/200",
	PlanType: PlanFree,
	Plan: &Plan{
		AppLimit: 2,
	},
	Integrations: AccountIntegrations{
		StripeSubscriptionID: "stripe_id_1",
		CovalentAPIKeyFree:   "covalent_api_key_1",
	},
	Notifications: map[NotificationType]AccountNotification{
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
	Users: map[UserID]AccountUser{
		"user_1": {
			AccountID: "",
			ID:        "user_1",
			Email:     "james.holden123@test.com",
			RoleName:  RoleOwner,
			Accepted:  true,
		},
		"user_2": {
			AccountID: "",
			ID:        "user_2",
			Email:     "paul.atreides456@test.com",
			RoleName:  RoleAdmin,
			Accepted:  true,
		},
		"user_8": {
			AccountID: "",
			ID:        "user_8",
			Email:     "rick.deckard456@test.com",
			RoleName:  RoleAdmin,
			Accepted:  false,
		},
		"user_9": {
			AccountID: "",
			ID:        "user_9",
			Email:     "tyrion.lannister789@test.com",
			RoleName:  RoleMember,
			Accepted:  false,
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
				Environment:       EnvProduction,
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
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
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
				Environment:       EnvProduction,
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
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		},
	},
	PartnerChainIDs:        map[RelayChainID]struct{}{"0001": {}, "0053": {}},
	PartnerThroughputLimit: 2_000,
	PartnerAppLimit:        1,
	CreatedAt:              timestamp,
	UpdatedAt:              timestamp,
}

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
			expected: testAccount.Users["user_1"],
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

func Test_GetMonthlyLimit(t *testing.T) {
	tests := []struct {
		name               string
		account            Account
		expectedRelayLimit int
	}{
		{
			name: "Should return monthly relay limit when plan type is PLAN_UNLIMITED and limit is set",
			account: Account{
				PlanType:         PlanUnlimited,
				MonthlyUserLimit: 100_000,
				Plan:             &Plan{MonthlyRelayLimit: 50_000},
			},
			expectedRelayLimit: 100_000,
		},
		{
			name: "Should return monthly relay limit when plan type is PLAN_FREE and limit is set",
			account: Account{
				PlanType:         PlanFree,
				MonthlyUserLimit: 100_000,
				Plan:             &Plan{MonthlyRelayLimit: 50_000},
			},
			expectedRelayLimit: 50_000,
		},
		{
			name: "Should return monthly relay limit when plan type is PLAN_UNLIMITED and limit is NOT set (0)",
			account: Account{
				PlanType:         PlanUnlimited,
				MonthlyUserLimit: 0,
				Plan:             &Plan{MonthlyRelayLimit: 50_000},
			},
			expectedRelayLimit: 50_000,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.account.GetMonthlyLimit()
			assert.Equal(t, test.expectedRelayLimit, result)
		})
	}
}

func Test_GetPortalAppLimit(t *testing.T) {
	tests := []struct {
		name     string
		account  Account
		expected int
	}{
		{
			name: "Should return partner app limit when it is set",
			account: Account{
				PartnerAppLimit: 5,
				Plan:            &Plan{AppLimit: 10},
			},
			expected: 5,
		},
		{
			name: "Should return plan app limit when partner app limit is not set",
			account: Account{
				Plan: &Plan{AppLimit: 10},
			},
			expected: 10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.account.GetPortalAppLimit()
			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_CanCreatePortalApp(t *testing.T) {
	tests := []struct {
		name     string
		account  Account
		expected bool
	}{
		{
			name: "Should return true when the number of portal apps is less than the app limit",
			account: Account{
				PortalApps: map[PortalAppID]*PortalApp{
					"app1": {},
					"app2": {},
				},
				Plan: &Plan{AppLimit: 5},
			},
			expected: true,
		},
		{
			name: "Should return false when the number of portal apps is equal to the app limit",
			account: Account{
				PortalApps: map[PortalAppID]*PortalApp{
					"app1": {},
					"app2": {},
					"app3": {},
					"app4": {},
					"app5": {},
				},
				Plan: &Plan{AppLimit: 5},
			},
			expected: false,
		},
		{
			name: "Should return true when the number of portal apps is less than the partner app limit",
			account: Account{
				PortalApps: map[PortalAppID]*PortalApp{
					"app1": {},
					"app2": {},
				},
				PartnerAppLimit: 5,
			},
			expected: true,
		},
		{
			name: "Should return false when the number of portal apps is equal to the partner app limit",
			account: Account{
				PortalApps: map[PortalAppID]*PortalApp{
					"app1": {},
					"app2": {},
					"app3": {},
					"app4": {},
					"app5": {},
				},
				PartnerAppLimit: 5,
			},
			expected: false,
		},
		{
			name: "Should return true when the account has an unlimited app limit",
			account: Account{
				PortalApps: map[PortalAppID]*PortalApp{
					"app1": {},
					"app2": {},
					"app3": {},
					"app4": {},
					"app5": {},
					"app6": {},
				},
				Plan: &Plan{AppLimit: 0},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.account.CanCreatePortalApp()
			assert.Equal(t, test.expected, result)
		})
	}
}
