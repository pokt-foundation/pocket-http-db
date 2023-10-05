// package cache stores data in memeory to reduce database calls and increase speed of data fetching
package cache

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/pokt-foundation/portal-db/v2/driver"
	"github.com/pokt-foundation/portal-db/v2/types"
	"github.com/pokt-foundation/utils-go/logger"
)

type (
	/*
		The Cache struct provides an in-memory cache for data used across the Portal API backend architecture.
		It allows for efficient querying, insertion, and updating of these entities through appropriate cache methods.

		It also uses a Listener that listens for Postgres Notifications in order to update multiple
		instances of the Cache in response to inserts, updates and deletes on a single Postgres database.
	*/
	Cache struct {
		reader  driver.Reader
		logger  *logger.Logger
		rwMutex sync.RWMutex

		// Cached Data (including pending sub-table data maps)
		users             map[types.UserID]*types.User
		usersByProviderID map[types.ProviderUserID]*types.User

		accounts                 map[types.AccountID]*types.Account
		accountsByUserID         map[types.UserID]AccountsForUser
		pendingAccountUserAccess map[types.AccountID][]types.AccountUserAccess
		pendingUsers             map[types.UserID]types.User
		pendingUserAuthProviders map[types.UserID]types.UserAuthProvider

		portalApps                map[types.PortalAppID]*types.PortalApp
		portalAppsByProtocolAppID map[types.ProtocolAppID]*types.PortalApp
		pendingPortalAppAATs      map[types.PortalAppID]types.AAT
		pendingSettings           map[types.PortalAppID]types.Settings
		pendingAppNotifications   map[types.PortalAppID]types.AppNotification
		pendingWhitelists         map[types.PortalAppID][]types.Whitelist

		gigastakeApps              map[types.GigastakeAppID]*types.GigastakeApp
		pendingChainsGigastakeApps map[types.GigastakeAppID]types.ChainGigastakeApp

		chains           map[types.RelayChainID]*types.Chain
		pendingAltruists map[types.RelayChainID][]types.Altruist
		pendingAliases   map[types.RelayChainID][]types.ChainAlias
		// DEPRECATED - TODO remove when move to only store aliases is complete
		pendingAliasDomains  map[types.RelayChainID][]types.AliasDomains
		pendingChecks        map[types.RelayChainID][]types.Check
		pendingGigastakeApps map[types.RelayChainID][]types.GigastakeApp

		plans map[types.PayPlanType]*types.Plan

		blockedContracts types.GlobalBlockedContracts
	}

	AccountsForUser map[types.AccountID]*types.Account

	Config struct {
		Reader              driver.Reader
		Logger              *logger.Logger
		CacheRefreshMinutes time.Duration
	}

	Options struct {
		IncludeDeleted            bool
		IncludeInactiveChains     bool
		ExcludeChainGigastakeApps bool
		FullDetails               bool
		FilterPortalApps          []types.RoleName
		Accepted                  *bool
	}

	userData struct {
		users             map[types.UserID]*types.User
		usersByProviderID map[types.ProviderUserID]*types.User
	}

	appsData struct {
		plans                     map[types.PayPlanType]*types.Plan
		accounts                  map[types.AccountID]*types.Account
		portalApps                map[types.PortalAppID]*types.PortalApp
		portalAppsByProtocolAppID map[types.ProtocolAppID]*types.PortalApp
		accountsByUserID          map[types.UserID]AccountsForUser
	}

	chainsData struct {
		chains        map[types.RelayChainID]*types.Chain
		gigastakeApps map[types.GigastakeAppID]*types.GigastakeApp
	}
)

// NewCache returns cache instance from reader interface and starts the cache refresh handler and listener in separate goroutines
func NewCache(ctx context.Context, config Config) (*Cache, error) {
	cache := &Cache{
		reader: config.Reader,
		logger: config.Logger,

		pendingAccountUserAccess:   make(map[types.AccountID][]types.AccountUserAccess),
		pendingUsers:               make(map[types.UserID]types.User),
		pendingUserAuthProviders:   make(map[types.UserID]types.UserAuthProvider),
		pendingPortalAppAATs:       make(map[types.PortalAppID]types.AAT),
		pendingSettings:            make(map[types.PortalAppID]types.Settings),
		pendingAppNotifications:    make(map[types.PortalAppID]types.AppNotification),
		pendingWhitelists:          make(map[types.PortalAppID][]types.Whitelist),
		pendingAltruists:           make(map[types.RelayChainID][]types.Altruist),
		pendingChainsGigastakeApps: make(map[types.GigastakeAppID]types.ChainGigastakeApp),
		pendingAliases:             make(map[types.RelayChainID][]types.ChainAlias),
		// DEPRECATED - TODO remove when move to only store aliases is complete
		pendingAliasDomains:  make(map[types.RelayChainID][]types.AliasDomains),
		pendingChecks:        make(map[types.RelayChainID][]types.Check),
		pendingGigastakeApps: make(map[types.RelayChainID][]types.GigastakeApp),
	}

	// Perform initial cache data fetch and log time to fetch
	startTime := time.Now()
	err := cache.setCache(ctx)
	elapsedTime := time.Since(startTime).Milliseconds()
	if err != nil {
		cache.logger.Error(fmt.Sprintf("error in setCache: %s", err))
		return nil, err
	}

	cache.logger.Info("cache successfully set",
		slog.Float64("cache_refresh_minutes", config.CacheRefreshMinutes.Minutes()),
		slog.String("db_time_to_fetch", fmt.Sprintf("%dms", elapsedTime)),
	)

	// Start cache refresh goroutine to refresh cache at regular interval
	go cache.cacheRefreshHandler(ctx, config.CacheRefreshMinutes)

	// Start listener goroutine for Postgres notifications
	go cache.listen()

	return cache, nil
}

/* ---------- Cache Set and Refresh ---------- */

// ResetCache manually resets the cache; called in the router when the `/reset_cache` endpoint is hit.
func (c *Cache) ResetCache(ctx context.Context) error {
	err := c.setCache(ctx)
	if err != nil {
		c.logger.Error(fmt.Sprintf("error in ResetCache - setCache: %s", err))
		return err
	}

	c.logger.Info("cache successfully manually reset")

	return nil
}

// setCache initializes the cache by setting user permissions, accounts and portal apps, chains, and plans.
func (c *Cache) setCache(ctx context.Context) error {
	t := NewRunTimer("setCache", c.logger)
	defer t.Stop()

	// Perform all the database reads first without locking the mutex
	userData, err := c.fetchUserData(ctx)
	if err != nil {
		return err
	}
	appsData, err := c.fetchAppsData(ctx)
	if err != nil {
		return err
	}
	chainsData, err := c.fetchChainsData(ctx)
	if err != nil {
		return err
	}
	blockedContractsData, err := c.fetchBlockedContractsData(ctx)
	if err != nil {
		return err
	}

	// Lock the mutex only once data is fetched
	// to minimize the time the lock is held.
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	// Set the cache with the fetched data
	c.users = userData.users
	c.usersByProviderID = userData.usersByProviderID

	c.plans = appsData.plans
	c.accounts = appsData.accounts
	c.accountsByUserID = appsData.accountsByUserID
	c.portalApps = appsData.portalApps
	c.portalAppsByProtocolAppID = appsData.portalAppsByProtocolAppID

	c.chains = chainsData.chains
	c.gigastakeApps = chainsData.gigastakeApps

	c.blockedContracts = blockedContractsData

	return nil
}

// fetchUserData reads UserPermissions from the database and returns them.
func (c *Cache) fetchUserData(ctx context.Context) (userData, error) {
	users, err := c.reader.ReadAllUsers(ctx)
	if err != nil {
		return userData{}, fmt.Errorf("ReadAllUsers - %w", err)
	}

	usersByProviderID := make(map[types.ProviderUserID]*types.User)
	for _, user := range users {
		for _, authProvider := range user.AuthProviders {
			usersByProviderID[authProvider.ProviderUserID] = user
		}
	}

	return userData{users, usersByProviderID}, nil
}

// fetchAppsData reads Plans, Accounts and PortalApps from the database,
// sets each Account's Plan and PortalApps and creates a map of Accounts by UserID.
// also creates a map of PortalApps by ProtocolAppID
func (c *Cache) fetchAppsData(ctx context.Context) (appsData, error) {
	// Fetch Plans, Accounts and PortalApps
	plans, err := c.reader.ReadPlans(ctx)
	if err != nil {
		return appsData{}, fmt.Errorf("ReadPlans - %w", err)
	}
	accounts, err := c.reader.ReadAccounts(ctx, types.DriverOptions{IncludeDeleted: false})
	if err != nil {
		return appsData{}, fmt.Errorf("ReadAccounts - %w", err)
	}
	portalApps, err := c.reader.ReadPortalApps(ctx, types.DriverOptions{IncludeDeleted: false})
	if err != nil {
		return appsData{}, fmt.Errorf("ReadPortalApps - %w", err)
	}

	// Set pointer to Account Plan
	for _, account := range accounts {
		if plan, ok := plans[account.PlanType]; ok {
			account.Plan = plan
		} else {
			c.logger.Warn("error setting plan for account: plan type does not exist in cache",
				slog.String("account_id", string(account.ID)),
				slog.String("plan_type", string(account.PlanType)),
			)
		}
	}

	portalAppsByProtocolAppID := make(map[types.ProtocolAppID]*types.PortalApp)

	// Set pointers to Account PortalApps and create map of PortalApps by ProtocolAppID
	for portalAppID, portalApp := range portalApps {
		if portalApp.Deleted {
			continue
		}
<<<<<<< HEAD
=======

		if account, ok := accounts[portalApp.AccountID]; ok {
			// Set users for the Portal App
			portalAppUsers := make(map[types.UserID]*types.AccountUserAccess)
			for userID, user := range account.Users {
				for portalAppID := range user.PortalAppRoles {
					if portalAppID == portalApp.ID {
						userCopy := user
						portalAppUsers[userID] = &userCopy
					}
				}
			}
			portalApp.Users = portalAppUsers
>>>>>>> 807487a9148e2c17fd6c0d9018a69d66a00f8742

		if account, ok := accounts[portalApp.AccountID]; ok {
			if account.PortalApps == nil {
				account.PortalApps = make(map[types.PortalAppID]*types.PortalApp)
			}
			account.PortalApps[portalAppID] = portalApp

		}

		for aatID := range portalApp.AATs {
			portalAppsByProtocolAppID[aatID] = portalApp
		}
	}

	// Create map of Accounts by UserID for all Account team members
	accountsByUserID := make(map[types.UserID]AccountsForUser)
	for _, account := range accounts {
		if account.Deleted {
			continue
		}

		for _, user := range account.Users {
			if _, ok := accountsByUserID[user.UserID]; !ok {
				accountsByUserID[user.UserID] = make(map[types.AccountID]*types.Account)
			}
			accountsByUserID[user.UserID][account.ID] = account
		}
	}

	return appsData{
		plans,
		accounts,
		portalApps,
		portalAppsByProtocolAppID,
		accountsByUserID,
	}, nil
}

// fetchChainsData reads Chains and GigastakeApps from the database and returns them.
// It also sets pointers to GigastakeApps on Chains.
func (c *Cache) fetchChainsData(ctx context.Context) (chainsData, error) {
	chains, err := c.reader.ReadChains(ctx, types.DriverOptions{IncludeDeleted: false})
	if err != nil {
		return chainsData{}, fmt.Errorf("ReadChains - %w", err)
	}
	gigastakeApps, err := c.reader.ReadGigastakeApps(ctx, types.DriverOptions{IncludeDeleted: false})
	if err != nil {
		return chainsData{}, fmt.Errorf("ReadGigastakeApps - %w", err)
	}

	// Set pointers to Chain GigastakeApps
	for aatID, gigastakeApp := range gigastakeApps {
		for chainID := range gigastakeApp.ChainIDs {
			if chain, ok := chains[chainID]; ok {
				if chain.GigastakeApps == nil {
					chain.GigastakeApps = make(map[types.GigastakeAppID]*types.GigastakeApp)
				}
				chain.GigastakeApps[aatID] = gigastakeApp

			}
		}
	}

	return chainsData{chains, gigastakeApps}, nil
}

// fetchBlockedContractsData reads BlockedContracts from the database and returns them.
func (c *Cache) fetchBlockedContractsData(ctx context.Context) (types.GlobalBlockedContracts, error) {
	blockedContracts, err := c.reader.ReadBlockedContracts(ctx)
	if err != nil {
		return types.GlobalBlockedContracts{}, fmt.Errorf("ReadBlockedContracts - %w", err)
	}

	return blockedContracts, nil
}

// cacheRefreshHandler sets the cache from the DB at a regular interval (default is 5 minutes).
func (c *Cache) cacheRefreshHandler(ctx context.Context, cacheRefreshMinutes time.Duration) {
	for {
		<-time.After(cacheRefreshMinutes)

		err := c.setCache(ctx)
		if err != nil {
			c.logger.Error("err in cacheRefreshHandler: %w", err)

			os.Exit(1) // Kill PHD if cache refresh fails to trigger pod restart
		}
	}
}

/* ---------- User Data ---------- */

// GetUserByID returns the Portal User for a given portal UserID
func (c *Cache) GetUserByID(userID types.UserID) types.User {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	user, ok := c.users[userID]
	if !ok {
		return types.User{}
	}

	return *user
}

// GetUserByProviderID returns the Portal User for a given provider UserID
func (c *Cache) GetUserByProviderID(providerUserID types.ProviderUserID) types.User {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	user, ok := c.usersByProviderID[providerUserID]
	if !ok {
		return types.User{}
	}

	return *user
}

/* ---------- Blocked Contracts ---------- */

// GetBlockedContracts returns all global blocked contracts
func (c *Cache) GetBlockedContracts() types.GlobalBlockedContracts {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.blockedContracts
}

// upsertGlobalBlockedContract inserts or updates a global blocked contract
func (c *Cache) upsertGlobalBlockedContract(blockedContract types.BlockedContract) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	switch blockedContract.Active {
	case false:
		delete(c.blockedContracts.BlockedAddresses, blockedContract.BlockedAddress)
	default:
		c.blockedContracts.BlockedAddresses[blockedContract.BlockedAddress] = struct{}{}
	}
}

// deleteGlobalBlockedContract deletes a global blocked contract
func (c *Cache) deleteGlobalBlockedContract(blockedContract types.BlockedContract) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	delete(c.blockedContracts.BlockedAddresses, blockedContract.BlockedAddress)
}

/* ---------- Accounts ---------- */

// GetAccounts returns all Accounts in cache
// May select whether to include deleted Accounts. Default: false
func (c *Cache) GetAccounts(options Options) []types.Account {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var accountsToReturn []types.Account

	for _, account := range c.accounts {
		if options.IncludeDeleted || !account.Deleted {
			accountCopy := *account
			if len(accountCopy.PortalApps) == 0 {
				accountCopy.PortalApps = map[types.PortalAppID]*types.PortalApp{}
			}
<<<<<<< HEAD

			accountCopy.AssignPortalAppUsers()

=======
>>>>>>> 807487a9148e2c17fd6c0d9018a69d66a00f8742
			accountsToReturn = append(accountsToReturn, accountCopy)
		}
	}

	return accountsToReturn
}

// GetUserAccount returns one Account by accountID with portal apps filtered by user ID
func (c *Cache) GetUserAccount(accountID types.AccountID, userID types.UserID, options Options) types.Account {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	if account, ok := c.accounts[accountID]; ok {
		// TODO remove once cache adapters removed
		if userID == "" {
			return *account
		}

		if userCanReadAccount(account, userID, options) {
<<<<<<< HEAD
			account.AssignPortalAppUsers()

=======
>>>>>>> 807487a9148e2c17fd6c0d9018a69d66a00f8742
			// filter portal apps by role name and accepted status
			return filterAccountPortalApps(account, userID, options)
		}
	}

	return types.Account{}
}

// GetUserAccounts returns all Accounts for a given user with portal apps filtered by user ID
func (c *Cache) GetUserAccounts(userID types.UserID, options Options) []types.Account {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var accountsToReturn []types.Account

	for _, account := range c.accountsByUserID[userID] {
		if userCanReadAccount(account, userID, options) {
			// filter portal apps by role name and accepted status
			accountCopy := filterAccountPortalApps(account, userID, options)
			if accountCopy.ID == "" {
				continue
			}

<<<<<<< HEAD
			// assign account users to all portal apps attached to the account
			accountCopy.AssignPortalAppUsers()
=======
>>>>>>> 807487a9148e2c17fd6c0d9018a69d66a00f8742
			accountsToReturn = append(accountsToReturn, accountCopy)
		}

	}

	return accountsToReturn
}

// userCanReadAccount returns a boolean indicating if the user can read the account with the given query params
func userCanReadAccount(a *types.Account, userID types.UserID, options Options) bool {
	// if accepted = false include all accounts, which will have their portal apps filtered by accepted
	if options.Accepted != nil && !*options.Accepted {
		return true
	}

	// if owner of account then return true
	ownerID, _ := a.GetOwnerID()
	if userID == ownerID {
		return true
	}

	// if role filters param set then filter accounts by whether user has access
	if len(options.FilterPortalApps) > 0 && !userHasAccountPerms(a, userID, options.FilterPortalApps) {
		return false
	}

	// if accepted param set then filter accounts by whether user has accepted
	return userAcceptedAccountInvite(a, userID)
}

// userAcceptedAccountInvite returns a boolean indicating if the user
// has accepted an invitation to any PortalApp in the Account
func userAcceptedAccountInvite(a *types.Account, userID types.UserID) bool {
	if user, ok := a.Users[userID]; ok {
		acceptedInvite := false

		for _, portalAppAccepted := range user.PortalAppsAccepted {
			if portalAppAccepted {
				acceptedInvite = true
				break
			}
		}

		return acceptedInvite
	}

	return false
}

// userHasAccountPerms returns a boolean indicating if the user
// has any account permissions for any PortalApps for the account
func userHasAccountPerms(a *types.Account, userID types.UserID, filterRoles []types.RoleName) bool {
	if user, ok := a.Users[userID]; ok {
		hasPermissions := false

		for _, portalAppRole := range user.PortalAppRoles {
			for _, filter := range filterRoles {
				if filter == portalAppRole {
					hasPermissions = true
					break
				}
			}
		}

		return hasPermissions
	}

	return false
}

// filterAccountPortalApps filters PortalApps based on the provided RoleNames and Accepted status
func filterAccountPortalApps(account *types.Account, userID types.UserID, options Options) types.Account {
	userAccess, ok := account.Users[userID]
	if !ok {
		return types.Account{}
	}

	accountCopy := *account

	// filter portal apps by user's role and accepted status
	accountCopy.PortalApps = filterAccountPortalAppsByRole(accountCopy.PortalApps, userAccess, userID, options)

	// filter portal apps by accepted status
	if options.Accepted != nil {
		accountCopy.PortalApps = filterAccountPortalAppsByAccepted(accountCopy.PortalApps, userAccess, userID, options)
	}

	return accountCopy
}

// filterAccountPortalAppsByRole filters PortalApps based on the provided RoleNames
func filterAccountPortalAppsByRole(
	accountPortalApps map[types.PortalAppID]*types.PortalApp,
	userAccess types.AccountUserAccess,
	userID types.UserID,
	options Options,
) map[types.PortalAppID]*types.PortalApp {
	var filterRoles []types.RoleName
	if len(options.FilterPortalApps) > 0 {
		// If filter params provided then filter by those params
		filterRoles = options.FilterPortalApps
	} else {
		// If filter params not provided then include any portal apps that user has any role for
		filterRoles = []types.RoleName{types.RoleOwner, types.RoleAdmin, types.RoleMember}
	}

	filteredPortalApps := make(map[types.PortalAppID]*types.PortalApp)
	for _, portalApp := range accountPortalApps {
		if userRole, ok := userAccess.PortalAppRoles[portalApp.ID]; ok {
			for _, filter := range filterRoles {
				if filter == userRole {
					filteredPortalApps[portalApp.ID] = portalApp
				}
			}
		}
	}

	return filteredPortalApps
}

// filterAccountPortalAppsByAccepted filters the PortalAppss based on the Accepted param
func filterAccountPortalAppsByAccepted(
	accountPortalApps map[types.PortalAppID]*types.PortalApp,
	userAccess types.AccountUserAccess,
	userID types.UserID,
	options Options,
) map[types.PortalAppID]*types.PortalApp {
	portalApps := make(map[types.PortalAppID]*types.PortalApp)

	for _, portalApp := range accountPortalApps {
		portalAppAccepted, ok := userAccess.PortalAppsAccepted[portalApp.ID]

		if ok && portalAppAccepted == *options.Accepted {
			portalApps[portalApp.ID] = portalApp
		}
	}

	return portalApps
}

// upsertAccount adds, updates or deletes Account in cache
func (c *Cache) upsertAccount(account types.Account) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	// If marked as deleted, remove Account from map
	// and from all user ID maps of all Account users
	if account.Deleted {
		delete(c.accounts, account.ID)
		for _, user := range account.Users {
			delete(c.accountsByUserID, user.UserID)
		}
		return
	}

	// Ensure upsert does not overwrite existing Account fields
	if existingAccount, ok := c.accounts[account.ID]; ok {
		account.PortalApps = existingAccount.PortalApps
		account.Users = existingAccount.Users
		account.Integrations = existingAccount.Integrations
	}
	// Set plan in cache to latest value
	if plan, ok := c.plans[account.PlanType]; ok {
		account.Plan = plan
	}

	if pendingAccountUsers, ok := c.pendingAccountUserAccess[account.ID]; ok {
		for _, accountUser := range pendingAccountUsers {
			if account.Users == nil {
				account.Users = make(map[types.UserID]types.AccountUserAccess)
			}
			if existingAccountUser, ok := account.Users[accountUser.UserID]; ok {
				accountUser.Email = existingAccountUser.Email
			}
			if user, ok := c.pendingUsers[accountUser.UserID]; ok {
				accountUser.Email = user.Email
				delete(c.pendingUsers, accountUser.UserID)
			}

			account.Users[accountUser.UserID] = accountUser
		}
		delete(c.pendingAccountUserAccess, account.ID)
	}

	c.accounts[account.ID] = &account

	for _, user := range account.Users {
		if _, ok := c.accountsByUserID[user.UserID]; !ok {
			c.accountsByUserID[user.UserID] = make(map[types.AccountID]*types.Account)
		}
		c.accountsByUserID[user.UserID][account.ID] = &account
	}
}

// upsertAccount adds or updates AccountUserAccess in cache
func (c *Cache) upsertAccountUser(accountUser types.AccountUserAccess) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	accountID := accountUser.AccountID
	accountUser.AccountID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	// Save incoming portal app ID and user role
	var incomingPortalAppID types.PortalAppID
	var incomingPortalAppRole types.RoleName
	var accepted bool
	for portalAppID, role := range accountUser.PortalAppRoles {
		incomingPortalAppID = portalAppID
		incomingPortalAppRole = role
		break
	}
	for _, portalAppAccepted := range accountUser.PortalAppsAccepted {
		accepted = portalAppAccepted
		break
	}

	// Add user to permissions map if they have accepted their invitation
	if accepted {
		if user, ok := c.users[accountUser.UserID]; ok {
			if user.ArePermissionsEmpty() {
				user.Permissions = make(map[types.PortalAppID]types.PortalAppPermissions)
			}

			switch accountUser.Owner {
			// If incoming account user is owner, set permissions for all account's portal apps
			case true:
				if account, ok := c.accounts[accountID]; ok {
					for portalAppID := range account.PortalApps {
						_, err := user.UpsertPermissions(portalAppID, accountID, types.RoleOwner)
						if err != nil {
							fmt.Printf("error upserting user permissions: %v\n", err)
						}
					}
				}

				// If incoming account user is not owner only set permissions for incoming portal app
			default:
				_, err := user.UpsertPermissions(incomingPortalAppID, accountID, incomingPortalAppRole)
				if err != nil {
					fmt.Printf("error upserting user permissions: %v\n", err)
				}
			}
		}
	}

	if user, ok := c.pendingUsers[accountUser.UserID]; ok {
		accountUser.Email = user.Email
		delete(c.pendingUsers, accountUser.UserID)
	}

	// Add or update user's Accounts
	if account, ok := c.accounts[accountID]; ok {
		if account.Users == nil {
			account.Users = make(map[types.UserID]types.AccountUserAccess)
		}

		// Set values from existing app to avoid overwriting cache
		if existingAccountUser, ok := account.Users[accountUser.UserID]; ok {
			accountUser.Email = existingAccountUser.Email
			accountUser.PortalAppRoles = existingAccountUser.PortalAppRoles
			accountUser.PortalAppsAccepted = existingAccountUser.PortalAppsAccepted
		}

<<<<<<< HEAD
=======
		switch accountUser.Owner {
		// If incoming account user is owner, set user for all account's portal apps
		case true:
			if account, ok := c.accounts[accountID]; ok {
				for portalAppID := range account.PortalApps {
					if portalApp, ok := c.portalApps[portalAppID]; ok {
						if portalApp.Users == nil {
							portalApp.Users = make(map[types.UserID]*types.AccountUserAccess)
						}

						portalApp.Users[accountUser.UserID] = &accountUser
					}
				}
			}
			// If incoming account user is not owner only set user for incoming portal app
		default:
			if portalApp, ok := c.portalApps[incomingPortalAppID]; ok {
				if portalApp.Users == nil {
					portalApp.Users = make(map[types.UserID]*types.AccountUserAccess)
				}

				portalApp.Users[accountUser.UserID] = &accountUser
			}
		}

>>>>>>> 807487a9148e2c17fd6c0d9018a69d66a00f8742
		account.Users[accountUser.UserID] = accountUser
		if user, ok := account.Users[accountUser.UserID]; ok {
			if user.PortalAppRoles == nil {
				user.PortalAppRoles = make(map[types.PortalAppID]types.RoleName)
			}
			if user.PortalAppsAccepted == nil {
				user.PortalAppsAccepted = make(map[types.PortalAppID]bool)
			}

			switch accountUser.Owner {
			// If incoming account user is owner, set owner for all account's portal apps
			case true:
				for portalAppID := range account.PortalApps {
					user.PortalAppRoles[portalAppID] = types.RoleOwner
					user.PortalAppsAccepted[portalAppID] = true
				}

			// If incoming account user is not owner only set role for incoming portal app
			default:
				user.PortalAppRoles[incomingPortalAppID] = incomingPortalAppRole
				user.PortalAppsAccepted[incomingPortalAppID] = accepted
			}
		}

		if _, ok := c.accountsByUserID[accountUser.UserID]; !ok {
			c.accountsByUserID[accountUser.UserID] = make(map[types.AccountID]*types.Account)
		}

		c.accountsByUserID[accountUser.UserID][account.ID] = account

		return
	}

	c.pendingAccountUserAccess[accountID] = append(c.pendingAccountUserAccess[accountID], accountUser)
}

// upsertAccountIntegrations updates the LBs integrations saved in cache
func (c *Cache) upsertAccountIntegrations(accountIntegrations *types.AccountIntegrations) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	accountID := accountIntegrations.AccountID
	if account, ok := c.accounts[accountID]; ok {
		account.Integrations = *accountIntegrations
	}
}

// deleteAccountUser deletes Account user saved in cache and updates permissions
func (c *Cache) deleteAccountUser(accountUser types.AccountUserAccess) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	accountID := accountUser.AccountID
	userID := accountUser.UserID
	accountUser.AccountID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	// Save incoming portal app ID and user role
	var incomingPortalAppID types.PortalAppID
	for portalAppID := range accountUser.PortalAppRoles {
		incomingPortalAppID = portalAppID
		break
	}

	if accountUser.Owner {
		return
	}

	// Remove user permissions for each PortalApp in the account
	if user, ok := c.users[userID]; ok {
		if !user.ArePermissionsEmpty() {
			for portalAppID := range accountUser.PortalAppRoles {
				user.DeletePermissions(portalAppID)
			}
		}
	}

	// Remove from Accounts map
	if account, ok := c.accounts[accountID]; ok {
		if accountUser, ok := account.Users[userID]; ok {
			delete(accountUser.PortalAppRoles, incomingPortalAppID)
			delete(accountUser.PortalAppsAccepted, incomingPortalAppID)

			if len(accountUser.PortalAppRoles) == 0 && len(accountUser.PortalAppsAccepted) == 0 {
				delete(account.Users, userID)
			}
		}

		if accountsForUser, ok := c.accountsByUserID[accountUser.UserID]; ok {
			delete(accountsForUser, accountID)
		}
	}

	// Remove from PortalApp map
	if portalApp, ok := c.portalApps[incomingPortalAppID]; ok {
		delete(portalApp.Users, userID)
	}
}

/* ---------- PortalApps ---------- */

// GetPortalApp returns PortalApp from cache by portalAppID
func (c *Cache) GetPortalApp(portalAppID types.PortalAppID) types.PortalApp {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalApp, ok := c.portalApps[portalAppID]
	if !ok {
		return types.PortalApp{}
	}

	portalAppCopy := *portalApp

	if account, ok := c.accounts[portalApp.AccountID]; ok {
		portalAppCopy.AssignUsers(account)
	}

	return portalAppCopy
}

// GetPortalAppByProtocolAppID returns PortalApp from cache by protocolAppID
func (c *Cache) GetPortalAppByProtocolAppID(protocolAppID types.ProtocolAppID) types.PortalApp {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalApp, ok := c.portalAppsByProtocolAppID[protocolAppID]
	if !ok {
		return types.PortalApp{}
	}

	portalAppCopy := *portalApp

	if account, ok := c.accounts[portalApp.AccountID]; ok {
		portalAppCopy.AssignUsers(account)
	}

	return portalAppCopy
}

// GetPortalApps returns all PortalApps in cache
func (c *Cache) GetPortalApps(options Options) []types.PortalApp {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var portalAppsToReturn []types.PortalApp

	for _, portalApp := range c.portalApps {
		if options.IncludeDeleted || !portalApp.Deleted {
			portalAppCopy := *portalApp

			if account, ok := c.accounts[portalApp.AccountID]; ok {
				portalAppCopy.AssignUsers(account)
			}

			portalAppsToReturn = append(portalAppsToReturn, portalAppCopy)
		}
	}

	return portalAppsToReturn
}

// GetPortalAppsForMiddleware returns all PortalAppLites in cache
func (c *Cache) GetPortalAppsForMiddleware() []types.PortalAppLite {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var portalAppLites []types.PortalAppLite

	for _, portalApp := range c.portalApps {
		if !portalApp.Deleted {
			var plan *types.Plan

			// Direct apps have no plan associated with them
			// TODO: REVIEW APPS WITHOUT PLANS
			if !portalApp.IsDirectApp() {
				if legacyPlan, ok := c.plans[portalApp.LegacyFields.PlanType]; ok {
					plan = legacyPlan
				}
			}

			portalAppLites = append(portalAppLites, portalApp.ConvertPortalAppToPortalAppLite(plan))
		}
	}

	return portalAppLites
}

func (c *Cache) GetPortalAppsByUserID(userID types.UserID, options Options) []types.PortalApp {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var portalAppsToReturn []types.PortalApp

	for _, account := range c.GetUserAccounts(userID, options) {
		var portalApps []types.PortalApp

		for _, portalApp := range account.PortalApps {
			portalApps = append(portalApps, *portalApp)
		}

		portalAppsToReturn = append(portalAppsToReturn, portalApps...)
	}

	return portalAppsToReturn
}

// upsertPortalApp adds, updates or deletes PortalApp in cache
func (c *Cache) upsertPortalApp(portalApp types.PortalApp) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	if existingPortalApp, ok := c.portalApps[portalApp.ID]; ok {
		portalApp.AATs = existingPortalApp.AATs
		portalApp.Settings = existingPortalApp.Settings
		portalApp.Notifications = existingPortalApp.Notifications
		portalApp.Whitelists = existingPortalApp.Whitelists
	}

	if portalApp.Deleted {
		delete(c.portalApps, portalApp.ID)
		for aatID := range portalApp.AATs {
			delete(c.portalAppsByProtocolAppID, aatID)
		}

		if account, ok := c.accounts[portalApp.AccountID]; ok {
			delete(account.PortalApps, portalApp.ID)
		}

		for _, account := range c.accounts {
			for _, user := range account.Users {
				delete(user.PortalAppRoles, portalApp.ID)
				delete(user.PortalAppsAccepted, portalApp.ID)
			}
		}

		return
	}

	if pendingAAT, ok := c.pendingPortalAppAATs[portalApp.ID]; ok {
		if portalApp.AATs == nil {
			portalApp.AATs = make(map[types.ProtocolAppID]types.AAT)
		}
		portalApp.AATs[pendingAAT.ID] = pendingAAT
		delete(c.pendingPortalAppAATs, portalApp.ID)
	}

	if pendingSettings, ok := c.pendingSettings[portalApp.ID]; ok {
		portalApp.Settings = pendingSettings
		delete(c.pendingSettings, portalApp.ID)
	}

	if pendingNotification, ok := c.pendingAppNotifications[portalApp.ID]; ok {
		if portalApp.Notifications == nil {
			portalApp.Notifications = make(map[types.NotificationType]types.AppNotification)
		}
		portalApp.Notifications[pendingNotification.Type] = pendingNotification
		delete(c.pendingAppNotifications, portalApp.ID)
	}

	if pendingWhitelists, ok := c.pendingWhitelists[portalApp.ID]; ok {
		for _, whitelist := range pendingWhitelists {
			portalApp.AddWhitelist(whitelist)
		}
		delete(c.pendingWhitelists, portalApp.ID)
	}

<<<<<<< HEAD
=======
	if pendingAccountUserAccess, ok := c.pendingAccountUserAccess[portalApp.AccountID]; ok {
		for _, accountUser := range pendingAccountUserAccess {
			if accountUser.Owner {
				if portalApp.Users == nil {
					portalApp.Users = make(map[types.UserID]*types.AccountUserAccess)
				}
				portalApp.Users[accountUser.UserID] = &accountUser
			}
		}
	}

	c.portalApps[portalApp.ID] = &portalApp
	for aatID := range portalApp.AATs {
		c.portalAppsByProtocolAppID[aatID] = &portalApp
	}

>>>>>>> 807487a9148e2c17fd6c0d9018a69d66a00f8742
	if account, ok := c.accounts[portalApp.AccountID]; ok {
		// Assign new app to account's PortalApps map
		if account.PortalApps == nil {
			account.PortalApps = make(map[types.PortalAppID]*types.PortalApp)
		}
		account.PortalApps[portalApp.ID] = &portalApp

		// Set account owner to OWNER role for new app
		if ownerID, err := account.GetOwnerID(); err == nil {
			if owner, ok := account.Users[ownerID]; ok {
				if owner.PortalAppRoles == nil {
					owner.PortalAppRoles = make(map[types.PortalAppID]types.RoleName)
				}
				owner.PortalAppRoles[portalApp.ID] = types.RoleOwner
				owner.PortalAppsAccepted[portalApp.ID] = true

				// Add owner to user permissions map
				if user, ok := c.users[ownerID]; ok {
					if user.ArePermissionsEmpty() {
						user.Permissions = make(map[types.PortalAppID]types.PortalAppPermissions)
					}

					_, err = user.UpsertPermissions(portalApp.ID, account.ID, types.RoleOwner)
					if err != nil {
						fmt.Printf("error upserting user permissions: %v\n", err)
					}
				}
			}
		}
	}

	c.portalApps[portalApp.ID] = &portalApp
	for aatID := range portalApp.AATs {
		c.portalAppsByProtocolAppID[aatID] = &portalApp
	}
}

// addPortalAppAAT adds or updates the AAT for a PortalApp in cache
func (c *Cache) addPortalAppAAT(aat types.AAT) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := aat.AppID
	aat.AppID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	c.pendingPortalAppAATs[appID] = aat

	portalApp, ok := c.portalApps[appID]
	if !ok {
		return
	}

	if portalApp.AATs == nil {
		portalApp.AATs = make(map[types.ProtocolAppID]types.AAT)
	}
	portalApp.AATs[aat.ID] = aat
}

// addPortalAppSettings adds or updates the Settings for a PortalApp in cache
func (c *Cache) addPortalAppSettings(settings types.Settings) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := settings.AppID
	settings.AppID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	c.pendingSettings[appID] = settings

	portalApp, ok := c.portalApps[appID]
	if !ok { // Add to pending map in case Settings notification arrives first
		return
	}

	portalApp.Settings = settings
}

// addPortalAppSettings adds or updates the AppNotifications for a PortalApp in cache
func (c *Cache) addAppNotification(notification types.AppNotification) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := notification.AppID
	notification.AppID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	c.pendingAppNotifications[appID] = notification

	portalApp, ok := c.portalApps[appID]
	if !ok {
		return
	}

	if portalApp.Notifications == nil {
		portalApp.Notifications = make(map[types.NotificationType]types.AppNotification)
	}
	portalApp.Notifications[notification.Type] = notification
}

// Adds or updates a Whitelist for a PortalApp in cache
func (c *Cache) addWhitelist(whitelist types.Whitelist) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := whitelist.AppID
	whitelist.AppID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	c.pendingWhitelists[appID] = append(c.pendingWhitelists[appID], whitelist)

	portalApp, ok := c.portalApps[appID]
	if !ok {
		return
	}

	portalApp.AddWhitelist(whitelist)
}

// Removes a Whitelist from a PortalApp in cache
func (c *Cache) deleteWhitelist(whitelist types.Whitelist) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := whitelist.AppID
	whitelist.AppID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	portalApp, ok := c.portalApps[appID]
	if !ok {
		return
	}

	portalApp.DeleteWhitelist(whitelist)
}

/* ---------- GigastakeApps ---------- */

// GetGigastakeAppByID returns GigastakeApp from cache by aatID
func (c *Cache) GetGigastakeAppByID(gigastakeAppID types.GigastakeAppID) types.GigastakeApp {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	gigastakeApp, ok := c.gigastakeApps[gigastakeAppID]
	if !ok {
		return types.GigastakeApp{}
	}

	return *gigastakeApp
}

// GetGigastakeApps returns all GigastakeApps in cache
func (c *Cache) GetGigastakeApps(options Options) []types.GigastakeApp {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var gigastakeAppsToReturn []types.GigastakeApp

	for _, gigastakeApp := range c.gigastakeApps {
		if options.IncludeDeleted || !gigastakeApp.Deleted {
			gigastakeAppsToReturn = append(gigastakeAppsToReturn, *gigastakeApp)
		}
	}

	return gigastakeAppsToReturn
}

// upsertGigastakeApp adds, updates or deletes GigastakeApp in cache
func (c *Cache) upsertGigastakeApp(gigastakeApp types.GigastakeApp) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	if gigastakeApp.Deleted {
		delete(c.gigastakeApps, gigastakeApp.ID)
		return
	}

	if existingGigastakeApp, ok := c.gigastakeApps[gigastakeApp.ID]; ok {
		gigastakeApp.ChainIDs = existingGigastakeApp.ChainIDs
	}

	if pendingChainGigastake, ok := c.pendingChainsGigastakeApps[gigastakeApp.ID]; ok {
		chainID := pendingChainGigastake.ChainID

		if gigastakeApp.ChainIDs == nil {
			gigastakeApp.ChainIDs = make(map[types.RelayChainID]struct{})
		}
		gigastakeApp.ChainIDs[chainID] = struct{}{}

		delete(c.pendingChainsGigastakeApps, gigastakeApp.ID)
	}

	c.gigastakeApps[gigastakeApp.ID] = &gigastakeApp

	for chainID := range gigastakeApp.ChainIDs {
		if chain, ok := c.chains[chainID]; ok {
			if chain.GigastakeApps == nil {
				chain.GigastakeApps = make(map[types.GigastakeAppID]*types.GigastakeApp)
			}
			chain.GigastakeApps[gigastakeApp.ID] = &gigastakeApp
			continue
		}

		c.pendingGigastakeApps[chainID] = append(c.pendingGigastakeApps[chainID], gigastakeApp)
	}
}

// addChainGigastakeApp adds the AAT for a GigastakeApp to the cache
func (c *Cache) addChainGigastakeApp(chainGigastake types.ChainGigastakeApp) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	gigastakeAppID := chainGigastake.GigastakeAppID
	chainID := chainGigastake.ChainID

	gigastakeApp, ok := c.gigastakeApps[gigastakeAppID]
	if !ok {
		c.pendingChainsGigastakeApps[gigastakeAppID] = chainGigastake
		return
	}

	if gigastakeApp.ChainIDs == nil {
		gigastakeApp.ChainIDs = make(map[types.RelayChainID]struct{})
	}
	gigastakeApp.ChainIDs[chainID] = struct{}{}

	if chain, ok := c.chains[chainID]; ok {
		if chain.GigastakeApps == nil {
			chain.GigastakeApps = make(map[types.GigastakeAppID]*types.GigastakeApp)
		}
		chain.GigastakeApps[gigastakeApp.ID] = gigastakeApp
		return
	}

	c.pendingGigastakeApps[chainID] = append(c.pendingGigastakeApps[chainID], *gigastakeApp)
}

/* ---------- Chains ---------- */

// GetChain returns Chain from cache by chainID
func (c *Cache) GetChain(chainID types.RelayChainID, options Options) types.Chain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	chain, ok := c.chains[chainID]
	if !ok {
		return types.Chain{}
	}

	if options.ExcludeChainGigastakeApps {
		return chain.ClearGigastakeApps()
	}

	return *chain
}

// GetChains returns all Chains from cache
func (c *Cache) GetChains(options Options) []types.Chain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var chainsToReturn []types.Chain

	for _, chain := range c.chains {
		if options.IncludeDeleted || !chain.Deleted {
			if options.IncludeInactiveChains || chain.Active {
				chainToReturn := *chain

				if options.ExcludeChainGigastakeApps {
					chainToReturn = chainToReturn.ClearGigastakeApps()
				}

				chainsToReturn = append(chainsToReturn, chainToReturn)
			}
		}
	}

	return chainsToReturn
}

// upsertChain adds, updates or deletes Chain saved in cache
func (c *Cache) upsertChain(chain types.Chain) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	if existingChain, ok := c.chains[chain.ID]; ok {
		chain.Altruists = existingChain.Altruists
		chain.Aliases = existingChain.Aliases
		chain.Checks = existingChain.Checks
		chain.GigastakeApps = existingChain.GigastakeApps
		// DEPRECATED - TODO remove when move to only store aliases is complete
		chain.AliasDomains = existingChain.AliasDomains
	}

	if pendingAltruists, ok := c.pendingAltruists[chain.ID]; ok {
		if chain.Altruists == nil {
			chain.Altruists = make(map[types.AltruistURL]types.Altruist)
		}
		for _, pendingAltruist := range pendingAltruists {
			chain.Altruists[pendingAltruist.URL] = pendingAltruist
		}
		delete(c.pendingAltruists, chain.ID)
	}

	if pendingChecks, ok := c.pendingChecks[chain.ID]; ok {
		if chain.Checks == nil {
			chain.Checks = make(map[types.ChainCheckType]types.Check)
		}
		for _, check := range pendingChecks {
			chain.Checks[check.Type] = check
		}
		delete(c.pendingChecks, chain.ID)
	}

	if pendingAliases, ok := c.pendingAliases[chain.ID]; ok {
		if chain.Aliases == nil {
			chain.Aliases = make(map[types.ChainAlias]struct{})
		}
		for _, alias := range pendingAliases {
			chain.Aliases[alias] = struct{}{}
		}
		delete(c.pendingAliases, chain.ID)
	}

	// DEPRECATED - TODO remove when move to only store aliases is complete
	if pendingAliasDomains, ok := c.pendingAliasDomains[chain.ID]; ok {
		if chain.AliasDomains == nil {
			chain.AliasDomains = make(map[types.ChainAlias][]types.ChainDomain)
		}
		for _, aliasDomain := range pendingAliasDomains {
			chain.AliasDomains[aliasDomain.Alias] = aliasDomain.Domains
		}
		delete(c.pendingAliasDomains, chain.ID)
	}

	if pendingGigastakeApps, ok := c.pendingGigastakeApps[chain.ID]; ok {
		if chain.GigastakeApps == nil {
			chain.GigastakeApps = make(map[types.GigastakeAppID]*types.GigastakeApp)
		}
		for _, gigastakeApp := range pendingGigastakeApps {
			chain.GigastakeApps[gigastakeApp.ID] = &gigastakeApp
		}
		delete(c.pendingGigastakeApps, chain.ID)
	}

	c.chains[chain.ID] = &chain
}

// containsDomain is a helper func that checks if a domain is already in a list of domains
func containsDomain(domains []types.ChainDomain, domain types.ChainDomain) bool {
	for _, d := range domains {
		if d == domain {
			return true
		}
	}
	return false
}

// addAltruist adds or updates Altruist saved in cache
func (c *Cache) addAltruist(altruist types.Altruist) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	chainID := altruist.ChainID
	altruist.ChainID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	if chain, ok := c.chains[chainID]; ok {
		if chain.Altruists == nil {
			chain.Altruists = make(map[types.AltruistURL]types.Altruist)
		}
		chain.Altruists[altruist.URL] = altruist
		return
	}

	c.pendingAltruists[chainID] = append(c.pendingAltruists[chainID], altruist)
}

// addAliasDomains adds or updates Alias saved in cache
func (c *Cache) addAlias(alias types.Alias) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	chainID := alias.ChainID
	alias.ChainID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	if chain, ok := c.chains[chainID]; ok {
		if chain.Aliases == nil {
			chain.Aliases = make(map[types.ChainAlias]struct{})
		}
		chain.Aliases[alias.Alias] = struct{}{}
		return
	}

	c.pendingAliases[chainID] = append(c.pendingAliases[chainID], alias.Alias)
}

// DEPRECATED - TODO remove when move to only store aliases is complete
// addAliasDomains adds or updates AliasDomains saved in cache
func (c *Cache) addAliasDomains(aliasDomains types.AliasDomains) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	chainID := aliasDomains.ChainID
	aliasDomains.ChainID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	domainExists := func(aliasDomainsMap map[types.ChainAlias][]types.ChainDomain, newAliasDomains types.AliasDomains) bool {
		existingDomains, ok := aliasDomainsMap[newAliasDomains.Alias]
		if !ok {
			return false
		}

		for _, newDomain := range newAliasDomains.Domains {
			if !containsDomain(existingDomains, newDomain) {
				aliasDomainsMap[newAliasDomains.Alias] = append(aliasDomainsMap[newAliasDomains.Alias], newDomain)
				return true
			}
		}
		return false
	}

	if chain, ok := c.chains[chainID]; ok {
		if !domainExists(chain.AliasDomains, aliasDomains) {
			if chain.AliasDomains == nil {
				chain.AliasDomains = make(map[types.ChainAlias][]types.ChainDomain)
			}
			chain.AliasDomains[aliasDomains.Alias] = aliasDomains.Domains
		}
		return
	}

	c.pendingAliasDomains[chainID] = append(c.pendingAliasDomains[chainID], aliasDomains)
}

// deleteAltruist removes an Altruist from cache
func (c *Cache) deleteAltruist(altruist types.Altruist) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	chainID := altruist.ChainID

	if chain, ok := c.chains[chainID]; ok {
		delete(chain.Altruists, altruist.URL)
	}
}

// deleteCheck removes a Check from cache
func (c *Cache) deleteCheck(check types.Check) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	chainID := check.ChainID

	if chain, ok := c.chains[chainID]; ok {
		delete(chain.Checks, check.Type)
	}
}

// deleteAliasDomains removes AliasDomains from cache
func (c *Cache) deleteAlias(alias types.Alias) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	chainID := alias.ChainID

	if chain, ok := c.chains[chainID]; ok {
		delete(chain.Aliases, alias.Alias)
	}
}

// DEPRECATED - TODO remove when move to only store aliases is complete
// deleteAliasDomains removes AliasDomains from cache
func (c *Cache) deleteAliasDomains(aliasDomains types.AliasDomains) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	chainID := aliasDomains.ChainID

	if chain, ok := c.chains[chainID]; ok {
		delete(chain.AliasDomains, aliasDomains.Alias)
	}
}

// addCheck adds or updates Check saved in cache
func (c *Cache) addCheck(check types.Check) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	chainID := check.ChainID
	check.ChainID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	if chain, ok := c.chains[chainID]; ok {
		if chain.Checks == nil {
			chain.Checks = make(map[types.ChainCheckType]types.Check)
		}
		chain.Checks[check.Type] = check
		return
	}

	c.pendingChecks[chainID] = append(c.pendingChecks[chainID], check)
}

/* ---------- Users ---------- */

// upsertUser updates or inserts user
func (c *Cache) upsertUser(user types.User) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	if existingUser, ok := c.users[user.ID]; ok {
		user.Permissions = existingUser.Permissions
		user.AuthProviders = existingUser.AuthProviders
	}

	if pendingUserAuth, ok := c.pendingUserAuthProviders[user.ID]; ok {
		if user.AuthProviders == nil {
			user.AuthProviders = make(map[types.AuthType]types.UserAuthProvider)
		}
		user.AuthProviders[pendingUserAuth.Type] = pendingUserAuth
		delete(c.pendingUserAuthProviders, user.ID)
	}

	c.users[user.ID] = &user
	for _, userAuthProvider := range user.AuthProviders {
		c.usersByProviderID[userAuthProvider.ProviderUserID] = &user
	}

	// Pending users map is used to update user info in accounts
	c.pendingUsers[user.ID] = user

	accountsForUser, ok := c.accountsByUserID[user.ID]
	if !ok {
		return
	}

	for accountID, account := range accountsForUser {
		if accountUser, ok := account.Users[user.ID]; ok {
			accountUser.Email = user.Email
			account.Users[user.ID] = accountUser
			c.accountsByUserID[user.ID][accountID] = account
		}
	}
}

func (c *Cache) addUserAuthProvider(userAuth types.UserAuthProvider) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	userID := userAuth.UserID
	userAuth.UserID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	if user, ok := c.users[userID]; ok {
		if user.AuthProviders == nil {
			user.AuthProviders = make(map[types.AuthType]types.UserAuthProvider)
		}

		userAuthProviders := user.AuthProviders
		userAuthProviders[userAuth.Type] = userAuth
		user.AuthProviders = userAuthProviders

		c.users[user.ID] = user
		c.usersByProviderID[userAuth.ProviderUserID] = user

		return
	}

	c.pendingUserAuthProviders[userID] = userAuth
}

// deleteUser deletes a user
func (c *Cache) deleteUser(user types.User) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	if cacheUser, ok := c.users[user.ID]; ok {
		delete(c.users, cacheUser.ID)
		for _, userAuthProvider := range cacheUser.AuthProviders {
			delete(c.usersByProviderID, userAuthProvider.ProviderUserID)
		}
	}
}

/* ---------- Plans ---------- */

// GetPlan returns Plan from cache by planType
func (c *Cache) GetPlan(planType types.PayPlanType) types.Plan {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	plan, ok := c.plans[planType]
	if !ok {
		return types.Plan{}
	}

	return *plan
}

// GetPlans returns all PayPlans in cache
func (c *Cache) GetPlans() []types.Plan {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var plansToReturn []types.Plan

	for _, plan := range c.plans {
		plansToReturn = append(plansToReturn, *plan)
	}

	return plansToReturn
}
