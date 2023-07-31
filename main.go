package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/mengbin92/browser/conf"
	"github.com/mengbin92/browser/log"
	"github.com/mengbin92/browser/service"
)

var (
	flagconf string
	id, _    = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
}

func main() {
	flag.Parse()

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()
	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	logger := log.DefaultLogger(bc.Log)
	service.NewServer(bc.Server, logger.Sugar())

	if err := service.Run(); err != nil {
		panic(err)
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	service.Shutdown()
	logger.Info("Shutdown server")
}
