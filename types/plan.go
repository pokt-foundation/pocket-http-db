package types

import "time"

/* Enums */
type PayPlanType string

var (
	StarterPlan    PayPlanType = "STARTER"
	AutoscalePlan  PayPlanType = "AUTOSCALE"
	EnterprisePlan PayPlanType = "ENTERPRISE"
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

		// TODO - remove when v2 migration finished
		// LegacyDailyLimit is the daily limit (required for legacy apps to function)
		LegacyDailyLimit int32     `json:"legacyDailyLimit"`
		CreatedAt        time.Time `json:"createdAt"`
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
	case StarterPlan, AutoscalePlan, EnterprisePlan:
		return true
	default:
		return false
	}
}
