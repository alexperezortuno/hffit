package compatibility

import (
	"github.com/alexperezortuno/hffit/internal/domain"
)

type GGUFCalculator struct {
	contextSize int
}

func NewGGUFCalculator(
	contextSize int,
) *GGUFCalculator {

	return &GGUFCalculator{
		contextSize: contextSize,
	}
}

func (c *GGUFCalculator) Calculate(
	model *domain.Model,
	hardware *domain.Hardware,
	variants []domain.GGUFVariant,
) []domain.GGUFVariant {

	vram :=
		largestGPUVRAM(hardware)

	for i := range variants {

		variant :=
			&variants[i]

		weights :=
			estimateGGUFWeights(
				model.Parameters,
				variant.BitsPerWeight,
			)

		/*
			Por ahora asumimos KV cache FP16.

			Después permitiremos:
			--kv-type f16
			--kv-type q8
			--kv-type q4
		*/
		kv :=
			estimateKVCache(
				model,
				c.contextSize,
				2,
			)

		/*
			llama.cpp necesita buffers adicionales.

			Para MVP dejamos 5% de los pesos.
		*/
		overhead :=
			uint64(
				float64(weights) * 0.05,
			)

		required :=
			weights +
				kv +
				overhead

		variant.EstimatedWeightsBytes =
			weights

		variant.EstimatedVRAMBytes =
			required

		evaluateGGUF(
			variant,
			required,
			vram,
			hardware.Memory.TotalBytes,
		)
	}

	markRecommended(
		variants,
	)

	return variants
}

func estimateGGUFWeights(
	parameters uint64,
	bitsPerWeight float64,
) uint64 {

	if bitsPerWeight <= 0 {
		return 0
	}

	bits :=
		float64(parameters) *
			bitsPerWeight

	return uint64(bits / 8)
}

func evaluateGGUF(
	result *domain.GGUFVariant,
	required uint64,
	vram uint64,
	ram uint64,
) {

	if vram == 0 {

		if required <= ram {

			result.Score = 40
			result.Level =
				domain.LevelPoor

			result.Message =
				"CPU inference only"

			return
		}

		result.Score = 0
		result.Level =
			domain.LevelImpossible

		result.Message =
			"Insufficient RAM"

		return
	}

	ratio :=
		float64(required) /
			float64(vram)

	switch {

	case ratio <= 0.65:

		result.Score = 100
		result.Level =
			domain.LevelExcellent

		result.Message =
			"Excellent VRAM headroom"

	case ratio <= 0.80:

		result.Score = 95
		result.Level =
			domain.LevelExcellent

		result.Message =
			"Excellent fit"

	case ratio <= 0.92:

		result.Score = 85
		result.Level =
			domain.LevelGood

		result.Message =
			"Good fit"

	case ratio <= 1:

		result.Score = 75
		result.Level =
			domain.LevelGood

		result.Message =
			"Fits with little VRAM headroom"

	case required <= vram+ram:

		result.Score = 55
		result.Level =
			domain.LevelLimited

		result.Message =
			"Partial CPU offload required"

	default:

		result.Score = 0
		result.Level =
			domain.LevelImpossible

		result.Message =
			"Insufficient memory"
	}
}

func markRecommended(
	variants []domain.GGUFVariant,
) {

	best := -1

	for i := range variants {

		variant :=
			variants[i]

		/*
			Queremos mínimo Good.

			Como vienen ordenados desde mayor
			calidad/bpw a menor, tomamos el
			primero que tenga >= 80.
		*/
		if variant.Score >= 80 {

			best = i
			break
		}
	}

	/*
		Si ninguno alcanza 80, usamos cualquier
		variante que pueda ejecutarse.
	*/
	if best == -1 {

		for i := range variants {

			if variants[i].Score >= 50 {
				best = i
				break
			}
		}
	}

	if best >= 0 {

		variants[best].Recommended =
			true
	}
}
