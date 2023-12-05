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
				types.StarterPlan:    testdata.PayPlans()["FREETIER_V0"],
				types.AutoscalePlan:  testdata.PayPlans()["PAY_AS_YOU_GO_V0"],
				types.EnterprisePlan: testdata.PayPlans()["ENTERPRISE"],
				types.TestPlan:       testdata.PayPlans()["TEST_PLAN_V0"],
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
