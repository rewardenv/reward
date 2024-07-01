//go:build go1.22

package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rewardenv/reward/cmd/root"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/globals"
)

var (
	APPNAME = "reward"
	VERSION = "0.0.1"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(
		sig,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	globals.InitGlobals(APPNAME, VERSION)

	app := config.New(globals.APPNAME, globals.VERSION)

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

	if err := root.NewCmdRoot(app).Execute(); err != nil {
		log.Error(err)

		os.Exit(1)
	}
	_ = app.Cleanup()
}
