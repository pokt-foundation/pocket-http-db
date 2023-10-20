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

func Test_GetGigastakeApps(t *testing.T) {
	tests := []struct {
		name     string
		chain    Chain
		expected []GigastakeApp
	}{
		{
			name: "Should return all gigastake apps of the chain",
			chain: Chain{
				GigastakeApps: map[GigastakeAppID]*GigastakeApp{
					"aat_1": {Version: "0.0.1", ClientPublicKey: "clientPublicKey1", PublicKey: "applicationPublicKey1", Signature: "applicationSignature1"},
					"aat_2": {Version: "0.0.2", ClientPublicKey: "clientPublicKey2", PublicKey: "applicationPublicKey2", Signature: "applicationSignature2"},
				},
			},
			expected: []GigastakeApp{
				{Version: "0.0.1", ClientPublicKey: "clientPublicKey1", PublicKey: "applicationPublicKey1", Signature: "applicationSignature1"},
				{Version: "0.0.2", ClientPublicKey: "clientPublicKey2", PublicKey: "applicationPublicKey2", Signature: "applicationSignature2"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			aats := test.chain.GetGigastakeApps()
			assert.ElementsMatch(t, test.expected, aats)
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
				GigastakeApps: map[GigastakeAppID]*GigastakeApp{
					"test_app_1": {
						ID: "test_app_1",
					},
					"test_app_2": {
						ID: "test_app_2",
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
