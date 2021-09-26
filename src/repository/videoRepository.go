package repository

import (
	"kard/src/dto"
	"kard/src/global/variable"
	"time"

	"gorm.io/gorm"
)

type VideoRepository struct {
	IsEnable bool,
	*gorm.DB
}

// 创建 VideoFactory
// 参数说明： 传递空值，默认使用 配置文件选项：UseDbType（mysql）
func VideoFactory() *VideoRepository {
	db := UseDbConn(variable.UseDbType)
	isEnable := db != nil
	return &VideoRepository{IsEnable：isEnable,DB: db}
}

// 表名
func (u *VideoRepository) TableName() string {
	return "video"
}

func (u *VideoRepository) GetCover(today time.Time) (*dto.VideoDto, error) {
	if !u.IsEnable {
       return nil,nil
	}
	sql := `select VideoName,VideoCoverFragment,VideoCoverImg from video 
			where isHomeCover=1 and homeCoverDate <=? order by homeCoverDate desc limit 1 `

	temp := &dto.VideoDto{}
	result := u.Raw(sql, today).First(temp)
	if result.Error == nil {
		return temp, nil
	} else {
		return nil, result.Error
	}
}
