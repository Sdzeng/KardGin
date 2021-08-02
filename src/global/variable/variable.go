package variable

import (
	"kard/src/global/interf"

	"github.com/olivere/elastic/v7"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	FecthPage string = "fecthPage"
	FecthList string = "fecthList"
	FecthInfo string = "fecthInfo"
	// FecthSelectDx1 string = "fecthSelectDx1"
	// FecthSource    string = "fecthSource"
	//Download       string = "download"
	ParseFile string = "parseFile"
)

var (
	BasePath           string       // 定义项目的根目录
	EventDestroyPrefix = "Destroy_" //  程序退出时需要销毁的事件前缀
	ConfigKeyPrefix    = "Config_"  //  配置文件键值缓存时，键的前缀

	// 全局日志指针
	ZapLog *zap.Logger

	// 全局配置文件
	WebYml  interf.YmlConfigInterf // 全局配置文件指针
	GormYml interf.YmlConfigInterf // 全局配置文件指针

	UseDbType string

	//gorm 数据库客户端，如果您操作数据库使用的是gorm，请取消以下注释，在 bootstrap>init 文件，进行初始化即可使用
	GormDbMysql      *gorm.DB        // 全局gorm的客户端连接
	GormDbSqlserver  *gorm.DB        // 全局gorm的客户端连接
	GormDbPostgreSql *gorm.DB        // 全局gorm的客户端连接
	ES               *elastic.Client //全局es客户端
)
