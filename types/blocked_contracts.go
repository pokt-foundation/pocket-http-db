package types

type (
	// Represents global blocked addresses across the entire Portal
	GlobalBlockedContracts struct {
		BlockedAddresses map[BlockedAddress]struct{} `json:"blockedAddresses"`
	}

	BlockedContract struct {
		BlockedAddress BlockedAddress `json:"blocked_address"`
		Active         bool           `json:"active"`
	}
)

func (g *GlobalBlockedContracts) IsContractBlocked(contract BlockedAddress) bool {
	_, ok := g.BlockedAddresses[contract]
	return ok
}

func (u *BlockedContract) Table() Table {
	return TableGlobalBlockedContracts
}
