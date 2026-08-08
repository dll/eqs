package model

import "time"

// ProjectFile 项目文件
// file_type: pdf/image/dwg/dxf/document
type ProjectFile struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ProjectID    uint      `json:"project_id" gorm:"index"`
	OrderID      uint      `json:"order_id"`
	UploaderID   uint      `json:"uploader_id"`
	OriginalName string    `json:"original_name" gorm:"size:200"`
	FileType     string    `json:"file_type" gorm:"size:20"`
	StorageKey   string    `json:"storage_key" gorm:"size:500"`
	Version      int       `json:"version" gorm:"default:1"`
	ParentFileID uint      `json:"parent_file_id"`
	SHA256       string    `json:"sha256" gorm:"size:64"`
	CreatedAt    time.Time `json:"created_at"`
}

// FileAnnotation 文件批注
type FileAnnotation struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	FileID    uint      `json:"file_id" gorm:"index"`
	AuthorID  uint      `json:"author_id"`
	PageNo    int       `json:"page_no"`
	XRatio    float64   `json:"x_ratio"`
	YRatio    float64   `json:"y_ratio"`
	Content   string    `json:"content" gorm:"type:text"`
	Status    string    `json:"status" gorm:"size:20;default:active"` // active/resolved/deleted
	CreatedAt time.Time `json:"created_at"`
}