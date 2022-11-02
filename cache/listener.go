package cache

import (
	"fmt"

	"github.com/pokt-foundation/portal-api-go/repository"
)

func (c *Cache) parseApplicationNotification(n repository.Notification) {
	app, ok := n.Data.(*repository.Application)
	if !ok {
		fmt.Println("parse application failed")
		return
	}

	if n.Action == repository.ActionInsert {
		c.addApplication(*app)
	}
	if n.Action == repository.ActionUpdate {
		c.updateApplication(*app)
	}
}

func (c *Cache) parseBlockchainNotification(n repository.Notification) {
	blockchain, ok := n.Data.(*repository.Blockchain)
	if !ok {
		fmt.Println("parse blockchain failed")
		return
	}

	if n.Action == repository.ActionInsert {
		c.addBlockchain(*blockchain)
	}
	if n.Action == repository.ActionUpdate {
		c.updateBlockchain(*blockchain)
	}
}

func (c *Cache) parseAppLimitNotification(n repository.Notification) {
	limit, ok := n.Data.(*repository.AppLimit)
	if !ok {
		fmt.Println("parse app limit failed")
		return
	}

	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addAppLimit(*limit)
	}
}

func (c *Cache) parseGatewayAATNotification(n repository.Notification) {
	aat, ok := n.Data.(*repository.GatewayAAT)
	if !ok {
		fmt.Println("parse gateway aat failed")
		return
	}

	if n.Action == repository.ActionInsert {
		c.addGatewayAAT(*aat)
	}
}

func (c *Cache) parseGatewaySettingsNotification(n repository.Notification) {
	settings, ok := n.Data.(*repository.GatewaySettings)
	if !ok {
		fmt.Println("parse gateway settings failed")
		return
	}

	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addGatewaySettings(*settings)
	}
}

func (c *Cache) parseLoadBalancerNotification(n repository.Notification) {
	lb, ok := n.Data.(*repository.LoadBalancer)
	if !ok {
		fmt.Println("parse load balancer failed")
		return
	}

	if n.Action == repository.ActionInsert {
		c.addLoadBalancer(*lb)
	}
	if n.Action == repository.ActionUpdate {
		c.updateLoadBalancer(*lb)
	}
}

func (c *Cache) parseNotificationSettingsNotification(n repository.Notification) {
	settings, ok := n.Data.(*repository.NotificationSettings)
	if !ok {
		fmt.Println("parse notification settings failed")
		return
	}

	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addNotificationSettings(*settings)
	}
}

func (c *Cache) parseRedirectNotification(n repository.Notification) {
	redirect, ok := n.Data.(*repository.Redirect)
	if !ok {
		fmt.Println("parse redirect failed")
		return
	}

	if n.Action == repository.ActionInsert {
		c.addRedirect(*redirect)
	}
}

func (c *Cache) parseStickinessOptionsNotification(n repository.Notification) {
	opts, ok := n.Data.(*repository.StickyOptions)
	if !ok {
		fmt.Println("parse stickiness options failed")
		return
	}

	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addStickinessOptions(*opts)
	}
}

func (c *Cache) parseSyncOptionsNotification(n repository.Notification) {
	opts, ok := n.Data.(*repository.SyncCheckOptions)
	if !ok {
		fmt.Println("parse sync check options failed")
		return
	}

	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addSyncOptions(*opts)
	}
}

func (c *Cache) parseLbApps(n repository.Notification) {
	lbApp, ok := n.Data.(*repository.LbApp)
	if !ok {
		fmt.Println("parse lb app failed")
		return
	}

	if n.Action == repository.ActionInsert {
		c.addLbApp(*lbApp)
	}
}

func (c *Cache) parseNotification(n repository.Notification) {

	switch n.Table {
	case repository.TableLoadBalancers:
		c.parseLoadBalancerNotification(n)
	case repository.TableStickinessOptions:
		c.parseStickinessOptionsNotification(n)

	case repository.TableLbApps:
		c.parseLbApps(n)

	case repository.TableApplications:
		c.parseApplicationNotification(n)
	case repository.TableAppLimits:
		c.parseAppLimitNotification(n)
	case repository.TableGatewayAAT:
		c.parseGatewayAATNotification(n)
	case repository.TableGatewaySettings:
		c.parseGatewaySettingsNotification(n)
	case repository.TableNotificationSettings:
		c.parseNotificationSettingsNotification(n)

	case repository.TableBlockchains:
		c.parseBlockchainNotification(n)
	case repository.TableRedirects:
		c.parseRedirectNotification(n)
	case repository.TableSyncCheckOptions:
		c.parseSyncOptionsNotification(n)
	}
}

func (c *Cache) listen() {
	c.listening = true

	for {
		n := <-c.reader.NotificationChannel()
		go c.parseNotification(*n)
	}
}
