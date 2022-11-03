package cache

import (
	"errors"
	"fmt"

	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/sirupsen/logrus"
)

var (
	errParseApplicationFailed          = errors.New("parse application failed")
	errParseBlockchainFailed           = errors.New("parse blockchain failed")
	errParseGatewayAATFailed           = errors.New("parse gateway aat failed")
	errParseGatewaySettingsFailed      = errors.New("parse gateway settings failed")
	errParseLoadBalancerFailed         = errors.New("parse load balancer failed")
	errParseNotificationSettingsFailed = errors.New("parse notification failed")
	errParseLBAppsFailed               = errors.New("parse lb app failed")
	errParseSyncCheckOptionsFailed     = errors.New("parse sync check options failed")
	errParseStickinessOptionsFailed    = errors.New("parse stickiness options failed")
	errParseRedirectFailed             = errors.New("parse redirect failed")
)

func (c *Cache) logError(err error) {
	if c.log == nil {
		c.log = logrus.New()
	}

	fields := logrus.Fields{
		"err": err.Error(),
	}

	c.log.WithFields(fields).Error(err)
}

func (c *Cache) parseApplicationNotification(n repository.Notification) {
	app, ok := n.Data.(*repository.Application)
	if !ok {
		c.logError(fmt.Errorf("parseApplicationNotification failed: %w", errParseApplicationFailed))
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
		c.logError(fmt.Errorf("parseBlockchainNotification failed: %w", errParseBlockchainFailed))
		return
	}

	if n.Action == repository.ActionInsert {
		c.addBlockchain(*blockchain)
	}
	if n.Action == repository.ActionUpdate {
		c.updateBlockchain(*blockchain)
	}
}

func (c *Cache) parseGatewayAATNotification(n repository.Notification) {
	aat, ok := n.Data.(*repository.GatewayAAT)
	if !ok {
		c.logError(fmt.Errorf("parseGatewayAATNotification failed: %w", errParseGatewayAATFailed))
		return
	}

	if n.Action == repository.ActionInsert {
		c.addGatewayAAT(*aat)
	}
}

func (c *Cache) parseGatewaySettingsNotification(n repository.Notification) {
	settings, ok := n.Data.(*repository.GatewaySettings)
	if !ok {
		c.logError(fmt.Errorf("parseGatewaySettingsNotification failed: %w", errParseGatewaySettingsFailed))
		return
	}

	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addGatewaySettings(*settings)
	}
}

func (c *Cache) parseLoadBalancerNotification(n repository.Notification) {
	lb, ok := n.Data.(*repository.LoadBalancer)
	if !ok {
		c.logError(fmt.Errorf("parseLoadBalancerNotification failed: %w", errParseLoadBalancerFailed))
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
		c.logError(fmt.Errorf("parseNotificationSettingsNotification failed: %w", errParseNotificationSettingsFailed))
		return
	}

	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addNotificationSettings(*settings)
	}
}

func (c *Cache) parseRedirectNotification(n repository.Notification) {
	redirect, ok := n.Data.(*repository.Redirect)
	if !ok {
		c.logError(fmt.Errorf("parseRedirectNotification failed: %w", errParseRedirectFailed))
		return
	}

	if n.Action == repository.ActionInsert {
		c.addRedirect(*redirect)
	}
}

func (c *Cache) parseStickinessOptionsNotification(n repository.Notification) {
	opts, ok := n.Data.(*repository.StickyOptions)
	if !ok {
		c.logError(fmt.Errorf("parseStickinessOptionsNotification failed: %w", errParseStickinessOptionsFailed))
		return
	}

	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addStickinessOptions(*opts)
	}
}

func (c *Cache) parseSyncOptionsNotification(n repository.Notification) {
	opts, ok := n.Data.(*repository.SyncCheckOptions)
	if !ok {
		c.logError(fmt.Errorf("parseSyncOptionsNotification failed: %w", errParseSyncCheckOptionsFailed))
		return
	}

	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addSyncOptions(*opts)
	}
}

func (c *Cache) parseLbApps(n repository.Notification) {
	lbApp, ok := n.Data.(*repository.LbApp)
	if !ok {
		c.logError(fmt.Errorf("parseLbApps failed: %w", errParseLBAppsFailed))
		return
	}

	if n.Action == repository.ActionInsert {
		c.addLbApp(*lbApp)
	}
}

func (c *Cache) parseNotification(n repository.Notification) {
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
		go c.parseNotification(*n)
	}
}
