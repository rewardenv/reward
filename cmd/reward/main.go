package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rewardenv/reward/cmd/root"
	"github.com/rewardenv/reward/internal/config"
)

var (
	APPNAME = "reward"
	VERSION = "v0.5.0-beta-20230925-1900"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(
		sig,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	app := config.New(APPNAME, VERSION)

	cobra.OnInitialize(func() {
		app.Init()
	})

	go func() {
		<-sig

		if err := app.Cleanup(); err != nil {
			os.Exit(1)
		}

		os.Exit(0)
	}()

	err := root.NewCmdRoot(app).Execute()
	if err != nil {
		log.Error(err)

		os.Exit(1)
	}
	_ = app.Cleanup()
}
