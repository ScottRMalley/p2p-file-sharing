package networking

import (
	"context"
	"github.com/rs/zerolog"
)

// SetAnnouncementTopic is the topic that is used to announce new file sets
// to the network. It is fixed for all participants
const SetAnnouncementTopic = "set-announcement"

type SetAnnouncement struct {
	pub *IOTopic[*setAnnouncementMsg]
}

func NewSetAnnouncement(
	logger zerolog.Logger,
	connection *Connection,
) (*SetAnnouncement, error) {
	pub, err := NewIOTopic[*setAnnouncementMsg](
		logger,
		connection.ps,
		SetAnnouncementTopic,
		connection.self,
	)
	if err != nil {
		return nil, err
	}
	return &SetAnnouncement{
		pub: pub,
	}, nil
}

func (sa *SetAnnouncement) Write(ctx context.Context, setId string) error {
	return sa.pub.Write(
		ctx, &setAnnouncementMsg{
			SetId: setId,
		},
	)
}

func (sa *SetAnnouncement) Read(ctx context.Context) <-chan string {
	// here we just want to transform the channel type from *setAnnouncementMsg to string
	// so we can return a channel of string
	ids := make(chan string)
	go func() {
		defer close(ids)
		for sam := range sa.pub.Read(ctx) {
			ids <- sam.SetId
		}
	}()
	return ids
}
