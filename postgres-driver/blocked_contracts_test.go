package postgresdriver

import (
	"context"
	"fmt"
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/testdata"
	"github.com/pokt-foundation/portal-http-db/v2/types"
)

func (ts *PGDriverTestSuite) Test_PGDriver_BlockedContracts() {
	ts.Run("Test_PGDriver_ReadBlockedContracts", func() {
		tests := []struct {
			name                     string
			expectedBlockedContracts types.GlobalBlockedContracts
			err                      error
		}{
			{
				name:                     "Should return all GlobalBlockedContracts from the DB",
				expectedBlockedContracts: testdata.GlobalBlockedContracts,
				err:                      nil,
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				blockedContracts, err := ts.driver.ReadBlockedContracts(context.Background())
				ts.Equal(test.err, err)
				ts.Equal(test.expectedBlockedContracts, blockedContracts)
			})
		}
	})

	ts.Run("Test_PGDriver_WriteBlockedContract", func() {
		tests := []struct {
			name                     string
			blockedAddress           types.BlockedAddress
			createdAt                time.Time
			expectedBlockedContracts types.GlobalBlockedContracts
			err                      error
		}{
			{
				name:           "Should add a new blocked address to the global blocked contracts table",
				blockedAddress: "0xtest_newabcdef0123456789abcdef01234567",
				expectedBlockedContracts: types.GlobalBlockedContracts{
					BlockedAddresses: map[types.BlockedAddress]struct{}{
						"0xtest_6789abcdef0123456789abcdef01234567":   {},
						"0xtest_f0123456789abcdef0123456789abcdef01":  {},
						"0xtest_cdef0123456789abcdef0123456789abcdef": {},
						"0xtest_56789abcdef0123456789abcdef01234567":  {},
						"0xtest_789abcdef0123456789abcdef0123456789":  {},
						"0xtest_newabcdef0123456789abcdef01234567":    {},
					},
				},
				createdAt: testdata.MockTimestamp,
				err:       nil,
			},
			{
				name:           "Should return an error if the address is empty",
				blockedAddress: "",
				createdAt:      testdata.MockTimestamp,
				err:            errNoAddress,
			},
			{
				name:           "Should return an error if the address is a duplicate",
				blockedAddress: "0xtest_cdef0123456789abcdef0123456789abcdef",
				createdAt:      testdata.MockTimestamp,
				err:            fmt.Errorf(errContractAlreadyBlocked.Error(), "0xtest_cdef0123456789abcdef0123456789abcdef"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				err := ts.driver.WriteBlockedContract(context.Background(), test.blockedAddress, test.createdAt)
				ts.Equal(test.err, err)
				if test.err == nil {
					blockedContracts, err := ts.driver.ReadBlockedContracts(context.Background())
					ts.NoError(err)
					ts.Equal(test.expectedBlockedContracts, blockedContracts)
				}
			})
		}
	})

	ts.Run("Test_PGDriver_UpdateBlockedContractActive", func() {
		tests := []struct {
			name                     string
			blockedAddress           types.BlockedAddress
			active                   bool
			updatedAt                time.Time
			expectedBlockedContracts types.GlobalBlockedContracts
			err                      error
		}{
			{
				name:           "Should deactivate a blocked address in the global blocked contracts table",
				blockedAddress: "0xtest_cdef0123456789abcdef0123456789abcdef",
				active:         false,
				expectedBlockedContracts: types.GlobalBlockedContracts{
					BlockedAddresses: map[types.BlockedAddress]struct{}{
						"0xtest_6789abcdef0123456789abcdef01234567":  {},
						"0xtest_f0123456789abcdef0123456789abcdef01": {},
						"0xtest_56789abcdef0123456789abcdef01234567": {},
						"0xtest_789abcdef0123456789abcdef0123456789": {},
						"0xtest_newabcdef0123456789abcdef01234567":   {},
					},
				},
				updatedAt: testdata.MockTimestamp,
				err:       nil,
			},
			{
				name:           "Should reactivate a blocked address in the global blocked contracts table",
				blockedAddress: "0xtest_cdef0123456789abcdef0123456789abcdef",
				active:         true,
				expectedBlockedContracts: types.GlobalBlockedContracts{
					BlockedAddresses: map[types.BlockedAddress]struct{}{
						"0xtest_6789abcdef0123456789abcdef01234567":   {},
						"0xtest_f0123456789abcdef0123456789abcdef01":  {},
						"0xtest_cdef0123456789abcdef0123456789abcdef": {},
						"0xtest_56789abcdef0123456789abcdef01234567":  {},
						"0xtest_789abcdef0123456789abcdef0123456789":  {},
						"0xtest_newabcdef0123456789abcdef01234567":    {},
					},
				},
				updatedAt: testdata.MockTimestamp,
				err:       nil,
			},
			{
				name:           "Should return an error if the address is empty",
				blockedAddress: "",
				updatedAt:      testdata.MockTimestamp,
				err:            errNoAddress,
			},
			{
				name:           "Should return an error if the address doesn't exist in the database",
				blockedAddress: "0xtest_34095u439fh49fh30fj239ru923kf3f09823fk",
				updatedAt:      testdata.MockTimestamp,
				err:            fmt.Errorf(errContractDoesntExist.Error(), "0xtest_34095u439fh49fh30fj239ru923kf3f09823fk"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				err := ts.driver.UpdateBlockedContractActive(context.Background(), test.blockedAddress, test.active, test.updatedAt)
				ts.Equal(test.err, err)
				if test.err == nil {
					blockedContracts, err := ts.driver.ReadBlockedContracts(context.Background())
					ts.NoError(err)
					ts.Equal(test.expectedBlockedContracts, blockedContracts)
				}
			})
		}
	})

	ts.Run("Test_PGDriver_RemoveBlockedContract", func() {
		tests := []struct {
			name                     string
			blockedAddress           types.BlockedAddress
			expectedBlockedContracts types.GlobalBlockedContracts
			err                      error
		}{
			{
				name:           "Should delete a blocked address in the global blocked contracts table",
				blockedAddress: "0xtest_789abcdef0123456789abcdef0123456789",
				expectedBlockedContracts: types.GlobalBlockedContracts{
					BlockedAddresses: map[types.BlockedAddress]struct{}{
						"0xtest_6789abcdef0123456789abcdef01234567":   {},
						"0xtest_f0123456789abcdef0123456789abcdef01":  {},
						"0xtest_cdef0123456789abcdef0123456789abcdef": {},
						"0xtest_56789abcdef0123456789abcdef01234567":  {},
						"0xtest_newabcdef0123456789abcdef01234567":    {},
					},
				},
				err: nil,
			},
			{
				name:           "Should return an error if the address is empty",
				blockedAddress: "",
				err:            errNoAddress,
			},
			{
				name:           "Should return an error if the address doesn't exist in the database",
				blockedAddress: "0xtest_34095u439fh49fh30fj239ru923kf3f09823fk",
				err:            fmt.Errorf(errContractDoesntExist.Error(), "0xtest_34095u439fh49fh30fj239ru923kf3f09823fk"),
			},
		}

		for _, test := range tests {
			ts.Run(test.name, func() {
				err := ts.driver.RemoveBlockedContract(context.Background(), test.blockedAddress)
				ts.Equal(test.err, err)
				if test.err == nil {
					blockedContracts, err := ts.driver.ReadBlockedContracts(context.Background())
					ts.NoError(err)
					ts.Equal(test.expectedBlockedContracts, blockedContracts)
				}
			})
		}
	})
}
