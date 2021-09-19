package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sitMCella/toggl-trello-kpi/application_errors"
	"go.uber.org/zap"
)

// DownloadStructAsCsv struct defines the Download of a struct as CSV service.
type DownloadStructAsCsv struct {
	logger *zap.Logger
}

// EmptyEntriesError defines the empty entries error.
type EmptyEntriesError struct {
}

func (error *EmptyEntriesError) Error() string {
	return fmt.Sprintf("The struct entries array is empty.")
}

// NewDownloadStructAsCsv creates a DownloadStructAsCsv.
func NewDownloadStructAsCsv(logger *zap.Logger) (*DownloadStructAsCsv, error) {
	if logger == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "logger"}
	}
	return &DownloadStructAsCsv{
		logger: logger,
	}, nil
}

// DownloadAll downloads all the entries from an array of struct as a CSV file.
func (downloadStructAsCsv *DownloadStructAsCsv) DownloadAll(entries []interface{}, name string) (err error) {
	if entries == nil || len(entries) == 0 {
		return &EmptyEntriesError{}
	}
	fileName := fmt.Sprintf("%s.csv", name)
	csvfile, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer func() {
		fileerr := csvfile.Close()
		if err == nil {
			err = fileerr
		}
	}()
	csvWriter := csv.NewWriter(csvfile)
	defer csvWriter.Flush()

	columnNames := retrieveColumnNames(entries)

	err = csvWriter.Write(columnNames)
	if err != nil {
		return
	}

	for j := 0; j < len(entries); j++ {
		values := downloadStructAsCsv.retrieveFieldValues(entries[j], columnNames)
		err = csvWriter.Write(values)
		if err != nil {
			return
		}
	}
	return
}

func retrieveColumnNames(entries []interface{}) []string {
	fields := reflect.Indirect(reflect.ValueOf(entries[0]))
	var columnNames = make([]string, fields.Type().NumField())
	for i := 0; i < fields.Type().NumField(); i++ {
		columnNames[i] = fields.Type().Field(i).Name
	}
	return columnNames
}

func (downloadStructAsCsv *DownloadStructAsCsv) retrieveFieldValues(entry interface{}, columnNames []string) []string {
	var values = make([]string, len(columnNames))
	fields := reflect.Indirect(reflect.ValueOf(entry))
	for i := 0; i < len(columnNames); i++ {
		field := fields.Field(i)
		fieldValue := field.Interface()
		switch fieldValue := fieldValue.(type) {
		case string:
			values[i] = fieldValue
		case int64:
			values[i] = strconv.FormatInt(fieldValue, 10)
		case uint64:
			values[i] = strconv.FormatUint(fieldValue, 10)
		case bool:
			values[i] = strconv.FormatBool(fieldValue)
		case time.Time:
			values[i] = fieldValue.Format(time.RFC3339Nano)
		case []string:
			values[i] = strings.Join(fieldValue, ",")
		default:
			downloadStructAsCsv.logger.Error("Cannot convert the data type", zap.String("Data type", fmt.Sprintf("%T", fieldValue)))
		}
	}
	return values
}
