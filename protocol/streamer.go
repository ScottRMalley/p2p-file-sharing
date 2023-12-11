package protocol

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/scottrmalley/p2p-file-challenge/model"
	"github.com/scottrmalley/p2p-file-challenge/networking"
)

type Streamer struct {
	logger     zerolog.Logger
	process    *Processor
	connection *networking.Connection
}

func NewStreamer(logger zerolog.Logger, process *Processor, connection *networking.Connection) *Streamer {
	return &Streamer{
		logger:     logger,
		process:    process,
		connection: connection,
	}
}

func (s *Streamer) WatchNew(ctx context.Context, fileSets <-chan string) func() error {
	return func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case setId := <-fileSets:
				files, err := networking.NewFileSet(
					s.logger,
					s.connection,
					setId,
				)
				if err != nil {
					return err
				}
				go func() {
					defer func() {
						if err := files.Close(); err != nil {
							s.logger.Error().Err(err).Msg("failed to close file set")
						}
					}()
					if err := s.ProcessStream(ctx, files.Read(ctx))(); err != nil {
						s.logger.Error().Err(err).Msg("failed to process set")
					}
				}()
			}
		}
	}
}

func (s *Streamer) ProcessStream(ctx context.Context, files <-chan model.File) func() error {
	return func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case file := <-files:
				err := s.process.ProcessFile(file)
				if err != nil {
					return err
				}
				if hash, err := s.process.CompleteSet(file.Metadata.SetId); err != nil {
					if errors.Is(err, ErrSetIncomplete) {
						continue
					} else {
						return err
					}
				} else {
					s.logger.Info().Msgf("root hash: %s", hexutil.Encode(hash))
					return nil
				}
			}
		}
	}
}
