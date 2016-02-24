package main

import (
	"flag"
	"github.com/Jumpscale/agentcontroller8/application"
	"log"
	"os"
)

// Gets the settings path from a CLI argument
func getSettingsPath() string {

	var settingsPath string

	flag.StringVar(&settingsPath, "c", "", "Path to config file")
	flag.Parse()

	if settingsPath == "" {
		log.Println("Missing required option -c")
		flag.PrintDefaults()
		os.Exit(1)
	}

	return settingsPath
}

func main() {

	app := application.NewApplication(getSettingsPath())

	app.Run()
}
