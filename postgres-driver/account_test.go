package postgresdriver

import (
	"context"
	"fmt"
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
					ID:       testdata.Users["user_1"].ID,
					Email:    testdata.Users["user_1"].Email,
					RoleName: types.RoleOwner,
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
				test.account.Users = test.users
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
			expectedAPIKeyRow: "updated_covalent_api_key_1",
			err:               nil,
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
			name: "Should update the account's PlanType field",
			update: types.UpdateAccount{
				AccountID: "account_1",
				PlanType:  types.Enterprise,
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
				AccountID: "account_1",
				PlanType:  types.TestPlan10K,
				Name:      "Willy Wonka",
				IconURL:   "https://picsum.photos/444",
			},
			account: testdata.Accounts["account_1"],
			err:     nil,
		},
		{
			name: "Should fail if update Account has an invalid icon URL set",
			update: types.UpdateAccount{
				AccountID: "account_8823",
				PlanType:  types.Enterprise,
				IconURL:   "who_did(this)",
			},
			err: fmt.Errorf(errInvalidIconURL.Error(), "who_did(this)"),
		},
		{
			name: "Should fail if update Account does not exist in the db",
			update: types.UpdateAccount{
				AccountID: "account_8823",
				PlanType:  types.Enterprise,
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
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			account, err := ts.driver.UpdateAccount(context.Background(), test.update, testdata.MockTimestamp)
			ts.Equal(test.err, err)

			if test.err == nil {
				if test.update.PlanType != "" {
					ts.Equal(test.update.PlanType, account.PlanType)
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
		createAccountUser                   types.CreateAccountUser
		accountUser, accountUserAfterCreate types.AccountUser
		testCreatedTime                     time.Time
		err                                 error
	}{
		{
			name: "Should create a new AccountUser row in the database for an existing User",
			createAccountUser: types.CreateAccountUser{
				AccountID: "account_1",
				Email:     "bernard.marx@test.com",
				RoleName:  types.RoleMember,
			},
			accountUser: testdata.AccountUser[13],
			accountUserAfterCreate: types.AccountUser{
				ID:       "user_11",
				Email:    "bernard.marx@test.com",
				RoleName: types.RoleMember,
				Accepted: false,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name:        "Should create a new AccountUser row in the database for a user that hasn't signed up yet",
			notSignedUp: true,
			createAccountUser: types.CreateAccountUser{
				AccountID: "account_2",
				Email:     "winston.smith@test.com",
				RoleName:  types.RoleAdmin,
			},
			accountUser: testdata.AccountUser[14],
			accountUserAfterCreate: types.AccountUser{
				ID:       "", // UserID created when user created
				Email:    "winston.smith@test.com",
				RoleName: types.RoleAdmin,
				Accepted: false,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should fail if an invalid email provided",
			createAccountUser: types.CreateAccountUser{
				AccountID: "account_4",
				Email:     "winston.smith",
				RoleName:  types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errInvalidEmail.Error(), types.Email("winston.smith")),
		},
		{
			name: "Should fail if account does not exist",
			createAccountUser: types.CreateAccountUser{
				AccountID: "account_674",
				Email:     "winston.smith@test.com",
				RoleName:  types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errAccountDoesntExist.Error(), "account_674"),
		},
		{
			name: "Should fail if an empty email string is provided",
			createAccountUser: types.CreateAccountUser{
				AccountID: "account_1",
				Email:     "",
				RoleName:  types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             errNoEmail,
		},
		{
			name: "Should fail if an invalid email string is provided",
			createAccountUser: types.CreateAccountUser{
				AccountID: "account_1",
				Email:     "invalid_email",
				RoleName:  types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errInvalidEmail.Error(), types.Email("invalid_email")),
		},
		{
			name: "Should fail if an empty AccountID string is provided",
			createAccountUser: types.CreateAccountUser{
				AccountID: "",
				Email:     "valid.email@test.com",
				RoleName:  types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             errCreateNoAccountID,
		},
		{
			name: "Should fail if the AccountID provided does not exist",
			createAccountUser: types.CreateAccountUser{
				AccountID: "non_existent_account",
				Email:     "valid.email@test.com",
				RoleName:  types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errAccountDoesntExist.Error(), "non_existent_account"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			userID, err := ts.driver.WriteAccountUser(context.Background(), test.createAccountUser, testdata.MockTimestamp)
			ts.Equal(test.err, err)

			if test.err == nil {
				ts.NotEmpty(userID)
				test.accountUserAfterCreate.ID = userID

				accounts, err := ts.driver.ReadAccounts(context.Background(), types.DriverOptions{})
				ts.NoError(err)

				user := accounts[test.createAccountUser.AccountID].Users[userID]
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
				AccountID: "account_3",
				UserID:    "user_7",
				RoleName:  types.RoleAdmin,
			},
			accountUsersAfterUpdate: map[types.UserID]types.AccountUser{
				"user_5":  testdata.AccountUser[5],
				"user_6":  testdata.AccountUser[6],
				"user_10": testdata.AccountUser[12],
				"user_7": {
					ID:       "user_7",
					Email:    "frodo.baggins123@test.com",
					RoleName: types.RoleAdmin,
					Accepted: true,
				},
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should update an existing AccountUser row's role back to original role",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID: "account_3",
				UserID:    "user_7",
				RoleName:  types.RoleMember,
			},
			accountUsersAfterUpdate: map[types.UserID]types.AccountUser{
				"user_5":  testdata.AccountUser[5],
				"user_6":  testdata.AccountUser[6],
				"user_10": testdata.AccountUser[12],
				"user_7": {
					ID:       "user_7",
					Email:    "frodo.baggins123@test.com",
					RoleName: types.RoleMember,
					Accepted: true,
				},
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should transfer the OWNER of an Account",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID: "account_2",
				UserID:    "user_4",
				RoleName:  types.RoleOwner,
			},
			accountUsersAfterUpdate: map[types.UserID]types.AccountUser{
				"user_9": testdata.AccountUser[9],
				"user_2": testdata.AccountUser[10],
				"user_3": {
					ID:       "user_3",
					Email:    "ellen.ripley789@test.com",
					RoleName: types.RoleAdmin,
					Accepted: true,
				},
				"user_4": {
					ID:       "user_4",
					Email:    "ulfric.stormcloak123@test.com",
					RoleName: types.RoleOwner,
					Accepted: true,
				},
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should transfer the OWNER of an Account back to original OWNER",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID: "account_2",
				UserID:    "user_3",
				RoleName:  types.RoleOwner,
			},
			accountUsersAfterUpdate: map[types.UserID]types.AccountUser{
				"user_9": testdata.AccountUser[9],
				"user_2": testdata.AccountUser[10],
				"user_3": {
					ID:       "user_3",
					Email:    "ellen.ripley789@test.com",
					RoleName: types.RoleOwner,
					Accepted: true,
				},
				"user_4": {
					ID:       "user_4",
					Email:    "ulfric.stormcloak123@test.com",
					RoleName: types.RoleAdmin,
					Accepted: true,
				},
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should work for users that have not accepted their invite yet",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID: "account_3",
				UserID:    "user_10",
				RoleName:  types.RoleMember,
			},
			accountUsersAfterUpdate: map[types.UserID]types.AccountUser{
				"user_5": testdata.AccountUser[5],
				"user_6": testdata.AccountUser[6],
				"user_7": testdata.AccountUser[7],
				"user_10": {
					ID:       "user_10",
					Email:    "daenerys.targaryen123@test.com",
					RoleName: types.RoleMember,
					Accepted: false,
				},
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should fail if attempting to transfer ownership to user that has not accepted their invite yet",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID: "account_3",
				UserID:    "user_10",
				RoleName:  types.RoleOwner,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errCannotTransferNotAccepted.Error(), "user_10", "account_3"),
		},
		{
			name: "Should fail if User is not a member of an Account",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID: "account_2",
				UserID:    "user_512",
				RoleName:  types.RoleMember,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errAccountUserDoesntExist.Error(), "user_512", "account_2"),
		},
		{
			name: "Should fail if RoleName is empty",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID: "account_3",
				UserID:    "user_7",
				RoleName:  "",
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             errNoRoleName,
		},
		{
			name: "Should fail if RoleName is invalid",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID: "account_3",
				UserID:    "user_7",
				RoleName:  "INVALID_ROLE_NAME",
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             errInvalidRoleName,
		},
		{
			name: "Should fail if AccountID is empty",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID: "",
				UserID:    "user_7",
				RoleName:  types.RoleAdmin,
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             errTransferNoAccountID,
		},
		{
			name: "Should fail if User is not a member of an Account",
			updateAccountUser: types.UpdateAccountUserRole{
				AccountID: "account_3",
				UserID:    "non_member_user",
				RoleName:  types.RoleMember,
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
				ts.Equal(test.accountUsersAfterUpdate, accounts[test.updateAccountUser.AccountID].Users)
			}
		})
	}
}
