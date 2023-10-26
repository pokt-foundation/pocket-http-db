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
				"FREETIER_V0":      testdata.PayPlans()["FREETIER_V0"],
				"PAY_AS_YOU_GO_V0": testdata.PayPlans()["PAY_AS_YOU_GO_V0"],
				"ENTERPRISE":       testdata.PayPlans()["ENTERPRISE"],
				"TEST_PLAN_V0":     testdata.PayPlans()["TEST_PLAN_V0"],
				"TEST_PLAN_90K":    testdata.PayPlans()["TEST_PLAN_90K"],
				"TEST_PLAN_10K":    testdata.PayPlans()["TEST_PLAN_10K"],
				"basic_plan":       testdata.PayPlans()["basic_plan"],
				"pro_plan":         testdata.PayPlans()["pro_plan"],
				"startup_plan":     testdata.PayPlans()["startup_plan"],
				"developer_plan":   testdata.PayPlans()["developer_plan"],
				"enterprise_plan":  testdata.PayPlans()["enterprise_plan"],
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
