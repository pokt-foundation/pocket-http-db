package dbclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pokt-foundation/portal-db/v2/types"
)

type (
	// DBClient struct contains all the possible methods to interact
	// with the Portal HTTP DB amd satisfies the IDBClient interface
	DBClient struct {
		IDBClient
		httpClient *http.Client
		config     Config
	}
	// Config struct to provide config options
	Config struct {
		BaseURL, APIKey string
		Retries         int
		Timeout         time.Duration
	}
	retryTransport struct {
		underlying http.RoundTripper
		retries    int
	}

	// IDBClient interface contains all read & write methods to interact with the Portal HTTP DB
	IDBClient interface {
		IDBReader
		IDBWriter
	}

	// IDBReader interface contains read-only methods for interacting with the Portal HTTP DB
	IDBReader interface {
		// GetChainByID returns a single Chain by its relay chain ID - GET `/v2/chain/{id}`
		GetChainByID(ctx context.Context, chainID types.RelayChainID) (*types.Chain, error)
		// GetGigastakeAppByID returns a single GigastakeApp by its GigastakeAppID - GET `/v2/gigastake/{id}`
		GetGigastakeAppByID(ctx context.Context, gigastakeAppID types.GigastakeAppID) (*types.GigastakeApp, error)
		// GetAllChains returns all chains - GET `/v2/chain`
		GetAllChains(ctx context.Context, options ...ChainOptions) ([]*types.Chain, error)
		// GetAllGigastakeApps returns all GigastakeApps - GET `/v2/gigastake`
		GetAllGigastakeApps(ctx context.Context, optionParams ...GigastakeAppOptions) ([]*types.GigastakeApp, error)
		// GetAllGigastakeAppsByChain returns all GigastakeApps for a single chain ID - GET `/v2/chain/{id}/gigastake`
		GetAllGigastakeAppsByChain(ctx context.Context, chainID types.RelayChainID) ([]*types.GigastakeApp, error)

		// GetPortalAppByID returns a single Portal App by its ID - GET `/v2/portal_app/{id}`
		GetPortalAppByID(ctx context.Context, portalAppID types.PortalAppID) (*types.PortalApp, error)
		// GetAllPortalApps returns all Portal Apps - GET `/v2/portal_app`
		GetAllPortalApps(ctx context.Context, options ...PortalAppOptions) ([]*types.PortalApp, error)
		// GetPortalAppsByUser fetches all portal applications - GET `/v2/user/{userID}/portal_app`
		GetPortalAppsByUser(ctx context.Context, userID types.UserID, options ...PortalAppOptions) ([]*types.PortalApp, error)

		// GetPortalAppsForMiddleware returns all Portal Apps - GET `/v2/middleware/portal_app`
		GetPortalAppsForMiddleware(ctx context.Context) ([]*types.PortalAppLite, error)

		// GetAllAccounts returns all Accounts - GET `/v2/account`
		GetAllAccounts(ctx context.Context, options ...AccountOptions) ([]*types.Account, error)
		// GetUserAccounts returns all Accounts for a given user ID - GET `/v2/user/{userID}/account`
		GetUserAccounts(ctx context.Context, userID types.UserID, options ...AccountOptions) ([]*types.Account, error)
		// GetUserAccount returns a single user Account by its account ID and user ID - GET `/v2/user/{userID}/account/{id}`
		GetUserAccount(ctx context.Context, accountID types.AccountID, userID types.UserID, options ...AccountOptions) (*types.Account, error)

		// GetPortalUser returns the Portal User for a given user ID, either provider ID or portal ID - GET `/v2/user/{userID}?full_details=true`
		// The userID is a plain string because you can provide the method with either a provider user ID or a portal user ID
		GetPortalUser(ctx context.Context, userID string) (*types.User, error)
		// GetPortalUserID returns the Portal User for a given user ID, either provider ID or portal ID - GET `/v2/user/{userID}`
		// The userID is a plain string because you can provide the method with either a provider user ID or a portal user ID
		GetPortalUserID(ctx context.Context, userID string) (types.UserID, error)

		// GetAllPlans returns all plans - GET `/v2/plan`
		GetAllPlans(ctx context.Context) ([]types.Plan, error)

		// GetBlockedContracts returns all blocked contracts - GET `/v2/blocked_contract`
		GetBlockedContracts(ctx context.Context) (types.GlobalBlockedContracts, error)
	}

	// IDBWriter interface contains write methods for interacting with the Portal HTTP DB
	IDBWriter interface {
		// CreateChainAndGigastakeApps creates a new blockchain and its Gigastake apps in the DB - POST `/v2/chain`
		CreateChainAndGigastakeApps(ctx context.Context, newChainInput types.NewChainInput) (*types.NewChainInput, error)
		// CreateGigastakeApp creates a new Gigastake app in the DB - POST `/v2/chain/gigastake`
		CreateGigastakeApp(ctx context.Context, gigastakeAppInput types.GigastakeApp) (*types.GigastakeApp, error)
		// UpdateChain updates an existing blockchain in the DB - PUT `/v2/chain/{id}`
		UpdateChain(ctx context.Context, chainUpdate types.UpdateChain) (*types.Chain, error)
		// UpdateGigastakeApp updates a Gigastake app in the DB - PUT `/v2/chain/gigastake/{id}`
		UpdateGigastakeApp(ctx context.Context, id types.GigastakeAppID, updateGigastakeApp types.UpdateGigastakeApp) (*types.UpdateGigastakeApp, error)
		// ActivateChain activates or deactivates a blockchain by ID in the DB - PUT `/v2/chain/{id}/activate`
		ActivateChain(ctx context.Context, chainID types.RelayChainID, active bool) (bool, error)

		// CreatePortalApp creates a new Portal App - POST `/v2/portal_app`
		CreatePortalApp(ctx context.Context, portalAppInput types.PortalApp) (*types.PortalApp, error)
		// UpdatePortalApp updates an existing Portal App - PUT `/v2/portal_app/{id}`
		UpdatePortalApp(ctx context.Context, portalAppUpdate types.UpdatePortalApp) (*types.UpdatePortalApp, error)
		// DeletePortalApp deletes a Portal App - DELETE `/v2/portal_app/{id}`
		DeletePortalApp(ctx context.Context, portalAppID types.PortalAppID) (map[string]string, error)
		// UpdatePortalAppsFirstDateSurpassed updates the FirstDateSurpassed field of one or more Portal Apps - POST `/v2/portal_app/first_date_surpassed`
		UpdatePortalAppsFirstDateSurpassed(ctx context.Context, firstDateSurpassedUpdate types.UpdateFirstDateSurpassed) (map[string]string, error)

		// CreateAccount creates a new Account in the database for a single user - POST `/v2/user/{userID}/account`
		CreateAccount(ctx context.Context, userID types.UserID, account types.Account, timestamp time.Time) (*types.Account, error)
		// UpdateAccount updates an existing account in the DB - PUT `/v2/account/{id}`
		UpdateAccount(ctx context.Context, account types.UpdateAccount) (*types.Account, error)
		// CreateAccountIntegration creates a new integration for an account - POST `/v2/account/{id}/integration`
		CreateAccountIntegration(ctx context.Context, accountID types.AccountID, integration types.AccountIntegrations) (*types.AccountIntegrations, error)
		// UpdateAccountIntegration updates an existing integration for an account - PUT `/v2/account/{id}/integration`
		UpdateAccountIntegration(ctx context.Context, accountID types.AccountID, integration types.AccountIntegrations) (*types.AccountIntegrations, error)
		// DeleteAccount deletes an account in the DB - DELETE `/v2/account/{id}`
		DeleteAccount(ctx context.Context, accountID types.AccountID) (map[string]string, error)

		// WriteAccountUser creates a single Account User - POST `/v2/account/user`
		WriteAccountUser(ctx context.Context, createUser types.CreateAccountUserAccess, time time.Time) (map[string]types.UserID, error)
		// SetAccountUserRole updates the role for a single Account User - PUT `/v2/account/user/update_role`
		SetAccountUserRole(ctx context.Context, updateUser types.UpdateAccountUserRole, time time.Time) (map[string]string, error)
		// UpdateAcceptAccountUser accepts or declines an Account User Access - PUT `/v2/account/user/accept`
		UpdateAcceptAccountUser(ctx context.Context, acceptUser types.UpdateAcceptAccountUser, time time.Time) (map[string]string, error)
		// RemoveAccountUser removes an Account User's Role - PUT `/v2/account/user/remove`
		RemoveAccountUser(ctx context.Context, removeUser types.UpdateRemoveAccountUser) (map[string]string, error)

		// CreateUser creates a new User in the database - POST `/v2/user`
		CreateUser(ctx context.Context, user types.CreateUser) (*types.CreateUserResponse, error)
		// UpdateUser updates an existing User in the database - PUT `/v2/user`
		UpdateUser(ctx context.Context, user types.UpdateUser) (*types.User, error)
		// DeleteUser deletes a User - DELETE `/v2/user/{userID}`
		DeleteUser(ctx context.Context, userID types.UserID) (map[string]string, error)

		// WriteBlockedContract adds a new blocked address to the global blocked contracts - POST `/v2/blocked_contract`
		WriteBlockedContract(ctx context.Context, blockedContract types.BlockedContract) (map[string]string, error)
		// UpdateBlockedContractActive updates the active status of a blocked contract - PUT `/v2/blocked_contract/{address}/active`
		UpdateBlockedContractActive(ctx context.Context, address types.BlockedAddress, isActive bool) (map[string]bool, error)
		// RemoveBlockedContract deletes a blocked address from the global blocked contracts - DELETE `/v2/blocked_contract/{address}`
		RemoveBlockedContract(ctx context.Context, address types.BlockedAddress) (map[string]string, error)
	}

	basePath   string
	subPath    string
	APIVersion string

	QueryParam        string
	commonQueryParams struct {
		includeDeleted QueryParam
	}
	chainQueryParams struct {
		includeInactive      QueryParam
		excludeGigastakeApps QueryParam
	}
	portalAppQueryParams struct {
		RoleNameFilters QueryParam
		Accepted        QueryParam
	}
	accountQueryParams struct {
		RoleNameFilters QueryParam
		Accepted        QueryParam
	}

	ChainOptions struct {
		ExcludeGigastakeApps *bool
		IncludeInactive      *bool
		IncludeDeleted       *bool
	}
	GigastakeAppOptions struct {
		IncludeDeleted *bool
	}
	PortalAppOptions struct {
		RoleNameFilters []types.RoleName
		IncludeDeleted  *bool
		Accepted        *bool
	}
	AccountOptions struct {
		RoleNameFilters []types.RoleName
		IncludeDeleted  *bool
		Accepted        *bool
	}
)

func BoolPtr(b bool) *bool {
	return &b
}

const (
	chainPath           basePath = "chain"
	portalAppPath       basePath = "portal_app"
	accountPath         basePath = "account"
	userPath            basePath = "user"
	planPath            basePath = "plan"
	blockedContractPath basePath = "blocked_contract"
	middlewarePath      basePath = "middleware"

	gigastakePath      subPath = "gigastake"
	activatePath       subPath = "activate"
	integrationSubPath subPath = "integration"
	updateRoleSubPath  subPath = "update_role"
	acceptSubPath      subPath = "accept"
	removeSubPath      subPath = "remove"
	permissionPath     subPath = "permission"
	activePath         subPath = "active"
)

var (
	commonParams = commonQueryParams{
		includeDeleted: "include_deleted",
	}
	ChainParams = chainQueryParams{
		includeInactive:      "include_inactive",
		excludeGigastakeApps: "exclude_gigastake_apps",
	}
	PortalAppParams = portalAppQueryParams{
		RoleNameFilters: "filters",
		Accepted:        "accepted",
	}
	AccountParams = accountQueryParams{
		RoleNameFilters: "filters",
		Accepted:        "accepted",
	}

	errBaseURLNotProvided error = errors.New("base URL not provided")
	errAPIKeyNotProvided  error = errors.New("API key not provided")

	errNoUserID           error = errors.New("no user ID")
	errNoChainID          error = errors.New("no chain ID")
	errNoGigastakeAppID   error = errors.New("no gigastake app ID")
	errNoPortalAppID      error = errors.New("no portal app ID")
	errNoAccountID        error = errors.New("no account ID")
	errNoRoleName         error = errors.New("no role name")
	errNoAuthProviderType error = errors.New("no auth provider type")
	errNoProviderUserID   error = errors.New("no provider user ID")
	errNoEmail            error = errors.New("no email")
	errNoPlanTypeSet      error = errors.New("no plan type set")
	errNoBlockedAddress   error = errors.New("no blocked address provided")

	errInvalidRoleName                     error = errors.New("invalid role name filter provided")
	errInvalidPortalAppJSON                error = errors.New("invalid portal app JSON")
	errInvalidFirstDateSurpassedUpdateJSON error = errors.New("invalid first date surpassed update JSON")
	errInvalidAccountJSON                  error = errors.New("invalid account JSON")
	errInvalidAccountIntegrationJSON       error = errors.New("invalid account integration JSON")
	errInvalidAppJSON                      error = errors.New("invalid application JSON")
	errInvalidChainJSON                    error = errors.New("invalid chain JSON")
	errInvalidGigastakeAppJSON             error = errors.New("invalid gigastake app JSON")
	errInvalidActiveStatusJSON             error = errors.New("invalid active status JSON")
	errInvalidBlockedContractJSON          error = errors.New("invalid blocked contract JSON")

	errMoreThanOneOption error = errors.New("may not provider more than one options parameter")

	errResponseNotOK error = errors.New("Response not OK")
)

// NewDBClient returns a read-write HTTP client to use the Portal HTTP DB - https://github.com/pokt-foundation/portal-http-db
func NewDBClient(config Config) (IDBClient, error) {
	if err := config.validateConfig(); err != nil {
		return nil, err
	}

	return &DBClient{httpClient: newHTTPClient(config), config: config}, nil
}

// NewReadOnlyDBClient returns a read-only HTTP client to use the Portal HTTP DB - https://github.com/pokt-foundation/portal-http-db
func NewReadOnlyDBClient(config Config) (IDBReader, error) {
	if err := config.validateConfig(); err != nil {
		return nil, err
	}

	return &DBClient{httpClient: newHTTPClient(config), config: config}, nil
}

// validateConfig ensures that a valid configuration is provided to the DB client
func (c Config) validateConfig() error {
	if c.BaseURL == "" {
		return errBaseURLNotProvided
	}
	if c.APIKey == "" {
		return errAPIKeyNotProvided
	}
	return nil
}

// v2BasePath returns the /v2/ base path for a given data type eg. `https://portal.http-db-url.com/v2/chain`
func (db *DBClient) v2BasePath(dataTypePath basePath) string {
	return fmt.Sprintf("%s/v2/%s", db.config.BaseURL, dataTypePath)
}

// getAuthHeaderForRead returns the auth header for read operations to PHD
func (db *DBClient) getAuthHeaderForRead() http.Header {
	return http.Header{"Authorization": {db.config.APIKey}}
}

// getAuthHeaderForRead returns the auth header for write operations to PHD
func (db *DBClient) getAuthHeaderForWrite() http.Header {
	return http.Header{"Authorization": {db.config.APIKey}, "Content-Type": {"application/json"}}
}

/* ------------ IDBReader Methods ------------ */

/* -- Chain Read Methods -- */

// GetChainByID returns a single Chain by its relay chain ID - GET `/v2/chain/{id}`
func (db *DBClient) GetChainByID(ctx context.Context, chainID types.RelayChainID) (*types.Chain, error) {
	if chainID == "" {
		return nil, errNoChainID
	}

	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(chainPath), chainID)

	return getReq[*types.Chain](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

// GetGigastakeAppByID returns a single GigastakeApp by its GigastakeAppID - GET `/v2/gigastake/{id}`
func (db *DBClient) GetGigastakeAppByID(ctx context.Context, gigastakeAppID types.GigastakeAppID) (*types.GigastakeApp, error) {
	if gigastakeAppID == "" {
		return nil, errNoGigastakeAppID
	}

	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(basePath(gigastakePath)), gigastakeAppID)

	return getReq[*types.GigastakeApp](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

// GetAllChains returns all chains - GET `/v2/chain`
func (db *DBClient) GetAllChains(ctx context.Context, optionParams ...ChainOptions) ([]*types.Chain, error) {
	endpoint := db.v2BasePath(chainPath)

	options := ChainOptions{}
	if len(optionParams) > 0 {
		options = optionParams[0]
	}

	queryParams := make([]string, 0)
	if options.IncludeInactive != nil {
		queryParams = append(queryParams, fmt.Sprintf("%s=%t", ChainParams.includeInactive, *options.IncludeInactive))
	}
	if options.ExcludeGigastakeApps != nil {
		queryParams = append(queryParams, fmt.Sprintf("%s=%t", ChainParams.excludeGigastakeApps, *options.ExcludeGigastakeApps))
	}
	if options.IncludeDeleted != nil {
		queryParams = append(queryParams, fmt.Sprintf("%s=%t", commonParams.includeDeleted, *options.IncludeDeleted))
	}

	if len(queryParams) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(queryParams, "&"))
	}

	return getReq[[]*types.Chain](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

// GetAllGigastakeApps returns all GigastakeApps - GET `/v2/gigastake`
func (db *DBClient) GetAllGigastakeApps(ctx context.Context, optionParams ...GigastakeAppOptions) ([]*types.GigastakeApp, error) {
	endpoint := db.v2BasePath(basePath(gigastakePath))

	options := GigastakeAppOptions{}
	if len(optionParams) > 0 {
		options = optionParams[0]
	}

	queryParams := make([]string, 0)
	if options.IncludeDeleted != nil {
		queryParams = append(queryParams, fmt.Sprintf("%s=%t", commonParams.includeDeleted, *options.IncludeDeleted))
	}

	if len(queryParams) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(queryParams, "&"))
	}

	return getReq[[]*types.GigastakeApp](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

// GetAllGigastakeAppsByChain returns all GigastakeApps for a single chain ID - GET `/v2/chain/{id}/gigastake`
func (db *DBClient) GetAllGigastakeAppsByChain(ctx context.Context, chainID types.RelayChainID) ([]*types.GigastakeApp, error) {
	if chainID == "" {
		return nil, errNoChainID
	}

	endpoint := fmt.Sprintf("%s/%s/gigastake", db.v2BasePath(chainPath), chainID)

	return getReq[[]*types.GigastakeApp](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

/* -- Portal App Read Methods -- */

// GetPortalAppByID returns a single Portal App by its ID - GET `/v2/portal_app/{id}`
func (db *DBClient) GetPortalAppByID(ctx context.Context, portalAppID types.PortalAppID) (*types.PortalApp, error) {
	if portalAppID == "" {
		return nil, errNoPortalAppID
	}

	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(portalAppPath), portalAppID)

	return getReq[*types.PortalApp](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

// GetAllPortalApps returns all Portal Apps - GET `/v2/portal_app`
func (db *DBClient) GetAllPortalApps(ctx context.Context, optionParams ...PortalAppOptions) ([]*types.PortalApp, error) {
	if len(optionParams) > 1 {
		return nil, errMoreThanOneOption
	}

	endpoint := db.v2BasePath(portalAppPath)

	options := PortalAppOptions{}
	if len(optionParams) > 0 {
		options = optionParams[0]
	}

	queryParams := make([]string, 0)

	if options.IncludeDeleted != nil {
		queryParams = append(queryParams, fmt.Sprintf("%s=%t", commonParams.includeDeleted, *options.IncludeDeleted))
	}

	if len(queryParams) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(queryParams, "&"))
	}

	return getReq[[]*types.PortalApp](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

// GetPortalAppsByUser fetches all portal applications - GET `/v2/user/{userID}/portal_app`
func (db *DBClient) GetPortalAppsByUser(ctx context.Context, userID types.UserID, optionParams ...PortalAppOptions) ([]*types.PortalApp, error) {
	if userID == "" {
		return nil, errNoUserID
	}
	if len(optionParams) > 1 {
		return nil, errMoreThanOneOption
	}

	endpoint := fmt.Sprintf("%s/%s/%s", db.v2BasePath(userPath), userID, portalAppPath)

	options := PortalAppOptions{}
	if len(optionParams) > 0 {
		options = optionParams[0]
	}

	queryParams := make([]string, 0)

	if len(options.RoleNameFilters) > 0 {
		roleNameStrs := make([]string, len(options.RoleNameFilters))
		for i, roleName := range options.RoleNameFilters {
			if !roleName.IsValid() {
				return nil, fmt.Errorf("%w: %s", errInvalidRoleName, roleName)
			}
			roleNameStrs[i] = string(roleName)
		}
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", PortalAppParams.RoleNameFilters, strings.Join(roleNameStrs, ",")))
	}

	if options.IncludeDeleted != nil {
		queryParams = append(queryParams, fmt.Sprintf("%s=%t", commonParams.includeDeleted, *options.IncludeDeleted))
	}

	if options.Accepted != nil {
		queryParams = append(queryParams, fmt.Sprintf("%s=%t", PortalAppParams.Accepted, *options.Accepted))
	}

	if len(queryParams) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(queryParams, "&"))
	}

	return getReq[[]*types.PortalApp](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

// GetPortalAppsForMiddleware returns all Portal Apps - GET `/v2/middleware/portal_app`
func (db *DBClient) GetPortalAppsForMiddleware(ctx context.Context) ([]*types.PortalAppLite, error) {
	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(middlewarePath), portalAppPath)

	return getReq[[]*types.PortalAppLite](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

/* -- Account Read Methods -- */

// GetAllAccounts returns all Accounts - GET `/v2/account`
func (db *DBClient) GetAllAccounts(ctx context.Context, optionParams ...AccountOptions) ([]*types.Account, error) {
	if len(optionParams) > 1 {
		return nil, errMoreThanOneOption
	}

	endpoint := db.v2BasePath(accountPath)

	options := AccountOptions{}
	if len(optionParams) > 0 {
		options = optionParams[0]
	}

	queryParams := make([]string, 0)

	if options.IncludeDeleted != nil {
		queryParams = append(queryParams, fmt.Sprintf("%s=%t", commonParams.includeDeleted, *options.IncludeDeleted))
	}

	if len(queryParams) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(queryParams, "&"))
	}

	return getReq[[]*types.Account](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

// GetUserAccounts returns all Accounts for a given user ID - GET `/v2/user/{userID}/account`
func (db *DBClient) GetUserAccounts(ctx context.Context, userID types.UserID, optionParams ...AccountOptions) ([]*types.Account, error) {
	if userID == "" {
		return nil, errNoUserID
	}
	if len(optionParams) > 1 {
		return nil, errMoreThanOneOption
	}

	endpoint := fmt.Sprintf("%s/%s/%s", db.v2BasePath(userPath), userID, accountPath)

	options := AccountOptions{}
	if len(optionParams) > 0 {
		options = optionParams[0]
	}

	queryParams := make([]string, 0)

	if len(options.RoleNameFilters) > 0 {
		roleNameStrs := make([]string, len(options.RoleNameFilters))
		for i, roleName := range options.RoleNameFilters {
			if !roleName.IsValid() {
				return nil, fmt.Errorf("%w: %s", errInvalidRoleName, roleName)
			}
			roleNameStrs[i] = string(roleName)
		}
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", AccountParams.RoleNameFilters, strings.Join(roleNameStrs, ",")))
	}

	if options.Accepted != nil {
		queryParams = append(queryParams, fmt.Sprintf("%s=%t", AccountParams.Accepted, *options.Accepted))
	}

	if len(queryParams) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(queryParams, "&"))
	}

	return getReq[[]*types.Account](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

// GetUserAccount returns a single user Account by its account ID and user ID - GET `/v2/user/{userID}/account/{id}`
func (db *DBClient) GetUserAccount(ctx context.Context, accountID types.AccountID, userID types.UserID, optionParams ...AccountOptions) (*types.Account, error) {
	if accountID == "" {
		return nil, errNoAccountID
	}
	if userID == "" {
		return nil, errNoUserID
	}
	if len(optionParams) > 1 {
		return nil, errMoreThanOneOption
	}

	endpoint := fmt.Sprintf("%s/%s/%s/%s", db.v2BasePath(userPath), userID, accountPath, accountID)

	options := AccountOptions{}
	if len(optionParams) > 0 {
		options = optionParams[0]
	}

	queryParams := make([]string, 0)

	if len(options.RoleNameFilters) > 0 {
		roleNameStrs := make([]string, len(options.RoleNameFilters))
		for i, roleName := range options.RoleNameFilters {
			if !roleName.IsValid() {
				return nil, fmt.Errorf("%w: %s", errInvalidRoleName, roleName)
			}
			roleNameStrs[i] = string(roleName)
		}
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", AccountParams.RoleNameFilters, strings.Join(roleNameStrs, ",")))
	}

	if options.Accepted != nil {
		queryParams = append(queryParams, fmt.Sprintf("%s=%t", AccountParams.Accepted, *options.Accepted))
	}

	if len(queryParams) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(queryParams, "&"))
	}

	return getReq[*types.Account](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

/* -- User Read Methods -- */

// GetPortalUser returns the Portal User for a given user ID, either provider ID or portal ID - GET `/v2/user/{userID}?full_details=true`
// The userID is a plain string because you can provide the method with either a provider user ID or a portal user ID
func (db *DBClient) GetPortalUser(ctx context.Context, userID string) (*types.User, error) {
	if userID == "" {
		return &types.User{}, errNoUserID
	}

	endpoint := fmt.Sprintf("%s/%s?full_details=true", db.v2BasePath(userPath), userID)

	return getReq[*types.User](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

// GetPortalUserID returns the Portal User for a given user ID, either provider ID or portal ID - GET `/v2/user/{userID}`
// The userID is a plain string because you can provide the method with either a provider user ID or a portal user ID
func (db *DBClient) GetPortalUserID(ctx context.Context, userID string) (types.UserID, error) {
	if userID == "" {
		return types.UserID(""), errNoUserID
	}

	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(userPath), userID)

	return getReq[types.UserID](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

/* -- Plans Read Methods -- */

// GetAllPlans returns all plans - GET `/v2/plan`
func (db *DBClient) GetAllPlans(ctx context.Context) ([]types.Plan, error) {
	endpoint := db.v2BasePath(planPath)

	return getReq[[]types.Plan](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

/* -- Blocked Contracts Read Methods -- */

// GetBlockedContracts returns all blocked contracts - GET `/v2/blocked_contract`
func (db *DBClient) GetBlockedContracts(ctx context.Context) (types.GlobalBlockedContracts, error) {
	endpoint := db.v2BasePath(blockedContractPath)

	return getReq[types.GlobalBlockedContracts](endpoint, db.getAuthHeaderForRead(), db.httpClient)
}

/* ------------ IDBWriter Methods ------------ */

/* -- Chain Write Methods -- */

// CreateChainAndGigastakeApps creates a new blockchain and its Gigastake apps in the DB - POST `/v2/chain`
func (db *DBClient) CreateChainAndGigastakeApps(ctx context.Context, newChainInput types.NewChainInput) (*types.NewChainInput, error) {
	newChainInputJSON, err := json.Marshal(newChainInput)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidChainJSON, err)
	}

	endpoint := db.v2BasePath(chainPath)

	return postReq[*types.NewChainInput](endpoint, db.getAuthHeaderForWrite(), newChainInputJSON, db.httpClient)
}

// CreateGigastakeApp creates a new Gigastake app in the DB - POST `/v2/chain/gigastake`
func (db *DBClient) CreateGigastakeApp(ctx context.Context, gigastakeAppInput types.GigastakeApp) (*types.GigastakeApp, error) {
	gigastakeAppInputJSON, err := json.Marshal(gigastakeAppInput)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidGigastakeAppJSON, err)
	}

	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(chainPath), gigastakePath)

	return postReq[*types.GigastakeApp](endpoint, db.getAuthHeaderForWrite(), gigastakeAppInputJSON, db.httpClient)
}

// UpdateChain updates an existing blockchain in the DB - PUT `/v2/chain/{id}`
func (db *DBClient) UpdateChain(ctx context.Context, chainUpdate types.UpdateChain) (*types.Chain, error) {
	if chainUpdate.ID == "" {
		return nil, errNoChainID
	}

	chainUpdateJSON, err := json.Marshal(chainUpdate)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidChainJSON, err)
	}

	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(chainPath), chainUpdate.ID)

	return putReq[*types.Chain](endpoint, db.getAuthHeaderForWrite(), chainUpdateJSON, db.httpClient)
}

// UpdateGigastakeApp updates a Gigastake app in the DB - PUT `/v2/chain/gigastake/{id}`
func (db *DBClient) UpdateGigastakeApp(ctx context.Context, id types.GigastakeAppID, updateGigastakeApp types.UpdateGigastakeApp) (*types.UpdateGigastakeApp, error) {
	if id == "" {
		return nil, errNoGigastakeAppID
	}

	updateGigastakeAppJSON, err := json.Marshal(updateGigastakeApp)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidGigastakeAppJSON, err)
	}

	endpoint := fmt.Sprintf("%s/%s/%s", db.v2BasePath(chainPath), gigastakePath, id)

	return putReq[*types.UpdateGigastakeApp](endpoint, db.getAuthHeaderForWrite(), updateGigastakeAppJSON, db.httpClient)
}

// ActivateChain activates or deactivates a blockchain by ID in the DB - PUT `/v2/chain/{id}/activate`
func (db *DBClient) ActivateChain(ctx context.Context, chainID types.RelayChainID, active bool) (bool, error) {
	activeJSON, err := json.Marshal(active)
	if err != nil {
		return false, fmt.Errorf("%w: %s", errInvalidActiveStatusJSON, err)
	}

	endpoint := fmt.Sprintf("%s/%s/%s", db.v2BasePath(chainPath), chainID, activatePath)

	return putReq[bool](endpoint, db.getAuthHeaderForWrite(), activeJSON, db.httpClient)
}

/* -- Portal App Write Methods -- */

// CreatePortalApp creates a new Portal App - POST `/v2/portal_app`
func (db *DBClient) CreatePortalApp(ctx context.Context, portalAppInput types.PortalApp) (*types.PortalApp, error) {
	portalAppInputJSON, err := json.Marshal(portalAppInput)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidPortalAppJSON, err)
	}

	endpoint := db.v2BasePath(portalAppPath)

	return postReq[*types.PortalApp](endpoint, db.getAuthHeaderForWrite(), portalAppInputJSON, db.httpClient)
}

// UpdatePortalApp updates an existing Portal App - PUT `/v2/portal_app/{id}`
func (db *DBClient) UpdatePortalApp(ctx context.Context, portalAppUpdate types.UpdatePortalApp) (*types.UpdatePortalApp, error) {
	portalAppUpdateJSON, err := json.Marshal(portalAppUpdate)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidPortalAppJSON, err)
	}

	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(portalAppPath), portalAppUpdate.AppID)

	return putReq[*types.UpdatePortalApp](endpoint, db.getAuthHeaderForWrite(), portalAppUpdateJSON, db.httpClient)
}

// DeletePortalApp deletes a Portal App - DELETE `/v2/portal_app/{id}`
func (db *DBClient) DeletePortalApp(ctx context.Context, portalAppID types.PortalAppID) (map[string]string, error) {
	if portalAppID == "" {
		return nil, errNoPortalAppID
	}

	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(portalAppPath), portalAppID)

	return deleteReq[map[string]string](endpoint, db.getAuthHeaderForWrite(), db.httpClient)
}

// UpdatePortalAppsFirstDateSurpassed updates the FirstDateSurpassed field of one or more Portal Apps - POST `/v2/portal_app/first_date_surpassed`
func (db *DBClient) UpdatePortalAppsFirstDateSurpassed(ctx context.Context, firstDateSurpassedUpdate types.UpdateFirstDateSurpassed) (map[string]string, error) {
	firstDateSurpassedUpdateJSON, err := json.Marshal(firstDateSurpassedUpdate)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidFirstDateSurpassedUpdateJSON, err)
	}

	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(portalAppPath), "first_date_surpassed")

	return postReq[map[string]string](endpoint, db.getAuthHeaderForWrite(), firstDateSurpassedUpdateJSON, db.httpClient)
}

/* -- Account Write Methods -- */

// CreateAccount creates a new Account in the database for a single user - POST `/v2/user/{userID}/account`
func (db *DBClient) CreateAccount(ctx context.Context, userID types.UserID, account types.Account, timestamp time.Time) (*types.Account, error) {
	if userID == "" {
		return nil, errNoUserID
	}
	if account.PlanType == "" {
		return nil, errNoPlanTypeSet
	}

	accountJSON, err := json.Marshal(account)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidAccountJSON, err)
	}

	endpoint := fmt.Sprintf("%s/%s/%s", db.v2BasePath(userPath), userID, accountPath)

	return postReq[*types.Account](endpoint, db.getAuthHeaderForWrite(), accountJSON, db.httpClient)
}

// UpdateAccount updates an Account in the DB - PUT `/v2/account/{id}`
func (db *DBClient) UpdateAccount(ctx context.Context, account types.UpdateAccount) (*types.Account, error) {
	accountJSON, err := json.Marshal(account)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidAccountJSON, err)
	}

	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(accountPath), account.AccountID)

	return putReq[*types.Account](endpoint, db.getAuthHeaderForWrite(), accountJSON, db.httpClient)
}

// CreateAccountIntegration creates an AccountIntegration in the DB - POST `/v2/account/{id}/integration`
func (db *DBClient) CreateAccountIntegration(ctx context.Context, accountID types.AccountID, integration types.AccountIntegrations) (*types.AccountIntegrations, error) {
	integrationJSON, err := json.Marshal(integration)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidAccountIntegrationJSON, err)
	}

	endpoint := fmt.Sprintf("%s/%s/%s", db.v2BasePath(accountPath), accountID, integrationSubPath)

	return postReq[*types.AccountIntegrations](endpoint, db.getAuthHeaderForWrite(), integrationJSON, db.httpClient)
}

// UpdateAccountIntegration updates an AccountIntegration in the DB - PUT `/v2/account/{id}/integration`
func (db *DBClient) UpdateAccountIntegration(ctx context.Context, accountID types.AccountID, integration types.AccountIntegrations) (*types.AccountIntegrations, error) {
	integrationJSON, err := json.Marshal(integration)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidAccountIntegrationJSON, err)
	}

	endpoint := fmt.Sprintf("%s/%s/%s", db.v2BasePath(accountPath), accountID, integrationSubPath)

	return putReq[*types.AccountIntegrations](endpoint, db.getAuthHeaderForWrite(), integrationJSON, db.httpClient)
}

// DeleteAccount deletes an Account in the DB - DELETE `/v2/account/{id}`
func (db *DBClient) DeleteAccount(ctx context.Context, accountID types.AccountID) (map[string]string, error) {
	if accountID == "" {
		return nil, errNoAccountID
	}

	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(accountPath), accountID)

	return deleteReq[map[string]string](endpoint, db.getAuthHeaderForWrite(), db.httpClient)
}

/* -- Account User Write Methods -- */

// WriteAccountUser creates a single Account User - POST `/v2/account/user`
func (db *DBClient) WriteAccountUser(ctx context.Context, createUser types.CreateAccountUserAccess, time time.Time) (map[string]types.UserID, error) {
	if createUser.AccountID == "" {
		return nil, errNoAccountID
	}
	if createUser.PortalAppID == "" {
		return nil, errNoPortalAppID
	}
	if createUser.Email == "" {
		return nil, errNoEmail
	}
	if createUser.RoleName == "" {
		return nil, errNoRoleName
	}

	createUserJSON, err := json.Marshal(createUser)
	if err != nil {
		return nil, fmt.Errorf("invalid createUser JSON: %w", err)
	}

	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(accountPath), userPath)

	return postReq[map[string]types.UserID](endpoint, db.getAuthHeaderForWrite(), createUserJSON, db.httpClient)
}

// SetAccountUserRole updates the role for a single Account User - PUT `/v2/account/user/update_role`
func (db *DBClient) SetAccountUserRole(ctx context.Context, updateUser types.UpdateAccountUserRole, time time.Time) (map[string]string, error) {
	if updateUser.PortalAppID == "" {
		return nil, errNoPortalAppID
	}
	if updateUser.UserID == "" {
		return nil, errNoUserID
	}
	if updateUser.AccountID == "" {
		return nil, errNoAccountID
	}
	if updateUser.RoleName == "" {
		return nil, errNoRoleName
	}

	updateUserJSON, err := json.Marshal(updateUser)
	if err != nil {
		return nil, fmt.Errorf("invalid updateUser JSON: %w", err)
	}

	endpoint := fmt.Sprintf("%s/%s/%s", db.v2BasePath(accountPath), userPath, updateRoleSubPath)

	return putReq[map[string]string](endpoint, db.getAuthHeaderForWrite(), updateUserJSON, db.httpClient)
}

// UpdateAcceptAccountUser accepts or declines an Account User Access - PUT `/v2/account/user/accept`
func (db *DBClient) UpdateAcceptAccountUser(ctx context.Context, acceptUser types.UpdateAcceptAccountUser, time time.Time) (map[string]string, error) {
	if acceptUser.PortalAppID == "" {
		return nil, errNoPortalAppID
	}
	if acceptUser.UserID == "" {
		return nil, errNoUserID
	}
	if acceptUser.AuthProviderType == "" {
		return nil, errNoAuthProviderType
	}
	if acceptUser.ProviderUserID == "" {
		return nil, errNoProviderUserID
	}

	acceptUserJSON, err := json.Marshal(acceptUser)
	if err != nil {
		return nil, fmt.Errorf("invalid acceptUser JSON: %w", err)
	}

	endpoint := fmt.Sprintf("%s/%s/%s", db.v2BasePath(accountPath), userPath, acceptSubPath)

	return putReq[map[string]string](endpoint, db.getAuthHeaderForWrite(), acceptUserJSON, db.httpClient)
}

// RemoveAccountUser removes an Account User's Role - PUT `/v2/account/user/remove`
func (db *DBClient) RemoveAccountUser(ctx context.Context, removeUser types.UpdateRemoveAccountUser) (map[string]string, error) {
	if removeUser.PortalAppID == "" {
		return nil, errNoPortalAppID
	}
	if removeUser.UserID == "" {
		return nil, errNoUserID
	}
	if removeUser.AccountID == "" {
		return nil, errNoAccountID
	}

	removeUserJSON, err := json.Marshal(removeUser)
	if err != nil {
		return nil, fmt.Errorf("invalid removeUser JSON: %w", err)
	}

	endpoint := fmt.Sprintf("%s/%s/%s", db.v2BasePath(accountPath), userPath, removeSubPath)

	return putReq[map[string]string](endpoint, db.getAuthHeaderForWrite(), removeUserJSON, db.httpClient)
}

/* -- User Write Methods -- */

// CreateUser creates a new User in the database - POST `/v2/user`
func (db *DBClient) CreateUser(ctx context.Context, user types.CreateUser) (*types.CreateUserResponse, error) {
	if user.Email == "" {
		return nil, errNoEmail
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("invalid user JSON: %w", err)
	}

	endpoint := db.v2BasePath(userPath)

	return postReq[*types.CreateUserResponse](endpoint, db.getAuthHeaderForWrite(), userJSON, db.httpClient)
}

// UpdateUser updates an existing User in the database - PUT `/v2/user`
func (db *DBClient) UpdateUser(ctx context.Context, user types.UpdateUser) (*types.User, error) {
	if user.ID == "" {
		return nil, errNoUserID
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("invalid update user JSON: %w", err)
	}

	endpoint := db.v2BasePath(userPath)

	return putReq[*types.User](endpoint, db.getAuthHeaderForWrite(), userJSON, db.httpClient)
}

// DeleteUser deletes a User - DELETE `/v2/user/{userID}`
func (db *DBClient) DeleteUser(ctx context.Context, userID types.UserID) (map[string]string, error) {
	if userID == "" {
		return nil, errNoUserID
	}

	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(userPath), userID)

	return deleteReq[map[string]string](endpoint, db.getAuthHeaderForWrite(), db.httpClient)
}

/* -- Blocked Contracts Write Methods -- */

// WriteBlockedContract adds a new blocked address to the global blocked contracts - POST `/v2/blocked_contract`
func (db *DBClient) WriteBlockedContract(ctx context.Context, blockedContract types.BlockedContract) (map[string]string, error) {
	blockedContractJSON, err := json.Marshal(blockedContract)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidBlockedContractJSON, err)
	}

	endpoint := db.v2BasePath(blockedContractPath)

	return postReq[map[string]string](endpoint, db.getAuthHeaderForWrite(), blockedContractJSON, db.httpClient)
}

// UpdateBlockedContractActive updates the active status of a blocked contract - PUT `/v2/blocked_contract/{address}/active`
func (db *DBClient) UpdateBlockedContractActive(ctx context.Context, address types.BlockedAddress, isActive bool) (map[string]bool, error) {
	if address == "" {
		return nil, errNoBlockedAddress
	}

	activeStatus := struct {
		Active bool `json:"active"`
	}{
		Active: isActive,
	}

	activeStatusJSON, err := json.Marshal(activeStatus)
	if err != nil {
		return nil, errInvalidAppJSON
	}

	endpoint := fmt.Sprintf("%s/%s/%s", db.v2BasePath(blockedContractPath), address, activePath)

	return putReq[map[string]bool](endpoint, db.getAuthHeaderForWrite(), activeStatusJSON, db.httpClient)
}

// RemoveBlockedContract deletes a blocked address from the global blocked contracts - DELETE `/v2/blocked_contract/{address}`
func (db *DBClient) RemoveBlockedContract(ctx context.Context, address types.BlockedAddress) (map[string]string, error) {
	if address == "" {
		return nil, errNoBlockedAddress
	}

	endpoint := fmt.Sprintf("%s/%s", db.v2BasePath(blockedContractPath), address)

	return deleteReq[map[string]string](endpoint, db.getAuthHeaderForWrite(), db.httpClient)
}

/* ------------ PHD Client HTTP Funcs ------------ */

func newHTTPClient(config Config) *http.Client {
	return &http.Client{
		Timeout: config.Timeout,
		Transport: &retryTransport{
			underlying: http.DefaultTransport,
			retries:    config.Retries,
		},
	}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.underlying
	if rt == nil {
		rt = http.DefaultTransport
	}

	var resp *http.Response
	var err error

	// Cache request body
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}

	for i := 0; i <= t.retries; i++ {
		// Recreate body reader
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err = rt.RoundTrip(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}

		if i < t.retries {
			time.Sleep(time.Duration(i*i) * 100 * time.Millisecond)
		}
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Generic HTTP GET request
func getReq[T any](endpoint string, header http.Header, httpClient *http.Client) (T, error) {
	var data T

	// Create a new request
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return data, err
	}

	// Set headers
	req.Header = header

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return data, parseErrorResponse(resp)
	}

	// Decode response body
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return data, err
	}

	return data, nil
}

// Generic HTTP POST request
func postReq[T any](endpoint string, header http.Header, postData []byte, httpClient *http.Client) (T, error) {
	var data T

	// Create a new request
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(postData))
	if err != nil {
		return data, err
	}

	// Set headers
	req.Header = header

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return data, parseErrorResponse(resp)
	}

	// Decode response body
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return data, err
	}

	return data, nil
}

// Generic HTTP PUT request
func putReq[T any](endpoint string, header http.Header, putData []byte, httpClient *http.Client) (T, error) {
	var data T

	// Create a new request
	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(putData))
	if err != nil {
		return data, err
	}

	// Set headers
	req.Header = header

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return data, parseErrorResponse(resp)
	}

	// Decode response body
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return data, err
	}

	return data, nil
}

// Generic HTTP DELETE request
func deleteReq[T any](endpoint string, header http.Header, httpClient *http.Client) (T, error) {
	var data T

	// Create a new request
	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return data, err
	}

	// Set headers
	req.Header = header

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return data, parseErrorResponse(resp)
	}

	// Decode response body
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return data, err
	}

	return data, nil
}

// Parses the error reponse and returns the status code and error message
func parseErrorResponse(errResponse *http.Response) error {
	code := errResponse.StatusCode
	text := http.StatusText(code)

	errString := fmt.Errorf("%s. %d %s", errResponseNotOK, code, text)

	body, err := io.ReadAll(errResponse.Body)
	if err != nil {
		return errString
	}

	var errorMap map[string]string
	err = json.Unmarshal(body, &errorMap)
	if err != nil {
		return errString
	}

	if errorMessage, ok := errorMap["error"]; ok {
		errString = fmt.Errorf("%s: %s", errString, errorMessage)
	}

	return errString
}
