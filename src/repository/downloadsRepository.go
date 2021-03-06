package repository

import (
	"encoding/json"
	"kard/src/global/variable"
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

func (repository *DownloadsRepository) KFirst(downloadId int32) *model.Downloads {
	if !repository.IsEnable {
		return nil
	}

	dl := new(model.Downloads)
	// repository.DB.Where("download_url=?", taskDto.DownloadUrl).Or("name=? and lan=?", taskDto.Name, taskDto.Lan).First(dl)
	repository.DB.First(dl, downloadId)
	return dl
}

func (repository *DownloadsRepository) TryCreate(esIndex, razor string, taskDto *dto.TaskDto) (bool, int32, error) {
	if !repository.IsEnable {
		return false, 0, nil
	}

	isCreate := true
	now := time.Now()
	dl := &model.Downloads{}
	err := repository.DB.Where("es_index=? and (info_url=? or download_url=? or name=?) ", esIndex, taskDto.InfoUrl, taskDto.DownloadUrl, taskDto.Name).First(dl).Error

	if err != nil {
		return false, 0, err
	}

	if dl.Id <= 0 {
		dl = &model.Downloads{
			EsIndex:       esIndex,
			BaseModel:     model.BaseModel{CreateTime: now},
			Name:          taskDto.Name,
			Page:          taskDto.PageNum,
			Razor:         razor,
			Lan:           taskDto.Lan,
			SubtitlesType: taskDto.SubtitlesType,
			InfoUrl:       taskDto.InfoUrl,
			UpdateTime:    now,
		}

		err = repository.DB.Create(dl).Error
		isCreate = true
	} else {
		if len(dl.DownloadUrl) <= 0 {
			dl.Name = taskDto.Name
			dl.Page = taskDto.PageNum
			dl.Razor = razor
			dl.Lan = taskDto.Lan
			dl.SubtitlesType = taskDto.SubtitlesType
			dl.InfoUrl = taskDto.InfoUrl
			dl.CreateTime = now
			dl.UpdateTime = now
			err = repository.DB.Model(&model.Downloads{BaseModel: model.BaseModel{Id: dl.Id}}).Updates(dl).Error
			isCreate = true
		} else {
			err = nil
			isCreate = false
		}
	}

	return isCreate, dl.Id, err
}

func (repository *DownloadsRepository) Update(id int32, name string) bool {
	if !repository.IsEnable {
		return false
	}

	now := time.Now()
	info := model.Downloads{Name: name, UpdateTime: now}
	result := repository.DB.Model(&model.Downloads{BaseModel: model.BaseModel{Id: id}}).Updates(info)

	if result.Error != nil {
		json, _ := json.Marshal(info)
		variable.ZapLog.Sugar().Errorf("更新名称失败%v err=%v", json, result.Error)
		return false
	}
	return true
}
