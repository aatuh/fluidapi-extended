package errorchecker

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/mattn/go-sqlite3"
	"github.com/pakkasys/fluidapi-extended/database"
)

// SQLiteErrorCode represents a SQLite error code.
type SQLiteErrorCode int

// Common SQLite error codes.
var (
	DuplicateEntryErrorCode    SQLiteErrorCode = SQLiteErrorCode(sqlite3.ErrConstraintUnique)
	ForeignConstraintErrorCode SQLiteErrorCode = SQLiteErrorCode(sqlite3.ErrConstraintForeignKey)
)

// ErrorChecker is used to check if an error is a SQLite error.
type ErrorChecker struct {
	systemId string
}

// NewErrorChecker returns a new ErrorChecker.
func NewErrorChecker(systemId string) *ErrorChecker {
	return &ErrorChecker{
		systemId: systemId,
	}
}

// Check attempts to match a given error against common SQLite errors.
//
// Parameters:
//   - err: The error to check.
//
// Returns:
//   - error: The checked error.
func (c *ErrorChecker) Check(err error) error {
	if err == nil {
		return nil
	}
	if isSQLiteErrorCode(err, DuplicateEntryErrorCode) {
		return database.DuplicateEntryError.WithData(err).WithOrigin(c.systemId)
	} else if isSQLiteErrorCode(err, ForeignConstraintErrorCode) {
		return database.ForeignConstraintError.WithData(err).WithOrigin(c.systemId)
	} else if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
		return database.NoRowsError.WithOrigin(c.systemId)
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
