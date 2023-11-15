package types

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	errAccountIDIsEmpty error = errors.New("account ID is empty")
	errInvalidRole      error = errors.New("invalid role provided")
)

/* Enums */
type (
	AuthProvider   string
	AuthType       string
	Permissions    string
	RoleName       string
	ProviderUserID string
)

const (
	AuthProviderAuth0 AuthProvider = "auth0"

	AuthTypeAuth0Github   AuthType = "auth0_github"
	AuthTypeAuth0Username AuthType = "auth0_username"

	PermReadEndpoint     Permissions = "read:endpoint"
	PermWriteEndpoint    Permissions = "write:endpoint"
	PermDeleteEndpoint   Permissions = "delete:endpoint"
	PermTransferEndpoint Permissions = "transfer:endpoint"

	RoleOwner  RoleName = "OWNER"
	RoleAdmin  RoleName = "ADMIN"
	RoleMember RoleName = "MEMBER"
)

// Currently only Auth0 auth types (`auth0` & `github`) supported.
// If new auth providers in addition to Auth0 are added in the future this
// method will need to be updated to parse the ID to determine the auth type.
func (pid ProviderUserID) AuthType() AuthType {
	prefix := strings.Split(string(pid), "|")[0]

	switch prefix {
	case "auth0":
		return AuthTypeAuth0Username
	case "github":
		return AuthTypeAuth0Github
	default:
		if prefix != "" {
			return AuthType(prefix)
		}
		return ""
	}
}

func (a AuthType) IsValid() bool {
	switch a {
	case AuthTypeAuth0Github, AuthTypeAuth0Username:
		return true
	default:
		return false
	}
}

func (a AuthType) IsFederated() bool {
	switch a {
	case AuthTypeAuth0Username:
		return false
	case AuthTypeAuth0Github:
		return true
	default:
		return false
	}
}

func (a AuthType) Provider() AuthProvider {
	switch a {
	case AuthTypeAuth0Username, AuthTypeAuth0Github:
		return AuthProviderAuth0
	default:
		return ""
	}
}

func (p Permissions) IsValid() bool {
	switch p {
	case PermReadEndpoint, PermWriteEndpoint, PermDeleteEndpoint, PermTransferEndpoint:
		return true
	default:
		return false
	}
}

func (r RoleName) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember:
		return true
	default:
		return false
	}
}

/* User Struct Definition and Methods */
type (
	// User represents a single Portal user
	User struct {
		ID               UserID                           `json:"id"`
		Email            Email                            `json:"email"`
		IconURL          string                           `json:"iconURL"`
		SignedUp         bool                             `json:"signedUp"`
		UpdatesProduct   bool                             `json:"updatesProduct"`
		UpdatesMarketing bool                             `json:"updatesMarketing"`
		BetaTester       bool                             `json:"betaTester"`
		Permissions      map[AccountID]AccountPermissions `json:"permissions"`
		AuthProviders    map[AuthType]UserAuthProvider    `json:"authProviders"`
		CreatedAt        time.Time                        `json:"createdAt"`
		UpdatedAt        time.Time                        `json:"updatedAt"`
	}

	// AccountPermissions stores user role and permissions for a given Account
	AccountPermissions struct {
		Accepted    bool          `json:"accepted"`
		RoleName    RoleName      `json:"roleName"`
		Permissions []Permissions `json:"permissions"`
	}

	// UserAuthProvider represents a single auth provider for a user (eg. Auth0)
	UserAuthProvider struct {
		UserID         UserID         `json:"userID,omitempty"`
		ProviderUserID ProviderUserID `json:"providerUserID"`
		Type           AuthType       `json:"type"`
		Provider       AuthProvider   `json:"provider"`
		Federated      bool           `json:"federated"`
	}

	CreateUser struct {
		Email          Email          `json:"email"`
		ProviderUserID ProviderUserID `json:"providerUserID"`
	}

	CreateUserResponse struct {
		User      User      `json:"user"`
		AccountID AccountID `json:"accountID"`
	}

	// UpdateUser contains all fields required to update a User
	UpdateUser struct {
		ID               UserID  `json:"id"`
		IconURL          *string `json:"iconURL"`
		UpdatesProduct   *bool   `json:"updatesProduct"`
		UpdatesMarketing *bool   `json:"updatesMarketing"`
		BetaTester       *bool   `json:"betaTester"`
	}
)

/* UserPermissions Struct Definition and Methods */
var permissionsList = map[RoleName][]Permissions{
	RoleOwner:  {PermReadEndpoint, PermWriteEndpoint, PermDeleteEndpoint, PermTransferEndpoint},
	RoleAdmin:  {PermReadEndpoint, PermWriteEndpoint},
	RoleMember: {PermReadEndpoint},
}

func (u *User) ToAccountUser(accountID AccountID) AccountUser {
	return AccountUser{
		ID:       u.ID,
		Email:    u.Email,
		RoleName: u.GetAccountRole(accountID),
		Accepted: u.Permissions[accountID].Accepted,
	}
}

func (u *User) ArePermissionsEmpty() bool {
	if u == nil || u.Permissions == nil || len(u.Permissions) == 0 {
		return true
	}
	return false
}

func (u *User) GetAccountRole(accountID AccountID) RoleName {
	app, ok := u.Permissions[accountID]
	if !ok {
		return RoleName("")
	}

	return app.RoleName
}

func (u *User) UpsertPermissions(accountID AccountID, role RoleName, accepted bool) (map[AccountID]AccountPermissions, error) {
	if accountID == "" {
		return nil, errAccountIDIsEmpty
	}
	if !role.IsValid() {
		return nil, errInvalidRole
	}

	u.Permissions[accountID] = AccountPermissions{
		Accepted:    accepted,
		RoleName:    role,
		Permissions: permissionsList[role],
	}

	return u.Permissions, nil
}

func (u *User) DeletePermissions(accountID AccountID) map[AccountID]AccountPermissions {
	delete(u.Permissions, accountID)

	return u.Permissions
}

func (u *User) HasPermission(accountID AccountID, permission Permissions) bool {
	accountPermissions, ok := u.Permissions[accountID]
	if !ok || !accountPermissions.Accepted {
		return false
	}

	for _, accountPermission := range accountPermissions.Permissions {
		if accountPermission == permission {
			return true
		}
	}

	return false
}

func (p *Permissions) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*p = Permissions(s)
	case string:
		*p = Permissions(s)
	default:
		return fmt.Errorf("unsupported scan type for Permissions: %T", src)
	}
	return nil
}

func (u *User) Table() Table {
	return TableUsers
}

func (u *UserAuthProvider) Table() Table {
	return TableUserAuthProviders
}
