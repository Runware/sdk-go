package runware

import (
	"context"
	"encoding/json"
)

type NewConnectReq struct {
	APIKey                string `json:"apiKey"`
	ConnectionSessionUUID string `json:"connectionSessionUUID,omitempty"`
}

type NewConnectResp struct {
	ConnectionSessionUUID string `json:"connectionSessionUUID"`
}

func (sdk *runware) Connect(ctx context.Context, req NewConnectReq) (*NewConnectResp, error) {
	sendReq := Request{
		Event:         NewConnection,
		ResponseEvent: NewConnectionSessionUUID,
		Data:          req,
	}
	res, err := sdk.SendAndResponse(ctx, sendReq)
	if err != nil {
		return nil, err
	}
	
	var newConnectResp *NewConnectResp
	err = json.Unmarshal(res.Data, &newConnectResp)
	if err != nil {
		return nil, err
	}
	
	// Set connection session key
	sdk.sessionKey = newConnectResp.ConnectionSessionUUID
	
	return newConnectResp, nil
}
