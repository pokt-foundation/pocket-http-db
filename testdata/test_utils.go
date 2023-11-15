package testdata

import (
	"sort"

	"github.com/pokt-foundation/portal-http-db/v2/types"
)

/* ---------- Test Utils ---------- */
func AccountsToMap(accounts []types.Account) map[types.AccountID]*types.Account {
	accountsMap := make(map[types.AccountID]*types.Account)
	for i := range accounts {
		accountsMap[accounts[i].ID] = &accounts[i]
	}
	return accountsMap
}

func PortalAppsToMap(portalApps []types.PortalApp) map[types.PortalAppID]*types.PortalApp {
	portalAppsMap := make(map[types.PortalAppID]*types.PortalApp)
	for i := range portalApps {
		portalAppsMap[portalApps[i].ID] = &portalApps[i]
	}
	return portalAppsMap
}

func PortalAppLitesToMap(portalAppLites []types.PortalAppLite) map[types.PortalAppID]*types.PortalAppLite {
	portalAppLitesMap := make(map[types.PortalAppID]*types.PortalAppLite)
	for i := range portalAppLites {
		strKeys := make([]string, len(portalAppLites[i].PublicKeys))
		for j, key := range portalAppLites[i].PublicKeys {
			strKeys[j] = string(key)
		}

		sort.Strings(strKeys)

		for j, key := range strKeys {
			portalAppLites[i].PublicKeys[j] = types.PortalAppPublicKey(key)
		}

		portalAppLitesMap[portalAppLites[i].ID] = &portalAppLites[i]
	}
	return portalAppLitesMap
}

func GigastakeAppsToMap(gigastakeApps []types.GigastakeApp) map[types.GigastakeAppID]*types.GigastakeApp {
	gigastakeAppsMap := make(map[types.GigastakeAppID]*types.GigastakeApp)
	for i := range gigastakeApps {
		gigastakeAppsMap[gigastakeApps[i].ID] = &gigastakeApps[i]
	}
	return gigastakeAppsMap
}

func ChainsToMap(chains []types.Chain) map[types.RelayChainID]*types.Chain {
	chainsMap := make(map[types.RelayChainID]*types.Chain)
	for i := range chains {
		chainsMap[chains[i].ID] = &chains[i]
	}
	return chainsMap
}

func PayPlansToMap(payPlans []types.Plan) map[types.PayPlanType]*types.Plan {
	payPlansMap := make(map[types.PayPlanType]*types.Plan)
	for i := range payPlans {
		payPlansMap[payPlans[i].Type] = &payPlans[i]
	}
	return payPlansMap
}

func FindAccountInAccounts(accounts []types.Account, accountID types.AccountID) bool {
	for _, account := range accounts {
		if account.ID == accountID {
			return true
		}
	}
	return false
}

func GetAccountFromAccounts(accounts []types.Account, accountID types.AccountID) types.Account {
	for _, account := range accounts {
		if account.ID == accountID {
			return account
		}
	}
	return types.Account{}
}

func FindPortalAppInPortalApps(portalApps []types.PortalApp, portalAppID types.PortalAppID) bool {
	for _, portalApp := range portalApps {
		if portalApp.ID == portalAppID {
			return true
		}
	}
	return false
}

func FindGigastakeAppInGigastakeApps(gigastakeApps []types.GigastakeApp, gigastakeAppID types.GigastakeAppID) bool {
	for _, app := range gigastakeApps {
		if app.ID == gigastakeAppID {
			return true
		}
	}
	return false
}

func FindChainInChains(chains []types.Chain, chainID types.RelayChainID) bool {
	for _, chain := range chains {
		if chain.ID == chainID {
			return true
		}
	}
	return false
}

func FilterActiveChains(chains map[types.RelayChainID]*types.Chain) map[types.RelayChainID]*types.Chain {
	activeChains := make(map[types.RelayChainID]*types.Chain)

	for id, chain := range chains {
		if chain.Active {
			activeChains[id] = chain
		}
	}

	return activeChains
}

// chainsToMap takes a slice of pointers to Chains and converts it into a map
// with the Chain ID as the key and the Chain as the value.
func ChainPointersToMap(chains []*types.Chain) map[types.RelayChainID]*types.Chain {
	chainMap := make(map[types.RelayChainID]*types.Chain)
	for _, chain := range chains {
		chainMap[chain.ID] = chain
	}
	return chainMap
}

func GigastakeAppPointersToMap(apps []*types.GigastakeApp) map[types.GigastakeAppID]*types.GigastakeApp {
	appMap := make(map[types.GigastakeAppID]*types.GigastakeApp)
	for _, app := range apps {
		appMap[app.ID] = app
	}
	return appMap
}

func PortalAppPointersToMap(apps []*types.PortalApp) map[types.PortalAppID]*types.PortalApp {
	result := make(map[types.PortalAppID]*types.PortalApp)
	for _, app := range apps {
		result[app.ID] = app
	}
	return result
}

func PortalAppLitePointersToMap(portalAppLites []*types.PortalAppLite) map[types.PortalAppID]*types.PortalAppLite {
	portalAppLitesMap := make(map[types.PortalAppID]*types.PortalAppLite)
	for i := range portalAppLites {
		strKeys := make([]string, len(portalAppLites[i].PublicKeys))
		for j, key := range portalAppLites[i].PublicKeys {
			strKeys[j] = string(key)
		}

		sort.Strings(strKeys)

		for j, key := range strKeys {
			portalAppLites[i].PublicKeys[j] = types.PortalAppPublicKey(key)
		}

		portalAppLitesMap[portalAppLites[i].ID] = portalAppLites[i]
	}
	return portalAppLitesMap
}

func ConvertAccountPointersToMap(accounts []*types.Account) map[types.AccountID]*types.Account {
	accountMap := make(map[types.AccountID]*types.Account)
	for _, account := range accounts {
		accountMap[account.ID] = account
	}
	return accountMap
}

func ConvertPlanPointersToMap(plans []types.Plan) map[types.PayPlanType]types.Plan {
	planMap := make(map[types.PayPlanType]types.Plan)
	for _, plan := range plans {
		planMap[plan.Type] = plan
	}
	return planMap
}

func DerefPlansMap(plans map[types.PayPlanType]*types.Plan) map[types.PayPlanType]types.Plan {
	planMap := make(map[types.PayPlanType]types.Plan)
	for _, plan := range plans {
		planMap[plan.Type] = *plan
	}
	return planMap
}

func BoolToPointer(value bool) *bool {
	return &value
}
