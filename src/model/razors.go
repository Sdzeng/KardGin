package model

import "time"

// razors表
type Razors struct {
	BaseModel
	Razor      string    `gorm:"razor"`
	EsIndex    string    `gorm:"es_index"` //es索引名称
	Domain     string    `gorm:"domain"`
	PathUrl    string    `gorm:"path_url"`
	Page       int       `gorm:"page"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

// 表名
func (t Razors) TableName() string {
	return "razors"
}
