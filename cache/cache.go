// package cache stores data in memeory to reduce database calls and increase speed of data fetching
package cache

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pokt-foundation/portal-db/v2/driver"
	"github.com/pokt-foundation/portal-db/v2/types"
	"github.com/sirupsen/logrus"
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
		log     *logrus.Logger
		rwMutex sync.RWMutex

		// Cached Data
		userPermissions map[types.UserID]*types.UserPermissions
		userIDs         map[types.ProviderUserID]types.UserID

		accounts                 map[types.AccountID]*types.Account
		accountsByUserID         map[types.UserID]AccountsForUser
		pendingAccountUserAccess map[types.AccountID][]types.AccountUserAccess
		pendingUsers             map[types.UserID]types.User

		portalApps              map[types.PortalAppID]*types.PortalApp
		pendingPortalAppAATs    map[types.PortalAppID]types.AAT
		pendingSettings         map[types.PortalAppID]types.Settings
		pendingAppNotifications map[types.PortalAppID]types.AppNotification
		pendingWhitelists       map[types.PortalAppID][]types.Whitelist

		gigastakeApps              map[types.GigastakeAppID]*types.GigastakeApp
		pendingChainsGigastakeApps map[types.GigastakeAppID]types.ChainGigastakeApp

		chains               map[types.RelayChainID]*types.Chain
		pendingAltruists     map[types.RelayChainID][]types.Altruist
		pendingAliasDomains  map[types.RelayChainID][]types.AliasDomains
		pendingChecks        map[types.RelayChainID][]types.Check
		pendingGigastakeApps map[types.RelayChainID][]types.GigastakeApp

		plans map[types.PayPlanType]*types.Plan

		blockedContracts types.GlobalBlockedContracts
	}

	AccountsForUser map[types.AccountID]*types.Account

	Config struct {
		Reader              driver.Reader
		Logger              *logrus.Logger
		CacheRefreshMinutes time.Duration
	}

	Options struct {
		IncludeDeleted,
		ExcludeChainGigastakeApps bool
	}
)

// NewCache returns cache instance from reader interface and starts the cache refresh handler and listener in separate goroutines
func NewCache(ctx context.Context, config Config) (*Cache, error) {
	cache := &Cache{
		reader: config.Reader,
		log:    config.Logger,

		pendingAccountUserAccess:   make(map[types.AccountID][]types.AccountUserAccess),
		pendingUsers:               make(map[types.UserID]types.User),
		pendingPortalAppAATs:       make(map[types.PortalAppID]types.AAT),
		pendingSettings:            make(map[types.PortalAppID]types.Settings),
		pendingAppNotifications:    make(map[types.PortalAppID]types.AppNotification),
		pendingWhitelists:          make(map[types.PortalAppID][]types.Whitelist),
		pendingAltruists:           make(map[types.RelayChainID][]types.Altruist),
		pendingChainsGigastakeApps: make(map[types.GigastakeAppID]types.ChainGigastakeApp),
		pendingAliasDomains:        make(map[types.RelayChainID][]types.AliasDomains),
		pendingChecks:              make(map[types.RelayChainID][]types.Check),
		pendingGigastakeApps:       make(map[types.RelayChainID][]types.GigastakeApp),
	}

	startTime := time.Now()
	err := cache.setCache(ctx)
	elapsedTime := time.Since(startTime).Milliseconds()
	if err != nil {
		cache.log.Error(fmt.Errorf("error in setCache: %w", err))
		return nil, err
	}

	cache.log.WithFields(
		logrus.Fields{
			"cache_refresh_minutes": config.CacheRefreshMinutes.Minutes(),
			"db_time_to_fetch":      fmt.Sprintf("%dms", elapsedTime),
		},
	).Info("cache successfully set")

	go cache.cacheRefreshHandler(ctx, config.CacheRefreshMinutes)

	go cache.listen()

	return cache, nil
}

/* ---------- Cache Set and Refresh ---------- */

// ResetCache manually resets the cache; called in the router when the `/reset_cache` endpoint is hit.
func (c *Cache) ResetCache(ctx context.Context) error {
	err := c.setCache(ctx)
	if err != nil {
		c.log.Error(fmt.Errorf("error in ResetCache - setCache: %w", err))
		return err
	}

	c.log.Info("cache successfully manually reset")

	return nil
}

// setCache initializes the cache by setting user permissions, accounts and portal apps, chains, and plans.
func (c *Cache) setCache(ctx context.Context) error {
	// Perform all the database reads first without locking the mutex
	userPermissions, userIDs, err := c.fetchUserData(ctx)
	if err != nil {
		return err
	}

	plans, accounts, portalApps, accountsByUserID, err := c.fetchPlansAccountsAndPortalApps(ctx)
	if err != nil {
		return err
	}

	chains, gigastakeApps, err := c.fetchChainsAndGigastakeApps(ctx)
	if err != nil {
		return err
	}

	blockedContracts, err := c.fetchBlockedContracts(ctx)
	if err != nil {
		return err
	}

	// Lock the mutex only once data is fetched
	// to minimize the time the lock is held.
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	// Set the cache with the fetched data
	c.userPermissions = userPermissions
	c.userIDs = userIDs
	c.plans = plans
	c.accounts = accounts
	c.accountsByUserID = accountsByUserID
	c.portalApps = portalApps
	c.chains = chains
	c.gigastakeApps = gigastakeApps
	c.blockedContracts = blockedContracts

	return nil
}

// fetchUserData reads UserPermissions from the database and returns them.
func (c *Cache) fetchUserData(ctx context.Context) (map[types.UserID]*types.UserPermissions, map[types.ProviderUserID]types.UserID, error) {
	userPermissions, err := c.reader.ReadUserPermissions(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("ReadUserPermissions - %w", err)
	}

	userIDs, err := c.reader.ReadUserIDsMap(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("ReadUserIDsMap - %w", err)
	}

	return userPermissions, userIDs, nil
}

// fetchPlansAccountsAndPortalApps reads Plans, Accounts and PortalApps from the database,
// sets each Account's Plan and PortalApps and creates a map of Accounts by UserID.
func (c *Cache) fetchPlansAccountsAndPortalApps(ctx context.Context) (map[types.PayPlanType]*types.Plan, map[types.AccountID]*types.Account, map[types.PortalAppID]*types.PortalApp, map[types.UserID]AccountsForUser, error) {
	// Fetch Plans, Accounts and PortalApps
	plans, err := c.reader.ReadPlans(ctx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("ReadPlans - %w", err)
	}
	accounts, err := c.reader.ReadAccounts(ctx, types.DriverOptions{IncludeDeleted: true})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("ReadAccounts - %w", err)
	}
	portalApps, err := c.reader.ReadPortalApps(ctx, types.DriverOptions{IncludeDeleted: true})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("ReadPortalApps - %w", err)
	}

	// Set pointer to Account Plan
	for _, account := range accounts {
		if plan, ok := plans[account.PlanType]; ok {
			account.Plan = plan
		} else {
			return nil, nil, nil, nil, fmt.Errorf("error setting account plan for account. plan type %s does not exist in cache", account.PlanType)
		}
	}

	// Set pointers to Account PortalApps
	for portalAppID, portalApp := range portalApps {
		if account, ok := accounts[portalApp.AccountID]; ok {
			if account.PortalApps == nil {
				account.PortalApps = make(map[types.PortalAppID]*types.PortalApp)
			}
			account.PortalApps[portalAppID] = portalApp
		}
	}

	// Create map of Accounts by UserID for all Account team members
	accountsByUserID := make(map[types.UserID]AccountsForUser)
	for _, account := range accounts {
		for _, user := range account.Users {
			if _, ok := accountsByUserID[user.UserID]; !ok {
				accountsByUserID[user.UserID] = make(map[types.AccountID]*types.Account)
			}
			accountsByUserID[user.UserID][account.ID] = account
		}
	}

	return plans, accounts, portalApps, accountsByUserID, nil
}

// fetchChainsAndGigastakeApps reads Chains and GigastakeApps from the database and returns them.
// It also sets pointers to GigastakeApps on Chains.
func (c *Cache) fetchChainsAndGigastakeApps(ctx context.Context) (map[types.RelayChainID]*types.Chain, map[types.GigastakeAppID]*types.GigastakeApp, error) {
	chains, err := c.reader.ReadChains(ctx, types.DriverOptions{IncludeDeleted: true})
	if err != nil {
		return nil, nil, fmt.Errorf("ReadChains - %w", err)
	}
	gigastakeApps, err := c.reader.ReadGigastakeApps(ctx, types.DriverOptions{IncludeDeleted: true})
	if err != nil {
		return nil, nil, fmt.Errorf("ReadGigastakeApps - %w", err)
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

	return chains, gigastakeApps, nil
}

// fetchBlockedContracts reads BlockedContracts from the database and returns them.
func (c *Cache) fetchBlockedContracts(ctx context.Context) (types.GlobalBlockedContracts, error) {
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
			c.log.Error("err in cacheRefreshHandler: %w", err)

			os.Exit(1) // Kill PHD if cache refresh fails to trigger pod restart
		}
	}
}

/* ---------- User Data ---------- */

// GetUserPermissionsByUserID returns all load balancer UserPermissions for a given UserID
func (c *Cache) GetUserPermissionsByUserID(userID types.UserID) types.UserPermissions {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	permission, ok := c.userPermissions[userID]
	if !ok {
		return types.UserPermissions{}
	}

	return *permission
}

// GetUserIDFromProviderUserID returns the Portal UserID for a given provider UserID
func (c *Cache) GetUserIDFromProviderUserID(providerUserID types.ProviderUserID) types.UserID {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	userID, ok := c.userIDs[providerUserID]
	if !ok {
		return types.UserID("")
	}

	return userID
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

// GetAccount returns Account by accountID
func (c *Cache) GetAccount(accountID types.AccountID) types.Account {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	account, ok := c.accounts[accountID]
	if !ok {
		return types.Account{}
	}

	return *account
}

// GetAccounts returns all Accounts in cache
// May select whether to include deleted Accounts. Default: false
func (c *Cache) GetAccounts(options Options) []types.Account {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var accountsToReturn []types.Account

	for _, account := range c.accounts {
		if options.IncludeDeleted || !account.Deleted {
			accountsToReturn = append(accountsToReturn, *account)
		}
	}

	return accountsToReturn
}

// GetAccounts returns all Accounts for a given user
func (c *Cache) GetAccountsByUserID(userID types.UserID) []types.Account {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var accountsToReturn []types.Account

	for _, account := range c.accountsByUserID[userID] {
		accountsToReturn = append(accountsToReturn, *account)
	}

	return accountsToReturn
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
	for portalAppID, role := range accountUser.PortalAppRoles {
		incomingPortalAppID = portalAppID
		incomingPortalAppRole = role
		break
	}

	// Add user to permissions map if they have accepted their invitation
	if accountUser.Accepted {
		userPermissions, ok := c.userPermissions[accountUser.UserID]
		if !ok {
			userPermissions = &types.UserPermissions{
				UserID:     accountUser.UserID,
				PortalApps: map[types.PortalAppID]types.PortalAppPermissions{},
			}
			c.userPermissions[accountUser.UserID] = userPermissions
		}

		switch accountUser.Owner {
		// If incoming account user is owner, set permissions for all account's portal apps
		case true:
			if account, ok := c.accounts[accountID]; ok {
				for portalAppID := range account.PortalApps {
					_, err := userPermissions.UpsertPermissions(portalAppID, types.RoleOwner)
					if err != nil {
						fmt.Printf("error upserting user permissions: %v\n", err)
					}
				}
			}

		// If incoming account user is not owner only set permissions for incoming portal app
		default:
			_, err := userPermissions.UpsertPermissions(incomingPortalAppID, incomingPortalAppRole)
			if err != nil {
				fmt.Printf("error upserting user permissions: %v\n", err)
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
		}

		account.Users[accountUser.UserID] = accountUser
		if user, ok := account.Users[accountUser.UserID]; ok {
			if user.PortalAppRoles == nil {
				user.PortalAppRoles = make(map[types.PortalAppID]types.RoleName)
			}

			switch accountUser.Owner {
			// If incoming account user is owner, set owner for all account's portal apps
			case true:
				for portalAppID := range account.PortalApps {
					user.PortalAppRoles[portalAppID] = types.RoleOwner
				}

			// If incoming account user is not owner only set role for incoming portal app
			default:
				user.PortalAppRoles[incomingPortalAppID] = incomingPortalAppRole
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

	if accountUser.Owner {
		return
	}

	// Remove user permissions for each PortalApp in the account
	if userPermissions, ok := c.userPermissions[userID]; ok {
		for portalAppID := range accountUser.PortalAppRoles {
			delete(userPermissions.PortalApps, portalAppID)
		}
	}

	// Remove from Accounts map
	account, ok := c.accounts[accountID]
	if ok {
		delete(account.Users, userID)

		if accountsForUser, ok := c.accountsByUserID[accountUser.UserID]; ok {
			delete(accountsForUser, accountID)
		}
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

	c.portalApps[portalApp.ID] = &portalApp

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

				// Add owner to user permissions map
				if userPermissions, ok := c.userPermissions[ownerID]; ok {
					_, err = userPermissions.UpsertPermissions(portalApp.ID, types.RoleOwner)
					if err != nil {
						fmt.Printf("error upserting user permissions: %v\n", err)
					}
				}
			}
		}
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
			chainToReturn := *chain

			if options.ExcludeChainGigastakeApps {
				chainToReturn = chainToReturn.ClearGigastakeApps()
			}

			chainsToReturn = append(chainsToReturn, chainToReturn)
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
		chain.AliasDomains = existingChain.AliasDomains
		chain.Checks = existingChain.Checks
		chain.GigastakeApps = existingChain.GigastakeApps
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

	c.userIDs[userAuth.ProviderUserID] = userAuth.UserID
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
