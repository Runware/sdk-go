package runware

import (
	"errors"
)

var (
	ErrWsDial         = errors.New("cannot connect to ws")
	ErrApiKeyRequired = errors.New("api key is required")
	ErrHandlerExists  = errors.New("handler already exists")
	ErrOutgoingIsNil  = errors.New("outgoing message cannot be nil")
)
