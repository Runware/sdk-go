package runware

type SDK struct {
	Client Runware
}

func NewSDK(client Runware) (*SDK, error) {
	sdk := &SDK{
		Client: client,
	}
	
	if !sdk.Client.Connected() {
		return nil, ErrWsNotConnected
	}
	
	return sdk, nil
}
