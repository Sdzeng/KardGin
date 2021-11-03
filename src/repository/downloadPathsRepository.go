package repository

import (
	"fmt"

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

	createTime := time.Now()
	df := &model.Downloads{
		BaseModel:   model.BaseModel{CreateTime: createTime},
		Name:        dto.Name,
		Crawler:     dto.Crawler,
		DownloadUrl: dto.DownloadUrl,
		// DownloadUrlFileName: dto.DownloadUrlFileName,
		Lan:           dto.Lan,
		SubtitlesType: dto.SubtitlesType,
		EsIndex:       dto.EsIndex,
	}

	trans := repository.DB.Begin()

	// result := trans.Debug().FirstOrCreate(df, model.Downloads{DownloadUrl: dto.DownloadUrl})
	// result := trans.Where(model.Downloads{DownloadUrl: dto.DownloadUrl}).FirstOrCreate(df)
	result := trans.Where("es_index=? and (download_url=? or name=?)", dto.EsIndex, dto.DownloadUrl, dto.Name).FirstOrCreate(df)
	// result := trans.Debug().Where("download_url=?", dto.DownloadUrl).FirstOrCreate(df)
	// result := trans.Debug().Where("name=? and lan=?", dto.Name, dto.Lan).FirstOrCreate(df)

	if result.Error != nil {
		trans.Rollback()
		return result.Error
	}

	//result.RowsAffected <= 0 ||
	if df.CreateTime.Before(createTime) {
		trans.Commit()
		dto.DbNew = false
		return nil
	} else {
		dto.DbNew = true
		// fmt.Printf("\n数据库新加：%v", dto.Name)
	}

	if len(dto.SubtitlesFiles) > 0 {
		for _, subtitlesFile := range dto.SubtitlesFiles {

			if len(subtitlesFile.SubtitleItems) <= 0 {
				continue
			}

			if repository.Exists(trans, subtitlesFile.FileName, subtitlesFile.FileSum, dto.EsIndex) {
				subtitlesFile.DbNew = false
				continue
			}

			downloadPath := &model.DownloadPaths{
				BaseModel:  model.BaseModel{CreateTime: df.CreateTime},
				DownloadId: df.Id,
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
