package application

import "errors"

const (
	SearchSpecialistsSpanName = "SearchSpecialistsCommand.Execute"

	StartingSearchMessage        = "Starting specialist search"
	SearchCompletedMessage       = "Specialist search completed"
	SearchNoResultsMessage       = "No specialists found matching criteria"
	ErrSearchExecutionMessage    = "Failed to execute search"
	ErrInvalidSearchInputMessage = "Invalid search input parameters"
)

var (
	ErrSearchExecution    = errors.New(ErrSearchExecutionMessage)
	ErrInvalidSearchInput = errors.New(ErrInvalidSearchInputMessage)
)
