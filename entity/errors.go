package entity

import "errors"

var (
	ErrGraphNotFound       = errors.New("graph not found")
	ErrDuplicatedGraphName = errors.New("duplicated graph name")

	ErrVertexNotFound = errors.New("vertex not found")

	ErrEdgeNotFound = errors.New("edge not found")
)
