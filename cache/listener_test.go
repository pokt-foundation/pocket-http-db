package cache

import (
	"testing"
	"time"

	"github.com/pokt-foundation/portal-db/driver"
	postgresdriver "github.com/pokt-foundation/portal-db/postgres-driver"
	"github.com/pokt-foundation/portal-db/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type ReaderMock struct {
	*driver.MockDriver
	lMock        *postgresdriver.ListenerMock
	notification chan *types.Notification
}

func NewReaderMock(t *testing.T) *ReaderMock {
	mock := &ReaderMock{
		MockDriver:   driver.NewMockDriver(t),
		lMock:        postgresdriver.NewListenerMock(),
		notification: make(chan *types.Notification, 32),
	}

	go postgresdriver.Listen(mock.lMock.NotificationChannel(), mock.notification)

	return mock
}

func newMockCache(readerMock *ReaderMock) *Cache {
	readerMock.On("ReadApplications").Return([]*types.Application{
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

	readerMock.On("ReadBlockchains").Return([]*types.Blockchain{
		{ID: "0021"},
	}, nil)

	readerMock.On("ReadLoadBalancers").Return([]*types.LoadBalancer{
		{
			ID:     "60ecb2bf67774900350d9c42",
			UserID: "60ecb35fts687463gh2h72gs",
			ApplicationIDs: []string{
				"5f62b7d8be3591c4dea8566d",
				"5f62b7d8be3591c4dea8566a",
			},
		},
	}, nil)

	readerMock.On("ReadPayPlans").Return([]*types.PayPlan{
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
	if err != nil {
		panic(err)
	}

	return cache
}

func TestCache_listenApplication(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)
	cache := newMockCache(readerMock)

	readerMock.lMock.MockEvent(types.ActionInsert, types.ActionInsert, &types.Application{
		ID:     "321",
		UserID: "user_id_12345",
		Name:   "pablo",
		GatewayAAT: types.GatewayAAT{
			Address: "123",
		},
		GatewaySettings: types.GatewaySettings{
			SecretKey: "123",
		},
		NotificationSettings: types.NotificationSettings{
			Full: true,
		},
		Limit: types.AppLimit{
			PayPlan: types.PayPlan{
				Type:  types.FreetierV0,
				Limit: 250000,
			},
		},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	app := cache.GetApplication("321")
	c.Equal("pablo", app.Name)
	c.Equal("123", app.GatewayAAT.Address)
	c.Equal("123", app.GatewaySettings.SecretKey)
	c.Equal(types.PayPlanType("FREETIER_V0"), app.Limit.PayPlan.Type)
	c.Equal(250000, app.DailyLimit())
	c.True(app.NotificationSettings.Full)

	readerMock.lMock.MockEvent(types.ActionUpdate, types.ActionUpdate, &types.Application{
		ID:     "321",
		UserID: "user_id_12345",
		Name:   "orlando",
		GatewaySettings: types.GatewaySettings{
			SecretKey: "1234",
		},
		NotificationSettings: types.NotificationSettings{
			Full:     true,
			SignedUp: true,
		},
		Limit: types.AppLimit{
			PayPlan: types.PayPlan{
				Type: types.Enterprise,
			},
			CustomLimit: 2000000,
		},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	app = cache.GetApplication("321")
	c.Equal("orlando", app.Name)
	c.Equal("1234", app.GatewaySettings.SecretKey)
	c.Equal(types.PayPlanType("ENTERPRISE"), app.Limit.PayPlan.Type)
	c.Equal(2000000, app.DailyLimit())

	// Delete application event
	apps := cache.GetApplicationsByUserID("user_id_12345")
	c.NotEmpty(apps)

	readerMock.lMock.MockEvent(types.ActionUpdate, types.ActionUpdate, &types.Application{
		ID:     "321",
		UserID: "",
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	apps = cache.GetApplicationsByUserID("user_id_12345")
	c.Empty(apps)
}

func TestCache_listenAppLimit(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)
	cache := newMockCache(readerMock)

	cache.payPlansMap = map[types.PayPlanType]*types.PayPlan{
		types.PayAsYouGoV0: {Type: types.PayAsYouGoV0, Limit: 0},
	}

	readerMock.lMock.MockEvent(types.ActionInsert, types.ActionInsert, &types.AppLimit{
		ID:      "321",
		PayPlan: types.PayPlan{Type: types.PayAsYouGoV0, Limit: 0},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	app := cache.GetApplication("321")
	c.Nil(app)

	pendingUpdates := cache.pendingAppLimit
	c.Len(pendingUpdates, 1)
	c.Equal(types.PayAsYouGoV0, cache.pendingAppLimit["321"].PayPlan.Type)

	readerMock.lMock.MockEvent(types.ActionInsert, types.ActionInsert, &types.Application{
		ID: "321",
		Limit: types.AppLimit{
			PayPlan: types.PayPlan{
				Type:  types.PayAsYouGoV0,
				Limit: 0,
			},
		},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	app = cache.GetApplication("321")
	c.Equal(types.PayPlanType("PAY_AS_YOU_GO_V0"), app.Limit.PayPlan.Type)
	c.Equal(0, app.DailyLimit())

	pendingUpdates = cache.pendingAppLimit
	c.Len(pendingUpdates, 0)
}

func TestCache_listenBlockchain(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)
	cache := newMockCache(readerMock)

	readerMock.lMock.MockEvent(types.ActionInsert, types.ActionInsert, &types.Blockchain{
		ID:   "0023",
		Path: "path",
		SyncCheckOptions: types.SyncCheckOptions{
			BlockchainID: "0023",
			Body:         "yeh",
		},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	blockchain := cache.GetBlockchain("0023")
	c.Equal("path", blockchain.Path)
	c.Equal("yeh", blockchain.SyncCheckOptions.Body)

	readerMock.lMock.MockEvent(types.ActionUpdate, types.ActionUpdate, &types.Blockchain{
		ID:     "0023",
		Active: true,
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	blockchain = cache.GetBlockchain("0023")
	c.True(blockchain.Active)
}

func TestCache_listenLoadBalancer(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)
	cache := newMockCache(readerMock)

	readerMock.lMock.MockEvent(types.ActionInsert, types.ActionInsert, &types.LoadBalancer{
		ID:     "123",
		UserID: "user_id_12345",
		Name:   "pablo",
		StickyOptions: types.StickyOptions{
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

	readerMock.lMock.MockEvent(types.ActionUpdate, types.ActionUpdate, &types.LoadBalancer{
		ID:     "123",
		UserID: "user_id_12345",
		Name:   "orlando",
		StickyOptions: types.StickyOptions{
			StickyOrigins: []string{"ohana"},
			Stickiness:    true,
		},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	lb = cache.GetLoadBalancer("123")
	c.Equal("orlando", lb.Name)
	c.Equal([]string{"ohana"}, lb.StickyOptions.StickyOrigins)

	// Delete load balancer event
	lbs := cache.GetLoadBalancersByUserID("user_id_12345")
	c.NotEmpty(lbs)

	readerMock.lMock.MockEvent(types.ActionUpdate, types.ActionUpdate, &types.LoadBalancer{
		ID:     "123",
		UserID: "",
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	lbs = cache.GetLoadBalancersByUserID("user_id_12345")
	c.Empty(lbs)
}

func TestCache_listenRedirect(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)
	cache := newMockCache(readerMock)

	readerMock.lMock.MockEvent(types.ActionInsert, types.ActionInsert, &types.Redirect{
		BlockchainID: "0021",
		Alias:        "papolo",
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	redirects := cache.GetBlockchain("0021").Redirects
	c.Equal("papolo", redirects[1].Alias)
}
