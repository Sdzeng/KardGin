package model

// download_paths表
type DownloadPaths struct {
	BaseModel
	DownloadId int32  `gorm:"column:download_id"`
	FilePath   string `gorm:"file_path"`
}

// 表名
func (t DownloadPaths) TableName() string {
	return "download_paths"
}
