package postgresdriver

import (
	"context"
	"encoding/json"

	"github.com/pokt-foundation/portal-http-db/v2/testdata"
	"github.com/pokt-foundation/portal-http-db/v2/types"
)

func (ts *PGDriverTestSuite) Test_PGDriver_ReadValidatorConfig() {
	tests := []struct {
		name     string
		expected *types.ValidatorPluginConfig
	}{
		{
			name:     "should read the method validator plugin config from the database",
			expected: testdata.ValidatorPluginConfig(),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			result, err := ts.driver.ReadValidatorConfig(context.Background())
			ts.NoError(err)

			ts.Equal(test.expected.ChainsValidRanges, result.ChainsValidRanges)
			ts.Equal(test.expected.ChainsLogLimitBlocks, result.ChainsLogLimitBlocks)

			for method, validators := range test.expected.MethodValidatorsMap {
				ts.Equal(validators, result.MethodValidatorsMap[method])
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_ReadOpenAPISpecs() {
	tests := []struct {
		name     string
		expected map[types.RelayChainID]json.RawMessage
	}{
		{
			name:     "should read the chain open api specs from the database",
			expected: testdata.OpenAPISpecs("../testdata/radix-spec.json"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			result, err := ts.driver.ReadOpenAPISpecs(context.Background())
			ts.NoError(err)
			for chainID, expectedSpec := range test.expected {
				ts.JSONEq(string(expectedSpec), string(result[chainID]))
			}
		})
	}
}
