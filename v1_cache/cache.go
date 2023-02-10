package v1_cache

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/pokt-foundation/portal-db/driver"
	"github.com/pokt-foundation/portal-db/types"
	stringUtils "github.com/pokt-foundation/utils-go/strings"
	"github.com/sirupsen/logrus"
)

// Cache struct handler for cache operations
type Cache struct {
	reader                   driver.Reader
	rwMutex                  sync.RWMutex
	applicationsMap          map[string]*types.Application
	applicationsMapByUserID  map[string][]*types.Application
	applications             []*types.Application
	blockchainsMap           map[string]*types.Blockchain
	blockchains              []*types.Blockchain
	loadBalancersMap         map[string]*types.LoadBalancer
	loadBalancersMapByUserID map[string][]*types.LoadBalancer
	userPermissions          map[types.UserID]*types.UserPermissions
	loadBalancers            []*types.LoadBalancer
	payPlansMap              map[types.PayPlanType]*types.PayPlan
	payPlans                 []*types.PayPlan

	listening bool

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

// NewCache returns cache instance from reader interface
func NewCache(reader driver.Reader, logger *logrus.Logger) *Cache {
	return &Cache{
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
}

// GetUserPermissionsByUserID returns all load balancer UserPermissions for a given UserID
func (c *Cache) GetUserPermissionsByUserID(userID types.UserID) *types.UserPermissions {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.userPermissions[userID]
}

// GetApplication returns Application from cache by applicationID
func (c *Cache) GetApplication(applicationID string) *types.Application {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.applicationsMap[applicationID]
}

// GetApplicationsByUserID returns Applications from cache by userID
func (c *Cache) GetApplicationsByUserID(userID string) []*types.Application {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.applicationsMapByUserID[userID]
}

// GetApplications returns all Applications in cache
func (c *Cache) GetApplications() []*types.Application {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.applications
}

// GetBlockchain returns Blockchain from cache by blockchainID
func (c *Cache) GetBlockchain(blockchainID string) *types.Blockchain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.blockchainsMap[blockchainID]
}

// GetBlockchains returns all Blockchains from cache
func (c *Cache) GetBlockchains() []*types.Blockchain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.blockchains
}

// GetLoadBalancer returns Loadbalancer by loadbalancerID
func (c *Cache) GetLoadBalancer(loadBalancerID string) *types.LoadBalancer {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.loadBalancersMap[loadBalancerID]
}

// GetLoadBalancers returns all Loadbalancers on cache
func (c *Cache) GetLoadBalancers() []*types.LoadBalancer {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.loadBalancers
}

func (c *Cache) GetLoadBalancersByUserID(userID string) []*types.LoadBalancer {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.loadBalancersMapByUserID[userID]
}

// GetPayPlan returns PayPlan from cache by planType
func (c *Cache) GetPayPlan(planType types.PayPlanType) *types.PayPlan {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.payPlansMap[planType]
}

// GetPayPlans returns all PayPlans in cache
func (c *Cache) GetPayPlans() []*types.PayPlan {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.payPlans
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

	if inApp.UserID == "" {
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

	appsForUser := c.applicationsMapByUserID[userID]
	appsForUserAfterRemove := []*types.Application{}

	for i := range appsForUser {
		if appsForUser[i].ID != inApp.ID {
			appsForUserAfterRemove = append(appsForUserAfterRemove, appsForUser[i])
		}
	}

	c.applicationsMapByUserID[userID] = appsForUserAfterRemove
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

func (c *Cache) addSyncOptions(opts types.SyncCheckOptions) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	blockchain := c.blockchainsMap[opts.BlockchainID]
	if blockchain != nil {
		blockchain.SyncCheckOptions = opts
		return
	}

	c.pendingSyncCheckOptions[opts.BlockchainID] = opts
}

func (c *Cache) addRedirect(redirect types.Redirect) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	blockchain := c.blockchainsMap[redirect.BlockchainID]
	if blockchain != nil {
		blockchain.Redirects = append(blockchain.Redirects, redirect)
		return
	}

	c.pendingRedirects[redirect.BlockchainID] = redirect
}

// updateBlockchain updates blockchain saved in cache
func (c *Cache) updateBlockchain(inBlockchain types.Blockchain) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	blockchain := c.blockchainsMap[inBlockchain.ID]
	blockchain.Active = inBlockchain.Active
	blockchain.UpdatedAt = inBlockchain.UpdatedAt
}

func (c *Cache) setLoadBalancers() error {
	loadBalancers, err := c.reader.ReadLoadBalancers(context.Background())
	if err != nil {
		return fmt.Errorf("err in ReadLoadBalancers: %w", err)
	}

	loadBalancersMap := make(map[string]*types.LoadBalancer)
	loadBalancersMapByUserID := make(map[string][]*types.LoadBalancer)

	for i, loadBalancer := range loadBalancers {
		for _, appID := range loadBalancer.ApplicationIDs {
			loadBalancer.Applications = append(loadBalancer.Applications, c.applicationsMap[appID])
		}

		loadBalancer.ApplicationIDs = nil // set to nil to avoid having two proofs of truth

		loadBalancers[i] = loadBalancer
		loadBalancersMap[loadBalancer.ID] = loadBalancer

		for _, user := range loadBalancer.Users { // User ID map stores all load balancers that user is a team member on
			loadBalancersMapByUserID[user.UserID] = append(loadBalancersMapByUserID[user.UserID], loadBalancer)
		}
	}

	c.loadBalancers = loadBalancers
	c.loadBalancersMap = loadBalancersMap
	c.loadBalancersMapByUserID = loadBalancersMapByUserID

	return nil
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

	// Add user to permissions map if they have a user ID
	if userAccess.UserID != "" {
		userID := types.UserID(userAccess.UserID)

		if userPermissions, ok := c.userPermissions[userID]; ok {
			_, err := userPermissions.UpsertPermissions(types.LoadBalancerID(lbID), userAccess.RoleName)
			fmt.Println("error upserting user permissions: %w", err)
		} else {
			permissions := &types.UserPermissions{
				UserID:        userID,
				LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{},
			}
			_, err := permissions.UpsertPermissions(types.LoadBalancerID(lbID), userAccess.RoleName)
			fmt.Println("error upserting user permissions: %w", err)

			c.userPermissions[userID] = permissions
		}
	}

	lb := c.loadBalancersMap[lbID]
	if lb != nil {
		lb.Users = append(lb.Users, userAccess)

		if userAccess.UserID != "" {
			c.loadBalancersMapByUserID[userAccess.UserID] = append(c.loadBalancersMapByUserID[userAccess.UserID], lb)
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

// updateLoadBalancer updates load balancer saved in cache
func (c *Cache) updateLoadBalancer(inLB types.LoadBalancer) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	lb := c.loadBalancersMap[inLB.ID]

	if inLB.UserID == "" {
		c.removeLoadBalancerFromUserIDMap(inLB, lb)
	}

	lb.Name = inLB.Name
	lb.UserID = inLB.UserID
	lb.UpdatedAt = inLB.UpdatedAt
}

// removeApplication removes load balancers saved in cache from the loadBalancersMapByUserID as they no longer have a userID
func (c *Cache) removeLoadBalancerFromUserIDMap(inLB types.LoadBalancer, oldLB *types.LoadBalancer) {
	userID := oldLB.UserID

	lbsForUser := c.loadBalancersMapByUserID[userID]
	lbsForUserAfterRemove := []*types.LoadBalancer{}

	for i := range lbsForUser {
		if lbsForUser[i].ID != inLB.ID {
			lbsForUserAfterRemove = append(lbsForUserAfterRemove, lbsForUser[i])
		}
	}

	c.loadBalancersMapByUserID[userID] = lbsForUserAfterRemove
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

// updateUserAccess updates load balancer user saved in cache with a new role and updates permissions map
func (c *Cache) updateUserAccess(inUser types.UserAccess) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	lbID := inUser.ID
	inUser.ID = "" // to avoid multiple sources of truth

	// Add user to permissions map if they have a user ID (have been approved)
	if inUser.UserID != "" {
		userID := types.UserID(inUser.UserID)

		if userPermissions, ok := c.userPermissions[userID]; ok {
			_, err := userPermissions.UpsertPermissions(types.LoadBalancerID(lbID), inUser.RoleName)
			fmt.Println("error upserting user permissions: %w", err)
		} else {
			permissions := &types.UserPermissions{
				UserID:        userID,
				LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{},
			}
			_, err := permissions.UpsertPermissions(types.LoadBalancerID(lbID), inUser.RoleName)
			fmt.Println("error upserting user permissions: %w", err)

			c.userPermissions[userID] = permissions
		}
	}

	// Update load balancer users slice
	lb := c.loadBalancersMap[lbID]
	if lb != nil {
		for i := range lb.Users {
			if lb.Users[i].UserID == inUser.UserID {
				lb.Users[i].RoleName = inUser.RoleName
			}

			// Update new user when they accept their invite
			if lb.Users[i].Email == inUser.Email && !lb.Users[i].Accepted && inUser.Accepted {
				lb.Users[i].Accepted = inUser.Accepted
				lb.Users[i].UserID = inUser.UserID

				c.loadBalancersMapByUserID[inUser.UserID] = append(c.loadBalancersMapByUserID[inUser.UserID], lb)
			}
		}
	}
}

// deleteUserAccess deletes load balancer user saved in cache  and updates permissions
func (c *Cache) deleteUserAccess(delUser types.UserAccess) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	lbID := delUser.ID
	userID := types.UserID(delUser.UserID)
	delUser.ID = "" // to avoid multiple sources of truth

	if userPermissions, ok := c.userPermissions[userID]; ok {
		_ = userPermissions.DeletePermissions(types.LoadBalancerID(lbID))
	}

	// Remove from load balancer users slice
	lb := c.loadBalancersMap[lbID]

	filteredUsers := []types.UserAccess{}
	for _, user := range lb.Users {
		if user.UserID != delUser.UserID {
			filteredUsers = append(filteredUsers, user)
		}
	}
	lb.Users = filteredUsers

	lbsForUser := c.loadBalancersMapByUserID[delUser.UserID]

	lbsForUserAfterRemove := []*types.LoadBalancer{}
	for i := range lbsForUser {
		if lbsForUser[i].ID != lb.ID {
			lbsForUserAfterRemove = append(lbsForUserAfterRemove, lbsForUser[i])
		}
	}

	c.loadBalancersMapByUserID[delUser.UserID] = lbsForUserAfterRemove
}

func (c *Cache) setUserPermissions() error {
	userPermissions, err := c.reader.ReadUserPermissions(context.Background())
	if err != nil {
		return fmt.Errorf("err in ReadLoadBalancers: %w", err)
	}

	c.userPermissions = userPermissions

	return nil
}

// SetCache gets all values from DB and stores them in cache
func (c *Cache) SetCache() error {
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

	if !c.listening {
		go c.listen()
	}

	return nil
}
