package runware

import (
	"context"
	"encoding/json"
	"fmt"
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
	req = *mergeWithDefaults(&req)
	if err := validate(req); err != nil {
		return nil, err
	}
	
	sendReq := Request{
		Event:         NewTask,
		ResponseEvent: NewImage,
		Data:          req,
	}
	
	resChan := make(chan *NewTaskResp)
	errChan := make(chan error)
	
	newTaskResp := &NewTaskResp{
		Images: make([]Image, 0),
	}
	
	go func() {
		currentCount := 0
		
		for {
			res, err := sdk.Client.SendAndResponse(ctx, sendReq)
			if err != nil {
				errChan <- err
				return
			}
			
			var iterTaskResp *NewTaskResp
			err = json.Unmarshal(res.Data, &iterTaskResp)
			if err != nil {
				errChan <- err
				return
			}
			
			newTaskResp.TotalAvailableResults += iterTaskResp.TotalAvailableResults
			newTaskResp.Images = mergeImageResults(iterTaskResp.Images, newTaskResp.Images)
			
			currentCount += len(iterTaskResp.Images)
			
			if currentCount >= req.NumberResults {
				resChan <- newTaskResp
				return
			}
			
			time.Sleep(500 * time.Millisecond)
		}
	}()
	
	select {
	case resp := <-resChan:
		return resp, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(10 * time.Second):
		newTaskResp.TimedOut = true
		return newTaskResp, nil
	}
}

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

func mergeWithDefaults(req *NewTaskReq) *NewTaskReq {
	_ = MergeEventRequestsWithDefaults[*NewTaskReq](req, NewTaskReqDefaults())
	return req
}

func validate(req NewTaskReq) error {
	if req.PromptText == "" {
		return fmt.Errorf("%w:[%s]", ErrPromptRequired, NewTask)
	}
	
	return nil
}
