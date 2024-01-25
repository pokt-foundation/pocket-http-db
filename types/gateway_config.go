package types

type ValidatorMethodStr = string

const (
	ValidatorMethodBlockRange     ValidatorMethodStr = "blockRange"
	ValidatorMethodLogLimitBlocks ValidatorMethodStr = "logLimitBlocks"
	ValidatorMethodETHAddress     ValidatorMethodStr = "ethAddress"
)

type (
	ValidatorPluginConfig struct {
		// ChainsValidRanges is a mapping of chain id to the maximum range of blocks (from the estimated current block height) allowed for any relay request
		ChainsValidRanges map[RelayChainID]int32 `json:"chains_valid_ranges,omitempty"`

		// ChainsLogLimitBlocks is a mapping of chain id to the maximum number of blocks allowed for any eth_getLogs relay request
		ChainsLogLimitBlocks map[RelayChainID]int32 `json:"chains_log_limit_blocks,omitempty"`

		// ArchivalChains is a mapping of mainnet chain IDs to their archival chain IDs (for all chains with an archival variant)
		ArchivalChains map[RelayChainID]RelayChainID `json:"archival_chains,omitempty"`

		// MethodValidatorsMap is a mapping of method to a list of validators to be applied to relay requests for that method
		MethodValidatorsMap map[Method][]ValidatorMethodStr `json:"method_validators,omitempty"`
	}

	ValidatorChainConfig struct {
		ChainID        RelayChainID `json:"chainID"`
		ValidRange     int32        `json:"validRange"`
		LogLimitBlocks int32        `json:"logLimitBlocks"`
	}

	ValidatorMethodConfig struct {
		Method    Method             `json:"method"`
		Validator ValidatorMethodStr `json:"validator"`
	}
)

func (v *ValidatorPluginConfig) UpdateChainConfigs(chainConfig *ValidatorChainConfig) {
	v.ChainsValidRanges[chainConfig.ChainID] = chainConfig.ValidRange
	v.ChainsLogLimitBlocks[chainConfig.ChainID] = chainConfig.LogLimitBlocks
}

func (v *ValidatorPluginConfig) RemoveChainConfigs(chainID RelayChainID) {
	delete(v.ChainsValidRanges, chainID)
	delete(v.ChainsLogLimitBlocks, chainID)
}

func (v *ValidatorPluginConfig) UpdateMethodValidator(methodConfig *ValidatorMethodConfig) {
	existingValidators, ok := v.MethodValidatorsMap[methodConfig.Method]
	if ok {
		for _, validator := range existingValidators {
			if validator == methodConfig.Validator {
				return
			}
		}
		v.MethodValidatorsMap[methodConfig.Method] = append(existingValidators, methodConfig.Validator)
	} else {
		v.MethodValidatorsMap[methodConfig.Method] = []ValidatorMethodStr{methodConfig.Validator}
	}
}

func (v *ValidatorPluginConfig) RemoveMethodValidator(methodConfig *ValidatorMethodConfig) {
	validators := v.MethodValidatorsMap[methodConfig.Method]
	for i, validator := range validators {
		if validator == methodConfig.Validator {
			// Remove the validator from the slice.
			validators = append(validators[:i], validators[i+1:]...)
			break
		}
	}

	if len(validators) == 0 {
		// If no validators left, delete the key from the map.
		delete(v.MethodValidatorsMap, methodConfig.Method)
	} else {
		// Otherwise, update the map with the new slice of validators.
		v.MethodValidatorsMap[methodConfig.Method] = validators
	}
}

func (v *ValidatorChainConfig) Table() Table {
	return TableChainConfig
}

func (v *ValidatorMethodConfig) Table() Table {
	return TableMethodValidators
}
