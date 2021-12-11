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

// func (repository *RazorsRepository) FirstOrCreate(razor, seedUrl, esIndex string, page int) *model.Razors {
// 	if !repository.IsEnable {
// 		return nil
// 	}
// 	now := time.Now()
// 	raz := &model.Razors{
// 		BaseModel:  model.BaseModel{CreateTime: now},
// 		Razor:      razor,
// 		EsIndex:    esIndex,
// 		Domain:     seedUrl,
// 		Page:       page,
// 		UpdateTime: now,
// 	}

// 	result := repository.DB.Where("razor=? and es_index=?", razor, esIndex).FirstOrCreate(raz)

// 	if result.Error != nil {
// 		variable.ZapLog.Sugar().Errorf("记录进度失败：%v", result)
// 		return nil
// 	}

// 	return raz
// }

func (repository *RazorsRepository) CreateOrUpdate(esIndex, razor, domain, pathUrl string, page int) bool {
	if !repository.IsEnable {
		return false
	}
	now := time.Now()
	raz := &model.Razors{}
	repository.DB.Where("razor=? and es_index=?", razor, esIndex).First(raz)

	var result *gorm.DB
	if raz.Id > 0 {
		result = repository.DB.Model(&model.Razors{BaseModel: model.BaseModel{Id: raz.Id}}).Select("domain", "path_url", "page", "update_time").Updates(model.Razors{Domain: domain, PathUrl: pathUrl, Page: page, UpdateTime: now})
	} else {
		raz.BaseModel = model.BaseModel{CreateTime: now}
		raz.Razor = razor
		raz.EsIndex = esIndex
		raz.Domain = domain
		raz.PathUrl = pathUrl
		raz.Page = page
		raz.UpdateTime = now
		result = repository.DB.Create(raz)
	}

	if result.Error != nil {
		return false
	}

	return true
}

func (repository *RazorsRepository) Update(razor, esIndex, domain, pathUrl string, page int) bool {
	if !repository.IsEnable {
		return false
	}

	createTime := time.Now()
	raz := model.Razors{Domain: domain, PathUrl: pathUrl, Page: page, UpdateTime: createTime}
	result := repository.DB.Model(&model.Razors{}).Where("razor=? and es_index=?", razor, esIndex).Select("domain", "path_url", "page", "update_time").Updates(raz)

	if result.Error != nil {
		json, _ := json.Marshal(raz)
		variable.ZapLog.Sugar().Errorf("更新进度失败%v err=%v", json, result.Error)
		return false
	}
	return true
}

func (repository *RazorsRepository) KFirst(razor, esIndex string) *model.Razors {
	if !repository.IsEnable {
		return nil
	}

	rz := new(model.Razors)
	// repository.DB.Where("download_url=?", taskDto.DownloadUrl).Or("name=? and lan=?", taskDto.Name, taskDto.Lan).First(dl)
	repository.DB.Where("razor=? and es_index=?", razor, esIndex).First(rz)
	return rz
}
