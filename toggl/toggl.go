package toggl

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/sitMCella/toggl-trello-kpi/application_errors"
	"github.com/sitMCella/toggl-trello-kpi/storage"
	"go.uber.org/zap"
)

// Client interface defines the Toggl client primitives.
type Client interface {
	GetRange(startTime time.Time, endTime time.Time) ([]TogglTimeEntry, error)
}

// TogglTime struct defines the Toggl service.
type TogglTime struct {
	logger             *zap.Logger
	togglClient        Client
	databaseConnection *sql.DB
}

// TogglTimeEntry struct defines the Toggl time entry.
type TogglTimeEntry struct {
	Id             uint64
	Description    string
	Start          time.Time
	Stop           time.Time
	Duration       int64
	Billable       bool
	Workspace_id   uint64
	Project_id     uint64
	Tags           []string
	Trello_card_id string
}

func (togglTimeEntry TogglTimeEntry) IsPrintable() bool {
	return true
}

// EmptyTimeResultError defines the empty time result error.
type EmptyTimeResultError struct {
}

func (error *EmptyTimeResultError) Error() string {
	return fmt.Sprintf("The Toggl time entries are empty.")
}

// NewTogglTime creates a new TogglTime.
func NewTogglTime(logger *zap.Logger, togglClient Client) (*TogglTime, error) {
	if logger == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "logger"}
	}
	if togglClient == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "togglClient"}
	}
	return &TogglTime{
		logger:             logger,
		togglClient:        togglClient,
		databaseConnection: nil,
	}, nil
}

// NewTogglTime creates a new TogglTime with a Database Connection.
func NewTogglTimeWithDatabaseConnection(logger *zap.Logger, togglClient Client, databaseConnection *sql.DB) (*TogglTime, error) {
	if logger == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "logger"}
	}
	if togglClient == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "togglClient"}
	}
	if databaseConnection == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "databaseConnection"}
	}
	return &TogglTime{
		logger:             logger,
		togglClient:        togglClient,
		databaseConnection: databaseConnection,
	}, nil
}

// DownloadAsCsv downloads the Toggl time entries as CSV file.
func (togglTime *TogglTime) DownloadAsCsv(startTime time.Time, endTime time.Time) (err error) {
	togglTimeEntries, err := togglTime.retrieve(startTime, endTime)
	if err != nil {
		return
	}
	if len(togglTimeEntries) == 0 {
		togglTime.logger.Error("Skip the creation of the Toggl time entries file.")
		return &EmptyTimeResultError{}
	}
	downloadStructAsCsv, err := storage.NewDownloadStructAsCsv(togglTime.logger)
	if err != nil {
		return
	}
	values := make([]interface{}, len(togglTimeEntries))
	for i, value := range togglTimeEntries {
		values[i] = value
	}
	return downloadStructAsCsv.DownloadAll(values, "toggl_time_entries")
}

// Store inserts the Toggl time entries into the database.
func (togglTime *TogglTime) Store(startTime time.Time, endTime time.Time) (err error) {
	togglTimeEntries, err := togglTime.retrieve(startTime, endTime)
	if err != nil {
		return
	}
	if len(togglTimeEntries) == 0 {
		togglTime.logger.Error("Skip the creation of the Toggl time entries into the database.")
		return &EmptyTimeResultError{}
	}
	for _, togglTimeEntry := range togglTimeEntries {
		err = togglTime.storeInDatabase(togglTimeEntry)
		if err != nil {
			return
		}
	}
	return nil
}

func (togglTime *TogglTime) retrieve(startTime time.Time, endTime time.Time) ([]TogglTimeEntry, error) {
	togglTimeEntries, err := togglTime.togglClient.GetRange(startTime, endTime)
	if err != nil {
		return nil, err
	}
	togglTime.logger.Info("Time entries", zap.Int("count", len(togglTimeEntries)))
	return togglTimeEntries, nil
}

func (togglTime *TogglTime) storeInDatabase(togglTimeEntry TogglTimeEntry) (err error) {
	if togglTime.databaseConnection == nil {
		err = &application_errors.DatabaseConnectionError{}
		return
	}
	tx, err := togglTime.databaseConnection.Begin()
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
	sqlStmt := `INSERT INTO toggl_time(id, description, start, stop, duration, billable, workspace_id, project_id, tags, trello_card_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, '')`
	_, err = togglTime.databaseConnection.Exec(
		sqlStmt,
		togglTimeEntry.Id, togglTimeEntry.Description, togglTimeEntry.Start, togglTimeEntry.Stop, togglTimeEntry.Duration,
		togglTimeEntry.Billable, togglTimeEntry.Workspace_id, togglTimeEntry.Project_id, pq.Array(togglTimeEntry.Tags))
	if err != nil {
		return
	}
	return
}
