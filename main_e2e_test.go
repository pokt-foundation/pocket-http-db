package main

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	dbclient "github.com/pokt-foundation/db-client/client"
	v1Types "github.com/pokt-foundation/portal-db/types"
	"github.com/pokt-foundation/portal-db/v2/testdata"
	v2Types "github.com/pokt-foundation/portal-db/v2/types"
	"github.com/stretchr/testify/suite"
)

func Test_V1_E2E_PocketHTTPDBTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end to end test")
	}

	readSuite := new(phdE2EReadTestSuite)
	writeSuite := new(phdE2EWriteTestSuite)

	err := initDBClient(readSuite)
	if err != nil {
		t.Fatal("Failed to initialize the DB client for read tests:", err)
	}

	err = initDBClient(writeSuite)
	if err != nil {
		t.Fatal("Failed to initialize the DB client for write tests:", err)
	}

	suite.Run(t, readSuite)
	suite.Run(t, writeSuite)
}

// Runs all the read-only endpoint tests first to compare to test DB seed data only
// ie. not yet including data written to the test DB by the test suite
func (ts *phdE2EReadTestSuite) Test_ReadTests() {

	/* ------ V2 Chain Read Tests ------ */

	ts.Run("Test_GetChainByID", func() {
		tests := []struct {
			name          string
			chainID       v2Types.RelayChainID
			err           error
			expectedChain *v2Types.Chain
			gigastakeApp  *v2Types.GigastakeApp
		}{
			{
				name:          "Should get chain by ID",
				chainID:       "0001",
				expectedChain: testdata.Chains["0001"],
				gigastakeApp:  testdata.GigastakeApps["test_gigastake_app_1"],
			},
			{
				name:    "Should return error if chain ID is empty",
				chainID: "",
				err:     fmt.Errorf("no chain ID"),
			},
			{
				name:    "Should return error if chain does not exist",
				chainID: "9999",
				err:     fmt.Errorf("Response not OK. 404 Not Found: blockchain not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				got, err := ts.client1.GetChainByID(testCtx, test.chainID)
				ts.Equal(test.err, err)

				if err == nil {
					// Assign GigastakeApp to the chain's GigastakeApps
					test.expectedChain.GigastakeApps = make(map[v2Types.ProtocolAppID]*v2Types.GigastakeApp)
					test.expectedChain.GigastakeApps[test.gigastakeApp.AATID] = test.gigastakeApp

					// Compare the expectedChain and actual
					ts.Equal(test.expectedChain, got)
				}
			})
		}
	})

	/* ------ V1 Read Tests ------ */

	ts.Run("Test_GetBlockchains", func() {
		tests := []struct {
			name                string
			expectedBlockchains map[string]*v1Types.Blockchain
			err                 error
		}{
			{
				name:                "Should fetch all blockchains in the database",
				expectedBlockchains: expectedLegacyBlockchains,
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				blockchains, err := ts.client1.GetBlockchains(testCtx)
				ts.ErrorIs(test.err, err)
				ts.Equal(test.expectedBlockchains, blockchainsToMap(blockchains))

				blockchains, err = ts.client2.GetBlockchains(testCtx)
				ts.NoError(err)
				ts.Equal(test.expectedBlockchains, blockchainsToMap(blockchains))
			})

		}
	})

	ts.Run("Test_GetBlockchain", func() {
		tests := []struct {
			name               string
			blockchainID       string
			expectedBlockchain *v1Types.Blockchain
			err                error
		}{
			{
				name:               "Should fetch one blockchain by ID",
				blockchainID:       "0021",
				expectedBlockchain: expectedLegacyBlockchains["0021"],
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
			expectedApplications map[string]*v1Types.Application
			err                  error
		}{
			{
				name:                 "Should fetch all applications in the database",
				expectedApplications: expectedLegacyApplications,
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				applications, err := ts.client1.GetApplications(testCtx)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedApplications, applicationsToMap(applications))

				applications, err = ts.client2.GetApplications(testCtx)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedApplications, applicationsToMap(applications))
			})
		}
	})

	ts.Run("Test_GetApplicationsByUserID", func() {
		tests := []struct {
			name                 string
			userID               string
			expectedApplications map[string]*v1Types.Application
			err                  error
		}{
			{
				name:   "Should fetch all applications for a single user ID",
				userID: "auth0|chrisjen_avasarala",
				expectedApplications: map[string]*v1Types.Application{
					"test_protocol_app_3": expectedLegacyApplications["test_protocol_app_3"],
					"test_protocol_app_4": expectedLegacyApplications["test_protocol_app_4"],
				},
			},
			{
				name:   "Should fail if the user does not have any applications in the DB",
				userID: "test_not_real_user",
				err:    fmt.Errorf("Response not OK. 404 Not Found: user not found for provider user ID test_not_real_user"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				applicationsByUserID, err := ts.client1.GetApplicationsByUserID(testCtx, test.userID)
				ts.Equal(test.err, err)
				if test.err == nil {
					ts.Equal(test.expectedApplications, applicationsToMap(applicationsByUserID))
				}

				applicationsByUserID, err = ts.client2.GetApplicationsByUserID(testCtx, test.userID)
				ts.Equal(test.err, err)
				if test.err == nil {
					ts.Equal(test.expectedApplications, applicationsToMap(applicationsByUserID))
				}
			})
		}
	})

	ts.Run("Test_GetLoadBalancers", func() {
		tests := []struct {
			name                  string
			expectedLoadBalancers map[string]*v1Types.LoadBalancer
			err                   error
		}{
			{
				name: "Should fetch all load balancers in the database",
				expectedLoadBalancers: map[string]*v1Types.LoadBalancer{
					"test_app_1":  expectedLegacyLoadBalancers["test_app_1"],
					"test_app_2":  expectedLegacyLoadBalancers["test_app_2"],
					"test_app_3":  expectedLegacyLoadBalancers["test_app_3"],
					"legacy_lb_1": expectedLegacyLoadBalancers["legacy_lb_1"],
					"legacy_lb_2": expectedLegacyLoadBalancers["legacy_lb_2"],
					"legacy_lb_3": expectedLegacyLoadBalancers["legacy_lb_3"],
				},
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				loadBalancers, err := ts.client1.GetLoadBalancers(testCtx)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedLoadBalancers, loadBalancersToMap(loadBalancers))

				loadBalancers, err = ts.client2.GetLoadBalancers(testCtx)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedLoadBalancers, loadBalancersToMap(loadBalancers))
			})
		}
	})

	ts.Run("Test_GetLoadBalancerByID", func() {
		tests := []struct {
			name                 string
			loadBalancerID       string
			expectedLoadBalancer *v1Types.LoadBalancer
			err                  error
		}{
			{
				name:                 "Should fetch one load balancer by ID",
				loadBalancerID:       "test_app_1",
				expectedLoadBalancer: expectedLegacyLoadBalancers["test_app_1"],
			},
			{
				name:           "Should fail if the load balancer does not exist in the DB",
				loadBalancerID: "test_not_real_load_balancer",
				err:            fmt.Errorf("Response not OK. 404 Not Found: portal app not found for load balancer ID test_not_real_load_balancer"),
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
			expectedLoadBalancers map[string]*v1Types.LoadBalancer
			roleNameFilter        v1Types.RoleName
			err                   error
		}{
			{
				name:   "Should fetch all load balancers for a single user ID when no filter provided",
				userID: "auth0|ulfric_stormcloak",
				expectedLoadBalancers: map[string]*v1Types.LoadBalancer{
					"test_app_2": expectedLegacyLoadBalancers["test_app_2"],
				},
			},
			{
				name:           "Should fetch all load balancers for a single user ID and role when a valid filter provided",
				userID:         "auth0|amos_burton",
				roleNameFilter: v1Types.RoleAdmin,
				expectedLoadBalancers: map[string]*v1Types.LoadBalancer{
					"test_app_3": expectedLegacyLoadBalancers["test_app_3"],
				},
			},
			{
				name:                  "Should return empty if the user does not have any load balancers",
				userID:                "auth0|bernard_marx",
				expectedLoadBalancers: map[string]*v1Types.LoadBalancer{},
			},
			{
				name:           "Should fail if an invalid role name provided as a filter",
				userID:         "test_user_1dbffbdfeeb225",
				roleNameFilter: v1Types.RoleName("not_real"),
				err:            fmt.Errorf("invalid role name filter"),
			},
			{
				name:   "Should fail if the user does not exist",
				userID: "test_not_real_user",
				err:    fmt.Errorf("Response not OK. 404 Not Found: user not found for provider user ID test_not_real_user"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				filter := &test.roleNameFilter
				if test.roleNameFilter == "" {
					filter = nil
				}

				loadBalancers, err := ts.client1.GetLoadBalancersByUserID(testCtx, test.userID, filter)
				ts.Equal(test.err, err)
				if test.err == nil {
					ts.Equal(test.expectedLoadBalancers, loadBalancersToMap(loadBalancers))
				}

				loadBalancers, err = ts.client2.GetLoadBalancersByUserID(testCtx, test.userID, filter)
				ts.Equal(test.err, err)
				if test.err == nil {
					ts.Equal(test.expectedLoadBalancers, loadBalancersToMap(loadBalancers))
				}
			})
		}
	})

	ts.Run("Test_GetPayPlans", func() {
		tests := []struct {
			name             string
			expectedPayPlans map[string]*v1Types.PayPlan
			err              error
		}{
			{
				name:             "Should fetch all pay plans in the DB",
				expectedPayPlans: expectedLegacyPayPlans,
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				payPlans, err := ts.client1.GetPayPlans(testCtx)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedPayPlans, payPlansToMap(payPlans))

				payPlans, err = ts.client2.GetPayPlans(testCtx)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedPayPlans, payPlansToMap(payPlans))
			})
		}
	})

	ts.Run("Test_GetPayPlanByType", func() {
		tests := []struct {
			name            string
			payPlanType     v1Types.PayPlanType
			expectedPayPlan *v1Types.PayPlan
			err             error
		}{
			{

				name:            "Should fetch a single pay plan by type",
				payPlanType:     "pro_plan",
				expectedPayPlan: expectedLegacyPayPlans["pro_plan"],
			},
			{
				name:        "Should fail if passed a pay plan type that is not in the DB",
				payPlanType: v1Types.PayPlanType("not_a_real_plan"),
				err:         fmt.Errorf("Response not OK. 404 Not Found: plan not found for type not_a_real_plan"),
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
			userID              v1Types.UserID
			expectedPermissions *v1Types.UserPermissions
			err                 error
		}{
			{

				name:   "Should fetch a single users load balancer permissions",
				userID: "auth0|paul_atreides",
				expectedPermissions: &v1Types.UserPermissions{
					UserID: "auth0|paul_atreides",
					LoadBalancers: map[v1Types.LoadBalancerID]v1Types.LoadBalancerPermissions{
						"test_app_1": {
							RoleName:    "ADMIN",
							Permissions: []v1Types.PermissionsEnum{"read:endpoint", "write:endpoint"},
						},
						"test_app_2": {
							RoleName:    "MEMBER",
							Permissions: []v1Types.PermissionsEnum{"read:endpoint"},
						},
					},
				},
			},
			{
				name:   "Should fetch another single users load balancer permissions",
				userID: "auth0|ulfric_stormcloak",
				expectedPermissions: &v1Types.UserPermissions{
					UserID: "auth0|ulfric_stormcloak",
					LoadBalancers: map[v1Types.LoadBalancerID]v1Types.LoadBalancerPermissions{
						"test_app_2": {
							RoleName: "MEMBER",
							Permissions: []v1Types.PermissionsEnum{
								"read:endpoint",
							},
						},
					},
				},
			},
			{
				name:   "Should fail if the user does not have any permissions",
				userID: "test_user_hey_who_am_i_wow",
				err:    fmt.Errorf("Response not OK. 404 Not Found: user not found for provider user ID test_user_hey_who_am_i_wow"),
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

	ts.Run("Test_GetPendingLoadBalancersByPortalID", func() {
		tests := []struct {
			name                  string
			userPortalID          string
			expectedLoadBalancers map[string]*v1Types.LoadBalancer
			err                   error
		}{
			{
				name:         "Should fetch all pending load balancers for a single portal user ID",
				userPortalID: "user_10",
				expectedLoadBalancers: map[string]*v1Types.LoadBalancer{
					"test_app_3": expectedLegacyLoadBalancers["test_app_3"],
				},
			},
			{
				name:                  "Should return empty if the portal user ID does not have any pending load balancers in the DB",
				userPortalID:          "user_1",
				expectedLoadBalancers: map[string]*v1Types.LoadBalancer{},
			},
			{
				name:         "Should fail if user portal ID not provided",
				userPortalID: "",
				err:          fmt.Errorf("no user ID"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				pendingLoadBalancersByPortalID, err := ts.client1.GetPendingLoadBalancersByUserID(testCtx, test.userPortalID)
				ts.Equal(test.err, err)
				if test.err == nil {
					ts.Equal(test.expectedLoadBalancers, loadBalancersToMap(pendingLoadBalancersByPortalID))
				}

				pendingLoadBalancersByPortalID, err = ts.client2.GetPendingLoadBalancersByUserID(testCtx, test.userPortalID)
				ts.Equal(test.err, err)
				if test.err == nil {
					ts.Equal(test.expectedLoadBalancers, loadBalancersToMap(pendingLoadBalancersByPortalID))
				}
			})
		}
	})

	ts.Run("Test_GetLoadBalancersCountByPortalID", func() {
		tests := []struct {
			name          string
			portalUserID  string
			expectedCount int
			err           error
		}{
			{
				name:          "Should return the count of load balancers owned by a portal user ID",
				portalUserID:  "user_1",
				expectedCount: 1,
			},
			{
				name:          "Should return the count of load balancers owned by a portal user ID",
				portalUserID:  "user_3",
				expectedCount: 1,
			},
			{
				name:          "Should return 0 if there's no load balancers owned by this portal user ID",
				portalUserID:  "user_9000",
				expectedCount: 0,
			},
			{
				name:         "Should return missing portal user ID if there's no portal user ID",
				portalUserID: "",
				err:          fmt.Errorf("no user ID"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				count, err := ts.client1.GetLoadBalancersCountByUserID(testCtx, test.portalUserID)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedCount, count)

				count, err = ts.client2.GetLoadBalancersCountByUserID(testCtx, test.portalUserID)
				ts.Equal(test.err, err)
				ts.Equal(test.expectedCount, count)
			})
		}
	})
}

// Runs all the write endpoint tests after the read tests
// This ensures the write tests do not modify the seed data expected by the read tests
func (ts *phdE2EWriteTestSuite) Test_WriteTests() {

	/* ------ V2 Chain Create/Update Tests ------ */

	ts.Run("Test_CreateChainAndGigastakeApps", func() {
		tests := []struct {
			name          string
			newChainInput v2Types.NewChainInput
			err           error
		}{
			{
				name:          "Should create a new blockchain and its Gigastake apps in the DB",
				newChainInput: testdata.TestCreateNewChainInput,
			},
			{
				name:          "Should fail if Chain is missing",
				newChainInput: v2Types.NewChainInput{},
				err:           fmt.Errorf("Response not OK. 400 Bad Request: error chain cannot be nil"),
			},
			{
				name: "Should fail if GigastakeApp is missing",
				newChainInput: v2Types.NewChainInput{
					Chain: &v2Types.Chain{ID: "1234"},
				},
				err: fmt.Errorf("Response not OK. 400 Bad Request: error gigastakeApps slice cannot be empty"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				createdChainResp, err := ts.client1.CreateChainAndGigastakeApps(testCtx, test.newChainInput)
				ts.Equal(test.err, err)

				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					createdChain := createdChainResp.Chain
					createdGigastakeApps := createdChainResp.GigastakeApps
					timestamp := createdChain.CreatedAt

					test.newChainInput.Chain.CreatedAt = timestamp
					test.newChainInput.Chain.UpdatedAt = timestamp

					ts.Equal(test.newChainInput.Chain, createdChain)
					for _, expectedApp := range test.newChainInput.GigastakeApps {
						expectedApp.CreatedAt = timestamp
						expectedApp.UpdatedAt = timestamp
						for _, createdApp := range createdGigastakeApps {
							expectedApp.AATID = createdApp.AATID
							expectedApp.ChainID = createdApp.ChainID
							expectedApp.AAT.ID = createdApp.AATID
							ts.Equal(test.newChainInput.GigastakeApps, createdGigastakeApps)
						}
					}

					createdChainByID, err := ts.client1.GetChainByID(testCtx, createdChain.ID)
					ts.NoError(err)
					createdChainByID.CreatedAt = timestamp
					createdChainByID.UpdatedAt = timestamp
					ts.Len(createdChainByID.GigastakeApps, 1)
					createdChainByID.GigastakeApps = nil
					ts.Equal(createdChain, createdChainByID)

					createdChainByID, err = ts.client2.GetChainByID(testCtx, createdChain.ID)
					ts.NoError(err)
					createdChainByID.CreatedAt = timestamp
					createdChainByID.UpdatedAt = timestamp
					ts.Len(createdChainByID.GigastakeApps, 1)
					createdChainByID.GigastakeApps = nil
					ts.Equal(createdChain, createdChainByID)
				}
			})
		}
	})

	ts.Run("Test_CreateGigastakeApp", func() {
		tests := []struct {
			name              string
			gigastakeAppInput v2Types.GigastakeApp
			err               error
			expected          *v2Types.GigastakeApp
		}{
			{
				name:              "Should create a new Gigastake app in the DB",
				gigastakeAppInput: testdata.TestCreateGigastakeApp,
				expected:          &testdata.TestCreateGigastakeApp,
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				createdGigastakeApp, err := ts.client1.CreateGigastakeApp(testCtx, test.gigastakeAppInput)
				ts.Equal(test.err, err)

				if err == nil {
					<-time.After(50 * time.Millisecond)
					timestamp := createdGigastakeApp.CreatedAt

					// Ensure timestamps are the same before comparing
					test.expected.CreatedAt = timestamp
					test.expected.UpdatedAt = timestamp
					test.expected.AATID = createdGigastakeApp.AATID
					test.expected.AAT.ID = createdGigastakeApp.AATID
					test.expected.AAT.PrivateKey = ""

					ts.Equal(test.expected, createdGigastakeApp)
				}
			})
		}
	})

	ts.Run("Test_UpdateChain", func() {
		tests := []struct {
			name        string
			chainUpdate v2Types.Chain
			noSubtables bool
			err         error
		}{
			{
				name:        "Should update the blockchain in the DB",
				chainUpdate: testdata.UpdateChainOne,
			},
			{
				name:        "Should update the blockchain again in the DB",
				chainUpdate: testdata.UpdateChainTwo,
			},
			{
				name:        "Should update the blockchain a third time in the DB without removing any subtables",
				chainUpdate: testdata.UpdateChainThree,
				noSubtables: true, // When no subtables are passed in the update do not modify the subtables of the expected chain
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				chainUpdateResponse, err := ts.client1.UpdateChain(testCtx, test.chainUpdate)
				ts.Equal(test.err, err)

				if test.err == nil {
					<-time.After(50 * time.Millisecond)

					ts.NotEmpty(chainUpdateResponse)

					timestamp := chainUpdateResponse.CreatedAt

					test.chainUpdate.CreatedAt = timestamp
					test.chainUpdate.UpdatedAt = timestamp

					ts.Equal(test.chainUpdate, *chainUpdateResponse)

					updatedChainByID, err := ts.client1.GetChainByID(testCtx, chainUpdateResponse.ID)
					ts.NoError(err)
					if test.noSubtables {
						test.chainUpdate.Altruists = updatedChainByID.Altruists
						test.chainUpdate.Checks = updatedChainByID.Checks
						test.chainUpdate.AliasDomains = updatedChainByID.AliasDomains
					}
					updatedChainByID.CreatedAt = timestamp
					updatedChainByID.UpdatedAt = timestamp
					ts.NotEmpty(updatedChainByID.GigastakeApps, 1)
					updatedChainByID.GigastakeApps = nil
					ts.Equal(test.chainUpdate, *updatedChainByID)

					updatedChainByID, err = ts.client2.GetChainByID(testCtx, chainUpdateResponse.ID)
					if test.noSubtables {
						test.chainUpdate.Altruists = updatedChainByID.Altruists
						test.chainUpdate.Checks = updatedChainByID.Checks
						test.chainUpdate.AliasDomains = updatedChainByID.AliasDomains
					}
					ts.NoError(err)
					updatedChainByID.CreatedAt = timestamp
					updatedChainByID.UpdatedAt = timestamp
					ts.NotEmpty(updatedChainByID.GigastakeApps, 1)
					updatedChainByID.GigastakeApps = nil
					ts.Equal(test.chainUpdate, *updatedChainByID)
				}
			})
		}
	})

	ts.Run("Test_ActivateChain", func() {
		tests := []struct {
			name    string
			chainID v2Types.RelayChainID
			active  bool
			err     error
		}{
			{
				name:    "Should activate a blockchain in the DB",
				chainID: "0064",
				active:  true,
			},
			{
				name:    "Should deactivate a blockchain in the DB",
				chainID: "0064",
				active:  false,
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				chainActive, err := ts.client1.ActivateChain(testCtx, test.chainID, test.active)
				ts.Equal(test.err, err)

				if err == nil {
					<-time.After(50 * time.Millisecond)
					ts.Equal(test.active, chainActive)

					fetchedChain, err := ts.client1.GetChainByID(testCtx, test.chainID)
					ts.NoError(err)
					ts.Equal(test.active, fetchedChain.Active)

					fetchedChain, err = ts.client2.GetChainByID(testCtx, test.chainID)
					ts.NoError(err)
					ts.Equal(test.active, fetchedChain.Active)
				}
			})
		}
	})

	/* ------ V1 Write Tests ------ */

	ts.Run("Test_CreatePortalUser", func() {
		tests := []struct {
			name             string
			userInput        v2Types.CreateUser
			expectedStatus   int
			expectedResponse *v2Types.CreateUserResponse
			err              error
		}{
			{
				name: "Should create a single user in the DB",
				userInput: v2Types.CreateUser{
					Email:          "test@test.com",
					ProviderUserID: "auth0|test",
				},
				expectedResponse: &v2Types.CreateUserResponse{
					User: v2Types.User{
						Email: "test@test.com",
						AuthProviders: map[v2Types.AuthType]v2Types.UserAuthProvider{
							v2Types.AuthTypeAuth0Username: {
								Type:           v2Types.AuthTypeAuth0Username,
								ProviderUserID: "auth0|test",
								Provider:       v2Types.AuthProviderAuth0,
								Federated:      false,
							},
						},
						SignedUp: true,
					},
					AccountID: "",
				},
			},
			{
				name: "Should fail if there's no email",
				userInput: v2Types.CreateUser{
					ProviderUserID: "auth0|test",
				},
				err: fmt.Errorf("Response not OK. 400 Bad Request: error email input is not a valid email address ''"),
			},
			{
				name: "Should fail if there's no provider type",
				userInput: v2Types.CreateUser{
					Email:          "email@test.com",
					ProviderUserID: "wtf|test",
				},
				err: fmt.Errorf("Response not OK. 400 Bad Request: error invalid auth provider type 'wtf'"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				createdUser, err := ts.client1.CreatePortalUser(testCtx, test.userInput)
				ts.Equal(test.err, err)

				if test.err == nil {
					<-time.After(50 * time.Millisecond)

					test.expectedResponse.User.ID = createdUser.User.ID
					test.expectedResponse.User.UpdatedAt = createdUser.User.UpdatedAt
					test.expectedResponse.User.CreatedAt = createdUser.User.CreatedAt
					test.expectedResponse.AccountID = createdUser.AccountID
					ts.Equal(test.expectedResponse, createdUser)
					ts.NotEmpty(createdUser.AccountID, "Should have an accountID")

					// If the user was created, it should have permissions
					providerID := createdUser.User.AuthProviders[v2Types.AuthTypeAuth0Username].ProviderUserID
					permission, err := ts.client1.GetUserPermissionsByUserID(testCtx, v1Types.UserID(providerID))
					ts.NoError(err)
					ts.NotNil(permission)

					permission, err = ts.client2.GetUserPermissionsByUserID(testCtx, v1Types.UserID(providerID))
					ts.NoError(err)
					ts.NotNil(permission)
				}
			})
		}
	})

	ts.Run("Test_CreateLoadBalancer", func() {
		tests := []struct {
			name                   string
			loadBalancer           *v1Types.LoadBalancer
			userID                 string
			expectedCovalentAPIKey string
			err                    error
		}{
			{
				name:                   "Should create a single loadBalancer in the DB",
				loadBalancer:           createLegacyLoadBalancer,
				userID:                 "user_1",
				expectedCovalentAPIKey: "covalent_api_key_1",
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				createdLB, err := ts.client1.CreateLoadBalancer(testCtx, *test.loadBalancer)
				ts.Equal(test.err, err)

				test.loadBalancer.Integrations.CovalentAPIKeyFree = test.expectedCovalentAPIKey
				test.loadBalancer.Applications[0].GatewayAAT.PrivateKey = ""
				test.loadBalancer.ID = createdLB.ID
				test.loadBalancer.UserID = test.userID
				test.loadBalancer.Applications[0].ID = createdLB.Applications[0].ID
				test.loadBalancer.Applications[0].UserID = test.userID
				clearTimeFields(createdLB)

				ts.Equal(test.loadBalancer, createdLB)

				if test.err == nil {
					<-time.After(50 * time.Millisecond)

					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, createdLB.ID)
					ts.Equal(test.err, err)
					clearTimeFields(loadBalancer)
					ts.Equal(test.loadBalancer, loadBalancer)

					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, createdLB.ID)
					ts.Equal(test.err, err)
					clearTimeFields(loadBalancer)
					ts.Equal(test.loadBalancer, loadBalancer)
				}
			})
		}
	})

	ts.Run("Test_UpdateLoadBalancerUserRole", func() {
		tests := []struct {
			name              string
			loadBalancerID    string
			userID            string
			update            v1Types.UpdateUserAccess
			loadBalancerUsers []v1Types.UserAccess
			userLBPermissions map[string]v1Types.LoadBalancerPermissions
			err               error
		}{
			{
				name:           "Should update a single user for an existing load balancer in the DB",
				loadBalancerID: "test_app_1",
				update: v1Types.UpdateUserAccess{
					UserID:   "user_2",
					RoleName: v1Types.RoleMember,
				},
				loadBalancerUsers: []v1Types.UserAccess{
					{RoleName: v1Types.RoleOwner, UserID: "user_1", Email: "james.holden123@test.com", Accepted: true},
					{RoleName: v1Types.RoleAdmin, UserID: "user_8", Email: "rick.deckard456@test.com", Accepted: false},
					{RoleName: v1Types.RoleMember, UserID: "user_2", Email: "paul.atreides456@test.com", Accepted: true},
				},
				userLBPermissions: map[string]v1Types.LoadBalancerPermissions{
					"auth0|paul_atreides": {
						RoleName: v1Types.RoleMember, Permissions: []v1Types.PermissionsEnum{
							"read:endpoint",
						},
					},
				},
			},
			{
				name:           "Should update a single user back for an existing load balancer in the DB",
				loadBalancerID: "test_app_1",
				update: v1Types.UpdateUserAccess{
					UserID:   "user_2",
					RoleName: v1Types.RoleAdmin,
				},
				loadBalancerUsers: []v1Types.UserAccess{
					{RoleName: v1Types.RoleOwner, UserID: "user_1", Email: "james.holden123@test.com", Accepted: true},
					{RoleName: v1Types.RoleAdmin, UserID: "user_2", Email: "paul.atreides456@test.com", Accepted: true},
					{RoleName: v1Types.RoleAdmin, UserID: "user_8", Email: "rick.deckard456@test.com", Accepted: false},
				},
				userLBPermissions: map[string]v1Types.LoadBalancerPermissions{
					"auth0|paul_atreides": {
						RoleName: v1Types.RoleAdmin, Permissions: []v1Types.PermissionsEnum{
							"read:endpoint", "write:endpoint",
						},
					},
				},
			},
			{
				name:           "Should transfer ownership for an existing load balancer in the DB",
				loadBalancerID: "test_app_1",
				update: v1Types.UpdateUserAccess{
					UserID:   "user_2",
					RoleName: v1Types.RoleOwner,
				},
				loadBalancerUsers: []v1Types.UserAccess{
					{RoleName: v1Types.RoleOwner, UserID: "user_2", Email: "paul.atreides456@test.com", Accepted: true},
					{RoleName: v1Types.RoleAdmin, UserID: "user_1", Email: "james.holden123@test.com", Accepted: true},
					{RoleName: v1Types.RoleAdmin, UserID: "user_8", Email: "rick.deckard456@test.com", Accepted: false},
				},
				userLBPermissions: map[string]v1Types.LoadBalancerPermissions{
					"auth0|james_holden": {
						RoleName: v1Types.RoleAdmin, Permissions: []v1Types.PermissionsEnum{
							"read:endpoint", "write:endpoint",
						},
					},
					"auth0|paul_atreides": {
						RoleName: v1Types.RoleOwner, Permissions: []v1Types.PermissionsEnum{
							"read:endpoint", "write:endpoint", "delete:endpoint", "transfer:endpoint",
						},
					},
				},
			},
			{
				name:           "Should transfer ownership back to original owner for an existing load balancer in the DB",
				loadBalancerID: "test_app_1",
				update: v1Types.UpdateUserAccess{
					UserID:   "user_1",
					RoleName: v1Types.RoleOwner,
				},
				loadBalancerUsers: []v1Types.UserAccess{
					{RoleName: v1Types.RoleOwner, UserID: "user_1", Email: "james.holden123@test.com", Accepted: true},
					{RoleName: v1Types.RoleAdmin, UserID: "user_2", Email: "paul.atreides456@test.com", Accepted: true},
					{RoleName: v1Types.RoleAdmin, UserID: "user_8", Email: "rick.deckard456@test.com", Accepted: false},
				},
				userLBPermissions: map[string]v1Types.LoadBalancerPermissions{
					"auth0|paul_atreides": {
						RoleName: v1Types.RoleAdmin, Permissions: []v1Types.PermissionsEnum{
							"read:endpoint", "write:endpoint",
						},
					},
					"auth0|james_holden": {
						RoleName: v1Types.RoleOwner, Permissions: []v1Types.PermissionsEnum{
							"read:endpoint", "write:endpoint", "delete:endpoint", "transfer:endpoint",
						},
					},
				},
			},
			{
				name:           "Should update a single unaccepted user to ADMIN for an existing load balancer in the DB",
				loadBalancerID: "test_app_1",
				update: v1Types.UpdateUserAccess{
					UserID:   "user_8",
					RoleName: v1Types.RoleAdmin,
				},
				loadBalancerUsers: []v1Types.UserAccess{
					{RoleName: v1Types.RoleOwner, UserID: "user_1", Email: "james.holden123@test.com", Accepted: true},
					{RoleName: v1Types.RoleAdmin, UserID: "user_2", Email: "paul.atreides456@test.com", Accepted: true},
					{RoleName: v1Types.RoleAdmin, UserID: "user_8", Email: "rick.deckard456@test.com", Accepted: false},
				},
			},
			{
				name:           "Should fail if attempting to transfer ownership and the user has not accepted their invite",
				loadBalancerID: "test_app_1",
				update: v1Types.UpdateUserAccess{
					UserID:   "user_8",
					RoleName: v1Types.RoleOwner,
				},
				err: fmt.Errorf("Response not OK. 500 Internal Server Error: error cannot transfer ownership to user ID 'user_8' for account ID 'account_1' because the user has not accepted their invite"),
			},
			{
				name:           "Should fail if load balancer ID not provided",
				loadBalancerID: "",
				err:            fmt.Errorf("no load balancer ID"),
			},
			{
				name:           "Should fail if invalid role name provided",
				loadBalancerID: "test_app_1",
				update: v1Types.UpdateUserAccess{
					UserID:   "user_8",
					RoleName: v1Types.RoleName("wrong_one"),
				},
				err: fmt.Errorf("invalid role name"),
			},
			{
				name:           "Should fail if load balancer cannot be found",
				loadBalancerID: "im_not_here",
				update: v1Types.UpdateUserAccess{
					UserID:   "user_8",
					RoleName: v1Types.RoleMember,
				},
				err: fmt.Errorf("Response not OK. 500 Internal Server Error: portal app not found for load balancer ID im_not_here"),
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

					if len(test.userLBPermissions) > 0 {
						for providerUserID, userLBPermissions := range test.userLBPermissions {
							permissionsByUserID, err := ts.client1.GetUserPermissionsByUserID(testCtx, v1Types.UserID(providerUserID))
							ts.Equal(test.err, err)
							ts.Equal(userLBPermissions, permissionsByUserID.LoadBalancers[v1Types.LoadBalancerID(test.loadBalancerID)])

							permissionsByUserID, err = ts.client2.GetUserPermissionsByUserID(testCtx, v1Types.UserID(providerUserID))
							ts.Equal(test.err, err)
							ts.Equal(userLBPermissions, permissionsByUserID.LoadBalancers[v1Types.LoadBalancerID(test.loadBalancerID)])
						}
					}
				}
			})
		}
	})

	ts.Run("Test_UpdateLoadBalancer", func() {
		tests := []struct {
			name                   string
			loadBalancerID         string
			applicationUpdate      v1Types.UpdateApplication
			applicationAfterUpdate v1Types.Application
			err                    error
		}{
			{
				name:           "Should update a single application in the DB",
				loadBalancerID: "test_app_1",
				applicationUpdate: v1Types.UpdateApplication{
					Name: "test_update_portal_app_123",
					GatewaySettings: &v1Types.UpdateGatewaySettings{
						SecretKey:            "test_90210ac4bdd3423e24877d1ff92",
						SecretKeyRequired:    boolToPointer(false),
						WhitelistOrigins:     []string{"https://portalgun.io", "https://subdomain.example.com", "https://www.example.com"},
						WhitelistBlockchains: []string{"0001", "0002", "003E", "0056"},
						WhitelistUserAgents:  []string{"Brave", "Google Chrome", "Mozilla Firefox", "Netscape Navigator", "Safari"},
						WhitelistContracts: []v1Types.WhitelistContracts{
							{BlockchainID: "0001", Contracts: []string{"0xtest_2f78db6436527729929aaf6c616361de0f7", "0xtest_5fbfe3e9af3971dd833d26ba9b5c936f0be"}},
							{BlockchainID: "0002", Contracts: []string{"0xtest_1111117dc0aa78b770fa6a738034120c302", "0xtest_a39b223fe8d0a0e5c4f27ead9083c756cc2"}},
							{BlockchainID: "003E", Contracts: []string{"0xtest_0a85d5af5bf1d1762f925bdaddc4201f984", "0xtest_f958d2ee523a2206206994597c13d831ec7"}},
							{BlockchainID: "0056", Contracts: []string{"0xtest_00000f279d81a1d3cc75430faa017fa5a2e", "0xtest_5068778dd592e39a122f4f5a5cf09c90fe2"}},
						},
						WhitelistMethods: []v1Types.WhitelistMethods{
							{BlockchainID: "0001", Methods: []string{"GET", "POST", "PUT"}},
							{BlockchainID: "0002", Methods: []string{"DELETE", "GET", "POST", "PUT"}},
							{BlockchainID: "003E", Methods: []string{"GET"}},
							{BlockchainID: "0056", Methods: []string{"GET", "POST"}},
						},
					},
					NotificationSettings: &v1Types.UpdateNotificationSettings{SignedUp: boolToPointer(true), Quarter: boolToPointer(true), Half: boolToPointer(false), ThreeQuarters: boolToPointer(true), Full: boolToPointer(false)},
				},
				applicationAfterUpdate: v1Types.Application{
					Name: "test_update_portal_app_123",
					GatewaySettings: v1Types.GatewaySettings{
						SecretKey:            "test_90210ac4bdd3423e24877d1ff92",
						WhitelistOrigins:     []string{"https://portalgun.io", "https://subdomain.example.com", "https://www.example.com"},
						WhitelistBlockchains: []string{"0001", "0002", "003E", "0056"},
						WhitelistUserAgents:  []string{"Brave", "Google Chrome", "Mozilla Firefox", "Netscape Navigator", "Safari"},
						WhitelistContracts: []v1Types.WhitelistContracts{
							{BlockchainID: "0001", Contracts: []string{"0xtest_2f78db6436527729929aaf6c616361de0f7", "0xtest_5fbfe3e9af3971dd833d26ba9b5c936f0be"}},
							{BlockchainID: "0002", Contracts: []string{"0xtest_1111117dc0aa78b770fa6a738034120c302", "0xtest_a39b223fe8d0a0e5c4f27ead9083c756cc2"}},
							{BlockchainID: "003E", Contracts: []string{"0xtest_0a85d5af5bf1d1762f925bdaddc4201f984", "0xtest_f958d2ee523a2206206994597c13d831ec7"}},
							{BlockchainID: "0056", Contracts: []string{"0xtest_00000f279d81a1d3cc75430faa017fa5a2e", "0xtest_5068778dd592e39a122f4f5a5cf09c90fe2"}},
						},
						WhitelistMethods: []v1Types.WhitelistMethods{
							{BlockchainID: "0001", Methods: []string{"GET", "POST", "PUT"}},
							{BlockchainID: "0002", Methods: []string{"DELETE", "GET", "POST", "PUT"}},
							{BlockchainID: "003E", Methods: []string{"GET"}},
							{BlockchainID: "0056", Methods: []string{"GET", "POST"}},
						},
					},
					NotificationSettings: v1Types.NotificationSettings{SignedUp: true, Quarter: true, Half: false, ThreeQuarters: true, Full: false},
				},
			},
			{
				name:           "Should remove all of a single application's whitelists",
				loadBalancerID: "test_app_1",
				applicationUpdate: v1Types.UpdateApplication{
					GatewaySettings: &v1Types.UpdateGatewaySettings{},
				},
				applicationAfterUpdate: v1Types.Application{
					Name: "test_update_portal_app_123",
					GatewaySettings: v1Types.GatewaySettings{
						SecretKey: "test_90210ac4bdd3423e24877d1ff92",
					},
					NotificationSettings: v1Types.NotificationSettings{SignedUp: true, Quarter: true, Half: false, ThreeQuarters: true, Full: false},
				},
			},
			{
				name:           "Should fail if application cannot be found",
				loadBalancerID: "9000",
				err:            fmt.Errorf("Response not OK. 404 Not Found: portal app not found for load balancer ID 9000"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				createdLB, err := ts.client1.UpdateLoadBalancer(testCtx, test.loadBalancerID, test.applicationUpdate)
				ts.Equal(test.err, err)

				if err == nil {
					<-time.After(50 * time.Millisecond)

					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, createdLB.ID)
					application := loadBalancer.Applications[0]
					ts.NoError(err)
					ts.Equal(test.applicationAfterUpdate.Name, application.Name)
					ts.Equal(test.applicationAfterUpdate.GatewaySettings, application.GatewaySettings)
					ts.Equal(test.applicationAfterUpdate.NotificationSettings, application.NotificationSettings)

					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, createdLB.ID)
					application = loadBalancer.Applications[0]
					ts.NoError(err)
					ts.Equal(test.applicationAfterUpdate.Name, application.Name)
					ts.Equal(test.applicationAfterUpdate.GatewaySettings, application.GatewaySettings)
					ts.Equal(test.applicationAfterUpdate.NotificationSettings, application.NotificationSettings)
				}
			})
		}
	})

	ts.Run("Test_UpdateAppFirstDateSurpassed", func() {
		tests := []struct {
			name           string
			update         v1Types.UpdateFirstDateSurpassed
			protocolAppIDs []string
			expectedDate   time.Time
			err            error
		}{
			{
				name: "Should update the app first date suprassed for the provided slice of app IDs",
				update: v1Types.UpdateFirstDateSurpassed{
					ApplicationIDs:     []string{"test_app_1", "test_app_2"},
					FirstDateSurpassed: time.Date(2022, time.November, 11, 11, 11, 11, 0, time.UTC),
				},
				protocolAppIDs: []string{"test_protocol_app_1", "test_protocol_app_2"},
				expectedDate:   time.Date(2022, time.November, 11, 11, 11, 11, 0, time.UTC),
				err:            nil,
			},
			{
				name: "Should fail if update contains no application IDs cannot be found",
				update: v1Types.UpdateFirstDateSurpassed{
					ApplicationIDs:     []string{},
					FirstDateSurpassed: time.Date(2022, time.November, 11, 11, 11, 11, 0, time.UTC),
				},
				err: fmt.Errorf("Response not OK. 400 Bad Request: no application IDs on input"),
			},
			{
				name: "Should fail if application cannot be found",
				update: v1Types.UpdateFirstDateSurpassed{
					ApplicationIDs:     []string{"9000"},
					FirstDateSurpassed: time.Date(2022, time.November, 11, 11, 11, 11, 0, time.UTC),
				},
				err: fmt.Errorf("Response not OK. 400 Bad Request: UpdateFirstDateSurpassed failed: 9000 not found"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.UpdateAppFirstDateSurpassed(testCtx, test.update)
				ts.Equal(test.err, err)

				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					for _, appID := range test.protocolAppIDs {
						applications, err := ts.client1.GetApplications(testCtx)
						ts.NoError(err)
						exists := false
						for _, application := range applications {
							if application.ID == appID {
								exists = true
								ts.Equal(test.expectedDate, application.FirstDateSurpassed)
							}
						}
						ts.True(exists)

						applications, err = ts.client2.GetApplications(testCtx)
						ts.NoError(err)
						exists = false
						for _, application := range applications {
							if application.ID == appID {
								exists = true
								ts.Equal(test.expectedDate, application.FirstDateSurpassed)
							}
						}
						ts.True(exists)
					}
				}
			})
		}
	})

	ts.Run("Test_AcceptLoadBalancerUser", func() {
		tests := []struct {
			name                                  string
			email, loadBalancerID, providerUserID string
			loadBalancerUsers                     []v1Types.UserAccess
			userLBPermissions                     v1Types.LoadBalancerPermissions
			err                                   error
			pendingResult                         error
		}{
			{
				name:           "Should update a single user's ID and Accepted field for an existing load balancer in the DB",
				loadBalancerID: "test_app_1",
				providerUserID: "auth0|rick_deckard",
				loadBalancerUsers: []v1Types.UserAccess{
					{RoleName: v1Types.RoleOwner, UserID: "user_1", Email: "james.holden123@test.com", Accepted: true},
					{RoleName: v1Types.RoleAdmin, UserID: "user_2", Email: "paul.atreides456@test.com", Accepted: true},
					{RoleName: v1Types.RoleAdmin, UserID: "user_8", Email: "rick.deckard456@test.com", Accepted: true},
				},
				userLBPermissions: v1Types.LoadBalancerPermissions{
					RoleName:    v1Types.RoleAdmin,
					Permissions: []v1Types.PermissionsEnum{v1Types.ReadEndpoint, v1Types.WriteEndpoint},
				},
				pendingResult: fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			},
			{
				name:           "Should fail if load balancer ID not provided",
				providerUserID: "auth0|rick_deckard",
				loadBalancerID: "",
				err:            fmt.Errorf("no load balancer ID"),
			},
			{
				name:           "Should fail if user ID not provided",
				loadBalancerID: "test_app_1",
				err:            fmt.Errorf("no user ID"),
			},
			{
				name:           "Should fail if load balancer cannot be found",
				providerUserID: "auth0|rick_deckard",
				loadBalancerID: "im_not_here",
				err:            fmt.Errorf("Response not OK. 500 Internal Server Error: portal app not found for load balancer ID im_not_here"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.AcceptLoadBalancerUser(testCtx, test.loadBalancerID, test.providerUserID)
				ts.Equal(test.err, err)
				if test.err == nil {
					<-time.After(50 * time.Millisecond)

					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.Equal(test.err, err)
					ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.Equal(test.err, err)
					ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

					if test.providerUserID != "" {
						// user should have been added to the user permissions in the cache
						permissionsByUserID, err := ts.client1.GetUserPermissionsByUserID(testCtx, v1Types.UserID(test.providerUserID))
						ts.Equal(test.err, err)
						ts.Equal(test.userLBPermissions, permissionsByUserID.LoadBalancers[v1Types.LoadBalancerID(test.loadBalancerID)])

						permissionsByUserID, err = ts.client2.GetUserPermissionsByUserID(testCtx, v1Types.UserID(test.providerUserID))
						ts.Equal(test.err, err)
						ts.Equal(test.userLBPermissions, permissionsByUserID.LoadBalancers[v1Types.LoadBalancerID(test.loadBalancerID)])

						exists := false
						loadBalancers, err := ts.client1.GetLoadBalancersByUserID(testCtx, test.providerUserID, nil)
						ts.Equal(test.err, err)
						for _, lb := range loadBalancers {
							if lb.ID == test.loadBalancerID {
								exists = true
								break
							}
						}
						ts.True(exists)

						exists = false
						loadBalancers, err = ts.client2.GetLoadBalancersByUserID(testCtx, test.providerUserID, nil)
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
						pendingLoadBalancers, _ := ts.client1.GetPendingLoadBalancersByUserID(testCtx, test.providerUserID)
						for _, lb := range pendingLoadBalancers {
							if lb.ID == test.loadBalancerID {
								removed = false
								break
							}
						}
						ts.True(removed)

						removed = true
						pendingLoadBalancers, _ = ts.client2.GetPendingLoadBalancersByUserID(testCtx, test.providerUserID)
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

	ts.Run("Test_CreateLoadBalancerUser", func() {
		tests := []struct {
			name              string
			loadBalancerID    string
			user              v1Types.UserAccess
			loadBalancerUsers map[string]v1Types.UserAccess
			err               error
		}{
			{
				name:           "Should add a single new user to an existing load balancer in the DB",
				loadBalancerID: "test_app_1",
				user: v1Types.UserAccess{
					RoleName: v1Types.RoleMember,
					Email:    "member_new@test.com",
				},
				loadBalancerUsers: map[string]v1Types.UserAccess{
					"user_1": {UserID: "user_1", RoleName: v1Types.RoleOwner, Email: "james.holden123@test.com", Accepted: true},
					"user_2": {UserID: "user_2", RoleName: v1Types.RoleAdmin, Email: "paul.atreides456@test.com", Accepted: true},
					"user_8": {UserID: "user_8", RoleName: v1Types.RoleAdmin, Email: "rick.deckard456@test.com", Accepted: true},
					// ID dynamically generated on creation
					"": {UserID: "", RoleName: v1Types.RoleMember, Email: "member_new@test.com", Accepted: false},
				},
			},
			{
				name:           "Should add a single existing user to an existing load balancer in the DB",
				loadBalancerID: "test_app_2",
				user: v1Types.UserAccess{
					RoleName: v1Types.RoleMember,
					Email:    "frodo.baggins123@test.com",
				},
				loadBalancerUsers: map[string]v1Types.UserAccess{
					"user_3": {UserID: "user_3", RoleName: v1Types.RoleOwner, Email: "ellen.ripley789@test.com", Accepted: true},
					"user_4": {UserID: "user_4", RoleName: v1Types.RoleMember, Email: "ulfric.stormcloak123@test.com", Accepted: true},
					"user_9": {UserID: "user_9", RoleName: v1Types.RoleMember, Email: "tyrion.lannister789@test.com", Accepted: false},
					"user_2": {UserID: "user_2", RoleName: v1Types.RoleMember, Email: "paul.atreides456@test.com", Accepted: true},
					// ID dynamically generated on creation
					"": {UserID: "", RoleName: v1Types.RoleMember, Email: "frodo.baggins123@test.com", Accepted: false},
				},
			},
			{
				name:           "Should fail if load balancer cannot be found",
				loadBalancerID: "sir_not_appearing_in_this_film",
				err:            fmt.Errorf("Response not OK. 500 Internal Server Error: portal app not found for load balancer ID sir_not_appearing_in_this_film"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				updatedLB, err := ts.client1.CreateLoadBalancerUser(testCtx, test.loadBalancerID, test.user)
				ts.Equal(test.err, err)
				if test.err == nil {
					<-time.After(50 * time.Millisecond)

					// Find the user in updatedLB.Users with the same email as test.user
					for _, updatedUser := range updatedLB.Users {
						if updatedUser.Email == test.user.Email {
							// Update the UserID in test.loadBalancerUsers
							lbUser := test.loadBalancerUsers[""]
							lbUser.UserID = updatedUser.UserID
							test.loadBalancerUsers[updatedUser.UserID] = lbUser
							delete(test.loadBalancerUsers, "")
							break
						}
					}
					ts.Equal(test.loadBalancerUsers, userAccessSliceToMap(updatedLB.Users))

					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.Equal(test.err, err)
					ts.Equal(test.loadBalancerUsers, userAccessSliceToMap(loadBalancer.Users))

					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.Equal(test.err, err)
					ts.Equal(test.loadBalancerUsers, userAccessSliceToMap(loadBalancer.Users))
				}
			})
		}
	})

	ts.Run("Test_DeleteLoadBalancerUser", func() {
		tests := []struct {
			name                       string
			loadBalancerID             string
			userID, providerUserID     string
			loadBalancerUsers          []v1Types.UserAccess
			lbsAfterDelete             int
			err, permissionsFetchError error
		}{
			{
				name:           "Should remove a single user from an existing load balancer in the DB",
				loadBalancerID: "test_app_3",
				userID:         "user_7",
				providerUserID: "auth0|frodo_baggins",
				loadBalancerUsers: []v1Types.UserAccess{
					{RoleName: v1Types.RoleOwner, UserID: "user_5", Email: "chrisjen.avasarala1@test.com", Accepted: true},
					{RoleName: v1Types.RoleAdmin, UserID: "user_6", Email: "amos.burton789@test.com", Accepted: true},
					{RoleName: v1Types.RoleMember, UserID: "user_10", Email: "daenerys.targaryen123@test.com", Accepted: false},
				},
				lbsAfterDelete:        1,
				permissionsFetchError: fmt.Errorf("Response not OK. 404 Not Found: user not found for provider user ID user_7"),
			},
			{
				name:           "Should fail if load balancer cannot be found",
				loadBalancerID: "why_am_i_not_a_load_balancer",
				userID:         "user_9",
				err:            fmt.Errorf("Response not OK. 500 Internal Server Error: portal app not found for load balancer ID why_am_i_not_a_load_balancer"),
			},
			{
				name:           "Should fail if load balancer ID not provided",
				loadBalancerID: "",
				userID:         "user_9",
				err:            fmt.Errorf("no load balancer ID"),
			},
			{
				name:           "Should fail if user ID not provided",
				loadBalancerID: "test_app_1",
				userID:         "",
				err:            fmt.Errorf("no user ID"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.DeleteLoadBalancerUser(testCtx, test.loadBalancerID, test.userID)
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
					_, err = ts.client1.GetUserPermissionsByUserID(testCtx, v1Types.UserID(test.userID))
					ts.Equal(test.permissionsFetchError, err)

					_, err = ts.client2.GetUserPermissionsByUserID(testCtx, v1Types.UserID(test.userID))
					ts.Equal(test.permissionsFetchError, err)

					// user's LB should have been removed from the user's LBs map in the cache
					lbs, err := ts.client1.GetLoadBalancersByUserID(testCtx, test.providerUserID, nil)
					ts.NoError(err)
					ts.Len(lbs, test.lbsAfterDelete)

					lbs, err = ts.client2.GetLoadBalancersByUserID(testCtx, test.providerUserID, nil)
					ts.NoError(err)
					ts.Len(lbs, test.lbsAfterDelete)
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
				name:           "should remove one load balancer by setting its delete flag into true",
				loadBalancerID: "test_app_3",
				expectedUserID: "",
			},
			{
				name:           "Should fail if load balancer cannot be found",
				loadBalancerID: "9000",
				err:            fmt.Errorf("Response not OK. 404 Not Found: portal app not found for load balancer ID 9000"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.RemoveLoadBalancer(testCtx, test.loadBalancerID)
				ts.Equal(test.err, err)

				if test.err == nil {
					<-time.After(50 * time.Millisecond)

					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.Equal("Response not OK. 404 Not Found: portal app not found for load balancer ID test_app_3", err.Error())
					ts.Nil(loadBalancer)

					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
					ts.Equal("Response not OK. 404 Not Found: portal app not found for load balancer ID test_app_3", err.Error())
					ts.Nil(loadBalancer)
				}
			})
		}
	})

	ts.Run("Test_RemoveApplication", func() {
		tests := []struct {
			name          string
			applicationID string
			protocolAppID string
			expectedError error
			err           error
		}{
			{
				name:          "should remove one application by setting its status to deleted",
				applicationID: "test_app_2",
				protocolAppID: "test_protocol_app_2",
				expectedError: fmt.Errorf("Response not OK. 404 Not Found: portal app not found for app ID test_protocol_app_2"),
			},
			{
				name:          "Should fail if application cannot be found",
				applicationID: "2348",
				err:           fmt.Errorf("Response not OK. 404 Not Found: portal app not found for load balancer ID 2348"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.RemoveApplication(testCtx, test.applicationID)
				ts.Equal(test.err, err)

				if test.err == nil {
					<-time.After(50 * time.Millisecond)

					applications, err := ts.client1.GetApplications(testCtx)
					ts.NoError(err)
					exists := false
					for _, application := range applications {
						if application.ID == test.applicationID {
							exists = true
						}
					}
					ts.False(exists)

					applications, err = ts.client2.GetApplications(testCtx)
					ts.NoError(err)
					exists = false
					for _, application := range applications {
						if application.ID == test.applicationID {
							exists = true
						}
					}
					ts.False(exists)
				}
			})
		}
	})

}

func blockchainsToMap(blockchains []*v1Types.Blockchain) map[string]*v1Types.Blockchain {
	expectedMap := make(map[string]*v1Types.Blockchain)
	for _, b := range blockchains {
		expectedMap[b.ID] = b
	}
	return expectedMap
}

func applicationsToMap(applications []*v1Types.Application) map[string]*v1Types.Application {
	expectedMap := make(map[string]*v1Types.Application)
	for _, a := range applications {
		expectedMap[a.ID] = a
	}
	return expectedMap
}

func loadBalancersToMap(loadBalancers []*v1Types.LoadBalancer) map[string]*v1Types.LoadBalancer {
	lbMap := make(map[string]*v1Types.LoadBalancer)
	for _, lb := range loadBalancers {
		sortUsersByRole(lb.Users)

		// Sort Applications by ID
		sort.Slice(lb.Applications, func(i, j int) bool {
			return lb.Applications[i].ID < lb.Applications[j].ID
		})

		lbMap[lb.ID] = lb
	}
	return lbMap
}

func sortUsersByRole(users []v1Types.UserAccess) {
	roleWeight := map[v1Types.RoleName]int{v1Types.RoleOwner: 0, v1Types.RoleAdmin: 1, v1Types.RoleMember: 2}

	sort.Slice(users, func(i, j int) bool {
		if roleWeight[users[i].RoleName] != roleWeight[users[j].RoleName] {
			return roleWeight[users[i].RoleName] < roleWeight[users[j].RoleName]
		}
		return users[i].UserID < users[j].UserID
	})
}

func payPlansToMap(payPlans []*v1Types.PayPlan) map[string]*v1Types.PayPlan {
	payPlansMap := make(map[string]*v1Types.PayPlan)
	for _, payPlan := range payPlans {
		payPlansMap[string(payPlan.Type)] = payPlan
	}
	return payPlansMap
}

func userAccessSliceToMap(users []v1Types.UserAccess) map[string]v1Types.UserAccess {
	userMap := make(map[string]v1Types.UserAccess)
	for _, user := range users {
		userMap[user.UserID] = user
	}
	return userMap
}

func clearTimeFields(lb *v1Types.LoadBalancer) {
	lb.CreatedAt = time.Time{}
	lb.UpdatedAt = time.Time{}
	if len(lb.Applications) > 0 {
		lb.Applications[0].CreatedAt = time.Time{}
		lb.Applications[0].UpdatedAt = time.Time{}
	}
}

func boolToPointer(value bool) *bool {
	return &value
}

var (
	mockTimestamp = time.Date(2022, time.November, 11, 11, 11, 11, 0, time.UTC)

	expectedLegacyApplications = map[string]*v1Types.Application{
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
			FirstDateSurpassed: mockTimestamp,
			CreatedAt:          mockTimestamp,
			UpdatedAt:          mockTimestamp,
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
			FirstDateSurpassed: mockTimestamp,
			CreatedAt:          mockTimestamp,
			UpdatedAt:          mockTimestamp,
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
			FirstDateSurpassed: mockTimestamp,
			CreatedAt:          mockTimestamp,
			UpdatedAt:          mockTimestamp,
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
			FirstDateSurpassed: mockTimestamp,
			CreatedAt:          mockTimestamp,
			UpdatedAt:          mockTimestamp,
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
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
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
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
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
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
	}

	expectedLegacyLoadBalancers = map[string]*v1Types.LoadBalancer{
		"test_app_1": {
			ID:                "test_app_1",
			AccountID:         "account_1",
			Name:              "pokt_app_123",
			UserID:            "user_1",
			GigastakeRedirect: true,
			RequestTimeout:    5_000,
			Applications: []*v1Types.Application{
				expectedLegacyApplications["test_protocol_app_1"],
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
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
		"test_app_2": {
			ID:                "test_app_2",
			AccountID:         "account_2",
			Name:              "pokt_app_456",
			UserID:            "user_3",
			GigastakeRedirect: true,
			RequestTimeout:    10_000,
			Applications: []*v1Types.Application{
				expectedLegacyApplications["test_protocol_app_2"],
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
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
		"test_app_3": {
			ID:                "test_app_3",
			AccountID:         "account_3",
			Name:              "pokt_app_789",
			UserID:            "user_5",
			GigastakeRedirect: true,
			RequestTimeout:    10_000,
			Applications: []*v1Types.Application{
				expectedLegacyApplications["test_protocol_app_3"],
				expectedLegacyApplications["test_protocol_app_4"],
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
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
		"legacy_lb_1": {
			ID:        "legacy_lb_1",
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
					CreatedAt: mockTimestamp,
					UpdatedAt: mockTimestamp,
				},
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
		"legacy_lb_2": {
			ID:        "legacy_lb_2",
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
					CreatedAt: mockTimestamp,
					UpdatedAt: mockTimestamp,
				},
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
		"legacy_lb_3": {
			ID:        "legacy_lb_3",
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
					CreatedAt: mockTimestamp,
					UpdatedAt: mockTimestamp,
				},
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
	}

	expectedLegacyPayPlans = map[string]*v1Types.PayPlan{
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

	expectedLegacyBlockchains = map[string]*v1Types.Blockchain{
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
				{Alias: "pokt-mainnet", Domain: "pokt-rpc.gateway.pokt.network", LoadBalancerID: "legacy_lb_1"},
			},
			SyncCheckOptions: v1Types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"query\"}",
				ResultKey: "result.sync_info", Allowance: 1,
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
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
				{Alias: "eth-mainnet", Domain: "eth-rpc.gateway.pokt.network", LoadBalancerID: ""},
			},
			SyncCheckOptions: v1Types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[]}",
				ResultKey: "result", Allowance: 5,
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
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
				{Alias: "harmony-0", Domain: "hmy-rpc.gateway.pokt.network", LoadBalancerID: "legacy_lb_3"},
			},
			SyncCheckOptions: v1Types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"hmy_blockNumber\",\"params\":[]}",
				ResultKey: "result",
				Allowance: 8,
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
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
				{Alias: "optimism-mainnet", Domain: "op-rpc.gateway.pokt.network", LoadBalancerID: "legacy_lb_2"},
			},
			SyncCheckOptions: v1Types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[]}",
				ResultKey: "result",
				Allowance: 2,
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
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
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
	}

	createLegacyLoadBalancer = &v1Types.LoadBalancer{
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
				Limit:                v1Types.AppLimit{PayPlan: v1Types.PayPlan{Type: "basic_plan", Limit: 1_000}, CustomLimit: 0},
				NotificationSettings: v1Types.NotificationSettings{SignedUp: true, Quarter: false, Half: false, ThreeQuarters: true, Full: true},
			},
		},
		Users: []v1Types.UserAccess{
			{UserID: "user_1", RoleName: v1Types.RoleOwner, Email: "james.holden123@test.com", Accepted: true},
		},
		AccountID: "account_1",
	}

	testCtx = context.Background()
)

/* ---------- Test Suite Util Interfaces ---------- */
type phdE2EReadTestSuite struct {
	suite.Suite
	client1, client2 dbclient.IDBClient
}

type phdE2EWriteTestSuite struct {
	suite.Suite
	client1, client2 dbclient.IDBClient
}

func initDBClient(ts DBClientInitializer) error {
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
	ts.SetClient1(client1)

	config2 := baseConfig
	config2.BaseURL = "http://localhost:8081"
	client2, err := dbclient.NewDBClient(config2)
	if err != nil {
		return err
	}
	ts.SetClient2(client2)

	return nil
}

type DBClientInitializer interface {
	SetClient1(client dbclient.IDBClient)
	SetClient2(client dbclient.IDBClient)
	NoError(err error)
}

func (ts *phdE2EReadTestSuite) SetClient1(client dbclient.IDBClient) {
	ts.client1 = client
}
func (ts *phdE2EReadTestSuite) SetClient2(client dbclient.IDBClient) {
	ts.client2 = client
}
func (ts *phdE2EWriteTestSuite) SetClient1(client dbclient.IDBClient) {
	ts.client1 = client
}
func (ts *phdE2EWriteTestSuite) SetClient2(client dbclient.IDBClient) {
	ts.client2 = client
}
func (ts *phdE2EReadTestSuite) NoError(err error) {
	ts.Suite.NoError(err)
}
func (ts *phdE2EWriteTestSuite) NoError(err error) {
	ts.Suite.NoError(err)
}
