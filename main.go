/*
Copyright © 2021-2023 JANOS MIKO <info@janosmiko.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"os"
	"os/signal"
	"syscall"

	"reward/cmd/root"
	"reward/internal/app"
)

var (
	APPNAME = "reward"
	VERSION = "v0.4.0-beta-20221222-1900"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(
		sig,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	app := app.New(APPNAME, VERSION)

	go func() {
		<-sig
		if err := app.Cleanup(); err != nil {
			os.Exit(1)
		}

		os.Exit(0)
	}()

	root.NewRootCmd(app).Execute()
	app.Cleanup()
}
