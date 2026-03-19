package application

import "errors"

const (
	ErrSearchExecutionMessage    = "Failed to execute search"
	ErrInvalidSearchInputMessage = "Invalid search input parameters"
)

var (
	ErrSearchExecution    = errors.New(ErrSearchExecutionMessage)
	ErrInvalidSearchInput = errors.New(ErrInvalidSearchInputMessage)
)
