package types

import (
	"sort"
	"strings"
	"time"
)

/* Enums */
type (
	Environment       string
	NotificationEvent string
	NotificationType  string
	WhitelistType     string

	PortalAppPublicKey string
	PortalAppOrigin    string

	AppEmoji string
)

const (
	EnvironmentProduction Environment = "production"
	EnvironmentTest       Environment = "test"

	NotificationEventFull          NotificationEvent = "full"
	NotificationEventHalf          NotificationEvent = "half"
	NotificationEventQuarter       NotificationEvent = "quarter"
	NotificationEventSignedUp      NotificationEvent = "signedUp"
	NotificationEventThreeQuarters NotificationEvent = "threeQuarters"

	NotificationTypeEmail   NotificationType = "email"
	NotificationTypePortal  NotificationType = "portal"
	NotificationTypeWebhook NotificationType = "webhook"

	WhitelistTypeOrigins     WhitelistType = "origins"
	WhitelistTypeBlockchains WhitelistType = "blockchains"
	WhitelistTypeUserAgents  WhitelistType = "userAgents"
	WhitelistTypeContracts   WhitelistType = "contracts"
	WhitelistTypeMethods     WhitelistType = "methods"
)

func (e Environment) IsValid() bool {
	switch e {
	case EnvironmentProduction, EnvironmentTest:
		return true
	default:
		return false
	}
}

func (n NotificationEvent) IsValid() bool {
	switch n {
	case NotificationEventFull, NotificationEventHalf, NotificationEventQuarter, NotificationEventSignedUp, NotificationEventThreeQuarters:
		return true
	default:
		return false
	}
}

func (n NotificationType) IsValid() bool {
	switch n {
	case NotificationTypeEmail, NotificationTypePortal, NotificationTypeWebhook:
		return true
	default:
		return false
	}
}

func (w WhitelistType) IsValid() bool {
	switch w {
	case WhitelistTypeBlockchains, WhitelistTypeContracts, WhitelistTypeMethods, WhitelistTypeOrigins, WhitelistTypeUserAgents:
		return true
	default:
		return false
	}
}

/* PortalApp Struct Definition and Methods */
type (
	// PortalApp represents a single application in the Portal
	PortalApp struct {
		ID                 PortalAppID                          `json:"id"`
		Name               string                               `json:"name"`
		Description        string                               `json:"description"`
		AppEmoji           AppEmoji                             `json:"appEmoji"`
		AccountID          AccountID                            `json:"accountID"`
		Settings           Settings                             `json:"settings"`
		Whitelists         Whitelists                           `json:"whitelists"`
		AATs               map[ProtocolAppID]AAT                `json:"aat"`
		Notifications      map[NotificationType]AppNotification `json:"notifications"`
		CreatedAt          time.Time                            `json:"createdAt"`
		UpdatedAt          time.Time                            `json:"updatedAt"`
		Deleted            bool                                 `json:"deleted"`
		FirstDateSurpassed time.Time                            `json:"firstDateSurpassed"`
		// TODO - remove when v2 migration finished
		// Fields required for compatibility with the old Portal API and Services (temporary)
		LegacyFields LegacyFields `json:"legacyFields"`

		// Users assigned in PHD from PortalApp's Account when PortalApp fetched from PUB
		PortalAppUsers map[UserID]PortalAppUser `json:"portalAppUsers"`
	}

	PortalAppUser struct {
		ID       UserID   `json:"id"`
		Email    Email    `json:"email"`
		RoleName RoleName `json:"roleName"`
		Accepted bool     `json:"accepted"`
	}

	// TODO - remove when v2 migration finished
	// Fields required for compatibility with the old Portal API and Services (temporary)
	LegacyFields struct {
		PlanType             PayPlanType `json:"planType"`
		DailyLimit           int32       `json:"dailyLimit"`
		CustomLimit          int32       `json:"customLimit"`
		RequestTimeout       int32       `json:"requestTimeout"`
		StripeSubscriptionID string      `json:"stripeSubscriptionID"`
	}

	Settings struct {
		AppID             PortalAppID               `json:"appID,omitempty"`
		Environment       Environment               `json:"environment"`
		SecretKey         string                    `json:"secretKey"`
		SecretKeyRequired bool                      `json:"secretKeyRequired"`
		FavoritedChainIDs map[RelayChainID]struct{} `json:"favoritedBlockchainIDs"`
		// MonthlyRelayLimit sets the monthly limit per-application
		// Sum of an Account's Apps MonthlyRelayLimits cannot exceed the Account's MonthlyRelayLimit
		MonthlyRelayLimit int32 `json:"monthlyRelayLimit"`
	}

	Whitelists struct {
		Origins     map[Origin]struct{}                    `json:"origins,omitempty"`
		UserAgents  map[UserAgent]struct{}                 `json:"userAgents,omitempty"`
		Blockchains map[RelayChainID]struct{}              `json:"blockchains,omitempty"`
		Contracts   map[RelayChainID]map[Contract]struct{} `json:"contracts,omitempty"`
		Methods     map[RelayChainID]map[Method]struct{}   `json:"methods,omitempty"`
	}

	// AAT contains the data needed to perform relays
	AAT struct {
		AppID           PortalAppID        `json:"appID,omitempty"`
		ID              ProtocolAppID      `json:"id"`
		PublicKey       PortalAppPublicKey `json:"publicKey"`
		Address         string             `json:"address"`
		ClientPublicKey string             `json:"clientPublicKey"`
		Signature       string             `json:"signature"`
		Version         string             `json:"version"`

		// PrivateKey used when read from the DB, will always be ""
		// Only used for saving to DB
		// TODO remove when decided to not support saving private key to DB
		PrivateKey string `json:"privateKey,omitempty"`
	}

	// Whitelist (singular) is used by the listener in PHD to receive a single whitelist row
	Whitelist struct {
		AppID   PortalAppID   `json:"applicationID"`
		Type    WhitelistType `json:"type"`
		Value   string        `json:"value"`
		ChainID RelayChainID  `json:"chain_id"`
	}

	AppNotification struct {
		AppID       PortalAppID                `json:"appID,omitempty"`
		Type        NotificationType           `json:"type"`
		Active      bool                       `json:"active"`
		Destination string                     `json:"destination"`
		Trigger     string                     `json:"trigger"`
		Events      map[NotificationEvent]bool `json:"events"`
	}

	// PortalAppLite contains only the PortalApp fields actually used in the Portal Middleware.
	PortalAppLite struct {
		ID         PortalAppID          `json:"id"`
		PublicKeys []PortalAppPublicKey `json:"publicKeys"`
		Settings   SettingsLite         `json:"settings,omitempty"`
		Whitelists Whitelists           `json:"whitelists,omitempty"`
		Plan       PlanLite             `json:"plan,omitempty"`
	}

	SettingsLite struct {
		SecretKey         string `json:"secretKey,omitempty"`
		SecretKeyRequired bool   `json:"secretKeyRequired,omitempty"`
	}

	PlanLite struct {
		PlanType        PayPlanType               `json:"planType,omitempty"`
		ChainIDs        map[RelayChainID]struct{} `json:"chainIDs,omitempty"`
		ThroughputLimit int32                     `json:"throughputLimit,omitempty"`
	}

	// WhitelistsObject is a GraphQL-compatible representation of all the whitelists for a given application (used for the Portal UI)
	// It is also used to update Whitelists for an app (sent from Portal UI to PUB to PHD)
	WhitelistsObject struct {
		AppWhitelists   [3]ApplicationWhitelists `json:"appWhitelists"`
		ChainWhitelists [2]ChainWhitelists       `json:"chainWhitelists"`
	}
	ApplicationWhitelists struct {
		Type   WhitelistType `json:"type"`
		Values []string      `json:"values"`
	}
	ChainWhitelists struct {
		Type   WhitelistType       `json:"type"`
		Values []ChainIDWhitelists `json:"values"`
	}
	ChainIDWhitelists struct {
		ChainID string   `json:"chainID"`
		Values  []string `json:"values"`
	}

	// UpdatePortalApp Struct Definition and Methods
	UpdatePortalApp struct {
		AppID         PortalAppID              `json:"appID,omitempty"`
		Name          string                   `json:"name,omitempty"`
		Description   string                   `json:"description,omitempty"`
		AppEmoji      AppEmoji                 `json:"appEmoji,omitempty"`
		Settings      *UpdateAppSettings       `json:"appSettings,omitempty"`
		Notifications []UpdateAppNotifications `json:"notificationSettings,omitempty"`
		Whitelists    *WhitelistsObject        `json:"whitelists,omitempty"`
		// TODO - remove when v2 migration finished
		PlanType             PayPlanType `json:"planType,omitempty"`
		CustomLimit          int32       `json:"customLimit,omitempty"`
		StripeSubscriptionID string      `json:"stripeSubscriptionID,omitempty"`
	}

	UpdateAppSettings struct {
		AppID             PortalAppID `json:"appID,omitempty"`
		Environment       Environment `json:"environment"`
		SecretKey         string      `json:"secretKey"`
		SecretKeyRequired bool        `json:"secretKeyRequired"`
		MonthlyRelayLimit int32       `json:"monthlyRelayLimit"`
		FavoritedChainIDs []string    `json:"favoritedChainIDs"`
	}

	UpdateAppNotifications struct {
		AppID            PortalAppID         `json:"appID,omitempty"`
		NotificationType NotificationType    `json:"notificationType"`
		Active           bool                `json:"active"`
		Destination      string              `json:"destination"`
		Trigger          string              `json:"trigger"`
		Events           []NotificationEvent `json:"events"`
	}

	UpdateFirstDateSurpassed struct {
		PortalAppIDs       []PortalAppID `json:"applicationIDs"`
		FirstDateSurpassed time.Time     `json:"firstDateSurpassed"`
	}

	Origin    string
	UserAgent string
	Method    string
	Contract  string
)

func (u *UpdatePortalApp) IsEmpty() bool {
	if u.Name != "" || u.PlanType != "" || u.Description != "" || u.AppEmoji != "" || u.CustomLimit != 0 || u.StripeSubscriptionID != "" {
		return false
	}
	return true
}

// TODO For Legacy Fields Only
// DailyLimit returns the daily relay limit for a given portal app
func (a *PortalApp) DailyLimit() int32 {
	if a.LegacyFields.PlanType == Enterprise {
		return a.LegacyFields.CustomLimit
	}
	return a.LegacyFields.DailyLimit
}

// TODO For Legacy Fields Only
// SubscriptionID returns the stripe subscription ID
func (a *PortalApp) SubscriptionID() string {
	return a.LegacyFields.StripeSubscriptionID
}

// MonthlyLimit returns the monthly relay limit for a given portal app
func (a *PortalApp) MonthlyLimit() int32 {
	return a.Settings.MonthlyRelayLimit
}

// GetUsers returns all the PortalApp's Users as a slice
func (a *PortalApp) GetUsers() []PortalAppUser {
	users := []PortalAppUser{}
	for _, user := range a.PortalAppUsers {
		users = append(users, user)
	}
	return users
}

// GetUserIDs returns all the PortalApp's Users' IDs as a slice
func (a *PortalApp) GetUserIDs() []UserID {
	users := []UserID{}
	for _, user := range a.PortalAppUsers {
		users = append(users, user.ID)
	}
	return users
}

// Owner ID returns the UserID of the PortalApp's owner
func (a *PortalApp) OwnerID() UserID {
	for _, user := range a.PortalAppUsers {
		if user.RoleName == RoleOwner {
			return user.ID
		}
	}
	return ""
}

// AssignUsers assigns the Account's AccountUsers to the PortalApp's PortalAppUsers field
// Only assigns users that have a role for the PortalApp
func (a *PortalApp) AssignUsers(account *Account) {
	if account.AccountUsers != nil && len(account.AccountUsers) > 0 {
		portalAppUsers := map[UserID]PortalAppUser{}

		for _, accountUser := range account.AccountUsers {
			if _, ok := accountUser.PortalAppRoles[a.ID]; ok {
				portalAppUsers[accountUser.UserID] = PortalAppUser{
					ID:       accountUser.UserID,
					Email:    accountUser.Email,
					RoleName: accountUser.PortalAppRoles[a.ID],
					Accepted: accountUser.PortalAppsAccepted[a.ID],
				}
			}
		}

		a.PortalAppUsers = portalAppUsers
	}
}

// AddWhitelist adds a whitelist to the PortalApp pointer's Whitelists field
func (a *PortalApp) AddWhitelist(whitelist Whitelist) {
	switch whitelist.Type {
	case WhitelistTypeOrigins:
		if a.Whitelists.Origins == nil {
			a.Whitelists.Origins = make(map[Origin]struct{})
		}
		a.Whitelists.Origins[Origin(whitelist.Value)] = struct{}{}
	case WhitelistTypeBlockchains:
		if a.Whitelists.Blockchains == nil {
			a.Whitelists.Blockchains = make(map[RelayChainID]struct{})
		}
		a.Whitelists.Blockchains[RelayChainID(whitelist.Value)] = struct{}{}
	case WhitelistTypeUserAgents:
		if a.Whitelists.UserAgents == nil {
			a.Whitelists.UserAgents = make(map[UserAgent]struct{})
		}
		a.Whitelists.UserAgents[UserAgent(whitelist.Value)] = struct{}{}
	case WhitelistTypeContracts:
		if a.Whitelists.Contracts == nil {
			a.Whitelists.Contracts = make(map[RelayChainID]map[Contract]struct{})
		}
		if _, ok := a.Whitelists.Contracts[whitelist.ChainID]; !ok {
			a.Whitelists.Contracts[whitelist.ChainID] = make(map[Contract]struct{})
		}
		a.Whitelists.Contracts[whitelist.ChainID][Contract(whitelist.Value)] = struct{}{}
	case WhitelistTypeMethods:
		if a.Whitelists.Methods == nil {
			a.Whitelists.Methods = make(map[RelayChainID]map[Method]struct{})
		}
		if _, ok := a.Whitelists.Methods[whitelist.ChainID]; !ok {
			a.Whitelists.Methods[whitelist.ChainID] = make(map[Method]struct{})
		}
		a.Whitelists.Methods[whitelist.ChainID][Method(whitelist.Value)] = struct{}{}
	}
}

// DeleteWhitelist deletes a whitelist from the PortalApp pointer's Whitelists field
func (a *PortalApp) DeleteWhitelist(whitelist Whitelist) {
	switch whitelist.Type {
	case WhitelistTypeOrigins:
		delete(a.Whitelists.Origins, Origin(whitelist.Value))
	case WhitelistTypeBlockchains:
		delete(a.Whitelists.Blockchains, RelayChainID(whitelist.Value))
	case WhitelistTypeUserAgents:
		delete(a.Whitelists.UserAgents, UserAgent(whitelist.Value))
	case WhitelistTypeContracts:
		if contracts, ok := a.Whitelists.Contracts[whitelist.ChainID]; ok {
			delete(contracts, Contract(whitelist.Value))
			if len(contracts) == 0 {
				delete(a.Whitelists.Contracts, whitelist.ChainID)
			}
		}
	case WhitelistTypeMethods:
		if methods, ok := a.Whitelists.Methods[whitelist.ChainID]; ok {
			delete(methods, Method(whitelist.Value))
			if len(methods) == 0 {
				delete(a.Whitelists.Methods, whitelist.ChainID)
			}
		}
	}
}

// GetWhitelistsObject returns a GraphQL-compatible WhitelistsObject struct that contains all whitelists for an application (used for Portal UI)
func (a *PortalApp) GetWhitelistsObject() *WhitelistsObject {
	var origins, userAgents, blockchains []string // App whitelists

	for origin := range a.Whitelists.Origins {
		origins = append(origins, string(origin))
	}
	sort.Strings(origins)

	for userAgent := range a.Whitelists.UserAgents {
		userAgents = append(userAgents, string(userAgent))
	}
	sort.Strings(userAgents)

	for blockchain := range a.Whitelists.Blockchains {
		blockchains = append(blockchains, string(blockchain))
	}
	sort.Strings(blockchains)

	var contractWhitelists, methodWhitelists []ChainIDWhitelists // Chain whitelists

	for chainID, chainContracts := range a.Whitelists.Contracts {
		contracts := []string{}
		for contract := range chainContracts {
			contracts = append(contracts, string(contract))
		}
		sort.Strings(contracts)
		contractWhitelists = append(contractWhitelists, ChainIDWhitelists{ChainID: string(chainID), Values: contracts})
	}
	sort.Slice(contractWhitelists, func(i, j int) bool {
		return contractWhitelists[i].ChainID < contractWhitelists[j].ChainID
	})

	for chainID, chainMethods := range a.Whitelists.Methods {
		methods := []string{}
		for method := range chainMethods {
			methods = append(methods, string(method))
		}
		sort.Strings(methods)
		methodWhitelists = append(methodWhitelists, ChainIDWhitelists{ChainID: string(chainID), Values: methods})
	}
	sort.Slice(methodWhitelists, func(i, j int) bool {
		return methodWhitelists[i].ChainID < methodWhitelists[j].ChainID
	})

	return &WhitelistsObject{
		AppWhitelists: [3]ApplicationWhitelists{
			{Type: WhitelistTypeOrigins, Values: origins},
			{Type: WhitelistTypeUserAgents, Values: userAgents},
			{Type: WhitelistTypeBlockchains, Values: blockchains},
		},
		ChainWhitelists: [2]ChainWhitelists{
			{Type: WhitelistTypeContracts, Values: contractWhitelists},
			{Type: WhitelistTypeMethods, Values: methodWhitelists},
		},
	}
}

func (w *Whitelists) IsEmpty() bool {
	if w == nil {
		return true
	}

	if len(w.Origins) != 0 {
		return false
	}
	if len(w.UserAgents) != 0 {
		return false
	}
	if len(w.Blockchains) != 0 {
		return false
	}
	if len(w.Contracts) != 0 {
		return false
	}
	if len(w.Methods) != 0 {
		return false
	}

	return true
}

// ConvertPortalAppToPortalAppLite converts a PortalApp to a PortalAppLite for use in the Portal Middleware
// Returns a copy rather than a pointer as its intended be used at the point of returning the request body
func (a *PortalApp) ConvertPortalAppToPortalAppLite(appPlan *Plan) PortalAppLite {
	portalAppLite := PortalAppLite{
		ID:         a.getIDForMiddleware(),
		PublicKeys: a.GetPublicKeys(),
	}
	if !a.Whitelists.IsEmpty() {
		portalAppLite.Whitelists = a.Whitelists
	}
	if a.Settings.SecretKeyRequired {
		portalAppLite.Settings = SettingsLite{
			SecretKey:         a.Settings.SecretKey,
			SecretKeyRequired: a.Settings.SecretKeyRequired,
		}
	}
	if appPlan != nil {
		portalAppLite.Plan = PlanLite{
			PlanType:        appPlan.Type,
			ChainIDs:        appPlan.ChainIDs,
			ThroughputLimit: appPlan.ThroughputLimit,
		}
	}

	return portalAppLite
}

// getIDForMiddleware returns the PortalAppID for a PortalApp for use in the middleware,
// with special handling to return the AAT.ID field in the case of a direct app.
func (a *PortalApp) getIDForMiddleware() PortalAppID {
	// Direct apps are Applications from V1 that did not have an associated LoadBalancer.
	if a.IsDirectApp() {
		return a.getDirectAppID()
	}
	// All other apps are V2 PortalApps.
	return a.ID
}

func (a *PortalApp) IsDirectApp() bool {
	return strings.Contains(string(a.ID), "direct_app")
}

// getDirectAppID returns the AAT.ID field for a direct app. This field was formerly
// the Application.ID field from V1 and is used in the Middleware to look up direct apps.
func (a *PortalApp) getDirectAppID() PortalAppID {
	for id := range a.AATs {
		return PortalAppID(id)
	}
	return ""
}

// GetPublicKeys returns all public keys for a PortalApp
func (a *PortalApp) GetPublicKeys() []PortalAppPublicKey {
	publicKeys := []PortalAppPublicKey{}

	for _, aat := range a.AATs {
		publicKeys = append(publicKeys, aat.PublicKey)
	}

	return publicKeys
}

func (a *PortalAppLite) PlanType() PayPlanType {
	return a.Plan.PlanType
}

func (a *PortalAppLite) IsChainIDOnPlan(chainID RelayChainID) bool {
	_, ok := a.Plan.ChainIDs[chainID]
	return ok
}

func (a *PortalAppLite) ThroughputLimit() int32 {
	return a.Plan.ThroughputLimit
}

// IsOriginWhitelisted returns a boolean indicating whether the given ORIGIN is whitelisted for an application
func (a *PortalAppLite) IsOriginWhitelisted(origin Origin) bool {
	_, ok := a.Whitelists.Origins[origin]
	return ok
}

// IsUserAgentWhitelisted returns a boolean indicating whether the given USER AGENT is whitelisted for an application
func (a *PortalAppLite) IsUserAgentWhitelisted(userAgent UserAgent) bool {
	_, ok := a.Whitelists.UserAgents[userAgent]
	return ok
}

// IsBlockchainWhitelisted returns a boolean indicating whether the given BLOCKCHAIN is whitelisted for an application
func (a *PortalAppLite) IsBlockchainWhitelisted(blockchain RelayChainID) bool {
	_, ok := a.Whitelists.Blockchains[blockchain]
	return ok
}

// IsContractWhitelisted returns a boolean indicating whether the given CONTRACT is whitelisted for a blockchain and application
func (a *PortalAppLite) IsContractWhitelisted(chainID RelayChainID, contract Contract) bool {
	if chainContracts, contractsOK := a.Whitelists.Contracts[chainID]; contractsOK {
		if _, contractOK := chainContracts[contract]; contractOK {
			return true
		}
	}
	return false
}

// IsMethodWhitelisted returns a boolean indicating whether the given METHOD is whitelisted for a blockchain and application
func (a *PortalAppLite) IsMethodWhitelisted(chainID RelayChainID, method Method) bool {
	if chainMethods, methodsOK := a.Whitelists.Methods[chainID]; methodsOK {
		if _, methodOK := chainMethods[method]; methodOK {
			return true
		}
	}
	return false
}

func (a *PortalApp) Table() Table {
	return TablePortalApps
}

func (a *Settings) Table() Table {
	return TableAppSettings
}

func (a *Whitelist) Table() Table {
	return TableAppWhitelists
}

func (a *AppNotification) Table() Table {
	return TableAppNotifications
}

func (a *AAT) Table() Table {
	return TablePortalAppAATs
}
