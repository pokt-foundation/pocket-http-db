package types

import "time"

/* Enums */
type PayPlanType string

var (
	StarterPlan    PayPlanType = "FREETIER_V0"      // `pay_plans.plan_type` in DB is `FREETIER_V0``
	AutoscalePlan  PayPlanType = "PAY_AS_YOU_GO_V0" // `pay_plans.plan_type` in DB is `PAY_AS_YOU_GO_V0``
	EnterprisePlan PayPlanType = "ENTERPRISE"       // `pay_plans.plan_type` in DB is `ENTERPRISE``
	TestPlan       PayPlanType = "TEST_PLAN_V0"     // `pay_plans.plan_type` in DB is `TEST_PLAN_V0``
)

/* Pay Plan Type and Methods */
type (
	Plan struct {
		ID          string                    `json:"id"`
		Name        string                    `json:"name"`
		Description string                    `json:"description"`
		Type        PayPlanType               `json:"planType"`
		ChainIDs    map[RelayChainID]struct{} `json:"chainIDs"`
		// DailyLimit is the daily relay limit for a pay plan
		DailyLimit int32 `json:"dailyLimit"`
		// ThroughputLimit is the number of relays-per-second for a pay plan
		ThroughputLimit int32 `json:"throughputLimit"`
		// AppLimit is the number of apps permitted for a pay plan
		AppLimit int32 `json:"appLimit"`
		// CURRENTLY UNUSED - MonthlyRelayLimit is the number of relays-per-month for a pay plan
		MonthlyRelayLimit int32     `json:"monthlyRelayLimit"`
		CreatedAt         time.Time `json:"createdAt"`
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
	case StarterPlan, AutoscalePlan, EnterprisePlan, TestPlan:
		return true
	default:
		return false
	}
}
