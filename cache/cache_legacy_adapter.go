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
)

var (
	errAccountNotFound           = errors.New("account not found for ID %s")
	errAccountNotFoundForUser    = errors.New("account not found for User ID %s")
	ErrPortalAppNotFound         = errors.New("portal app not found for load balancer ID %s")
	errPortalAppNotFoundForApp   = errors.New("portal app not found for app ID %s")
	errUserNotFound              = errors.New("user not found for provider user ID %s")
	errPlanNotFound              = errors.New("plan not found for type %s")
	errPortalAppsNotFoundForUser = errors.New("portal app not found for user ID %s")
	errBlockchainNotFound        = errors.New("blockchain not found for chain ID %s")
	errWritingPortalApp          = errors.New("error writing portal app %w")
)

func (c *Cache) getUserIDAndAccount(accountID types.AccountID) (string, *types.Account, error) {
	account, ok := c.accounts[accountID]
	if ok {
		return account.LegacyUserID(), account, nil
	}

	return "", nil, fmt.Errorf(errAccountNotFound.Error(), accountID)
}

/* ---------- Read Methods ---------- */

// GetLegacyApplications returns all V2 PortalApps in the cache converted to V1 Applications
func (c *Cache) GetLegacyApplications(options Options) []v1Types.Application {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalApps := c.GetPortalApps(options)

	var applications []v1Types.Application

	for _, portalApp := range portalApps {
		userID, account, err := c.getUserIDAndAccount(portalApp.AccountID)
		if err != nil {
			c.log.Warn(fmt.Sprintf("error in GetLegacyApplications: %s", err.Error()))
			continue
		}

		legacyApps := legacyadapters.ConvertToLegacyApplications(portalApp, userID, account.PlanType, account.Plan.LegacyDailyLimit)

		var apps []v1Types.Application
		for _, app := range legacyApps {
			apps = append(apps, *app)
		}

		applications = append(applications, apps...)
	}

	return applications
}

// GetLegacyApplication returns one V2 PortalApp in the cache converted to a V1 Application
func (c *Cache) GetLegacyApplication(appID string) (v1Types.Application, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	PortalAppID := types.ProtocolAppID(appID)

	portalApp := c.GetPortalAppByProtocolID(PortalAppID)

	if portalApp.ID == "" {
		return v1Types.Application{}, fmt.Errorf(errPortalAppNotFoundForApp.Error(), appID)
	}

	_, account, err := c.getUserIDAndAccount(portalApp.AccountID)
	if err != nil {
		return v1Types.Application{}, err
	}

	application := legacyadapters.ConvertToLegacyApplication(portalApp, *account, PortalAppID)

	return application, nil
}

// GetLegacyLoadBalancersByUserID returns all V1 LoadBalancers for a User ID
func (c *Cache) GetLegacyLoadBalancersByUserID(ctx context.Context, userID string, roleNameFilter v1Types.RoleName) ([]v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	loadBalancers := []v1Types.LoadBalancer{}

	portalUserID := c.GetUserIDFromProviderUserID(types.ProviderUserID(userID))
	if portalUserID == "" {
		return nil, fmt.Errorf(errUserNotFound.Error(), userID)
	}

	accounts := c.GetAccountsByUserID(portalUserID)

	filteredAccounts := []types.Account{}
	switch {
	case roleNameFilter == v1Types.RoleName(""):
		filteredAccounts = accounts

	case roleNameFilter != v1Types.RoleName(""):
		for _, account := range accounts {
			for _, user := range account.Users {
				if user.UserID == portalUserID && user.RoleName == types.RoleName(roleNameFilter) {
					filteredAccounts = append(filteredAccounts, account)
				}
			}
		}
	}

	for _, account := range filteredAccounts {
		for _, portalApp := range account.PortalApps {
			legacyLB := legacyadapters.ConvertToLegacyLoadBalancer(*portalApp, account)

			loadBalancers = append(loadBalancers, legacyLB)
		}
	}

	return loadBalancers, nil
}

// GetLegacyLoadBalancer returns one V1 LoadBalancer by ID
func (c *Cache) GetLegacyLoadBalancer(ctx context.Context, lbID string) (v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalApp := c.GetPortalApp(types.PortalAppID(lbID))
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(ErrPortalAppNotFound.Error(), lbID)
	}

	account := c.GetAccount(portalApp.AccountID)
	if account.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errAccountNotFound.Error(), portalApp.AccountID)
	}

	return legacyadapters.ConvertToLegacyLoadBalancer(portalApp, account), nil
}

// GetLegacyLoadBalancers returns all V1 LoadBalancers
func (c *Cache) GetLegacyLoadBalancers(ctx context.Context) []v1Types.LoadBalancer {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var lbs []v1Types.LoadBalancer

	for _, account := range c.accounts {
		for _, portalApp := range account.PortalApps {
			lbs = append(lbs, legacyadapters.ConvertToLegacyLoadBalancer(*portalApp, *account))
		}
	}

	return lbs
}

// GetLegacyApplicationsByUserID returns all V1 Applicatons for a User ID
func (c *Cache) GetLegacyApplicationsByUserID(ctx context.Context, userID string) ([]v1Types.Application, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	applications := []v1Types.Application{}

	portalUserID := c.GetUserIDFromProviderUserID(types.ProviderUserID(userID))
	if portalUserID == "" {
		return nil, fmt.Errorf(errUserNotFound.Error(), userID)
	}

	accounts := c.GetAccountsByUserID(portalUserID)

	for _, account := range accounts {
		for _, portalApp := range account.PortalApps {
			_, account, err := c.getUserIDAndAccount(portalApp.AccountID)
			if err != nil {
				return applications, err
			}

			for protocolAppID := range portalApp.AATs {
				legacyApp := legacyadapters.ConvertToLegacyApplication(*portalApp, *account, protocolAppID)

				applications = append(applications, legacyApp)
			}
		}
	}

	if len(applications) == 0 {
		return applications, fmt.Errorf(errPortalAppsNotFoundForUser.Error(), portalUserID)
	}

	return applications, nil
}

// GetLegacyBlockchains returns all V1 Blockchains
func (c *Cache) GetLegacyBlockchains(ctx context.Context) []v1Types.Blockchain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	blockchains := []v1Types.Blockchain{}

	chains := c.GetChains(Options{IncludeDeleted: false})

	for _, chain := range chains {
		blockchains = append(blockchains, legacyadapters.ConvertToLegacyBlockchain(chain))
	}

	return blockchains
}

// GetLegacyBlockchain returns one V1 Blockchain for a blockchain ID
func (c *Cache) GetLegacyBlockchain(ctx context.Context, chainID string) v1Types.Blockchain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	chain := c.GetChain(types.RelayChainID(chainID), Options{})
	if chain.ID == "" {
		return v1Types.Blockchain{}
	}

	return legacyadapters.ConvertToLegacyBlockchain(chain)
}

// GetLegacyBlockchains returns all V1 Applicatons for a User ID
func (c *Cache) GetLegacyUserPermissionsByUserID(ctx context.Context, userID string) (v1Types.UserPermissions, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalUserID := c.GetUserIDFromProviderUserID(types.ProviderUserID(userID))
	if portalUserID == "" {
		return v1Types.UserPermissions{}, fmt.Errorf(errUserNotFound.Error(), userID)
	}

	userPermissions := c.GetUserPermissionsByUserID(portalUserID)

	legacyUserPermissions := v1Types.UserPermissions{
		UserID:        v1Types.UserID(userID),
		LoadBalancers: make(map[v1Types.LoadBalancerID]v1Types.LoadBalancerPermissions),
	}

	for accountID, accountPermissions := range userPermissions.Accounts {
		account, ok := c.accounts[accountID]
		if ok {
			for _, portalApp := range account.PortalApps {
				lbID := v1Types.LoadBalancerID(portalApp.ID)
				// TODO: remove after v2 migration
				rolePermissions := []v1Types.PermissionsEnum{}
				for _, permission := range accountPermissions.Permissions {
					rolePermissions = append(rolePermissions, v1Types.PermissionsEnum(permission))
				}

				legacyUserPermissions.LoadBalancers[lbID] = v1Types.LoadBalancerPermissions{
					RoleName:    v1Types.RoleName(accountPermissions.RoleName),
					Permissions: rolePermissions,
				}
			}
		}
	}

	return legacyUserPermissions, nil
}

// GetLegacyPayPlas returns all V1 Applicatons for a User ID
func (c *Cache) GetLegacyPayPlan(ctx context.Context, payPlanType string) (v1Types.PayPlan, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	plan := c.GetPlan(types.PayPlanType(payPlanType))
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

	userID := types.UserID(portalUserID)

	accounts := c.GetAccountsByUserID(userID)

	pendingAccounts := []types.Account{}
	for _, account := range accounts {
		for _, user := range account.Users {
			if user.UserID == userID && !user.Accepted {
				pendingAccounts = append(pendingAccounts, account)
			}
		}
	}

	for _, account := range pendingAccounts {
		for _, portalApp := range account.PortalApps {
			loadBalancers = append(loadBalancers, legacyadapters.ConvertToLegacyLoadBalancer(*portalApp, account))
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

	accounts := c.GetAccountsByUserID(types.UserID(portalUserID))
	for _, account := range accounts {
		for _, user := range account.Users {
			if user.UserID == types.UserID(userID) && user.RoleName == types.RoleOwner {
				count++
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

	portalUserID := c.GetUserIDFromProviderUserID(types.ProviderUserID(lb.UserID))
	if portalUserID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errUserNotFound.Error(), lb.UserID)
	}

	var account types.Account

	accounts := c.GetAccountsByUserID(portalUserID)
	for _, userAccount := range accounts {
		if userAccount.Users[portalUserID].RoleName == types.RoleOwner {
			account = userAccount
			break
		}
	}
	if account.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errAccountNotFoundForUser.Error(), portalUserID)
	}

	portalApp.AccountID = account.ID

	// Get the account plan based on the input plan type
	inputPlan := lb.Applications[0].Limit.PayPlan.Type
	accountPlan := c.GetPlan(types.PayPlanType(inputPlan))

	// Write the portal app to the data store
	createdPortalApp, err := writer.WritePortalApp(ctx, portalApp, aat, createdAt)
	if err != nil {
		return v1Types.LoadBalancer{}, fmt.Errorf(errWritingPortalApp.Error(), err)
	}

	// Add the created portal app to the PortalApps map of the created account
	account.PortalApps = map[types.PortalAppID]*types.PortalApp{
		createdPortalApp.ID: createdPortalApp,
	}

	// Set the account plan to the retrieved plan and return the V1 LoadBalancer object
	account.Plan = &accountPlan

	return legacyadapters.ConvertToLegacyLoadBalancer(*createdPortalApp, account), nil
}

// CreateLegacyLoadBalancerUser creates a single V2 AccountUserAccess row for a given account
func (c *Cache) CreateLegacyLoadBalancerUser(ctx context.Context, writer driver.Writer, lbID string, user v1Types.UserAccess) (v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	// Get the current time to use as the created time for the account and portal app
	createdAt := time.Now()

	portalAppID := types.PortalAppID(lbID)

	portalApp := c.GetPortalApp(portalAppID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(ErrPortalAppNotFound.Error(), portalAppID)
	}
	account := c.GetAccount(portalApp.AccountID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errAccountNotFound.Error(), portalApp.AccountID)
	}

	// Convert the V2 LoadBalancer object to a V2 account and portal app using the ConvertToV2AccountAndPortalApp function
	accountUserAccess := legacyadapters.ConvertToV2AccountUserAccess(user)
	createUser := types.CreateAccountUserAccess{
		AccountID: account.ID,
		Email:     accountUserAccess.Email,
		RoleName:  accountUserAccess.RoleName,
	}

	newUser, err := writer.WriteAccountUser(ctx, createUser, createdAt)
	if err != nil {
		return v1Types.LoadBalancer{}, err
	}

	account.Users[newUser.UserID] = *newUser

	return legacyadapters.ConvertToLegacyLoadBalancer(portalApp, account), nil
}

// CreateLegacyBlockchain creates one blockchain
func (c *Cache) CreateLegacyBlockchain(ctx context.Context, writer driver.Writer, blockchain v1Types.Blockchain) (v1Types.Blockchain, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	createdAt := time.Now()

	chain := legacyadapters.ConvertToV2Chain(blockchain)

	_, err := writer.WriteChain(ctx, chain, createdAt)
	if err != nil {
		return v1Types.Blockchain{}, fmt.Errorf("CreateLegacyBlockchain failed: %w", err)
	}

	return legacyadapters.ConvertToLegacyBlockchain(chain), nil
}

/* ---------- Update Methods ---------- */

// UpdateLegacyApplicationsFirstDateSurpassed updates the first date surpassed for a list of application IDs
func (c *Cache) UpdateLegacyApplicationsFirstDateSurpassed(ctx context.Context, writer driver.Writer, updateInput types.UpdateFirstDateSurpassed) ([]*v1Types.Application, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var portalAppsToUpdate []types.PortalApp
	var appsToUpdate []*v1Types.Application

	if len(updateInput.PortalAppIDs) == 0 {
		return appsToUpdate, fmt.Errorf("no application IDs on input")
	}

	for _, appID := range updateInput.PortalAppIDs {
		portalAppID := types.PortalAppID(appID)

		portalApp := c.GetPortalApp(portalAppID)
		if portalApp.ID == "" {
			return appsToUpdate, fmt.Errorf("UpdateFirstDateSurpassed failed: %s not found", portalAppID)
		}

		portalAppsToUpdate = append(portalAppsToUpdate, portalApp)
	}

	err := writer.UpdatePortalAppsFirstDateSurpassed(ctx, &updateInput)
	if err != nil {
		return appsToUpdate, fmt.Errorf("UpdateFirstDateSurpassed failed: %W", err)
	}

	for _, portalApp := range portalAppsToUpdate {
		userID, account, err := c.getUserIDAndAccount(portalApp.AccountID)
		if err != nil {
			return appsToUpdate, err
		}

		appsToUpdate = append(appsToUpdate, legacyadapters.ConvertToLegacyApplications(portalApp, userID, account.Plan.Type, account.Plan.AppLimit)...)
	}

	for _, app := range appsToUpdate {
		app.FirstDateSurpassed = updateInput.FirstDateSurpassed
	}

	return appsToUpdate, nil
}

// UpdateLegacyLoadBalancer updates one V2 PortalApp in the cache from a legacy LoadBalancer
func (c *Cache) UpdateLegacyLoadBalancer(ctx context.Context, writer driver.Writer, lbID string, update v1Types.UpdateApplication) (v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalAppID := types.PortalAppID(lbID)
	updatedAt := time.Now()

	portalApp := c.GetPortalApp(portalAppID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(ErrPortalAppNotFound.Error(), portalAppID)
	}
	account := c.GetAccount(portalApp.AccountID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errAccountNotFound.Error(), portalApp.AccountID)
	}

	updatePortalApp := legacyadapters.ConvertToV2UpdatePortalApp(update, lbID)

	switch update.Remove {
	case true:
		err := writer.SetPortalAppDeleted(ctx, portalAppID, updatedAt)
		if err != nil {
			return v1Types.LoadBalancer{}, err
		}
	default:
		err := writer.UpdatePortalApp(ctx, updatePortalApp, updatedAt)
		if err != nil {
			return v1Types.LoadBalancer{}, err
		}
	}

	return legacyadapters.ConvertToLegacyLoadBalancer(portalApp, account), nil
}

// UpdateLegacyUserRole updates the role for a single LB user
func (c *Cache) UpdateLegacyUserRole(ctx context.Context, writer driver.Writer, lbID string, updateUser v1Types.UpdateUserAccess) (v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalUserID := types.UserID(updateUser.UserID)

	portalAppID := types.PortalAppID(lbID)
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

	for i := range account.Users {
		if account.Users[i].UserID == portalUserID {
			updatedUser := account.Users[i]
			updatedUser.RoleName = types.RoleName(updateUser.RoleName)
			account.Users[i] = updatedUser
			break
		}
	}

	return legacyadapters.ConvertToLegacyLoadBalancer(portalApp, account), nil
}

// AcceptLegacyUser updates the role for a single LB user
func (c *Cache) AcceptLegacyUser(ctx context.Context, writer driver.Writer, lbID string, updateUser v1Types.UpdateUserAccess) (v1Types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalUserID := types.UserID(updateUser.UserID)

	portalAppID := types.PortalAppID(lbID)
	updatedAt := time.Now()

	portalApp := c.GetPortalApp(portalAppID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(ErrPortalAppNotFound.Error(), portalAppID)
	}
	account := c.GetAccount(portalApp.AccountID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errAccountNotFound.Error(), portalApp.AccountID)
	}

	updateAccountUser := legacyadapters.ConvertToV2UpdateAcceptAccountUserAccess(updateUser, account.ID, portalUserID)

	err := writer.UpdateAcceptAccountUser(ctx, updateAccountUser, updatedAt)
	if err != nil {
		return v1Types.LoadBalancer{}, err
	}

	for i := range account.Users {
		if account.Users[i].UserID == portalUserID {
			updatedUser := account.Users[i]
			updatedUser.RoleName = types.RoleName(updateUser.RoleName)
			updatedUser.Accepted = true
			account.Users[i] = updatedUser
			break
		}
	}

	return legacyadapters.ConvertToLegacyLoadBalancer(portalApp, account), nil
}

// UpdateLegacyBlockchain updates one blockchain
func (c *Cache) UpdateLegacyBlockchain(ctx context.Context, writer driver.Writer, blockchainID string, updateBlockchain v1Types.UpdateBlockchain) (v1Types.Blockchain, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	chainID := types.RelayChainID(blockchainID)
	updatedAt := time.Now()

	chain := c.GetChain(chainID, Options{})
	if chain.ID == "" {
		return v1Types.Blockchain{}, fmt.Errorf(errBlockchainNotFound.Error(), chainID)
	}

	update := legacyadapters.ConvertToV2UpdateChain(blockchainID, updateBlockchain)

	err := writer.UpdateChain(ctx, update, updatedAt)
	if err != nil {
		return v1Types.Blockchain{}, fmt.Errorf("UpdateChain failed: %w", err)
	}

	chain.UpdateBlockchain(&update)
	blockchain := legacyadapters.ConvertToLegacyBlockchain(chain)

	return blockchain, nil
}

// CreateLegacyRedirect creates one Redirect for one blockchain
func (c *Cache) CreateLegacyRedirect(ctx context.Context, writer driver.Writer, redirect v1Types.Redirect) (v1Types.Redirect, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	chainID := types.RelayChainID(redirect.BlockchainID)
	updatedAt := time.Now()

	chain := c.GetChain(chainID, Options{})
	if chain.ID == "" {
		return redirect, fmt.Errorf(errBlockchainNotFound.Error(), chainID)
	}

	chainUpdate := types.Chain{
		ID: chainID,
	}
	chainUpdate.Redirects = append(chainUpdate.Redirects, chain.Redirects...)
	chainUpdate.Redirects = append(chainUpdate.Redirects, legacyadapters.ConvertToV2Redirect(redirect))

	err := writer.UpdateChain(ctx, chainUpdate, updatedAt)
	if err != nil {
		return redirect, fmt.Errorf("CreateRedirect failed: %w", err)
	}

	return redirect, nil
}

// ActivateLegacyBlockchain activates one blockchain
func (c *Cache) ActivateLegacyBlockchain(ctx context.Context, writer driver.Writer, blockchainID string, active bool) (bool, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	chainID := types.RelayChainID(blockchainID)
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

	portalUserID := types.UserID(userID)

	portalAppID := types.PortalAppID(lbID)

	portalApp := c.GetPortalApp(portalAppID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(ErrPortalAppNotFound.Error(), portalAppID)
	}
	account := c.GetAccount(portalApp.AccountID)
	if portalApp.ID == "" {
		return v1Types.LoadBalancer{}, fmt.Errorf(errAccountNotFound.Error(), portalApp.AccountID)
	}

	err := writer.RemoveAccountUser(ctx, portalUserID, account.ID)
	if err != nil {
		return v1Types.LoadBalancer{}, fmt.Errorf("RemoveAccountUser failed: %w", err)
	}

	filteredUsers := map[types.UserID]types.AccountUserAccess{}
	for i := range account.Users {
		user := account.Users[i]
		if user.UserID != portalUserID {
			filteredUsers[user.UserID] = user
		}
	}
	account.Users = filteredUsers

	return legacyadapters.ConvertToLegacyLoadBalancer(portalApp, account), nil
}
