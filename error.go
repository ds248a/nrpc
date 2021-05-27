package nrpc

import "errors"

// ------------------
//   Client error
// ------------------

var (
	// Represents a timeout error because of timer or context.
	ErrClientTimeout = errors.New("timeout")

	// Represents an error of 0 time parameter.
	ErrClientInvalidTimeoutZero = errors.New("invalid timeout, should not be 0")

	// Represents an error of less than 0 time parameter.
	ErrClientInvalidTimeoutLessThanZero = errors.New("invalid timeout, should not be < 0")

	// Represents an error with 0 time parameter but with non-nil callback.
	ErrClientInvalidTimeoutZeroWithNonNilCallback = errors.New("invalid timeout 0 with non-nil callback")

	// Represents an error of Client's send queue is full.
	ErrClientOverstock = errors.New("timeout: rpc Client's send queue is full")

	// Represents an error that Client is reconnecting.
	ErrClientReconnecting = errors.New("client reconnecting")

	// Represents an error that Client is stopped.
	ErrClientStopped = errors.New("client stopped")

	// Represents an error of empty dialer array.
	ErrClientInvalidPoolDialers = errors.New("invalid dialers: empty array")
)

// ------------------
//   Message error
// ------------------

var (
	// Invalid message CMD.
	ErrInvalidRspMessage = errors.New("invalid response message cmd")

	// Method not found.
	ErrMethodNotFound = errors.New("method not found")

	// Invlaid flag bit index.
	ErrInvalidFlagBitIndex = errors.New("invalid index, should be 0-7")
)

// ------------------
//   Context error
// ------------------

var (
	ErrContextResponseToNotify = errors.New("should not response to a context with notify message")
)

// ------------------
//   General errors
// ------------------

var (
	ErrTimeout = errors.New("timeout")
)
