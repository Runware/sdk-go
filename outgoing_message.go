package runware

// OutgoingMessageHandler interface implementation for outgoing messages
type OutgoingMessageHandler interface {
	MarshalBinary() ([]byte, error)
}
