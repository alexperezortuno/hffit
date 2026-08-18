package compatibility

import "github.com/alexperezortuno/hffit/internal/domain"

func estimateKVCache(
	model *domain.Model,
	contextSize int,
	bytesPerElement uint64,
) uint64 {

	if contextSize <= 0 {
		return 0
	}

	if model.HiddenLayers <= 0 ||
		model.KeyValueHeads <= 0 ||
		model.HeadDim <= 0 {

		return 0
	}

	/*
		KV cache:

		2
		× layers
		× context
		× kv_heads
		× head_dim
		× bytes

		2 = Key + Value

		Batch size = 1
	*/

	return uint64(2) *
		uint64(model.HiddenLayers) *
		uint64(contextSize) *
		uint64(model.KeyValueHeads) *
		uint64(model.HeadDim) *
		bytesPerElement
}
