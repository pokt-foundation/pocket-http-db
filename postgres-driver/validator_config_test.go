package postgresdriver

import (
	"context"

	"github.com/pokt-foundation/portal-http-db/v2/testdata"
	"github.com/pokt-foundation/portal-http-db/v2/types"
)

func (ts *PGDriverTestSuite) Test_PGDriver_ReadValidatorConfig() {
	tests := []struct {
		name     string
		expected *types.ValidatorPluginConfig
	}{
		{
			name:     "Test_PGDriver_ReadValidatorConfig",
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
