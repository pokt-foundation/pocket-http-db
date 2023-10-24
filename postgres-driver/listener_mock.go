package postgresdriver

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgxlisten"
	"github.com/pokt-foundation/portal-http-db/v2/types"
)

type (
	// ListenerMock simulates a listener that receives notifications from a PostgreSQL database.
	ListenerMock struct {
		Notify   chan *pgconn.Notification
		handlers map[string]pgxlisten.Handler
	}
	inputStruct struct {
		action types.Action
		table  types.Table
		input  any
	}
)

func NewListenerMock(outCh chan *types.Notification) *ListenerMock {
	listenerMock := &ListenerMock{
		Notify:   make(chan *pgconn.Notification, 32),
		handlers: make(map[string]pgxlisten.Handler),
	}

	handler := &PGXNotificationHandler{outCh: outCh}
	listenerMock.Handle(listenerChannel, handler)

	return listenerMock
}

// NotificationChannel returns a channel that receives pq.Notification instances.
func (l *ListenerMock) NotificationChannel() <-chan *pgconn.Notification {
	return l.Notify
}

func (l *ListenerMock) Listen(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case notif := <-l.Notify:
				if handler, ok := l.handlers[notif.Channel]; ok {
					pgxNotif := &pgconn.Notification{
						Channel: notif.Channel,
						Payload: notif.Payload,
					}

					err := handler.HandleNotification(ctx, pgxNotif, nil)
					if err != nil {
						log.Println(err)
						return
					}
				}
			}
		}
	}()

	return nil
}

func (l *ListenerMock) Handle(channel string, handler pgxlisten.Handler) {
	if l.handlers == nil {
		l.handlers = make(map[string]pgxlisten.Handler)
	}

	l.handlers[channel] = handler
}

// MockEvent simulates a database event by sending mock notifications to the ListenerMock's Notify channel.
// To exclude main or side table actions simply pass an empty string as the parameter.
func (l *ListenerMock) MockEvent(mainTableAction, sideTablesAction types.Action, content types.SavedOnDB) {
	notifications := mockContent(mainTableAction, sideTablesAction, content)

	for _, notification := range notifications {
		l.Notify <- notification
	}
}

// mockInput creates a mock pq.Notification based on the inputStruct provided.
func mockInput(inStruct inputStruct) *pgconn.Notification {
	notification, _ := json.Marshal(notification{
		Table:  inStruct.table,
		Action: inStruct.action,
		Data:   inStruct.input,
	})

	return &pgconn.Notification{
		Channel: listenerChannel,
		Payload: string(notification),
	}
}

// mockContent creates a list of mock notifications for a given mainTableAction, sideTablesAction, and content.
func mockContent(mainTableAction, sideTablesAction types.Action, content types.SavedOnDB) []*pgconn.Notification {
	var inputs []inputStruct

	switch content.(type) {
	case *types.Account:
		inputs = accountInputs(mainTableAction, sideTablesAction, content)
	case *types.PortalApp:
		inputs = portalAppInputs(mainTableAction, sideTablesAction, content)
	case *types.GigastakeApp:
		inputs = gigastakeAppInputs(mainTableAction, sideTablesAction, content)
	case *types.Chain:
		inputs = chainInputs(mainTableAction, sideTablesAction, content)
	default:
		panic("invalid content type")
	}

	var notifications []*pgconn.Notification

	for _, input := range inputs {
		notifications = append(notifications, mockInput(input))
	}

	return notifications
}

// accountInputs generates the mock data for a listener notification for an Account struct
func accountInputs(mainTableAction, sideTablesAction types.Action, content types.SavedOnDB) []inputStruct {
	account := content.(*types.Account)

	var inputs []inputStruct

	partnerChainIDs := make([]string, 0, len(account.PartnerChainIDs))
	for chainID := range account.PartnerChainIDs {
		partnerChainIDs = append(partnerChainIDs, string(chainID))
	}

	if mainTableAction != "" {
		inputs = append(inputs, inputStruct{
			action: mainTableAction,
			table:  types.TableAccounts,
			input: dbAccount{
				ID:                      account.ID,
				PlanType:                account.PlanType,
				PartnerChainIDs:         partnerChainIDs,
				PartnerThroughputLimit:  account.PartnerThroughputLimit,
				PartnerApplicationLimit: account.PartnerAppLimit,
				CreatedAt:               account.CreatedAt,
				UpdatedAt:               account.UpdatedAt,
				Deleted:                 account.Deleted,
			},
		})
	}

	if sideTablesAction != "" {
		for _, user := range account.Users {
			inputs = append(inputs, inputStruct{
				action: sideTablesAction,
				table:  types.TableAccountUserAccess,
				input: dbAccountUser{
					AccountID: account.ID,
					UserID:    user.ID,
					RoleName:  user.RoleName,
					Accepted:  user.Accepted,
				},
			})
		}
	}

	return inputs
}

// portalAppInputs generates the mock data for a listener notification for a PortalApp struct
func portalAppInputs(mainTableAction, sideTablesAction types.Action, content types.SavedOnDB) []inputStruct {
	portalApp := content.(*types.PortalApp)

	var inputs []inputStruct

	if mainTableAction != "" {
		inputs = append(inputs, inputStruct{
			action: mainTableAction,
			table:  types.TablePortalApps,
			input: dbPortalApplication{
				ID:                   portalApp.ID,
				AccountID:            string(portalApp.AccountID),
				Name:                 portalApp.Name,
				FirstDateSurpassed:   portalApp.FirstDateSurpassed,
				CreatedAt:            portalApp.CreatedAt,
				UpdatedAt:            portalApp.UpdatedAt,
				Deleted:              portalApp.Deleted,
				RequestTimeout:       portalApp.LegacyFields.RequestTimeout,
				CustomLimit:          portalApp.LegacyFields.CustomLimit,
				StripeSubscriptionID: portalApp.LegacyFields.StripeSubscriptionID,
			},
		})
	}

	if sideTablesAction != "" {
		aat := types.AAT{}
		for _, portalAAT := range portalApp.AATs {
			aat = portalAAT
			break
		}

		inputs = append(inputs, inputStruct{
			action: sideTablesAction,
			table:  types.TablePortalAppAATs,
			input: dbPortalApplicationAAT{
				PortalAppID:     portalApp.ID,
				ID:              aat.ID,
				Address:         aat.Address,
				PublicKey:       aat.PublicKey,
				ClientPublicKey: aat.ClientPublicKey,
				Signature:       aat.Signature,
				Version:         aat.Version,
			},
		})

		favoritedChainIDs := make([]string, 0, len(portalApp.Settings.FavoritedChainIDs))
		for chainID := range portalApp.Settings.FavoritedChainIDs {
			favoritedChainIDs = append(favoritedChainIDs, string(chainID))
		}

		inputs = append(inputs, inputStruct{
			action: sideTablesAction,
			table:  types.TableAppSettings,
			input: dbPortalApplicationSetting{
				ApplicationID:     portalApp.ID,
				MonthlyRelayLimit: portalApp.Settings.MonthlyRelayLimit,
				Environment:       portalApp.Settings.Environment,
				FavoritedChainIDs: favoritedChainIDs,
				SecretKey:         portalApp.Settings.SecretKey,
				SecretKeyRequired: portalApp.Settings.SecretKeyRequired,
			},
		})

		for chainID := range portalApp.Whitelists.Blockchains {
			inputs = append(inputs, inputStruct{
				action: sideTablesAction,
				table:  types.TableAppWhitelists,
				input: dbPortalApplicationWhitelist{
					ApplicationID: portalApp.ID,
					Type:          types.WhitelistTypeBlockchains,
					Value:         string(chainID),
				},
			})
		}

		for _, notification := range portalApp.Notifications {
			events := []types.NotificationEvent{}
			for event, active := range notification.Events {
				if active {
					events = append(events, event)
				}
			}

			inputs = append(inputs, inputStruct{
				action: sideTablesAction,
				table:  types.TableAppNotifications,
				input: dbPortalApplicationNotification{
					ApplicationID: portalApp.ID,
					Active:        notification.Active,
					Type:          notification.Type,
					Destination:   notification.Destination,
					Trigger:       notification.Trigger,
					Events:        events,
				},
			})
		}

		for _, notification := range portalApp.AATs {
			inputs = append(inputs, inputStruct{
				action: sideTablesAction,
				table:  types.TablePortalAppAATs,
				input: dbPortalApplicationAAT{
					PortalAppID:     portalApp.ID,
					ID:              notification.ID,
					Address:         notification.Address,
					PublicKey:       notification.PublicKey,
					ClientPublicKey: notification.ClientPublicKey,
					Signature:       notification.Signature,
					Version:         notification.Version,
				},
			})
		}
	}

	return inputs
}

// gigastakeAppInputs generates the mock data for a listener notification for a GigastakeApp struct
func gigastakeAppInputs(mainTableAction types.Action, sideTablesAction types.Action, content types.SavedOnDB) []inputStruct {
	gigastakeApp := content.(*types.GigastakeApp)

	var inputs []inputStruct

	if mainTableAction != "" {
		inputs = append(inputs, inputStruct{
			action: mainTableAction,
			table:  types.TableGigastakeApps,
			input: dbGigastakeApp{
				ID:              gigastakeApp.ID,
				Name:            gigastakeApp.Name,
				Address:         gigastakeApp.Address,
				PublicKey:       string(gigastakeApp.PublicKey),
				ClientPublicKey: gigastakeApp.ClientPublicKey,
				Signature:       gigastakeApp.Signature,
				Version:         gigastakeApp.Version,
				CreatedAt:       gigastakeApp.CreatedAt,
				UpdatedAt:       gigastakeApp.UpdatedAt,
				Deleted:         gigastakeApp.Deleted,
			},
		})
	}

	if sideTablesAction != "" {
		var chainID types.RelayChainID
		for appChainID := range gigastakeApp.ChainIDs {
			chainID = appChainID
			break
		}

		inputs = append(inputs, inputStruct{
			action: sideTablesAction,
			table:  types.TableChainGigastakeApps,
			input: dbChainGigastakeApp{
				GigastakeAppID: gigastakeApp.ID,
				ChainID:        chainID,
			},
		})
	}

	return inputs
}

// chainInputs generates the mock data for a listener notification for a Chain struct
func chainInputs(mainTableAction, sideTablesAction types.Action, content types.SavedOnDB) []inputStruct {
	chain := content.(*types.Chain)

	var inputs []inputStruct

	if mainTableAction != "" {
		inputs = append(inputs, inputStruct{
			action: mainTableAction,
			table:  types.TableChains,
			input: dbChain{
				ID:             chain.ID,
				Blockchain:     chain.Blockchain,
				Description:    chain.Description,
				EnforceResult:  chain.EnforceResult,
				Ticker:         chain.Ticker,
				Path:           chain.Path,
				RequestTimeout: chain.RequestTimeout,
				LogLimitBlocks: chain.LogLimitBlocks,
				AllowedMethods: chain.AllowedMethods,
				Active:         chain.Active,
				CreatedAt:      chain.CreatedAt,
				UpdatedAt:      chain.UpdatedAt,
			},
		})
	}

	if sideTablesAction != "" {
		for _, altruist := range chain.Altruists {
			inputs = append(inputs, inputStruct{
				action: sideTablesAction,
				table:  types.TableChainAltruists,
				input: dbChainAltruist{
					ChainID:  chain.ID,
					URL:      altruist.URL,
					Auth:     altruist.Auth,
					AuthType: altruist.AuthType,
				},
			})
		}

		for _, check := range chain.Checks {
			inputs = append(inputs, inputStruct{
				action: sideTablesAction,
				table:  types.TableChainChecks,
				input: dbChainCheck{
					ChainID:    chain.ID,
					Type:       check.Type,
					Payload:    check.Payload,
					ResultKey:  check.ResultKey,
					Allowance:  check.Allowance,
					EVMChainID: check.EVMChainID,
				},
			})
		}

		for alias := range chain.Aliases {
			inputs = append(inputs, inputStruct{
				action: sideTablesAction,
				table:  types.TableChainAliases,
				input: dbChainAlias{
					ChainID: chain.ID,
					Alias:   alias,
				},
			})
		}

		// DEPRECATED - TODO remove when move to only store aliases is complete
		for alias, domains := range chain.AliasDomains {
			inputs = append(inputs, inputStruct{
				action: sideTablesAction,
				table:  types.TableChainAliasDomains,
				input: dbChainAliasDomains{
					ChainID: chain.ID,
					Alias:   alias,
					Domains: domains,
				},
			})
		}
	}

	return inputs
}
