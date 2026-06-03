package repository

import "errors"

var (
	ErrNotFound          = errors.New("record not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrDuplicateTxHash   = errors.New("transaction hash already used")
)
