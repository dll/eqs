package model

import "time"

// CaseShowcase 服务案例（企业案例沉淀：服务方展示已完成项目，作为获客资产）
// status: published / hidden（管理员可隐藏违规案例）
type CaseShowcase struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	SupplierID    uint      `json:"supplier_id" gorm:"index"` // 展示方（服务方）
	ProjectID     uint      `json:"project_id"`               // 关联项目（可选）
	OrderID       uint      `json:"order_id"`                 // 关联订单（可选，已完成订单可一键沉淀为案例）
	Title         string    `json:"title" gorm:"size:200" binding:"required"`
	Description   string    `json:"description" gorm:"type:text"`
	ServiceType   string    `json:"service_type" gorm:"size:50"` // cost/supervision/geotech/design
	ImageFileIDs  string    `json:"image_file_ids" gorm:"type:text"` // JSON 数组：成果图/现场照片 file_id 列表
	Status        string    `json:"status" gorm:"size:20;default:published"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Supplier      User      `json:"supplier" gorm:"foreignKey:SupplierID"`
}
