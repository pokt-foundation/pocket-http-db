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

	readerMock.On("ReadBlockchains").Return([]*repository.Blockchain{
		{
			ID: "0021",
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

	readerMock.On("ReadUsers").Return([]*repository.User{
		{
			ID: "60ecb2bf67774900350d9c43",
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

	c.NotEmpty(cache.GetUser("60ecb2bf67774900350d9c43"))
	c.Len(cache.GetUsers(), 1)
}

func TestCache_SetCacheFailure(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

	readerMock.On("ReadApplications").Return([]*repository.Application{}, errors.New("error on applications")).Once()

	cache := NewCache(readerMock)

	err := cache.SetCache()
	c.EqualError(err, "error on applications")

	readerMock.On("ReadApplications").Return([]*repository.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
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

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{}, errors.New("error on loadbalancers")).Once()

	err = cache.SetCache()
	c.EqualError(err, "error on loadbalancers")

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{
		{
			ID: "60ecb2bf67774900350d9c42",
		},
	}, nil)

	readerMock.On("ReadUsers").Return([]*repository.User{}, errors.New("error on users")).Once()

	err = cache.SetCache()
	c.EqualError(err, "error on users")
}

func TestCache_AddApplication(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

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

	err := cache.setApplications()
	c.NoError(err)

	cache.AddApplication(&repository.Application{
		ID:     "5f62b7d8be3591c4dea8566b",
		UserID: "60ecb2bf67774900350d9c43",
	})

	c.Len(cache.GetApplications(), 4)
	c.Len(cache.GetApplicationsByUserID("60ecb2bf67774900350d9c43"), 3)
}

func TestCache_UpdateApplication(t *testing.T) {
	c := require.New(t)

	readerMock := &ReaderMock{}

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

	err := cache.setApplications()
	c.NoError(err)

	err = cache.setLoadBalancers()
	c.NoError(err)

	cache.UpdateApplication(&repository.Application{
		ID:     "5f62b7d8be3591c4dea8566a",
		UserID: "60ecb2bf67774900350d9c44",
		Name:   "papolo",
	}, "60ecb2bf67774900350d9c43")

	c.Len(cache.GetApplications(), 3)
	c.Len(cache.GetApplicationsByUserID("60ecb2bf67774900350d9c43"), 1)
	c.Len(cache.GetApplicationsByUserID("60ecb2bf67774900350d9c44"), 2)
	c.Equal("papolo", cache.GetApplication("5f62b7d8be3591c4dea8566a").Name)
	c.Equal("papolo", cache.GetLoadBalancer("60ecb2bf67774900350d9c42").Applications[1].Name)
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

	cache := NewCache(readerMock)

	err := cache.setLoadBalancers()
	c.NoError(err)

	cache.AddLoadBalancer(&repository.LoadBalancer{
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

	cache.UpdateLoadBalancer(&repository.LoadBalancer{
		ID:     "5f62b7d8be3591c4dea8566a",
		UserID: "60ecb2bf67774900350d9c44",
		Name:   "papolo",
	}, "60ecb2bf67774900350d9c43")

	c.Len(cache.GetLoadBalancers(), 3)
	c.Len(cache.GetLoadBalancersByUserID("60ecb2bf67774900350d9c43"), 1)
	c.Len(cache.GetLoadBalancersByUserID("60ecb2bf67774900350d9c44"), 2)
	c.Equal("papolo", cache.GetLoadBalancer("5f62b7d8be3591c4dea8566a").Name)
}
