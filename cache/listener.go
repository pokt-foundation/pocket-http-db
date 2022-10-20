package cache

import (
	"time"

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

func (c *Cache) addGatewayAAT(aat *repository.GatewayAAT) {
	c.addGatewayAATWithRetries(aat, 0)
}

func (c *Cache) addGatewayAATWithRetries(aat *repository.GatewayAAT, retries int) {
	if retries >= retriesSideTable {
		return
	}

	app := c.GetApplication(aat.ID)
	if app != nil {
		c.rwMutex.Lock()
		defer c.rwMutex.Unlock()

		aat.ID = "" // to avoid multiple sources of truth
		app.GatewayAAT = *aat
		return
	}

	time.Sleep(1 * time.Second)
	retries++
	c.addGatewayAATWithRetries(aat, retries)
}

func (c *Cache) parseGatewayAATNotification(n *repository.Notification) {
	aat := n.Data.(*repository.GatewayAAT)
	if n.Action == repository.ActionInsert {
		c.addGatewayAAT(aat)
	}
}

func (c *Cache) addGatewaySettings(settings *repository.GatewaySettings) {
	c.addGatewaySettingsWithRetries(settings, 0)
}

func (c *Cache) addGatewaySettingsWithRetries(settings *repository.GatewaySettings, retries int) {
	if retries >= retriesSideTable {
		return
	}

	app := c.GetApplication(settings.ID)
	if app != nil {
		c.rwMutex.Lock()
		defer c.rwMutex.Unlock()

		settings.ID = "" // to avoid multiple sources of truth
		app.GatewaySettings = *settings
		return
	}

	time.Sleep(1 * time.Second)
	retries++
	c.addGatewaySettingsWithRetries(settings, retries)
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

func (c *Cache) addNotificationSettings(settings *repository.NotificationSettings) {
	c.addNotificationSettingsWithRetries(settings, 0)
}

func (c *Cache) addNotificationSettingsWithRetries(settings *repository.NotificationSettings, retries int) {
	if retries >= retriesSideTable {
		return
	}

	app := c.GetApplication(settings.ID)
	if app != nil {
		c.rwMutex.Lock()
		defer c.rwMutex.Unlock()

		settings.ID = "" // to avoid multiple sources of truth
		app.NotificationSettings = *settings
		return
	}

	time.Sleep(1 * time.Second)
	retries++
	c.addNotificationSettingsWithRetries(settings, retries)
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

func (c *Cache) addStickinessOptions(opts *repository.StickyOptions) {
	c.addStickinessOptionsWithRetries(opts, 0)
}

func (c *Cache) addStickinessOptionsWithRetries(opts *repository.StickyOptions, retries int) {
	if retries >= retriesSideTable {
		return
	}

	lb := c.GetLoadBalancer(opts.ID)
	if lb != nil {
		c.rwMutex.Lock()
		defer c.rwMutex.Unlock()

		opts.ID = "" // to avoid multiple sources of truth
		lb.StickyOptions = *opts
		return
	}

	time.Sleep(1 * time.Second)
	retries++
	c.addStickinessOptionsWithRetries(opts, retries)
}

func (c *Cache) parseStickinessOptionsNotification(n *repository.Notification) {
	opts := n.Data.(*repository.StickyOptions)
	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addStickinessOptions(opts)
	}
}

func (c *Cache) addSyncOptions(opts *repository.SyncCheckOptions) {
	c.addSyncOptionsWithRetries(opts, 0)
}

func (c *Cache) addSyncOptionsWithRetries(opts *repository.SyncCheckOptions, retries int) {
	if retries >= retriesSideTable {
		return
	}

	blockchain := c.GetBlockchain(opts.BlockchainID)
	if blockchain != nil {
		c.rwMutex.Lock()
		defer c.rwMutex.Unlock()

		blockchain.SyncCheckOptions = *opts
		return
	}

	time.Sleep(1 * time.Second)
	retries++
	c.addSyncOptionsWithRetries(opts, retries)
}

func (c *Cache) parseSyncOptionsNotification(n *repository.Notification) {
	opts := n.Data.(*repository.SyncCheckOptions)
	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addSyncOptions(opts)
	}
}

func (c *Cache) addLbApp(lbApp *repository.LbApp) {
	c.addLbAppWithRetries(lbApp, 0)
}

func (c *Cache) addLbAppWithRetries(lbApp *repository.LbApp, retries int) {
	if retries >= retriesSideTable {
		return
	}

	lb := c.GetLoadBalancer(lbApp.LbID)
	if lb != nil {
		c.rwMutex.Lock()
		defer c.rwMutex.Unlock()

		lb.Applications = append(lb.Applications, c.applicationsMap[lbApp.AppID])
		return
	}

	time.Sleep(1 * time.Second)
	retries++
	c.addLbAppWithRetries(lbApp, retries)
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
