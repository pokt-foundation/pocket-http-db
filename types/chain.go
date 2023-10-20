package types

import (
	"net/url"
	"regexp"
	"strings"
	"time"
)

/* Enums */
type (
	ChainAuthType  string
	ChainCheckType string

	AltruistURL string
	ChainAlias  string
	ChainDomain string
)

const (
	ChainAuthTypeBasicAuth   ChainAuthType = "basic_auth"
	ChainAuthTypeBearerToken ChainAuthType = "bearer_token"
	ChainAuthTypeNone        ChainAuthType = "none"

	ChainCheckTypeArchival ChainCheckType = "archival"
	ChainCheckTypeChain    ChainCheckType = "chain"
	ChainCheckTypeMerge    ChainCheckType = "merge"
	ChainCheckTypeSync     ChainCheckType = "sync"
)

func (c ChainAuthType) IsValid() bool {
	switch c {
	case ChainAuthTypeBasicAuth, ChainAuthTypeBearerToken, ChainAuthTypeNone:
		return true
	default:
		return false
	}
}

func (c ChainCheckType) IsValid() bool {
	switch c {
	case ChainCheckTypeArchival, ChainCheckTypeChain, ChainCheckTypeMerge, ChainCheckTypeSync:
		return true
	default:
		return false
	}
}

func (a AltruistURL) IsValid() bool {
	parsedURL, err := url.Parse(string(a))
	if err != nil {
		return false
	}
	if parsedURL.Scheme == "http" || parsedURL.Scheme == "https" {
		return true
	}
	return false
}

func (d ChainDomain) IsValid() bool {
	pattern := `^([a-zA-Z0-9-*]+\.)+[a-zA-Z0-9-*]{2,}\.[a-zA-Z]{2,}(:[0-9]{1,5})?$`
	matched, err := regexp.MatchString(pattern, string(d))
	if err != nil {
		return false
	}
	return matched
}

func (d ChainDomain) IsPublicEndpoint() bool {
	domainComponents := strings.Split(string(d), ".")
	if len(domainComponents) < 1 {
		return false
	}
	// Public endpoints end with -rpc
	return strings.HasSuffix(domainComponents[0], "-rpc")
}

func (d ChainDomain) IsWildcardDomain() bool {
	// Wildcard domains are private
	return strings.Contains(string(d), "*")
}

func (d ChainDomain) GetAlias() ChainAlias {
	domainComponents := strings.Split(string(d), ".")
	if len(domainComponents) < 1 {
		return ""
	}

	return ChainAlias(domainComponents[0])
}

/* Chain Struct and Methods */
type (
	Chain struct {
		ID             RelayChainID             `json:"id"`
		Blockchain     ChainAlias               `json:"blockchain"`
		Description    string                   `json:"description"`
		IconURL        string                   `json:"iconURL"`
		EnforceResult  string                   `json:"enforceResult"`
		Path           string                   `json:"path"`
		Ticker         string                   `json:"ticker"`
		AllowedMethods []string                 `json:"allowedMethods"`
		LogLimitBlocks int32                    `json:"logLimitBlocks"`
		RequestTimeout int32                    `json:"requestTimeout"`
		Active         bool                     `json:"active"`
		Altruists      map[AltruistURL]Altruist `json:"altruists,omitempty"`
		Checks         map[ChainCheckType]Check `json:"chainChecks,omitempty"`
		Aliases        map[ChainAlias]struct{}  `json:"chainAliases,omitempty"`
		// DEPRECATED - TODO remove when move to only store aliases is complete
		AliasDomains map[ChainAlias][]ChainDomain `json:"domains,omitempty"`
		CreatedAt    time.Time                    `json:"createdAt"`
		UpdatedAt    time.Time                    `json:"updatedAt"`
		Deleted      bool                         `json:"deleted"`

		// GigastakeApps are set inside PHD
		GigastakeApps map[GigastakeAppID]*GigastakeApp `json:"chainGigastakeApps,omitempty"`
	}
	Altruist struct {
		ChainID  RelayChainID  `json:"chainID,omitempty"`
		URL      AltruistURL   `json:"url"`
		Auth     string        `json:"auth"`
		AuthType ChainAuthType `json:"authType"`
	}
	Check struct {
		ChainID    RelayChainID   `json:"chainID,omitempty"`
		Type       ChainCheckType `json:"type"`
		Payload    string         `json:"payload"`
		ResultKey  string         `json:"resultKey"`
		Allowance  int32          `json:"allowance"`
		EVMChainID int32          `json:"evmChainID"`
	}

	// Used for mapping listener notification
	Alias struct {
		ChainID RelayChainID `json:"chainID,omitempty"`
		Alias   ChainAlias   `json:"alias"`
	}
	// DEPRECATED - TODO remove when move to only store aliases is complete
	AliasDomains struct {
		ChainID RelayChainID  `json:"chainID,omitempty"`
		Alias   ChainAlias    `json:"alias"`
		Domains []ChainDomain `json:"domains"`
	}

	// NewChainInput is used for creating a new Chain, including its Gigastake App
	NewChainInput struct {
		Chain         *Chain          `json:"chain"`
		GigastakeApps []*GigastakeApp `json:"gigastakeApps"`
	}

	UpdateChain struct {
		ID             RelayChainID              `json:"id"`
		IconURL        *string                   `json:"iconURL"`
		Blockchain     *ChainAlias               `json:"blockchain"`
		Description    *string                   `json:"description"`
		EnforceResult  *string                   `json:"enforceResult"`
		Path           *string                   `json:"path"`
		Ticker         *string                   `json:"ticker"`
		AllowedMethods []string                  `json:"allowedMethods"`
		LogLimitBlocks *int32                    `json:"logLimitBlocks"`
		RequestTimeout *int32                    `json:"requestTimeout"`
		Active         *bool                     `json:"active"`
		Aliases        *map[ChainAlias]struct{}  `json:"chainAliases,omitempty"`
		Altruists      *map[AltruistURL]Altruist `json:"altruists,omitempty"`
		Checks         *map[ChainCheckType]Check `json:"chainChecks,omitempty"`
		// DEPRECATED - TODO remove when move to only store aliases is complete
		AliasDomains *map[ChainAlias][]ChainDomain `json:"domains,omitempty"`
	}
)

// GetChainAltruists returns a slice of all of a Chain's altruists
func (c *Chain) GetChainAltruists() []Altruist {
	altruists := []Altruist{}

	for _, altruist := range c.Altruists {
		altruists = append(altruists, altruist)
	}

	return altruists
}

// IsEVM returns a simple bool if the Chain has a Check with type Chain
// Only EVM chains may have a Check with type Chain
func (c *Chain) IsEVM() bool {
	_, ok := c.Checks[ChainCheckTypeChain]
	return ok
}

// Get ChainCheck returns a single Chain Check by its type
func (c *Chain) GetChainCheck(checkType ChainCheckType) Check {
	return c.Checks[checkType]
}

// GetChainAliases returns a slice of all of a Chain's aliases
func (c *Chain) GetChainAliases() []ChainAlias {
	aliasesSlice := []ChainAlias{}

	for alias := range c.Aliases {
		aliasesSlice = append(aliasesSlice, alias)
	}

	return aliasesSlice
}

// DEPRECATED - TODO remove when move to only store aliases is complete
// LegacyGetChainAliases returns a slice of all of a Chain's aliases
func (c *Chain) LegacyGetChainAliases() []ChainAlias {
	chainAliases := []ChainAlias{}

	for chainAlias := range c.AliasDomains {
		chainAliases = append(chainAliases, chainAlias)
	}

	return chainAliases
}

// GetChainDomains returns a slice of all of a Chain's domains
func (c *Chain) GetChainDomains() []ChainDomain {
	chainDomains := []ChainDomain{}

	for _, chainAliasDomains := range c.AliasDomains {
		chainDomains = append(chainDomains, chainAliasDomains...)
	}

	return chainDomains
}

// GetGigastakeAATs returns a slice of all of a Chain's GigastakeApps
func (c *Chain) GetGigastakeApps() []GigastakeApp {
	gigastakeAppsSlice := []GigastakeApp{}

	for _, gigastakeApp := range c.GigastakeApps {
		gigastakeAppsSlice = append(gigastakeAppsSlice, *gigastakeApp)
	}

	return gigastakeAppsSlice
}

// ClearGigastakeApps clears the Chain's GigastakesApps map and returns the updated Chain
// Used in the PHD cache when cache.Options.ExcludeGigastakeApps is true
func (c Chain) ClearGigastakeApps() Chain {
	c.GigastakeApps = nil
	return c
}

func (c *Chain) Table() Table {
	return TableChains
}

func (c *Altruist) Table() Table {
	return TableChainAltruists
}

func (c *Check) Table() Table {
	return TableChainChecks
}

func (c *Alias) Table() Table {
	return TableChainAliases
}

// DEPRECATED - TODO remove when move to only store aliases is complete
func (c *AliasDomains) Table() Table {
	return TableChainAliasDomains
}
