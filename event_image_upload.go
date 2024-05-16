package runware

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
	
	"github.com/google/uuid"
)

type NewImageUploadReq struct {
	ImageBase64 string `json:"imageBase64"`
	TaskUUID    string `json:"taskUUID"`
}

type NewImageUploadResp struct {
	NewImageSrc  string `json:"newImageSrc"`
	NewImageUUID string `json:"newImageUUID"` // Pointer to handle null values
	TaskUUID     string `json:"taskUUID"`
	TimedOut     bool   `json:"timedOut"`
}

func (sdk *SDK) ImageUpload(ctx context.Context, req NewImageUploadReq) (*NewImageUploadResp, error) {
	req = *mergeNewControlNetsReqDefaults(&req)
	if err := validateNewImageUploadReq(req); err != nil {
		return nil, err
	}
	
	sendReq := Request{
		ID:            uuid.New().String(),
		Event:         NewImageUpload,
		ResponseEvent: NewUploadedImageUUID,
		Data:          req,
	}
	
	newImageUploadResp := &NewImageUploadResp{}
	
	responseChan := make(chan *NewImageUploadResp)
	errChan := make(chan error)
	
	go func() {
		defer close(responseChan)
		defer close(errChan)
		
		for msg := range sdk.Client.Listen() {
			var msgData map[string]interface{}
			if err := json.Unmarshal(msg, &msgData); err != nil {
				errChan <- fmt.Errorf("%w:[%s]", ErrDecodeMessage, err.Error())
				return
			}
			
			// Check if is an error message first
			if errMsg, ok := sdk.OnError(msgData); ok {
				errChan <- errMsg
				return
			}
			
			for k, v := range msgData {
				if k != sendReq.ResponseEvent {
					log.Println("Skipping event", k, "Currently handling", sendReq.ResponseEvent)
					continue
				}
				
				bValue, err := interfaceToByte(v)
				if err != nil {
					errChan <- err
					return
				}
				
				err = json.Unmarshal(bValue, &newImageUploadResp)
				if err != nil {
					errChan <- err
					return
				}
				
				responseChan <- newImageUploadResp
				return
			}
		}
		
	}()
	
	bSendReq, err := sendReq.ToEvent()
	if err != nil {
		return nil, err
	}
	
	if err = sdk.Client.Send(bSendReq); err != nil {
		return nil, err
	}
	
	select {
	case resp := <-responseChan:
		return resp, nil
	case err = <-errChan:
		return nil, err
	case <-time.After(timeoutSendResponse * time.Second):
		newImageUploadResp.TimedOut = true
		return newImageUploadResp, fmt.Errorf("%w:[%s]", ErrRequestTimeout, sendReq.Event)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func NewImageUploadReqDefaults() *NewImageUploadReq {
	return &NewImageUploadReq{
		TaskUUID: uuid.New().String(),
	}
}

func mergeNewControlNetsReqDefaults(req *NewImageUploadReq) *NewImageUploadReq {
	_ = MergeEventRequestsWithDefaults[*NewImageUploadReq](req, NewImageUploadReqDefaults())
	return req
}

func validateNewImageUploadReq(req NewImageUploadReq) error {
	if req.ImageBase64 == "" {
		return fmt.Errorf("%w:[%s]", ErrFieldRequired, "imageBase64")
	}
	_, err := isValidBase64Image(req.ImageBase64)
	if err != nil {
		return fmt.Errorf("%w:[%s]", err, "imageBase64")
	}
	return nil
}

func isValidBase64Image(v string) (bool, error) {
	
	// Check for Data URI scheme (e.g., "data:image/png;base64,")
	if strings.HasPrefix(v, "data:image") {
		// Extract base64 data after the comma
		commaIndex := strings.Index(v, ",")
		if commaIndex == -1 {
			return false, ErrImageWrongSchema
		}
		v = v[commaIndex+1:]
	}
	
	// Decode string to base64
	decoded, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return false, ErrImageIsNotBase64
	}
	
	// Validate image format
	reader := bytes.NewReader(decoded)
	if _, err = decodeImage(reader); err != nil {
		return false, err
	}
	
	return true, nil
}

func decodeImage(reader io.Reader) (string, error) {
	var (
		format       = ""
		err    error = nil
	)
	
	// Read the first few bytes for format detection
	header := make([]byte, 8)
	n, err := reader.Read(header)
	if err != nil {
		return format, fmt.Errorf("%w: [%s]", ErrImageHeader, err.Error())
	}
	if n < 8 {
		return format, fmt.Errorf("%w: [%s]", ErrImageHeader, "insufficient image data")
	}
	
	// TODO: Uncomment the rest as they become supported
	switch {
	case bytes.HasPrefix(header, []byte{0x52, 0x49, 0x46, 0x46}) &&
		bytes.Contains(header, []byte("WEBPVP8")): // WEBP
		format = "webp"
	case bytes.HasPrefix(header, []byte{0xFF, 0xD8, 0xFF}): // JPEG
		format = "jpeg"
	case bytes.HasPrefix(header, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}): // PNG
		format = "png"
	
	// case bytes.HasPrefix(header, []byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}) ||
	// 	bytes.HasPrefix(header, []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}): // GIF
	// 	format = "gif"
	// case bytes.HasPrefix(header, []byte{0x42, 0x4D}): // BMP
	// 	format = "bmp"
	// case bytes.HasPrefix(header, []byte{0x49, 0x49, 0x2A, 0x00}) ||
	// 	bytes.HasPrefix(header, []byte{0x4D, 0x4D, 0x00, 0x2A}): // TIFF
	//
	// 	format = "tiff"
	default:
		
		return format, ErrImageUnsupported
	}
	
	// if err != nil {
	// 	return nil, "", fmt.Errorf("%w:[%s]", ErrImageDecode, err.Error())
	// }
	
	return format, nil
}
