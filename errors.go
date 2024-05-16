package runware

import (
	"errors"
)

var (
	ErrWsDial            = errors.New("cannot connect to ws")
	ErrApiKeyRequired    = errors.New("api key is required")
	ErrOutgoingIsNil     = errors.New("outgoing message cannot be nil")
	ErrFieldRequired     = errors.New("field is required")
	ErrWsUnknownError    = errors.New("unknown error")
	ErrInvalidApiKey     = errors.New("invalid api key")
	ErrRequestTimeout    = errors.New("request timeout")
	ErrDecodeMessage     = errors.New("cannot decode message")
	ErrGuidImageRequired = errors.New("guid image is required")
)
