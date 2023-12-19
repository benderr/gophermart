package utils_test

import (
	"testing"

	"github.com/benderr/gophermart/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidateOrder(t *testing.T) {

	tests := []struct {
		number string
		valid  bool
	}{
		{
			number: "4242424242424242",
			valid:  true,
		},
		{
			number: "4012888888881881",
			valid:  true,
		},
		{
			number: "5496198584584769",
			valid:  true,
		},
		{
			number: "5555555555554444",
			valid:  true,
		},
		{
			number: "4111110000000112",
			valid:  true,
		},
		{
			number: "4000000000000069",
			valid:  true,
		},
		{
			number: "4000000000000002",
			valid:  true,
		},
		{
			number: "5105105105105100",
			valid:  true,
		},
		{
			number: "510510510517777",
			valid:  false,
		},
		{
			number: "123",
			valid:  false,
		},
		{
			number: "0001",
			valid:  false,
		},
		{
			number: "620282061",
			valid:  true,
		},
	}

	for _, test := range tests {
		t.Run("Validate "+test.number, func(t *testing.T) {

			err := utils.ValidateOrder(test.number)

			if test.valid {
				assert.NoError(t, err, "error validate")
			} else {
				assert.ErrorIs(t, err, utils.ErrInvalidNumber)
			}
		})
	}
}
