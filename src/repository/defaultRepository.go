package repository

import (
	"errors"
	"fmt"
	"kard/src/global/kardError"
	"kard/src/global/variable"
	"kard/src/model"
	"strings"

	"gorm.io/gorm"
)

func UseDbConn() *gorm.DB {
	return UseDbConnWithType("")
}

func UseDbConnWithType(sqlType string) *gorm.DB {
	var db *gorm.DB
	sqlType = strings.Trim(sqlType, " ")
	if sqlType == "" {
		sqlType = variable.UseDbType // variable.GormYml.GetString("Gormv2.UseDbType")
	}
	switch strings.ToLower(sqlType) {
	case "mysql":
		if variable.GormDbMysql == nil {
			variable.ZapLog.Warn(fmt.Sprintf(kardError.ErrorsGormNotInitGlobalPointer, sqlType, sqlType))
		} else {
			db = variable.GormDbMysql
		}
	case "sqlserver":
		if variable.GormDbSqlserver == nil {
			variable.ZapLog.Warn(fmt.Sprintf(kardError.ErrorsGormNotInitGlobalPointer, sqlType, sqlType))
		} else {
			db = variable.GormDbSqlserver
		}
	case "postgres", "postgre", "postgresql":
		if variable.GormDbPostgreSql == nil {
			variable.ZapLog.Warn(fmt.Sprintf(kardError.ErrorsGormNotInitGlobalPointer, sqlType, sqlType))
		} else {
			db = variable.GormDbPostgreSql
		}
	default:
		variable.ZapLog.Error(kardError.ErrorsDbDriverNotExists + sqlType)
	}
	return db
}

func Insert(inter interface{}) (id int32, err error) {
	db := UseDbConn()
	result := db.Create(inter)
	if result.RowsAffected <= 0 {
		return 0, result.Error
	}

	t, ok := inter.(model.BaseModel)
	if !ok {
		return 0, errors.New("未组合repository.Table类型")
	}
	return t.Id, nil
}
