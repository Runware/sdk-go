package runware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	
	"github.com/gorilla/websocket"
)

type ConnAddr string

func (c ConnAddr) String() string {
	return string(c)
}

const (
	DevEnv  ConnAddr = ""
	ProdEnv ConnAddr = "wss://ws-api.runware.ai/v1/"
)

const (
	pingInterval        = 15 * time.Second
	pongWait            = 30 * time.Second
	timeoutSendResponse = 30 // In sec
)

type Runware interface {
	Connected() bool
	Close() error
	Send([]byte) error
	SendAndResponse(ctx context.Context, req Request) (*Response, error)
}

type runware struct {
	apiKey           string
	sessionKey       string
	connStr          ConnAddr
	client           *websocket.Conn
	incomingMessages chan []byte
	
	reconnectAttempt int
	reconnectChan    chan struct{}
}

func (sdk *runware) Connected() bool {
	if sdk.client == nil {
		return false
	}
	
	if sdk.sessionKey == "" {
		return false
	}
	
	return true
}

// Close connection to socket
func (sdk *runware) Close() error {
	return sdk.client.Close()
}

// Send socket message
func (sdk *runware) Send(msg []byte) error {
	if msg == nil {
		return ErrOutgoingIsNil
	}
	
	return sdk.client.WriteMessage(websocket.TextMessage, msg)
}

type Request struct {
	Event         string
	ResponseEvent string
	Count         int
	Data          interface{}
}

func (req Request) MarshalBinary() ([]byte, error) {
	return json.Marshal(req)
}

func (req Request) ToSDKEvent() ([]byte, error) {
	reqM := map[string]interface{}{
		req.Event: req.Data,
	}
	return json.Marshal(reqM)
}

type Response struct {
	Event string
	Data  []byte
}

func (sdk *runware) SendAndResponse(ctx context.Context, req Request) (*Response, error) {
	responseChan := make(chan Response)
	errChan := make(chan error)
	
	go func() {
		for msg := range sdk.incomingMessages {
			var msgData map[string]interface{}
			if err := json.Unmarshal(msg, &msgData); err != nil {
				errChan <- fmt.Errorf("%w:[%s]", ErrDecodeMessage, err.Error())
				return
			}
			
			// Check if is an error message first
			if errMsg, ok := sdk.handleSendAndResponseError(msgData); ok {
				errChan <- errMsg
				return
			}
			
			for k, v := range msgData {
				if !incomingEventExist(k) {
					errChan <- fmt.Errorf("%w:[%s]", ErrUnknownEvent, k)
					return
				}
				
				// Skip is current event is not the one from request
				if k != req.ResponseEvent {
					log.Println("Skipping event", k, "Currently handling", req.ResponseEvent)
					continue
				}
				
				bValue, err := interfaceToByte(v)
				if err != nil {
					errChan <- err
					return
				}
				
				res := Response{
					Event: k,
					Data:  bValue,
				}
				
				responseChan <- res
				return
			}
		}
	}()
	
	bSendReq, err := req.ToSDKEvent()
	if err != nil {
		return nil, err
	}
	
	if err = sdk.Send(bSendReq); err != nil {
		return nil, err
	}
	
	select {
	case resp := <-responseChan:
		return &resp, nil
	case err = <-errChan:
		return nil, err
	case <-time.After(timeoutSendResponse * time.Second):
		return nil, fmt.Errorf("%w:[%s]", ErrRequestTimeout, req.Event)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// readLoop incoming message monitoring
func (sdk *runware) readLoop() {
	defer func() {
		_ = sdk.Close()
	}()
	
	_ = sdk.client.SetReadDeadline(time.Now().Add(pongWait))
	sdk.client.SetPongHandler(func(string) error {
		_ = sdk.client.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	
	for {
		_, msg, err := sdk.client.ReadMessage()
		if err != nil {
			ok := websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure)
			if ok {
				log.Println("Abnormal close", err)
				sdk.reconnectAttempt = 3
				sdk.reconnectChan <- struct{}{}
			} else {
				log.Println("Error reading message", err)
				sdk.reconnectAttempt = 1
				sdk.reconnectChan <- struct{}{}
			}
			break
		}
		
		sdk.incomingMessages <- msg
	}
}

func (sdk *runware) handleSendAndResponseError(msg map[string]interface{}) (error, bool) {
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
			}
		}
	}
	
	return err, hasError
}

func (sdk *runware) reconnectLoop() {
	for {
		select {
		case <-sdk.reconnectChan:
			log.Println("Reconnecting to runware...")
			
			for i := 0; i < sdk.reconnectAttempt; i++ {
				var err error
				_ = sdk.Close()
				sdk.client, err = wsConnect(sdk.connStr.String())
				if err != nil {
					log.Printf("Cannot recover: %s\n", err.Error())
					sdk.reconnectAttempt = 0
					return
				}
				
				// Restart reading loop
				go sdk.readLoop()
				go sdk.reconnectLoop()
				
				connRes, err := sdk.Connect(context.Background(), NewConnectReq{
					APIKey:                sdk.apiKey,
					ConnectionSessionUUID: sdk.sessionKey,
				})
				if err != nil {
					log.Printf("Cannot reconnect: %s\n", err.Error())
					return
				}
				
				if connRes != nil {
					log.Printf("Reconnected to runware: %s\n", sdk.sessionKey)
					sdk.reconnectAttempt = 0
					return
				}
				
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// New create a new client and initiate connection
func New(cfg Config) (Runware, error) {
	
	if cfg.APIKey == "" {
		return nil, ErrApiKeyRequired
	}
	
	if cfg.ConnAddr == "" {
		cfg.ConnAddr = ProdEnv
	}
	
	client, err := wsConnect(cfg.ConnAddr.String())
	if err != nil {
		return nil, fmt.Errorf("%w:[%s]", ErrWsDial, cfg.ConnAddr.String())
	}
	
	r := &runware{
		apiKey:           cfg.APIKey,
		connStr:          cfg.ConnAddr,
		client:           client,
		incomingMessages: make(chan []byte),
		reconnectChan:    make(chan struct{}),
	}
	
	go r.readLoop()
	go r.reconnectLoop()
	
	// Attempt initial connect
	conRes, err := r.Connect(context.Background(), NewConnectReq{
		APIKey: cfg.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("%w:[%s]", ErrWsDial, err.Error())
	}
	r.sessionKey = conRes.ConnectionSessionUUID
	
	return r, nil
}

func interfaceToByte(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	default:
		return json.Marshal(v)
	}
}

func wsConnect(connStr string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(connStr, nil)
	return conn, err
}
