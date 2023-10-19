package cache

import (
	"fmt"
	"os"

	"log/slog"

	"github.com/pokt-foundation/portal-db/v2/types"
)

func (c *Cache) listen() {
	for {
		n, ok := <-c.reader.NotificationChannel()
		if !ok {
			c.logger.Error("notification channel closed, exiting PHD")
			os.Exit(1)
		}

		queueSize := len(c.reader.NotificationChannel())
		if queueSize > 0 {
			c.logger.Info("Notification channel queue > 0",
				slog.Int("queue_size", queueSize),
			)
		}

		if n == nil {
			// Temporary fix until proper reconnection logic is in place.
			c.logger.Error("got a nil value from notification listener")
			continue
		}

		go c.parseNotification(*n)
	}
}

func (c *Cache) parseNotification(n types.Notification) {
	c.logger.Info("Starting parseNotification",
		slog.String("action", string(n.Action)),
		slog.String("table", string(n.Table)),
	)

	t := NewRunTimer(fmt.Sprintf("parseNotification %s %s", n.Action, n.Table), c.logger)
	defer t.Stop()

	switch n.Table {
	case types.TableAccounts:
		c.parseAccountNotification(n)
	case types.TableAccountUserAccess:
		c.parseAccountUserAccessNotification(n)
	case types.TableAccountIntegrations:
		c.parseAccountIntegrationsNotification(n)

	case types.TablePortalApps:
		c.parsePortalAppNotification(n)
	case types.TableAppSettings:
		c.parseSettingsNotification(n)
	case types.TableAppWhitelists:
		c.parseWhitelistNotification(n)
	case types.TableAppNotifications:
		c.parseAppNotificationNotification(n)
	case types.TablePortalAppAATs:
		c.parsePortalAppAATNotification(n)

	case types.TableGigastakeApps:
		c.parseGigastakeAppNotification(n)

	case types.TableChainGigastakeApps:
		c.parseChainGigastakeAppNotification(n)

	case types.TableChains:
		c.parseChainNotification(n)
	case types.TableChainAltruists:
		c.parseChainAltruistNotification(n)
	case types.TableChainAliases:
		c.parseChainAliasNotification(n)
	// DEPRECATED - TODO remove when move to only store aliases is complete
	case types.TableChainAliasDomains:
		c.parseChainAliasDomainsNotification(n)
	case types.TableChainChecks:
		c.parseChainCheckNotification(n)

	case types.TableUsers:
		c.parseUsersNotification(n)
	case types.TableUserAuthProviders:
		c.parseUserAuthProviderNotification(n)

	case types.TableGlobalBlockedContracts:
		c.parseGlobalBlockedContractsNotification(n)
	}
}

/* ---------- Accounts ---------- */

func (c *Cache) parseAccountNotification(n types.Notification) {
	account, ok := n.Data.(*types.Account)
	if !ok {
		c.logger.Error("parseAccountNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.upsertAccount(*account)
	}
}

func (c *Cache) parseAccountUserAccessNotification(n types.Notification) {
	accountUserAccess, ok := n.Data.(*types.AccountUser)
	if !ok {
		c.logger.Error("parseAccountUserAccessNotification failed")
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
		c.logger.Error("parseAccountIntegrationsNotification failed")
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
		c.logger.Error("parsePortalAppNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.upsertPortalApp(*portalApp)
	}
}

func (c *Cache) parseSettingsNotification(n types.Notification) {
	appSettings, ok := n.Data.(*types.Settings)
	if !ok {
		c.logger.Error("parseSettingsNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addPortalAppSettings(*appSettings)
	}
}

func (c *Cache) parseWhitelistNotification(n types.Notification) {
	appWhitelist, ok := n.Data.(*types.Whitelist)
	if !ok {
		c.logger.Error("parseWhitelistNotification failed")
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
		c.logger.Error("parseAppNotificationNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addAppNotification(*appNotification)
	}
}

func (c *Cache) parsePortalAppAATNotification(n types.Notification) {
	appAAT, ok := n.Data.(*types.AAT)
	if !ok {
		c.logger.Error("parsePortalAppAATNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addPortalAppAAT(*appAAT)
	}
}

/* ---------- GigastakeApps ---------- */

func (c *Cache) parseGigastakeAppNotification(n types.Notification) {
	gigastakeApp, ok := n.Data.(*types.GigastakeApp)
	if !ok {
		c.logger.Error("parseGigastakeAppsNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.upsertGigastakeApp(*gigastakeApp)
	}
}

func (c *Cache) parseChainGigastakeAppNotification(n types.Notification) {
	chainGigastakeApp, ok := n.Data.(*types.ChainGigastakeApp)
	if !ok {
		c.logger.Error("parseChainGigastakeAppNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addChainGigastakeApp(*chainGigastakeApp)
	}
}

/* ---------- Chains ---------- */

func (c *Cache) parseChainNotification(n types.Notification) {
	chain, ok := n.Data.(*types.Chain)
	if !ok {
		c.logger.Error("parseChainNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.upsertChain(*chain)
	}
}

func (c *Cache) parseChainAltruistNotification(n types.Notification) {
	chainAltruist, ok := n.Data.(*types.Altruist)
	if !ok {
		c.logger.Error("parseChainAltruistNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addAltruist(*chainAltruist)
	}
	if n.Action == types.ActionDelete {
		c.deleteAltruist(*chainAltruist)
	}
}

func (c *Cache) parseChainCheckNotification(n types.Notification) {
	chainCheck, ok := n.Data.(*types.Check)
	if !ok {
		c.logger.Error("parseChainCheckNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addCheck(*chainCheck)
	}
	if n.Action == types.ActionDelete {
		c.deleteCheck(*chainCheck)
	}
}

func (c *Cache) parseChainAliasNotification(n types.Notification) {
	aliasDomains, ok := n.Data.(*types.Alias)
	if !ok {
		c.logger.Error("parseChainAliasNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addAlias(*aliasDomains)
	}
	if n.Action == types.ActionDelete {
		c.deleteAlias(*aliasDomains)
	}
}

// DEPRECATED - TODO remove when move to only store aliases is complete
func (c *Cache) parseChainAliasDomainsNotification(n types.Notification) {
	aliasDomains, ok := n.Data.(*types.AliasDomains)
	if !ok {
		c.logger.Error("parseChainAliasDomainsNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addAliasDomains(*aliasDomains)
	}
	if n.Action == types.ActionDelete {
		c.deleteAliasDomains(*aliasDomains)
	}
}

/* ---------- Users ---------- */

func (c *Cache) parseUsersNotification(n types.Notification) {
	user, ok := n.Data.(*types.User)
	if !ok {
		c.logger.Error("parseUsersNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.upsertUser(*user)
	}
	if n.Action == types.ActionDelete {
		c.deleteUser(*user)
	}
}

func (c *Cache) parseUserAuthProviderNotification(n types.Notification) {
	authProvider, ok := n.Data.(*types.UserAuthProvider)
	if !ok {
		c.logger.Error("parseUserAuthProviderNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addUserAuthProvider(*authProvider)
	}
}

/* ---------- Global Blocked Contracts ---------- */

func (c *Cache) parseGlobalBlockedContractsNotification(n types.Notification) {
	blockedContract, ok := n.Data.(*types.BlockedContract)
	if !ok {
		c.logger.Error("parseGlobalBlockedContractsNotification failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.upsertGlobalBlockedContract(*blockedContract)
	}
	if n.Action == types.ActionDelete {
		c.deleteGlobalBlockedContract(*blockedContract)
	}
}
