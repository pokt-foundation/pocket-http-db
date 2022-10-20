package cache

import (
	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	retriesSideTable = 5
)

func (c *Cache) parseApplicationNotification(n *repository.Notification) {
	app := n.Data.(*repository.Application)
	if n.Action == repository.ActionInsert {
		c.addApplication(app)
	}
	if n.Action == repository.ActionUpdate {
		c.updateApplication(app)
	}
}

func (c *Cache) parseBlockchainNotification(n *repository.Notification) {
	blockchain := n.Data.(*repository.Blockchain)
	if n.Action == repository.ActionInsert {
		c.addBlockchain(blockchain)
	}
	if n.Action == repository.ActionUpdate {
		c.updateBlockchain(blockchain)
	}
}

func (c *Cache) parseGatewayAATNotification(n *repository.Notification) {
	aat := n.Data.(*repository.GatewayAAT)
	if n.Action == repository.ActionInsert {
		c.addGatewayAAT(aat)
	}
}

func (c *Cache) parseGatewaySettingsNotification(n *repository.Notification) {
	settings := n.Data.(*repository.GatewaySettings)
	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addGatewaySettings(settings)
	}
}

func (c *Cache) parseLoadBalancerNotification(n *repository.Notification) {
	lb := n.Data.(*repository.LoadBalancer)
	if n.Action == repository.ActionInsert {
		c.addLoadBalancer(lb)
	}
	if n.Action == repository.ActionUpdate {
		c.updateLoadBalancer(lb)
	}
}

func (c *Cache) parseNotificationSettingsNotification(n *repository.Notification) {
	settings := n.Data.(*repository.NotificationSettings)
	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addNotificationSettings(settings)
	}
}

func (c *Cache) parseRedirectNotification(n *repository.Notification) {
	redirect := n.Data.(*repository.Redirect)
	if n.Action == repository.ActionInsert {
		c.AddRedirect(redirect)
	}
}

func (c *Cache) parseStickinessOptionsNotification(n *repository.Notification) {
	opts := n.Data.(*repository.StickyOptions)
	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addStickinessOptions(opts)
	}
}

func (c *Cache) parseSyncOptionsNotification(n *repository.Notification) {
	opts := n.Data.(*repository.SyncCheckOptions)
	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addSyncOptions(opts)
	}
}

func (c *Cache) parseLbApps(n *repository.Notification) {
	lbApp := n.Data.(*repository.LbApp)
	if n.Action == repository.ActionInsert {
		c.addLbApp(lbApp)
	}
}

func (c *Cache) parseNotification(n *repository.Notification) {
	switch n.Table {
	case repository.TableApplications:
		c.parseApplicationNotification(n)
	case repository.TableBlockchains:
		c.parseBlockchainNotification(n)
	case repository.TableGatewayAAT:
		c.parseGatewayAATNotification(n)
	case repository.TableGatewaySettings:
		c.parseGatewaySettingsNotification(n)
	case repository.TableLoadBalancers:
		c.parseLoadBalancerNotification(n)
	case repository.TableNotificationSettings:
		c.parseNotificationSettingsNotification(n)
	case repository.TableRedirects:
		c.parseRedirectNotification(n)
	case repository.TableStickinessOptions:
		c.parseStickinessOptionsNotification(n)
	case repository.TableSyncCheckOptions:
		c.parseSyncOptionsNotification(n)
	case repository.TableLbApps:
		c.parseLbApps(n)
	}
}

func (c *Cache) listen() {
	c.listening = true

	for {
		n := <-c.reader.NotificationChannel()
		go c.parseNotification(n)
	}
}
