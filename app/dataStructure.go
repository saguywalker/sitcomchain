package app

// SetValidatorParam for updating new validator
type SetValidatorParam struct {
	PublicKey string `json:"public_key"`
	Power     int64  `json:"power"`
}
