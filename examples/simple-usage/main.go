package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	
	picfinder "github.com/Runware/sdk-go"
	"github.com/google/uuid"
)

func main() {
	
	sdk, err := picfinder.NewSDK(picfinder.SDKConfig{
		APIKey:    os.Getenv("RUNWARE_API"),
		KeepAlive: true,
	})
	if err != nil {
		panic(err)
	}
	
	ctx := context.Background()
	// ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	// defer cancel()
	
	log.Println("Text to Image")
	imagesRes, err := sdk.NewImage(ctx, picfinder.NewTaskReq{
		PromptText:    "neon punk retro futuristic 1970 fallout game vibes like theme rocky deserted villages outside the cities. Air looks dusty un unclean with a tint of red. Debris and rusty cars here and there set the scene",
		NumberResults: 12,
	})
	if err != nil {
		if !errors.Is(err, picfinder.ErrRequestTimeout) {
			panic(err)
		}
	}
	
	jsonPrint(imagesRes)
	
	workImage := imagesRes.Images[0]
	log.Println("Image to Image image: UUID", workImage.ImageUUID)
	imagesRes, err = sdk.NewImage(ctx, picfinder.NewTaskReq{
		PromptText:         "neon punk retro futuristic 1970 fallout game vibes like theme rocky deserted villages outside the cities. Air looks dusty un unclean with a tint of red. Debris and rusty cars here and there set the scene",
		NumberResults:      12,
		ImageInitiatorUUID: workImage.ImageUUID,
	})
	if err != nil {
		if !errors.Is(err, picfinder.ErrRequestTimeout) {
			panic(err)
		}
	}
	jsonPrint(imagesRes)
	
	log.Println("ControlNets image: UUID", workImage.ImageUUID)
	
	taskID := uuid.New().String()
	log.Println("ControlNets TaskUUID:", taskID)
	cnRes, err := sdk.NewControlNets(ctx, picfinder.NewControlNetsReq{
		TaskUUID:         taskID,
		PreProcessorType: picfinder.ProcessorDepth,
		GuideImageUUID:   workImage.ImageUUID,
	})
	if err != nil {
		panic(err)
	}
	
	jsonPrint(cnRes)
	
	log.Println("ControlNets image: UUID", workImage.ImageUUID)
	upscaleRes, err := sdk.ImageUpscale(context.Background(), picfinder.NewUpscaleGanReq{
		ImageUUID:     workImage.ImageUUID,
		TaskUUID:      workImage.TaskUUID,
		UpscaleFactor: 2,
	})
	
	if err != nil {
		panic(err)
	}
	
	jsonPrint(upscaleRes)
	
	select {}
}

func jsonPrint(data any) {
	prettyJSON, _ := json.MarshalIndent(data, "", "  ")
	log.Println(string(prettyJSON))
}
