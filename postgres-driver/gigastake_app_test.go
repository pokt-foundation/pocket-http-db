package postgresdriver

import (
	"context"
	"fmt"
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/testdata"
	"github.com/pokt-foundation/portal-http-db/v2/types"
)

func (ts *PGDriverTestSuite) Test_PGDriver_ReadGigastakeApps() {
	tests := []struct {
		name          string
		gigastakeApps map[types.GigastakeAppID]*types.GigastakeApp
		options       types.DriverOptions
		testNoChains  bool
		err           error
	}{
		{
			name: "Should return all non-deleted GigastakeApps from the database",
			gigastakeApps: map[types.GigastakeAppID]*types.GigastakeApp{
				testdata.GigastakeApps()["test_gigastake_app_1"].ID: testdata.GigastakeApps()["test_gigastake_app_1"],
				testdata.GigastakeApps()["test_gigastake_app_2"].ID: testdata.GigastakeApps()["test_gigastake_app_2"],
				testdata.GigastakeApps()["test_gigastake_app_3"].ID: testdata.GigastakeApps()["test_gigastake_app_3"],
			},
			options: types.DriverOptions{},
			err:     nil,
		},
		{
			name: "Should work even if gigastake app has no chains associated",
			gigastakeApps: map[types.GigastakeAppID]*types.GigastakeApp{
				testdata.GigastakeApps()["test_gigastake_app_1"].ID: testdata.GigastakeApps()["test_gigastake_app_1"],
				testdata.GigastakeApps()["test_gigastake_app_2"].ID: testdata.GigastakeApps()["test_gigastake_app_2"],
				testdata.GigastakeApps()["test_gigastake_app_3"].ID: {
					ID:              "test_gigastake_app_3",
					ProtocolID:      "morse-mainnet",
					ChainIDs:        map[types.RelayChainID]struct{}{},
					Name:            "snorlax",
					Address:         "test_e570c841d5cd4f6197e0428ed7c517fd",
					PublicKey:       types.GigastakeAppPublicKey("test_4f805bbbf96c4a649efc3f4f95616f2e"),
					ClientPublicKey: "test_789f9d6adcc846f1a079bf68237b5f5c",
					Signature:       "test_01eac46efc9242a2be73879f1d09f1dc",
					Version:         "0.0.1",
					CreatedAt:       testdata.MockTimestamp,
					UpdatedAt:       testdata.MockTimestamp,
				},
			},
			options:      types.DriverOptions{},
			testNoChains: true,
			err:          nil,
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			if test.testNoChains {
				ts.NoError(ts.deleteChainsForGigastakeApp3())
			}

			gigastakeApps, err := ts.driver.ReadGigastakeApps(context.Background(), test.options)
			ts.Equal(test.err, err)
			ts.Equal(test.gigastakeApps, gigastakeApps)
		})
	}
}

func (ts *PGDriverTestSuite) deleteChainsForGigastakeApp3() error {
	_, err := ts.driver.db.Exec(context.Background(), `
		DELETE FROM chains_gigastake_applications
		WHERE gigastake_application_id = 'test_gigastake_app_3'
	`)
	return err
}

func (ts *PGDriverTestSuite) Test_PGDriver_WriteGigastakeApp() {
	tests := []struct {
		name            string
		gigastakeApp    types.GigastakeApp
		testCreatedTime time.Time
		err             error
	}{
		{
			name:            "Should create a new GigastakeApp in the database",
			gigastakeApp:    testdata.TestCreateGigastakeApp,
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should return an error if the chain doesn't exist",
			gigastakeApp: types.GigastakeApp{
				Name:     "whatever",
				ChainIDs: map[types.RelayChainID]struct{}{"0666": {}},
			},
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errChainDoesntExist.Error(), "0666"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			createdGigastakeApp, err := ts.driver.WriteGigastakeApp(context.Background(), &test.gigastakeApp, test.testCreatedTime)
			ts.Equal(test.err, err)

			if test.err == nil {
				// Ensure the timestamps are the same as the created app
				test.gigastakeApp.ID = createdGigastakeApp.ID
				test.gigastakeApp.CreatedAt = createdGigastakeApp.CreatedAt
				test.gigastakeApp.UpdatedAt = createdGigastakeApp.UpdatedAt
				ts.Equal(&test.gigastakeApp, createdGigastakeApp)

				gigastakeApps, err := ts.driver.ReadGigastakeApps(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				ts.Equal(&test.gigastakeApp, gigastakeApps[test.gigastakeApp.ID])
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_UpdateGigastakeApp() {
	tests := []struct {
		name               string
		gigastakeAppUpdate types.UpdateGigastakeApp
		testUpdatedTime    time.Time
		err                error
	}{
		{
			name: "Should update GigastakeApp ChainIDs in the database",
			gigastakeAppUpdate: types.UpdateGigastakeApp{
				ID:       "test_gigastake_app_1",
				ChainIDs: []types.RelayChainID{"0001", "0040"},
			},
			testUpdatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should return an error if the GigastakeApp ChainIDs is empty",
			gigastakeAppUpdate: types.UpdateGigastakeApp{
				ID:       "test_gigastake_app_1",
				ChainIDs: []types.RelayChainID{},
			},
			testUpdatedTime: testdata.MockTimestamp,
			err:             errEmptyChainIDsForGigastakeApp,
		},
		{
			name: "Should return an error if the chain doesn't exist",
			gigastakeAppUpdate: types.UpdateGigastakeApp{
				ID:       "test_gigastake_app_1",
				ChainIDs: []types.RelayChainID{"0666"},
			},
			testUpdatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errChainDoesntExist.Error(), "0666"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			err := ts.driver.UpdateGigastakeApp(context.Background(), test.gigastakeAppUpdate, test.testUpdatedTime)
			ts.Equal(test.err, err)

			if test.err == nil {
				gigastakeApps, err := ts.driver.ReadGigastakeApps(context.Background(), types.DriverOptions{})
				ts.NoError(err)

				updatedApp := gigastakeApps[test.gigastakeAppUpdate.ID]

				if len(test.gigastakeAppUpdate.ChainIDs) > 0 {
					// Validate ChainIDs
					expectedChainIDs := make(map[types.RelayChainID]struct{})
					for _, chainID := range test.gigastakeAppUpdate.ChainIDs {
						expectedChainIDs[chainID] = struct{}{}
					}
					ts.Equal(expectedChainIDs, updatedApp.ChainIDs)
				}
				ts.Equal(test.testUpdatedTime, updatedApp.UpdatedAt)
			}
		})
	}
}
