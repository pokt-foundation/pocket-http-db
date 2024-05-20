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

type (
	altruistDBRow struct {
		ChainID  string `json:"chain_id"`
		URL      string `json:"url"`
		Auth     string `json:"auth"`
		AuthType string `json:"auth_type"`
	}

	checkDBRow struct {
		ChainID    string `json:"chain_id"`
		Type       string `json:"type"`
		Payload    string `json:"payload"`
		ResultKey  string `json:"result_key"`
		Allowance  int32  `json:"allowance"`
		EVMChainID int32  `json:"evm_chain_id"`
	}

	ChainRow struct {
		ID             types.RelayChainID `json:"id"`
		IconURL        string             `json:"icon_url"`
		Blockchain     string             `json:"blockchain"`
		Description    string             `json:"description"`
		EnforceResult  string             `json:"enforce_result"`
		Ticker         string             `json:"ticker"`
		Path           pgtype.Text        `json:"path"`
		RequestTimeout pgtype.Int4        `json:"request_timeout"`
		LogLimitBlocks pgtype.Int4        `json:"log_limit_blocks"`
		AllowedMethods []string           `json:"allowed_methods"`
		Active         bool               `json:"active"`
		CreatedAt      pgtype.Timestamptz `json:"created_at"`
		UpdatedAt      pgtype.Timestamptz `json:"updated_at"`
		Deleted        pgtype.Bool        `json:"deleted"`
		DeletedAt      pgtype.Timestamptz `json:"deleted_at"`
		ChainAltruists []byte             `json:"chain_altruists"`
		ChainChecks    []byte             `json:"chain_checks"`
		ChainAliases   []string           `json:"chain_aliases"`
		// DEPRECATED - TODO remove when move to only store aliases is complete
		AliasDomainsMap []byte `json:"alias_domains_map"`
		Overrides       []byte `json:"overrides"`
	}
)

var (
	errChainCannotBeNil     = errors.New("error chain cannot be nil")
	errInvalidAltruistURL   = errors.New("error altruist URL '%s' is an invalid URL")
	errInvalidDomain        = errors.New("error domain '%s' for alias '%s' is invalid")
	errChainExists          = errors.New("error chain already exists for chain ID '%s'")
	errChainDoesntExist     = errors.New("error chain does not exist for chain ID '%s'")
	errUnmarshallingDomains = errors.New("error unmarshalling domains: %w")
)

/* ----- postgresdriver Chain Read Methods ----- */

// ReadChains returns all Chains in the database as Chains structs
func (pg *PostgresDriver) ReadChains(ctx context.Context, options types.DriverOptions) (map[types.RelayChainID]*types.Chain, error) {
	dbChains, err := pg.SelectChains(ctx, options.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	chains := make(map[types.RelayChainID]*types.Chain, len(dbChains))
	for _, dbChain := range dbChains {
		chain, err := dbChain.ToChainRow().toChain()
		if err != nil {
			return nil, err
		}

		chains[types.RelayChainID(dbChain.ID)] = chain
	}

	return chains, nil
}

func (c *SelectChainsRow) ToChainRow() *ChainRow {
	return &ChainRow{
		ID:             c.ID,
		IconURL:        c.IconURL.String,
		Blockchain:     c.Blockchain.String,
		Description:    c.Description.String,
		EnforceResult:  c.EnforceResult.String,
		Ticker:         c.Ticker.String,
		Path:           c.Path,
		RequestTimeout: c.RequestTimeout,
		LogLimitBlocks: c.LogLimitBlocks,
		AllowedMethods: c.AllowedMethods,
		Active:         c.Active,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
		Deleted:        c.Deleted,
		DeletedAt:      c.DeletedAt,
		ChainAliases:   c.ChainAliases,
		ChainAltruists: c.ChainAltruists,
		ChainChecks:    c.ChainChecks,
		Overrides:      c.Overrides,
		// DEPRECATED - TODO remove when move to only store aliases is complete
		AliasDomainsMap: c.AliasDomainsMap,
	}
}

func (c *SelectChainRow) ToChainRow() *ChainRow {
	return &ChainRow{
		ID:             c.ID,
		IconURL:        c.IconURL.String,
		Blockchain:     c.Blockchain.String,
		Description:    c.Description.String,
		EnforceResult:  c.EnforceResult.String,
		Ticker:         c.Ticker.String,
		Path:           c.Path,
		RequestTimeout: c.RequestTimeout,
		LogLimitBlocks: c.LogLimitBlocks,
		AllowedMethods: c.AllowedMethods,
		Active:         c.Active,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
		Deleted:        c.Deleted,
		DeletedAt:      c.DeletedAt,
		ChainAliases:   c.ChainAliases,
		ChainAltruists: c.ChainAltruists,
		ChainChecks:    c.ChainChecks,
		Overrides:      c.Overrides,
		// DEPRECATED - TODO remove when move to only store aliases is complete
		AliasDomainsMap: c.AliasDomainsMap,
	}
}

// toChain converts SelectChainsRow to Chain struct
func (c *ChainRow) toChain() (*types.Chain, error) {
	altruists, err := c.toAltruists()
	if err != nil {
		return nil, err
	}
	checks, err := c.toChecks()
	if err != nil {
		return nil, err
	}
	aliases, err := c.toAliases()
	if err != nil {
		return nil, err
	}
	domains, err := c.toDomains()
	if err != nil {
		return nil, err
	}
	overrides, err := c.toOverrideSession()
	if err != nil {
		return nil, err
	}

	chain := &types.Chain{
		ID:             types.RelayChainID(c.ID),
		IconURL:        c.IconURL,
		Blockchain:     types.ChainAlias(c.Blockchain),
		Description:    c.Description,
		EnforceResult:  c.EnforceResult,
		Ticker:         c.Ticker,
		AllowedMethods: c.AllowedMethods,
		Path:           c.Path.String,
		LogLimitBlocks: c.LogLimitBlocks.Int32,
		RequestTimeout: c.RequestTimeout.Int32,
		Active:         c.Active,
		Aliases:        aliases,
		Altruists:      altruists,
		Checks:         checks,
		Overrides:      overrides,
		// DEPRECATED - TODO remove when move to only store aliases is complete
		AliasDomains: domains,
		CreatedAt:    c.CreatedAt.Time.UTC(),
		UpdatedAt:    c.UpdatedAt.Time.UTC(),
	}

	return chain, nil
}

// toAltruists converts altruists from DB rows to Altruist structs
func (c *ChainRow) toAltruists() (map[types.AltruistURL]types.Altruist, error) {
	var altruistRows []altruistDBRow
	if err := json.Unmarshal(c.ChainAltruists, &altruistRows); err != nil {
		return nil, err
	}

	altruists := make(map[types.AltruistURL]types.Altruist, len(altruistRows))

	for _, altruistRow := range altruistRows {
		url := types.AltruistURL(altruistRow.URL)
		altruists[url] = types.Altruist{
			URL:      types.AltruistURL(altruistRow.URL),
			Auth:     altruistRow.Auth,
			AuthType: types.ChainAuthType(altruistRow.AuthType),
		}
	}

	return altruists, nil
}

// toChecks converts checks from DB rows to Check structs
func (c *ChainRow) toChecks() (map[types.ChainCheckType]types.Check, error) {
	var checkRows []checkDBRow
	if err := json.Unmarshal(c.ChainChecks, &checkRows); err != nil {
		return nil, err
	}

	checks := make(map[types.ChainCheckType]types.Check, len(checkRows))

	for _, checkRow := range checkRows {
		checkType := types.ChainCheckType(checkRow.Type)
		checks[checkType] = types.Check{
			Type:       checkType,
			Payload:    checkRow.Payload,
			ResultKey:  checkRow.ResultKey,
			Allowance:  checkRow.Allowance,
			EVMChainID: checkRow.EVMChainID,
		}
	}

	return checks, nil
}

// toAliases converts chain aliases from DB rows to aliases map
func (c *ChainRow) toAliases() (map[types.ChainAlias]struct{}, error) {
	aliases := make(map[types.ChainAlias]struct{}, len(c.ChainAliases))

	for _, alias := range c.ChainAliases {
		aliases[types.ChainAlias(alias)] = struct{}{}
	}

	return aliases, nil
}

// DEPRECATED - TODO remove when move to only store aliases is complete
// toDomains converts chain aliases and domains from DB rows to alias domains map
func (c *ChainRow) toDomains() (map[types.ChainAlias][]types.ChainDomain, error) {
	var domains map[types.ChainAlias][]types.ChainDomain
	if len(string(c.AliasDomainsMap)) > 2 { // length of empty JSON array in bytes
		if err := json.Unmarshal(c.AliasDomainsMap, &domains); err != nil {
			return nil, fmt.Errorf("%s: %w", errUnmarshallingDomains, err)
		}
	}

	return domains, nil
}

func (c *ChainRow) toOverrideSession() (map[types.ProtocolID]types.OverrideSession, error) {
	var overrideSessionsMap map[types.ProtocolID]types.OverrideSession
	if len(c.Overrides) > 0 {
		if err := json.Unmarshal(c.Overrides, &overrideSessionsMap); err != nil {
			return nil, fmt.Errorf("error unmarshalling override sessions: %w", err)
		}
	}

	return overrideSessionsMap, nil
}

// /* ----- postgresdriver Chain Create Methods ----- */

// WriteChainAndGigastakeApp creates a single Chain along with its GigastakeApps in the database as a single transaction.
// Used specifically for running the `new-chains-ci` and adding a new Chain along with one or more GigastakeApps.
func (pg *PostgresDriver) WriteChainAndGigastakeApps(ctx context.Context, input types.NewChainInput, createdAt time.Time) (*types.NewChainInput, error) {
	err := validateWriteChainAndGigastakeApps(input)
	if err != nil {
		return nil, err
	}

	tx, err := pg.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := pg.WithTx(tx)

	chain := input.Chain
	chain.CreatedAt = createdAt
	chain.UpdatedAt = createdAt

	err = pg.insertChain(ctx, qtx, *chain)
	if err != nil {
		return nil, err
	}

	for _, gigastakeApp := range input.GigastakeApps {
		gigastakeAppID, err := pg.generateID(ctx)
		if err != nil {
			return nil, err
		}

		gigastakeApp.ID = types.GigastakeAppID(gigastakeAppID)
		if gigastakeApp.ChainIDs != nil && len(gigastakeApp.ChainIDs) > 0 {
			gigastakeApp.ChainIDs[chain.ID] = struct{}{}
		} else {
			gigastakeApp.ChainIDs = map[types.RelayChainID]struct{}{chain.ID: {}}
		}
		gigastakeApp.CreatedAt = createdAt
		gigastakeApp.UpdatedAt = createdAt

		err = pg.insertGigastakeApp(ctx, qtx, gigastakeApp)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &input, nil
}

// Validate checks that the fields in NewChainInput are correct
func validateWriteChainAndGigastakeApps(input types.NewChainInput) error {
	if input.Chain == nil {
		return errChainCannotBeNil
	}

	for _, gigastakeApp := range input.GigastakeApps {
		if !gigastakeApp.ProtocolID.IsValid() {
			return fmt.Errorf(errInvalidProtocolID.Error(), gigastakeApp.ProtocolID)
		}
	}

	return nil
}

// WriteChain creates a single Chain in the database
func (pg *PostgresDriver) WriteChain(ctx context.Context, chain types.Chain, createdAt time.Time) (*types.Chain, error) {
	tx, err := pg.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := pg.WithTx(tx)

	chain.CreatedAt = createdAt
	chain.UpdatedAt = createdAt

	err = pg.insertChain(ctx, qtx, chain)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &chain, nil
}

// UpdateChain updates a single Chain in the database
func (pg *PostgresDriver) UpdateChain(ctx context.Context, update types.UpdateChain, updatedAt time.Time) (*types.Chain, error) {
	tx, err := pg.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := pg.WithTx(tx)

	err = pg.updateChain(ctx, qtx, update, updatedAt)
	if err != nil {
		return nil, err
	}

	err = pg.removeUnusedChainRows(ctx, qtx, update)
	if err != nil {
		return nil, err
	}

	updatedChainRow, err := qtx.SelectChain(ctx, update.ID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	updatedChain, err := updatedChainRow.ToChainRow().toChain()
	if err != nil {
		return nil, err
	}

	return updatedChain, nil
}

// insertChain performs either an insert on a chain in the database
func (pg *PostgresDriver) insertChain(ctx context.Context, qtx *Queries, chain types.Chain) error {
	err := pg.validateChainInput(ctx, qtx, chain)
	if err != nil {
		return err
	}

	createdChainID, err := qtx.InsertChain(ctx, InsertChainParams{
		ID:             chain.ID,
		IconURL:        newText(chain.IconURL),
		Blockchain:     newText(string(chain.Blockchain)),
		Description:    newText(chain.Description),
		EnforceResult:  newText(chain.EnforceResult),
		Ticker:         newText(chain.Ticker),
		Path:           newText(chain.Path),
		RequestTimeout: newInt4(chain.RequestTimeout, true),
		LogLimitBlocks: newInt4(chain.LogLimitBlocks, true),
		AllowedMethods: chain.AllowedMethods,
		CreatedAt:      newTimestamptz(chain.CreatedAt),
		UpdatedAt:      newTimestamptz(chain.UpdatedAt),
	})
	if err != nil {
		return err
	}

	for _, altruist := range chain.Altruists {
		err := qtx.UpsertChainAltruist(ctx, UpsertChainAltruistParams{
			ChainID:   createdChainID,
			URL:       altruist.URL,
			AuthType:  altruist.AuthType,
			Auth:      newText(altruist.Auth),
			CreatedAt: newTimestamptz(chain.CreatedAt),
			UpdatedAt: newTimestamptz(chain.UpdatedAt),
		})
		if err != nil {
			return err
		}
	}
	for checkType, check := range chain.Checks {
		err := qtx.UpsertChainCheck(ctx, UpsertChainCheckParams{
			ChainID:    createdChainID,
			Type:       checkType,
			Payload:    newText(check.Payload),
			ResultKey:  newText(check.ResultKey),
			Allowance:  newInt4(check.Allowance, false),
			EVMChainID: newInt4(check.EVMChainID, false),
			CreatedAt:  newTimestamptz(chain.CreatedAt),
			UpdatedAt:  newTimestamptz(chain.UpdatedAt),
		})
		if err != nil {
			return err
		}
	}
	for alias := range chain.Aliases {
		err := qtx.InsertChainAlias(ctx, InsertChainAliasParams{
			ChainID:   createdChainID,
			Alias:     alias,
			CreatedAt: newTimestamptz(chain.UpdatedAt),
		})
		if err != nil {
			return err
		}
	}
	// DEPRECATED - TODO remove when move to only store aliases is complete
	for alias, domains := range chain.AliasDomains {
		domainsStrs := []string{}
		for _, domain := range domains {
			domainsStrs = append(domainsStrs, string(domain))
		}
		err := qtx.UpsertChainAliasDomains(ctx, UpsertChainAliasDomainsParams{
			ChainID:   createdChainID,
			Alias:     alias,
			Domains:   domainsStrs,
			UpdatedAt: newTimestamptz(chain.UpdatedAt),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// updateChain performs either an update on a chain in the database
func (pg *PostgresDriver) updateChain(ctx context.Context, qtx *Queries, update types.UpdateChain, updatedAt time.Time) error {
	err := pg.validateChainUpdate(ctx, qtx, update)
	if err != nil {
		return err
	}

	createdChainID, err := qtx.UpdateChain(ctx, UpdateChainParams{
		ID:             update.ID,
		IconURL:        newNullString(update.IconURL),
		Blockchain:     newNullString((*string)(update.Blockchain)),
		Description:    newNullString(update.Description),
		EnforceResult:  newNullString(update.EnforceResult),
		Ticker:         newNullString(update.Ticker),
		Path:           newNullString(update.Path),
		RequestTimeout: newNullInt(update.RequestTimeout, true),
		LogLimitBlocks: newNullInt(update.LogLimitBlocks, true),
		AllowedMethods: update.AllowedMethods,
		UpdatedAt:      newTimestamptz(updatedAt),
	})
	if err != nil {
		return err
	}

	if update.Altruists != nil {
		for _, altruist := range *update.Altruists {
			err := qtx.UpsertChainAltruist(ctx, UpsertChainAltruistParams{
				ChainID:   createdChainID,
				URL:       altruist.URL,
				AuthType:  altruist.AuthType,
				Auth:      newText(altruist.Auth),
				CreatedAt: newTimestamptz(updatedAt),
				UpdatedAt: newTimestamptz(updatedAt),
			})
			if err != nil {
				return err
			}
		}
	}
	if update.Checks != nil {
		for checkType, check := range *update.Checks {
			err := qtx.UpsertChainCheck(ctx, UpsertChainCheckParams{
				ChainID:    createdChainID,
				Type:       checkType,
				Payload:    newText(check.Payload),
				ResultKey:  newText(check.ResultKey),
				Allowance:  newInt4(check.Allowance, false),
				EVMChainID: newInt4(check.EVMChainID, false),
				CreatedAt:  newTimestamptz(updatedAt),
				UpdatedAt:  newTimestamptz(updatedAt),
			})
			if err != nil {
				return err
			}
		}
	}
	if update.Aliases != nil {
		for alias := range *update.Aliases {
			err := qtx.InsertChainAlias(ctx, InsertChainAliasParams{
				ChainID:   createdChainID,
				Alias:     alias,
				CreatedAt: newTimestamptz(updatedAt),
			})
			if err != nil {
				return err
			}
		}
	}
	// DEPRECATED - TODO remove when move to only store aliases is complete
	if update.AliasDomains != nil {
		for alias, domains := range *update.AliasDomains {
			domainsStrs := []string{}
			for _, domain := range domains {
				domainsStrs = append(domainsStrs, string(domain))
			}
			err := qtx.UpsertChainAliasDomains(ctx, UpsertChainAliasDomainsParams{
				ChainID:   createdChainID,
				Alias:     alias,
				Domains:   domainsStrs,
				UpdatedAt: newTimestamptz(updatedAt),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// validateChainInput performs all necessary data validation checks on incoming Chain data for insert
func (pg *PostgresDriver) validateChainInput(ctx context.Context, qtx *Queries, chain types.Chain) error {
	if chain.IconURL != "" {
		_, err := url.ParseRequestURI(chain.IconURL)
		if err != nil {
			return fmt.Errorf(errInvalidIconURL.Error(), chain.IconURL)
		}
	}

	for url := range chain.Altruists {
		if !url.IsValid() {
			return fmt.Errorf(errInvalidAltruistURL.Error(), url)
		}
	}

	for alias, domains := range chain.AliasDomains {
		for _, domain := range domains {
			if !domain.IsValid() {
				return fmt.Errorf(errInvalidDomain.Error(), domain, alias)
			}
		}
	}

	chainExists, err := qtx.CheckChainExists(ctx, chain.ID)
	if err != nil {
		return err
	}
	if chainExists {
		return fmt.Errorf(errChainExists.Error(), chain.ID)
	}

	return nil
}

// validateChainUpdate performs all necessary data validation checks on incoming Chain data for update
func (pg *PostgresDriver) validateChainUpdate(ctx context.Context, qtx *Queries, chain types.UpdateChain) error {
	if chain.IconURL != nil && *chain.IconURL != "" {
		_, err := url.ParseRequestURI(*chain.IconURL)
		if err != nil {
			return fmt.Errorf(errInvalidIconURL.Error(), *chain.IconURL)
		}
	}

	if chain.Altruists != nil {
		for url := range *chain.Altruists {
			if !url.IsValid() {
				return fmt.Errorf(errInvalidAltruistURL.Error(), url)
			}
		}
	}

	if chain.AliasDomains != nil {
		for alias, domains := range *chain.AliasDomains {
			for _, domain := range domains {
				if !domain.IsValid() {
					return fmt.Errorf(errInvalidDomain.Error(), domain, alias)
				}
			}
		}
	}

	chainExists, err := qtx.CheckChainExists(ctx, chain.ID)
	if err != nil {
		return err
	}
	if !chainExists {
		return fmt.Errorf(errChainDoesntExist.Error(), chain.ID)
	}

	return nil
}

// removeUnusedChainRows removes chain subtables (altruists, checks or alias domains)
// on a Chain update if they are not present in the update data.
// For example:
// - Chain.Altruists = {url_1: <ALTRUIST_1>, url_2: <ALTRUIST_2>} - Chain.Altruists will be set to this map
// - Chain.Altruists = nil - No changes will be made to Chain.Altruists
// - Chain.Altruists = {} - Chain.Altruists will be set to empty
func (pg *PostgresDriver) removeUnusedChainRows(ctx context.Context, qtx *Queries, chain types.UpdateChain) error {
	if chain.Altruists != nil {
		deleteAltruistParams := DeleteUnusedChainAltruistsParams{ChainID: chain.ID}
		for _, altruist := range *chain.Altruists {
			deleteAltruistParams.URLs = append(deleteAltruistParams.URLs, string(altruist.URL))
		}
		err := qtx.DeleteUnusedChainAltruists(ctx, deleteAltruistParams)
		if err != nil {
			return err
		}
	}

	if chain.Checks != nil {
		deleteCheckParams := DeleteUnusedChainChecksParams{ChainID: chain.ID}
		for checkType := range *chain.Checks {
			deleteCheckParams.Types = append(deleteCheckParams.Types, checkType)
		}
		err := qtx.DeleteUnusedChainChecks(ctx, deleteCheckParams)
		if err != nil {
			return err
		}
	}

	if chain.Aliases != nil {
		deleteAliasDomainsParams := DeleteUnusedChainAliasParams{ChainID: chain.ID}
		for alias := range *chain.Aliases {
			deleteAliasDomainsParams.Aliases = append(deleteAliasDomainsParams.Aliases, string(alias))
		}
		err := qtx.DeleteUnusedChainAlias(ctx, deleteAliasDomainsParams)
		if err != nil {
			return err
		}
	}

	// DEPRECATED - TODO remove when move to only store aliases is complete
	if chain.AliasDomains != nil {
		deleteAliasDomainsParams := DeleteUnusedChainAliasDomainsParams{ChainID: chain.ID}
		for alias := range *chain.AliasDomains {
			deleteAliasDomainsParams.Aliases = append(deleteAliasDomainsParams.Aliases, string(alias))
		}
		err := qtx.DeleteUnusedChainAliasDomains(ctx, deleteAliasDomainsParams)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pg *PostgresDriver) SetChainActiveStatus(ctx context.Context, chainID types.RelayChainID, active bool, updatedAt time.Time) (bool, error) {
	chainExists, err := pg.CheckChainExists(ctx, chainID)
	if err != nil {
		return false, err
	}
	if !chainExists {
		return false, fmt.Errorf(errChainDoesntExist.Error(), chainID)
	}

	params := UpdateChainActiveParams{ID: chainID, Active: active, UpdatedAt: newTimestamptz(updatedAt)}

	activeStatus, err := pg.UpdateChainActive(ctx, params)
	if err != nil {
		return false, err
	}

	return activeStatus, nil
}

/* ----- Used by Listener ----- */
func (json dbChain) toOutput() *types.Chain {
	return &types.Chain{
		ID:             json.ID,
		IconURL:        json.IconURL,
		Blockchain:     json.Blockchain,
		Description:    json.Description,
		EnforceResult:  json.EnforceResult,
		Path:           json.Path,
		Ticker:         json.Ticker,
		AllowedMethods: json.AllowedMethods,
		LogLimitBlocks: json.LogLimitBlocks,
		RequestTimeout: json.RequestTimeout,
		Active:         json.Active,
		CreatedAt:      json.CreatedAt,
		UpdatedAt:      json.UpdatedAt,
	}
}

func (json dbChainAltruist) toOutput() *types.Altruist {
	return &types.Altruist{
		ChainID:  json.ChainID,
		URL:      json.URL,
		Auth:     json.Auth,
		AuthType: json.AuthType,
	}
}

func (json dbChainCheck) toOutput() *types.Check {
	return &types.Check{
		ChainID:    json.ChainID,
		Type:       json.Type,
		Payload:    json.Payload,
		ResultKey:  json.ResultKey,
		Allowance:  json.Allowance,
		EVMChainID: json.EVMChainID,
	}
}

func (json dbChainAlias) toOutput() *types.Alias {
	return &types.Alias{
		ChainID: json.ChainID,
		Alias:   json.Alias,
	}
}

// DEPRECATED - TODO remove when move to only store aliases is complete
func (json dbChainAliasDomains) toOutput() *types.AliasDomains {
	return &types.AliasDomains{
		ChainID: json.ChainID,
		Alias:   json.Alias,
		Domains: json.Domains,
	}
}

type dbChain struct {
	ID                       types.RelayChainID  `json:"id"`
	IconURL                  string              `json:"icon_url"`
	Blockchain               types.ChainAlias    `json:"blockchain"`
	Description              string              `json:"description"`
	EnforceResult            string              `json:"enforce_result"`
	Ticker                   string              `json:"ticker"`
	Path                     string              `json:"path"`
	RequestTimeout           int32               `json:"request_timeout"`
	LogLimitBlocks           int32               `json:"log_limit_blocks"`
	ChainAliases             []string            `json:"chain_aliases"`
	AllowedMethods           []string            `json:"allowed_methods"`
	GigastakeRedirectDomains []types.ChainDomain `json:"gigastake_redirect_domains"`
	Active                   bool                `json:"active"`
	CreatedAt                time.Time           `json:"created_at"`
	UpdatedAt                time.Time           `json:"updated_at"`
	Deleted                  bool                `json:"deleted"`
	DeletedAt                time.Time           `json:"deleted_at"`
}

type dbChainAltruist struct {
	ID        int32               `json:"id"`
	ChainID   types.RelayChainID  `json:"chain_id"`
	URL       types.AltruistURL   `json:"url"`
	Auth      string              `json:"auth"`
	AuthType  types.ChainAuthType `json:"auth_type"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type dbChainCheck struct {
	ID         int32                `json:"id"`
	ChainID    types.RelayChainID   `json:"chain_id"`
	Type       types.ChainCheckType `json:"type"`
	Payload    string               `json:"payload"`
	ResultKey  string               `json:"result_key"`
	Allowance  int32                `json:"allowance"`
	EVMChainID int32                `json:"evm_chain_id"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
}

type dbChainAlias struct {
	ChainID types.RelayChainID `json:"chain_id"`
	Alias   types.ChainAlias   `json:"alias"`
}

// DEPRECATED - TODO remove when move to only store aliases is complete
type dbChainAliasDomains struct {
	ChainID types.RelayChainID  `json:"chain_id"`
	Alias   types.ChainAlias    `json:"alias"`
	Domains []types.ChainDomain `json:"domains"`
}
