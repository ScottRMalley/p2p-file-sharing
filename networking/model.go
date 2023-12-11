package networking

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

// basically the same as model.File with some small adjustments
// but allows us to keep the DTO separate from what the service uses
// internally
type fileMetadata struct {
	SenderId   string `json:"senderId"`
	SetId      string `json:"setId"`
	SetCount   int    `json:"setCount"`
	FileNumber int    `json:"fileNumber"`
}

type fileMsg struct {
	Metadata fileMetadata `json:"metadata"`
	Contents string       `json:"contents"`
}

type Connection struct {
	ps   *pubsub.PubSub
	self peer.ID
}

func NewConnection(ps *pubsub.PubSub, self peer.ID) *Connection {
	return &Connection{
		ps:   ps,
		self: self,
	}
}
