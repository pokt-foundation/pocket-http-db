package postgresdriver

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/pokt-foundation/portal-http-db/v2/types"
)

var (
	errGatewayConfigEmptyEnv       = errors.New("Env cannot be empty for gateway config")
	errGatewayConfigInvalidVersion = errors.New("Version should be greater than 0 for gateway config")
)

// ReadValidatorConfig reads the method validator plugin config from the database
func (q *Queries) ReadValidatorConfig(ctx context.Context) (*types.ValidatorPluginConfig, error) {
	chainConfigs, err := q.SelectValidatorChainConfig(ctx)
	if err != nil {
		return nil, err
	}

	methodConfigs, err := q.SelectValidatorMethodConfig(ctx)
	if err != nil {
		return nil, err
	}

	config := &types.ValidatorPluginConfig{
		ChainsValidRanges:    make(map[types.RelayChainID]int32),
		ChainsLogLimitBlocks: make(map[types.RelayChainID]int32),
		ArchivalChains:       make(map[types.RelayChainID]types.RelayChainID),
		MethodValidatorsMap:  make(map[types.Method][]types.ValidatorMethodStr),
	}

	for _, cc := range chainConfigs {
		config.ChainsValidRanges[types.RelayChainID(cc.ChainID.String)] = cc.ValidRange.Int32
		config.ChainsLogLimitBlocks[types.RelayChainID(cc.ChainID.String)] = cc.LogLimitBlocks.Int32
		if cc.ArchivalChainID.String != "" {
			config.ArchivalChains[types.RelayChainID(cc.ChainID.String)] = types.RelayChainID(cc.ArchivalChainID.String)
		}
	}

	for _, mc := range methodConfigs {
		config.MethodValidatorsMap[types.Method(mc.Method)] = append(config.MethodValidatorsMap[types.Method(mc.Method)], types.ValidatorMethodStr(mc.Validator.ValidatorMethodStr))
	}

	return config, nil
}

// ReadGatewayConfig reads the raw JSON config from the database
func (q *Queries) ReadGatewayConfig(ctx context.Context) (map[types.Environment]types.GatewayConfig, error) {
	latestGatewayConfigs, err := q.SelectLatestConfigurations(ctx)
	if err != nil {
		return nil, err
	}

	gatewayConfigs := make(map[types.Environment]types.GatewayConfig, len(latestGatewayConfigs))

	for _, gatewayConfig := range latestGatewayConfigs {
		gatewayConfigs[gatewayConfig.Environment] = types.GatewayConfig{
			Environment: gatewayConfig.Environment,
			Version:     int(gatewayConfig.Version),
			Description: gatewayConfig.Description,
			Config:      json.RawMessage(gatewayConfig.Configuration),
		}
	}
	return gatewayConfigs, nil
}

// ReadOpenAPISpecs reads the chain OpenAPI specs from the database
func (q *Queries) ReadOpenAPISpecs(ctx context.Context) (map[types.RelayChainID]json.RawMessage, error) {
	specs, err := q.SelectOpenAPISpecs(ctx)
	if err != nil {
		return nil, err
	}

	chainSpecs := make(map[types.RelayChainID]json.RawMessage)
	for _, spec := range specs {
		chainSpecs[types.RelayChainID(spec.ChainID)] = json.RawMessage(spec.OpenApiSpec)
	}

	return chainSpecs, nil
}

/* ----- postgresdriver Gateway Config Create Methods ----- */

// CreateGatewayConfig creates a new gateway config entry in the database.
func (pg *PostgresDriver) CreateGatewayConfig(ctx context.Context, config *types.GatewayConfig) (*types.GatewayConfig, error) {
	err := pg.validateGatewayConfigInput(ctx, config)
	if err != nil {
		return nil, err
	}

	tx, err := pg.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := pg.WithTx(tx)

	_, err = qtx.InsertGatewayConfig(ctx, InsertGatewayConfigParams{
		Environment:   config.Environment,
		Version:       int32(config.Version),
		Description:   config.Description,
		Configuration: config.Config,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// validateGatewayConfigInput performs all necessary data validation checks on incoming GatewayConfig data
func (pg *PostgresDriver) validateGatewayConfigInput(ctx context.Context, config *types.GatewayConfig) error {
	if config.Environment == "" {
		return errGatewayConfigEmptyEnv
	}

	if config.Version <= 0 {
		return errGatewayConfigInvalidVersion
	}

	return nil
}

/* ----- Used by Listener ----- */

func (cc dbChainConfig) toOutput() *types.ValidatorChainConfig {
	return &types.ValidatorChainConfig{
		ChainID:        types.RelayChainID(cc.ChainID),
		ValidRange:     cc.ValidRange,
		LogLimitBlocks: cc.LogLimitBlocks,
	}
}

func (mv dbMethodValidator) toOutput() *types.ValidatorMethodConfig {
	return &types.ValidatorMethodConfig{
		Method:    types.Method(mv.Method),
		Validator: types.ValidatorMethodStr(mv.Validator),
	}
}

func (c dbGatewayConfig) toOutput() *types.GatewayConfig {
	return &types.GatewayConfig{
		Environment: types.Environment(c.Environment),
		Description: c.Description,
		Version:     int(c.Version),
		Config:      c.Configuration,
	}
}

type dbChainConfig struct {
	ChainID        types.RelayChainID `json:"chain_id"`
	ValidRange     int32              `json:"valid_range"`
	LogLimitBlocks int32              `json:"log_limit_blocks"`
}

type dbMethodValidator struct {
	Method    types.Method `json:"method"`
	Validator string       `json:"validator"`
}

type dbGatewayConfig struct {
	Environment   types.Environment `json:"environment"`
	Description   string            `json:"description"`
	Version       int32             `json:"version"`
	Configuration json.RawMessage   `json:"configuration"`
}
