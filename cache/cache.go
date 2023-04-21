// package cache stores data in memeory to reduce database calls and increase speed of data fetching
package cache

import (
	"context"
	"fmt"
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
		portalAppsByProtocolID  map[types.ProtocolAppID]*types.PortalApp
		pendingAAT              map[types.PortalAppID]types.AAT
		pendingSettings         map[types.PortalAppID]types.Settings
		pendingAppNotifications map[types.PortalAppID]types.AppNotification
		pendingWhitelists       map[types.PortalAppID][]types.Whitelist
		pendingStickyOptions    map[types.PortalAppID]types.StickyOptions

		chains                    map[types.RelayChainID]*types.Chain
		pendingAltruists          map[types.RelayChainID][]types.Altruist
		pendingGigastakeRedirects map[types.RelayChainID][]types.GigastakeRedirect
		pendingChecks             map[types.RelayChainID][]types.Check

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
		IncludeDeleted          bool
		SetGigastakeRedirectApp bool
	}
)

// NewCache returns cache instance from reader interface and starts the cache refresh handler and listener in separate goroutines
func NewCache(ctx context.Context, config Config) (*Cache, error) {
	cache := &Cache{
		reader: config.Reader,
		log:    config.Logger,

		pendingAccountUserAccess:  make(map[types.AccountID][]types.AccountUserAccess),
		pendingUsers:              make(map[types.UserID]types.User),
		pendingAAT:                make(map[types.PortalAppID]types.AAT),
		pendingSettings:           make(map[types.PortalAppID]types.Settings),
		pendingAppNotifications:   make(map[types.PortalAppID]types.AppNotification),
		pendingStickyOptions:      make(map[types.PortalAppID]types.StickyOptions),
		pendingAltruists:          make(map[types.RelayChainID][]types.Altruist),
		pendingGigastakeRedirects: make(map[types.RelayChainID][]types.GigastakeRedirect),
		pendingChecks:             make(map[types.RelayChainID][]types.Check),
	}

	err := cache.setCache(ctx)
	if err != nil {
		cache.log.Error(fmt.Errorf("error in setCache: %w", err))
		return nil, err
	}

	cache.log.WithFields(
		logrus.Fields{"cache_refresh_minutes": config.CacheRefreshMinutes.Minutes()},
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
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	err := c.setUserData(ctx)
	if err != nil {
		return fmt.Errorf("setUserData - %w", err)
	}

	err = c.setPlansAccountsAndPortalApps(ctx)
	if err != nil {
		return fmt.Errorf("setAccountsAndPortalApps - %w", err)
	}

	err = c.setChains(ctx)
	if err != nil {
		return fmt.Errorf("setChains - %w", err)
	}

	err = c.setBlockedContracts(ctx)
	if err != nil {
		return fmt.Errorf("setBlockedContracts - %w", err)
	}

	return nil
}

// setUserPermissions reads UserPermissions from the database and sets them in the cache.
func (c *Cache) setUserData(ctx context.Context) error {
	userPermissions, err := c.reader.ReadUserPermissions(ctx)
	if err != nil {
		return fmt.Errorf("ReadUserPermissions - %w", err)
	}

	c.userPermissions = userPermissions

	userIDs, err := c.reader.ReadUserIDsMap(ctx)
	if err != nil {
		return fmt.Errorf("ReadUserIDsMap - %w", err)
	}
	c.userIDs = userIDs

	return nil
}

// setAccountsAndPortalApps reads Plans, Accounts and PortalApps from the database,
// sets each Account's Plan and PortalApps and creates a map of Accounts by UserID.
func (c *Cache) setPlansAccountsAndPortalApps(ctx context.Context) error {
	// Fetch Plans, Accounts and PortalApps
	plans, err := c.reader.ReadPlans(ctx)
	if err != nil {
		return fmt.Errorf("ReadPlans - %w", err)
	}
	accounts, err := c.reader.ReadAccounts(ctx, types.DriverOptions{IncludeDeleted: true})
	if err != nil {
		return fmt.Errorf("ReadAccounts - %w", err)
	}
	portalApps, err := c.reader.ReadPortalApps(ctx, types.DriverOptions{IncludeDeleted: true})
	if err != nil {
		return fmt.Errorf("ReadPortalApps - %w", err)
	}

	// Set pointer to Account Plan
	for _, account := range accounts {
		if plan, ok := plans[account.PlanType]; ok {
			account.Plan = plan
		} else {
			return fmt.Errorf("error setting account plan for account. plan type %s does not exist in cache", account.PlanType)
		}
	}

	// Set pointers to Account PortalApps and PortalApps by ProtocolID (applications.application_id in V1)
	portalAppsByProtocolID := make(map[types.ProtocolAppID]*types.PortalApp)

	for portalAppID, portalApp := range portalApps {
		if account, ok := accounts[portalApp.AccountID]; ok {
			if account.PortalApps == nil {
				account.PortalApps = make(map[types.PortalAppID]*types.PortalApp)
			}
			account.PortalApps[portalAppID] = portalApp
		}

		for protocolAppID := range portalApp.AATs {
			portalAppsByProtocolID[protocolAppID] = portalApp
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

	c.plans = plans
	c.accounts = accounts
	c.accountsByUserID = accountsByUserID
	c.portalApps = portalApps
	c.portalAppsByProtocolID = portalAppsByProtocolID

	return nil
}

// setChains reads Chains from the database and sets them in the cache.
func (c *Cache) setChains(ctx context.Context) error {
	chains, err := c.reader.ReadChains(ctx, types.DriverOptions{IncludeDeleted: true})
	if err != nil {
		return fmt.Errorf("ReadChains - %w", err)
	}

	c.chains = chains

	return nil
}

// setBlockedContracts reads BlockedContracts from the database and sets them in the cache.
func (c *Cache) setBlockedContracts(ctx context.Context) error {
	blockedContracts, err := c.reader.ReadBlockedContracts(ctx)
	if err != nil {
		return fmt.Errorf("ReadBlockedContracts - %w", err)
	}

	c.blockedContracts = blockedContracts

	return nil
}

// cacheRefreshHandler sets the cache from the DB at a regular interval (default is 5 minutes)
func (c *Cache) cacheRefreshHandler(ctx context.Context, cacheRefreshMinutes time.Duration) {
	for {
		<-time.After(cacheRefreshMinutes)

		err := c.setCache(ctx)
		if err != nil {
			c.log.Error("err in cacheRefreshHandler: %w", err)
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

// GetAccounts returns all Accounts for a given user
func (c *Cache) GetUserFromUserID(userID types.UserID) []types.AccountUserAccess {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	accounts := c.GetAccountsByUserID(userID)

	user := []types.AccountUserAccess{}
	for _, account := range accounts {
		userAccess := account.Users[userID]
		userAccess.AccountID = account.ID
		user = append(user, userAccess)
	}

	return user
}

// upsertAccount adds, updates or deletes Account in cache
func (c *Cache) upsertAccount(account types.Account) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	pendingUsers, ok := c.pendingAccountUserAccess[account.ID]
	if ok {
		for _, user := range pendingUsers {
			account.Users[user.UserID] = user
		}
		delete(c.pendingAccountUserAccess, account.ID)
	}

	c.accounts[account.ID] = &account

	for _, user := range account.Users {
		switch account.Deleted {
		// If marked as deleted, remove Account from all user ID maps of all Account users
		case true:
			delete(c.accountsByUserID, user.UserID)

		// If not deleted, set Account to all user ID maps of all Account users
		default:
			if _, ok := c.accountsByUserID[user.UserID]; !ok {
				c.accountsByUserID[user.UserID] = make(map[types.AccountID]*types.Account)
			}
			c.accountsByUserID[user.UserID][account.ID] = &account
		}
	}
}

// upsertAccount adds or updates AccountUserAccess in cache
func (c *Cache) upsertAccountUser(accountUser types.AccountUserAccess) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	accountID := accountUser.AccountID
	accountUser.AccountID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	// Add user to permissions map if they have accepted their invitation
	if accountUser.Accepted {
		userPermissions, ok := c.userPermissions[accountUser.UserID]
		if !ok {
			userPermissions = &types.UserPermissions{
				UserID:   accountUser.UserID,
				Accounts: map[types.AccountID]types.AccountPermissions{},
			}
			c.userPermissions[accountUser.UserID] = userPermissions
		}

		_, err := userPermissions.UpsertPermissions(accountID, accountUser.RoleName)
		if err != nil {
			fmt.Printf("error upserting user permissions: %v\n", err)
		}
	}

	// Add or update user's Accounts
	if account, ok := c.accounts[accountID]; ok {
		if account.Users == nil {
			account.Users = make(map[types.UserID]types.AccountUserAccess)
		}
		if user, ok := c.pendingUsers[accountUser.UserID]; ok {
			accountUser.Email = user.Email
			delete(c.pendingUsers, accountUser.UserID)
		}

		account.Users[accountUser.UserID] = accountUser

		if _, ok := c.accountsByUserID[accountUser.UserID]; !ok {
			c.accountsByUserID[accountUser.UserID] = make(map[types.AccountID]*types.Account)
		}
		c.accountsByUserID[accountUser.UserID][account.ID] = account

		return
	}

	c.pendingAccountUserAccess[accountID] = append(c.pendingAccountUserAccess[accountID], accountUser)
}

// deleteAccountUser deletes Account user saved in cache and updates permissions
func (c *Cache) deleteAccountUser(accountUser types.AccountUserAccess) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	accountID := accountUser.AccountID
	userID := accountUser.UserID
	accountUser.AccountID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	if userPermissions, ok := c.userPermissions[userID]; ok {
		delete(userPermissions.Accounts, accountID)
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

// TODO: Add cache test for CreateNewPortalUser since it's the first writer method in the new cache package
// CreateNewPortalUser creates a new portal user in the db
func (c *Cache) CreateNewPortalUser(ctx context.Context, writer driver.Writer, createUserInput types.CreateUser) (types.User, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	newUserID, err := writer.WriteUserNewSignUp(ctx, createUserInput, time.Now())
	if err != nil {
		return types.User{}, fmt.Errorf("error WriteUserNewSignUp: %v", err)
	}

	user, err := c.reader.ReadUserByUserID(ctx, newUserID)
	if err != nil {
		return types.User{}, fmt.Errorf("error ReadUserByUserID: %v", err)
	}

	newAccountInput := types.Account{PlanType: types.BasicPlan, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	newAccount, err := writer.WriteAccount(ctx, newUserID, newAccountInput, time.Now())
	if err != nil {
		// rollback the user if the account failed
		_, Derr := writer.DeletePortalUser(ctx, user.ID)
		if Derr != nil {
			return types.User{}, fmt.Errorf("error deleting new user with failed account: %v", Derr)
		}

		return types.User{}, fmt.Errorf("error WriteAccount: %v", err)
	}

	c.userIDs[createUserInput.ProviderUserID] = user.ID

	if _, ok := c.accountsByUserID[user.ID]; !ok {
		c.accountsByUserID[user.ID] = make(map[types.AccountID]*types.Account)
	}
	c.accountsByUserID[user.ID][newAccount.ID] = newAccount

	return *user, nil
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

// GetPortalApp returns PortalApp from cache by portalAppID
func (c *Cache) GetPortalAppByProtocolID(protocolAppID types.ProtocolAppID) types.PortalApp {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	portalApp, ok := c.portalAppsByProtocolID[protocolAppID]
	if !ok {
		return types.PortalApp{}
	}

	portalAppDeref := *portalApp

	// Ensure only the requested protocol app AAT is returned
	for aatProtocalAppID := range portalAppDeref.AATs {
		if aatProtocalAppID != protocolAppID {
			delete(portalAppDeref.AATs, aatProtocalAppID)
		}
	}

	return portalAppDeref
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
		portalApp.LegacyFields.StickyOptions = existingPortalApp.LegacyFields.StickyOptions
	}

	if portalApp.Deleted {
		delete(c.portalApps, portalApp.ID)
		delete(c.portalAppsByProtocolID, portalApp.ProtocolAppID())
		if account, ok := c.accounts[portalApp.AccountID]; ok {
			delete(account.PortalApps, portalApp.ID)
		}

		return
	}

	pendingAAT, ok := c.pendingAAT[portalApp.ID]
	if ok {
		if portalApp.AATs == nil {
			portalApp.AATs = make(map[types.ProtocolAppID]types.AAT)
		}
		portalApp.AATs[pendingAAT.ID] = pendingAAT
		delete(c.pendingAAT, portalApp.ID)
	}

	pendingSettings, ok := c.pendingSettings[portalApp.ID]
	if ok {
		portalApp.Settings = pendingSettings
		delete(c.pendingSettings, portalApp.ID)
	}

	pendingNotification, ok := c.pendingAppNotifications[portalApp.ID]
	if ok {
		if portalApp.Notifications == nil {
			portalApp.Notifications = make(map[types.NotificationType]types.AppNotification)
		}
		portalApp.Notifications[pendingNotification.Type] = pendingNotification
		delete(c.pendingAppNotifications, portalApp.ID)
	}

	pendingWhitelists, ok := c.pendingWhitelists[portalApp.ID]
	if ok {
		for _, whitelist := range pendingWhitelists {
			portalApp.AddWhitelist(whitelist)
		}
		delete(c.pendingWhitelists, portalApp.ID)
	}

	// TODO - remove when v2 migration finished
	pendingStickyOptions, ok := c.pendingStickyOptions[portalApp.ID]
	if ok {
		portalApp.LegacyFields.StickyOptions = pendingStickyOptions
		delete(c.pendingStickyOptions, portalApp.ID)
	}

	c.portalApps[portalApp.ID] = &portalApp

	if protocolAppID := portalApp.ProtocolAppID(); protocolAppID != "" {
		c.portalAppsByProtocolID[protocolAppID] = &portalApp
	}

	if account, ok := c.accounts[portalApp.AccountID]; ok {
		if account.PortalApps == nil {
			account.PortalApps = make(map[types.PortalAppID]*types.PortalApp)
		}
		account.PortalApps[portalApp.ID] = &portalApp
	}
}

// addPortalAppAAT adds or updates the AAT for a PortalApp in cache
func (c *Cache) addPortalAppAAT(aat types.AAT) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := aat.AppID
	aat.AppID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	c.pendingAAT[appID] = aat

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

	if c.pendingWhitelists == nil {
		c.pendingWhitelists = make(map[types.PortalAppID][]types.Whitelist)
	}
	c.pendingWhitelists[appID] = append(c.pendingWhitelists[appID], whitelist)

	portalApp, ok := c.portalApps[appID]
	if !ok {
		return
	}

	portalApp.AddWhitelist(whitelist)
}

// addStickinessOptions adds or updates the StickinesSOptions for a PortalApp in cache
// TODO - remove when v2 migration finished
func (c *Cache) addStickinessOptions(stickyOptions types.StickyOptions) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := types.PortalAppID(stickyOptions.ID)
	stickyOptions.ID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	portalApp, ok := c.portalApps[appID]
	if !ok {
		c.pendingStickyOptions[appID] = stickyOptions
		return
	}

	portalApp.LegacyFields.StickyOptions = stickyOptions
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

/* ---------- Chains ---------- */

// GetChain returns Chain from cache by chainID
func (c *Cache) GetChain(chainID types.RelayChainID, options Options) types.Chain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	chain, ok := c.chains[chainID]
	if !ok {
		return types.Chain{}
	}

	if options.SetGigastakeRedirectApp {
		c.setGigastakeRedirectApp(chain)
	}

	return *chain
}

// GetChains returns all Chains from cache
func (c *Cache) GetChains(options Options) []types.Chain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var chainsToReturn []types.Chain

	for _, chain := range c.chains {
		if options.SetGigastakeRedirectApp {
			c.setGigastakeRedirectApp(chain)
		}
		if options.IncludeDeleted || !chain.Deleted {
			chainsToReturn = append(chainsToReturn, *chain)
		}
	}

	return chainsToReturn
}

// setGigastakeRedirectApp sets the GigastakeAccount for each Redirect in a Chain if options.SetGigastakeRedirectApp is true
func (c *Cache) setGigastakeRedirectApp(chain *types.Chain) {
	for i := range chain.Redirects {
		portalAppID := chain.Redirects[i].PortalApplicationID
		if portalApp, ok := c.portalApps[portalAppID]; ok {
			chain.Redirects[i].GigastakePortalApp = portalApp
		} else {
			c.log.Warn(fmt.Sprintf("error setting gigastake portal app for chain ID %s: portal app ID %s does not exist in cache", chain.ID, portalAppID))
		}
	}
}

// upsertChain adds, updates or deletes Chain saved in cache
func (c *Cache) upsertChain(chain types.Chain) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	if existingChain, ok := c.chains[chain.ID]; ok {
		chain.Altruists = existingChain.Altruists
		chain.Redirects = existingChain.Redirects
		chain.Checks = existingChain.Checks
	}

	pendingAltruists, ok := c.pendingAltruists[chain.ID]
	if ok {
		chain.Altruists = pendingAltruists
		delete(c.pendingAltruists, chain.ID)
	}

	pendingRedirects, ok := c.pendingGigastakeRedirects[chain.ID]
	if ok {
		chain.Redirects = pendingRedirects
		delete(c.pendingGigastakeRedirects, chain.ID)
	}

	pendingChecks, ok := c.pendingChecks[chain.ID]
	if ok {
		if chain.Checks == nil {
			chain.Checks = make(map[types.ChainCheckType]types.Check)
		}
		for _, check := range pendingChecks {
			chain.Checks[check.Type] = check
		}
		delete(c.pendingChecks, chain.ID)
	}

	c.chains[chain.ID] = &chain
}

// addAltruist adds or updates Altruist saved in cache
func (c *Cache) addAltruist(altruist types.Altruist) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	chainID := altruist.ChainID
	altruist.ChainID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	altruistExists := func(altruists []types.Altruist) bool {
		for i, existingAltruist := range altruists {
			if existingAltruist.URL == altruist.URL {
				altruists[i] = altruist
				return true
			}
		}
		return false
	}

	if chain, ok := c.chains[chainID]; ok {
		if !altruistExists(chain.Altruists) {
			chain.Altruists = append(chain.Altruists, altruist)
		}
		return
	}

	if !altruistExists(c.pendingAltruists[chainID]) {
		c.pendingAltruists[chainID] = append(c.pendingAltruists[chainID], altruist)
	}
}

// addGigastakeRedirect adds or updates GigastakeRedirect saved in cache
func (c *Cache) addGigastakeRedirect(redirect types.GigastakeRedirect) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	chainID := redirect.ChainID
	redirect.ChainID = "" // to avoid multiple sources of truth TODO: update this by sending row ID in notification struct and passing as param

	redirectExists := func(redirects []types.GigastakeRedirect) bool {
		for i, existingRedirect := range redirects {
			if existingRedirect.PortalApplicationID == redirect.PortalApplicationID && existingRedirect.Domain == redirect.Domain {
				redirects[i] = redirect
				return true
			}
		}
		return false
	}

	if chain, ok := c.chains[chainID]; ok {
		if !redirectExists(chain.Redirects) {
			chain.Redirects = append(chain.Redirects, redirect)
		}
		return
	}

	if !redirectExists(c.pendingGigastakeRedirects[chainID]) {
		c.pendingGigastakeRedirects[chainID] = append(c.pendingGigastakeRedirects[chainID], redirect)
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
// TODO update more fields on the account user when users creation logic updated
func (c *Cache) upsertUser(user types.User) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

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
