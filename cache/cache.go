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
	ReadUsers() ([]*repository.User, error)
}

// Cache struct handler for cache operations
type Cache struct {
	reader                   Reader
	applicationsMap          map[string]*repository.Application
	applicationsMapByUserID  map[string][]*repository.Application
	applications             []*repository.Application
	applicationsMux          sync.Mutex
	blockchainsMap           map[string]*repository.Blockchain
	blockchains              []*repository.Blockchain
	blockchainsMux           sync.Mutex
	loadBalancersMap         map[string]*repository.LoadBalancer
	loadBalancersMapByUserID map[string][]*repository.LoadBalancer
	loadBalancers            []*repository.LoadBalancer
	loadBalancersMux         sync.Mutex
	usersMap                 map[string]*repository.User
	users                    []*repository.User
	usersMux                 sync.Mutex
}

// NewCache returns cache instance from reader interface
func NewCache(reader Reader) *Cache {
	return &Cache{
		reader: reader,
	}
}

// GetApplication returns Application from cache by applicationID
func (c *Cache) GetApplication(applicationID string) *repository.Application {
	return c.applicationsMap[applicationID]
}

// GetApplicationsByUserID returns Applications from cache by userID
func (c *Cache) GetApplicationsByUserID(userID string) []*repository.Application {
	return c.applicationsMapByUserID[userID]
}

// GetApplications returns all Applications in cache
func (c *Cache) GetApplications() []*repository.Application {
	return c.applications
}

// GetBlockchain returns Blockchain from cache by blockchainID
func (c *Cache) GetBlockchain(blockchainID string) *repository.Blockchain {
	return c.blockchainsMap[blockchainID]
}

// GetBlockchains returns all Blockchains from cache
func (c *Cache) GetBlockchains() []*repository.Blockchain {
	return c.blockchains
}

// GetLoadBalancer returns Loadbalancer by loadbalancerID
func (c *Cache) GetLoadBalancer(loadBalancerID string) *repository.LoadBalancer {
	return c.loadBalancersMap[loadBalancerID]
}

// GetLoadBalancers returns all Loadbalancers on cache
func (c *Cache) GetLoadBalancers() []*repository.LoadBalancer {
	return c.loadBalancers
}

func (c *Cache) GetLoadBalancersByUserID(userID string) []*repository.LoadBalancer {
	return c.loadBalancersMapByUserID[userID]
}

// GetUser returns User from cache by userID
func (c *Cache) GetUser(userID string) *repository.User {
	return c.usersMap[userID]
}

// GetUsers returns all Users in cache
func (c *Cache) GetUsers() []*repository.User {
	return c.users
}

func (c *Cache) setApplications() error {
	applications, err := c.reader.ReadApplications()
	if err != nil {
		return err
	}

	applicationsMap := make(map[string]*repository.Application)
	applicationsMapByUserID := make(map[string][]*repository.Application)

	for _, application := range applications {
		applicationsMap[application.ID] = application
		applicationsMapByUserID[application.UserID] = append(applicationsMapByUserID[application.UserID], application)
	}

	c.applicationsMux.Lock()
	defer c.applicationsMux.Unlock()

	c.applications = applications
	c.applicationsMap = applicationsMap
	c.applicationsMapByUserID = applicationsMapByUserID

	return nil
}

// AddApplication adds application to cache
func (c *Cache) AddApplication(app *repository.Application) {
	c.applications = append(c.applications, app)
	c.applicationsMap[app.ID] = app
	c.applicationsMapByUserID[app.UserID] = append(c.applicationsMapByUserID[app.UserID], app)
}

// UpdateApplication updates application saved in cache
func (c *Cache) UpdateApplication(app *repository.Application, oldUserID string) {
	c.applications = updateApplicationFromSlice(app, c.applications)

	c.applicationsMap[app.ID] = app

	c.applicationsMapByUserID[oldUserID] = deleteApplicationFromSlice(app.ID, c.applicationsMapByUserID[oldUserID])
	c.applicationsMapByUserID[app.UserID] = append(c.applicationsMapByUserID[app.UserID], app)

	for i := 0; i < len(c.loadBalancers); i++ {
		updateApplicationFromSlice(app, c.loadBalancers[i].Applications)
	}
}

func deleteApplicationFromSlice(appID string, apps []*repository.Application) []*repository.Application {
	for i := 0; i < len(apps); i++ {
		if apps[i].ID == appID {
			apps[i] = apps[len(apps)-1]
			apps = apps[:len(apps)-1]

			break
		}
	}

	return apps
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
		blockchainsMap[blockchain.ID] = blockchain
	}

	c.blockchainsMux.Lock()
	defer c.blockchainsMux.Unlock()

	c.blockchains = blockchains
	c.blockchainsMap = blockchainsMap

	return nil
}

func (c *Cache) setLoadBalancers() error {
	loadBalancers, err := c.reader.ReadLoadBalancers()
	if err != nil {
		return err
	}

	loadBalancersMap := make(map[string]*repository.LoadBalancer)
	loadBalancersMapByUserID := make(map[string][]*repository.LoadBalancer)

	c.applicationsMux.Lock()

	for i, loadBalancer := range loadBalancers {
		for _, appID := range loadBalancer.ApplicationIDs {
			loadBalancer.Applications = append(loadBalancer.Applications, c.applicationsMap[appID])
		}

		loadBalancer.ApplicationIDs = nil // set to nil to avoid having two proofs of truth

		loadBalancers[i] = loadBalancer
		loadBalancersMap[loadBalancer.ID] = loadBalancer
		loadBalancersMapByUserID[loadBalancer.UserID] = append(loadBalancersMapByUserID[loadBalancer.UserID], loadBalancer)
	}

	c.applicationsMux.Unlock()

	c.loadBalancersMux.Lock()
	defer c.loadBalancersMux.Unlock()

	c.loadBalancers = loadBalancers
	c.loadBalancersMap = loadBalancersMap
	c.loadBalancersMapByUserID = loadBalancersMapByUserID

	return nil
}

// AddLoadBalancer adds load balancer to cache
func (c *Cache) AddLoadBalancer(lb *repository.LoadBalancer) {
	c.loadBalancers = append(c.loadBalancers, lb)
	c.loadBalancersMap[lb.ID] = lb
	c.loadBalancersMapByUserID[lb.UserID] = append(c.loadBalancersMapByUserID[lb.UserID], lb)
}

// UpdateLoadBalancer updates load balancer saved in cache
func (c *Cache) UpdateLoadBalancer(lb *repository.LoadBalancer, oldUserID string) {
	c.loadBalancers = updateLoadBalancerFromSlice(lb, c.loadBalancers)

	c.loadBalancersMap[lb.ID] = lb

	c.loadBalancersMapByUserID[oldUserID] = deleteLoadBalancerFromSlice(lb.ID, c.loadBalancersMapByUserID[oldUserID])
	c.loadBalancersMapByUserID[lb.UserID] = append(c.loadBalancersMapByUserID[lb.UserID], lb)
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

func (c *Cache) setUsers() error {
	users, err := c.reader.ReadUsers()
	if err != nil {
		return err
	}

	userMap := make(map[string]*repository.User)

	for _, user := range users {
		userMap[user.ID] = user
	}

	c.usersMux.Lock()
	defer c.usersMux.Unlock()

	c.users = users
	c.usersMap = userMap

	return nil
}

// SetCache gets all values from DB and stores them in cache
func (c *Cache) SetCache() error {
	err := c.setApplications()
	if err != nil {
		return err
	}

	err = c.setBlockchains()
	if err != nil {
		return err
	}

	// always call after setApplications func
	err = c.setLoadBalancers()
	if err != nil {
		return err
	}

	return c.setUsers()
}
