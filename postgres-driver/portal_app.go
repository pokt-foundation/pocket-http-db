package postgresdriver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/types"
)

type (
	whitelistDBRow struct {
		Type         types.WhitelistType `json:"type"`
		Value        string              `json:"value"`
		BlockchainID string              `json:"chain_id"`
	}
	aatDBRow struct {
		ID              types.ProtocolAppID      `json:"id"`
		PublicKey       types.PortalAppPublicKey `json:"public_key"`
		Address         string                   `json:"address"`
		ClientPublicKey string                   `json:"client_public_key"`
		Signature       string                   `json:"signature"`
		Version         string                   `json:"version"`
	}
)

var (
	errEmptyPortalAppName         = errors.New("portal app name cannot be empty")
	errInvalidEnvironment         = errors.New("invalid portal app environment provided: %s")
	errUnmarshallingWhitelists    = errors.New("error unmarshalling whitelists")
	errUnmarshallingNotifications = errors.New("error unmarshalling notifications")
	errUnmarshallingAATs          = errors.New("error unmarshalling AATs")
)

/* ----- postgresdriver PortalApp Read Methods ----- */

// ReadPortalApps returns all portal applications in the database as PortalApp structs
func (pg *PostgresDriver) ReadPortalApps(ctx context.Context, options types.DriverOptions) (map[types.PortalAppID]*types.PortalApp, error) {
	dbPortalApps, err := pg.SelectPortalApplications(ctx, options.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	portalApps := make(map[types.PortalAppID]*types.PortalApp, len(dbPortalApps))
	for _, dbPortalApp := range dbPortalApps {
		portalApp, err := dbPortalApp.toPortalApp()
		if err != nil {
			return nil, err
		}

		portalApps[dbPortalApp.ID] = portalApp
	}

	return portalApps, nil
}

// toPortalApp converts PortalApp SELECT output to PortalApp struct
func (a *SelectPortalApplicationsRow) toPortalApp() (*types.PortalApp, error) {
	var appAATs map[types.ProtocolAppID]types.AAT
	if len(string(a.AATs)) > 2 { // length of empty JSON array in bytes
		aats, err := a.toAATs()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errUnmarshallingAATs, err)
		}
		appAATs = aats
	}

	var appWhitelists types.Whitelists
	if len(string(a.Whitelists)) > 2 { // length of empty JSON array in bytes
		whitelists, err := a.toWhitelists()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errUnmarshallingWhitelists, err)
		}
		appWhitelists = whitelists
	}

	var notifications map[types.NotificationType]types.AppNotification
	if len(string(a.Notifications)) > 2 { // length of empty JSON array in bytes
		if err := json.Unmarshal(a.Notifications, &notifications); err != nil {
			return nil, fmt.Errorf("%s: %w", errUnmarshallingNotifications, err)
		}
	}

	var favoritedChainIDs map[types.RelayChainID]struct{}
	if len(a.FavoritedChainIDs) != 0 {
		favoritedChainIDs = make(map[types.RelayChainID]struct{})
		for _, chainID := range a.FavoritedChainIDs {
			favoritedChainIDs[types.RelayChainID(chainID)] = struct{}{}
		}
	}

	// TODO remove legacy fields when migration to V2 schema complete
	legacyFields := types.LegacyFields{
		PlanType:             a.PlanType,
		DailyLimit:           a.DailyLimit.Int32,
		CustomLimit:          a.CustomLimit.Int32,
		RequestTimeout:       a.RequestTimeout.Int32,
		StripeSubscriptionID: a.StripeSubscriptionID.String,
	}

	return &types.PortalApp{
		ID:          types.PortalAppID(a.ID),
		AccountID:   types.AccountID(a.AccountID.String),
		Name:        a.Name,
		Description: a.Description.String,
		AppEmoji:    types.AppEmoji(a.AppEmoji.String),
		Settings: types.Settings{
			Environment:       types.Environment(a.Environment.Environment),
			SecretKey:         a.SecretKey.String,
			SecretKeyRequired: a.SecretKeyRequired.Bool,
			MonthlyRelayLimit: a.MonthlyRelayLimit.Int32,
			FavoritedChainIDs: favoritedChainIDs,
		},
		AATs:               appAATs,
		Whitelists:         appWhitelists,
		Notifications:      notifications,
		FirstDateSurpassed: a.FirstDateSurpassed.Time.UTC(),
		CreatedAt:          a.CreatedAt.Time.UTC(),
		UpdatedAt:          a.UpdatedAt.Time.UTC(),
		Deleted:            a.Deleted,
		// TODO remove legacy fields when migration to V2 schema complete
		LegacyFields: legacyFields,
	}, nil
}

// toAATs converts AATs from DB rows to map-based PortalApp.AATs struct
func (a *SelectPortalApplicationsRow) toAATs() (map[types.ProtocolAppID]types.AAT, error) {
	var dbAATs map[types.ProtocolAppID]aatDBRow
	if err := json.Unmarshal(a.AATs, &dbAATs); err != nil {
		return nil, err
	}

	aats := make(map[types.ProtocolAppID]types.AAT, len(dbAATs))

	for aatID, dbAAT := range dbAATs {
		aats[aatID] = types.AAT{
			ID:              aatID,
			Address:         dbAAT.Address,
			PublicKey:       dbAAT.PublicKey,
			ClientPublicKey: dbAAT.ClientPublicKey,
			Signature:       dbAAT.Signature,
			Version:         dbAAT.Version,
		}
	}

	return aats, nil
}

// toWhitelists converts whitelists from DB rows to map-based PortalApp.Whitelists struct
func (a *SelectPortalApplicationsRow) toWhitelists() (types.Whitelists, error) {
	whitelists := types.Whitelists{
		Origins:     make(map[types.Origin]struct{}),
		UserAgents:  make(map[types.UserAgent]struct{}),
		Blockchains: make(map[types.RelayChainID]struct{}),
		Contracts:   make(map[types.RelayChainID]map[types.Contract]struct{}),
		Methods:     make(map[types.RelayChainID]map[types.Method]struct{}),
	}

	var whitelistRows []whitelistDBRow
	if err := json.Unmarshal(a.Whitelists, &whitelistRows); err != nil {
		return whitelists, err
	}

	for _, wl := range whitelistRows {
		switch wl.Type {

		case types.WhitelistTypeBlockchains:
			whitelists.Blockchains[types.RelayChainID(wl.Value)] = struct{}{}
		case types.WhitelistTypeOrigins:
			whitelists.Origins[types.Origin(wl.Value)] = struct{}{}
		case types.WhitelistTypeUserAgents:
			whitelists.UserAgents[types.UserAgent(wl.Value)] = struct{}{}

		case types.WhitelistTypeContracts:
			if _, ok := whitelists.Contracts[types.RelayChainID(wl.BlockchainID)]; !ok {
				whitelists.Contracts[types.RelayChainID(wl.BlockchainID)] = make(map[types.Contract]struct{})
			}
			whitelists.Contracts[types.RelayChainID(wl.BlockchainID)][types.Contract(wl.Value)] = struct{}{}
		case types.WhitelistTypeMethods:
			if _, ok := whitelists.Methods[types.RelayChainID(wl.BlockchainID)]; !ok {
				whitelists.Methods[types.RelayChainID(wl.BlockchainID)] = make(map[types.Method]struct{})
			}
			whitelists.Methods[types.RelayChainID(wl.BlockchainID)][types.Method(wl.Value)] = struct{}{}
		}
	}

	return whitelists, nil
}

/* ----- postgresdriver PortalApp Create Methods ----- */

// WritePortalApp creates a single PortalApp in the database, including its AAT and Settings rows
func (pg *PostgresDriver) WritePortalApp(ctx context.Context, portalApp types.PortalApp, aat types.AAT, createdAt time.Time) (*types.PortalApp, error) {
	err := pg.validatePortalAppInput(ctx, portalApp, aat)
	if err != nil {
		return nil, err
	}

	portalAppID, protocolAppID, err := pg.generatePortalAppIDs(ctx)
	if err != nil {
		return nil, err
	}

	portalApp.ID = portalAppID
	portalApp.CreatedAt = createdAt
	portalApp.UpdatedAt = createdAt

	aat.ID = protocolAppID
	portalApp.AATs = map[types.ProtocolAppID]types.AAT{
		protocolAppID: aat,
	}

	tx, err := pg.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := pg.WithTx(tx)

	insert := InsertPortalApplicationParams{
		ID:          portalApp.ID,
		AccountID:   newText(string(portalApp.AccountID)),
		Name:        portalApp.Name,
		Description: newText(portalApp.Description),
		AppEmoji:    newText(string(portalApp.AppEmoji)),
		CreatedAt:   newTimestamptz(portalApp.CreatedAt),
		UpdatedAt:   newTimestamptz(portalApp.UpdatedAt),
		// TODO remove legacy fields when migration to V2 schema complete
		PlanType:           portalApp.LegacyFields.PlanType,
		CustomLimit:        newInt4(portalApp.LegacyFields.CustomLimit, true),
		RequestTimeout:     newInt4(portalApp.LegacyFields.RequestTimeout, true),
		FirstDateSurpassed: newTimestamptz(portalApp.FirstDateSurpassed),
	}

	if portalApp.LegacyFields.PlanType != types.Enterprise {
		dailyLimit, err := qtx.GetPlanDailyLimit(ctx, portalApp.LegacyFields.PlanType)
		if err != nil {
			switch {
			case errNoRows(err):
				return nil, fmt.Errorf(errPayPlanDoesntExist.Error(), portalApp.LegacyFields.PlanType)
			default:
				return nil, err
			}
		}

		insert.DailyLimit = newInt4(dailyLimit.Int32, false)
	}

	_, err = qtx.InsertPortalApplication(ctx, insert)
	if err != nil {
		return nil, err
	}
	createdAAT, err := qtx.InsertPortalApplicationAAT(ctx, InsertPortalApplicationAATParams{
		ID:                  aat.ID,
		PortalApplicationID: portalApp.ID,
		Address:             aat.Address,
		PublicKey:           aat.PublicKey,
		ClientPublicKey:     aat.ClientPublicKey,
		Signature:           aat.Signature,
		Version:             aat.Version,
		PrivateKey:          newText(aat.PrivateKey),
	})
	if err != nil {
		return nil, err
	}
	_, err = qtx.InsertPortalApplicationSetting(ctx, InsertPortalApplicationSettingParams{
		ApplicationID:     portalApp.ID,
		Environment:       newText(string(portalApp.Settings.Environment)),
		SecretKey:         newText(portalApp.Settings.SecretKey),
		SecretKeyRequired: newBool(&portalApp.Settings.SecretKeyRequired),
		MonthlyRelayLimit: portalApp.Settings.MonthlyRelayLimit,
	})
	if err != nil {
		return nil, err
	}

	// TODO - Notifications will be moved to Account level
	// accountOwner, err := qtx.GetAccountOwner(ctx, portalApp.AccountID)
	// if err != nil {
	// 	return nil, err
	// }

	// err = qtx.UpdateUpsertPortalAppNotification(ctx, UpdateUpsertPortalAppNotificationParams{
	// 	ApplicationID: portalApp.ID,
	// 	Active:        true,
	// 	Type:          types.NotificationTypeEmail,
	// 	Destination:   string(accountOwner.Email),
	// 	Events: []types.NotificationEvent{
	// 		types.NotificationEventSignedUp,
	// 		types.NotificationEventThreeQuarters,
	// 		types.NotificationEventFull,
	// 	},
	// 	UpdatedAt: newTimestamptz(portalApp.UpdatedAt),
	// })
	// if err != nil {
	// 	return nil, err
	// }
	// portalApp.Notifications = map[types.NotificationType]types.AppNotification{
	// 	types.NotificationTypeEmail: {
	// 		Type:        types.NotificationTypeEmail,
	// 		Active:      true,
	// 		Destination: string(accountOwner.Email),
	// 		Events: map[types.NotificationEvent]bool{
	// 			types.NotificationEventSignedUp:      true,
	// 			types.NotificationEventThreeQuarters: true,
	// 			types.NotificationEventFull:          true,
	// 		},
	// 	},
	// }

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	portalApp.AATs = map[types.ProtocolAppID]types.AAT{
		createdAAT.ID: {
			ID:              createdAAT.ID,
			Address:         createdAAT.Address,
			PublicKey:       createdAAT.PublicKey,
			ClientPublicKey: createdAAT.ClientPublicKey,
			Signature:       createdAAT.Signature,
			Version:         createdAAT.Version,
		},
	}

	// Don't return PrivateKey in response
	for _, aat := range portalApp.AATs {
		aat.PrivateKey = ""
	}

	return &portalApp, nil
}

// generatePortalAppIDs generates a unique 8 character hex ID for both the Portal App and its AAT
func (pg *PostgresDriver) generatePortalAppIDs(ctx context.Context) (types.PortalAppID, types.ProtocolAppID, error) {
	portalAppID, err := pg.generateID(ctx)
	if err != nil {
		return "", "", err
	}
	protocolAppID, err := pg.generateID(ctx)
	if err != nil {
		return "", "", err
	}
	for protocolAppID == portalAppID { // Ensure the protocol app ID cannot match the portal app ID
		protocolAppID, err = pg.generateID(ctx)
		if err != nil {
			return "", "", err
		}
	}

	return types.PortalAppID(portalAppID), types.ProtocolAppID(protocolAppID), nil
}

// validatePortalAppInput performs all necessary data validation checks on incoming PortalApp data
func (pg *PostgresDriver) validatePortalAppInput(ctx context.Context, portalApp types.PortalApp, aat types.AAT) error {
	if portalApp.Name == "" {
		return errEmptyPortalAppName
	}
	if !portalApp.Settings.Environment.IsValid() {
		return fmt.Errorf(errInvalidEnvironment.Error(), portalApp.Settings.Environment)
	}

	planExists, err := pg.CheckPlanExists(ctx, portalApp.LegacyFields.PlanType)
	if err != nil {
		return err
	}
	if !planExists {
		return fmt.Errorf(errPayPlanDoesntExist.Error(), portalApp.LegacyFields.PlanType)
	}
	accountExists, err := pg.CheckAccountExists(ctx, portalApp.AccountID)
	if err != nil {
		return err
	}
	if !accountExists {
		return fmt.Errorf(errAccountDoesntExist.Error(), portalApp.AccountID)
	}

	return nil
}

/* ----- postgresdriver PortalApp Update Methods ----- */

// UpdatePortalApp updates a single PortalApp in the database: Name field and its Notifications, Whitelists and Settings
func (pg *PostgresDriver) UpdatePortalApp(ctx context.Context, update types.UpdatePortalApp, updatedAt time.Time) error {
	tx, err := pg.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := pg.WithTx(tx)

	if !update.IsEmpty() {
		updateApp := UpdatePortalAppFieldsParams{ID: update.AppID, UpdatedAt: newTimestamptz(updatedAt)}

		if update.Name != "" {
			updateApp.Name = update.Name
		}

		if update.Description != "" {
			updateApp.Description = update.Description
		}

		if update.AppEmoji != "" {
			updateApp.AppEmoji = string(update.AppEmoji)
		}

		if update.PlanType != "" {
			updateApp.PlanType = string(update.PlanType)
			// If changing plan, pull the plan daily limit from the DB for non-enterprise plans
			// TODO remove when change to account-based plan limits completed
			if update.PlanType != types.Enterprise {
				dailyLimit, err := qtx.GetPlanDailyLimit(ctx, update.PlanType)
				if err != nil {
					switch {
					case errNoRows(err):
						return fmt.Errorf(errPayPlanDoesntExist.Error(), update.PlanType)
					default:
						return err
					}
				}

				updateApp.DailyLimit = newInt4(dailyLimit.Int32, false)
			}
		}

		if update.CustomLimit != 0 {
			updateApp.CustomLimit = newInt4(update.CustomLimit, false)
		}

		if update.StripeSubscriptionID != "" {
			updateApp.StripeSubscriptionID = newText(update.StripeSubscriptionID)
		}

		err := qtx.UpdatePortalAppFields(ctx, updateApp)
		if err != nil {
			return err
		}
	}
	if update.Settings != nil {
		err := pg.updateSettings(ctx, qtx, update, updatedAt)
		if err != nil {
			return err
		}
	}
	if update.Notifications != nil && len(update.Notifications) > 0 {
		err := pg.updateNotifications(ctx, qtx, update, updatedAt)
		if err != nil {
			return err
		}
	}
	if update.Whitelists != nil {
		err := pg.updateWhitelists(ctx, qtx, update, updatedAt)
		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

// updateSettings updates the PortalApp's settings row in the portal_application_settings table
func (pg *PostgresDriver) updateSettings(ctx context.Context, qtx *Queries, update types.UpdatePortalApp, updatedAt time.Time) error {
	updateSettings := UpdatePortalAppSettingsParams{
		ApplicationID:     update.AppID,
		SecretKey:         newText(update.Settings.SecretKey),
		SecretKeyRequired: newBool(&update.Settings.SecretKeyRequired),
		MonthlyRelayLimit: update.Settings.MonthlyRelayLimit,
		Environment:       newText(string(update.Settings.Environment)),
		FavoritedChainIDs: update.Settings.FavoritedChainIDs,
		UpdatedAt:         newTimestamptz(updatedAt),
	}

	err := qtx.UpdatePortalAppSettings(ctx, updateSettings)
	if err != nil {
		return err
	}

	return nil
}

// updateNotifications updates the PortalApp's notifications rows in the portal_application_notifications table
func (pg *PostgresDriver) updateNotifications(ctx context.Context, qtx *Queries, update types.UpdatePortalApp, updatedAt time.Time) error {
	for _, appNotification := range update.Notifications {
		updateNotification := UpdateUpsertPortalAppNotificationParams{
			ApplicationID: update.AppID,
			Type:          appNotification.NotificationType,
			Active:        appNotification.Active,
			Destination:   appNotification.Destination,
			Trigger:       appNotification.Trigger,
			Events:        appNotification.Events,
			UpdatedAt:     newTimestamptz(updatedAt),
		}

		// Upsert notification row for application_id & type if active: true in update struct
		err := qtx.UpdateUpsertPortalAppNotification(ctx, updateNotification)
		if err != nil {
			return err
		}

		// Delete notification row for application_id & type in DB if active: false in update struct
		if !appNotification.Active {
			err := qtx.UpdateDeletePortalAppNotification(ctx, UpdateDeletePortalAppNotificationParams{
				ApplicationID: update.AppID, Type: appNotification.NotificationType,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// updateWhitelists updates the PortalApp's whitelists rows in the portal_application_whitelists table
func (pg *PostgresDriver) updateWhitelists(ctx context.Context, qtx *Queries, update types.UpdatePortalApp, updatedAt time.Time) error {
	updateWhitelists := UpdateInsertWhitelistsParams{ApplicationID: update.AppID, CreatedAt: newTimestamptz(updatedAt)}
	updateDeleteWhitelists := UpdateDeleteWhitelistsParams{ApplicationID: update.AppID}

	for _, appWhitelist := range update.Whitelists.AppWhitelists {
		for _, whitelistValue := range appWhitelist.Values {
			exists, err := pg.CheckIfAppWhitelistExists(ctx, CheckIfAppWhitelistExistsParams{
				ApplicationID: updateWhitelists.ApplicationID,
				Type:          appWhitelist.Type,
				Value:         whitelistValue,
			})
			if err != nil {
				return err
			}

			updateDeleteWhitelists.Types = append(updateDeleteWhitelists.Types, appWhitelist.Type)
			updateDeleteWhitelists.Values = append(updateDeleteWhitelists.Values, whitelistValue)
			updateDeleteWhitelists.ChainIDs = append(updateDeleteWhitelists.ChainIDs, "")

			if !exists {
				updateWhitelists.Types = append(updateWhitelists.Types, appWhitelist.Type)
				updateWhitelists.Values = append(updateWhitelists.Values, whitelistValue)
				updateWhitelists.ChainIDs = append(updateWhitelists.ChainIDs, "")
			}
		}
	}

	for _, chainWhitelist := range update.Whitelists.ChainWhitelists {
		for _, blockchainValues := range chainWhitelist.Values {
			for _, whitelistValue := range blockchainValues.Values {
				updateWhitelists.Types = append(updateWhitelists.Types, chainWhitelist.Type)
				updateWhitelists.ChainIDs = append(updateWhitelists.ChainIDs, blockchainValues.ChainID)
				updateWhitelists.Values = append(updateWhitelists.Values, whitelistValue)

				updateDeleteWhitelists.Types = append(updateDeleteWhitelists.Types, chainWhitelist.Type)
				updateDeleteWhitelists.ChainIDs = append(updateDeleteWhitelists.ChainIDs, blockchainValues.ChainID)
				updateDeleteWhitelists.Values = append(updateDeleteWhitelists.Values, whitelistValue)
			}
		}
	}

	err := qtx.UpdateInsertWhitelists(ctx, updateWhitelists)
	if err != nil {
		return err
	}

	err = qtx.UpdateDeleteWhitelists(ctx, updateDeleteWhitelists)
	if err != nil {
		return err
	}

	return nil
}

// UpdatePortalAppsFirstDateSurpassed updates multiple PortalApps' LegacyFields.FirstDateSurpassed fields
func (pg *PostgresDriver) UpdatePortalAppsFirstDateSurpassed(ctx context.Context, update *types.UpdateFirstDateSurpassed) error {
	portalAppIDs := []string{}
	for _, portalAppID := range update.PortalAppIDs {
		portalAppIDs = append(portalAppIDs, string(portalAppID))
	}

	params := UpdateFirstDatesSurpassedParams{
		PortalAppIds:       portalAppIDs,
		FirstDateSurpassed: newTimestamptz(update.FirstDateSurpassed),
	}

	err := pg.UpdateFirstDatesSurpassed(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

/* ----- postgresdriver PortalApp Delete Methods ----- */

// SetPortalAppDeleted updates a single PortalApp in the database's Deleted field to true
func (pg *PostgresDriver) SetPortalAppDeleted(ctx context.Context, portalAppID types.PortalAppID, deletedAt time.Time) error {
	params := DeletePortalAppParams{ID: portalAppID, DeletedAt: newTimestamptz(deletedAt)}

	err := pg.DeletePortalApp(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

/* ----- Used by Listener ----- */
func (json dbPortalApplication) toOutput() *types.PortalApp {
	return &types.PortalApp{
		ID:                 json.ID,
		AccountID:          types.AccountID(json.AccountID),
		Name:               json.Name,
		Description:        json.Description,
		AppEmoji:           types.AppEmoji(json.AppEmoji),
		FirstDateSurpassed: json.FirstDateSurpassed,
		CreatedAt:          json.CreatedAt,
		UpdatedAt:          json.UpdatedAt,
		Deleted:            json.Deleted,
		LegacyFields: types.LegacyFields{
			PlanType:             types.PayPlanType(json.PlanType),
			DailyLimit:           json.DailyLimit,
			CustomLimit:          json.CustomLimit,
			RequestTimeout:       json.RequestTimeout,
			StripeSubscriptionID: json.StripeSubscriptionID,
		},
	}
}

func (json dbPortalApplicationSetting) toOutput() *types.Settings {
	var favoritedChainIDs map[types.RelayChainID]struct{}
	if len(json.FavoritedChainIDs) != 0 {
		favoritedChainIDs = make(map[types.RelayChainID]struct{})
		for _, chainID := range json.FavoritedChainIDs {
			favoritedChainIDs[types.RelayChainID(chainID)] = struct{}{}
		}

	}

	return &types.Settings{
		AppID:             json.ApplicationID,
		Environment:       types.Environment(json.Environment),
		SecretKey:         json.SecretKey,
		SecretKeyRequired: json.SecretKeyRequired,
		FavoritedChainIDs: favoritedChainIDs,
		MonthlyRelayLimit: json.MonthlyRelayLimit,
	}
}

func (json dbPortalApplicationWhitelist) toOutput() *types.Whitelist {
	return &types.Whitelist{
		AppID:   json.ApplicationID,
		Type:    json.Type,
		Value:   json.Value,
		ChainID: types.RelayChainID(json.ChainID),
	}
}

func (json dbPortalApplicationNotification) toOutput() *types.AppNotification {
	return &types.AppNotification{
		AppID:       json.ApplicationID,
		Type:        json.Type,
		Active:      json.Active,
		Destination: json.Destination,
		Trigger:     json.Trigger,
		Events:      json.mapEvents(),
	}
}

func (json dbPortalApplicationNotification) mapEvents() map[types.NotificationEvent]bool {
	events := make(map[types.NotificationEvent]bool)
	for _, event := range json.Events {
		events[event] = true
	}
	return events
}

func (json dbPortalApplicationAAT) toOutput() *types.AAT {
	return &types.AAT{
		AppID:           json.PortalAppID,
		ID:              json.ID,
		Address:         json.Address,
		PublicKey:       json.PublicKey,
		ClientPublicKey: json.ClientPublicKey,
		Signature:       json.Signature,
		Version:         json.Version,
	}
}

type dbPortalApplication struct {
	ID                   types.PortalAppID `json:"id"`
	AccountID            string            `json:"account_id"`
	Name                 string            `json:"name"`
	Description          string            `json:"description"`
	AppEmoji             string            `json:"app_emoji"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
	Deleted              bool              `json:"deleted"`
	DeletedAt            time.Time         `json:"deleted_at"`
	RequestTimeout       int32             `json:"request_timeout"`
	FirstDateSurpassed   time.Time         `json:"first_date_surpassed"`
	PlanType             string            `json:"plan_type"`
	DailyLimit           int32             `json:"daily_limit"`
	CustomLimit          int32             `json:"custom_limit"`
	StripeSubscriptionID string            `json:"stripe_subscription_id"`
}

type dbPortalApplicationNotification struct {
	ID            int32                     `json:"id"`
	ApplicationID types.PortalAppID         `json:"application_id"`
	Active        bool                      `json:"active"`
	Type          types.NotificationType    `json:"type"`
	Destination   string                    `json:"destination"`
	Trigger       string                    `json:"trigger"`
	Events        []types.NotificationEvent `json:"events"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

type dbPortalApplicationSetting struct {
	ID                int32             `json:"id"`
	ApplicationID     types.PortalAppID `json:"application_id"`
	MonthlyRelayLimit int32             `json:"monthly_relay_limit"`
	Environment       types.Environment `json:"environment"`
	FavoritedChainIDs []string          `json:"favorited_chain_ids"`
	SecretKey         string            `json:"secret_key"`
	SecretKeyRequired bool              `json:"secret_key_required"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

type dbPortalApplicationWhitelist struct {
	ID            int32               `json:"id"`
	ApplicationID types.PortalAppID   `json:"application_id"`
	Type          types.WhitelistType `json:"type"`
	Value         string              `json:"value"`
	ChainID       string              `json:"chain_id"`
	CreatedAt     time.Time           `json:"created_at"`
}

type dbPortalApplicationAAT struct {
	ID              types.ProtocolAppID      `json:"id"`
	PortalAppID     types.PortalAppID        `json:"portal_application_id"`
	PublicKey       types.PortalAppPublicKey `json:"public_key"`
	Address         string                   `json:"address"`
	ClientPublicKey string                   `json:"client_public_key"`
	Signature       string                   `json:"signature"`
	Version         string                   `json:"version"`
}
