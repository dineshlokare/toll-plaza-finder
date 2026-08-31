package domain

// TollRequest represents the incoming API request payload
type TollRequest struct {
	SourcePincode      string `json:"sourcePincode" binding:"required"`
	DestinationPincode string `json:"destinationPincode" binding:"required"`
}
