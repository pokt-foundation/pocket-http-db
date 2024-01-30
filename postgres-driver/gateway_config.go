package postgresdriver

import (
	"context"
	"encoding/json"

	"github.com/pokt-foundation/portal-http-db/v2/types"
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

// ReadConfigJSON reads the raw JSON config from the database
func (q *Queries) ReadConfigJSON(ctx context.Context) (json.RawMessage, error) {
	configJSON, err := q.SelectConfiguration(ctx)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(configJSON), nil
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

type dbChainConfig struct {
	ChainID        types.RelayChainID
	ValidRange     int32
	LogLimitBlocks int32
}

type dbMethodValidator struct {
	Method    types.Method
	Validator string
}
