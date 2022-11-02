package cache

import (
	"testing"
	"time"

	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/stretchr/testify/require"
)

func newMockCache(readerMock *ReaderMock) *Cache {
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
	if err != nil {
		panic(err)
	}

	return cache
}

func TestCache_listenApplication(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock()
	cache := newMockCache(readerMock)

	readerMock.lMock.MockEvent(repository.ActionInsert, repository.ActionInsert, &repository.Application{
		ID:   "321",
		Name: "pablo",
		GatewayAAT: repository.GatewayAAT{
			Address: "123",
		},
		GatewaySettings: repository.GatewaySettings{
			SecretKey: "123",
		},
		NotificationSettings: repository.NotificationSettings{
			Full: true,
		},
		Limit: repository.AppLimit{
			PayPlan: repository.PayPlan{
				Type:  repository.FreetierV0,
				Limit: 250000,
			},
		},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	app := cache.GetApplication("321")
	c.Equal("pablo", app.Name)
	c.Equal("123", app.GatewayAAT.Address)
	c.Equal("123", app.GatewaySettings.SecretKey)
	c.Equal(repository.PayPlanType("FREETIER_V0"), app.Limit.PayPlan.Type)
	c.Equal(250000, app.DailyLimit())
	c.True(app.NotificationSettings.Full)

	readerMock.lMock.MockEvent(repository.ActionUpdate, repository.ActionUpdate, &repository.Application{
		ID:   "321",
		Name: "orlando",
		GatewaySettings: repository.GatewaySettings{
			SecretKey: "1234",
		},
		NotificationSettings: repository.NotificationSettings{
			Full:     true,
			SignedUp: true,
		},
		Limit: repository.AppLimit{
			PayPlan: repository.PayPlan{
				Type: repository.Enterprise,
			},
			CustomLimit: 2000000,
		},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	app = cache.GetApplication("321")
	c.Equal("orlando", app.Name)
	c.Equal("1234", app.GatewaySettings.SecretKey)
	c.Equal(repository.PayPlanType("ENTERPRISE"), app.Limit.PayPlan.Type)
	c.Equal(2000000, app.DailyLimit())
	c.True(app.NotificationSettings.SignedUp)
}

func TestCache_listenBlockchain(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock()
	cache := newMockCache(readerMock)

	readerMock.lMock.MockEvent(repository.ActionInsert, repository.ActionInsert, &repository.Blockchain{
		ID:   "0023",
		Path: "path",
		SyncCheckOptions: repository.SyncCheckOptions{
			BlockchainID: "0023",
			Body:         "yeh",
		},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	blockchain := cache.GetBlockchain("0023")
	c.Equal("path", blockchain.Path)
	c.Equal("yeh", blockchain.SyncCheckOptions.Body)

	readerMock.lMock.MockEvent(repository.ActionUpdate, repository.ActionUpdate, &repository.Blockchain{
		ID:     "0023",
		Active: true,
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	blockchain = cache.GetBlockchain("0023")
	c.True(blockchain.Active)
}

func TestCache_listenLoadBalancer(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock()
	cache := newMockCache(readerMock)

	readerMock.lMock.MockEvent(repository.ActionInsert, repository.ActionInsert, &repository.LoadBalancer{
		ID:   "123",
		Name: "pablo",
		StickyOptions: repository.StickyOptions{
			StickyOrigins: []string{"oahu"},
			Stickiness:    true,
		},
		ApplicationIDs: []string{"5f62b7d8be3591c4dea8566a"},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	lb := cache.GetLoadBalancer("123")
	c.Equal("pablo", lb.Name)
	c.Equal([]string{"oahu"}, lb.StickyOptions.StickyOrigins)
	c.Equal("5f62b7d8be3591c4dea8566a", lb.Applications[0].ID)

	readerMock.lMock.MockEvent(repository.ActionUpdate, repository.ActionUpdate, &repository.LoadBalancer{
		ID:   "123",
		Name: "orlando",
		StickyOptions: repository.StickyOptions{
			StickyOrigins: []string{"ohana"},
			Stickiness:    true,
		},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	lb = cache.GetLoadBalancer("123")
	c.Equal("orlando", lb.Name)
	c.Equal([]string{"ohana"}, lb.StickyOptions.StickyOrigins)
}

func TestCache_listenRedirect(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock()
	cache := newMockCache(readerMock)

	readerMock.lMock.MockEvent(repository.ActionInsert, repository.ActionInsert, &repository.Redirect{
		BlockchainID: "0021",
		Alias:        "papolo",
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	redirects := cache.GetRedirects("0021")
	c.Equal("papolo", redirects[1].Alias)
}
