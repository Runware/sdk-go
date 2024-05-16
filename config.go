package runware

type RunwareConfig struct {
	APIKey    string
	ConnAddr  ConnAddr
	KeepAlive bool
}

type SDKConfig struct {
	APIKey    string
	ConnAddr  ConnAddr
	KeepAlive bool
	Client    Runware
}
