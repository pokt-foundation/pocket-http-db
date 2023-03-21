package cache

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/pokt-foundation/portal-db/driver"
	"github.com/pokt-foundation/portal-db/types"
	stringUtils "github.com/pokt-foundation/utils-go/strings"
	"github.com/sirupsen/logrus"
)

// Cache struct handler for cache operations
type Cache struct {
	reader                         driver.Reader
	rwMutex                        sync.RWMutex
	applicationsMap                map[string]*types.Application
	applicationsMapByUserID        map[string][]*types.Application
	applications                   []*types.Application
	blockchainsMap                 map[string]*types.Blockchain
	blockchains                    []*types.Blockchain
	loadBalancersMap               map[string]*types.LoadBalancer
	loadBalancersMapByUserID       map[string][]*types.LoadBalancer
	pendingLoadBalancersMapByEmail map[string][]*types.LoadBalancer
	loadBalancersOwnedByEmail      map[string][]*types.LoadBalancer
	userPermissions                map[types.UserID]*types.UserPermissions
	loadBalancers                  []*types.LoadBalancer
	payPlansMap                    map[types.PayPlanType]*types.PayPlan
	payPlans                       []*types.PayPlan

	pendingAppLimit            map[string]types.AppLimit
	pendingGatewayAAT          map[string]types.GatewayAAT
	pendingGatewaySettings     map[string]types.GatewaySettings
	pendingNotifactionSettings map[string]types.NotificationSettings
	pendingRedirects           map[string]types.Redirect
	pendingSyncCheckOptions    map[string]types.SyncCheckOptions
	pendingStickyOptions       map[string]types.StickyOptions
	pendingUserAccess          map[string][]types.UserAccess
	pendingLbApps              map[string][]types.LbApp

	log *logrus.Logger
}

// NewCache returns cache instance from reader interface and starts the cache refresh handler and listener in separate goroutines
func NewCache(reader driver.Reader, cacheRefreshMinutes time.Duration, logger *logrus.Logger) (*Cache, error) {
	cache := &Cache{
		reader: reader,

		pendingAppLimit:            make(map[string]types.AppLimit),
		pendingGatewayAAT:          make(map[string]types.GatewayAAT),
		pendingGatewaySettings:     make(map[string]types.GatewaySettings),
		pendingNotifactionSettings: make(map[string]types.NotificationSettings),
		pendingRedirects:           make(map[string]types.Redirect),
		pendingSyncCheckOptions:    make(map[string]types.SyncCheckOptions),
		pendingStickyOptions:       make(map[string]types.StickyOptions),
		pendingUserAccess:          make(map[string][]types.UserAccess),
		pendingLbApps:              make(map[string][]types.LbApp),

		log: logger,
	}

	err := cache.setCache()
	if err != nil {
		cache.log.Error(fmt.Errorf("err in NewCache: %w", err))
		return nil, err
	}
	cache.log.WithFields(logrus.Fields{"cache_refresh_minutes": cacheRefreshMinutes.Minutes()}).Info("cache successfully set")

	go cache.cacheRefreshHandler(cacheRefreshMinutes)

	go cache.listen()

	return cache, nil
}

// cacheRefreshHandler sets the cache from the DB at a regular interval (default is 5 minutes)
func (c *Cache) cacheRefreshHandler(cacheRefreshMinutes time.Duration) {
	for {
		<-time.After(cacheRefreshMinutes)

		err := c.setCache()
		if err != nil {
			c.log.Error("err in cacheRefreshHandler: %w", err)
		}

		c.log.WithFields(logrus.Fields{"cache_refresh_minutes": cacheRefreshMinutes.Minutes()}).Info("cache successfully set")
	}
}

// setCache gets all values from DB and stores them in cache
func (c *Cache) setCache() error {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	err := c.setPayPlans()
	if err != nil {
		return fmt.Errorf("err in setPayPlans: %w", err)
	}

	err = c.setBlockchains()
	if err != nil {
		return fmt.Errorf("err in setBlockchains: %w", err)
	}

	// Always call after setPayPlans
	err = c.setApplications()
	if err != nil {
		return fmt.Errorf("err in setApplications: %w", err)
	}

	// Always call after setApplications
	err = c.setLoadBalancers()
	if err != nil {
		return err
	}

	err = c.setUserPermissions()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) setPayPlans() error {
	payPlans, err := c.reader.ReadPayPlans(context.Background())
	if err != nil {
		return err
	}

	payPlansMap := make(map[types.PayPlanType]*types.PayPlan)

	for _, payPlan := range payPlans {
		payPlansMap[payPlan.Type] = payPlan
	}

	c.payPlans = payPlans
	c.payPlansMap = payPlansMap

	return nil
}

func (c *Cache) setBlockchains() error {
	blockchains, err := c.reader.ReadBlockchains(context.Background())
	if err != nil {
		return err
	}

	blockchainsMap := make(map[string]*types.Blockchain)

	for _, blockchain := range blockchains {
		blockchainsMap[blockchain.ID] = blockchain
	}

	c.blockchains = blockchains
	c.blockchainsMap = blockchainsMap

	return nil
}

func (c *Cache) setApplications() error {
	applications, err := c.reader.ReadApplications(context.Background())
	if err != nil {
		return err
	}

	applicationsMap := make(map[string]*types.Application)
	applicationsMapByUserID := make(map[string][]*types.Application)

	for i := 0; i < len(applications); i++ {
		app := applications[i]
		applicationID, userID := app.ID, app.UserID

		applicationsMap[applicationID] = app
		applicationsMapByUserID[userID] = append(applicationsMapByUserID[userID], app)
	}

	c.applications = applications
	c.applicationsMap = applicationsMap
	c.applicationsMapByUserID = applicationsMapByUserID

	return nil
}

func (c *Cache) setLoadBalancers() error {
	loadBalancers, err := c.reader.ReadLoadBalancers(context.Background())
	if err != nil {
		return fmt.Errorf("err in ReadLoadBalancers: %w", err)
	}

	loadBalancersMap := make(map[string]*types.LoadBalancer)
	loadBalancersMapByUserID := make(map[string][]*types.LoadBalancer)
	pendingLoadBalancersMapByEmail := make(map[string][]*types.LoadBalancer)
	loadBalancersMapByEmail := make(map[string][]*types.LoadBalancer)

	for i, loadBalancer := range loadBalancers {
		for _, appID := range loadBalancer.ApplicationIDs {
			loadBalancer.Applications = append(loadBalancer.Applications, c.applicationsMap[appID])
		}

		loadBalancer.ApplicationIDs = nil // set to nil to avoid having two proofs of truth

		loadBalancers[i] = loadBalancer
		loadBalancersMap[loadBalancer.ID] = loadBalancer

		for _, user := range loadBalancer.Users {
			// user ID map stores all load balancers that user is a team member on
			// if the user has not accepted their invite they do not have a user ID and won't be added to this map
			loadBalancersMapByUserID[user.UserID] = append(loadBalancersMapByUserID[user.UserID], loadBalancer)
			// if not yet accepted, add to pending LBs map
			if !user.Accepted {
				pendingLoadBalancersMapByEmail[user.Email] = append(pendingLoadBalancersMapByEmail[user.Email], loadBalancer)
			}

			if user.RoleName == types.RoleOwner {
				loadBalancersMapByEmail[user.Email] = append(loadBalancersMapByEmail[user.Email], loadBalancer)
			}
		}
	}

	c.loadBalancers = loadBalancers
	c.loadBalancersMap = loadBalancersMap
	c.loadBalancersMapByUserID = loadBalancersMapByUserID
	c.pendingLoadBalancersMapByEmail = pendingLoadBalancersMapByEmail
	c.loadBalancersOwnedByEmail = loadBalancersMapByEmail

	return nil
}

func (c *Cache) setUserPermissions() error {
	userPermissions, err := c.reader.ReadUserPermissions(context.Background())
	if err != nil {
		return fmt.Errorf("err in ReadLoadBalancers: %w", err)
	}

	c.userPermissions = userPermissions

	return nil
}

/* ----- Read Methods ----- */

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

// GetApplication returns Application from cache by applicationID
func (c *Cache) GetApplication(applicationID string) types.Application {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	app, ok := c.applicationsMap[applicationID]
	if !ok {
		return types.Application{}
	}

	return *app
}

// GetApplicationsByUserID returns Applications from cache by userID
func (c *Cache) GetApplicationsByUserID(userID string) []types.Application {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var appsToReturn []types.Application

	for _, app := range c.applicationsMapByUserID[userID] {
		appsToReturn = append(appsToReturn, *app)
	}

	return appsToReturn
}

// GetApplications returns all Applications in cache
func (c *Cache) GetApplications() []types.Application {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var appsToReturn []types.Application

	for _, app := range c.applications {
		appsToReturn = append(appsToReturn, *app)
	}

	return appsToReturn
}

// GetBlockchain returns Blockchain from cache by blockchainID
func (c *Cache) GetBlockchain(blockchainID string) types.Blockchain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	blockchain, ok := c.blockchainsMap[blockchainID]
	if !ok {
		return types.Blockchain{}
	}

	return *blockchain
}

// GetBlockchains returns all Blockchains from cache
func (c *Cache) GetBlockchains() []types.Blockchain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var blockchainsToReturn []types.Blockchain

	for _, blockchain := range c.blockchains {
		blockchainsToReturn = append(blockchainsToReturn, *blockchain)
	}

	return blockchainsToReturn
}

// GetLoadBalancer returns Loadbalancer by loadbalancerID
func (c *Cache) GetLoadBalancer(loadBalancerID string) types.LoadBalancer {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	lb, ok := c.loadBalancersMap[loadBalancerID]
	if !ok {
		return types.LoadBalancer{}
	}

	return *lb
}

// GetLoadBalancers returns all Loadbalancers on cache
func (c *Cache) GetLoadBalancers() []types.LoadBalancer {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var lbsToReturn []types.LoadBalancer

	for _, lb := range c.loadBalancers {
		lbsToReturn = append(lbsToReturn, *lb)
	}

	return lbsToReturn
}

func (c *Cache) GetLoadBalancersByUserID(userID string) []types.LoadBalancer {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var lbsToReturn []types.LoadBalancer

	for _, lb := range c.loadBalancersMapByUserID[userID] {
		lbsToReturn = append(lbsToReturn, *lb)
	}

	return lbsToReturn
}

func (c *Cache) GetPendingLoadBalancersByEmail(email string) []types.LoadBalancer {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var lbsToReturn []types.LoadBalancer

	for _, lb := range c.pendingLoadBalancersMapByEmail[email] {
		lbsToReturn = append(lbsToReturn, *lb)
	}

	return lbsToReturn
}

func (c *Cache) GetLoadBalancersCountByEmail(email string) int {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return len(c.loadBalancersOwnedByEmail[email])
}

// GetPayPlan returns PayPlan from cache by planType
func (c *Cache) GetPayPlan(planType types.PayPlanType) types.PayPlan {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	payPlan, ok := c.payPlansMap[planType]
	if !ok {
		return types.PayPlan{}
	}

	return *payPlan
}

// GetPayPlans returns all PayPlans in cache
func (c *Cache) GetPayPlans() []types.PayPlan {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	var payPlansToReturn []types.PayPlan

	for _, payPlan := range c.payPlans {
		payPlansToReturn = append(payPlansToReturn, *payPlan)
	}

	return payPlansToReturn
}

/* ----- Write Methods ----- */

// addApplication adds application to cache
func (c *Cache) addApplication(app types.Application) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	limit, ok := c.pendingAppLimit[app.ID]
	if ok {
		app.Limit = limit
		delete(c.pendingAppLimit, app.ID)
	}

	aat, ok := c.pendingGatewayAAT[app.ID]
	if ok {
		app.GatewayAAT = aat
		delete(c.pendingGatewayAAT, app.ID)
	}

	gSettings, ok := c.pendingGatewaySettings[app.ID]
	if ok {
		// These are the only gateway settings fields set on a newly created app
		app.GatewaySettings.SecretKey = gSettings.SecretKey
		app.GatewaySettings.SecretKeyRequired = gSettings.SecretKeyRequired
		delete(c.pendingGatewaySettings, app.ID)
	}

	nSettings, ok := c.pendingNotifactionSettings[app.ID]
	if ok {
		app.NotificationSettings = nSettings
		delete(c.pendingNotifactionSettings, app.ID)
	}

	c.applications = append(c.applications, &app)
	c.applicationsMap[app.ID] = &app
	c.applicationsMapByUserID[app.UserID] = append(c.applicationsMapByUserID[app.UserID], &app)
}

func (c *Cache) addAppLimit(limit types.AppLimit) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := limit.ID
	limit.ID = "" // to avoid multiple sources of truth

	if limit.PayPlan.Type != types.Enterprise {
		payPlan, ok := c.payPlansMap[limit.PayPlan.Type]
		if !ok {
			fmt.Println("invalid pay plan type on add app limit")
			return
		}
		limit.PayPlan.Limit = payPlan.Limit
	}

	if app, ok := c.applicationsMap[appID]; ok {
		app.Limit = limit
		return
	}

	c.pendingAppLimit[appID] = limit
}

func (c *Cache) addGatewayAAT(aat types.GatewayAAT) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := aat.ID
	aat.ID = "" // to avoid multiple sources of truth

	app := c.applicationsMap[appID]
	if app != nil {
		app.GatewayAAT = aat
		return
	}

	c.pendingGatewayAAT[appID] = aat
}

// Updates the gateway settings from a listener notification
func (c *Cache) addGatewaySettings(settings types.GatewaySettings) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := settings.ID
	settings.ID = "" // to avoid multiple sources of truth

	app, ok := c.applicationsMap[appID]
	if !ok { // Add to pending map in case gateway settings notification arrives first
		c.pendingGatewaySettings[appID] = settings
		return
	}

	// The below fields are sent from the listener on the gateway setting struct
	app.GatewaySettings.SecretKey = settings.SecretKey
	app.GatewaySettings.SecretKeyRequired = settings.SecretKeyRequired
	app.GatewaySettings.WhitelistOrigins = settings.WhitelistOrigins
	app.GatewaySettings.WhitelistUserAgents = settings.WhitelistUserAgents
	app.GatewaySettings.WhitelistBlockchains = settings.WhitelistBlockchains
}

// Updates the gateway settings whitelist contracts from a listener notification
func (c *Cache) addWhitelistContract(inContract types.WhitelistContract) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := inContract.ID
	inContract.ID = "" // to avoid multiple sources of truth

	app, ok := c.applicationsMap[appID]
	if !ok { // No pending map since whitelist methods are never set on a new app
		return
	}

	appSettings := app.GatewaySettings

	var chainIDExists bool
	for i := range appSettings.WhitelistContracts {
		if appSettings.WhitelistContracts[i].BlockchainID == inContract.BlockchainID {
			chainIDExists = true

			if stringUtils.ExactContains(appSettings.WhitelistContracts[i].Contracts, inContract.Contract) {
				break
			}

			appSettings.WhitelistContracts[i].Contracts = append(
				appSettings.WhitelistContracts[i].Contracts, inContract.Contract,
			)
			sort.Strings(appSettings.WhitelistContracts[i].Contracts)

			break
		}
	}

	if !chainIDExists {
		appSettings.WhitelistContracts = append(appSettings.WhitelistContracts, types.WhitelistContracts{
			BlockchainID: inContract.BlockchainID,
			Contracts:    []string{inContract.Contract},
		})
	}

	app.GatewaySettings = appSettings
}

// Removes a contract from the gateway settings whitelist contracts from a listener notification
func (c *Cache) deleteWhitelistContract(inContract types.WhitelistContract) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := inContract.ID
	inContract.ID = "" // to avoid multiple sources of truth

	app, ok := c.applicationsMap[appID]
	if !ok { // No pending map since whitelist methods are never set on a new app
		return
	}

	appSettings := app.GatewaySettings

	for i, whitelistContract := range appSettings.WhitelistContracts {
		if whitelistContract.BlockchainID != inContract.BlockchainID {
			continue
		}

		filteredContracts := []string{}
		for _, contract := range whitelistContract.Contracts {
			if contract != inContract.Contract {
				filteredContracts = append(filteredContracts, contract)
			}
		}

		// If no contracts remain remove the WhitelistContracts from the slice
		if len(filteredContracts) == 0 {
			appSettings.WhitelistContracts = append(appSettings.WhitelistContracts[:i], appSettings.WhitelistContracts[i+1:]...)
			break
		}

		appSettings.WhitelistContracts[i].Contracts = filteredContracts
		break
	}

	app.GatewaySettings = appSettings
}

// Updates the gateway settings whitelist methods from a listener notification
func (c *Cache) addWhitelistMethod(inMethod types.WhitelistMethod) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := inMethod.ID
	inMethod.ID = "" // to avoid multiple sources of truth

	app, ok := c.applicationsMap[appID]
	if !ok { // No pending map since whitelist methods are never set on a new app
		return
	}

	appSettings := app.GatewaySettings

	var chainIDExists bool
	for i := range appSettings.WhitelistMethods {
		if appSettings.WhitelistMethods[i].BlockchainID == inMethod.BlockchainID {
			chainIDExists = true

			if stringUtils.ExactContains(appSettings.WhitelistMethods[i].Methods, inMethod.Method) {
				break
			}

			appSettings.WhitelistMethods[i].Methods = append(
				appSettings.WhitelistMethods[i].Methods, inMethod.Method,
			)
			sort.Strings(appSettings.WhitelistMethods[i].Methods)

			break
		}
	}

	if !chainIDExists {
		appSettings.WhitelistMethods = append(appSettings.WhitelistMethods, types.WhitelistMethods{
			BlockchainID: inMethod.BlockchainID,
			Methods:      []string{inMethod.Method},
		})
	}

	app.GatewaySettings = appSettings
}

// Removes a method from the gateway settings whitelist methods from a listener notification
func (c *Cache) deleteWhitelistMethod(inMethod types.WhitelistMethod) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := inMethod.ID
	inMethod.ID = "" // to avoid multiple sources of truth

	app, ok := c.applicationsMap[appID]
	if !ok { // No pending map since whitelist methods are never set on a new app
		return
	}

	appSettings := app.GatewaySettings

	for i, whitelistMethod := range appSettings.WhitelistMethods {
		if whitelistMethod.BlockchainID != inMethod.BlockchainID {
			continue
		}

		filteredMethods := []string{}
		for _, method := range whitelistMethod.Methods {
			if method != inMethod.Method {
				filteredMethods = append(filteredMethods, method)
			}
		}

		// If no methods remain remove the WhitelistMethods from the slice
		if len(filteredMethods) == 0 {
			appSettings.WhitelistMethods = append(appSettings.WhitelistMethods[:i], appSettings.WhitelistMethods[i+1:]...)
			break
		}

		appSettings.WhitelistMethods[i].Methods = filteredMethods
		break
	}

	app.GatewaySettings = appSettings
}

func (c *Cache) addNotificationSettings(settings types.NotificationSettings) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := settings.ID
	settings.ID = "" // to avoid multiple sources of truth

	app := c.applicationsMap[appID]
	if app != nil {
		app.NotificationSettings = settings
		return
	}

	c.pendingNotifactionSettings[appID] = settings
}

// updateApplication updates application saved in cache
func (c *Cache) updateApplication(inApp types.Application) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	app := c.applicationsMap[inApp.ID]

	if inApp.Status == types.AwaitingGracePeriod {
		c.removeApplicationFromUserIDMap(inApp, app)
	}

	limit := app.Limit

	if limit.PayPlan.Type != types.Enterprise {
		payPlan, ok := c.payPlansMap[limit.PayPlan.Type]
		if ok {
			limit.PayPlan.Limit = payPlan.Limit
		}
	}

	app.Name = inApp.Name
	app.Status = inApp.Status
	app.Limit = limit
	app.FirstDateSurpassed = inApp.FirstDateSurpassed
	app.UpdatedAt = inApp.UpdatedAt
}

// removeApplicationFromUserIDMap removes applications saved in cache from the applicationsMapByUserID as they no longer have a userID
func (c *Cache) removeApplicationFromUserIDMap(inApp types.Application, oldApp *types.Application) {
	userID := oldApp.UserID

	var appsForUserAfterRemove []*types.Application
	for _, app := range c.applicationsMapByUserID[userID] {
		if app.ID != inApp.ID {
			appsForUserAfterRemove = append(appsForUserAfterRemove, app)
		}
	}

	c.applicationsMapByUserID[userID] = appsForUserAfterRemove
}

// addBlockchain adds blockchain to cache
func (c *Cache) addBlockchain(blockchain types.Blockchain) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	opts, ok := c.pendingSyncCheckOptions[blockchain.ID]
	if ok {
		blockchain.SyncCheckOptions = opts
		delete(c.pendingSyncCheckOptions, blockchain.ID)
	}

	c.blockchains = append(c.blockchains, &blockchain)
	c.blockchainsMap[blockchain.ID] = &blockchain
}

func (c *Cache) addSyncOptions(inOpts types.SyncCheckOptions) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	if blockchain, ok := c.blockchainsMap[inOpts.BlockchainID]; ok {
		blockchain.SyncCheckOptions = inOpts
		return
	}

	c.pendingSyncCheckOptions[inOpts.BlockchainID] = inOpts
}

func (c *Cache) addRedirect(redirect types.Redirect) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	if blockchain, ok := c.blockchainsMap[redirect.BlockchainID]; ok {
		blockchain.Redirects = append(blockchain.Redirects, redirect)
		return
	}

	c.pendingRedirects[redirect.BlockchainID] = redirect
}

// updateBlockchain updates blockchain saved in cache
func (c *Cache) updateBlockchain(inBlockchain types.Blockchain) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	if blockchain, ok := c.blockchainsMap[inBlockchain.ID]; ok {
		// keep the update from overwriting
		inBlockchain.SyncCheckOptions = blockchain.SyncCheckOptions
		inBlockchain.Redirects = blockchain.Redirects

		// update cache with all incoming fields
		*blockchain = inBlockchain
	}
}

// addLoadBalancer adds load balancer to cache
func (c *Cache) addLoadBalancer(lb types.LoadBalancer) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	opts, ok := c.pendingStickyOptions[lb.ID]
	if ok {
		lb.StickyOptions = opts
		delete(c.pendingStickyOptions, lb.ID)
	}

	users, ok := c.pendingUserAccess[lb.ID]
	if ok {
		lb.Users = append(lb.Users, users...)
		delete(c.pendingUserAccess, lb.ID)
	}

	lbApps, ok := c.pendingLbApps[lb.ID]
	if ok {
		for _, lbApp := range lbApps {
			lb.Applications = append(lb.Applications, c.applicationsMap[lbApp.AppID])
		}
		delete(c.pendingLbApps, lb.ID)
	}

	c.loadBalancers = append(c.loadBalancers, &lb)
	c.loadBalancersMap[lb.ID] = &lb

	for _, user := range lb.Users { // User ID map stores all load balancers that user is a team member on
		c.loadBalancersMapByUserID[user.UserID] = append(c.loadBalancersMapByUserID[user.UserID], &lb)
		if user.RoleName == types.RoleOwner {
			c.loadBalancersOwnedByEmail[user.Email] = append(c.loadBalancersOwnedByEmail[user.Email], &lb)
		}
	}
}

func (c *Cache) addStickinessOptions(opts types.StickyOptions) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	lbID := opts.ID
	opts.ID = "" // to avoid multiple sources of truth

	lb := c.loadBalancersMap[lbID]
	if lb != nil {
		lb.StickyOptions = opts
		return
	}

	c.pendingStickyOptions[lbID] = opts
}

func (c *Cache) addUserAccess(userAccess types.UserAccess) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	lbID := userAccess.ID
	userAccess.ID = "" // to avoid multiple sources of truth

	// Add user to permissions map if they have a user ID (have accepted their invitation)
	if userAccess.UserID != "" {
		userID := types.UserID(userAccess.UserID)

		userPermissions, ok := c.userPermissions[userID]
		if !ok {
			userPermissions = &types.UserPermissions{
				UserID:        userID,
				LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{},
			}
			c.userPermissions[userID] = userPermissions
		}

		_, err := userPermissions.UpsertPermissions(types.LoadBalancerID(lbID), userAccess.RoleName)
		if err != nil {
			fmt.Printf("error upserting user permissions: %v\n", err)
		}
	}

	// Add or update user's LBs
	lb := c.loadBalancersMap[lbID]
	if lb != nil {
		emailExists := false
		for _, user := range lb.Users {
			if user.Email == userAccess.Email {
				emailExists = true
				break
			}
		}
		if !emailExists {
			lb.Users = append(lb.Users, userAccess)
		}

		switch {
		// If user has just been created add to pending LBs map by email
		case !userAccess.Accepted || userAccess.UserID == "":
			c.pendingLoadBalancersMapByEmail[userAccess.Email] = append(c.pendingLoadBalancersMapByEmail[userAccess.Email], lb)

		// If user has an ID (has accepted invitation) add to LBs map by user ID and remove from pending map
		case userAccess.UserID != "" || userAccess.Accepted:
			c.loadBalancersMapByUserID[userAccess.UserID] = append(c.loadBalancersMapByUserID[userAccess.UserID], lb)

			filteredPendingLBs := []*types.LoadBalancer{}
			for _, pendingLB := range c.pendingLoadBalancersMapByEmail[userAccess.Email] {
				if pendingLB.ID != lbID {
					filteredPendingLBs = append(filteredPendingLBs, pendingLB)
				}
			}
			c.pendingLoadBalancersMapByEmail[userAccess.Email] = filteredPendingLBs
		}

		return
	}

	c.pendingUserAccess[lbID] = append(c.pendingUserAccess[lbID], userAccess)
}

func (c *Cache) addLbApp(lbApp types.LbApp) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	lb := c.loadBalancersMap[lbApp.LbID]
	if lb != nil {
		lb.Applications = append(lb.Applications, c.applicationsMap[lbApp.AppID])
		return
	}

	c.pendingLbApps[lbApp.LbID] = append(c.pendingLbApps[lbApp.LbID], lbApp)
}

func (c *Cache) RemovePendingInvitationFromUser(email, loadBalancerID string) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.pendingLoadBalancersMapByEmail[email] = removeLoadBalancerByID(c.pendingLoadBalancersMapByEmail[email], loadBalancerID)
}

// updateLoadBalancer updates load balancer saved in cache
func (c *Cache) updateLoadBalancer(inLB types.LoadBalancer) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	lb := c.loadBalancersMap[inLB.ID]

	if inLB.UserID == "" {
		c.removeLoadBalancersFromUsersMap(inLB, lb)
	}

	lb.Name = inLB.Name
	lb.UserID = inLB.UserID
	lb.UpdatedAt = inLB.UpdatedAt
}

// removeLoadBalancersFromUsersMap removes load balancers saved in cache from the loadBalancersMapByUserID as they no longer have a userID
func (c *Cache) removeLoadBalancersFromUsersMap(inLB types.LoadBalancer, oldLB *types.LoadBalancer) {
	userID := oldLB.UserID

	for _, ownerLB := range c.loadBalancersMapByUserID[userID] {
		for _, user := range ownerLB.Users {
			c.loadBalancersMapByUserID[user.UserID] = removeLoadBalancerByID(c.loadBalancersMapByUserID[user.UserID], inLB.ID)
			c.pendingLoadBalancersMapByEmail[user.Email] = removeLoadBalancerByID(c.pendingLoadBalancersMapByEmail[user.Email], inLB.ID)
			c.loadBalancersOwnedByEmail[user.Email] = removeLoadBalancerByID(c.loadBalancersOwnedByEmail[user.Email], inLB.ID)
		}
	}
}

func removeLoadBalancerByID(lbs []*types.LoadBalancer, lbID string) []*types.LoadBalancer {
	for i, lb := range lbs {
		if lb.ID == lbID {
			lbs[i] = lbs[len(lbs)-1]
			lbs = lbs[:len(lbs)-1]
			break
		}
	}

	return lbs
}

// updateUserAccess updates updates permissions map, load balancer user and user LBs map saved in cache
func (c *Cache) updateUserAccess(inUser types.UserAccess) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	lbID := inUser.ID
	inUser.ID = "" // to avoid multiple sources of truth

	// add user to permissions map if they have accepted their invitation
	if inUser.Accepted {
		userID := types.UserID(inUser.UserID)

		userPermissions, ok := c.userPermissions[userID]
		if !ok {
			userPermissions = &types.UserPermissions{
				UserID:        userID,
				LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{},
			}
			c.userPermissions[userID] = userPermissions
		}

		_, err := userPermissions.UpsertPermissions(types.LoadBalancerID(lbID), inUser.RoleName)
		if err != nil {
			fmt.Printf("error upserting user permissions: %v\n", err)
		}
	}

	if lb, ok := c.loadBalancersMap[lbID]; ok {
		for i := range lb.Users {
			newOwner := false
			if lb.Users[i].Email == inUser.Email {
				// remove the LB from the user if he's no longer the owner
				if lb.Users[i].RoleName == types.RoleOwner && inUser.RoleName != types.RoleOwner {
					c.loadBalancersOwnedByEmail[inUser.Email] = removeLoadBalancerByID(c.loadBalancersOwnedByEmail[inUser.Email], lb.ID)
				}

				if lb.Users[i].RoleName != types.RoleOwner && inUser.RoleName == types.RoleOwner {
					newOwner = true
				}

				// update user fields when user_access listener event for create/update received
				lb.Users[i].RoleName = inUser.RoleName
				lb.Users[i].UserID = inUser.UserID
				lb.Users[i].Accepted = inUser.Accepted

				// add LB to user's LBs map if it doesn't already exist there
				var exists bool
				for _, userLB := range c.loadBalancersMapByUserID[inUser.UserID] {
					if userLB.ID == lb.ID {
						exists = true
						break
					}
				}
				if !exists {
					c.loadBalancersMapByUserID[inUser.UserID] = append(c.loadBalancersMapByUserID[inUser.UserID], lb)
				}
				if newOwner {
					c.loadBalancersOwnedByEmail[inUser.Email] = append(c.loadBalancersOwnedByEmail[inUser.Email], lb)
				}
			}

		}
	}
}

// deleteUserAccess deletes load balancer user saved in cache and updates permissions
func (c *Cache) deleteUserAccess(delUser types.UserAccess) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	lbID := delUser.ID
	userID := types.UserID(delUser.UserID)
	delUser.ID = "" // to avoid multiple sources of truth

	if userPermissions, ok := c.userPermissions[userID]; ok {
		delete(userPermissions.LoadBalancers, types.LoadBalancerID(lbID))
	}

	// Remove from load balancer users slice
	lb, ok := c.loadBalancersMap[lbID]
	if ok {
		filteredUsers := []types.UserAccess{}
		for _, user := range lb.Users {
			if user.Email != delUser.Email {
				filteredUsers = append(filteredUsers, user)
			}
		}
		lb.Users = filteredUsers

		// Just a safety check since the owner can't leave the loadbalancer
		if delUser.RoleName == types.RoleOwner {
			c.loadBalancersOwnedByEmail[delUser.Email] = removeLoadBalancerByID(c.loadBalancersOwnedByEmail[delUser.Email], lb.ID)
		}

		switch delUser.UserID {
		case "":
			c.pendingLoadBalancersMapByEmail[delUser.Email] = removeLoadBalancerByID(c.pendingLoadBalancersMapByEmail[delUser.Email], lb.ID)
		default:
			c.loadBalancersMapByUserID[delUser.UserID] = removeLoadBalancerByID(c.loadBalancersMapByUserID[delUser.UserID], lb.ID)
		}
	}
}
