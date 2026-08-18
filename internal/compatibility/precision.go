package compatibility

import "github.com/alexperezortuno/hffit/internal/domain"

func weightBytesPerParameter(
	precision domain.Precision,
) float64 {

	switch precision {

	case domain.PrecisionFP32:
		return 4

	case domain.PrecisionFP16:
		return 2

	case domain.PrecisionINT8:
		return 1

	case domain.PrecisionINT4:
		return 0.5

	default:
		return 4
	}
}

func kvBytesPerElement(
	precision domain.Precision,
) uint64 {

	switch precision {

	case domain.PrecisionFP32:
		return 4

	default:
		// Inicialmente BF16 / FP16
		return 2
	}
}
