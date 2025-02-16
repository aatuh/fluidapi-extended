package util

import (
	"github.com/pakkasys/fluidapi/core"
)

var (
	DuplicateEntryError    = core.NewAPIError("DUPLICATE_ENTRY")
	ForeignConstraintError = core.NewAPIError("FOREIGN_CONSTRAINT_ERROR")
	NoRowsError            = core.NewAPIError("NO_ROWS")
)
