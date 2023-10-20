package driver

import (
	"testing"

	postgresdriver "github.com/pokt-foundation/portal-http-db/v2/postgres-driver"
	"github.com/stretchr/testify/assert"
)

func Test_PostgresDriverImplementsDriverInterface(t *testing.T) {
	tests := []struct {
		name   string
		driver Driver
	}{
		{
			name:   "Should verify that PostgresDriver implements the Driver interface",
			driver: &postgresdriver.PostgresDriver{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Implements(t, (*Driver)(nil), test.driver)
		})
	}
}
