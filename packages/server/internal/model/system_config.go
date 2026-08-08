package model

import "time"

// SystemConfig 系统配置项（V7 配置中心）
type SystemConfig struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ConfigKey   string    `json:"config_key" gorm:"size:100;uniqueIndex"`
	ConfigValue string    `json:"config_value" gorm:"type:text"`
	ValueType   string    `json:"value_type" gorm:"size:20;default:string"` // string/int/bool/json
	Description string    `json:"description" gorm:"size:255"`
	IsPublic    bool      `json:"is_public" gorm:"default:false"`
	UpdatedBy   uint      `json:"updated_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserSetting 用户偏好设置（主题/语言）
type UserSetting struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"uniqueIndex"`
	Theme     string    `json:"theme" gorm:"size:20;default:print"` // print/dark/light
	Lang      string    `json:"lang" gorm:"size:10;default:zh-CN"`  // zh-CN/en-US
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SystemVersion 系统版本记录
type SystemVersion struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Version      string    `json:"version" gorm:"size:20"`
	Build        int       `json:"build"`
	Platform     string    `json:"platform" gorm:"size:20;default:all"` // all/h5/mp-weixin/app
	UpdateURL    string    `json:"update_url" gorm:"size:255"`
	ReleaseNotes string    `json:"release_notes" gorm:"type:text"`
	Mandatory    bool      `json:"mandatory" gorm:"default:false"`
	ReleasedAt   time.Time `json:"released_at"`
	CreatedAt    time.Time `json:"created_at"`
}
