package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var testUserPermissions = map[UserID]User{
	"user_1": {
		ID: "user_1",
		Permissions: map[AccountID]AccountPermissions{
			"account_1": {
				Accepted:    true,
				RoleName:    RoleOwner,
				Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint, PermDeleteEndpoint, PermTransferEndpoint},
			},
		},
	},
	"user_2": {
		ID: "user_2",
		Permissions: map[AccountID]AccountPermissions{
			"account_1": {
				Accepted:    true,
				RoleName:    RoleAdmin,
				Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint},
			},
			"account_2": {
				Accepted:    false,
				RoleName:    RoleMember,
				Permissions: []Permissions{PermReadEndpoint},
			},
		},
	},
	"user_3": {
		ID: "user_3",
		Permissions: map[AccountID]AccountPermissions{
			"account_2": {
				Accepted:    true,
				RoleName:    RoleOwner,
				Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint, PermDeleteEndpoint, PermTransferEndpoint},
			},
		},
	},
}

// deepCopyUserPermissions returns a deep copy of a User object
func deepCopyUserPermissions(user User) User {
	newUser := User{
		ID:          user.ID,
		Permissions: make(map[AccountID]AccountPermissions, len(user.Permissions)),
	}

	for k, v := range user.Permissions {
		newPerms := make([]Permissions, len(v.Permissions))
		copy(newPerms, v.Permissions)

		newUser.Permissions[k] = AccountPermissions{
			Accepted:    v.Accepted,
			RoleName:    v.RoleName,
			Permissions: newPerms,
		}
	}
	return newUser
}

// getUserPermissions returns a deep copy of a single User in the testUserPermissions map
func getUserPermissions(userID UserID) User {
	userPermissions := testUserPermissions[userID]
	return deepCopyUserPermissions(userPermissions)
}

func Test_UserPermissions_GetAccountRole(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name      string
		userID    UserID
		accountID AccountID
		roleName  RoleName
	}{
		{
			name:      "Should return the role for a given user and account ID",
			userID:    "user_1",
			accountID: "account_1",
			roleName:  RoleOwner,
		},
		{
			name:      "Should return the role for a given user and account ID",
			userID:    "user_2",
			accountID: "account_2",
			roleName:  RoleMember,
		},
		{
			name:      "Should return an empty string if user does not have a role for a account",
			userID:    "user_1",
			accountID: "account_77",
			roleName:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPermissions := getUserPermissions(test.userID)

			roleName := userPermissions.GetAccountRole(test.accountID)
			c.Equal(test.roleName, roleName)
		})
	}
}

func Test_UserPermissions_UpsertPermissions(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name                string
		userID              UserID
		accountID           AccountID
		accepted            bool
		roleName            RoleName
		expectedPermissions map[AccountID]AccountPermissions
		err                 error
	}{
		{
			name:      "Should add permissions for an App if the user doesn't have any for that App and they have accepted their invite",
			userID:    "user_1",
			accountID: "account_2",
			accepted:  true,
			roleName:  RoleAdmin,
			expectedPermissions: map[AccountID]AccountPermissions{
				"account_1": {
					Accepted:    true,
					RoleName:    RoleOwner,
					Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint, PermDeleteEndpoint, PermTransferEndpoint},
				},
				"account_2": {
					Accepted:    true,
					RoleName:    RoleAdmin,
					Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint},
				},
			},
		},
		{
			name:      "Should add permissions for an App if the user doesn't have any for that App and they have not accepted their invite",
			userID:    "user_3",
			accountID: "account_3",
			accepted:  false,
			roleName:  RoleAdmin,
			expectedPermissions: map[AccountID]AccountPermissions{
				"account_2": {
					Accepted:    true,
					RoleName:    RoleOwner,
					Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint, PermDeleteEndpoint, PermTransferEndpoint},
				},
				"account_3": {
					Accepted:    false,
					RoleName:    RoleAdmin,
					Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint},
				},
			},
		},
		{
			name:      "Should update permissions for an App if the user already has them for that App",
			userID:    "user_1",
			accountID: "account_2",
			roleName:  RoleMember,
			accepted:  true,
			expectedPermissions: map[AccountID]AccountPermissions{
				"account_1": {
					Accepted:    true,
					RoleName:    RoleOwner,
					Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint, PermDeleteEndpoint, PermTransferEndpoint},
				},
				"account_2": {
					Accepted:    true,
					RoleName:    RoleMember,
					Permissions: []Permissions{PermReadEndpoint},
				},
			},
		},
		{
			name:      "Should fail if passed an empty account ID",
			userID:    "user_1",
			accountID: "",
			err:       errAccountIDIsEmpty,
		},
		{
			name:      "Should fail if passed an invalid role",
			userID:    "user_1",
			accountID: "account_1",
			roleName:  RoleName("not_real"),
			err:       errInvalidRole,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPermissions := getUserPermissions(test.userID)

			updatedUserPermissions, err := userPermissions.UpsertPermissions(test.accountID, test.roleName, test.accepted)
			c.Equal(test.err, err)
			c.Equal(test.expectedPermissions, updatedUserPermissions)
		})
	}
}

func Test_UserPermissions_DeletePermissions(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name                string
		userID              UserID
		accountID           AccountID
		roleName            RoleName
		expectedPermissions map[AccountID]AccountPermissions
		err                 error
	}{
		{
			name:      "Should delete UserPermissions for a given user and App",
			userID:    "user_3",
			accountID: "account_1",
			roleName:  RoleMember,
			expectedPermissions: map[AccountID]AccountPermissions{
				"account_2": {
					Accepted:    true,
					RoleName:    RoleOwner,
					Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint, PermDeleteEndpoint, PermTransferEndpoint},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPermissions := getUserPermissions(test.userID)

			updatedUserPermissions, err := userPermissions.UpsertPermissions(test.accountID, test.roleName, true)
			c.Equal(test.err, err)
			c.Len(updatedUserPermissions, 2)

			deletedUserPermissions := userPermissions.DeletePermissions(test.accountID)
			c.Equal(test.expectedPermissions, deletedUserPermissions)
		})
	}
}

func Test_UserPermissions_HasPermission_Read(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name              string
		userID            UserID
		accountID         AccountID
		hasReadPermission bool
	}{
		{
			name:              "Should return a boolean indicating whether a given user has read permission for a given account",
			userID:            "user_1",
			accountID:         "account_1",
			hasReadPermission: true,
		},
		{
			name:              "Should return a boolean indicating whether a given user has read permission for a given account",
			userID:            "user_1",
			accountID:         "account_89",
			hasReadPermission: false,
		},
		{
			name:              "Should return a boolean indicating whether a given user has read permission for a given account",
			userID:            "user_2",
			accountID:         "account_2",
			hasReadPermission: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPermissions := getUserPermissions(test.userID)

			hasReadPermission := userPermissions.HasPermission(test.accountID, PermReadEndpoint)
			c.Equal(test.hasReadPermission, hasReadPermission)
		})
	}
}

func Test_UserPermissions_HasPermission_Write(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name               string
		userID             UserID
		accountID          AccountID
		hasWritePermission bool
	}{
		{
			name:               "Should return a boolean indicating whether a given user has write permission for a given account",
			userID:             "user_1",
			accountID:          "account_1",
			hasWritePermission: true,
		},
		{
			name:               "Should return a boolean indicating whether a given user has write permission for a given account",
			userID:             "user_1",
			accountID:          "account_89",
			hasWritePermission: false,
		},
		{
			name:               "Should return a boolean indicating whether a given user has write permission for a given account",
			userID:             "user_3",
			accountID:          "account_2",
			hasWritePermission: true,
		},
		{
			name:               "Should return a boolean indicating whether a given user has write permission for a given account",
			userID:             "user_3",
			accountID:          "account_3",
			hasWritePermission: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPermissions := getUserPermissions(test.userID)

			hasWritePermission := userPermissions.HasPermission(test.accountID, PermWriteEndpoint)
			c.Equal(test.hasWritePermission, hasWritePermission)
		})
	}
}

func Test_UserPermissions_HasPermission_Delete(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name                string
		userID              UserID
		accountID           AccountID
		hasDeletePermission bool
	}{
		{
			name:                "Should return a boolean indicating whether a given user has delete permission for a given account",
			userID:              "user_1",
			accountID:           "account_1",
			hasDeletePermission: true,
		},
		{
			name:                "Should return a boolean indicating whether a given user has delete permission for a given account",
			userID:              "user_1",
			accountID:           "account_89",
			hasDeletePermission: false,
		},
		{
			name:                "Should return a boolean indicating whether a given user has delete permission for a given account",
			userID:              "user_3",
			accountID:           "account_2",
			hasDeletePermission: true,
		},
		{
			name:                "Should return a boolean indicating whether a given user has delete permission for a given account",
			userID:              "user_3",
			accountID:           "account_3",
			hasDeletePermission: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPermissions := getUserPermissions(test.userID)

			hasDeletePermission := userPermissions.HasPermission(test.accountID, PermDeleteEndpoint)
			c.Equal(test.hasDeletePermission, hasDeletePermission)
		})
	}
}

func Test_UserPermissions_HasPermission_Transfer(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name                  string
		userID                UserID
		accountID             AccountID
		hasTransferPermission bool
	}{
		{
			name:                  "Should return a boolean indicating whether a given user has transfer permission for a given account",
			userID:                "user_1",
			accountID:             "account_1",
			hasTransferPermission: true,
		},
		{
			name:                  "Should return a boolean indicating whether a given user has transfer permission for a given account",
			userID:                "user_1",
			accountID:             "account_89",
			hasTransferPermission: false,
		},
		{
			name:                  "Should return a boolean indicating whether a given user has transfer permission for a given account",
			userID:                "user_3",
			accountID:             "account_2",
			hasTransferPermission: true,
		},
		{
			name:                  "Should return a boolean indicating whether a given user has transfer permission for a given account",
			userID:                "user_3",
			accountID:             "account_3",
			hasTransferPermission: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPermissions := getUserPermissions(test.userID)

			hasTransferPermission := userPermissions.HasPermission(test.accountID, PermTransferEndpoint)
			c.Equal(test.hasTransferPermission, hasTransferPermission)
		})
	}
}
