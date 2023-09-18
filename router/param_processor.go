package router

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pokt-foundation/pocket-http-db/cache"
	"github.com/pokt-foundation/portal-db/v2/types"
)

type (
	queryParamDetail struct {
		required bool
	}

	queryParam string

	// NOTE: Add new query params here
	commonQueryParams struct {
		includeDeleted queryParam
	}
	chainQueryParams struct {
		includeInactive      queryParam
		excludeGigastakeApps queryParam
	}
	portalAppQueryParams struct {
		roleNameFilters queryParam
	}
	accountQueryParams struct {
		roleNameFilters queryParam
	}
	userQueryParams struct {
		fullDetails queryParam
	}

	paramHandler func(v string, options *cache.Options) error
)

const (
	apiVersion2       = "/v2"
	chainEndpoint     = "/chain"
	portalAppEndpoint = "/portal_app"
	accountEndpoint   = "/account"
	userEndpoint      = "/user"
	// TODO remove when v1 endpoints are removed
	apiVersion1          = "/v1"
	blockchainEndpointV1 = "/blockchain"
)

var (
	v2ChainEndpoint     = fmt.Sprintf("%s%s", apiVersion2, chainEndpoint)
	v2PortalAppEndpoint = fmt.Sprintf("%s%s", apiVersion2, portalAppEndpoint)
	v2AccountEndpoint   = fmt.Sprintf("%s%s", apiVersion2, accountEndpoint)
	v2UserEndpoint      = fmt.Sprintf("%s%s", apiVersion2, userEndpoint)
	// TODO remove when v1 endpoints are removed
	v1BlockchainEndpoint = fmt.Sprintf("%s%s", apiVersion1, blockchainEndpointV1)

	commonParams = commonQueryParams{
		includeDeleted: "include_deleted",
	}
	chainParams = chainQueryParams{
		includeInactive:      "include_inactive",
		excludeGigastakeApps: "exclude_gigastake_apps",
	}
	portalAppParams = portalAppQueryParams{
		roleNameFilters: "filters",
	}
	accountParams = accountQueryParams{
		roleNameFilters: "filters",
	}
	userParams = userQueryParams{
		fullDetails: "full_details",
	}

	// NOTE: Add new parameter handlers here when new query param options added

	// Contains the mapping of URL endpoints to their corresponding query parameters and their handlers.
	// These handlers assign the relevant field from the query param to the Options struct.
	paramsMap = map[string]map[queryParam]paramHandler{
		"": {
			commonParams.includeDeleted: func(v string, options *cache.Options) error {
				val, err := parseBool(v)
				if err != nil {
					return err
				}
				options.IncludeDeleted = val
				return nil
			},
		},

		v2ChainEndpoint: {
			chainParams.excludeGigastakeApps: func(v string, options *cache.Options) error {
				val, err := parseBool(v)
				if err != nil {
					return err
				}
				options.ExcludeChainGigastakeApps = val
				return nil
			},

			chainParams.includeInactive: func(v string, options *cache.Options) error {
				val, err := parseBool(v)
				if err != nil {
					return err
				}
				options.IncludeInactiveChains = val
				return nil
			},
		},

		v2PortalAppEndpoint: {
			portalAppParams.roleNameFilters: func(v string, options *cache.Options) error {
				roleNameFilters, err := parseRoleNameList(v)
				if err != nil {
					return err
				}
				options.FilterPortalApps = roleNameFilters
				return nil
			},
		},

		v2AccountEndpoint: {
			accountParams.roleNameFilters: func(v string, options *cache.Options) error {
				roleNameFilters, err := parseRoleNameList(v)
				if err != nil {
					return err
				}
				options.FilterPortalApps = roleNameFilters
				return nil
			},
		},

		v2UserEndpoint: {
			userParams.fullDetails: func(v string, options *cache.Options) error {
				val, err := parseBool(v)
				if err != nil {
					return err
				}
				options.FullDetails = val
				return nil
			},
		},

		// TODO remove when v1 endpoints are removed
		v1BlockchainEndpoint: {
			chainParams.excludeGigastakeApps: func(v string, options *cache.Options) error {
				val, err := parseBool(v)
				if err != nil {
					return err
				}
				options.ExcludeChainGigastakeApps = val
				return nil
			},

			chainParams.includeInactive: func(v string, options *cache.Options) error {
				val, err := parseBool(v)
				if err != nil {
					return err
				}
				options.IncludeInactiveChains = val
				return nil
			},
		},
	}
)

// paramProcessor validates and processes the request query parameters based on the provided specifications.
// It returns a processed cache.Options object or an error if the parameters are invalid.
func paramProcessor(url *url.URL, queryParams map[queryParam]queryParamDetail) (cache.Options, error) {
	params := url.Query()

	result := make(map[queryParam]string)

	for param, detail := range queryParams {
		val, ok := params[string(param)]
		if !ok {
			if !detail.required {
				continue
			} else {
				return cache.Options{}, fmt.Errorf("required parameter %s not found", param)
			}
		}

		result[param] = val[0]
	}

	options := &cache.Options{}
	for param, value := range result {
		var setter func(string, *cache.Options) error
		var ok bool
		var path string

		switch {
		case param.isCommon():
			path = ""
		case param.isChain() || param.isPortalApp() || param.isUser() || param.isAccount():
			path = pathMatcher(url.Path)
		default:
			return cache.Options{}, fmt.Errorf("invalid parameter %s", param)
		}

		setter, ok = paramsMap[path][param]
		if !ok {
			return cache.Options{}, fmt.Errorf("invalid parameter %s for path %s", param, url.Path)
		}

		if err := setter(value, options); err != nil {
			return cache.Options{}, fmt.Errorf("failed to set parameter %s: %w", param, err)
		}
	}

	return *options, nil
}

// pathMatcher takes a path as input and matches it with predefined endpoints, returning the matching endpoint as a string.
func pathMatcher(path string) string {
	if strings.HasPrefix(path, v2ChainEndpoint) {
		return v2ChainEndpoint
	}
	if strings.HasPrefix(path, v2PortalAppEndpoint) || strings.Contains(path, portalAppEndpoint) {
		return v2PortalAppEndpoint
	}
	if strings.HasPrefix(path, v2AccountEndpoint) || strings.Contains(path, accountEndpoint) {
		return v2AccountEndpoint
	}
	if strings.HasPrefix(path, v2UserEndpoint) || strings.Contains(path, v2UserEndpoint) {
		return v2UserEndpoint
	}
	// TODO remove when v1 endpoints are removed
	if strings.HasPrefix(path, v1BlockchainEndpoint) {
		return v1BlockchainEndpoint
	}
	return ""
}

func (qp queryParam) isCommon() bool {
	return qp == commonParams.includeDeleted
}

func (qp queryParam) isChain() bool {
	return qp == chainParams.includeInactive || qp == chainParams.excludeGigastakeApps
}

func (qp queryParam) isPortalApp() bool {
	return qp == portalAppParams.roleNameFilters
}

func (qp queryParam) isAccount() bool {
	return qp == accountParams.roleNameFilters
}

func (qp queryParam) isUser() bool {
	return qp == userParams.fullDetails
}

func parseBool(v string) (bool, error) {
	val, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("failed to parse bool value: %w", err)
	}
	return val, nil
}

func parseRoleNameList(v string) ([]types.RoleName, error) {
	if v == "" {
		return []types.RoleName{}, nil
	}

	var roleNameFilters []types.RoleName

	for _, roleName := range strings.Split(v, ",") {
		roleName := types.RoleName(roleName)

		if !roleName.IsValid() {
			return nil, fmt.Errorf("invalid role name provided: %s", roleName)
		}

		roleNameFilters = append(roleNameFilters, roleName)
	}

	return roleNameFilters, nil
}
