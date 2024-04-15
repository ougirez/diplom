package constants

import (
	"errors"
	"net/http"
)

// CodedError is an error wrapper which wraps errors with http status codes.
type CodedError struct {
	err  error
	code int
}

func (ce *CodedError) Error() string {
	return ce.err.Error()
}

func (ce *CodedError) Code() int {
	return ce.code
}

func NewCodedError(msg string, code int) *CodedError {
	return &CodedError{errors.New(msg), http.StatusNotFound}
}

var (
	// Unathorized
	ErrUnauthorized      = &CodedError{errors.New("unauthorized"), http.StatusUnauthorized}
	ErrMissingAuthCookie = &CodedError{errors.New("missing authorization cookie"), http.StatusUnauthorized}

	ErrPasswordMismatch = &CodedError{errors.New("password mismatch"), http.StatusUnauthorized}

	ErrAuthTokenInvalid        = &CodedError{errors.New("authorization token is invalid"), http.StatusUnauthorized}
	ErrUnexpectedSigningMethod = &CodedError{errors.New("unexpected signing method"), http.StatusUnauthorized}
	ErrHashInvalid             = &CodedError{errors.New("hash is invalid"), http.StatusUnauthorized}

	// Forbidden
	ErrAuthTokenExpired = &CodedError{errors.New("authorization token is expired"), http.StatusForbidden}

	// Bad Request
	ErrBindRequest     = &CodedError{errors.New("failed to bind request"), http.StatusBadRequest}
	ErrValidateRequest = &CodedError{errors.New("failed to validate request"), http.StatusBadRequest}
	ErrDBNotFound      = &CodedError{errors.New("not found in the database"), http.StatusBadRequest}
	ErrAlreadyExists   = &CodedError{errors.New("already exists"), http.StatusBadRequest}
	ErrNegativeDebit   = &CodedError{errors.New("non-positive debit amount for refill"), http.StatusBadRequest}
	ErrNegativeCredit  = &CodedError{errors.New("non-positive credit amount for withdraw"), http.StatusBadRequest}
	ErrAccountNotFound = &CodedError{errors.New("account not found with such a number"), http.StatusBadRequest}
	ErrNotEnoughMoney  = &CodedError{errors.New("not enough money"), http.StatusBadRequest}
	ErrBadRequest      = &CodedError{errors.New("bad request"), http.StatusBadRequest}

	// Internal
	ErrSignToken      = &CodedError{errors.New("failed to sign token"), http.StatusInternalServerError}
	ErrGenerateUUID   = &CodedError{errors.New("failed to generate UUID"), http.StatusInternalServerError}
	ErrParseAuthToken = &CodedError{errors.New("failed to parse authorization token"), http.StatusInternalServerError}

	// Conflict
	ErrEmailAlreadyTaken = &CodedError{errors.New("email is taken already by other user"), http.StatusConflict}
	AccountAlreadyExists = &CodedError{errors.New("account with this currency is already exists for this user"), http.StatusConflict}
)
