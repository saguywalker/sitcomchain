package app

// SetValidatorParam for updating new validator
type SetValidatorParam struct {
	PublicKey []byte `json:"public_key"`
	Power     int64  `json:"power"`
}

// GiveBadge for adding new data
type GiveBadge struct {
	StudentID    string `json:"student_id"`
	CompetenceID uint32 `json:"competence_id"`
	Semester     uint32 `json:"semester"`
	Giver        []byte `json:"-"`
}
