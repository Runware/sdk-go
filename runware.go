package runware

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"
	
	"github.com/gorilla/websocket"
)

type ConnAddr string

const (
	DevEnv  ConnAddr = "wss://dev-ws-api.diffusionmaster.com/v1/"
	ProdEnv ConnAddr = "wss://ws-api.diffusionmaster.com/v1/"
)

const (
	pingInterval = 15 * time.Second
	pongTimeout  = 10 * time.Second
)

type Runware struct {
	apiKey          string
	connStr         ConnAddr
	client          *websocket.Conn
	messageHandlers map[string]IncomingMessageHandler
	sendQueue       chan []byte
}

// Connect shorthand function for sending a newConnect message
func (sdk *Runware) Connect() {
	panic("Not implemented!")
}

// Reconnect shorthand function for sending a newConnect message with reconnect session key
func (sdk *Runware) Reconnect() {
	panic("Not implemented!")
}

// Close connection to socket
func (sdk *Runware) Close() error {
	return sdk.client.Close()
}

func (sdk *Runware) Listen() {
	pingTicker := time.NewTicker(pingInterval)
	done := make(chan struct{})
	
	// Write loop
	go func() {
		for msg := range sdk.sendQueue {
			err := sdk.client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				// TODO: Improve write error handling
				log.Println("Error writing message", err)
			}
		}
	}()
	
	// Read loop
	go func() {
		for {
			_, msg, err := sdk.client.ReadMessage()
			if err != nil {
				close(done)
				log.Println("Error reading message:", err)
				return
			}
			sdk.handleIncomingMessage(msg)
		}
	}()
	
	// Heartbeat loop
	go func() {
		for {
			select {
			case <-pingTicker.C:
				err := sdk.client.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(pongTimeout))
				if err != nil {
					close(done)
					return
				}
			case <-done:
				pingTicker.Stop() // Stop the ticker
				return
			}
		}
	}()
}

// Send socket message
func (sdk *Runware) Send(msg OutgoingMessageHandler) error {
	if msg == nil {
		return ErrOutgoingIsNil
	}
	
	bMsg, err := msg.MarshalBinary()
	if err != nil {
		return err
	}
	
	sdk.sendQueue <- bMsg
	
	return nil
}

// RegisterHandler for message types.
// refer to API Docs https://picfinder.ai/support/en/collections/4049537-api-docs for message types
// refer to message.go for all supported types
func (sdk *Runware) RegisterHandler(messageType string, handler IncomingMessageHandler) error {
	if _, ok := sdk.messageHandlers[messageType]; ok {
		return fmt.Errorf("%w:[%s]", ErrHandlerExists, messageType)
	}
	
	sdk.messageHandlers[messageType] = handler
	return nil
}

// handleIncomingMessage handle incoming message pre-wrap and
// assign the message to its proper handler
func (sdk *Runware) handleIncomingMessage(message []byte) {
	// Extract the message root key here
	var msgData map[string]interface{}
	if err := json.Unmarshal(message, &msgData); err != nil {
		log.Println("Error decoding response message:", err)
		return
	}
	
	for key, value := range msgData {
		switch reflect.TypeOf(value).Kind() {
		case reflect.Bool:
			if key == "error" {
				errorMessage := fmt.Sprintf("Error from server: %v", msgData)
				log.Println(errorMessage)
				return
			}
		case reflect.Map, reflect.Struct, reflect.String:
			if handler, ok := sdk.messageHandlers[key]; ok {
				bValue, err := interfaceToByte(value)
				if err != nil {
					log.Println("Error converting interface to byte:", err)
				}
				handler(bValue)
			}
		default:
			log.Printf("Unsupported value type for key '%s': %v", key, reflect.TypeOf(value))
		}
	}
}

// New create a new client and initiate connection
func New(connEnv ConnAddr, apiKey string) (*Runware, error) {
	
	if apiKey == "" {
		return nil, ErrApiKeyRequired
	}
	
	client, err := wsConnect(string(connEnv))
	if err != nil {
		return nil, fmt.Errorf("%w:[%s]", ErrWsDial, string(connEnv))
	}
	
	runware := &Runware{
		apiKey:          apiKey,
		connStr:         connEnv,
		client:          client,
		messageHandlers: make(map[string]IncomingMessageHandler),
		sendQueue:       make(chan []byte),
	}
	
	go runware.Listen()
	
	return runware, nil
}

func interfaceToByte(v interface{}) ([]byte, error) {
	vBit, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return vBit, nil
}

func wsConnect(connStr string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(connStr, nil)
	return conn, err
}
