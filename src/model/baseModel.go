package model

// type BaseModel struct {
// 	Id           int32 `gorm:"primarykey" json:"id"`
// 	CreateTime   string `json:"creationTime"` //日期时间字段统一设置为字符串即可
// 	CreateUserId int32  `json:"creator"`
// }

type BaseModel struct {
	Id         int32 `json:"id" gorm:"column:id;primaryKey;"`
	CreateTime int64 `json:"create_time" gorm:"column:create_time"`
}
