package postgresdriver

import (
	"context"

	"github.com/pokt-foundation/portal-http-db/v2/testdata"
	"github.com/pokt-foundation/portal-http-db/v2/types"
)

func (ts *PGDriverTestSuite) Test_PGDriver_ReadPlans() {
	tests := []struct {
		name  string
		plans map[types.PayPlanType]*types.Plan
		err   error
	}{
		{
			name: "Should return all plans from the database",
			plans: map[types.PayPlanType]*types.Plan{
				types.StarterPlan:    testdata.PayPlans()[types.StarterPlan],
				types.AutoscalePlan:  testdata.PayPlans()[types.AutoscalePlan],
				types.EnterprisePlan: testdata.PayPlans()[types.EnterprisePlan],
			},
			err: nil,
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			plans, err := ts.driver.ReadPlans(context.Background())
			ts.Equal(test.err, err)
			ts.Equal(test.plans, plans)

		})
	}
}
