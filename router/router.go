package router

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pokt-foundation/pocket-http-db/cache"
	"github.com/pokt-foundation/pocket-http-db/environment"
)

var (
	apiKeys = environment.GetStringMap("API_KEYS", "", ",")
)

// Router struct handler for router requests
type Router struct {
	Cache *cache.Cache
}

// NewRouter returns router instance
func NewRouter(reader cache.Reader) (*Router, error) {
	cache := cache.NewCache(reader)

	err := cache.SetCache()
	if err != nil {
		return nil, err
	}

	return &Router{
		Cache: cache,
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

func (rt *Router) AuthorizationHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is the path of the health check endpoint
		if r.URL.Path == "/" {
			h.ServeHTTP(w, r)

			return
		}

		if !apiKeys[r.Header.Get("Authorization")] {
			w.WriteHeader(http.StatusUnauthorized)
			_, err := w.Write([]byte("Unauthorized"))
			if err != nil {
				panic(err)
			}

			return
		}

		h.ServeHTTP(w, r)
	})
}

func (rt *Router) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Pocket HTTP DB is up and running!"))
	if err != nil {
		panic(err)
	}
}

func (rt *Router) GetApplications(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, rt.Cache.GetApplications())
}

func (rt *Router) GetApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	respondWithJSON(w, http.StatusOK, rt.Cache.GetApplication(vars["id"]))
}

func (rt *Router) GetApplicationByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	respondWithJSON(w, http.StatusOK, rt.Cache.GetApplicationsByUserID(vars["id"]))
}

func (rt *Router) GetLoadBalancerByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	respondWithJSON(w, http.StatusOK, rt.Cache.GetLoadBalancersByUserID(vars["id"]))
}

func (rt *Router) GetBlockchain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	respondWithJSON(w, http.StatusOK, rt.Cache.GetBlockchain(vars["id"]))
}

func (rt *Router) GetBlockchains(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, rt.Cache.GetBlockchains())
}

func (rt *Router) GetLoadBalancer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	respondWithJSON(w, http.StatusOK, rt.Cache.GetLoadBalancer(vars["id"]))
}

func (rt *Router) GetLoadBalancers(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, rt.Cache.GetLoadBalancers())
}

func (rt *Router) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	respondWithJSON(w, http.StatusOK, rt.Cache.GetUser(vars["id"]))
}

func (rt *Router) GetUsers(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, rt.Cache.GetUsers())
}
