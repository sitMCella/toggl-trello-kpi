package storage

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/bmizerany/assert"
	"github.com/sitMCella/toggl-trello-kpi/application_errors"
	"go.uber.org/zap"
)

type ExampleStruct struct {
	StringField string
	Int64Field  int64
	Uint64Field uint64
	BoolField   bool
	TimeField   time.Time
	StringArray []string
}

func TestDownloadStructAsCsvCreateThrowsErrorOnNilLogger(t *testing.T) {
	_, err := NewDownloadStructAsCsv(nil)
	if err == nil {
		t.Fatalf("Expect an error while creating DownloadStructAsCsv with nil logger.")
	}
	switch err.(type) {
	case *application_errors.NilParameterError:
		return
	default:
		t.Errorf("Expect a NilParameterError while creating DownloadStructAsCsv with nil logger.")
	}
}

func TestDownloadStructAsCsv(t *testing.T) {
	logger, err := getLogger()
	if err != nil {
		t.Fatalf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()
	downloadStructAsCsv, err := NewDownloadStructAsCsv(logger)
	if err != nil {
		t.Fatalf("Error creating DownloadStructAsCsv: %v", err)
	}
	exampleStructEntry1 := ExampleStruct{
		StringField: "string field value",
		Int64Field:  75,
		Uint64Field: 9,
		BoolField:   true,
		TimeField:   time.Date(2021, time.Month(01), 00, 0, 0, 0, 0, time.UTC),
		StringArray: []string{"string array field 1", "string array field 2"},
	}
	var exampleStructEntries []ExampleStruct
	exampleStructEntries = append(exampleStructEntries, exampleStructEntry1)
	values := make([]interface{}, len(exampleStructEntries))
	for i, value := range exampleStructEntries {
		values[i] = value
	}

	err = downloadStructAsCsv.DownloadAll(values, "example_struct_entries")
	if err != nil {
		t.Fatalf("Error in DownloadStructAsCsv DownloadAll: %v", err)
	}

	fileName := "example_struct_entries.csv"
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Couldn't find file %s: %v", fileName, err)
		} else {
			t.Fatalf("Couldn't get file %s: %v", fileName, err)
		}
	}
	defer func() {
		oserr := os.Remove(fileName)
		if oserr != nil {
			t.Fatalf("Error on removing temp file: %v", oserr)
		}
	}()
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Error while reading the file %s: %v", fileName, err)
	}
	expectedData := `StringField,Int64Field,Uint64Field,BoolField,TimeField,StringArray
string field value,75,9,true,2020-12-31T00:00:00Z,"string array field 1,string array field 2"
`
	assert.Equal(t, []byte(expectedData), data, "Expected same file content")
}

func TestDownloadStructAsCsvThrowsEmptyEntriesErrorOnNilEntries(t *testing.T) {
	logger, err := getLogger()
	if err != nil {
		t.Fatalf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()
	downloadStructAsCsv, err := NewDownloadStructAsCsv(logger)
	if err != nil {
		t.Fatalf("Error creating DownloadStructAsCsv: %v", err)
	}

	err = downloadStructAsCsv.DownloadAll(nil, "example_struct_entries")

	if err == nil {
		t.Fatalf("Expect an error in DownloadStructAsCsv DownloadAll with empty entries parameter")
	}
	switch err.(type) {
	case *EmptyEntriesError:
		return
	default:
		t.Errorf("Expect an EmptyEntriesError error in DownloadStructAsCsv DownloadAll with empty entries parameter")
	}
}

func TestDownloadStructAsCsvThrowsEmptyEntriesErrorOnEmptyEntries(t *testing.T) {
	logger, err := getLogger()
	if err != nil {
		t.Fatalf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()
	downloadStructAsCsv, err := NewDownloadStructAsCsv(logger)
	if err != nil {
		t.Fatalf("Error creating DownloadStructAsCsv: %v", err)
	}
	var values = make([]interface{}, 0)

	err = downloadStructAsCsv.DownloadAll(values, "example_struct_entries")

	if err == nil {
		t.Fatalf("Expect an error in DownloadStructAsCsv DownloadAll with empty entries parameter")
	}
	switch err.(type) {
	case *EmptyEntriesError:
		return
	default:
		t.Errorf("Expect an EmptyEntriesError error in DownloadStructAsCsv DownloadAll with empty entries parameter")
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
