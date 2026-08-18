package huggingface

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

	Siblings []struct {
		Filename string `json:"rfilename"`
	} `json:"siblings"`
}

type configResponse struct {
	Architectures []string `json:"architectures"`
	ModelType     string   `json:"model_type"`

	HiddenSize            int `json:"hidden_size"`
	NumHiddenLayers       int `json:"num_hidden_layers"`
	NumAttentionHeads     int `json:"num_attention_heads"`
	NumKeyValueHeads      int `json:"num_key_value_heads"`
	MaxPositionEmbeddings int `json:"max_position_embeddings"`
	HeadDim               int `json:"head_dim"`

	// Algunos modelos/configs pueden usar nombres distintos.
	NLayer int `json:"n_layer"`
	NHead  int `json:"n_head"`
	NEmbd  int `json:"n_embd"`
}

func (c *Client) GetModel(
	ctx context.Context,
	modelID string,
) (*domain.Model, error) {

	modelID = normalizeModelID(modelID)

	metadata, err := c.getMetadata(ctx, modelID)
	if err != nil {
		return nil, err
	}

	config, err := c.getConfig(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf(
			"load model config: %w",
			err,
		)
	}

	model := &domain.Model{
		ID:          metadata.ID,
		Library:     metadata.LibraryName,
		PipelineTag: metadata.PipelineTag,
		Parameters:  metadata.Safetensors.Total,

		ModelType: config.ModelType,

		HiddenSize:            config.HiddenSize,
		HiddenLayers:          config.NumHiddenLayers,
		AttentionHeads:        config.NumAttentionHeads,
		KeyValueHeads:         config.NumKeyValueHeads,
		MaxPositionEmbeddings: config.MaxPositionEmbeddings,
		HeadDim:               config.HeadDim,
	}

	for _, sibling := range metadata.Siblings {
		model.Files = append(
			model.Files,
			sibling.Filename,
		)
	}

	if len(config.Architectures) > 0 {
		model.Architecture = config.Architectures[0]
	}

	applyFallbacks(model, config)

	return model, nil
}

func (c *Client) getMetadata(
	ctx context.Context,
	modelID string,
) (*modelResponse, error) {

	endpoint := fmt.Sprintf(
		"%s/api/models/%s",
		c.baseURL,
		modelID,
	)

	var result modelResponse

	if err := c.getJSON(ctx, endpoint, &result); err != nil {
		return nil, fmt.Errorf(
			"get model metadata: %w",
			err,
		)
	}

	return &result, nil
}

func (c *Client) getConfig(
	ctx context.Context,
	modelID string,
) (*configResponse, error) {

	endpoint := fmt.Sprintf(
		"%s/%s/resolve/main/config.json",
		c.baseURL,
		modelID,
	)

	var result configResponse

	if err := c.getJSON(ctx, endpoint, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) getJSON(
	ctx context.Context,
	endpoint string,
	target any,
) error {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return err
	}

	req.Header.Set(
		"User-Agent",
		"hffit/0.2",
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"HTTP %d from %s",
			resp.StatusCode,
			endpoint,
		)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf(
			"decode JSON: %w",
			err,
		)
	}

	return nil
}

func normalizeModelID(value string) string {
	value = strings.TrimSpace(value)

	value = strings.TrimPrefix(
		value,
		"https://huggingface.co/",
	)

	value = strings.TrimSuffix(value, "/")

	return value
}

func applyFallbacks(
	model *domain.Model,
	config *configResponse,
) {

	if model.HiddenSize == 0 {
		model.HiddenSize = config.NEmbd
	}

	if model.HiddenLayers == 0 {
		model.HiddenLayers = config.NLayer
	}

	if model.AttentionHeads == 0 {
		model.AttentionHeads = config.NHead
	}

	// Si no existe num_key_value_heads asumimos MHA.
	if model.KeyValueHeads == 0 {
		model.KeyValueHeads = model.AttentionHeads
	}

	// Algunos modelos declaran explícitamente head_dim.
	// Si no, lo derivamos.
	if model.HeadDim == 0 &&
		model.HiddenSize > 0 &&
		model.AttentionHeads > 0 {

		model.HeadDim =
			model.HiddenSize / model.AttentionHeads
	}
}
