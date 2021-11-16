package repository

import (
	"fmt"

	"kard/src/global/variable"
	"kard/src/model"
	"kard/src/model/dto"
	"strings"
	"time"

	"gorm.io/gorm"
)

type DownloadPathsRepository struct {
	IsEnable bool
	*gorm.DB
}

// 创建 DownloadPathsFactory
func DownloadPathsFactory() *DownloadPathsRepository {
	db := UseDbConn()
	isEnable := db != nil
	return &DownloadPathsRepository{IsEnable: isEnable, DB: db}
}

func (repository *DownloadPathsRepository) Exists(db *gorm.DB, fileName, fileSum, esIndex string) bool {
	if !repository.IsEnable {
		return false
	}

	dp := new(model.DownloadPaths)
	// repository.DB.Where("download_url=?", taskDto.DownloadUrl).Or("name=? and lan=?", taskDto.Name, taskDto.Lan).First(dl)
	err := db.
		Table("download_paths").
		Select("download_paths.id").
		Joins("left join downloads ON download_paths.download_id = downloads.id").
		Where("downloads.es_index=? and (download_paths.file_sum=? or download_paths.file_name=?)", esIndex, fileSum, fileName).
		First(dp).
		Error
	if err != nil {
		fmt.Print(err)
		return false
	}
	return dp.Id > 0
}

func (repository *DownloadPathsRepository) Save(dto *dto.TaskDto) error {
	if !repository.IsEnable {
		return nil
	}

	now := time.Now()
	dl := &model.Downloads{
		BaseModel: model.BaseModel{Id: dto.DownloadId},
	}

	trans := repository.DB.Begin()

	// result := trans.Debug().FirstOrCreate(df, model.Downloads{DownloadUrl: dto.DownloadUrl})
	// result := trans.Where(model.Downloads{DownloadUrl: dto.DownloadUrl}).FirstOrCreate(df)

	result := trans.Model(dl).Select("download_url", "page", "update_time").Updates(model.Downloads{DownloadUrl: dto.DownloadUrl, Page: dto.PageNum, UpdateTime: now})

	if result.Error != nil {
		trans.Rollback()
		return result.Error
	}

	//result.RowsAffected <= 0 ||
	// if df.CreateTime.Before(createTime) {
	// 	trans.Commit()
	// 	dto.DbNew = false
	// 	return nil
	// } else {
	// 	dto.DbNew = true
	// }

	if len(dto.SubtitlesFiles) > 0 {
		for _, subtitlesFile := range dto.SubtitlesFiles {

			if len(subtitlesFile.SubtitleItems) <= 0 {
				variable.ZapLog.Sugar().Infof("跳过解析不到字幕的文件：%v---%v", dto.Name, subtitlesFile.FileName)
				continue
			}

			if repository.Exists(trans, subtitlesFile.FileName, subtitlesFile.FileSum, variable.IndexName) {
				subtitlesFile.DbNew = false
				variable.ZapLog.Sugar().Infof("跳过已存在文件：%v---%v", dto.Name, subtitlesFile.FileName)
				continue
			}

			downloadPath := &model.DownloadPaths{
				BaseModel:  model.BaseModel{CreateTime: now},
				DownloadId: dto.DownloadId,
				Name:       subtitlesFile.Name,
				FileName:   subtitlesFile.FileName,
				FilePath:   subtitlesFile.FilePath,
				FileSum:    subtitlesFile.FileSum,
			}

			result = trans.Create(downloadPath)
			if result.Error != nil {
				trans.Rollback()
				return result.Error
			}

			subtitlesFile.DownloadPathId = downloadPath.Id
			subtitlesFile.DbNew = true
			// downloadPathSubtitlesSlice := []*model.DownloadPathSubtitles{}
			// for _, subtitleItems := range subtitlesFile.SubtitleItems {

			// 	var buffer bytes.Buffer
			// 	for _, text := range subtitleItems.Text {
			// 		buffer.WriteString(text + " ")
			// 	}

			// 	downloadPathSubtitles := &model.DownloadPathSubtitles{
			// 		BaseModel:      model.BaseModel{CreateTime: df.CreateTime},
			// 		DownloadPathId: downloadPath.Id,
			// 		StartAt:        int32(subtitleItems.StartAt.Seconds()),
			// 		Text:           buffer.String(),
			// 	}

			// 	downloadPathSubtitlesSlice = append(downloadPathSubtitlesSlice, downloadPathSubtitles)
			// }

			// result = trans.CreateInBatches(downloadPathSubtitlesSlice, len(downloadPathSubtitlesSlice))
			// if result.Error != nil {
			// 	trans.Rollback()
			// 	return result.Error
			// }
		}
	} else {
		variable.ZapLog.Sugar().Warn("跳过下载不到的文件：%v---%v", dto.Name)
	}

	if len(dto.Refers) > 0 {
		valueStrings := make([]string, 0, len(dto.Refers))
		valueArgs := make([]interface{}, 0, len(dto.Refers)*4)
		sort := 0
		for _, refer := range dto.Refers {
			sort++
			valueStrings = append(valueStrings, "(?, ?, ?, ?)")
			valueArgs = append(valueArgs, dto.DownloadId)
			valueArgs = append(valueArgs, refer)
			valueArgs = append(valueArgs, sort)
			valueArgs = append(valueArgs, now)
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
