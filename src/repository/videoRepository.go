package repository

import (
	"kard/src/dto"
	"kard/src/global/variable"
	"time"

	"gorm.io/gorm"
)

type VideoRepository struct {
	*gorm.DB
}

// 创建 userFactory
// 参数说明： 传递空值，默认使用 配置文件选项：UseDbType（mysql）
func CreateVideoFactory() *VideoRepository {
	return &VideoRepository{DB: UseDbConn(variable.UseDbType)}
}

// 表名
func (u *VideoRepository) TableName() string {
	return "video"
}

func (u *VideoRepository) GetCover(today time.Time) (*dto.VideoDto, error) {
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
