package cache

import (
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
	reader                  Reader
	applicationsMap         map[string]*repository.Application
	applicationsMapByUserID map[string][]*repository.Application
	applications            []*repository.Application
	blockchainsMap          map[string]*repository.Blockchain
	blockchains             []*repository.Blockchain
	loadBalancersMap        map[string]*repository.LoadBalancer
	loadBalancers           []*repository.LoadBalancer
	usersMap                map[string]*repository.User
	users                   []*repository.User
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

	c.applications = applications
	c.applicationsMap = applicationsMap
	c.applicationsMapByUserID = applicationsMapByUserID

	return nil
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

	for i, loadBalancer := range loadBalancers {
		for _, appID := range loadBalancer.ApplicationIDs {
			loadBalancer.Applications = append(loadBalancer.Applications, c.applicationsMap[appID])
		}

		loadBalancers[i] = loadBalancer
		loadBalancersMap[loadBalancer.ID] = loadBalancer
	}

	c.loadBalancers = loadBalancers
	c.loadBalancersMap = loadBalancersMap

	return nil
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
