package model

// downloads表
type Downloads struct {
	BaseModel
	Name        string `gorm:"column:name"`
	DownloadUrl string `gorm:"downloag_url"`
	FileName    string `gorm:"file_name"`
	Lan         string `gorm:"lan"`
	Subtitles   string `gorm:"subtitles"`
}

// 表名
func (t Downloads) TableName() string {
	return "downloads"
}
