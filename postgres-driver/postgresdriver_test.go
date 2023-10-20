package postgresdriver

import (
	"context"
)

func (ts *PGDriverTestSuite) Test_PGDriver_GenerateID() {
	tests := []struct {
		name        string
		existingIDs map[string]bool
		expectError bool
	}{
		{
			name:        "Should generate a unique ID that doesn't exist in the DB",
			expectError: false,
		},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			result, err := ts.driver.generateID(context.Background())
			if test.expectError {
				ts.Require().Error(err)
			} else {
				ts.Require().NoError(err)
			}
			ts.NotEmpty(result)
			ts.Len(result, idLength)
		})
	}
}
