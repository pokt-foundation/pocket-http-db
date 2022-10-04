package cache

import (
	"sync"

	"github.com/pokt-foundation/portal-api-go/repository"
)

// Reader represents implementation of reader interface
type Reader interface {
	ReadApplications() ([]*repository.Application, error)
	ReadBlockchains() ([]*repository.Blockchain, error)
	ReadLoadBalancers() ([]*repository.LoadBalancer, error)
	ReadPayPlans() ([]*repository.PayPlan, error)
	ReadRedirects() ([]*repository.Redirect, error)
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

// AddApplication adds application to cache
func (c *Cache) AddApplication(app *repository.Application) {
	if app.PayPlanType != "" {
		newPlan := c.GetPayPlan(app.PayPlanType)
		app.Limits = repository.AppLimits{
			PlanType:   newPlan.PlanType,
			DailyLimit: newPlan.DailyLimit,
		}
	}

	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.applications = append(c.applications, app)
	c.applicationsMap[app.ID] = app
	c.applicationsMapByUserID[app.UserID] = append(c.applicationsMapByUserID[app.UserID], app)
}

// UpdateApplication updates application saved in cache
func (c *Cache) UpdateApplication(app *repository.Application) {
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

	c.applications = updateApplicationFromSlice(app, c.applications)
	c.applicationsMap[app.ID] = app
	c.applicationsMapByUserID[app.UserID] = updateApplicationFromSlice(app, c.applicationsMapByUserID[app.UserID])

	for _, lb := range c.loadBalancers {
		lb.Applications = updateApplicationFromSlice(app, lb.Applications)

		c.loadBalancers = updateLoadBalancerFromSlice(lb, c.loadBalancers)

		c.loadBalancersMap[lb.ID] = lb

		c.loadBalancersMapByUserID[lb.UserID] = updateLoadBalancerFromSlice(lb, c.loadBalancersMapByUserID[lb.UserID])
	}
}

func updateApplicationFromSlice(updatedApp *repository.Application, apps []*repository.Application) []*repository.Application {
	for i := 0; i < len(apps); i++ {
		if apps[i].ID == updatedApp.ID {
			apps[i] = apps[len(apps)-1]
			apps = apps[:len(apps)-1]
			apps = append(apps, updatedApp)

			break
		}
	}

	return apps
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

// AddBlockchain adds blockchain to cache
func (c *Cache) AddBlockchain(blockchain *repository.Blockchain) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.blockchains = append(c.blockchains, blockchain)
	c.blockchainsMap[blockchain.ID] = blockchain
}

// ActivateBlockchain updates application saved in cache
func (c *Cache) ActivateBlockchain(id string, active bool) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	if blockchain, exists := c.blockchainsMap[id]; exists {
		blockchain.Active = active
		c.blockchainsMap[id] = blockchain
	}
}

func updateBlockchainFromSlice(updatedChain *repository.Blockchain, chains []*repository.Blockchain) []*repository.Blockchain {
	for i := 0; i < len(chains); i++ {
		if chains[i].ID == updatedChain.ID {
			chains[i] = chains[len(chains)-1]
			chains = chains[:len(chains)-1]
			chains = append(chains, updatedChain)

			break
		}
	}

	return chains
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

// AddLoadBalancer adds load balancer to cache
func (c *Cache) AddLoadBalancer(lb *repository.LoadBalancer) {
	for _, appID := range lb.ApplicationIDs {
		lb.Applications = append(lb.Applications, c.GetApplication(appID))
	}

	lb.ApplicationIDs = nil // set to nil to avoid having two proofs of truth

	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.loadBalancers = append(c.loadBalancers, lb)
	c.loadBalancersMap[lb.ID] = lb
	c.loadBalancersMapByUserID[lb.UserID] = append(c.loadBalancersMapByUserID[lb.UserID], lb)
}

// UpdateLoadBalancer updates load balancer saved in cache
func (c *Cache) UpdateLoadBalancer(lb *repository.LoadBalancer) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.loadBalancers = updateLoadBalancerFromSlice(lb, c.loadBalancers)

	c.loadBalancersMap[lb.ID] = lb

	c.loadBalancersMapByUserID[lb.UserID] = updateLoadBalancerFromSlice(lb, c.loadBalancersMapByUserID[lb.UserID])
}

// DeleteLoadBalancer removes the load balancer from the cache
func (c *Cache) DeleteLoadBalancer(lb *repository.LoadBalancer, oldUserID string) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.loadBalancersMapByUserID[oldUserID] = deleteLoadBalancerFromSlice(lb.ID, c.loadBalancersMapByUserID[oldUserID])
}

func deleteLoadBalancerFromSlice(lbID string, lbs []*repository.LoadBalancer) []*repository.LoadBalancer {
	for i := 0; i < len(lbs); i++ {
		if lbs[i].ID == lbID {
			lbs[i] = lbs[len(lbs)-1]
			lbs = lbs[:len(lbs)-1]

			break
		}
	}

	return lbs
}

func updateLoadBalancerFromSlice(updatedlb *repository.LoadBalancer, lbs []*repository.LoadBalancer) []*repository.LoadBalancer {
	for i := 0; i < len(lbs); i++ {
		if lbs[i].ID == updatedlb.ID {
			lbs[i] = lbs[len(lbs)-1]
			lbs = lbs[:len(lbs)-1]
			lbs = append(lbs, updatedlb)

			break
		}
	}

	return lbs
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
func (c *Cache) AddRedirect(redirect *repository.Redirect) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.redirectsMapByBlockchainID[redirect.BlockchainID] = append(c.redirectsMapByBlockchainID[redirect.BlockchainID], redirect)

	if blockchain, exists := c.blockchainsMap[redirect.BlockchainID]; exists {
		blockchain.Redirects = append(blockchain.Redirects, *redirect)

		c.blockchains = updateBlockchainFromSlice(blockchain, c.blockchains)
		c.blockchainsMap[redirect.BlockchainID] = blockchain
	}
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
	return c.setLoadBalancers()
}
