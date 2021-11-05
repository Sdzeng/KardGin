package model

import "time"

// downloads表
type Downloads struct {
	BaseModel
	EsIndex     string `gorm:"es_index"` //es索引名称
	Name        string `gorm:"column:name"`
	Crawler     string `gorm:"column:crawler"`
	InfoUrl     string `gorm:"info_url"`
	DownloadUrl string `gorm:"download_url"`
	// DownloadUrlFileName string `gorm:"download_url_file_name"`
	Lan           string `gorm:"lan"`
	SubtitlesType string `gorm:"subtitles_type"`
	PicPath       string `gorm:"pic_path"`
	Making        string `gorm:"making"` //制作
	Edit          string `gorm:"edit"`   //校订
	Source        string `gorm:"source"` //来源

	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

// 表名
func (t Downloads) TableName() string {
	return "downloads"
}
