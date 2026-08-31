package validator_test

import (
	"testing"
	"toll_plaza/internal/domain"
	"toll_plaza/internal/validator"

	"github.com/stretchr/testify/assert"
)

func TestIsValidIndianPincode(t *testing.T) {
	tests := []struct {
		name    string
		pincode string
		valid   bool
	}{
		{"Valid Delhi", "110001", true},
		{"Valid Bengaluru", "560064", true},
		{"Valid Pune", "411045", true},
		{"Invalid starts with 0", "010001", false},
		{"Invalid short length", "56006", false},
		{"Invalid long length", "5600641", false},
		{"Invalid alphanumeric", "56006A", false},
		{"Invalid special chars", "560-64", false},
		{"Invalid empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, validator.IsValidIndianPincode(tt.pincode))
		})
	}
}

func TestValidateTollRequest(t *testing.T) {
	// Valid request
	err := validator.ValidateTollRequest(&domain.TollRequest{
		SourcePincode:      "110001",
		DestinationPincode: "560001",
	})
	assert.NoError(t, err)

	// Same pincode error
	errSame := validator.ValidateTollRequest(&domain.TollRequest{
		SourcePincode:      "110001",
		DestinationPincode: "110001",
	})
	assert.Equal(t, domain.ErrSamePincode, errSame)

	// Invalid source pincode
	errInvalidSrc := validator.ValidateTollRequest(&domain.TollRequest{
		SourcePincode:      "010001",
		DestinationPincode: "560001",
	})
	assert.Equal(t, domain.ErrInvalidPincode, errInvalidSrc)

	// Invalid destination pincode
	errInvalidDst := validator.ValidateTollRequest(&domain.TollRequest{
		SourcePincode:      "110001",
		DestinationPincode: "ABCDEF",
	})
	assert.Equal(t, domain.ErrInvalidPincode, errInvalidDst)
}
