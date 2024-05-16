package runware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

type SDK struct {
	Client Runware
	
	sessionKey string
}

func NewSDK(cfg SDKConfig) (*SDK, error) {
	
	client, err := makeClient(cfg)
	if err != nil {
		return nil, err
	}
	
	sdk := &SDK{
		Client: client,
	}
	
	res, err := sdk.Connect(context.Background(), NewConnectReq{
		APIKey: sdk.Client.APIKey(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w:[%s]", ErrWsDial, err.Error())
	}
	
	sdk.sessionKey = res.ConnectionSessionUUID
	
	log.Println("Connected", sdk.sessionKey)
	
	// Start reconnection monitor
	go sdk.onReconnected()
	
	return sdk, nil
}

func (sdk *SDK) OnError(msg map[string]interface{}) (error, bool) {
	var (
		hasError       = false
		err      error = nil
	)
	
	for k, v := range msg {
		if k == "error" {
			if v == true {
				hasError = true
			}
		}
		if k == "errorId" {
			switch v {
			case float64(19):
				err = ErrInvalidApiKey
				// Add more
			default:
				err = ErrWsUnknownError
			}
		}
	}
	
	return fmt.Errorf("%w:[%v:%s]", err, msg["errorId"], msg["errorMessage"]), hasError
}

func (sdk *SDK) onReconnected() {
	for {
		select {
		case <-sdk.Client.Reconnected():
			_, err := sdk.Connect(context.Background(), NewConnectReq{
				APIKey:                sdk.Client.APIKey(),
				ConnectionSessionUUID: sdk.sessionKey,
			})
			if err != nil {
				log.Println("Reconnect failed:", err)
			}
		}
	}
}

type Request struct {
	ID            string
	Event         string
	ResponseEvent string
	Count         int
	Data          interface{}
}

func (req Request) ToEvent() ([]byte, error) {
	reqM := map[string]interface{}{
		req.Event: req.Data,
	}
	return json.Marshal(reqM)
}

func makeClient(cfg SDKConfig) (Runware, error) {
	if cfg.Client != nil {
		return cfg.Client, nil
	}
	
	client, err := New(RunwareConfig{
		APIKey:    cfg.APIKey,
		ConnAddr:  cfg.ConnAddr,
		KeepAlive: false,
	})
	if err != nil {
		return nil, err
	}
	
	return client, nil
}
