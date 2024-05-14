package runware

import (
	"errors"
)

var (
	ErrWsDial         = errors.New("cannot connect to ws")
	ErrApiKeyRequired = errors.New("api key is required")
	ErrWsNotConnected = errors.New("websocket not connected")
	ErrOutgoingIsNil  = errors.New("outgoing message cannot be nil")
	ErrPromptRequired = errors.New("prompt is required")
	ErrInvalidApiKey  = errors.New("invalid api key")
	ErrRequestTimeout = errors.New("request timeout")
	ErrUnknownEvent   = errors.New("unknown event")
	ErrDecodeMessage  = errors.New("cannot decode message")
)
