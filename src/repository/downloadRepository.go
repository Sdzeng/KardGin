package repository

import (
	"kard/src/model"
	"kard/src/model/dto"

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

func (repository *DownloadRepository) Exists(taskDto *dto.TaskDto) bool {
	if !repository.IsEnable {
		return false
	}

	dl := new(model.Downloads)
	repository.DB.Where("download_url=?", taskDto.DownloadUrl).Or("name=? and lan=?", taskDto.Name, taskDto.Lan).First(dl)
	return dl.Id > 0
}
