package networking

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/scottrmalley/p2p-file-challenge/model"
)

type FileSet struct {
	pub *IOTopic[*fileMsg]
}

func NewFileSet(
	logger zerolog.Logger,
	connection *Connection,
	topicName string,
) (*FileSet, error) {
	pub, err := NewIOTopic[*fileMsg](logger, connection.ps, topicName, connection.self)
	if err != nil {
		return nil, err
	}
	return &FileSet{
		pub: pub,
	}, nil
}

func (fs *FileSet) Write(ctx context.Context, file model.File) error {
	fm := &fileMsg{
		Metadata: fileMetadata{
			Id:         hexutil.Encode(crypto.Keccak256(file.Contents)),
			SenderId:   fs.pub.self.String(),
			SetId:      file.Metadata.SetId,
			SetCount:   file.Metadata.SetCount,
			FileNumber: file.Metadata.FileNumber,
		},
		Contents: hexutil.Encode(file.Contents),
	}

	return fs.pub.Write(ctx, fm)
}

func (fs *FileSet) Read(ctx context.Context) <-chan model.File {
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
					Id:         fm.Metadata.Id,
					SetId:      fm.Metadata.SetId,
					SetCount:   fm.Metadata.SetCount,
					FileNumber: fm.Metadata.FileNumber,
				},
				Contents: content,
			}
			files <- f
		}
		close(files)
	}()
	return files
}

func (fs *FileSet) Close() error {
	return fs.pub.Close()
}
