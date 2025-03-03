package mysql

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/pakkasys/fluidapi/core"
)

type MySQLErrorCode uint16

const (
	ForeignConstraintErrorCode MySQLErrorCode = 1452
	DuplicateEntryErrorCode    MySQLErrorCode = 1062
)

var (
	DuplicateEntryError    = core.NewAPIError("DUPLICATE_ENTRY")
	ForeignConstraintError = core.NewAPIError("FOREIGN_CONSTRAINT_ERROR")
	NoRowsError            = core.NewAPIError("NO_ROWS")
)

type ErrorChecker struct{}

// Check is a function used to check if an error is a MySQL error.
func (c *ErrorChecker) Check(err error) error {
	if isMySQLErrorCode(err, DuplicateEntryErrorCode) {
		return DuplicateEntryError.WithData(err)
	} else if isMySQLErrorCode(err, ForeignConstraintErrorCode) {
		return ForeignConstraintError.WithData(err)
	} else if err == sql.ErrNoRows {
		return NoRowsError
	}
	return err
}

func isMySQLErrorCode(err error, code MySQLErrorCode) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == uint16(code)
	}
	return false
}
