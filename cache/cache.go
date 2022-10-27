package cache

import (
	"sync"
	"time"

	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	retriesSideTable = 5
)

// Reader represents implementation of reader interface
type Reader interface {
	ReadApplications() ([]*repository.Application, error)
	ReadBlockchains() ([]*repository.Blockchain, error)
	ReadLoadBalancers() ([]*repository.LoadBalancer, error)
	ReadPayPlans() ([]*repository.PayPlan, error)
	ReadRedirects() ([]*repository.Redirect, error)
	NotificationChannel() <-chan *repository.Notification
}

// Cache struct handler for cache operations
type Cache struct {
	reader                     Reader
	rwMutex                    sync.RWMutex
	applicationsMap            map[string]*repository.Application
	applicationsMapByUserID    map[string][]*repository.Application
	applications               []*repository.Application
	blockchainsMap             map[string]*repository.Blockchain
	blockchains                []*repository.Blockchain
	loadBalancersMap           map[string]*repository.LoadBalancer
	loadBalancersMapByUserID   map[string][]*repository.LoadBalancer
	loadBalancers              []*repository.LoadBalancer
	payPlansMap                map[repository.PayPlanType]*repository.PayPlan
	payPlans                   []*repository.PayPlan
	redirectsMapByBlockchainID map[string][]*repository.Redirect
	listening                  bool
	pendingGatewayAAT          map[string]repository.GatewayAAT
}

// NewCache returns cache instance from reader interface
func NewCache(reader Reader) *Cache {
	return &Cache{
		reader: reader,
	}
}

// GetApplication returns Application from cache by applicationID
func (c *Cache) GetApplication(applicationID string) *repository.Application {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.applicationsMap[applicationID]
}

// GetApplicationsByUserID returns Applications from cache by userID
func (c *Cache) GetApplicationsByUserID(userID string) []*repository.Application {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.applicationsMapByUserID[userID]
}

// GetApplications returns all Applications in cache
func (c *Cache) GetApplications() []*repository.Application {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.applications
}

// GetBlockchain returns Blockchain from cache by blockchainID
func (c *Cache) GetBlockchain(blockchainID string) *repository.Blockchain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.blockchainsMap[blockchainID]
}

// GetBlockchains returns all Blockchains from cache
func (c *Cache) GetBlockchains() []*repository.Blockchain {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.blockchains
}

// GetLoadBalancer returns Loadbalancer by loadbalancerID
func (c *Cache) GetLoadBalancer(loadBalancerID string) *repository.LoadBalancer {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.loadBalancersMap[loadBalancerID]
}

// GetLoadBalancers returns all Loadbalancers on cache
func (c *Cache) GetLoadBalancers() []*repository.LoadBalancer {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.loadBalancers
}

func (c *Cache) GetLoadBalancersByUserID(userID string) []*repository.LoadBalancer {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.loadBalancersMapByUserID[userID]
}

// GetPayPlan returns PayPlan from cache by planType
func (c *Cache) GetPayPlan(planType repository.PayPlanType) *repository.PayPlan {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.payPlansMap[planType]
}

// GetPayPlans returns all PayPlans in cache
func (c *Cache) GetPayPlans() []*repository.PayPlan {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.payPlans
}

// GetRedirects returns all Redirects from cache by blockchainID
func (c *Cache) GetRedirects(blockchainID string) []*repository.Redirect {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return c.redirectsMapByBlockchainID[blockchainID]
}

func (c *Cache) setApplications() error {
	applications, err := c.reader.ReadApplications()
	if err != nil {
		return err
	}

	applicationsMap := make(map[string]*repository.Application)
	applicationsMapByUserID := make(map[string][]*repository.Application)

	for i := 0; i < len(applications); i++ {
		plan := c.payPlansMap[applications[i].PayPlanType]

		if plan != nil {
			applications[i].Limits = repository.AppLimits{
				PlanType:   plan.PlanType,
				DailyLimit: plan.DailyLimit,
			}
		}

		applications[i].PayPlanType = "" // set to empty to avoid two sources of truth

		applicationsMap[applications[i].ID] = applications[i]
		applicationsMapByUserID[applications[i].UserID] = append(applicationsMapByUserID[applications[i].UserID], applications[i])
	}

	c.applications = applications
	c.applicationsMap = applicationsMap
	c.applicationsMapByUserID = applicationsMapByUserID

	return nil
}

// addApplication adds application to cache
func (c *Cache) addApplication(app repository.Application) {
	if app.PayPlanType != "" {
		newPlan := c.GetPayPlan(app.PayPlanType)
		app.Limits = repository.AppLimits{
			PlanType:   newPlan.PlanType,
			DailyLimit: newPlan.DailyLimit,
		}

		app.PayPlanType = "" // set to empty to avoid two sources of truth
	}

	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	aat := c.pendingGatewayAAT[app.ID]
	if aat != (repository.GatewayAAT{}) {
		app.GatewayAAT = aat
		delete(c.pendingGatewayAAT, app.ID)
	}

	c.applications = append(c.applications, &app)
	c.applicationsMap[app.ID] = &app
	c.applicationsMapByUserID[app.UserID] = append(c.applicationsMapByUserID[app.UserID], &app)
}

func (c *Cache) addGatewayAAT(aat repository.GatewayAAT) {
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

func (c *Cache) addGatewaySettings(settings repository.GatewaySettings) {
	for i := 0; i < retriesSideTable; i++ {
		app := c.GetApplication(settings.ID)
		if app != nil {
			c.rwMutex.Lock()
			defer c.rwMutex.Unlock()

			settings.ID = "" // to avoid multiple sources of truth
			app.GatewaySettings = settings
			return
		}

		time.Sleep(1 * time.Second)
	}
}

func (c *Cache) addNotificationSettings(settings repository.NotificationSettings) {
	for i := 0; i < retriesSideTable; i++ {
		app := c.GetApplication(settings.ID)
		if app != nil {
			c.rwMutex.Lock()
			defer c.rwMutex.Unlock()

			settings.ID = "" // to avoid multiple sources of truth
			app.NotificationSettings = settings
			return
		}

		time.Sleep(1 * time.Second)
	}
}

// updateApplication updates application saved in cache
func (c *Cache) updateApplication(inApp repository.Application) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	app := c.applicationsMap[inApp.ID]

	if inApp.PayPlanType != "" {
		newPlan := c.payPlansMap[inApp.PayPlanType]
		app.Limits = repository.AppLimits{
			PlanType:   newPlan.PlanType,
			DailyLimit: newPlan.DailyLimit,
		}
	}

	app.Name = inApp.Name
	app.Status = inApp.Status
	app.FirstDateSurpassed = inApp.FirstDateSurpassed
	app.UpdatedAt = inApp.UpdatedAt
}

func (c *Cache) setBlockchains() error {
	blockchains, err := c.reader.ReadBlockchains()
	if err != nil {
		return err
	}

	blockchainsMap := make(map[string]*repository.Blockchain)

	for _, blockchain := range blockchains {
		if blockchainRedirects, exists := c.redirectsMapByBlockchainID[blockchain.ID]; exists {
			redirects := []repository.Redirect{}
			for _, redirect := range blockchainRedirects {
				redirects = append(redirects, *redirect)
			}
			blockchain.Redirects = redirects
		}

		blockchainsMap[blockchain.ID] = blockchain
	}

	c.blockchains = blockchains
	c.blockchainsMap = blockchainsMap

	return nil
}

// addBlockchain adds blockchain to cache
func (c *Cache) addBlockchain(blockchain repository.Blockchain) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.blockchains = append(c.blockchains, &blockchain)
	c.blockchainsMap[blockchain.ID] = &blockchain
}

func (c *Cache) addSyncOptions(opts repository.SyncCheckOptions) {
	for i := 0; i < retriesSideTable; i++ {
		blockchain := c.GetBlockchain(opts.BlockchainID)
		if blockchain != nil {
			c.rwMutex.Lock()
			defer c.rwMutex.Unlock()

			blockchain.SyncCheckOptions = opts
			return
		}

		time.Sleep(1 * time.Second)
	}
}

// updateBlockchain updates blockchain saved in cache
func (c *Cache) updateBlockchain(inBlockchain repository.Blockchain) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	blockchain := c.blockchainsMap[inBlockchain.ID]
	blockchain.Active = inBlockchain.Active
	blockchain.UpdatedAt = inBlockchain.UpdatedAt
}

func (c *Cache) setLoadBalancers() error {
	loadBalancers, err := c.reader.ReadLoadBalancers()
	if err != nil {
		return err
	}

	loadBalancersMap := make(map[string]*repository.LoadBalancer)
	loadBalancersMapByUserID := make(map[string][]*repository.LoadBalancer)

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
func (c *Cache) addLoadBalancer(lb repository.LoadBalancer) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.loadBalancers = append(c.loadBalancers, &lb)
	c.loadBalancersMap[lb.ID] = &lb
	c.loadBalancersMapByUserID[lb.UserID] = append(c.loadBalancersMapByUserID[lb.UserID], &lb)
}

func (c *Cache) addStickinessOptions(opts repository.StickyOptions) {
	for i := 0; i < retriesSideTable; i++ {
		lb := c.GetLoadBalancer(opts.ID)
		if lb != nil {
			c.rwMutex.Lock()
			defer c.rwMutex.Unlock()

			opts.ID = "" // to avoid multiple sources of truth
			lb.StickyOptions = opts
			return
		}

		time.Sleep(1 * time.Second)
	}
}

func (c *Cache) addLbApp(lbApp repository.LbApp) {
	for i := 0; i < retriesSideTable; i++ {
		lb := c.GetLoadBalancer(lbApp.LbID)
		if lb != nil {
			c.rwMutex.Lock()
			defer c.rwMutex.Unlock()

			lb.Applications = append(lb.Applications, c.applicationsMap[lbApp.AppID])
			return
		}

		time.Sleep(1 * time.Second)
	}
}

// updateLoadBalancer updates load balancer saved in cache
func (c *Cache) updateLoadBalancer(inLb repository.LoadBalancer) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	lb := c.loadBalancersMap[inLb.ID]

	lb.Name = inLb.Name
	lb.UserID = inLb.UserID
	lb.UpdatedAt = inLb.UpdatedAt
}

func (c *Cache) setPayPlans() error {
	payPlans, err := c.reader.ReadPayPlans()
	if err != nil {
		return err
	}

	payPlansMap := make(map[repository.PayPlanType]*repository.PayPlan)

	for _, payPlan := range payPlans {
		payPlansMap[payPlan.PlanType] = payPlan
	}

	c.payPlans = payPlans
	c.payPlansMap = payPlansMap

	return nil
}

func (c *Cache) setRedirects() error {
	redirects, err := c.reader.ReadRedirects()
	if err != nil {
		return err
	}

	redirectsMap := make(map[string][]*repository.Redirect)

	for _, redirect := range redirects {
		redirectsMap[redirect.BlockchainID] = append(redirectsMap[redirect.BlockchainID], redirect)
	}

	c.redirectsMapByBlockchainID = redirectsMap

	return nil
}

// AddRedirects adds blockchain redirect to cache and updates cached blockchain entry
func (c *Cache) addRedirect(redirect repository.Redirect) {
	c.rwMutex.Lock()
	c.redirectsMapByBlockchainID[redirect.BlockchainID] = append(c.redirectsMapByBlockchainID[redirect.BlockchainID], &redirect)
	c.rwMutex.Unlock()

	blockchain := c.GetBlockchain(redirect.BlockchainID)
	blockchain.Redirects = append(blockchain.Redirects, redirect)
}

// SetCache gets all values from DB and stores them in cache
func (c *Cache) SetCache() error {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	err := c.setPayPlans()
	if err != nil {
		return err
	}

	err = c.setRedirects()
	if err != nil {
		return err
	}

	// always call after setPayPlans func
	err = c.setApplications()
	if err != nil {
		return err
	}

	// always call after setRedirects func
	err = c.setBlockchains()
	if err != nil {
		return err
	}

	// always call after setApplications func
	err = c.setLoadBalancers()
	if err != nil {
		return err
	}

	if !c.listening {
		go c.listen()
	}

	return nil
}
