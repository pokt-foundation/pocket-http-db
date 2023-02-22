package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pokt-foundation/portal-db/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var (
	testCtx     = context.Background()
	testRefresh = time.Minute * 5
)

func NewTestCache(t *testing.T) *Cache {
	c := require.New(t)

	readerMock := NewReaderMock(t)

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
		{ID: "0021", Active: true},
		{
			ID:     "0001",
			Active: false,
			Ticker: "POKT",
			Redirects: []types.Redirect{
				{BlockchainID: "0001", Alias: "pokt-mainnet-1"},
				{BlockchainID: "0001", Alias: "pokt-mainnet-2"},
			},
		},
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
		{
			ID:     "test_lb_4nf37fhfy27dj28d",
			UserID: "test_user_774900350d9c43",
			ApplicationIDs: []string{
				"test_app_e3591c4dea8566f",
			},
			Users: []types.UserAccess{
				{UserID: "test_user_774900350d9c43", RoleName: types.RoleOwner},
				{UserID: "test_user_47fhf7rwejf943", RoleName: types.RoleMember},
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

	readerMock.On("ReadUserPermissions", testCtx).Return(map[types.UserID]*types.UserPermissions{
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
				"test_lb_4nf37fhfy27dj28d": {
					RoleName:    types.RoleMember,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint},
				},
			},
		},
		"test_user_774900350d9c43": {
			UserID: "test_user_774900350d9c43",
			LoadBalancers: map[types.LoadBalancerID]types.LoadBalancerPermissions{
				"test_lb_4nf37fhfy27dj28d": {
					RoleName:    types.RoleOwner,
					Permissions: []types.PermissionsEnum{types.ReadEndpoint, types.WriteEndpoint},
				},
			},
		},
	}, nil)

	cache, err := NewCache(readerMock, testRefresh, logrus.New())
	c.NoError(err)

	return cache
}

func TestCache_SetCache(t *testing.T) {
	c := require.New(t)

	cache := NewTestCache(t)

	c.NotEmpty(cache.GetApplication("test_app_e3591c4dea8566d"))
	c.Len(cache.GetApplications(), 3)
	c.Len(cache.GetApplicationsByUserID("test_user_774900350d9c43"), 2)

	c.NotEmpty(cache.GetBlockchain("0001"))
	c.NotEmpty(cache.GetBlockchain("0021"))
	c.Len(cache.GetBlockchains(), 2)

	c.NotEmpty(cache.GetLoadBalancer("test_lb_67774900350d9c42"))
	c.Len(cache.GetLoadBalancers(), 2)
	c.Len(cache.GetLoadBalancersByUserID("test_user_687463gh2h72gs"), 1)
	c.Len(cache.GetLoadBalancersByUserID("test_user_774900350d9c43"), 1)
	c.Len(cache.GetLoadBalancersByUserID("test_user_47fhf7rwejf943"), 2)

	c.Len(cache.GetUserPermissionsByUserID("test_user_687463gh2h72gs").LoadBalancers, 1)
	c.Len(cache.GetUserPermissionsByUserID("test_user_47fhf7rwejf943").LoadBalancers, 2)
	c.Len(cache.GetUserPermissionsByUserID("test_user_774900350d9c43").LoadBalancers, 1)

	c.NotEmpty(cache.GetPayPlan(types.FreetierV0))
	c.Len(cache.GetPayPlans(), 2)
}

func TestCache_SetCacheFailure(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name         string
		failedMethod string
		err          error
	}{
		{
			name:         "should fail if ReadPayPlans fails",
			failedMethod: "ReadPayPlans",
			err:          errors.New("error on pay plans"),
		},
		{
			name:         "should fail if ReadBlockchains fails",
			failedMethod: "ReadBlockchains",
			err:          errors.New("error on blockchains"),
		},
		{
			name:         "should fail if ReadApplications fails",
			failedMethod: "ReadApplications",
			err:          errors.New("error on applications"),
		},
		{
			name:         "should fail if ReadLoadBalancers fails",
			failedMethod: "ReadLoadBalancers",
			err:          errors.New("error on loadbalancers"),
		},
		{
			name:         "should fail if ReadUserPermissions fails",
			failedMethod: "ReadUserPermissions",
			err:          errors.New("error on user permissions"),
		},
	}

	for _, test := range tests {
		readerMock := NewReaderMock(t)

		if test.failedMethod == "ReadBlockchains" {
			readerMock.On("ReadPayPlans", testCtx).Return([]*types.PayPlan{}, nil)
		}
		if test.failedMethod == "ReadApplications" {
			readerMock.On("ReadPayPlans", testCtx).Return([]*types.PayPlan{}, nil)
			readerMock.On("ReadBlockchains", testCtx).Return([]*types.Blockchain{}, nil)
		}
		if test.failedMethod == "ReadLoadBalancers" {
			readerMock.On("ReadPayPlans", testCtx).Return([]*types.PayPlan{}, nil)
			readerMock.On("ReadBlockchains", testCtx).Return([]*types.Blockchain{}, nil)
			readerMock.On("ReadApplications", testCtx).Return([]*types.Application{}, nil)
		}
		if test.failedMethod == "ReadUserPermissions" {
			readerMock.On("ReadPayPlans", testCtx).Return([]*types.PayPlan{}, nil)
			readerMock.On("ReadBlockchains", testCtx).Return([]*types.Blockchain{}, nil)
			readerMock.On("ReadApplications", testCtx).Return([]*types.Application{}, nil)
			readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{}, nil)
		}

		readerMock.On(test.failedMethod, testCtx).Return(nil, test.err)

		_, err := NewCache(readerMock, testRefresh, logrus.New())
		c.ErrorIs(err, test.err)
	}
}

func TestCache_AddApplication(t *testing.T) {
	c := require.New(t)

	cache := NewTestCache(t)

	cache.addApplication(types.Application{
		ID:     "test_app_e3591c4dea8566b",
		UserID: "test_user_774900350d9c43",
		Limit: types.AppLimit{
			PayPlan: types.PayPlan{
				Type:  types.FreetierV0,
				Limit: 250000,
			},
		},
	})

	c.Len(cache.GetApplications(), 4)
	c.Len(cache.GetApplicationsByUserID("test_user_774900350d9c43"), 3)
	c.Equal(cache.GetApplication("test_app_e3591c4dea8566b").Limit.PayPlan.Limit, 250000)
}

func TestCache_UpdateApplication(t *testing.T) {
	c := require.New(t)

	cache := NewTestCache(t)

	cache.updateApplication(types.Application{
		ID:     "test_app_e3591c4dea8566a",
		UserID: "test_user_774900350d9c43",
		Name:   "papolo",
	})

	c.Len(cache.GetApplications(), 3)
	c.Len(cache.GetApplicationsByUserID("test_user_774900350d9c43"), 2)
	c.Len(cache.GetApplicationsByUserID("test_user_774900350d9c44"), 1)
	c.Equal("papolo", cache.GetApplication("test_app_e3591c4dea8566a").Name)
	c.Equal("papolo", cache.GetLoadBalancer("test_lb_67774900350d9c42").Applications[1].Name)
	c.Equal("papolo", cache.GetApplicationsByUserID("test_user_774900350d9c43")[1].Name)

	c.Equal("papolo", cache.GetLoadBalancer("test_lb_67774900350d9c42").Applications[1].Name)
	c.Equal("papolo", cache.GetLoadBalancersByUserID("test_user_687463gh2h72gs")[0].Applications[1].Name)
}

func TestCache_RemoveApplication(t *testing.T) {
	c := require.New(t)

	cache := NewTestCache(t)

	err := cache.setPayPlans()
	c.NoError(err)
	err = cache.setApplications()
	c.NoError(err)
	err = cache.setLoadBalancers()
	c.NoError(err)
	c.Len(cache.GetApplicationsByUserID("test_user_774900350d9c43"), 2)

	cache.updateApplication(types.Application{
		ID:     "test_app_e3591c4dea8566a",
		Status: types.AwaitingGracePeriod,
	})
	c.Len(cache.GetApplicationsByUserID("test_user_774900350d9c43"), 1)
}

func TestCache_AddLoadBalancer(t *testing.T) {
	c := require.New(t)

	cache := NewTestCache(t)

	cache.addLoadBalancer(types.LoadBalancer{
		ID:     "test_lb_47cjd72kd82kd98d",
		UserID: "test_user_774900350d9c43",
		Users: []types.UserAccess{
			{UserID: "test_user_774900350d9c43", RoleName: types.RoleOwner},
			{UserID: "test_user_774900350d9c44", RoleName: types.RoleAdmin},
		},
	})

	c.Len(cache.GetLoadBalancers(), 3)
	c.Len(cache.GetLoadBalancersByUserID("test_user_774900350d9c43"), 2)
	c.Len(cache.GetLoadBalancersByUserID("test_user_774900350d9c44"), 1)
}

func TestCache_UpdateLoadBalancer(t *testing.T) {
	c := require.New(t)

	cache := NewTestCache(t)

	cache.updateLoadBalancer(types.LoadBalancer{
		ID:     "test_lb_4nf37fhfy27dj28d",
		UserID: "test_user_774900350d9c43",
		Name:   "papolo",
	})

	c.Len(cache.GetLoadBalancers(), 2)
	c.Len(cache.GetLoadBalancersByUserID("test_user_774900350d9c43"), 1)
	c.Len(cache.GetLoadBalancersByUserID("test_user_774900350d9c44"), 0)
	c.Equal("papolo", cache.GetLoadBalancersByUserID("test_user_774900350d9c43")[0].Name)
	c.Equal("papolo", cache.GetLoadBalancer("test_lb_4nf37fhfy27dj28d").Name)
}

func TestCache_RemoveLoadBalancer(t *testing.T) {
	c := require.New(t)

	cache := NewTestCache(t)

	c.Len(cache.GetLoadBalancersByUserID("test_user_774900350d9c43"), 1)

	cache.updateLoadBalancer(types.LoadBalancer{
		ID:     "test_lb_4nf37fhfy27dj28d",
		UserID: "",
	})

	c.Len(cache.GetLoadBalancersByUserID("test_user_774900350d9c43"), 0)
}

func TestCache_AddBlockchain(t *testing.T) {
	c := require.New(t)

	cache := NewTestCache(t)

	c.Len(cache.GetBlockchains(), 2)

	cache.addBlockchain(types.Blockchain{ID: "0002", Ticker: "ETH"})

	c.Len(cache.GetBlockchains(), 3)
}

func TestCache_UpdateBlockchain(t *testing.T) {
	c := require.New(t)

	cache := NewTestCache(t)

	cache.updateBlockchain(types.Blockchain{
		ID:     "0001",
		Active: true,
	})

	c.Len(cache.GetBlockchains(), 2)
	c.Equal(true, cache.GetBlockchain("0001").Active)

	cache.updateBlockchain(types.Blockchain{
		ID:      "0001",
		Ticker:  "CHAIN2",
		Network: "the_network",
	})

	c.Len(cache.GetBlockchains(), 2)
	c.Equal("CHAIN2", cache.GetBlockchain("0001").Ticker)
	c.Equal("the_network", cache.GetBlockchain("0001").Network)
}

func TestCache_AddRedirect(t *testing.T) {
	c := require.New(t)

	cache := NewTestCache(t)

	c.Len(cache.GetBlockchains(), 2)
	c.Len(cache.GetBlockchain("0001").Redirects, 2)

	cache.addRedirect(types.Redirect{BlockchainID: "0001", Alias: "pokt-mainnet-3"})

	c.Len(cache.GetBlockchain("0001").Redirects, 3)
}

func TestCache_UpsertUserAccess(t *testing.T) {
	c := require.New(t)

	cache := NewTestCache(t)

	c.Len(cache.GetLoadBalancers(), 2)

	cache.addUserAccess(types.UserAccess{
		ID:       "test_lb_4nf37fhfy27dj28d",
		UserID:   "",
		RoleName: types.RoleMember,
		Email:    "test@test.com",
		Accepted: false,
	})

	c.Len(cache.GetLoadBalancer("test_lb_4nf37fhfy27dj28d").Users, 3)
	c.Len(cache.GetPendingLoadBalancersByEmail("test@test.com"), 1)
	c.Empty(cache.GetUserPermissionsByUserID("test_user_access_1234").LoadBalancers)

	cache.addUserAccess(types.UserAccess{
		ID:       "test_lb_4nf37fhfy27dj28d",
		UserID:   "test_user_access_1234",
		RoleName: types.RoleMember,
		Email:    "test@test.com",
		Accepted: true,
	})

	c.Len(cache.GetLoadBalancer("test_lb_4nf37fhfy27dj28d").Users, 3)
	c.Len(cache.GetLoadBalancersByUserID("test_user_access_1234"), 1)
	c.Empty(cache.GetPendingLoadBalancersByEmail("test@test.com"))
	c.Len(cache.GetUserPermissionsByUserID("test_user_access_1234").LoadBalancers, 1)
	c.Equal(types.RoleMember, cache.GetUserPermissionsByUserID("test_user_access_1234").LoadBalancers["test_lb_4nf37fhfy27dj28d"].RoleName)

	cache.addUserAccess(types.UserAccess{
		ID:       "test_lb_4nf37fhfy27dj28d",
		UserID:   "test_user_access_1234",
		RoleName: types.RoleAdmin,
		Email:    "test@test.com",
		Accepted: true,
	})

	c.Equal(types.RoleAdmin, cache.GetUserPermissionsByUserID("test_user_access_1234").LoadBalancers["test_lb_4nf37fhfy27dj28d"].RoleName)
}

func TestCache_DeleteUserAccess(t *testing.T) {
	c := require.New(t)

	cache := NewTestCache(t)

	c.Len(cache.GetLoadBalancers(), 2)

	cache.addUserAccess(types.UserAccess{
		ID:       "test_lb_4nf37fhfy27dj28d",
		UserID:   "test_user_access_1234",
		RoleName: types.RoleMember,
		Email:    "test@test.com",
		Accepted: true,
	})

	c.Len(cache.GetLoadBalancer("test_lb_4nf37fhfy27dj28d").Users, 3)
	c.Len(cache.GetLoadBalancersByUserID("test_user_access_1234"), 1)
	c.Len(cache.GetUserPermissionsByUserID("test_user_access_1234").LoadBalancers, 1)
	c.Equal(types.RoleMember, cache.GetUserPermissionsByUserID("test_user_access_1234").LoadBalancers["test_lb_4nf37fhfy27dj28d"].RoleName)

	cache.deleteUserAccess(types.UserAccess{
		ID:       "test_lb_4nf37fhfy27dj28d",
		UserID:   "test_user_access_1234",
		RoleName: types.RoleAdmin,
		Email:    "test@test.com",
		Accepted: false,
	})

	c.Len(cache.GetLoadBalancer("test_lb_4nf37fhfy27dj28d").Users, 2)
	c.Len(cache.GetLoadBalancersByUserID("test_user_access_1234"), 0)
	c.Len(cache.GetUserPermissionsByUserID("test_user_access_1234").LoadBalancers, 0)
}
