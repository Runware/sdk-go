package runware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	
	"github.com/google/uuid"
)

type NewUpscaleGanReq struct {
	TaskUUID      string `json:"taskUUID"`
	ImageUUID     string `json:"imageUUID"`
	UpscaleFactor int    `json:"upscaleFactor"`
}

type NewUpscaleGanResp struct {
	Images   []Image `json:"images"`
	TimedOut bool    `json:"timedOut"`
}

func (sdk *SDK) ImageUpscale(ctx context.Context, req NewUpscaleGanReq) (*NewUpscaleGanResp, error) {
	req = *mergeNewUpscaleGanReqWithDefaults(&req)
	if err := validateNewUpscaleGanReq(req); err != nil {
		return nil, err
	}
	
	sendReq := Request{
		ID:            uuid.New().String(),
		Event:         NewUpscaleGan,
		ResponseEvent: NewUpscaleGan,
		Data:          req,
	}
	
	newUpscaleGanResp := &NewUpscaleGanResp{}
	
	responseChan := make(chan *NewUpscaleGanResp)
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
				
				err = json.Unmarshal(bValue, &newUpscaleGanResp)
				if err != nil {
					errChan <- err
					return
				}
				
				responseChan <- newUpscaleGanResp
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
		newUpscaleGanResp.TimedOut = true
		return newUpscaleGanResp, fmt.Errorf("%w:[%s]", ErrRequestTimeout, sendReq.Event)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func NewUpscaleGanReqDefaults() *NewUpscaleGanReq {
	return &NewUpscaleGanReq{
		TaskUUID: uuid.New().String(),
	}
}

func mergeNewUpscaleGanReqWithDefaults(req *NewUpscaleGanReq) *NewUpscaleGanReq {
	_ = MergeEventRequestsWithDefaults[*NewUpscaleGanReq](req, NewUpscaleGanReqDefaults())
	return req
}

func validateNewUpscaleGanReq(req NewUpscaleGanReq) error {
	if req.ImageUUID == "" {
		return fmt.Errorf("%w:[%s]", ErrFieldRequired, "imageUUID")
	}
	
	if req.UpscaleFactor == 0 {
		return fmt.Errorf("%w:[%s]", ErrFieldRequired, "upscaleFactor")
	}
	
	return nil
}
