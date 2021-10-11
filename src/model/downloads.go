package model

// downloads表
type Downloads struct {
	BaseModel
	Name                string `gorm:"column:name"`
	DownloadUrl         string `gorm:"download_url"`
	DownloadUrlFileName string `gorm:"download_url_file_name"`
	Lan                 string `gorm:"lan"`
	SubtitlesType       string `gorm:"subtitles_type"`
}

// 表名
func (t Downloads) TableName() string {
	return "downloads"
}
