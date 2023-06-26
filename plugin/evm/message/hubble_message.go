package message

type MessageCode uint

const (
	PlaceOrder MessageCode = iota
)

type HubbleGossip struct {
	Message     []byte      `serialize:"true"`
	Signature   []byte      `serialize:"true"`
	MessageCode MessageCode `serialize:"true"`
}
