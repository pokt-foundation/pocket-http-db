package router

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/pokt-foundation/portal-db/v2/types"
	"github.com/pokt-foundation/portal-http-db/cache"
	"github.com/stretchr/testify/require"
)

func Test_paramProcessor(t *testing.T) {
	expectedParams := map[queryParam]queryParamDetail{
		commonParams.includeDeleted:      {required: false},
		chainParams.includeInactive:      {required: false},
		chainParams.excludeGigastakeApps: {required: false},
		portalAppParams.roleNameFilters:  {required: false},
	}

	tests := []struct {
		name          string
		params        url.Values
		path          string
		expected      cache.Options
		expectedError string
	}{
		{
			name: "Should process valid params correctly for common endpoint",
			params: url.Values{
				string(commonParams.includeDeleted): []string{"true"},
			},
			path: "/common_endpoint",
			expected: cache.Options{
				IncludeDeleted: true,
			},
		},
		{
			name: "Should process valid params correctly for chain endpoint",
			params: url.Values{
				string(chainParams.includeInactive):      []string{"true"},
				string(chainParams.excludeGigastakeApps): []string{"false"},
			},
			path: v2ChainEndpoint,
			expected: cache.Options{
				IncludeInactiveChains:     true,
				ExcludeChainGigastakeApps: false,
			},
		},
		{
			name: "Should process valid params correctly for chain endpoint with include deleted",
			params: url.Values{
				string(chainParams.includeInactive):      []string{"true"},
				string(chainParams.excludeGigastakeApps): []string{"false"},
				string(commonParams.includeDeleted):      []string{"true"},
			},
			path: v2ChainEndpoint,
			expected: cache.Options{
				IncludeInactiveChains:     true,
				ExcludeChainGigastakeApps: false,
				IncludeDeleted:            true,
			},
		},
		{
			name: "Should process valid params correctly for portal_app endpoint",
			params: url.Values{
				string(portalAppParams.roleNameFilters): []string{"OWNER,ADMIN"},
			},
			path: v2PortalAppEndpoint,
			expected: cache.Options{
				FilterPortalApps: []types.RoleName{types.RoleOwner, types.RoleAdmin},
			},
		},
		{
			name: "Should return error for invalid boolean param in common endpoint",
			params: url.Values{
				string(commonParams.includeDeleted): []string{"not-a-bool"},
			},
			path:          "/common_endpoint",
			expectedError: "failed to parse bool value",
		},
		{
			name: "Should return error for invalid param for common endpoint",
			params: url.Values{
				string(chainParams.includeInactive): []string{"true"},
			},
			path:          "/common_endpoint",
			expectedError: "invalid parameter",
		},
		{
			name: "Should return error for invalid boolean param",
			params: url.Values{
				string(chainParams.includeInactive): []string{"not-a-bool"},
			},
			path:          v2ChainEndpoint,
			expectedError: "failed to parse bool value",
		},
		{
			name: "Should return error for invalid param for path",
			params: url.Values{
				string(portalAppParams.roleNameFilters): []string{"role1,role2"},
			},
			path:          v2ChainEndpoint,
			expectedError: "invalid parameter",
		},
		{
			name: "Should return error for invalid param for different path",
			params: url.Values{
				string(chainParams.includeInactive): []string{"true"},
			},
			path:          v2PortalAppEndpoint,
			expectedError: fmt.Sprintf("invalid parameter %s for path %s", chainParams.includeInactive, v2PortalAppEndpoint),
		},
		{
			name: "Should return error for parameter not expected for the path",
			params: url.Values{
				string(chainParams.includeInactive): []string{"true"},
			},
			path:          "/some_unexpected_endpoint",
			expectedError: "invalid parameter include_inactive for path /some_unexpected_endpoint",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			urlTest := &url.URL{
				Scheme:   "https",
				Host:     "test.phd.com",
				RawQuery: test.params.Encode(),
				Path:     test.path,
			}
			result, err := paramProcessor(urlTest, expectedParams)

			if test.expectedError != "" {
				c.Error(err)
				c.Contains(err.Error(), test.expectedError)
			} else {
				c.NoError(err)
				c.Equal(test.expected, result)
			}
		})
	}
}

func Test_pathMatcher(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedRes string
	}{
		{
			name:        "Should match v2ChainEndpoint",
			input:       "/v2/chain/someextrapath",
			expectedRes: v2ChainEndpoint,
		},
		{
			name:        "Should match v1BlockchainEndpoint",
			input:       "/v1/blockchain/someextrapath",
			expectedRes: v1BlockchainEndpoint,
		},
		{
			name:        "Should match v2PortalAppEndpoint",
			input:       "/v2/portal_app/someextrapath",
			expectedRes: v2PortalAppEndpoint,
		},
		{
			name:        "Should match user portal apps endpoint",
			input:       "/v2/user/user_id_123/portal_app",
			expectedRes: v2PortalAppEndpoint,
		},
		{
			name:        "Should return empty for unmatched path",
			input:       "/v3/unmatchedpath",
			expectedRes: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			c.Equal(test.expectedRes, pathMatcher(test.input))
		})
	}
}

func Test_QueryParamMethods(t *testing.T) {
	tests := []struct {
		name            string
		param           queryParam
		expectCommon    bool
		expectChain     bool
		expectPortalApp bool
	}{
		{
			name:            "Test includeDeleted param",
			param:           commonParams.includeDeleted,
			expectCommon:    true,
			expectChain:     false,
			expectPortalApp: false,
		},
		{
			name:            "Test includeInactive param",
			param:           chainParams.includeInactive,
			expectCommon:    false,
			expectChain:     true,
			expectPortalApp: false,
		},
		{
			name:            "Test excludeGigastakeApps param",
			param:           chainParams.excludeGigastakeApps,
			expectCommon:    false,
			expectChain:     true,
			expectPortalApp: false,
		},
		{
			name:            "Test roleNameFilters param",
			param:           portalAppParams.roleNameFilters,
			expectCommon:    false,
			expectChain:     false,
			expectPortalApp: true,
		},
		{
			name:            "Test unknown param",
			param:           "unknown_param",
			expectCommon:    false,
			expectChain:     false,
			expectPortalApp: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			c.Equal(test.expectCommon, test.param.isCommon())
			c.Equal(test.expectChain, test.param.isChain())
			c.Equal(test.expectPortalApp, test.param.isPortalApp())
		})
	}
}

func Test_parseBool(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expected     bool
		expectingErr bool
	}{
		{
			name:         "Should correctly parse 'true'",
			input:        "true",
			expected:     true,
			expectingErr: false,
		},
		{
			name:         "Should correctly parse 'false'",
			input:        "false",
			expected:     false,
			expectingErr: false,
		},
		{
			name:         "Should return an error for invalid boolean input",
			input:        "not_bool",
			expectingErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			result, err := parseBool(test.input)

			if test.expectingErr {
				c.Error(err)
			} else {
				c.NoError(err)
				c.Equal(test.expected, result)
			}
		})
	}
}

func Test_parseRoleNameList(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expected     []types.RoleName
		expectingErr bool
	}{
		{
			name:         "Should parse multiple valid role names",
			input:        "OWNER,ADMIN",
			expected:     []types.RoleName{types.RoleOwner, types.RoleAdmin},
			expectingErr: false,
		},
		{
			name:         "Should parse single valid role name",
			input:        "OWNER",
			expected:     []types.RoleName{types.RoleOwner},
			expectingErr: false,
		},
		{
			name:         "Should not parse invalid role name",
			input:        "OWNER,GUEST",
			expectingErr: true,
		},
		{
			name:         "Should parse empty role name",
			input:        "",
			expected:     []types.RoleName{},
			expectingErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			result, err := parseRoleNameList(test.input)

			if test.expectingErr {
				c.Error(err)
			} else {
				c.NoError(err)
				c.Equal(test.expected, result)
			}
		})
	}
}
