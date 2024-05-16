package runware

import (
	"errors"
)

var (
	ErrWsDial         = errors.New("cannot connect to ws")
	ErrApiKeyRequired = errors.New("api key is required")
	ErrOutgoingIsNil  = errors.New("outgoing message cannot be nil")
	ErrFieldRequired  = errors.New("field is required")
	ErrWsUnknownError = errors.New("unknown error")
	ErrInvalidApiKey  = errors.New("invalid api key")
	ErrRequestTimeout = errors.New("request timeout")
	ErrDecodeMessage  = errors.New("cannot decode message")
)

// Base64 Err validations
var (
	ErrImageWrongSchema = errors.New("image scheme is wrong")
	ErrImageIsNotBase64 = errors.New("image is not base64")
	ErrImageUnsupported = errors.New("unsupported image format")
	ErrImageHeader      = errors.New("image header is invalid")
	ErrImageDecode      = errors.New("image decode error")
)
