package application_errors

import "fmt"

// NilParameterError defines the function Nil parameter error.
type NilParameterError struct {
	ParameterName string
}

func (err *NilParameterError) Error() string {
	return fmt.Sprintf("The function parameter %s is nil.", err.ParameterName)
}

// DatabaseConnectionError defines the database connection error.
type DatabaseConnectionError struct {
}

func (err *DatabaseConnectionError) Error() string {
	return fmt.Sprintf("The database connection is null.")
}
