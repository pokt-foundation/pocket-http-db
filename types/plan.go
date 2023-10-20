package types

import "time"

/* Enums */
type PayPlanType string

var (
	TestPlanV0   PayPlanType = "TEST_PLAN_V0"
	TestPlan10K  PayPlanType = "TEST_PLAN_10K"
	TestPlan90k  PayPlanType = "TEST_PLAN_90K"
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
	case TestPlanV0, TestPlan10K, TestPlan90k, FreetierV0, PayAsYouGoV0, Enterprise:
		return true
	default:
		return false
	}
}
