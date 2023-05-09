package cache

import (
	"github.com/pokt-foundation/portal-db/v2/types"
)

/* ---------- Accounts ---------- */

func (c *Cache) parseAccountNotification(n types.Notification) {
	account, ok := n.Data.(*types.Account)
	if !ok {
		c.log.Error("parseAccountNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.upsertAccount(*account)
	}
}

func (c *Cache) parseAccountUserAccessNotification(n types.Notification) {
	accountUserAccess, ok := n.Data.(*types.AccountUserAccess)
	if !ok {
		c.log.Error("parseAccountUserAccessNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.upsertAccountUser(*accountUserAccess)
	}
	if n.Action == types.ActionDelete {
		c.deleteAccountUser(*accountUserAccess)
	}
}

func (c *Cache) parseAccountIntegrationsNotification(n types.Notification) {
	accountIntegrations, ok := n.Data.(*types.AccountIntegrations)
	if !ok {
		c.log.Error("parseAccountIntegrationsNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.upsertAccountIntegrations(accountIntegrations)
	}
}

/* ---------- PortalApps ---------- */

func (c *Cache) parsePortalAppNotification(n types.Notification) {
	portalApp, ok := n.Data.(*types.PortalApp)
	if !ok {
		c.log.Error("parsePortalAppNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.upsertPortalApp(*portalApp)
	}
}

func (c *Cache) parseAATNotification(n types.Notification) {
	appAAT, ok := n.Data.(*types.AAT)
	if !ok {
		c.log.Error("parseAppAATsNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addPortalAppAAT(*appAAT)
	}
}

func (c *Cache) parseSettingsNotification(n types.Notification) {
	appSettings, ok := n.Data.(*types.Settings)
	if !ok {
		c.log.Error("parseSettingsNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addPortalAppSettings(*appSettings)
	}
}

func (c *Cache) parseWhitelistNotification(n types.Notification) {
	appWhitelist, ok := n.Data.(*types.Whitelist)
	if !ok {
		c.log.Error("parseWhitelistNotification failed")
		return
	}

	if n.Action == types.ActionInsert {
		c.addWhitelist(*appWhitelist)
	} else if n.Action == types.ActionDelete {
		c.deleteWhitelist(*appWhitelist)
	}
}

func (c *Cache) parseAppNotificationNotification(n types.Notification) {
	appNotification, ok := n.Data.(*types.AppNotification)
	if !ok {
		c.log.Error("parseAppNotificationNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addAppNotification(*appNotification)
	}
}

/* ---------- Chains ---------- */

func (c *Cache) parseChainNotification(n types.Notification) {
	chain, ok := n.Data.(*types.Chain)
	if !ok {
		c.log.Error("parseChainNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.upsertChain(*chain)
	}
}

func (c *Cache) parseChainAltruistNotification(n types.Notification) {
	chainAltruist, ok := n.Data.(*types.Altruist)
	if !ok {
		c.log.Error("parseChainAltruistNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addAltruist(*chainAltruist)
	}
}

func (c *Cache) parseGigastakeRedirectNotification(n types.Notification) {
	redirect, ok := n.Data.(*types.GigastakeRedirect)
	if !ok {
		c.log.Error("parseGigastakeRedirectNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addGigastakeRedirect(*redirect)
	}
}

func (c *Cache) parseChainCheckNotification(n types.Notification) {
	chainCheck, ok := n.Data.(*types.Check)
	if !ok {
		c.log.Error("parseChainCheckNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addCheck(*chainCheck)
	}
}

/* ---------- Users ---------- */

func (c *Cache) parseUsersNotification(n types.Notification) {
	user, ok := n.Data.(*types.User)
	if !ok {
		c.log.Error("parseUsersNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.upsertUser(*user)
	}
}

func (c *Cache) parseUserAuthProviderNotification(n types.Notification) {
	authProvider, ok := n.Data.(*types.UserAuthProvider)
	if !ok {
		c.log.Error("parseUserAuthProviderNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addUserAuthProvider(*authProvider)
	}
}

/* ---------- Plans ---------- */

// TODO - implement in cache
// func (c *Cache) parsePayPlanNotification(n types.Notification) {
// 	payPlan, ok := n.Data.(*types.Plan)
// 	if !ok {
// 		c.log.Error("parsePayPlanNotification failed")
// 		return
// 	}

// 	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
// 		c.upsertPayPlan(*payPlan)
// 	}
// }

// TODO - implement in cache
// func (c *Cache) parseGlobalBlockedContractNotification(n types.Notification) {
// 	blockedContract, ok := n.Data.(*types.BlockedContract)
// 	if !ok {
// 		c.log.Error("parseGlobalBlockedContractNotification failed")
// 		return
// 	}

// 	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
// 		c.upsertGlobalBlockedContract(*blockedContract)
// 	}
// }

func (c *Cache) parseNotification(n types.Notification) {
	switch n.Table {
	case types.TableAccounts:
		c.parseAccountNotification(n)
	case types.TableAccountUserAccess:
		c.parseAccountUserAccessNotification(n)
	case types.TableAccountIntegrations:
		c.parseAccountIntegrationsNotification(n)

	case types.TablePortalApps:
		c.parsePortalAppNotification(n)
	case types.TableAppAATs:
		c.parseAATNotification(n)
	case types.TableAppSettings:
		c.parseSettingsNotification(n)
	case types.TableAppWhitelists:
		c.parseWhitelistNotification(n)
	case types.TableAppNotifications:
		c.parseAppNotificationNotification(n)
	case types.TableChains:
		c.parseChainNotification(n)
	case types.TableChainAltruists:
		c.parseChainAltruistNotification(n)
	case types.TableChainGigastakeRedirects:
		c.parseGigastakeRedirectNotification(n)
	case types.TableChainChecks:
		c.parseChainCheckNotification(n)

	case types.TableUsers:
		c.parseUsersNotification(n)
	case types.TableUserAuthProviders:
		c.parseUserAuthProviderNotification(n)

		// TODO - implement in cache
		// case types.TablePayPlans:
		// 	c.parsePayPlanNotification(n)

		// case types.TableGlobalBlockedContracts:
		// 	c.parseGlobalBlockedContractNotification(n)
	}
}

func (c *Cache) listen() {
	for {
		n := <-c.reader.NotificationChannel()
		go c.parseNotification(*n)
	}
}
