package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var testUserPermissions = map[UserID]User{
	"user_1": {
		ID: "user_1",
		Permissions: map[PortalAppID]PortalAppPermissions{
			"test_app_1": {
				AccountID:   "account_1",
				RoleName:    RoleOwner,
				Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint, PermDeleteEndpoint, PermTransferEndpoint},
			},
		},
	},
	"user_2": {
		ID: "user_2",
		Permissions: map[PortalAppID]PortalAppPermissions{
			"test_app_1": {
				AccountID:   "account_1",
				RoleName:    RoleAdmin,
				Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint},
			},
			"test_app_2": {
				AccountID:   "account_2",
				RoleName:    RoleMember,
				Permissions: []Permissions{PermReadEndpoint},
			},
		},
	},
	"user_3": {
		ID: "user_3",
		Permissions: map[PortalAppID]PortalAppPermissions{
			"test_app_3": {
				AccountID:   "account_2",
				RoleName:    RoleOwner,
				Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint, PermDeleteEndpoint, PermTransferEndpoint},
			},
		},
	},
}

func Test_UserPermissions_GetPortalAppRole(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name        string
		userID      UserID
		portalAppID PortalAppID
		roleName    RoleName
	}{
		{
			name:        "Should return the role for a given user and portal application ID",
			userID:      "user_1",
			portalAppID: "test_app_1",
			roleName:    RoleOwner,
		},
		{
			name:        "Should return the role for a given user and portal application ID",
			userID:      "user_2",
			portalAppID: "test_app_2",
			roleName:    RoleMember,
		},
		{
			name:        "Should return an empty string if user does not have a role for a portal application",
			userID:      "user_1",
			portalAppID: "test_app_77",
			roleName:    "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPermissions := testUserPermissions[test.userID]

			roleName := userPermissions.GetPortalAppRole(test.portalAppID)
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
		portalAppID         PortalAppID
		roleName            RoleName
		expectedPermissions map[PortalAppID]PortalAppPermissions
		err                 error
	}{
		{
			name:        "Should add permissions for an App if the user doesn't have any for that App",
			userID:      "user_1",
			accountID:   "account_2",
			portalAppID: "test_app_4",
			roleName:    RoleAdmin,
			expectedPermissions: map[PortalAppID]PortalAppPermissions{
				"test_app_1": {
					AccountID:   "account_1",
					RoleName:    RoleOwner,
					Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint, PermDeleteEndpoint, PermTransferEndpoint},
				},
				"test_app_4": {
					AccountID:   "account_2",
					RoleName:    RoleAdmin,
					Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint},
				},
			},
		},
		{
			name:        "Should update permissions for an App if the user already has them for that App",
			userID:      "user_1",
			accountID:   "account_2",
			portalAppID: "test_app_4",
			roleName:    RoleMember,
			expectedPermissions: map[PortalAppID]PortalAppPermissions{
				"test_app_1": {
					AccountID:   "account_1",
					RoleName:    RoleOwner,
					Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint, PermDeleteEndpoint, PermTransferEndpoint},
				},
				"test_app_4": {
					AccountID:   "account_2",
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
			name:        "Should fail if passed an empty portal app ID",
			userID:      "user_1",
			accountID:   "account_1",
			portalAppID: "",
			err:         errPortalAppIDIsEmpty,
		},
		{
			name:        "Should fail if passed an invalid role",
			userID:      "user_1",
			accountID:   "account_1",
			portalAppID: "test_app_4",
			roleName:    RoleName("not_real"),
			err:         errInvalidRole,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPermissions := testUserPermissions[test.userID]

			updatedUserPermissions, err := userPermissions.UpsertPermissions(test.portalAppID, test.accountID, test.roleName)
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
		portalAppID         PortalAppID
		roleName            RoleName
		expectedPermissions map[PortalAppID]PortalAppPermissions
		err                 error
	}{
		{
			name:        "Should delete UserPermissions for a given user and App",
			userID:      "user_3",
			accountID:   "account_1",
			portalAppID: "test_app_5",
			roleName:    RoleMember,
			expectedPermissions: map[PortalAppID]PortalAppPermissions{
				"test_app_3": {
					AccountID:   "account_2",
					RoleName:    RoleOwner,
					Permissions: []Permissions{PermReadEndpoint, PermWriteEndpoint, PermDeleteEndpoint, PermTransferEndpoint},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPermissions := testUserPermissions[test.userID]

			updatedUserPermissions, err := userPermissions.UpsertPermissions(test.portalAppID, test.accountID, test.roleName)
			c.Equal(test.err, err)
			c.Len(updatedUserPermissions, 2)

			deletedUserPermissions := userPermissions.DeletePermissions(test.portalAppID)
			c.Equal(test.expectedPermissions, deletedUserPermissions)
		})
	}
}

func Test_UserPermissions_HasPermission_Read(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name              string
		userID            UserID
		portalAppID       PortalAppID
		hasReadPermission bool
	}{
		{
			name:              "Should return a boolean indicating whether a given user has read permission for a given portal application",
			userID:            "user_1",
			portalAppID:       "test_app_1",
			hasReadPermission: true,
		},
		{
			name:              "Should return a boolean indicating whether a given user has read permission for a given portal application",
			userID:            "user_1",
			portalAppID:       "test_app_89",
			hasReadPermission: false,
		},
		{
			name:              "Should return a boolean indicating whether a given user has read permission for a given portal application",
			userID:            "user_2",
			portalAppID:       "test_app_2",
			hasReadPermission: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPermissions := testUserPermissions[test.userID]

			hasReadPermission := userPermissions.HasPermission(test.portalAppID, PermReadEndpoint)
			c.Equal(test.hasReadPermission, hasReadPermission)
		})
	}
}

func Test_UserPermissions_HasPermission_Write(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name               string
		userID             UserID
		portalAppID        PortalAppID
		hasWritePermission bool
	}{
		{
			name:               "Should return a boolean indicating whether a given user has write permission for a given portal application",
			userID:             "user_1",
			portalAppID:        "test_app_1",
			hasWritePermission: true,
		},
		{
			name:               "Should return a boolean indicating whether a given user has write permission for a given portal application",
			userID:             "user_1",
			portalAppID:        "test_app_89",
			hasWritePermission: false,
		},
		{
			name:               "Should return a boolean indicating whether a given user has write permission for a given portal application",
			userID:             "user_3",
			portalAppID:        "test_app_3",
			hasWritePermission: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPermissions := testUserPermissions[test.userID]

			hasWritePermission := userPermissions.HasPermission(test.portalAppID, PermWriteEndpoint)
			c.Equal(test.hasWritePermission, hasWritePermission)
		})
	}
}

func Test_UserPermissions_HasPermission_Delete(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name                string
		userID              UserID
		portalAppID         PortalAppID
		hasDeletePermission bool
	}{
		{
			name:                "Should return a boolean indicating whether a given user has delete permission for a given portal application",
			userID:              "user_1",
			portalAppID:         "test_app_1",
			hasDeletePermission: true,
		},
		{
			name:                "Should return a boolean indicating whether a given user has delete permission for a given portal application",
			userID:              "user_1",
			portalAppID:         "test_app_89",
			hasDeletePermission: false,
		},
		{
			name:                "Should return a boolean indicating whether a given user has delete permission for a given portal application",
			userID:              "user_3",
			portalAppID:         "test_app_3",
			hasDeletePermission: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPermissions := testUserPermissions[test.userID]

			hasReadPermission := userPermissions.HasPermission(test.portalAppID, PermDeleteEndpoint)
			c.Equal(test.hasDeletePermission, hasReadPermission)
		})
	}
}

func Test_UserPermissions_HasPermission_Transfer(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name                string
		userID              UserID
		portalAppID         PortalAppID
		hasDeletePermission bool
	}{
		{
			name:                "Should return a boolean indicating whether a given user has transfer permission for a given portal application",
			userID:              "user_1",
			portalAppID:         "test_app_1",
			hasDeletePermission: true,
		},
		{
			name:                "Should return a boolean indicating whether a given user has transfer permission for a given portal application",
			userID:              "user_1",
			portalAppID:         "test_app_89",
			hasDeletePermission: false,
		},
		{
			name:                "Should return a boolean indicating whether a given user has transfer permission for a given portal application",
			userID:              "user_3",
			portalAppID:         "test_app_3",
			hasDeletePermission: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userPermissions := testUserPermissions[test.userID]

			hasReadPermission := userPermissions.HasPermission(test.portalAppID, PermTransferEndpoint)
			c.Equal(test.hasDeletePermission, hasReadPermission)
		})
	}
}

func Test_HasAccountPermission(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name          string
		userID        UserID
		accountID     AccountID
		permission    Permissions
		hasPermission bool
	}{
		// Test cases for user_1
		{
			name:          "Should return true for user_1 with RoleOwner on account_1 for PermReadEndpoint",
			userID:        "user_1",
			accountID:     "account_1",
			permission:    PermReadEndpoint,
			hasPermission: true,
		},
		{
			name:          "Should return true for user_1 with RoleOwner on account_1 for PermWriteEndpoint",
			userID:        "user_1",
			accountID:     "account_1",
			permission:    PermWriteEndpoint,
			hasPermission: true,
		},
		{
			name:          "Should return true for user_1 with RoleOwner on account_1 for PermDeleteEndpoint",
			userID:        "user_1",
			accountID:     "account_1",
			permission:    PermDeleteEndpoint,
			hasPermission: true,
		},
		{
			name:          "Should return true for user_1 with RoleOwner on account_1 for PermTransferEndpoint",
			userID:        "user_1",
			accountID:     "account_1",
			permission:    PermTransferEndpoint,
			hasPermission: true,
		},
		// Test cases for user_2
		{
			name:          "Should return true for user_2 with RoleAdmin on account_1 for PermReadEndpoint",
			userID:        "user_2",
			accountID:     "account_1",
			permission:    PermReadEndpoint,
			hasPermission: true,
		},
		{
			name:          "Should return true for user_2 with RoleAdmin on account_1 for PermWriteEndpoint",
			userID:        "user_2",
			accountID:     "account_1",
			permission:    PermWriteEndpoint,
			hasPermission: true,
		},
		{
			name:          "Should return false for user_2 with RoleAdmin on account_1 for PermDeleteEndpoint",
			userID:        "user_2",
			accountID:     "account_1",
			permission:    PermDeleteEndpoint,
			hasPermission: false,
		},
		{
			name:          "Should return false for user_2 with RoleAdmin on account_1 for PermTransferEndpoint",
			userID:        "user_2",
			accountID:     "account_1",
			permission:    PermTransferEndpoint,
			hasPermission: false,
		},
		// Test cases for user_3
		{
			name:          "Should return true for user_3 with RoleOwner on account_2 for PermReadEndpoint",
			userID:        "user_3",
			accountID:     "account_2",
			permission:    PermReadEndpoint,
			hasPermission: true,
		},
		{
			name:          "Should return true for user_3 with RoleOwner on account_2 for PermWriteEndpoint",
			userID:        "user_3",
			accountID:     "account_2",
			permission:    PermWriteEndpoint,
			hasPermission: true,
		},
		{
			name:          "Should return true for user_3 with RoleOwner on account_2 for PermDeleteEndpoint",
			userID:        "user_3",
			accountID:     "account_2",
			permission:    PermDeleteEndpoint,
			hasPermission: true,
		},
		{
			name:          "Should return true for user_3 with RoleOwner on account_2 for PermTransferEndpoint",
			userID:        "user_3",
			accountID:     "account_2",
			permission:    PermTransferEndpoint,
			hasPermission: true,
		},
		// Test cases for user_2 with RoleMember on account_2
		{
			name:          "Should return true for user_2 with RoleMember on account_2 for PermReadEndpoint",
			userID:        "user_2",
			accountID:     "account_2",
			permission:    PermReadEndpoint,
			hasPermission: true,
		},
		{
			name:          "Should return false for user_2 with RoleMember on account_2 for PermWriteEndpoint",
			userID:        "user_2",
			accountID:     "account_2",
			permission:    PermWriteEndpoint,
			hasPermission: false,
		},
		{
			name:          "Should return false for user_2 with RoleMember on account_2 for PermDeleteEndpoint",
			userID:        "user_2",
			accountID:     "account_2",
			permission:    PermDeleteEndpoint,
			hasPermission: false,
		},
		{
			name:          "Should return false for user_2 with RoleMember on account_2 for PermTransferEndpoint",
			userID:        "user_2",
			accountID:     "account_2",
			permission:    PermTransferEndpoint,
			hasPermission: false,
		},
		// Test case for non-existent user
		{
			name:          "Should return false for non-existent user with account_1 for PermReadEndpoint",
			userID:        "user_4",
			accountID:     "account_1",
			permission:    PermReadEndpoint,
			hasPermission: false,
		},
		{
			name:          "Should return false for user_2 with non-existent account for PermReadEndpoint",
			userID:        "user_2",
			accountID:     "account_5",
			permission:    PermReadEndpoint,
			hasPermission: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user := testUserPermissions[test.userID]
			hasPermission := user.HasAccountPermission(test.accountID, test.permission)
			c.Equal(test.hasPermission, hasPermission)
		})
	}
}
