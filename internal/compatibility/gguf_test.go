package compatibility

import (
	"testing"

	"github.com/alexperezortuno/hffit/internal/domain"
)

func TestDetectQuantization(t *testing.T) {

	tests := []struct {
		filename string
		expected string
	}{
		{
			"Qwen3-8B-Q4_K_M.gguf",
			"Q4_K_M",
		},
		{
			"Qwen3-8B-Q5_K_S.gguf",
			"Q5_K_S",
		},
		{
			"model-Q8_0.gguf",
			"Q8_0",
		},
		{
			"MODEL-q3_k_m.GGUF",
			"Q3_K_M",
		},
	}

	for _, tt := range tests {

		t.Run(
			tt.filename,
			func(t *testing.T) {

				got :=
					detectQuantization(
						tt.filename,
					)

				if got != tt.expected {

					t.Fatalf(
						"expected %s, got %s",
						tt.expected,
						got,
					)
				}
			},
		)
	}
}

func TestDiscoverGGUF(t *testing.T) {

	model :=
		&domain.Model{
			Files: []string{
				"README.md",
				"config.json",
				"model-Q4_K_M.gguf",
				"model-Q8_0.gguf",
				"model.safetensors",
			},
		}

	results :=
		DiscoverGGUF(model)

	if len(results) != 2 {

		t.Fatalf(
			"expected 2 GGUF files, got %d",
			len(results),
		)
	}

	if results[0].Quantization != "Q8_0" {

		t.Fatalf(
			"expected highest quality first, got %s",
			results[0].Quantization,
		)
	}
}
