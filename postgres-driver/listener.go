package postgresdriver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgxlisten"
	"github.com/pokt-foundation/portal-http-db/v2/types"
)

const listenerChannel = "events"

type (
	// Listener interface establishes the structure for handling and listening to PostgreSQL notifications.
	Listener interface {
		Handle(channel string, handler pgxlisten.Handler)
		Listen(ctx context.Context) error
	}
	// PGXNotificationHandler structure for handling PostgreSQL notifications.
	PGXNotificationHandler struct {
		outCh chan *types.Notification
	}
	// notification structure for parsing PostgreSQL notifications.
	notification struct {
		Table  types.Table  `json:"table"`
		Action types.Action `json:"action"`
		Data   any          `json:"data"`
	}
)

// NewPGXPoolListener creates a new pgxlisten.Listener with a connection from the provided pool and output channel.
func NewPGXPoolListener(pool *pgxpool.Pool, outCh chan *types.Notification) *pgxlisten.Listener {
	connectFunc := func(ctx context.Context) (*pgx.Conn, error) {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to acquire connection: %v", err)
		}
		return conn.Conn(), nil
	}

	listener := &pgxlisten.Listener{
		Connect:  connectFunc,
		LogError: logError,
	}

	handler := &PGXNotificationHandler{outCh: outCh}
	listener.Handle(listenerChannel, handler)

	return listener
}

// HandleNotification handles incoming PostgreSQL notifications by parsing them in a separate goroutine.
func (h *PGXNotificationHandler) HandleNotification(ctx context.Context, n *pgconn.Notification, conn *pgx.Conn) error {
	go h.parsePGXNotification(n)
	return nil
}

// parsePGXNotification parses incoming PostgreSQL notifications and sends them to the output channel.
func (h *PGXNotificationHandler) parsePGXNotification(n *pgconn.Notification) {
	if n != nil {
		var notification *notification

		err := json.Unmarshal([]byte(n.Payload), &notification)
		if err != nil {
			return
		}

		h.outCh <- notification.parseNotification()
	}
}

// logError logs errors that occur while listening to PostgreSQL notifications.
func logError(ctx context.Context, err error) {
	fmt.Printf("listener notification error: %v", err)
}

// CloseListener closes the notification channel and stops listening for PostgreSQL notifications.
func (d *PostgresDriver) CloseListener() {
	close(d.notification)
}

// parseNotification parses the notification based on the notified table and returns a types.Notification.
func (n notification) parseNotification() *types.Notification {
	switch n.Table {
	case types.TableAccounts:
		return n.parseAccountNotification()
	case types.TableAccountUserAccess:
		return n.parseAccountUserAccessNotification()
	case types.TableAccountIntegrations:
		return n.parseAccountIntegrationsNotification()

	case types.TablePortalApps:
		return n.parsePortalAppNotification()
	case types.TableAppSettings:
		return n.parseSettingsNotification()
	case types.TableAppWhitelists:
		return n.parseWhitelistNotification()
	case types.TableAppNotifications:
		return n.parseAppNotificationNotification()
	case types.TablePortalAppAATs:
		return n.parseAATNotification()

	case types.TableGigastakeApps:
		return n.parseGigastakeAppsNotification()

	case types.TableChainGigastakeApps:
		return n.parseChainGigastakeAppNotification()

	case types.TableChains:
		return n.parseChainNotification()
	case types.TableChainAltruists:
		return n.parseChainAltruistNotification()
	case types.TableChainAliases:
		return n.parseChainAliasNotification()
	// DEPRECATED - TODO remove when move to only store aliases is complete
	case types.TableChainAliasDomains:
		return n.parseChainAliasDomainsNotification()
	case types.TableChainChecks:
		return n.parseChainCheckNotification()

	case types.TableUsers:
		return n.parseUsersNotification()
	case types.TableUserAuthProviders:
		return n.parseUserAuthProviderNotification()

	case types.TablePayPlans:
		return n.parsePayPlanNotification()

	case types.TableGlobalBlockedContracts:
		return n.parseGlobalBlockedContractNotification()

	default:
		return nil
	}
}

/* ---------- Table Data Parser Methods ---------- */

func (n notification) parseAccountNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbAccount dbAccount
	_ = json.Unmarshal(rawData, &dbAccount)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbAccount.toOutput(),
	}
}

func (n notification) parseAccountUserAccessNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbAccountUser dbAccountUserAccess
	_ = json.Unmarshal(rawData, &dbAccountUser)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbAccountUser.toOutput(),
	}
}

func (n notification) parseAccountIntegrationsNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbAccountIntegrations dbAccountIntegration
	_ = json.Unmarshal(rawData, &dbAccountIntegrations)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbAccountIntegrations.toOutput(),
	}
}

func (n notification) parsePortalAppNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbPortalApp dbPortalApplication
	_ = json.Unmarshal(rawData, &dbPortalApp)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbPortalApp.toOutput(),
	}
}

func (n notification) parseSettingsNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbAppSettings dbPortalApplicationSetting
	_ = json.Unmarshal(rawData, &dbAppSettings)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbAppSettings.toOutput(),
	}
}

func (n notification) parseWhitelistNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbAppWhitelist dbPortalApplicationWhitelist
	_ = json.Unmarshal(rawData, &dbAppWhitelist)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbAppWhitelist.toOutput(),
	}
}

func (n notification) parseAppNotificationNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbAppNotification dbPortalApplicationNotification
	_ = json.Unmarshal(rawData, &dbAppNotification)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbAppNotification.toOutput(),
	}
}

func (n notification) parseAATNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbAAT dbPortalApplicationAAT
	_ = json.Unmarshal(rawData, &dbAAT)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbAAT.toOutput(),
	}
}

func (n notification) parseGigastakeAppsNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbGigastakeApp dbGigastakeApp
	_ = json.Unmarshal(rawData, &dbGigastakeApp)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbGigastakeApp.toOutput(),
	}
}

func (n notification) parseChainGigastakeAppNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbChainGigastakeApp dbChainGigastakeApp
	_ = json.Unmarshal(rawData, &dbChainGigastakeApp)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbChainGigastakeApp.toOutput(),
	}
}

func (n notification) parseChainNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbChain dbChain
	_ = json.Unmarshal(rawData, &dbChain)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbChain.toOutput(),
	}
}

func (n notification) parseChainAltruistNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbChainAltruist dbChainAltruist
	_ = json.Unmarshal(rawData, &dbChainAltruist)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbChainAltruist.toOutput(),
	}
}

func (n notification) parseChainCheckNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbChainCheck dbChainCheck
	_ = json.Unmarshal(rawData, &dbChainCheck)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbChainCheck.toOutput(),
	}
}

func (n notification) parseChainAliasNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbAlias dbChainAlias
	_ = json.Unmarshal(rawData, &dbAlias)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbAlias.toOutput(),
	}
}

// DEPRECATED - TODO remove when move to only store aliases is complete
func (n notification) parseChainAliasDomainsNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbAlias dbChainAliasDomains
	_ = json.Unmarshal(rawData, &dbAlias)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbAlias.toOutput(),
	}
}

func (n notification) parseUsersNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbUser dbUser
	_ = json.Unmarshal(rawData, &dbUser)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbUser.toOutput(),
	}
}

func (n notification) parseUserAuthProviderNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbUserAuthProvider dbUserAuthProvider
	_ = json.Unmarshal(rawData, &dbUserAuthProvider)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbUserAuthProvider.toOutput(),
	}
}

func (n notification) parsePayPlanNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbPayPlan dbPayPlan
	_ = json.Unmarshal(rawData, &dbPayPlan)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbPayPlan.toOutput(),
	}
}

func (n notification) parseGlobalBlockedContractNotification() *types.Notification {
	rawData, _ := json.Marshal(n.Data)
	var dbBlockedContract dbGlobalBlockedContract
	_ = json.Unmarshal(rawData, &dbBlockedContract)

	return &types.Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   dbBlockedContract.toOutput(),
	}
}
