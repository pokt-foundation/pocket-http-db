package cache

import (
	"errors"
	"testing"

	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/stretchr/testify/require"
)

func TestCache_SetCache(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadApplications").Return([]*repository.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
			Limit: repository.AppLimit{
				PayPlan: repository.PayPlan{
					Type:  repository.FreetierV0,
					Limit: 250000,
				},
			},
		},
		{
			ID:     "5f62b7d8be3591c4dea8566a",
			UserID: "60ecb2bf67774900350d9c43",
			Limit: repository.AppLimit{
				PayPlan: repository.PayPlan{
					Type: repository.Enterprise,
				},
				CustomLimit: 2000000,
			},
		},
		{
			ID:     "5f62b7d8be3591c4dea8566f",
			UserID: "60ecb2bf67774900350d9c44",
			Limit: repository.AppLimit{
				PayPlan: repository.PayPlan{
					Type:  repository.PayAsYouGoV0,
					Limit: 0,
				},
			},
		},
	}, nil)

	readerMock.On("ReadBlockchains").Return([]*repository.Blockchain{
		{ID: "0021"},
	}, nil)

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{
		{
			ID:     "60ecb2bf67774900350d9c42",
			UserID: "60ecb35fts687463gh2h72gs",
			ApplicationIDs: []string{
				"5f62b7d8be3591c4dea8566d",
				"5f62b7d8be3591c4dea8566a",
			},
		},
	}, nil)

	readerMock.On("ReadPayPlans").Return([]*repository.PayPlan{
		{
			Type:  repository.FreetierV0,
			Limit: 250000,
		},
		{
			Type:  repository.PayAsYouGoV0,
			Limit: 0,
		},
	}, nil)
	readerMock.On("ReadRedirects").Return([]*repository.Redirect{
		{
			BlockchainID:   "0021",
			Alias:          "pokt-mainnet",
			Domain:         "pokt-mainnet.gateway.network",
			LoadBalancerID: "12345",
		},
		{
			BlockchainID:   "0022",
			Alias:          "eth-mainnet",
			Domain:         "eth-mainnet.gateway.network",
			LoadBalancerID: "45678",
		},
	}, nil)

	cache := NewCache(readerMock)

	err := cache.SetCache()
	c.NoError(err)

	c.NotEmpty(cache.GetApplication("5f62b7d8be3591c4dea8566d"))
	c.Len(cache.GetApplications(), 3)
	c.Len(cache.GetApplicationsByUserID("60ecb2bf67774900350d9c43"), 2)

	c.NotEmpty(cache.GetBlockchain("0021"))
	c.Len(cache.GetBlockchains(), 1)

	c.NotEmpty(cache.GetLoadBalancer("60ecb2bf67774900350d9c42"))
	c.Len(cache.GetLoadBalancers(), 1)
	c.Len(cache.GetLoadBalancersByUserID("60ecb35fts687463gh2h72gs"), 1)

	c.NotEmpty(cache.GetPayPlan(repository.FreetierV0))
	c.Len(cache.GetPayPlans(), 2)

	c.NotEmpty(cache.GetRedirects("0021"))
	c.Len(cache.GetRedirects("0021"), 1)
}

func TestCache_SetCacheFailure(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}
	cache := NewCache(readerMock)

	readerMock.On("ReadPayPlans").Return([]*repository.PayPlan{}, errors.New("error on pay plans")).Once()

	err := cache.SetCache()
	c.EqualError(err, "error on pay plans")

	readerMock.On("ReadPayPlans").Return([]*repository.PayPlan{
		{
			Type:  repository.FreetierV0,
			Limit: 250000,
		},
		{
			Type:  repository.PayAsYouGoV0,
			Limit: 0,
		},
	}, nil)

	readerMock.On("ReadRedirects").Return([]*repository.Redirect{}, errors.New("error on redirects")).Once()

	err = cache.SetCache()
	c.EqualError(err, "error on redirects")

	readerMock.On("ReadRedirects").Return([]*repository.Redirect{
		{
			BlockchainID:   "0021",
			Alias:          "pokt-mainnet",
			Domain:         "pokt-mainnet.gateway.network",
			LoadBalancerID: "12345",
		},
		{
			BlockchainID:   "0022",
			Alias:          "eth-mainnet",
			Domain:         "eth-mainnet.gateway.network",
			LoadBalancerID: "45678",
		},
	}, nil)

	readerMock.On("ReadBlockchains").Return([]*repository.Blockchain{}, errors.New("error on blockchains")).Once()

	err = cache.SetCache()
	c.EqualError(err, "error on blockchains")

	readerMock.On("ReadBlockchains").Return([]*repository.Blockchain{
		{
			ID: "0021",
		},
	}, nil)

	readerMock.On("ReadApplications").Return([]*repository.Application{}, errors.New("error on applications")).Once()

	err = cache.SetCache()
	c.EqualError(err, "error on applications")

	readerMock.On("ReadApplications").Return([]*repository.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
		},
	}, nil)

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{}, errors.New("error on loadbalancers")).Once()

	err = cache.SetCache()
	c.EqualError(err, "error on loadbalancers")

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{
		{
			ID: "60ecb2bf67774900350d9c42",
		},
	}, nil)
}

func TestCache_AddApplication(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadPayPlans").Return([]*repository.PayPlan{
		{
			Type:  repository.FreetierV0,
			Limit: 250000,
		},
		{
			Type:  repository.PayAsYouGoV0,
			Limit: 0,
		},
	}, nil)

	readerMock.On("ReadApplications").Return([]*repository.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
		},
		{
			ID:     "5f62b7d8be3591c4dea8566a",
			UserID: "60ecb2bf67774900350d9c43",
		},
		{
			ID:     "5f62b7d8be3591c4dea8566f",
			UserID: "60ecb2bf67774900350d9c44",
		},
	}, nil)

	cache := NewCache(readerMock)

	err := cache.setPayPlans()
	c.NoError(err)

	err = cache.setApplications()
	c.NoError(err)

	cache.addApplication(repository.Application{
		ID:     "5f62b7d8be3591c4dea8566b",
		UserID: "60ecb2bf67774900350d9c43",
		Limit: repository.AppLimit{
			PayPlan: repository.PayPlan{
				Type:  repository.FreetierV0,
				Limit: 250000,
			},
		},
	})

	c.Len(cache.GetApplications(), 4)
	c.Len(cache.GetApplicationsByUserID("60ecb2bf67774900350d9c43"), 3)
	c.Equal(cache.GetApplication("5f62b7d8be3591c4dea8566b").DailyLimit(), 250000)
}

func TestCache_UpdateApplication(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadPayPlans").Return([]*repository.PayPlan{
		{
			Type:  repository.FreetierV0,
			Limit: 250000,
		},
		{
			Type:  repository.PayAsYouGoV0,
			Limit: 0,
		},
	}, nil)

	readerMock.On("ReadApplications").Return([]*repository.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
		},
		{
			ID:     "5f62b7d8be3591c4dea8566a",
			UserID: "60ecb2bf67774900350d9c43",
		},
		{
			ID:     "5f62b7d8be3591c4dea8566f",
			UserID: "60ecb2bf67774900350d9c44",
		},
	}, nil)

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{
		{
			ID:     "60ecb2bf67774900350d9c42",
			UserID: "60ecb35fts687463gh2h72gs",
			ApplicationIDs: []string{
				"5f62b7d8be3591c4dea8566d",
				"5f62b7d8be3591c4dea8566a",
			},
		},
	}, nil)

	cache := NewCache(readerMock)

	err := cache.setPayPlans()
	c.NoError(err)

	err = cache.setApplications()
	c.NoError(err)

	err = cache.setLoadBalancers()
	c.NoError(err)

	cache.updateApplication(repository.Application{
		ID:     "5f62b7d8be3591c4dea8566a",
		UserID: "60ecb2bf67774900350d9c43",
		Name:   "papolo",
	})

	c.Len(cache.GetApplications(), 3)
	c.Len(cache.GetApplicationsByUserID("60ecb2bf67774900350d9c43"), 2)
	c.Len(cache.GetApplicationsByUserID("60ecb2bf67774900350d9c44"), 1)
	c.Equal("papolo", cache.GetApplication("5f62b7d8be3591c4dea8566a").Name)
	c.Equal("papolo", cache.GetLoadBalancer("60ecb2bf67774900350d9c42").Applications[1].Name)
	c.Equal("papolo", cache.GetApplicationsByUserID("60ecb2bf67774900350d9c43")[1].Name)
	c.Equal("papolo", cache.GetLoadBalancer("60ecb2bf67774900350d9c42").Applications[1].Name)
	c.Equal("papolo", cache.GetLoadBalancersByUserID("60ecb35fts687463gh2h72gs")[0].Applications[1].Name)
}

func TestCache_AddLoadBalancer(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
		},
		{
			ID:     "5f62b7d8be3591c4dea8566a",
			UserID: "60ecb2bf67774900350d9c43",
		},
		{
			ID:     "5f62b7d8be3591c4dea8566f",
			UserID: "60ecb2bf67774900350d9c44",
		},
	}, nil)

	readerMock.On("ReadApplications").Return([]*repository.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
		},
	}, nil)

	cache := NewCache(readerMock)

	err := cache.setLoadBalancers()
	c.NoError(err)

	err = cache.setApplications()
	c.NoError(err)

	cache.addLoadBalancer(repository.LoadBalancer{
		ID:     "5f62b7d8be3591c4dea8566b",
		UserID: "60ecb2bf67774900350d9c43",
	})

	c.Len(cache.GetLoadBalancers(), 4)
	c.Len(cache.GetLoadBalancersByUserID("60ecb2bf67774900350d9c43"), 3)
}

func TestCache_UpdateLoadBalancer(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
		},
		{
			ID:     "5f62b7d8be3591c4dea8566a",
			UserID: "60ecb2bf67774900350d9c43",
		},
		{
			ID:     "5f62b7d8be3591c4dea8566f",
			UserID: "60ecb2bf67774900350d9c44",
		},
	}, nil)

	cache := NewCache(readerMock)

	err := cache.setLoadBalancers()
	c.NoError(err)

	cache.updateLoadBalancer(repository.LoadBalancer{
		ID:     "5f62b7d8be3591c4dea8566a",
		UserID: "60ecb2bf67774900350d9c43",
		Name:   "papolo",
	})

	c.Len(cache.GetLoadBalancers(), 3)
	c.Len(cache.GetLoadBalancersByUserID("60ecb2bf67774900350d9c43"), 2)
	c.Len(cache.GetLoadBalancersByUserID("60ecb2bf67774900350d9c44"), 1)
	c.Equal("papolo", cache.GetLoadBalancersByUserID("60ecb2bf67774900350d9c43")[1].Name)
	c.Equal("papolo", cache.GetLoadBalancer("5f62b7d8be3591c4dea8566a").Name)
}

func TestCache_AddBlockchain(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadBlockchains").Return([]*repository.Blockchain{
		{ID: "0001", Ticker: "POKT"},
	}, nil)

	cache := NewCache(readerMock)

	err := cache.setBlockchains()
	c.NoError(err)

	c.Len(cache.GetBlockchains(), 1)

	cache.addBlockchain(repository.Blockchain{ID: "0002", Ticker: "ETH"})

	c.Len(cache.GetBlockchains(), 2)
}

func TestCache_UpdateBlockchain(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadBlockchains").Return([]*repository.Blockchain{
		{ID: "0001", Active: false},
	}, nil)

	cache := NewCache(readerMock)

	err := cache.setBlockchains()
	c.NoError(err)

	c.Len(cache.GetBlockchains(), 1)
	c.Equal(cache.GetBlockchains()[0].Active, false)

	cache.updateBlockchain(repository.Blockchain{
		ID:     "0001",
		Active: true,
	})

	c.Len(cache.GetBlockchains(), 1)
	c.Equal(cache.GetBlockchains()[0].Active, true)
}

func TestCache_AddRedirect(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadBlockchains").Return([]*repository.Blockchain{
		{ID: "0001", Ticker: "POKT"},
	}, nil)

	readerMock.On("ReadRedirects").Return([]*repository.Redirect{
		{BlockchainID: "0001", Alias: "pokt-mainnet-1"},
		{BlockchainID: "0001", Alias: "pokt-mainnet-2"},
		{BlockchainID: "0002", Alias: "eth-mainnet"},
	}, nil)

	cache := NewCache(readerMock)

	err := cache.setRedirects()
	c.NoError(err)

	err = cache.setBlockchains()
	c.NoError(err)

	c.Len(cache.GetBlockchains(), 1)
	c.Len(cache.GetBlockchains()[0].Redirects, 2)
	c.Len(cache.GetRedirects("0001"), 2)

	cache.addRedirect(repository.Redirect{BlockchainID: "0001", Alias: "pokt-mainnet-3"})

	c.Len(cache.GetBlockchains()[0].Redirects, 3)
	c.Len(cache.GetRedirects("0001"), 3)
}
