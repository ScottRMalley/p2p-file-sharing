package api

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/rs/zerolog"
)

type Controller struct {
	logger  zerolog.Logger
	service *Service
}

func NewController(logger zerolog.Logger, service *Service) *Controller {
	return &Controller{
		logger:  logger,
		service: service,
	}
}

func (c *Controller) PostFiles(ctx *gin.Context, in *FilesInput) (*FilesOutput, error) {
	setId := uuid.New()
	hash, err := c.service.Files(setId, bytes(in.Files))
	if err != nil {
		return nil, err
	}
	return &FilesOutput{
		Success: true,
		SetId:   setId.String(),
		Root:    hash,
	}, nil
}

func (c *Controller) GetFile(ctx *gin.Context, in *GetFileInput) (*GetFileOutput, error) {
	setId, err := uuid.Parse(in.SetId)
	if err != nil {
		return nil, err
	}
	file, proof, index, err := c.service.File(setId, in.Index)
	if err != nil {
		return nil, err
	}
	return &GetFileOutput{
		File: hexutil.Encode(file),
		Proof: ProofResponse{
			Proof: strings(proof),
			Index: index,
		},
	}, nil
}

func (c *Controller) RegisterRoutes(router *gin.RouterGroup) error {
	router.POST("/sets", tonic.Handler(c.PostFiles, 201))
	router.GET("/sets/:setId/files/:file", tonic.Handler(c.GetFile, 200))
	return nil
}

func bytes(in []string) [][]byte {
	out := make([][]byte, len(in))
	for i, s := range in {
		out[i] = []byte(s)
	}
	return out
}

func strings(in [][]byte) []string {
	out := make([]string, len(in))
	for i, b := range in {
		out[i] = hexutil.Encode(b)
	}
	return out
}
