package postgresdriver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/types"
)

var (
	errNoAddress              = errors.New("error blockchain address must be provided")
	errContractDoesntExist    = errors.New("error blockchain address %s does not exist")
	errContractAlreadyBlocked = errors.New("error blockchain address %s is already blocked")
)

/* ----- postgresdriver GlobalBlockedContracts Read Methods ----- */

// ReadBlockedContracts returns all GlobalBlockedContracts in the DB
func (pg *PostgresDriver) ReadBlockedContracts(ctx context.Context) (types.GlobalBlockedContracts, error) {
	dbBlockedContracts, err := pg.SelectGlobalBlockedContract(ctx)
	if err != nil {
		switch {
		case errNoRows(err):
			return types.GlobalBlockedContracts{}, nil
		default:
			return types.GlobalBlockedContracts{}, err
		}
	}

	blockedContracts := types.GlobalBlockedContracts{
		BlockedAddresses: make(map[types.BlockedAddress]struct{}, len(dbBlockedContracts)),
	}

	for _, contract := range dbBlockedContracts {
		blockedContracts.BlockedAddresses[contract.BlockedAddress] = struct{}{}
	}

	return blockedContracts, nil
}

/* ----- postgresdriver GlobalBlockedContracts Methods ----- */

// WriteBlockedContract adds a new blocked address to the global blocked contracts table.
func (pg *PostgresDriver) WriteBlockedContract(ctx context.Context, blockedAddress types.BlockedAddress, createdAt time.Time) error {
	if blockedAddress == "" {
		return errNoAddress
	}

	params := AddGlobalBlockedContractParams{
		BlockedAddress: blockedAddress,
		CreatedAt:      newTimestamptz(createdAt),
		UpdatedAt:      newTimestamptz(createdAt),
	}

	err := pg.AddGlobalBlockedContract(ctx, params)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "duplicate key value violates unique constraint"):
			return fmt.Errorf(errContractAlreadyBlocked.Error(), blockedAddress)
		default:
			return err
		}
	}

	return nil
}

// UpdateBlockedContractActive activates or deactives a blocked address in the global blocked contracts table.
func (pg *PostgresDriver) UpdateBlockedContractActive(ctx context.Context, blockedAddress types.BlockedAddress, active bool, updatedAt time.Time) error {
	if blockedAddress == "" {
		return errNoAddress
	}

	params := SetGlobalBlockedContractActiveParams{
		BlockedAddress: blockedAddress,
		Active:         active,
		UpdatedAt:      newTimestamptz(updatedAt),
	}

	_, err := pg.SetGlobalBlockedContractActive(ctx, params)
	if err != nil {
		switch {
		case errNoRows(err):
			return fmt.Errorf(errContractDoesntExist.Error(), blockedAddress)
		default:
			return err
		}
	}

	return nil
}

// RemoveBlockedContract deletes a blocked address from the global blocked contracts table.
func (pg *PostgresDriver) RemoveBlockedContract(ctx context.Context, blockedAddress types.BlockedAddress) error {
	if blockedAddress == "" {
		return errNoAddress
	}

	_, err := pg.RemoveGlobalBlockedContract(ctx, blockedAddress)
	if err != nil {
		switch {
		case errNoRows(err):
			return fmt.Errorf(errContractDoesntExist.Error(), blockedAddress)
		default:
			return err
		}
	}

	return nil
}

/* ----- Used by Listener ----- */
func (gbc dbGlobalBlockedContract) toOutput() *types.BlockedContract {
	return &types.BlockedContract{
		BlockedAddress: gbc.BlockedAddress,
		Active:         gbc.Active,
	}
}

type dbGlobalBlockedContract struct {
	ID             int32                `json:"id"`
	BlockedAddress types.BlockedAddress `json:"blocked_address"`
	Active         bool                 `json:"active"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}
