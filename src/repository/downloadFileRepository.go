package repository

import (
	"fmt"
	"kard/src/dto"
	"kard/src/model"
	"strings"
	"time"
)

type DownloadFileRepository struct {
	IsEnable bool,
	*gorm.DB
}

// 创建 DownloadFileFactory
// 参数说明： 传递空值，默认使用 配置文件选项：UseDbType（mysql）
func DownloadFileFactory() *DownloadFileRepository {
	db := UseDbConn(variable.UseDbType)
	isEnable := db != nil
	return &DownloadFileRepository{IsEnable：isEnable,DB: db}
}


func (repository *DownloadFileRepository) Save(dto *dto.UrlDto) error {
	if !u.IsEnable {
		return nil,nil
	 }

	df := &model.Downloads{
		BaseModel:   model.BaseModel{CreateTime: time.Now().Unix()},
		Name:        dto.Name,
		DownloadUrl: dto.DownloadUrl,
		FileName:    dto.FileName,
		Lan:         dto.Lan,
		Subtitles:   dto.Subtitles,
	}

	trans := Db.Begin()
	result := trans.FirstOrCreate(df, model.Downloads{DownloadUrl: dto.DownloadUrl})

	if result.Error != nil {
		trans.Rollback()
		return result.Error
	}

	if result.RowsAffected <= 0 {
		trans.Commit()
		return nil
	} else {
		fmt.Printf("\n数据库新加：%v", dto.Name)
	}

	if len(dto.FilePaths) > 0 {
		for _, filePath := range dto.FilePaths {
			downloadPath := &model.DownloadPaths{
				BaseModel:  model.BaseModel{CreateTime: df.CreateTime},
				DownloadId: df.Id,
				FilePath:   filePath,
			}

			result = trans.Create(downloadPath)
			if result.Error != nil {
				trans.Rollback()
				return result.Error
			}
		}
	}

	if len(dto.Refers) > 0 {
		valueStrings := make([]string, 0, len(dto.Refers))
		valueArgs := make([]interface{}, 0, len(dto.Refers)*4)
		sort := 0
		for _, refer := range dto.Refers {
			sort++
			valueStrings = append(valueStrings, "(?, ?, ?, ?)")
			valueArgs = append(valueArgs, df.Id)
			valueArgs = append(valueArgs, refer)
			valueArgs = append(valueArgs, sort)
			valueArgs = append(valueArgs, df.CreateTime)
		}
		stmt := fmt.Sprintf("INSERT INTO `kard_gin`.`download_refers` ( `download_id`, `refer`,`sort`, `create_time`) VALUES %s", strings.Join(valueStrings, ","))
		result = trans.Exec(stmt, valueArgs...)
		if result.Error != nil {
			trans.Rollback()
			return result.Error
		}
	}

	trans.Commit()
	return nil
}
