package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pokt-foundation/pocket-http-db/cache"
	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/stretchr/testify/require"
)

func newTestRouter() (*mux.Router, error) {
	readerMock := &cache.ReaderMock{}

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
		{
			ID: "0022",
		},
	}, nil)

	readerMock.On("ReadLoadBalancers").Return([]*repository.LoadBalancer{
		{
			ID: "60ecb2bf67774900350d9c42",
			ApplicationIDs: []string{
				"5f62b7d8be3591c4dea8566d",
				"5f62b7d8be3591c4dea8566a",
			},
		},
		{
			ID: "60ecb2bf67774900350d9c43",
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
		{
			ID: "60ecb2bf67774900350d9c44",
		},
	}, nil)

	router, err := NewRouter(readerMock)
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()
	r.HandleFunc("/", router.HealthCheck)
	r.HandleFunc("/blockchain", router.GetBlockchains)
	r.HandleFunc("/blockchain/{id}", router.GetBlockchain)
	r.HandleFunc("/application", router.GetApplications)
	r.HandleFunc("/application/{id}", router.GetApplication)
	r.HandleFunc("/load_balancer", router.GetLoadBalancers)
	r.HandleFunc("/load_balancer/{id}", router.GetLoadBalancer)
	r.HandleFunc("/user", router.GetUsers)
	r.HandleFunc("/user/{id}", router.GetUser)
	r.HandleFunc("/user/{id}/application", router.GetApplicationByUserID)

	return r, nil
}

func TestRouter_HealthCheck(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)
}

func TestRouter_GetApplications(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest("GET", "/application", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*repository.Application{
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
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func TestRouter_GetApplication(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest("GET", "/application/5f62b7d8be3591c4dea8566d", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal(&repository.Application{
		ID:     "5f62b7d8be3591c4dea8566d",
		UserID: "60ecb2bf67774900350d9c43",
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func TestRouter_GetApplicationsByUserID(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest("GET", "/user/60ecb2bf67774900350d9c43/application", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*repository.Application{
		{
			ID:     "5f62b7d8be3591c4dea8566d",
			UserID: "60ecb2bf67774900350d9c43",
		},
		{
			ID:     "5f62b7d8be3591c4dea8566a",
			UserID: "60ecb2bf67774900350d9c43",
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func TestRouter_GetBlockchains(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest("GET", "/blockchain", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*repository.Blockchain{
		{
			ID: "0021",
		},
		{
			ID: "0022",
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func TestRouter_GetBlockchain(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest("GET", "/blockchain/0021", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal(&repository.Blockchain{
		ID: "0021",
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func TestRouter_GetLoadBalancers(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest("GET", "/load_balancer", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	var marshaledBody []*repository.LoadBalancer

	err = json.Unmarshal(rr.Body.Bytes(), &marshaledBody)
	c.NoError(err)

	c.Len(marshaledBody, 2)
	c.Equal("60ecb2bf67774900350d9c42", marshaledBody[0].ID)
	c.Equal("60ecb2bf67774900350d9c43", marshaledBody[1].ID)
}

func TestRouter_GetLoadBalancer(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest("GET", "/load_balancer/60ecb2bf67774900350d9c42", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	var marshaledBody repository.LoadBalancer

	err = json.Unmarshal(rr.Body.Bytes(), &marshaledBody)
	c.NoError(err)

	c.Equal("60ecb2bf67774900350d9c42", marshaledBody.ID)
}

func TestRouter_GetUsers(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest("GET", "/user", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal([]*repository.User{
		{
			ID: "60ecb2bf67774900350d9c43",
		},
		{
			ID: "60ecb2bf67774900350d9c44",
		},
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}

func TestRouter_GetUser(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest("GET", "/user/60ecb2bf67774900350d9c43", nil)
	c.NoError(err)

	rr := httptest.NewRecorder()

	router, err := newTestRouter()
	c.NoError(err)

	router.ServeHTTP(rr, req)

	c.Equal(http.StatusOK, rr.Code)

	expectedBody, err := json.Marshal(&repository.User{
		ID: "60ecb2bf67774900350d9c43",
	})
	c.NoError(err)

	c.Equal(expectedBody, rr.Body.Bytes())
}
