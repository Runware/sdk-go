package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	
	picfinder "github.com/Runware/sdk-go"
)

func main() {
	
	picSdk, err := picfinder.New(picfinder.Config{
		APIKey: os.Getenv("RUNWARE_API"),
	})
	if err != nil {
		panic(err)
	}
	
	sdk, err := picfinder.NewSDK(picSdk)
	if err != nil {
		panic(err)
	}
	
	ctx := context.Background()
	// ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	// defer cancel()
	//
	imagesRes, err := sdk.NewImage(ctx, picfinder.NewTaskReq{
		PromptText:    "futuristic with steampunk vibes of a beautiful cityscape dark morning vibes. Neon signs here and there are tuned on the stores. Abandoned cities skylines mixed with vegetation",
		NumberResults: 12,
	})
	if err != nil {
		panic(err)
	}
	
	prettyJSON, _ := json.MarshalIndent(imagesRes, "", "  ")
	log.Println(string(prettyJSON), "Count", len(imagesRes.Images))
	
	picfinder.NewTaskReq{
		TaskUUID:           "",
		ImageInitiatorUUID: "",
		PromptText:         "",
		NumberResults:      0,
		ModelId:            "",
		SizeId:             0,
		TaskType:           0,
		PromptLanguageId:   nil,
		Offset:             0,
		Lora:               nil,
		ControlNet:         nil,
	}
	
	select {}
}
