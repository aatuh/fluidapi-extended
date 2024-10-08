package mysqlutil

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	databaseerrors "github.com/pakkasys/fluidapi/database/errors"
)

type MySQLErrorCode uint16

const (
	ForeignConstraintError MySQLErrorCode = 1452
	DuplicateEntryError    MySQLErrorCode = 1062
)

type MySQLUtil struct{}

// CheckDBError is a function used to check if an error is a MySQL error and
// matches the given code.
func (c *MySQLUtil) CheckDBError(err error) error {
	if isMySQLError(err, DuplicateEntryError) {
		return databaseerrors.DuplicateEntryError.WithData(err)
	} else if isMySQLError(err, ForeignConstraintError) {
		return databaseerrors.ForeignConstraintError.WithData(err)
	}
	return err
}

func isMySQLError(err error, code MySQLErrorCode) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == uint16(code)
	}
	return false
}
