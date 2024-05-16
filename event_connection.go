package runware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	
	"github.com/google/uuid"
)

type NewConnectReq struct {
	APIKey                string `json:"apiKey"`
	ConnectionSessionUUID string `json:"connectionSessionUUID,omitempty"`
}

type NewConnectResp struct {
	ConnectionSessionUUID string `json:"connectionSessionUUID"`
}

func (sdk *SDK) Connect(ctx context.Context, req NewConnectReq) (*NewConnectResp, error) {
	sendReq := Request{
		ID:            uuid.New().String(),
		Event:         NewConnection,
		ResponseEvent: NewConnectionSessionUUID,
		Data:          req,
	}
	
	responseChan := make(chan *NewConnectResp)
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
				
				var newConnectResp *NewConnectResp
				err = json.Unmarshal(bValue, &newConnectResp)
				if err != nil {
					errChan <- err
					return
				}
				
				responseChan <- newConnectResp
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
		return nil, fmt.Errorf("%w:[%s]", ErrRequestTimeout, sendReq.Event)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
