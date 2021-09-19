package main

import (
	"log"

	"github.com/sitMCella/toggl-trello-kpi/cli"
	"github.com/sitMCella/toggl-trello-kpi/configuration"
	"github.com/sitMCella/toggl-trello-kpi/logger"
)

func main() {
	config, err := configuration.NewConfiguration()
	if err != nil {
		log.Fatalf("Couldn't initialize Configuration: %+v", err)
	}

	logger, err := logger.NewLogger(config.ApplicationConfiguration.LogLevel)
	if err != nil {
		log.Fatalf("Couldn't initialize logger: %+v", err)
	}

	commandLine := cli.NewCommandLine(config, logger)
	commandLine.Execute()
}
