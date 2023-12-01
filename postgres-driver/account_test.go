package postgresdriver

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/testdata"
	"github.com/pokt-foundation/portal-http-db/v2/types"
)

func (ts *PGDriverTestSuite) Test_PGDriver_ReadAccounts() {
	tests := []struct {
		name     string
		accounts map[types.AccountID]*types.Account
		options  types.DriverOptions
		err      error
	}{
		{
			name: "Should return all non-deleted Accounts from the database",
			accounts: map[types.AccountID]*types.Account{
				types.AccountID("account_1"): testdata.Accounts[types.AccountID("account_1")],
				types.AccountID("account_2"): testdata.Accounts[types.AccountID("account_2")],
				types.AccountID("account_3"): testdata.Accounts[types.AccountID("account_3")],
				types.AccountID("account_4"): testdata.Accounts[types.AccountID("account_4")],
				types.AccountID("account_5"): testdata.Accounts[types.AccountID("account_5")],
			},
			options: types.DriverOptions{},
			err:     nil,
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			accounts, err := ts.driver.ReadAccounts(context.Background(), test.options)
			ts.Equal(test.err, err)
			ts.Equal(test.accounts, accounts)
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_WriteAccount() {
	tests := []struct {
		name            string
		ownerID         types.UserID
		account         types.Account
		testCreatedTime time.Time
		users           map[types.UserID]types.AccountUser
		err             error
	}{
		{
			name:            "Should create a new Account in the database",
			ownerID:         "user_1",
			account:         *testdata.TestCreateAccount,
			testCreatedTime: testdata.MockTimestamp,
			users: map[types.UserID]types.AccountUser{
				"user_1": {
					UserID:             testdata.Users["user_1"].ID,
					Email:              testdata.Users["user_1"].Email,
					Owner:              true,
					PortalAppRoles:     map[types.PortalAppID]types.RoleName{},
					PortalAppsAccepted: map[types.PortalAppID]bool{},
				},
			},
		},
		{
			name:            "Should fail if input Account has an invalid icon URL set",
			ownerID:         "user_1",
			account:         types.Account{IconURL: "what-even~am_I"},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errInvalidIconURL.Error(), "what-even~am_I"),
		},
		{
			name:            "Should fail if input Account does not have a PayPlanType set",
			ownerID:         "user_1",
			account:         types.Account{PlanType: ""},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errPayPlanDoesntExist.Error(), ""),
		},
		{
			name:            "Should fail if input Account has an invalid plan type",
			ownerID:         "user_1",
			account:         types.Account{PlanType: types.PayPlanType("turbo_ultra_mega_plan")},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errPayPlanDoesntExist.Error(), types.PayPlanType("turbo_ultra_mega_plan")),
		},
		{
			name:            "Should fail if input User does not exist in the db",
			ownerID:         "user_451",
			account:         *testdata.Accounts[types.AccountID("account_5")],
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errUserDoesntExist.Error(), "user_451"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			createdAccount, err := ts.driver.WriteAccount(context.Background(), test.ownerID, test.account, test.testCreatedTime)
			ts.Equal(test.err, err)

			if test.err == nil {
				test.account.ID = createdAccount.ID
				test.account.AccountUsers = test.users
				ts.Equal(&test.account, createdAccount)

				accounts, err := ts.driver.ReadAccounts(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				ts.Equal(&test.account, accounts[createdAccount.ID])
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_UpsertAccountIntegration() {
	tests := []struct {
		name              string
		integrationsInput types.AccountIntegrations
		expectedAPIKeyRow string
		err               error
	}{
		{
			name: "Should create a single account integrations row",
			integrationsInput: types.AccountIntegrations{
				AccountID:          "account_5",
				CovalentAPIKeyFree: "created_covalent_api_key_1",
			},
			expectedAPIKeyRow: "created_covalent_api_key_1",
			err:               nil,
		},
		{
			name: "Should update an existing account integrations row",
			integrationsInput: types.AccountIntegrations{
				AccountID:          "account_1",
				CovalentAPIKeyFree: "updated_covalent_api_key_1",
			},
			expectedIntegrations: types.AccountIntegrations{
				AccountID:            "account_1",
				StripeSubscriptionID: "stripe_id_1",
				CovalentAPIKeyFree:   "updated_covalent_api_key_1",
			},
		},
		{
			name: "Should update an existing account integrations row with Stripe subscription ID",
			integrationsInput: types.AccountIntegrations{
				AccountID:            "account_2",
				StripeSubscriptionID: "updated_stripe_id_2",
			},
			expectedIntegrations: types.AccountIntegrations{
				AccountID:            "account_2",
				StripeSubscriptionID: "updated_stripe_id_2",
				CovalentAPIKeyFree:   "covalent_api_key_2",
			},
		},
		{
			name: "Should update an existing account integrations row with both values",
			integrationsInput: types.AccountIntegrations{
				AccountID:            "account_3",
				StripeSubscriptionID: "updated_stripe_id_3",
				CovalentAPIKeyFree:   "updated_covalent_api_key_3",
			},
			expectedIntegrations: types.AccountIntegrations{
				AccountID:            "account_3",
				StripeSubscriptionID: "updated_stripe_id_3",
				CovalentAPIKeyFree:   "updated_covalent_api_key_3",
			},
		},
	}

	for _, test := range tests {
		createdIntegrations, err := ts.driver.UpsertAccountIntegration(context.Background(), test.integrationsInput)
		ts.Equal(test.err, err)
		if err == nil {
			ts.Equal(test.integrationsInput, *createdIntegrations)

			if test.err == nil {
				accounts, err := ts.driver.ReadAccounts(context.Background(), types.DriverOptions{IncludeDeleted: false})
				ts.NoError(err)
				exists := false
				for _, account := range accounts {
					if account.ID == test.integrationsInput.AccountID {
						exists = true
						ts.Equal(test.expectedAPIKeyRow, account.Integrations.CovalentAPIKeyFree)
						break
					}
				}
				ts.True(exists)
			}
		}
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_UpdateAccount() {
	tests := []struct {
		name    string
		update  types.UpdateAccount
		account *types.Account
		err     error
	}{
		{
			name: "Should update the account's PlanType field to Enterprise",
			update: types.UpdateAccount{
				AccountID:       "account_1",
				PlanType:        types.EnterprisePlan,
				EnterpriseLimit: 3_000_000,
			},
			account: testdata.Accounts()["account_1"],
			err:     nil,
		},
		{
			name: "Should update the account's PlanType field to non-Enterprise",
			update: types.UpdateAccount{
				AccountID: "account_1",
				PlanType:  types.StarterPlan,
			},
			account: testdata.Accounts["account_1"],
			err:     nil,
		},
		{
			name: "Should update the account's Name field",
			update: types.UpdateAccount{
				AccountID: "account_1",
				Name:      "I Can Show You the World",
			},
			account: testdata.Accounts["account_1"],
			err:     nil,
		},
		{
			name: "Should update the account's IconURL field",
			update: types.UpdateAccount{
				AccountID: "account_1",
				IconURL:   "https://picsum.photos/222",
			},
			account: testdata.Accounts["account_1"],
			err:     nil,
		},
		{
			name: "Should update all fields",
			update: types.UpdateAccount{
				AccountID:            "account_3",
				PlanType:             types.AutoscalePlan,
				StripeSubscriptionID: testdata.StringToPointer("stripe_id_update_all_123"),
				Name:                 "Willy Wonka",
				IconURL:              "https://picsum.photos/444",
			},
<<<<<<< HEAD
			account: testdata.Accounts["account_1"],
=======
			account: testdata.Accounts()["account_3"],
>>>>>>> 37bc7c3 (feat: added back deprecated plan types for backwards compatibility)
			err:     nil,
		},
		{
			name: "Should fail if update Account has an invalid icon URL set",
			update: types.UpdateAccount{
				AccountID: "account_8823",
				PlanType:  types.EnterprisePlan,
				IconURL:   "who_did(this)",
			},
			err: fmt.Errorf(errInvalidIconURL.Error(), "who_did(this)"),
		},
		{
			name: "Should fail if update Account does not exist in the db",
			update: types.UpdateAccount{
				AccountID: "account_8823",
				PlanType:  types.StarterPlan,
			},
			err: fmt.Errorf(errAccountDoesntExist.Error(), "account_8823"),
		},
		{
			name: "Should fail if update Account has an invalid plan type",
			update: types.UpdateAccount{
				AccountID: "account_1",
				PlanType:  "turbo_ultra_mega_plan",
			},
			err: fmt.Errorf(errPayPlanDoesntExist.Error(), types.PayPlanType("turbo_ultra_mega_plan")),
		},
		{
			name: "Should fail if update Account has plan type Enterprise but no EnterpriseLimit",
			update: types.UpdateAccount{
				AccountID: "account_1",
				PlanType:  types.EnterprisePlan,
			},
			err: errEnterpriseMustHaveLimit,
		},
		{
			name: "Should fail if update Account has a non-Enterprise plan type but has an EnterpriseLimit",
			update: types.UpdateAccount{
				AccountID:       "account_1",
				PlanType:        types.StarterPlan,
				EnterpriseLimit: 123,
			},
			err: errNonEnterpriseMustNotHaveLimit,
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			account, err := ts.driver.UpdateAccount(context.Background(), test.update, testdata.MockTimestamp)
			ts.Equal(test.err, err)

			if test.err == nil {
				if test.update.PlanType != "" {
					ts.Equal(test.update.PlanType, account.PlanType)
					if test.update.PlanType == types.EnterprisePlan {
						ts.Equal(test.update.EnterpriseLimit, account.EnterpriseLimit)
					}
				}
<<<<<<< HEAD
				if test.update.StripeSubscriptionID != "" {
					ts.Equal(test.update.StripeSubscriptionID, account.Integrations.StripeSubscriptionID)
=======
				if test.update.StripeSubscriptionID != nil && *test.update.StripeSubscriptionID != "" {
					ts.Equal(test.update.StripeSubscriptionID, &account.Integrations.StripeSubscriptionID)
>>>>>>> 37bc7c3 (feat: added back deprecated plan types for backwards compatibility)
				}
				if test.update.Name != "" {
					ts.Equal(test.update.Name, account.Name)
				}
				if test.update.IconURL != "" {
					ts.Equal(test.update.IconURL, account.IconURL)
				}
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_WriteAccountUser() {
	tests := []struct {
		name                                string
		notSignedUp                         bool
		createAccountUser                   types.CreateAccountUserAccess
		accountUser, accountUserAfterCreate types.AccountUser
		testCreatedTime                     time.Time
		err                                 error
	}{
		{
			name: "Should create a new AccountUser row in the database for an existing User",
			createAccountUser: types.CreateAccountUserAccess{
				AccountID:   "account_1",
				PortalAppID: "test_app_1",
				Email:       "bernard.marx@test.com",
				RoleName:    types.RoleMember,
			},
			accountUser: testdata.AccountUser[13],
			accountUserAfterCreate: types.AccountUser{
				UserID:             "user_11",
				Email:              "bernard.marx@test.com",
				PortalAppRoles:     map[types.PortalAppID]types.RoleName{"test_app_1": types.RoleMember},
				PortalAppsAccepted: map[types.PortalAppID]bool{"test_app_1": false},
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name:        "Should create a new AccountUser row in the database for a user that hasn't signed up yet",
			notSignedUp: true,
			createAccountUser: types.CreateAccountUserAccess{
				AccountID:   "account_2",
				PortalAppID: "test_app_2",
				Email:       "winston.smith@test.com",
				RoleName:    types.RoleAdmin,
			},
			accountUser: testdata.AccountUser[14],
			accountUserAfterCreate: types.AccountUser{
				UserID:             "", // UserID created when user created
				Email:              "winston.smith@test.com",
				PortalAppRoles:     map[types.PortalAppID]types.RoleName{"test_app_2": types.RoleAdmin},
				PortalAppsAccepted: map[types.PortalAppID]bool{"test_app_2": false},
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should fail if an invalid email provided",
			createAccountUser: types.CreateAccountUserAccess{
				AccountID: "account_4",
				Email:     "winston.smith",
				RoleName:  types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errInvalidEmail.Error(), types.Email("winston.smith")),
		},
		{
			name: "Should fail if account does not exist",
			createAccountUser: types.CreateAccountUserAccess{
				AccountID:   "account_674",
				PortalAppID: "test_app_222",
				Email:       "winston.smith@test.com",
				RoleName:    types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errAccountDoesntExist.Error(), "account_674"),
		},
		{
			name: "Should fail if an empty email string is provided",
			createAccountUser: types.CreateAccountUserAccess{
				AccountID: "account_1",
				Email:     "",
				RoleName:  types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             errNoEmail,
		},
		{
			name: "Should fail if an invalid email string is provided",
			createAccountUser: types.CreateAccountUserAccess{
				AccountID: "account_1",
				Email:     "invalid_email",
				RoleName:  types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errInvalidEmail.Error(), types.Email("invalid_email")),
		},
		{
			name: "Should fail if an empty PortalAppID string is provided",
			createAccountUser: types.CreateAccountUserAccess{
				AccountID:   "account_1",
				PortalAppID: "",
				Email:       "valid.email@test.com",
				RoleName:    types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             errCreateNoPortalAppID,
		},
		{
			name: "Should fail if an empty AccountID string is provided",
			createAccountUser: types.CreateAccountUserAccess{
				AccountID:   "",
				PortalAppID: "test_app_1",
				Email:       "valid.email@test.com",
				RoleName:    types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             errCreateNoAccountID,
		},
		{
			name: "Should fail if the AccountID provided does not exist",
			createAccountUser: types.CreateAccountUserAccess{
				AccountID:   "non_existent_account",
				PortalAppID: "test_app_1",
				Email:       "valid.email@test.com",
				RoleName:    types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errAccountDoesntExist.Error(), "non_existent_account"),
		},
		{
			name: "Should fail if the PortalAppID provided does not exist",
			createAccountUser: types.CreateAccountUserAccess{
				AccountID:   "account_1",
				PortalAppID: "non_existent_app",
				Email:       "valid.email@test.com",
				RoleName:    types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errPortalAppDoesntExist.Error(), "non_existent_app"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			userID, err := ts.driver.WriteAccountUser(context.Background(), test.createAccountUser, testdata.MockTimestamp)
			ts.Equal(test.err, err)

			if test.err == nil {
				ts.NotEmpty(userID)
				test.accountUserAfterCreate.UserID = userID

				accounts, err := ts.driver.ReadAccounts(context.Background(), types.DriverOptions{})
				ts.NoError(err)

				user := accounts[test.createAccountUser.AccountID].AccountUsers[userID]
				ts.NotEmpty(user)
				ts.Equal(test.accountUserAfterCreate, user)
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_SetAccountUserRole() {
	tests := []struct {
		name                    string
		updateAccountUser       types.UpdateAccountUserRole
		accountUsersAfterUpdate map[types.UserID]types.AccountUser
		testCreatedTime         time.Time
		err                     error
	}{
		{
			name: "Should update an existing AccountUser row's role to non-OWNER role",
			updateAccountUser: types.UpdateAccountUserRole{
				PortalAppID: "test_app_3",
				AccountID:   "account_3",
				UserID:      "user_7",
				RoleName:    types.RoleAdmin,
			},
			accountUsersAfterUpdate: map[types.UserID]types.AccountUser{
				"user_5":  testdata.AccountUser[5],
				"user_6":  testdata.AccountUser[6],
				"user_10": testdata.AccountUser[12],
				"user_7": {
					UserID:             "user_7",
					Email:              "frodo.baggins123@test.com",
					PortalAppRoles:     map[types.PortalAppID]types.RoleName{"test_app_3": types.RoleAdmin},
					PortalAppsAccepted: map[types.PortalAppID]bool{"test_app_3": true},
				},
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should update an existing AccountUser row's role back to original role",
			updateAccountUser: types.UpdateAccountUserRole{
				PortalAppID: "test_app_3",
				AccountID:   "account_3",
				UserID:      "user_7",
				RoleName:    types.RoleMember,
			},
			accountUsersAfterUpdate: map[types.UserID]types.AccountUser{
				"user_5":  testdata.AccountUser[5],
				"user_6":  testdata.AccountUser[6],
				"user_10": testdata.AccountUser[12],
				"user_7": {
					UserID:             "user_7",
					Email:              "frodo.baggins123@test.com",
					PortalAppRoles:     map[types.PortalAppID]types.RoleName{"test_app_3": types.RoleMember},
					PortalAppsAccepted: map[types.PortalAppID]bool{"test_app_3": true},
				},
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should transfer the OWNER of an Account",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID:   "account_2",
				PortalAppID: "test_app_2",
				UserID:      "user_4",
				RoleName:    types.RoleOwner,
			},
			accountUsersAfterUpdate: map[types.UserID]types.AccountUser{
				"user_9": testdata.AccountUser[9],
				"user_2": testdata.AccountUser[10],
				"user_3": {
					UserID:             "user_3",
					Email:              "ellen.ripley789@test.com",
					PortalAppRoles:     map[types.PortalAppID]types.RoleName{"test_app_2": types.RoleAdmin},
					PortalAppsAccepted: map[types.PortalAppID]bool{"test_app_2": true},
				},
				"user_4": {
					UserID:             "user_4",
					Email:              "ulfric.stormcloak123@test.com",
					Owner:              true,
					PortalAppRoles:     map[types.PortalAppID]types.RoleName{"test_app_2": types.RoleOwner},
					PortalAppsAccepted: map[types.PortalAppID]bool{"test_app_2": true},
				},
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should transfer the OWNER of an Account back to original OWNER",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID:   "account_2",
				PortalAppID: "test_app_2",
				UserID:      "user_3",
				RoleName:    types.RoleOwner,
			},
			accountUsersAfterUpdate: map[types.UserID]types.AccountUser{
				"user_9": testdata.AccountUser[9],
				"user_2": testdata.AccountUser[10],
				"user_3": {
					UserID:             "user_3",
					Email:              "ellen.ripley789@test.com",
					Owner:              true,
					PortalAppRoles:     map[types.PortalAppID]types.RoleName{"test_app_2": types.RoleOwner},
					PortalAppsAccepted: map[types.PortalAppID]bool{"test_app_2": true},
				},
				"user_4": {
					UserID:             "user_4",
					Email:              "ulfric.stormcloak123@test.com",
					PortalAppRoles:     map[types.PortalAppID]types.RoleName{"test_app_2": types.RoleAdmin},
					PortalAppsAccepted: map[types.PortalAppID]bool{"test_app_2": true},
				},
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should work for users that have not accepted their invite yet",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID:   "account_3",
				PortalAppID: "test_app_3",
				UserID:      "user_10",
				RoleName:    types.RoleMember,
			},
			accountUsersAfterUpdate: map[types.UserID]types.AccountUser{
				"user_5": testdata.AccountUser[5],
				"user_6": testdata.AccountUser[6],
				"user_7": testdata.AccountUser[7],
				"user_10": {
					UserID:             "user_10",
					Email:              "daenerys.targaryen123@test.com",
					PortalAppRoles:     map[types.PortalAppID]types.RoleName{"test_app_3": types.RoleMember},
					PortalAppsAccepted: map[types.PortalAppID]bool{"test_app_3": false},
				},
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should fail if attempting to transfer ownership to user that has not accepted their invite yet",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID:   "account_3",
				PortalAppID: "test_app_3",
				UserID:      "user_10",
				RoleName:    types.RoleOwner,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errCannotTransferNotAccepted.Error(), "user_10", "account_3"),
		},
		{
			name: "Should fail if User is not a member of an Account",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID:   "account_2",
				PortalAppID: "test_app_2",
				UserID:      "user_512",
				RoleName:    types.RoleMember,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errAccountUserDoesntExist.Error(), "user_512", "account_2"),
		},
		{
			name: "Should fail if RoleName is empty",
			updateAccountUser: types.UpdateAccountUserRole{
				PortalAppID: "test_app_3",
				AccountID:   "account_3",
				UserID:      "user_7",
				RoleName:    "",
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             errNoRoleName,
		},
		{
			name: "Should fail if RoleName is invalid",
			updateAccountUser: types.UpdateAccountUserRole{
				PortalAppID: "test_app_3",
				AccountID:   "account_3",
				UserID:      "user_7",
				RoleName:    "INVALID_ROLE_NAME",
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             errInvalidRoleName,
		},
		{
			name: "Should fail if PortalAppID is empty",
			updateAccountUser: types.UpdateAccountUserRole{
				PortalAppID: "",
				AccountID:   "account_3",
				UserID:      "user_7",
				RoleName:    types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             errTransferNoPortalAppID,
		},
		{
			name: "Should fail if AccountID is empty",
			updateAccountUser: types.UpdateAccountUserRole{
				PortalAppID: "test_app_3",
				AccountID:   "",
				UserID:      "user_7",
				RoleName:    types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             errTransferNoAccountID,
		},
		{
			name: "Should fail if PortalAppID does not exist",
			updateAccountUser: types.UpdateAccountUserRole{
				PortalAppID: "non_existent_app",
				AccountID:   "account_3",
				UserID:      "user_7",
				RoleName:    types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errPortalAppDoesntExist.Error(), "non_existent_app"),
		},
		{
			name: "Should fail if User is not a member of an Account",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID:   "account_3",
				PortalAppID: "test_app_3",
				UserID:      "non_member_user",
				RoleName:    types.RoleMember,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errAccountUserDoesntExist.Error(), "non_member_user", "account_3"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			err := ts.driver.SetAccountUserRole(context.Background(), test.updateAccountUser, test.testCreatedTime)
			ts.Equal(test.err, err)

			if test.err == nil {
				accounts, err := ts.driver.ReadAccounts(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				ts.Equal(test.accountUsersAfterUpdate, accounts[test.updateAccountUser.AccountID].AccountUsers)
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_ZSetAccountUserRole_MultiplePortalApps() {
	// This test tests the transfer of ownership from one user to another
	// when multiple PortalApps are present on the account.
	// The Z is because it must run last in the suite.
	test := struct {
		updateAccountUser       types.UpdateAccountUserRole
		accountUsersAfterUpdate map[types.UserID]types.AccountUser
		testCreatedTime         time.Time
		testCreatePortalApp     types.PortalApp
		testCreateUsers         map[types.Email]types.CreateAccountUserAccess
		testAcceptUsers         map[types.Email]types.UpdateAcceptAccountUser
		allAppsUsers            []types.Email
		err                     error
	}{
		updateAccountUser: types.UpdateAccountUserRole{
			UserID:      "user_8",
			AccountID:   "account_5",
			RoleName:    types.RoleOwner,
			PortalAppID: "", // set dynamically in test
		},
		testCreatedTime: testdata.MockTimestamp,
		testCreatePortalApp: types.PortalApp{
			AccountID: "account_5",
			AppEmoji:  "1F336",
			Name:      "create_admin_role_pokt_app",
			Settings: types.Settings{
				Environment: types.EnvironmentProduction,
			},
			LegacyFields: types.LegacyFields{
				PlanType:       types.FreetierV0,
				DailyLimit:     250_000,
				CustomLimit:    0,
				RequestTimeout: 15_000,
			},
		},
		testCreateUsers: map[types.Email]types.CreateAccountUserAccess{
			"frodo.baggins123@test.com": {
				AccountID: "account_5",
				Email:     "frodo.baggins123@test.com",
				RoleName:  types.RoleMember,
			},
			"rick.deckard456@test.com": {
				AccountID: "account_5",
				Email:     "rick.deckard456@test.com",
				RoleName:  types.RoleAdmin,
			},
			"chrisjen.avasarala1@test.com": {
				AccountID: "account_5",
				Email:     "chrisjen.avasarala1@test.com",
				RoleName:  types.RoleAdmin,
			},
			"tyrion.lannister789@test.com": {
				AccountID: "account_5",
				Email:     "tyrion.lannister789@test.com",
				RoleName:  types.RoleMember,
			},
		},
		testAcceptUsers: map[types.Email]types.UpdateAcceptAccountUser{
			"frodo.baggins123@test.com": {
				PortalAppID:      "", // set dynamically in test
				UserID:           "user_7",
				AuthProviderType: types.AuthTypeAuth0Username,
				ProviderUserID:   "auth0|frodo_baggins",
			},
			"rick.deckard456@test.com": {
				PortalAppID:      "", // set dynamically in test
				UserID:           "user_8",
				AuthProviderType: types.AuthTypeAuth0Username,
				ProviderUserID:   "auth0|rick_deckard",
			},
			"chrisjen.avasarala1@test.com": {
				PortalAppID:      "", // set dynamically in test
				UserID:           "user_5",
				AuthProviderType: types.AuthTypeAuth0Username,
				ProviderUserID:   "auth0|chrisjen_avasarala",
			},
			"tyrion.lannister789@test.com": {
				PortalAppID:      "", // set dynamically in test
				UserID:           "user_9",
				AuthProviderType: types.AuthTypeAuth0Username,
				ProviderUserID:   "auth0|tyrion_lannister",
			},
		},
		accountUsersAfterUpdate: map[types.UserID]types.AccountUser{
			"user_4": {
				Owner:              false,
				UserID:             "user_4",
				Email:              "ulfric.stormcloak123@test.com",
				PortalAppRoles:     map[types.PortalAppID]types.RoleName{"placeholder": types.RoleAdmin}, // set dynamically in test
				PortalAppsAccepted: map[types.PortalAppID]bool{"placeholder": true},
			},
			"user_7": {
				Owner:              false,
				UserID:             "user_7",
				Email:              "frodo.baggins123@test.com",
				PortalAppRoles:     map[types.PortalAppID]types.RoleName{"placeholder": types.RoleMember}, // set dynamically in test
				PortalAppsAccepted: map[types.PortalAppID]bool{"placeholder": true},
			},
			"user_8": {
				Owner:              true,
				UserID:             "user_8",
				Email:              "rick.deckard456@test.com",
				PortalAppRoles:     map[types.PortalAppID]types.RoleName{"placeholder": types.RoleOwner}, // set dynamically in test
				PortalAppsAccepted: map[types.PortalAppID]bool{"placeholder": true},
			},
			"user_5": {
				Owner:              false,
				UserID:             "user_5",
				Email:              "chrisjen.avasarala1@test.com",
				PortalAppRoles:     map[types.PortalAppID]types.RoleName{"placeholder": types.RoleAdmin}, // set dynamically in test
				PortalAppsAccepted: map[types.PortalAppID]bool{"placeholder": true},
			},
			"user_9": {
				Owner:              false,
				UserID:             "user_9",
				Email:              "tyrion.lannister789@test.com",
				PortalAppRoles:     map[types.PortalAppID]types.RoleName{"placeholder": types.RoleMember}, // set dynamically in test
				PortalAppsAccepted: map[types.PortalAppID]bool{"placeholder": true},
			},
		},
		allAppsUsers: []types.Email{
			"ulfric.stormcloak123@test.com",
			"frodo.baggins123@test.com",
			"rick.deckard456@test.com",
		},
	}

	// Create portal apps and users
	portalAppIDs := []types.PortalAppID{} // Store the created PortalAppIDs
	for i := 0; i < 10; i++ {
		createdPortalApp, err := ts.driver.WritePortalApp(
			context.Background(),
			test.testCreatePortalApp,
			testdata.TestCreatePortalAppAAT,
			testdata.MockTimestamp,
		)
		ts.NoError(err)

		if i == 0 {
			test.updateAccountUser.PortalAppID = createdPortalApp.ID
		}

		portalAppIDs = append(portalAppIDs, createdPortalApp.ID) // Capture the created PortalAppID

		for email, createUser := range test.testCreateUsers {
			if i < 3 || slices.Contains(test.allAppsUsers, email) {
				createUser.PortalAppID = createdPortalApp.ID
				_, err := ts.driver.WriteAccountUser(context.Background(), createUser, testdata.MockTimestamp)
				ts.NoError(err)

				acceptUser := test.testAcceptUsers[email]
				acceptUser.PortalAppID = createdPortalApp.ID
				test.testAcceptUsers[email] = acceptUser
				err = ts.driver.UpdateAcceptAccountUser(context.Background(), acceptUser, testdata.MockTimestamp)
				ts.NoError(err)
			}
		}
	}

	// Update the expected data with the dynamically generated PortalAppIDs
	for userID, accountUserAccess := range test.accountUsersAfterUpdate {
		for i, portalAppID := range portalAppIDs {
			if i < 3 || slices.Contains(test.allAppsUsers, accountUserAccess.Email) {
				accountUserAccess.PortalAppRoles[portalAppID] = accountUserAccess.PortalAppRoles["placeholder"]
				accountUserAccess.PortalAppsAccepted[portalAppID] = accountUserAccess.PortalAppsAccepted["placeholder"]
			}
		}
		delete(accountUserAccess.PortalAppRoles, "placeholder")
		delete(accountUserAccess.PortalAppsAccepted, "placeholder")
		test.accountUsersAfterUpdate[userID] = accountUserAccess
	}

	// Perform the role change and assert
	err := ts.driver.SetAccountUserRole(context.Background(), test.updateAccountUser, test.testCreatedTime)
	ts.NoError(err)

	accounts, err := ts.driver.ReadAccounts(context.Background(), types.DriverOptions{})
	ts.NoError(err)
	ts.Equal(test.accountUsersAfterUpdate, accounts[test.updateAccountUser.AccountID].AccountUsers)
}

func (ts *PGDriverTestSuite) Test_PGDriver_UpdateAcceptAccountUser() {
	tests := []struct {
		name              string
		accountID         types.AccountID
		userID            types.UserID
		acceptAccountUser types.UpdateAcceptAccountUser
		user              *types.User
		newUser           bool
		newUserEmail      types.Email
		accountUsers      map[types.UserID]types.AccountUser
		err               error
	}{
		{
			name:      "Should update a user to Accepted if they accept their invite",
			accountID: "account_3",
			userID:    "user_10",
			acceptAccountUser: types.UpdateAcceptAccountUser{
				PortalAppID:      "test_app_3",
				UserID:           "user_10",
				AuthProviderType: types.AuthTypeAuth0Username,
				ProviderUserID:   "auth0|daenerys_targaryen",
			},
			user: &types.User{
				ID:               "user_10",
				Email:            "daenerys.targaryen123@test.com",
				SignedUp:         true,
				IconURL:          "https://picsum.photos/200",
				UpdatesProduct:   false,
				UpdatesMarketing: false,
				BetaTester:       true,
				AuthProviders: map[types.AuthType]types.UserAuthProvider{
					types.AuthTypeAuth0Username: {
						ProviderUserID: "auth0|daenerys_targaryen", Provider: types.AuthProviderAuth0,
						Type: types.AuthTypeAuth0Username, Federated: false,
					},
				},
				CreatedAt: testdata.MockTimestamp,
				UpdatedAt: testdata.MockTimestamp,
			},
			accountUsers: map[types.UserID]types.AccountUser{
				"user_5": testdata.AccountUser[5],
				"user_6": testdata.AccountUser[6],
				"user_7": testdata.AccountUser[7],
				"user_10": {
					UserID:             "user_10",
					Email:              "daenerys.targaryen123@test.com",
					PortalAppRoles:     map[types.PortalAppID]types.RoleName{"test_app_3": types.RoleMember},
					PortalAppsAccepted: map[types.PortalAppID]bool{"test_app_3": true},
				},
			},
		},
		{
			name:         "Should delete an existing user's account_user_access row if they decline their invite",
			accountID:    "account_1",
			userID:       "user_3",
			newUserEmail: "ellen.ripley789@test.com",
			acceptAccountUser: types.UpdateAcceptAccountUser{
				DeclinedInvite:   true,
				PortalAppID:      "test_app_1",
				UserID:           "user_3",
				AuthProviderType: types.AuthTypeAuth0Username,
				ProviderUserID:   "auth0|ellen_ripley",
			},
			user: &types.User{
				ID:               "user_3",
				Email:            "ellen.ripley789@test.com",
				IconURL:          "https://picsum.photos/200",
				SignedUp:         true,
				UpdatesProduct:   true,
				UpdatesMarketing: true,
				AuthProviders: map[types.AuthType]types.UserAuthProvider{
					"auth0_username": {ProviderUserID: "auth0|ellen_ripley", Type: "auth0_username", Provider: "auth0"},
				},
				CreatedAt: testdata.MockTimestamp,
				UpdatedAt: testdata.MockTimestamp,
			},
			accountUsers: map[types.UserID]types.AccountUser{
				"user_1": testdata.AccountUser[1],
				"user_2": testdata.AccountUser[2],
				"user_8": testdata.AccountUser[8],
			},
		},
		{
			name:         "Should delete a new user's account_user_access row if they decline their invite",
			accountID:    "account_1",
			userID:       "", // set in test case
			newUser:      true,
			newUserEmail: "red.rackham@test.com",
			acceptAccountUser: types.UpdateAcceptAccountUser{
				DeclinedInvite:   true,
				PortalAppID:      "test_app_3",
				UserID:           "", // set in test case
				AuthProviderType: types.AuthTypeAuth0Username,
				ProviderUserID:   "auth0|red_rackham",
			},
			user: &types.User{
				ID:       "e2bb5de8",
				Email:    "red.rackham@test.com",
				SignedUp: true,
				AuthProviders: map[types.AuthType]types.UserAuthProvider{
					"auth0_username": {ProviderUserID: "auth0|red_rackham", Type: "auth0_username", Provider: "auth0"},
				},
				CreatedAt: testdata.MockTimestamp,
				UpdatedAt: testdata.MockTimestamp,
			},
			accountUsers: map[types.UserID]types.AccountUser{
				"user_1": testdata.AccountUser[1],
				"user_2": testdata.AccountUser[2],
				"user_8": testdata.AccountUser[8],
			},
		},
		{
			name: "Should fail to decline invite if user has already accepted their invite",
			acceptAccountUser: types.UpdateAcceptAccountUser{
				DeclinedInvite:   true,
				PortalAppID:      "test_app_2",
				UserID:           "user_4",
				AuthProviderType: types.AuthTypeAuth0Username,
				ProviderUserID:   "auth0|ulfric_stormcloak",
			},
			err: fmt.Errorf(errCannotDeclineAlreadyAccepted.Error(), "user_4", "test_app_2"),
		},
		{
			name: "Should fail if an invalid auth provider type provided",
			acceptAccountUser: types.UpdateAcceptAccountUser{
				PortalAppID:      "test_app_3",
				UserID:           "user_10",
				AuthProviderType: types.AuthType("ask_jeeves"),
			},
			err: fmt.Errorf(errInvalidAuthProviderType.Error(), types.AuthType("ask_jeeves")),
		},
		{
			name: "Should fail if AuthProviderType is not provided",
			acceptAccountUser: types.UpdateAcceptAccountUser{
				PortalAppID:      "test_app_3",
				UserID:           "user_10",
				AuthProviderType: "",
			},
			err: errNoAuthProviderType,
		},
		{
			name: "Should fail if PortalAppID is not provided",
			acceptAccountUser: types.UpdateAcceptAccountUser{
				PortalAppID:      "",
				UserID:           "user_10",
				AuthProviderType: types.AuthTypeAuth0Username,
			},
			err: errAcceptNoPortalAppID,
		},
		{
			name: "Should fail if user does not exist",
			acceptAccountUser: types.UpdateAcceptAccountUser{
				PortalAppID:      "test_app_3",
				UserID:           "user_123",
				AuthProviderType: types.AuthTypeAuth0Username,
			},
			err: fmt.Errorf(errAccountUserDoesntExist.Error(), "user_123", "test_app_3"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			if test.acceptAccountUser.DeclinedInvite && test.err == nil {
				if test.newUser {
					newUser, _, err := ts.driver.WriteUserNewSignUp(context.Background(), types.CreateUser{
						Email:          test.newUserEmail,
						ProviderUserID: test.acceptAccountUser.ProviderUserID,
					}, testdata.MockTimestamp)
					ts.NoError(err)
					test.user.ID = newUser.ID
					test.userID = newUser.ID
					test.acceptAccountUser.UserID = newUser.ID
				}

				userID, err := ts.driver.WriteAccountUser(context.Background(), types.CreateAccountUserAccess{
					AccountID:   test.accountID,
					PortalAppID: test.acceptAccountUser.PortalAppID,
					Email:       test.newUserEmail,
					RoleName:    types.RoleMember,
				}, testdata.MockTimestamp)
				ts.NoError(err)
				ts.Equal(test.acceptAccountUser.UserID, userID)

				// Verify declined user is present in account before update
				accounts, err := ts.driver.ReadAccounts(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				userAccount := accounts[test.accountID]
				ts.NotEmpty(userAccount)
				ts.Equal(test.acceptAccountUser.UserID, accounts[test.accountID].AccountUsers[test.acceptAccountUser.UserID].UserID)
				ts.NotEmpty(accounts[test.accountID].AccountUsers[test.acceptAccountUser.UserID].PortalAppRoles[test.acceptAccountUser.PortalAppID])
				ts.NotEmpty(accounts[test.accountID].AccountUsers[test.acceptAccountUser.UserID].PortalAppsAccepted)
				ts.False(accounts[test.accountID].AccountUsers[test.acceptAccountUser.UserID].PortalAppsAccepted[test.acceptAccountUser.PortalAppID])
			}

			err := ts.driver.UpdateAcceptAccountUser(context.Background(), test.acceptAccountUser, testdata.MockTimestamp)
			ts.Equal(test.err, err)

			if test.err == nil {
				user, err := ts.driver.readUserByUserID(context.Background(), test.userID)
				ts.NoError(err)
				ts.Equal(test.user, user)

				accounts, err := ts.driver.ReadAccounts(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				ts.Equal(test.accountUsers, accounts[test.accountID].AccountUsers)
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_DeleteAccount() {
	tests := []struct {
		name                    string
		accountID               types.AccountID
		numAccountsBeforeDelete int
		accountsAfterDelete     map[types.AccountID]*types.Account
		err                     error
	}{
		{
			name:                    "Should set a Account's deleted field to true, causing it to not appear in the ReadAccounts query",
			numAccountsBeforeDelete: 6,
			accountsAfterDelete: map[types.AccountID]*types.Account{
				types.AccountID("account_1"): testdata.Accounts[types.AccountID("account_1")],
				types.AccountID("account_2"): testdata.Accounts[types.AccountID("account_2")],
				types.AccountID("account_3"): testdata.Accounts[types.AccountID("account_3")],
				types.AccountID("account_4"): testdata.Accounts[types.AccountID("account_4")],
				types.AccountID("account_5"): testdata.Accounts[types.AccountID("account_5")],
			},
		},
		{
			name:                    "Should fail if account does not exist",
			accountID:               "account_347",
			numAccountsBeforeDelete: 5,
			err:                     fmt.Errorf(errAccountDoesntExist.Error(), "account_347"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			if test.accountID == types.AccountID("") {
				// Create test Account to delete
				createdAccount, err := ts.driver.WriteAccount(
					context.Background(),
					"user_1",
					types.Account{PlanType: types.PayPlanType("developer_plan")},
					testdata.MockTimestamp,
				)
				ts.NoError(err)
				test.accountID = createdAccount.ID
			}

			// Check all Accounts exist before delete
			accountsBeforeDelete, err := ts.driver.ReadAccounts(context.Background(), types.DriverOptions{IncludeDeleted: false})
			ts.NoError(err)
			ts.Len(accountsBeforeDelete, test.numAccountsBeforeDelete)

			// Delete Account
			err = ts.driver.SetAccountDeleted(context.Background(), test.accountID, testdata.MockTimestamp)
			ts.Equal(test.err, err)

			if test.err == nil {
				// Check Account was deleted
				accounts, err := ts.driver.ReadAccounts(context.Background(), types.DriverOptions{IncludeDeleted: false})
				ts.NoError(err)
				ts.Len(accounts, test.numAccountsBeforeDelete-1)
				ts.Equal(test.accountsAfterDelete, accounts)

				// Check Account still appears if IncludeDeleted: true
				accounts, err = ts.driver.ReadAccounts(context.Background(), types.DriverOptions{IncludeDeleted: true})
				ts.NoError(err)
				testDeletedApp, ok := accountsBeforeDelete[test.accountID]
				ts.True(ok)
				testDeletedApp.Deleted = true
				ts.Equal(accountsBeforeDelete, accounts)
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_RemoveAccountUser() {
	tests := []struct {
		name                    string
		userID                  types.UserID
		accountID               types.AccountID
		portalAppID             types.PortalAppID
		numAccountsBeforeDelete int
		accountUsersAfterDelete map[types.UserID]types.AccountUser
		err                     error
	}{
		{
			name:                    "Should delete a single AccountUser row",
			accountID:               "account_1",
			portalAppID:             "test_app_1",
			numAccountsBeforeDelete: 5,
			accountUsersAfterDelete: map[types.UserID]types.AccountUser{
				"user_1": testdata.AccountUser[1],
				"user_2": testdata.AccountUser[2],
				"user_8": testdata.AccountUser[8],
			},
		},
		{
			name:                    "Should fail if provided a UserID that doesn't exist for the Account",
			userID:                  "user_789",
			accountID:               "account_3",
			portalAppID:             "test_app_3",
			numAccountsBeforeDelete: 5,
			err:                     fmt.Errorf(errAccountUserDoesntExist.Error(), "user_789", "test_app_3"),
		},
		{
			name:        "Should fail if attempting to delete the current Account OWNER",
			userID:      "user_1",
			accountID:   "account_1",
			portalAppID: "test_app_1",
			err:         fmt.Errorf(errCannotDeleteOwner.Error(), "user_1", "account_1"),
		},
		{
			name:                    "Should fail if provided a UserID that doesn't exist for the Account",
			userID:                  "user_nonexistent",
			accountID:               "account_1",
			portalAppID:             "test_app_1",
			numAccountsBeforeDelete: 5,
			err:                     fmt.Errorf(errAccountUserDoesntExist.Error(), "user_nonexistent", "test_app_1"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			if test.err == nil {
				if test.userID == types.UserID("") {
					// create test User to delete
					userID, err := ts.driver.WriteAccountUser(context.Background(), types.CreateAccountUserAccess{
						AccountID:   test.accountID,
						PortalAppID: test.portalAppID,
						Email:       "hermaeus.mora@example.com",
						RoleName:    types.RoleMember,
					}, testdata.MockTimestamp)
					ts.NoError(err)
					test.userID = userID
				}

				// check all Accounts exist before delete
				accounts, err := ts.driver.ReadAccounts(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				ts.Len(accounts, test.numAccountsBeforeDelete)
			}

			err := ts.driver.RemoveAccountUser(context.Background(), test.userID, test.portalAppID, test.accountID)
			ts.Equal(test.err, err)

			if test.err == nil {
				accounts, err := ts.driver.ReadAccounts(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				ts.Equal(test.accountUsersAfterDelete, accounts[test.accountID].AccountUsers)
			}
		})
	}
}
