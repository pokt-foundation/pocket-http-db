package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"github.com/gorilla/mux"
	v1Types "github.com/pokt-foundation/portal-db/types"
	"github.com/pokt-foundation/portal-db/v2/driver"
	"github.com/pokt-foundation/portal-db/v2/types"
	"github.com/pokt-foundation/portal-http-db/cache"
	jsonresponse "github.com/pokt-foundation/utils-go/json-response"
	"github.com/pokt-foundation/utils-go/logger"
)

var (
	errNoChainsFound                  error = errors.New("no chains were found")
	errNoGigastakeAppsFound           error = errors.New("no gigastake apps were found")
	errNoGigastakeAppsFoundForChain   error = errors.New("no gigastake apps were found for chain")
	errChainNotFound                  error = errors.New("chain not found")
	errGigastakeAppNotFound           error = errors.New("gigastake app not found")
	errNoPortalAppsFound              error = errors.New("no portal apps were found")
	errPortalAppNotFound              error = errors.New("portal app not found")
	errAccountNotFound                error = errors.New("account not found")
	errNoAccountsFoundForUserID       error = errors.New("no accounts were found for user ID")
	errUserIDNotFoundByProviderUserID error = errors.New("portal user ID not found for provider user ID: %s")
	errUserNotFoundByID               error = errors.New("user not found for ID: %s")

	errBlockchainNotFound    error = errors.New("blockchain not found")
	errNoBlockchainsFound    error = errors.New("no blockchains were found")
	errInvalidRoleNameFilter error = errors.New("invalid role name filter")
	errPermissionsNotFound   error = errors.New("no permissions found for given user ID")
)

// router struct handler for router requests
type (
	Router struct {
		Router   *mux.Router
		writer   driver.Writer
		cache    *cache.Cache
		apiKeys  map[string]bool
		logger   *logger.Logger
		imageTag string
	}
	Config struct {
		Cache    *cache.Cache
		Writer   driver.Writer
		APIKeys  map[string]bool
		Logger   *logger.Logger
		ImageTag string
	}
)

func (rt *Router) handleError(w http.ResponseWriter, methodName string, errorCode int, err error) {
	// Create error label containing method name
	errorMethod := fmt.Sprintf("error in %s", methodName)
	// Log error as JSON
	rt.logger.Error(errorMethod, slog.String("error", err.Error()))
	// Wrap error with method name
	errorString := fmt.Sprintf("%s: %s", errorMethod, err.Error())
	// Return error string as JSON
	jsonresponse.RespondWithError(w, errorCode, errorString)
}

func timingMiddleware(logger *logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("HTTP Request Starting", slog.String("uri", r.RequestURI))

			startTime := time.Now()
			next.ServeHTTP(w, r)
			elapsed := time.Since(startTime)

			logger.Info("HTTP Request Completed",
				slog.String("uri", r.RequestURI),
				slog.Int64("milliseconds", elapsed.Milliseconds()),
			)
		})
	}
}

// NewRouter returns router instance
func NewRouter(config Config) (*Router, error) {
	rt := &Router{
		Router:   mux.NewRouter(),
		writer:   config.Writer,
		cache:    config.Cache,
		apiKeys:  config.APIKeys,
		logger:   config.Logger,
		imageTag: config.ImageTag,
	}

	rt.Router.Use(timingMiddleware(rt.logger))

	rt.Router.HandleFunc("/healthz", rt.healthCheck).Methods(http.MethodGet)
	rt.Router.HandleFunc("/reset_cache", rt.resetCache).Methods(http.MethodPost)

	/* ----- V2 Endpoints ----- */
	// V2 endpoints are for the new DB schema and access the cache and V2 structs directly.

	rt.Router.HandleFunc("/v2/chain/{id}", rt.getChainByID).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/gigastake/{id}", rt.getGigastakeAppByID).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/chain", rt.getAllChains).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/gigastake", rt.getAllGigastakeApps).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/chain/{id}/gigastake", rt.getAllGigastakeAppsByChain).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/chain", rt.createNewChainAndGigastakeApps).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v2/chain/gigastake", rt.createNewGigastakeApp).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v2/chain/gigastake/{id}", rt.updateGigastakeApp).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v2/chain/{id}", rt.updateChain).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v2/chain/gigastake", rt.updateGigastakeApp).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v2/chain/{id}/activate", rt.activateChain).Methods(http.MethodPut)

	rt.Router.HandleFunc("/v2/portal_app/{id}", rt.getPortalAppByID).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/portal_app", rt.getAllPortalApps).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/portal_app", rt.createNewPortalApp).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v2/portal_app/{id}", rt.updatePortalApp).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v2/portal_app/{id}", rt.deletePortalApp).Methods(http.MethodDelete)
	rt.Router.HandleFunc("/v2/portal_app/first_date_surpassed", rt.updatePortalAppFirstDateSurpassed).Methods(http.MethodPost)

	rt.Router.HandleFunc("/v2/middleware/portal_app", rt.getPortalAppsForMiddleware).Methods(http.MethodGet)

	rt.Router.HandleFunc("/v2/account", rt.getAccounts).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/account/{id}", rt.updateAccount).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v2/account/{id}/integration", rt.createAccountIntegration).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v2/account/{id}/integration", rt.updateAccountIntegration).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v2/account/{id}", rt.deleteAccount).Methods(http.MethodDelete)

	rt.Router.HandleFunc("/v2/account/user", rt.writeAccountUser).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v2/account/user/update_role", rt.updateAccountUserRole).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v2/account/user/accept", rt.acceptAccountUser).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v2/account/user/remove", rt.removeAccountUser).Methods(http.MethodPut)

	rt.Router.HandleFunc("/v2/user", rt.createUser).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v2/user", rt.updateUser).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v2/user/{userID}", rt.getPortalUser).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/user/{userID}/portal_app", rt.getPortalAppsByUser).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/user/{userID}/account", rt.getUserAccounts).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/user/{userID}/account/{id}", rt.getUserAccount).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/user/{userID}/account", rt.createAccount).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v2/user/{userID}", rt.deleteUser).Methods(http.MethodDelete)

	rt.Router.HandleFunc("/v2/blocked_contract", rt.getBlockedContracts).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v2/blocked_contract", rt.writeBlockedContract).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v2/blocked_contract/{address}/active", rt.updateBlockedContractActive).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v2/blocked_contract/{address}", rt.removeBlockedContract).Methods(http.MethodDelete)

	rt.Router.HandleFunc("/v2/plan", rt.getPlans).Methods(http.MethodGet)

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
	rt.Router.HandleFunc("/v1/user/{userID}/portal_id", rt.getPortalUserID).Methods(http.MethodGet)
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
		if r.URL.Path == "/healthz" {
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
	var msg string
	var statusCode int
	if err := rt.writer.Ping(r.Context()); err == nil {
		statusCode = http.StatusOK
		msg = fmt.Sprintf("DB Check Done. Portal HTTP DB is up and running!\nImage Tag: %s", rt.imageTag)
	} else {
		statusCode = http.StatusServiceUnavailable
		msg = fmt.Sprintf("Could not connect to the db: %s", err.Error())
	}

	w.WriteHeader(statusCode)
	_, err := w.Write([]byte(msg))
	if err != nil {
		panic(err)
	}
}

func (rt *Router) resetCache(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	err := rt.cache.ResetCache(r.Context())
	elapsedTime := time.Since(startTime).Milliseconds()
	if err != nil {
		rt.handleError(w, "resetCache", http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := fmt.Sprintf("cache manually reset successfully in %dms", elapsedTime)
	_, err = w.Write([]byte(resp))
	if err != nil {
		rt.handleError(w, "resetCache", http.StatusInternalServerError, err)
		return
	}
}

/* ---------- V2 Chain Handler Methods ---------- */

// GET /v2/chain/{id} - Fetches a single blockchain by its RelayChainID.
func (rt *Router) getChainByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chainID := types.RelayChainID(vars["id"])
	if chainID == "" {
		rt.handleError(w, "getChainByID", http.StatusBadRequest, fmt.Errorf("chain ID must be provided"))
		return
	}

	chain := rt.cache.GetChain(chainID, cache.Options{})
	if chain.ID == "" {
		rt.handleError(w, "getChainByID", http.StatusNotFound, errChainNotFound)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, chain)
}

// GET /v2/gigastake/{id} - Fetches a single gigastake app by its GigastakeAppID.
func (rt *Router) getGigastakeAppByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gigastakeAppID := types.GigastakeAppID(vars["id"])
	if gigastakeAppID == "" {
		rt.handleError(w, "getGigastakeAppByID", http.StatusBadRequest, fmt.Errorf("gigastake app ID must be provided"))
		return
	}

	gigastakeApp := rt.cache.GetGigastakeAppByID(gigastakeAppID)
	if gigastakeApp.ID == "" {
		rt.handleError(w, "getGigastakeAppByID", http.StatusNotFound, errGigastakeAppNotFound)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, gigastakeApp)
}

// GET /v2/chain - Fetches all blockchains.
func (rt *Router) getAllChains(w http.ResponseWriter, r *http.Request) {
	queryParams := map[queryParam]queryParamDetail{
		commonParams.includeDeleted: {required: false},
		chainParams.includeInactive: {required: false},
	}

	options, err := paramProcessor(r.URL, queryParams)
	if err != nil {
		rt.handleError(w, "getAllChains", http.StatusBadRequest, err)
		return
	}

	chains := rt.cache.GetChains(options)
	if len(chains) == 0 {
		rt.handleError(w, "getAllChains", http.StatusNotFound, errNoChainsFound)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, chains)
}

// GET /v2/gigastake - Fetches all gigastake apps.
func (rt *Router) getAllGigastakeApps(w http.ResponseWriter, r *http.Request) {
	queryParams := map[queryParam]queryParamDetail{
		commonParams.includeDeleted: {required: false},
	}

	options, err := paramProcessor(r.URL, queryParams)
	if err != nil {
		rt.handleError(w, "getAllGigastakeApps", http.StatusBadRequest, err)
		return
	}

	gigastakeApps := rt.cache.GetGigastakeApps(options)
	if len(gigastakeApps) == 0 {
		rt.handleError(w, "getAllGigastakeApps", http.StatusNotFound, errNoGigastakeAppsFound)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, gigastakeApps)
}

// GET /v2/chain/{id}/gigastake - Fetches all gigastake apps for a single chain ID.
func (rt *Router) getAllGigastakeAppsByChain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chainID := types.RelayChainID(vars["id"])
	if chainID == "" {
		rt.handleError(w, "getAllGigastakeAppsByChain", http.StatusBadRequest, fmt.Errorf("chain ID must be provided"))
		return
	}

	chain := rt.cache.GetChain(chainID, cache.Options{})
	if chain.ID == "" {
		rt.handleError(w, "getAllGigastakeAppsByChain", http.StatusBadRequest, errChainNotFound)
		return
	}

	gigastakeAppsByChain := []types.GigastakeApp{}
	for _, gigastakeApp := range chain.GigastakeApps {
		gigastakeAppsByChain = append(gigastakeAppsByChain, *gigastakeApp)
	}

	if len(gigastakeAppsByChain) == 0 {
		rt.handleError(w, "getAllGigastakeAppsByChain", http.StatusNotFound, errNoGigastakeAppsFoundForChain)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, gigastakeAppsByChain)
}

// POST /v2/chain - Creates a new blockchains and its Gigastake apps.
func (rt *Router) createNewChainAndGigastakeApps(w http.ResponseWriter, r *http.Request) {
	var newChainInput types.NewChainInput

	err := json.NewDecoder(r.Body).Decode(&newChainInput)
	if err != nil {
		rt.handleError(w, "createNewChainAndGigastakeApps", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	createdChainAndGigastakeApps, err := rt.writer.WriteChainAndGigastakeApps(r.Context(), newChainInput, time.Now())
	if err != nil {
		rt.handleError(w, "createNewChainAndGigastakeApps", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, createdChainAndGigastakeApps)
}

// POST /v2/chain/gigastake - Creates a new Gigastake app.
func (rt *Router) createNewGigastakeApp(w http.ResponseWriter, r *http.Request) {
	var gigastakeAppInput types.GigastakeApp

	err := json.NewDecoder(r.Body).Decode(&gigastakeAppInput)
	if err != nil {
		rt.handleError(w, "createNewGigastakeApp", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	createdGigastakeApp, err := rt.writer.WriteGigastakeApp(r.Context(), gigastakeAppInput, time.Now())
	if err != nil {
		rt.handleError(w, "createNewGigastakeApp", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, createdGigastakeApp)
}

// PUT /v2/chain/{id} - Updates an existing blockchain.
func (rt *Router) updateChain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chainID := types.RelayChainID(vars["id"])
	if chainID == "" {
		rt.handleError(w, "updateChain", http.StatusBadRequest, fmt.Errorf("chain ID must be provided"))
		return
	}

	var chainUpdate types.UpdateChain

	err := json.NewDecoder(r.Body).Decode(&chainUpdate)
	if err != nil {
		rt.handleError(w, "updateChain", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	chainUpdate.ID = chainID

	updatedChain, err := rt.writer.UpdateChain(r.Context(), chainUpdate, time.Now())
	if err != nil {
		rt.handleError(w, "updateChain", http.StatusInternalServerError, err)
		return
	}

	chainInCache := rt.cache.GetChain(chainID, cache.Options{})
	if chainInCache.ID != "" {
		updatedChain.GigastakeApps = chainInCache.GigastakeApps
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedChain)
}

// PUT /v2/chain/gigastake/{id} - Updates a Gigastake app by ID.
func (rt *Router) updateGigastakeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gigastakeAppID := types.GigastakeAppID(vars["id"])
	if gigastakeAppID == "" {
		rt.handleError(w, "updateGigastakeApp", http.StatusBadRequest, fmt.Errorf("gigastake app ID must be provided"))
		return
	}

	var gigastakeAppUpdate types.UpdateGigastakeApp

	err := json.NewDecoder(r.Body).Decode(&gigastakeAppUpdate)
	if err != nil {
		rt.handleError(w, "updateGigastakeApp", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	gigastakeAppUpdate.ID = gigastakeAppID

	err = rt.writer.UpdateGigastakeApp(r.Context(), gigastakeAppUpdate, time.Now())
	if err != nil {
		rt.handleError(w, "updateGigastakeApp", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, gigastakeAppUpdate)
}

// PUT /v2/chain/{id}/activate - Activates or deactivates a blockchain by ID.
func (rt *Router) activateChain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chainID := types.RelayChainID(vars["id"])
	if chainID == "" {
		rt.handleError(w, "activateChain", http.StatusBadRequest, fmt.Errorf("chain ID must be provided"))
		return
	}

	var active bool

	err := json.NewDecoder(r.Body).Decode(&active)
	if err != nil {
		rt.handleError(w, "activateChain", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	activated, err := rt.writer.SetChainActiveStatus(r.Context(), chainID, active, time.Now())
	if err != nil {
		rt.handleError(w, "activateChain", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, activated)
}

/* ---------- V2 Portal App Handler Methods ---------- */

// GET /v2/portal_app/{id} - Fetches a single portal application by its PortalAppID.
func (rt *Router) getPortalAppByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portalAppID := types.PortalAppID(vars["id"])
	if portalAppID == "" {
		rt.handleError(w, "getPortalAppByID", http.StatusBadRequest, fmt.Errorf("portal app ID must be provided"))
	}

	portalApp := rt.cache.GetPortalApp(portalAppID)
	if portalApp.ID == "" {
		rt.handleError(w, "getPortalAppByID", http.StatusNotFound, errPortalAppNotFound)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, portalApp)
}

// GET /v2/portal_app - Fetches all portal applications.
func (rt *Router) getAllPortalApps(w http.ResponseWriter, r *http.Request) {
	queryParams := map[queryParam]queryParamDetail{
		commonParams.includeDeleted: {required: false},
	}

	options, err := paramProcessor(r.URL, queryParams)
	if err != nil {
		rt.handleError(w, "getAllPortalApps", http.StatusBadRequest, err)
		return
	}

	portalApps := rt.cache.GetPortalApps(options)
	if len(portalApps) == 0 {
		rt.handleError(w, "getAllPortalApps", http.StatusNotFound, errNoPortalAppsFound)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, portalApps)
}

// GET /v2/middleware/portal_app - Fetches all middleware portal applications.
func (rt *Router) getPortalAppsForMiddleware(w http.ResponseWriter, r *http.Request) {
	middlewarePortalApps := rt.cache.GetPortalAppsForMiddleware()
	if len(middlewarePortalApps) == 0 {
		rt.handleError(w, "getPortalAppsForMiddleware", http.StatusNotFound, errNoPortalAppsFound)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, middlewarePortalApps)
}

// GET /v2/user/{userID}/portal_app - Fetches all portal applications.
func (rt *Router) getPortalAppsByUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := types.UserID(vars["userID"])
	if userID == "" {
		rt.handleError(w, "getPortalAppsByUser", http.StatusBadRequest, fmt.Errorf("user ID must be provided"))
	}

	queryParams := map[queryParam]queryParamDetail{
		portalAppParams.roleNameFilters: {required: false},
		portalAppParams.accepted:        {required: false},
	}

	options, err := paramProcessor(r.URL, queryParams)
	if err != nil {
		rt.handleError(w, "getPortalAppsByUser", http.StatusBadRequest, err)
		return
	}

	portalAppsForUser := rt.cache.GetPortalAppsByUserID(userID, options)
	if len(portalAppsForUser) == 0 {
		rt.handleError(w, "getPortalAppsByUser", http.StatusNotFound, errNoPortalAppsFound)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, portalAppsForUser)
}

// POST /v2/portal_app - Creates a new portal application.
func (rt *Router) createNewPortalApp(w http.ResponseWriter, r *http.Request) {
	var portalAppInput types.PortalApp

	err := json.NewDecoder(r.Body).Decode(&portalAppInput)
	if err != nil {
		rt.handleError(w, "createNewPortalApp", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	// portalAppInput.AATs only ever contains a single AAT
	var aat types.AAT
	for _, inputAAT := range portalAppInput.AATs {
		aat = inputAAT
		break
	}

	createdPortalApp, err := rt.writer.WritePortalApp(r.Context(), portalAppInput, aat, time.Now())
	if err != nil {
		rt.handleError(w, "createNewPortalApp", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, createdPortalApp)
}

// PUT /v2/portal_app/{id} - Updates a portal application.
func (rt *Router) updatePortalApp(w http.ResponseWriter, r *http.Request) {
	var portalAppUpdate types.UpdatePortalApp

	err := json.NewDecoder(r.Body).Decode(&portalAppUpdate)
	if err != nil {
		rt.handleError(w, "updatePortalApp", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	vars := mux.Vars(r)
	portalAppUpdate.AppID = types.PortalAppID(vars["id"])

	err = rt.writer.UpdatePortalApp(r.Context(), portalAppUpdate, time.Now())
	if err != nil {
		rt.handleError(w, "updatePortalApp", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, portalAppUpdate)
}

// DELETE /v2/portal_app/{id} - Deletes a portal application.
func (rt *Router) deletePortalApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portalAppID := types.PortalAppID(vars["id"])
	if portalAppID == "" {
		rt.handleError(w, "deletePortalApp", http.StatusBadRequest, fmt.Errorf("portal app ID must be provided"))
	}

	err := rt.writer.SetPortalAppDeleted(r.Context(), portalAppID, time.Now())
	if err != nil {
		rt.handleError(w, "deletePortalApp", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// POST /v2/portal_app/first_date_surpassed - Updates one or more PortalApps' FirstDateSurpassed field.
func (rt *Router) updatePortalAppFirstDateSurpassed(w http.ResponseWriter, r *http.Request) {
	var firstDateSurpassedUpdate *types.UpdateFirstDateSurpassed

	err := json.NewDecoder(r.Body).Decode(&firstDateSurpassedUpdate)
	if err != nil {
		rt.handleError(w, "updatePortalAppFirstDateSurpassed", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	if len(firstDateSurpassedUpdate.PortalAppIDs) == 0 {
		rt.handleError(w, "updatePortalAppFirstDateSurpassed", http.StatusBadRequest, fmt.Errorf("no portal app IDs provided"))
	}

	err = rt.writer.UpdatePortalAppsFirstDateSurpassed(r.Context(), firstDateSurpassedUpdate)
	if err != nil {
		rt.handleError(w, "updatePortalAppFirstDateSurpassed", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

/* ---------- V2 Account Handler Methods ---------- */

// GET /v2/account - Get all Accounts.
func (rt *Router) getAccounts(w http.ResponseWriter, r *http.Request) {
	queryParams := map[queryParam]queryParamDetail{
		commonParams.includeDeleted: {required: false},
	}

	options, err := paramProcessor(r.URL, queryParams)
	if err != nil {
		rt.handleError(w, "getAccounts", http.StatusBadRequest, err)
		return
	}

	accounts := rt.cache.GetAccounts(options)
	if len(accounts) == 0 {
		rt.handleError(w, "getAccounts", http.StatusNotFound, errAccountNotFound)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, accounts)
}

// GET /v2/{userID}/account - Get all user accounts by user ID.
func (rt *Router) getUserAccounts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := types.UserID(vars["userID"])
	if userID == "" {
		rt.handleError(w, "getUserAccounts", http.StatusBadRequest, fmt.Errorf("user ID must be provided"))
	}

	queryParams := map[queryParam]queryParamDetail{
		accountParams.roleNameFilters: {required: false},
		accountParams.accepted:        {required: false},
	}

	options, err := paramProcessor(r.URL, queryParams)
	if err != nil {
		rt.handleError(w, "getUserAccounts", http.StatusBadRequest, err)
		return
	}

	accounts := rt.cache.GetUserAccounts(userID, options)

	if len(accounts) == 0 {
		rt.handleError(w, "getUserAccounts", http.StatusNotFound, errNoAccountsFoundForUserID)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, accounts)
}

// GET /v2/user/{userID}/account/{id} - Get one user account by account ID.
func (rt *Router) getUserAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := types.UserID(vars["userID"])
	if userID == "" {
		rt.handleError(w, "getUserAccount", http.StatusBadRequest, fmt.Errorf("user ID must be provided"))
	}
	accountID := types.AccountID(vars["id"])
	if accountID == "" {
		rt.handleError(w, "getUserAccount", http.StatusBadRequest, fmt.Errorf("account ID must be provided"))
	}

	queryParams := map[queryParam]queryParamDetail{
		accountParams.roleNameFilters: {required: false},
		accountParams.accepted:        {required: false},
	}

	options, err := paramProcessor(r.URL, queryParams)
	if err != nil {
		rt.handleError(w, "getUserAccount", http.StatusBadRequest, err)
		return
	}

	account := rt.cache.GetUserAccount(accountID, userID, options)
	if account.ID == "" {
		rt.handleError(w, "getUserAccount", http.StatusNotFound, errAccountNotFound)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, account)
}

// POST /v2/{userID}/account - Creates an Account in the database for a single user.
func (rt *Router) createAccount(w http.ResponseWriter, r *http.Request) {
	var account types.Account

	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		rt.handleError(w, "createAccount", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	vars := mux.Vars(r)
	userID := types.UserID(vars["userID"])
	if userID == "" {
		rt.handleError(w, "createAccount", http.StatusBadRequest, fmt.Errorf("user ID must be provided"))
	}

	createdAccount, err := rt.writer.WriteAccount(r.Context(), userID, account, time.Now())
	if err != nil {
		rt.handleError(w, "createAccount", http.StatusInternalServerError, err)
		return
	}

	plan := rt.cache.GetPlan(createdAccount.PlanType)
	if plan.Type == "" {
		rt.handleError(w, "createAccount", http.StatusInternalServerError, fmt.Errorf("plan type %s not found", createdAccount.PlanType))
	}
	createdAccount.Plan = &plan

	jsonresponse.RespondWithJSON(w, http.StatusOK, createdAccount)
}

// PUT /v2/account/{id} - Update an Account.
func (rt *Router) updateAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := types.AccountID(vars["id"])
	if accountID == "" {
		rt.handleError(w, "updateAccount", http.StatusBadRequest, fmt.Errorf("account ID must be provided"))
	}

	var update types.UpdateAccount

	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		rt.handleError(w, "updateAccount", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	update.AccountID = accountID

	updatedAccount, err := rt.writer.UpdateAccount(r.Context(), update, time.Now())
	if err != nil {
		rt.handleError(w, "updateAccount", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedAccount)
}

// POST /v2/account/{id}/integration - Add the Account Integrations for a single Account.
func (rt *Router) createAccountIntegration(w http.ResponseWriter, r *http.Request) {
	var integrations types.AccountIntegrations

	err := json.NewDecoder(r.Body).Decode(&integrations)
	if err != nil {
		rt.handleError(w, "createAccountIntegration", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	vars := mux.Vars(r)
	accountID := types.AccountID(vars["id"])
	integrations.AccountID = accountID

	createdIntegration, err := rt.writer.UpsertAccountIntegration(r.Context(), integrations)
	if err != nil {
		rt.handleError(w, "createAccountIntegration", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, createdIntegration)
}

// PUT /v2/account/{id}/integration - Updates the Account Integrations for a single Account.
func (rt *Router) updateAccountIntegration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := types.AccountID(vars["id"])
	if accountID == "" {
		rt.handleError(w, "updateAccountIntegration", http.StatusBadRequest, fmt.Errorf("account ID must be provided"))
	}

	var integrations types.AccountIntegrations

	err := json.NewDecoder(r.Body).Decode(&integrations)
	if err != nil {
		rt.handleError(w, "updateAccountIntegration", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	integrations.AccountID = accountID

	updatedIntegration, err := rt.writer.UpsertAccountIntegration(r.Context(), integrations)
	if err != nil {
		rt.handleError(w, "updateAccountIntegration", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedIntegration)
}

// DELETE /v2/account/{id} - Set the Account's Deleted field to true.
func (rt *Router) deleteAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := types.AccountID(vars["id"])
	if accountID == "" {
		rt.handleError(w, "deleteAccount", http.StatusBadRequest, fmt.Errorf("account ID must be provided"))
	}

	err := rt.writer.SetAccountDeleted(r.Context(), accountID, time.Now())
	if err != nil {
		rt.handleError(w, "deleteAccount", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

/* ---------- V2 Account User Handler Methods ---------- */

// POST /v2/account/user - Create a single Account User.
func (rt *Router) writeAccountUser(w http.ResponseWriter, r *http.Request) {
	var createUser types.CreateAccountUserAccess

	err := json.NewDecoder(r.Body).Decode(&createUser)
	if err != nil {
		rt.handleError(w, "writeAccountUser", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	userID, err := rt.writer.WriteAccountUser(r.Context(), createUser, time.Now())
	if err != nil {
		rt.handleError(w, "writeAccountUser", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, map[string]types.UserID{"userID": userID})
}

// PUT /v2/account/user/update_role - Update the role for a single Account User.
func (rt *Router) updateAccountUserRole(w http.ResponseWriter, r *http.Request) {
	var updateUser types.UpdateAccountUserRole

	err := json.NewDecoder(r.Body).Decode(&updateUser)
	if err != nil {
		rt.handleError(w, "updateAccountUserRole", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	err = rt.writer.SetAccountUserRole(r.Context(), updateUser, time.Now())
	if err != nil {
		rt.handleError(w, "updateAccountUserRole", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// PUT /v2/account/user/accept - Accept an AccountUser.
func (rt *Router) acceptAccountUser(w http.ResponseWriter, r *http.Request) {
	var acceptUser types.UpdateAcceptAccountUser

	err := json.NewDecoder(r.Body).Decode(&acceptUser)
	if err != nil {
		rt.handleError(w, "acceptAccountUser", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	err = rt.writer.UpdateAcceptAccountUser(r.Context(), acceptUser, time.Now())
	if err != nil {
		rt.handleError(w, "acceptAccountUser", http.StatusInternalServerError, err)
		return
	}

	var status string
	switch acceptUser.DeclinedInvite {
	case true:
		status = "declined"
	default:
		status = "accepted"
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, map[string]string{"status": status})
}

// PUT /v2/account/user/remove - Remove an AccountUser.
func (rt *Router) removeAccountUser(w http.ResponseWriter, r *http.Request) {
	var removeUser types.UpdateRemoveAccountUser

	err := json.NewDecoder(r.Body).Decode(&removeUser)
	if err != nil {
		rt.handleError(w, "removeAccountUser", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	err = rt.writer.RemoveAccountUser(r.Context(), removeUser.UserID, removeUser.PortalAppID, removeUser.AccountID)
	if err != nil {
		rt.handleError(w, "removeAccountUser", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

/* ---------- V2 User Handler Methods ---------- */

// GET /v2/user/{userID} - Gets the Portal User ID for a given auth provider user ID.
func (rt *Router) getPortalUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	queryParams := map[queryParam]queryParamDetail{
		userParams.fullDetails: {required: false},
	}

	options, err := paramProcessor(r.URL, queryParams)
	if err != nil {
		rt.handleError(w, "getPortalUser", http.StatusBadRequest, err)
		return
	}

	var user types.User

	// First try to get user by provider user ID
	userByProviderID := rt.cache.GetUserByProviderID(types.ProviderUserID(userID))
	if userByProviderID.ID != "" {
		user = userByProviderID
	} else {
		// Then try and get user by portal user ID
		userByPortalID := rt.cache.GetUserByID(types.UserID(userID))
		if userByPortalID.ID != "" {
			user = userByPortalID
		}
	}
	// If both user ID types fail to find a user, return a 404
	if user.ID == "" {
		rt.handleError(w, "getPortalUser", http.StatusNotFound, fmt.Errorf(errUserNotFoundByID.Error(), userID))
		return
	}

	switch {
	// If 'full_details' param provided return entire user
	case options.FullDetails:
		jsonresponse.RespondWithJSON(w, http.StatusOK, user)
	// If 'full_details' param not provided or false return only user ID
	default:
		jsonresponse.RespondWithJSON(w, http.StatusOK, user.ID)
	}
}

// GET /v2/user/{userID} - Gets the Portal User ID for a given auth provider user ID.
func (rt *Router) getPortalUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerUserID := types.ProviderUserID(vars["userID"])

	portalUser := rt.cache.GetUserByProviderID(providerUserID)
	if portalUser.ID == "" {
		rt.handleError(w, "getPortalUserID", http.StatusNotFound, fmt.Errorf(errUserIDNotFoundByProviderUserID.Error(), providerUserID))
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, portalUser.ID)
}

// POST /v2/user - Creates a new User in the database. Called in response to a new user sign up.
func (rt *Router) createUser(w http.ResponseWriter, r *http.Request) {
	var user types.CreateUser

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		rt.handleError(w, "createUser", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	createdUser, accountID, err := rt.writer.WriteUserNewSignUp(r.Context(), user, time.Now())
	if err != nil {
		rt.handleError(w, "createUser", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, types.CreateUserResponse{
		User:      *createdUser,
		AccountID: accountID,
	})
}

// PUT /v2/user - Updates a User in the database.
func (rt *Router) updateUser(w http.ResponseWriter, r *http.Request) {
	var updateUser types.UpdateUser

	err := json.NewDecoder(r.Body).Decode(&updateUser)
	if err != nil {
		rt.handleError(w, "updateUser", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	updatedUser, err := rt.writer.UpdateUser(r.Context(), updateUser, time.Now())
	if err != nil {
		rt.handleError(w, "updateUser", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedUser)
}

// DELETE /v2/user/{userID} - Deletes a User (WARNING: this will perform a hard delete in the database).
func (rt *Router) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := types.UserID(vars["userID"])

	_, err := rt.writer.DeletePortalUser(r.Context(), userID)
	if err != nil {
		rt.handleError(w, "deleteUser", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

/* ---------- V2 Plan Handler Methods ---------- */

func (rt *Router) getPlans(w http.ResponseWriter, r *http.Request) {
	jsonresponse.RespondWithJSON(w, http.StatusOK, rt.cache.GetPlans())
}

/* ---------- V2 Blocked Contracts Handler Methods ---------- */

// GET /v2/blocked_contract - Get all Blocked contracts.
func (rt *Router) getBlockedContracts(w http.ResponseWriter, r *http.Request) {
	blockedContracts := rt.cache.GetBlockedContracts()

	jsonresponse.RespondWithJSON(w, http.StatusOK, blockedContracts)
}

// POST /v2/blocked_contract - Add a new blocked address to the global blocked contracts table.
func (rt *Router) writeBlockedContract(w http.ResponseWriter, r *http.Request) {
	var blockedContract types.BlockedContract

	err := json.NewDecoder(r.Body).Decode(&blockedContract)
	if err != nil {
		rt.handleError(w, "writeBlockedContract", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	err = rt.writer.WriteBlockedContract(r.Context(), blockedContract.BlockedAddress, time.Now())
	if err != nil {
		rt.handleError(w, "writeBlockedContract", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "added"})
}

// PUT /v2/blocked_contract/{address}/active - Activate or deactivate a blocked address in the global blocked contracts table.
func (rt *Router) updateBlockedContractActive(w http.ResponseWriter, r *http.Request) {
	var blockedContract types.BlockedContract

	err := json.NewDecoder(r.Body).Decode(&blockedContract)
	if err != nil {
		rt.handleError(w, "updateBlockedContractActive", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	vars := mux.Vars(r)
	blockedAddress := types.BlockedAddress(vars["address"])

	err = rt.writer.UpdateBlockedContractActive(r.Context(), blockedAddress, blockedContract.Active, time.Now())
	if err != nil {
		rt.handleError(w, "updateBlockedContractActive", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, map[string]bool{"contract_active": blockedContract.Active})
}

// DELETE /v2/blocked_contract/{address} - Delete a blocked address from the global blocked contracts table.
func (rt *Router) removeBlockedContract(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockedAddress := types.BlockedAddress(vars["address"])

	err := rt.writer.RemoveBlockedContract(r.Context(), blockedAddress)
	if err != nil {
		rt.handleError(w, "removeBlockedContract", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

/* ---------- V1 Read Methods ---------- */

func (rt *Router) getUserPermissionsByUserID(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - getUserPermissionsByUserID")

	vars := mux.Vars(r)
	userID := vars["id"]

	permissions, err := rt.cache.GetLegacyUserPermissionsByUserID(r.Context(), userID)
	if err != nil {
		rt.handleError(w, "getUserPermissionsByUserID", http.StatusNotFound, err)
		return
	}
	if permissions.UserID == "" {
		rt.handleError(w, "getUserPermissionsByUserID", http.StatusNotFound, errPermissionsNotFound)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, permissions)
}

func (rt *Router) getApplicationsByUserID(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - getApplicationsByUserID")

	vars := mux.Vars(r)

	apps, err := rt.cache.GetLegacyApplicationsByUserID(r.Context(), vars["id"])
	if err != nil {
		rt.handleError(w, "getApplicationsByUserID", http.StatusNotFound, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, apps)
}

func (rt *Router) getApplications(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - getApplications")

	applications := rt.cache.GetLegacyApplications(cache.Options{
		IncludeDeleted: false,
	})

	jsonresponse.RespondWithJSON(w, http.StatusOK, applications)
}

func (rt *Router) getLoadBalancersByUserID(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - getLoadBalancersByUserID")

	vars := mux.Vars(r)
	userID := vars["id"]

	roleNameFilter := v1Types.RoleName(r.URL.Query().Get("filter"))
	if roleNameFilter != v1Types.RoleName("") && !v1Types.ValidRoleNames[roleNameFilter] {
		jsonresponse.RespondWithError(w, http.StatusBadRequest, errInvalidRoleNameFilter.Error())
		return
	}

	lbs, err := rt.cache.GetLegacyLoadBalancersByUserID(r.Context(), userID, roleNameFilter)
	if err != nil {
		rt.handleError(w, "getLoadBalancersByUserID", http.StatusNotFound, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, lbs)
}

func (rt *Router) getLoadBalancers(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - getLoadBalancers")

	lbs := rt.cache.GetLegacyLoadBalancers(r.Context())

	jsonresponse.RespondWithJSON(w, http.StatusOK, lbs)
}

func (rt *Router) getLoadBalancer(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - getLoadBalancer")

	vars := mux.Vars(r)

	lb, err := rt.cache.GetLegacyLoadBalancer(r.Context(), vars["id"])
	if err != nil {
		rt.handleError(w, "getLoadBalancer", http.StatusNotFound, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, lb)
}

func (rt *Router) getBlockchains(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - getBlockchains")

	queryParams := map[queryParam]queryParamDetail{
		chainParams.includeInactive: {required: false},
	}

	options, err := paramProcessor(r.URL, queryParams)
	if err != nil {
		rt.handleError(w, "getBlockchains", http.StatusBadRequest, err)
		return
	}

	blockchains := rt.cache.GetLegacyBlockchains(r.Context(), options.IncludeInactiveChains)
	if len(blockchains) == 0 {
		rt.handleError(w, "getBlockchains", http.StatusNotFound, errNoBlockchainsFound)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, blockchains)
}

func (rt *Router) getBlockchain(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - getBlockchain")

	vars := mux.Vars(r)

	blockchain := rt.cache.GetLegacyBlockchain(r.Context(), vars["id"])
	if blockchain.ID == "" {
		rt.handleError(w, "getBlockchain", http.StatusNotFound, errBlockchainNotFound)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, blockchain)
}

func (rt *Router) getPayPlans(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - getPayPlans")

	jsonresponse.RespondWithJSON(w, http.StatusOK, rt.cache.GetLegacyPayPlans(r.Context()))
}

func (rt *Router) getPayPlan(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - getPayPlan")

	vars := mux.Vars(r)

	plan, err := rt.cache.GetLegacyPayPlan(r.Context(), vars["type"])
	if err != nil {
		rt.handleError(w, "getPayPlan", http.StatusNotFound, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, plan)
}

/* Returns all pending load balancers for a given portal user ID. */
func (rt *Router) getPendingLoadBalancersByPortalID(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - getPendingLoadBalancersByPortalID")

	vars := mux.Vars(r)
	userPortalID := vars["portalID"]

	lbs, err := rt.cache.GetLegacyPendingLoadBalancers(r.Context(), userPortalID)
	if err != nil {
		rt.handleError(w, "getPendingLoadBalancersByPortalID", http.StatusNotFound, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, lbs)
}

/* Returns the number of load balancers a given user owns, by portal ID */
func (rt *Router) getLoadBalancersCountByPortalID(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - getLoadBalancersCountByPortalID")

	vars := mux.Vars(r)
	userPortalID := vars["portalID"]

	lbsCount := rt.cache.GetLegacyLoadBalancersCount(r.Context(), userPortalID)

	jsonresponse.RespondWithJSON(w, http.StatusOK, lbsCount)
}

/* ---------- V1 Create Methods ---------- */
func (rt *Router) createLoadBalancer(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - createLoadBalancer")

	var lb v1Types.LoadBalancer

	err := json.NewDecoder(r.Body).Decode(&lb)
	if err != nil {
		rt.handleError(w, "createLoadBalancer", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	createdLB, err := rt.cache.CreateLegacyLoadBalancer(r.Context(), rt.writer, lb)
	if err != nil {
		rt.handleError(w, "createLoadBalancer", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, createdLB)
}

func (rt *Router) createLoadBalancerUser(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - createLoadBalancerUser")

	vars := mux.Vars(r)
	lbID := vars["id"]

	var user v1Types.UserAccess

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		rt.handleError(w, "createLoadBalancerUser", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	updatedLB, err := rt.cache.CreateLegacyLoadBalancerUser(r.Context(), rt.writer, lbID, user)
	if err != nil {
		rt.handleError(w, "createLoadBalancerUser", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedLB)
}

func (rt *Router) createLoadBalancerIntegration(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - createLoadBalancerIntegration")

	vars := mux.Vars(r)
	lbID := vars["id"]

	var integrationsInput types.AccountIntegrations

	err := json.NewDecoder(r.Body).Decode(&integrationsInput)
	if err != nil {
		rt.handleError(w, "createLoadBalancerIntegration", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	updatedLB, err := rt.cache.CreateLegacyLoadBalancerIntegration(r.Context(), rt.writer, lbID, integrationsInput)
	if err != nil {
		rt.handleError(w, "createLoadBalancerIntegration", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedLB)
}

func (rt *Router) createPortalUser(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - createPortalUser")

	var userInput types.CreateUser

	err := json.NewDecoder(r.Body).Decode(&userInput)
	if err != nil {
		rt.handleError(w, "createPortalUser", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	newUser, accountID, err := rt.writer.WriteUserNewSignUp(r.Context(), userInput, time.Now())
	if err != nil {
		rt.handleError(w, "createPortalUser", http.StatusBadRequest, err)
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
	rt.logger.Info("V1 endpoint hit - updateLoadBalancer")

	vars := mux.Vars(r)
	lbID := vars["id"]

	var updateInput v1Types.UpdateApplication

	err := json.NewDecoder(r.Body).Decode(&updateInput)
	if err != nil {
		rt.handleError(w, "updateLoadBalancer", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	loadBalancer, err := rt.cache.UpdateLegacyLoadBalancer(r.Context(), rt.writer, lbID, updateInput)
	if err != nil {
		if err.Error() == fmt.Errorf(cache.ErrPortalAppNotFound.Error(), lbID).Error() {
			rt.handleError(w, "updateLoadBalancer", http.StatusNotFound, err)
			return
		}

		rt.handleError(w, "updateLoadBalancer", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, loadBalancer)
}

func (rt *Router) updateFirstDateSurpassed(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - updateFirstDateSurpassed")

	var updateInput v1Types.UpdateFirstDateSurpassed

	err := json.NewDecoder(r.Body).Decode(&updateInput)
	if err != nil {
		rt.handleError(w, "updateFirstDateSurpassed", http.StatusBadRequest, err)
		return
	}

	defer r.Body.Close()

	appsToUpdate, err := rt.cache.UpdateLegacyApplicationsFirstDateSurpassed(r.Context(), rt.writer, updateInput)
	if err != nil {
		rt.handleError(w, "updateFirstDateSurpassed", http.StatusBadRequest, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, appsToUpdate)
}

func (rt *Router) updateLoadBalancerUserRole(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - updateLoadBalancerUserRole")

	vars := mux.Vars(r)
	lbID := vars["id"]

	var updateUser v1Types.UpdateUserAccess

	err := json.NewDecoder(r.Body).Decode(&updateUser)
	if err != nil {
		rt.handleError(w, "updateLoadBalancerUserRole", http.StatusBadRequest, err)
		return
	}

	defer r.Body.Close()

	updatedLB, err := rt.cache.UpdateLegacyUserRole(r.Context(), rt.writer, lbID, updateUser)
	if err != nil {
		rt.handleError(w, "updateLoadBalancerUserRole", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedLB)
}

func (rt *Router) acceptLoadBalancerUser(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - acceptLoadBalancerUser")

	vars := mux.Vars(r)
	lbID := vars["id"]

	var acceptUserUpdate v1Types.UpdateUserAccess

	err := json.NewDecoder(r.Body).Decode(&acceptUserUpdate)
	if err != nil {
		rt.handleError(w, "acceptLoadBalancerUser", http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	updatedLB, err := rt.cache.AcceptLegacyUser(r.Context(), rt.writer, lbID, acceptUserUpdate)
	if err != nil {
		rt.handleError(w, "acceptLoadBalancerUser", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedLB)
}

func (rt *Router) deleteLoadBalancerUser(w http.ResponseWriter, r *http.Request) {
	rt.logger.Info("V1 endpoint hit - deleteLoadBalancerUser")

	vars := mux.Vars(r)

	lb, err := rt.cache.RemoveLegacyAccountUser(r.Context(), rt.writer, vars["userID"], vars["id"])
	if err != nil {
		rt.handleError(w, "deleteLoadBalancerUser", http.StatusInternalServerError, err)
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, lb)
}
