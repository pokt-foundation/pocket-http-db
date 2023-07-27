package phdtestdata

import (
	"sort"
	"time"

	v1Types "github.com/pokt-foundation/portal-db/types"
	"github.com/pokt-foundation/portal-db/v2/testdata"
	v2Types "github.com/pokt-foundation/portal-db/v2/types"
)

/* ---------- Test Utils ---------- */
func AccountsToMap(accounts []v2Types.Account) map[v2Types.AccountID]*v2Types.Account {
	accountsMap := make(map[v2Types.AccountID]*v2Types.Account)
	for i := range accounts {
		accountsMap[accounts[i].ID] = &accounts[i]
	}
	return accountsMap
}

func PortalAppsToMap(portalApps []v2Types.PortalApp) map[v2Types.PortalAppID]*v2Types.PortalApp {
	portalAppsMap := make(map[v2Types.PortalAppID]*v2Types.PortalApp)
	for i := range portalApps {
		portalAppsMap[portalApps[i].ID] = &portalApps[i]
	}
	return portalAppsMap
}

func PortalAppLitesToMap(portalAppLites []v2Types.PortalAppLite) map[v2Types.PortalAppID]*v2Types.PortalAppLite {
	portalAppLitesMap := make(map[v2Types.PortalAppID]*v2Types.PortalAppLite)
	for i := range portalAppLites {
		strKeys := make([]string, len(portalAppLites[i].PublicKeys))
		for j, key := range portalAppLites[i].PublicKeys {
			strKeys[j] = string(key)
		}

		sort.Strings(strKeys)

		for j, key := range strKeys {
			portalAppLites[i].PublicKeys[j] = v2Types.PortalAppPublicKey(key)
		}

		portalAppLitesMap[portalAppLites[i].ID] = &portalAppLites[i]
	}
	return portalAppLitesMap
}

func GigastakeAppsToMap(gigastakeApps []v2Types.GigastakeApp) map[v2Types.GigastakeAppID]*v2Types.GigastakeApp {
	gigastakeAppsMap := make(map[v2Types.GigastakeAppID]*v2Types.GigastakeApp)
	for i := range gigastakeApps {
		gigastakeAppsMap[gigastakeApps[i].ID] = &gigastakeApps[i]
	}
	return gigastakeAppsMap
}

func ChainsToMap(chains []v2Types.Chain) map[v2Types.RelayChainID]*v2Types.Chain {
	chainsMap := make(map[v2Types.RelayChainID]*v2Types.Chain)
	for i := range chains {
		chainsMap[chains[i].ID] = &chains[i]
	}
	return chainsMap
}

func PayPlansToMap(payPlans []v2Types.Plan) map[v2Types.PayPlanType]*v2Types.Plan {
	payPlansMap := make(map[v2Types.PayPlanType]*v2Types.Plan)
	for i := range payPlans {
		payPlansMap[payPlans[i].Type] = &payPlans[i]
	}
	return payPlansMap
}

func FindAccountInAccounts(accounts []v2Types.Account, accountID v2Types.AccountID) bool {
	for _, account := range accounts {
		if account.ID == accountID {
			return true
		}
	}
	return false
}

func GetAccountFromAccounts(accounts []v2Types.Account, accountID v2Types.AccountID) v2Types.Account {
	for _, account := range accounts {
		if account.ID == accountID {
			return account
		}
	}
	return v2Types.Account{}
}

func FindPortalAppInPortalApps(portalApps []v2Types.PortalApp, portalAppID v2Types.PortalAppID) bool {
	for _, portalApp := range portalApps {
		if portalApp.ID == portalAppID {
			return true
		}
	}
	return false
}

func FindGigastakeAppInGigastakeApps(gigastakeApps []v2Types.GigastakeApp, gigastakeAppID v2Types.GigastakeAppID) bool {
	for _, app := range gigastakeApps {
		if app.ID == gigastakeAppID {
			return true
		}
	}
	return false
}

func FindChainInChains(chains []v2Types.Chain, chainID v2Types.RelayChainID) bool {
	for _, chain := range chains {
		if chain.ID == chainID {
			return true
		}
	}
	return false
}

func FilterActiveBlockchains(blockchains map[string]v1Types.Blockchain) map[string]v1Types.Blockchain {
	activeBlockchains := make(map[string]v1Types.Blockchain)

	for id, blockchain := range blockchains {
		if blockchain.Active {
			activeBlockchains[id] = blockchain
		}
	}

	return activeBlockchains
}

func FilterActiveChains(chains map[v2Types.RelayChainID]*v2Types.Chain) map[v2Types.RelayChainID]*v2Types.Chain {
	activeChains := make(map[v2Types.RelayChainID]*v2Types.Chain)

	for id, chain := range chains {
		if chain.Active {
			activeChains[id] = chain
		}
	}

	return activeChains
}

// This functIon is needed because the cache alter the DB output
func ModifiedAccountToTest() map[v2Types.AccountID]*v2Types.Account {
	portalAppsCopy := testdata.PortalApps
	payPlanCopy := testdata.PayPlans
	accCopy := testdata.Accounts
	accCopy["account_1"].Plan = payPlanCopy["basic_plan"]
	accCopy["account_2"].Plan = payPlanCopy["pro_plan"]
	accCopy["account_2"].PortalApps["test_app_2"] = portalAppsCopy["test_app_2"]
	accCopy["account_3"].Plan = payPlanCopy["startup_plan"]
	accCopy["account_4"].Plan = payPlanCopy["enterprise_plan"]
	accCopy["account_5"].Plan = payPlanCopy["basic_plan"]

	return accCopy
}

// chainsToMap takes a slice of pointers to Chains and converts it into a map
// with the Chain ID as the key and the Chain as the value.
func ChainPointersToMap(chains []*v2Types.Chain) map[v2Types.RelayChainID]*v2Types.Chain {
	chainMap := make(map[v2Types.RelayChainID]*v2Types.Chain)
	for _, chain := range chains {
		chainMap[chain.ID] = chain
	}
	return chainMap
}

func PortalAppPointersToMap(apps []*v2Types.PortalApp) map[v2Types.PortalAppID]*v2Types.PortalApp {
	result := make(map[v2Types.PortalAppID]*v2Types.PortalApp)
	for _, app := range apps {
		result[app.ID] = app
	}
	return result
}

func PortalAppLitePointersToMap(portalAppLites []*v2Types.PortalAppLite) map[v2Types.PortalAppID]*v2Types.PortalAppLite {
	portalAppLitesMap := make(map[v2Types.PortalAppID]*v2Types.PortalAppLite)
	for i := range portalAppLites {
		strKeys := make([]string, len(portalAppLites[i].PublicKeys))
		for j, key := range portalAppLites[i].PublicKeys {
			strKeys[j] = string(key)
		}

		sort.Strings(strKeys)

		for j, key := range strKeys {
			portalAppLites[i].PublicKeys[j] = v2Types.PortalAppPublicKey(key)
		}

		portalAppLitesMap[portalAppLites[i].ID] = portalAppLites[i]
	}
	return portalAppLitesMap
}

func ConvertAccountPointersToMap(accounts []*v2Types.Account) map[v2Types.AccountID]*v2Types.Account {
	accountMap := make(map[v2Types.AccountID]*v2Types.Account)
	for _, account := range accounts {
		accountMap[account.ID] = account
	}
	return accountMap
}

func ConvertPlanPointersToMap(plans []v2Types.Plan) map[v2Types.PayPlanType]v2Types.Plan {
	planMap := make(map[v2Types.PayPlanType]v2Types.Plan)
	for _, plan := range plans {
		planMap[plan.Type] = plan
	}
	return planMap
}

func DerefPlansMap(plans map[v2Types.PayPlanType]*v2Types.Plan) map[v2Types.PayPlanType]v2Types.Plan {
	planMap := make(map[v2Types.PayPlanType]v2Types.Plan)
	for _, plan := range plans {
		planMap[plan.Type] = *plan
	}
	return planMap
}

func BlockchainPointersToMap(blockchains []*v1Types.Blockchain) map[string]*v1Types.Blockchain {
	expectedMap := make(map[string]*v1Types.Blockchain)
	for _, b := range blockchains {
		expectedMap[b.ID] = b
	}
	return expectedMap
}

func ApplicationPointersToMap(applications []*v1Types.Application) map[string]*v1Types.Application {
	expectedMap := make(map[string]*v1Types.Application)
	for _, a := range applications {
		expectedMap[a.ID] = a
	}
	return expectedMap
}

func LoadBalancerPointersToMap(loadBalancers []*v1Types.LoadBalancer) map[string]*v1Types.LoadBalancer {
	lbMap := make(map[string]*v1Types.LoadBalancer)
	for _, lb := range loadBalancers {
		SortUsersByRole(lb.Users)

		// Sort Applications by ID
		sort.Slice(lb.Applications, func(i, j int) bool {
			return lb.Applications[i].ID < lb.Applications[j].ID
		})

		lbMap[lb.ID] = lb
	}
	return lbMap
}

func SortUsersByRole(users []v1Types.UserAccess) {
	roleWeight := map[v1Types.RoleName]int{v1Types.RoleOwner: 0, v1Types.RoleAdmin: 1, v1Types.RoleMember: 2}

	sort.Slice(users, func(i, j int) bool {
		if roleWeight[users[i].RoleName] != roleWeight[users[j].RoleName] {
			return roleWeight[users[i].RoleName] < roleWeight[users[j].RoleName]
		}
		return users[i].UserID < users[j].UserID
	})
}

func PayPlanPointersToMap(payPlans []*v1Types.PayPlan) map[string]*v1Types.PayPlan {
	payPlansMap := make(map[string]*v1Types.PayPlan)
	for _, payPlan := range payPlans {
		payPlansMap[string(payPlan.Type)] = payPlan
	}
	return payPlansMap
}

func UserAccessSliceToMap(users []v1Types.UserAccess) map[string]v1Types.UserAccess {
	userMap := make(map[string]v1Types.UserAccess)
	for _, user := range users {
		userMap[user.UserID] = user
	}
	return userMap
}

func ClearTimeFields(lb *v1Types.LoadBalancer) {
	lb.CreatedAt = time.Time{}
	lb.UpdatedAt = time.Time{}
	if len(lb.Applications) > 0 {
		lb.Applications[0].CreatedAt = time.Time{}
		lb.Applications[0].UpdatedAt = time.Time{}
	}
}

func BoolToPointer(value bool) *bool {
	return &value
}
