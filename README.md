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
## Advanced settings 

### Context adjustments

By default, all tasks have a 30-second timeout, but sometimes you might want a fast expiry. This can be passed via context

```go
ctx := context.Background()
ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
defer cancel()

imagesRes, err := sdk.NewImage(ctx, picfinder.NewTaskReq{
    PromptText:    "Some prompt text",
    NumberResults: 1,
})
```
to close this request after 5 seconds.


### Custom UUID for Requests

If at some point you need to group your execution your self and you need to do something with them based 
on your business needs you can pass your own UUID v4 to any sdk request via `TaskUUID`

## Roadmap

- Add custom handler support for API events

