package client

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/scottrmalley/p2p-file-challenge/api"
	"github.com/scottrmalley/p2p-file-challenge/proof"
)

type Persistence interface {
	SetFileSet(setId string, root []byte, count int) error
	FileSet(setId string) ([]byte, int, error)
	Sets() ([]string, error)
}

// Client is a client for the file service. It is separate from the
// api.Client which is strictly a client for the api. This client
// requires a persistence layer for storing the root hash, but in our
// case we use an in-memory persistence layer.
type Client struct {
	persistence Persistence
	apiClient   *api.Client
}

func NewClient(persistence Persistence, apiClient *api.Client) *Client {
	return &Client{
		persistence: persistence,
		apiClient:   apiClient,
	}
}

// CreateSet will not assume the provided fileset is complete
// the api will only return valid proofs once the number of files uploaded
// matches the setCount
func (c *Client) CreateSet(root []byte, setCount int) (string, error) {
	setId := uuid.New()
	if err := c.persistence.SetFileSet(setId.String(), root, setCount); err != nil {
		return "", err
	}
	return setId.String(), nil
}

// AddFile will add a file to the file set. If the file set is complete
// it will return an error
func (c *Client) AddFile(setId string, index int, file []byte) error {
	_, count, err := c.persistence.FileSet(setId)
	if err != nil {
		return err
	}
	if count <= index {
		return errors.Errorf("index %d out of range for file set %s", index, setId)
	}
	if _, err := c.apiClient.PostFile(
		&api.PostFileRequest{
			SetId:    setId,
			SetCount: count,
			Index:    index,
			Content:  hexutil.Encode(file),
		},
	); err != nil {
		return err
	}
	return nil
}

// PostFiles will assume the file set provided is complete
func (c *Client) PostFiles(files [][]byte) (string, error) {
	setId := uuid.New()
	for i, file := range files {
		if _, err := c.apiClient.PostFile(
			&api.PostFileRequest{
				SetId:    setId.String(),
				SetCount: len(files),
				Index:    i,
				Content:  hexutil.Encode(file),
			},
		); err != nil {
			return "", err
		}
	}

	root, err := proof.Root(files)
	if err != nil {
		return "", err
	}
	if err := c.persistence.SetFileSet(setId.String(), root, len(files)); err != nil {
		return "", err
	}
	return setId.String(), nil
}

// GetFile will verify the proof returned by the api or return an error
func (c *Client) GetFile(setId string, index int) ([]byte, error) {
	root, count, err := c.persistence.FileSet(setId)
	if err != nil {
		return nil, err
	}
	if count <= index {
		return nil, errors.Errorf("index %d out of range for file set %s", index, setId)
	}
	out, err := c.apiClient.GetFile(
		setId,
		index,
	)
	if err != nil {
		return nil, err
	}
	hashes, position, err := decodeProofResponse(out.Proof)
	if err != nil {
		return nil, err
	}
	file, err := hexutil.Decode(out.File)
	if err != nil {
		return nil, err
	}
	if success, err := proof.Verify(file, hashes, position, root); err != nil {
		return nil, errors.Wrap(err, "failed to verify proof")
	} else if !success {
		return nil, errors.New("proof verification failed")
	}
	return file, nil
}

func (c *Client) Sets() ([]string, error) {
	return c.persistence.Sets()
}

func (c *Client) SetSize(setId string) (int, error) {
	_, count, err := c.persistence.FileSet(setId)
	return count, err
}

func decodeProofResponse(in api.ProofResponse) (hashes [][]byte, index uint64, err error) {
	hashes = make([][]byte, len(in.Proof))
	for i, hash := range in.Proof {
		hashes[i], err = hexutil.Decode(hash)
		if err != nil {
			return nil, 0, err
		}
	}
	return hashes, in.Index, nil
}
