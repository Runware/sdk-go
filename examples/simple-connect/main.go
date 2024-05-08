package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	
	picfinder "github.com/Runware/sdk-go"
)

func main() {
	picSdk, err := picfinder.New(picfinder.ProdEnv, "abcd")
	if err != nil {
		panic(err)
	}
	
	picSdk.RegisterHandler("newConnectionSessionUUID", handleNewConnection)
	
	newConnMsg := &picfinder.NewConnection{APIKey: os.Getenv("RUNWARE_API")}
	
	go func() {
		time.Sleep(1 * time.Second)
		err = picSdk.Send(newConnMsg)
		fmt.Println("retry Send error:", err)
	}()
	
	select {}
}

type ConnectionResponse struct {
	ConnectionSessionUUID string `json:"connectionSessionUUID"`
}

func handleNewConnection(msg []byte) {
	res := &ConnectionResponse{}
	if err := json.Unmarshal(msg, res); err != nil {
		panic(err)
	}
	
	fmt.Printf("%+v", res)
	
}
