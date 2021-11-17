package repository

import (
	"kard/src/model"

	"gorm.io/gorm"
)

type DownloadRefersRepository struct {
	IsEnable bool
	*gorm.DB
}

// 创建 DownloadRefersFactory
func DownloadRefersFactory() *DownloadRefersRepository {
	db := UseDbConn()
	isEnable := db != nil
	return &DownloadRefersRepository{IsEnable: isEnable, DB: db}
}

func (repository *DownloadRefersRepository) KFind(downloadId int32) []*model.DownloadRefers {
	if !repository.IsEnable {
		return nil
	}

	refers := []*model.DownloadRefers{}
	// repository.DB.Where("download_url=?", taskDto.DownloadUrl).Or("name=? and lan=?", taskDto.Name, taskDto.Lan).First(dl)
	repository.DB.Where("download_id= ?", downloadId).Order("sort asc").Find(&refers)
	return refers
}
