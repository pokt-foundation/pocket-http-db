package types

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	testPortalApplication = PortalApp{
		ID:        "test_app_1",
		Name:      "test_portal_app_123",
		AccountID: "account_1",
		AATs: map[ProtocolAppID]AAT{
			"test_protocol_app_1": {
				Address:         "test_34715cae753e67c75fbb340442e7de8e",
				PublicKey:       "test_11b8d394ca331d7c7a71ca1896d630f6",
				ClientPublicKey: "test_9e9ca4fe13725d412003f4bc518f6974",
				PrivateKey:      "test_89a3af6a587aec02cfade6f5000424c2",
				Signature:       "test_1dc39a2e5a84a35bf030969a0b3231f7",
				Version:         "0.0.1",
			},
		},
		Settings: Settings{
			Environment:       EnvProduction,
			SecretKey:         "test_90210ac4bdd3423e24877d1ff92",
			SecretKeyRequired: true,
			MonthlyRelayLimit: 250_000,
			FavoritedChainIDs: map[RelayChainID]struct{}{"0001": {}, "0056": {}, "0002": {}, "003E": {}},
		},
		Whitelists: Whitelists{
			Origins:     map[Origin]struct{}{"https://www.example.com": {}, "https://subdomain.example.com": {}, "https://portalgun.io": {}},
			UserAgents:  map[UserAgent]struct{}{"Mozilla Firefox": {}, "Brave": {}, "Google Chrome": {}, "Safari": {}, "Netscape Navigator": {}},
			Blockchains: map[RelayChainID]struct{}{"0001": {}, "0056": {}, "0002": {}, "003E": {}},
			Contracts: map[RelayChainID]map[Contract]struct{}{
				"0001": {"0xtest_5fbfe3e9af3971dd833d26ba9b5c936f0be": {}, "0xtest_2f78db6436527729929aaf6c616361de0f7": {}},
				"0056": {"0xtest_5068778dd592e39a122f4f5a5cf09c90fe2": {}, "0xtest_00000f279d81a1d3cc75430faa017fa5a2e": {}},
				"0002": {"0xtest_1111117dc0aa78b770fa6a738034120c302": {}, "0xtest_a39b223fe8d0a0e5c4f27ead9083c756cc2": {}},
				"003E": {"0xtest_f958d2ee523a2206206994597c13d831ec7": {}, "0xtest_0a85d5af5bf1d1762f925bdaddc4201f984": {}},
			},
			Methods: map[RelayChainID]map[Method]struct{}{
				"0001": {"GET": {}, "POST": {}, "PUT": {}},
				"0056": {"GET": {}, "POST": {}},
				"0002": {"GET": {}, "POST": {}, "PUT": {}, "DELETE": {}},
				"003E": {"GET": {}},
			},
		},
		CreatedAt: time.Date(2023, time.February, 14, 11, 11, 11, 0, time.UTC),
		UpdatedAt: time.Date(2023, time.February, 27, 13, 13, 13, 0, time.UTC),
	}

	testPlan = &Plan{
		Type:              PayPlanType("FREETIER_V0"),
		ChainIDs:          map[RelayChainID]struct{}{"0001": {}, "0053": {}},
		MonthlyRelayLimit: 5_000_000,
		ThroughputLimit:   5_000,
		AppLimit:          5,
		DailyLimit:        250_000,
		CreatedAt:         time.Date(2023, time.February, 14, 11, 11, 11, 0, time.UTC),
	}

	testPortalAppLite = PortalAppLite{
		ID:         "test_app_1",
		PublicKeys: []PortalAppPublicKey{"test_11b8d394ca331d7c7a71ca1896d630f6"},
		Settings: SettingsLite{
			SecretKey:         "test_90210ac4bdd3423e24877d1ff92",
			SecretKeyRequired: true,
		},
		Whitelists: Whitelists{
			Origins:     map[Origin]struct{}{"https://www.example.com": {}, "https://subdomain.example.com": {}, "https://portalgun.io": {}},
			UserAgents:  map[UserAgent]struct{}{"Mozilla Firefox": {}, "Brave": {}, "Google Chrome": {}, "Safari": {}, "Netscape Navigator": {}},
			Blockchains: map[RelayChainID]struct{}{"0001": {}, "0056": {}, "0002": {}, "003E": {}},
			Contracts: map[RelayChainID]map[Contract]struct{}{
				"0001": {"0xtest_5fbfe3e9af3971dd833d26ba9b5c936f0be": {}, "0xtest_2f78db6436527729929aaf6c616361de0f7": {}},
				"0056": {"0xtest_5068778dd592e39a122f4f5a5cf09c90fe2": {}, "0xtest_00000f279d81a1d3cc75430faa017fa5a2e": {}},
				"0002": {"0xtest_1111117dc0aa78b770fa6a738034120c302": {}, "0xtest_a39b223fe8d0a0e5c4f27ead9083c756cc2": {}},
				"003E": {"0xtest_f958d2ee523a2206206994597c13d831ec7": {}, "0xtest_0a85d5af5bf1d1762f925bdaddc4201f984": {}},
			},
			Methods: map[RelayChainID]map[Method]struct{}{
				"0001": {"GET": {}, "POST": {}, "PUT": {}},
				"0056": {"GET": {}, "POST": {}},
				"0002": {"GET": {}, "POST": {}, "PUT": {}, "DELETE": {}},
				"003E": {"GET": {}},
			},
		},
		Plan: PlanLite{
			PlanType:        PayPlanType("FREETIER_V0"),
			ChainIDs:        map[RelayChainID]struct{}{"0001": {}, "0053": {}},
			ThroughputLimit: 5_000,
		},
	}
)

func Test_ConvertPortalAppToPortalAppLite(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name           string
		portalApp      PortalApp
		appPlan        *Plan
		expectedMidApp PortalAppLite
	}{
		{
			name:           "Should correctly convert PortalApp to PortalAppLite",
			portalApp:      testPortalApplication,
			appPlan:        testPlan,
			expectedMidApp: testPortalAppLite,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			middlewarePortalApp := test.portalApp.ConvertPortalAppToPortalAppLite(test.appPlan)
			c.Equal(test.expectedMidApp.ID, middlewarePortalApp.ID)
			c.Equal(test.expectedMidApp.Settings, middlewarePortalApp.Settings)
			c.Equal(test.expectedMidApp.Whitelists, middlewarePortalApp.Whitelists)
		})
	}
}

func Test_PortalApp_GetPublicKeys(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name           string
		portalApp      *PortalApp
		expectedResult []PortalAppPublicKey
	}{
		{
			name:           "Should return all public keys of the app",
			portalApp:      &testPortalApplication,
			expectedResult: []PortalAppPublicKey{"test_11b8d394ca331d7c7a71ca1896d630f6"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			publicKeys := test.portalApp.GetPublicKeys()
			c.Equal(test.expectedResult, publicKeys)
		})
	}
}

func Test_PortalAppLite_PlanType(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name          string
		portalAppLite PortalAppLite
		expectedType  PayPlanType
	}{
		{
			name:          "Should correctly return PlanType",
			portalAppLite: testPortalAppLite,
			expectedType:  "FREETIER_V0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c.Equal(test.expectedType, test.portalAppLite.PlanType())
		})
	}
}

func Test_PortalAppLite_IsChainIDOnPlan(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name           string
		portalAppLite  PortalAppLite
		chainID        RelayChainID
		expectedResult bool
	}{
		{
			name:           "Should correctly return IsChainIDOnPlan - existing ID",
			portalAppLite:  testPortalAppLite,
			chainID:        "0001",
			expectedResult: true,
		},
		{
			name:           "Should correctly return IsChainIDOnPlan - non-existing ID",
			portalAppLite:  testPortalAppLite,
			chainID:        "0002",
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c.Equal(test.expectedResult, test.portalAppLite.IsChainIDOnPlan(test.chainID))
		})
	}
}

func Test_PortalAppLite_ThroughputLimit(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name          string
		portalAppLite PortalAppLite
		expectedLimit int32
	}{
		{
			name:          "Should correctly return ThroughputLimit",
			portalAppLite: testPortalAppLite,
			expectedLimit: 5_000,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c.Equal(test.expectedLimit, test.portalAppLite.ThroughputLimit())
		})
	}
}

func Test_PortalApp_IsOriginWhitelisted(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name           string
		portalApp      PortalAppLite
		origin         Origin
		expectedResult bool
	}{
		{
			name:           "Should return true if a given origin is whitelisted for a given app",
			portalApp:      testPortalAppLite,
			origin:         Origin("https://portalgun.io"),
			expectedResult: true,
		},
		{
			name:           "Should return true if a given origin is whitelisted for a given app",
			portalApp:      testPortalAppLite,
			origin:         Origin("https://www.example.com"),
			expectedResult: true,
		},
		{
			name:           "Should return false if a given origin is not whitelisted for a given app",
			portalApp:      testPortalAppLite,
			origin:         Origin("https://ricksanchez.io"),
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isOriginWhitelisted := test.portalApp.IsOriginWhitelisted(test.origin)
			c.Equal(test.expectedResult, isOriginWhitelisted)
		})
	}
}

func Test_PortalApp_IsUserAgentWhitelisted(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name           string
		portalApp      PortalAppLite
		userAgent      UserAgent
		expectedResult bool
	}{
		{
			name:           "Should return true if a given user agent is whitelisted for a given app",
			portalApp:      testPortalAppLite,
			userAgent:      UserAgent("Brave"),
			expectedResult: true,
		},
		{
			name:           "Should return true if a given user agent is whitelisted for a given app",
			portalApp:      testPortalAppLite,
			userAgent:      UserAgent("Safari"),
			expectedResult: true,
		},
		{
			name:           "Should return false if a given user agent is not whitelisted for a given app",
			portalApp:      testPortalAppLite,
			userAgent:      UserAgent("Bird Person"),
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isUserAgentWhitelisted := test.portalApp.IsUserAgentWhitelisted(test.userAgent)
			c.Equal(test.expectedResult, isUserAgentWhitelisted)
		})
	}
}

func Test_PortalApp_IsBlockchainWhitelisted(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name           string
		portalApp      PortalAppLite
		blockchain     RelayChainID
		expectedResult bool
	}{
		{
			name:           "Should return true if a given blockchain is whitelisted for a given app",
			portalApp:      testPortalAppLite,
			blockchain:     RelayChainID("0001"),
			expectedResult: true,
		},
		{
			name:           "Should return true if a given blockchain is whitelisted for a given app",
			portalApp:      testPortalAppLite,
			blockchain:     RelayChainID("003E"),
			expectedResult: true,
		},
		{
			name:           "Should return false if a given blockchain is not whitelisted for a given app",
			portalApp:      testPortalAppLite,
			blockchain:     RelayChainID("7009"),
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isBlockchainWhitelisted := test.portalApp.IsBlockchainWhitelisted(test.blockchain)
			c.Equal(test.expectedResult, isBlockchainWhitelisted)
		})
	}
}

func Test_PortalApp_IsContractWhitelisted(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name           string
		portalApp      PortalAppLite
		blockchain     RelayChainID
		contract       Contract
		expectedResult bool
	}{
		{
			name:           "Should return true if a given contract is whitelisted for a given app and blockchain",
			portalApp:      testPortalAppLite,
			blockchain:     RelayChainID("0001"),
			contract:       Contract("0xtest_5fbfe3e9af3971dd833d26ba9b5c936f0be"),
			expectedResult: true,
		},
		{
			name:           "Should return true if a given contract is whitelisted for a given app and blockchain",
			portalApp:      testPortalAppLite,
			blockchain:     RelayChainID("003E"),
			contract:       Contract("0xtest_0a85d5af5bf1d1762f925bdaddc4201f984"),
			expectedResult: true,
		},
		{
			name:           "Should return false if a given contract is not whitelisted for a given app and blockchain",
			portalApp:      testPortalAppLite,
			blockchain:     RelayChainID("0056"),
			contract:       Contract("0xtest_04938rfj439fj3409jf0439fjf4304f4444"),
			expectedResult: false,
		},
		{
			name:           "Should return false if a given contract is not whitelisted for a given app and blockchain",
			portalApp:      testPortalAppLite,
			blockchain:     RelayChainID("7009"),
			contract:       Contract("0xtest_439834fnin3f2032f03re3j2f30fj33f3f3"),
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isBlockchainWhitelisted := test.portalApp.IsContractWhitelisted(test.blockchain, test.contract)
			c.Equal(test.expectedResult, isBlockchainWhitelisted)
		})
	}
}

func Test_PortalApp_IsMethodWhitelisted(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name           string
		portalApp      PortalAppLite
		blockchain     RelayChainID
		method         Method
		expectedResult bool
	}{
		{
			name:           "Should return true if a given method is whitelisted for a given app and blockchain",
			portalApp:      testPortalAppLite,
			blockchain:     RelayChainID("0001"),
			method:         Method("POST"),
			expectedResult: true,
		},
		{
			name:           "Should return true if a given method is whitelisted for a given app and blockchain",
			portalApp:      testPortalAppLite,
			blockchain:     RelayChainID("003E"),
			method:         Method("GET"),
			expectedResult: true,
		},
		{
			name:           "Should return false if a given method is not whitelisted for a given app and blockchain",
			portalApp:      testPortalAppLite,
			blockchain:     RelayChainID("0056"),
			method:         Method("PUT"),
			expectedResult: false,
		},
		{
			name:           "Should return false if a given method is not whitelisted for a given app and blockchain",
			portalApp:      testPortalAppLite,
			blockchain:     RelayChainID("7009"),
			method:         Method("GET"),
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isBlockchainWhitelisted := test.portalApp.IsMethodWhitelisted(test.blockchain, test.method)
			c.Equal(test.expectedResult, isBlockchainWhitelisted)
		})
	}
}

func Test_PortalApp_GetWhitelistsObject(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name           string
		portalApp      PortalApp
		expectedResult *WhitelistsObject
		expectedJSON   string
	}{
		{
			name:      "Should return an applications whitelists in the form of a WhitelistContracts struct",
			portalApp: testPortalApplication,
			expectedResult: &WhitelistsObject{
				AppWhitelists: [3]ApplicationWhitelists{
					{Type: "origins", Values: []string{"https://portalgun.io", "https://subdomain.example.com", "https://www.example.com"}},
					{Type: "userAgents", Values: []string{"Brave", "Google Chrome", "Mozilla Firefox", "Netscape Navigator", "Safari"}},
					{Type: "blockchains", Values: []string{"0001", "0002", "003E", "0056"}},
				},
				ChainWhitelists: [2]ChainWhitelists{
					{Type: "contracts", Values: []ChainIDWhitelists{
						{ChainID: "0001", Values: []string{"0xtest_2f78db6436527729929aaf6c616361de0f7", "0xtest_5fbfe3e9af3971dd833d26ba9b5c936f0be"}},
						{ChainID: "0002", Values: []string{"0xtest_1111117dc0aa78b770fa6a738034120c302", "0xtest_a39b223fe8d0a0e5c4f27ead9083c756cc2"}},
						{ChainID: "003E", Values: []string{"0xtest_0a85d5af5bf1d1762f925bdaddc4201f984", "0xtest_f958d2ee523a2206206994597c13d831ec7"}},
						{ChainID: "0056", Values: []string{"0xtest_00000f279d81a1d3cc75430faa017fa5a2e", "0xtest_5068778dd592e39a122f4f5a5cf09c90fe2"}},
					},
					},
					{Type: "methods", Values: []ChainIDWhitelists{
						{ChainID: "0001", Values: []string{"GET", "POST", "PUT"}},
						{ChainID: "0002", Values: []string{"DELETE", "GET", "POST", "PUT"}},
						{ChainID: "003E", Values: []string{"GET"}},
						{ChainID: "0056", Values: []string{"GET", "POST"}},
					},
					},
				},
			},
			expectedJSON: `{
				"appWhitelists": [
					{
						"type": "origins",
						"values": [
							"https://portalgun.io",
							"https://subdomain.example.com",
							"https://www.example.com"
						]
					},
					{
						"type": "userAgents",
						"values": [
							"Brave",
							"Google Chrome",
							"Mozilla Firefox",
							"Netscape Navigator",
							"Safari"
						]
					},
					{
						"type": "blockchains",
						"values": [
							"0001",
							"0002",
							"003E",
							"0056"
						]
					}
				],
				"chainWhitelists": [
					{
						"type": "contracts",
						"values": [
							{
								"chainID": "0001",
								"values": [
									"0xtest_2f78db6436527729929aaf6c616361de0f7",
									"0xtest_5fbfe3e9af3971dd833d26ba9b5c936f0be"
								]
							},
							{
								"chainID": "0002",
								"values": [
									"0xtest_1111117dc0aa78b770fa6a738034120c302",
									"0xtest_a39b223fe8d0a0e5c4f27ead9083c756cc2"
								]
							},
							{
								"chainID": "003E",
								"values": [
									"0xtest_0a85d5af5bf1d1762f925bdaddc4201f984",
									"0xtest_f958d2ee523a2206206994597c13d831ec7"
								]
							},
							{
								"chainID": "0056",
								"values": [
									"0xtest_00000f279d81a1d3cc75430faa017fa5a2e",
									"0xtest_5068778dd592e39a122f4f5a5cf09c90fe2"
								]
							}
						]
					},
					{
						"type": "methods",
						"values": [
							{
								"chainID": "0001",
								"values": [
									"GET",
									"POST",
									"PUT"
								]
							},
							{
								"chainID": "0002",
								"values": [
									"DELETE",
									"GET",
									"POST",
									"PUT"
								]
							},
							{
								"chainID": "003E",
								"values": [
									"GET"
								]
							},
							{
								"chainID": "0056",
								"values": [
									"GET",
									"POST"
								]
							}
						]
					}
				]
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			whitelistsObject := test.portalApp.GetWhitelistsObject()
			c.Equal(test.expectedResult, whitelistsObject)

			// check that JSON is equal as well
			resultJSON, err := json.MarshalIndent(whitelistsObject, "", "  ")
			c.NoError(err)
			expectedJSON := strings.ReplaceAll(strings.ReplaceAll(test.expectedJSON, " ", ""), "\t", "")
			actualJSON := strings.ReplaceAll(strings.ReplaceAll(string(resultJSON), " ", ""), "\t", "")
			c.Equal(expectedJSON, actualJSON)
		})
	}
}
