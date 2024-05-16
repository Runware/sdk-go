package runware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	
	"github.com/google/uuid"
)

type NewReverseImageClipReq struct {
	ImageUUID string `json:"imageUUID"`
	TaskUUID  string `json:"taskUUID"`
}

type NewReverseImageClipResp struct {
	Texts    []Text `json:"texts"`
	TimedOut bool   `json:"timedOut"`
}

func (sdk *SDK) ImageToText(ctx context.Context, req NewReverseImageClipReq) (*NewReverseImageClipResp, error) {
	req = *mergeNewReverseImageClipReqDefaults(&req)
	if err := validateNewReverseImageClipReq(req); err != nil {
		return nil, err
	}
	
	sendReq := Request{
		ID:            uuid.New().String(),
		Event:         NewReverseImageClip,
		ResponseEvent: NewReverseClip,
		Data:          req,
	}
	
	newReverseImageClipResp := &NewReverseImageClipResp{}
	
	responseChan := make(chan *NewReverseImageClipResp)
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
				
				err = json.Unmarshal(bValue, &newReverseImageClipResp)
				if err != nil {
					errChan <- err
					return
				}
				
				responseChan <- newReverseImageClipResp
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
		newReverseImageClipResp.TimedOut = true
		return newReverseImageClipResp, fmt.Errorf("%w:[%s]", ErrRequestTimeout, sendReq.Event)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	
}

func NewReverseImageClipReqDefaults() *NewReverseImageClipReq {
	return &NewReverseImageClipReq{
		TaskUUID: uuid.New().String(),
	}
}

func mergeNewReverseImageClipReqDefaults(req *NewReverseImageClipReq) *NewReverseImageClipReq {
	_ = MergeEventRequestsWithDefaults[*NewReverseImageClipReq](req, NewReverseImageClipReqDefaults())
	return req
}

func validateNewReverseImageClipReq(req NewReverseImageClipReq) error {
	if req.ImageUUID == "" {
		return fmt.Errorf("%w:[%s]", ErrFieldRequired, "imageUUID")
	}
	return nil
}
