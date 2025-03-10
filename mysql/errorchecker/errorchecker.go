package errorchecker

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/pakkasys/fluidapi-extended/database"
)

// MySQLErrorCode represents a MySQL error code.
type MySQLErrorCode uint16

// Constants for MySQL error codes.
const (
	ForeignConstraintErrorCode MySQLErrorCode = 1452
	DuplicateEntryErrorCode    MySQLErrorCode = 1062
)

// ErrorChecker is a MySQL error checker.
type ErrorChecker struct{}

// NewErrorChecker returns a new ErrorChecker.
func NewErrorChecker() *ErrorChecker {
	return &ErrorChecker{}
}

// Check is a function used to check if an error is a MySQL error.
//
// Parameters:
//   - err: The error to check.
//
// Returns:
//   - error: The checked error.
func (c *ErrorChecker) Check(err error) error {
	if isMySQLErrorCode(err, DuplicateEntryErrorCode) {
		return database.DuplicateEntryError.WithData(err)
	} else if isMySQLErrorCode(err, ForeignConstraintErrorCode) {
		return database.ForeignConstraintError.WithData(err)
	} else if err == sql.ErrNoRows {
		return database.NoRowsError
	}
	return err
}

// isMySQLErrorCode checks whether 'err' is a mysql.MySQLError and if its Number
// matches the desired code.
func isMySQLErrorCode(err error, code MySQLErrorCode) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == uint16(code)
	}
	return false
}
