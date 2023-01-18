package cache

import (
	"context"
	"fmt"
	"sync"

	"github.com/pokt-foundation/portal-db/driver"
	"github.com/pokt-foundation/portal-db/types"
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
	pendingLbApps              map[string][]types.LbApp
	log                        *logrus.Logger
}

// NewCache returns cache instance from reader interface
func NewCache(reader driver.Reader, logger *logrus.Logger) *Cache {
	return &Cache{
		reader:                     reader,
		pendingAppLimit:            make(map[string]types.AppLimit),
		pendingGatewayAAT:          make(map[string]types.GatewayAAT),
		pendingGatewaySettings:     make(map[string]types.GatewaySettings),
		pendingNotifactionSettings: make(map[string]types.NotificationSettings),
		pendingSyncCheckOptions:    make(map[string]types.SyncCheckOptions),
		pendingStickyOptions:       make(map[string]types.StickyOptions),
		pendingLbApps:              make(map[string][]types.LbApp),
		log:                        logger,
	}
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
		app.GatewaySettings = gSettings
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

func (c *Cache) addGatewaySettings(settings types.GatewaySettings) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	appID := settings.ID
	settings.ID = "" // to avoid multiple sources of truth

	app := c.applicationsMap[appID]
	if app != nil {
		app.GatewaySettings = settings
		return
	}

	c.pendingGatewaySettings[appID] = settings
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
		loadBalancersMapByUserID[loadBalancer.UserID] = append(loadBalancersMapByUserID[loadBalancer.UserID], loadBalancer)
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

	lbApps, ok := c.pendingLbApps[lb.ID]
	if ok {
		for _, lbApp := range lbApps {
			lb.Applications = append(lb.Applications, c.applicationsMap[lbApp.AppID])
		}
		delete(c.pendingLbApps, lb.ID)
	}

	c.loadBalancers = append(c.loadBalancers, &lb)
	c.loadBalancersMap[lb.ID] = &lb
	c.loadBalancersMapByUserID[lb.UserID] = append(c.loadBalancersMapByUserID[lb.UserID], &lb)
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

	if !c.listening {
		go c.listen()
	}

	return nil
}
