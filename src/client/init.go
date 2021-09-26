package client

import (
	"kard/src/global/config"
	"kard/src/global/kardError"
	"kard/src/global/kardLog"
	"kard/src/global/variable"
	"kard/src/gorm"
	"log"
	"os"
	"strings"

	"github.com/olivere/elastic/v7"
)

func init() {
	// 1.初始化程序根目录
	os.Chdir("../../")
	if path, err := os.Getwd(); err == nil {
		// 路径进行处理，兼容单元测试程序程序启动时的奇怪路径
		if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "-test") {
			variable.BasePath = strings.Replace(strings.Replace(path, `\test`, "", 1), `/test`, "", 1)
		} else {
			variable.BasePath = path
		}
	} else {
		log.Fatal(kardError.ErrorsBasePath)
	}

	//2.检查配置文件以及日志目录等非编译性的必要条件
	checkRequiredFolders()

	// 4.启动针对配置文件(confgi.yml、gorm_v2.yml)变化的监听， 配置文件操作指针，初始化为全局变量
	variable.WebYml = config.CreateYamlFactory()
	variable.WebYml.ConfigFileChangeListen()
	// config>gorm.yml 启动文件变化监听事件
	variable.GormYml = variable.WebYml.Clone("gorm")
	variable.GormYml.ConfigFileChangeListen()

	//5. 初始化ES
	esEnable := variable.WebYml.GetBool("ElasticSearch.Enable")
	if esEnable {
		esUrl := variable.WebYml.GetString("ElasticSearch.Url")
		ps := variable.ES.Ping(esUrl)
		if ps == nil {
			log.Fatal("初始化es客户端ping失败")
		}

		userName := variable.WebYml.GetString("ElasticSearch.UserName")
		password := variable.WebYml.GetString("ElasticSearch.Password")
		es, err := elastic.NewClient(elastic.SetURL(esUrl), elastic.SetSniff(false), elastic.SetBasicAuth(userName, password))

		if err != nil {
			log.Fatal("初始化es客户端连接失败" + err.Error())
		}
		variable.ES = es
	}
	// 6.初始化全局日志句柄，并载入日志钩子处理函数
	variable.ZapLog = kardLog.CreateZapFactory(kardLog.ZapLogHandler)

	variable.UseDbType = variable.GormYml.GetString("Gormv2.UseDbType")

	// 7.根据配置初始化 gorm mysql 全局 *gorm.Db
	if variable.GormYml.GetInt("Gormv2.Mysql.IsInitGolobalGormMysql") == 1 {
		if dbMysql, err := gorm.GetOneMysqlClient(); err != nil {
			log.Fatal(kardError.ErrorsGormInitFail + err.Error())
		} else {
			variable.GormDbMysql = dbMysql
		}
	}

	// 8.根据配置初始化 gorm sqlserver 全局 *gorm.Db
	if variable.GormYml.GetInt("Gormv2.Sqlserver.IsInitGolobalGormSqlserver") == 1 {
		if dbSqlserver, err := gorm.GetOneSqlserverClient(); err != nil {
			log.Fatal(kardError.ErrorsGormInitFail + err.Error())
		} else {
			variable.GormDbSqlserver = dbSqlserver
		}
	}
	// 9.根据配置初始化 gorm postgresql 全局 *gorm.Db
	if variable.GormYml.GetInt("Gormv2.PostgreSql.IsInitGolobalGormPostgreSql") == 1 {
		if dbPostgre, err := gorm.GetOnePostgreSqlClient(); err != nil {
			log.Fatal(kardError.ErrorsGormInitFail + err.Error())
		} else {
			variable.GormDbPostgreSql = dbPostgre
		}
	}

}

// 检查项目必须的非编译目录是否存在，避免编译后调用的时候缺失相关目录
func checkRequiredFolders() {
	//1.检查配置文件是否存在
	if _, err := os.Stat(variable.BasePath + "/config/web.yml"); err != nil {
		log.Fatal(kardError.ErrorsConfigYamlNotExists + err.Error())
	}
	if _, err := os.Stat(variable.BasePath + "/config/gorm.yml"); err != nil {
		log.Fatal(kardError.ErrorsConfigGormNotExists + err.Error())
	}
	//2.检查public目录是否存在
	if _, err := os.Stat(variable.BasePath + "/client/web/wwwroot"); err != nil {
		log.Fatal(kardError.ErrorsPublicNotExists + err.Error())
	}
	//3.检查storage/logs 目录是否存在
	// if _, err := os.Stat(variable.BasePath + "/storage/logs/"); err != nil {
	// 	log.Fatal(kardError.ErrorsStorageLogsNotExists + err.Error())
	// }
	// 4.自动创建软连接、更好的管理静态资源
	// if _, err := os.Stat(variable.BasePath + "/public/storage"); err == nil {
	// 	if err = os.Remove(variable.BasePath + "/public/storage"); err != nil {
	// 		log.Fatal(kardError.ErrorsSoftLinkDeleteFail + err.Error())
	// 	}
	// }

}
