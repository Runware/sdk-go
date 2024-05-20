package runware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
	
	"github.com/google/uuid"
)

type NewTaskReq Task

type NewTaskResp struct {
	Images                []Image `json:"images"`
	TotalAvailableResults int     `json:"totalAvailableResults"`
	TimedOut              bool    `json:"timedOut"`
}

func (sdk *SDK) NewImage(ctx context.Context, req NewTaskReq) (*NewTaskResp, error) {
	if err := validateNewTaskReq(req); err != nil {
		return nil, err
	}
	
	// In case `req.TaskType` is empty try to evaluate it
	if req.TaskType == 0 {
		req.TaskType = getTaskType(req.PromptText, req.ControlNet, "", req.ImageInitiatorUUID)
	}
	
	req = *mergeNewTaskReqWithDefaults(&req)
	
	newTaskReq := Request{
		ID:            uuid.New().String(),
		Event:         NewTask,
		ResponseEvent: NewImage,
		Data:          req,
	}
	
	newTaskResp := &NewTaskResp{
		Images: make([]Image, 0),
	}
	
	responseChan := make(chan *NewTaskResp)
	errChan := make(chan error)
	
	go func() {
		defer close(responseChan)
		defer close(errChan)
		
		currentCount := 0
		for msg := range sdk.Client.Listen() {
			var msgData map[string]interface{}
			if err := json.Unmarshal(msg, &msgData); err != nil {
				errChan <- fmt.Errorf("%w:[%s]", ErrDecodeMessage, err.Error())
				return
			}
			
			// Check if is an error message first
			if errMsg, ok := sdk.OnError(msgData); ok {
				errChan <- errMsg
				return
			}
			
			for k, v := range msgData {
				if k != newTaskReq.ResponseEvent {
					log.Println("Skipping event", k, "Currently handling", newTaskReq.ResponseEvent)
					continue
				}
				
				bValue, err := interfaceToByte(v)
				if err != nil {
					errChan <- err
					return
				}
				
				var iterTaskResp *NewTaskResp
				err = json.Unmarshal(bValue, &iterTaskResp)
				if err != nil {
					errChan <- err
					return
				}
				
				newTaskResp.Images = mergeImageResults(iterTaskResp.Images, newTaskResp.Images)
				currentCount += len(iterTaskResp.Images)
				fmt.Println("-> SET:", currentCount)
				
				if currentCount >= req.NumberResults {
					responseChan <- newTaskResp
					return
				}
				
				time.Sleep(30 * time.Millisecond)
			}
		}
	}()
	
	bSendReq, err := newTaskReq.ToEvent()
	if err != nil {
		return nil, err
	}
	
	if err = sdk.Client.Send(bSendReq); err != nil {
		return nil, err
	}
	
	select {
	case resp := <-responseChan:
		return resp, nil
	case err = <-errChan:
		return nil, err
	case <-time.After(timeoutSendResponse * time.Second):
		newTaskResp.TimedOut = true
		return newTaskResp, fmt.Errorf("%w:[%s]", ErrRequestTimeout, newTaskReq.Event)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	
}

// NewTaskReqDefaults set requests defaults
// TODO: Add task type determination function helper
func NewTaskReqDefaults() *NewTaskReq {
	return &NewTaskReq{
		TaskUUID:      uuid.New().String(),
		TaskType:      TextToImage,
		SizeId:        SizeSquare512,
		Offset:        0,
		NumberResults: 4,
		ModelId:       strconv.Itoa(ModelAbsolutereality),
	}
}

func mergeImageResults(src, dest []Image) []Image {
	for _, img := range src {
		for idx, destImg := range dest {
			if img.ImageUUID == destImg.ImageUUID {
				dest[idx].ImageAltText = img.ImageAltText
				dest[idx].BNSFWContent = img.BNSFWContent
				dest[idx].ImageSrc = img.ImageSrc
				
				return dest
			}
		}
		
		// If not found add the entire object
		dest = append(dest, img)
	}
	return dest
}

func mergeNewTaskReqWithDefaults(req *NewTaskReq) *NewTaskReq {
	_ = MergeEventRequestsWithDefaults[*NewTaskReq](req, NewTaskReqDefaults())
	return req
}

func validateNewTaskReq(req NewTaskReq) error {
	if req.PromptText == "" {
		return fmt.Errorf("%w:[%s]", ErrFieldRequired, "promptText")
	}
	
	return nil
}

func getTaskType(promptText string, controlNet []ControlNet, imageMaskUUID, imageInitiatorUUID string) int {
	hasPrompt := promptText != ""
	hasControlNet := len(controlNet) > 0
	hasMask := imageMaskUUID != ""
	hasInitiator := imageInitiatorUUID != ""
	
	switch {
	case hasPrompt && !hasControlNet && !hasMask && !hasInitiator:
		return TextToImage
	case hasPrompt && !hasControlNet && !hasMask && hasInitiator:
		return ImageToImage
	case hasPrompt && !hasControlNet && hasMask && hasInitiator:
		return Inpainting
	case hasPrompt && hasControlNet && !hasMask && !hasInitiator:
		return ControlNetTextToImage
	case hasPrompt && hasControlNet && !hasMask && hasInitiator:
		return ControlNetImageToImage
	case hasPrompt && hasControlNet && hasMask && hasInitiator:
		return ControlNetPreprocessImage
	default:
		return 0
	}
}
