package repository

import (
	"encoding/json"
	"kard/src/global/variable"
	"kard/src/model"
	"time"

	"gorm.io/gorm"
)

type RazorsRepository struct {
	IsEnable bool
	*gorm.DB
}

// 创建 RazorsFactory
func RazorsFactory() *RazorsRepository {
	db := UseDbConn()
	isEnable := db != nil
	return &RazorsRepository{IsEnable: isEnable, DB: db}
}

func (repository *RazorsRepository) FirstOrCreate(razor, seedUrl, esIndex string, page int) *model.Razors {
	if !repository.IsEnable {
		return nil
	}
	now := time.Now()
	raz := &model.Razors{
		BaseModel:  model.BaseModel{CreateTime: now},
		Razor:      razor,
		EsIndex:    esIndex,
		SeedUrl:    seedUrl,
		Page:       page,
		UpdateTime: now,
	}

	result := repository.DB.Where("razor=? and es_index=?", razor, esIndex).FirstOrCreate(raz)

	if result.Error != nil {
		variable.ZapLog.Sugar().Errorf("记录进度失败：%v", result)
		return nil
	}

	return raz
}

func (repository *RazorsRepository) CreateOrUpdate(razor, seedUrl, esIndex string, page int) bool {
	if !repository.IsEnable {
		return false
	}
	now := time.Now()
	raz := &model.Razors{}
	repository.DB.Where("razor=? and es_index=?", razor, esIndex).First(raz)

	var result *gorm.DB
	if raz.Id > 0 {
		result = repository.DB.Model(&model.Razors{BaseModel: model.BaseModel{Id: raz.Id}}).Select("seed_url", "page", "update_time").Updates(model.Razors{SeedUrl: seedUrl, Page: page, UpdateTime: now})
	} else {
		raz.BaseModel = model.BaseModel{CreateTime: now}
		raz.Razor = razor
		raz.EsIndex = esIndex
		raz.SeedUrl = seedUrl
		raz.Page = page
		raz.UpdateTime = now
		result = repository.DB.Create(raz)
	}

	if result.Error != nil {
		return false
	}

	return true
}

func (repository *RazorsRepository) Update(razor, seedUrl, esIndex string, page int) bool {
	if !repository.IsEnable {
		return false
	}

	createTime := time.Now()
	raz := model.Razors{SeedUrl: seedUrl, Page: page, UpdateTime: createTime}
	result := repository.DB.Model(&model.Razors{}).Where("razor=? and es_index=?", razor, esIndex).Select("seed_url", "page", "update_time").Updates(raz)

	if result.Error != nil {
		json, _ := json.Marshal(raz)
		variable.ZapLog.Sugar().Errorf("更新进度失败%v err=%v", json, result.Error)
		return false
	}
	return true
}
