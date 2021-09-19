package toggl

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bmizerany/assert"
	"github.com/lib/pq"
	"github.com/sitMCella/toggl-trello-kpi/application_errors"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockTogglClient struct {
	mock.Mock
	emptyTogglEntries bool
}

func (mockTogglClient *MockTogglClient) GetRange(start time.Time, end time.Time) ([]TogglTimeEntry, error) {
	var togglTimeEntries []TogglTimeEntry
	if mockTogglClient.emptyTogglEntries {
		return togglTimeEntries, nil
	}
	togglTimeEntry := TogglTimeEntry{
		Id:             86854567,
		Description:    "description",
		Start:          time.Date(2021, time.Month(02), 01, 9, 15, 0, 0, time.UTC),
		Stop:           time.Date(2021, time.Month(02), 01, 9, 30, 0, 0, time.UTC),
		Duration:       900,
		Billable:       true,
		Workspace_id:   2245503,
		Project_id:     7458839,
		Tags:           []string{"tag1"},
		Trello_card_id: "",
	}
	togglTimeEntries = append(togglTimeEntries, togglTimeEntry)
	return togglTimeEntries, nil
}

func TestTogglCreateThrowsErrorOnNilLogger(t *testing.T) {
	mockTogglClient := &MockTogglClient{}

	_, err := NewTogglTime(nil, mockTogglClient)
	verifyNilParameterError(t, err, "logger")
}

func TestTogglCreateThrowsErrorOnNilTogglClient(t *testing.T) {
	logger, err := getLogger()
	if err != nil {
		t.Fatalf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()

	_, err = NewTogglTime(logger, nil)
	verifyNilParameterError(t, err, "togglClient")
}

func TestTogglCreateWithDatabaseConnectionThrowsErrorOnNilLogger(t *testing.T) {
	mockTogglClient := &MockTogglClient{}
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	_, err = NewTogglTimeWithDatabaseConnection(nil, mockTogglClient, db)
	verifyNilParameterError(t, err, "logger")
}

func TestTogglCreateWithDatabaseConnectionThrowsErrorOnNilTogglClient(t *testing.T) {
	logger, err := getLogger()
	if err != nil {
		t.Fatalf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	_, err = NewTogglTimeWithDatabaseConnection(logger, nil, db)
	verifyNilParameterError(t, err, "togglClient")
}

func TestTogglCreateWithDatabaseConnectionThrowsErrorOnNilDatabaseConnection(t *testing.T) {
	logger, err := getLogger()
	if err != nil {
		t.Fatalf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()
	mockTogglClient := &MockTogglClient{}

	_, err = NewTogglTimeWithDatabaseConnection(logger, mockTogglClient, nil)
	verifyNilParameterError(t, err, "databaseConnection")
}

func verifyNilParameterError(t *testing.T, err error, parameterName string) {
	if err == nil {
		t.Fatalf("Expect an error while creating TogglTime with nil %s.", parameterName)
	}
	switch err.(type) {
	case *application_errors.NilParameterError:
		return
	default:
		t.Errorf("Expect a NilParameterError while creating TogglTime with nil %s.", parameterName)
	}
}

func TestTogglDownloadAsCsv(t *testing.T) {
	logger, err := getLogger()
	if err != nil {
		t.Fatalf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()
	mockTogglClient := &MockTogglClient{}
	togglTime, err := NewTogglTime(logger, mockTogglClient)
	if err != nil {
		t.Fatalf("Error creating TogglTime: %v", err)
	}
	startTime := time.Date(2021, time.Month(02), 01, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2021, time.Month(02), 28, 0, 0, 0, 0, time.UTC)

	err = togglTime.DownloadAsCsv(startTime, endTime)
	if err != nil {
		t.Fatalf("Error in DownloadAsCsv: %v", err)
	}

	togglTimeEntriesFileName := "toggl_time_entries.csv"
	if _, err := os.Stat(togglTimeEntriesFileName); err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Couldn't find file %s: %v", togglTimeEntriesFileName, err)
		} else {
			t.Fatalf("Couldn't get file %s: %v", togglTimeEntriesFileName, err)
		}
	}
	defer func() {
		oserr := os.Remove(togglTimeEntriesFileName)
		if oserr != nil {
			t.Fatalf("Error on removing temp file: %v", oserr)
		}
	}()
	data, err := ioutil.ReadFile(togglTimeEntriesFileName)
	if err != nil {
		t.Fatalf("Error while reading the file %s: %v", togglTimeEntriesFileName, err)
	}
	expectedData := `Id,Description,Start,Stop,Duration,Billable,Workspace_id,Project_id,Tags,Trello_card_id
86854567,description,2021-02-01T09:15:00Z,2021-02-01T09:30:00Z,900,true,2245503,7458839,tag1,
`
	assert.Equal(t, []byte(expectedData), data, "Expected same file content")
}

func TestTogglDownloadAsCsvThrowsEmptyTimeResultErrorOnEmptyTogglEntries(t *testing.T) {
	logger, err := getLogger()
	if err != nil {
		t.Fatalf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()
	mockTogglClient := &MockTogglClient{emptyTogglEntries: true}

	togglTime, err := NewTogglTime(logger, mockTogglClient)
	if err != nil {
		t.Fatalf("Error creating TogglTime: %v", err)
	}
	startTime := time.Date(2021, time.Month(02), 01, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2021, time.Month(02), 28, 0, 0, 0, 0, time.UTC)

	err = togglTime.DownloadAsCsv(startTime, endTime)
	if err == nil {
		t.Fatalf("Expect an error in TogglTime DownloadAsCsv with empty Toggl entries from togglClient")
	}
	switch err.(type) {
	case *EmptyTimeResultError:
		return
	default:
		t.Errorf("Expect an EmptyTimeResultError in TogglTime DownloadAsCsv with empty Toggl entries from togglClient")
	}
}

func TestTogglStoreInDatabase(t *testing.T) {
	logger, err := getLogger()
	if err != nil {
		t.Fatalf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()
	mockTogglClient := &MockTogglClient{}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	togglTime, err := NewTogglTimeWithDatabaseConnection(logger, mockTogglClient, db)
	if err != nil {
		t.Fatalf("Error creating TogglTime: %v", err)
	}
	startTime := time.Date(2021, time.Month(02), 01, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2021, time.Month(02), 28, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO toggl_time").
		WithArgs(86854567, "description", time.Date(2021, time.Month(02), 01, 9, 15, 0, 0, time.UTC), time.Date(2021, time.Month(02), 01, 9, 30, 0, 0, time.UTC), 900, true, 2245503, 7458839, pq.Array([]string{"tag1"})).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = togglTime.Store(startTime, endTime)
	if err != nil {
		t.Errorf("Error in TogglTime Store: %v", err)
	}
}

func getLogger() (*zap.Logger, error) {
	zapCfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.FatalLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	return zapCfg.Build()
}
