package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pokt-foundation/portal-db/v2/driver"
	legacyadapters "github.com/pokt-foundation/portal-db/v2/legacy-adapters"
	"github.com/pokt-foundation/portal-db/v2/types"
)

var (
	errAccountNotFound           = errors.New("account not found for ID %s")
	errAccountNotFoundForLB      = errors.New("load balancer not found for LB ID %s")
	errAppNotFound               = errors.New("application not found")
	errUserNotFound              = errors.New("user not found for provider user ID %s")
	errPlanNotFound              = errors.New("plan not found for type %s")
	errPortalAppsNotFoundForUser = errors.New("portal app not found for user ID %s")
	errBlockchainNotFound        = errors.New("blockchain not found for chain ID %s")
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
func (c *Cache) GetLegacyApplications(options Options) []types.Application {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalApps := c.GetPortalApps(options)

	var applications []types.Application

	for _, portalApp := range portalApps {
		userID, account, err := c.getUserIDAndAccount(portalApp.AccountID)
		if err != nil {
			c.log.Warn(fmt.Sprintf("error in GetLegacyApplications: %s", err.Error()))
			continue
		}

		legacyApp := legacyadapters.ConvertToLegacyApplication(portalApp, userID, account.PlanType, account.Plan.LegacyDailyLimit)

		applications = append(applications, legacyApp)
	}

	return applications
}

// GetLegacyApplication returns one V2 PortalApp in the cache converted to a V1 Application
func (c *Cache) GetLegacyApplication(portalAppID string) (types.Application, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalApp := c.GetPortalApp(types.PortalAppID(portalAppID))

	if portalApp.ID == "" {
		return types.Application{}, errAppNotFound
	}

	userID, account, err := c.getUserIDAndAccount(portalApp.AccountID)
	if err != nil {
		return types.Application{}, err
	}

	application := legacyadapters.ConvertToLegacyApplication(portalApp, userID, account.PlanType, account.Plan.LegacyDailyLimit)

	return application, nil
}

// GetLegacyLoadBalancersByUserID returns all V1 LoadBalancers for a User ID
func (c *Cache) GetLegacyLoadBalancersByUserID(ctx context.Context, userID string, roleNameFilter types.RoleName) ([]types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	loadBalancers := []types.LoadBalancer{}

	portalUserID := c.GetUserIDFromProviderUserID(types.ProviderUserID(userID))
	if portalUserID == "" {
		return nil, fmt.Errorf(errUserNotFound.Error(), userID)
	}

	accounts := c.GetAccountsByUserID(portalUserID)

	filteredAccounts := []types.Account{}
	switch {
	case roleNameFilter == types.RoleName(""):
		filteredAccounts = accounts

	case roleNameFilter != types.RoleName(""):
		for _, account := range accounts {
			for _, user := range account.Users {
				if user.UserID == portalUserID && user.RoleName == roleNameFilter {
					filteredAccounts = append(filteredAccounts, account)
				}
			}
		}
	}

	for _, account := range filteredAccounts {
		legacyLB := legacyadapters.ConvertToLegacyLoadBalancer(account)

		loadBalancers = append(loadBalancers, legacyLB)
	}

	return loadBalancers, nil
}

// GetLegacyLoadBalancer returns one V1 LoadBalancer by ID
func (c *Cache) GetLegacyLoadBalancer(ctx context.Context, lbID string) (types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var lbAccount types.Account
	for _, account := range c.accounts {
		if account.LegacyLoadBalancerID == lbID {
			lbAccount = *account
			break
		}
	}
	if lbAccount.ID == "" {
		return types.LoadBalancer{}, fmt.Errorf(errAccountNotFoundForLB.Error(), lbID)
	}

	return legacyadapters.ConvertToLegacyLoadBalancer(lbAccount), nil
}

// GetLegacyLoadBalancers returns all V1 LoadBalancers
func (c *Cache) GetLegacyLoadBalancers(ctx context.Context) []types.LoadBalancer {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var lbs []types.LoadBalancer

	for _, account := range c.accounts {
		lbs = append(lbs, legacyadapters.ConvertToLegacyLoadBalancer(*account))
	}

	return lbs
}

// GetLegacyApplicationsByUserID returns all V1 Applicatons for a User ID
func (c *Cache) GetLegacyApplicationsByUserID(ctx context.Context, userID string) ([]types.Application, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	applications := []types.Application{}

	portalUserID := c.GetUserIDFromProviderUserID(types.ProviderUserID(userID))
	if portalUserID == "" {
		return nil, fmt.Errorf(errUserNotFound.Error(), userID)
	}

	accounts := c.GetAccountsByUserID(portalUserID)

	for _, account := range accounts {
		for _, portalApp := range account.PortalApps {
			userID, account, err := c.getUserIDAndAccount(portalApp.AccountID)
			if err != nil {
				return applications, err
			}

			legacyApp := legacyadapters.ConvertToLegacyApplication(*portalApp, userID, account.PlanType, account.Plan.LegacyDailyLimit)

			applications = append(applications, legacyApp)
		}
	}

	if len(applications) == 0 {
		return applications, errPortalAppsNotFoundForUser
	}

	return applications, nil
}

// GetLegacyBlockchains returns all V1 Blockchains
func (c *Cache) GetLegacyBlockchains(ctx context.Context) []types.Blockchain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	blockchains := []types.Blockchain{}

	chains := c.GetChains(Options{IncludeDeleted: false})

	for _, chain := range chains {
		blockchains = append(blockchains, legacyadapters.ConvertToLegacyBlockchain(chain))
	}

	return blockchains
}

// GetLegacyBlockchain returns one V1 Blockchain for a blockchain ID
func (c *Cache) GetLegacyBlockchain(ctx context.Context, chainID string) types.Blockchain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	chain := c.GetChain(types.RelayChainID(chainID), Options{})
	if chain.ID == "" {
		return types.Blockchain{}
	}

	return legacyadapters.ConvertToLegacyBlockchain(chain)
}

// GetLegacyBlockchains returns all V1 Applicatons for a User ID
func (c *Cache) GetLegacyUserPermissionsByUserID(ctx context.Context, userID string) (types.LegacyUserPermissions, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalUserID := c.GetUserIDFromProviderUserID(types.ProviderUserID(userID))
	if portalUserID == "" {
		return types.LegacyUserPermissions{}, fmt.Errorf(errUserNotFound.Error(), userID)
	}

	userPermissions := c.GetUserPermissionsByUserID(portalUserID)

	legacyUserPermissions := types.LegacyUserPermissions{
		UserID:        types.UserID(userID),
		LoadBalancers: make(map[types.LoadBalancerID]types.LoadBalancerPermissions),
	}

	for accountID, accountPermissions := range userPermissions.Accounts {
		account, ok := c.accounts[accountID]
		if ok {
			lbID := types.LoadBalancerID(account.LegacyLoadBalancerID)
			legacyUserPermissions.LoadBalancers[lbID] = types.LoadBalancerPermissions(accountPermissions)
		}
	}

	return legacyUserPermissions, nil
}

// GetLegacyPayPlas returns all V1 Applicatons for a User ID
func (c *Cache) GetLegacyPayPlan(ctx context.Context, payPlanType string) (types.PayPlan, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	plan := c.GetPlan(types.PayPlanType(payPlanType))
	if plan.Type == "" {
		return types.PayPlan{}, fmt.Errorf(errPlanNotFound.Error(), payPlanType)
	}

	return legacyadapters.ConvertToLegacyPayPlan(plan), nil
}

// GetLegacyPayPlans returns all V1 Applicatons for a User ID
func (c *Cache) GetLegacyPayPlans(ctx context.Context) []types.PayPlan {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	payPlans := []types.PayPlan{}

	plans := c.GetPlans()

	for _, plan := range plans {
		payPlans = append(payPlans, legacyadapters.ConvertToLegacyPayPlan(plan))
	}

	return payPlans
}

// GetLegacyPendingLoadBalancers returns all pending load balancers for a given portal user ID.
func (c *Cache) GetLegacyPendingLoadBalancers(ctx context.Context, portalUserID string) ([]types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	loadBalancers := []types.LoadBalancer{}

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
		legacyLB := legacyadapters.ConvertToLegacyLoadBalancer(account)

		loadBalancers = append(loadBalancers, legacyLB)
	}

	return loadBalancers, nil
}

// GetLegacyLoadBalancersCount returns all pending load balancers for a given portal user ID.
func (c *Cache) GetLegacyLoadBalancersCount(ctx context.Context, portalUserID string) int {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	count := 0

	userID := types.UserID(portalUserID)

	accounts := c.GetAccountsByUserID(types.UserID(portalUserID))
	for _, account := range accounts {
		for _, user := range account.Users {
			if user.UserID == userID && user.RoleName == types.RoleOwner {
				count++
			}
		}
	}

	return count
}

/* ---------- Create Methods ---------- */

// CreateLegacyLoadBalancer returns one V2 PortalApp in the cache converted to a V1 Application
func (c *Cache) CreateLegacyLoadBalancer(ctx context.Context, writer driver.Writer, lb types.LoadBalancer) (types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	// Get the current time to use as the created time for the account and portal app
	createdAt := time.Now()

	// Convert the V2 LoadBalancer object to a V2 account and portal app using the ConvertToV2AccountAndPortalApp function
	account, portalApp := legacyadapters.ConvertToV2AccountAndPortalApp(lb, lb.UserID)

	portalUserID := c.GetUserIDFromProviderUserID(types.ProviderUserID(lb.UserID))
	if portalUserID == "" {
		return types.LoadBalancer{}, fmt.Errorf(errUserNotFound.Error(), lb.UserID)
	}

	// Get the account plan based on the input plan type
	inputPlan := lb.Applications[0].Limit.PayPlan.Type
	accountPlan := c.GetPlan(inputPlan)

	// Write the account to the data store
	createdAccount, err := writer.WriteAccount(ctx, portalUserID, account, createdAt)
	if err != nil {
		return types.LoadBalancer{}, errors.New("WHATVER")
	}

	// Write the portal app to the data store
	createdPortalApp, err := writer.WritePortalApp(ctx, portalApp, createdAt)
	if err != nil {
		return types.LoadBalancer{}, errors.New("WHATVER")
	}

	// Add the created portal app to the PortalApps map of the created account
	createdAccount.PortalApps = map[types.PortalAppID]*types.PortalApp{
		createdPortalApp.ID: createdPortalApp,
	}

	// Set the account plan to the retrieved plan and return the V1 LoadBalancer object
	createdAccount.Plan = &accountPlan
	return legacyadapters.ConvertToLegacyLoadBalancer(*createdAccount), nil
}

// CreateLegacyLoadBalancerUser creates a single V2 AccountUserAccess row for a given account
func (c *Cache) CreateLegacyLoadBalancerUser(ctx context.Context, writer driver.Writer, lbID string, user types.UserAccess) (types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	// Get the current time to use as the created time for the account and portal app
	createdAt := time.Now()

	var lbAccount types.Account
	for _, account := range c.accounts {
		if account.LegacyLoadBalancerID == lbID {
			lbAccount = *account
			break
		}
	}
	if lbAccount.ID == "" {
		return types.LoadBalancer{}, fmt.Errorf(errAccountNotFound.Error(), lbID)
	}

	// Convert the V2 LoadBalancer object to a V2 account and portal app using the ConvertToV2AccountAndPortalApp function
	accountUserAccess := legacyadapters.ConvertToV2AccountUserAccess(user)
	createUser := types.CreateAccountUserAccess{
		AccountID: lbAccount.ID,
		Email:     accountUserAccess.Email,
		RoleName:  accountUserAccess.RoleName,
	}

	newUser, err := writer.WriteAccountUser(ctx, createUser, createdAt)
	if err != nil {
		return types.LoadBalancer{}, err
	}

	lbAccount.Users[newUser.UserID] = *newUser

	return legacyadapters.ConvertToLegacyLoadBalancer(lbAccount), nil
}

// CreateLegacyBlockchain creates one blockchain
func (c *Cache) CreateLegacyBlockchain(ctx context.Context, writer driver.Writer, blockchain types.Blockchain) (types.Blockchain, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	createdAt := time.Now()

	chain := legacyadapters.ConvertToV2Chain(blockchain)

	_, err := writer.WriteChain(ctx, chain, createdAt)
	if err != nil {
		return types.Blockchain{}, fmt.Errorf("ActivateLegacyBlockchain failed: %w", err)
	}

	return blockchain, nil
}

/* ---------- Update Methods ---------- */

// UpdateLegacyApplication returns one V2 PortalApp in the cache converted to a V1 Application
func (c *Cache) UpdateLegacyApplication(ctx context.Context, writer driver.Writer, appID string, update types.UpdateApplication) (types.Application, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalAppID := types.PortalAppID(appID)
	updatedAt := time.Now()

	app := c.GetPortalApp(portalAppID)
	if app.ID == "" {
		return types.Application{}, errAppNotFound
	}

	updatePortalApp := legacyadapters.ConvertToV2UpdatePortalApp(update, appID)

	switch update.Remove {
	case true:
		err := writer.SetPortalAppDeleted(ctx, portalAppID, updatedAt)
		if err != nil {
			return types.Application{}, err
		}
	default:
		err := writer.UpdatePortalApp(ctx, updatePortalApp, updatedAt)
		if err != nil {
			return types.Application{}, err
		}
	}

	return types.Application{}, nil // TODO fix this
}

// UpdateLegacyApplicationsFirstDateSurpassed updates the first date surpassed for a list of application IDs
func (c *Cache) UpdateLegacyApplicationsFirstDateSurpassed(ctx context.Context, writer driver.Writer, updateInput types.UpdateFirstDateSurpassed) ([]types.Application, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var appsToUpdate []types.Application

	if len(updateInput.PortalAppIDs) == 0 {
		return appsToUpdate, fmt.Errorf("no application IDs on input")
	}

	for _, appID := range updateInput.PortalAppIDs {
		portalAppID := types.PortalAppID(appID)

		portalApp := c.GetPortalApp(portalAppID)
		if portalApp.ID == "" {
			return appsToUpdate, fmt.Errorf("UpdateFirstDateSurpassed failed: %s not found", portalAppID)
		}

		userID, account, err := c.getUserIDAndAccount(portalApp.AccountID)
		if err != nil {
			return appsToUpdate, err
		}

		appsToUpdate = append(appsToUpdate, legacyadapters.ConvertToLegacyApplication(portalApp, userID, account.PlanType, account.Plan.LegacyDailyLimit))
	}

	err := writer.UpdatePortalAppsFirstDateSurpassed(ctx, &updateInput)
	if err != nil {
		return appsToUpdate, fmt.Errorf("UpdateFirstDateSurpassed failed: %W", err)
	}

	for _, app := range appsToUpdate {
		app.FirstDateSurpassed = updateInput.FirstDateSurpassed
	}

	return appsToUpdate, nil
}

// UpdateLegacyLoadBalancer updates one V2 Account in the cache from a legacy LoadBalancer
func (c *Cache) UpdateLegacyLoadBalancer(ctx context.Context, writer driver.Writer, lbID string, update types.UpdateLoadBalancer) (types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var lbAccount types.Account
	for _, account := range c.accounts {
		if account.LegacyLoadBalancerID == lbID {
			lbAccount = *account
			break
		}
	}
	if lbAccount.ID == "" {
		return types.LoadBalancer{}, fmt.Errorf(errAccountNotFoundForLB.Error(), lbID)
	}

	updatedAt := time.Now()

	switch update.Remove {
	case true:
		err := writer.SetAccountDeleted(ctx, lbAccount.ID, updatedAt)
		if err != nil {
			return types.LoadBalancer{}, err
		}
	default:
		updateAccount := types.UpdateAccount{
			AccountID: lbAccount.ID,
			Name:      update.Name,
		}
		err := writer.UpdateAccount(ctx, updateAccount, updatedAt)
		if err != nil {
			return types.LoadBalancer{}, err
		}
	}

	return types.LoadBalancer{}, nil // TODO fix this
}

// UpdateLegacyUserRole updates the role for a single LB user
func (c *Cache) UpdateLegacyUserRole(ctx context.Context, writer driver.Writer, lbID string, updateUser types.UpdateUserAccess) (types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalUserID := c.GetUserIDFromProviderUserID(types.ProviderUserID(updateUser.UserID))
	if portalUserID == "" {
		return types.LoadBalancer{}, fmt.Errorf(errUserNotFound.Error(), updateUser.UserID)
	}

	var lbAccount types.Account
	for _, account := range c.accounts {
		if account.LegacyLoadBalancerID == lbID {
			lbAccount = *account
			break
		}
	}
	if lbAccount.ID == "" {
		return types.LoadBalancer{}, fmt.Errorf(errAccountNotFoundForLB.Error(), lbID)
	}

	updatedAt := time.Now()

	updateAccountUser := legacyadapters.ConvertToV2UpdateAccountUserAccess(updateUser, lbID, string(portalUserID))

	err := writer.SetAccountUserRole(ctx, updateAccountUser, updatedAt)
	if err != nil {
		return types.LoadBalancer{}, err
	}

	for i := range lbAccount.Users {
		if lbAccount.Users[i].UserID == portalUserID {
			updatedUser := lbAccount.Users[i]
			updatedUser.RoleName = updateUser.RoleName
			lbAccount.Users[i] = updatedUser
			break
		}
	}

	updatedLB := legacyadapters.ConvertToLegacyLoadBalancer(lbAccount)

	return updatedLB, nil
}

// AcceptLegacyUser updates the role for a single LB user
func (c *Cache) AcceptLegacyUser(ctx context.Context, writer driver.Writer, lbID string, updateUser types.UpdateUserAccess) (types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalUserID := c.GetUserIDFromProviderUserID(types.ProviderUserID(updateUser.UserID))
	if portalUserID == "" {
		return types.LoadBalancer{}, fmt.Errorf(errUserNotFound.Error(), updateUser.UserID)
	}

	var lbAccount types.Account
	for _, account := range c.accounts {
		if account.LegacyLoadBalancerID == lbID {
			lbAccount = *account
			break
		}
	}
	if lbAccount.ID == "" {
		return types.LoadBalancer{}, fmt.Errorf(errAccountNotFoundForLB.Error(), lbID)
	}

	updatedAt := time.Now()

	updateAccountUser := legacyadapters.ConvertToV2UpdateAcceptAccountUserAccess(updateUser, lbAccount.ID, portalUserID)

	err := writer.UpdateAcceptAccountUser(ctx, updateAccountUser, updatedAt)
	if err != nil {
		return types.LoadBalancer{}, err
	}

	for i := range lbAccount.Users {
		if lbAccount.Users[i].UserID == portalUserID {
			updatedUser := lbAccount.Users[i]
			updatedUser.RoleName = updateUser.RoleName
			updatedUser.Accepted = true
			lbAccount.Users[i] = updatedUser
			break
		}
	}

	updatedLB := legacyadapters.ConvertToLegacyLoadBalancer(lbAccount)

	return updatedLB, nil
}

// UpdateLegacyBlockchain updates one blockchain
func (c *Cache) UpdateLegacyBlockchain(ctx context.Context, writer driver.Writer, blockchainID string, updateBlockchain types.UpdateBlockchain) (types.Blockchain, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	chainID := types.RelayChainID(blockchainID)
	updatedAt := time.Now()

	chain := c.GetChain(chainID, Options{})
	if chain.ID == "" {
		return types.Blockchain{}, fmt.Errorf(errBlockchainNotFound.Error(), chainID)
	}

	update := legacyadapters.ConvertToV2UpdateChain(updateBlockchain)
	err := writer.UpdateChain(ctx, update, updatedAt)
	if err != nil {
		return types.Blockchain{}, fmt.Errorf("UpdateChain failed: %w", err)
	}

	chain.UpdateBlockchain(&update)
	blockchain := legacyadapters.ConvertToLegacyBlockchain(chain)

	return blockchain, nil
}

// CreateLegacyRedirect creates one Redirect for one blockchain
func (c *Cache) CreateLegacyRedirect(ctx context.Context, writer driver.Writer, redirect types.Redirect) (types.Redirect, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	chainID := types.RelayChainID(redirect.ChainID)
	updatedAt := time.Now()

	chain := c.GetChain(chainID, Options{})
	if chain.ID == "" {
		return redirect, fmt.Errorf(errBlockchainNotFound.Error(), chainID)
	}

	redirectUpdate := legacyadapters.ConvertToV2Redirect(redirect)
	chainUpdate := types.Chain{
		ID:        chainID,
		Redirects: []types.GigastakeRedirect{redirectUpdate},
	}
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
func (c *Cache) RemoveLegacyAccountUser(ctx context.Context, writer driver.Writer, userID string, lbID string) (types.LoadBalancer, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalUserID := c.GetUserIDFromProviderUserID(types.ProviderUserID(userID))
	if portalUserID == "" {
		return types.LoadBalancer{}, fmt.Errorf(errUserNotFound.Error(), userID)
	}

	var lbAccount types.Account
	for _, account := range c.accounts {
		if account.LegacyLoadBalancerID == lbID {
			lbAccount = *account
			break
		}
	}
	if lbAccount.ID == "" {
		return types.LoadBalancer{}, fmt.Errorf(errAccountNotFoundForLB.Error(), lbID)
	}

	err := writer.RemoveAccountUser(ctx, portalUserID, lbAccount.ID)
	if err != nil {
		return types.LoadBalancer{}, fmt.Errorf("RemoveAccountUser failed: %w", err)
	}

	filteredUsers := map[types.UserID]types.AccountUserAccess{}
	for i := range lbAccount.Users {
		user := lbAccount.Users[i]
		if user.UserID != portalUserID {
			filteredUsers[user.UserID] = user
		}
	}
	lbAccount.Users = filteredUsers

	return legacyadapters.ConvertToLegacyLoadBalancer(lbAccount), nil
}
