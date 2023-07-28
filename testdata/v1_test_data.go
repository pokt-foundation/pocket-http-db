package phdtestdata

import (
	"time"

	v1Types "github.com/pokt-foundation/portal-db/types"
)

var (
	MockTimestamp = time.Date(2022, time.November, 11, 11, 11, 11, 0, time.UTC)

	ExpectedLegacyApplications = map[string]*v1Types.Application{
		"test_protocol_app_1": {
			ID:     "test_protocol_app_1",
			UserID: "user_1",
			Name:   "pokt_app_123",
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_34715cae753e67c75fbb340442e7de8e",
				ApplicationPublicKey: "test_34715cae753e67c75fbb340442e7de8e",
				ApplicationSignature: "test_1dc39a2e5a84a35bf030969a0b3231f7",
				ClientPublicKey:      "test_89a3af6a587aec02cfade6f5000424c2",
				PrivateKey:           "",
				Version:              "0.0.1",
			},
			GatewaySettings: v1Types.GatewaySettings{
				SecretKey:           "test_40f482d91a5ef2300ebb4e2308c",
				SecretKeyRequired:   true,
				WhitelistOrigins:    []string{"https://test.com"},
				WhitelistUserAgents: []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"},
				WhitelistContracts: []v1Types.WhitelistContracts{
					{BlockchainID: "0001", Contracts: []string{"0x1234567890abcdef"}},
				},
				WhitelistMethods: []v1Types.WhitelistMethods{
					{BlockchainID: "0001", Methods: []string{"GET"}},
				},
				WhitelistBlockchains: []string{"0053"},
			},
			Limit: v1Types.AppLimit{
				PayPlan: v1Types.PayPlan{
					Type:  v1Types.FreetierV0,
					Limit: 250_000,
				},
			},
			NotificationSettings: v1Types.NotificationSettings{
				Quarter:       true,
				ThreeQuarters: true,
				Full:          true,
			},
			FirstDateSurpassed: MockTimestamp,
			CreatedAt:          MockTimestamp,
			UpdatedAt:          MockTimestamp,
		},
		"test_protocol_app_2": {
			ID:     "test_protocol_app_2",
			UserID: "user_3",
			Name:   "pokt_app_456",
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_8237c72345f12d1b1a8b64a1a7f66fa4",
				ApplicationPublicKey: "test_8237c72345f12d1b1a8b64a1a7f66fa4",
				ApplicationSignature: "test_f48d33b30ddaf60a1e5bb50d2ba8da5a",
				ClientPublicKey:      "test_04c71d90a92f40416b6f1d7d8af17e02",
				PrivateKey:           "",
				Version:              "0.0.1",
			},
			Limit: v1Types.AppLimit{
				PayPlan: v1Types.PayPlan{
					Type:  v1Types.PayAsYouGoV0,
					Limit: 0,
				},
			},
			GatewaySettings: v1Types.GatewaySettings{
				SecretKey:           "test_9c9e3b193cfba5348f93bb2f3e3fb794",
				WhitelistOrigins:    []string{"https://example.com"},
				WhitelistUserAgents: []string{"Mozilla/5.0 (Linux; Android 10; SM-A205U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36"},
				WhitelistContracts: []v1Types.WhitelistContracts{
					{BlockchainID: "0064", Contracts: []string{"0x0987654321abcdef"}},
				},
				WhitelistMethods: []v1Types.WhitelistMethods{
					{BlockchainID: "0064", Methods: []string{"POST"}},
				},
				WhitelistBlockchains: []string{"0021"},
			},
			NotificationSettings: v1Types.NotificationSettings{
				Half: true,
				Full: true,
			},
			FirstDateSurpassed: MockTimestamp,
			CreatedAt:          MockTimestamp,
			UpdatedAt:          MockTimestamp,
		},
		"test_protocol_app_3": {
			ID:     "test_protocol_app_3",
			UserID: "user_5",
			Name:   "pokt_app_789",
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_b5e07928fc80083c13ad0201b81bae9b",
				ApplicationPublicKey: "test_f608500e4fe3e09014fe2411b4a560b5",
				ApplicationSignature: "test_c3cd8be16ba32e24dd49fdb0247fc9b8",
				ClientPublicKey:      "test_328a9cf1b35085eeaa669aa858f6fba9",
				PrivateKey:           "",
				Version:              "0.0.1",
			},
			Limit: v1Types.AppLimit{
				PayPlan: v1Types.PayPlan{
					Type:  v1Types.Enterprise,
					Limit: 0,
				},
				CustomLimit: 4_200_000,
			},
			GatewaySettings: v1Types.GatewaySettings{
				SecretKey: "test_9f48b13e2bc5fd31ab367841f11495c1",
			},
			FirstDateSurpassed: MockTimestamp,
			CreatedAt:          MockTimestamp,
			UpdatedAt:          MockTimestamp,
		},
		"test_protocol_app_4": {
			ID:     "test_protocol_app_4",
			UserID: "user_5",
			Name:   "pokt_app_789",
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_eb2e5bcba557cfe8fa76fd7fff54f9d1",
				ApplicationPublicKey: "test_f6a5d8690ecb669865bd752b7796a920",
				ApplicationSignature: "test_cf05cf9bb26111c548e88fb6157af708",
				ClientPublicKey:      "test_6ee5ea553408f0895923fd1569dc5072",
				PrivateKey:           "",
				Version:              "0.0.1",
			},
			Limit: v1Types.AppLimit{
				PayPlan: v1Types.PayPlan{
					Type:  v1Types.Enterprise,
					Limit: 0,
				},
				CustomLimit: 4200000,
			},
			GatewaySettings: v1Types.GatewaySettings{
				SecretKey: "test_9f48b13e2bc5fd31ab367841f11495c1",
			},
			FirstDateSurpassed: MockTimestamp,
			CreatedAt:          MockTimestamp,
			UpdatedAt:          MockTimestamp,
		},
		"test_gigastake_app_1": {
			ID:   "test_gigastake_app_1",
			Name: "pokt_gigastake",
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_8d4f6a5b0c6e9f1db12c1f662e5ec8c5",
				ApplicationPublicKey: "test_37a0e8437f5149dc98a9a5b207efc2d0",
				ApplicationSignature: "test_f22651fb566346fca30b605e5f46e3ca",
				ClientPublicKey:      "test_65c29f0cc82e418b81a528a0c0682a9f",
				PrivateKey:           "",
				Version:              "0.0.1",
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"test_gigastake_app_2": {
			ID:   "test_gigastake_app_2",
			Name: "optimism_gigastake",
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_5c60d434db4e42d2b5d2ea6eeb8933c4",
				ApplicationPublicKey: "test_a7e28f8d716541a0a332a5dc6b7e4e6e",
				ApplicationSignature: "test_52e991c26da841bc882ad3a3ee9ee964",
				ClientPublicKey:      "test_ba4e53dada8f4f939048e56dc8f88f37",
				PrivateKey:           "",
				Version:              "0.0.1",
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"test_gigastake_app_3": {
			ID:   "test_gigastake_app_3",
			Name: "harmony_gigastake",
			GatewayAAT: v1Types.GatewayAAT{
				Address:              "test_e570c841d5cd4f6197e0428ed7c517fd",
				ApplicationPublicKey: "test_4f805bbbf96c4a649efc3f4f95616f2e",
				ApplicationSignature: "test_01eac46efc9242a2be73879f1d09f1dc",
				ClientPublicKey:      "test_789f9d6adcc846f1a079bf68237b5f5c",
				PrivateKey:           "",
				Version:              "0.0.1",
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
	}

	ExpectedLegacyLoadBalancers = map[string]*v1Types.LoadBalancer{
		"test_app_1": {
			ID:                "test_app_1",
			AccountID:         "account_1",
			Name:              "pokt_app_123",
			UserID:            "user_1",
			GigastakeRedirect: true,
			RequestTimeout:    5_000,
			Applications: []*v1Types.Application{
				ExpectedLegacyApplications["test_protocol_app_1"],
			},
			Users: []v1Types.UserAccess{
				{
					UserID:   "user_1",
					RoleName: "OWNER",
					Email:    "james.holden123@test.com",
					Accepted: true,
				},
				{
					UserID:   "user_2",
					RoleName: "ADMIN",
					Email:    "paul.atreides456@test.com",
					Accepted: true,
				},
				{
					UserID:   "user_8",
					RoleName: "ADMIN",
					Email:    "rick.deckard456@test.com",
					Accepted: false,
				},
			},
			Integrations: v1Types.AccountIntegrations{
				CovalentAPIKeyFree: "covalent_api_key_1",
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"test_app_2": {
			ID:                "test_app_2",
			AccountID:         "account_2",
			Name:              "pokt_app_456",
			UserID:            "user_3",
			GigastakeRedirect: true,
			RequestTimeout:    10_000,
			Applications: []*v1Types.Application{
				ExpectedLegacyApplications["test_protocol_app_2"],
			},
			Users: []v1Types.UserAccess{
				{
					UserID:   "user_3",
					RoleName: "OWNER",
					Email:    "ellen.ripley789@test.com",
					Accepted: true,
				},
				{
					UserID:   "user_2",
					RoleName: "MEMBER",
					Email:    "paul.atreides456@test.com",
					Accepted: true,
				},
				{
					UserID:   "user_4",
					RoleName: "MEMBER",
					Email:    "ulfric.stormcloak123@test.com",
					Accepted: true,
				},
				{
					UserID:   "user_9",
					RoleName: "MEMBER",
					Email:    "tyrion.lannister789@test.com",
					Accepted: false,
				},
			},
			Integrations: v1Types.AccountIntegrations{
				CovalentAPIKeyFree: "covalent_api_key_2",
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"test_app_3": {
			ID:                "test_app_3",
			AccountID:         "account_3",
			Name:              "pokt_app_789",
			UserID:            "user_5",
			GigastakeRedirect: true,
			RequestTimeout:    10_000,
			Applications: []*v1Types.Application{
				ExpectedLegacyApplications["test_protocol_app_3"],
				ExpectedLegacyApplications["test_protocol_app_4"],
			},
			Users: []v1Types.UserAccess{
				{
					UserID:   "user_5",
					RoleName: "OWNER",
					Email:    "chrisjen.avasarala1@test.com",
					Accepted: true,
				},
				{
					UserID:   "user_6",
					RoleName: "ADMIN",
					Email:    "amos.burton789@test.com",
					Accepted: true,
				},
				{
					UserID:   "user_10",
					RoleName: "MEMBER",
					Email:    "daenerys.targaryen123@test.com",
					Accepted: false,
				},
				{
					UserID:   "user_7",
					RoleName: "MEMBER",
					Email:    "frodo.baggins123@test.com",
					Accepted: true,
				},
			},
			Integrations: v1Types.AccountIntegrations{
				CovalentAPIKeyFree: "covalent_api_key_3",
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"0001-POKT-pokt-mainnet": {
			ID:        "0001-POKT-pokt-mainnet",
			Gigastake: true,
			Applications: []*v1Types.Application{
				{
					ID:   "test_gigastake_app_1",
					Name: "pokt_gigastake",
					GatewayAAT: v1Types.GatewayAAT{
						Address:              "test_8d4f6a5b0c6e9f1db12c1f662e5ec8c5",
						ApplicationPublicKey: "test_37a0e8437f5149dc98a9a5b207efc2d0",
						ApplicationSignature: "test_f22651fb566346fca30b605e5f46e3ca",
						ClientPublicKey:      "test_65c29f0cc82e418b81a528a0c0682a9f",
						PrivateKey:           "",
						Version:              "0.0.1",
					},
					CreatedAt: MockTimestamp,
					UpdatedAt: MockTimestamp,
				},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"0053-OP-optimism-mainnet": {
			ID:        "0053-OP-optimism-mainnet",
			Gigastake: true,
			Applications: []*v1Types.Application{
				{
					ID:   "test_gigastake_app_2",
					Name: "optimism_gigastake",
					GatewayAAT: v1Types.GatewayAAT{
						Address:              "test_5c60d434db4e42d2b5d2ea6eeb8933c4",
						ApplicationPublicKey: "test_a7e28f8d716541a0a332a5dc6b7e4e6e",
						ApplicationSignature: "test_52e991c26da841bc882ad3a3ee9ee964",
						ClientPublicKey:      "test_ba4e53dada8f4f939048e56dc8f88f37",
						PrivateKey:           "",
						Version:              "0.0.1",
					},
					CreatedAt: MockTimestamp,
					UpdatedAt: MockTimestamp,
				},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"0040-HMY-harmony-0": {
			ID:        "0040-HMY-harmony-0",
			Gigastake: true,
			Applications: []*v1Types.Application{
				{
					ID:   "test_gigastake_app_3",
					Name: "harmony_gigastake",
					GatewayAAT: v1Types.GatewayAAT{
						Address:              "test_e570c841d5cd4f6197e0428ed7c517fd",
						ApplicationPublicKey: "test_4f805bbbf96c4a649efc3f4f95616f2e",
						ApplicationSignature: "test_01eac46efc9242a2be73879f1d09f1dc",
						ClientPublicKey:      "test_789f9d6adcc846f1a079bf68237b5f5c",
						PrivateKey:           "",
						Version:              "0.0.1",
					},
					CreatedAt: MockTimestamp,
					UpdatedAt: MockTimestamp,
				},
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
	}

	ExpectedLegacyPayPlans = map[string]*v1Types.PayPlan{
		"basic_plan":       {Type: "basic_plan", Limit: 1_000},
		"developer_plan":   {Type: "developer_plan", Limit: 100},
		"enterprise_plan":  {Type: "enterprise_plan", Limit: 10_000},
		"pro_plan":         {Type: "pro_plan", Limit: 5_000},
		"startup_plan":     {Type: "startup_plan", Limit: 500},
		"ENTERPRISE":       {Type: "ENTERPRISE", Limit: 0},
		"FREETIER_V0":      {Type: "FREETIER_V0", Limit: 250_000},
		"PAY_AS_YOU_GO_V0": {Type: "PAY_AS_YOU_GO_V0", Limit: 0},
		"TEST_PLAN_10K":    {Type: "TEST_PLAN_10K", Limit: 10_000},
		"TEST_PLAN_90K":    {Type: "TEST_PLAN_90K", Limit: 90_000},
		"TEST_PLAN_V0":     {Type: "TEST_PLAN_V0", Limit: 0},
	}

	ExpectedLegacyBlockchains = map[string]*v1Types.Blockchain{
		"0001": {
			ID:                "0001",
			Altruist:          "https://test_pocket:auth123456@altruist-0001.com:1234", // pragma: allowlist secret
			Blockchain:        "pokt-mainnet",
			Description:       "Pocket Network Mainnet",
			EnforceResult:     "JSON",
			Path:              "/v1/query/height",
			Ticker:            "POKT",
			BlockchainAliases: []string{"pokt-mainnet"},
			Active:            true,
			Redirects: []v1Types.Redirect{
				{Alias: "pokt-mainnet", Domain: "pokt-rpc.gateway.pokt.network", LoadBalancerID: "0001-POKT-pokt-mainnet"},
			},
			SyncCheckOptions: v1Types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"query\"}",
				ResultKey: "result.sync_info", Allowance: 1,
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"0021": {
			ID:                "0021",
			Altruist:          "https://test_pocket:auth123456@altruist-0021.com:1234", // pragma: allowlist secret
			Blockchain:        "eth-mainnet",
			ChainID:           "1",
			ChainIDCheck:      "{\"method\":\"eth_chainId\",\"id\":1,\"jsonrpc\":\"2.0\"}",
			Description:       "Ethereum Mainnet",
			EnforceResult:     "JSON",
			Ticker:            "ETH",
			BlockchainAliases: []string{"eth-mainnet"},
			LogLimitBlocks:    100000,
			Active:            true,
			Redirects: []v1Types.Redirect{
				{Alias: "eth-mainnet", Domain: "eth-rpc.gateway.pokt.network", LoadBalancerID: "0021-ETH-eth-mainnet"},
			},
			SyncCheckOptions: v1Types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[]}",
				ResultKey: "result", Allowance: 5,
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"0040": {
			ID:                "0040",
			Altruist:          "https://test_pocket:auth123456@altruist-0040.com:1234", // pragma: allowlist secret
			Blockchain:        "harmony-0",
			Description:       "Harmony Shard 0",
			EnforceResult:     "JSON",
			Ticker:            "HMY",
			BlockchainAliases: []string{"harmony-0"},
			Active:            true,
			Redirects: []v1Types.Redirect{
				{Alias: "harmony-0", Domain: "hmy-rpc.gateway.pokt.network", LoadBalancerID: "0040-HMY-harmony-0"},
			},
			SyncCheckOptions: v1Types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"hmy_blockNumber\",\"params\":[]}",
				ResultKey: "result",
				Allowance: 8,
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"0053": {
			ID:                "0053",
			Altruist:          "https://test_pocket:auth123456@altruist-0053.com:1234", // pragma: allowlist secret
			Blockchain:        "optimism-mainnet",
			Description:       "Optimism Mainnet",
			EnforceResult:     "JSON",
			Ticker:            "OP",
			BlockchainAliases: []string{"optimism-mainnet"},
			LogLimitBlocks:    100000,
			Active:            true,
			Redirects: []v1Types.Redirect{
				{Alias: "optimism-mainnet", Domain: "op-rpc.gateway.pokt.network", LoadBalancerID: "0053-OP-optimism-mainnet"},
			},
			SyncCheckOptions: v1Types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[]}",
				ResultKey: "result",
				Allowance: 2,
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
		"0064": {
			ID:             "0064",
			Altruist:       "https://test_pocket:auth123456@altruist-0064.com:1234", // pragma: allowlist secret
			Blockchain:     "sui-testnet",
			Description:    "Sui Testnet",
			EnforceResult:  "JSON",
			Ticker:         "SUI-TESTNET",
			LogLimitBlocks: 100000,
			RequestTimeout: 60000,
			SyncCheckOptions: v1Types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"sui_blockNumber\",\"params\":[]}",
				ResultKey: "result",
				Allowance: 7,
			},
			CreatedAt: MockTimestamp,
			UpdatedAt: MockTimestamp,
		},
	}

	CreateLegacyLoadBalancer = &v1Types.LoadBalancer{
		Name:              "new_pokt_app_777",
		UserID:            "auth0|james_holden",
		ApplicationIDs:    []string(nil),
		RequestTimeout:    5000,
		GigastakeRedirect: true,
		Applications: []*v1Types.Application{
			{
				UserID: "auth0|james_holden",
				Name:   "new_pokt_app_777",
				GatewayAAT: v1Types.GatewayAAT{
					Address:              "test_34715cae753e67c75fbb340442e7de8e",
					ApplicationPublicKey: "test_34715cae753e67c75fbb340442e7de8e",
					ApplicationSignature: "test_1dc39a2e5a84a35bf030969a0b3231f7",
					ClientPublicKey:      "test_89a3af6a587aec02cfade6f5000424c2",
					PrivateKey:           "test_11b8d394ca331d7c7a71ca1896d630f6",
					Version:              "0.0.1",
				},
				GatewaySettings:      v1Types.GatewaySettings{SecretKey: "test_40f482d91a5ef2300ebb4e2308c"},
				Limit:                v1Types.AppLimit{PayPlan: v1Types.PayPlan{Type: "FREETIER_V0", Limit: 250_000}, CustomLimit: 0},
				NotificationSettings: v1Types.NotificationSettings{SignedUp: true, Quarter: false, Half: false, ThreeQuarters: true, Full: true},
			},
		},
		Users: []v1Types.UserAccess{
			{UserID: "user_1", RoleName: v1Types.RoleOwner, Email: "james.holden123@test.com", Accepted: true},
		},
		AccountID: "account_1",
	}
)
