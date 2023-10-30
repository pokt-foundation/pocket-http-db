package driver

import (
	"context"
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/types"
)

type (
	// The Driver interface represents all database operations required by the Portal HTTP DB
	Driver interface {
		Reader
		Writer
	}

	// Pinger ensures to provide a healthcheck to perform against the db
	Pinger interface {
		Ping(ctx context.Context) error
	}

	Reader interface {
		/* ReadAccounts returns all Accounts in the database. Can specify if deleted Accounts should be included. */
		ReadAccounts(ctx context.Context, options types.DriverOptions) (map[types.AccountID]*types.Account, error)
		/* ReadChains returns all Chains in the databas. Can specify if deleted Chains should be included. */
		ReadChains(ctx context.Context, options types.DriverOptions) (map[types.RelayChainID]*types.Chain, error)
		/* ReadPortalApps returns all PortalApps in the database. Can specify if deleted PortalApps should be included.  */
		ReadPortalApps(ctx context.Context, options types.DriverOptions) (map[types.PortalAppID]*types.PortalApp, error)
		/* ReadGigastakeApps returns all GigastakeApps in the database. Can specify if deleted GigastakeApps should be included.  */
		ReadGigastakeApps(ctx context.Context, options types.DriverOptions) (map[types.GigastakeAppID]*types.GigastakeApp, error)

		/* ReadPlans returns all Plans in the database */
		ReadPlans(ctx context.Context) (map[types.PayPlanType]*types.Plan, error)

		/* ReadAllUsers returns all users as a map that takes the form map[types.UserID]*types.User */
		ReadAllUsers(ctx context.Context) (map[types.UserID]*types.User, error)

		/* ReadBlockedContracts returns all GlobalBlockedContracts in the DB */
		ReadBlockedContracts(ctx context.Context) (types.GlobalBlockedContracts, error)

		NotificationChannel() <-chan *types.Notification

		Pinger
	}

	Writer interface {
		/* WritePortalApp saves input PortalApp to the database. */
		WritePortalApp(ctx context.Context, portalApp types.PortalApp, aat types.AAT, createdAt time.Time) (*types.PortalApp, error)
		/* UpdateLoadBalancer updates PortalApp and related table rows. */
		UpdatePortalApp(ctx context.Context, update types.UpdatePortalApp, updatedAt time.Time) error
		/* SetPortalAppDeleted sets the portal app Deleted field to true. */
		SetPortalAppDeleted(ctx context.Context, portalAppID types.PortalAppID, deletedAt time.Time) error

		/* WriteUserNewSignUp creates a new portal User in the DB from a CreateUser input. */
		WriteUserNewSignUp(ctx context.Context, user types.CreateUser, createdAt time.Time) (*types.User, types.AccountID, error)
		/* DeletePortalUser deletes a portal User from the DB. WARNING will do a full delete in the case of users. */
		DeletePortalUser(ctx context.Context, userID types.UserID) (types.UserID, error)

		/* WriteAccount saves input Account to the database. */
		WriteAccount(ctx context.Context, creatorID types.UserID, account types.Account, createdAt time.Time) (*types.Account, error)
		/* UpdateAccount updates an Account. */
		UpdateAccount(ctx context.Context, update types.UpdateAccount, updatedAt time.Time) (*types.Account, error)
		/* UpsertAccountIntegration adds or updates the Account Integrations for a single Account. */
		UpsertAccountIntegration(ctx context.Context, integrations types.AccountIntegrations) (*types.AccountIntegrations, error)
		/* SetAccountDeleted sets the Account Deleted field to true. */
		SetAccountDeleted(ctx context.Context, accountID types.AccountID, deletedAt time.Time) error
		/* UpdateAccountsFirstDateSurpassed updates multiple Accounts firstDateSurpassed field. */
		UpdateAccountsFirstDateSurpassed(ctx context.Context, update *types.UpdateFirstDateSurpassed) error

		/* WriteAccountUser saves input AccountUser to the database. */
		WriteAccountUser(ctx context.Context, createAccountUser types.CreateAccountUser, createdAt time.Time) (types.UserID, error)
		/* SetAccountUserRole updates the role for an existing AccountUser row. If transferring ownership the account owner becomes an admin. */
		SetAccountUserRole(ctx context.Context, updateAccountUser types.UpdateAccountUserRole, updatedAt time.Time) error
		/* UpdateAcceptAccountUser sets the User ID and the Accepted field to true for an AccountUser row. */
		UpdateAcceptAccountUser(ctx context.Context, acceptAccountUser types.UpdateAcceptAccountUser, updatedAt time.Time) error
		/* RemoveAccountUser deletes a AccountUser row for a given user and account ID. */
		RemoveAccountUser(ctx context.Context, userID types.UserID, accountID types.AccountID) error

		/* UpdateUser updates a user's updatable fields in the database. */
		UpdateUser(ctx context.Context, update types.UpdateUser, updatedAt time.Time) (*types.User, error)

		/* WriteChainAndGigastakeApps saves input Chain and one or more GigastakeApp structs to the database as one transaction. */
		WriteChainAndGigastakeApps(ctx context.Context, input types.NewChainInput, createdAt time.Time) (*types.NewChainInput, error)
		/* WriteGigastakeApp saves input GigastakeApp and AAT to the database as one transaction. */
		WriteGigastakeApp(ctx context.Context, gigastakeApp types.GigastakeApp, createdAt time.Time) (*types.GigastakeApp, error)
		/* WriteChain saves input Chain struct to the database. */
		WriteChain(ctx context.Context, chain types.Chain, createdAt time.Time) (*types.Chain, error)
		/* UpdateChain updates Chain and ChainCheck for a given Chain. */
		UpdateChain(ctx context.Context, chain types.UpdateChain, updatedAt time.Time) (*types.Chain, error)
		/* UpdateGigastakeApp updates fields on a GigastakeApp. */
		UpdateGigastakeApp(ctx context.Context, gigastakeApp types.UpdateGigastakeApp, updatedAt time.Time) error
		/* ActivateChain toggles Chain.Active field on or off. */
		SetChainActiveStatus(ctx context.Context, chainID types.RelayChainID, active bool, updatedAt time.Time) (bool, error)

		/* WriteBlockedContract adds a new blocked address to the global blocked contracts table. */
		WriteBlockedContract(ctx context.Context, blockedAddress types.BlockedAddress, createdAt time.Time) error
		/* UpdateBlockedContractActive activates or deactives a blocked address in the global blocked contracts table. */
		UpdateBlockedContractActive(ctx context.Context, blockedAddress types.BlockedAddress, active bool, updatedAt time.Time) error
		/* RemoveBlockedContract deletes a blocked address from the global blocked contracts table. */
		RemoveBlockedContract(ctx context.Context, blockedAddress types.BlockedAddress) error

		Ping(ctx context.Context) error
	}
)
