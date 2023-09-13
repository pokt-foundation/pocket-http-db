package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1Types "github.com/pokt-foundation/portal-db/types"
	"github.com/pokt-foundation/portal-db/v2/driver"
	legacyadapters "github.com/pokt-foundation/portal-db/v2/legacy-adapters"
	"github.com/pokt-foundation/portal-db/v2/types"
	v2Types "github.com/pokt-foundation/portal-db/v2/types"
)

var (
	errAccountNotFound           = errors.New("account not found for ID %s")
	ErrPortalAppNotFound         = errors.New("portal app not found for load balancer ID %s")
	errUserNotFound              = errors.New("user not found for provider user ID %s")
	errPlanNotFound              = errors.New("plan not found for type %s")
	errPortalAppsNotFoundForUser = errors.New("portal app not found for user ID %s")
	errBlockchainNotFound        = errors.New("blockchain not found for chain ID %s")
	errWritingPortalApp          = errors.New("error writing portal app %w")
	errWritingAccount            = errors.New("error writing account %w")
)

/* ---------- Read Methods ---------- */

// GetLegacyApplications returns all V2 PortalApps in the cache converted to V1 Applications
func (c *Cache) GetLegacyApplications(options Options) []v1Types.Application {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var applications []v1Types.Application

	portalApps := c.portalApps

	for _, portalApp := range portalApps {
		if portalApp != nil {
			var appOwnerID string
			if account, ok := c.accounts[portalApp.AccountID]; ok && account != nil {
				if userID, err := account.GetOwnerID(); err == nil {
					appOwnerID = string(userID)
				}
			}

			legacyApps := legacyadapters.ConvertPortalAppToLegacyApplications(*portalApp, appOwnerID)

			var apps []v1Types.Application
			for _, app := range legacyApps {
				apps = append(apps, *app)
			}

			applications = append(applications, apps...)
		}
	}

	gigastakeApps := c.gigastakeApps
	for _, gigastakeApp := range gigastakeApps {
		if gigastakeApp != nil {
			legacyApp := legacyadapters.ConvertGigastakeAppToLegacyApplication(gigastakeApp)

			if legacyApp != nil {
				applications = append(applications, *legacyApp)
			}
		}
	}

	return applications
}

// GetLegacyLoadBalancersByUserID returns all V1 LoadBalancers for a User ID
func (c *Cache) GetLegacyLoadBalancersByUserID(ctx context.Context, userID string, roleNameFilter v1Types.RoleName) ([]v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	loadBalancers := []v1Types.LoadBalancer{}

	portalUser := c.GetUserByProviderID(v2Types.ProviderUserID(userID))
	if portalUser.ID == "" {
		return nil, fmt.Errorf(errUserNotFound.Error(), userID)
	}

	accounts := c.GetAccountsByUserID(portalUser.ID)

	filteredAccounts := make(map[types.AccountID]v2Types.Account)
	switch {
	case roleNameFilter == v1Types.RoleName(""):
		for _, account := range accounts {
			if !account.Deleted {
				filteredAccounts[account.ID] = account
			}
		}

	case roleNameFilter != v1Types.RoleName(""):
		for _, account := range accounts {
			if !account.Deleted {
				if user, ok := account.Users[portalUser.ID]; ok {
					if user.Accepted {
						for _, roleName := range user.PortalAppRoles {
							if roleName == v2Types.RoleName(roleNameFilter) {
								if _, ok := filteredAccounts[account.ID]; ok {
									continue
								}
								filteredAccounts[account.ID] = account
							}
						}
					}
				}
			}
		}
	}

	for _, account := range filteredAccounts {
		for _, portalApp := range account.PortalApps {
			if portalApp != nil && !portalApp.Deleted {
				legacyLB := legacyadapters.ConvertPortalAppToLegacyLoadBalancer(*portalApp, account)

				loadBalancers = append(loadBalancers, legacyLB)
			}
		}
	}

	return loadBalancers, nil
}

// GetLegacyLoadBalancer returns one V1 LoadBalancer by ID
func (c *Cache) GetLegacyLoadBalancer(ctx context.Context, lbID string) (v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalApp := c.GetPortalApp(v2Types.PortalAppID(lbID))
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(ErrPortalAppNotFound.Error(), lbID)
	}

	account := c.GetAccount(portalApp.AccountID)
	if account.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errAccountNotFound.Error(), portalApp.AccountID)
	}

	return legacyadapters.ConvertPortalAppToLegacyLoadBalancer(portalApp, account), nil
}

// GetLegacyLoadBalancers returns all V1 LoadBalancers
func (c *Cache) GetLegacyLoadBalancers(ctx context.Context) []v1Types.LoadBalancer {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var lbs []v1Types.LoadBalancer

	for _, account := range c.accounts {
		if account != nil {
			for _, portalApp := range account.PortalApps {
				if portalApp != nil {
					lbs = append(lbs, legacyadapters.ConvertPortalAppToLegacyLoadBalancer(*portalApp, *account))
				}
			}
		}
	}

	for _, chain := range c.chains {
		if len(chain.GigastakeApps) > 0 {
			lbs = append(lbs, legacyadapters.ConvertChainToLegacyGigastakeLoadBalancer(*chain))
		}
	}

	return lbs
}

// GetLegacyApplicationsByUserID returns all V1 Applicatons for a User ID
func (c *Cache) GetLegacyApplicationsByUserID(ctx context.Context, userID string) ([]v1Types.Application, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	applications := []v1Types.Application{}

	portalUser := c.GetUserByProviderID(v2Types.ProviderUserID(userID))
	if portalUser.ID == "" {
		return nil, fmt.Errorf(errUserNotFound.Error(), userID)
	}

	accounts := c.GetAccountsByUserID(portalUser.ID)

	for _, account := range accounts {
		if !account.Deleted {
			for _, portalApp := range account.PortalApps {
				if portalApp != nil && !portalApp.Deleted {
					for _, aat := range portalApp.AATs {
						if userID, err := account.GetOwnerID(); err == nil {
							legacyApp := legacyadapters.ConvertPortalAppToLegacyApplication(*portalApp, string(userID), aat)

							applications = append(applications, legacyApp)
						}
					}
				}
			}
		}
	}

	if len(applications) == 0 {
		return applications, fmt.Errorf(errPortalAppsNotFoundForUser.Error(), portalUser.ID)
	}

	return applications, nil
}

// GetLegacyBlockchains returns all V1 Blockchains
func (c *Cache) GetLegacyBlockchains(ctx context.Context, includeInactive bool) []v1Types.Blockchain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	blockchains := []v1Types.Blockchain{}

	chains := c.GetChains(Options{IncludeDeleted: false, IncludeInactiveChains: includeInactive})

	for _, chain := range chains {
		blockchains = append(blockchains, legacyadapters.ConvertToLegacyBlockchain(chain))
	}

	return blockchains
}

// GetLegacyBlockchain returns one V1 Blockchain for a blockchain ID
func (c *Cache) GetLegacyBlockchain(ctx context.Context, chainID string) v1Types.Blockchain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	chain := c.GetChain(v2Types.RelayChainID(chainID), Options{})
	if chain.ID == "" {
		return v1Types.Blockchain{}
	}

	return legacyadapters.ConvertToLegacyBlockchain(chain)
}

// GetLegacyBlockchains returns all V1 Applicatons for a User ID
func (c *Cache) GetLegacyUserPermissionsByUserID(ctx context.Context, userID string) (v1Types.UserPermissions, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalUser := c.GetUserByProviderID(v2Types.ProviderUserID(userID))
	if portalUser.ID == "" {
		return v1Types.UserPermissions{}, fmt.Errorf(errUserNotFound.Error(), userID)
	}

	user := c.GetUserByID(portalUser.ID)

	legacyUserPermissions := v1Types.UserPermissions{
		UserID:        v1Types.UserID(userID),
		LoadBalancers: make(map[v1Types.LoadBalancerID]v1Types.LoadBalancerPermissions),
	}

	for portalAppID, portalAppPermissions := range user.Permissions {
		lbID := v1Types.LoadBalancerID(portalAppID)

		rolePermissions := []v1Types.PermissionsEnum{}
		for _, permission := range portalAppPermissions.Permissions {
			rolePermissions = append(rolePermissions, v1Types.PermissionsEnum(permission))
		}

		legacyUserPermissions.LoadBalancers[lbID] = v1Types.LoadBalancerPermissions{
			RoleName:    v1Types.RoleName(portalAppPermissions.RoleName),
			Permissions: rolePermissions,
		}
	}

	return legacyUserPermissions, nil
}

// GetLegacyPayPlas returns all V1 Applicatons for a User ID
func (c *Cache) GetLegacyPayPlan(ctx context.Context, payPlanType string) (v1Types.PayPlan, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	plan := c.GetPlan(v2Types.PayPlanType(payPlanType))
	if plan.Type == "" {
		return v1Types.PayPlan{}, fmt.Errorf(errPlanNotFound.Error(), payPlanType)
	}

	return legacyadapters.ConvertToLegacyPayPlan(plan), nil
}

// GetLegacyPayPlans returns all V1 Applicatons for a User ID
func (c *Cache) GetLegacyPayPlans(ctx context.Context) []v1Types.PayPlan {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	payPlans := []v1Types.PayPlan{}

	plans := c.GetPlans()

	for _, plan := range plans {
		payPlans = append(payPlans, legacyadapters.ConvertToLegacyPayPlan(plan))
	}

	return payPlans
}

// GetLegacyPendingLoadBalancers returns all pending load balancers for a given portal user ID.
func (c *Cache) GetLegacyPendingLoadBalancers(ctx context.Context, portalUserID string) ([]v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	loadBalancers := []v1Types.LoadBalancer{}

	userID := v2Types.UserID(portalUserID)

	accounts := c.GetAccountsByUserID(userID)

	pendingAccounts := []v2Types.Account{}
	for _, account := range accounts {
		for _, user := range account.Users {
			if !user.Accepted && user.UserID == userID {
				pendingAccounts = append(pendingAccounts, account)
			}
		}
	}

	for _, account := range pendingAccounts {
		for _, portalApp := range account.PortalApps {
			if portalApp != nil {
				loadBalancers = append(loadBalancers, legacyadapters.ConvertPortalAppToLegacyLoadBalancer(*portalApp, account))
			}
		}
	}

	return loadBalancers, nil
}

// GetLegacyLoadBalancersCount returns all pending load balancers for a given portal user ID.
func (c *Cache) GetLegacyLoadBalancersCount(ctx context.Context, portalUserID string) int {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	count := 0

	userID := v1Types.UserID(portalUserID)

	accounts := c.GetAccountsByUserID(v2Types.UserID(portalUserID))
	for _, account := range accounts {
		if user, ok := account.Users[v2Types.UserID(userID)]; ok {
			if user.Accepted {
				for _, roleName := range user.PortalAppRoles {
					if roleName == v2Types.RoleOwner {
						count++
					}
				}
			}
		}
	}

	return count
}

/* ---------- Create Methods ---------- */

// CreateLegacyLoadBalancer returns one V2 PortalApp in the cache converted to a V1 Application
func (c *Cache) CreateLegacyLoadBalancer(ctx context.Context, writer driver.Writer, lb v1Types.LoadBalancer) (v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	// Get the current time to use as the created time for the account and portal app
	createdAt := time.Now()

	// Convert the V2 LoadBalancer object to a V2 portal app & AAT using the ConvertToV2PortalAppAndAAT function
	portalApp, aat := legacyadapters.ConvertToV2PortalAppAndAAT(lb)

	portalUser := c.GetUserByProviderID(v2Types.ProviderUserID(lb.UserID))
	if portalUser.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errUserNotFound.Error(), lb.UserID)
	}

	// Get the Account of the user creating the load balancer
	var account v2Types.Account

	accounts := c.GetAccountsByUserID(portalUser.ID)
	for _, userAccount := range accounts {
		if user, ok := userAccount.Users[portalUser.ID]; ok {
			for _, roleName := range user.PortalAppRoles {
				if roleName == v2Types.RoleOwner {
					account = userAccount
					break
				}
			}
		}
	}

	// If user does not have an Account, create a default one for them with FreetierV0 plan
	if account.ID == "" {
		createdAccount, err := writer.WriteAccount(ctx, portalUser.ID, v2Types.Account{PlanType: v2Types.FreetierV0}, createdAt)
		if err != nil {
			return v1Types.LoadBalancer{}, fmt.Errorf(errWritingAccount.Error(), err)
		}
		account = *createdAccount
	}

	account.PlanType = v2Types.FreetierV0
	freetierPlan, ok := c.plans[v2Types.FreetierV0]
	if !ok {
		return v1Types.LoadBalancer{}, fmt.Errorf("error in CreateLegacyLoadBalancer: FreetierV0 plan not found in cache")
	}
	account.Plan = freetierPlan

	portalApp.AccountID = account.ID
	portalApp.LegacyFields.PlanType = account.PlanType
	portalApp.LegacyFields.DailyLimit = int32(account.Plan.LegacyDailyLimit)

	// Write the portal app to the data store
	createdPortalApp, err := writer.WritePortalApp(ctx, portalApp, aat, createdAt)
	if err != nil {
		return v1Types.LoadBalancer{}, fmt.Errorf(errWritingPortalApp.Error(), err)
	}

	// Add the created portal app to the PortalApps map of the created account
	account.PortalApps = map[v2Types.PortalAppID]*v2Types.PortalApp{
		createdPortalApp.ID: createdPortalApp,
	}

	// Add the creator as the OWNER to the account before returning converted legacy load balancer
	if accountUsers, ok := account.Users[portalUser.ID]; ok {
		if accountUsers.PortalAppRoles == nil {
			accountUsers.PortalAppRoles = make(map[v2Types.PortalAppID]v2Types.RoleName)
		}
		accountUsers.PortalAppRoles[createdPortalApp.ID] = v2Types.RoleOwner
	}

	derefCreatedPortalApp := types.PortalApp{}
	if createdPortalApp != nil {
		derefCreatedPortalApp = *createdPortalApp
	}

	return legacyadapters.ConvertPortalAppToLegacyLoadBalancer(derefCreatedPortalApp, account), nil
}

// CreateLegacyLoadBalancerUser creates a single V2 AccountUserAccess row for a given account
func (c *Cache) CreateLegacyLoadBalancerUser(ctx context.Context, writer driver.Writer, lbID string, user v1Types.UserAccess) (v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	// Get the current time to use as the created time for the account and portal app
	createdAt := time.Now()

	portalAppID := v2Types.PortalAppID(lbID)

	portalApp := c.GetPortalApp(portalAppID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(ErrPortalAppNotFound.Error(), portalAppID)
	}
	account := c.GetAccount(portalApp.AccountID)
	if account.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errAccountNotFound.Error(), portalApp.AccountID)
	}

	// Convert the V2 LoadBalancer object to a V2 account and portal app using the ConvertToV2AccountAndPortalApp function
	accountUserAccess := legacyadapters.ConvertToV2AccountUserAccess(user, lbID, string(account.ID))
	createUser := v2Types.CreateAccountUserAccess{
		AccountID:   account.ID,
		PortalAppID: portalAppID,
		Email:       accountUserAccess.Email,
		RoleName:    accountUserAccess.PortalAppRoles[portalAppID],
	}

	userID, err := writer.WriteAccountUser(ctx, createUser, createdAt)
	if err != nil {
		return v1Types.LoadBalancer{}, err
	}

	accountUserAccess.UserID = userID
	account.Users[userID] = accountUserAccess

	return legacyadapters.ConvertPortalAppToLegacyLoadBalancer(portalApp, account), nil
}

// CreateLegacyLoadBalancerIntegration creates a single integration row for a given account
func (c *Cache) CreateLegacyLoadBalancerIntegration(ctx context.Context, writer driver.Writer, lbID string, integrations v2Types.AccountIntegrations) (v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	accountIntegrations, err := writer.UpsertAccountIntegration(ctx, integrations)
	if err != nil {
		return v1Types.LoadBalancer{}, err
	}

	return v1Types.LoadBalancer{
		Integrations: v1Types.AccountIntegrations{
			CovalentAPIKeyFree: accountIntegrations.CovalentAPIKeyFree,
		},
	}, nil
}

/* ---------- Update Methods ---------- */

// UpdateLegacyApplicationsFirstDateSurpassed updates the first date surpassed for a list of application IDs
func (c *Cache) UpdateLegacyApplicationsFirstDateSurpassed(ctx context.Context, writer driver.Writer, update v1Types.UpdateFirstDateSurpassed) ([]*v1Types.Application, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var portalAppsToUpdate []v2Types.PortalApp
	var appsToUpdate []*v1Types.Application

	if len(update.ApplicationIDs) == 0 {
		c.logger.Warn("no application IDs on input")
		return appsToUpdate, nil
	}

	v2Update := &v2Types.UpdateFirstDateSurpassed{
		FirstDateSurpassed: update.FirstDateSurpassed,
	}

	for _, appID := range update.ApplicationIDs {
		protocolAppID := v2Types.ProtocolAppID(appID)

		portalApp := c.GetPortalAppByProtocolAppID(protocolAppID)
		if portalApp.ID == "" {
			c.logger.Error(fmt.Sprintf("UpdateFirstDateSurpassed: protocol app ID %s not found", protocolAppID))
			continue
		}

		portalAppsToUpdate = append(portalAppsToUpdate, portalApp)
		v2Update.PortalAppIDs = append(v2Update.PortalAppIDs, portalApp.ID)
	}

	err := writer.UpdatePortalAppsFirstDateSurpassed(ctx, v2Update)
	if err != nil {
		return appsToUpdate, fmt.Errorf("UpdateFirstDateSurpassed failed: %W", err)
	}

	for _, portalApp := range portalAppsToUpdate {
		if account, ok := c.accounts[portalApp.AccountID]; ok {
			var userID string
			if ownerID, err := account.GetOwnerID(); err == nil {
				userID = string(ownerID)
			}

			appsToUpdate = append(appsToUpdate, legacyadapters.ConvertPortalAppToLegacyApplications(portalApp, userID)...)
		}
	}

	for _, app := range appsToUpdate {
		app.FirstDateSurpassed = update.FirstDateSurpassed
	}

	return appsToUpdate, nil
}

// UpdateLegacyLoadBalancer updates one V2 PortalApp in the cache from a legacy LoadBalancer
func (c *Cache) UpdateLegacyLoadBalancer(ctx context.Context, writer driver.Writer, lbID string, update v1Types.UpdateApplication) (v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalAppID := v2Types.PortalAppID(lbID)
	updatedAt := time.Now()

	portalApp := c.GetPortalApp(portalAppID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(ErrPortalAppNotFound.Error(), portalAppID)
	}
	account := c.GetAccount(portalApp.AccountID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errAccountNotFound.Error(), portalApp.AccountID)
	}

	updatedLB := legacyadapters.ConvertPortalAppToLegacyLoadBalancer(portalApp, account)

	switch update.Remove {
	case true:
		err := writer.SetPortalAppDeleted(ctx, portalAppID, updatedAt)
		if err != nil {
			return v1Types.LoadBalancer{}, err
		}

	default:
		updatePortalApp := legacyadapters.ConvertToV2UpdatePortalApp(update, lbID)

		err := writer.UpdatePortalApp(ctx, updatePortalApp, updatedAt)
		if err != nil {
			return v1Types.LoadBalancer{}, err
		}

		if update.Limit != nil {
			updateAccount := v2Types.UpdateAccount{
				AccountID: account.ID,
				PlanType:  v2Types.PayPlanType(update.Limit.PayPlan.Type),
			}

			_, err := writer.UpdateAccount(ctx, updateAccount, updatedAt)
			if err != nil {
				return v1Types.LoadBalancer{}, err
			}
		}

		app := updatedLB.Applications[0]

		if update.Name != "" {
			updatedLB.Name = update.Name
			app.Name = update.Name
		}
		if update.Limit != nil {
			if update.Limit.PayPlan.Type != v1Types.Enterprise {
				newPlan := c.plans[v2Types.PayPlanType(update.Limit.PayPlan.Type)]
				update.Limit.PayPlan.Limit = int(newPlan.LegacyDailyLimit)
			}
			app.Limit = *update.Limit
		}
		if update.GatewaySettings != nil {
			app.GatewaySettings = v1Types.GatewaySettings{
				SecretKey:            update.GatewaySettings.SecretKey,
				SecretKeyRequired:    pointerToBool(update.GatewaySettings.SecretKeyRequired),
				WhitelistOrigins:     update.GatewaySettings.WhitelistOrigins,
				WhitelistUserAgents:  update.GatewaySettings.WhitelistUserAgents,
				WhitelistContracts:   update.GatewaySettings.WhitelistContracts,
				WhitelistMethods:     update.GatewaySettings.WhitelistMethods,
				WhitelistBlockchains: update.GatewaySettings.WhitelistBlockchains,
			}
		}
		if update.NotificationSettings != nil {
			app.NotificationSettings = v1Types.NotificationSettings{
				SignedUp:      pointerToBool(update.NotificationSettings.SignedUp),
				Quarter:       pointerToBool(update.NotificationSettings.Quarter),
				Half:          pointerToBool(update.NotificationSettings.Half),
				ThreeQuarters: pointerToBool(update.NotificationSettings.ThreeQuarters),
				Full:          pointerToBool(update.NotificationSettings.Full),
			}
		}
	}

	updatedLB.UpdatedAt = updatedAt

	return updatedLB, nil
}

func pointerToBool(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

// UpdateLegacyUserRole updates the role for a single LB user
func (c *Cache) UpdateLegacyUserRole(ctx context.Context, writer driver.Writer, lbID string, updateUser v1Types.UpdateUserAccess) (v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalUserID := v2Types.UserID(updateUser.UserID)

	portalAppID := v2Types.PortalAppID(lbID)
	updatedAt := time.Now()

	portalApp := c.GetPortalApp(portalAppID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(ErrPortalAppNotFound.Error(), portalAppID)
	}
	account := c.GetAccount(portalApp.AccountID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errAccountNotFound.Error(), portalApp.AccountID)
	}

	updateAccountUser := legacyadapters.ConvertToV2UpdateAccountUserAccess(updateUser, lbID, string(portalUserID))
	updateAccountUser.AccountID = portalApp.AccountID

	err := writer.SetAccountUserRole(ctx, updateAccountUser, updatedAt)
	if err != nil {
		return v1Types.LoadBalancer{}, err
	}

	for userID, user := range account.Users {
		if user.UserID == "" || user.PortalAppRoles == nil {
			continue
		}

		if updateUser.RoleName == v1Types.RoleOwner && user.PortalAppRoles[portalAppID] == v2Types.RoleOwner {
			updatedUser := user
			updatedUser.PortalAppRoles[portalAppID] = v2Types.RoleAdmin
			account.Users[userID] = updatedUser
		}

		if userID == portalUserID {
			updatedUser := user
			updatedUser.PortalAppRoles[portalAppID] = v2Types.RoleName(updateUser.RoleName)
			account.Users[userID] = updatedUser
		}
	}

	return legacyadapters.ConvertPortalAppToLegacyLoadBalancer(portalApp, account), nil
}

// AcceptLegacyUser updates the role for a single LB user
func (c *Cache) AcceptLegacyUser(ctx context.Context, writer driver.Writer, lbID string, updateUser v1Types.UpdateUserAccess) (v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalUser := c.GetUserByProviderID(v2Types.ProviderUserID(updateUser.UserID))
	if portalUser.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errUserNotFound.Error(), updateUser.UserID)

	}

	portalAppID := v2Types.PortalAppID(lbID)
	updatedAt := time.Now()

	portalApp := c.GetPortalApp(portalAppID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(ErrPortalAppNotFound.Error(), portalAppID)
	}
	account := c.GetAccount(portalApp.AccountID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errAccountNotFound.Error(), portalApp.AccountID)
	}

	updateAccountUser := legacyadapters.ConvertToV2UpdateAcceptAccountUserAccess(updateUser, portalAppID, portalUser.ID)

	err := writer.UpdateAcceptAccountUser(ctx, updateAccountUser, updatedAt)
	if err != nil {
		return v1Types.LoadBalancer{}, err
	}

	for userID, user := range account.Users {
		if user.UserID == "" {
			continue
		}

		if userID == portalUser.ID {
			updatedUser := user
			updatedUser.Accepted = true
			account.Users[userID] = updatedUser
			break
		}
	}

	return legacyadapters.ConvertPortalAppToLegacyLoadBalancer(portalApp, account), nil
}

// ActivateLegacyBlockchain activates one blockchain
func (c *Cache) ActivateLegacyBlockchain(ctx context.Context, writer driver.Writer, blockchainID string, active bool) (bool, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	chainID := v2Types.RelayChainID(blockchainID)
	updatedAt := time.Now()

	chain := c.GetChain(chainID, Options{})
	if chain.ID == "" {
		return false, fmt.Errorf(errBlockchainNotFound.Error(), chainID)
	}

	activated, err := writer.SetChainActiveStatus(ctx, chainID, active, updatedAt)
	if err != nil {
		return false, fmt.Errorf("SetChainActiveStatus failed: %w", err)
	}

	return activated, nil
}

// RemoveLegacyAccountUser activates one blockchain
func (c *Cache) RemoveLegacyAccountUser(ctx context.Context, writer driver.Writer, userID string, lbID string) (v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalUserID := v2Types.UserID(userID)

	portalAppID := v2Types.PortalAppID(lbID)

	portalApp := c.GetPortalApp(portalAppID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(ErrPortalAppNotFound.Error(), portalAppID)
	}
	account := c.GetAccount(portalApp.AccountID)
	if account.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errAccountNotFound.Error(), portalApp.AccountID)
	}

	err := writer.RemoveAccountUser(ctx, portalUserID, portalAppID, account.ID)
	if err != nil {
		return v1Types.LoadBalancer{}, fmt.Errorf("RemoveAccountUser failed: %w", err)
	}

	filteredUsers := map[v2Types.UserID]v2Types.AccountUserAccess{}

	for userID, user := range account.Users {
		if user.UserID == "" {
			continue
		}

		if userID != portalUserID {
			filteredUsers[userID] = user
		}
	}
	account.Users = filteredUsers

	return legacyadapters.ConvertPortalAppToLegacyLoadBalancer(portalApp, account), nil
}
