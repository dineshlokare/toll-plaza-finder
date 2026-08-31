package validator

import (
	"strings"
	"toll_plaza/internal/domain"
	"unicode"
)

// IsValidIndianPincode validates whether a given string is a valid 6-digit Indian PIN code.
func IsValidIndianPincode(pincode string) bool {
	pin := strings.TrimSpace(pincode)
	if len(pin) != 6 {
		return false
	}
	// First digit must be 1-9 (Indian pincodes do not start with 0)
	if pin[0] < '1' || pin[0] > '9' {
		return false
	}
	for _, ch := range pin {
		if !unicode.IsDigit(ch) {
			return false
		}
	}
	return true
}

// ValidateTollRequest checks the validity of source and destination pincodes in the request.
func ValidateTollRequest(req *domain.TollRequest) error {
	src := strings.TrimSpace(req.SourcePincode)
	dst := strings.TrimSpace(req.DestinationPincode)

	if !IsValidIndianPincode(src) || !IsValidIndianPincode(dst) {
		return domain.ErrInvalidPincode
	}

	if src == dst {
		return domain.ErrSamePincode
	}

	return nil
}
