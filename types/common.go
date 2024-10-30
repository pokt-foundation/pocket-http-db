package types

import "regexp"

/* String ID types */
type (
	AccountID        string
	GCPAccountID     string
	GCPEntitlementID string
	UserID           string
	GigastakeAppID   string
	RelayChainID     string
	PortalAppID      string
	ProtocolID       string
	ProtocolAppID    string
	Email            string
	BlockedAddress   string

	Environment string
)

const (
	EnvProduction Environment = "production"
	EnvCanary     Environment = "canary"
	EnvStaging    Environment = "staging"
	EnvTest       Environment = "test"
	EnvLocal      Environment = "local"
)

func (e Environment) IsValid() bool {
	switch e {
	case EnvProduction, EnvCanary, EnvStaging, EnvTest, EnvLocal:
		return true
	default:
		return false
	}
}

const (
	ProtocolMorseMainnet   ProtocolID = "morse-mainnet"
	ProtocolShannonMainnet ProtocolID = "shannon-mainnet"
	ProtocolShannonTestnet ProtocolID = "shannon-testnet"
)

func (id ProtocolID) IsValid() bool {
	switch id {
	case ProtocolMorseMainnet, ProtocolShannonMainnet, ProtocolShannonTestnet:
		return true
	default:
		return false
	}
}

// Validates that an Email fits a valid email format eg. test@example.com
func (e Email) IsValid() bool {
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(emailPattern)
	return regex.MatchString(string(e))
}

/* Config Structs */

// Provides options for passing to Driver interface methods
type DriverOptions struct {
	IncludeDeleted bool
}

/* Listener types */
type (
	Table  string
	Action string

	Notification struct {
		Table  Table
		Action Action
		Data   SavedOnDB
	}
)

const (
	TablePortalApps    Table = "portal_applications"
	TableAppSettings   Table = "portal_application_settings"
	TableAppWhitelists Table = "portal_application_whitelists"
	TablePortalAppAATs Table = "portal_application_aats"

	TableGigastakeApps Table = "gigastake_applications"

	TableAccounts             Table = "accounts"
	TableAccountUsers         Table = "account_users"
	TableAccountIntegrations  Table = "account_integrations"
	TableAccountNotifications Table = "account_notifications"

	TableUsers             Table = "users"
	TableUserAuthProviders Table = "user_auth_providers"

	TablePayPlans Table = "pay_plans"

	TableChains         Table = "chains"
	TableChainAltruists Table = "chain_altruists"
	TableChainChecks    Table = "chain_checks"
	TableChainAliases   Table = "chain_aliases"
	// DEPRECATED - TODO remove when move to only store aliases is complete
	TableChainAliasDomains Table = "chain_alias_domains"

	TableChainGigastakeApps Table = "chains_gigastake_applications"

	TableGlobalBlockedContracts Table = "global_blocked_contracts"

	TableChainConfig      Table = "chain_config"
	TableMethodValidators Table = "method_validators"

	TableGatewayConfig Table = "config_gateway"

	ActionInsert Action = "INSERT"
	ActionUpdate Action = "UPDATE"
	ActionDelete Action = "DELETE"
)

type SavedOnDB interface {
	Table() Table
}
