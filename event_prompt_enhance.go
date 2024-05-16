package runware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	
	"github.com/google/uuid"
)

type NewPromptEnhanceReq struct {
	TaskUUID         string `json:"taskUUID"`
	PromptText       string `json:"prompt"` // PromptText json should be `promptText`
	PromptMaxLength  int    `json:"promptMaxLength"`
	PromptVersions   int    `json:"promptVersions"`
	PromptLanguageId int    `json:"promptLanguageId"`
}

type NewPromptEnhanceRes struct {
	Texts    []Text `json:"texts"`
	TimedOut bool   `json:"timedOut"`
}

func (sdk *SDK) PromptEnhancer(ctx context.Context, req NewPromptEnhanceReq) (*NewPromptEnhanceRes, error) {
	req = *mergeNewPromptEnhanceReqDefaults(&req)
	if err := validateNewPromptEnhanceReq(req); err != nil {
		return nil, err
	}
	
	sendReq := Request{
		ID:            uuid.New().String(),
		Event:         NewPromptEnhance,
		ResponseEvent: NewPromptEnhancer,
		Data:          req,
	}
	
	newPromptEnhanceRes := &NewPromptEnhanceRes{}
	
	responseChan := make(chan *NewPromptEnhanceRes)
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
				
				err = json.Unmarshal(bValue, &newPromptEnhanceRes)
				if err != nil {
					errChan <- err
					return
				}
				
				responseChan <- newPromptEnhanceRes
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
		newPromptEnhanceRes.TimedOut = true
		return newPromptEnhanceRes, fmt.Errorf("%w:[%s]", ErrRequestTimeout, sendReq.Event)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func NewPromptEnhanceReqDefaults() *NewPromptEnhanceReq {
	return &NewPromptEnhanceReq{
		TaskUUID:         uuid.New().String(),
		PromptLanguageId: 1,
		PromptVersions:   3,
	}
}

func mergeNewPromptEnhanceReqDefaults(req *NewPromptEnhanceReq) *NewPromptEnhanceReq {
	_ = MergeEventRequestsWithDefaults[*NewPromptEnhanceReq](req, NewPromptEnhanceReqDefaults())
	return req
}

func validateNewPromptEnhanceReq(req NewPromptEnhanceReq) error {
	if req.PromptMaxLength < 1 || req.PromptVersions > 380 {
		return fmt.Errorf("%w:[%s][1-380]", ErrFieldIncorrectVal, "promptMaxLength")
	}
	
	if req.PromptVersions < 1 || req.PromptVersions > 5 {
		return fmt.Errorf("%w:[%s][1-380]", ErrFieldIncorrectVal, "promptVersions")
	}
	
	return nil
}
