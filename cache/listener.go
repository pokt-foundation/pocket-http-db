package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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

	errInvalidFormatInput = errors.New("input could not be marshaled")

	log = logrus.New()
)

func logError(msg string, inputs []interface{}, err error) {
	inputStr := ""

	for _, v := range inputs {
		inputStr += fmt.Sprintf("%v,", interfaceToString(v))
	}

	inputStr = strings.TrimRight(inputStr, ",")

	fields := logrus.Fields{
		"err":    err.Error(),
		"inputs": inputStr,
	}

	log.WithFields(fields).Error(fmt.Sprintf("%s with error: %s", msg, err.Error()))
}

func interfaceToString(inter interface{}) string {
	marshaledInterface, err := json.Marshal(inter)
	if err != nil {
		return errInvalidFormatInput.Error()
	}

	return string(marshaledInterface)
}

func (c *Cache) parseApplicationNotification(n repository.Notification) {
	app, ok := n.Data.(*repository.Application)
	if !ok {
		logError("parseApplicationNotification failed", []interface{}{n.Data}, errParseApplicationFailed)
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
		logError("parseBlockchainNotification failed", []interface{}{n.Data}, errParseBlockchainFailed)
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
		logError("parseGatewayAATNotification failed", []interface{}{n.Data}, errParseGatewayAATFailed)
		return
	}

	if n.Action == repository.ActionInsert {
		c.addGatewayAAT(*aat)
	}
}

func (c *Cache) parseGatewaySettingsNotification(n repository.Notification) {
	settings, ok := n.Data.(*repository.GatewaySettings)
	if !ok {
		logError("parseGatewaySettingsNotification failed", []interface{}{n.Data}, errParseGatewaySettingsFailed)
		return
	}

	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addGatewaySettings(*settings)
	}
}

func (c *Cache) parseLoadBalancerNotification(n repository.Notification) {
	lb, ok := n.Data.(*repository.LoadBalancer)
	if !ok {
		logError("parseLoadBalancerNotification failed", []interface{}{n.Data}, errParseLoadBalancerFailed)
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
		logError("parseNotificationSettingsNotification failed", []interface{}{n.Data}, errParseNotificationSettingsFailed)
		return
	}

	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addNotificationSettings(*settings)
	}
}

func (c *Cache) parseRedirectNotification(n repository.Notification) {
	redirect, ok := n.Data.(*repository.Redirect)
	if !ok {
		logError("parseRedirectNotification failed", []interface{}{n.Data}, errParseRedirectFailed)
		return
	}

	if n.Action == repository.ActionInsert {
		c.addRedirect(*redirect)
	}
}

func (c *Cache) parseStickinessOptionsNotification(n repository.Notification) {
	opts, ok := n.Data.(*repository.StickyOptions)
	if !ok {
		logError("parseStickinessOptionsNotification failed", []interface{}{n.Data}, errParseStickinessOptionsFailed)
		return
	}

	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addStickinessOptions(*opts)
	}
}

func (c *Cache) parseSyncOptionsNotification(n repository.Notification) {
	opts, ok := n.Data.(*repository.SyncCheckOptions)
	if !ok {
		logError("parseSyncOptionsNotification failed", []interface{}{n.Data}, errParseSyncCheckOptionsFailed)
		return
	}

	if n.Action == repository.ActionInsert || n.Action == repository.ActionUpdate {
		c.addSyncOptions(*opts)
	}
}

func (c *Cache) parseLbApps(n repository.Notification) {
	lbApp, ok := n.Data.(*repository.LbApp)
	if !ok {
		logError("parseLbApps failed", []interface{}{n.Data}, errParseLBAppsFailed)
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
