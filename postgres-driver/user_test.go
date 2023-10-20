package postgresdriver

import (
	"context"
	"fmt"
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/testdata"
	"github.com/pokt-foundation/portal-http-db/v2/types"
)

func (ts *PGDriverTestSuite) Test_PGDriver_ReadUserIDsMap() {
	tests := []struct {
		name               string
		expectedUserIDsMap map[types.ProviderUserID]types.UserID
		err                error
	}{
		{
			name:               "Should return the Portal UserID when passed the auth provider user ID",
			expectedUserIDsMap: testdata.UserIDs,
			err:                nil,
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			userIDsMap, err := ts.driver.ReadUserIDsMap(context.Background())
			ts.Equal(test.err, err)
			ts.Equal(test.expectedUserIDsMap, userIDsMap)
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_readUserByUserID() {
	tests := []struct {
		name   string
		userID types.UserID
		user   *types.User
		err    error
	}{
		{
			name:   "Should return a Portal User when passed the portal UserID",
			userID: "user_1",
			user:   testdata.Users["user_1"],
			err:    nil,
		},
		{
			name:   "Should return a Portal User when passed the portal UserID",
			userID: "user_7",
			user:   testdata.Users["user_7"],
			err:    nil,
		},
		{
			name:   "Should fail if the user does not exist in the database",
			userID: "user_42",
			err:    fmt.Errorf(errUserDoesntExist.Error(), "user_42"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			user, err := ts.driver.readUserByUserID(context.Background(), test.userID)
			ts.Equal(test.err, err)

			if test.err == nil {
				// Method is only used internally, so we can safely set the permissions to nil
				testUser := test.user
				testUser.Permissions = nil
				ts.Equal(testUser, user)
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_ReadAllUsers() {
	tests := []struct {
		name          string
		expectedUsers map[types.UserID]*types.User
		err           error
	}{
		{
			name:          "Should return all users from the database",
			expectedUsers: testdata.Users,
			err:           nil,
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			usersMap, err := ts.driver.ReadAllUsers(context.Background())
			ts.Equal(test.err, err)

			if test.err == nil {
				ts.Equal(test.expectedUsers, usersMap)
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_WriteNewUser() {
	tests := []struct {
		name       string
		createUser types.CreateUser
		user       *types.User
		err        error
	}{
		{
			name: "Should create a new portal User in the DB",
			createUser: types.CreateUser{
				Email:          "geralt.of.rivia623@example.com",
				ProviderUserID: "auth0|geralt_of_rivia",
			},
			user: &types.User{
				ID:       "", // user ID set in test case
				Email:    "geralt.of.rivia623@example.com",
				SignedUp: true,
				AuthProviders: map[types.AuthType]types.UserAuthProvider{
					types.AuthTypeAuth0Username: {
						ProviderUserID: "auth0|geralt_of_rivia",
						Type:           types.AuthTypeAuth0Username,
						Provider:       types.AuthProviderAuth0,
						Federated:      false,
					},
				},
				CreatedAt: testdata.MockTimestamp,
				UpdatedAt: testdata.MockTimestamp,
			},
			err: nil,
		},
		{
			name: "Should fail if an invalid email provided",
			createUser: types.CreateUser{
				Email: "jar.jar.binks3",
			},
			err: fmt.Errorf(errInvalidEmail.Error(), types.Email("jar.jar.binks3")),
		},
		{
			name: "Should fail if an invalid auth provider type provided",
			createUser: types.CreateUser{
				Email:          "jar.jar.binks3@example.com",
				ProviderUserID: "wtf|jar_jar_binks",
			},
			err: fmt.Errorf(errInvalidAuthProviderType.Error(), types.AuthType("wtf")),
		},
		{
			name: "Should fail if user already exists with provided email and auth provider type",
			createUser: types.CreateUser{
				Email:          "geralt.of.rivia623@example.com",
				ProviderUserID: "auth0|geralt_of_rivia",
			},
			err: fmt.Errorf(errUserAlreadyExists.Error(), "geralt.of.rivia623@example.com", "auth0_username"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			user, accountID, err := ts.driver.WriteUserNewSignUp(context.Background(), test.createUser, testdata.MockTimestamp)
			ts.Equal(test.err, err)

			if test.err == nil {
				test.user.ID = user.ID
				ts.Equal(test.user, user)

				ts.NotEmpty(accountID)

				accounts, err := ts.driver.ReadAccounts(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				exists := false
				for _, account := range accounts {
					for _, user := range account.AccountUsers {
						if user.UserID == test.user.ID {
							exists = true
							ts.Equal(test.createUser.Email, user.Email)
							break
						}
					}
				}
				ts.True(exists)
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_UpdateUser() {
	tests := []struct {
		name            string
		update          types.UpdateUser
		testUpdatedTime time.Time
		err             error
	}{
		{
			name:            "Should update an existing user in the database",
			update:          testdata.UpdateUserOne,
			testUpdatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name:            "Should update an existing user again in the database",
			update:          testdata.UpdateUserTwo,
			testUpdatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name:            "Should fail if input user has an invalid icon URL set",
			update:          testdata.UpdateUserInvalidURL,
			testUpdatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errInvalidIconURL.Error(), "i-am-not-a-url"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			var initialUser *types.User
			if test.err == nil {
				user, err := ts.driver.readUserByUserID(context.Background(), test.update.ID)
				ts.NoError(err)
				initialUser = user
			}

			updatedUser, err := ts.driver.UpdateUser(context.Background(), test.update, test.testUpdatedTime)
			ts.Equal(test.err, err)

			if test.err == nil {
				ts.Equal(updatedUser.UpdatedAt, test.testUpdatedTime)

				// Assert that the fields in the updated user match the expected values
				// Only check the fields present in the update struct
				if test.update.IconURL != nil && *test.update.IconURL != "" {
					ts.Equal(*test.update.IconURL, updatedUser.IconURL)
				} else {
					ts.Equal(initialUser.IconURL, updatedUser.IconURL)
				}

				if test.update.UpdatesProduct != nil {
					ts.Equal(*test.update.UpdatesProduct, updatedUser.UpdatesProduct)
				} else {
					ts.Equal(initialUser.UpdatesProduct, updatedUser.UpdatesProduct)
				}

				if test.update.UpdatesMarketing != nil {
					ts.Equal(*test.update.UpdatesMarketing, updatedUser.UpdatesMarketing)
				} else {
					ts.Equal(initialUser.UpdatesMarketing, updatedUser.UpdatesMarketing)
				}

				if test.update.BetaTester != nil {
					ts.Equal(*test.update.BetaTester, updatedUser.BetaTester)
				} else {
					ts.Equal(initialUser.BetaTester, updatedUser.BetaTester)
				}
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_Z_DeletePortalUser() {
	tests := []struct {
		name        string
		userID      types.UserID
		expectedErr error
		err         error
	}{
		{
			name:        "Should fail to delete a portal User from the DB if they are on the team of any accounts",
			userID:      "user_1",
			expectedErr: errUserDoesntExist,
			err:         errUserHasAccount,
		},
		{
			name:        "Should delete a portal User from the DB",
			userID:      "user_12",
			expectedErr: errUserDoesntExist,
			err:         nil,
		},
		{
			name:   "Should fail if the user does not exist in the database",
			userID: "user_42",
			err:    fmt.Errorf(errUserDoesntExist.Error(), "user_42"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			userID, err := ts.driver.DeletePortalUser(context.Background(), test.userID)
			ts.Equal(test.err, err)

			if test.err == nil {
				_, err := ts.driver.readUserByUserID(context.Background(), userID)
				ts.Equal(fmt.Errorf(test.expectedErr.Error(), userID), err)
			}
		})
	}
}
