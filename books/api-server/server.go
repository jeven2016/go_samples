package main

import (
	"api/common"
	"api/initialization"
	"context"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
)

var configPath *string = flag.StringP("config", "c", "", "The path of defined config file")

func main() {
	// 解析传入的参数, 运行时可以添加 -h查看细节
	flag.Parse()

	//读取配置文件
	config, err := initialization.SetupViper(*configPath)
	if err != nil {
		panic(fmt.Sprintf("Program exited, %v", err))
	}

	app := new(common.App)

	//log初始化
	logger := initialization.SetupLog(*config)
	defer logger.Sync()

	//设置logger
	app.Log = logger

	if json, err := convertor.ToJson(config); err == nil {
		app.Log.Info("the configuration parsed", zap.String("content", json))
	}

	//初始化Mongodb
	client, db, err := initialization.SetupMongodb(config, app.Log)
	if err == nil {
		app.Log.Info("Connecting mongoDB successfully")
		defer func() {
			// 延迟释放连接
			if err = client.Disconnect(context.TODO()); err != nil {
				app.Log.Error("The mongodb client cannot disconnect.", zap.Error(err))
			}
		}()
		//设置DB
		app.Db = db
	}

	//web初始化
	engine := initialization.SetupEngine(config, app)

	bind := fmt.Sprintf("%v:%v", config.ApiServerConfig.BindAddress, config.ApiServerConfig.Port)
	if err := engine.Run(bind); err != nil {
		panic(errors.New(fmt.Sprintf("failed to start: %v", err)))
	} else {
		defer app.Log.Info("Service is stopped")
	}
}