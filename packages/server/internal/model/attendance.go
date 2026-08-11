package model

import "time"

// SupplierQualification 服务方资质
// verification_method: OCR/manual/external
// verification_status: pending/approved/rejected/expired
type SupplierQualification struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	SupplierID         uint       `json:"supplier_id" gorm:"index"`
	QualificationType  string     `json:"qualification_type" gorm:"size:50"`
	CertificateNo      EncryptedString `json:"certificate_no"` // P1-09：透明加密，列中存密文
	Level              string     `json:"level" gorm:"size:50"`
	Scope              string     `json:"scope" gorm:"size:200"`
	ValidFrom          *time.Time `json:"valid_from"`
	ValidTo            *time.Time `json:"valid_to"`
	VerificationMethod string     `json:"verification_method" gorm:"size:20;default:manual"`
	VerificationStatus string     `json:"verification_status" gorm:"size:20;default:pending"`
	EvidenceFileID     uint       `json:"evidence_file_id"`
	ReviewedBy         uint       `json:"reviewed_by"`
	ReviewedAt         *time.Time `json:"reviewed_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// AttendanceRecord 现场打卡
// verification_status: valid/exception/manual_approved/rejected
type AttendanceRecord struct {
	ID                 uint            `json:"id" gorm:"primaryKey"`
	OrderID            uint            `json:"order_id" gorm:"index"`
	UserID             uint            `json:"user_id" gorm:"index"`
	CheckInAt          time.Time       `json:"check_in_at"`
	Longitude          EncryptedFloat  `json:"longitude"` // P1-09：透明加密，列中存密文
	Latitude           EncryptedFloat  `json:"latitude"`
	DistanceMeters     int             `json:"distance_meters"`
	EvidenceFileID     uint            `json:"evidence_file_id"`
	VerificationStatus string          `json:"verification_status" gorm:"size:20;default:ok"`
	CreatedAt          time.Time       `json:"created_at"`
}