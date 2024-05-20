package runware

import (
	"bytes"
	"errors"
	"io"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ImageUploadSuite struct {
	suite.Suite
	originalDecodeImage func(io.Reader) (string, error)
}

// Mock function for testing
func (s *ImageUploadSuite) mockDecodeImage(reader io.Reader) (string, error) {
	header := make([]byte, 8)
	n, err := reader.Read(header)
	if err != nil {
		return "", err
	}
	if n < 8 {
		return "", errors.New("insufficient image data")
	}
	
	switch {
	case bytes.HasPrefix(header, []byte{0x52, 0x49, 0x46, 0x46}): // WEBP
		return "webp", nil
	case bytes.HasPrefix(header, []byte{0xFF, 0xD8, 0xFF}): // JPEG
		return "jpeg", nil
	case bytes.HasPrefix(header, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}): // PNG
		return "png", nil
	default:
		return "", ErrImageUnsupported
	}
}

// SetupTest method is run before each test
func (s *ImageUploadSuite) SetupTest() {
	s.originalDecodeImage = decodeImage
}

// TearDownTest method is run after each test
func (s *ImageUploadSuite) TearDownTest() {

}

func (s *ImageUploadSuite) TestValidateNewImageUploadReq() {
	testCases := []struct {
		name    string
		req     NewImageUploadReq
		wantErr error
	}{
		{
			name:    "EmptyImageBase64",
			req:     NewImageUploadReq{},
			wantErr: ErrFieldRequired,
		},
		{
			name: "InvalidBase64",
			req: NewImageUploadReq{
				ImageBase64: "invalid-base64-string",
			},
			wantErr: ErrImageIsNotBase64,
		},
		{
			name: "InvalidDataURISchema",
			req: NewImageUploadReq{
				ImageBase64: "data:image/jpeg",
			},
			wantErr: ErrImageWrongSchema,
		},
		{
			name: "ValidJPEG",
			req: NewImageUploadReq{
				ImageBase64: "data:image/jpeg;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=",
			},
			wantErr: nil,
		},
	}
	
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := validateNewImageUploadReq(tc.req)
			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr), "Error should wrap the expected error")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *ImageUploadSuite) TestIsValidBase64Image() {
	testCases := []struct {
		name    string
		input   string
		want    bool
		wantErr error
	}{
		{
			name:  "ValidBase64JPEG",
			input: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=",
			want:  true,
		},
	}
	
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			got, err := isValidBase64Image(tc.input)
			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr), "Error should wrap the expected error")
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

// Run the test suite
func TestImageUploadSuite(t *testing.T) {
	suite.Run(t, new(ImageUploadSuite))
}
