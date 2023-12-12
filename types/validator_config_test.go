package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ValidatorPluginConfig_UpdateChainConfigs(t *testing.T) {
	tests := []struct {
		name                   string
		initialConfig          *ValidatorPluginConfig
		chainConfig            *ValidatorChainConfig
		expectedValidRanges    map[RelayChainID]int32
		expectedLogLimitBlocks map[RelayChainID]int32
	}{
		{
			name: "Should update ChainsValidRanges and ChainsLogLimitBlocks with the provided ValidatorChainConfig",
			initialConfig: &ValidatorPluginConfig{
				ChainsValidRanges:    map[RelayChainID]int32{},
				ChainsLogLimitBlocks: map[RelayChainID]int32{},
			},
			chainConfig: &ValidatorChainConfig{
				ChainID:        "0001",
				ValidRange:     100,
				LogLimitBlocks: 500,
			},
			expectedValidRanges: map[RelayChainID]int32{
				"0001": 100,
			},
			expectedLogLimitBlocks: map[RelayChainID]int32{
				"0001": 500,
			},
		},
		{
			name: "Should update existing ChainsValidRanges and ChainsLogLimitBlocks with the provided ValidatorChainConfig",
			initialConfig: &ValidatorPluginConfig{
				ChainsValidRanges: map[RelayChainID]int32{
					"0001": 50,
				},
				ChainsLogLimitBlocks: map[RelayChainID]int32{
					"0001": 250,
				},
			},
			chainConfig: &ValidatorChainConfig{
				ChainID:        "0001",
				ValidRange:     100,
				LogLimitBlocks: 500,
			},
			expectedValidRanges: map[RelayChainID]int32{
				"0001": 100,
			},
			expectedLogLimitBlocks: map[RelayChainID]int32{
				"0001": 500,
			},
		},
		{
			name: "Should not affect other chains when updating a specific chain",
			initialConfig: &ValidatorPluginConfig{
				ChainsValidRanges: map[RelayChainID]int32{
					"0001": 50,
					"0002": 150,
				},
				ChainsLogLimitBlocks: map[RelayChainID]int32{
					"0001": 250,
					"0002": 350,
				},
			},
			chainConfig: &ValidatorChainConfig{
				ChainID:        "0001",
				ValidRange:     100,
				LogLimitBlocks: 500,
			},
			expectedValidRanges: map[RelayChainID]int32{
				"0001": 100,
				"0002": 150,
			},
			expectedLogLimitBlocks: map[RelayChainID]int32{
				"0001": 500,
				"0002": 350,
			},
		},
		{
			name: "Should add new chain to ChainsValidRanges and ChainsLogLimitBlocks with the provided ValidatorChainConfig",
			initialConfig: &ValidatorPluginConfig{
				ChainsValidRanges:    map[RelayChainID]int32{},
				ChainsLogLimitBlocks: map[RelayChainID]int32{},
			},
			chainConfig: &ValidatorChainConfig{
				ChainID:        "0003",
				ValidRange:     200,
				LogLimitBlocks: 600,
			},
			expectedValidRanges: map[RelayChainID]int32{
				"0003": 200,
			},
			expectedLogLimitBlocks: map[RelayChainID]int32{
				"0003": 600,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.initialConfig.UpdateChainConfigs(test.chainConfig)
			assert.Equal(t, test.expectedValidRanges, test.initialConfig.ChainsValidRanges)
			assert.Equal(t, test.expectedLogLimitBlocks, test.initialConfig.ChainsLogLimitBlocks)
		})
	}
}

func Test_ValidatorPluginConfig_RemoveChainConfigs(t *testing.T) {
	tests := []struct {
		name                   string
		initialConfig          *ValidatorPluginConfig
		chainID                RelayChainID
		expectedValidRanges    map[RelayChainID]int32
		expectedLogLimitBlocks map[RelayChainID]int32
	}{
		{
			name: "Should remove chain from ChainsValidRanges and ChainsLogLimitBlocks",
			initialConfig: &ValidatorPluginConfig{
				ChainsValidRanges: map[RelayChainID]int32{
					"0001": 100,
					"0002": 200,
				},
				ChainsLogLimitBlocks: map[RelayChainID]int32{
					"0001": 500,
					"0002": 600,
				},
			},
			chainID: "0001",
			expectedValidRanges: map[RelayChainID]int32{
				"0002": 200,
			},
			expectedLogLimitBlocks: map[RelayChainID]int32{
				"0002": 600,
			},
		},
		{
			name: "Should not affect other chains when removing a specific chain",
			initialConfig: &ValidatorPluginConfig{
				ChainsValidRanges: map[RelayChainID]int32{
					"0001": 100,
					"0002": 200,
					"0003": 300,
				},
				ChainsLogLimitBlocks: map[RelayChainID]int32{
					"0001": 500,
					"0002": 600,
					"0003": 700,
				},
			},
			chainID: "0002",
			expectedValidRanges: map[RelayChainID]int32{
				"0001": 100,
				"0003": 300,
			},
			expectedLogLimitBlocks: map[RelayChainID]int32{
				"0001": 500,
				"0003": 700,
			},
		},
		{
			name: "Should handle case when chain is not in the map",
			initialConfig: &ValidatorPluginConfig{
				ChainsValidRanges: map[RelayChainID]int32{
					"0001": 100,
					"0002": 200,
				},
				ChainsLogLimitBlocks: map[RelayChainID]int32{
					"0001": 500,
					"0002": 600,
				},
			},
			chainID: "0003",
			expectedValidRanges: map[RelayChainID]int32{
				"0001": 100,
				"0002": 200,
			},
			expectedLogLimitBlocks: map[RelayChainID]int32{
				"0001": 500,
				"0002": 600,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			test.initialConfig.RemoveChainConfigs(test.chainID)

			c.Equal(test.expectedValidRanges, test.initialConfig.ChainsValidRanges)
			c.Equal(test.expectedLogLimitBlocks, test.initialConfig.ChainsLogLimitBlocks)
		})
	}
}

func Test_ValidatorPluginConfig_UpdateMethodValidator(t *testing.T) {
	tests := []struct {
		name                        string
		initialMethodValidatorsMap  map[Method][]ValidatorMethodStr
		methodConfig                *ValidatorMethodConfig
		expectedMethodValidatorsMap map[Method][]ValidatorMethodStr
	}{
		{
			name:                       "should add new method and validator to method validators map",
			initialMethodValidatorsMap: map[Method][]ValidatorMethodStr{},
			methodConfig: &ValidatorMethodConfig{
				Method:    "new_method",
				Validator: "new_validator",
			},
			expectedMethodValidatorsMap: map[Method][]ValidatorMethodStr{
				"new_method": {"new_validator"},
			},
		},
		{
			name: "should not duplicate validator for the same method",
			initialMethodValidatorsMap: map[Method][]ValidatorMethodStr{
				"existing_method": {"existing_validator"},
			},
			methodConfig: &ValidatorMethodConfig{
				Method:    "existing_method",
				Validator: "existing_validator",
			},
			expectedMethodValidatorsMap: map[Method][]ValidatorMethodStr{
				"existing_method": {"existing_validator"},
			},
		},
		{
			name: "should handle multiple validators for the same method",
			initialMethodValidatorsMap: map[Method][]ValidatorMethodStr{
				"existing_method": {"existing_validator1", "existing_validator2"},
			},
			methodConfig: &ValidatorMethodConfig{
				Method:    "existing_method",
				Validator: "new_validator",
			},
			expectedMethodValidatorsMap: map[Method][]ValidatorMethodStr{
				"existing_method": {"existing_validator1", "existing_validator2", "new_validator"},
			},
		},
		{
			name: "should handle multiple methods",
			initialMethodValidatorsMap: map[Method][]ValidatorMethodStr{
				"existing_method1": {"existing_validator1"},
				"existing_method2": {"existing_validator2"},
			},
			methodConfig: &ValidatorMethodConfig{
				Method:    "new_method",
				Validator: "new_validator",
			},
			expectedMethodValidatorsMap: map[Method][]ValidatorMethodStr{
				"existing_method1": {"existing_validator1"},
				"existing_method2": {"existing_validator2"},
				"new_method":       {"new_validator"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			v := &ValidatorPluginConfig{
				MethodValidatorsMap: test.initialMethodValidatorsMap,
			}

			v.UpdateMethodValidator(test.methodConfig)

			c.Equal(test.expectedMethodValidatorsMap, v.MethodValidatorsMap)
		})
	}
}

func Test_ValidatorPluginConfig_RemoveMethodValidator(t *testing.T) {
	tests := []struct {
		name                        string
		initialMethodValidatorsMap  map[Method][]ValidatorMethodStr
		methodConfig                *ValidatorMethodConfig
		expectedMethodValidatorsMap map[Method][]ValidatorMethodStr
	}{
		{
			name: "should remove validator from method validators map",
			initialMethodValidatorsMap: map[Method][]ValidatorMethodStr{
				"existing_method": {"existing_validator1", "existing_validator2"},
			},
			methodConfig: &ValidatorMethodConfig{
				Method:    "existing_method",
				Validator: "existing_validator1",
			},
			expectedMethodValidatorsMap: map[Method][]ValidatorMethodStr{
				"existing_method": {"existing_validator2"},
			},
		},
		{
			name: "should handle case when validator is not in the map",
			initialMethodValidatorsMap: map[Method][]ValidatorMethodStr{
				"existing_method": {"existing_validator1", "existing_validator2"},
			},
			methodConfig: &ValidatorMethodConfig{
				Method:    "existing_method",
				Validator: "non_existing_validator",
			},
			expectedMethodValidatorsMap: map[Method][]ValidatorMethodStr{
				"existing_method": {"existing_validator1", "existing_validator2"},
			},
		},
		{
			name: "should handle case when method is not in the map",
			initialMethodValidatorsMap: map[Method][]ValidatorMethodStr{
				"existing_method": {"existing_validator1", "existing_validator2"},
			},
			methodConfig: &ValidatorMethodConfig{
				Method:    "non_existing_method",
				Validator: "existing_validator1",
			},
			expectedMethodValidatorsMap: map[Method][]ValidatorMethodStr{
				"existing_method": {"existing_validator1", "existing_validator2"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			v := &ValidatorPluginConfig{
				MethodValidatorsMap: test.initialMethodValidatorsMap,
			}

			v.RemoveMethodValidator(test.methodConfig)

			c.Equal(test.expectedMethodValidatorsMap, v.MethodValidatorsMap)
		})
	}
}
