package router

import (
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

	rt.Router.HandleFunc("/", rt.HealthCheck).Methods(http.MethodGet)
	rt.Router.HandleFunc("/reset_cache", rt.ResetCache).Methods(http.MethodPost)

	// TODO update Blockchain create/update endpoints to /v2
	// rt.Router.HandleFunc("/v1/blockchain", rt.CreateBlockchain).Methods(http.MethodPost)
	// rt.Router.HandleFunc("/v1/blockchain/{id}", rt.UpdateBlockchain).Methods(http.MethodPut)
	// rt.Router.HandleFunc("/v1/blockchain/{id}/activate", rt.ActivateBlockchain).Methods(http.MethodPost)
	// rt.Router.HandleFunc("/v1/blockchain/redirect", rt.CreateRedirect).Methods(http.MethodPost)

	rt.Router.HandleFunc("/v1/blockchain", rt.GetBlockchains).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/blockchain/{id}", rt.GetBlockchain).Methods(http.MethodGet)

	rt.Router.HandleFunc("/v1/application", rt.GetApplications).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/application/first_date_surpassed", rt.UpdateFirstDateSurpassed).Methods(http.MethodPost)

	rt.Router.HandleFunc("/v1/load_balancer", rt.GetLoadBalancers).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/load_balancer", rt.CreateLoadBalancer).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v1/load_balancer/{id}", rt.GetLoadBalancer).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/load_balancer/{id}", rt.UpdateLoadBalancer).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v1/load_balancer/{id}/user", rt.CreateLoadBalancerUser).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v1/load_balancer/{id}/integration", rt.CreateLoadBalancerIntegration).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v1/load_balancer/{id}/user", rt.UpdateLoadBalancerUserRole).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v1/load_balancer/{id}/user/accept", rt.AcceptLoadBalancerUser).Methods(http.MethodPut)
	rt.Router.HandleFunc("/v1/load_balancer/{id}/user/{userID}", rt.DeleteLoadBalancerUser).Methods(http.MethodDelete) // TODO update to use userID not email

	rt.Router.HandleFunc("/v1/user", rt.CreatePortalUser).Methods(http.MethodPost)
	rt.Router.HandleFunc("/v1/user/{id}/portal_id", rt.GetPortalUserID).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/user/{id}/application", rt.GetApplicationsByUserID).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/user/{id}/load_balancer", rt.GetLoadBalancersByUserID).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/user/{id}/permission", rt.GetUserPermissionsByUserID).Methods(http.MethodGet)

	rt.Router.HandleFunc("/v1/user/{portalID}/load_balancer/pending", rt.GetPendingLoadBalancersByPortalID).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/user/{portalID}/load_balancer/count", rt.GetLoadBalancersCountByPortalID).Methods(http.MethodGet)

	rt.Router.HandleFunc("/v1/pay_plan", rt.GetPayPlans).Methods(http.MethodGet)
	rt.Router.HandleFunc("/v1/pay_plan/{type}", rt.GetPayPlan).Methods(http.MethodGet)

	rt.Router.Use(rt.AuthorizationHandler)

	return rt, nil
}

func (rt *Router) AuthorizationHandler(h http.Handler) http.Handler {
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

func (rt *Router) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Pocket HTTP DB is up and running!"))
	if err != nil {
		panic(err)
	}
}

func (rt *Router) ResetCache(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	err := rt.cache.ResetCache(r.Context())
	elapsedTime := time.Since(startTime).Milliseconds()
	if err != nil {
		rt.logError(fmt.Errorf("ResetCache failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := fmt.Sprintf("cache manually reset successfully in %dms", elapsedTime)
	_, err = w.Write([]byte(resp))
	if err != nil {
		rt.logError(fmt.Errorf("ResetCache failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

/* ---------- Read Methods ---------- */

func (rt *Router) GetUserPermissionsByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	permissions, err := rt.cache.GetLegacyUserPermissionsByUserID(r.Context(), userID)
	if err != nil {
		rt.logError(fmt.Errorf("GetUserPermissionsByUserID failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	if permissions.UserID == "" {
		rt.logError(fmt.Errorf("GetUserPermissionsByUserID failed: %w", errPermissionsNotFound))
		jsonresponse.RespondWithError(w, http.StatusNotFound, errPermissionsNotFound.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, permissions)
}

func (rt *Router) GetPortalUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	portalUserID := rt.cache.GetUserIDFromProviderUserID(types.ProviderUserID(vars["id"]))

	jsonresponse.RespondWithJSON(w, http.StatusOK, portalUserID)
}

func (rt *Router) GetApplicationsByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	apps, err := rt.cache.GetLegacyApplicationsByUserID(r.Context(), vars["id"])
	if err != nil {
		rt.logError(fmt.Errorf("GetApplicationsByUserID failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, apps)
}

func (rt *Router) GetApplications(w http.ResponseWriter, r *http.Request) {
	applications := rt.cache.GetLegacyApplications(cache.Options{
		IncludeDeleted: false,
	})

	jsonresponse.RespondWithJSON(w, http.StatusOK, applications)
}

func (rt *Router) GetLoadBalancersByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	roleNameFilter := v1Types.RoleName(r.URL.Query().Get("filter"))
	if roleNameFilter != v1Types.RoleName("") && !v1Types.ValidRoleNames[roleNameFilter] {
		jsonresponse.RespondWithError(w, http.StatusBadRequest, errInvalidRoleNameFilter.Error())
		return
	}

	lbs, err := rt.cache.GetLegacyLoadBalancersByUserID(r.Context(), userID, roleNameFilter)
	if err != nil {
		rt.logError(fmt.Errorf("GetLoadBalancersByUserID failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, lbs)
}

func (rt *Router) GetLoadBalancers(w http.ResponseWriter, r *http.Request) {
	jsonresponse.RespondWithJSON(w, http.StatusOK, rt.cache.GetLegacyLoadBalancers(r.Context()))
}

func (rt *Router) GetLoadBalancer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	lb, err := rt.cache.GetLegacyLoadBalancer(r.Context(), vars["id"])
	if err != nil {
		rt.logError(fmt.Errorf("GetLoadBalancer failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, lb)
}

func (rt *Router) GetBlockchains(w http.ResponseWriter, r *http.Request) {
	jsonresponse.RespondWithJSON(w, http.StatusOK, rt.cache.GetLegacyBlockchains(r.Context()))
}

func (rt *Router) GetBlockchain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	blockchain := rt.cache.GetLegacyBlockchain(r.Context(), vars["id"])
	if blockchain.ID == "" {
		rt.logError(fmt.Errorf("GetBlockchain failed: %w", errBlockchainNotFound))
		jsonresponse.RespondWithError(w, http.StatusNotFound, errBlockchainNotFound.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, blockchain)
}

func (rt *Router) GetPayPlans(w http.ResponseWriter, r *http.Request) {
	jsonresponse.RespondWithJSON(w, http.StatusOK, rt.cache.GetLegacyPayPlans(r.Context()))
}

func (rt *Router) GetPayPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	plan, err := rt.cache.GetLegacyPayPlan(r.Context(), vars["type"])
	if err != nil {
		rt.logError(fmt.Errorf("GetLegacyPayPlan inGetPayPlan failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, plan)
}

/* Returns all pending load balancers for a given portal user ID. */
func (rt *Router) GetPendingLoadBalancersByPortalID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userPortalID := vars["portalID"]

	lbs, err := rt.cache.GetLegacyPendingLoadBalancers(r.Context(), userPortalID)
	if err != nil {
		rt.logError(fmt.Errorf("GetLegacyPendingLoadBalancers in GetPendingLoadBalancersByEmail failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, lbs)
}

/* Returns the number of load balancers a given user owns, by portal ID */
func (rt *Router) GetLoadBalancersCountByPortalID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userPortalID := vars["portalID"]

	lbsCount := rt.cache.GetLegacyLoadBalancersCount(r.Context(), userPortalID)

	jsonresponse.RespondWithJSON(w, http.StatusOK, lbsCount)
}

/* ---------- Create Methods ---------- */

func (rt *Router) CreateLoadBalancer(w http.ResponseWriter, r *http.Request) {
	var lb v1Types.LoadBalancer

	err := json.NewDecoder(r.Body).Decode(&lb)
	if err != nil {
		rt.logError(fmt.Errorf("CreateLoadBalancer Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	createdLB, err := rt.cache.CreateLegacyLoadBalancer(r.Context(), rt.writer, lb)
	if err != nil {
		rt.logError(fmt.Errorf("CreateLoadBalancer Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, createdLB)
}

func (rt *Router) CreateLoadBalancerUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	lbID := vars["id"]

	var user v1Types.UserAccess

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		rt.logError(fmt.Errorf("CreateLoadBalancerUser Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	updatedLB, err := rt.cache.CreateLegacyLoadBalancerUser(r.Context(), rt.writer, lbID, user)
	if err != nil {
		rt.logError(fmt.Errorf("WriteLoadBalancerUser in CreateLoadBalancerUser failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedLB)
}

func (rt *Router) CreateLoadBalancerIntegration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	lbID := vars["id"]

	var integrationsInput types.AccountIntegrations

	err := json.NewDecoder(r.Body).Decode(&integrationsInput)
	if err != nil {
		rt.logError(fmt.Errorf("CreateLoadBalancerUser Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	updatedLB, err := rt.cache.CreateLegacyLoadBalancerIntegration(r.Context(), rt.writer, lbID, integrationsInput)
	if err != nil {
		rt.logError(fmt.Errorf("WriteLoadBalancerUser in CreateLoadBalancerUser failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedLB)
}

// func (rt *Router) CreateBlockchain(w http.ResponseWriter, r *http.Request) {
// 	var blockchain v1Types.Blockchain

// 	err := json.NewDecoder(r.Body).Decode(&blockchain)
// 	if err != nil {
// 		rt.logError(fmt.Errorf("CreateBlockchain decode failed: %w", err))
// 		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
// 		return
// 	}
// 	defer r.Body.Close()

// 	fullBlockchain, err := rt.cache.CreateLegacyBlockchain(r.Context(), rt.writer, blockchain)
// 	if err != nil {
// 		rt.logError(fmt.Errorf("CreateLegacyBlockchain in CreateBlockchain failed: %w", err))
// 		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	jsonresponse.RespondWithJSON(w, http.StatusOK, fullBlockchain)
// }

// func (rt *Router) CreateRedirect(w http.ResponseWriter, r *http.Request) {
// 	var redirect v1Types.Redirect

// 	err := json.NewDecoder(r.Body).Decode(&redirect)
// 	if err != nil {
// 		rt.logError(fmt.Errorf("CreateRedirect decode failed: %w", err))
// 		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
// 		return
// 	}

// 	defer r.Body.Close()

// 	fullRedirect, err := rt.cache.CreateLegacyRedirect(r.Context(), rt.writer, redirect)
// 	if err != nil {
// 		rt.logError(fmt.Errorf("CreateLegacyRedirect in CreateRedirect failed: %w", err))
// 		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	jsonresponse.RespondWithJSON(w, http.StatusOK, fullRedirect)
// }

func (rt *Router) CreatePortalUser(w http.ResponseWriter, r *http.Request) {
	var userInput types.CreateUser

	err := json.NewDecoder(r.Body).Decode(&userInput)
	if err != nil {
		rt.logError(fmt.Errorf("CreatePortalUser Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	newUser, accountID, err := rt.writer.WriteUserNewSignUp(r.Context(), userInput, time.Now())
	if err != nil {
		rt.logError(fmt.Errorf("WriteUserNewSignUp failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	createPortalUserResponse := types.CreateUserResponse{
		User:      *newUser,
		AccountID: accountID,
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, createPortalUserResponse)
}

/* ---------- Update Methods ---------- */

func (rt *Router) UpdateLoadBalancer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	lbID := vars["id"]

	// Note: It is intended to use UpdateApplication here
	var updateInput v1Types.UpdateApplication

	err := json.NewDecoder(r.Body).Decode(&updateInput)
	if err != nil {
		rt.logError(fmt.Errorf("UpdateLoadBalancer decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	loadBalancer, err := rt.cache.UpdateLegacyLoadBalancer(r.Context(), rt.writer, lbID, updateInput)
	if err != nil {
		rt.logError(fmt.Errorf("UpdateLegacyLoadBalancer in UpdateLoadBalancer failed: %w", err))
		if err.Error() == fmt.Errorf(cache.ErrPortalAppNotFound.Error(), lbID).Error() { // prevent a 500 when it should be a 404
			jsonresponse.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, loadBalancer)
}

func (rt *Router) UpdateFirstDateSurpassed(w http.ResponseWriter, r *http.Request) {
	rt.log.Info("Received UpdateFirstDateSurpassed request")

	var updateInput types.UpdateFirstDateSurpassed

	err := json.NewDecoder(r.Body).Decode(&updateInput)
	if err != nil {
		rt.logError(fmt.Errorf("UpdateFirstDateSurpassed decode failed: %w", err))
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
		rt.logError(fmt.Errorf("UpdateLegacyApplicationsFirstDateSurpassed in UpdateFirstDateSurpassed failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, appsToUpdate)
}

func (rt *Router) UpdateLoadBalancerUserRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lbID := vars["id"]

	var updateUser v1Types.UpdateUserAccess

	err := json.NewDecoder(r.Body).Decode(&updateUser)
	if err != nil {
		rt.logError(fmt.Errorf("UpdateLoadBalancerUser Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	updatedLB, err := rt.cache.UpdateLegacyUserRole(r.Context(), rt.writer, lbID, updateUser)
	if err != nil {
		rt.logError(fmt.Errorf("UpdateLegacyUserRole for team member in UpdateLoadBalancerUser failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedLB)
}

func (rt *Router) AcceptLoadBalancerUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lbID := vars["id"]

	var acceptUserUpdate v1Types.UpdateUserAccess

	err := json.NewDecoder(r.Body).Decode(&acceptUserUpdate)
	if err != nil {
		rt.logError(fmt.Errorf("AcceptLoadBalancerUser Decode failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	updatedLB, err := rt.cache.AcceptLegacyUser(r.Context(), rt.writer, lbID, acceptUserUpdate)
	if err != nil {
		rt.logError(fmt.Errorf("AcceptLegacyUser in AcceptLoadBalancerUser failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedLB)
}

func (rt *Router) DeleteLoadBalancerUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	lb, err := rt.cache.RemoveLegacyAccountUser(r.Context(), rt.writer, vars["userID"], vars["id"])
	if err != nil {
		rt.logError(fmt.Errorf("RemoveLegacyAccountUser in DeleteLoadBalancerUser failed: %w", err))
		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonresponse.RespondWithJSON(w, http.StatusOK, lb)
}

// func (rt *Router) UpdateBlockchain(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	blockchainID := vars["id"]

// 	var update v1Types.UpdateBlockchain

// 	err := json.NewDecoder(r.Body).Decode(&update)
// 	if err != nil {
// 		rt.logError(fmt.Errorf("UpdateBlockchain decode failed: %w", err))
// 		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
// 		return
// 	}
// 	defer r.Body.Close()

// 	updatedChain, err := rt.cache.UpdateLegacyBlockchain(r.Context(), rt.writer, blockchainID, update)
// 	if err != nil {
// 		rt.logError(fmt.Errorf("UpdateLegacyBlockchain in UpdateBlockchain failed: %w", err))
// 		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	jsonresponse.RespondWithJSON(w, http.StatusOK, updatedChain)
// }

// func (rt *Router) ActivateBlockchain(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	blockchainID := vars["id"]

// 	var active bool

// 	err := json.NewDecoder(r.Body).Decode(&active)
// 	if err != nil {
// 		rt.logError(fmt.Errorf("ActivateBlockchain decode failed: %w", err))
// 		jsonresponse.RespondWithError(w, http.StatusBadRequest, err.Error())
// 		return
// 	}
// 	defer r.Body.Close()

// 	activated, err := rt.cache.ActivateLegacyBlockchain(r.Context(), rt.writer, blockchainID, active)
// 	if err != nil {
// 		rt.logError(fmt.Errorf("ActivateLegacyBlockchain in ActivateBlockchain failed: %w", err))
// 		jsonresponse.RespondWithError(w, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	jsonresponse.RespondWithJSON(w, http.StatusOK, activated)
// }
