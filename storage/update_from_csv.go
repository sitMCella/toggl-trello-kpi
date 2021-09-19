package storage

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"

	"go.uber.org/zap"
)

// UpdateFromCsv defines the update database entries from CSV file service.
type UpdateFromCsv struct {
	logger             *zap.Logger
	databaseConnection *sql.DB
}

// NewUpdateFromCsv creates a new UpdateFromCsv.
func NewUpdateFromCsv(logger *zap.Logger, databaseConnection *sql.DB) UpdateFromCsv {
	updateFromCsv := UpdateFromCsv{
		logger:             logger,
		databaseConnection: databaseConnection,
	}
	return updateFromCsv
}

// Upload updates the table column values for each entries in a database table from a CSV file.
func (updateFromCsv *UpdateFromCsv) Upload(fileName string, databaseTableName string, columnName string) (err error) {
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
		updateFromCsv.logger.Error("The CSV file is empty.")
		return
	}
	idColumnIndex := SliceIndex(len(lines[0]), func(i int) bool { return lines[0][i] == "id" })
	if idColumnIndex == -1 {
		updateFromCsv.logger.Error("The CSV file must contain the ID column.")
		return
	}
	columnIndex := SliceIndex(len(lines[0]), func(i int) bool { return lines[0][i] == columnName })
	if columnIndex == -1 {
		updateFromCsv.logger.Error("The CSV file does not contain a required column.", zap.String("Column name", columnName))
		return
	}
	for i, line := range lines {
		if i == 0 {
			continue
		}
		updateFromCsv.updateRow(databaseTableName, columnName, line[columnIndex], line[idColumnIndex])
	}
	return nil
}

func (updateFromCsv *UpdateFromCsv) updateRow(databaseTableName string, columnName string, value string, id string) (err error) {
	sqlStmt := fmt.Sprintf(`UPDATE %s SET %s = $1 WHERE id = $2`, databaseTableName, columnName)
	_, err = updateFromCsv.databaseConnection.Exec(sqlStmt, value, id)
	return
}
