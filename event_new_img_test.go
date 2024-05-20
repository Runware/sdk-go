package runware

import (
	"testing"
)

func TestGetTaskType(t *testing.T) {
	type args struct {
		promptText         string
		controlNet         []ControlNet
		imageMaskUUID      string
		imageInitiatorUUID string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "TextToImage",
			args: args{
				promptText:         "A beautiful landscape",
				controlNet:         nil,
				imageMaskUUID:      "",
				imageInitiatorUUID: "",
			},
			want: TextToImage,
		},
		{
			name: "ImageToImage",
			args: args{
				promptText:         "A cat with a hat",
				controlNet:         nil,
				imageMaskUUID:      "",
				imageInitiatorUUID: "some-image-uuid",
			},
			want: ImageToImage,
		},
		{
			name: "Inpainting",
			args: args{
				promptText:         "Fill the sky with stars",
				controlNet:         nil,
				imageMaskUUID:      "mask-uuid",
				imageInitiatorUUID: "image-uuid",
			},
			want: Inpainting,
		},
		{
			name: "ControlNetTextToImage",
			args: args{
				promptText:         "A fantasy scene with a dragon",
				controlNet:         []ControlNet{{}},
				imageMaskUUID:      "",
				imageInitiatorUUID: "",
			},
			want: ControlNetTextToImage,
		},
		{
			name: "ControlNetImageToImage",
			args: args{
				promptText:         "Make the cat more colorful",
				controlNet:         []ControlNet{{}},
				imageMaskUUID:      "",
				imageInitiatorUUID: "cat-image-uuid",
			},
			want: ControlNetImageToImage,
		},
		{
			name: "ControlNetPreprocessImage",
			args: args{
				promptText:         "Add a glowing aura to the figure",
				controlNet:         []ControlNet{{}},
				imageMaskUUID:      "mask-uuid",
				imageInitiatorUUID: "figure-uuid",
			},
			want: ControlNetPreprocessImage,
		},
		{
			name: "InvalidInput Default to 0",
			args: args{
				promptText:         "",
				controlNet:         nil,
				imageMaskUUID:      "",
				imageInitiatorUUID: "",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTaskType(tt.args.promptText, tt.args.controlNet, tt.args.imageMaskUUID, tt.args.imageInitiatorUUID); got != tt.want {
				t.Errorf("getTaskType() = %v, want %v", got, tt.want)
			}
		})
	}
}
