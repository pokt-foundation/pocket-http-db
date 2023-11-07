// package cache stores data in memeory to reduce database calls and increase speed of data fetching
package cache

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/driver"
	"github.com/pokt-foundation/portal-http-db/v2/types"
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
		usersByAccountID  map[types.AccountID]UsersForAccount

		accounts                    map[types.AccountID]*types.Account
		pendingAccountNotifications map[types.AccountID]types.AccountNotification
		accountsByUserID            map[types.UserID]AccountsForUser
		pendingAccountUsers         map[types.UserID][]types.AccountUser
		pendingUserAuthProviders    map[types.UserID]types.UserAuthProvider

		portalApps                map[types.PortalAppID]*types.PortalApp
		portalAppsByProtocolAppID map[types.ProtocolAppID]*types.PortalApp
		pendingPortalAppAATs      map[types.PortalAppID]types.AAT
		pendingSettings           map[types.PortalAppID]types.Settings
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

	UsersForAccount map[types.UserID]*types.User
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
		FilterAccounts            []types.RoleName
		Accepted                  *bool
	}

	userData struct {
		users             map[types.UserID]*types.User
		usersByProviderID map[types.ProviderUserID]*types.User
		usersByAccountID  map[types.AccountID]UsersForAccount
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

		pendingAccountNotifications: make(map[types.AccountID]types.AccountNotification),
		pendingAccountUsers:         make(map[types.UserID][]types.AccountUser),
		pendingUserAuthProviders:    make(map[types.UserID]types.UserAuthProvider),
		pendingPortalAppAATs:        make(map[types.PortalAppID]types.AAT),
		pendingSettings:             make(map[types.PortalAppID]types.Settings),
		pendingWhitelists:           make(map[types.PortalAppID][]types.Whitelist),
		pendingAltruists:            make(map[types.RelayChainID][]types.Altruist),
		pendingChainsGigastakeApps:  make(map[types.GigastakeAppID]types.ChainGigastakeApp),
		pendingAliases:              make(map[types.RelayChainID][]types.ChainAlias),
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
	appsData, err := c.fetchAppsData(ctx, userData.usersByAccountID)
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
	c.usersByAccountID = userData.usersByAccountID

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
	usersByAccountID := make(map[types.AccountID]UsersForAccount)

	for _, user := range users {
		for _, authProvider := range user.AuthProviders {
			usersByProviderID[authProvider.ProviderUserID] = user
		}

		for accountID := range user.Permissions {
			if _, ok := usersByAccountID[accountID]; !ok {
				usersByAccountID[accountID] = make(map[types.UserID]*types.User)
			}
			usersByAccountID[accountID][user.ID] = user
		}
	}

	return userData{users, usersByProviderID, usersByAccountID}, nil
}

// fetchAppsData reads Plans, Accounts and PortalApps from the database,
// sets each Account's Plan and PortalApps and creates a map of Accounts by UserID.
// also creates a map of PortalApps by ProtocolAppID
func (c *Cache) fetchAppsData(ctx context.Context, usersByAccountID map[types.AccountID]UsersForAccount) (appsData, error) {
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
	for accountID, usersForAccount := range usersByAccountID {
		if account, ok := accounts[accountID]; ok {
			if account.Deleted {
				continue
			}

			for userID := range usersForAccount {
				if _, ok := accountsByUserID[userID]; !ok {
					accountsByUserID[userID] = make(map[types.AccountID]*types.Account)
				}
				accountsByUserID[userID][account.ID] = account
			}
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

			if accountCopy.PortalApps == nil {
				accountCopy.PortalApps = map[types.PortalAppID]*types.PortalApp{}
			}

			if usersForAccount, ok := c.usersByAccountID[account.ID]; ok {
				accountCopy.AssignAccountUsers(usersForAccount)
			}

			accountsToReturn = append(accountsToReturn, accountCopy)
		}
	}

	return accountsToReturn
}

// GetAccounts returns one Accounts from cache
func (c *Cache) GetAccount(accountID types.AccountID) types.Account {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	if account, ok := c.accounts[accountID]; ok {
		return *account
	}

	return types.Account{}
}

// GetUserAccount returns one Account by accountID with portal apps filtered by user ID
func (c *Cache) GetUserAccount(accountID types.AccountID, userID types.UserID, options Options) types.Account {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	if account, ok := c.accounts[accountID]; ok {
		if c.userAcceptedAccountInvite(account, userID) {
			accountCopy := *account

			if accountCopy.PortalApps == nil {
				accountCopy.PortalApps = map[types.PortalAppID]*types.PortalApp{}
			}

			if usersForAccount, ok := c.usersByAccountID[account.ID]; ok {
				accountCopy.AssignAccountUsers(usersForAccount)
			}

			return accountCopy
		}
	}

	return types.Account{}
}

func (c *Cache) GetUserAccounts(userID types.UserID, options Options) []types.Account {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var accountsToReturn []types.Account

	for _, account := range c.accountsByUserID[userID] {
		// Skip if the accepted filter is passed and user's accepted status doesn't match
		if options.Accepted != nil {
			if c.userAcceptedAccountInvite(account, userID) != *options.Accepted {
				continue
			}
		}

		// Skip if role name filters are passed and the user's role isn't in the filters slice
		if options.FilterAccounts != nil && len(options.FilterAccounts) > 0 {
			if accountUsers, ok := c.usersByAccountID[account.ID]; ok {
				if user, ok := accountUsers[userID]; ok {
					userRole := user.GetAccountRole(account.ID)
					if !slices.Contains(options.FilterAccounts, userRole) {
						continue
					}
				}
			}
		}

		accountCopy := *account

		if accountCopy.PortalApps == nil {
			accountCopy.PortalApps = map[types.PortalAppID]*types.PortalApp{}
		}

		if usersForAccount, ok := c.usersByAccountID[account.ID]; ok {
			accountCopy.AssignAccountUsers(usersForAccount)
		}

		accountsToReturn = append(accountsToReturn, accountCopy)
	}

	return accountsToReturn
}

// userAcceptedAccountInvite returns a boolean indicating if the user
// has accepted an invitation to the Account
func (c *Cache) userAcceptedAccountInvite(a *types.Account, userID types.UserID) bool {
	if usersForAccount, ok := c.usersByAccountID[a.ID]; ok {
		if user, ok := usersForAccount[userID]; ok {
			if accountPermissions, ok := user.Permissions[a.ID]; ok {
				return accountPermissions.Accepted
			}
		}
	}

	return false
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
			delete(c.accountsByUserID, user.ID)
		}
		return
	}

	// Ensure upsert does not overwrite existing Account fields
	if existingAccount, ok := c.accounts[account.ID]; ok {
		account.PortalApps = existingAccount.PortalApps
		account.Integrations = existingAccount.Integrations
		account.Notifications = existingAccount.Notifications
	}

	if pendingNotification, ok := c.pendingAccountNotifications[account.ID]; ok {
		if account.Notifications == nil {
			account.Notifications = make(map[types.NotificationType]types.AccountNotification)
		}
		account.Notifications[pendingNotification.Type] = pendingNotification
		delete(c.pendingAccountNotifications, account.ID)
	}

	// Set plan in cache to latest value
	if plan, ok := c.plans[account.PlanType]; ok {
		account.Plan = plan
	}
	c.accounts[account.ID] = &account

	if usersForAccount, ok := c.usersByAccountID[account.ID]; ok {
		for _, user := range usersForAccount {
			if _, ok := c.accountsByUserID[user.ID]; !ok {
				c.accountsByUserID[user.ID] = make(map[types.AccountID]*types.Account)
			}
			c.accountsByUserID[user.ID][account.ID] = &account
		}
	}
}

// addAccountNotification adds or updates the AccountNotifications for an Account in cache
func (c *Cache) addAccountNotification(notification types.AccountNotification) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	accountID := notification.AccountID
	notification.AccountID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	if account, ok := c.accounts[accountID]; ok {
		if account.Notifications == nil {
			account.Notifications = make(map[types.NotificationType]types.AccountNotification)
		}
		account.Notifications[notification.Type] = notification
		return
	}

	c.pendingAccountNotifications[accountID] = notification
}

// upsertAccount adds or updates AccountUser in cache
func (c *Cache) upsertAccountUser(accountUser types.AccountUser) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	if user, ok := c.users[accountUser.ID]; ok {
		c.updateUserAccountData(accountUser, user)
		return
	}

	c.pendingAccountUsers[accountUser.ID] = append(c.pendingAccountUsers[accountUser.ID], accountUser)
}

func (c *Cache) updateUserAccountData(accountUser types.AccountUser, user *types.User) {
	accountID := accountUser.AccountID
	accountUser.AccountID = ""

	if user.ArePermissionsEmpty() {
		user.Permissions = make(map[types.AccountID]types.AccountPermissions)
	}

	_, err := user.UpsertPermissions(accountID, accountUser.RoleName, accountUser.Accepted)
	if err != nil {
		c.logger.Error(fmt.Sprintf("error upserting user permissions: %v\n", err))
	}

	// Add user to users by account ID map
	if _, ok := c.usersByAccountID[accountID]; !ok {
		c.usersByAccountID[accountID] = make(map[types.UserID]*types.User)
	}
	c.usersByAccountID[accountID][user.ID] = user

	// Add or update user's Accounts
	if account, ok := c.accounts[accountID]; ok {
		if _, ok := c.accountsByUserID[accountUser.ID]; !ok {
			c.accountsByUserID[accountUser.ID] = make(map[types.AccountID]*types.Account)
		}

		c.accountsByUserID[accountUser.ID][account.ID] = account

		return
	}
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
func (c *Cache) deleteAccountUser(accountUser types.AccountUser) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	accountID := accountUser.AccountID
	userID := accountUser.ID
	accountUser.AccountID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	if accountUser.RoleName == types.RoleOwner {
		c.logger.Warn("cannot delete owner from account")
		return
	}

	// Remove user permissions for account
	if user, ok := c.users[userID]; ok {
		if !user.ArePermissionsEmpty() {
			user.DeletePermissions(accountID)
		}
	}

	// Remove from Accounts map
	if account, ok := c.accounts[accountID]; ok {
		delete(account.Users, userID)

		if accountsForUser, ok := c.accountsByUserID[accountUser.ID]; ok {
			delete(accountsForUser, accountID)
		}
	}

	// Remove from users by account ID map
	if usersByAccountID, ok := c.usersByAccountID[accountID]; ok {
		delete(usersByAccountID, userID)
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

	return *portalApp
}

// GetPortalAppByProtocolAppID returns PortalApp from cache by protocolAppID
func (c *Cache) GetPortalAppByProtocolAppID(protocolAppID types.ProtocolAppID) types.PortalApp {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalApp, ok := c.portalAppsByProtocolAppID[protocolAppID]
	if !ok {
		return types.PortalApp{}
	}

	return *portalApp
}

// GetPortalApps returns all PortalApps in cache
func (c *Cache) GetPortalApps(options Options) []types.PortalApp {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var portalAppsToReturn []types.PortalApp

	for _, portalApp := range c.portalApps {
		if options.IncludeDeleted || !portalApp.Deleted {
			portalAppsToReturn = append(portalAppsToReturn, *portalApp)
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
			if !portalApp.IsDirectApp() {
				if account, ok := c.accounts[portalApp.AccountID]; ok {
					plan = account.Plan
				}
			}

			portalAppLites = append(portalAppLites, portalApp.ConvertPortalAppToPortalAppLite(plan))
		}
	}

	return portalAppLites
}

func (c *Cache) GetPortalAppsByUserID(userID types.UserID) []types.PortalApp {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var portalAppsToReturn []types.PortalApp

	for _, account := range c.GetUserAccounts(userID, Options{}) {
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

	if pendingWhitelists, ok := c.pendingWhitelists[portalApp.ID]; ok {
		for _, whitelist := range pendingWhitelists {
			portalApp.AddWhitelist(whitelist)
		}
		delete(c.pendingWhitelists, portalApp.ID)
	}

	if account, ok := c.accounts[portalApp.AccountID]; ok {
		// Assign new app to account's PortalApps map
		if account.PortalApps == nil {
			account.PortalApps = make(map[types.PortalAppID]*types.PortalApp)
		}
		account.PortalApps[portalApp.ID] = &portalApp
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

	if pendingAccountUsers, ok := c.pendingAccountUsers[user.ID]; ok {
		for _, accountUser := range pendingAccountUsers {
			c.updateUserAccountData(accountUser, &user)
		}
		delete(c.pendingAccountUsers, user.ID)
	}

	c.users[user.ID] = &user
	for _, userAuthProvider := range user.AuthProviders {
		c.usersByProviderID[userAuthProvider.ProviderUserID] = &user
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
