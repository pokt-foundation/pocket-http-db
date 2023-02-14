package v1_cache

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
		{ID: "0021"},
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

	cache := NewCache(readerMock, logrus.New())

	err := cache.SetCache()
	c.NoError(err)

	c.NotEmpty(cache.GetApplication("test_app_e3591c4dea8566d"))
	c.Len(cache.GetApplications(), 3)
	c.Len(cache.GetApplicationsByUserID("test_user_774900350d9c43"), 2)

	c.NotEmpty(cache.GetBlockchain("0021"))
	c.Len(cache.GetBlockchains(), 1)

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
			ID:     "test_app_e3591c4dea8566d",
			UserID: "test_user_774900350d9c43",
		},
	}, nil)

	errOnLoadBalancer := errors.New("error on loadbalancers")
	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{}, errOnLoadBalancer).Once()

	err = cache.SetCache()
	c.ErrorIs(err, errOnLoadBalancer)

	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{
		{
			ID: "test_lb_67774900350d9c42",
		},
	}, nil)

	errOnUserPermissions := errors.New("error on user permissions")
	readerMock.On("ReadUserPermissions", testCtx).Return(map[types.UserID]*types.UserPermissions{}, errOnUserPermissions).Once()

	err = cache.SetCache()
	c.ErrorIs(err, errOnUserPermissions)
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
			ID:     "test_app_e3591c4dea8566d",
			UserID: "test_user_774900350d9c43",
		},
		{
			ID:     "test_app_e3591c4dea8566a",
			UserID: "test_user_774900350d9c43",
		},
		{
			ID:     "test_app_e3591c4dea8566f",
			UserID: "test_user_774900350d9c44",
		},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setPayPlans()
	c.NoError(err)

	err = cache.setApplications()
	c.NoError(err)

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
			ID:     "test_app_e3591c4dea8566d",
			UserID: "test_user_774900350d9c43",
		},
		{
			ID:     "test_app_e3591c4dea8566a",
			UserID: "test_user_774900350d9c43",
		},
		{
			ID:     "test_app_e3591c4dea8566f",
			UserID: "test_user_774900350d9c44",
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

	readerMock := NewReaderMock(t)

	readerMock.On("ReadPayPlans", testCtx).Return([]*types.PayPlan{
		{
			Type:  types.FreetierV0,
			Limit: 250000,
		},
	}, nil)

	readerMock.On("ReadApplications", testCtx).Return([]*types.Application{
		{
			ID:     "test_app_e3591c4dea8566d",
			UserID: "test_user_774900350d9c43",
		},
		{
			ID:     "test_app_e3591c4dea8566a",
			UserID: "test_user_774900350d9c43",
		},
	}, nil)

	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{
		{
			ID:     "test_lb_67774900350d9c42",
			UserID: "test_user_774900350d9c43",
			ApplicationIDs: []string{
				"test_app_e3591c4dea8566d",
				"test_app_e3591c4dea8566a",
			},
			Users: []types.UserAccess{
				{UserID: "test_user_774900350d9c43", RoleName: types.RoleOwner},
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
	c.Len(cache.GetApplicationsByUserID("test_user_774900350d9c43"), 2)

	cache.updateApplication(types.Application{
		ID:     "test_app_e3591c4dea8566a",
		UserID: "",
	})
	c.Len(cache.GetApplicationsByUserID("test_user_774900350d9c43"), 1)
}

func TestCache_AddLoadBalancer(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)

	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{
		{
			ID:     "test_lb_476fhdf72jdh2dwhk",
			UserID: "test_user_774900350d9c43",
			Users: []types.UserAccess{
				{UserID: "test_user_774900350d9c43", RoleName: types.RoleOwner},
			},
		},
		{
			ID:     "test_lb_4nf37fhfy27dj28d",
			UserID: "test_user_774900350d9c43",
			Users: []types.UserAccess{
				{UserID: "test_user_774900350d9c43", RoleName: types.RoleOwner},
			},
		},
		{
			ID:     "test_lb_7s283jkd8dk2ufj9",
			UserID: "test_user_774900350d9c44",
			Users: []types.UserAccess{
				{UserID: "test_user_774900350d9c44", RoleName: types.RoleOwner},
			},
		},
	}, nil)

	readerMock.On("ReadApplications", testCtx).Return([]*types.Application{
		{
			ID:     "test_app_e3591c4dea8566d",
			UserID: "test_user_774900350d9c43",
		},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setLoadBalancers()
	c.NoError(err)

	err = cache.setApplications()
	c.NoError(err)

	cache.addLoadBalancer(types.LoadBalancer{
		ID:     "test_lb_47cjd72kd82kd98d",
		UserID: "test_user_774900350d9c43",
		Users: []types.UserAccess{
			{UserID: "test_user_774900350d9c43", RoleName: types.RoleOwner},
			{UserID: "test_user_774900350d9c44", RoleName: types.RoleAdmin},
		},
	})

	c.Len(cache.GetLoadBalancers(), 4)
	c.Len(cache.GetLoadBalancersByUserID("test_user_774900350d9c43"), 3)
	c.Len(cache.GetLoadBalancersByUserID("test_user_774900350d9c44"), 2)
}

func TestCache_UpdateLoadBalancer(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)

	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{
		{
			ID:     "test_lb_476fhdf72jdh2dwhk",
			UserID: "test_user_774900350d9c43",
			Users: []types.UserAccess{
				{UserID: "test_user_774900350d9c43", RoleName: types.RoleOwner},
			},
		},
		{
			ID:     "test_lb_4nf37fhfy27dj28d",
			UserID: "test_user_774900350d9c43",
			Users: []types.UserAccess{
				{UserID: "test_user_774900350d9c43", RoleName: types.RoleOwner},
			},
		},
		{
			ID:     "test_lb_7s283jkd8dk2ufj9",
			UserID: "test_user_774900350d9c44",
			Users: []types.UserAccess{
				{UserID: "test_user_774900350d9c44", RoleName: types.RoleOwner},
			},
		},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setLoadBalancers()
	c.NoError(err)

	cache.updateLoadBalancer(types.LoadBalancer{
		ID:     "test_lb_4nf37fhfy27dj28d",
		UserID: "test_user_774900350d9c43",
		Name:   "papolo",
	})

	c.Len(cache.GetLoadBalancers(), 3)
	c.Len(cache.GetLoadBalancersByUserID("test_user_774900350d9c43"), 2)
	c.Len(cache.GetLoadBalancersByUserID("test_user_774900350d9c44"), 1)
	c.Equal("papolo", cache.GetLoadBalancersByUserID("test_user_774900350d9c43")[1].Name)
	c.Equal("papolo", cache.GetLoadBalancer("test_lb_4nf37fhfy27dj28d").Name)
}

func TestCache_RemoveLoadBalancer(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadLoadBalancers", testCtx).Return([]*types.LoadBalancer{
		{
			ID:     "test_lb_476fhdf72jdh2dwhk",
			UserID: "test_user_774900350d9c43",
			Users: []types.UserAccess{
				{UserID: "test_user_774900350d9c43", RoleName: types.RoleOwner},
			},
		},
		{
			ID:     "test_lb_4nf37fhfy27dj28d",
			UserID: "test_user_774900350d9c43",
			Users: []types.UserAccess{
				{UserID: "test_user_774900350d9c43", RoleName: types.RoleOwner},
			},
		},
	}, nil)

	cache := NewCache(readerMock, logrus.New())

	err := cache.setLoadBalancers()
	c.NoError(err)
	c.Len(cache.GetLoadBalancersByUserID("test_user_774900350d9c43"), 2)

	cache.updateLoadBalancer(types.LoadBalancer{
		ID:     "test_lb_4nf37fhfy27dj28d",
		UserID: "",
	})

	c.Len(cache.GetLoadBalancersByUserID("test_user_774900350d9c43"), 1)
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

func TestCache_AddUserAccess(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)

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

	cache := NewCache(readerMock, logrus.New())

	err := cache.setLoadBalancers()
	c.NoError(err)
	err = cache.setUserPermissions()
	c.NoError(err)

	c.Len(cache.GetLoadBalancers(), 2)

	cache.addUserAccess(types.UserAccess{
		ID:       "test_lb_4nf37fhfy27dj28d",
		UserID:   "test_user_access_1234",
		RoleName: types.RoleMember,
		Email:    "test@test.com",
		Accepted: false,
	})

	c.Len(cache.GetLoadBalancer("test_lb_4nf37fhfy27dj28d").Users, 3)
	c.Len(cache.GetLoadBalancersByUserID("test_user_access_1234"), 1)
	c.Len(cache.GetUserPermissionsByUserID("test_user_access_1234").LoadBalancers, 1)
	c.Equal(types.RoleMember, cache.GetUserPermissionsByUserID("test_user_access_1234").LoadBalancers["test_lb_4nf37fhfy27dj28d"].RoleName)
}

func TestCache_UpdateUserAccess(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)

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

	cache := NewCache(readerMock, logrus.New())

	err := cache.setLoadBalancers()
	c.NoError(err)
	err = cache.setUserPermissions()
	c.NoError(err)

	c.Len(cache.GetLoadBalancers(), 2)

	cache.addUserAccess(types.UserAccess{
		ID:       "test_lb_4nf37fhfy27dj28d",
		UserID:   "test_user_access_1234",
		RoleName: types.RoleMember,
		Email:    "test@test.com",
		Accepted: false,
	})

	c.Len(cache.GetLoadBalancer("test_lb_4nf37fhfy27dj28d").Users, 3)
	c.Len(cache.GetLoadBalancersByUserID("test_user_access_1234"), 1)
	c.Len(cache.GetUserPermissionsByUserID("test_user_access_1234").LoadBalancers, 1)
	c.Equal(types.RoleMember, cache.GetUserPermissionsByUserID("test_user_access_1234").LoadBalancers["test_lb_4nf37fhfy27dj28d"].RoleName)

	cache.updateUserAccess(types.UserAccess{
		ID:       "test_lb_4nf37fhfy27dj28d",
		UserID:   "test_user_access_1234",
		RoleName: types.RoleAdmin,
		Email:    "test@test.com",
		Accepted: false,
	})

	c.Len(cache.GetLoadBalancer("test_lb_4nf37fhfy27dj28d").Users, 3)
	c.Len(cache.GetLoadBalancersByUserID("test_user_access_1234"), 1)
	c.Len(cache.GetUserPermissionsByUserID("test_user_access_1234").LoadBalancers, 1)
	c.Equal(types.RoleAdmin, cache.GetUserPermissionsByUserID("test_user_access_1234").LoadBalancers["test_lb_4nf37fhfy27dj28d"].RoleName)
}
func TestCache_DeleteUserAccess(t *testing.T) {
	c := require.New(t)

	readerMock := NewReaderMock(t)

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

	cache := NewCache(readerMock, logrus.New())

	err := cache.setLoadBalancers()
	c.NoError(err)
	err = cache.setUserPermissions()
	c.NoError(err)

	c.Len(cache.GetLoadBalancers(), 2)

	cache.addUserAccess(types.UserAccess{
		ID:       "test_lb_4nf37fhfy27dj28d",
		UserID:   "test_user_access_1234",
		RoleName: types.RoleMember,
		Email:    "test@test.com",
		Accepted: false,
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
