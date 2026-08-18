package compatibility

import (
	"fmt"

	"github.com/alexperezortuno/hffit/internal/domain"
)

const runtimeOverhead = 1.20

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Calculate(
	model *domain.Model,
	hardware *domain.Hardware,
) []domain.CompatibilityResult {

	var gpuVRAM uint64

	for _, gpu := range hardware.GPUs {
		// MVP:
		// tomamos la GPU con mayor VRAM.
		if gpu.VRAMBytes > gpuVRAM {
			gpuVRAM = gpu.VRAMBytes
		}
	}

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
		required := estimateVRAM(
			model.Parameters,
			precision,
		)

		result := evaluate(
			precision,
			required,
			gpuVRAM,
			hardware.Memory.TotalBytes,
		)

		results = append(results, result)
	}

	return results
}

func estimateVRAM(
	parameters uint64,
	precision domain.Precision,
) uint64 {

	var bytesPerParameter float64

	switch precision {
	case domain.PrecisionFP32:
		bytesPerParameter = 4

	case domain.PrecisionFP16:
		bytesPerParameter = 2

	case domain.PrecisionINT8:
		bytesPerParameter = 1

	case domain.PrecisionINT4:
		bytesPerParameter = 0.5

	default:
		bytesPerParameter = 4
	}

	weights := float64(parameters) * bytesPerParameter

	return uint64(weights * runtimeOverhead)
}

func evaluate(
	precision domain.Precision,
	required uint64,
	vram uint64,
	ram uint64,
) domain.CompatibilityResult {

	result := domain.CompatibilityResult{
		Precision:     precision,
		RequiredVRAM:  required,
		AvailableVRAM: vram,
		AvailableRAM:  ram,
	}

	if vram == 0 {
		return evaluateWithoutGPU(result)
	}

	ratio := float64(required) / float64(vram)

	switch {
	case ratio <= 0.70:
		result.Score = 100
		result.Level = domain.LevelExcellent
		result.CanFitVRAM = true
		result.Message = "Excellent fit with plenty of VRAM headroom"

	case ratio <= 0.85:
		result.Score = 90
		result.Level = domain.LevelExcellent
		result.CanFitVRAM = true
		result.Message = "Excellent fit"

	case ratio <= 1.0:
		result.Score = 80
		result.Level = domain.LevelGood
		result.CanFitVRAM = true
		result.Message = "Fits in GPU VRAM"

	case required <= vram+ram:
		result.Score = 55
		result.Level = domain.LevelLimited
		result.CanFitWithOffload = true
		result.Message = "Requires RAM/CPU offloading"

	case required <= ram:
		result.Score = 40
		result.Level = domain.LevelPoor
		result.Message = "CPU inference may be possible"

	default:
		result.Score = 10
		result.Level = domain.LevelImpossible
		result.Message = "Insufficient system memory"
	}

	return result
}

func evaluateWithoutGPU(
	result domain.CompatibilityResult,
) domain.CompatibilityResult {

	if result.RequiredVRAM <= result.AvailableRAM {
		result.Score = 35
		result.Level = domain.LevelPoor
		result.Message = "No supported GPU detected; CPU inference may be possible"

		return result
	}

	result.Score = 0
	result.Level = domain.LevelImpossible
	result.Message = fmt.Sprintf(
		"requires more memory than available",
	)

	return result
}
