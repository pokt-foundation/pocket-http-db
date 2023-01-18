package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/pokt-foundation/portal-db/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var testCtx = context.Background()

func TestCache_SetCache(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)

	readerMock.On("ReadApplications", testCtx).Return([]*types.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
			Limit: types.AppLimit{
				PayPlan: types.PayPlan{
					Type:  types.FreetierV0,
					Limit: 250000,
				},
			},
		},
		{
			ID:     "5f62b7d8be3591c4dea8566a",
			UserID: "60ecb2bf67774900350d9c43",
			Limit: types.AppLimit{
				PayPlan: types.PayPlan{
					Type: types.Enterprise,
				},
				CustomLimit: 2000000,
			},
		},
		{
			ID:     "5f62b7d8be3591c4dea8566f",
			UserID: "60ecb2bf67774900350d9c44",
			Limit: types.AppLimit{
				PayPlan: types.PayPlan{
					Type:  types.PayAsYouGoV0,
					Limit: 0,
				},
			},
		},
	}, nil)

	readerMock.On("ReadBlockchains", testCtx).Return([]*types.Blockchain{
		{ID: "0021"},
	}, nil)

	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{
		{
			ID:     "60ecb2bf67774900350d9c42",
			UserID: "60ecb35fts687463gh2h72gs",
			ApplicationIDs: []string{
				"5f62b7d8be3591c4dea8566d",
				"5f62b7d8be3591c4dea8566a",
			},
		},
	}, nil)

	readerMock.On("ReadPayPlans", testCtx).Return([]*types.PayPlan{
		{
			Type:  types.FreetierV0,
			Limit: 250000,
		},
		{
			Type:  types.PayAsYouGoV0,
			Limit: 0,
		},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

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

	c.NotEmpty(cache.GetPayPlan(types.FreetierV0))
	c.Len(cache.GetPayPlans(), 2)
}

func TestCache_SetCacheFailure(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)
	cache := NewCache(readerMock, logrus.New())

	errOnPay := errors.New("error on pay plans")
	readerMock.On("ReadPayPlans", testCtx).Return([]*types.PayPlan{}, errOnPay).Once()

	err := cache.SetCache()
	c.ErrorIs(err, errOnPay)

	readerMock.On("ReadPayPlans", testCtx).Return([]*types.PayPlan{
		{
			Type:  types.FreetierV0,
			Limit: 250000,
		},
		{
			Type:  types.PayAsYouGoV0,
			Limit: 0,
		},
	}, nil)

	errOnBlockchain := errors.New("error on blockchains")
	readerMock.On("ReadBlockchains", testCtx).Return([]*types.Blockchain{}, errOnBlockchain).Once()

	err = cache.SetCache()
	c.ErrorIs(err, errOnBlockchain)

	readerMock.On("ReadBlockchains", testCtx).Return([]*types.Blockchain{
		{
			ID: "0021",
		},
	}, nil)

	errOnApplications := errors.New("error on applications")
	readerMock.On("ReadApplications", testCtx).Return([]*types.Application{}, errOnApplications).Once()

	err = cache.SetCache()
	c.ErrorIs(err, errOnApplications)

	readerMock.On("ReadApplications", testCtx).Return([]*types.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
		},
	}, nil)

	errOnLoadBalancer := errors.New("error on loadbalancers")
	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{}, errOnLoadBalancer).Once()

	err = cache.SetCache()
	c.ErrorIs(err, errOnLoadBalancer)

	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{
		{
			ID: "60ecb2bf67774900350d9c42",
		},
	}, nil)
}

func TestCache_AddApplication(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)

	readerMock.On("ReadPayPlans", testCtx).Return([]*types.PayPlan{
		{
			Type:  types.FreetierV0,
			Limit: 250000,
		},
		{
			Type:  types.PayAsYouGoV0,
			Limit: 0,
		},
	}, nil)

	readerMock.On("ReadApplications", testCtx).Return([]*types.Application{
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

	cache := NewCache(readerMock, logrus.New())

	err := cache.setPayPlans()
	c.NoError(err)

	err = cache.setApplications()
	c.NoError(err)

	cache.addApplication(types.Application{
		ID:     "5f62b7d8be3591c4dea8566b",
		UserID: "60ecb2bf67774900350d9c43",
		Limit: types.AppLimit{
			PayPlan: types.PayPlan{
				Type:  types.FreetierV0,
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

	readerMock := NewReaderMock(t)

	readerMock.On("ReadPayPlans", testCtx).Return([]*types.PayPlan{
		{
			Type:  types.FreetierV0,
			Limit: 250000,
		},
		{
			Type:  types.PayAsYouGoV0,
			Limit: 0,
		},
	}, nil)

	readerMock.On("ReadApplications", testCtx).Return([]*types.Application{
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

	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{
		{
			ID:     "60ecb2bf67774900350d9c42",
			UserID: "60ecb35fts687463gh2h72gs",
			ApplicationIDs: []string{
				"5f62b7d8be3591c4dea8566d",
				"5f62b7d8be3591c4dea8566a",
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

	cache.updateApplication(types.Application{
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

func TestCache_RemoveApplication(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)

	readerMock.On("ReadPayPlans", testCtx).Return([]*types.PayPlan{
		{
			Type:  types.FreetierV0,
			Limit: 250000,
		},
	}, nil)

	readerMock.On("ReadApplications", testCtx).Return([]*types.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
		},
		{
			ID:     "5f62b7d8be3591c4dea8566a",
			UserID: "60ecb2bf67774900350d9c43",
		},
	}, nil)

	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{
		{
			ID:     "60ecb2bf67774900350d9c42",
			UserID: "60ecb2bf67774900350d9c43",
			ApplicationIDs: []string{
				"5f62b7d8be3591c4dea8566d",
				"5f62b7d8be3591c4dea8566a",
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
	c.Len(cache.GetApplicationsByUserID("60ecb2bf67774900350d9c43"), 2)

	cache.updateApplication(types.Application{
		ID:     "5f62b7d8be3591c4dea8566a",
		UserID: "",
	})
	c.Len(cache.GetApplicationsByUserID("60ecb2bf67774900350d9c43"), 1)
}

func TestCache_AddLoadBalancer(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)

	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{
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

	readerMock.On("ReadApplications", testCtx).Return([]*types.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
		},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setLoadBalancers()
	c.NoError(err)

	err = cache.setApplications()
	c.NoError(err)

	cache.addLoadBalancer(types.LoadBalancer{
		ID:     "5f62b7d8be3591c4dea8566b",
		UserID: "60ecb2bf67774900350d9c43",
	})

	c.Len(cache.GetLoadBalancers(), 4)
	c.Len(cache.GetLoadBalancersByUserID("60ecb2bf67774900350d9c43"), 3)
}

func TestCache_UpdateLoadBalancer(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)

	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{
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

	cache := NewCache(readerMock, logrus.New())

	err := cache.setLoadBalancers()
	c.NoError(err)

	cache.updateLoadBalancer(types.LoadBalancer{
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

func TestCache_RemoveLoadBalancer(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
		},
		{
			ID:     "5f62b7d8be3591c4dea8566a",
			UserID: "60ecb2bf67774900350d9c43",
		},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setLoadBalancers()
	c.NoError(err)
	c.Len(cache.GetLoadBalancersByUserID("60ecb2bf67774900350d9c43"), 2)

	cache.updateLoadBalancer(types.LoadBalancer{
		ID:     "5f62b7d8be3591c4dea8566a",
		UserID: "",
	})

	c.Len(cache.GetLoadBalancersByUserID("60ecb2bf67774900350d9c43"), 1)
}

func TestCache_AddBlockchain(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)

	readerMock.On("ReadBlockchains", testCtx).Return([]*types.Blockchain{
		{ID: "0001", Ticker: "POKT"},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setBlockchains()
	c.NoError(err)

	c.Len(cache.GetBlockchains(), 1)

	cache.addBlockchain(types.Blockchain{ID: "0002", Ticker: "ETH"})

	c.Len(cache.GetBlockchains(), 2)
}

func TestCache_UpdateBlockchain(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)

	readerMock.On("ReadBlockchains", testCtx).Return([]*types.Blockchain{
		{ID: "0001", Active: false},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setBlockchains()
	c.NoError(err)

	c.Len(cache.GetBlockchains(), 1)
	c.Equal(cache.GetBlockchains()[0].Active, false)

	cache.updateBlockchain(types.Blockchain{
		ID:     "0001",
		Active: true,
	})

	c.Len(cache.GetBlockchains(), 1)
	c.Equal(cache.GetBlockchains()[0].Active, true)
}

func TestCache_AddRedirect(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)

	readerMock.On("ReadBlockchains", testCtx).Return([]*types.Blockchain{
		{ID: "0001", Ticker: "POKT", Redirects: []types.Redirect{
			{BlockchainID: "0001", Alias: "pokt-mainnet-1"},
			{BlockchainID: "0001", Alias: "pokt-mainnet-2"},
		}},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setBlockchains()
	c.NoError(err)

	c.Len(cache.GetBlockchains(), 1)
	c.Len(cache.GetBlockchains()[0].Redirects, 2)
	c.Len(cache.GetBlockchain("0001").Redirects, 2)

	cache.addRedirect(types.Redirect{BlockchainID: "0001", Alias: "pokt-mainnet-3"})

	c.Len(cache.GetBlockchains()[0].Redirects, 3)
	c.Len(cache.GetBlockchain("0001").Redirects, 3)
}
