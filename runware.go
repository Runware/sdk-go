package runware

import (
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
	pongWait            = 5 * time.Second
	pingInterval        = (pongWait * 9) / 10
	timeoutSendResponse = 30 // In sec
)

type Runware interface {
	APIKey() string
	Connected() bool
	Close() error
	Send([]byte) error
	Listen() chan []byte
	Reconnected() chan struct{}
}

type runware struct {
	apiKey           string
	sessionKey       string
	connStr          ConnAddr
	client           *websocket.Conn
	incomingMessages chan []byte
	
	reconnectAttempt int
	reconnectChan    chan struct{}
	reconnectedChan  chan struct{}
}

func (r *runware) APIKey() string {
	return r.apiKey
}

func (r *runware) Connected() bool {
	if r.client == nil {
		return false
	}
	
	if r.sessionKey == "" {
		return false
	}
	
	return true
}

// Close connection to socket
func (r *runware) Close() error {
	return r.client.Close()
}

// Send socket message
func (r *runware) Send(msg []byte) error {
	if msg == nil {
		return ErrOutgoingIsNil
	}
	
	return r.client.WriteMessage(websocket.TextMessage, msg)
}

func (r *runware) Listen() chan []byte {
	return r.incomingMessages
}

func (r *runware) Reconnected() chan struct{} {
	return r.reconnectedChan
}

func (r *runware) handleSendAndResponseError(msg map[string]interface{}) (error, bool) {
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

func (r *runware) heartbeatLoop() {
	
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			log.Println("Ping ...")
			if err := r.Send([]byte(`{"ping": true}`)); err != nil {
				log.Println("Ping err", err)
				r.reconnectAttempt = 1
				r.reconnectChan <- struct{}{}
			}
		}
	}
}

// readLoop incoming message monitoring
func (r *runware) readLoop() {
	defer func() {
		_ = r.Close()
	}()
	
	// TODO: This deadline causes unexpected connection close
	// _ = r.client.SetReadDeadline(time.Now().Add(pongWait))
	// r.client.SetPongHandler(func(string) error {
	// 	_ = r.client.SetReadDeadline(time.Now().Add(pongWait))
	// 	return nil
	// })
	
	for {
		_, msg, err := r.client.ReadMessage()
		if err != nil {
			ok := websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure)
			if ok {
				log.Println("Abnormal close", err)
				r.reconnectAttempt = 3
				r.reconnectChan <- struct{}{}
			} else {
				log.Println("Error reading message", err)
				r.reconnectAttempt = 1
				r.reconnectChan <- struct{}{}
			}
			break
		}
		
		var msgData map[string]interface{}
		_ = json.Unmarshal(msg, &msgData)
		if _, ok := msgData[Pong]; ok {
			log.Println("Pong", msgData[Pong])
			continue
		}
		
		fmt.Printf("[readLoop]: %+v\n", string(msg))
		
		r.incomingMessages <- msg
	}
	
	fmt.Print("[readLoop] closed")
}

// reconnectLoop monitor and attempts to reconnect
func (r *runware) reconnectLoop() {
	for {
		select {
		case <-r.reconnectChan:
			log.Println("Reconnecting to runware...")
			
			for i := 0; i < r.reconnectAttempt; i++ {
				var err error
				_ = r.Close()
				r.client, err = wsConnect(r.connStr.String())
				if err != nil {
					log.Printf("Reconnect attempt %d failed: %s\n", i+1, err.Error())
					time.Sleep(5 * time.Second)
					continue
				}
				
				// Restart loops
				go r.readLoop()
				go r.reconnectLoop()
				
				r.reconnectedChan <- struct{}{}
				fmt.Printf("Attempt: %d\n", i+1)
				return
			}
			
			log.Println("Reconnection failed after 3 attempts. Aborted")
			return
		}
	}
}

// New create a new client and initiate connection
func New(cfg RunwareConfig) (Runware, error) {
	
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
		reconnectedChan:  make(chan struct{}),
	}
	
	go r.readLoop()
	go r.reconnectLoop()
	if cfg.KeepAlive {
		go r.heartbeatLoop()
		
	}
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
