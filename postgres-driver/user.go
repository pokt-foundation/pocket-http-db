package postgresdriver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pokt-foundation/portal-http-db/v2/types"
)

const (
	newUser      newUserCase = "new"
	existingUser newUserCase = "existing"
	invitedUser  newUserCase = "invited"
)

type (
	newUserCase string

	authProviderDBRow struct {
		UserID         types.UserID         `json:"user_id"`
		Type           types.AuthType       `json:"type"`
		Provider       types.AuthProvider   `json:"provider"`
		ProviderUserID types.ProviderUserID `json:"provider_user_id"`
		Federated      bool                 `json:"federated"`
	}

	UserRow struct {
		ID               types.UserID       `json:"id"`
		Email            types.Email        `json:"email"`
		SignedUp         bool               `json:"signed_up"`
		IconURL          pgtype.Text        `json:"icon_url"`
		UpdatesProduct   pgtype.Bool        `json:"updates_product"`
		UpdatesMarketing pgtype.Bool        `json:"updates_marketing"`
		BetaTester       pgtype.Bool        `json:"beta_tester"`
		CreatedAt        pgtype.Timestamptz `json:"created_at"`
		UpdatedAt        pgtype.Timestamptz `json:"updated_at"`
		AuthProviders    []byte             `json:"auth_providers"`
	}
)

var (
	errUserDoesntExist          = errors.New("error user does not exist for portal ID '%s'")
	errUserAlreadyExists        = errors.New("error user already exists for email'%s' and provider type '%s'")
	errUserHasAccount           = errors.New("error cannot delete user because they are still on an account team")
	errNoEmail                  = errors.New("error no email")
	errInvalidEmail             = errors.New("error email input is not a valid email address '%s'")
	errNoAuthProviderType       = errors.New("error no auth provider type")
	errInvalidAuthProviderType  = errors.New("error invalid auth provider type '%s'")
	errAuthProviderTypeNotFound = errors.New("error no auth provider type found")
)

// /* ----- postgresdriver User Read Methods ----- */

// ReadUserIDsMap returns all user IDs in the database mapped by their provider user IDs
func (pg *PostgresDriver) ReadUserIDsMap(ctx context.Context) (map[types.ProviderUserID]types.UserID, error) {
	userIDs, err := pg.SelectUserIDs(ctx)
	if err != nil {
		return nil, err
	}

	userIDsMap := make(map[types.ProviderUserID]types.UserID, len(userIDs))
	for _, userIDRow := range userIDs {
		userIDsMap[userIDRow.ProviderUserID] = userIDRow.UserID
	}

	return userIDsMap, nil
}

// readUserByUserID takes a portal UserID and returns a single user in the database as a User struct
func (pg *PostgresDriver) readUserByUserID(ctx context.Context, userID types.UserID) (*types.User, error) {
	userData, err := pg.GetUserDataFromPortalUserID(ctx, userID)
	if err != nil {
		switch {
		case errNoRows(err):
			return nil, fmt.Errorf(errUserDoesntExist.Error(), userID)
		default:
			return nil, err
		}
	}

	user, err := userData.toUserRow().toUser()
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ReadAllUsers returns all users in the database as a map of User structs by UserID
func (pg *PostgresDriver) ReadAllUsers(ctx context.Context) (map[types.UserID]*types.User, error) {
	userData, err := pg.SelectAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	userRoles, err := pg.SelectUserPermissions(ctx)
	if err != nil {
		return nil, err
	}

	usersMap := make(map[types.UserID]*types.User, len(userData))

	for _, userRow := range userData {
		user, err := userRow.toUserRow().toUser()
		if err != nil {
			return nil, err
		}

		usersMap[user.ID] = user
	}

	for _, userRoleRow := range userRoles {
		if user, ok := usersMap[userRoleRow.UserID]; ok {
			if user.Permissions == nil {
				user.Permissions = make(map[types.AccountID]types.AccountPermissions)
			}

			_, err := user.UpsertPermissions(userRoleRow.AccountID, userRoleRow.RoleName, userRoleRow.Accepted)
			if err != nil {
				return nil, err
			}
		}
	}

	return usersMap, nil
}

func (u *GetUserDataFromPortalUserIDRow) toUserRow() *UserRow {
	return &UserRow{
		ID:               u.ID,
		Email:            u.Email,
		SignedUp:         u.SignedUp,
		IconURL:          u.IconURL,
		UpdatesProduct:   u.UpdatesProduct,
		UpdatesMarketing: u.UpdatesMarketing,
		BetaTester:       u.BetaTester,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
		AuthProviders:    u.AuthProviders,
	}
}

func (u *SelectAllUsersRow) toUserRow() *UserRow {
	return &UserRow{
		ID:               u.ID,
		Email:            u.Email,
		SignedUp:         u.SignedUp,
		IconURL:          u.IconURL,
		UpdatesProduct:   u.UpdatesProduct,
		UpdatesMarketing: u.UpdatesMarketing,
		BetaTester:       u.BetaTester,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
		AuthProviders:    u.AuthProviders,
	}
}

// toUser converts User SELECT output to User struct
func (u *UserRow) toUser() (*types.User, error) {
	var providerRows []authProviderDBRow
	if err := json.Unmarshal(u.AuthProviders, &providerRows); err != nil {
		return nil, err
	}

	authProviders := make(map[types.AuthType]types.UserAuthProvider, len(providerRows))
	for _, provider := range providerRows {
		if provider.Type != "" {
			authProviders[provider.Type] = types.UserAuthProvider{
				ProviderUserID: provider.ProviderUserID,
				Type:           provider.Type,
				Provider:       provider.Provider,
				Federated:      provider.Federated,
			}
		}
	}

	return &types.User{
		ID:               u.ID,
		IconURL:          u.IconURL.String,
		Email:            u.Email,
		SignedUp:         u.SignedUp,
		AuthProviders:    authProviders,
		UpdatesProduct:   u.UpdatesProduct.Bool,
		UpdatesMarketing: u.UpdatesMarketing.Bool,
		BetaTester:       u.BetaTester.Bool,
		CreatedAt:        u.CreatedAt.Time.UTC(),
		UpdatedAt:        u.UpdatedAt.Time.UTC(),
	}, nil
}

/* ----- postgresdriver User Create Methods ----- */

// WriteUserNewSignUp creates a new portal User and UserAuthProviderin the DB from a CreateUser input when a new user signs up.
// If a brand new user is created they need to have a default account so we also create a default Account for them using the free tier plan.
// If a new account is created for the user, the account ID is returned; otherwise it is an empty string.
func (pg *PostgresDriver) WriteUserNewSignUp(ctx context.Context, user types.CreateUser, createdAt time.Time) (*types.User, types.AccountID, error) {
	var (
		userID    types.UserID
		accountID types.AccountID
	)

	err := pg.validateWriteUserNewSignUpInput(ctx, user)
	if err != nil {
		return nil, "", err
	}

	tx, err := pg.DB.Begin(ctx)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := pg.WithTx(tx)

	// A new user creation may be one of 3 types: new, existing, or invited
	newUserCase, err := pg.determineNewUserCase(ctx, user.Email)
	if err != nil {
		return nil, "", err
	}

	switch newUserCase {
	// 1. brand new user - never signed up with any auth provider
	case newUser:
		generatedUserID, err := pg.generateID(ctx)
		if err != nil {
			return nil, "", err
		}
		userID = types.UserID(generatedUserID)
	// 2. existing user - signing up with a new auth provider (eg. Github or Google)
	case existingUser:
		userID, err = pg.GetUserIDFromEmail(ctx, user.Email)
		if err != nil {
			return nil, "", err
		}
	// 3. invited user - a user that has been invited to a team signing up for the first time
	case invitedUser:
		userID, err = pg.GetUserIDFromEmail(ctx, user.Email)
		if err != nil {
			return nil, "", err
		}
	}

	authType := user.ProviderUserID.AuthType()

	newUserParams := CreateUserNewSignUpParams{
		ID:             userID,
		Email:          user.Email,
		ProviderUserID: user.ProviderUserID,
		Type:           authType,
		Provider:       authType.Provider(),
		Federated:      authType.IsFederated(),
		CreatedAt:      newTimestamptz(createdAt),
		UpdatedAt:      newTimestamptz(createdAt),
	}

	createdUserID, err := qtx.CreateUserNewSignUp(ctx, newUserParams)
	if err != nil {
		return nil, "", err
	}

	if newUserCase != existingUser {
		// If a new user is created they need to have a default account
		// so we also create a default Account for them using the free tier plan.
		userHasAccount, err := pg.CheckUserHasOwnAccount(ctx, createdUserID)
		if err != nil {
			return nil, "", err
		}
		if !userHasAccount {
			accountID, err = pg.createDefaultAccountForNewUser(ctx, qtx, newUserParams, createdAt)
			if err != nil {
				return nil, "", err
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, "", err
	}

	createdUser, err := pg.readUserByUserID(ctx, createdUserID)
	if err != nil {
		return nil, "", err
	}

	return createdUser, accountID, nil
}

// determineNewUserCase checks the user's status and returns a string indicating the status.
func (pg *PostgresDriver) determineNewUserCase(ctx context.Context, email types.Email) (newUserCase, error) {
	isNewUser, err := pg.IsUserNew(ctx, email)
	if err != nil {
		return "", err
	}
	if isNewUser {
		return newUser, nil
	}

	isExistingUser, err := pg.IsUserExisting(ctx, email)
	if err != nil {
		return "", err
	}
	if isExistingUser {
		return existingUser, nil
	}

	isInvitedUser, err := pg.IsUserInvited(ctx, email)
	if err != nil {
		return "", err
	}
	if isInvitedUser {
		return invitedUser, nil
	}

	return "", errors.New("user status could not be determined")
}

// createDefaultAccountForNewUser creates a default Account for a new User using the free tier plan.
func (pg *PostgresDriver) createDefaultAccountForNewUser(ctx context.Context, qtx *Queries, newUserParams CreateUserNewSignUpParams, createdAt time.Time) (types.AccountID, error) {
	id, err := pg.generateID(ctx)
	if err != nil {
		return "", err
	}
	accountID := types.AccountID(id)

	_, err = qtx.InsertAccount(ctx, InsertAccountParams{
		ID:        accountID,
		PlanType:  newText(string(types.FreetierV0)),
		CreatedAt: newTimestamptz(createdAt),
		UpdatedAt: newTimestamptz(createdAt),
	})
	if err != nil {
		return "", err
	}

	// New user is Account OWNER
	_, err = qtx.InsertAccountUser(ctx, InsertAccountUserParams{
		AccountID: accountID,
		UserID:    newUserParams.ID,
		Email:     newUserParams.Email,
		RoleName:  types.RoleOwner,
		Accepted:  true,
		CreatedAt: newTimestamptz(createdAt),
		UpdatedAt: newTimestamptz(createdAt),
	})
	if err != nil {
		return "", err
	}

	err = qtx.UpdateUpsertAccountNotification(ctx, UpdateUpsertAccountNotificationParams{
		AccountID:   accountID,
		Type:        types.NotificationTypeEmail,
		Active:      true,
		Destination: string(newUserParams.Email),
		Events:      defaultAccountNotifications,
		UpdatedAt:   newTimestamptz(createdAt),
	})
	if err != nil {
		return "", err
	}

	return accountID, nil
}

// validateWriteUserNewSignUpInput validates the input to create a new User and User Auth Provider
func (pg *PostgresDriver) validateWriteUserNewSignUpInput(ctx context.Context, user types.CreateUser) error {
	if !user.Email.IsValid() {
		return fmt.Errorf(errInvalidEmail.Error(), user.Email)
	}
	if !user.ProviderUserID.AuthType().IsValid() {
		if user.ProviderUserID.AuthType() != types.AuthType("") {
			return fmt.Errorf(errInvalidAuthProviderType.Error(), user.ProviderUserID.AuthType())
		}
		return errAuthProviderTypeNotFound
	}

	userExists, err := pg.CheckUserProviderExists(ctx, CheckUserProviderExistsParams{
		Email: user.Email,
		Type:  user.ProviderUserID.AuthType(),
	})
	if err != nil {
		return err
	}
	if userExists {
		return fmt.Errorf(errUserAlreadyExists.Error(), user.Email, user.ProviderUserID.AuthType())
	}

	return nil
}

// UpdateUser updates a User's fields in the DB.
func (pg *PostgresDriver) UpdateUser(ctx context.Context, update types.UpdateUser, updatedAt time.Time) (*types.User, error) {
	err := pg.validateUpdateUserInput(ctx, update)
	if err != nil {
		return nil, err
	}

	params := UpdateUserFieldsParams{
		ID:               update.ID,
		IconURL:          newNullString(update.IconURL),
		UpdatesProduct:   newBool(update.UpdatesProduct),
		UpdatesMarketing: newBool(update.UpdatesMarketing),
		BetaTester:       newBool(update.BetaTester),
		UpdatedAt:        newTimestamptz(updatedAt),
	}

	err = pg.UpdateUserFields(ctx, params)
	if err != nil {
		return nil, err
	}

	user, err := pg.readUserByUserID(ctx, update.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// validateUpdateUserInput validates the input to update a User
func (pg *PostgresDriver) validateUpdateUserInput(ctx context.Context, update types.UpdateUser) error {
	if update.IconURL != nil && *update.IconURL != "" {
		_, err := url.ParseRequestURI(*update.IconURL)
		if err != nil {
			return fmt.Errorf(errInvalidIconURL.Error(), *update.IconURL)
		}
	}

	return nil
}

/* ----- postgresdriver User Delete Methods ----- */

// DeletePortalUser deletes a portal User from the DB. WARNING will do a hard delete.
// Will also delete the user's `account_user_access` and `user_auth_providers` rows.
func (pg *PostgresDriver) DeletePortalUser(ctx context.Context, userID types.UserID) (types.UserID, error) {
	err := pg.validateDeletePortalUserInput(ctx, userID)
	if err != nil {
		return types.UserID(""), err
	}

	deletedUserID, err := pg.DeleteUser(ctx, userID)
	if err != nil {
		return types.UserID(""), err
	}

	return types.UserID(deletedUserID), nil
}

// validateDeletePortalUserInput validates the input to delete a User
func (pg *PostgresDriver) validateDeletePortalUserInput(ctx context.Context, userID types.UserID) error {
	userExists, err := pg.CheckUserExists(ctx, userID)
	if err != nil {
		return err
	}
	if err != nil || !userExists {
		return fmt.Errorf(errUserDoesntExist.Error(), userID)
	}

	// Check if the user is part of any accounts
	userHasAccount, err := pg.CheckUserAccountExists(ctx, userID)
	if err != nil {
		return err
	}
	if userHasAccount {
		return errUserHasAccount
	}

	return nil
}

/* ----- Used by Listener ----- */
func (json dbUser) toOutput() *types.User {
	return &types.User{
		ID:               json.ID,
		IconURL:          json.IconURL,
		Email:            json.Email,
		SignedUp:         json.SignedUp,
		UpdatesProduct:   json.UpdatesProduct,
		UpdatesMarketing: json.UpdatesMarketing,
		BetaTester:       json.BetaTester,
		CreatedAt:        json.CreatedAt,
		UpdatedAt:        json.UpdatedAt,
	}
}

func (json dbUserAuthProvider) toOutput() *types.UserAuthProvider {
	return &types.UserAuthProvider{
		UserID:         json.UserID,
		ProviderUserID: json.ProviderUserID,
		Type:           json.Type,
		Provider:       json.Provider,
		Federated:      json.Federated,
	}
}

type dbUser struct {
	ID               types.UserID `json:"id"`
	IconURL          string       `json:"icon_url"`
	Email            types.Email  `json:"email"`
	SignedUp         bool         `json:"signed_up"`
	UpdatesProduct   bool         `json:"updates_product"`
	UpdatesMarketing bool         `json:"updates_marketing"`
	BetaTester       bool         `json:"beta_tester"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

type dbUserAuthProvider struct {
	ID             int32                `json:"id"`
	UserID         types.UserID         `json:"user_id"`
	Type           types.AuthType       `json:"type"`
	Provider       types.AuthProvider   `json:"provider"`
	ProviderUserID types.ProviderUserID `json:"provider_user_id"`
	Federated      bool                 `json:"federated"`
	CreatedAt      time.Time            `json:"created_at"`
}
