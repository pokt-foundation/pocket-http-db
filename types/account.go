package types

import (
	"errors"
	"time"
)

var errNoOwner error = errors.New("account does not have an owner")

/* Account Struct Definition and Methods */
type (
	// Account represents a single account for a single application in the Portal
	Account struct {
		ID                     AccountID                 `json:"id"`
		Name                   string                    `json:"name"`
		IconURL                string                    `json:"iconURL"`
		PlanType               PayPlanType               `json:"planType"`
		PartnerChainIDs        map[RelayChainID]struct{} `json:"partnerBlockchainIDs"`
		PartnerThroughputLimit int32                     `json:"partnerThroughputLimit"`
		PartnerAppLimit        int32                     `json:"partnerAppLimit"`
		Integrations           AccountIntegrations       `json:"integrations"`
		CreatedAt              time.Time                 `json:"createdAt"`
		UpdatedAt              time.Time                 `json:"updatedAt"`
		Deleted                bool                      `json:"deleted"`

		// PortalApps and Plan are set in cache.go
		PortalApps map[PortalAppID]*PortalApp `json:"portalApps"`
		Plan       *Plan                      `json:"payPlan"`

		// Users assigned when account is fetched in cache.go
		Users map[UserID]AccountUser `json:"users"`

		// DEPRECATED
		AccountUsers map[UserID]AccountUserAccess `json:"accountUsers"`
	}

	AccountUser struct {
		AccountID AccountID `json:"accountID,omitempty"` // used for listener
		ID        UserID    `json:"id"`
		Email     Email     `json:"email"`
		RoleName  RoleName  `json:"roleName"`
		Accepted  bool      `json:"accepted"`
	}

	// DEPRECATED - AccountUser represents a single Portal user for a single Account
	AccountUserAccess struct {
		AccountID          AccountID                `json:"id,omitempty"` // used for listener
		Owner              bool                     `json:"owner"`
		UserID             UserID                   `json:"userID"`
		Email              Email                    `json:"email"`
		PortalAppRoles     map[PortalAppID]RoleName `json:"portalApplicationRoles"`
		PortalAppsAccepted map[PortalAppID]bool     `json:"portalApplicationsAccepted"`
	}

	// AccountUser represents fields used for integrations with other platforms
	AccountIntegrations struct {
		AccountID          AccountID `json:"id,omitempty"` // used for listener
		CovalentAPIKeyFree string    `json:"covalentAPIKeyFree"`
		CovalentAPIKeyPaid string    `json:"covalentAPIKeyPaid"`
	}

	// CreateAccountUserAccess contains all fields required to create a new Account User
	CreateAccountUserAccess struct {
		AccountID   AccountID   `json:"accountID"`
		PortalAppID PortalAppID `json:"portalAppID"`
		Email       Email       `json:"email"`
		RoleName    RoleName    `json:"roleName"`
	}

	// UpdateAccount contains all fields required to update an Account
	UpdateAccount struct {
		AccountID AccountID   `json:"accountID"`
		Name      string      `json:"name"`
		IconURL   string      `json:"iconURL"`
		PlanType  PayPlanType `json:"planType"`
	}

	// UpdateAccountUserRole contains all fields required to update an Account User's Role
	UpdateAccountUserRole struct {
		PortalAppID PortalAppID `json:"portalAppID"`
		UserID      UserID      `json:"userID"`
		AccountID   AccountID   `json:"accountID"`
		RoleName    RoleName    `json:"roleName"`
	}

	// UpdateAccountUserRole contains all fields required to accept an Account User
	UpdateAcceptAccountUser struct {
		DeclinedInvite   bool           `json:"declined"`
		PortalAppID      PortalAppID    `json:"portalAppID"`
		UserID           UserID         `json:"userID"`
		AuthProviderType AuthType       `json:"type"`
		ProviderUserID   ProviderUserID `json:"providerUserID"`
	}

	// UpdateRemoveAccountUser contains all fields required to remove an Account User's Role
	UpdateRemoveAccountUser struct {
		UserID      UserID      `json:"userID"`
		PortalAppID PortalAppID `json:"portalAppID"`
		AccountID   AccountID   `json:"accountID"`
	}
)

// GetOwner returns the Account OWNER
func (a *Account) GetOwner() (AccountUser, error) {
	for _, user := range a.Users {
		if user.RoleName == RoleOwner {
			return user, nil
		}
	}
	return AccountUser{}, errNoOwner
}

// GetOwnerID returns the Account OWNER's ID
func (a *Account) GetOwnerID() (UserID, error) {
	for _, user := range a.Users {
		if user.RoleName == RoleOwner {
			return user.ID, nil
		}
	}
	return UserID(""), errNoOwner
}

func (a *Account) Table() Table {
	return TableAccounts
}

func (a *AccountUser) Table() Table {
	return TableAccountUserAccess
}

func (s *AccountIntegrations) Table() Table {
	return TableAccountIntegrations
}

// AssignAccountUsers assigns all users for the account
// Called in the PHD cache when an account is fetched by any account endpoints
func (a *Account) AssignAccountUsers(usersForAccountID map[UserID]*User) {
	if a.Users == nil {
		a.Users = make(map[UserID]AccountUser)
	}

	for _, user := range usersForAccountID {
		a.Users[user.ID] = user.ToAccountUser(a.ID)
	}
}
