package api

import (
	"github.com/go-resty/resty/v2"
	"gopkg.in/errgo.v2/fmt/errors"
	"net/url"
)

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

func (c *Client) PostFiles(in *FilesInput) (*FilesOutput, error) {
	out := new(FilesOutput)
	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetBody(in).
		SetResult(out).
		Post(c.baseUrl.String() + "/files")
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, errors.Newf("error posting files: %s", res.String())
	}
	return out, nil
}

func (c *Client) GetFile(setId string, index int) (*GetFileOutput, error) {
	out := new(GetFileOutput)
	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetPathParams(
			map[string]string{
				"setId": setId,
				"file":  string(index),
			},
		).
		SetResult(out).
		Get(c.baseUrl.String() + "/sets/{setId}/files/{file}")
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, errors.Newf("error getting file: %s", res.String())
	}
	return out, nil
}
