package compatibility

import (
	"github.com/alexperezortuno/hffit/internal/domain"
)

const runtimeOverheadRatio = 0.10

type Options struct {
	ContextSize int
}

type Calculator struct {
	options Options
}

func NewCalculator(
	options Options,
) *Calculator {

	return &Calculator{
		options: options,
	}
}

func (c *Calculator) Calculate(
	model *domain.Model,
	hardware *domain.Hardware,
) []domain.CompatibilityResult {

	gpuVRAM := largestGPUVRAM(hardware)

	precisions := []domain.Precision{
		domain.PrecisionFP32,
		domain.PrecisionFP16,
		domain.PrecisionINT8,
		domain.PrecisionINT4,
	}

	results := make(
		[]domain.CompatibilityResult,
		0,
		len(precisions),
	)

	for _, precision := range precisions {

		weights := estimateWeights(
			model.Parameters,
			precision,
		)

		kvCache := estimateKVCache(
			model,
			c.options.ContextSize,
			kvBytesPerElement(precision),
		)

		overhead := estimateRuntimeOverhead(
			weights,
		)

		required :=
			weights +
				kvCache +
				overhead

		result := evaluate(
			precision,
			weights,
			kvCache,
			overhead,
			required,
			gpuVRAM,
			hardware.Memory.TotalBytes,
			c.options.ContextSize,
		)

		results = append(
			results,
			result,
		)
	}

	return results
}

func largestGPUVRAM(
	hardware *domain.Hardware,
) uint64 {

	var largest uint64

	for _, gpu := range hardware.GPUs {
		if gpu.VRAMBytes > largest {
			largest = gpu.VRAMBytes
		}
	}

	return largest
}

func estimateWeights(
	parameters uint64,
	precision domain.Precision,
) uint64 {

	bytes :=
		float64(parameters) *
			weightBytesPerParameter(precision)

	return uint64(bytes)
}

func estimateRuntimeOverhead(
	weights uint64,
) uint64 {

	return uint64(
		float64(weights) *
			runtimeOverheadRatio,
	)
}

func evaluate(
	precision domain.Precision,
	weights uint64,
	kvCache uint64,
	overhead uint64,
	required uint64,
	vram uint64,
	ram uint64,
	contextSize int,
) domain.CompatibilityResult {

	result := domain.CompatibilityResult{
		Precision: precision,

		WeightsBytes:  weights,
		KVCacheBytes:  kvCache,
		OverheadBytes: overhead,
		RequiredVRAM:  required,

		AvailableVRAM: vram,
		AvailableRAM:  ram,

		ContextSize: contextSize,
	}

	if vram == 0 {
		return evaluateWithoutGPU(result)
	}

	ratio :=
		float64(required) /
			float64(vram)

	switch {

	case ratio <= 0.70:

		result.Score = 100
		result.Level = domain.LevelExcellent
		result.CanFitVRAM = true

		result.Message =
			"Excellent fit with VRAM headroom"

	case ratio <= 0.85:

		result.Score = 90
		result.Level = domain.LevelExcellent
		result.CanFitVRAM = true

		result.Message =
			"Excellent fit"

	case ratio <= 1.0:

		result.Score = 80
		result.Level = domain.LevelGood
		result.CanFitVRAM = true

		result.Message =
			"Fits in GPU VRAM"

	case required <= vram+ram:

		result.Score = 55
		result.Level = domain.LevelLimited
		result.CanFitWithOffload = true

		result.Message =
			"Requires CPU/RAM offloading"

	case required <= ram:

		result.Score = 40
		result.Level = domain.LevelPoor

		result.Message =
			"CPU inference may be possible"

	default:

		result.Score = 0
		result.Level = domain.LevelImpossible

		result.Message =
			"Insufficient system memory"
	}

	return result
}

func evaluateWithoutGPU(
	result domain.CompatibilityResult,
) domain.CompatibilityResult {

	if result.RequiredVRAM <= result.AvailableRAM {

		result.Score = 35
		result.Level = domain.LevelPoor

		result.Message =
			"No supported GPU detected; CPU inference may be possible"

		return result
	}

	result.Score = 0

	result.Level =
		domain.LevelImpossible

	result.Message =
		"Insufficient system RAM"

	return result
}
