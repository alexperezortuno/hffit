package huggingface

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/alexperezortuno/hffit/internal/domain"
)

const defaultBaseURL = "https://huggingface.co"

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type modelResponse struct {
	ID          string `json:"id"`
	LibraryName string `json:"library_name"`
	PipelineTag string `json:"pipeline_tag"`

	Config struct {
		Architectures []string `json:"architectures"`
		ModelType     string   `json:"model_type"`
	} `json:"config"`

	Safetensors struct {
		Parameters map[string]uint64 `json:"parameters"`
		Total      uint64            `json:"total"`
	} `json:"safetensors"`
}

func (c *Client) GetModel(
	ctx context.Context,
	modelID string,
) (*domain.Model, error) {

	endpoint := fmt.Sprintf(
		"%s/api/models/%s",
		c.baseURL,
		modelID,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"User-Agent",
		"hffit/0.1",
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("huggingface request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"huggingface returned HTTP %d",
			resp.StatusCode,
		)
	}

	var response modelResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf(
			"decode huggingface response: %w",
			err,
		)
	}

	model := &domain.Model{
		ID:          response.ID,
		Library:     response.LibraryName,
		PipelineTag: response.PipelineTag,
		Parameters:  response.Safetensors.Total,
		ModelType:   response.Config.ModelType,
	}

	if len(response.Config.Architectures) > 0 {
		model.Architecture = response.Config.Architectures[0]
	}

	return model, nil
}

func escapeModelID(modelID string) string {
	return url.PathEscape(modelID)
}
