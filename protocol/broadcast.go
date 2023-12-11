package protocol

import (
	"context"
	"github.com/rs/zerolog"

	"github.com/scottrmalley/p2p-file-challenge/model"
	"github.com/scottrmalley/p2p-file-challenge/networking"
)

type Broadcaster struct {
	logger          zerolog.Logger
	connection      *networking.Connection
	setAnnouncement *networking.SetAnnouncement
}

func NewBroadcaster(
	logger zerolog.Logger,
	connection *networking.Connection,
	announcement *networking.SetAnnouncement,
) *Broadcaster {
	return &Broadcaster{
		logger:          logger,
		connection:      connection,
		setAnnouncement: announcement,
	}
}

func (b *Broadcaster) Broadcast(ctx context.Context, files []model.File) error {
	if len(files) == 0 {
		return nil
	}
	fs, err := networking.NewFileSet(
		b.logger,
		b.connection,
		files[0].Metadata.SetId,
	)
	if err != nil {
		return err
	}

	if err := b.setAnnouncement.Write(ctx, files[0].Metadata.SetId); err != nil {
		return err
	}

	if err != nil {
		return err
	}
	for _, file := range files {
		if err := fs.Write(ctx, file); err != nil {
			return err
		}
	}
	return nil
}
