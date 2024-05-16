package runware

import (
	"reflect"
)

const (
	NewConnectionSessionUUID = "newConnectionSessionUUID"
	NewImage                 = "newImages"
	NewUpscaleGan            = "newUpscaleGan"
	NewUploadedImageUUID     = "newUploadedImageUUID"
	NewReverseClip           = "newReverseClip"
	NewPromptEnhancer        = "newPromptEnhancer"
	NewConnection            = "newConnection"
	NewTask                  = "newTask"
	NewPreProcessControlNet  = "newPreProcessControlNet"
	NewImageUpload           = "newImageUpload"
	NewReverseImageClip      = "newReverseImageClip"
	NewPromptEnhance         = "newPromptEnhance"
	Pong                     = "pong"
)

func MergeEventRequestsWithDefaults[T any](cfgDest, defaultCfgDest T) error {
	dstVal := reflect.ValueOf(cfgDest).Elem()
	defaultVal := reflect.ValueOf(defaultCfgDest).Elem()
	
	for i := 0; i < dstVal.NumField(); i++ {
		dstField := dstVal.Field(i)
		defaultField := defaultVal.Field(i)
		
		if isZeroValues(dstField) {
			setField(dstField, defaultField)
		}
	}
	
	return nil
}

func isZeroValues(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.String:
		return field.Len() == 0
	case reflect.Ptr:
		return field.IsNil()
	default:
		return reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface())
	}
}

func setField(target, source reflect.Value) {
	if target.CanSet() {
		if source.Kind() == reflect.Ptr && source.IsNil() {
			return
		}
		// Dereference pointers if necessary
		if target.Kind() == reflect.Ptr && !target.IsNil() {
			target.Elem().Set(source)
		} else {
			target.Set(source)
		}
	}
}
