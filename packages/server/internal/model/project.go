package model

import "time"

// Project 项目表
// status: 0-草稿 1-已发布 2-已接单 3-进行中 4-已完成
type Project struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"index"`
	ProjectType string    `json:"project_type" gorm:"size:50"`
	ServiceType string    `json:"service_type" gorm:"size:50"` // cost/supervision/geotech/design
	Title       string    `json:"title" gorm:"size:200"`
	Description string    `json:"description" gorm:"type:text"`
	Address     string    `json:"address" gorm:"size:300"`
	Longitude   float64   `json:"longitude"`
	Latitude    float64   `json:"latitude"`
	BudgetMin   float64   `json:"budget_min"`
	BudgetMax   float64   `json:"budget_max"`
	PublishScope string   `json:"publish_scope" gorm:"size:20;default:public"` // public/invited
	Theme       string   `json:"theme" gorm:"size:20"`                        // 项目主题：空=跟随系统, print/dark/light
	Status      int       `json:"status" gorm:"default:0"`
	PublishTime *time.Time `json:"publish_time"`
	Deadline    *time.Time `json:"deadline"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	User        User      `json:"user" gorm:"foreignKey:UserID"`
}