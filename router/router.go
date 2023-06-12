package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pokt-foundation/pocket-http-db/cache"
	v1Types "github.com/pokt-foundation/portal-db/types"
	"github.com/pokt-foundation/portal-db/v2/driver"
	"github.com/pokt-foundation/portal-db/v2/types"
	jsonresponse "github.com/pokt-foundation/utils-go/json-response"
	"github.com/sirupsen/logrus"
)

var (
	errNoChainsFound     error = errors.New("no chains were found")
	errChainNotFound     error = errors.New("chain not found")
	errNoPortalAppsFound error = errors.New("no portal apps were found")
	errPortalAppNotFound error = errors.New("portal app not found")

	errBlockchainNotFound    error = errors.New("blockchain not found")
	errInvalidRoleNameFilter error = errors.New("invalid role name filter")
	errPermissionsNotFound   error = errors.New("no permissions found for given user ID")
)

// router struct handler for router requests
type (
	Router struct {
		Router  *mux.Router
		writer  driver.Writer
		cache   *cache.Cache
		apiKeys map[string]bool
		log     *logrus.Logger
	}
	Config struct {
		Cache   *cache.Cache
		Writer  driver.Writer
		APIKeys map[string]bool
		Log     *logrus.Logger
	}
)

func (rt *Router) logError(err error) {
	fields := logrus.Fields{
		"err": err.Error(),
	}

	rt.log.WithFields(fields).Error(err)
}

// NewRouter returns router instance
func NewRouter(config Config) (*Router, error) {
	rt := &Router{
		Router:  mux.NewRouter(),
		writer:  config.Writer,
		cache:   config.Cache,
		apiKeys: config.APIKeys,
		log:     config.Log,
	}

	rt.Router.HandleFunc("/", rt.healthCheck).Methods(http.MethodGet)
	rt.Router.HandleFunc("/reset_cache", rt.resetCache).Methods(http.MethodPost)

	/* ----- V2 Endpoints ----- */
	// V2 endpoints are for the new DB schema and access the cache and V2 structs directly.

	rt.Router.HandleFunc("/v2/chain/{id}", rt.getChainByID).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/chain", rt.getAllChains).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/chain", rt.createNewChainAndGigastakeApps).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v2/chain/gigastake", rt.createNewGigastakeApp).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v2/chain/gigastake", rt.updateGigastakeApp).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v2/chain", rt.updateChain).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v2/chain/{id}/activate", rt.activateChain).Methods(http.MethodPut)

	rt.Router.HandleFunc("/v2/portal_app/{id}", rt.getPortalAppByID).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/portal_app", rt.getAllPortalApps).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/portal_app", rt.createNewPortalApp).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v2/portal_app/{id}", rt.updatePortalApp).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v2/portal_app/{id}", rt.deletePortalApp).Methods(http.MethodDelete)

	rt.Router.HandleFunc("/v2/user/{userID}/portal_app", rt.getPortalAppsByUser).Methods(http.MethodGet)

	/* ----- V1 Endpoints ----- */
	// V1 endpoints are for the legacy DB schema and access the cache through the legacy adapters.

	rt.Router.HandleFunc("/v1/blockchain", rt.getBlockchains).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/blockchain/{id}", rt.getBlockchain).Methods(http.MethodGet)

	rt.Router.HandleFunc("/v1/application", rt.getApplications).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/application/first_date_surpassed", rt.updateFirstDateSurpassed).Methods(http.MethodPost)

	rt.Router.HandleFunc("/v1/load_balancer", rt.getLoadBalancers).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/load_balancer", rt.createLoadBalancer).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v1/load_balancer/{id}", rt.getLoadBalancer).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/load_balancer/{id}", rt.updateLoadBalancer).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v1/load_balancer/{id}/user", rt.createLoadBalancerUser).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v1/load_balancer/{id}/integration", rt.createLoadBalancerIntegration).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v1/load_balancer/{id}/user", rt.updateLoadBalancerUserRole).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v1/load_balancer/{id}/user/accept", rt.acceptLoadBalancerUser).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v1/load_balancer/{id}/user/{userID}", rt.deleteLoadBalancerUser).Methods(http.MethodDelete)

	rt.Router.HandleFunc("/v1/user", rt.createPortalUser).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v1/user/{id}/portal_id", rt.getPortalUserID).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/user/{id}/application", rt.getApplicationsByUserID).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/user/{id}/load_balancer", rt.getLoadBalancersByUserID).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/user/{id}/permission", rt.getUserPermissionsByUserID).Methods(http.MethodGet)

	rt.Router.HandleFunc("/v1/user/{portalID}/load_balancer/pending", rt.getPendingLoadBalancersByPortalID).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/user/{portalID}/load_balancer/count", rt.getLoadBalancersCountByPortalID).Methods(http.MethodGet)

	rt.Router.HandleFunc("/v1/pay_plan", rt.getPayPlans).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/pay_plan/{type}", rt.getPayPlan).Methods(http.MethodGet)

	rt.Router.Use(rt.authorizationHandler)

	return rt, nil
}

/* ---------- Middleware and Health Checks ---------- */

func (rt *Router) authorizationHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is the path of the health check endpoint
		if r.URL.Path == "/" {
			h.ServeHTTP(w, r)

			return
		}

		if !rt.apiKeys[r.Header.Get("Authorization")] {
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

func (rt *Router) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Pocket HTTP DB is up and running!"))
	if err != nil {
		panic(err)
	}
}

func (rt *Router) resetCache(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	err := rt.cache.ResetCache(r.Context())
	elapsedTime := time.Since(startTime).Milliseconds()
	if err != nil {
		rt.logError(fmt.Errorf("resetCache failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := fmt.Sprintf("cache manually reset successfully in %dms", elapsedTime)
	_, err = w.Write([]byte(resp))
	if err != nil {
		rt.logError(fmt.Errorf("resetCache failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

/* ---------- V2 Chain Handler Methods ---------- */

// GET /v2/chain/{id} - Fetches a single blockchains by its RelayChainID.
func (rt *Router) getChainByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chainID := types.RelayChainID(vars["id"])

	chain := rt.cache.GetChain(chainID, cache.Options{})
	if chain.ID == "" {
		rt.logError(fmt.Errorf("getChainByID failed: %w", errChainNotFound))
		jsonresponse.RespondWithError(w, http.StatusNotFound, errChainNotFound.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, chain)
}

// GET /v2/chain - Fetches all blockchains.
func (rt *Router) getAllChains(w http.ResponseWriter, r *http.Request) {
	chains := rt.cache.GetChains(cache.Options{})
	if len(chains) == 0 {
		rt.logError(fmt.Errorf("getAllChains failed: %w", errNoChainsFound))
		jsonresponse.RespondWithError(w, http.StatusNotFound, errNoChainsFound.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, chains)
}

// POST /v2/chain - Creates a new blockchains and its Gigastake apps.
func (rt *Router) createNewChainAndGigastakeApps(w http.ResponseWriter, r *http.Request) {
	var newChainInput types.NewChainInput

	err := json.NewDecoder(r.Body).Decode(&newChainInput)
	if err != nil {
		rt.logError(fmt.Errorf("createNewChainAndGigastakeApps Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	createdChainAndGigastakeApps, err := rt.writer.WriteChainAndGigastakeApps(r.Context(), newChainInput, time.Now())
	if err != nil {
		rt.logError(fmt.Errorf("WriteChainAndGigastakeApps failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, createdChainAndGigastakeApps)
}

// POST /v2/chain/gigastake - Creates a new Gigastake app.
func (rt *Router) createNewGigastakeApp(w http.ResponseWriter, r *http.Request) {
	var gigastakeAppInput types.GigastakeApp

	err := json.NewDecoder(r.Body).Decode(&gigastakeAppInput)
	if err != nil {
		rt.logError(fmt.Errorf("createNewGigastakeApp Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	createdGigastakeApp, err := rt.writer.WriteGigastakeApp(r.Context(), gigastakeAppInput, time.Now())
	if err != nil {
		rt.logError(fmt.Errorf("WriteGigastakeApp failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, createdGigastakeApp)
}

// PUT /v2/chain - Updates an existing blockchain.
func (rt *Router) updateChain(w http.ResponseWriter, r *http.Request) {
	var chainUpdate types.Chain

	err := json.NewDecoder(r.Body).Decode(&chainUpdate)
	if err != nil {
		rt.logError(fmt.Errorf("updateChain Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	err = rt.writer.UpdateChain(r.Context(), chainUpdate, time.Now())
	if err != nil {
		rt.logError(fmt.Errorf("UpdateChain failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, chainUpdate)
}

// PUT /v2/chain/gigastake - Updates a Gigastake app.
func (rt *Router) updateGigastakeApp(w http.ResponseWriter, r *http.Request) {
	var gigastakeAppUpdate types.UpdateGigastakeApp

	err := json.NewDecoder(r.Body).Decode(&gigastakeAppUpdate)
	if err != nil {
		rt.logError(fmt.Errorf("updateGigastakeApp Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	err = rt.writer.UpdateGigastakeApp(r.Context(), gigastakeAppUpdate, time.Now())
	if err != nil {
		rt.logError(fmt.Errorf("UpdateGigastakeApp failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, gigastakeAppUpdate)
}

// PUT /v2/chain/{id}/activate - Activates or deactivates a blockchain by ID.
func (rt *Router) activateChain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chainID := types.RelayChainID(vars["id"])

	var active bool

	err := json.NewDecoder(r.Body).Decode(&active)
	if err != nil {
		rt.logError(fmt.Errorf("activateChain decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	activated, err := rt.writer.SetChainActiveStatus(r.Context(), chainID, active, time.Now())
	if err != nil {
		rt.logError(fmt.Errorf("SetChainActiveStatus in ActivateChain failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, activated)
}

/* ---------- V2 Portal App Handler Methods ---------- */

// GET /v2/portal_app/{id} - Fetches a single portal application by its PortalAppID.
func (rt *Router) getPortalAppByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portalAppID := types.PortalAppID(vars["id"])

	portalApp := rt.cache.GetPortalApp(portalAppID)
	if portalApp.ID == "" {
		rt.logError(fmt.Errorf("getPortalAppByID failed: %w", errPortalAppNotFound))
		jsonresponse.RespondWithError(w, http.StatusNotFound, errPortalAppNotFound.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, portalApp)
}

// GET /v2/portal_app - Fetches all portal applications.
func (rt *Router) getAllPortalApps(w http.ResponseWriter, r *http.Request) {
	portalApps := rt.cache.GetPortalApps(cache.Options{})
	if len(portalApps) == 0 {
		rt.logError(fmt.Errorf("getAllPortalApps failed: %w", errNoPortalAppsFound))
		jsonresponse.RespondWithError(w, http.StatusNotFound, errNoPortalAppsFound.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, portalApps)
}

// GET /v2/user/{userID}/portal_app - Fetches all portal applications.
func (rt *Router) getPortalAppsByUser(w http.ResponseWriter, r *http.Request) {
	roleNameFilter := types.RoleName(r.URL.Query().Get("filter"))
	if roleNameFilter != types.RoleName("") && !roleNameFilter.IsValid() {
		jsonresponse.RespondWithError(w, http.StatusBadRequest, errInvalidRoleNameFilter.Error())
		return
	}

	vars := mux.Vars(r)
	userID := types.UserID(vars["userID"])

	userAccounts := rt.cache.GetAccountsByUserID(userID)
	Plog("userAccounts", userAccounts)

	filteredPortalApps := []types.PortalApp{}
	for _, account := range userAccounts {
		if user, ok := account.Users[userID]; ok && user.Accepted {
			for portalAppID, roleName := range user.PortalAppRoles {
				if roleNameFilter == "" || roleName == types.RoleName(roleNameFilter) {
					if portalApp, ok := account.PortalApps[portalAppID]; ok {
						filteredPortalApps = append(filteredPortalApps, *portalApp)
					}
				}
			}
		}
	}
	if len(filteredPortalApps) == 0 {
		rt.logError(fmt.Errorf("getPortalAppsByUser failed: %w", errNoPortalAppsFound))
		jsonresponse.RespondWithError(w, http.StatusNotFound, errNoPortalAppsFound.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, filteredPortalApps)
}

func Plog(args ...interface{}) {
	for _, arg := range args {
		var prettyJSON bytes.Buffer
		jsonArg, _ := json.Marshal(arg)
		str := string(jsonArg)
		_ = json.Indent(&prettyJSON, []byte(str), "", "    ")
		output := prettyJSON.String()

		fmt.Println(output)
	}
}

// POST /v2/portal_app - Creates a new portal application.
func (rt *Router) createNewPortalApp(w http.ResponseWriter, r *http.Request) {
	var portalAppInput types.PortalApp

	err := json.NewDecoder(r.Body).Decode(&portalAppInput)
	if err != nil {
		rt.logError(fmt.Errorf("createNewPortalApp Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	createdPortalApp, err := rt.writer.WritePortalApp(r.Context(), portalAppInput, portalAppInput.AAT(), time.Now())
	if err != nil {
		rt.logError(fmt.Errorf("WritePortalApp failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, createdPortalApp)
}

// PUT /v2/portal_app/{id} - Updates a portal application.
func (rt *Router) updatePortalApp(w http.ResponseWriter, r *http.Request) {
	var portalAppUpdate types.UpdatePortalApp

	err := json.NewDecoder(r.Body).Decode(&portalAppUpdate)
	if err != nil {
		rt.logError(fmt.Errorf("updatePortalApp Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	vars := mux.Vars(r)
	portalAppUpdate.AppID = types.PortalAppID(vars["id"])

	err = rt.writer.UpdatePortalApp(r.Context(), portalAppUpdate, time.Now())
	if err != nil {
		rt.logError(fmt.Errorf("UpdatePortalApp failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, portalAppUpdate)
}

// DELETE /v2/portal_app/{id} - Deletes a portal application.
func (rt *Router) deletePortalApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portalAppID := types.PortalAppID(vars["id"])

	err := rt.writer.SetPortalAppDeleted(r.Context(), portalAppID, time.Now())
	if err != nil {
		rt.logError(fmt.Errorf("SetPortalAppDeleted failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, portalAppID)
}

/* ---------- V1 Read Methods ---------- */

func (rt *Router) getUserPermissionsByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	permissions, err := rt.cache.GetLegacyUserPermissionsByUserID(r.Context(), userID)
	if err != nil {
		rt.logError(fmt.Errorf("getUserPermissionsByUserID failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	if permissions.UserID == "" {
		rt.logError(fmt.Errorf("getUserPermissionsByUserID failed: %w", errPermissionsNotFound))
		jsonresponse.RespondWithError(w, http.StatusNotFound, errPermissionsNotFound.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, permissions)
}

func (rt *Router) getPortalUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	portalUserID := rt.cache.GetUserIDFromProviderUserID(types.ProviderUserID(vars["id"]))

	jsonresponse.RespondWithJSON(w, http.StatusOK, portalUserID)
}

func (rt *Router) getApplicationsByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	apps, err := rt.cache.GetLegacyApplicationsByUserID(r.Context(), vars["id"])
	if err != nil {
		rt.logError(fmt.Errorf("getApplicationsByUserID failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, apps)
}

func (rt *Router) getApplications(w http.ResponseWriter, r *http.Request) {
	applications := rt.cache.GetLegacyApplications(cache.Options{
		IncludeDeleted: false,
	})

	jsonresponse.RespondWithJSON(w, http.StatusOK, applications)
}

func (rt *Router) getLoadBalancersByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	roleNameFilter := v1Types.RoleName(r.URL.Query().Get("filter"))
	if roleNameFilter != v1Types.RoleName("") && !v1Types.ValidRoleNames[roleNameFilter] {
		jsonresponse.RespondWithError(w, http.StatusBadRequest, errInvalidRoleNameFilter.Error())
		return
	}

	lbs, err := rt.cache.GetLegacyLoadBalancersByUserID(r.Context(), userID, roleNameFilter)
	if err != nil {
		rt.logError(fmt.Errorf("getLoadBalancersByUserID failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, lbs)
}

func (rt *Router) getLoadBalancers(w http.ResponseWriter, r *http.Request) {
	jsonresponse.RespondWithJSON(w, http.StatusOK, rt.cache.GetLegacyLoadBalancers(r.Context()))
}

func (rt *Router) getLoadBalancer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	lb, err := rt.cache.GetLegacyLoadBalancer(r.Context(), vars["id"])
	if err != nil {
		rt.logError(fmt.Errorf("getLoadBalancer failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, lb)
}

func (rt *Router) getBlockchains(w http.ResponseWriter, r *http.Request) {
	jsonresponse.RespondWithJSON(w, http.StatusOK, rt.cache.GetLegacyBlockchains(r.Context()))
}

func (rt *Router) getBlockchain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	blockchain := rt.cache.GetLegacyBlockchain(r.Context(), vars["id"])
	if blockchain.ID == "" {
		rt.logError(fmt.Errorf("getBlockchain failed: %w", errBlockchainNotFound))
		jsonresponse.RespondWithError(w, http.StatusNotFound, errBlockchainNotFound.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, blockchain)
}

func (rt *Router) getPayPlans(w http.ResponseWriter, r *http.Request) {
	jsonresponse.RespondWithJSON(w, http.StatusOK, rt.cache.GetLegacyPayPlans(r.Context()))
}

func (rt *Router) getPayPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	plan, err := rt.cache.GetLegacyPayPlan(r.Context(), vars["type"])
	if err != nil {
		rt.logError(fmt.Errorf("getLegacyPayPlan inGetPayPlan failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, plan)
}

/* Returns all pending load balancers for a given portal user ID. */
func (rt *Router) getPendingLoadBalancersByPortalID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userPortalID := vars["portalID"]

	lbs, err := rt.cache.GetLegacyPendingLoadBalancers(r.Context(), userPortalID)
	if err != nil {
		rt.logError(fmt.Errorf("getLegacyPendingLoadBalancers in GetPendingLoadBalancersByEmail failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, lbs)
}

/* Returns the number of load balancers a given user owns, by portal ID */
func (rt *Router) getLoadBalancersCountByPortalID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userPortalID := vars["portalID"]

	lbsCount := rt.cache.GetLegacyLoadBalancersCount(r.Context(), userPortalID)

	jsonresponse.RespondWithJSON(w, http.StatusOK, lbsCount)
}

/* ---------- V1 Create Methods ---------- */

func (rt *Router) createLoadBalancer(w http.ResponseWriter, r *http.Request) {
	var lb v1Types.LoadBalancer

	err := json.NewDecoder(r.Body).Decode(&lb)
	if err != nil {
		rt.logError(fmt.Errorf("createLoadBalancer Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	createdLB, err := rt.cache.CreateLegacyLoadBalancer(r.Context(), rt.writer, lb)
	if err != nil {
		rt.logError(fmt.Errorf("createLoadBalancer Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, createdLB)
}

func (rt *Router) createLoadBalancerUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	lbID := vars["id"]

	var user v1Types.UserAccess

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		rt.logError(fmt.Errorf("createLoadBalancerUser Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	updatedLB, err := rt.cache.CreateLegacyLoadBalancerUser(r.Context(), rt.writer, lbID, user)
	if err != nil {
		rt.logError(fmt.Errorf("writeLoadBalancerUser in CreateLoadBalancerUser failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedLB)
}

func (rt *Router) createLoadBalancerIntegration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	lbID := vars["id"]

	var integrationsInput types.AccountIntegrations

	err := json.NewDecoder(r.Body).Decode(&integrationsInput)
	if err != nil {
		rt.logError(fmt.Errorf("createLoadBalancerUser Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	updatedLB, err := rt.cache.CreateLegacyLoadBalancerIntegration(r.Context(), rt.writer, lbID, integrationsInput)
	if err != nil {
		rt.logError(fmt.Errorf("writeLoadBalancerUser in CreateLoadBalancerUser failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedLB)
}

func (rt *Router) createPortalUser(w http.ResponseWriter, r *http.Request) {
	var userInput types.CreateUser

	err := json.NewDecoder(r.Body).Decode(&userInput)
	if err != nil {
		rt.logError(fmt.Errorf("createPortalUser Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	newUser, accountID, err := rt.writer.WriteUserNewSignUp(r.Context(), userInput, time.Now())
	if err != nil {
		rt.logError(fmt.Errorf("writeUserNewSignUp failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	createPortalUserResponse := types.CreateUserResponse{
		User:      *newUser,
		AccountID: accountID,
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, createPortalUserResponse)
}

/* ---------- V1 Update Methods ---------- */

func (rt *Router) updateLoadBalancer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	lbID := vars["id"]

	// Note: It is intended to use UpdateApplication here
	var updateInput v1Types.UpdateApplication

	err := json.NewDecoder(r.Body).Decode(&updateInput)
	if err != nil {
		rt.logError(fmt.Errorf("updateLoadBalancer decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	loadBalancer, err := rt.cache.UpdateLegacyLoadBalancer(r.Context(), rt.writer, lbID, updateInput)
	if err != nil {
		rt.logError(fmt.Errorf("updateLegacyLoadBalancer in UpdateLoadBalancer failed: %w", err))
		if err.Error() == fmt.Errorf(cache.ErrPortalAppNotFound.Error(), lbID).Error() { // prevent a 500 when it should be a 404
			jsonresponse.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, loadBalancer)
}

func (rt *Router) updateFirstDateSurpassed(w http.ResponseWriter, r *http.Request) {
	rt.log.Info("Received UpdateFirstDateSurpassed request")

	var updateInput types.UpdateFirstDateSurpassed

	err := json.NewDecoder(r.Body).Decode(&updateInput)
	if err != nil {
		rt.logError(fmt.Errorf("updateFirstDateSurpassed decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	rt.log.WithFields(logrus.Fields{
		"applicationIDs":     updateInput.PortalAppIDs,
		"firstDateSurpassed": updateInput.FirstDateSurpassed,
	}).Info("Received UpdateFirstDateSurpassed request")

	appsToUpdate, err := rt.cache.UpdateLegacyApplicationsFirstDateSurpassed(r.Context(), rt.writer, updateInput)
	if err != nil {
		rt.logError(fmt.Errorf("updateLegacyApplicationsFirstDateSurpassed in UpdateFirstDateSurpassed failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, appsToUpdate)
}

func (rt *Router) updateLoadBalancerUserRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lbID := vars["id"]

	var updateUser v1Types.UpdateUserAccess

	err := json.NewDecoder(r.Body).Decode(&updateUser)
	if err != nil {
		rt.logError(fmt.Errorf("updateLoadBalancerUser Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	updatedLB, err := rt.cache.UpdateLegacyUserRole(r.Context(), rt.writer, lbID, updateUser)
	if err != nil {
		rt.logError(fmt.Errorf("updateLegacyUserRole for team member in UpdateLoadBalancerUser failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedLB)
}

func (rt *Router) acceptLoadBalancerUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lbID := vars["id"]

	var acceptUserUpdate v1Types.UpdateUserAccess

	err := json.NewDecoder(r.Body).Decode(&acceptUserUpdate)
	if err != nil {
		rt.logError(fmt.Errorf("acceptLoadBalancerUser Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	updatedLB, err := rt.cache.AcceptLegacyUser(r.Context(), rt.writer, lbID, acceptUserUpdate)
	if err != nil {
		rt.logError(fmt.Errorf("acceptLegacyUser in AcceptLoadBalancerUser failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedLB)
}

func (rt *Router) deleteLoadBalancerUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	lb, err := rt.cache.RemoveLegacyAccountUser(r.Context(), rt.writer, vars["userID"], vars["id"])
	if err != nil {
		rt.logError(fmt.Errorf("removeLegacyAccountUser in DeleteLoadBalancerUser failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, lb)
}
