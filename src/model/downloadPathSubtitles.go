package model

// download_path_subtitles表
type DownloadPathSubtitles struct {
	BaseModel
	DownloadPathId int32  `gorm:"column:download_path_id"`
	StartAt        string `gorm:"start_at"`
	Text           string `gorm:"text"`
}

// 表名
func (t DownloadPathSubtitles) TableName() string {
	return "download_path_subtitles"
}
