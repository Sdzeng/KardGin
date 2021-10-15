package repository

import (
	"bytes"
	"fmt"
	"kard/src/model"
	"kard/src/model/dto"
	"strings"
	"time"

	"gorm.io/gorm"
)

type DownloadFileRepository struct {
	IsEnable bool
	*gorm.DB
}

// 创建 DownloadFileFactory
func DownloadFileFactory() *DownloadFileRepository {
	db := UseDbConn()
	isEnable := db != nil
	return &DownloadFileRepository{IsEnable: isEnable, DB: db}
}

func (repository *DownloadFileRepository) Save(dto *dto.TaskDto) error {
	if !repository.IsEnable {
		return nil
	}

	df := &model.Downloads{
		BaseModel:           model.BaseModel{CreateTime: time.Now()},
		Name:                dto.Name,
		DownloadUrl:         dto.DownloadUrl,
		DownloadUrlFileName: dto.DownloadUrlFileName,
		Lan:                 dto.Lan,
		SubtitlesType:       dto.SubtitlesType,
	}

	trans := repository.DB.Begin()
	result := trans.FirstOrCreate(df, model.Downloads{DownloadUrl: dto.DownloadUrl})

	if result.Error != nil {
		trans.Rollback()
		return result.Error
	}

	if result.RowsAffected <= 0 {
		trans.Commit()
		dto.DbNew = false
		return nil
	} else {
		dto.DbNew = true
		// fmt.Printf("\n数据库新加：%v", dto.Name)
	}

	if len(dto.SubtitlesFiles) > 0 {
		for _, subtitlesFile := range dto.SubtitlesFiles {
			downloadPath := &model.DownloadPaths{
				BaseModel:  model.BaseModel{CreateTime: df.CreateTime},
				DownloadId: df.Id,
				FileName:   subtitlesFile.FileName,
				FilePath:   subtitlesFile.FilePath,
			}

			result = trans.Create(downloadPath)
			if result.Error != nil {
				trans.Rollback()
				return result.Error
			}

			subtitlesFile.DownloadPathId = downloadPath.Id

			downloadPathSubtitlesSlice := []*model.DownloadPathSubtitles{}
			for _, subtitleItems := range subtitlesFile.SubtitleItems {

				var buffer bytes.Buffer
				for _, text := range subtitleItems.Text {
					buffer.WriteString(text + " ")
				}

				downloadPathSubtitles := &model.DownloadPathSubtitles{
					BaseModel:      model.BaseModel{CreateTime: df.CreateTime},
					DownloadPathId: downloadPath.Id,
					StartAt:        int32(subtitleItems.StartAt.Seconds()),
					Text:           buffer.String(),
				}

				downloadPathSubtitlesSlice = append(downloadPathSubtitlesSlice, downloadPathSubtitles)
			}

			result = trans.CreateInBatches(downloadPathSubtitlesSlice, len(downloadPathSubtitlesSlice))
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
