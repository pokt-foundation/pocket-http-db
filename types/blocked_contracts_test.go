package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GlobalBlockedContracts_IsContractBlocked(t *testing.T) {
	c := require.New(t)

	tests := []struct {
		name             string
		blockedContracts GlobalBlockedContracts
		contract         BlockedAddress
		expectedResult   bool
	}{
		{
			name: "Should return true if a given contract is blocked",
			blockedContracts: GlobalBlockedContracts{
				BlockedAddresses: map[BlockedAddress]struct{}{
					"0xtest_5Cc6634C0532925a3b844Bc454e4438f44e": {},
					"0xtest_C90EfD254f685c5f3F3E3b3CcefA96B7242": {},
					"0xtest_fDa3C5EfBf33E2f2D67EbF5f5D45f5f5F5f": {},
				},
			},
			contract:       BlockedAddress("0xtest_fDa3C5EfBf33E2f2D67EbF5f5D45f5f5F5f"),
			expectedResult: true,
		},
		{
			name: "Should return false if a given contract is not blocked",
			blockedContracts: GlobalBlockedContracts{
				BlockedAddresses: map[BlockedAddress]struct{}{
					"0xtest_5Cc6634C0532925a3b844Bc454e4438f44e": {},
					"0xtest_C90EfD254f685c5f3F3E3b3CcefA96B7242": {},
					"0xtest_fDa3C5EfBf33E2f2D67EbF5f5D45f5f5F5f": {},
				},
			},
			contract:       BlockedAddress("0xtest_Ff0F0f0f0F0F0f0f0F0f0F0f0f0f0F0f0f0F"),
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isContractBlocked := test.blockedContracts.IsContractBlocked(test.contract)
			c.Equal(test.expectedResult, isContractBlocked)
		})
	}
}
