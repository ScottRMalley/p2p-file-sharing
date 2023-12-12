package api

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/rs/zerolog"
)

// Controller should handle encoding/decoding of requests
// and uses Gin to handle HTTP requests / tonic to extract
// request parameters
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

func (c *Controller) PostFile(_ *gin.Context, in *PostFileRequest) (*PostFileResponse, error) {
	setId, err := uuid.Parse(in.SetId)
	if err != nil {
		return nil, err
	}

	fileBytes, err := hexutil.Decode(in.Content)
	if err != nil {
		return nil, err
	}
	hash, err := c.service.SaveFile(setId, in.Index, in.SetCount, fileBytes)
	if err != nil {
		return nil, err
	}
	return &PostFileResponse{
		Success: true,
		Hash:    hash,
	}, nil
}

func (c *Controller) GetFile(_ *gin.Context, in *GetFileRequest) (*GetFileResponse, error) {
	setId, err := uuid.Parse(in.SetId)
	if err != nil {
		return nil, err
	}
	file, proof, index, err := c.service.File(setId, in.Index)
	if err != nil {
		return nil, err
	}
	return &GetFileResponse{
		File: hexutil.Encode(file),
		Proof: ProofResponse{
			Proof: strings(proof),
			Index: index,
		},
	}, nil
}

// RegisterRoutes registers the routes on the given router group
func (c *Controller) RegisterRoutes(router *gin.RouterGroup) error {
	router.POST("/sets/:setId/files/:index", tonic.Handler(c.PostFile, 200))
	router.GET("/sets/:setId/files/:index", tonic.Handler(c.GetFile, 200))
	return nil
}

// strings converts a [][]byte to []string
// and is just a helper function for the controller
func strings(in [][]byte) []string {
	out := make([]string, len(in))
	for i, b := range in {
		out[i] = hexutil.Encode(b)
	}
	return out
}
