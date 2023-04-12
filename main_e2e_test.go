package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"

	dbclient "github.com/pokt-foundation/db-client/client"
	"github.com/pokt-foundation/portal-db/v2/types"
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

func PrettyString(label string, thing interface{}) {
	jsonThing, _ := json.Marshal(thing)
	str := string(jsonThing)

	var prettyJSON bytes.Buffer
	_ = json.Indent(&prettyJSON, []byte(str), "", "    ")
	output := prettyJSON.String()

	fmt.Println(label, output)
}

// func dereferenceSlice(slice interface{}) interface{} {
// 	sliceValue := reflect.ValueOf(slice)
// 	if sliceValue.Kind() != reflect.Slice {
// 		panic("dereferenceSlice: not a slice")
// 	}

// 	elemType := sliceValue.Type().Elem()
// 	if elemType.Kind() != reflect.Ptr {
// 		panic("dereferenceSlice: slice element type is not a pointer")
// 	}

// 	dereferencedSlice := reflect.MakeSlice(reflect.SliceOf(elemType.Elem()), sliceValue.Len(), sliceValue.Cap())
// 	for i := 0; i < sliceValue.Len(); i++ {
// 		dereferencedSlice.Index(i).Set(sliceValue.Index(i).Elem())
// 	}

// 	return dereferencedSlice.Interface()
// }

// func dereferenceMap(m map[string]*types.LoadBalancer) map[string]types.LoadBalancer {
// 	dereferencedMap := make(map[string]types.LoadBalancer)
// 	for k, v := range m {
// 		dereferencedMap[k] = *v
// 	}
// 	return dereferencedMap
// }

// Runs all the read-only endpoint tests first to compare to test DB seed data only
// ie. not yet including data written to the test DB by the test suite
func (ts *phdE2EReadTestSuite) Test_ReadTests() {

	ts.Run("Test_GetBlockchains", func() {
		tests := []struct {
			name                string
			expectedBlockchains map[string]*types.Blockchain
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
				ts.ErrorIs(test.err, err)
				ts.Equal(test.expectedBlockchains, blockchainsToMap(blockchains))
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
			expectedApplications map[string]*types.Application
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

	ts.Run("Test_GetApplicationByID", func() {
		tests := []struct {
			name                string
			applicationID       string
			expectedApplication *types.Application
			err                 error
		}{
			{
				name:                "Should fetch one application by ID",
				applicationID:       "test_protocol_app_1",
				expectedApplication: expectedLegacyApplications["test_protocol_app_1"],
			},
			{
				name:          "Should fail if the application does not exist in the DB",
				applicationID: "test_not_real_app",
				err:           fmt.Errorf("Response not OK. 404 Not Found: portal app not found for app ID test_not_real_app"),
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
			expectedApplications map[string]*types.Application
			err                  error
		}{
			{
				name:   "Should fetch all applications for a single user ID",
				userID: "auth0|chrisjen_avasarala",
				expectedApplications: map[string]*types.Application{
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
			expectedLoadBalancers map[string]*types.LoadBalancer
			err                   error
		}{
			{
				name: "Should fetch all load balancers in the database (for portal apps that have an account)",
				expectedLoadBalancers: map[string]*types.LoadBalancer{
					"test_app_1": expectedLegacyLoadBalancers["test_app_1"],
					"test_app_2": expectedLegacyLoadBalancers["test_app_2"],
					"test_app_3": expectedLegacyLoadBalancers["test_app_3"],
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
			expectedLoadBalancer *types.LoadBalancer
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
			expectedLoadBalancers map[string]*types.LoadBalancer
			roleNameFilter        types.RoleName
			err                   error
		}{
			{
				name:   "Should fetch all load balancers for a single user ID when no filter provided",
				userID: "auth0|ulfric_stormcloak",
				expectedLoadBalancers: map[string]*types.LoadBalancer{
					"test_app_2": expectedLegacyLoadBalancers["test_app_2"],
				},
			},
			{
				name:           "Should fetch all load balancers for a single user ID and role when a valid filter provided",
				userID:         "auth0|amos_burton",
				roleNameFilter: types.RoleAdmin,
				expectedLoadBalancers: map[string]*types.LoadBalancer{
					"test_app_3": expectedLegacyLoadBalancers["test_app_3"],
				},
			},
			{
				name:                  "Should return empty if the user does not have any load balancers",
				userID:                "auth0|bernard_marx",
				expectedLoadBalancers: map[string]*types.LoadBalancer{},
			},
			{
				name:           "Should fail if an invalid role name provided as a filter",
				userID:         "test_user_1dbffbdfeeb225",
				roleNameFilter: types.RoleName("not_real"),
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
			expectedPayPlans map[string]*types.PayPlan
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
			payPlanType     types.PayPlanType
			expectedPayPlan *types.PayPlan
			err             error
		}{
			{

				name:            "Should fetch a single pay plan by type",
				payPlanType:     "pro_plan",
				expectedPayPlan: expectedLegacyPayPlans["pro_plan"],
			},
			{
				name:        "Should fail if passed a pay plan type that is not in the DB",
				payPlanType: types.PayPlanType("not_a_real_plan"),
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
			userID              types.UserID
			expectedPermissions *types.LegacyUserPermissions
			err                 error
		}{
			{

				name:   "Should fetch a single users load balancer permissions",
				userID: "auth0|paul_atreides",
				expectedPermissions: &types.LegacyUserPermissions{
					UserID: "auth0|paul_atreides",
					LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{
						"test_app_1": {
							RoleName:    "ADMIN",
							Permissions: []types.Permissions{"read:endpoint", "write:endpoint"},
						},
						"test_app_2": {
							RoleName:    "MEMBER",
							Permissions: []types.Permissions{"read:endpoint"},
						},
					},
				},
			},
			{
				name:   "Should fetch another single users load balancer permissions",
				userID: "auth0|ulfric_stormcloak",
				expectedPermissions: &types.LegacyUserPermissions{
					UserID: "auth0|ulfric_stormcloak",
					LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{
						"test_app_2": {
							RoleName: "MEMBER",
							Permissions: []types.Permissions{
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
			expectedLoadBalancers map[string]*types.LoadBalancer
			err                   error
		}{
			{
				name:         "Should fetch all pending load balancers for a single portal user ID",
				userPortalID: "user_10",
				expectedLoadBalancers: map[string]*types.LoadBalancer{
					"test_app_3": expectedLegacyLoadBalancers["test_app_3"],
				},
			},
			{
				name:                  "Should return empty if the portal user ID does not have any pending load balancers in the DB",
				userPortalID:          "user_1",
				expectedLoadBalancers: map[string]*types.LoadBalancer{},
			},
			{
				name:         "Should fail if user portal ID not provided",
				userPortalID: "",
				err:          fmt.Errorf("no user email"),
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
				name:          "Should return the count of load balancer owned by a portal user ID",
				portalUserID:  "user_1",
				expectedCount: 1,
			},
			{
				name:          "Should return the count of load balancer owned by a portal user ID",
				portalUserID:  "user_4",
				expectedCount: 2,
			},
			{
				name:          "Should return 0 if there's no lb owned by this portal user ID",
				portalUserID:  "user_9000",
				expectedCount: 0,
			},
			{
				name:         "Should return missing portal user ID if there's no portal user ID",
				portalUserID: "",
				err:          fmt.Errorf("no user email"),
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

	ts.Run("Test_CreateBlockchain", func() {
		tests := []struct {
			name       string
			blockchain *types.Blockchain
			err        error
		}{
			{
				name:       "Should create a single blockchain in the DB",
				blockchain: createLegacyBlockchain,
			},
			{
				name:       "Should fail if attempting to create a duplicate record",
				blockchain: createLegacyBlockchain,
				err:        fmt.Errorf("Response not OK. 500 Internal Server Error: CreateLegacyBlockchain failed: error chain already exists for chain ID '0127'"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.CreateBlockchain(testCtx, *test.blockchain)
				ts.Equal(test.err, err)

				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					blockchain, err := ts.client1.GetBlockchainByID(testCtx, test.blockchain.ID)
					ts.Equal(test.err, err)
					blockchain.CreatedAt = mockTimestamp
					blockchain.UpdatedAt = mockTimestamp
					ts.Equal(test.blockchain, blockchain)

					blockchain, err = ts.client2.GetBlockchainByID(testCtx, test.blockchain.ID)
					ts.Equal(test.err, err)
					blockchain.CreatedAt = mockTimestamp
					blockchain.UpdatedAt = mockTimestamp
					ts.Equal(test.blockchain, blockchain)
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
					Alias:          "altruist-0002",
					Domain:         "pokt-rpc-2.gateway.pokt.network",
					LoadBalancerID: "test_app_2",
				},
				redirects: []types.Redirect{
					{Alias: "altruist-0001", Domain: "pokt-rpc.gateway.pokt.network", LoadBalancerID: "test_app_1"},
					{Alias: "altruist-0002", Domain: "pokt-rpc-2.gateway.pokt.network", LoadBalancerID: "test_app_2"},
				},
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

	ts.Run("Test_CreateLoadBalancer", func() {
		tests := []struct {
			name                   string
			loadBalancer           *types.LoadBalancer
			expectedCovalentAPIKey string
			err                    error
		}{
			{
				name:                   "Should create a single loadBalancer in the DB",
				loadBalancer:           createLegacyLoadBalancer,
				expectedCovalentAPIKey: "covalent_api_key_1",
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				createdLB, err := ts.client1.CreateLoadBalancer(testCtx, *test.loadBalancer)
				ts.Equal(test.err, err)

				test.loadBalancer.Integrations.CovalentAPIKeyFree = test.expectedCovalentAPIKey
				test.loadBalancer.ID = createdLB.ID
				test.loadBalancer.Applications[0].ID = createdLB.Applications[0].ID
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

	ts.Run("Test_CreateLoadBalancerUser", func() {
		tests := []struct {
			name              string
			loadBalancerID    string
			user              types.UserAccess
			loadBalancerUsers map[string]types.UserAccess
			err               error
		}{
			{
				name:           "Should add a single new user to an existing load balancer in the DB",
				loadBalancerID: "test_app_1",
				user: types.UserAccess{
					RoleName: types.RoleMember,
					Email:    "member_new@test.com",
				},
				loadBalancerUsers: map[string]types.UserAccess{
					"user_1": {UserID: "user_1", RoleName: types.RoleOwner, Email: "james.holden123@test.com", Accepted: true},
					"user_2": {UserID: "user_2", RoleName: types.RoleAdmin, Email: "paul.atreides456@test.com", Accepted: true},
					"user_8": {UserID: "user_8", RoleName: types.RoleAdmin, Email: "rick.deckard456@test.com", Accepted: false},
					// ID dynamically generated on creation
					"": {UserID: "", RoleName: types.RoleMember, Email: "member_new@test.com", Accepted: false},
				},
			},
			{
				name:           "Should add a single existing user to an existing load balancer in the DB",
				loadBalancerID: "test_app_2",
				user: types.UserAccess{
					RoleName: types.RoleMember,
					Email:    "frodo.baggins123@test.com",
				},
				loadBalancerUsers: map[string]types.UserAccess{
					"user_3": {UserID: "user_3", RoleName: types.RoleOwner, Email: "ellen.ripley789@test.com", Accepted: true},
					"user_4": {UserID: "user_4", RoleName: types.RoleMember, Email: "ulfric.stormcloak123@test.com", Accepted: true},
					"user_9": {UserID: "user_9", RoleName: types.RoleMember, Email: "tyrion.lannister789@test.com", Accepted: false},
					"user_2": {UserID: "user_2", RoleName: types.RoleMember, Email: "paul.atreides456@test.com", Accepted: true},
					// ID dynamically generated on creation
					"": {UserID: "", RoleName: types.RoleMember, Email: "frodo.baggins123@test.com", Accepted: false},
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

					<-time.After(50 * time.Millisecond)
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
				err:          fmt.Errorf("Response not OK. 500 Internal Server Error: blockchain not found for chain ID 5440"),
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

	ts.Run("Test_UpdateLoadBalancer", func() {
		tests := []struct {
			name                   string
			loadBalancerID         string
			applicationUpdate      types.UpdateApplication
			applicationAfterUpdate types.Application
			err                    error
		}{
			{
				name:           "Should update a single application in the DB",
				loadBalancerID: "test_app_1",
				applicationUpdate: types.UpdateApplication{
					Name: "test_update_portal_app_123",
					GatewaySettings: &types.UpdateGatewaySettings{
						SecretKey:            "test_90210ac4bdd3423e24877d1ff92",
						SecretKeyRequired:    boolToPointer(false),
						WhitelistOrigins:     []string{"https://portalgun.io", "https://subdomain.example.com", "https://www.example.com"},
						WhitelistBlockchains: []string{"0001", "0002", "003E", "0056"},
						WhitelistUserAgents:  []string{"Brave", "Google Chrome", "Mozilla Firefox", "Netscape Navigator", "Safari"},
						WhitelistContracts: []types.WhitelistContracts{
							{BlockchainID: "0001", Contracts: []string{"0xtest_2f78db6436527729929aaf6c616361de0f7", "0xtest_5fbfe3e9af3971dd833d26ba9b5c936f0be"}},
							{BlockchainID: "0002", Contracts: []string{"0xtest_1111117dc0aa78b770fa6a738034120c302", "0xtest_a39b223fe8d0a0e5c4f27ead9083c756cc2"}},
							{BlockchainID: "003E", Contracts: []string{"0xtest_0a85d5af5bf1d1762f925bdaddc4201f984", "0xtest_f958d2ee523a2206206994597c13d831ec7"}},
							{BlockchainID: "0056", Contracts: []string{"0xtest_00000f279d81a1d3cc75430faa017fa5a2e", "0xtest_5068778dd592e39a122f4f5a5cf09c90fe2"}},
						},
						WhitelistMethods: []types.WhitelistMethods{
							{BlockchainID: "0001", Methods: []string{"GET", "POST", "PUT"}},
							{BlockchainID: "0002", Methods: []string{"DELETE", "GET", "POST", "PUT"}},
							{BlockchainID: "003E", Methods: []string{"GET"}},
							{BlockchainID: "0056", Methods: []string{"GET", "POST"}},
						},
					},
					NotificationSettings: &types.UpdateNotificationSettings{SignedUp: boolToPointer(true), Quarter: boolToPointer(true), Half: boolToPointer(false), ThreeQuarters: boolToPointer(true), Full: boolToPointer(false)},
				},
				applicationAfterUpdate: types.Application{
					Name: "test_update_portal_app_123",
					GatewaySettings: types.GatewaySettings{
						SecretKey:            "test_90210ac4bdd3423e24877d1ff92",
						WhitelistOrigins:     []string{"https://portalgun.io", "https://subdomain.example.com", "https://www.example.com"},
						WhitelistBlockchains: []string{"0001", "0002", "003E", "0056"},
						WhitelistUserAgents:  []string{"Brave", "Google Chrome", "Mozilla Firefox", "Netscape Navigator", "Safari"},
						WhitelistContracts: []types.WhitelistContracts{
							{BlockchainID: "0001", Contracts: []string{"0xtest_2f78db6436527729929aaf6c616361de0f7", "0xtest_5fbfe3e9af3971dd833d26ba9b5c936f0be"}},
							{BlockchainID: "0002", Contracts: []string{"0xtest_1111117dc0aa78b770fa6a738034120c302", "0xtest_a39b223fe8d0a0e5c4f27ead9083c756cc2"}},
							{BlockchainID: "003E", Contracts: []string{"0xtest_0a85d5af5bf1d1762f925bdaddc4201f984", "0xtest_f958d2ee523a2206206994597c13d831ec7"}},
							{BlockchainID: "0056", Contracts: []string{"0xtest_00000f279d81a1d3cc75430faa017fa5a2e", "0xtest_5068778dd592e39a122f4f5a5cf09c90fe2"}},
						},
						WhitelistMethods: []types.WhitelistMethods{
							{BlockchainID: "0001", Methods: []string{"GET", "POST", "PUT"}},
							{BlockchainID: "0002", Methods: []string{"DELETE", "GET", "POST", "PUT"}},
							{BlockchainID: "003E", Methods: []string{"GET"}},
							{BlockchainID: "0056", Methods: []string{"GET", "POST"}},
						},
					},
					NotificationSettings: types.NotificationSettings{SignedUp: true, Quarter: true, Half: false, ThreeQuarters: true, Full: false},
				},
			},
			{
				name:           "Should remove all of a single application's whitelists",
				loadBalancerID: "test_app_1",
				applicationUpdate: types.UpdateApplication{
					GatewaySettings: &types.UpdateGatewaySettings{},
				},
				applicationAfterUpdate: types.Application{
					Name: "test_update_portal_app_123",
					GatewaySettings: types.GatewaySettings{
						SecretKey: "test_90210ac4bdd3423e24877d1ff92",
					},
					NotificationSettings: types.NotificationSettings{SignedUp: true, Quarter: true, Half: false, ThreeQuarters: true, Full: false},
				},
			},
			{
				name:           "Should fail if application cannot be found",
				loadBalancerID: "9000",
				err:            fmt.Errorf("Response not OK. 500 Internal Server Error"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				createdLB, err := ts.client1.UpdateLoadBalancer(testCtx, test.loadBalancerID, test.applicationUpdate)
				ts.Equal(test.err, err)

				if err == nil {
					<-time.After(50 * time.Millisecond)
					applicationID := createdLB.Applications[0].ID

					application, err := ts.client1.GetApplicationByID(testCtx, applicationID)
					ts.NoError(err)
					ts.Equal(test.applicationAfterUpdate.Name, application.Name)
					ts.Equal(test.applicationAfterUpdate.GatewaySettings, application.GatewaySettings)
					ts.Equal(test.applicationAfterUpdate.NotificationSettings, application.NotificationSettings)

					application, err = ts.client2.GetApplicationByID(testCtx, applicationID)
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
			update         types.UpdateFirstDateSurpassed
			protocolAppIDs []string
			expectedDate   time.Time
			err            error
		}{
			{
				name: "Should update the app first date suprassed for the provided slice of app IDs",
				update: types.UpdateFirstDateSurpassed{
					PortalAppIDs:       []string{"test_app_1", "test_app_2"},
					FirstDateSurpassed: time.Date(2022, time.November, 11, 11, 11, 11, 0, time.UTC),
				},
				protocolAppIDs: []string{"test_protocol_app_1", "test_protocol_app_2"},
				expectedDate:   time.Date(2022, time.November, 11, 11, 11, 11, 0, time.UTC),
				err:            nil,
			},
			{
				name: "Should fail if update contains no application IDs cannot be found",
				update: types.UpdateFirstDateSurpassed{
					PortalAppIDs:       []string{},
					FirstDateSurpassed: time.Date(2022, time.November, 11, 11, 11, 11, 0, time.UTC),
				},
				err: fmt.Errorf("Response not OK. 400 Bad Request: no application IDs on input"),
			},
			{
				name: "Should fail if application cannot be found",
				update: types.UpdateFirstDateSurpassed{
					PortalAppIDs:       []string{"9000"},
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
			name          string
			applicationID string
			protocolAppID string
			expectedError error
			err           error
		}{
			{
				name:          "should remove one application by setting its status to deleted",
				applicationID: "test_app_1",
				protocolAppID: "test_protocol_app_1",
				expectedError: fmt.Errorf("Response not OK. 404 Not Found: portal app not found for app ID test_protocol_app_1"),
			},
			{
				name:          "Should fail if application cannot be found",
				applicationID: "2348",
				err:           fmt.Errorf("Response not OK. 500 Internal Server Error"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				_, err := ts.client1.RemoveApplication(testCtx, test.applicationID)
				ts.Equal(test.err, err)

				if test.err == nil {
					<-time.After(50 * time.Millisecond)
					_, err := ts.client1.GetApplicationByID(testCtx, test.protocolAppID)
					ts.Equal(test.expectedError, err)

					_, err = ts.client2.GetApplicationByID(testCtx, test.protocolAppID)
					ts.Equal(test.expectedError, err)
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
					EnforceResult:     "JSUN",
					Path:              "new-path",
					Body:              `{"new-body": "alliance"}`,
					Allowance:         &allowanceInt,
				},
				expectedChain: types.Blockchain{
					ID:                "0001",
					Altruist:          "https://test-update:test-password123@shared-test2.nodes.pokt.network:12345", // pragma: allowlist secret
					Blockchain:        "pokt-mainnet-updated",
					Description:       "POKT Network Mainnet Updated",
					EnforceResult:     "JSUN",
					Ticker:            "SUCH-WOW",
					Path:              "new-path",
					BlockchainAliases: []string{"pokt-mainnet", "another-one"},
					LogLimitBlocks:    100_010,
					Active:            true,
					RequestTimeout:    66_654,
					// Redirects are handled via their own update process
					SyncCheckOptions: types.SyncCheckOptions{
						Body:      `{"new-body": "alliance"}`,
						ResultKey: "updated-key",
						Allowance: 123,
					},
				},
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

	ts.Run("Test_UpdateLoadBalancerUserRole", func() {
		tests := []struct {
			name              string
			loadBalancerID    string
			providerUserID    string
			update            types.UpdateUserAccess
			loadBalancerUsers []types.UserAccess
			userLBPermissions types.LoadBalancerPermissions
			err               error
		}{
			{
				name:           "Should update a single user TO ADMIN for an existing load balancer in the DB",
				loadBalancerID: "test_app_2",
				providerUserID: "auth0|ulfric_stormcloak",
				update: types.UpdateUserAccess{
					UserID:   "user_4",
					RoleName: types.RoleAdmin,
				},
				loadBalancerUsers: []types.UserAccess{
					{UserID: "user_3", RoleName: "OWNER", Email: "ellen.ripley789@test.com", Accepted: true},
					{UserID: "user_2", RoleName: "MEMBER", Email: "paul.atreides456@test.com", Accepted: true},
					{UserID: "user_4", RoleName: "MEMBER", Email: "ulfric.stormcloak123@test.com", Accepted: true},
					{UserID: "user_7", RoleName: "MEMBER", Email: "frodo.baggins123@test.com", Accepted: false},
					{UserID: "user_9", RoleName: "MEMBER", Email: "tyrion.lannister789@test.com", Accepted: false},
				},
				userLBPermissions: types.LoadBalancerPermissions{
					RoleName: "MEMBER", Permissions: []types.Permissions{"read:endpoint"},
				},
			},
			// {
			// 	name:           "Should update a single user to MEMBER for an existing load balancer in the DB",
			// 	loadBalancerID: "test_lb_34987u329rfn23f",
			// 	update: types.UpdateUserAccess{
			// 		Email:    "member1@test.com",
			// 		RoleName: types.RoleMember,
			// 	},
			// 	loadBalancerUsers: []types.UserAccess{
			// 		{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
			// 		{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
			// 		{RoleName: types.RoleMember, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
			// 		{RoleName: types.RoleMember, UserID: "", Email: "member_new@test.com", Accepted: false},
			// 	},
			// 	userLBPermissions: types.LoadBalancerPermissions{
			// 		RoleName:    types.RoleMember,
			// 		Permissions: []types.PermissionsEnum{types.ReadEndpoint},
			// 	},
			// },
			// {
			// 	name:           "Should transfer OWNERship for an existing load balancer in the DB",
			// 	loadBalancerID: "test_lb_34987u329rfn23f",
			// 	userID:         "test_user_admin1234",
			// 	update: types.UpdateUserAccess{
			// 		Email:        "admin1@test.com",
			// 		RoleName:     types.RoleOwner,
			// 		UpdaterEmail: "owner1@test.com",
			// 	},
			// 	loadBalancerUsers: []types.UserAccess{
			// 		{RoleName: types.RoleAdmin, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
			// 		{RoleName: types.RoleOwner, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
			// 		{RoleName: types.RoleMember, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
			// 		{RoleName: types.RoleMember, UserID: "", Email: "member_new@test.com", Accepted: false},
			// 	},
			// 	userLBPermissions: types.LoadBalancerPermissions{
			// 		RoleName:    types.RoleOwner,
			// 		Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint, types.DeleteEndpoint, types.TransferEndpoint},
			// 	},
			// },
			// {
			// 	name:           "Should transfer OWNERship again for an existing load balancer in the DB",
			// 	loadBalancerID: "test_lb_34987u329rfn23f",
			// 	userID:         "test_user_member1234",
			// 	update: types.UpdateUserAccess{
			// 		Email:        "member1@test.com",
			// 		RoleName:     types.RoleOwner,
			// 		UpdaterEmail: "admin1@test.com",
			// 	},
			// 	loadBalancerUsers: []types.UserAccess{
			// 		{RoleName: types.RoleAdmin, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
			// 		{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
			// 		{RoleName: types.RoleOwner, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
			// 		{RoleName: types.RoleMember, UserID: "", Email: "member_new@test.com", Accepted: false},
			// 	},
			// 	userLBPermissions: types.LoadBalancerPermissions{
			// 		RoleName:    types.RoleOwner,
			// 		Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint, types.DeleteEndpoint, types.TransferEndpoint},
			// 	},
			// },
			// {
			// 	name:           "Should transfer OWNERship back to original owner for an existing load balancer in the DB",
			// 	loadBalancerID: "test_lb_34987u329rfn23f",
			// 	userID:         "test_user_1dbffbdfeeb225",
			// 	update: types.UpdateUserAccess{
			// 		Email:        "owner1@test.com",
			// 		RoleName:     types.RoleOwner,
			// 		UpdaterEmail: "member1@test.com",
			// 	},
			// 	loadBalancerUsers: []types.UserAccess{
			// 		{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
			// 		{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
			// 		{RoleName: types.RoleAdmin, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
			// 		{RoleName: types.RoleMember, UserID: "", Email: "member_new@test.com", Accepted: false},
			// 	},
			// 	userLBPermissions: types.LoadBalancerPermissions{
			// 		RoleName:    types.RoleOwner,
			// 		Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint, types.DeleteEndpoint, types.TransferEndpoint},
			// 	},
			// },
			// {
			// 	name:           "Should update a single unaccepted user to ADMIN for an existing load balancer in the DB",
			// 	loadBalancerID: "test_lb_34987u329rfn23f",
			// 	userID:         "",
			// 	update: types.UpdateUserAccess{
			// 		Email:    "member_new@test.com",
			// 		RoleName: types.RoleAdmin,
			// 	},
			// 	loadBalancerUsers: []types.UserAccess{
			// 		{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
			// 		{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
			// 		{RoleName: types.RoleAdmin, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
			// 		{RoleName: types.RoleAdmin, UserID: "", Email: "member_new@test.com", Accepted: false},
			// 	},
			// 	userLBPermissions: types.LoadBalancerPermissions{
			// 		RoleName:    types.RoleAdmin,
			// 		Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint},
			// 	},
			// },
			// {
			// 	name:           "Should update a single unaccepted user back to MEMBER for an existing load balancer in the DB",
			// 	loadBalancerID: "test_lb_34987u329rfn23f",
			// 	userID:         "",
			// 	update: types.UpdateUserAccess{
			// 		Email:    "member_new@test.com",
			// 		RoleName: types.RoleMember,
			// 	},
			// 	loadBalancerUsers: []types.UserAccess{
			// 		{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
			// 		{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
			// 		{RoleName: types.RoleAdmin, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
			// 		{RoleName: types.RoleMember, UserID: "", Email: "member_new@test.com", Accepted: false},
			// 	},
			// 	userLBPermissions: types.LoadBalancerPermissions{
			// 		RoleName:    types.RoleMember,
			// 		Permissions: []types.PermissionsEnum{types.ReadEndpoint},
			// 	},
			// },
			// {
			// 	name:           "Should update a single user back to member for an existing load balancer in the DB",
			// 	loadBalancerID: "test_lb_34987u329rfn23f",
			// 	userID:         "test_user_member1234",
			// 	update: types.UpdateUserAccess{
			// 		Email:    "member1@test.com",
			// 		RoleName: types.RoleMember,
			// 	},
			// 	loadBalancerUsers: []types.UserAccess{
			// 		{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
			// 		{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
			// 		{RoleName: types.RoleMember, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
			// 		{RoleName: types.RoleMember, UserID: "", Email: "member_new@test.com", Accepted: false},
			// 	},
			// 	userLBPermissions: types.LoadBalancerPermissions{
			// 		RoleName:    types.RoleMember,
			// 		Permissions: []types.PermissionsEnum{types.ReadEndpoint},
			// 	},
			// },
			// {
			// 	name:           "Should fail if attempting to transfer ownership and the UpdaterEmail is not provided",
			// 	loadBalancerID: "test_lb_34987u329rfn23f",
			// 	update: types.UpdateUserAccess{
			// 		Email:        "member1@test.com",
			// 		RoleName:     types.RoleOwner,
			// 		UpdaterEmail: "",
			// 	},
			// 	err: fmt.Errorf("transferring ownership requires providing the updater's email"),
			// },
			// {
			// 	name:           "Should fail if attempting to transfer ownership and the user has not accepted their invite",
			// 	loadBalancerID: "test_lb_34987u329rfn23f",
			// 	update: types.UpdateUserAccess{
			// 		Email:        "member_new@test.com",
			// 		RoleName:     types.RoleOwner,
			// 		UpdaterEmail: "owner1@test.com",
			// 	},
			// 	err: fmt.Errorf("Response not OK. 500 Internal Server Error: error: cannot set a user to owner if they have not yet accepted their invitation"),
			// },
			// {
			// 	name:           "Should fail if load balancer ID not provided",
			// 	loadBalancerID: "",
			// 	err:            fmt.Errorf("no load balancer ID"),
			// },
			// {
			// 	name:           "Should fail if user email not provided",
			// 	loadBalancerID: "test_lb_34987u329rfn23f",
			// 	update: types.UpdateUserAccess{
			// 		Email:    "",
			// 		RoleName: types.RoleAdmin,
			// 	},
			// 	err: fmt.Errorf("no user email"),
			// },
			// {
			// 	name:           "Should fail if invalid role name provided",
			// 	loadBalancerID: "test_lb_34987u329rfn23f",
			// 	update: types.UpdateUserAccess{
			// 		Email:    "member1@test.com",
			// 		RoleName: types.RoleName("wrong_one"),
			// 	},
			// 	err: fmt.Errorf("invalid role name"),
			// },
			// {
			// 	name:           "Should fail if load balancer cannot be found",
			// 	loadBalancerID: "im_not_here",
			// 	update: types.UpdateUserAccess{
			// 		Email:    "member1@test.com",
			// 		RoleName: types.RoleMember,
			// 	},
			// 	err: fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
			// },
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

					if test.providerUserID != "" {
						permissionsByUserID, err := ts.client1.GetUserPermissionsByUserID(testCtx, types.UserID(test.providerUserID))
						ts.Equal(test.err, err)
						ts.Equal(test.userLBPermissions, permissionsByUserID.LoadBalancers[types.LoadBalancerID(test.loadBalancerID)])

						permissionsByUserID, err = ts.client2.GetUserPermissionsByUserID(testCtx, types.UserID(test.providerUserID))
						ts.Equal(test.err, err)
						ts.Equal(test.userLBPermissions, permissionsByUserID.LoadBalancers[types.LoadBalancerID(test.loadBalancerID)])
					}
				}
			})
		}
	})

	// 	ts.Run("Test_AcceptLoadBalancerUser", func() {
	// 		tests := []struct {
	// 			name                          string
	// 			email, loadBalancerID, userID string
	// 			loadBalancerUsers             []types.UserAccess
	// 			userLBPermissions             types.LoadBalancerPermissions
	// 			err                           error
	// 			pendingResult                 error
	// 		}{
	// 			{
	// 				name:           "Should update a single user's ID and Accepted field for an existing load balancer in the DB",
	// 				email:          "member_new@test.com",
	// 				loadBalancerID: "test_lb_34987u329rfn23f",
	// 				userID:         "test_user_accept_member",
	// 				loadBalancerUsers: []types.UserAccess{
	// 					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
	// 					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
	// 					{RoleName: types.RoleMember, UserID: "test_user_member1234", Email: "member1@test.com", Accepted: true},
	// 					{RoleName: types.RoleMember, UserID: "test_user_accept_member", Email: "member_new@test.com", Accepted: true},
	// 				},
	// 				userLBPermissions: types.LoadBalancerPermissions{
	// 					RoleName:    types.RoleMember,
	// 					Permissions: []types.PermissionsEnum{types.ReadEndpoint},
	// 				},
	// 				pendingResult: fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
	// 			},
	// 			{
	// 				name:  "Should fail if email not provided",
	// 				email: "",
	// 				err:   fmt.Errorf("no user email"),
	// 			},
	// 			{
	// 				name:           "Should fail if load balancer ID not provided",
	// 				email:          "member_new@test.com",
	// 				loadBalancerID: "",
	// 				err:            fmt.Errorf("no load balancer ID"),
	// 			},
	// 			{
	// 				name:           "Should fail if user ID not provided",
	// 				email:          "member_new@test.com",
	// 				loadBalancerID: "test_lb_34987u329rfn23f",
	// 				userID:         "",
	// 				err:            fmt.Errorf("no user ID"),
	// 			},
	// 			{
	// 				name:           "Should fail if load balancer cannot be found",
	// 				email:          "member_new@test.com",
	// 				loadBalancerID: "im_not_here",
	// 				userID:         "test_user_accept_member",
	// 				err:            fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
	// 			},
	// 		}

	// 		for _, test := range tests {
	// 			ts.Run(test.name, func() {
	// 				_, err := ts.client1.AcceptLoadBalancerUser(testCtx, test.email, test.loadBalancerID, test.userID)
	// 				ts.Equal(test.err, err)
	// 				_, err = ts.client2.AcceptLoadBalancerUser(testCtx, test.email, test.loadBalancerID, test.userID)
	// 				ts.Equal(test.err, err)
	// 				if test.err == nil {
	// 					<-time.After(50 * time.Millisecond)
	// 					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
	// 					ts.Equal(test.err, err)
	// 					ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

	// 					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
	// 					ts.Equal(test.err, err)
	// 					ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

	// 					if test.userID != "" {
	// 						// user should have been added to the user permissions in the cache
	// 						permissionsByUserID, err := ts.client1.GetUserPermissionsByUserID(testCtx, types.UserID(test.userID))
	// 						ts.Equal(test.err, err)
	// 						ts.Equal(test.userLBPermissions, permissionsByUserID.LoadBalancers[types.LoadBalancerID(test.loadBalancerID)])

	// 						permissionsByUserID, err = ts.client2.GetUserPermissionsByUserID(testCtx, types.UserID(test.userID))
	// 						ts.Equal(test.err, err)
	// 						ts.Equal(test.userLBPermissions, permissionsByUserID.LoadBalancers[types.LoadBalancerID(test.loadBalancerID)])

	// 						// user's LB should have been added to the user's LBs map in the cache
	// 						roleMember := types.RoleMember

	// 						exists := false
	// 						loadBalancers, err := ts.client1.GetLoadBalancersByUserID(testCtx, test.userID, &roleMember)
	// 						ts.Equal(test.err, err)
	// 						for _, lb := range loadBalancers {
	// 							if lb.ID == test.loadBalancerID {
	// 								exists = true
	// 								break
	// 							}
	// 						}
	// 						ts.True(exists)

	// 						exists = false
	// 						loadBalancers, err = ts.client2.GetLoadBalancersByUserID(testCtx, test.userID, &roleMember)
	// 						ts.Equal(test.err, err)
	// 						for _, lb := range loadBalancers {
	// 							if lb.ID == test.loadBalancerID {
	// 								exists = true
	// 								break
	// 							}
	// 						}
	// 						ts.True(exists)

	// 						// user should be removed from pending lb list
	// 						removed := true
	// 						pendingLoadBalancers, _ := ts.client1.GetPendingLoadBalancersByUserID(testCtx, test.email)
	// 						for _, lb := range pendingLoadBalancers {
	// 							if lb.ID == test.loadBalancerID {
	// 								removed = false
	// 								break
	// 							}
	// 						}
	// 						ts.True(removed)

	// 						removed = true
	// 						pendingLoadBalancers, _ = ts.client2.GetPendingLoadBalancersByUserID(testCtx, test.email)
	// 						for _, lb := range pendingLoadBalancers {
	// 							if lb.ID == test.loadBalancerID {
	// 								removed = false
	// 								break
	// 							}
	// 						}
	// 						ts.True(removed)
	// 					}
	// 				}
	// 			})
	// 		}
	// 	})

	// 	ts.Run("Test_RemoveLoadBalancer", func() {
	// 		tests := []struct {
	// 			name           string
	// 			loadBalancerID string
	// 			expectedUserID string
	// 			err            error
	// 		}{
	// 			{
	// 				name:           "should remove one load balancer by setting its user ID to an empty string",
	// 				loadBalancerID: "test_lb_3890ru23jfi32fj",
	// 				expectedUserID: "",
	// 			},
	// 			{
	// 				name:           "Should fail if load balancer cannot be found",
	// 				loadBalancerID: "9000",
	// 				err:            fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
	// 			},
	// 		}

	// 		for _, test := range tests {
	// 			ts.Run(test.name, func() {
	// 				_, err := ts.client1.RemoveLoadBalancer(testCtx, test.loadBalancerID)
	// 				ts.Equal(test.err, err)
	// 				if test.err == nil {
	// 					<-time.After(50 * time.Millisecond)
	// 					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
	// 					ts.NoError(err)
	// 					ts.Equal(test.expectedUserID, loadBalancer.UserID)

	// 					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
	// 					ts.NoError(err)
	// 					ts.Equal(test.expectedUserID, loadBalancer.UserID)
	// 				}
	// 			})
	// 		}
	// 	})

	// 	ts.Run("Test_DeleteLoadBalancerUser", func() {
	// 		tests := []struct {
	// 			name                       string
	// 			loadBalancerID, email      string
	// 			userID                     string
	// 			loadBalancerUsers          []types.UserAccess
	// 			testPendingUsers           []types.UserAccess
	// 			err, permissionsFetchError error
	// 		}{
	// 			{
	// 				name:           "Should delete a single user for an existing load balancer in the DB",
	// 				loadBalancerID: "test_lb_34987u329rfn23f",
	// 				email:          "member1@test.com",
	// 				userID:         "test_user_member1234",
	// 				loadBalancerUsers: []types.UserAccess{
	// 					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
	// 					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
	// 					{RoleName: types.RoleMember, UserID: "test_user_accept_member", Email: "member_new@test.com", Accepted: true},
	// 				},
	// 				permissionsFetchError: errors.New("Response not OK. 404 Not Found: no permissions found for given user ID"),
	// 			},
	// 			{
	// 				name:           "Should delete a single pending user for an existing load balancer in the DB",
	// 				loadBalancerID: "test_lb_34987u329rfn23f",
	// 				testPendingUsers: []types.UserAccess{
	// 					{Email: "owner2@test.com", RoleName: types.RoleMember},
	// 				},
	// 				email:  "owner2@test.com",
	// 				userID: "test_user_04228205bd261a",
	// 				loadBalancerUsers: []types.UserAccess{
	// 					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
	// 					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
	// 					{RoleName: types.RoleMember, UserID: "test_user_accept_member", Email: "member_new@test.com", Accepted: true},
	// 				},
	// 				permissionsFetchError: nil, // this test user has other permissions for other LBs so will still return permissions
	// 			},
	// 			{
	// 				name:           "Should delete only one pending user even if multiple pending users exist for an existing load balancer in the DB",
	// 				loadBalancerID: "test_lb_34987u329rfn23f",
	// 				testPendingUsers: []types.UserAccess{
	// 					{Email: "owner2@test.com", RoleName: types.RoleMember},
	// 					{Email: "test-delete-user-2@test.com", RoleName: types.RoleAdmin},
	// 					{Email: "test-delete-user-3@test.com", RoleName: types.RoleMember},
	// 				},
	// 				email:  "owner2@test.com",
	// 				userID: "test_user_04228205bd261a",
	// 				loadBalancerUsers: []types.UserAccess{
	// 					{RoleName: types.RoleOwner, UserID: "test_user_1dbffbdfeeb225", Email: "owner1@test.com", Accepted: true},
	// 					{RoleName: types.RoleAdmin, UserID: "test_user_admin1234", Email: "admin1@test.com", Accepted: true},
	// 					{RoleName: types.RoleMember, UserID: "test_user_accept_member", Email: "member_new@test.com", Accepted: true},
	// 					{RoleName: types.RoleAdmin, UserID: "", Email: "test-delete-user-2@test.com", Accepted: false},
	// 					{RoleName: types.RoleMember, UserID: "", Email: "test-delete-user-3@test.com", Accepted: false},
	// 				},
	// 				permissionsFetchError: nil, // this test user has other permissions for other LBs so will still return permissions
	// 			},
	// 			{
	// 				name:           "Should fail if load balancer cannot be found",
	// 				loadBalancerID: "why_am_i_not_a_load_balancer",
	// 				email:          "member1@test.com",
	// 				userID:         "test_user_member1234",
	// 				err:            fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"),
	// 			},
	// 			{
	// 				name:           "Should fail if attempting to delete the owner of a load balancer",
	// 				loadBalancerID: "test_lb_34987u329rfn23f",
	// 				email:          "owner1@test.com",
	// 				userID:         "test_user_1dbffbdfeeb225",
	// 				err:            fmt.Errorf("Response not OK. 500 Internal Server Error: cannot delete the owner of a load balancer"),
	// 			},
	// 			{
	// 				name:           "Should fail if load balancer ID not provided",
	// 				loadBalancerID: "",
	// 				email:          "member1@test.com",
	// 				err:            fmt.Errorf("no load balancer ID"),
	// 			},
	// 			{
	// 				name:           "Should fail if email not provided",
	// 				loadBalancerID: "test_lb_34987u329rfn23f",
	// 				email:          "",
	// 				err:            fmt.Errorf("no user email"),
	// 			},
	// 		}

	// 		for _, test := range tests {
	// 			ts.Run(test.name, func() {
	// 				if len(test.testPendingUsers) != 0 {
	// 					var updatedLB *types.LoadBalancer
	// 					for _, user := range test.testPendingUsers {
	// 						lb, err := ts.client1.CreateLoadBalancerUser(testCtx, test.loadBalancerID, user)
	// 						ts.NoError(err)
	// 						updatedLB = lb
	// 					}
	// 					ts.Len(updatedLB.Users, 3+len(test.testPendingUsers))
	// 				}

	// 				_, err := ts.client1.DeleteLoadBalancerUser(testCtx, test.loadBalancerID, test.email)
	// 				ts.Equal(test.err, err)
	// 				if test.err == nil {
	// 					<-time.After(50 * time.Millisecond)
	// 					loadBalancer, err := ts.client1.GetLoadBalancerByID(testCtx, test.loadBalancerID)
	// 					ts.Equal(test.err, err)
	// 					ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

	// 					loadBalancer, err = ts.client2.GetLoadBalancerByID(testCtx, test.loadBalancerID)
	// 					ts.Equal(test.err, err)
	// 					ts.Equal(test.loadBalancerUsers, loadBalancer.Users)

	// 					// user should have been removed from the user permissions in the cache
	// 					_, err = ts.client1.GetUserPermissionsByUserID(testCtx, types.UserID(test.userID))
	// 					ts.Equal(test.permissionsFetchError, err)

	// 					_, err = ts.client2.GetUserPermissionsByUserID(testCtx, types.UserID(test.userID))
	// 					ts.Equal(test.permissionsFetchError, err)

	// 					// user's LB should have been removed from the user's LBs map in the cache
	// 					_, err = ts.client1.GetLoadBalancersByUserID(testCtx, test.userID, nil)
	// 					ts.Equal(fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"), err)

	//					_, err = ts.client2.GetLoadBalancersByUserID(testCtx, test.userID, nil)
	//					ts.Equal(fmt.Errorf("Response not OK. 404 Not Found: load balancer not found"), err)
	//				}
	//			})
	//		}
	// })
}

func blockchainsToMap(blockchains []*types.Blockchain) map[string]*types.Blockchain {
	expectedMap := make(map[string]*types.Blockchain)
	for _, b := range blockchains {
		expectedMap[b.ID] = b
	}
	return expectedMap
}

func applicationsToMap(applications []*types.Application) map[string]*types.Application {
	expectedMap := make(map[string]*types.Application)
	for _, a := range applications {
		expectedMap[a.ID] = a
	}
	return expectedMap
}

func loadBalancersToMap(loadBalancers []*types.LoadBalancer) map[string]*types.LoadBalancer {
	lbMap := make(map[string]*types.LoadBalancer)
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

func sortUsersByRole(users []types.UserAccess) {
	roleWeight := map[types.RoleName]int{types.RoleOwner: 0, types.RoleAdmin: 1, types.RoleMember: 2}

	sort.Slice(users, func(i, j int) bool {
		if roleWeight[users[i].RoleName] != roleWeight[users[j].RoleName] {
			return roleWeight[users[i].RoleName] < roleWeight[users[j].RoleName]
		}
		return users[i].UserID < users[j].UserID
	})
}

func payPlansToMap(payPlans []*types.PayPlan) map[string]*types.PayPlan {
	payPlansMap := make(map[string]*types.PayPlan)
	for _, payPlan := range payPlans {
		payPlansMap[string(payPlan.Type)] = payPlan
	}
	return payPlansMap
}

func userAccessSliceToMap(users []types.UserAccess) map[string]types.UserAccess {
	userMap := make(map[string]types.UserAccess)
	for _, user := range users {
		userMap[user.UserID] = user
	}
	return userMap
}

func clearTimeFields(lb *types.LoadBalancer) {
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

	expectedLegacyApplications = map[string]*types.Application{
		"test_protocol_app_1": {
			ID:                 "test_protocol_app_1",
			UserID:             "auth0|james_holden",
			Name:               "pokt_app_123",
			FirstDateSurpassed: mockTimestamp,
			GatewayAAT: types.GatewayAAT{
				Address:              "test_34715cae753e67c75fbb340442e7de8e",
				ApplicationPublicKey: "test_34715cae753e67c75fbb340442e7de8e",
				ApplicationSignature: "test_1dc39a2e5a84a35bf030969a0b3231f7",
				ClientPublicKey:      "test_89a3af6a587aec02cfade6f5000424c2",
				PrivateKey:           "test_11b8d394ca331d7c7a71ca1896d630f6",
				Version:              "0.0.1",
			},
			GatewaySettings: types.GatewaySettings{
				SecretKey:           "test_40f482d91a5ef2300ebb4e2308c",
				SecretKeyRequired:   true,
				WhitelistOrigins:    []string{"https://test.com"},
				WhitelistUserAgents: []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"},
				WhitelistContracts: []types.WhitelistContracts{
					{BlockchainID: "0001", Contracts: []string{"0x1234567890abcdef"}},
				},
				WhitelistMethods: []types.WhitelistMethods{
					{BlockchainID: "0001", Methods: []string{"GET"}},
				},
				WhitelistBlockchains: []string{"0053"},
			},
			Limit: types.AppLimit{
				PayPlan: types.PayPlan{
					Type:  "basic_plan",
					Limit: 1000,
				},
				CustomLimit: 0,
			},
			NotificationSettings: types.NotificationSettings{
				Full:          true,
				Half:          false,
				Quarter:       true,
				SignedUp:      false,
				ThreeQuarters: true,
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
		"test_protocol_app_2": {
			ID:                 "test_protocol_app_2",
			UserID:             "auth0|ellen_ripley",
			Name:               "pokt_app_456",
			FirstDateSurpassed: mockTimestamp,
			GatewayAAT: types.GatewayAAT{
				Address:              "test_8237c72345f12d1b1a8b64a1a7f66fa4",
				ApplicationPublicKey: "test_8237c72345f12d1b1a8b64a1a7f66fa4",
				ApplicationSignature: "test_f48d33b30ddaf60a1e5bb50d2ba8da5a",
				ClientPublicKey:      "test_04c71d90a92f40416b6f1d7d8af17e02",
				PrivateKey:           "test_2e83c836a29b423a47d8e18c779fd422",
				Version:              "0.0.1",
			},
			GatewaySettings: types.GatewaySettings{
				SecretKey:           "test_9c9e3b193cfba5348f93bb2f3e3fb794",
				SecretKeyRequired:   false,
				WhitelistOrigins:    []string{"https://example.com"},
				WhitelistUserAgents: []string{"Mozilla/5.0 (Linux; Android 10; SM-A205U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36"},
				WhitelistContracts: []types.WhitelistContracts{
					{BlockchainID: "0064", Contracts: []string{"0x0987654321abcdef"}},
				},
				WhitelistMethods: []types.WhitelistMethods{
					{BlockchainID: "0064", Methods: []string{"POST"}},
				},
				WhitelistBlockchains: []string{"0021"},
			},
			Limit: types.AppLimit{
				PayPlan: types.PayPlan{
					Type:  "pro_plan",
					Limit: 5000,
				},
				CustomLimit: 0,
			},
			NotificationSettings: types.NotificationSettings{
				Half: true,
				Full: true,
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
		"test_protocol_app_3": {
			ID:                 "test_protocol_app_3",
			UserID:             "auth0|chrisjen_avasarala",
			Name:               "pokt_app_789",
			FirstDateSurpassed: mockTimestamp,
			GatewayAAT: types.GatewayAAT{
				Address:              "test_b5e07928fc80083c13ad0201b81bae9b",
				ApplicationPublicKey: "test_f608500e4fe3e09014fe2411b4a560b5",
				ApplicationSignature: "test_c3cd8be16ba32e24dd49fdb0247fc9b8",
				ClientPublicKey:      "test_328a9cf1b35085eeaa669aa858f6fba9",
				PrivateKey:           "test_8663e187c19f3c6e27317eab4ed6d7d5",
				Version:              "0.0.1",
			},
			GatewaySettings: types.GatewaySettings{SecretKey: "test_9f48b13e2bc5fd31ab367841f11495c1"},
			Limit: types.AppLimit{
				PayPlan:     types.PayPlan{Type: "startup_plan", Limit: 500},
				CustomLimit: 0,
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
		"test_protocol_app_4": {
			ID:                 "test_protocol_app_4",
			UserID:             "auth0|chrisjen_avasarala",
			Name:               "pokt_app_789",
			FirstDateSurpassed: mockTimestamp,
			GatewayAAT: types.GatewayAAT{
				Address:              "test_eb2e5bcba557cfe8fa76fd7fff54f9d1",
				ApplicationPublicKey: "test_f6a5d8690ecb669865bd752b7796a920",
				ApplicationSignature: "test_cf05cf9bb26111c548e88fb6157af708",
				ClientPublicKey:      "test_6ee5ea553408f0895923fd1569dc5072",
				PrivateKey:           "test_838d29d61a65401f7d56d084cb6e4783",
				Version:              "0.0.1",
			},
			GatewaySettings: types.GatewaySettings{SecretKey: "test_9f48b13e2bc5fd31ab367841f11495c1"},
			Limit: types.AppLimit{
				PayPlan:     types.PayPlan{Type: "startup_plan", Limit: 500},
				CustomLimit: 0,
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
	}

	expectedLegacyLoadBalancers = map[string]*types.LoadBalancer{
		"test_app_1": {
			ID:                "test_app_1",
			Name:              "pokt_app_123",
			UserID:            "auth0|james_holden",
			RequestTimeout:    5000,
			Gigastake:         true,
			GigastakeRedirect: true,
			StickyOptions: types.StickyOptions{
				Duration: "60",
				StickyOrigins: []string{
					"chrome-extension://",
					"moz-extension://",
				},
				StickyMax:  300,
				Stickiness: true,
			},
			Applications: []*types.Application{
				expectedLegacyApplications["test_protocol_app_1"],
			},
			Users: []types.UserAccess{
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
			Integrations: types.AccountIntegrations{
				CovalentAPIKeyFree: "covalent_api_key_1",
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
		"test_app_2": {
			ID:             "test_app_2",
			Name:           "pokt_app_456",
			UserID:         "auth0|ellen_ripley",
			RequestTimeout: 10000,
			Applications: []*types.Application{
				expectedLegacyApplications["test_protocol_app_2"],
			},
			StickyOptions: types.StickyOptions{
				Duration: "30",
				StickyOrigins: []string{
					"https://example.com",
					"https://test.com",
				},
				StickyMax:  600,
				Stickiness: true,
			},
			Users: []types.UserAccess{
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
			Integrations: types.AccountIntegrations{
				CovalentAPIKeyFree: "covalent_api_key_2",
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
		"test_app_3": {
			ID:             "test_app_3",
			Name:           "pokt_app_789",
			UserID:         "auth0|chrisjen_avasarala",
			RequestTimeout: 10000,
			Applications: []*types.Application{
				expectedLegacyApplications["test_protocol_app_3"],
				expectedLegacyApplications["test_protocol_app_4"],
			},
			Users: []types.UserAccess{
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
			Integrations: types.AccountIntegrations{
				CovalentAPIKeyFree: "covalent_api_key_3",
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
		"test_app_4": {
			ID:                "test_app_4",
			UserID:            "auth0|ulfric_stormcloak",
			Name:              "test_account_4",
			RequestTimeout:    0,
			Applications:      []*types.Application(nil),
			Gigastake:         false,
			GigastakeRedirect: false,
			StickyOptions: types.StickyOptions{
				Stickiness: false,
			},
			Users: []types.UserAccess{
				{
					UserID:   "user_4",
					RoleName: "OWNER",
					Email:    "ulfric.stormcloak123@test.com",
					Accepted: true,
				},
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
		"test_app_5": {
			ID:     "test_app_5",
			Name:   "test_account_5",
			UserID: "auth0|ulfric_stormcloak",
			Users: []types.UserAccess{
				{
					UserID:   "user_4",
					RoleName: "OWNER",
					Email:    "ulfric.stormcloak123@test.com",
					Accepted: true,
				},
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
	}

	expectedLegacyPayPlans = map[string]*types.PayPlan{
		"developer_plan":  {Type: "developer_plan", Limit: 100},
		"startup_plan":    {Type: "startup_plan", Limit: 500},
		"basic_plan":      {Type: "basic_plan", Limit: 1000},
		"pro_plan":        {Type: "pro_plan", Limit: 5000},
		"enterprise_plan": {Type: "enterprise_plan", Limit: 10000},
	}

	expectedLegacyBlockchains = map[string]*types.Blockchain{
		"0001": {
			ID:                "0001",
			Altruist:          "https://test_pocket:auth123456@altruist-0001.com:1234", // pragma: allowlist secret
			Blockchain:        "mainnet",
			Description:       "Pocket Network Mainnet",
			EnforceResult:     "JSON",
			Path:              "/v1/query/height",
			Ticker:            "POKT",
			BlockchainAliases: []string{"mainnet"},
			Active:            true,
			Redirects: []types.Redirect{
				{
					Alias:          "altruist-0001",
					Domain:         "pokt-rpc.gateway.pokt.network",
					LoadBalancerID: "test_app_1",
					CreatedAt:      time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:      time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			SyncCheckOptions: types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"query\"}",
				ResultKey: "result.sync_info",
				Allowance: 1,
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
			Redirects: []types.Redirect{
				{
					Alias:          "altruist-0021",
					Domain:         "eth-rpc.gateway.pokt.network",
					LoadBalancerID: "test_app_3",
					CreatedAt:      time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:      time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			SyncCheckOptions: types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[]}",
				ResultKey: "result",
				Allowance: 5,
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
			Redirects: []types.Redirect{
				{
					Alias:          "altruist-0040",
					Domain:         "hmy-rpc.gateway.pokt.network",
					LoadBalancerID: "test_app_3",
					CreatedAt:      time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:      time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			SyncCheckOptions: types.SyncCheckOptions{
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
			Redirects: []types.Redirect{
				{
					Alias:          "altruist-0053",
					Domain:         "op-rpc.gateway.pokt.network",
					LoadBalancerID: "test_app_2",
					CreatedAt:      time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:      time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			SyncCheckOptions: types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[]}",
				ResultKey: "result",
				Allowance: 2,
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
		"0064": {
			ID:                "0064",
			Altruist:          "https://test_pocket:auth123456@altruist-0064.com:1234", // pragma: allowlist secret
			Blockchain:        "sui-testnet",
			Description:       "Sui Testnet",
			EnforceResult:     "JSON",
			Ticker:            "SUI-TESTNET",
			BlockchainAliases: []string{"sui-testnet"},
			LogLimitBlocks:    100000,
			RequestTimeout:    60000,
			Active:            false,
			SyncCheckOptions: types.SyncCheckOptions{
				Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"sui_blockNumber\",\"params\":[]}",
				ResultKey: "result",
				Allowance: 7,
			},
			CreatedAt: mockTimestamp,
			UpdatedAt: mockTimestamp,
		},
	}

	createLegacyBlockchain = &types.Blockchain{
		ID:                "0127",
		Altruist:          "https://test_pocket:auth123456@altruist-0127.com:1234", // pragma: allowlist secret
		Blockchain:        "bsc-mainnet",
		Description:       "Binance Smart Chain Mainnet",
		EnforceResult:     "JSON",
		Ticker:            "BNB",
		BlockchainAliases: []string{"bsc-mainnet", "binance-smart-chain-mainnet"},
		LogLimitBlocks:    200000,
		Redirects: []types.Redirect{
			{
				Alias:          "altruist-0127",
				Domain:         "bnbrpc.gateway.pokt.network",
				LoadBalancerID: "test_app_3",
			},
		},
		SyncCheckOptions: types.SyncCheckOptions{
			Body:      "{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"eth_syncing\",\"params\":[]}",
			ResultKey: "result",
			Allowance: 3,
		},
		CreatedAt: mockTimestamp,
		UpdatedAt: mockTimestamp,
	}

	createLegacyLoadBalancer = &types.LoadBalancer{
		Name:              "new_pokt_app_777",
		UserID:            "auth0|james_holden",
		ApplicationIDs:    []string(nil),
		RequestTimeout:    5000,
		Gigastake:         true,
		GigastakeRedirect: true,
		StickyOptions:     types.StickyOptions{Duration: "60", StickyOrigins: []string{"chrome-extension://", "moz-extension://"}, StickyMax: 300, Stickiness: true},
		Applications: []*types.Application{
			{
				UserID: "auth0|james_holden",
				Name:   "new_pokt_app_777",
				GatewayAAT: types.GatewayAAT{
					Address:              "test_34715cae753e67c75fbb340442e7de8e",
					ApplicationPublicKey: "test_34715cae753e67c75fbb340442e7de8e",
					ApplicationSignature: "test_1dc39a2e5a84a35bf030969a0b3231f7",
					ClientPublicKey:      "test_89a3af6a587aec02cfade6f5000424c2",
					PrivateKey:           "test_11b8d394ca331d7c7a71ca1896d630f6",
					Version:              "0.0.1",
				},
				GatewaySettings:      types.GatewaySettings{SecretKey: "test_40f482d91a5ef2300ebb4e2308c"},
				Limit:                types.AppLimit{PayPlan: types.PayPlan{Type: "basic_plan", Limit: 1_000}, CustomLimit: 0},
				NotificationSettings: types.NotificationSettings{SignedUp: true, Quarter: false, Half: false, ThreeQuarters: true, Full: true},
			},
		},
		Users: []types.UserAccess{
			{UserID: "user_1", RoleName: types.RoleOwner, Email: "james.holden123@test.com", Accepted: true},
			{UserID: "user_2", RoleName: types.RoleAdmin, Email: "paul.atreides456@test.com", Accepted: true},
			{UserID: "user_8", RoleName: types.RoleAdmin, Email: "rick.deckard456@test.com", Accepted: false},
		},
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
