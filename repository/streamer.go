package repository

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/scottrmalley/p2p-file-sharing/model"
)

type persistence interface {
	SaveFile(file model.File) error
}

// Streamer is responsible for watching new files as they are read from the
// file topic and saving them to the persistence layer
type Streamer struct {
	logger zerolog.Logger

	repo persistence
}

func NewStreamer(logger zerolog.Logger, repo persistence) *Streamer {
	return &Streamer{
		logger: logger,
		repo:   repo,
	}
}

// WatchNew returns a func() error in order to be easily used with
// errgroup.Group
func (s *Streamer) WatchNew(ctx context.Context, files <-chan model.File) func() error {
	return func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case file := <-files:
				s.logger.Debug().
					Int("file-number", file.Metadata.FileNumber).
					Str("set-id", file.Metadata.SetId).
					Msg("received file")
				if err := s.repo.SaveFile(file); err != nil {
					s.logger.Error().Err(err).Msg("failed to save file")
				}
			}
		}
	}
}
