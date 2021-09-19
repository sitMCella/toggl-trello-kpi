package storage

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/sitMCella/toggl-trello-kpi/application_errors"
	"go.uber.org/zap"
)

// DatabaseConnectionError struct defines the database connection error.
type DatabaseConnectionError struct {
}

func (error *DatabaseConnectionError) Error() string {
	return fmt.Sprintf("The database connection is null.")
}

// InsertFromCsv struct defines the insert data in database from CSV service.
type InsertFromCsv struct {
	logger             *zap.Logger
	databaseConnection *sql.DB
}

// NewInsertFromCsv creates a new InsertFromCsv.
func NewInsertFromCsv(logger *zap.Logger, databaseConnection *sql.DB) (*InsertFromCsv, error) {
	if logger == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "logger"}
	}
	if databaseConnection == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "databaseConnection"}
	}
	return &InsertFromCsv{
		logger:             logger,
		databaseConnection: databaseConnection,
	}, nil
}

// Insert inserts all the entries from the provided CSV file with a specific data type to a database table.
func (insertFromCsv *InsertFromCsv) Insert(fileName string, databaseTableName string, dataType interface{}) (err error) {
	if insertFromCsv.databaseConnection == nil {
		err = &DatabaseConnectionError{}
		return
	}
	file, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer func() {
		fileerr := file.Close()
		if err == nil {
			err = fileerr
		}
	}()

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return
	}
	if len(lines) == 0 {
		insertFromCsv.logger.Info("The CSV file is empty.")
		return
	}
	for i, line := range lines {
		if i == 0 {
			continue
		}
		err = insertFromCsv.insertRow(databaseTableName, lines[0], line, dataType)
		if err != nil {
			return
		}
	}
	return nil
}

func (insertFromCsv *InsertFromCsv) insertRow(databaseTableName string, columnNames []string, line []string, dataType interface{}) error {
	columns := strings.Join(columnNames, ",")
	var valuesQuery = make([]string, len(columnNames))
	for i := 0; i < len(columnNames); i++ {
		valuesQuery[i] += fmt.Sprintf("$%d", i+1)
	}
	valuesQueryJoin := strings.Join(valuesQuery, ",")
	sqlStmt := fmt.Sprintf(`INSERT INTO %s(%s) VALUES (%s)`, databaseTableName, columns, valuesQueryJoin)

	var args []reflect.Value
	args = append(args, reflect.ValueOf(sqlStmt))
	togglTimeEntryFields := reflect.Indirect(reflect.ValueOf(dataType))
	for i := 0; i < len(line); i++ {
		field := togglTimeEntryFields.FieldByName(columnNames[i])
		fieldValue := field.Interface()
		switch fieldValue.(type) {
		case string:
			args = append(args, reflect.ValueOf(line[i]))
		case int64:
			value, err := strconv.ParseInt(line[i], 10, 64)
			if err != nil {
				return err
			}
			args = append(args, reflect.ValueOf(value))
		case uint64:
			value, err := strconv.ParseUint(line[i], 10, 64)
			if err != nil {
				return err
			}
			args = append(args, reflect.ValueOf(value))
		case bool:
			value, err := strconv.ParseBool(line[i])
			if err != nil {
				return err
			}
			args = append(args, reflect.ValueOf(value))
		case time.Time:
			value, err := time.Parse(time.RFC3339Nano, line[i])
			if err != nil {
				return err
			}
			args = append(args, reflect.ValueOf(value))
		case []string:
			value := pq.Array(strings.Split(line[i], ","))
			args = append(args, reflect.ValueOf(value))
		default:
			insertFromCsv.logger.Error("Cannot convert the data type", zap.String("Data type", fmt.Sprintf("%T", fieldValue)))
		}
	}
	fun := reflect.ValueOf(insertFromCsv.databaseConnection.Exec)
	result := fun.Call(args)
	if len(result) > 0 {
		if reflect.TypeOf(result[0]) == reflect.TypeOf((*error)(nil)).Elem() {
			err := result[0].Interface().(error)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
