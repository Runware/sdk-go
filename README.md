# Golang Runware SDK

> The SDK is used to run image inference with the Runware API, powered by the RunWare inference platform. It can be used to generate imaged with text-to-image and image-to-image. It also allows the use of an existing gallery of models or selecting any model or LoRA from the CivitAI gallery. 

## Get API access

For an API Key and trial credits, [Create a free account](https://my.runware.ai/) with [Runware](https://runware.ai)

### NB: Please keep your API key private

## Usage

```shell
go get github.com/Runware/sdk-go
```

### Basic image task

```go

import (
    runware "github.com/Runware/sdk-go"
)


func main() {
    runwareCli, err := runware.New(picfinder.Config{
        APIKey: os.Getenv("RUNWARE_API"),
    })
    if err != nil {
        panic(err)
    }
    
    sdk, err := runware.NewSDK(runwareCli)
    if err != nil {
        panic(err)
    }

    imagesRes, err := sdk.NewImage(ctx, picfinder.NewTaskReq{
        PromptText:    "Prompt text",
        NumberResults: 12,
    })
    if err != nil {
        panic(err)
    }
    
    prettyJSON, _ := json.MarshalIndent(imagesRes, "", "  ")
    log.Println(string(prettyJSON), "Count", len(imagesRes.Images))
}
```

For advanced requests refer to `types.go` file 

#### Sample 

```go
req := runware.NewTaskReq{
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
```

## Roadmap

- Add custom handler support for API events
- Implement SDK support for ControlNets 
- Implement SDK support for Image upscaling
- Implement SDK support for Image upload
- Implement SDK support for Image interrogator
- Implement SDK support for Prompt enhancer

