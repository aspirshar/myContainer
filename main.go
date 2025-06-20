package main

import (
	"os"

	"github.com/aspirshar/myContainer/config"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const usage = `myContainer is a container runtime implementation.`

func main() {
	app := cli.NewApp()
	app.Name = "myContainer"
	app.Usage = usage

	app.Commands = []cli.Command{
		initCommand,
		RunCommand,
		commitCommand,
		listCommand,
		logCommand,
		execCommand,
		stopCommand,
		removeCommand,
		networkCommand,
	}

	app.Before = func(context *cli.Context) error {
		// 设置日志输出格式
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)

		// 初始化配置
		if err := config.Init(); err != nil {
			log.Errorf("Failed to initialize config: %v", err)
			return err
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}