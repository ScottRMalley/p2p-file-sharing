package networking

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rs/zerolog"
	"github.com/scottrmalley/p2p-file-challenge/model"
)

const FileTopicName = "file-set"

type FileTopic struct {
	pub *IOTopic[*fileMsg]
}

func NewFileTopic(
	logger zerolog.Logger,
	connection *Connection,
) (*FileTopic, error) {
	pub, err := NewIOTopic[*fileMsg](logger, connection.ps, FileTopicName, connection.self)
	if err != nil {
		return nil, err
	}
	return &FileTopic{
		pub: pub,
	}, nil
}

func (fs *FileTopic) Write(ctx context.Context, file model.File) error {
	fm := &fileMsg{
		Metadata: fileMetadata{
			SenderId:   fs.pub.self.String(),
			SetId:      file.Metadata.SetId,
			SetCount:   file.Metadata.SetCount,
			FileNumber: file.Metadata.FileNumber,
		},
		Contents: hexutil.Encode(file.Contents),
	}

	return fs.pub.Write(ctx, fm)
}

func (fs *FileTopic) Read(ctx context.Context) <-chan model.File {
	// here we just want to transform the channel type from *fileMsg to model.File
	// so we can return a channel of model.File
	files := make(chan model.File)
	go func() {
		defer close(files)
		for fm := range fs.pub.Read(ctx) {
			content, err := hexutil.Decode(fm.Contents)
			if err != nil {
				fs.pub.logger.Error().Err(err).Msg("failed to decode file contents")
				continue
			}
			f := model.File{
				Metadata: model.FileMetadata{
					SetId:      fm.Metadata.SetId,
					SetCount:   fm.Metadata.SetCount,
					FileNumber: fm.Metadata.FileNumber,
				},
				Contents: content,
			}
			files <- f
		}
	}()
	return files
}

func (fs *FileTopic) Close() error {
	return fs.pub.Close()
}
