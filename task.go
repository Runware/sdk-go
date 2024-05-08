package runware

const (
	TextToImage = iota + 1 // Start iota at 1 to match your values
	ImageToImage
	Inpainting
	ImageToText
	PromptEnhancer
	ImageUpscale
	ImageUpload
	RemoveBackground
	ControlNetTextToImage
	ControlNetImageToImage
	ControlNetPreprocessImage
)

type Lora struct {
	ModelID string  `json:"modelId"`
	Weight  float64 `json:"weight"`
}

type Task struct {
	ID   string `json:"taskUUID"`
	Lora []Lora `json:"lora"`
	
	// 	Add missing properties
}
