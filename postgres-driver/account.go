package postgresdriver

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/types"
)

type (
	accountUserDBRow struct {
		UserID         string `json:"user_id"`
		types.RoleName `json:"role_name"`
	}
)

var (
	errNoRoleName                   = errors.New("error no role name set")
	errInvalidRoleName              = errors.New("error invalid role name set")
	errInvalidIconURL               = errors.New("error icon URL '%s' is not a valid URL")
	errPayPlanDoesntExist           = errors.New("error pay plan '%s' does not exist")
	errAccountDoesntExist           = errors.New("error account does not exist for account ID '%s'")
	errAccountUserDoesntExist       = errors.New("error user ID '%s' does not exist for account ID '%s'")
	errCannotDeleteOwner            = errors.New("error cannot delete user ID '%s' for account ID '%s' because this user is the current account owner")
	errCreateNoAccountID            = errors.New("error must provide account ID when creating user")
	errTransferNoAccountID          = errors.New("error must provide account ID when transferring user")
	errAcceptNoAccountID            = errors.New("error must provide account ID when accepting user")
	errCannotTransferNotAccepted    = errors.New("error cannot transfer ownership to user ID '%s' for account ID '%s' because the user has not accepted their invite")
	errCannotDeclineAlreadyAccepted = errors.New("error user ID '%s' cannot decline invite for account ID '%s' because they have already accepted it")
)

/* ----- postgresdriver Account Read Methods ----- */

// ReadAccounts returns all Accounts in the database as Accounts structs
func (pg *PostgresDriver) ReadAccounts(ctx context.Context, options types.DriverOptions) (map[types.AccountID]*types.Account, error) {
	dbAccounts, err := pg.SelectAccounts(ctx, options.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	accounts := make(map[types.AccountID]*types.Account, len(dbAccounts))
	for _, dbAccount := range dbAccounts {
		account, err := dbAccount.toAccount()
		if err != nil {
			return nil, err
		}

		accounts[dbAccount.ID] = account
	}

	return accounts, nil
}

// toAccount converts Account SELECT output to Account struct
func (a *SelectAccountsRow) toAccount() (*types.Account, error) {
	var chainIDs map[types.RelayChainID]struct{}
	if len(a.PartnerChainIDs) != 0 {
		chainIDs = make(map[types.RelayChainID]struct{}, len(a.PartnerChainIDs))
		for _, chainID := range a.PartnerChainIDs {
			chainIDs[types.RelayChainID(chainID)] = struct{}{}
		}
	}

	var partnerChainIDs map[types.RelayChainID]struct{}
	if len(a.PartnerChainIDs) != 0 {
		partnerChainIDs = make(map[types.RelayChainID]struct{}, len(a.PartnerChainIDs))
		for _, chainID := range a.PartnerChainIDs {
			partnerChainIDs[types.RelayChainID(chainID)] = struct{}{}
		}
	}

	return &types.Account{
		ID:                     a.ID,
		Name:                   a.Name.String,
		IconURL:                a.IconURL.String,
		PlanType:               types.PayPlanType(a.PlanType.String),
		PartnerChainIDs:        partnerChainIDs,
		PartnerThroughputLimit: a.PartnerThroughputLimit.Int32,
		PartnerAppLimit:        a.PartnerApplicationLimit.Int32,
		Integrations: types.AccountIntegrations{
			StripeSubscriptionID: a.StripeSubscriptionID.String,
			CovalentAPIKeyFree:   a.CovalentAPIKeyFree.String,
			CovalentAPIKeyPaid:   a.CovalentAPIKeyPaid.String,
		},
		CreatedAt: a.CreatedAt.Time.UTC(),
		UpdatedAt: a.UpdatedAt.Time.UTC(),
		Deleted:   a.Deleted,

		// PortalApps map set in cache when cache is set
		PortalApps: map[types.PortalAppID]*types.PortalApp{},
		// Users map set in cache when Account is fetched
		Users: map[types.UserID]types.AccountUser{},
	}, nil
}

/* ----- postgresdriver Account Create Methods ----- */

// WriteAccount creates a single Account in the database, including its OWNER's AccountUser row
func (pg *PostgresDriver) WriteAccount(ctx context.Context, creatorID types.UserID, account types.Account, createdAt time.Time) (*types.Account, error) {
	tx, err := pg.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := pg.WithTx(tx)

	err = pg.validateWriteAccountInput(ctx, creatorID, account)
	if err != nil {
		return nil, err
	}

	id, err := pg.generateID(ctx)
	if err != nil {
		return nil, err
	}
	account.ID = types.AccountID(id)

	createdAccount, err := qtx.InsertAccount(ctx, InsertAccountParams{
		ID:        account.ID,
		Name:      newText(account.Name),
		IconURL:   newText(account.IconURL),
		PlanType:  newText(string(account.PlanType)),
		CreatedAt: newTimestamptz(createdAt),
		UpdatedAt: newTimestamptz(createdAt),
	})
	if err != nil {
		return nil, err
	}

	account.ID = createdAccount.ID
	account.CreatedAt = createdAt
	account.UpdatedAt = createdAt

	user, err := qtx.GetUserFields(ctx, creatorID)
	if err != nil {
		return nil, err
	}

	// Account creator becomes Account OWNER
	_, err = qtx.InsertAccountUser(ctx, InsertAccountUserParams{
		AccountID: createdAccount.ID,
		UserID:    creatorID,
		Email:     user.Email,
		RoleName:  types.RoleOwner,
		Accepted:  true,
		CreatedAt: newTimestamptz(createdAt),
		UpdatedAt: newTimestamptz(createdAt),
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	// Assign OWNER to returned Account struct
	account.Users = map[types.UserID]types.AccountUser{
		types.UserID(creatorID): {
			ID:       types.UserID(creatorID),
			Email:    types.Email(user.Email),
			RoleName: types.RoleOwner,
			Accepted: true,
		},
	}

	return &account, nil
}

// validateWriteAccountInput validates the input to create a new Account
func (pg *PostgresDriver) validateWriteAccountInput(ctx context.Context, creatorID types.UserID, account types.Account) error {
	if account.IconURL != "" {
		_, err := url.ParseRequestURI(account.IconURL)
		if err != nil {
			return fmt.Errorf(errInvalidIconURL.Error(), account.IconURL)
		}
	}

	planExists, err := pg.CheckPlanTypeExists(ctx, account.PlanType)
	if err != nil {
		return err
	}
	if !planExists {
		return fmt.Errorf(errPayPlanDoesntExist.Error(), account.PlanType)
	}

	userExists, err := pg.CheckUserExists(ctx, creatorID)
	if err != nil {
		return err
	}
	if !userExists {
		return fmt.Errorf(errUserDoesntExist.Error(), creatorID)
	}

	return nil
}

/* UpsertAccountIntegration saves or updates input AccountIntegrations in the database */
func (pg *PostgresDriver) UpsertAccountIntegration(ctx context.Context, integrations types.AccountIntegrations) (*types.AccountIntegrations, error) {
	time := time.Now()

	accountIntegrations, err := pg.UpsertAccountIntegrations(ctx, UpsertAccountIntegrationsParams{
		AccountID:            integrations.AccountID,
		StripeSubscriptionID: newText(integrations.StripeSubscriptionID),
		CovalentAPIKeyFree:   newText(integrations.CovalentAPIKeyFree),
		CovalentAPIKeyPaid:   newText(integrations.CovalentAPIKeyPaid),
		CreatedAt:            newTimestamptz(time),
		UpdatedAt:            newTimestamptz(time),
	})
	if err != nil {
		return nil, err
	}

	return &types.AccountIntegrations{
		AccountID:            accountIntegrations.AccountID,
		StripeSubscriptionID: accountIntegrations.StripeSubscriptionID.String,
		CovalentAPIKeyFree:   accountIntegrations.CovalentAPIKeyFree.String,
		CovalentAPIKeyPaid:   accountIntegrations.CovalentAPIKeyPaid.String,
	}, nil
}

/* ----- postgresdriver Account Update Methods ----- */

// UpdateAccount updates a single Account in the database's PlanType field
func (pg *PostgresDriver) UpdateAccount(ctx context.Context, update types.UpdateAccount, updatedAt time.Time) (*types.Account, error) {
	err := pg.validateUpdateAccountInput(ctx, update)
	if err != nil {
		return nil, err
	}

	err = pg.UpdateAccountFields(ctx, UpdateAccountFieldsParams{
		ID:        update.AccountID,
		Name:      newText(update.Name),
		IconURL:   newText(update.IconURL),
		PlanType:  newText(string(update.PlanType)),
		UpdatedAt: newTimestamptz(updatedAt),
	})
	if err != nil {
		return nil, err
	}

	accountResult, err := pg.SelectAccount(ctx, update.AccountID)
	if err != nil {
		return nil, err
	}

	accountData := &SelectAccountsRow{
		ID:                      accountResult.ID,
		Name:                    accountResult.Name,
		IconURL:                 accountResult.IconURL,
		PlanType:                accountResult.PlanType,
		PartnerChainIDs:         accountResult.PartnerChainIDs,
		PartnerThroughputLimit:  accountResult.PartnerThroughputLimit,
		PartnerApplicationLimit: accountResult.PartnerApplicationLimit,
		CovalentAPIKeyFree:      accountResult.CovalentAPIKeyFree,
		CovalentAPIKeyPaid:      accountResult.CovalentAPIKeyPaid,
		CreatedAt:               accountResult.CreatedAt,
		UpdatedAt:               accountResult.UpdatedAt,
		Deleted:                 accountResult.Deleted,
		DeletedAt:               accountResult.DeletedAt,
	}
	account, err := accountData.toAccount()
	if err != nil {
		return nil, err
	}

	return account, nil
}

// validateUpdateAccountInput validates the input to udapte an Account
func (pg *PostgresDriver) validateUpdateAccountInput(ctx context.Context, update types.UpdateAccount) error {
	if update.IconURL != "" {
		_, err := url.ParseRequestURI(update.IconURL)
		if err != nil {
			return fmt.Errorf(errInvalidIconURL.Error(), update.IconURL)
		}
	}

	accountExists, err := pg.CheckAccountExists(ctx, update.AccountID)
	if err != nil {
		return err
	}
	if !accountExists {
		return fmt.Errorf(errAccountDoesntExist.Error(), update.AccountID)
	}

	if update.PlanType != "" {
		planExists, err := pg.CheckPlanTypeExists(ctx, update.PlanType)
		if err != nil {
			return err
		}
		if !planExists {
			return fmt.Errorf(errPayPlanDoesntExist.Error(), update.PlanType)
		}
	}

	return nil
}

/* ----- postgresdriver Account Delete Methods ----- */

// SetAccountDeleted updates a single Account in the database's Deleted field to true
func (pg *PostgresDriver) SetAccountDeleted(ctx context.Context, accountID types.AccountID, deletedAt time.Time) error {
	err := pg.validateDeleteAccountInput(ctx, accountID)
	if err != nil {
		return err
	}

	params := DeleteAccountParams{ID: accountID, DeletedAt: newTimestamptz(deletedAt)}

	err = pg.DeleteAccount(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

// validateDeleteAccountInput validates the input to delete an existing Account
func (pg *PostgresDriver) validateDeleteAccountInput(ctx context.Context, accountID types.AccountID) error {
	accountExists, err := pg.CheckAccountExists(ctx, accountID)
	if err != nil {
		return err
	}
	if !accountExists {
		return fmt.Errorf(errAccountDoesntExist.Error(), accountID)
	}

	return nil
}

/* ----- postgresdriver AccountUser Write Methods ----- */

// WriteAccountUser saves a single input AccountUser to the database.
func (pg *PostgresDriver) WriteAccountUser(ctx context.Context, createAccountUser types.CreateAccountUser, createdAt time.Time) (types.UserID, error) {
	err := pg.validateWriteAccountUserInput(ctx, createAccountUser)
	if err != nil {
		return "", err
	}

	// determine if user for a given email already exists
	userID, err := pg.CheckUserIDFromEmail(ctx, createAccountUser.Email)
	if err != nil {
		switch {

		case errNoRows(err):
			// user with provided email does not exist in DB so create a new User and AccountUser entry
			createdUserID, err := pg.writeAccountUserAccessNoUser(ctx, createAccountUser, createdAt)
			if err != nil {
				return "", err
			}

			return createdUserID, nil

		default:
			return "", err
		}
	}

	// user with provided email already exists in DB so create a new AccountUser entry
	createdUserID, err := pg.writeAccountUser(ctx, userID, createAccountUser, createdAt)
	if err != nil {
		return "", err
	}

	return createdUserID, nil
}

// validateWriteAccountUserInput validates the input to create a new AccountUser row
func (pg *PostgresDriver) validateWriteAccountUserInput(ctx context.Context, createAccountUser types.CreateAccountUser) error {
	if createAccountUser.Email == "" {
		return errNoEmail
	}
	if !createAccountUser.Email.IsValid() {
		return fmt.Errorf(errInvalidEmail.Error(), createAccountUser.Email)
	}
	if createAccountUser.AccountID == "" {
		return errCreateNoAccountID
	}

	accountExists, err := pg.CheckAccountExists(ctx, createAccountUser.AccountID)
	if err != nil {
		return err
	}
	if !accountExists {
		return fmt.Errorf(errAccountDoesntExist.Error(), createAccountUser.AccountID)
	}

	return nil
}

// writeAccountUserAccessNoUser creates a new User in the database and then creates a new AccountUser for that user
// Called when a user is invited to a new team but does not yet have a Portal Account for the provided email
func (pg *PostgresDriver) writeAccountUserAccessNoUser(
	ctx context.Context,
	createAccountUser types.CreateAccountUser,
	createdAt time.Time,
) (types.UserID, error) {
	id, err := pg.generateID(ctx)
	if err != nil {
		return "", err
	}
	userID := types.UserID(id)

	createdUserID, err := pg.InsertAccountUserAccessNoUser(ctx, InsertAccountUserAccessNoUserParams{
		ID:        userID,
		AccountID: createAccountUser.AccountID,
		Email:     createAccountUser.Email,
		RoleName:  createAccountUser.RoleName,
		CreatedAt: newTimestamptz(createdAt),
		UpdatedAt: newTimestamptz(createdAt),
	})
	if err != nil {
		return "", err
	}

	return createdUserID, nil
}

// writeAccountUserAccessNoUser creates a new AccountUser row for an existing user
// Called when an existing Portal user is invited to a new team
func (pg *PostgresDriver) writeAccountUser(
	ctx context.Context,
	userID types.UserID,
	createAccountUser types.CreateAccountUser,
	createdAt time.Time,
) (types.UserID, error) {
	createdUserID, err := pg.InsertAccountUser(ctx, InsertAccountUserParams{
		UserID:    userID,
		AccountID: createAccountUser.AccountID,
		Email:     createAccountUser.Email,
		RoleName:  createAccountUser.RoleName,
		Accepted:  false,
		CreatedAt: newTimestamptz(createdAt),
		UpdatedAt: newTimestamptz(createdAt),
	})
	if err != nil {
		return "", err
	}

	return createdUserID, nil
}

/* ----- postgresdriver AccountUser Update Methods ----- */

// SetAccountUserRole updates the role for an existing AccountUser row. If transferring ownership the account owner becomes an admin.
func (pg *PostgresDriver) SetAccountUserRole(ctx context.Context, updateAccountUser types.UpdateAccountUserRole, updatedAt time.Time) error {
	err := pg.validateSetAccountUserRoleInput(ctx, updateAccountUser)
	if err != nil {
		return err
	}

	tx, err := pg.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := pg.WithTx(tx)

	// If transferring ownership, set the previous owner to ADMIN
	if updateAccountUser.RoleName == types.RoleOwner {
		err = qtx.UpdateAccountOwnerToAdmin(ctx, UpdateAccountOwnerToAdminParams{
			AccountID: updateAccountUser.AccountID,
			UserID:    updateAccountUser.UserID,
			UpdatedAt: newTimestamptz(updatedAt),
		})
		if err != nil {
			return err
		}
	}

	err = qtx.UpdateAccountUserRole(ctx, UpdateAccountUserRoleParams{
		AccountID: updateAccountUser.AccountID,
		UserID:    updateAccountUser.UserID,
		RoleName:  updateAccountUser.RoleName,
		UpdatedAt: newTimestamptz(updatedAt),
	})
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

// validateSetAccountUserRoleInput validates the input to update an existing AccountUser role
func (pg *PostgresDriver) validateSetAccountUserRoleInput(ctx context.Context, updateAccountUser types.UpdateAccountUserRole) error {
	if updateAccountUser.RoleName == "" {
		return errNoRoleName
	}
	if !updateAccountUser.RoleName.IsValid() {
		return errInvalidRoleName
	}
	if updateAccountUser.AccountID == "" {
		return errTransferNoAccountID
	}

	// If transferring OWNER role
	if updateAccountUser.RoleName == types.RoleOwner {
		// Cannot transfer OWNER to a user who has not accepted their invite
		userAccepted, err := pg.CheckAccountUserAccepted(ctx, CheckAccountUserAcceptedParams{
			AccountID: updateAccountUser.AccountID,
			UserID:    updateAccountUser.UserID,
		})
		if err != nil || !userAccepted {
			return fmt.Errorf(errCannotTransferNotAccepted.Error(), updateAccountUser.UserID, updateAccountUser.AccountID)
		}
	}

	accountUserExists, err := pg.CheckAccountUserExists(ctx, CheckAccountUserExistsParams{
		UserID:    updateAccountUser.UserID,
		AccountID: updateAccountUser.AccountID,
	})
	if err != nil {
		return err
	}
	if !accountUserExists {
		return fmt.Errorf(errAccountUserDoesntExist.Error(), updateAccountUser.UserID, updateAccountUser.AccountID)
	}

	return nil
}

// UpdateAcceptAccountUser creates a new portal UserAuthProvider in the DB when a user accepts their team invite.
// If they accept their invite it also updates `users.signed_up` and `account_user_access.accepted` fields to true.
// If they have declined their invite it will delete their `account_user_access` row and their `users` row if they have not yet signed up with an auth provider.
func (pg *PostgresDriver) UpdateAcceptAccountUser(ctx context.Context, acceptAccountUser types.UpdateAcceptAccountUser, updatedAt time.Time) error {
	err := pg.validateUpdateAcceptAccountUserInput(ctx, acceptAccountUser)
	if err != nil {
		return err
	}

	switch acceptAccountUser.DeclinedInvite {
	// If user has declined their invite delete the `account_user_access` row for the given user and portal app IDs
	// If the user has not yet signed up with an auth provider their `users` row will be deleted as well.
	case true:
		// ensure the user has not already accepted the invite
		userAccepted, err := pg.CheckAccountUserAccepted(ctx, CheckAccountUserAcceptedParams{
			UserID:    acceptAccountUser.UserID,
			AccountID: acceptAccountUser.AccountID,
		})
		if err != nil {
			return err
		}
		if userAccepted {
			return fmt.Errorf(errCannotDeclineAlreadyAccepted.Error(), acceptAccountUser.UserID, acceptAccountUser.AccountID)
		}

		err = pg.UpdateUserDeclinedInvite(ctx, UpdateUserDeclinedInviteParams{
			AccountID: acceptAccountUser.AccountID,
			ID:        acceptAccountUser.UserID,
		})
		if err != nil {
			return err
		}
	// If user has accepted invite then update their `account_user_access` row to set `accepted` to true
	default:
		params := UpdateUserAcceptedInviteParams{
			AccountID:      acceptAccountUser.AccountID,
			UserID:         acceptAccountUser.UserID,
			ProviderUserID: acceptAccountUser.ProviderUserID,
			Type:           acceptAccountUser.AuthProviderType,
			Provider:       acceptAccountUser.AuthProviderType.Provider(),
			Federated:      acceptAccountUser.AuthProviderType.IsFederated(),
		}

		err = pg.UpdateUserAcceptedInvite(ctx, params)
		if err != nil {
			return err
		}
	}

	return nil
}

// validateUpdateAcceptAccountUserInput validates the input to set an AccountUser role to Accepted
func (pg *PostgresDriver) validateUpdateAcceptAccountUserInput(ctx context.Context, acceptAccountUser types.UpdateAcceptAccountUser) error {
	if acceptAccountUser.AuthProviderType == "" {
		return errNoAuthProviderType
	}
	if !acceptAccountUser.AuthProviderType.IsValid() {
		return fmt.Errorf(errInvalidAuthProviderType.Error(), acceptAccountUser.AuthProviderType)
	}
	if acceptAccountUser.AccountID == "" {
		return errAcceptNoAccountID
	}

	existsParams := CheckAccountUserExistsParams{UserID: acceptAccountUser.UserID, AccountID: acceptAccountUser.AccountID}
	accountUserExists, err := pg.CheckAccountUserExists(ctx, existsParams)
	if err != nil {
		return err
	}
	if !accountUserExists {
		return fmt.Errorf(errAccountUserDoesntExist.Error(), acceptAccountUser.UserID, acceptAccountUser.AccountID)
	}

	return nil
}

/* ----- postgresdriver AccountUser Delete Methods ----- */

// RemoveAccountUser deletes a AccountUser row for a given user and account ID.
func (pg *PostgresDriver) RemoveAccountUser(ctx context.Context, userID types.UserID, accountID types.AccountID) error {
	err := pg.validateRemoveAccountUserInput(ctx, userID, accountID)
	if err != nil {
		return err
	}

	err = pg.DeleteAccountUser(ctx, DeleteAccountUserParams{
		AccountID: accountID,
		UserID:    userID,
	})
	if err != nil {
		return err
	}

	return nil
}

// validateRemoveAccountUserInput validates the input to remove an AccountUser row
func (pg *PostgresDriver) validateRemoveAccountUserInput(ctx context.Context, userID types.UserID, accountID types.AccountID) error {
	accountUserRole, err := pg.CheckAccountUserRole(ctx, CheckAccountUserRoleParams{
		UserID:    userID,
		AccountID: accountID,
	})
	if err != nil {
		switch {
		case errNoRows(err):
			return fmt.Errorf(errAccountUserDoesntExist.Error(), userID, accountID)
		default:
			return err
		}
	}

	if accountUserRole == types.RoleOwner {
		return fmt.Errorf(errCannotDeleteOwner.Error(), userID, accountID)
	}

	return nil
}

/* ----- Used by Listener ----- */
func (json dbAccount) toOutput() *types.Account {
	var partnerChainIDs map[types.RelayChainID]struct{}
	if len(json.PartnerChainIDs) != 0 {
		partnerChainIDs = make(map[types.RelayChainID]struct{})
		for _, chainID := range json.PartnerChainIDs {
			partnerChainIDs[types.RelayChainID(chainID)] = struct{}{}
		}
	}

	return &types.Account{
		ID:                     json.ID,
		Name:                   json.Name,
		IconURL:                json.IconURL,
		PlanType:               json.PlanType,
		PartnerChainIDs:        partnerChainIDs,
		PartnerThroughputLimit: json.PartnerThroughputLimit,
		PartnerAppLimit:        json.PartnerApplicationLimit,
		CreatedAt:              json.CreatedAt,
		UpdatedAt:              json.UpdatedAt,
		Deleted:                json.Deleted,
	}
}

func (json dbAccountUser) toOutput() *types.AccountUser {
	return &types.AccountUser{
		AccountID: json.AccountID,
		ID:        json.UserID,
		RoleName:  json.RoleName,
		Accepted:  json.Accepted,
	}
}

func (j dbAccountIntegration) toOutput() *types.AccountIntegrations {
	return &types.AccountIntegrations{
		AccountID:          j.AccountID,
		CovalentAPIKeyFree: j.CovalentAPIKeyFree,
		CovalentAPIKeyPaid: j.CovalentAPIKeyPaid,
	}
}

type dbAccount struct {
	ID                      types.AccountID   `json:"id"`
	Name                    string            `json:"name"`
	IconURL                 string            `json:"icon_url"`
	PlanType                types.PayPlanType `json:"plan_type"`
	PartnerChainIDs         []string          `json:"partner_chain_ids"`
	PartnerThroughputLimit  int32             `json:"partner_throughput_limit"`
	PartnerApplicationLimit int32             `json:"partner_application_limit"`
	CreatedAt               time.Time         `json:"created_at"`
	UpdatedAt               time.Time         `json:"updated_at"`
	Deleted                 bool              `json:"deleted"`
	DeletedAt               time.Time         `json:"deleted_at"`
}

type dbAccountIntegration struct {
	ID                 int32           `json:"id"`
	AccountID          types.AccountID `json:"account_id"`
	CovalentAPIKeyFree string          `json:"covalent_api_key_free"`
	CovalentAPIKeyPaid string          `json:"covalent_api_key_paid"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type dbAccountUser struct {
	ID        int32           `json:"id"`
	UserID    types.UserID    `json:"user_id"`
	AccountID types.AccountID `json:"account_id"`
	RoleName  types.RoleName  `json:"role_name"`
	Accepted  bool            `json:"accepted"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
