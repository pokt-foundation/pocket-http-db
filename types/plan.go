package types

/* Enums */
type PayPlanType string

var (
	// CURRENT PLAN TYPES
	PlanFree      PayPlanType = "PLAN_FREE"
	PlanUnlimited PayPlanType = "PLAN_UNLIMITED"

	// DEPRECATED PLAN TYPES - Remove Eventually
	FreetierV0   PayPlanType = "FREETIER_V0"
	PayAsYouGoV0 PayPlanType = "PAY_AS_YOU_GO_V0"
	Enterprise   PayPlanType = "ENTERPRISE"
)

/* Pay Plan Type and Methods */
type (
	Plan struct {
		ID          string                    `json:"id"`
		Name        string                    `json:"name"`
		Description string                    `json:"description"`
		Type        PayPlanType               `json:"planType"`
		ChainIDs    map[RelayChainID]struct{} `json:"chainIDs"`
		// MonthlyRelayLimit is the number of relays-per-month for a pay plan
		MonthlyRelayLimit int32 `json:"monthlyRelayLimit"`
		// ThroughputLimit is the number of relays-per-second for a pay plan
		ThroughputLimit int32 `json:"throughputLimit"`
		// AppLimit is the number of apps permitted for a pay plan
		AppLimit int32 `json:"appLimit"`

		// DEPRECATED - Remove Eventually
		// DailyLimit is the daily relay limit for a pay plan
		DailyLimit int32 `json:"dailyLimit"`
	}
)

func (p *Plan) IsEmpty() bool {
	return p == nil || (p.Type == "" && len(p.ChainIDs) == 0 && p.ThroughputLimit == 0)
}

func (p *Plan) Table() Table {
	return TablePayPlans
}

func (p PayPlanType) Valid() bool {
	switch p {
	case
		// CURRENT PLAN TYPES
		PlanFree, PlanUnlimited,
		// DEPRECATED PLAN TYPES - Remove Eventually
		FreetierV0, PayAsYouGoV0, Enterprise:
		return true
	default:
		return false
	}
}
