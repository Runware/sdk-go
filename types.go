package runware

// Task types
const (
	TextToImage               = 1
	ImageToImage              = 2
	Inpainting                = 3
	ImageToText               = 4
	PromptEnhancer            = 5
	ImageUpscale              = 6
	ImageUpload               = 7
	RemoveBackground          = 8
	ControlNetTextToImage     = 9
	ControlNetImageToImage    = 10
	ControlNetPreprocessImage = 11
)

// Available models
const (
	ModelSDXL               = 4
	ModelRevAnimated        = 13
	ModelAbsolutereality    = 18
	ModelCyberrealistic     = 19
	ModelDreamshaper        = 20
	ModelGhostmixBakedvae   = 22
	ModelSamaritan3DCartoon = 25
)

// Available processors
const (
	ProcessorCanny        = "canny"
	ProcessorDepth        = "depth"
	ProcessorMlsd         = "mlsd"
	ProcessorNormalbae    = "normalbae"
	ProcessorOpenpose     = "openpose"
	ProcessorTile         = "tile"
	ProcessorSeg          = "seg"
	ProcessorLineart      = "lineart"
	ProcessorLineartAnime = "lineart_anime"
	ProcessorShuffle      = "shuffle"
	ProcessorScribble     = "scribble"
	ProcessorSoftedge     = "softedge"
)

// Available sizes
const (
	SizeSquare512          = 1
	SizePortrait2to3       = 2
	SizePortrait1to2       = 3
	SizeLandscape2to3      = 4
	SizeLandscape2to1      = 5
	SizeLandscape4to3      = 6
	SizeLandscape16to9     = 7
	SizePortrait9to16      = 8
	SizePortrait3to4       = 9
	SizeSquare1024SDXL     = 11
	SizeLandscape16to9SDXL = 16
	SizePortrait9to16SDXL  = 17
	SizePortrait2to3SDXL   = 20
	SizeLandscape3to2SDXL  = 21
)

type ControlNet struct {
	Preprocessor   string  `json:"preprocessor"`
	Weight         float64 `json:"weight"`
	StartStep      int     `json:"startStep"`
	EndStep        int     `json:"endStep"`
	GuideImageUUID string  `json:"guideImageUUID"`
	ControlMode    string  `json:"controlMode"`
}

type Image struct {
	ImageSrc     string `json:"imageSrc"`
	ImageUUID    string `json:"imageUUID"`
	BNSFWContent bool   `json:"bNSFWContent"`
	ImageAltText string `json:"imageAltText"`
	TaskUUID     string `json:"taskUUID"`
}

type Text struct {
	TaskUUID string `json:"taskUUID"`
	Text     string `json:"text"`
}

type Lora struct {
	ModelID string  `json:"modelId"`
	Weight  float64 `json:"weight"`
}

type PreProcessControlNet struct {
	TaskUUID           string `json:"taskUUID"`
	PreProcessorType   string `json:"preProcessorType"`
	GuideImageUUID     string `json:"guideImageUUID"`
	TaskType           int    `json:"taskType"`
	Width              int    `json:"width"`
	Height             int    `json:"height"`
	LowThresholdCanny  int    `json:"lowThresholdCanny"`
	HighThresholdCanny int    `json:"highThresholdCanny"`
}

type Task struct {
	TaskUUID           string       `json:"taskUUID"`
	ImageInitiatorUUID string       `json:"imageInitiatorUUID,omitempty"`
	PromptText         string       `json:"promptText"`
	NumberResults      int          `json:"numberResults"`
	ModelId            string       `json:"modelId"`
	SizeId             int          `json:"sizeId"`
	TaskType           int          `json:"taskType"`
	PromptLanguageId   *string      `json:"promptLanguageId"`
	Offset             int          `json:"offset"`
	Lora               []Lora       `json:"lora"`
	ControlNet         []ControlNet `json:"controlNet"`
}
