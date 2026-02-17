package services

import "errors"

var (
	ErrInvalidInput      = errors.New("invalid input data")
	ErrAlreadyExists     = errors.New("entity already exists")
	ErrNotFound          = errors.New("entity not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInsufficientStock = errors.New("insufficient product quantity")
	ErrInvalidAmount     = errors.New("amount must be positive")
	ErrInvalidQuantity   = errors.New("quantity must be positive")
)
