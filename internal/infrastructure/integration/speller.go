package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	apiURL     string
	httpClient doer
}

func NewClient(
	apiURL string,
	httpClient doer,
) *Client {
	return &Client{
		apiURL:     apiURL,
		httpClient: httpClient,
	}
}

type ValidateTextResponse struct {
	Code int      `json:"code"`
	Pos  int      `json:"pos"`
	Row  int      `json:"row"`
	Col  int      `json:"col"`
	Len  int      `json:"len"`
	Word string   `json:"word"`
	S    []string `json:"s"`
}

func (c Client) Validate(
	ctx context.Context,
	text string,
) ([]ValidateTextResponse, error) {

	endpoint := c.apiURL + "/checkText?" + url.Values{
		"text": {text},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return []ValidateTextResponse{}, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return []ValidateTextResponse{}, fmt.Errorf("httpClient.Do: %w", err)
	}

	defer resp.Body.Close()
	var response []ValidateTextResponse

	if resp.Status != "200 OK" {
		return response, fmt.Errorf("error status code: %s", err)
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return []ValidateTextResponse{}, fmt.Errorf("json.Decode: %w", err)
	}

	return response, nil
}
