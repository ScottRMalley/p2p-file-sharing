package repository

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/scottrmalley/p2p-file-challenge/model"
)

type persistence interface {
	SaveFile(file model.File) error
}

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

func (s *Streamer) WatchNew(ctx context.Context, files <-chan model.File) func() error {
	return func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case file := <-files:
				s.logger.Info().
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
