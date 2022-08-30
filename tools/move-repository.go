package main

import (
	"context"
	"github.com/duke-git/lancet/v2/convertor"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
	"move-repository/pkg/common"
	"move-repository/pkg/handler"
	"runtime"
)

var configPath *string = flag.StringP("config", "c", "", "The path of config file")
var command *string = flag.StringP("command", "m", "", "The supported command: download, upload ")

func main() {
	flag.Parse()
	config, _ := common.SetupViper(*configPath)

	if len(*command) == 0 {
		panic("you should specify the command to run: handler or upload")
	}

	//log初始化
	logger := common.SetupLog(*config)
	defer logger.Sync()

	if json, err := convertor.ToJson(config); err == nil {
		logger.Info("the configuration parsed", zap.String("content", json))
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "config", config)
	ctx = context.WithValue(ctx, "logger", logger)

	runtime.GOMAXPROCS(5)

	switch *command {
	case "download":
		handler.Download(ctx)
	case "upload":
		handler.Upload(ctx)
	}
	select {}
}