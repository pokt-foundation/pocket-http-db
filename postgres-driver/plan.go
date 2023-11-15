package postgresdriver

import (
	"context"
	"time"

	"github.com/pokt-foundation/portal-http-db/v2/types"
)

/* ----- postgresdriver Plan Read Methods ----- */

// ReadPlans returns all Chains in the database as Chains structs
func (pg *PostgresDriver) ReadPlans(ctx context.Context) (map[types.PayPlanType]*types.Plan, error) {
	dbPlans, err := pg.SelectPlans(ctx)
	if err != nil {
		return nil, err
	}

	plans := make(map[types.PayPlanType]*types.Plan, len(dbPlans))
	for _, dbPlan := range dbPlans {
		plan, err := dbPlan.toPlan()
		if err != nil {
			return nil, err
		}

		plans[dbPlan.PlanType] = plan
	}

	return plans, nil
}

// toPlan converts PayPlan to Plan struct
func (p *SelectPlansRow) toPlan() (*types.Plan, error) {
	plan := &types.Plan{
		Name:              p.Name.String,
		Description:       p.Description.String,
		Type:              p.PlanType,
		ChainIDs:          make(map[types.RelayChainID]struct{}, len(p.ChainIDs)),
		MonthlyRelayLimit: p.MonthlyRelayLimit,
		ThroughputLimit:   p.ThroughputLimit,
		AppLimit:          p.ApplicationLimit,
		DailyLimit:        p.DailyLimit.Int32,
		CreatedAt:         p.CreatedAt.Time.UTC(),
	}

	for _, chainID := range p.ChainIDs {
		plan.ChainIDs[types.RelayChainID(chainID)] = struct{}{}
	}

	return plan, nil
}

/* ----- Used by Listener ----- */
func (json dbPayPlan) toOutput() *types.Plan {
	chainIDs := make(map[types.RelayChainID]struct{})
	for _, id := range json.ChainIDs {
		chainIDs[types.RelayChainID(id)] = struct{}{}
	}

	return &types.Plan{
		Name:              json.Name,
		Description:       json.Description,
		Type:              json.PlanType,
		ChainIDs:          chainIDs,
		MonthlyRelayLimit: json.MonthlyRelayLimit,
		ThroughputLimit:   json.ThroughputLimit,
		AppLimit:          json.ApplicationLimit,
		DailyLimit:        json.DailyLimit,
	}
}

type dbPayPlan struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	PlanType          types.PayPlanType `json:"plan_type"`
	ChainIDs          []string          `json:"chain_ids"`
	MonthlyRelayLimit int32             `json:"monthly_relay_limit"`
	ThroughputLimit   int32             `json:"throughput_limit"`
	ApplicationLimit  int32             `json:"application_limit"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	DailyLimit        int32             `json:"daily_limit"`
}
