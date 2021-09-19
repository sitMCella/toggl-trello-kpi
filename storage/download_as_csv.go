package storage

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/sitMCella/toggl-trello-kpi/application_errors"
	"go.uber.org/zap"
)

// DownloadAsCsv struct defines the download from database as CSV service.
type DownloadAsCsv struct {
	logger             *zap.Logger
	databaseConnection *sql.DB
}

// NewDownloadAsCsv creates a new DownloadAsCsv.
func NewDownloadAsCsv(logger *zap.Logger, databaseConnection *sql.DB) (*DownloadAsCsv, error) {
	if logger == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "logger"}
	}
	if databaseConnection == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "databaseConnection"}
	}
	return &DownloadAsCsv{
		logger:             logger,
		databaseConnection: databaseConnection,
	}, nil
}

// DownloadAll downloads all the entries from a specific table in the database.
func (downloadAsCsv *DownloadAsCsv) DownloadAll(databaseTableName string) (err error) {
	sqlStmt := fmt.Sprintf(`SELECT * FROM %s`, databaseTableName)
	rows, err := downloadAsCsv.databaseConnection.Query(sqlStmt)
	if err != nil {
		return
	}
	defer func() {
		sqlerr := rows.Close()
		if err == nil {
			err = sqlerr
		}
	}()

	fileName := fmt.Sprintf("%s.csv", databaseTableName)
	err = downloadAsCsv.createCsv(fileName, rows)
	if err != nil {
		return
	}
	return rows.Err()
}

// Download downloads the all the entries from a specific table in the database with a filter on the columns.
func (downloadAsCsv *DownloadAsCsv) Download(databaseTableName string, columnsFilter []string) (err error) {
	columns := strings.Join(columnsFilter, ",")
	sqlStmt := fmt.Sprintf(`SELECT %s FROM %s`, columns, databaseTableName)
	rows, err := downloadAsCsv.databaseConnection.Query(sqlStmt)
	if err != nil {
		return
	}
	defer func() {
		sqlerr := rows.Close()
		if err == nil {
			err = sqlerr
		}
	}()

	fileName := fmt.Sprintf("%s.csv", databaseTableName)
	err = downloadAsCsv.createCsv(fileName, rows)
	if err != nil {
		return
	}
	return rows.Err()
}

func (downloadAsCsv *DownloadAsCsv) createCsv(fileName string, rows *sql.Rows) (err error) {
	csvfile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func() {
		fileerr := csvfile.Close()
		if err == nil {
			err = fileerr
		}
	}()
	csvWriter := csv.NewWriter(csvfile)
	defer csvWriter.Flush()

	columnNames, err := rows.Columns()
	if err != nil {
		return
	}

	err = csvWriter.Write(columnNames)
	if err != nil {
		return err
	}

	for rows.Next() {
		values, err := rowMapString(columnNames, rows)
		if err != nil {
			return err
		}
		var row = make([]string, len(values))
		for columnName, value := range values {
			i := SliceIndex(len(columnNames), func(i int) bool { return columnNames[i] == columnName })
			row[i] = fmt.Sprintf(string(value))
		}
		err = csvWriter.Write(row)
		if err != nil {
			return err
		}
	}
	return
}

func rowMapString(columnNames []string, rows *sql.Rows) (map[string]string, error) {
	lenCN := len(columnNames)
	ret := make(map[string]string, lenCN)

	columnPointers := make([]interface{}, lenCN)
	for i := 0; i < lenCN; i++ {
		columnPointers[i] = new(sql.RawBytes)
	}

	if err := rows.Scan(columnPointers...); err != nil {
		return nil, err
	}

	for i := 0; i < lenCN; i++ {
		if rb, ok := columnPointers[i].(*sql.RawBytes); ok {
			ret[columnNames[i]] = string(*rb)
		} else {
			return nil, fmt.Errorf("Cannot convert index %d column %s to type *sql.RawBytes", i, columnNames[i])
		}
	}

	return ret, nil
}

func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}
