package database

import (
	"github.com/pakkasys/fluidapi/core"
)

// Commmon database errors.
var (
	DuplicateEntryError    = core.NewAPIError("DUPLICATE_ENTRY")
	ForeignConstraintError = core.NewAPIError("FOREIGN_CONSTRAINT_ERROR")
	NoRowsError            = core.NewAPIError("NO_ROWS")
)
