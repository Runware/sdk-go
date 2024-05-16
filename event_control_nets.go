package runware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	
	"github.com/google/uuid"
)

type NewControlNetsReq PreProcessControlNet

type NewControlNetsResp struct {
	NewImageSrc   string `json:"newImageSrc"`
	NewImageUUID  string `json:"newImageUUID"`
	InitImageUUID string `json:"initImageUUID"`
	NNsfwContent  *bool  `json:"nNsfwContent"` // Pointer to handle null values
	TaskUUID      string `json:"taskUUID"`
	TimedOut      bool   `json:"timedOut"`
}

func (sdk *SDK) NewControlNets(ctx context.Context, req NewControlNetsReq) (*NewControlNetsResp, error) {
	req = *mergeControlNetsReqWithDefaults(&req)
	if err := validateNewControlNetsReq(req); err != nil {
		return nil, err
	}
	
	sendReq := Request{
		Event:         NewPreProcessControlNet,
		ResponseEvent: NewPreProcessControlNet,
		Data:          req,
	}
	
	newControlNetsResp := &NewControlNetsResp{}
	
	responseChan := make(chan *NewControlNetsResp)
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
				
				err = json.Unmarshal(bValue, &newControlNetsResp)
				if err != nil {
					errChan <- err
					return
				}
				
				responseChan <- newControlNetsResp
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
		newControlNetsResp.TimedOut = true
		return newControlNetsResp, fmt.Errorf("%w:[%s]", ErrRequestTimeout, sendReq.Event)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func NewControlNetsReqDefaults() *NewControlNetsReq {
	return &NewControlNetsReq{
		TaskUUID:           uuid.New().String(),
		TaskType:           ControlNetPreprocessImage,
		LowThresholdCanny:  100,
		HighThresholdCanny: 200,
	}
}

func mergeControlNetsReqWithDefaults(req *NewControlNetsReq) *NewControlNetsReq {
	_ = MergeEventRequestsWithDefaults[*NewControlNetsReq](req, NewControlNetsReqDefaults())
	return req
}

func validateNewControlNetsReq(req NewControlNetsReq) error {
	if req.GuideImageUUID == "" {
		return fmt.Errorf("%w:[%s]", ErrFieldRequired, "guideImageUUID")
	}
	return nil
}
