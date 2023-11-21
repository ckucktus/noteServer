package main

import (
	"fmt"
	"os"
	"test_task/internal/application"
)

var appVersion = "v0.0.0" //nolint:gochecknoglobals

func main() {
	app := application.NewApp(appVersion)
	err := app.SetupLogger()
	if err != nil {
		fmt.Printf("SetupLogger %w", err)
	}
	err = app.SetupConfig()

	if err := app.Run(); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}
