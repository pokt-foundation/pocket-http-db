package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ChainDomain_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		domain   ChainDomain
		expected bool
	}{
		{
			name:     "Should return true for a valid domain without port",
			domain:   ChainDomain("pokt-rpc.gateway.pokt.network"),
			expected: true,
		},
		{
			name:     "Should return true for a valid domain with port",
			domain:   ChainDomain("pokt-rpc.gateway.pokt.network:80"),
			expected: true,
		},
		{
			name:     "Should return false for a domain without TLD",
			domain:   ChainDomain("pokt-rpc.gateway"),
			expected: false,
		},
		{
			name:     "Should return false for an invalid domain",
			domain:   ChainDomain("pokt-rpc..gateway.pokt.network"),
			expected: false,
		},
		{
			name:     "Should return false for an empty domain",
			domain:   ChainDomain(""),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isValid := test.domain.IsValid()
			assert.Equal(t, test.expected, isValid)
		})
	}
}

func Test_ChainDomain_IsPublicEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		domain   ChainDomain
		expected bool
	}{
		{
			name:     "Should return true for public endpoint",
			domain:   ChainDomain("pokt-rpc.gateway.pokt.network"),
			expected: true,
		},
		{
			name:     "Should return false for non-public endpoint",
			domain:   ChainDomain("pokt.gateway.pokt.network"),
			expected: false,
		},
		{
			name:     "Should return false for empty domain",
			domain:   ChainDomain(""),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isPublic := test.domain.IsPublicEndpoint()
			assert.Equal(t, test.expected, isPublic)
		})
	}
}

func Test_ChainDomain_IsWildcardDomain(t *testing.T) {
	tests := []struct {
		name     string
		domain   ChainDomain
		expected bool
	}{
		{
			name:     "Should return true for wildcard domain",
			domain:   ChainDomain("*.gateway.pokt.network"),
			expected: true,
		},
		{
			name:     "Should return false for non-wildcard domain",
			domain:   ChainDomain("pokt-rpc.gateway.pokt.network"),
			expected: false,
		},
		{
			name:     "Should return false for empty domain",
			domain:   ChainDomain(""),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isWildcard := test.domain.IsWildcardDomain()
			assert.Equal(t, test.expected, isWildcard)
		})
	}
}

func Test_ChainDomain_GetAlias(t *testing.T) {
	tests := []struct {
		name     string
		domain   ChainDomain
		expected ChainAlias
	}{
		{
			name:     "Should return the alias for valid domain",
			domain:   ChainDomain("pokt-mainnet.gateway.pokt.network"),
			expected: ChainAlias("pokt-mainnet"),
		},
		{
			name:     "Should return empty alias for empty domain",
			domain:   ChainDomain(""),
			expected: ChainAlias(""),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alias := test.domain.GetAlias()
			assert.Equal(t, test.expected, alias)
		})
	}
}

func Test_GetChainAltruists(t *testing.T) {
	tests := []struct {
		name     string
		chain    Chain
		expected []Altruist
	}{
		{
			name: "Should return all altruists of the chain",
			chain: Chain{
				Altruists: map[AltruistURL]Altruist{
					"https://fluffy-bunny.com": {
						ChainID:  "testChainID",
						URL:      "https://fluffy-bunny.com",
						Auth:     "carrotAuth",
						AuthType: ChainAuthTypeBearerToken,
					},
					"https://super-kitten.io": {
						ChainID:  "testChainID",
						URL:      "https://super-kitten.io",
						Auth:     "meowAuth",
						AuthType: ChainAuthTypeBearerToken,
					},
				},
			},
			expected: []Altruist{
				{
					ChainID:  "testChainID",
					URL:      "https://fluffy-bunny.com",
					Auth:     "carrotAuth",
					AuthType: ChainAuthTypeBearerToken,
				},
				{
					ChainID:  "testChainID",
					URL:      "https://super-kitten.io",
					Auth:     "meowAuth",
					AuthType: ChainAuthTypeBearerToken,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			altruists := test.chain.GetChainAltruists()
			assert.ElementsMatch(t, test.expected, altruists)
		})
	}
}

func Test_IsEVM(t *testing.T) {
	tests := []struct {
		name     string
		chain    Chain
		expected bool
	}{
		{
			name: "Chain with EVM check",
			chain: Chain{
				ID: "1001",
				Checks: map[ChainCheckType]Check{
					ChainCheckTypeChain: {
						ChainID:    "1001",
						Type:       ChainCheckTypeChain,
						Payload:    `{"method":"eth_chainId","id":10,"jsonrpc":"2.0"}`,
						ResultKey:  "id",
						Allowance:  10,
						EVMChainID: 10,
					},
				},
			},
			expected: true,
		},
		{
			name: "Chain without EVM check",
			chain: Chain{
				ID: "2002",
				Checks: map[ChainCheckType]Check{
					ChainCheckTypeSync: {
						ChainID:    "2002",
						Type:       ChainCheckTypeSync,
						Payload:    `{"id":20,"jsonrpc":"2.0","method":"hmy_blockNumber","params":[]}`,
						ResultKey:  "result",
						Allowance:  20,
						EVMChainID: 20,
					},
				},
			},
			expected: false,
		},
		{
			name: "Chain with no checks",
			chain: Chain{
				ID:     "3003",
				Checks: map[ChainCheckType]Check{},
			},
			expected: false,
		},
		{
			name: "Chain is nil",
			chain: Chain{
				ID:     "4004",
				Checks: nil,
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isEVM := test.chain.IsEVM()
			assert.Equal(t, test.expected, isEVM)
		})
	}
}

func Test_GetChainAliases(t *testing.T) {
	tests := []struct {
		name     string
		chain    Chain
		expected []ChainAlias
	}{
		{
			name: "Should return all aliases of the chain",
			chain: Chain{
				Aliases: map[ChainAlias]struct{}{
					"pokt-mainnet": {},
					"pokt-testnet": {},
				},
			},
			expected: []ChainAlias{"pokt-mainnet", "pokt-testnet"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			aliases := test.chain.GetChainAliases()
			assert.ElementsMatch(t, test.expected, aliases)
		})
	}
}

func Test_GetChainCheck(t *testing.T) {
	chain := &Chain{
		Checks: map[ChainCheckType]Check{
			ChainCheckTypeChain: {
				Type:       ChainCheckTypeChain,
				Payload:    "test payload",
				ResultKey:  "test key",
				Allowance:  1000,
				EVMChainID: 1,
			},
			ChainCheckTypeArchival: {
				Type:       ChainCheckTypeArchival,
				Payload:    "test payload",
				ResultKey:  "test key",
				Allowance:  1000,
				EVMChainID: 1,
			},
		},
	}

	tests := []struct {
		name      string
		checkType ChainCheckType
		want      Check
	}{
		{
			name:      "Test getting Chain check",
			checkType: ChainCheckTypeChain,
			want: Check{
				Type:       ChainCheckTypeChain,
				Payload:    "test payload",
				ResultKey:  "test key",
				Allowance:  1000,
				EVMChainID: 1,
			},
		},
		{
			name:      "Test getting Archival check",
			checkType: ChainCheckTypeArchival,
			want: Check{
				Type:       ChainCheckTypeArchival,
				Payload:    "test payload",
				ResultKey:  "test key",
				Allowance:  1000,
				EVMChainID: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := chain.GetChainCheck(tt.checkType)
			if got != tt.want {
				t.Errorf("GetChainCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetChainDomains(t *testing.T) {
	tests := []struct {
		name     string
		chain    Chain
		expected []ChainDomain
	}{
		{
			name: "Should return all domains of the chain",
			chain: Chain{
				AliasDomains: map[ChainAlias][]ChainDomain{
					"pokt-mainnet": {"pokt-rpc.gateway.pokt.network", "pokt-rpc-2.gateway.pokt.network"},
					"pokt-testnet": {"pokt-rpc-test.gateway.pokt.network"},
				},
			},
			expected: []ChainDomain{"pokt-rpc.gateway.pokt.network", "pokt-rpc-2.gateway.pokt.network", "pokt-rpc-test.gateway.pokt.network"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			domains := test.chain.GetChainDomains()
			assert.ElementsMatch(t, test.expected, domains)
		})
	}
}

func Test_GetAllGigastakeApps(t *testing.T) {
	tests := []struct {
		name     string
		chain    Chain
		expected []GigastakeApp
	}{
		{
			name: "Should return all GigastakeApps across all protocols",
			chain: Chain{
				GigastakeApps: map[ProtocolID]map[GigastakeAppID]*GigastakeApp{
					"protocol_1": {
						"app_1": {Version: "1.0", ClientPublicKey: "key1", PublicKey: "pubKey1", Signature: "sig1"},
						"app_2": {Version: "1.1", ClientPublicKey: "key2", PublicKey: "pubKey2", Signature: "sig2"},
					},
					"protocol_2": {
						"app_3": {Version: "2.0", ClientPublicKey: "key3", PublicKey: "pubKey3", Signature: "sig3"},
					},
				},
			},
			expected: []GigastakeApp{
				{Version: "1.0", ClientPublicKey: "key1", PublicKey: "pubKey1", Signature: "sig1"},
				{Version: "1.1", ClientPublicKey: "key2", PublicKey: "pubKey2", Signature: "sig2"},
				{Version: "2.0", ClientPublicKey: "key3", PublicKey: "pubKey3", Signature: "sig3"},
			},
		},
		{
			name: "Should return an empty slice when there are no GigastakeApps",
			chain: Chain{
				GigastakeApps: map[ProtocolID]map[GigastakeAppID]*GigastakeApp{},
			},
			expected: []GigastakeApp{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gigastakeApps := test.chain.GetAllGigastakeApps()
			assert.ElementsMatch(t, test.expected, gigastakeApps)
		})
	}
}

func Test_GetGigastakeAppsByProtocolID(t *testing.T) {
	tests := []struct {
		name       string
		chain      Chain
		protocolID ProtocolID
		expected   []GigastakeApp
	}{
		{
			name: "Should return all gigastake apps for a given protocol ID",
			chain: Chain{
				GigastakeApps: map[ProtocolID]map[GigastakeAppID]*GigastakeApp{
					"protocol_1": {
						"app_1": {Version: "1.0", ClientPublicKey: "key1", PublicKey: "pubKey1", Signature: "sig1"},
						"app_2": {Version: "1.1", ClientPublicKey: "key2", PublicKey: "pubKey2", Signature: "sig2"},
					},
					"protocol_2": {
						"app_3": {Version: "2.0", ClientPublicKey: "key3", PublicKey: "pubKey3", Signature: "sig3"},
					},
				},
			},
			protocolID: "protocol_1",
			expected: []GigastakeApp{
				{Version: "1.0", ClientPublicKey: "key1", PublicKey: "pubKey1", Signature: "sig1"},
				{Version: "1.1", ClientPublicKey: "key2", PublicKey: "pubKey2", Signature: "sig2"},
			},
		},
		{
			name: "Should return an empty slice for a non-existent protocol ID",
			chain: Chain{
				GigastakeApps: map[ProtocolID]map[GigastakeAppID]*GigastakeApp{
					"protocol_1": {
						"app_1": {Version: "1.0", ClientPublicKey: "key1", PublicKey: "pubKey1", Signature: "sig1"},
					},
				},
			},
			protocolID: "protocol_3",
			expected:   []GigastakeApp{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gigastakeApps := test.chain.GetGigastakeAppsByProtocolID(test.protocolID)
			assert.ElementsMatch(t, test.expected, gigastakeApps)
		})
	}
}

func Test_ClearGigastakeApps(t *testing.T) {
	tests := []struct {
		name     string
		chain    Chain
		expected Chain
	}{
		{
			name: "Should clear all GigastakeApps of the chain",
			chain: Chain{
				GigastakeApps: map[ProtocolID]map[GigastakeAppID]*GigastakeApp{
					ProtocolMorseMainnet: {
						"test_app_1": {
							ID: "test_app_1",
						},
						"test_app_2": {
							ID: "test_app_2",
						},
					},
				},
			},
			expected: Chain{
				GigastakeApps: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clearedChain := test.chain.ClearGigastakeApps()
			assert.Equal(t, test.expected, clearedChain)
		})
	}
}

func Test_Chain_IsHTTPS(t *testing.T) {
	tests := []struct {
		name     string
		chain    Chain
		expected bool
	}{
		{
			name:     "should return true for HTTPS scheme",
			chain:    Chain{Scheme: ChainSchemeHTTPS},
			expected: true,
		},
		{
			name:     "should return false for non-HTTPS scheme",
			chain:    Chain{Scheme: ChainSchemeWSS},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.chain.IsHTTPS()
			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_Chain_IsWSS(t *testing.T) {
	tests := []struct {
		name     string
		chain    Chain
		expected bool
	}{
		{
			name:     "should return true for WSS scheme",
			chain:    Chain{Scheme: ChainSchemeWSS},
			expected: true,
		},
		{
			name:     "should return false for non-WSS scheme",
			chain:    Chain{Scheme: ChainSchemeHTTPS},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.chain.IsWSS()
			assert.Equal(t, test.expected, result)
		})
	}
}
