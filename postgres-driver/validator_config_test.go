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
			name:     "should read the validator plugin config from the database",
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
		chainID  types.RelayChainID
		expected map[types.RelayChainID][]byte
	}{
		{
			name:     "should read the validator plugin config from the database",
			chainID:  "0083",
			expected: testdata.OpenAPISpecs("../testdata/radix-spec.json"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			result, err := ts.driver.ReadOpenAPISpecs(context.Background())
			ts.NoError(err)

			ts.Equal(test.expected, result)

			// Test if the result is a valid JSON structure
			var dummyInterface interface{}
			err = json.Unmarshal(result[test.chainID], &dummyInterface)
			ts.NoError(err, "The result should be a valid JSON structure")
		})
	}
}
