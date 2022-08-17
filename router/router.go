package router

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pokt-foundation/pocket-http-db/cache"
	"github.com/pokt-foundation/pocket-http-db/environment"
	"github.com/pokt-foundation/portal-api-go/repository"
)

var (
	apiKeys = environment.GetStringMap("API_KEYS", "", ",")
)

// Writer represents the implementation of writer interface
type Writer interface {
	WriteLoadBalancer(loadBalancer *repository.LoadBalancer) (*repository.LoadBalancer, error)
	UpdateLoadBalancer(id string, options *repository.UpdateLoadBalancer) error
	RemoveLoadBalancer(id string) error
	WriteApplication(app *repository.Application) (*repository.Application, error)
	UpdateApplication(id string, options *repository.UpdateApplication) error
	RemoveApplication(id string) error
}

// Router struct handler for router requests
type Router struct {
	Cache  *cache.Cache
	Router *mux.Router
	Writer Writer
}

// NewRouter returns router instance
func NewRouter(reader cache.Reader, writer Writer) (*Router, error) {
	cache := cache.NewCache(reader)

	err := cache.SetCache()
	if err != nil {
		return nil, err
	}

	rt := &Router{
		Cache:  cache,
		Writer: writer,
		Router: mux.NewRouter(),
	}

	rt.Router.HandleFunc("/", rt.HealthCheck).Methods(http.MethodGet)
	rt.Router.HandleFunc("/blockchain", rt.GetBlockchains).Methods(http.MethodGet)
	rt.Router.HandleFunc("/blockchain/{id}", rt.GetBlockchain).Methods(http.MethodGet)
	rt.Router.HandleFunc("/application", rt.GetApplications).Methods(http.MethodGet)
	rt.Router.HandleFunc("/application", rt.CreateApplication).Methods(http.MethodPost)
	rt.Router.HandleFunc("/application/limits", rt.GetApplicationsLimits).Methods(http.MethodGet)
	rt.Router.HandleFunc("/application/{id}", rt.GetApplication).Methods(http.MethodGet)
	rt.Router.HandleFunc("/application/{id}", rt.UpdateApplication).Methods(http.MethodPut)
	rt.Router.HandleFunc("/load_balancer", rt.GetLoadBalancers).Methods(http.MethodGet)
	rt.Router.HandleFunc("/load_balancer", rt.CreateLoadBalancer).Methods(http.MethodPost)
	rt.Router.HandleFunc("/load_balancer/{id}", rt.GetLoadBalancer).Methods(http.MethodGet)
	rt.Router.HandleFunc("/load_balancer/{id}", rt.UpdateLoadBalancer).Methods(http.MethodPut)
	rt.Router.HandleFunc("/user", rt.GetUsers).Methods(http.MethodGet)
	rt.Router.HandleFunc("/user/{id}", rt.GetUser).Methods(http.MethodGet)
	rt.Router.HandleFunc("/user/{id}/application", rt.GetApplicationByUserID).Methods(http.MethodGet)
	rt.Router.HandleFunc("/user/{id}/load_balancer", rt.GetLoadBalancerByUserID).Methods(http.MethodGet)

	rt.Router.Use(rt.AuthorizationHandler)

	return rt, nil
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

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
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

func (rt *Router) GetApplicationsLimits(w http.ResponseWriter, r *http.Request) {
	apps := rt.Cache.GetApplications()

	var appsLimits []repository.AppLimits

	for _, app := range apps {
		appsLimits = append(appsLimits, repository.AppLimits{
			AppID:      app.ID,
			PublicKey:  app.FreeTierApplicationAccount.PublicKey,
			DailyLimit: app.Limits.DailyLimit,
		})
	}

	respondWithJSON(w, http.StatusOK, appsLimits)
}

func (rt *Router) GetApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	app := rt.Cache.GetApplication(vars["id"])

	if app == nil {
		respondWithError(w, http.StatusNotFound, "application not found")
		return
	}

	respondWithJSON(w, http.StatusOK, app)
}

func (rt *Router) CreateApplication(w http.ResponseWriter, r *http.Request) {
	var app repository.Application

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&app)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	fullApp, err := rt.Writer.WriteApplication(&app)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rt.Cache.AddApplication(fullApp)

	respondWithJSON(w, http.StatusOK, fullApp)
}

func (rt *Router) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	app := rt.Cache.GetApplication(vars["id"])
	if app == nil {
		respondWithError(w, http.StatusNotFound, "application not found")
		return
	}

	var updateInput repository.UpdateApplication

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&updateInput)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	if updateInput.Remove {
		err = rt.Writer.RemoveApplication(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		app.Status = repository.AwaitingGracePeriod
	} else {
		err = rt.Writer.UpdateApplication(vars["id"], &updateInput)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if updateInput.Name != "" {
			app.Name = updateInput.Name
		}
		if updateInput.AppLimits != nil {
			app.Limits = *updateInput.AppLimits
		}
		if updateInput.GatewaySettings != nil {
			app.GatewaySettings = *updateInput.GatewaySettings
		}
		if updateInput.NotificationSettings != nil {
			app.NotificationSettings = *updateInput.NotificationSettings
		}
	}

	rt.Cache.UpdateApplication(app)

	respondWithJSON(w, http.StatusOK, app)
}

func (rt *Router) GetApplicationByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	apps := rt.Cache.GetApplicationsByUserID(vars["id"])

	if len(apps) == 0 {
		respondWithError(w, http.StatusNotFound, "applications not found")
		return
	}

	respondWithJSON(w, http.StatusOK, apps)
}

func (rt *Router) GetLoadBalancerByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	lbs := rt.Cache.GetLoadBalancersByUserID(vars["id"])

	if len(lbs) == 0 {
		respondWithError(w, http.StatusNotFound, "load balancers not found")
		return
	}

	respondWithJSON(w, http.StatusOK, lbs)
}

func (rt *Router) GetBlockchain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	blockchain := rt.Cache.GetBlockchain(vars["id"])

	if blockchain == nil {
		respondWithError(w, http.StatusNotFound, "blockchain not found")
		return
	}

	respondWithJSON(w, http.StatusOK, blockchain)
}

func (rt *Router) GetBlockchains(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, rt.Cache.GetBlockchains())
}

func (rt *Router) GetLoadBalancer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	lb := rt.Cache.GetLoadBalancer(vars["id"])

	if lb == nil {
		respondWithError(w, http.StatusNotFound, "load balancer not found")
		return
	}

	respondWithJSON(w, http.StatusOK, lb)
}

func (rt *Router) CreateLoadBalancer(w http.ResponseWriter, r *http.Request) {
	var lb repository.LoadBalancer

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&lb)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	fullLB, err := rt.Writer.WriteLoadBalancer(&lb)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rt.Cache.AddLoadBalancer(fullLB)

	respondWithJSON(w, http.StatusOK, fullLB)
}

func (rt *Router) UpdateLoadBalancer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	lb := rt.Cache.GetLoadBalancer(vars["id"])
	if lb == nil {
		respondWithError(w, http.StatusNotFound, "load balancer not found")
		return
	}

	var updateInput repository.UpdateLoadBalancer

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&updateInput)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	if updateInput.Remove {
		err = rt.Writer.RemoveLoadBalancer(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		oldUserID := lb.UserID
		lb.UserID = ""

		rt.Cache.DeleteLoadBalancer(lb, oldUserID)
	} else {
		err = rt.Writer.UpdateLoadBalancer(vars["id"], &updateInput)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if updateInput.Name != "" {
			lb.Name = updateInput.Name
		}

		rt.Cache.UpdateLoadBalancer(lb)
	}

	respondWithJSON(w, http.StatusOK, lb)
}

func (rt *Router) GetLoadBalancers(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, rt.Cache.GetLoadBalancers())
}

func (rt *Router) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	user := rt.Cache.GetUser(vars["id"])

	if user == nil {
		respondWithError(w, http.StatusNotFound, "user not found")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (rt *Router) GetUsers(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, rt.Cache.GetUsers())
}
