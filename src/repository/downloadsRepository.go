package repository

import (
	"kard/src/model"
	"kard/src/model/dto"
	"time"

	"gorm.io/gorm"
)

type DownloadsRepository struct {
	IsEnable bool
	*gorm.DB
}

// 创建 DownloadsFactory
func DownloadsFactory() *DownloadsRepository {
	db := UseDbConn()
	isEnable := db != nil
	return &DownloadsRepository{IsEnable: isEnable, DB: db}
}

// func (repository *DownloadsRepository) Exists(taskDto *dto.TaskDto) bool {
// 	if !repository.IsEnable {
// 		return false
// 	}

// 	dl := new(model.Downloads)
// 	// repository.DB.Where("download_url=?", taskDto.DownloadUrl).Or("name=? and lan=?", taskDto.Name, taskDto.Lan).First(dl)
// 	repository.DB.Where("es_index=? and (download_url=? or name=?)", taskDto.EsIndex, taskDto.DownloadUrl, taskDto.Name).First(dl)
// 	return dl.Id > 0
// }

func (repository *DownloadsRepository) TryCreate(taskDto *dto.TaskDto) (bool, int32) {
	if !repository.IsEnable {
		return false, 0
	}

	createTime := time.Now()
	dl := &model.Downloads{
		EsIndex:       taskDto.EsIndex,
		BaseModel:     model.BaseModel{CreateTime: createTime},
		Name:          taskDto.Name,
		Crawler:       taskDto.Crawler,
		Lan:           taskDto.Lan,
		SubtitlesType: taskDto.SubtitlesType,
		InfoUrl:       taskDto.InfoUrl,
		UpdateTime:    createTime,
	}

	repository.DB.Where("es_index=? and (info_url=? and download_url=? or name=?)", taskDto.EsIndex, taskDto.InfoUrl, taskDto.DownloadUrl, taskDto.Name).FirstOrCreate(dl)

	isCreate := true
	if dl.CreateTime.Before(createTime) {
		isCreate = false
	}

	return isCreate, dl.Id
}
