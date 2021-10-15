package repository

import (
	"kard/src/model"

	"gorm.io/gorm"
)

type DownloadRepository struct {
	IsEnable bool
	*gorm.DB
}

// 创建 DownloadFactory
func DownloadFactory() *DownloadRepository {
	db := UseDbConn()
	isEnable := db != nil
	return &DownloadRepository{IsEnable: isEnable, DB: db}
}

func (repository *DownloadRepository) Exists(downloadUrl string) bool {
	if !repository.IsEnable {
		return false
	}

	dl := new(model.Downloads)
	repository.DB.First(dl, &model.Downloads{DownloadUrl: downloadUrl})
	return dl.Id > 0
}
