package main

import (
	"errors"
	"fmt"
	"net/http"
	"promotions/controllers"
	"promotions/packages/connection"
	"promotions/packages/tools"
	_ "promotions/routers"

	"github.com/astaxie/beego"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func init() {
	//配置文件处理
	err := initConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	tools.InitCode()
}

func main() {
	//链接到数据库
	dbClient, err := connection.GetDbConnect()
	if err != nil {
		fmt.Println(err)
		return
	}

	//连接redis
	redisClient, err := connection.GetRedisClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	//错误处理
	beego.ErrorController(&controllers.ErrorController{})
	beego.RunWithMiddleWares(":"+tools.AppConfig.HttpPort, MiddlewareHandler)
	defer func() {
		_ = dbClient.Close()
		_ = redisClient.Close()
	}()
}

/**
初始化配置文件
*/
func initConfig() (err error) {
	runMode := beego.AppConfig.String("runmode")
	httpPort := beego.AppConfig.String("httpport")
	limitValue := beego.AppConfig.DefaultInt64("limitValue", 60)
	if len(runMode) == 0 || len(httpPort) == 0 {
		return errors.New("请配置运行环境或运行端口！")
	}
	err = beego.LoadAppConfig("ini", "conf/"+runMode+"_app.conf")
	if err != nil {
		return err
	}
	beego.BConfig.RecoverPanic = false
	beego.BConfig.EnableErrorsRender = false
	tools.AppConfig = &tools.Config{
		HttpPort:   httpPort,
		RunMode:    runMode,
		LimitValue: limitValue,
	}

	return nil
}

/**
中间件处理
*/
func MiddlewareHandler(httpHandler http.Handler) http.Handler {
	return tools.NewMiddlewareHandler(httpHandler)
}
