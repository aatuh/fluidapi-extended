package sqlite

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/mattn/go-sqlite3"
	"github.com/pakkasys/fluidapi-extended/util"
)

type SQLiteErrorCode int

var (
	DuplicateEntryErrorCode    SQLiteErrorCode = SQLiteErrorCode(sqlite3.ErrConstraintUnique)
	ForeignConstraintErrorCode SQLiteErrorCode = SQLiteErrorCode(sqlite3.ErrConstraintForeignKey)
)

type ErrorChecker struct{}

// Check attempts to match a given error against common SQLite errors.
func (c *ErrorChecker) Check(err error) error {
	if err == nil {
		return nil
	}
	if isSQLiteErrorCode(err, DuplicateEntryErrorCode) {
		return util.DuplicateEntryError.WithData(err)
	} else if isSQLiteErrorCode(err, ForeignConstraintErrorCode) {
		return util.ForeignConstraintError.WithData(err)
	} else if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
		return util.NoRowsError
	}
	return err
}

// isSQLiteErrorCode checks whether 'err' is a sqlite3.Error
// and if its ExtendedCode or Code matches the desired code.
func isSQLiteErrorCode(err error, code SQLiteErrorCode) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return int(sqliteErr.ExtendedCode) == int(code)
	}
	return false
}
