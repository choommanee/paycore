package domain

import "errors"

// Sentinel domain errors. Handlers map these to HTTP status + stable error codes.
var (
	ErrNotImplemented      = errors.New("not implemented")
	ErrInvalidRequest      = errors.New("invalid request")
	ErrPaymentNotFound     = errors.New("payment not found")
	ErrDuplicateRequest    = errors.New("duplicate idempotency key")
	ErrIdempotencyKeyReuse = errors.New("idempotency key reused with a different request")
	ErrCardDeclined        = errors.New("card declined")
	ErrInsufficientFund    = errors.New("insufficient funds")
	ErrThreeDSRequired     = errors.New("3ds authentication required")
	ErrRefundExceeds       = errors.New("refund exceeds captured amount")
	ErrInvalidState        = errors.New("invalid payment state transition")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrMerchantNotFound    = errors.New("merchant not found")
	ErrDisputeNotFound     = errors.New("dispute not found")
	ErrForbidden           = errors.New("forbidden")
)
