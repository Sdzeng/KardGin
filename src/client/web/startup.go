package main

import (
	"io"
	"kard/src/client/web/controller"
	"kard/src/global/variable"
	"net/http"
	"os"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	_ "kard/src/client"
)

func main() {
	var router *gin.Engine
	// 非调试模式（生产模式） 日志写到日志文件
	if !variable.WebYml.GetBool("AppDebug") {
		//1.将日志写入日志文件
		gin.DisableConsoleColor()
		f, _ := os.Create(variable.BasePath + variable.WebYml.GetString("Logs.GinLogName"))
		gin.DefaultWriter = io.MultiWriter(f)
		// 2.如果是有nginx前置做代理，基本不需要gin框架记录访问日志，开启下面一行代码，屏蔽上面的三行代码，性能提升 5%
		//gin.SetMode(gin.ReleaseMode)

		router = gin.Default()
	} else {
		// 调试模式，开启 pprof 包，便于开发阶段分析程序性能
		router = gin.Default()
		pprof.Register(router)
	}

	//根据配置进行设置跨域
	if variable.WebYml.GetBool("HttpServer.AllowCrossDomain") {
		router.Use(cors())
	}

	runKard(router)

}

// 允许跨域
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Access-Control-Allow-Headers,Authorization,User-Agent, Keep-Alive, Content-Type, X-Requested-With,X-CSRF-Token,AccessToken,Token")
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, PATCH, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		// 放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusAccepted)
		}
		c.Next()
	}
}

func runKard(router *gin.Engine) {

	//处理静态资源
	router.Static("/assert", variable.BasePath+"/web/wwwroot") //  定义静态资源路由与实际目录映射关系

	homeController := &controller.HomeController{}
	homeGroup := router.Group("/home")
	{
		// homeGroup.GET("/", homeController.GetCover)
		// homeGroup.GET("/subtitles", homeController.ExtractSubtitles)
		homeGroup.POST("/search", homeController.Search)
		homeGroup.POST("/scroll_search", homeController.SearchScroll)
	}

	router.Run(variable.WebYml.GetString("HttpServer.Api.Port"))

}
