package model

import "time"

// razors表
type Razors struct {
	BaseModel
	Razor      string    `gorm:"razor"`
	SeedUrl    string    `gorm:"seed_url"`
	Page       int       `gorm:"page"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

// 表名
func (t Razors) TableName() string {
	return "razors"
}
