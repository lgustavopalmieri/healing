package application

import "errors"

const (
	SearchSpecialistsSpanName = "SearchSpecialistsCommand.Execute"

	ErrSearchExecutionMessage    = "Failed to execute search"
	ErrInvalidSearchInputMessage = "Invalid search input parameters"
)

var (
	ErrSearchExecution    = errors.New(ErrSearchExecutionMessage)
	ErrInvalidSearchInput = errors.New(ErrInvalidSearchInputMessage)
)
