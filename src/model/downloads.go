package model

// downloads表
type Downloads struct {
	BaseModel
	Name                string `gorm:"column:name"`
	DownloadUrl         string `gorm:"download_url"`
	DownloadUrlFileName string `gorm:"download_url_file_name"`
	Lan                 string `gorm:"lan"`
	SubtitlesType       string `gorm:"subtitles_type"`
	PicPath             string `gorm:"pic_path"`
	Making              string `gorm:"making"`   //制作
	Edit                string `gorm:"edit"`     //校订
	Source              string `gorm:"source"`   //来源
	EsIndex             string `gorm:"es_index"` //es索引名称
}

// 表名
func (t Downloads) TableName() string {
	return "downloads"
}
