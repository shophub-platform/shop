package repository

import "errors"

var (
	ErrNotFound          = errors.New("record not found")
	ErrAlreadyExists     = errors.New("record already exists")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrDuplicateTxHash   = errors.New("transaction hash already used")
)
