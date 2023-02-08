package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	dbclient "github.com/pokt-foundation/db-client/client"
	"github.com/pokt-foundation/portal-db/types"
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
						Network:           "POKT-mainnet",
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
							Path:      "/v1/query/height",
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
						Network:           "ETH-1",
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
			blockchains, err := ts.client1.GetBlockchains(testCtx)
			ts.ErrorIs(test.err, err)
			ts.Equal(test.expectedBlockchains, blockchains)

			blockchains, err = ts.client2.GetBlockchains(testCtx)
			ts.ErrorIs(test.err, err)
			ts.Equal(test.expectedBlockchains, blockchains)
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
					Network:           "ETH-1",
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
			blockchain, err := ts.client1.GetBlockchainByID(testCtx, test.blockchainID)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedBlockchain, blockchain)

			blockchain, err = ts.client2.GetBlockchainByID(testCtx, test.blockchainID)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedBlockchain, blockchain)
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
							SecretKey:         "test_40f482d91a5ef2300ebb4e2308c",
							SecretKeyRequired: true,
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
				},
			},
		}

		for _, test := range tests {
			applications, err := ts.client1.GetApplications(testCtx)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedApplications, applications)

			applications, err = ts.client2.GetApplications(testCtx)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedApplications, applications)
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
			applicationByID, err := ts.client1.GetApplicationByID(testCtx, test.applicationID)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedApplication, applicationByID)

			applicationByID, err = ts.client2.GetApplicationByID(testCtx, test.applicationID)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedApplication, applicationByID)
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
			applicationsByUserID, err := ts.client1.GetApplicationsByUserID(testCtx, test.userID)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedApplications, applicationsByUserID)

			applicationsByUserID, err = ts.client2.GetApplicationsByUserID(testCtx, test.userID)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedApplications, applicationsByUserID)
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
									SecretKey:         "test_40f482d91a5ef2300ebb4e2308c",
									SecretKeyRequired: true,
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
							{RoleName: "OWNER", UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
							{RoleName: "ADMIN", UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
							{RoleName: "MEMBER", UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
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
						Applications: []*types.Application{nil},
						Users: []types.UserAccess{
							{RoleName: "OWNER", UserID: "test_user_redirect233344", Email: "owner3@test.com", Accepted: true},
							{RoleName: "MEMBER", UserID: "test_user_member5678", Email: "member2@test.com", Accepted: false},
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
							{RoleName: "OWNER", UserID: "test_user_04228205bd261a", Email: "owner2@test.com", Accepted: true},
							{RoleName: "ADMIN", UserID: "test_user_admin5678", Email: "admin2@test.com", Accepted: true},
						},
						CreatedAt: mockTimestamp,
						UpdatedAt: mockTimestamp,
					},
				},
			},
		}

		for _, test := range tests {
			loadBalancers, err := ts.client1.GetLoadBalancers(testCtx)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedLoadBalancers, loadBalancers)

			loadBalancers, err = ts.client2.GetLoadBalancers(testCtx)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedLoadBalancers, loadBalancers)
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
						{RoleName: "OWNER", UserID: "test_user_04228205bd261a", Email: "owner2@test.com", Accepted: true},
						{RoleName: "ADMIN", UserID: "test_user_admin5678", Email: "admin2@test.com", Accepted: true},
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
			loadBalancerByID, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedLoadBalancer, loadBalancerByID)

			loadBalancerByID, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedLoadBalancer, loadBalancerByID)
		}
	})

	ts.Run("Test_GetLoadBalancersByUserID", func() {
		tests := []struct {
			name                  string
			userID                string
			expectedLoadBalancers []*types.LoadBalancer
			err                   error
		}{
			{
				name:   "Should fetch all load balancers for a single user ID",
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
									SecretKeyRequired: true,
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
							{RoleName: "OWNER", UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
							{RoleName: "ADMIN", UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
							{RoleName: "MEMBER", UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
						},
						CreatedAt: mockTimestamp,
						UpdatedAt: mockTimestamp,
					},
				},
			},
			{
				name:   "Should fail if the user does not have any load balancers in the DB",
				userID: "test_not_real_user",
				err:    fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
		}

		for _, test := range tests {
			loadBalancersByUserID, err := ts.client1.GetLoadBalancersByUserID(testCtx, test.userID)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedLoadBalancers, loadBalancersByUserID)

			loadBalancersByUserID, err = ts.client2.GetLoadBalancersByUserID(testCtx, test.userID)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedLoadBalancers, loadBalancersByUserID)
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
			payPlans, err := ts.client1.GetPayPlans(testCtx)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedPayPlans, payPlans)

			payPlans, err = ts.client2.GetPayPlans(testCtx)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedPayPlans, payPlans)
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
			payPlanByType, err := ts.client1.GetPayPlanByType(testCtx, test.payPlanType)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedPayPlan, payPlanByType)

			payPlanByType, err = ts.client2.GetPayPlanByType(testCtx, test.payPlanType)
			ts.Equal(test.err, err)
			ts.Equal(test.expectedPayPlan, payPlanByType)
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
					Network:           "POL-mainnet",
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
					Network:           "POL-mainnet",
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
					Network:           "POL-mainnet",
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
			_, err := ts.client1.CreateBlockchain(testCtx, test.blockchainInput)
			ts.Equal(test.err, err)
			if test.err == nil {
				blockchain, err := ts.client1.GetBlockchainByID(testCtx, test.blockchainInput.ID)
				ts.Equal(test.err, err)
				ts.Equal(test.blockchain.ID, blockchain.ID)
				ts.Equal(test.blockchain.Altruist, blockchain.Altruist)
				ts.Equal(test.blockchain.Blockchain, blockchain.Blockchain)
				ts.Equal(test.blockchain.ChainID, blockchain.ChainID)
				ts.Equal(test.blockchain.ChainIDCheck, blockchain.ChainIDCheck)
				ts.Equal(test.blockchain.Description, blockchain.Description)
				ts.Equal(test.blockchain.EnforceResult, blockchain.EnforceResult)
				ts.Equal(test.blockchain.Network, blockchain.Network)
				ts.Equal(test.blockchain.Path, blockchain.Path)
				ts.Equal(test.blockchain.SyncCheck, blockchain.SyncCheck)
				ts.Equal(test.blockchain.Ticker, blockchain.Ticker)
				ts.Equal(test.blockchain.BlockchainAliases, blockchain.BlockchainAliases)
				ts.Equal(test.blockchain.LogLimitBlocks, blockchain.LogLimitBlocks)
				ts.Equal(test.blockchain.RequestTimeout, blockchain.RequestTimeout)
				ts.Equal(test.blockchain.SyncAllowance, blockchain.SyncAllowance)
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
				ts.Equal(test.blockchain.Network, blockchain.Network)
				ts.Equal(test.blockchain.Path, blockchain.Path)
				ts.Equal(test.blockchain.SyncCheck, blockchain.SyncCheck)
				ts.Equal(test.blockchain.Ticker, blockchain.Ticker)
				ts.Equal(test.blockchain.BlockchainAliases, blockchain.BlockchainAliases)
				ts.Equal(test.blockchain.LogLimitBlocks, blockchain.LogLimitBlocks)
				ts.Equal(test.blockchain.RequestTimeout, blockchain.RequestTimeout)
				ts.Equal(test.blockchain.SyncAllowance, blockchain.SyncAllowance)
				ts.Equal(test.blockchain.Active, blockchain.Active)
				ts.Equal(test.blockchain.SyncCheckOptions, blockchain.SyncCheckOptions)
				ts.NotEmpty(blockchain.CreatedAt)
				ts.NotEmpty(blockchain.UpdatedAt)
			}
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
			_, err := ts.client1.CreateBlockchainRedirect(testCtx, test.redirectInput)
			ts.Equal(test.err, err)
			if test.err == nil {
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
					Name:   "pokt_app_789",
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
					Name:   "pokt_app_789",
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
			createdApp, err := ts.client1.CreateApplication(testCtx, test.applicationInput)
			ts.Equal(test.err, err)
			if test.err == nil {
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
								SecretKey:         "test_40f482d91a5ef2300ebb4e2308c",
								SecretKeyRequired: true,
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
			createdLB, err := ts.client1.CreateLoadBalancer(testCtx, test.loadBalancerInput)
			ts.Equal(test.err, err)
			if test.err == nil {
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
					RoleName: "MEMBER",
					UserID:   "test_user_create_new_member",
					Email:    "member_new@test.com",
				},
				loadBalancerUsers: []types.UserAccess{
					{RoleName: "OWNER", UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: "ADMIN", UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: "MEMBER", UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
					{RoleName: "MEMBER", UserID: "test_user_create_new_member", Email: "member_new@test.com", Accepted: false},
				},
			},
			{
				name:           "Should fail if load balancer cannot be found",
				loadBalancerID: "sir_not_appearing_in_this_film",
				err:            fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
		}

		for _, test := range tests {
			_, err := ts.client1.CreateLoadBalancerUser(testCtx, test.loadBalancerID, test.user)
			ts.Equal(test.err, err)
			if test.err == nil {
				loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
				ts.Equal(test.err, err)
				ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

				loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
				ts.Equal(test.err, err)
				ts.Equal(test.loadBalancerUsers, loadBalancer.Users)
			}
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
			_, err := ts.client1.ActivateBlockchain(testCtx, test.blockchainID, test.active)
			ts.Equal(test.err, err)
			if test.err == nil {
				blockchain, err := ts.client1.GetBlockchainByID(testCtx, test.blockchainID)
				ts.Equal(test.err, err)
				ts.Equal(test.active, blockchain.Active)
			}
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
						WhitelistContracts:   []types.WhitelistContract{{BlockchainID: "01", Contracts: []string{"test-contract1"}}},
						WhitelistMethods:     []types.WhitelistMethod{{BlockchainID: "01", Methods: []string{"test-method1"}}},
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
						WhitelistContracts:   []types.WhitelistContract{{BlockchainID: "01", Contracts: []string{"test-contract1"}}},
						WhitelistMethods:     []types.WhitelistMethod{{BlockchainID: "01", Methods: []string{"test-method1"}}},
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
						WhitelistContracts:   []types.WhitelistContract{{BlockchainID: "01", Contracts: []string{"test-contract1"}}},
						WhitelistMethods:     []types.WhitelistMethod{{BlockchainID: "01", Methods: []string{"test-method1"}}},
						WhitelistBlockchains: []string{"test-chain1"},
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
			createdApp, err := ts.client1.UpdateApplication(testCtx, test.applicationID, test.applicationUpdate)
			ts.Equal(test.err, err)
			if err == nil {
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
			_, err := ts.client1.UpdateAppFirstDateSurpassed(testCtx, test.update)
			ts.Equal(test.err, err)
			if test.err == nil {
				for _, appID := range test.update.ApplicationIDs {
					application, err := ts.client1.GetApplicationByID(testCtx, appID)
					ts.NoError(err)
					ts.Equal(test.expectedDate, application.FirstDateSurpassed)

					application, err = ts.client2.GetApplicationByID(testCtx, appID)
					ts.NoError(err)
					ts.Equal(test.expectedDate, application.FirstDateSurpassed)
				}
			}
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
			_, err := ts.client1.RemoveApplication(testCtx, test.applicationID)
			ts.Equal(test.err, err)
			if test.err == nil {
				application, err := ts.client1.GetApplicationByID(testCtx, test.applicationID)
				ts.NoError(err)
				ts.Equal(test.expectedStatus, application.Status)

				application, err = ts.client2.GetApplicationByID(testCtx, test.applicationID)
				ts.NoError(err)
				ts.Equal(test.expectedStatus, application.Status)
			}
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
			createdLB, err := ts.client1.UpdateLoadBalancer(testCtx, test.loadBalancerID, test.loadBalancerUpdate)
			ts.Equal(test.err, err)
			if err == nil {
				loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, createdLB.ID)
				ts.Equal(test.err, err)
				ts.Equal(test.loadBalancerAfterUpdate.Name, loadBalancer.Name)
				ts.Equal(test.loadBalancerAfterUpdate.StickyOptions, loadBalancer.StickyOptions)

				loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, createdLB.ID)
				ts.Equal(test.err, err)
				ts.Equal(test.loadBalancerAfterUpdate.Name, loadBalancer.Name)
				ts.Equal(test.loadBalancerAfterUpdate.StickyOptions, loadBalancer.StickyOptions)
			}
		}
	})

	ts.Run("Test_UpdateLoadBalancerUserRole", func() {
		tests := []struct {
			name              string
			loadBalancerID    string
			userUpdate        types.UpdateUserAccess
			loadBalancerUsers []types.UserAccess
			err               error
		}{
			{
				name:           "Should add a single user to an existing load balancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				userUpdate: types.UpdateUserAccess{
					RoleName: types.RoleAdmin,
					UserID:   "test_user_create_new_member",
				},
				loadBalancerUsers: []types.UserAccess{
					{RoleName: "OWNER", UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: "ADMIN", UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: "MEMBER", UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
					{RoleName: "ADMIN", UserID: "test_user_create_new_member", Email: "member_new@test.com", Accepted: false},
				},
			},
			{
				name:           "Should fail if load balancer cannot be found",
				loadBalancerID: "im_not_here",
				err:            fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
		}

		for _, test := range tests {
			_, err := ts.client1.UpdateLoadBalancerUserRole(testCtx, test.loadBalancerID, test.userUpdate)
			ts.Equal(test.err, err)
			if test.err == nil {
				loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
				ts.Equal(test.err, err)
				ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

				loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
				ts.Equal(test.err, err)
				ts.Equal(test.loadBalancerUsers, loadBalancer.Users)
			}
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
			_, err := ts.client1.RemoveLoadBalancer(testCtx, test.loadBalancerID)
			ts.Equal(test.err, err)
			if test.err == nil {
				loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
				ts.NoError(err)
				ts.Equal(test.expectedUserID, loadBalancer.UserID)

				loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
				ts.NoError(err)
				ts.Equal(test.expectedUserID, loadBalancer.UserID)
			}
		}
	})

	ts.Run("Test_DeleteLoadBalancerUser", func() {
		tests := []struct {
			name                   string
			loadBalancerID, userID string
			loadBalancerUsers      []types.UserAccess
			err                    error
		}{
			{
				name:           "Should add a single user to an existing load balancer in the DB",
				loadBalancerID: "test_lb_34987u329rfn23f",
				userID:         "test_user_member1234",
				loadBalancerUsers: []types.UserAccess{
					{RoleName: "OWNER", UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
					{RoleName: "ADMIN", UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
					{RoleName: "ADMIN", UserID: "test_user_create_new_member", Email: "member_new@test.com", Accepted: false},
				},
			},
			{
				name:           "Should fail if load balancer cannot be found",
				loadBalancerID: "why_am_i_not_a_load_balancer",
				userID:         "test_user_member1234",
				err:            fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
		}

		for _, test := range tests {
			_, err := ts.client1.DeleteLoadBalancerUser(testCtx, test.loadBalancerID, test.userID)
			ts.Equal(test.err, err)
			if test.err == nil {
				loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
				ts.Equal(test.err, err)
				ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

				loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
				ts.Equal(test.err, err)
				ts.Equal(test.loadBalancerUsers, loadBalancer.Users)
			}
		}
	})
}
