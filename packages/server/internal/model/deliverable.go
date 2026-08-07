package model

import "time"

type Deliverable struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	OrderID   uint      `json:"order_id" gorm:"index"`
	Milestone string    `json:"milestone" gorm:"size:50"`
	FileURL   string    `json:"file_url" gorm:"size:500"`
	Version   int       `json:"version" gorm:"default:1"`
	Status    int       `json:"status" gorm:"default:0"` // 0:待审核 1:已通过 2:已驳回
	CreatedAt time.Time `json:"created_at"`
	Order     Order     `json:"order" gorm:"foreignKey:OrderID"`
}
