package model

// StaffAddCompetence contains data that will be stored into blockchain
type StaffAddCompetence struct {
	StudentID    string `json:"student_id"`
	CompetenceID uint16 `json:"competence_id"`
	By           string `json:"by"`
	Semester     uint16 `json:"semester"`
	Nonce        uint64 `json:"nonce"`
}

// AttendedActivity contains data that will be stored into blockchain
type AttendedActivity struct {
	StudentID  string `json:"student_id"`
	ActivityID uint32 `json:"activity_id"`
	Approver   []byte `json:"approver"`
	Semester   uint16 `json:"semester"`
	Nonce      uint64 `json:"nonce"`
}
