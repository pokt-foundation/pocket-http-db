package v1_cache

import (
	"testing"
	"time"

	"github.com/pokt-foundation/portal-db/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func newMockCache(readerMock *ReaderMock) *Cache {
	readerMock.On("ReadApplications", testCtx).Return([]*types.Application{
		{
			ID:     "test_app_e3591c4dea8566d",
			UserID: "test_user_774900350d9c43",
			Limit: types.AppLimit{
				PayPlan: types.PayPlan{
					Type:  types.FreetierV0,
					Limit: 250000,
				},
			},
		},
		{
			ID:     "test_app_e3591c4dea8566a",
			UserID: "test_user_774900350d9c43",
			Limit: types.AppLimit{
				PayPlan: types.PayPlan{
					Type: types.Enterprise,
				},
				CustomLimit: 2000000,
			},
		},
		{
			ID:     "test_app_e3591c4dea8566f",
			UserID: "test_user_774900350d9c44",
			Limit: types.AppLimit{
				PayPlan: types.PayPlan{
					Type:  types.PayAsYouGoV0,
					Limit: 0,
				},
			},
		},
	}, nil)

	readerMock.On("ReadBlockchains", testCtx).Return([]*types.Blockchain{
		{ID: "0021", Redirects: []types.Redirect{}},
	}, nil)

	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{
		{
			ID:     "test_lb_67774900350d9c42",
			UserID: "test_user_687463gh2h72gs",
			ApplicationIDs: []string{
				"test_app_e3591c4dea8566d",
				"test_app_e3591c4dea8566a",
			},
			Users: []types.UserAccess{
				{UserID: "test_user_687463gh2h72gs", RoleName: types.RoleOwner},
				{UserID: "test_user_47fhf7rwejf943", RoleName: types.RoleAdmin},
			},
		},
	}, nil)

	readerMock.On("ReadUserPermissions", testCtx).Return(map[types.UserID]types.UserPermissions{
		"test_user_687463gh2h72gs": {
			UserID: "test_user_687463gh2h72gs",
			LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{
				"test_lb_67774900350d9c42": {
					RoleName:    types.RoleOwner,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint},
				},
			},
		},
		"test_user_47fhf7rwejf943": {
			UserID: "test_user_47fhf7rwejf943",
			LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{
				"test_lb_67774900350d9c42": {
					RoleName:    types.RoleAdmin,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint},
				},
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
	if err != nil {
		panic(err)
	}

	return cache
}

func Test_V1_Cache_listenApplication(t *testing.T) {
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

func Test_V1_Cache_listenAppLimit(t *testing.T) {
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

func Test_V1_Cache_listenBlockchain(t *testing.T) {
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

func Test_V1_Cache_listenLoadBalancer(t *testing.T) {
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
		Users: []types.UserAccess{
			{RoleName: "OWNER", UserID: "user_id_12345", Email: "owner1@test.com", Accepted: true},
		},
		ApplicationIDs: []string{"test_app_e3591c4dea8566a"},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	lb := cache.GetLoadBalancer("123")
	c.Equal("pablo", lb.Name)
	c.Equal([]string{"oahu"}, lb.StickyOptions.StickyOrigins)
	c.Equal("test_app_e3591c4dea8566a", lb.Applications[0].ID)
	c.Equal("user_id_12345", lb.Users[0].UserID)
	c.Equal("owner1@test.com", lb.Users[0].Email)

	// Update load balancer event (including updating User Access)
	readerMock.lMock.MockEvent(types.ActionUpdate, types.ActionUpdate, &types.LoadBalancer{
		ID:     "123",
		UserID: "user_id_12345",
		Name:   "orlando",
		StickyOptions: types.StickyOptions{
			StickyOrigins: []string{"ohana"},
			Stickiness:    true,
		},
		Users: []types.UserAccess{
			{RoleName: "ADMIN", UserID: "user_id_12345", Email: "owner1@test.com", Accepted: true},
		},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	lb = cache.GetLoadBalancer("123")
	c.Equal("orlando", lb.Name)
	c.Equal([]string{"ohana"}, lb.StickyOptions.StickyOrigins)
	c.Equal("user_id_12345", lb.Users[0].UserID)
	c.Equal(types.RoleAdmin, lb.Users[0].RoleName)
	c.Len(cache.GetLoadBalancersByUserID("user_id_12345"), 1)
	c.Len(cache.GetUserPermissionsByUserID("user_id_12345").LoadBalancers, 1)
	c.Equal(types.RoleAdmin, cache.GetUserPermissionsByUserID("user_id_12345").LoadBalancers["123"].RoleName)

	// Add new load balancer User Access event
	readerMock.lMock.MockEvent(types.ActionUpdate, types.ActionInsert, &types.LoadBalancer{
		ID:     "123",
		UserID: "user_id_12345",
		Name:   "orlando",
		StickyOptions: types.StickyOptions{
			StickyOrigins: []string{"ohana"},
			Stickiness:    true,
		},
		Users: []types.UserAccess{
			{RoleName: "ADMIN", UserID: "test_user_admin4309tug4g", Email: "admin1@test.com", Accepted: true},
		},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	lb = cache.GetLoadBalancer("123")
	c.Len(lb.Users, 2)
	c.Equal("user_id_12345", lb.Users[0].UserID)
	c.Equal("test_user_admin4309tug4g", lb.Users[1].UserID)
	c.Equal("admin1@test.com", lb.Users[1].Email)
	c.Len(cache.GetLoadBalancersByUserID("test_user_admin4309tug4g"), 1)
	c.Len(cache.GetUserPermissionsByUserID("test_user_admin4309tug4g").LoadBalancers, 1)
	c.Equal(types.RoleAdmin, cache.GetUserPermissionsByUserID("test_user_admin4309tug4g").LoadBalancers["123"].RoleName)

	// Delete load balancer User Access event
	readerMock.lMock.MockEvent(types.ActionUpdate, types.ActionDelete, &types.LoadBalancer{
		ID:     "123",
		UserID: "user_id_12345",
		Name:   "orlando",
		StickyOptions: types.StickyOptions{
			StickyOrigins: []string{"ohana"},
			Stickiness:    true,
		},
		Users: []types.UserAccess{
			{RoleName: "ADMIN", UserID: "test_user_admin4309tug4g", Email: "admin1@test.com", Accepted: true},
		},
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	lb = cache.GetLoadBalancer("123")
	c.Len(lb.Users, 1)
	c.Equal("user_id_12345", lb.Users[0].UserID)
	c.Len(cache.GetLoadBalancersByUserID("test_user_admin4309tug4g"), 0)
	c.Len(cache.GetUserPermissionsByUserID("test_user_admin4309tug4g").LoadBalancers, 0)

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

func Test_V1_Cache_listenRedirect(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)
	cache := newMockCache(readerMock)

	readerMock.lMock.MockEvent(types.ActionInsert, types.ActionInsert, &types.Redirect{
		BlockchainID: "0021",
		Alias:        "papolo",
	})

	time.Sleep(1 * time.Second) // need time for cache refresh

	redirects := cache.GetBlockchain("0021").Redirects
	c.Equal("papolo", redirects[0].Alias)
}
