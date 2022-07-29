package router

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pokt-foundation/pocket-http-db/cache"
)

// Router struct handler for router requests
type Router struct {
	cache *cache.Cache
}

// NewRouter returns router instance
func NewRouter(reader cache.Reader) (*Router, error) {
	cache := cache.NewCache(reader)

	err := cache.SetCache()
	if err != nil {
		return nil, err
	}

	return &Router{
		cache: cache,
	}, nil
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		panic(err)
	}
}

func (rt *Router) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Pocket HTTP DB is up and running!"))
	if err != nil {
		panic(err)
	}
}

func (rt *Router) GetApplications(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, rt.cache.GetApplications())
}

func (rt *Router) GetApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	respondWithJSON(w, http.StatusOK, rt.cache.GetApplication(vars["id"]))
}

func (rt *Router) GetApplicationByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	respondWithJSON(w, http.StatusOK, rt.cache.GetApplicationsByUserID(vars["id"]))
}

func (rt *Router) GetBlockchain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	respondWithJSON(w, http.StatusOK, rt.cache.GetBlockchain(vars["id"]))
}

func (rt *Router) GetBlockchains(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, rt.cache.GetBlockchains())
}

func (rt *Router) GetLoadBalancer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	respondWithJSON(w, http.StatusOK, rt.cache.GetLoadBalancer(vars["id"]))
}

func (rt *Router) GetLoadBalancers(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, rt.cache.GetLoadBalancers())
}

func (rt *Router) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	respondWithJSON(w, http.StatusOK, rt.cache.GetUser(vars["id"]))
}

func (rt *Router) GetUsers(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, rt.cache.GetUsers())
}