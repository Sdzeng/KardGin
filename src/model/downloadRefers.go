package model

// download_refers表
type DownloadRefers struct {
	BaseModel
	DownloadId int32  `gorm:"column:download_id"`
	Refer      string `gorm:"refer"`
	Sort       int32  `gorm:"sort"`
}

// 表名
func (t DownloadRefers) TableName() string {
	return "download_refers"
}
