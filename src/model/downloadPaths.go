package model

// download_paths表
type DownloadPaths struct {
	BaseModel
	DownloadId int32  `gorm:"column:download_id"`
	Name       string `gorm:"name"`
	FileName   string `gorm:"file_name"`
	FilePath   string `gorm:"file_path"`
	FileSum    string `gorm:"file_sum"`
	Remark     string `gorm:"remark"`
}

// 表名
func (t DownloadPaths) TableName() string {
	return "download_paths"
}
