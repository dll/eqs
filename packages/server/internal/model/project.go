package model

import "time"

type Project struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"index"`
	ProjectType string    `json:"project_type" gorm:"size:50"`
	Title       string    `json:"title" gorm:"size:200"`
	BudgetMin   float64   `json:"budget_min"`
	BudgetMax   float64   `json:"budget_max"`
	Status      int       `json:"status" gorm:"default:0"` // 0:草稿 1:已发布 2:已接单 3:进行中 4:已完成
	PublishTime *time.Time `json:"publish_time"`
	Deadline    *time.Time `json:"deadline"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	User        User       `json:"user" gorm:"foreignKey:UserID"`
}
