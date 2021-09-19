package trello

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/sitMCella/toggl-trello-kpi/application_errors"
	"github.com/sitMCella/toggl-trello-kpi/storage"
	"go.uber.org/zap"
)

// TrelloClient interface defines the Trello client primitives.
type Client interface {
	GetCards() ([]TrelloCardEntry, error)
}

// Trello struct defines the Trello service.
type Trello struct {
	logger             *zap.Logger
	trelloClient       Client
	databaseConnection *sql.DB
}

// TrelloCardEntry struct defines the Trello card entry.
type TrelloCardEntry struct {
	Id       string
	Name     string
	Closed   bool
	Labels   []string
	Project  string
	Customer string
	Team     string
	Type     string
}

// EmptyTrelloCardsError defines the empty Trello cards error.
type EmptyTrelloCardsError struct {
}

func (error *EmptyTrelloCardsError) Error() string {
	return fmt.Sprintf("The Trello card entries are empty.")
}

// NewTrello creates a new Trello.
func NewTrello(logger *zap.Logger, trelloClient Client) (*Trello, error) {
	if logger == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "logger"}
	}
	if trelloClient == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "trelloClient"}
	}
	return &Trello{
		logger:             logger,
		trelloClient:       trelloClient,
		databaseConnection: nil,
	}, nil
}

// NewTrello creates a new Trello with a Database Connection.
func NewTrelloWithDatabaseConnection(logger *zap.Logger, trelloClient Client, databaseConnection *sql.DB) (*Trello, error) {
	if logger == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "logger"}
	}
	if trelloClient == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "trelloClient"}
	}
	if databaseConnection == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "databaseConnection"}
	}
	return &Trello{
		logger:             logger,
		trelloClient:       trelloClient,
		databaseConnection: databaseConnection,
	}, nil
}

// DownloadAsCsv downloads the Trello card entries as CSV file.
func (trello *Trello) DownloadAsCsv() (err error) {
	trelloCardEntries, err := trello.trelloClient.GetCards()
	if err != nil {
		return
	}
	if len(trelloCardEntries) == 0 {
		trello.logger.Error("Skip the creation of the Trello card entries file.")
		return &EmptyTrelloCardsError{}
	}
	downloadStructAsCsv, err := storage.NewDownloadStructAsCsv(trello.logger)
	if err != nil {
		return
	}
	values := make([]interface{}, len(trelloCardEntries))
	trello.logger.Info("Trello card entries", zap.Int("count", len(trelloCardEntries)))
	for i, value := range trelloCardEntries {
		values[i] = value
	}
	return downloadStructAsCsv.DownloadAll(values, "trello_entries")
}

// Store inserts the Trello card entries into the database.
func (trello *Trello) Store() (err error) {
	trelloCardEntries, err := trello.trelloClient.GetCards()
	if err != nil {
		return
	}
	if len(trelloCardEntries) == 0 {
		trello.logger.Error("Skip the creation of the Trello card entries file.")
		return &EmptyTrelloCardsError{}
	}
	for _, trelloCardEntry := range trelloCardEntries {
		//fmt.Printf("%+v", trelloCardEntry)
		err = trello.storeInDatabase(trelloCardEntry)
		if err != nil {
			return
		}
	}
	return
}

func (trello *Trello) storeInDatabase(trelloCardEntry TrelloCardEntry) (err error) {
	if trello.databaseConnection == nil {
		err = &application_errors.DatabaseConnectionError{}
		return
	}
	tx, err := trello.databaseConnection.Begin()
	if err != nil {
		return
	}
	defer func() {
		switch err {
		case nil:
			sqlerr := tx.Commit()
			if err == nil {
				err = sqlerr
			}
		default:
			sqlerr := tx.Rollback()
			if err == nil {
				err = sqlerr
			}
		}
	}()
	sqlStmt := `INSERT INTO trello_card(id, name, closed, labels, project, customer, team, type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = trello.databaseConnection.Exec(
		sqlStmt,
		trelloCardEntry.Id, trelloCardEntry.Name, trelloCardEntry.Closed, pq.Array(trelloCardEntry.Labels),
		trelloCardEntry.Project, trelloCardEntry.Customer, trelloCardEntry.Team, trelloCardEntry.Type)
	if err != nil {
		return
	}
	return
}

func contains(labels []string, value string) bool {
	for _, label := range labels {
		if label == value {
			return true
		}
	}
	return false
}
