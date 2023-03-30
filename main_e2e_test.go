package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	dbclient "github.com/pokt-foundation/db-client/client"
	"github.com/pokt-foundation/portal-db/v2/types"
	"github.com/stretchr/testify/suite"
)

var (
	testCtx       = context.Background()
	mockTimestamp = time.Date(2022, 11, 11, 11, 11, 11, 0, time.UTC)
)

type phdE2ETestSuite struct {
	suite.Suite
	client1, client2 dbclient.IDBClient
	mu               sync.Mutex
}

func Test_V1_E2E_PocketHTTPDBTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end to end test")
	}

	suite.Run(t, new(phdE2ETestSuite))
}

// SetupSuite runs before each test suite run
func (ts *phdE2ETestSuite) SetupSuite() {
	err := ts.initDBClient()
	ts.NoError(err)
}

// Initializes two separate DB clients to talk to two separate PHD instances.
// This is to test the Postgres listener is updating multiple instances as intended.
func (ts *phdE2ETestSuite) initDBClient() error {
	baseConfig := dbclient.Config{
		APIKey:  "test_api_key_6789",
		Version: dbclient.V1,
		Retries: 1,
		Timeout: 10 * time.Second,
	}

	config1 := baseConfig
	config1.BaseURL = "http://localhost:8080"
	client1, err := dbclient.NewDBClient(config1)
	if err != nil {
		return err
	}
	ts.client1 = client1

	config2 := baseConfig
	config2.BaseURL = "http://localhost:8081"
	client2, err := dbclient.NewDBClient(config2)
	if err != nil {
		return err
	}
	ts.client2 = client2

	return nil
}

func boolPointer(value bool) *bool {
	return &value
}

// Runs all the read-only endpoint tests first to compare to test DB seed data only
// ie. not yet including data written to the test DB by the test suite
func (ts *phdE2ETestSuite) Test_ReadTests() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.Run("Test_GetBlockchains", func() {
		tests := []struct {
			name                string
			expectedBlockchains []*types.Blockchain
			err                 error
		}{
			{
				name: "Should fetch all blockchains in the database",
				expectedBlockchains: []*types.Blockchain{
					{
						ID:                "0001",
						Altruist:          "https://test:test_93uhfniu23f8@shared-test2.nodes.pokt.network:12345",
						Blockchain:        "pokt-mainnet",
						Description:       "POKT Network Mainnet",
						EnforceResult:     "JSON",
						Ticker:            "POKT",
						BlockchainAliases: []string{"pokt-mainnet"},
						LogLimitBlocks:    100_000,
						Active:            true,
						Redirects: []types.Redirect{
							{
								Alias:          "test-mainnet",
								Domain:         "test-rpc1.testnet.pokt.network",
								LoadBalancerID: "test_lb_34gg4g43g34g5hh",
							},
							{
								Alias:          "test-mainnet",
								Domain:         "test-rpc2.testnet.pokt.network",
								LoadBalancerID: "test_lb_34gg4g43g34g5hh",
							},
						},
						SyncCheckOptions: types.SyncCheckOptions{
							Body:      `{}`,
							ResultKey: "height",
							Allowance: 1,
						},
						CreatedAt: mockTimestamp,
						UpdatedAt: mockTimestamp,
					},
					{
						ID:                "0021",
						Altruist:          "https://test:test_u32fh239hf@shared-test2.nodes.eth.network:12345",
						Blockchain:        "eth-mainnet",
						ChainID:           "1",
						ChainIDCheck:      `{\"method\":\"eth_chainId\",\"id\":1,\"jsonrpc\":\"2.0\"}`,
						Description:       "Ethereum Mainnet",
						EnforceResult:     "JSON",
						Ticker:            "ETH",
						BlockchainAliases: []string{"eth-mainnet"},
						LogLimitBlocks:    100_000,
						Active:            true,
						Redirects: []types.Redirect{
							{
								Alias:          "eth-mainnet",
								Domain:         "test-rpc.testnet.eth.network",
								LoadBalancerID: "test_lb_34gg4g43g34g5hh",
							},
						},
						SyncCheckOptions: types.SyncCheckOptions{
							Body:      `{\"method\":\"eth_blockNumber\",\"id\":1,\"jsonrpc\":\"2.0\"}`,
							ResultKey: "result",
							Allowance: 5,
						},
						CreatedAt: mockTimestamp,
						UpdatedAt: mockTimestamp,
					},
				},
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				blockchains, err := ts.client1.GetBlockchains(testCtx)
				ts.ErrorIs(test.err, err)
				ts.Equal(test.expectedBlockchains, blockchains)

				blockchains, err = ts.client2.GetBlockchains(testCtx)
				ts.ErrorIs(test.err, err)
				ts.Equal(test.expectedBlockchains, blockchains)
			})
		}
	})
	ts.Run("Test_GetBlockchain", func() {
		tests := []struct {
			name               string
			blockchainID       string
			expectedBlockchain *types.Blockchain
			err                error
		}{
			{
				name:         "Should fetch one blockchain by ID",
				blockchainID: "0021",
				expectedBlockchain: &types.Blockchain{
					ID:                "0021",
					Altruist:          "https://test:test_u32fh239hf@shared-test2.nodes.eth.network:12345",
					Blockchain:        "eth-mainnet",
					ChainID:           "1",
					ChainIDCheck:      `{\"method\":\"eth_chainId\",\"id\":1,\"jsonrpc\":\"2.0\"}`,
					Description:       "Ethereum Mainnet",
					EnforceResult:     "JSON",
					Ticker:            "ETH",
					BlockchainAliases: []string{"eth-mainnet"},
					LogLimitBlocks:    100_000,
					Active:            true,
					Redirects: []types.Redirect{
						{
							Alias:          "eth-mainnet",
							Domain:         "test-rpc.testnet.eth.network",
							LoadBalancerID: "test_lb_34gg4g43g34g5hh",
						},
					},
					SyncCheckOptions: types.SyncCheckOptions{
						Body:      `{\"method\":\"eth_blockNumber\",\"id\":1,\"jsonrpc\":\"2.0\"}`,
						ResultKey: "result",
						Allowance: 5,
					},
					CreatedAt: mockTimestamp,
					UpdatedAt: mockTimestamp,
				},
			},
			{
				name:         "Should fail if the blockchain does not exist in the DB",
				blockchainID: "666",
				err:          fmt.Errorf("Response not OK. 404 Not Found: blockchain not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				blockchain, err := ts.client1.GetBlockchainByID(testCtx, test.blockchainID)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedBlockchain, blockchain)

				blockchain, err = ts.client2.GetBlockchainByID(testCtx, test.blockchainID)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedBlockchain, blockchain)
			})
		}
	})

	ts.Run("Test_GetApplications", func() {
		tests := []struct {
			name                 string
			expectedApplications []*types.Application
			err                  error
		}{
			{
				name: "Should fetch all applications in the database",
				expectedApplications: []*types.Application{
					{
						ID:     "test_app_47hfnths73j2se",
						UserID: "test_user_1dbffbdfeeb225",
						Name:   "pokt_app_123",
						URL:    "https://test.app123.io",
						Dummy:  true,
						Status: types.InService,
						GatewayAAT: types.GatewayAAT{
							Address:              "test_34715cae753e67c75fbb340442e7de8e",
							ApplicationPublicKey: "test_11b8d394ca331d7c7a71ca1896d630f6",
							ApplicationSignature: "test_89a3af6a587aec02cfade6f5000424c2",
							ClientPublicKey:      "test_1dc39a2e5a84a35bf030969a0b3231f7",
							PrivateKey:           "test_d2ce53f115f4ecb2208e9188800a85cf",
						},
						GatewaySettings: types.GatewaySettings{
							SecretKey:            "test_40f482d91a5ef2300ebb4e2308c",
							SecretKeyRequired:    true,
							WhitelistBlockchains: []string{"chain_1"},
							WhitelistOrigins:     []string{"origin_1", "origin_2"},
							WhitelistUserAgents:  []string{"user_agent_1", "user_agent_2", "user_agent_3"},
							WhitelistContracts: []types.WhitelistContracts{
								{BlockchainID: "TST2", Contracts: []string{"test_address_67890"}},
								{BlockchainID: "TST1", Contracts: []string{"test_address_12345", "test_address_44444"}},
							},
							WhitelistMethods: []types.WhitelistMethods{
								{BlockchainID: "TST1", Methods: []string{"GET"}},
								{BlockchainID: "TST2", Methods: []string{"PUT", "POST"}},
							},
						},
						Limit: types.AppLimit{
							PayPlan: types.PayPlan{Type: types.FreetierV0, Limit: 250_000},
						},
						NotificationSettings: types.NotificationSettings{
							SignedUp:      true,
							Quarter:       false,
							Half:          false,
							ThreeQuarters: true,
							Full:          true,
						},
						CreatedAt: mockTimestamp,
						UpdatedAt: mockTimestamp,
					},
					{
						ID:     "test_app_5hdf7sh23jd828",
						UserID: "test_user_04228205bd261a",
						Name:   "pokt_app_456",
						URL:    "https://test.app456.io",
						Dummy:  true,
						Status: types.InService,
						GatewayAAT: types.GatewayAAT{
							Address:              "test_558c0225c7019e14ccf2e7379ad3eb50",
							ApplicationPublicKey: "test_96c981db344ab6920b7e87853838e285",
							ApplicationSignature: "test_1272a8ab4cbbf636f09bf4fa5395b885",
							ClientPublicKey:      "test_d709871777b89ed3051190f229ea3f01",
							PrivateKey:           "test_53e50765d8bc1fb41b3b0065dd8094de",
						},
						GatewaySettings: types.GatewaySettings{
							SecretKey:         "test_90210ac4bdd3423e24877d1ff92",
							SecretKeyRequired: false,
						},
						Limit: types.AppLimit{
							PayPlan:     types.PayPlan{Type: types.Enterprise},
							CustomLimit: 2_000_000,
						},
						NotificationSettings: types.NotificationSettings{
							SignedUp:      true,
							Quarter:       false,
							Half:          false,
							ThreeQuarters: true,
							Full:          true,
						},
						CreatedAt: mockTimestamp,
						UpdatedAt: mockTimestamp,
					},
					{
						ID:     "test_app_97djd63jd78r7f",
						UserID: "test_user_redirect233344",
						Name:   "pokt_app_789",
						URL:    "https://test.app789.io",
						Dummy:  true,
						Status: types.InService,
						GatewayAAT: types.GatewayAAT{
							Address:              "test_58a37510a103f3de57296af882b2603c",
							ApplicationPublicKey: "test_cc49729227c22f3934a966d99a6be72b",
							ApplicationSignature: "test_8d5f4a3006b5603bcecfb44a819c272a",
							ClientPublicKey:      "test_28a54385e49b65a860f61db62449703b",
							PrivateKey:           "test_8d5f4a3006b5603bcecfb44a819c272a",
						},
						GatewaySettings: types.GatewaySettings{
							SecretKey:         "test_8fd84jf74jf83kd83kd92okg355",
							SecretKeyRequired: false,
						},
						Limit: types.AppLimit{
							PayPlan: types.PayPlan{Type: types.PayAsYouGoV0, Limit: 0},
						},
						NotificationSettings: types.NotificationSettings{
							SignedUp:      true,
							Quarter:       false,
							Half:          false,
							ThreeQuarters: true,
							Full:          true,
						},
						CreatedAt: mockTimestamp,
						UpdatedAt: mockTimestamp,
					},
				},
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				applications, err := ts.client1.GetApplications(testCtx)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedApplications, applications)

				applications, err = ts.client2.GetApplications(testCtx)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedApplications, applications)
			})
		}
	})

	ts.Run("Test_GetApplicationByID", func() {
		tests := []struct {
			name                string
			applicationID       string
			expectedApplication *types.Application
			err                 error
		}{
			{
				name:          "Should fetch one application by ID",
				applicationID: "test_app_5hdf7sh23jd828",
				expectedApplication: &types.Application{
					ID:     "test_app_5hdf7sh23jd828",
					UserID: "test_user_04228205bd261a",
					Name:   "pokt_app_456",
					URL:    "https://test.app456.io",
					Dummy:  true,
					Status: types.InService,
					GatewayAAT: types.GatewayAAT{
						Address:              "test_558c0225c7019e14ccf2e7379ad3eb50",
						ApplicationPublicKey: "test_96c981db344ab6920b7e87853838e285",
						ApplicationSignature: "test_1272a8ab4cbbf636f09bf4fa5395b885",
						ClientPublicKey:      "test_d709871777b89ed3051190f229ea3f01",
						PrivateKey:           "test_53e50765d8bc1fb41b3b0065dd8094de",
					},
					GatewaySettings: types.GatewaySettings{
						SecretKey:         "test_90210ac4bdd3423e24877d1ff92",
						SecretKeyRequired: false,
					},
					Limit: types.AppLimit{
						PayPlan:     types.PayPlan{Type: types.Enterprise},
						CustomLimit: 2_000_000,
					},
					NotificationSettings: types.NotificationSettings{
						SignedUp:      true,
						Quarter:       false,
						Half:          false,
						ThreeQuarters: true,
						Full:          true,
					},
					CreatedAt: mockTimestamp,
					UpdatedAt: mockTimestamp,
				},
			},
			{
				name:          "Should fail if the application does not exist in the DB",
				applicationID: "test_not_real_app",
				err:           fmt.Errorf("Response not OK. 404 Not Found: applications not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				applicationByID, err := ts.client1.GetApplicationByID(testCtx, test.applicationID)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedApplication, applicationByID)

				applicationByID, err = ts.client2.GetApplicationByID(testCtx, test.applicationID)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedApplication, applicationByID)
			})
		}
	})

	ts.Run("Test_GetApplicationsByUserID", func() {
		tests := []struct {
			name                 string
			userID               string
			expectedApplications []*types.Application
			err                  error
		}{
			{
				name:   "Should fetch all applications for a single user ID",
				userID: "test_user_04228205bd261a",
				expectedApplications: []*types.Application{
					{
						ID:     "test_app_5hdf7sh23jd828",
						UserID: "test_user_04228205bd261a",
						Name:   "pokt_app_456",
						URL:    "https://test.app456.io",
						Dummy:  true,
						Status: types.InService,
						GatewayAAT: types.GatewayAAT{
							Address:              "test_558c0225c7019e14ccf2e7379ad3eb50",
							ApplicationPublicKey: "test_96c981db344ab6920b7e87853838e285",
							ApplicationSignature: "test_1272a8ab4cbbf636f09bf4fa5395b885",
							ClientPublicKey:      "test_d709871777b89ed3051190f229ea3f01",
							PrivateKey:           "test_53e50765d8bc1fb41b3b0065dd8094de",
						},
						GatewaySettings: types.GatewaySettings{
							SecretKey:         "test_90210ac4bdd3423e24877d1ff92",
							SecretKeyRequired: false,
						},
						Limit: types.AppLimit{
							PayPlan:     types.PayPlan{Type: types.Enterprise},
							CustomLimit: 2_000_000,
						},
						NotificationSettings: types.NotificationSettings{
							SignedUp:      true,
							Quarter:       false,
							Half:          false,
							ThreeQuarters: true,
							Full:          true,
						},
						CreatedAt: mockTimestamp,
						UpdatedAt: mockTimestamp,
					},
				},
			},
			{
				name:   "Should fail if the user does not have any applications in the DB",
				userID: "test_not_real_user",
				err:    fmt.Errorf("Response not OK. 404 Not Found: applications not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				applicationsByUserID, err := ts.client1.GetApplicationsByUserID(testCtx, test.userID)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedApplications, applicationsByUserID)

				applicationsByUserID, err = ts.client2.GetApplicationsByUserID(testCtx, test.userID)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedApplications, applicationsByUserID)
			})
		}
	})

	ts.Run("Test_GetLoadBalancers", func() {
		tests := []struct {
			name                  string
			expectedLoadBalancers []*types.LoadBalancer
			err                   error
		}{
			{
				name: "Should fetch all load balancers in the database",
				expectedLoadBalancers: []*types.LoadBalancer{
					{
						ID:                "test_lb_34987u329rfn23f",
						Name:              "pokt_app_123",
						UserID:            "test_user_1dbffbdfeeb225",
						RequestTimeout:    5_000,
						Gigastake:         true,
						GigastakeRedirect: true,
						StickyOptions: types.StickyOptions{
							Duration:      "60",
							StickyOrigins: []string{"chrome-extension://", "moz-extension://"},
							StickyMax:     300,
							Stickiness:    true,
						},
						Applications: []*types.Application{
							{
								ID:     "test_app_47hfnths73j2se",
								UserID: "test_user_1dbffbdfeeb225",
								Name:   "pokt_app_123",
								URL:    "https://test.app123.io",
								Dummy:  true,
								Status: types.InService,
								GatewayAAT: types.GatewayAAT{
									Address:              "test_34715cae753e67c75fbb340442e7de8e",
									ApplicationPublicKey: "test_11b8d394ca331d7c7a71ca1896d630f6",
									ApplicationSignature: "test_89a3af6a587aec02cfade6f5000424c2",
									ClientPublicKey:      "test_1dc39a2e5a84a35bf030969a0b3231f7",
									PrivateKey:           "test_d2ce53f115f4ecb2208e9188800a85cf",
								},
								GatewaySettings: types.GatewaySettings{
									SecretKey:            "test_40f482d91a5ef2300ebb4e2308c",
									SecretKeyRequired:    true,
									WhitelistBlockchains: []string{"chain_1"},
									WhitelistOrigins:     []string{"origin_1", "origin_2"},
									WhitelistUserAgents:  []string{"user_agent_1", "user_agent_2", "user_agent_3"},
									WhitelistContracts: []types.WhitelistContracts{
										{BlockchainID: "TST2", Contracts: []string{"test_address_67890"}},
										{BlockchainID: "TST1", Contracts: []string{"test_address_12345", "test_address_44444"}},
									},
									WhitelistMethods: []types.WhitelistMethods{
										{BlockchainID: "TST1", Methods: []string{"GET"}},
										{BlockchainID: "TST2", Methods: []string{"PUT", "POST"}},
									},
								},
								Limit: types.AppLimit{
									PayPlan: types.PayPlan{Type: types.FreetierV0, Limit: 250_000},
								},
								NotificationSettings: types.NotificationSettings{
									SignedUp:      true,
									Quarter:       false,
									Half:          false,
									ThreeQuarters: true,
									Full:          true,
								},
								CreatedAt: mockTimestamp,
								UpdatedAt: mockTimestamp,
							},
						},
						Users: []types.UserAccess{
							{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
							{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
							{RoleName: types.RoleMember, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
						},
						CreatedAt: mockTimestamp,
						UpdatedAt: mockTimestamp,
					},
					{
						ID:                "test_lb_34gg4g43g34g5hh",
						Name:              "test_lb_redirect",
						UserID:            "test_user_redirect233344",
						RequestTimeout:    5_000,
						Gigastake:         false,
						GigastakeRedirect: false,
						StickyOptions: types.StickyOptions{
							Duration:      "20",
							StickyOrigins: []string{"test-extension://", "test-extension2://"},
							StickyMax:     600,
							Stickiness:    false,
						},
						Applications: []*types.Application{
							{
								ID:     "test_app_97djd63jd78r7f",
								UserID: "test_user_redirect233344",
								Name:   "pokt_app_789",
								URL:    "https://test.app789.io",
								Dummy:  true,
								Status: types.InService,
								GatewayAAT: types.GatewayAAT{
									Address:              "test_58a37510a103f3de57296af882b2603c",
									ApplicationPublicKey: "test_cc49729227c22f3934a966d99a6be72b",
									ApplicationSignature: "test_8d5f4a3006b5603bcecfb44a819c272a",
									ClientPublicKey:      "test_28a54385e49b65a860f61db62449703b",
									PrivateKey:           "test_8d5f4a3006b5603bcecfb44a819c272a",
								},
								GatewaySettings: types.GatewaySettings{
									SecretKey:         "test_8fd84jf74jf83kd83kd92okg355",
									SecretKeyRequired: false,
								},
								Limit: types.AppLimit{
									PayPlan: types.PayPlan{Type: types.PayAsYouGoV0, Limit: 0},
								},
								NotificationSettings: types.NotificationSettings{
									SignedUp:      true,
									Quarter:       false,
									Half:          false,
									ThreeQuarters: true,
									Full:          true,
								},
								CreatedAt: mockTimestamp,
								UpdatedAt: mockTimestamp,
							},
						},
						Users: []types.UserAccess{
							{RoleName: types.RoleOwner, UserID: "test_user_redirect233344", Email: "owner3@test.com", Accepted: true},
							{RoleName: types.RoleMember, UserID: "", Email: "member2@test.com", Accepted: false},
						},
						CreatedAt: mockTimestamp,
						UpdatedAt: mockTimestamp,
					},
					{
						ID:                "test_lb_3890ru23jfi32fj",
						Name:              "pokt_app_456",
						UserID:            "test_user_04228205bd261a",
						RequestTimeout:    5_000,
						Gigastake:         true,
						GigastakeRedirect: true,
						StickyOptions: types.StickyOptions{
							Duration:      "40",
							StickyOrigins: []string{"chrome-extension://"},
							StickyMax:     400,
							Stickiness:    true,
						},
						Applications: []*types.Application{
							{
								ID:     "test_app_5hdf7sh23jd828",
								UserID: "test_user_04228205bd261a",
								Name:   "pokt_app_456",
								URL:    "https://test.app456.io",
								Dummy:  true,
								Status: types.InService,
								GatewayAAT: types.GatewayAAT{
									Address:              "test_558c0225c7019e14ccf2e7379ad3eb50",
									ApplicationPublicKey: "test_96c981db344ab6920b7e87853838e285",
									ApplicationSignature: "test_1272a8ab4cbbf636f09bf4fa5395b885",
									ClientPublicKey:      "test_d709871777b89ed3051190f229ea3f01",
									PrivateKey:           "test_53e50765d8bc1fb41b3b0065dd8094de",
								},
								GatewaySettings: types.GatewaySettings{
									SecretKey:         "test_90210ac4bdd3423e24877d1ff92",
									SecretKeyRequired: false,
								},
								Limit: types.AppLimit{
									PayPlan:     types.PayPlan{Type: types.Enterprise},
									CustomLimit: 2_000_000,
								},
								NotificationSettings: types.NotificationSettings{
									SignedUp:      true,
									Quarter:       false,
									Half:          false,
									ThreeQuarters: true,
									Full:          true,
								},
								CreatedAt: mockTimestamp,
								UpdatedAt: mockTimestamp,
							},
						},
						Users: []types.UserAccess{
							{RoleName: types.RoleOwner, UserID: "test_user_04228205bd261a", Email: "owner2@test.com", Accepted: true},
							{RoleName: types.RoleAdmin, UserID: "test_user_admin5678", Email: "admin2@test.com", Accepted: true},
						},
						CreatedAt: mockTimestamp,
						UpdatedAt: mockTimestamp,
					},
				},
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				loadBalancers, err := ts.client1.GetLoadBalancers(testCtx)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedLoadBalancers, loadBalancers)

				loadBalancers, err = ts.client2.GetLoadBalancers(testCtx)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedLoadBalancers, loadBalancers)
			})
		}
	})

	ts.Run("Test_GetLoadBalancerByID", func() {
		tests := []struct {
			name                 string
			loadBalancerID       string
			expectedLoadBalancer *types.LoadBalancer
			err                  error
		}{
			{
				name:           "Should fetch one load balancer by ID",
				loadBalancerID: "test_lb_3890ru23jfi32fj",
				expectedLoadBalancer: &types.LoadBalancer{
					ID:                "test_lb_3890ru23jfi32fj",
					Name:              "pokt_app_456",
					UserID:            "test_user_04228205bd261a",
					RequestTimeout:    5_000,
					Gigastake:         true,
					GigastakeRedirect: true,
					StickyOptions: types.StickyOptions{
						Duration:      "40",
						StickyOrigins: []string{"chrome-extension://"},
						StickyMax:     400,
						Stickiness:    true,
					},
					Applications: []*types.Application{
						{
							ID:     "test_app_5hdf7sh23jd828",
							UserID: "test_user_04228205bd261a",
							Name:   "pokt_app_456",
							URL:    "https://test.app456.io",
							Dummy:  true,
							Status: types.InService,
							GatewayAAT: types.GatewayAAT{
								Address:              "test_558c0225c7019e14ccf2e7379ad3eb50",
								ApplicationPublicKey: "test_96c981db344ab6920b7e87853838e285",
								ApplicationSignature: "test_1272a8ab4cbbf636f09bf4fa5395b885",
								ClientPublicKey:      "test_d709871777b89ed3051190f229ea3f01",
								PrivateKey:           "test_53e50765d8bc1fb41b3b0065dd8094de",
							},
							GatewaySettings: types.GatewaySettings{
								SecretKey:         "test_90210ac4bdd3423e24877d1ff92",
								SecretKeyRequired: false,
							},
							Limit: types.AppLimit{
								PayPlan:     types.PayPlan{Type: types.Enterprise},
								CustomLimit: 2_000_000,
							},
							NotificationSettings: types.NotificationSettings{
								SignedUp:      true,
								Quarter:       false,
								Half:          false,
								ThreeQuarters: true,
								Full:          true,
							},
							CreatedAt: mockTimestamp,
							UpdatedAt: mockTimestamp,
						},
					},
					Users: []types.UserAccess{
						{RoleName: types.RoleOwner, UserID: "test_user_04228205bd261a", Email: "owner2@test.com", Accepted: true},
						{RoleName: types.RoleAdmin, UserID: "test_user_admin5678", Email: "admin2@test.com", Accepted: true},
					},
					CreatedAt: mockTimestamp,
					UpdatedAt: mockTimestamp,
				},
			},
			{
				name:           "Should fail if the load balancer does not exist in the DB",
				loadBalancerID: "test_not_real_load_balancer",
				err:            fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				loadBalancerByID, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedLoadBalancer, loadBalancerByID)

				loadBalancerByID, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedLoadBalancer, loadBalancerByID)
			})
		}
	})

	ts.Run("Test_GetLoadBalancersByUserID", func() {
		tests := []struct {
			name                  string
			userID                string
			expectedLoadBalancers []*types.LoadBalancer
			roleNameFilter        types.RoleName
			err                   error
		}{
			{
				name:   "Should fetch all load balancers for a single user ID when no filter provided",
				userID: "test_user_1dbffbdfeeb225",
				expectedLoadBalancers: []*types.LoadBalancer{
					{
						ID:                "test_lb_34987u329rfn23f",
						Name:              "pokt_app_123",
						UserID:            "test_user_1dbffbdfeeb225",
						RequestTimeout:    5_000,
						Gigastake:         true,
						GigastakeRedirect: true,
						StickyOptions: types.StickyOptions{
							Duration:      "60",
							StickyOrigins: []string{"chrome-extension://", "moz-extension://"},
							StickyMax:     300,
							Stickiness:    true,
						},
						Applications: []*types.Application{
							{
								ID:     "test_app_47hfnths73j2se",
								UserID: "test_user_1dbffbdfeeb225",
								Name:   "pokt_app_123",
								URL:    "https://test.app123.io",
								Dummy:  true,
								Status: types.InService,
								GatewayAAT: types.GatewayAAT{
									Address:              "test_34715cae753e67c75fbb340442e7de8e",
									ApplicationPublicKey: "test_11b8d394ca331d7c7a71ca1896d630f6",
									ApplicationSignature: "test_89a3af6a587aec02cfade6f5000424c2",
									ClientPublicKey:      "test_1dc39a2e5a84a35bf030969a0b3231f7",
									PrivateKey:           "test_d2ce53f115f4ecb2208e9188800a85cf",
								},
								GatewaySettings: types.GatewaySettings{
									SecretKey:         "test_40f482d91a5ef2300ebb4e2308c",
									SecretKeyRequired: true, WhitelistBlockchains: []string{"chain_1"},
									WhitelistOrigins:    []string{"origin_1", "origin_2"},
									WhitelistUserAgents: []string{"user_agent_1", "user_agent_2", "user_agent_3"},
									WhitelistContracts: []types.WhitelistContracts{
										{BlockchainID: "TST2", Contracts: []string{"test_address_67890"}},
										{BlockchainID: "TST1", Contracts: []string{"test_address_12345", "test_address_44444"}},
									},
									WhitelistMethods: []types.WhitelistMethods{
										{BlockchainID: "TST1", Methods: []string{"GET"}},
										{BlockchainID: "TST2", Methods: []string{"PUT", "POST"}},
									},
								},
								Limit: types.AppLimit{
									PayPlan: types.PayPlan{Type: types.FreetierV0, Limit: 250_000},
								},
								NotificationSettings: types.NotificationSettings{
									SignedUp:      true,
									Quarter:       false,
									Half:          false,
									ThreeQuarters: true,
									Full:          true,
								},
								CreatedAt: mockTimestamp,
								UpdatedAt: mockTimestamp,
							},
						},
						Users: []types.UserAccess{
							{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
							{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
							{RoleName: types.RoleMember, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
						},
						CreatedAt: mockTimestamp,
						UpdatedAt: mockTimestamp,
					},
				},
			},
			{
				name:           "Should fetch all load balancers for a single user ID and role when a valid filter provided",
				userID:         "test_user_admin1234",
				roleNameFilter: types.RoleAdmin,
				expectedLoadBalancers: []*types.LoadBalancer{
					{
						ID:                "test_lb_34987u329rfn23f",
						Name:              "pokt_app_123",
						UserID:            "test_user_1dbffbdfeeb225",
						RequestTimeout:    5_000,
						Gigastake:         true,
						GigastakeRedirect: true,
						StickyOptions: types.StickyOptions{
							Duration:      "60",
							StickyOrigins: []string{"chrome-extension://", "moz-extension://"},
							StickyMax:     300,
							Stickiness:    true,
						},
						Applications: []*types.Application{
							{
								ID:     "test_app_47hfnths73j2se",
								UserID: "test_user_1dbffbdfeeb225",
								Name:   "pokt_app_123",
								URL:    "https://test.app123.io",
								Dummy:  true,
								Status: types.InService,
								GatewayAAT: types.GatewayAAT{
									Address:              "test_34715cae753e67c75fbb340442e7de8e",
									ApplicationPublicKey: "test_11b8d394ca331d7c7a71ca1896d630f6",
									ApplicationSignature: "test_89a3af6a587aec02cfade6f5000424c2",
									ClientPublicKey:      "test_1dc39a2e5a84a35bf030969a0b3231f7",
									PrivateKey:           "test_d2ce53f115f4ecb2208e9188800a85cf",
								},
								GatewaySettings: types.GatewaySettings{
									SecretKey:            "test_40f482d91a5ef2300ebb4e2308c",
									SecretKeyRequired:    true,
									WhitelistBlockchains: []string{"chain_1"},
									WhitelistOrigins:     []string{"origin_1", "origin_2"},
									WhitelistUserAgents:  []string{"user_agent_1", "user_agent_2", "user_agent_3"},
									WhitelistContracts: []types.WhitelistContracts{
										{BlockchainID: "TST2", Contracts: []string{"test_address_67890"}},
										{BlockchainID: "TST1", Contracts: []string{"test_address_12345", "test_address_44444"}},
									},
									WhitelistMethods: []types.WhitelistMethods{
										{BlockchainID: "TST1", Methods: []string{"GET"}},
										{BlockchainID: "TST2", Methods: []string{"PUT", "POST"}},
									},
								},
								Limit: types.AppLimit{
									PayPlan: types.PayPlan{Type: types.FreetierV0, Limit: 250_000},
								},
								NotificationSettings: types.NotificationSettings{
									SignedUp:      true,
									Quarter:       false,
									Half:          false,
									ThreeQuarters: true,
									Full:          true,
								},
								CreatedAt: mockTimestamp,
								UpdatedAt: mockTimestamp,
							},
						},
						Users: []types.UserAccess{
							{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
							{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
							{RoleName: types.RoleMember, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
						},
						CreatedAt: mockTimestamp,
						UpdatedAt: mockTimestamp,
					},
				},
			},
			{
				name:           "Should fail if an invalid role name provided as a filter",
				userID:         "test_user_1dbffbdfeeb225",
				roleNameFilter: types.RoleName("not_real"),
				err:            fmt.Errorf("invalid role name filter"),
			},
			{
				name:   "Should fail if the user does not have any load balancers in the DB",
				userID: "test_not_real_user",
				err:    fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				filter := &test.roleNameFilter
				if test.roleNameFilter == "" {
					filter = nil
				}

				loadBalancersByUserID, err := ts.client1.GetLoadBalancersByUserID(testCtx, test.userID, filter)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedLoadBalancers, loadBalancersByUserID)

				loadBalancersByUserID, err = ts.client2.GetLoadBalancersByUserID(testCtx, test.userID, filter)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedLoadBalancers, loadBalancersByUserID)
			})
		}
	})

	ts.Run("Test_GetPendingLoadBalancersByEmail", func() {
		tests := []struct {
			name                  string
			userEmail             string
			expectedLoadBalancers []*types.LoadBalancer
			err                   error
		}{
			{
				name:      "Should fetch all pending load balancers for a single user email",
				userEmail: "member2@test.com",
				expectedLoadBalancers: []*types.LoadBalancer{
					{
						ID:                "test_lb_34gg4g43g34g5hh",
						Name:              "test_lb_redirect",
						UserID:            "test_user_redirect233344",
						RequestTimeout:    5_000,
						Gigastake:         false,
						GigastakeRedirect: false,
						StickyOptions: types.StickyOptions{
							Duration:      "20",
							StickyOrigins: []string{"test-extension://", "test-extension2://"},
							StickyMax:     600,
							Stickiness:    false,
						},
						Applications: []*types.Application{
							{
								ID:     "test_app_97djd63jd78r7f",
								UserID: "test_user_redirect233344",
								Name:   "pokt_app_789",
								URL:    "https://test.app789.io",
								Dummy:  true,
								Status: types.InService,
								GatewayAAT: types.GatewayAAT{
									Address:              "test_58a37510a103f3de57296af882b2603c",
									ApplicationPublicKey: "test_cc49729227c22f3934a966d99a6be72b",
									ApplicationSignature: "test_8d5f4a3006b5603bcecfb44a819c272a",
									ClientPublicKey:      "test_28a54385e49b65a860f61db62449703b",
									PrivateKey:           "test_8d5f4a3006b5603bcecfb44a819c272a",
								},
								GatewaySettings: types.GatewaySettings{
									SecretKey:         "test_8fd84jf74jf83kd83kd92okg355",
									SecretKeyRequired: false,
								},
								Limit: types.AppLimit{
									PayPlan: types.PayPlan{Type: types.PayAsYouGoV0, Limit: 0},
								},
								NotificationSettings: types.NotificationSettings{
									SignedUp:      true,
									Quarter:       false,
									Half:          false,
									ThreeQuarters: true,
									Full:          true,
								},
								CreatedAt: mockTimestamp,
								UpdatedAt: mockTimestamp,
							},
						},
						Users: []types.UserAccess{
							{RoleName: types.RoleOwner, UserID: "test_user_redirect233344", Email: "owner3@test.com", Accepted: true},
							{RoleName: types.RoleMember, UserID: "", Email: "member2@test.com", Accepted: false},
						},
						CreatedAt: mockTimestamp,
						UpdatedAt: mockTimestamp,
					},
				},
			},
			{
				name:      "Should fail if user email not provided",
				userEmail: "",
				err:       fmt.Errorf("no user email"),
			},
			{
				name:      "Should fail if the email does not have any pending load balancers in the DB",
				userEmail: "test_not_real@user.com",
				err:       fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				pendingLoadBalancersByEmail, err := ts.client1.GetPendingLoadBalancersByEmail(testCtx, test.userEmail)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedLoadBalancers, pendingLoadBalancersByEmail)

				pendingLoadBalancersByEmail, err = ts.client2.GetPendingLoadBalancersByEmail(testCtx, test.userEmail)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedLoadBalancers, pendingLoadBalancersByEmail)
			})
		}
	})

	ts.Run("Test_GetPayPlans", func() {
		tests := []struct {
			name             string
			expectedPayPlans []*types.PayPlan
			err              error
		}{
			{
				name: "Should fetch all pay plans in the DB",
				expectedPayPlans: []*types.PayPlan{
					{Type: types.Enterprise, Limit: 0},
					{Type: types.FreetierV0, Limit: 250000},
					{Type: types.PayAsYouGoV0, Limit: 0},
					{Type: types.TestPlan10K, Limit: 10000},
					{Type: types.TestPlan90k, Limit: 90000},
					{Type: types.TestPlanV0, Limit: 100},
				},
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				payPlans, err := ts.client1.GetPayPlans(testCtx)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedPayPlans, payPlans)

				payPlans, err = ts.client2.GetPayPlans(testCtx)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedPayPlans, payPlans)
			})
		}
	})

	ts.Run("Test_GetPayPlanByType", func() {
		tests := []struct {
			name            string
			payPlanType     types.PayPlanType
			expectedPayPlan *types.PayPlan
			err             error
		}{
			{

				name:        "Should fetch a single pay plan by type",
				payPlanType: types.FreetierV0,
				expectedPayPlan: &types.PayPlan{
					Type: types.FreetierV0, Limit: 250000,
				},
			},
			{
				name:        "Should fail if passed a pay plan type that is not in the DB",
				payPlanType: types.PayPlanType("not_a_real_plan"),
				err:         fmt.Errorf("Response not OK. 404 Not Found: pay plan not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				payPlanByType, err := ts.client1.GetPayPlanByType(testCtx, test.payPlanType)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedPayPlan, payPlanByType)

				payPlanByType, err = ts.client2.GetPayPlanByType(testCtx, test.payPlanType)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedPayPlan, payPlanByType)
			})
		}
	})

	ts.Run("Test_GetUserPermissionsByUserID", func() {
		tests := []struct {
			name                string
			userID              types.UserID
			expectedPermissions *types.UserPermissions
			err                 error
		}{
			{

				name:   "Should fetch a single users load balancer permissions",
				userID: "test_user_1dbffbdfeeb225",
				expectedPermissions: &types.UserPermissions{
					UserID: "test_user_1dbffbdfeeb225",
					LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{
						"test_lb_34987u329rfn23f": {
							RoleName:    types.RoleOwner,
							Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint, types.DeleteEndpoint, types.TransferEndpoint},
						},
					},
				},
			},
			{

				name:   "Should fetch another single users load balancer permissions",
				userID: "test_user_member1234",
				expectedPermissions: &types.UserPermissions{
					UserID: "test_user_member1234",
					LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{
						"test_lb_34987u329rfn23f": {
							RoleName:    types.RoleMember,
							Permissions: []types.PermissionsEnum{types.ReadEndpoint},
						},
					},
				},
			},
			{
				name:   "Should fail is the user does not have any permissions",
				userID: "test_user_hey_who_am_i_wow",
				err:    fmt.Errorf("Response not OK. 404 Not Found: no permissions found for given user ID"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				permissionsByUserID, err := ts.client1.GetUserPermissionsByUserID(testCtx, test.userID)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedPermissions, permissionsByUserID)

				permissionsByUserID, err = ts.client2.GetUserPermissionsByUserID(testCtx, test.userID)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedPermissions, permissionsByUserID)
			})
		}
	})

	ts.Run("Test_GetLoadBalancersCountByEmail", func() {
		tests := []struct {
			name          string
			email         string
			expectedCount int
			err           error
		}{
			{
				name:          "Should return the count of load balancer owned by an email",
				email:         "owner1@test.com",
				expectedCount: 1,
			},
			{
				name:          "Should return the count of load balancer owned by an email (the remix)",
				email:         "owner2@test.com",
				expectedCount: 1,
			},
			{
				name:  "Should return 0 if there's no lb owned by this email",
				email: "ownerfake@test.com",
			},
			{
				name:  "Should return missing email if there's no email",
				email: "",
				err:   fmt.Errorf("no user email"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				count, err := ts.client1.GetLoadBalancersCountByEmail(testCtx, test.email)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedCount, count)

				count, err = ts.client2.GetLoadBalancersCountByEmail(testCtx, test.email)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedCount, count)
			})
		}
	})
}

// Runs all the write endpoint tests after the read tests
// This ensures the write tests to not modify the seed data expected by the read tests
func (ts *phdE2ETestSuite) Test_WriteTests() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.Run("Test_CreateBlockchain", func() {
		tests := []struct {
			name                        string
			blockchainInput, blockchain types.Blockchain
			err                         error
		}{
			{
				name: "Should create a single blockchain in the DB",
				blockchainInput: types.Blockchain{
					ID:                "003",
					Altruist:          "https://test:test_fg332f@shared-test3.nodes.pol.network:12345",
					Blockchain:        "pol-mainnet",
					Description:       "Polygon Mainnet",
					EnforceResult:     "JSON",
					Ticker:            "POL",
					BlockchainAliases: []string{"pol-mainnet"},
					LogLimitBlocks:    100000,
					Active:            true,
					SyncCheckOptions: types.SyncCheckOptions{
						Body:      "{}",
						ResultKey: "result",
						Allowance: 3,
					},
				},
				blockchain: types.Blockchain{
					ID:                "003",
					Altruist:          "https://test:test_fg332f@shared-test3.nodes.pol.network:12345",
					Blockchain:        "pol-mainnet",
					Description:       "Polygon Mainnet",
					EnforceResult:     "JSON",
					Ticker:            "POL",
					BlockchainAliases: []string{"pol-mainnet"},
					LogLimitBlocks:    100000,
					Active:            true,
					SyncCheckOptions: types.SyncCheckOptions{
						BlockchainID: "003",
						Body:         "{}",
						ResultKey:    "result",
						Allowance:    3,
					},
				},
			},
			{
				name: "Should fail if attempting to create a duplicate record",
				blockchainInput: types.Blockchain{
					ID:                "003",
					Altruist:          "https://test:test_fg332f@shared-test3.nodes.pol.network:12345",
					Blockchain:        "pol-mainnet",
					Description:       "Polygon Mainnet",
					EnforceResult:     "JSON",
					Ticker:            "POL",
					BlockchainAliases: []string{"pol-mainnet"},
					LogLimitBlocks:    100000,
					Active:            true,
					SyncCheckOptions: types.SyncCheckOptions{
						Body:      "{}",
						ResultKey: "result",
						Allowance: 3,
					},
				},
				err: fmt.Errorf("Response not OK. 500 Internal Server Error: pq: duplicate key value violates unique constraint \"blockchains_pkey\""),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.CreateBlockchain(testCtx, test.blockchainInput)
				ts.Equal(test.err, err)
				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					blockchain, err := ts.client1.GetBlockchainByID(testCtx, test.blockchainInput.ID)
					ts.Equal(test.err, err)
					ts.Equal(test.blockchain.ID, blockchain.ID)
					ts.Equal(test.blockchain.Altruist, blockchain.Altruist)
					ts.Equal(test.blockchain.Blockchain, blockchain.Blockchain)
					ts.Equal(test.blockchain.ChainID, blockchain.ChainID)
					ts.Equal(test.blockchain.ChainIDCheck, blockchain.ChainIDCheck)
					ts.Equal(test.blockchain.Description, blockchain.Description)
					ts.Equal(test.blockchain.EnforceResult, blockchain.EnforceResult)
					ts.Equal(test.blockchain.Path, blockchain.Path)
					ts.Equal(test.blockchain.Ticker, blockchain.Ticker)
					ts.Equal(test.blockchain.BlockchainAliases, blockchain.BlockchainAliases)
					ts.Equal(test.blockchain.LogLimitBlocks, blockchain.LogLimitBlocks)
					ts.Equal(test.blockchain.RequestTimeout, blockchain.RequestTimeout)
					ts.Equal(test.blockchain.Active, blockchain.Active)
					ts.Equal(test.blockchain.SyncCheckOptions, blockchain.SyncCheckOptions)
					ts.NotEmpty(blockchain.CreatedAt)
					ts.NotEmpty(blockchain.UpdatedAt)

					blockchain, err = ts.client2.GetBlockchainByID(testCtx, test.blockchainInput.ID)
					ts.Equal(test.err, err)
					ts.Equal(test.blockchain.ID, blockchain.ID)
					ts.Equal(test.blockchain.Altruist, blockchain.Altruist)
					ts.Equal(test.blockchain.Blockchain, blockchain.Blockchain)
					ts.Equal(test.blockchain.ChainID, blockchain.ChainID)
					ts.Equal(test.blockchain.ChainIDCheck, blockchain.ChainIDCheck)
					ts.Equal(test.blockchain.Description, blockchain.Description)
					ts.Equal(test.blockchain.EnforceResult, blockchain.EnforceResult)
					ts.Equal(test.blockchain.Path, blockchain.Path)
					ts.Equal(test.blockchain.Ticker, blockchain.Ticker)
					ts.Equal(test.blockchain.BlockchainAliases, blockchain.BlockchainAliases)
					ts.Equal(test.blockchain.LogLimitBlocks, blockchain.LogLimitBlocks)
					ts.Equal(test.blockchain.RequestTimeout, blockchain.RequestTimeout)
					ts.Equal(test.blockchain.Active, blockchain.Active)
					ts.Equal(test.blockchain.SyncCheckOptions, blockchain.SyncCheckOptions)
					ts.NotEmpty(blockchain.CreatedAt)
					ts.NotEmpty(blockchain.UpdatedAt)
				}
			})
		}
	})

	ts.Run("Test_CreateRedirect", func() {
		tests := []struct {
			name          string
			redirectInput types.Redirect
			redirects     []types.Redirect
			err           error
		}{
			{
				name: "Should create a new redirect for an existing blockchain in the DB",
				redirectInput: types.Redirect{
					BlockchainID:   "0001",
					Alias:          "test-mainnet-3",
					Domain:         "test-rpc3.testnet.pokt.network",
					LoadBalancerID: "test_lb_34gg4g43g34g5hh",
				},
				redirects: []types.Redirect{
					{
						Alias:          "test-mainnet",
						Domain:         "test-rpc1.testnet.pokt.network",
						LoadBalancerID: "test_lb_34gg4g43g34g5hh",
					},
					{
						Alias:          "test-mainnet",
						Domain:         "test-rpc2.testnet.pokt.network",
						LoadBalancerID: "test_lb_34gg4g43g34g5hh",
					},
					{
						Alias:          "test-mainnet-3",
						Domain:         "test-rpc3.testnet.pokt.network",
						LoadBalancerID: "test_lb_34gg4g43g34g5hh",
					},
				},
			},
			{
				name: "Should fail if attempting to create a duplicate record",
				redirectInput: types.Redirect{
					BlockchainID:   "0001",
					Alias:          "test-mainnet-3",
					Domain:         "test-rpc3.testnet.pokt.network",
					LoadBalancerID: "test_lb_34gg4g43g34g5hh",
				},
				err: fmt.Errorf("Response not OK. 500 Internal Server Error: pq: duplicate key value violates unique constraint \"redirects_blockchain_id_domain_key\""),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.CreateBlockchainRedirect(testCtx, test.redirectInput)
				ts.Equal(test.err, err)
				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					blockchain, err := ts.client1.GetBlockchainByID(testCtx, test.redirectInput.BlockchainID)
					ts.Equal(test.err, err)
					ts.Len(blockchain.Redirects, len(test.redirects))
					for i, redirect := range blockchain.Redirects {
						ts.Equal(test.redirects[i].Alias, redirect.Alias)
						ts.Equal(test.redirects[i].Domain, redirect.Domain)
						ts.Equal(test.redirects[i].LoadBalancerID, redirect.LoadBalancerID)
					}

					blockchain, err = ts.client2.GetBlockchainByID(testCtx, test.redirectInput.BlockchainID)
					ts.Equal(test.err, err)
					ts.Len(blockchain.Redirects, len(test.redirects))
					for i, redirect := range blockchain.Redirects {
						ts.Equal(test.redirects[i].Alias, redirect.Alias)
						ts.Equal(test.redirects[i].Domain, redirect.Domain)
						ts.Equal(test.redirects[i].LoadBalancerID, redirect.LoadBalancerID)
					}
				}
			})
		}
	})

	ts.Run("Test_CreateApplication", func() {
		tests := []struct {
			name                          string
			applicationInput, application types.Application
			err                           error
		}{
			{
				name: "Should create a single application in the DB",
				applicationInput: types.Application{
					Name:   "pokt_app_101",
					UserID: "test_user_47fhsd75jd756sh",
					Dummy:  true,
					Status: types.InService,
					GatewayAAT: types.GatewayAAT{
						Address:              "test_e209a2d1f3454ddc69cb9333d547bbcf",
						ApplicationPublicKey: "test_b95c35affacf6df4a5585388490542f0",
						ApplicationSignature: "test_e59760339d9ce02972d1080d73446c90",
						ClientPublicKey:      "test_d591178ab3f48f45b243303fe77dc8c3",
						PrivateKey:           "test_f403700aed7e039c0a8fc2dd22da6fd9",
					},
					GatewaySettings: types.GatewaySettings{
						SecretKey:         "test_489574398f34uhf4uhjf9328jf23f98j",
						SecretKeyRequired: true,
					},
					Limit: types.AppLimit{
						PayPlan: types.PayPlan{Type: types.FreetierV0},
					},
					NotificationSettings: types.NotificationSettings{SignedUp: true, Quarter: false, Half: false, ThreeQuarters: true, Full: true},
				},
				application: types.Application{
					Name:   "pokt_app_101",
					UserID: "test_user_47fhsd75jd756sh",
					Dummy:  true,
					Status: types.InService,
					GatewayAAT: types.GatewayAAT{
						Address:              "test_e209a2d1f3454ddc69cb9333d547bbcf",
						ApplicationPublicKey: "test_b95c35affacf6df4a5585388490542f0",
						ApplicationSignature: "test_e59760339d9ce02972d1080d73446c90",
						ClientPublicKey:      "test_d591178ab3f48f45b243303fe77dc8c3",
						PrivateKey:           "test_f403700aed7e039c0a8fc2dd22da6fd9",
					},
					GatewaySettings: types.GatewaySettings{
						SecretKey:         "test_489574398f34uhf4uhjf9328jf23f98j",
						SecretKeyRequired: true,
					},
					Limit: types.AppLimit{
						PayPlan: types.PayPlan{Type: types.FreetierV0, Limit: 250_000},
					},
					NotificationSettings: types.NotificationSettings{SignedUp: true, Quarter: false, Half: false, ThreeQuarters: true, Full: true},
				},
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				createdApp, err := ts.client1.CreateApplication(testCtx, test.applicationInput)
				ts.Equal(test.err, err)
				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					application, err := ts.client1.GetApplicationByID(testCtx, createdApp.ID)
					ts.Equal(test.err, err)
					ts.Equal(createdApp.ID, application.ID)
					ts.Equal(test.application.Dummy, application.Dummy)
					ts.Equal(test.application.Status, application.Status)
					ts.Equal(test.application.GatewayAAT, application.GatewayAAT)
					ts.Equal(test.application.GatewaySettings, application.GatewaySettings)
					ts.Equal(test.application.Limit, application.Limit)
					ts.Equal(test.application.NotificationSettings, application.NotificationSettings)
					ts.NotEmpty(application.CreatedAt)
					ts.NotEmpty(application.UpdatedAt)

					application, err = ts.client2.GetApplicationByID(testCtx, createdApp.ID)
					ts.Equal(test.err, err)
					ts.Equal(createdApp.ID, application.ID)
					ts.Equal(test.application.Dummy, application.Dummy)
					ts.Equal(test.application.Status, application.Status)
					ts.Equal(test.application.GatewayAAT, application.GatewayAAT)
					ts.Equal(test.application.GatewaySettings, application.GatewaySettings)
					ts.Equal(test.application.Limit, application.Limit)
					ts.Equal(test.application.NotificationSettings, application.NotificationSettings)
					ts.NotEmpty(application.CreatedAt)
					ts.NotEmpty(application.UpdatedAt)
				}
			})
		}
	})

	ts.Run("Test_CreateLoadBalancer", func() {
		tests := []struct {
			name                            string
			loadBalancerInput, loadBalancer types.LoadBalancer
			err                             error
		}{
			{
				name: "Should create a single loadBalancer in the DB",
				loadBalancerInput: types.LoadBalancer{
					Name:              "pokt_app_789",
					UserID:            "test_user_47fhsd75jd756sh",
					RequestTimeout:    5000,
					Gigastake:         true,
					GigastakeRedirect: true,
					ApplicationIDs:    []string{"test_app_47hfnths73j2se"},
					StickyOptions: types.StickyOptions{
						Duration:      "70",
						StickyOrigins: []string{"chrome-extension://"},
						StickyMax:     400,
						Stickiness:    true,
					},
					Users: []types.UserAccess{
						{
							UserID:   "test_user_47fhsd75jd756sh",
							RoleName: types.RoleOwner,
							Email:    "owner4@test.com",
							Accepted: true,
						},
					},
				},
				loadBalancer: types.LoadBalancer{
					Name:              "pokt_app_789",
					UserID:            "test_user_47fhsd75jd756sh",
					RequestTimeout:    5000,
					Gigastake:         true,
					GigastakeRedirect: true,
					Applications: []*types.Application{
						{
							ID:     "test_app_47hfnths73j2se",
							UserID: "test_user_1dbffbdfeeb225",
							Name:   "pokt_app_123",
							URL:    "https://test.app123.io",
							Dummy:  true,
							Status: types.InService,
							GatewayAAT: types.GatewayAAT{
								Address:              "test_34715cae753e67c75fbb340442e7de8e",
								ApplicationPublicKey: "test_11b8d394ca331d7c7a71ca1896d630f6",
								ApplicationSignature: "test_89a3af6a587aec02cfade6f5000424c2",
								ClientPublicKey:      "test_1dc39a2e5a84a35bf030969a0b3231f7",
								PrivateKey:           "test_d2ce53f115f4ecb2208e9188800a85cf",
							},
							GatewaySettings: types.GatewaySettings{
								SecretKey:            "test_40f482d91a5ef2300ebb4e2308c",
								SecretKeyRequired:    true,
								WhitelistBlockchains: []string{"chain_1"},
								WhitelistOrigins:     []string{"origin_1", "origin_2"},
								WhitelistUserAgents:  []string{"user_agent_1", "user_agent_2", "user_agent_3"},
								WhitelistContracts: []types.WhitelistContracts{
									{BlockchainID: "TST2", Contracts: []string{"test_address_67890"}},
									{BlockchainID: "TST1", Contracts: []string{"test_address_12345", "test_address_44444"}},
								},
								WhitelistMethods: []types.WhitelistMethods{
									{BlockchainID: "TST1", Methods: []string{"GET"}},
									{BlockchainID: "TST2", Methods: []string{"PUT", "POST"}},
								},
							},
							Limit: types.AppLimit{
								PayPlan: types.PayPlan{Type: types.FreetierV0, Limit: 250_000},
							},
							NotificationSettings: types.NotificationSettings{
								SignedUp:      true,
								Quarter:       false,
								Half:          false,
								ThreeQuarters: true,
								Full:          true,
							},
							CreatedAt: mockTimestamp,
							UpdatedAt: mockTimestamp,
						},
					},
					StickyOptions: types.StickyOptions{
						Duration:      "70",
						StickyOrigins: []string{"chrome-extension://"},
						StickyMax:     400,
						Stickiness:    true,
					},
					Users: []types.UserAccess{
						{
							UserID:   "test_user_47fhsd75jd756sh",
							RoleName: types.RoleOwner,
							Email:    "owner4@test.com",
							Accepted: true,
						},
					},
				},
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				createdLB, err := ts.client1.CreateLoadBalancer(testCtx, test.loadBalancerInput)
				ts.Equal(test.err, err)
				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, createdLB.ID)
					ts.Equal(test.err, err)
					ts.Equal(createdLB.ID, loadBalancer.ID)
					ts.Equal(test.loadBalancer.UserID, loadBalancer.UserID)
					ts.Equal(test.loadBalancer.Name, loadBalancer.Name)
					ts.Equal(test.loadBalancer.UserID, loadBalancer.UserID)
					ts.Equal(test.loadBalancer.RequestTimeout, loadBalancer.RequestTimeout)
					ts.Equal(test.loadBalancer.Gigastake, loadBalancer.Gigastake)
					ts.Equal(test.loadBalancer.GigastakeRedirect, loadBalancer.GigastakeRedirect)
					ts.Equal(test.loadBalancer.ApplicationIDs, loadBalancer.ApplicationIDs)
					ts.Equal(test.loadBalancer.Applications, loadBalancer.Applications)
					ts.Equal(test.loadBalancer.StickyOptions, loadBalancer.StickyOptions)
					ts.Equal(test.loadBalancer.Users, loadBalancer.Users)
					ts.NotEmpty(loadBalancer.CreatedAt)
					ts.NotEmpty(loadBalancer.UpdatedAt)

					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, createdLB.ID)
					ts.Equal(test.err, err)
					ts.Equal(createdLB.ID, loadBalancer.ID)
					ts.Equal(test.loadBalancer.UserID, loadBalancer.UserID)
					ts.Equal(test.loadBalancer.Name, loadBalancer.Name)
					ts.Equal(test.loadBalancer.UserID, loadBalancer.UserID)
					ts.Equal(test.loadBalancer.RequestTimeout, loadBalancer.RequestTimeout)
					ts.Equal(test.loadBalancer.Gigastake, loadBalancer.Gigastake)
					ts.Equal(test.loadBalancer.GigastakeRedirect, loadBalancer.GigastakeRedirect)
					ts.Equal(test.loadBalancer.ApplicationIDs, loadBalancer.ApplicationIDs)
					ts.Equal(test.loadBalancer.Applications, loadBalancer.Applications)
					ts.Equal(test.loadBalancer.StickyOptions, loadBalancer.StickyOptions)
					ts.Equal(test.loadBalancer.Users, loadBalancer.Users)
					ts.NotEmpty(loadBalancer.CreatedAt)
					ts.NotEmpty(loadBalancer.UpdatedAt)
				}
			})
		}
	})

	ts.Run("Test_CreateLoadBalancerUser", func() {
		tests := []struct {
			name              string
			loadBalancerID    string
			user              types.UserAccess
			loadBalancerUsers []types.UserAccess
			err               error
		}{
			{
				name:           "Should add a single user to an existing load balancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				user: types.UserAccess{
					RoleName: types.RoleMember,
					Email:    "member_new@test.com",
				},
				loadBalancerUsers: []types.UserAccess{
					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "", Email: "member_new@test.com", Accepted: false},
				},
			},
			{
				name:           "Should fail if load balancer cannot be found",
				loadBalancerID: "sir_not_appearing_in_this_film",

				err: fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.CreateLoadBalancerUser(testCtx, test.loadBalancerID, test.user)
				ts.Equal(test.err, err)
				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.Equal(test.err, err)
					ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.Equal(test.err, err)
					ts.Equal(test.loadBalancerUsers, loadBalancer.Users)
				}
			})
		}
	})

	ts.Run("Test_ActivateBlockchain", func() {
		tests := []struct {
			name         string
			blockchainID string
			active       bool
			err          error
		}{
			{
				name:         "Should set a blockchain's active field to false",
				blockchainID: "0001",
				active:       false,
			},
			{
				name:         "Should set a blockchain's active field to true",
				blockchainID: "0001",
				active:       true,
			},
			{
				name:         "Should fail if blockchain cannot be found",
				blockchainID: "5440",
				err:          fmt.Errorf("Response not OK. 404 Not Found: blockchain not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.ActivateBlockchain(testCtx, test.blockchainID, test.active)
				ts.Equal(test.err, err)
				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					blockchain, err := ts.client1.GetBlockchainByID(testCtx, test.blockchainID)
					ts.Equal(test.err, err)
					ts.Equal(test.active, blockchain.Active)
				}
			})
		}
	})

	ts.Run("Test_UpdateApplication", func() {
		tests := []struct {
			name                   string
			applicationID          string
			applicationUpdate      types.UpdateApplication
			applicationAfterUpdate types.Application
			err                    error
		}{
			{
				name:          "Should update a single application in the DB",
				applicationID: "test_app_47hfnths73j2se",
				applicationUpdate: types.UpdateApplication{
					Name: "pokt_app_updated_lb",
					GatewaySettings: &types.UpdateGatewaySettings{
						WhitelistOrigins:     []string{"test-origin1", "test-origin2"},
						WhitelistUserAgents:  []string{"test-agent1"},
						WhitelistContracts:   []types.WhitelistContracts{{BlockchainID: "01", Contracts: []string{"test-contract1"}}},
						WhitelistMethods:     []types.WhitelistMethods{{BlockchainID: "01", Methods: []string{"test-method1"}}},
						WhitelistBlockchains: []string{"test-chain1"},
					},
					NotificationSettings: &types.UpdateNotificationSettings{SignedUp: boolPointer(false), Quarter: boolPointer(true), Half: boolPointer(true), ThreeQuarters: boolPointer(false), Full: boolPointer(false)},
					Limit:                &types.AppLimit{PayPlan: types.PayPlan{Type: types.Enterprise}, CustomLimit: 4_200_000},
				},
				applicationAfterUpdate: types.Application{
					Name: "pokt_app_updated_lb",
					GatewaySettings: types.GatewaySettings{
						SecretKey:            "test_40f482d91a5ef2300ebb4e2308c",
						SecretKeyRequired:    true,
						WhitelistOrigins:     []string{"test-origin1", "test-origin2"},
						WhitelistUserAgents:  []string{"test-agent1"},
						WhitelistContracts:   []types.WhitelistContracts{{BlockchainID: "01", Contracts: []string{"test-contract1"}}},
						WhitelistMethods:     []types.WhitelistMethods{{BlockchainID: "01", Methods: []string{"test-method1"}}},
						WhitelistBlockchains: []string{"test-chain1"},
					},
					NotificationSettings: types.NotificationSettings{SignedUp: false, Quarter: true, Half: true, ThreeQuarters: false, Full: false},
					Limit:                types.AppLimit{PayPlan: types.PayPlan{Type: types.Enterprise}, CustomLimit: 4_200_000},
				},
			},
			{
				name:          "Should update a single application's pay plan type",
				applicationID: "test_app_47hfnths73j2se",
				applicationUpdate: types.UpdateApplication{
					Limit: &types.AppLimit{
						PayPlan: types.PayPlan{
							Type: types.FreetierV0,
						},
					},
				},
				applicationAfterUpdate: types.Application{
					Name: "pokt_app_updated_lb",
					GatewaySettings: types.GatewaySettings{
						SecretKey:            "test_40f482d91a5ef2300ebb4e2308c",
						SecretKeyRequired:    true,
						WhitelistOrigins:     []string{"test-origin1", "test-origin2"},
						WhitelistUserAgents:  []string{"test-agent1"},
						WhitelistContracts:   []types.WhitelistContracts{{BlockchainID: "01", Contracts: []string{"test-contract1"}}},
						WhitelistMethods:     []types.WhitelistMethods{{BlockchainID: "01", Methods: []string{"test-method1"}}},
						WhitelistBlockchains: []string{"test-chain1"}},
					NotificationSettings: types.NotificationSettings{SignedUp: false, Quarter: true, Half: true, ThreeQuarters: false, Full: false},
					Limit: types.AppLimit{
						PayPlan: types.PayPlan{
							Type:  types.FreetierV0,
							Limit: 250_000,
						},
						CustomLimit: 0,
					},
				},
			},
			{
				name:          "Should remove all of a  single application's whitelists",
				applicationID: "test_app_47hfnths73j2se",
				applicationUpdate: types.UpdateApplication{
					GatewaySettings: &types.UpdateGatewaySettings{},
				},
				applicationAfterUpdate: types.Application{
					Name: "pokt_app_updated_lb",
					GatewaySettings: types.GatewaySettings{
						SecretKey:         "test_40f482d91a5ef2300ebb4e2308c",
						SecretKeyRequired: true,
					},
					NotificationSettings: types.NotificationSettings{SignedUp: false, Quarter: true, Half: true, ThreeQuarters: false, Full: false},
					Limit: types.AppLimit{
						PayPlan: types.PayPlan{
							Type:  types.FreetierV0,
							Limit: 250_000,
						},
						CustomLimit: 0,
					},
				},
			},
			{
				name:          "Should fail if application cannot be found",
				applicationID: "9000",
				err:           fmt.Errorf("Response not OK. 404 Not Found: applications not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				createdApp, err := ts.client1.UpdateApplication(testCtx, test.applicationID, test.applicationUpdate)
				ts.Equal(test.err, err)
				if err == nil {
					<-time.After(50 * time.Millisecond)
					application, err := ts.client1.GetApplicationByID(testCtx, createdApp.ID)
					ts.NoError(err)
					ts.Equal(test.applicationAfterUpdate.Name, application.Name)
					ts.Equal(test.applicationAfterUpdate.GatewaySettings, application.GatewaySettings)
					ts.Equal(test.applicationAfterUpdate.NotificationSettings, application.NotificationSettings)
					ts.Equal(test.applicationAfterUpdate.Limit, application.Limit)

					application, err = ts.client2.GetApplicationByID(testCtx, createdApp.ID)
					ts.NoError(err)
					ts.Equal(test.applicationAfterUpdate.Name, application.Name)
					ts.Equal(test.applicationAfterUpdate.GatewaySettings, application.GatewaySettings)
					ts.Equal(test.applicationAfterUpdate.NotificationSettings, application.NotificationSettings)
					ts.Equal(test.applicationAfterUpdate.Limit, application.Limit)
				}
			})
		}
	})

	ts.Run("Test_UpdateApplication_HugePayload", func() {
		tests := []struct {
			name                                       string
			applicationID, applicationUpdateJSONPath   string
			contractsLength, numOfChains, truncateList int
			totalPayloadSizeKB                         float64
			err                                        error
		}{
			{
				name:                      "Should update a single application's whitelists even with an extremely large payload",
				applicationID:             "test_app_47hfnths73j2se",
				applicationUpdateJSONPath: "./testdata/long-whitelist-testdata.json",
				contractsLength:           800,
				numOfChains:               7,
				totalPayloadSizeKB:        229.6875, // The maximum size of a single Postgres notification is 8kb
				err:                       nil,
			},
			{
				name:                      "Should update a single application's whitelists to a shorter list",
				applicationID:             "test_app_47hfnths73j2se",
				applicationUpdateJSONPath: "./testdata/long-whitelist-testdata.json",
				contractsLength:           257,
				numOfChains:               7,
				truncateList:              257,
				totalPayloadSizeKB:        73.787109375, // The maximum size of a single Postgres notification is 8kb
				err:                       nil,
			},
			{
				name:                      "Should update a single application's whitelists to a longer list",
				applicationID:             "test_app_47hfnths73j2se",
				applicationUpdateJSONPath: "./testdata/long-whitelist-testdata.json",
				contractsLength:           673,
				numOfChains:               7,
				truncateList:              673,
				totalPayloadSizeKB:        193.224609375, // The maximum size of a single Postgres notification is 8kb
				err:                       nil,
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				file, err := os.Open(test.applicationUpdateJSONPath)
				ts.NoError(err)
				defer file.Close()

				decoder := json.NewDecoder(file)
				var hugePayloadUpdate types.UpdateApplication
				err = decoder.Decode(&hugePayloadUpdate)
				ts.NoError(err)

				if test.truncateList != 0 {
					for i := range hugePayloadUpdate.GatewaySettings.WhitelistContracts {
						hugePayloadUpdate.GatewaySettings.WhitelistContracts[i].Contracts =
							hugePayloadUpdate.GatewaySettings.WhitelistContracts[i].Contracts[:test.truncateList]
					}
				}

				createdApp, err := ts.client1.UpdateApplication(testCtx, test.applicationID, hugePayloadUpdate)
				ts.Equal(test.err, err)

				if err == nil {
					<-time.After(3 * time.Second)
					application, err := ts.client1.GetApplicationByID(testCtx, createdApp.ID)
					ts.NoError(err)
					ts.Len(application.GatewaySettings.WhitelistContracts, test.numOfChains)

					var totalByteSize int
					for _, contract := range application.GatewaySettings.WhitelistContracts {
						ts.Len(contract.Contracts, test.contractsLength)
						for _, contractStr := range contract.Contracts {
							totalByteSize += len([]byte(contractStr))
						}
					}
					ts.Equal(float64(totalByteSize)/1024, test.totalPayloadSizeKB)

					application, err = ts.client2.GetApplicationByID(testCtx, createdApp.ID)
					ts.NoError(err)
					ts.Len(application.GatewaySettings.WhitelistContracts, test.numOfChains)

					var totalByteSize2 int
					for _, contract := range application.GatewaySettings.WhitelistContracts {
						ts.Len(contract.Contracts, test.contractsLength)
						for _, contractStr := range contract.Contracts {
							totalByteSize2 += len([]byte(contractStr))
						}
					}
					ts.Equal(float64(totalByteSize2)/1024, test.totalPayloadSizeKB)
				}
			})
		}
	})

	ts.Run("Test_UpdateAppFirstDateSurpassed", func() {
		tests := []struct {
			name         string
			update       types.UpdateFirstDateSurpassed
			expectedDate time.Time
			err          error
		}{
			{
				name: "Should update the app first date suprassed for the provided slice of app IDs",
				update: types.UpdateFirstDateSurpassed{
					ApplicationIDs:     []string{"test_app_47hfnths73j2se", "test_app_5hdf7sh23jd828"},
					FirstDateSurpassed: time.Date(2022, time.December, 13, 5, 15, 0, 0, time.UTC),
				},
				expectedDate: time.Date(2022, time.December, 13, 5, 15, 0, 0, time.UTC),
				err:          nil,
			},
			{
				name: "Should fail if update contains no application IDs cannot be found",
				update: types.UpdateFirstDateSurpassed{
					ApplicationIDs:     []string{},
					FirstDateSurpassed: time.Date(2022, time.December, 13, 5, 15, 0, 0, time.UTC),
				},
				err: fmt.Errorf("Response not OK. 400 Bad Request: no application IDs on input"),
			},
			{
				name: "Should fail if application cannot be found",
				update: types.UpdateFirstDateSurpassed{
					ApplicationIDs:     []string{"9000"},
					FirstDateSurpassed: time.Date(2022, time.December, 13, 5, 15, 0, 0, time.UTC),
				},
				err: fmt.Errorf("Response not OK. 404 Not Found: 9000 not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.UpdateAppFirstDateSurpassed(testCtx, test.update)
				ts.Equal(test.err, err)
				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					for _, appID := range test.update.ApplicationIDs {
						application, err := ts.client1.GetApplicationByID(testCtx, appID)
						ts.NoError(err)
						ts.Equal(test.expectedDate, application.FirstDateSurpassed)

						application, err = ts.client2.GetApplicationByID(testCtx, appID)
						ts.NoError(err)
						ts.Equal(test.expectedDate, application.FirstDateSurpassed)
					}
				}
			})
		}
	})

	ts.Run("Test_RemoveApplication", func() {
		tests := []struct {
			name           string
			applicationID  string
			expectedStatus types.AppStatus
			err            error
		}{
			{
				name:           "should remove one application by setting its status to AWAITING_GRACE_PERIOD",
				applicationID:  "test_app_5hdf7sh23jd828",
				expectedStatus: types.AwaitingGracePeriod,
			},
			{
				name:          "Should fail if application cannot be found",
				applicationID: "2348",
				err:           fmt.Errorf("Response not OK. 404 Not Found: applications not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.RemoveApplication(testCtx, test.applicationID)
				ts.Equal(test.err, err)
				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					application, err := ts.client1.GetApplicationByID(testCtx, test.applicationID)
					ts.NoError(err)
					ts.Equal(test.expectedStatus, application.Status)

					application, err = ts.client2.GetApplicationByID(testCtx, test.applicationID)
					ts.NoError(err)
					ts.Equal(test.expectedStatus, application.Status)
				}
			})
		}
	})

	ts.Run("Test_UpdateBlockchain", func() {
		allowanceInt := 123

		tests := []struct {
			name             string
			blockchainID     string
			blockchainUpdate types.UpdateBlockchain
			expectedChain    types.Blockchain
			err              error
		}{
			{
				name:         "Should update a single blockchain in the DB",
				blockchainID: "0001",
				blockchainUpdate: types.UpdateBlockchain{
					Altruist:          "https://test-update:test-password123@shared-test2.nodes.pokt.network:12345", // pragma: allowlist secret
					Blockchain:        "pokt-mainnet-updated",
					Description:       "POKT Network Mainnet Updated",
					RequestTimeout:    66_654,
					ResultKey:         "updated-key",
					LogLimitBlocks:    100_010,
					Ticker:            "SUCH-WOW",
					BlockchainAliases: []string{"pokt-mainnet", "another-one"},
					EnforceResult:     "JSON-2",
					Path:              "new-path",
					Body:              `{"new-body": "alliance"}`,
				},
				expectedChain: types.Blockchain{
					ID:                "0001",
					Altruist:          "https://test-update:test-password123@shared-test2.nodes.pokt.network:12345", // pragma: allowlist secret
					Blockchain:        "pokt-mainnet-updated",
					Description:       "POKT Network Mainnet Updated",
					EnforceResult:     "JSON-2",
					Ticker:            "SUCH-WOW",
					Path:              "new-path",
					BlockchainAliases: []string{"pokt-mainnet", "another-one"},
					LogLimitBlocks:    100_010,
					Active:            true,
					RequestTimeout:    66_654,
					// Redirects are handled via their own update process
					SyncCheckOptions: types.SyncCheckOptions{
						BlockchainID: "0001",
						Body:         `{"new-body": "alliance"}`,
						ResultKey:    "updated-key",
						Allowance:    1,
					},
				},
			},
			{
				name:         "Should update Allowance for a single blockchain in the DB",
				blockchainID: "0001",
				blockchainUpdate: types.UpdateBlockchain{
					Allowance: &allowanceInt,
				},
				expectedChain: types.Blockchain{
					ID:                "0001",
					Altruist:          "https://test-update:test-password123@shared-test2.nodes.pokt.network:12345", // pragma: allowlist secret
					Blockchain:        "pokt-mainnet-updated",
					Description:       "POKT Network Mainnet Updated",
					EnforceResult:     "JSON-2",
					Ticker:            "SUCH-WOW",
					Path:              "new-path",
					BlockchainAliases: []string{"pokt-mainnet", "another-one"},
					LogLimitBlocks:    100_010,
					Active:            true,
					RequestTimeout:    66_654,
					// Redirects are handled via their own update process
					SyncCheckOptions: types.SyncCheckOptions{
						BlockchainID: "0001",
						Body:         `{"new-body": "alliance"}`,
						ResultKey:    "updated-key",
						Allowance:    123,
					},
				},
			},
			{
				name:         "Should fail if blockchain cannot be found",
				blockchainID: "9000",
				err:          fmt.Errorf("Response not OK. 404 Not Found: blockchain not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				updatedChain, err := ts.client1.UpdateBlockchain(testCtx, test.blockchainID, test.blockchainUpdate)
				ts.Equal(test.err, err)
				if err == nil {
					<-time.After(50 * time.Millisecond)
					blockchain, err := ts.client1.GetBlockchainByID(testCtx, updatedChain.ID)
					ts.NoError(err)
					ts.Equal(test.expectedChain.Blockchain, blockchain.Blockchain)
					ts.Equal(test.expectedChain.Description, blockchain.Description)
					ts.Equal(test.expectedChain.RequestTimeout, blockchain.RequestTimeout)
					ts.Equal(test.expectedChain.LogLimitBlocks, blockchain.LogLimitBlocks)
					ts.Equal(test.expectedChain.Ticker, blockchain.Ticker)
					ts.Equal(test.expectedChain.BlockchainAliases, blockchain.BlockchainAliases)
					ts.Equal(test.expectedChain.EnforceResult, blockchain.EnforceResult)
					ts.Equal(test.expectedChain.Path, blockchain.Path)
					ts.Equal(test.expectedChain.SyncCheckOptions, blockchain.SyncCheckOptions)

					blockchain, err = ts.client2.GetBlockchainByID(testCtx, updatedChain.ID)
					ts.NoError(err)
					ts.Equal(test.expectedChain.Blockchain, blockchain.Blockchain)
					ts.Equal(test.expectedChain.Description, blockchain.Description)
					ts.Equal(test.expectedChain.RequestTimeout, blockchain.RequestTimeout)
					ts.Equal(test.expectedChain.LogLimitBlocks, blockchain.LogLimitBlocks)
					ts.Equal(test.expectedChain.Ticker, blockchain.Ticker)
					ts.Equal(test.expectedChain.BlockchainAliases, blockchain.BlockchainAliases)
					ts.Equal(test.expectedChain.EnforceResult, blockchain.EnforceResult)
					ts.Equal(test.expectedChain.Path, blockchain.Path)
					ts.Equal(test.expectedChain.SyncCheckOptions, blockchain.SyncCheckOptions)
				}
			})
		}
	})

	ts.Run("Test_UpdateLoadBalancer", func() {
		tests := []struct {
			name                    string
			loadBalancerID          string
			loadBalancerUpdate      types.UpdateLoadBalancer
			loadBalancerAfterUpdate types.LoadBalancer
			err                     error
		}{
			{
				name:           "Should update a single loadBalancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				loadBalancerUpdate: types.UpdateLoadBalancer{
					Name: "pokt_app_updated",
					StickyOptions: &types.UpdateStickyOptions{
						Duration:      "100",
						StickyOrigins: []string{"chrome-extension://", "test-ext://"},
						StickyMax:     500,
						Stickiness:    boolPointer(false),
					},
				},
				loadBalancerAfterUpdate: types.LoadBalancer{
					Name: "pokt_app_updated",
					StickyOptions: types.StickyOptions{
						Duration:      "100",
						StickyOrigins: []string{"chrome-extension://", "test-ext://"},
						StickyMax:     500,
						Stickiness:    false,
					},
				},
			},
			{
				name:           "Should update only the name of a single loadBalancer in the DB",
				loadBalancerID: "test_lb_3890ru23jfi32fj",
				loadBalancerUpdate: types.UpdateLoadBalancer{
					Name: "pokt_app_updated_2",
				},
				loadBalancerAfterUpdate: types.LoadBalancer{
					Name: "pokt_app_updated_2",
					StickyOptions: types.StickyOptions{
						Duration:      "40",
						StickyOrigins: []string{"chrome-extension://"},
						StickyMax:     400,
						Stickiness:    true,
					},
				},
			},
			{
				name:           "Should fail if load balancer cannot be found",
				loadBalancerID: "9000",
				err:            fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				updatedLB, err := ts.client1.UpdateLoadBalancer(testCtx, test.loadBalancerID, test.loadBalancerUpdate)
				ts.Equal(test.err, err)
				if err == nil {
					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, updatedLB.ID)
					ts.Equal(test.err, err)
					ts.Equal(test.loadBalancerAfterUpdate.Name, loadBalancer.Name)
					ts.Equal(test.loadBalancerAfterUpdate.StickyOptions, loadBalancer.StickyOptions)

					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, updatedLB.ID)
					ts.Equal(test.err, err)
					ts.Equal(test.loadBalancerAfterUpdate.Name, loadBalancer.Name)
					ts.Equal(test.loadBalancerAfterUpdate.StickyOptions, loadBalancer.StickyOptions)
				}
			})
		}
	})

	ts.Run("Test_UpdateLoadBalancerUserRole", func() {
		tests := []struct {
			name                   string
			loadBalancerID, userID string
			update                 types.UpdateUserAccess
			loadBalancerUsers      []types.UserAccess
			userLBPermissions      types.LoadBalancerPermissions
			err                    error
		}{
			{
				name:           "Should update a single user TO ADMIN for an existing load balancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				userID:         "test_user_member1234",
				update: types.UpdateUserAccess{
					Email:    "member1@test.com",
					RoleName: types.RoleAdmin,
				},
				loadBalancerUsers: []types.UserAccess{
					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "", Email: "member_new@test.com", Accepted: false},
				},
				userLBPermissions: types.LoadBalancerPermissions{
					RoleName:    types.RoleAdmin,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint},
				},
			},
			{
				name:           "Should update a single user to MEMBER for an existing load balancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				userID:         "test_user_member1234",
				update: types.UpdateUserAccess{
					Email:    "member1@test.com",
					RoleName: types.RoleMember,
				},
				loadBalancerUsers: []types.UserAccess{
					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "", Email: "member_new@test.com", Accepted: false},
				},
				userLBPermissions: types.LoadBalancerPermissions{
					RoleName:    types.RoleMember,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint},
				},
			},
			{
				name:           "Should transfer OWNERship for an existing load balancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				userID:         "test_user_admin1234",
				update: types.UpdateUserAccess{
					Email:        "admin1@test.com",
					RoleName:     types.RoleOwner,
					UpdaterEmail: "owner1@test.com",
				},
				loadBalancerUsers: []types.UserAccess{
					{RoleName: types.RoleAdmin, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: types.RoleOwner, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "", Email: "member_new@test.com", Accepted: false},
				},
				userLBPermissions: types.LoadBalancerPermissions{
					RoleName:    types.RoleOwner,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint, types.DeleteEndpoint, types.TransferEndpoint},
				},
			},
			{
				name:           "Should transfer OWNERship again for an existing load balancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				userID:         "test_user_member1234",
				update: types.UpdateUserAccess{
					Email:        "member1@test.com",
					RoleName:     types.RoleOwner,
					UpdaterEmail: "admin1@test.com",
				},
				loadBalancerUsers: []types.UserAccess{
					{RoleName: types.RoleAdmin, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: types.RoleOwner, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "", Email: "member_new@test.com", Accepted: false},
				},
				userLBPermissions: types.LoadBalancerPermissions{
					RoleName:    types.RoleOwner,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint, types.DeleteEndpoint, types.TransferEndpoint},
				},
			},
			{
				name:           "Should transfer OWNERship back to original owner for an existing load balancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				userID:         "test_user_1dbffbdfeeb225",
				update: types.UpdateUserAccess{
					Email:        "owner1@test.com",
					RoleName:     types.RoleOwner,
					UpdaterEmail: "member1@test.com",
				},
				loadBalancerUsers: []types.UserAccess{
					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "", Email: "member_new@test.com", Accepted: false},
				},
				userLBPermissions: types.LoadBalancerPermissions{
					RoleName:    types.RoleOwner,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint, types.DeleteEndpoint, types.TransferEndpoint},
				},
			},
			{
				name:           "Should update a single unaccepted user to ADMIN for an existing load balancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				userID:         "",
				update: types.UpdateUserAccess{
					Email:    "member_new@test.com",
					RoleName: types.RoleAdmin,
				},
				loadBalancerUsers: []types.UserAccess{
					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "", Email: "member_new@test.com", Accepted: false},
				},
				userLBPermissions: types.LoadBalancerPermissions{
					RoleName:    types.RoleAdmin,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint},
				},
			},
			{
				name:           "Should update a single unaccepted user back to MEMBER for an existing load balancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				userID:         "",
				update: types.UpdateUserAccess{
					Email:    "member_new@test.com",
					RoleName: types.RoleMember,
				},
				loadBalancerUsers: []types.UserAccess{
					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "", Email: "member_new@test.com", Accepted: false},
				},
				userLBPermissions: types.LoadBalancerPermissions{
					RoleName:    types.RoleMember,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint},
				},
			},
			{
				name:           "Should update a single user back to member for an existing load balancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				userID:         "test_user_member1234",
				update: types.UpdateUserAccess{
					Email:    "member1@test.com",
					RoleName: types.RoleMember,
				},
				loadBalancerUsers: []types.UserAccess{
					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "", Email: "member_new@test.com", Accepted: false},
				},
				userLBPermissions: types.LoadBalancerPermissions{
					RoleName:    types.RoleMember,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint},
				},
			},
			{
				name:           "Should fail if attempting to transfer ownership and the UpdaterEmail is not provided",
				loadBalancerID: "test_lb_34987u329rfn23f",
				update: types.UpdateUserAccess{
					Email:        "member1@test.com",
					RoleName:     types.RoleOwner,
					UpdaterEmail: "",
				},
				err: fmt.Errorf("transferring ownership requires providing the updater's email"),
			},
			{
				name:           "Should fail if attempting to transfer ownership and the user has not accepted their invite",
				loadBalancerID: "test_lb_34987u329rfn23f",
				update: types.UpdateUserAccess{
					Email:        "member_new@test.com",
					RoleName:     types.RoleOwner,
					UpdaterEmail: "owner1@test.com",
				},
				err: fmt.Errorf("Response not OK. 500 Internal Server Error: error: cannot set a user to owner if they have not yet accepted their invitation"),
			},
			{
				name:           "Should fail if load balancer ID not provided",
				loadBalancerID: "",
				err:            fmt.Errorf("no load balancer ID"),
			},
			{
				name:           "Should fail if user email not provided",
				loadBalancerID: "test_lb_34987u329rfn23f",
				update: types.UpdateUserAccess{
					Email:    "",
					RoleName: types.RoleAdmin,
				},
				err: fmt.Errorf("no user email"),
			},
			{
				name:           "Should fail if invalid role name provided",
				loadBalancerID: "test_lb_34987u329rfn23f",
				update: types.UpdateUserAccess{
					Email:    "member1@test.com",
					RoleName: types.RoleName("wrong_one"),
				},
				err: fmt.Errorf("invalid role name"),
			},
			{
				name:           "Should fail if load balancer cannot be found",
				loadBalancerID: "im_not_here",
				update: types.UpdateUserAccess{
					Email:    "member1@test.com",
					RoleName: types.RoleMember,
				},
				err: fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.UpdateLoadBalancerUserRole(testCtx, test.loadBalancerID, test.update)
				ts.Equal(test.err, err)
				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.Equal(test.err, err)
					ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.Equal(test.err, err)
					ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

					if test.userID != "" {
						permissionsByUserID, err := ts.client1.GetUserPermissionsByUserID(testCtx, types.UserID(test.userID))
						ts.Equal(test.err, err)
						ts.Equal(test.userLBPermissions, permissionsByUserID.LoadBalancers[types.LoadBalancerID(test.loadBalancerID)])

						permissionsByUserID, err = ts.client2.GetUserPermissionsByUserID(testCtx, types.UserID(test.userID))
						ts.Equal(test.err, err)
						ts.Equal(test.userLBPermissions, permissionsByUserID.LoadBalancers[types.LoadBalancerID(test.loadBalancerID)])
					}
				}
			})
		}
	})

	ts.Run("Test_AcceptLoadBalancerUser", func() {
		tests := []struct {
			name                          string
			email, loadBalancerID, userID string
			loadBalancerUsers             []types.UserAccess
			userLBPermissions             types.LoadBalancerPermissions
			err                           error
			pendingResult                 error
		}{
			{
				name:           "Should update a single user's ID and Accepted field for an existing load balancer in the DB",
				email:          "member_new@test.com",
				loadBalancerID: "test_lb_34987u329rfn23f",
				userID:         "test_user_accept_member",
				loadBalancerUsers: []types.UserAccess{
					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "test_user_accept_member", Email: "member_new@test.com", Accepted: true},
				},
				userLBPermissions: types.LoadBalancerPermissions{
					RoleName:    types.RoleMember,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint},
				},
				pendingResult: fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
			{
				name:  "Should fail if email not provided",
				email: "",
				err:   fmt.Errorf("no user email"),
			},
			{
				name:           "Should fail if load balancer ID not provided",
				email:          "member_new@test.com",
				loadBalancerID: "",
				err:            fmt.Errorf("no load balancer ID"),
			},
			{
				name:           "Should fail if user ID not provided",
				email:          "member_new@test.com",
				loadBalancerID: "test_lb_34987u329rfn23f",
				userID:         "",
				err:            fmt.Errorf("no user ID"),
			},
			{
				name:           "Should fail if load balancer cannot be found",
				email:          "member_new@test.com",
				loadBalancerID: "im_not_here",
				userID:         "test_user_accept_member",
				err:            fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.AcceptLoadBalancerUser(testCtx, test.email, test.loadBalancerID, test.userID)
				ts.Equal(test.err, err)
				_, err = ts.client2.AcceptLoadBalancerUser(testCtx, test.email, test.loadBalancerID, test.userID)
				ts.Equal(test.err, err)
				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.Equal(test.err, err)
					ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.Equal(test.err, err)
					ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

					if test.userID != "" {
						// user should have been added to the user permissions in the cache
						permissionsByUserID, err := ts.client1.GetUserPermissionsByUserID(testCtx, types.UserID(test.userID))
						ts.Equal(test.err, err)
						ts.Equal(test.userLBPermissions, permissionsByUserID.LoadBalancers[types.LoadBalancerID(test.loadBalancerID)])

						permissionsByUserID, err = ts.client2.GetUserPermissionsByUserID(testCtx, types.UserID(test.userID))
						ts.Equal(test.err, err)
						ts.Equal(test.userLBPermissions, permissionsByUserID.LoadBalancers[types.LoadBalancerID(test.loadBalancerID)])

						// user's LB should have been added to the user's LBs map in the cache
						roleMember := types.RoleMember

						exists := false
						loadBalancers, err := ts.client1.GetLoadBalancersByUserID(testCtx, test.userID, &roleMember)
						ts.Equal(test.err, err)
						for _, lb := range loadBalancers {
							if lb.ID == test.loadBalancerID {
								exists = true
								break
							}
						}
						ts.True(exists)

						exists = false
						loadBalancers, err = ts.client2.GetLoadBalancersByUserID(testCtx, test.userID, &roleMember)
						ts.Equal(test.err, err)
						for _, lb := range loadBalancers {
							if lb.ID == test.loadBalancerID {
								exists = true
								break
							}
						}
						ts.True(exists)

						// user should be removed from pending lb list
						removed := true
						pendingLoadBalancers, _ := ts.client1.GetPendingLoadBalancersByEmail(testCtx, test.email)
						for _, lb := range pendingLoadBalancers {
							if lb.ID == test.loadBalancerID {
								removed = false
								break
							}
						}
						ts.True(removed)

						removed = true
						pendingLoadBalancers, _ = ts.client2.GetPendingLoadBalancersByEmail(testCtx, test.email)
						for _, lb := range pendingLoadBalancers {
							if lb.ID == test.loadBalancerID {
								removed = false
								break
							}
						}
						ts.True(removed)
					}
				}
			})
		}
	})

	ts.Run("Test_RemoveLoadBalancer", func() {
		tests := []struct {
			name           string
			loadBalancerID string
			expectedUserID string
			err            error
		}{
			{
				name:           "should remove one load balancer by setting its user ID to an empty string",
				loadBalancerID: "test_lb_3890ru23jfi32fj",
				expectedUserID: "",
			},
			{
				name:           "Should fail if load balancer cannot be found",
				loadBalancerID: "9000",
				err:            fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.RemoveLoadBalancer(testCtx, test.loadBalancerID)
				ts.Equal(test.err, err)
				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.NoError(err)
					ts.Equal(test.expectedUserID, loadBalancer.UserID)

					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.NoError(err)
					ts.Equal(test.expectedUserID, loadBalancer.UserID)
				}
			})
		}
	})

	ts.Run("Test_DeleteLoadBalancerUser", func() {
		tests := []struct {
			name                       string
			loadBalancerID, email      string
			userID                     string
			loadBalancerUsers          []types.UserAccess
			testPendingUsers           []types.UserAccess
			err, permissionsFetchError error
		}{
			{
				name:           "Should delete a single user for an existing load balancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				email:          "member1@test.com",
				userID:         "test_user_member1234",
				loadBalancerUsers: []types.UserAccess{
					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "test_user_accept_member", Email: "member_new@test.com", Accepted: true},
				},
				permissionsFetchError: errors.New("Response not OK. 404 Not Found: no permissions found for given user ID"),
			},
			{
				name:           "Should delete a single pending user for an existing load balancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				testPendingUsers: []types.UserAccess{
					{Email: "owner2@test.com", RoleName: types.RoleMember},
				},
				email:  "owner2@test.com",
				userID: "test_user_04228205bd261a",
				loadBalancerUsers: []types.UserAccess{
					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "test_user_accept_member", Email: "member_new@test.com", Accepted: true},
				},
				permissionsFetchError: nil, // this test user has other permissions for other LBs so will still return permissions
			},
			{
				name:           "Should delete only one pending user even if multiple pending users exist for an existing load balancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				testPendingUsers: []types.UserAccess{
					{Email: "owner2@test.com", RoleName: types.RoleMember},
					{Email: "test-delete-user-2@test.com", RoleName: types.RoleAdmin},
					{Email: "test-delete-user-3@test.com", RoleName: types.RoleMember},
				},
				email:  "owner2@test.com",
				userID: "test_user_04228205bd261a",
				loadBalancerUsers: []types.UserAccess{
					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: types.RoleMember, UserID: "test_user_accept_member", Email: "member_new@test.com", Accepted: true},
					{RoleName: types.RoleAdmin, UserID: "", Email: "test-delete-user-2@test.com", Accepted: false},
					{RoleName: types.RoleMember, UserID: "", Email: "test-delete-user-3@test.com", Accepted: false},
				},
				permissionsFetchError: nil, // this test user has other permissions for other LBs so will still return permissions
			},
			{
				name:           "Should fail if load balancer cannot be found",
				loadBalancerID: "why_am_i_not_a_load_balancer",
				email:          "member1@test.com",
				userID:         "test_user_member1234",
				err:            fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
			{
				name:           "Should fail if attempting to delete the owner of a load balancer",
				loadBalancerID: "test_lb_34987u329rfn23f",
				email:          "owner1@test.com",
				userID:         "test_user_1dbffbdfeeb225",
				err:            fmt.Errorf("Response not OK. 500 Internal Server Error: cannot delete the owner of a load balancer"),
			},
			{
				name:           "Should fail if load balancer ID not provided",
				loadBalancerID: "",
				email:          "member1@test.com",
				err:            fmt.Errorf("no load balancer ID"),
			},
			{
				name:           "Should fail if email not provided",
				loadBalancerID: "test_lb_34987u329rfn23f",
				email:          "",
				err:            fmt.Errorf("no user email"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				if len(test.testPendingUsers) != 0 {
					var updatedLB *types.LoadBalancer
					for _, user := range test.testPendingUsers {
						lb, err := ts.client1.CreateLoadBalancerUser(testCtx, test.loadBalancerID, user)
						ts.NoError(err)
						updatedLB = lb
					}
					ts.Len(updatedLB.Users, 3+len(test.testPendingUsers))
				}

				_, err := ts.client1.DeleteLoadBalancerUser(testCtx, test.loadBalancerID, test.email)
				ts.Equal(test.err, err)
				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.Equal(test.err, err)
					ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.Equal(test.err, err)
					ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

					// user should have been removed from the user permissions in the cache
					_, err = ts.client1.GetUserPermissionsByUserID(testCtx, types.UserID(test.userID))
					ts.Equal(test.permissionsFetchError, err)

					_, err = ts.client2.GetUserPermissionsByUserID(testCtx, types.UserID(test.userID))
					ts.Equal(test.permissionsFetchError, err)

					// user's LB should have been removed from the user's LBs map in the cache
					_, err = ts.client1.GetLoadBalancersByUserID(testCtx, test.userID, nil)
					ts.Equal(fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"), err)

					_, err = ts.client2.GetLoadBalancersByUserID(testCtx, test.userID, nil)
					ts.Equal(fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"), err)
				}
			})
		}
	})
}
