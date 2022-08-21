package cli

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	trelloLib "github.com/adlio/trello"
	"github.com/sitMCella/toggl-trello-kpi/configuration"
	"github.com/sitMCella/toggl-trello-kpi/grafana"
	"github.com/sitMCella/toggl-trello-kpi/storage"
	"github.com/sitMCella/toggl-trello-kpi/toggl"
	"github.com/sitMCella/toggl-trello-kpi/trello"
	"go.uber.org/zap"
)

// CommandLine struct defines the command line service.
type CommandLine struct {
	config configuration.Configuration
	logger *zap.Logger
}

// NewCommandLine creates a new CommandLine.
func NewCommandLine(config configuration.Configuration, logger *zap.Logger) CommandLine {
	commandLine := CommandLine{
		config: config,
		logger: logger,
	}
	return commandLine
}

// Execute runs a command based on the command line choice and parameters.
func (commandLine *CommandLine) Execute() {
	executionChoice := flag.Int("choice", 1, "Application execution choice")
	flag.Parse()

	switch choice := *executionChoice; choice {
	case 1:
		commandLine.downloadTogglTimeAsCsv(flag.Args())
	case 2:
		commandLine.downloadTrelloCardsAsCsv()
	case 3:
		commandLine.insertFromCsv(flag.Args())
	case 4:
		commandLine.storeTogglTime()
	case 5:
		commandLine.storeTrelloBoard()
	case 6:
		commandLine.downloadTableAsCsv(flag.Args())
	case 7:
		commandLine.updateFromCsv(flag.Args())
	case 8:
		commandLine.createGrafanaDashboard()
	default:
		commandLine.logger.Fatal("Couldn't find the application choice", zap.Int("Choice", choice))
	}
}

// downloadTogglTimeAsCsv downloads and stores the Toggl Time entries in a CSV file.
func (commandLine *CommandLine) downloadTogglTimeAsCsv(args []string) {
	fmt.Println("Execute: Download Toggl Time as CSV file.")
	if len(args) < 2 {
		commandLine.logger.Fatal("Provide the year, and the month as arguments")
	}
	year, err := strconv.Atoi(args[0])
	if err != nil {
		commandLine.logger.Fatal("Error converting the year argument to numeric value", zap.Error(err))
	}
	month, err := strconv.Atoi(args[1])
	if err != nil {
		commandLine.logger.Fatal("Error converting the month argument to numeric value", zap.Error(err))
	}
	togglClient := toggl.NewTogglClient(commandLine.config, commandLine.logger)
	togglTime, err := toggl.NewTogglTime(commandLine.logger, togglClient)
	if err != nil {
		commandLine.logger.Fatal("Error creating TogglTime", zap.Error(err))
	}

	startTime := time.Date(year, time.Month(month), 01, 0, 0, 0, 0, time.UTC)
	endTime := startTime.AddDate(0, 1, -1)

	err = togglTime.DownloadAsCsv(startTime, endTime)
	if err != nil {
		commandLine.logger.Fatal("Error retrieving and storing the time range from Toggl", zap.Error(err))
	}
}

// downloadTrelloCardsAsCsv downloads and stores the Trello Card entries in a CSV file.
func (commandLine *CommandLine) downloadTrelloCardsAsCsv() {
	fmt.Println("Execute: Download Trello cards as CSV file.")
	client := trelloLib.NewClient(commandLine.config.TrelloConfiguration.AppKey, commandLine.config.TrelloConfiguration.ApiToken)
	trelloClient := trello.NewTrelloClient(commandLine.config, commandLine.logger, client)
	trello, err := trello.NewTrello(commandLine.logger, trelloClient)
	if err != nil {
		commandLine.logger.Fatal("Error creating Trello", zap.Error(err))
	}
	err = trello.DownloadAsCsv()
	if err != nil {
		commandLine.logger.Fatal("Error retrieving and storing the cards from Trello", zap.Error(err))
	}
}

// insertFromCsv inserts the database entries for either the Toggl Time or the Trello Cards from a CSV file.
func (commandLine *CommandLine) insertFromCsv(args []string) {
	fmt.Println("Execute: Insert from CSV file.")
	if len(args) < 2 {
		commandLine.logger.Fatal("Provide the file name, and the database table name as arguments")
	}
	fileName := args[0]
	databaseTableName := args[1]
	if databaseTableName != "toggl_time" && databaseTableName != "trello_card" {
		commandLine.logger.Fatal("Provide the correct database table name as argument. Choose from 'toggl_time' and 'trello_card'.")
	}
	postgresqlConnection := initPostgresqlConnection(commandLine.config, commandLine.logger)
	defer func() {
		dberr := postgresqlConnection.Close()
		if dberr != nil {
			commandLine.logger.Fatal("Error closing the PostgreSQL connection", zap.Error(dberr))
		}
	}()
	insertFromCsv, err := storage.NewInsertFromCsv(commandLine.logger, postgresqlConnection.GetDb())
	if err != nil {
		commandLine.logger.Fatal("Error creating InsertFromCsv", zap.Error(err))
	}
	err = executeInsertFromCsv(insertFromCsv, fileName, databaseTableName)
	if err != nil {
		commandLine.logger.Fatal("Error inserting the CSV file entries into the database", zap.String("Databse table name", databaseTableName), zap.Error(err))
	}
}

func executeInsertFromCsv(insertFromCsv *storage.InsertFromCsv, fileName string, databaseTableName string) error {
	switch databaseTableName {
	case "toggl_time":
		return insertFromCsv.Insert(fileName, databaseTableName, toggl.TogglTimeEntry{})
	case "trello_card":
		return insertFromCsv.Insert(fileName, databaseTableName, trello.TrelloCardEntry{})
	}
	return nil
}

// storeTogglTime downloads and stores the Toggl Time entries in the database.
func (commandLine *CommandLine) storeTogglTime() {
	fmt.Println("Execute: Store Toggl Time.")
	postgresqlConnection := initPostgresqlConnection(commandLine.config, commandLine.logger)
	defer func() {
		dberr := postgresqlConnection.Close()
		if dberr != nil {
			commandLine.logger.Fatal("Error closing the PostgreSQL connection", zap.Error(dberr))
		}
	}()
	togglClient := toggl.NewTogglClient(commandLine.config, commandLine.logger)
	togglTime, err := toggl.NewTogglTimeWithDatabaseConnection(commandLine.logger, togglClient, postgresqlConnection.GetDb())
	if err != nil {
		commandLine.logger.Fatal("Error creating TogglTime", zap.Error(err))
	}

	startTime := time.Date(2021, 02, 01, 01, 00, 00, 0, time.UTC)
	endTime := time.Date(2021, 02, 06, 23, 59, 59, 999999999, time.UTC)

	err = togglTime.Store(startTime, endTime)
	if err != nil {
		commandLine.logger.Fatal("Error retrieving and storing the time range from Toggl", zap.Error(err))
	}
}

// storeTrelloBoard downloads and stores the Trello Card entries in the database.
func (commandLine *CommandLine) storeTrelloBoard() {
	fmt.Println("Execute: Store Trello Board.")
	postgresqlConnection := initPostgresqlConnection(commandLine.config, commandLine.logger)
	defer func() {
		dberr := postgresqlConnection.Close()
		if dberr != nil {
			commandLine.logger.Fatal("Error closing the PostgreSQL connection", zap.Error(dberr))
		}
	}()
	client := trelloLib.NewClient(commandLine.config.TrelloConfiguration.AppKey, commandLine.config.TrelloConfiguration.ApiToken)
	trelloClient := trello.NewTrelloClient(commandLine.config, commandLine.logger, client)
	trello, err := trello.NewTrelloWithDatabaseConnection(commandLine.logger, trelloClient, postgresqlConnection.GetDb())
	if err != nil {
		commandLine.logger.Fatal("Error creating Trello", zap.Error(err))
	}
	err = trello.Store()
	if err != nil {
		commandLine.logger.Fatal("Error retrieving and storing the cards from Trello", zap.Error(err))
	}
}

// downloadTableAsCsv downloads either the Toggl Time or the Trello Cards from the database to a CSV file.
func (commandLine *CommandLine) downloadTableAsCsv(args []string) {
	fmt.Println("Execute: Download table as CSV.")
	if len(args) == 0 {
		commandLine.logger.Fatal("Provide the database table name as argument")
	}
	databaseTableName := args[0]
	if databaseTableName != "toggl_time" && databaseTableName != "trello_card" {
		commandLine.logger.Fatal("Provide the correct database table name as argument. Choose from 'toggl_time' and 'trello_card'.")
	}
	postgresqlConnection := initPostgresqlConnection(commandLine.config, commandLine.logger)
	defer func() {
		dberr := postgresqlConnection.Close()
		if dberr != nil {
			commandLine.logger.Fatal("Error closing the PostgreSQL connection", zap.Error(dberr))
		}
	}()
	downloadAsCsv, err := storage.NewDownloadAsCsv(commandLine.logger, postgresqlConnection.GetDb())
	if err != nil {
		commandLine.logger.Fatal("Error creating NewDownloadAsCsv", zap.Error(err))
	}
	if len(args) == 1 {
		err := downloadAsCsv.DownloadAll(databaseTableName)
		if err != nil {
			commandLine.logger.Fatal("Cannot download the database table as CSV", zap.String("Databse table name", databaseTableName), zap.Error(err))
		}
	} else {
		columnsFilter := strings.Split(args[1], ",")
		err := downloadAsCsv.Download(databaseTableName, columnsFilter)
		if err != nil {
			commandLine.logger.Fatal("Cannot download the database table as CSV", zap.String("Databse table name", databaseTableName), zap.Error(err))
		}
	}
}

// updateFromCsv updates the database entries for the specified table from a CSV file.
func (commandLine *CommandLine) updateFromCsv(args []string) {
	fmt.Println("Execute: Update table from CSV.")
	if len(args) < 3 {
		commandLine.logger.Fatal("Provide the file name, the database table name, and the column name as arguments")
	}
	postgresqlConnection := initPostgresqlConnection(commandLine.config, commandLine.logger)
	defer func() {
		dberr := postgresqlConnection.Close()
		if dberr != nil {
			commandLine.logger.Fatal("Error closing the PostgreSQL connection", zap.Error(dberr))
		}
	}()
	fileName := args[0]
	updateFromCsv := storage.NewUpdateFromCsv(commandLine.logger, postgresqlConnection.GetDb())
	databaseTableName := args[1]
	columnName := args[2]
	err := updateFromCsv.Upload(fileName, databaseTableName, columnName)
	if err != nil {
		commandLine.logger.Fatal("Cannot update the database table from CSV", zap.String("File name", fileName), zap.String("Database table name", databaseTableName), zap.Error(err))
	}
}

func initPostgresqlConnection(config configuration.Configuration, logger *zap.Logger) (postgresqlConnection storage.PostgresqlConnection) {
	postgresqlConnection, err := storage.NewPostgresConnection(config.DBConfiguration)
	if err != nil {
		logger.Fatal("Couldn't connect to the database", zap.Error(err))
	}
	err = postgresqlConnection.InitDatabase()
	if err != nil {
		logger.Fatal("Couldn't initialize the database", zap.Error(err))
	}
	return
}

// createGrafanaDashboard creates the Grafana json definition from the application configuration.
func (commandLine *CommandLine) createGrafanaDashboard() {
	fmt.Println("Execute: Create the Grafana Dashboard.")
	grafanaDashboard, err := grafana.NewGrafanaDashboard(commandLine.config, commandLine.logger)
	if err != nil {
		commandLine.logger.Fatal("Error creating GrafanaDashboard", zap.Error(err))
	}
	err = grafanaDashboard.CreateDashboard()
	if err != nil {
		commandLine.logger.Fatal("Cannot create the Grafana Dashboard", zap.Error(err))
	}
}
