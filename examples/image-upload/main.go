package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	
	picfinder "github.com/Runware/sdk-go"
	"github.com/google/uuid"
)

func main() {
	
	// Read file
	file, err := os.Open(os.Getenv("RUNWARE_IMG"))
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	file.Seek(0, 0)
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	
	// Determine the MIME type
	mimeType := http.DetectContentType(data)
	
	// Encode the byte slice to base64
	encoded := base64.StdEncoding.EncodeToString(data)
	
	// Construct the URI
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
	
	sdk, err := picfinder.NewSDK(picfinder.SDKConfig{
		APIKey:    os.Getenv("RUNWARE_API"),
		KeepAlive: true,
	})
	if err != nil {
		panic(err)
	}
	
	ctx := context.Background()
	
	log.Println("Image Upload")
	taskID := uuid.New().String()
	res, err := sdk.ImageUpload(ctx, picfinder.NewImageUploadReq{
		ImageBase64: dataURI,
		TaskUUID:    taskID,
	})
	
	if err != nil {
		if !errors.Is(err, picfinder.ErrRequestTimeout) {
			panic(err)
		}
	}
	jsonPrint(res)
	
	log.Println("Image to Text")
	imgToText, err := sdk.ImageToText(context.Background(), picfinder.NewReverseImageClipReq{
		ImageUUID: res.NewImageUUID,
	})
	
	if err != nil {
		if !errors.Is(err, picfinder.ErrRequestTimeout) {
			panic(err)
		}
	}
	jsonPrint(imgToText)
	
	log.Println("Prompt enhance")
	promptEnhance, err := sdk.PromptEnhancer(context.Background(), picfinder.NewPromptEnhanceReq{
		PromptText:      imgToText.Texts[0].Text,
		PromptMaxLength: 250,
	})
	
	if err != nil {
		if !errors.Is(err, picfinder.ErrRequestTimeout) {
			panic(err)
		}
	}
	jsonPrint(promptEnhance)
	
}

func jsonPrint(data any) {
	prettyJSON, _ := json.MarshalIndent(data, "", "  ")
	log.Println(string(prettyJSON))
}
