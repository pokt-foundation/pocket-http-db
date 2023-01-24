package v0_cache

import (
	"errors"
	"testing"

	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func Test_V0_Cache_SetCache(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadApplications").Return([]*repository.Application{
		{
			ID:     "test_7d8be3591c4dea8566d",
			UserID: "test_2bf67774900350d9c43",
			Limit: repository.AppLimit{
				PayPlan: repository.PayPlan{
					Type:  repository.FreetierV0,
					Limit: 250000,
				},
			},
		},
		{
			ID:     "test_d8be3591c4dea8566a",
			UserID: "test_2bf67774900350d9c43",
			Limit: repository.AppLimit{
				PayPlan: repository.PayPlan{
					Type: repository.Enterprise,
				},
				CustomLimit: 2000000,
			},
		},
		{
			ID:     "test_7d8be3591c4dea8566f",
			UserID: "test_2bf67774900350d9c44",
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
			ID:     "test_bf67774900350d9c42",
			UserID: "test_35fts687463gh2h72gs",
			ApplicationIDs: []string{
				"test_7d8be3591c4dea8566d",
				"test_d8be3591c4dea8566a",
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

	cache := NewCache(readerMock, logrus.New())

	err := cache.SetCache()
	c.NoError(err)

	c.NotEmpty(cache.GetApplication("test_7d8be3591c4dea8566d"))
	c.Len(cache.GetApplications(), 3)
	c.Len(cache.GetApplicationsByUserID("test_2bf67774900350d9c43"), 2)

	c.NotEmpty(cache.GetBlockchain("0021"))
	c.Len(cache.GetBlockchains(), 1)

	c.NotEmpty(cache.GetLoadBalancer("test_bf67774900350d9c42"))
	c.Len(cache.GetLoadBalancers(), 1)
	c.Len(cache.GetLoadBalancersByUserID("test_35fts687463gh2h72gs"), 1)

	c.NotEmpty(cache.GetPayPlan(repository.FreetierV0))
	c.Len(cache.GetPayPlans(), 2)

	c.NotEmpty(cache.GetRedirects("0021"))
	c.Len(cache.GetRedirects("0021"), 1)
}

func Test_V0_Cache_SetCacheFailure(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}
	cache := NewCache(readerMock, logrus.New())

	errOnPay := errors.New("error on pay plans")
	readerMock.On("ReadPayPlans").Return([]*repository.PayPlan{}, errOnPay).Once()

	err := cache.SetCache()
	c.ErrorIs(err, errOnPay)

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

	errOnRedirect := errors.New("error on redirects")
	readerMock.On("ReadRedirects").Return([]*repository.Redirect{}, errOnRedirect).Once()

	err = cache.SetCache()
	c.ErrorIs(err, errOnRedirect)

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

	errOnBlockchain := errors.New("error on blockchains")
	readerMock.On("ReadBlockchains").Return([]*repository.Blockchain{}, errOnBlockchain).Once()

	err = cache.SetCache()
	c.ErrorIs(err, errOnBlockchain)

	readerMock.On("ReadBlockchains").Return([]*repository.Blockchain{
		{
			ID: "0021",
		},
	}, nil)

	errOnApplications := errors.New("error on applications")
	readerMock.On("ReadApplications").Return([]*repository.Application{}, errOnApplications).Once()

	err = cache.SetCache()
	c.ErrorIs(err, errOnApplications)

	readerMock.On("ReadApplications").Return([]*repository.Application{
		{
			ID:     "test_7d8be3591c4dea8566d",
			UserID: "test_2bf67774900350d9c43",
		},
	}, nil)

	errOnLoadBalancer := errors.New("error on loadbalancers")
	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{}, errOnLoadBalancer).Once()

	err = cache.SetCache()
	c.ErrorIs(err, errOnLoadBalancer)

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{
		{
			ID: "test_bf67774900350d9c42",
		},
	}, nil)
}

func Test_V0_Cache_AddApplication(t *testing.T) {
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
			ID:     "test_7d8be3591c4dea8566d",
			UserID: "test_2bf67774900350d9c43",
		},
		{
			ID:     "test_d8be3591c4dea8566a",
			UserID: "test_2bf67774900350d9c43",
		},
		{
			ID:     "test_7d8be3591c4dea8566f",
			UserID: "test_2bf67774900350d9c44",
		},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setPayPlans()
	c.NoError(err)

	err = cache.setApplications()
	c.NoError(err)

	cache.addApplication(repository.Application{
		ID:     "test_7d8be3591c4dea8566b",
		UserID: "test_2bf67774900350d9c43",
		Limit: repository.AppLimit{
			PayPlan: repository.PayPlan{
				Type:  repository.FreetierV0,
				Limit: 250000,
			},
		},
	})

	c.Len(cache.GetApplications(), 4)
	c.Len(cache.GetApplicationsByUserID("test_2bf67774900350d9c43"), 3)
	c.Equal(cache.GetApplication("test_7d8be3591c4dea8566b").DailyLimit(), 250000)
}

func Test_V0_Cache_UpdateApplication(t *testing.T) {
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
			ID:     "test_7d8be3591c4dea8566d",
			UserID: "test_2bf67774900350d9c43",
		},
		{
			ID:     "test_d8be3591c4dea8566a",
			UserID: "test_2bf67774900350d9c43",
		},
		{
			ID:     "test_7d8be3591c4dea8566f",
			UserID: "test_2bf67774900350d9c44",
		},
	}, nil)

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{
		{
			ID:     "test_bf67774900350d9c42",
			UserID: "test_35fts687463gh2h72gs",
			ApplicationIDs: []string{
				"test_7d8be3591c4dea8566d",
				"test_d8be3591c4dea8566a",
			},
		},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setPayPlans()
	c.NoError(err)

	err = cache.setApplications()
	c.NoError(err)

	err = cache.setLoadBalancers()
	c.NoError(err)

	cache.updateApplication(repository.Application{
		ID:     "test_d8be3591c4dea8566a",
		UserID: "test_2bf67774900350d9c43",
		Name:   "papolo",
	})

	c.Len(cache.GetApplications(), 3)
	c.Len(cache.GetApplicationsByUserID("test_2bf67774900350d9c43"), 2)
	c.Len(cache.GetApplicationsByUserID("test_2bf67774900350d9c44"), 1)
	c.Equal("papolo", cache.GetApplication("test_d8be3591c4dea8566a").Name)
	c.Equal("papolo", cache.GetLoadBalancer("test_bf67774900350d9c42").Applications[1].Name)
	c.Equal("papolo", cache.GetApplicationsByUserID("test_2bf67774900350d9c43")[1].Name)

	c.Equal("papolo", cache.GetLoadBalancer("test_bf67774900350d9c42").Applications[1].Name)
	c.Equal("papolo", cache.GetLoadBalancersByUserID("test_35fts687463gh2h72gs")[0].Applications[1].Name)
}

func Test_V0_Cache_RemoveApplication(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadPayPlans").Return([]*repository.PayPlan{
		{
			Type:  repository.FreetierV0,
			Limit: 250000,
		},
	}, nil)

	readerMock.On("ReadApplications").Return([]*repository.Application{
		{
			ID:     "test_7d8be3591c4dea8566d",
			UserID: "test_2bf67774900350d9c43",
		},
		{
			ID:     "test_d8be3591c4dea8566a",
			UserID: "test_2bf67774900350d9c43",
		},
	}, nil)

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{
		{
			ID:     "test_bf67774900350d9c42",
			UserID: "test_2bf67774900350d9c43",
			ApplicationIDs: []string{
				"test_7d8be3591c4dea8566d",
				"test_d8be3591c4dea8566a",
			},
		},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setPayPlans()
	c.NoError(err)
	err = cache.setApplications()
	c.NoError(err)
	err = cache.setLoadBalancers()
	c.NoError(err)
	c.Len(cache.GetApplicationsByUserID("test_2bf67774900350d9c43"), 2)

	cache.updateApplication(repository.Application{
		ID:     "test_d8be3591c4dea8566a",
		UserID: "",
	})
	c.Len(cache.GetApplicationsByUserID("test_2bf67774900350d9c43"), 1)
}

func Test_V0_Cache_AddLoadBalancer(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{
		{
			ID:     "test_7d8be3591c4dea8566d",
			UserID: "test_2bf67774900350d9c43",
		},
		{
			ID:     "test_d8be3591c4dea8566a",
			UserID: "test_2bf67774900350d9c43",
		},
		{
			ID:     "test_7d8be3591c4dea8566f",
			UserID: "test_2bf67774900350d9c44",
		},
	}, nil)

	readerMock.On("ReadApplications").Return([]*repository.Application{
		{
			ID:     "test_7d8be3591c4dea8566d",
			UserID: "test_2bf67774900350d9c43",
		},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setLoadBalancers()
	c.NoError(err)

	err = cache.setApplications()
	c.NoError(err)

	cache.addLoadBalancer(repository.LoadBalancer{
		ID:     "test_7d8be3591c4dea8566b",
		UserID: "test_2bf67774900350d9c43",
	})

	c.Len(cache.GetLoadBalancers(), 4)
	c.Len(cache.GetLoadBalancersByUserID("test_2bf67774900350d9c43"), 3)
}

func Test_V0_Cache_UpdateLoadBalancer(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{
		{
			ID:     "test_7d8be3591c4dea8566d",
			UserID: "test_2bf67774900350d9c43",
		},
		{
			ID:     "test_d8be3591c4dea8566a",
			UserID: "test_2bf67774900350d9c43",
		},
		{
			ID:     "test_7d8be3591c4dea8566f",
			UserID: "test_2bf67774900350d9c44",
		},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setLoadBalancers()
	c.NoError(err)

	cache.updateLoadBalancer(repository.LoadBalancer{
		ID:     "test_d8be3591c4dea8566a",
		UserID: "test_2bf67774900350d9c43",
		Name:   "papolo",
	})

	c.Len(cache.GetLoadBalancers(), 3)
	c.Len(cache.GetLoadBalancersByUserID("test_2bf67774900350d9c43"), 2)
	c.Len(cache.GetLoadBalancersByUserID("test_2bf67774900350d9c44"), 1)
	c.Equal("papolo", cache.GetLoadBalancersByUserID("test_2bf67774900350d9c43")[1].Name)
	c.Equal("papolo", cache.GetLoadBalancer("test_d8be3591c4dea8566a").Name)
}

func Test_V0_Cache_RemoveLoadBalancer(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{
		{
			ID:     "test_7d8be3591c4dea8566d",
			UserID: "test_2bf67774900350d9c43",
		},
		{
			ID:     "test_d8be3591c4dea8566a",
			UserID: "test_2bf67774900350d9c43",
		},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setLoadBalancers()
	c.NoError(err)
	c.Len(cache.GetLoadBalancersByUserID("test_2bf67774900350d9c43"), 2)

	cache.updateLoadBalancer(repository.LoadBalancer{
		ID:     "test_d8be3591c4dea8566a",
		UserID: "",
	})

	c.Len(cache.GetLoadBalancersByUserID("test_2bf67774900350d9c43"), 1)
}

func Test_V0_Cache_AddBlockchain(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadBlockchains").Return([]*repository.Blockchain{
		{ID: "0001", Ticker: "POKT"},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setBlockchains()
	c.NoError(err)

	c.Len(cache.GetBlockchains(), 1)

	cache.addBlockchain(repository.Blockchain{ID: "0002", Ticker: "ETH"})

	c.Len(cache.GetBlockchains(), 2)
}

func Test_V0_Cache_UpdateBlockchain(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadBlockchains").Return([]*repository.Blockchain{
		{ID: "0001", Active: false},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

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

func Test_V0_Cache_AddRedirect(t *testing.T) {
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

	cache := NewCache(readerMock, logrus.New())

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
