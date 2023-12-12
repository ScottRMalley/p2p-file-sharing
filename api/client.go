package api

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

// Client is intended to be strictly a client for the api. It is
// separate from the client.Client which requires a persistence layer for
// storing the root hash.
type Client struct {
	r       *resty.Client
	baseUrl *url.URL
}

func NewClient(baseUrl string) (*Client, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	return &Client{r: resty.New(), baseUrl: u}, nil
}

func (c *Client) PostFile(in *PostFileRequest) (*PostFileResponse, error) {
	out := new(PostFileResponse)
	path := fmt.Sprintf("%s/sets/%s/files/%s", c.baseUrl.String(), in.SetId, strconv.Itoa(in.Index))
	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetBody(
			&PostFileRequest{
				Content:  in.Content,
				SetCount: in.SetCount,
			},
		).
		SetResult(out).
		Post(path)
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, errors.Errorf("error posting file: %s", res.String())
	}
	return out, nil
}

func (c *Client) GetFile(setId string, index int) (*GetFileResponse, error) {
	out := new(GetFileResponse)
	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetResult(out).
		Get(fmt.Sprintf("%s/sets/%s/files/%s", c.baseUrl.String(), setId, strconv.Itoa(index)))
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, errors.Errorf("error getting file: %s", res.String())
	}
	return out, nil
}
