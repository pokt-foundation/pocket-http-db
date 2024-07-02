package postgresdriver

import (
	"context"
	"fmt"
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/testdata"
	"github.com/pokt-foundation/portal-http-db/v2/types"
)

func (ts *PGDriverTestSuite) Test_PGDriver_ReadChains() {
	tests := []struct {
		name    string
		chains  map[types.RelayChainID]*types.Chain
		options types.DriverOptions
		err     error
	}{
		{
			name: "Should return all non-deleted Chains from the database",
			chains: map[types.RelayChainID]*types.Chain{
				types.RelayChainID("0001"): testdata.Chains()[types.RelayChainID("0001")],
				types.RelayChainID("0053"): testdata.Chains()[types.RelayChainID("0053")],
				types.RelayChainID("0021"): testdata.Chains()[types.RelayChainID("0021")],
				types.RelayChainID("0064"): testdata.Chains()[types.RelayChainID("0064")],
				types.RelayChainID("0040"): testdata.Chains()[types.RelayChainID("0040")],
				types.RelayChainID("0083"): testdata.Chains()[types.RelayChainID("0083")],
			},
			options: types.DriverOptions{},
			err:     nil,
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			chains, err := ts.driver.ReadChains(context.Background(), test.options)
			ts.Equal(test.err, err)
			ts.Equal(test.chains, chains)
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_WriteChainAndGigastakeApps() {
	tests := []struct {
		name            string
		newChainInput   types.NewChainInput
		testCreatedTime time.Time
		err             error
	}{
		{
			name:            "Should create a new Chain along with its GigastakeApps in the database",
			newChainInput:   testdata.TestCreateNewChainInput,
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name: "Should fail if chain is nil",
			newChainInput: types.NewChainInput{
				Chain: nil,
				GigastakeApps: []*types.GigastakeApp{{
					ProtocolID: "morse-mainnet",
				}},
			},
			err: errChainCannotBeNil,
		},
		{
			name: "Should fail if altruist URL is invalid",
			newChainInput: types.NewChainInput{
				Chain: &types.Chain{
					ID:         "testID",
					Blockchain: "testBlockchain",
					Altruists:  map[types.AltruistURL]types.Altruist{"invalid_url": {URL: "invalid_url"}},
				},
				GigastakeApps: []*types.GigastakeApp{{
					ProtocolID: "morse-mainnet",
				}},
			},
			err: fmt.Errorf(errInvalidAltruistURL.Error(), "invalid_url"),
		},
		{
			name: "Should fail if domain is invalid",
			newChainInput: types.NewChainInput{
				Chain: &types.Chain{
					ID:         "testID",
					Blockchain: "testBlockchain",
					AliasDomains: map[types.ChainAlias][]types.ChainDomain{
						"testAlias": {"invalid_domain"},
					},
				},
				GigastakeApps: []*types.GigastakeApp{{
					ProtocolID: "morse-mainnet",
				}},
			},
			err: fmt.Errorf(errInvalidDomain.Error(), "invalid_domain", "testAlias"),
		},
		{
			name: "Should fail if domain is invalid",
			newChainInput: types.NewChainInput{
				Chain: &types.Chain{
					ID:         "testID",
					Blockchain: "testBlockchain",
					AliasDomains: map[types.ChainAlias][]types.ChainDomain{
						"testAlias": {"invalid_domain"},
					},
				},
				GigastakeApps: []*types.GigastakeApp{{
					ProtocolID: "morse-mainnet",
				}},
			},
			err: fmt.Errorf(errInvalidDomain.Error(), "invalid_domain", "testAlias"),
		},
		{
			name: "Should fail if trying to insert an existing chain",
			newChainInput: types.NewChainInput{
				Chain: &types.Chain{
					ID:         "0001",
					Blockchain: "testBlockchain",
				},
				GigastakeApps: []*types.GigastakeApp{{
					ProtocolID: "morse-mainnet",
				}},
			},
			err: fmt.Errorf(errChainExists.Error(), "0001"),
		},
		{
			name: "Should fail if protocol ID is invalid",
			newChainInput: types.NewChainInput{
				Chain: &types.Chain{
					ID:         "0001",
					Blockchain: "testBlockchain",
				},
				GigastakeApps: []*types.GigastakeApp{{
					ProtocolID: "invalid_protocol",
				}},
			},
			err: fmt.Errorf(errInvalidProtocolID.Error(), "invalid_protocol"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			_, err := ts.driver.WriteChainAndGigastakeApps(context.Background(), test.newChainInput, test.testCreatedTime)
			ts.Equal(test.err, err)

			if test.err == nil {
				chains, err := ts.driver.ReadChains(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				exists := false
				for chainID, chain := range chains {
					if chainID == test.newChainInput.Chain.ID {
						exists = true
						ts.Equal(test.newChainInput.Chain, chain)
						break
					}
				}
				ts.True(exists)

				gigastakeApps, err := ts.driver.ReadGigastakeApps(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				exists = false
				for gigastakeAppID, gigastakeApp := range gigastakeApps {
					for _, testGigastakeApp := range test.newChainInput.GigastakeApps {
						if gigastakeAppID == testGigastakeApp.ID {
							exists = true
							ts.Equal(testGigastakeApp, gigastakeApp)
							break
						}
					}
				}
				ts.True(exists)
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_WriteChain() {
	tests := []struct {
		name            string
		chain           types.Chain
		testCreatedTime time.Time
		iconURL         string
		altruistURL     types.AltruistURL
		aliasDomains    map[types.ChainAlias][]types.ChainDomain
		err             error
	}{
		{
			name:            "Should create a new Chain in the database",
			chain:           *testdata.TestCreateChain,
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name:            "Should fail if chain already exists in the database",
			chain:           *testdata.Chains()[types.RelayChainID("0064")],
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errChainExists.Error(), "0064"),
		},
		{
			name:            "Should fail if input Chain has an invalid icon URL set",
			iconURL:         "what-is+this*wow",
			chain:           *testdata.TestCreateChain,
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errInvalidIconURL.Error(), "what-is+this*wow"),
		},
		{
			name:            "Should fail if any input Altruist has an invalid URL",
			altruistURL:     "htz:/bad-domain2",
			chain:           *testdata.TestCreateChain,
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errInvalidAltruistURL.Error(), "htz:/bad-domain2"),
		},
		{
			name:        "Should fail if any input alias has an invalid domain",
			altruistURL: "http://www.good-domain.com",
			aliasDomains: map[types.ChainAlias][]types.ChainDomain{
				"sol-mainnet": {"im-not-a-domain"},
			},
			chain:           *testdata.TestCreateChain,
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errInvalidDomain.Error(), "im-not-a-domain", "sol-mainnet"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			testChain := test.chain

			if test.iconURL != "" {
				testChain.IconURL = test.iconURL
			}
			if test.altruistURL != "" {
				altruist := testChain.Altruists[test.altruistURL]
				altruist.URL = test.altruistURL
				testChain.Altruists = make(map[types.AltruistURL]types.Altruist)
				testChain.Altruists[test.altruistURL] = altruist
			}
			if len(test.aliasDomains) > 0 {
				testChain.AliasDomains = test.aliasDomains
			}

			createdChain, err := ts.driver.WriteChain(context.Background(), testChain, test.testCreatedTime)
			ts.Equal(test.err, err)

			if test.err == nil {
				ts.Equal(&testChain, createdChain)

				chains, err := ts.driver.ReadChains(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				ts.Equal(&testChain, chains[createdChain.ID])
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_UpdateChain() {
	tests := []struct {
		name            string
		update          types.UpdateChain
		testCreatedTime time.Time
		err             error
	}{
		{
			name:            "Should update an existing chain in the database",
			update:          testdata.UpdateChainOne,
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name:            "Should update an existing chain again in the database",
			update:          testdata.UpdateChainTwo,
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name:            "Should not remove any subtables if they are not present in the update struct",
			update:          testdata.UpdateChainTwo,
			testCreatedTime: testdata.MockTimestamp,
			err:             nil,
		},
		{
			name:            "Should fail if input Chain has an invalid icon URL set",
			update:          testdata.UpdateChainInvalidIconURL,
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errInvalidIconURL.Error(), "what_is_555"),
		},
		{
			name:            "Should fail if chain doesn't exist in the database",
			update:          testdata.UpdateChainNotExists,
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errChainDoesntExist.Error(), "0073"),
		},
		{
			name:            "Should fail if any input Altruist has an invalid URL",
			update:          testdata.UpdateChainInvalidURL,
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errInvalidAltruistURL.Error(), "htz:/bad-domain2"),
		},
		{
			name:            "Should fail if any input alias has an invalid domain",
			update:          testdata.UpdateChainInvalidDomain,
			testCreatedTime: testdata.MockTimestamp,
			err:             fmt.Errorf(errInvalidDomain.Error(), "im-not-a-domain", "sol-mainnet"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			var initialChain *types.Chain
			if test.err == nil {
				chains, err := ts.driver.ReadChains(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				chain, ok := chains[test.update.ID]
				ts.True(ok)
				initialChain = chain
			}

			updatedChain, err := ts.driver.UpdateChain(context.Background(), test.update, test.testCreatedTime)
			ts.Equal(test.err, err)

			if test.err == nil {
				// Assert that the fields in the updated chain match the expected values
				// Only check the fields present in the update struct
				if test.update.Blockchain != nil {
					ts.Equal(*test.update.Blockchain, updatedChain.Blockchain)
				} else {
					ts.Equal(initialChain.Blockchain, updatedChain.Blockchain)
				}

				if test.update.IconURL != nil {
					ts.Equal(*test.update.IconURL, updatedChain.IconURL)
				} else {
					ts.Equal(initialChain.IconURL, updatedChain.IconURL)
				}

				if test.update.Description != nil {
					ts.Equal(*test.update.Description, updatedChain.Description)
				} else {
					ts.Equal(initialChain.Description, updatedChain.Description)
				}

				if test.update.EnforceResult != nil {
					ts.Equal(*test.update.EnforceResult, updatedChain.EnforceResult)
				} else {
					ts.Equal(initialChain.EnforceResult, updatedChain.EnforceResult)
				}

				if test.update.Path != nil {
					ts.Equal(*test.update.Path, updatedChain.Path)
				} else {
					ts.Equal(initialChain.Path, updatedChain.Path)
				}

				if test.update.Ticker != nil {
					ts.Equal(*test.update.Ticker, updatedChain.Ticker)
				} else {
					ts.Equal(initialChain.Ticker, updatedChain.Ticker)
				}

				if test.update.AllowedMethods != nil {
					ts.Equal(test.update.AllowedMethods, updatedChain.AllowedMethods)
				} else {
					ts.Equal(initialChain.AllowedMethods, updatedChain.AllowedMethods)
				}

				if test.update.LogLimitBlocks != nil {
					ts.Equal(*test.update.LogLimitBlocks, updatedChain.LogLimitBlocks)
				} else {
					ts.Equal(initialChain.LogLimitBlocks, updatedChain.LogLimitBlocks)
				}

				if test.update.RequestTimeout != nil {
					ts.Equal(*test.update.RequestTimeout, updatedChain.RequestTimeout)
				} else {
					ts.Equal(initialChain.RequestTimeout, updatedChain.RequestTimeout)
				}

				if test.update.NodeRequestTimeout != nil {
					ts.Equal(*test.update.NodeRequestTimeout, updatedChain.NodeRequestTimeout)
				} else {
					ts.Equal(initialChain.NodeRequestTimeout, updatedChain.NodeRequestTimeout)
				}

				if test.update.Active != nil {
					ts.Equal(*test.update.Active, updatedChain.Active)
				} else {
					ts.Equal(initialChain.Active, updatedChain.Active)
				}

				if len(*test.update.Altruists) > 0 {
					ts.Equal(*test.update.Altruists, updatedChain.Altruists)
				} else {
					ts.Equal(initialChain.Altruists, updatedChain.Altruists)
				}

				if len(*test.update.Checks) > 0 {
					ts.Equal(*test.update.Checks, updatedChain.Checks)
				} else {
					ts.Equal(initialChain.Checks, updatedChain.Checks)
				}

				if len(*test.update.Aliases) > 0 {
					ts.Equal(*test.update.Aliases, updatedChain.Aliases)
				} else {
					ts.Equal(initialChain.Aliases, updatedChain.Aliases)
				}
			}
		})
	}
}

func (ts *PGDriverTestSuite) Test_PGDriver_SetChainActiveStatus() {
	tests := []struct {
		name    string
		chainID types.RelayChainID
		active  bool
		err     error
	}{
		{
			name:    "Should deactivate an existing chain in the database",
			chainID: "0001",
			active:  false,
			err:     nil,
		},
		{
			name:    "Should activate an existing chain in the database",
			chainID: "0001",
			active:  true,
			err:     nil,
		},
		{
			name:    "Should fail if chain does not exist in the database",
			chainID: "7701",
			err:     fmt.Errorf(errChainDoesntExist.Error(), "7701"),
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			active, err := ts.driver.SetChainActiveStatus(context.Background(), test.chainID, test.active, testdata.MockTimestamp)
			ts.Equal(test.err, err)

			if test.err == nil {
				ts.Equal(test.active, active)

				chains, err := ts.driver.ReadChains(context.Background(), types.DriverOptions{})
				ts.NoError(err)
				ts.Equal(test.active, chains[test.chainID].Active)
			}
		})
	}
}
