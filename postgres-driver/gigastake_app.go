package postgresdriver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/types"
)

var (
	errEmptyChainIDsForGigastakeApp = errors.New("chainIDs cannot be empty for gigastake app")
)

/* ----- postgresdriver GigastakeApp Read Methods ----- */

// ReadGigastakeApps returns all gigastake applications in the database as GigastakeApp structs
func (pg *PostgresDriver) ReadGigastakeApps(ctx context.Context, options types.DriverOptions) (map[types.GigastakeAppID]*types.GigastakeApp, error) {
	dbGigastakeApps, err := pg.SelectGigastakeApplications(ctx)
	if err != nil {
		return nil, err
	}

	gigastakeApps := make(map[types.GigastakeAppID]*types.GigastakeApp, len(dbGigastakeApps))
	for _, dbGigastakeApp := range dbGigastakeApps {
		gigastakeApp, err := dbGigastakeApp.toGigastakeApp()
		if err != nil {
			return nil, err
		}

		gigastakeApps[dbGigastakeApp.ID] = gigastakeApp
	}

	return gigastakeApps, nil
}

// toGigastakeApp converts GigastakeApp SELECT output to GigastakeApp struct
func (g *SelectGigastakeApplicationsRow) toGigastakeApp() (*types.GigastakeApp, error) {
	chainIDs := make(map[types.RelayChainID]struct{}, len(g.ChainIDs))
	for _, chainID := range g.ChainIDs {
		chainIDs[types.RelayChainID(chainID)] = struct{}{}
	}

	return &types.GigastakeApp{
		ID:              g.ID,
		ChainIDs:        chainIDs,
		Name:            g.Name,
		Address:         g.Address,
		PublicKey:       types.GigastakeAppPublicKey(g.PublicKey),
		ClientPublicKey: g.ClientPublicKey,
		Signature:       g.Signature,
		Version:         g.Version,
		CreatedAt:       g.CreatedAt.Time.UTC(),
		UpdatedAt:       g.UpdatedAt.Time.UTC(),
		Deleted:         g.Deleted,
	}, nil
}

/* ----- postgresdriver GigastakeApp Create Methods ----- */

// WriteGigastakeApp creates a single GigastakeApp in the database
func (pg *PostgresDriver) WriteGigastakeApp(ctx context.Context, gigastakeApp *types.GigastakeApp, createdAt time.Time) (*types.GigastakeApp, error) {
	err := pg.validateGigastakeAppInput(ctx, gigastakeApp)
	if err != nil {
		return nil, err
	}

	gigastakeAppID, err := pg.generateID(ctx)
	if err != nil {
		return nil, err
	}

	gigastakeApp.ID = types.GigastakeAppID(gigastakeAppID)
	gigastakeApp.CreatedAt = createdAt
	gigastakeApp.UpdatedAt = createdAt

	tx, err := pg.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := pg.WithTx(tx)

	err = pg.insertGigastakeApp(ctx, qtx, gigastakeApp)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return gigastakeApp, nil
}

// insertGigastakeApp performs an insert on a GigastakeApp in the database.
// It must take a transaction as it is used in both WriteGigastakeApp and WriteChainAndGigastakeApps
func (pg *PostgresDriver) insertGigastakeApp(ctx context.Context, qtx *Queries, gigastakeApp *types.GigastakeApp) error {
	chainIDs := []string{}
	for chainID := range gigastakeApp.ChainIDs {
		chainIDs = append(chainIDs, string(chainID))
	}

	gigastakesCount, err := pg.GetGigastakeApplicationsCount(ctx)
	if err != nil {
		return err
	}

	gigastakeApp.Name = pg.getNextGigastakeName(gigastakesCount, gigastakeApp.ID)

	err = qtx.InsertGigastakeApp(ctx, InsertGigastakeAppParams{
		ID:              gigastakeApp.ID,
		ChainIDs:        chainIDs,
		Name:            gigastakeApp.Name,
		Address:         gigastakeApp.Address,
		PublicKey:       string(gigastakeApp.PublicKey),
		ClientPublicKey: gigastakeApp.ClientPublicKey,
		Signature:       gigastakeApp.Signature,
		Version:         gigastakeApp.Version,
		CreatedAt:       newTimestamptz(gigastakeApp.CreatedAt),
		UpdatedAt:       newTimestamptz(gigastakeApp.UpdatedAt),
	})
	if err != nil {
		return err
	}

	err = qtx.UpdateGigastakeAppChainIDs(ctx, UpdateGigastakeAppChainIDsParams{
		ID:        gigastakeApp.ID,
		ChainIDs:  chainIDs,
		UpdatedAt: newTimestamptz(gigastakeApp.UpdatedAt),
	})
	if err != nil {
		return err
	}

	return nil
}

// getNextGigastakeName generates a new gigastake name and checks if the name already exists in the DB
// getNextGigastakeName generates a new gigastake name and checks if the name already exists in the DB
// If the name can't be fetched, it returns the fallback name so it will always return a string
func (pg *PostgresDriver) getNextGigastakeName(latestEntry int64, fallback types.GigastakeAppID) string {
	number := latestEntry + 1

	for i := 0; i < 10; i++ {
		name, err := generateNextGigastakeName(number, fallback)
		if err == nil {
			exists, err := pg.CheckGigastakeAppNameExists(context.Background(), name)
			if err == nil && !exists {
				return name
			}
		}

		number++
	}

	return string(fallback)
}

// generateNextGigastakeName gets the next name for a gigastake. This field is only for human readability.
// If the name can't be fetched, the calling function returns the fallback name so it will always return a string
func generateNextGigastakeName(entry int64, fallback types.GigastakeAppID) (string, error) {
	// I want to be the very best, like no one ever was
	// To catch them is my real test; to train them is my cause!
	resp, err := http.Get(fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%d", entry))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return string(fallback), nil
	}

	return response.Name, nil
}

// validateGigastakeAppInput performs all necessary data validation checks on incoming GigastakeApp data
func (pg *PostgresDriver) validateGigastakeAppInput(ctx context.Context, gigastakeApp *types.GigastakeApp) error {
	if len(gigastakeApp.ChainIDs) == 0 {
		return errEmptyChainIDsForGigastakeApp
	}

	for chainID := range gigastakeApp.ChainIDs {
		chainExists, err := pg.CheckChainExists(ctx, chainID)
		if err != nil {
			return err
		}
		if !chainExists {
			return fmt.Errorf(errChainDoesntExist.Error(), chainID)
		}
	}

	return nil
}

/* ----- postgresdriver GigastakeApp Create Methods ----- */

// UpdateGigastakeApp updates a single GigastakeApp in the database.
func (pg *PostgresDriver) UpdateGigastakeApp(ctx context.Context, gigastakeAppUpdate types.UpdateGigastakeApp, updatedAt time.Time) error {
	err := pg.validateUpdateGigastakeAppInput(ctx, gigastakeAppUpdate)
	if err != nil {
		return err
	}

	chainIDs := []string{}
	for _, chainID := range gigastakeAppUpdate.ChainIDs {
		chainIDs = append(chainIDs, string(chainID))
	}

	err = pg.UpdateGigastakeAppChainIDs(ctx, UpdateGigastakeAppChainIDsParams{
		ID:        gigastakeAppUpdate.ID,
		ChainIDs:  chainIDs,
		UpdatedAt: newTimestamptz(updatedAt),
	})
	if err != nil {
		return err
	}

	return nil
}

// validateGigastakeAppInput performs all necessary data validation checks on incoming GigastakeApp data
func (pg *PostgresDriver) validateUpdateGigastakeAppInput(ctx context.Context, gigastakeAppUpdate types.UpdateGigastakeApp) error {
	if len(gigastakeAppUpdate.ChainIDs) == 0 {
		return errEmptyChainIDsForGigastakeApp
	}

	for _, chainID := range gigastakeAppUpdate.ChainIDs {
		chainExists, err := pg.CheckChainExists(ctx, chainID)
		if err != nil {
			return err
		}
		if !chainExists {
			return fmt.Errorf(errChainDoesntExist.Error(), chainID)
		}
	}

	return nil
}

/* ----- Used by Listener ----- */
func (json dbGigastakeApp) toOutput() *types.GigastakeApp {
	return &types.GigastakeApp{
		ID:              json.ID,
		Name:            json.Name,
		Address:         json.Address,
		PublicKey:       types.GigastakeAppPublicKey(json.PublicKey),
		ClientPublicKey: json.ClientPublicKey,
		Signature:       json.Signature,
		Version:         json.Version,
		CreatedAt:       json.CreatedAt,
		UpdatedAt:       json.UpdatedAt,
		Deleted:         json.Deleted,
	}
}

func (json dbChainGigastakeApp) toOutput() *types.ChainGigastakeApp {
	return &types.ChainGigastakeApp{
		GigastakeAppID: json.GigastakeAppID,
		ChainID:        json.ChainID,
	}
}

type dbGigastakeApp struct {
	ID              types.GigastakeAppID `json:"id"`
	Name            string               `json:"name"`
	Address         string               `json:"address"`
	PublicKey       string               `json:"public_key"`
	ClientPublicKey string               `json:"client_public_key"`
	Signature       string               `json:"signature"`
	Version         string               `json:"version"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	Deleted         bool                 `json:"deleted"`
}

type dbChainGigastakeApp struct {
	GigastakeAppID types.GigastakeAppID `json:"gigastake_application_id"`
	ChainID        types.RelayChainID   `json:"chain_id"`
}
