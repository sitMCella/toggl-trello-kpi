package trello

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bmizerany/assert"
	"github.com/lib/pq"
	"github.com/sitMCella/toggl-trello-kpi/application_errors"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockTrelloClient struct {
	mock.Mock
	emptyTrelloCardEntries bool
}

func (mockTrelloClient *MockTrelloClient) GetCards() ([]TrelloCardEntry, error) {
	var trelloCardEntries []TrelloCardEntry
	if mockTrelloClient.emptyTrelloCardEntries {
		return trelloCardEntries, nil
	}
	trelloCardEntry := TrelloCardEntry{
		Id:       "45636633",
		Name:     "Card name",
		Closed:   false,
		Labels:   []string{"Project name", "Customer name", "Task type"},
		Project:  "Project name",
		Customer: "Customer name",
		Team:     "Team name",
		Type:     "Task type",
	}
	trelloCardEntries = append(trelloCardEntries, trelloCardEntry)
	return trelloCardEntries, nil
}

func TestTrelloCreateThrowsErrorOnNilLogger(t *testing.T) {
	mockTrelloClient := &MockTrelloClient{}

	_, err := NewTrello(nil, mockTrelloClient)
	verifyNilParameterError(t, err, "logger")
}

func TestTrelloCreateThrowsErrorOnNilTrelloClient(t *testing.T) {
	logger, err := getLogger()
	if err != nil {
		t.Fatalf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()

	_, err = NewTrello(logger, nil)
	verifyNilParameterError(t, err, "trelloClient")
}

func TestTrelloCreateWithDatabaseConnectionThrowsErrorOnNilLogger(t *testing.T) {
	mockTrelloClient := &MockTrelloClient{}
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	_, err = NewTrelloWithDatabaseConnection(nil, mockTrelloClient, db)
	verifyNilParameterError(t, err, "logger")
}

func TestTrelloCreateWithDatabaseConnectionThrowsErrorOnNilTrelloClient(t *testing.T) {
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

	_, err = NewTrelloWithDatabaseConnection(logger, nil, db)
	verifyNilParameterError(t, err, "trelloClient")
}

func TestTrelloCreateWithDatabaseConnectionThrowsErrorOnNilDatabaseConnection(t *testing.T) {
	logger, err := getLogger()
	if err != nil {
		t.Fatalf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()
	mockTrelloClient := &MockTrelloClient{}

	_, err = NewTrelloWithDatabaseConnection(logger, mockTrelloClient, nil)
	verifyNilParameterError(t, err, "databaseConnection")
}

func verifyNilParameterError(t *testing.T, err error, parameterName string) {
	if err == nil {
		t.Fatalf("Expect an error while creating Trello with nil %s.", parameterName)
	}
	switch err.(type) {
	case *application_errors.NilParameterError:
		return
	default:
		t.Errorf("Expect a NilParameterError while creating Trello with nil %s.", parameterName)
	}
}

func TestTrelloDownloadAsCsv(t *testing.T) {
	logger, err := getLogger()
	if err != nil {
		t.Fatalf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()
	mockTrelloClient := &MockTrelloClient{}
	trello, err := NewTrello(logger, mockTrelloClient)
	if err != nil {
		t.Fatalf("Error creating Trello: %v", err)
	}

	err = trello.DownloadAsCsv()
	if err != nil {
		t.Fatalf("Error in Trello DownloadAsCsv: %v", err)
	}

	trelloEntriesFileName := "trello_entries.csv"
	if _, err := os.Stat(trelloEntriesFileName); err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Couldn't find file %s: %v", trelloEntriesFileName, err)
		} else {
			t.Fatalf("Couldn't get file %s: %v", trelloEntriesFileName, err)
		}
	}
	defer func() {
		oserr := os.Remove(trelloEntriesFileName)
		if oserr != nil {
			t.Fatalf("Error on removing temp file: %v", oserr)
		}
	}()
	data, err := ioutil.ReadFile(trelloEntriesFileName)
	if err != nil {
		t.Fatalf("Error while reading the file %s: %v", trelloEntriesFileName, err)
	}
	expectedData := `Id,Name,Closed,Labels,Project,Customer,Team,Type
45636633,Card name,false,"Project name,Customer name,Task type",Project name,Customer name,Team name,Task type
`
	assert.Equal(t, []byte(expectedData), data, "Expected same file content")
}

func TestTrelloDownloadAsCsvThrowsEmptyTrelloCardsErrorOnEmptyTrelloCardEntries(t *testing.T) {
	logger, err := getLogger()
	if err != nil {
		t.Fatalf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()
	mockTrelloClient := &MockTrelloClient{emptyTrelloCardEntries: true}
	trello, err := NewTrello(logger, mockTrelloClient)
	if err != nil {
		t.Fatalf("Error creating Trello: %v", err)
	}

	err = trello.DownloadAsCsv()

	if err == nil {
		t.Fatalf("Expect an error in Trello DownloadAsCsv with empty Trello card entries from trelloClient")
	}
	switch err.(type) {
	case *EmptyTrelloCardsError:
		return
	default:
		t.Errorf("Expect an EmptyTrelloCardsError error in Trello DownloadAsCsv with empty Trello card entries from trelloClient")
	}
}

func TestTrelloStoreInDatabase(t *testing.T) {
	logger, err := getLogger()
	if err != nil {
		t.Fatalf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()
	mockTrelloClient := &MockTrelloClient{}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	trello, err := NewTrelloWithDatabaseConnection(logger, mockTrelloClient, db)
	if err != nil {
		t.Fatalf("Error creating Trello: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO trello_card").
		WithArgs("45636633", "Card name", false, pq.Array([]string{"Project name", "Customer name", "Task type"}), "Project name", "Customer name", "Team name", "Task type").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = trello.Store()
	if err != nil {
		t.Errorf("Error in Trello Store: %v", err)
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
