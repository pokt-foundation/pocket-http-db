package cache

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pokt-foundation/portal-db/types"
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
	fields := logrus.Fields{
		"err": err.Error(),
	}

	c.log.WithFields(fields).Error(err)
}

func PrettyString(label string, thing interface{}) {
	jsonThing, _ := json.Marshal(thing)
	str := string(jsonThing)

	var prettyJSON bytes.Buffer
	_ = json.Indent(&prettyJSON, []byte(str), "", "    ")
	output := prettyJSON.String()

	fmt.Println(label, output)
}

func (c *Cache) parseApplicationNotification(n types.Notification) {
	app, ok := n.Data.(*types.Application)
	PrettyString("APP JSON", app)
	if !ok {
		c.logError(fmt.Errorf("parseApplicationNotification failed: %w", errParseApplicationFailed))
		return
	}

	if n.Action == types.ActionInsert {
		c.addApplication(*app)
	}
	if n.Action == types.ActionUpdate {
		c.updateApplication(*app)
	}
}

func (c *Cache) parseBlockchainNotification(n types.Notification) {
	blockchain, ok := n.Data.(*types.Blockchain)
	if !ok {
		c.logError(fmt.Errorf("parseBlockchainNotification failed: %w", errParseBlockchainFailed))
		return
	}

	if n.Action == types.ActionInsert {
		c.addBlockchain(*blockchain)
	}
	if n.Action == types.ActionUpdate {
		c.updateBlockchain(*blockchain)
	}
}

func (c *Cache) parseAppLimitNotification(n types.Notification) {
	limit, ok := n.Data.(*types.AppLimit)
	if !ok {
		fmt.Println("parse app limit failed")
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addAppLimit(*limit)
	}
}

func (c *Cache) parseGatewayAATNotification(n types.Notification) {
	aat, ok := n.Data.(*types.GatewayAAT)
	if !ok {
		c.logError(fmt.Errorf("parseGatewayAATNotification failed: %w", errParseGatewayAATFailed))
		return
	}

	if n.Action == types.ActionInsert {
		c.addGatewayAAT(*aat)
	}
}

func (c *Cache) parseGatewaySettingsNotification(n types.Notification) {
	settings, ok := n.Data.(*types.GatewaySettings)
	if !ok {
		c.logError(fmt.Errorf("parseGatewaySettingsNotification failed: %w", errParseGatewaySettingsFailed))
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addGatewaySettings(*settings)
	}
}

func (c *Cache) parseLoadBalancerNotification(n types.Notification) {
	lb, ok := n.Data.(*types.LoadBalancer)
	if !ok {
		c.logError(fmt.Errorf("parseLoadBalancerNotification failed: %w", errParseLoadBalancerFailed))
		return
	}

	if n.Action == types.ActionInsert {
		c.addLoadBalancer(*lb)
	}
	if n.Action == types.ActionUpdate {
		c.updateLoadBalancer(*lb)
	}
}

func (c *Cache) parseNotificationSettingsNotification(n types.Notification) {
	settings, ok := n.Data.(*types.NotificationSettings)
	if !ok {
		c.logError(fmt.Errorf("parseNotificationSettingsNotification failed: %w", errParseNotificationSettingsFailed))
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addNotificationSettings(*settings)
	}
}

func (c *Cache) parseRedirectNotification(n types.Notification) {
	redirect, ok := n.Data.(*types.Redirect)
	if !ok {
		c.logError(fmt.Errorf("parseRedirectNotification failed: %w", errParseRedirectFailed))
		return
	}

	if n.Action == types.ActionInsert {
		c.addRedirect(*redirect)
	}
}

func (c *Cache) parseStickinessOptionsNotification(n types.Notification) {
	opts, ok := n.Data.(*types.StickyOptions)
	if !ok {
		c.logError(fmt.Errorf("parseStickinessOptionsNotification failed: %w", errParseStickinessOptionsFailed))
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addStickinessOptions(*opts)
	}
}

func (c *Cache) parseSyncOptionsNotification(n types.Notification) {
	opts, ok := n.Data.(*types.SyncCheckOptions)
	if !ok {
		c.logError(fmt.Errorf("parseSyncOptionsNotification failed: %w", errParseSyncCheckOptionsFailed))
		return
	}

	if n.Action == types.ActionInsert || n.Action == types.ActionUpdate {
		c.addSyncOptions(*opts)
	}
}

func (c *Cache) parseLbApps(n types.Notification) {
	lbApp, ok := n.Data.(*types.LbApp)
	if !ok {
		c.logError(fmt.Errorf("parseLbApps failed: %w", errParseLBAppsFailed))
		return
	}

	if n.Action == types.ActionInsert {
		c.addLbApp(*lbApp)
	}
}

func (c *Cache) parseNotification(n types.Notification) {
	switch n.Table {
	case types.TableLoadBalancers:
		c.parseLoadBalancerNotification(n)
	case types.TableStickinessOptions:
		c.parseStickinessOptionsNotification(n)

	case types.TableLbApps:
		c.parseLbApps(n)

	case types.TableApplications:
		c.parseApplicationNotification(n)
	case types.TableAppLimits:
		c.parseAppLimitNotification(n)
	case types.TableGatewayAAT:
		c.parseGatewayAATNotification(n)
	case types.TableGatewaySettings:
		c.parseGatewaySettingsNotification(n)
	case types.TableNotificationSettings:
		c.parseNotificationSettingsNotification(n)

	case types.TableBlockchains:
		c.parseBlockchainNotification(n)
	case types.TableRedirects:
		c.parseRedirectNotification(n)
	case types.TableSyncCheckOptions:
		c.parseSyncOptionsNotification(n)
	}
}

func (c *Cache) listen() {
	c.listening = true

	for {
		n := <-c.reader.NotificationChannel()
		PrettyString("NOTIFICATION", n)
		go c.parseNotification(*n)
	}
}
