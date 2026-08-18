package compatibility

import (
	"testing"

	"github.com/alexperezortuno/hffit/internal/domain"
)

func TestEstimateKVCache(t *testing.T) {

	model := &domain.Model{
		HiddenLayers:   36,
		AttentionHeads: 32,
		KeyValueHeads:  8,
		HiddenSize:     4096,
		HeadDim:        128,
	}

	got := estimateKVCache(
		model,
		8192,
		2,
	)

	expected :=
		uint64(2) *
			36 *
			8192 *
			8 *
			128 *
			2

	if got != expected {

		t.Fatalf(
			"expected %d bytes, got %d",
			expected,
			got,
		)
	}
}
